package devsecops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

type workflow struct {
	Name        string            `yaml:"name"`
	On          map[string]any    `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Concurrency struct {
		Group            string `yaml:"group"`
		CancelInProgress bool   `yaml:"cancel-in-progress"`
	} `yaml:"concurrency"`
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	TimeoutMinutes  int               `yaml:"timeout-minutes"`
	Permissions     map[string]string `yaml:"permissions"`
	Steps           []step            `yaml:"steps"`
	ContinueOnError bool              `yaml:"continue-on-error"`
	If              string            `yaml:"if"`
	Needs           any               `yaml:"needs"`
	Env             map[string]string `yaml:"env"`
	Defaults        any               `yaml:"defaults"`
	Container       any               `yaml:"container"`
}

type step struct {
	Name             string            `yaml:"name"`
	Uses             string            `yaml:"uses"`
	Run              string            `yaml:"run"`
	WorkingDirectory string            `yaml:"working-directory"`
	With             map[string]any    `yaml:"with"`
	Env              map[string]string `yaml:"env"`
	Shell            string            `yaml:"shell"`
	TimeoutMinutes   int               `yaml:"timeout-minutes"`
	ContinueOnError  bool              `yaml:"continue-on-error"`
	If               string            `yaml:"if"`
	Container        any               `yaml:"container"`
}

// pinnedScannerStepShell pins the execution shell for every scanner step
// whose body the policy pins. No pinned scanner step may carry step-level
// env: env alters execution without tripping the body pin, so any deviation
// fails closed.
var pinnedScannerStepShell = map[string]string{
	"Check vulnerabilities including tests":                          "",
	"Check nested cloud bootstrap vulnerabilities including tests":   "",
	"Check nested adapter vulnerabilities including tests":           "",
	"Install pinned security tools":                                  "",
	"Check test-inclusive dependency licenses":                       "bash",
	"Check nested adapter test-inclusive dependency licenses":        "bash",
	"Verify license policy rejects a test-only forbidden dependency": "",
	"Lint workflow files":                                            "",
	"Install pinned Gitleaks":                                        "",
	"Verify Gitleaks clean, finding, and error behavior":             "",
	"Reject unmanaged Gitleaks config and ignore files":              "",
	"Gitleaks full-history scan":                                     "",
}

type exceptionFile struct {
	Version    int         `yaml:"version"`
	Exceptions []exception `yaml:"exceptions"`
}

type exception struct {
	ID      string   `yaml:"id"`
	Paths   []string `yaml:"paths"`
	Rules   []string `yaml:"rules"`
	Reason  string   `yaml:"reason"`
	Owner   string   `yaml:"owner"`
	BeadsID string   `yaml:"beads_id"`
	Expires string   `yaml:"expires"`
}

var (
	actionSHAPattern  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	beadsIDPattern    = regexp.MustCompile(`^dap-[a-z0-9]+(?:\.[a-z0-9]+)*$`)
	ownerPattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{1,38}$`)
	probeSleepPattern = regexp.MustCompile(`^sleep ([0-9]+)m$`)
)

func readWorkflow(t *testing.T, name string) (workflow, string) {
	t.Helper()
	path := filepath.Join("..", "..", ".github", "workflows", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got workflow
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return got, string(data)
}

func TestSecurityWorkflowUsesExpectedTriggersAndLeastPrivilege(t *testing.T) {
	got, raw := readWorkflow(t, "security-scans.yml")
	if err := validateWorkflow(got); err != nil {
		t.Fatal(err)
	}
	if err := validateActionPins(got, raw); err != nil {
		t.Fatal(err)
	}
	if err := validateMainAndWeeklyTriggers(got.On); err != nil {
		t.Fatal(err)
	}
	for jobName, configured := range got.Jobs {
		if configured.TimeoutMinutes < 5 || configured.TimeoutMinutes > 30 {
			t.Errorf("job %s timeout must be between 5 and 30 minutes", jobName)
		}
		if len(configured.Permissions) != 0 {
			t.Errorf("job %s unexpectedly overrides the read-only workflow token", jobName)
		}
	}
}

func TestNestedAdapterCarriesRootLicense(t *testing.T) {
	rootLicense, err := os.ReadFile(filepath.Join("..", "..", "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	adapterLicense, err := os.ReadFile(filepath.Join("..", "..", "ci", "pdfua2", "adapter", "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	if string(adapterLicense) != string(rootLicense) {
		t.Fatal("nested CI-only module license must match the repository license")
	}
}

func TestScannerStepsFailClosedAndUseFindingSensitiveModes(t *testing.T) {
	got, raw := readWorkflow(t, "security-scans.yml")
	assertGovulncheckStep(t, got, "Check vulnerabilities including tests", "")
	assertGovulncheckStep(t, got, "Check nested cloud bootstrap vulnerabilities including tests", "cmd/cloud-tools-bootstrap")
	assertGovulncheckStep(t, got, "Check nested adapter vulnerabilities including tests", "ci/pdfua2/adapter")
	if err := validateScannerSteps(got); err != nil {
		t.Fatal(err)
	}

	badCases := []struct {
		name   string
		mutate func(*workflow)
	}{
		{"govulncheck JSON mode exits successfully on findings", func(w *workflow) {
			mutateRun(w, "Check vulnerabilities including tests", func(s *step) { s.Run = "govulncheck -test -json ./..." })
		}},
		{"nested adapter module omitted from vulnerability scan", func(w *workflow) {
			mutateRun(w, "Check nested adapter vulnerabilities including tests", func(s *step) { s.WorkingDirectory = "." })
		}},
		{"nested cloud bootstrap module omitted from vulnerability scan", func(w *workflow) {
			mutateRun(w, "Check nested cloud bootstrap vulnerabilities including tests", func(s *step) { s.WorkingDirectory = "." })
		}},
		{"nested cloud bootstrap scan uses finding-insensitive output", func(w *workflow) {
			mutateRun(w, "Check nested cloud bootstrap vulnerabilities including tests", func(s *step) { s.Run = "govulncheck -test -json ./..." })
		}},
		{"nested cloud bootstrap scanner finding is swallowed", func(w *workflow) {
			mutateRun(w, "Check nested cloud bootstrap vulnerabilities including tests", func(s *step) { s.Run += " || true" })
		}},
		{"nested cloud bootstrap scanner continues after failure", func(w *workflow) {
			mutateRun(w, "Check nested cloud bootstrap vulnerabilities including tests", func(s *step) { s.ContinueOnError = true })
		}},
		{"scanner finding is swallowed", func(w *workflow) {
			mutateRun(w, "Check vulnerabilities including tests", func(s *step) { s.Run += " || true" })
		}},
		{"scanner step continues after failure", func(w *workflow) {
			mutateRun(w, "Check vulnerabilities including tests", func(s *step) { s.ContinueOnError = true })
		}},
		{"Gitleaks report may expose secrets", func(w *workflow) {
			mutateRun(w, "Gitleaks full-history scan", func(s *step) { s.Run += " --report-format=json --report-path=report.json" })
		}},
		{"Gitleaks inline allow comments bypass scan", func(w *workflow) {
			mutateRun(w, "Gitleaks full-history scan", func(s *step) { s.Run = strings.ReplaceAll(s.Run, " --ignore-gitleaks-allow", "") })
		}},
		{"Gitleaks repository config can replace defaults", func(w *workflow) {
			mutateRun(w, "Gitleaks full-history scan", func(s *step) { s.Run = strings.ReplaceAll(s.Run, "--config=.security/gitleaks.toml", "") })
		}},
		{"Gitleaks source-level suppression files are not rejected", func(w *workflow) {
			mutateRun(w, "Reject unmanaged Gitleaks config and ignore files", func(s *step) { s.Run = "printf 'suppression guard skipped'" })
		}},
		{"Gitleaks ignore file can suppress findings", func(w *workflow) {
			mutateRun(w, "Gitleaks full-history scan", func(s *step) { s.Run = strings.ReplaceAll(s.Run, "--gitleaks-ignore-path=/dev/null", "") })
		}},
		{"license scan omits test dependencies", func(w *workflow) {
			mutateRun(w, "Check test-inclusive dependency licenses", func(s *step) { s.Run = strings.ReplaceAll(s.Run, "--include_tests ", "") })
		}},
		{"license scan fails open", func(w *workflow) {
			mutateRun(w, "Check test-inclusive dependency licenses", func(s *step) { s.Run += " || true" })
		}},
		{"license scan does not use explicit allowlist", func(w *workflow) {
			mutateRun(w, "Check test-inclusive dependency licenses", func(s *step) { s.Run = "go-licenses check --include_tests ./..." })
		}},
		{"license scan continues after failure", func(w *workflow) {
			mutateRun(w, "Check test-inclusive dependency licenses", func(s *step) { s.ContinueOnError = true })
		}},
		{"nested adapter module omitted from license scan", func(w *workflow) {
			mutateRun(w, "Check nested adapter test-inclusive dependency licenses", func(s *step) { s.WorkingDirectory = "." })
		}},
		{"actionlint failure is ignored", func(w *workflow) {
			mutateRun(w, "Lint workflow files", func(s *step) { s.Run += " || true" })
		}},
		{"actionlint invocation is missing", func(w *workflow) {
			mutateRun(w, "Lint workflow files", func(s *step) { s.Run = "printf 'lint skipped'" })
		}},
	}
	for _, tc := range badCases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := got
			mutated.Jobs = cloneJobs(got.Jobs)
			tc.mutate(&mutated)
			if err := validateScannerSteps(mutated); err == nil {
				t.Fatal("expected policy rejection")
			}
		})
	}

	condition := strings.Replace(raw, "      - name: Gitleaks full-history scan\n", "      - name: Gitleaks full-history scan\n        if: ${{ false }}\n", 1)
	var conditional workflow
	if err := yaml.Unmarshal([]byte(condition), &conditional); err != nil {
		t.Fatal(err)
	}
	if err := validateScannerSteps(conditional); err == nil {
		t.Fatal("scanner step guarded by a condition was accepted")
	}

	conditionalGovulncheck := strings.Replace(raw, "      - name: Check vulnerabilities including tests\n", "      - name: Check vulnerabilities including tests\n        if: ${{ false }}\n", 1)
	var conditionalGovuln workflow
	if err := yaml.Unmarshal([]byte(conditionalGovulncheck), &conditionalGovuln); err != nil {
		t.Fatal(err)
	}
	if err := validateScannerSteps(conditionalGovuln); err == nil {
		t.Fatal("govulncheck step guarded by a condition was accepted")
	}

	failOpenJob := strings.Replace(raw, "  go-security:\n", "  go-security:\n    continue-on-error: true\n", 1)
	var failOpen workflow
	if err := yaml.Unmarshal([]byte(failOpenJob), &failOpen); err != nil {
		t.Fatal(err)
	}
	if err := validateWorkflow(failOpen); err == nil {
		t.Fatal("job-level continue-on-error was accepted")
	}
	disabledSecurityJob := strings.Replace(raw, "  go-security:\n", "  go-security:\n    if: ${{ false }}\n", 1)
	var disabled workflow
	if err := yaml.Unmarshal([]byte(disabledSecurityJob), &disabled); err != nil {
		t.Fatal(err)
	}
	if err := validateWorkflow(disabled); err == nil {
		t.Fatal("security job guarded by a condition was accepted")
	}
}

func assertGovulncheckStep(t *testing.T, w workflow, name, workingDirectory string) {
	t.Helper()
	for _, configured := range w.Jobs {
		for _, s := range configured.Steps {
			if s.Name != name {
				continue
			}
			if s.WorkingDirectory != workingDirectory {
				t.Fatalf("%s working-directory = %q, want %q", name, s.WorkingDirectory, workingDirectory)
			}
			if strings.TrimSpace(s.Run) != "govulncheck -test ./..." {
				t.Fatalf("%s run = %q, want test-inclusive exit-status govulncheck", name, s.Run)
			}
			if s.ContinueOnError || strings.TrimSpace(s.If) != "" || strings.Contains(s.Run, "|| true") {
				t.Fatalf("%s must be unconditional and fail closed", name)
			}
			return
		}
	}
	t.Fatalf("required govulncheck step %q was not found", name)
}

func TestSecurityScannersUseSeparateRunnerFromProjectTests(t *testing.T) {
	security, _ := readWorkflow(t, "security-scans.yml")
	project, _ := readWorkflow(t, "go-checks.yml")
	if err := validateScannerIsolation(security, project); err != nil {
		t.Fatal(err)
	}

	t.Run("project test added to scanner job", func(t *testing.T) {
		mutated := security
		mutated.Jobs = cloneJobs(security.Jobs)
		scanner := mutated.Jobs["go-security"]
		scanner.Steps = append(scanner.Steps, step{Name: "Run PR-controlled tests", Run: "go test -race ./..."})
		mutated.Jobs["go-security"] = scanner
		if err := validateScannerIsolation(mutated, project); err == nil || !strings.Contains(err.Error(), "executes project Go code") {
			t.Fatalf("scanner job running PR-controlled tests rejection = %v, want project-code isolation error", err)
		}
	})

	t.Run("project tests removed from isolated workflow", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		checks := mutated.Jobs["checks"]
		for i, candidate := range checks.Steps {
			if candidate.Name == "Test with race detector" {
				checks.Steps = append(checks.Steps[:i], checks.Steps[i+1:]...)
				break
			}
		}
		mutated.Jobs["checks"] = checks
		if err := validateScannerIsolation(security, mutated); err == nil || !strings.Contains(err.Error(), "must keep") {
			t.Fatalf("workflow without isolated project tests rejection = %v, want missing-project-check error", err)
		}
	})
	t.Run("root module tidy check removed from isolated workflow", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		checks := mutated.Jobs["checks"]
		for i, candidate := range checks.Steps {
			if candidate.Name == "Check module files" {
				checks.Steps = append(checks.Steps[:i], checks.Steps[i+1:]...)
				break
			}
		}
		mutated.Jobs["checks"] = checks
		if err := validateScannerIsolation(security, mutated); err == nil || !strings.Contains(err.Error(), `keep "Check module files" running "go mod tidy -diff"`) {
			t.Fatalf("workflow without root module tidy check rejection = %v, want missing-root-tidy error", err)
		}
	})

	t.Run("scanner Go cache shared with project tests", func(t *testing.T) {
		mutated := security
		mutated.Jobs = cloneJobs(security.Jobs)
		scanner := mutated.Jobs["go-security"]
		for i := range scanner.Steps {
			if scanner.Steps[i].Name == "Set up Go" {
				scanner.Steps[i].With["cache"] = true
			}
		}
		mutated.Jobs["go-security"] = scanner
		if err := validateScannerIsolation(mutated, project); err == nil || !strings.Contains(err.Error(), "must not restore or save a Go build cache") {
			t.Fatalf("scanner job with shared Go cache rejection = %v, want cache-isolation error", err)
		}
		original := security.Jobs["go-security"]
		for _, candidate := range original.Steps {
			if candidate.Name == "Set up Go" && candidate.With["cache"] != false {
				t.Fatalf("mutating cloned scanner job changed original setup-go cache setting to %#v", candidate.With["cache"])
			}
		}
	})

	t.Run("scanner binaries installed outside the runner temp directory", func(t *testing.T) {
		mutated := security
		mutated.Jobs = cloneJobs(security.Jobs)
		mutateRun(&mutated, "Install pinned security tools", func(s *step) {
			s.Run = strings.ReplaceAll(s.Run, "GOBIN=\"$tool_dir\" ", "")
		})
		if err := validateScannerIsolation(mutated, project); err == nil || !strings.Contains(err.Error(), "private RUNNER_TEMP directory") {
			t.Fatalf("scanner installation without private runner-temp GOBIN rejection = %v, want GOBIN-isolation error", err)
		}
	})
}

// TestGoChecksModuleGateRemovalIsRejected is the dap-jo9.5.6 negative proof:
// removing any of the nine root and nested-module gates from the go-checks
// "checks" job must fail validateScannerIsolation with an error naming the
// removed gate. The unmutated presence check keeps each subtest from going
// vacuous if a gate step is ever renamed rather than removed.
func TestGoChecksModuleGateRemovalIsRejected(t *testing.T) {
	security, _ := readWorkflow(t, "security-scans.yml")
	project, _ := readWorkflow(t, "go-checks.yml")

	for _, gate := range requiredGoChecksModuleGates {
		gate := gate
		t.Run(gate.name, func(t *testing.T) {
			matches := func(candidate step) bool {
				return candidate.Name == gate.name &&
					candidate.WorkingDirectory == gate.workingDirectory &&
					strings.TrimSpace(candidate.Run) == gate.command
			}

			// The gate must exist unmutated, or the removal probe proves nothing.
			unmutatedFound := false
			for _, candidate := range project.Jobs["checks"].Steps {
				if matches(candidate) {
					unmutatedFound = true
					break
				}
			}
			if !unmutatedFound {
				t.Fatalf("gate %q running %q (working directory %q) missing from the go-checks job; test is vacuous against a renamed step", gate.name, gate.command, gate.workingDirectory)
			}

			mutated := project
			mutated.Jobs = cloneJobs(project.Jobs)
			checks := mutated.Jobs["checks"]
			removed := false
			for i, candidate := range checks.Steps {
				if matches(candidate) {
					checks.Steps = append(checks.Steps[:i], checks.Steps[i+1:]...)
					removed = true
					break
				}
			}
			if !removed {
				t.Fatalf("could not splice gate %q out of the cloned go-checks job", gate.name)
			}
			mutated.Jobs["checks"] = checks

			err := validateScannerIsolation(security, mutated)
			if err == nil || !strings.Contains(err.Error(), gate.name) {
				t.Fatalf("workflow without gate %q rejection = %v, want error naming the removed gate", gate.name, err)
			}
			if !strings.Contains(err.Error(), gate.command) {
				t.Fatalf("workflow without gate %q rejection = %v, want error naming the removed command", gate.name, err)
			}
		})
	}
}

// TestGoChecksGatesRejectDisabledSteps is the F1 negative proof: a pinned
// go-checks gate step neutered with `if:` or `continue-on-error:` must fail
// validateScannerIsolation, naming the neutered step. The gate body pin
// (name/working-directory/run) is byte-identical in both bypass variants, so
// the validator must read those parsed fields on this path.
func TestGoChecksGatesRejectDisabledSteps(t *testing.T) {
	const gate = "Test with race detector"
	security, _ := readWorkflow(t, "security-scans.yml")
	project, _ := readWorkflow(t, "go-checks.yml")

	t.Run("if-guarded gate step is rejected", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		mutateRun(&mutated, gate, func(s *step) { s.If = "${{ false }}" })
		err := validateScannerIsolation(security, mutated)
		if err == nil {
			t.Fatal("if-guarded go-checks gate step was accepted")
		}
		if !strings.Contains(err.Error(), gate) {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("continue-on-error gate step is rejected", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		mutateRun(&mutated, gate, func(s *step) { s.ContinueOnError = true })
		err := validateScannerIsolation(security, mutated)
		if err == nil {
			t.Fatal("continue-on-error go-checks gate step was accepted")
		}
		if !strings.Contains(err.Error(), gate) {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})
}

// TestGoChecksJobRejectsJobLevelDisablers is the F2 negative proof: the
// go-checks "checks" job neutered with a job-level `if:` or
// `continue-on-error:` disables all nine module gates at once, so
// validateScannerIsolation must reject the workflow naming the job. The
// fields are parsed on the job struct already; they are just never read on
// this path yet.
func TestGoChecksJobRejectsJobLevelDisablers(t *testing.T) {
	security, _ := readWorkflow(t, "security-scans.yml")
	project, _ := readWorkflow(t, "go-checks.yml")

	t.Run("if-guarded checks job is rejected", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		checks := mutated.Jobs["checks"]
		checks.If = "${{ false }}"
		mutated.Jobs["checks"] = checks
		err := validateScannerIsolation(security, mutated)
		if err == nil {
			t.Fatal("if-guarded go-checks checks job was accepted")
		}
		if !strings.Contains(err.Error(), "checks") {
			t.Fatalf("rejection did not name the job: %v", err)
		}
	})

	t.Run("continue-on-error checks job is rejected", func(t *testing.T) {
		mutated := project
		mutated.Jobs = cloneJobs(project.Jobs)
		checks := mutated.Jobs["checks"]
		checks.ContinueOnError = true
		mutated.Jobs["checks"] = checks
		err := validateScannerIsolation(security, mutated)
		if err == nil {
			t.Fatal("continue-on-error go-checks checks job was accepted")
		}
		if !strings.Contains(err.Error(), "checks") {
			t.Fatalf("rejection did not name the job: %v", err)
		}
	})
}

// TestGoChecksJobRejectsUnmodeledExecutionKeys is the F3 negative proof:
// job-level `env:`/`defaults:`/`container:` and step-level `container:` are
// silently dropped by non-strict yaml.Unmarshal, so every pin is bypassable
// one level up (job env inherited by all gates, defaults.run.shell
// overriding pinned shells, container running gates in an attacker image).
// Each bypass is constructed at YAML level and must fail
// validateScannerIsolation, naming the job or the step. The pinned
// go-checks.yml uses none of these keys, so the policy forbids them outright.
func TestGoChecksJobRejectsUnmodeledExecutionKeys(t *testing.T) {
	security, _ := readWorkflow(t, "security-scans.yml")
	_, raw := readWorkflow(t, "go-checks.yml")
	const gate = "Test with race detector"

	cases := []struct {
		name   string
		anchor string
		insert string
		wantIn string
	}{
		{
			name:   "job-level env is rejected",
			anchor: "  checks:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n",
			insert: "    env:\n      GOFLAGS: -run=^$\n",
			wantIn: "checks",
		},
		{
			name:   "job-level defaults is rejected",
			anchor: "  checks:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n",
			insert: "    defaults:\n      run:\n        shell: sh\n",
			wantIn: "checks",
		},
		{
			name:   "job-level container map is rejected",
			anchor: "  checks:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n",
			insert: "    container:\n      image: ghcr.io/attacker/runner:latest\n",
			wantIn: "checks",
		},
		{
			name:   "step-level container map is rejected",
			anchor: "      - name: Test with race detector\n",
			insert: "        container:\n          image: ghcr.io/attacker/runner:latest\n",
			wantIn: gate,
		},
		{
			name:   "step-level container string is rejected",
			anchor: "      - name: Test with race detector\n",
			insert: "        container: ghcr.io/attacker/runner:latest\n",
			wantIn: gate,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(raw, tc.anchor, tc.anchor+tc.insert, 1)
			if mutated == raw {
				t.Fatalf("anchor not found for %s", tc.name)
			}
			var project workflow
			if err := yaml.Unmarshal([]byte(mutated), &project); err != nil {
				t.Fatal(err)
			}
			err := validateScannerIsolation(security, project)
			if err == nil {
				t.Fatalf("%s was accepted", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Fatalf("rejection did not name the job/step: %v", err)
			}
		})
	}
}

func TestProjectWorkspaceMutationCannotChangeScannerToolSelection(t *testing.T) {
	scannerRunnerTemp := t.TempDir()
	scannerTools := filepath.Join(scannerRunnerTemp, "security-tools")
	projectRunnerWorkspace := t.TempDir()
	projectBin := filepath.Join(projectRunnerWorkspace, "bin")
	for _, dir := range []string{scannerTools, projectBin} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	scannerTool := filepath.Join(scannerTools, "govulncheck")
	projectTool := filepath.Join(projectBin, "govulncheck")
	if err := os.WriteFile(scannerTool, []byte("#!/bin/sh\nprintf '%s\\n' scanner-pinned-tool\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectTool, []byte("#!/bin/sh\nprintf '%s\\n' project-tool\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	// Model a PR-controlled project test replacing its own same-named tool.
	if err := os.WriteFile(projectTool, []byte("#!/bin/sh\nprintf '%s\\n' project-replaced-tool\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	projectCmd := exec.Command("/bin/sh", "-c", "govulncheck")
	projectCmd.Env = append(os.Environ(), "PATH="+projectBin+string(os.PathListSeparator)+"/usr/bin:/bin")
	projectOutput, err := projectCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run project-selected dummy tool: %v; output: %s", err, projectOutput)
	}
	if got := strings.TrimSpace(string(projectOutput)); got != "project-replaced-tool" {
		t.Fatalf("project PATH selected %q, want project-replaced-tool", got)
	}

	cmd := exec.Command("/bin/sh", "-c", "govulncheck")
	cmd.Env = append(os.Environ(), "PATH="+scannerTools+string(os.PathListSeparator)+projectBin+string(os.PathListSeparator)+"/usr/bin:/bin")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run scanner-selected tool: %v; output: %s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != "scanner-pinned-tool" {
		t.Fatalf("scanner PATH selected %q after project workspace mutation; want scanner-pinned-tool", got)
	}
}

func TestGoWorkflowDefinesBoundedTimeout(t *testing.T) {
	got, raw := readWorkflow(t, "go-checks.yml")
	if err := validateJobTimeouts(got); err != nil {
		t.Fatalf("Go checks workflow must have a valid timeout: %v", err)
	}
	configured := got.Jobs["checks"].TimeoutMinutes
	if configured < 5 || configured > 30 {
		t.Fatalf("Go checks timeout = %d minutes, want 5..30", configured)
	}
	mutations := []struct {
		name string
		raw  string
	}{
		{"missing timeout", strings.Replace(raw, "    timeout-minutes: 20\n", "", 1)},
		{"timeout too large", strings.Replace(raw, "    timeout-minutes: 20\n", "    timeout-minutes: 31\n", 1)},
		{"timeout too small", strings.Replace(raw, "    timeout-minutes: 20\n", "    timeout-minutes: 4\n", 1)},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			var mutated workflow
			if err := yaml.Unmarshal([]byte(tc.raw), &mutated); err != nil {
				t.Fatal(err)
			}
			if err := validateJobTimeouts(mutated); err == nil {
				t.Fatalf("policy accepted %s", tc.name)
			}
		})
	}
}

func TestTimeoutCancellationProbeMatchesGoChecksBound(t *testing.T) {
	goChecks, _ := readWorkflow(t, "go-checks.yml")
	probe, raw := readWorkflow(t, "timeout-cancellation-probe.yml")
	productionBound := goChecks.Jobs["checks"].TimeoutMinutes
	if err := validateTimeoutCancellationProbe(probe, productionBound); err != nil {
		t.Fatalf("timeout cancellation probe must exercise the Go checks bound: %v", err)
	}

	mutations := []struct {
		name string
		edit func(*workflow)
	}{
		{"timeout differs from production", func(w *workflow) {
			job := w.Jobs["probe"]
			job.TimeoutMinutes--
			w.Jobs["probe"] = job
		}},
		{"probe no longer stalls past the bound", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Steps[0].Run = "sleep 19m"
			w.Jobs["probe"] = job
		}},
		{"compound sleep duration is unsupported", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Steps[0].Run = "sleep 25m0s"
			w.Jobs["probe"] = job
		}},
		{"branch pushes can run the probe automatically", func(w *workflow) { w.On["push"] = map[string]any{"branches": []any{"**"}} }},
		{"tag pushes can run the probe automatically", func(w *workflow) { w.On["push"] = map[string]any{"tags": []any{"v*"}} }},
		{"job condition can skip the probe", func(w *workflow) {
			job := w.Jobs["probe"]
			job.If = "false"
			w.Jobs["probe"] = job
		}},
		{"step condition can skip the stall", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Steps[0].If = "false"
			w.Jobs["probe"] = job
		}},
		{"step timeout can expire before the job bound", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Steps[0].TimeoutMinutes = 5
			w.Jobs["probe"] = job
		}},
		{"job can ignore timeout failure", func(w *workflow) {
			job := w.Jobs["probe"]
			job.ContinueOnError = true
			w.Jobs["probe"] = job
		}},
		{"step can ignore timeout failure", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Steps[0].ContinueOnError = true
			w.Jobs["probe"] = job
		}},
		{"probe can depend on a missing job", func(w *workflow) {
			job := w.Jobs["probe"]
			job.Needs = []string{"missing"}
			w.Jobs["probe"] = job
		}},
	}
	for _, tc := range mutations {
		t.Run(tc.name, func(t *testing.T) {
			var mutated workflow
			if err := yaml.Unmarshal([]byte(raw), &mutated); err != nil {
				t.Fatal(err)
			}
			tc.edit(&mutated)
			if err := validateTimeoutCancellationProbe(mutated, productionBound); err == nil {
				t.Fatalf("policy accepted %s", tc.name)
			}
		})
	}

	var upperBound workflow
	if err := yaml.Unmarshal([]byte(raw), &upperBound); err != nil {
		t.Fatal(err)
	}
	job := upperBound.Jobs["probe"]
	job.TimeoutMinutes = 30
	upperBound.Jobs["probe"] = job
	if err := validateTimeoutCancellationProbe(upperBound, 30); err == nil {
		t.Fatal("policy accepted a 25-minute stall for a 30-minute production timeout")
	}
}

func TestPinnedStepsRejectEnvAndShellOverrides(t *testing.T) {
	// Regression test: step-level env/shell alter execution without tripping
	// the body pin. A rewritten step could keep the pinned body while smuggling
	// behavior through these keys, so the validators must pin them per step.
	t.Run("project step env override rejected", func(t *testing.T) {
		_, raw := readWorkflow(t, "go-checks.yml")
		mutated := strings.Replace(raw,
			"      - name: Test with race detector\n        run: go test -race ./...",
			"      - name: Test with race detector\n        env:\n          GOTOOLCHAIN: local\n        run: go test -race ./...",
			1)
		var project workflow
		if err := yaml.Unmarshal([]byte(mutated), &project); err != nil {
			t.Fatal(err)
		}
		security, _ := readWorkflow(t, "security-scans.yml")
		err := validateScannerIsolation(security, project)
		if err == nil {
			t.Fatal("step-level env override on a pinned project step was accepted")
		}
		if !strings.Contains(err.Error(), "Test with race detector") {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("project step shell override rejected", func(t *testing.T) {
		_, raw := readWorkflow(t, "go-checks.yml")
		mutated := strings.Replace(raw,
			"      - name: Test with race detector\n        run: go test -race ./...",
			"      - name: Test with race detector\n        shell: bash\n        run: go test -race ./...",
			1)
		var project workflow
		if err := yaml.Unmarshal([]byte(mutated), &project); err != nil {
			t.Fatal(err)
		}
		security, _ := readWorkflow(t, "security-scans.yml")
		err := validateScannerIsolation(security, project)
		if err == nil {
			t.Fatal("step-level shell override on a pinned project step was accepted")
		}
		if !strings.Contains(err.Error(), "Test with race detector") {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("scanner step shell change rejected", func(t *testing.T) {
		_, raw := readWorkflow(t, "security-scans.yml")
		mutated := strings.Replace(raw,
			"      - name: Check test-inclusive dependency licenses\n        shell: bash\n",
			"      - name: Check test-inclusive dependency licenses\n        shell: sh\n",
			1)
		var w workflow
		if err := yaml.Unmarshal([]byte(mutated), &w); err != nil {
			t.Fatal(err)
		}
		err := validateScannerSteps(w)
		if err == nil {
			t.Fatal("shell change on a pinned scanner step was accepted")
		}
		if !strings.Contains(err.Error(), "Check test-inclusive dependency licenses") {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("scanner step env override rejected", func(t *testing.T) {
		_, raw := readWorkflow(t, "security-scans.yml")
		mutated := strings.Replace(raw,
			"      - name: Check vulnerabilities including tests\n        run: govulncheck -test ./...",
			"      - name: Check vulnerabilities including tests\n        env:\n          GOFLAGS: -mod=mod\n        run: govulncheck -test ./...",
			1)
		var w workflow
		if err := yaml.Unmarshal([]byte(mutated), &w); err != nil {
			t.Fatal(err)
		}
		err := validateScannerSteps(w)
		if err == nil {
			t.Fatal("step-level env override on a pinned scanner step was accepted")
		}
		if !strings.Contains(err.Error(), "Check vulnerabilities including tests") {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("probe step env override rejected", func(t *testing.T) {
		_, raw := readWorkflow(t, "timeout-cancellation-probe.yml")
		mutated := strings.Replace(raw,
			"      - name: Stall beyond the configured timeout\n        run: sleep 25m",
			"      - name: Stall beyond the configured timeout\n        env:\n          SLEEP_EXTRA: \"60\"\n        run: sleep 25m",
			1)
		var probe workflow
		if err := yaml.Unmarshal([]byte(mutated), &probe); err != nil {
			t.Fatal(err)
		}
		goChecks, _ := readWorkflow(t, "go-checks.yml")
		err := validateTimeoutCancellationProbe(probe, goChecks.Jobs["checks"].TimeoutMinutes)
		if err == nil {
			t.Fatal("step-level env override on the probe step was accepted")
		}
		if !strings.Contains(err.Error(), "Stall beyond the configured timeout") {
			t.Fatalf("rejection did not name the step: %v", err)
		}
	})

	t.Run("env and shell are parsed, not dropped", func(t *testing.T) {
		got, _ := readWorkflow(t, "security-scans.yml")
		license := findStep(t, got, "Check test-inclusive dependency licenses")
		if license.Shell != "bash" {
			t.Fatalf("license step shell = %q, want %q (shell must be parsed, not dropped)", license.Shell, "bash")
		}
		if len(license.Env) != 0 {
			t.Fatalf("license step env = %v, want empty", license.Env)
		}
		verapdf, _ := readWorkflow(t, "verapdf-integration.yml")
		image := findStep(t, verapdf, "Pull the verified official Temurin base and build a local validator image")
		if !strings.HasPrefix(image.Env["BASE_IMAGE"], "eclipse-temurin:21.0.12.1_1-jre-noble@sha256:") {
			t.Fatalf("validator image step BASE_IMAGE = %q, want the pinned digest reference (env must be parsed, not dropped)", image.Env["BASE_IMAGE"])
		}
	})
}

func validateTimeoutCancellationProbe(probe workflow, productionBound int) error {
	if productionBound < 1 || len(probe.Jobs) != 1 {
		return fmt.Errorf("probe must define exactly one job and production must have a timeout")
	}
	job, ok := probe.Jobs["probe"]
	if !ok || job.TimeoutMinutes != productionBound {
		return fmt.Errorf("probe timeout must match Go checks timeout of %d minutes", productionBound)
	}
	if strings.TrimSpace(job.If) != "" || job.ContinueOnError || job.Needs != nil || len(job.Steps) != 1 {
		return fmt.Errorf("probe must sleep longer than the configured timeout")
	}
	if strings.TrimSpace(job.Steps[0].If) != "" || job.Steps[0].ContinueOnError {
		return fmt.Errorf("probe stall must not be skipped or ignored")
	}
	if job.Steps[0].Shell != "" || len(job.Steps[0].Env) != 0 {
		return fmt.Errorf("probe step %q must not set shell or step-level env", job.Steps[0].Name)
	}
	if job.Steps[0].TimeoutMinutes != 0 && job.Steps[0].TimeoutMinutes < productionBound {
		return fmt.Errorf("probe step timeout must not expire before the configured job timeout")
	}
	command := strings.TrimSpace(job.Steps[0].Run)
	sleepMatch := probeSleepPattern.FindStringSubmatch(command)
	if len(sleepMatch) != 2 {
		return fmt.Errorf("probe must sleep longer than the configured timeout")
	}
	stallMinutes, err := strconv.Atoi(sleepMatch[1])
	if err != nil || stallMinutes <= productionBound {
		return fmt.Errorf("probe stall must exceed the configured timeout of %d minutes", productionBound)
	}
	if len(probe.On) != 1 {
		return fmt.Errorf("probe must have only a manual workflow_dispatch trigger")
	}
	if _, ok := probe.On["workflow_dispatch"]; !ok {
		return fmt.Errorf("probe must be manually dispatched")
	}
	return nil
}

func TestWorkflowPolicyRejectsMutableActionsAndWritePermissions(t *testing.T) {
	got, raw := readWorkflow(t, "security-scans.yml")
	if err := validateWorkflow(got); err != nil {
		t.Fatal(err)
	}
	if err := validateActionPins(got, raw); err != nil {
		t.Fatal(err)
	}

	mutable := got
	mutable.Jobs = cloneJobs(got.Jobs)
	mutateUse(&mutable, "Checkout repository", "actions/checkout@v7")
	if err := validateActionPins(mutable, raw); err == nil {
		t.Fatal("mutable action tag accepted")
	}

	writable := got
	writable.Permissions = map[string]string{"contents": "write"}
	if err := validateWorkflow(writable); err == nil {
		t.Fatal("write permission accepted")
	}

	withoutCancellation := got
	withoutCancellation.Concurrency.CancelInProgress = false
	if err := validateWorkflow(withoutCancellation); err == nil {
		t.Fatal("workflow without cancellation accepted")
	}

	wrongBranch := got
	wrongBranch.On = cloneEvents(got.On)
	wrongBranch.On["pull_request"].(map[string]any)["branches"] = []any{"release"}
	if err := validateMainAndWeeklyTriggers(wrongBranch.On); err == nil {
		t.Fatal("pull request trigger outside main accepted")
	}

	noSchedule := got
	noSchedule.On = cloneEvents(got.On)
	delete(noSchedule.On, "schedule")
	if err := validateMainAndWeeklyTriggers(noSchedule.On); err == nil {
		t.Fatal("missing weekly scan trigger accepted")
	}
}

func TestSecretWorkflowScansAllHistoryWithRedactedOutput(t *testing.T) {
	got, raw := readWorkflow(t, "security-scans.yml")
	if err := validateActionPins(got, raw); err != nil {
		t.Fatal(err)
	}
	if err := validateMainAndWeeklyTriggers(got.On); err != nil {
		t.Fatal(err)
	}
	checkout := findStep(t, got, "Checkout repository")
	if checkout.With["persist-credentials"] != false {
		t.Fatal("checkout must not persist credentials")
	}
	secretJob, ok := got.Jobs["secrets"]
	if !ok {
		t.Fatal("missing full-history secret-scan job")
	}
	if secretJob.Permissions != nil {
		t.Fatal("secret-scan job must use only workflow-level read permissions")
	}
	fullHistory := false
	for _, s := range secretJob.Steps {
		if s.Uses != "" && strings.HasPrefix(s.Uses, "actions/checkout@") && s.With["fetch-depth"] == 0 {
			fullHistory = true
		}
	}
	if !fullHistory {
		t.Fatal("Gitleaks checkout must fetch full history")
	}
	if err := validateScannerSteps(got); err != nil {
		t.Fatal(err)
	}
}

func TestExceptionConfigAndMetadataPolicy(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".security", "secret-exceptions.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var config exceptionFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse exception config: %v", err)
	}
	if err := validateExceptions(config, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if len(config.Exceptions) != 0 {
		t.Fatal("no secret exceptions are currently approved")
	}

	valid := exception{ID: "synthetic-key-example", Paths: []string{"internal/devsecops/testdata/gitleaks/example.txt"}, Rules: []string{"aws-access-token"}, Reason: "Synthetic scanner fixture only", Owner: "security-owner", BeadsID: "dap-jo9.7", Expires: "2026-11-12"}
	if err := validateExceptions(exceptionFile{Version: 1, Exceptions: []exception{valid}}, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("valid synthetic exception rejected: %v", err)
	}

	badCases := []struct {
		name   string
		mutate func(*exception)
	}{
		{"expired", func(e *exception) { e.Expires = "2026-10-11" }},
		{"missing owner", func(e *exception) { e.Owner = "" }},
		{"missing Beads issue", func(e *exception) { e.BeadsID = "" }},
		{"invalid Beads issue", func(e *exception) { e.BeadsID = "other-123" }},
		{"broad non-fixture path", func(e *exception) { e.Paths = []string{"**/*.go"} }},
		{"missing expiration", func(e *exception) { e.Expires = "" }},
	}
	for _, tc := range badCases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := valid
			candidate.Paths = append([]string(nil), valid.Paths...)
			candidate.Rules = append([]string(nil), valid.Rules...)
			tc.mutate(&candidate)
			if err := validateExceptions(exceptionFile{Version: 1, Exceptions: []exception{candidate}}, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)); err == nil {
				t.Fatal("invalid exception accepted")
			}
		})
	}
}

func TestExceptionConfigRejectsMalformedYAML(t *testing.T) {
	var config exceptionFile
	if err := yaml.Unmarshal([]byte("version: ["), &config); err == nil {
		t.Fatal("malformed YAML accepted")
	}
}

func TestLicenseAllowlistIsExplicitAndRejectsUnknown(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".security", "allowed-go-licenses.txt"))
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		license := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if license != "" {
			allowed[license] = true
		}
	}
	for _, license := range []string{"Apache-2.0", "MIT", "BSD-3-Clause"} {
		if !allowed[license] {
			t.Errorf("reviewed permissive license %q missing from allowlist", license)
		}
	}
	for _, unapproved := range []string{"", "Unknown", "GPL-3.0", "GPL-3.0-only", "LicenseRef-unreviewed"} {
		if allowed[unapproved] {
			t.Errorf("unapproved license %q was allowlisted", unapproved)
		}
	}
}

func TestGitleaksConfigHasNoSuppressionRules(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".security", "gitleaks.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateGitleaksConfig(data); err != nil {
		t.Fatal(err)
	}
	for _, mutated := range [][]byte{
		append(append([]byte(nil), data...), []byte("\n[allowlist]\npaths = ['**']\n")...),
		[]byte("[extend]\nuseDefault = false\n"),
	} {
		if err := validateGitleaksConfig(mutated); err == nil {
			t.Fatal("Gitleaks suppression or custom-rule config accepted")
		}
	}
}

func TestGitleaksSmokeCoversRepositorySuppressionBypasses(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "scripts", "test-gitleaks.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{".gitleaks.toml", ".gitleaksignore", "config_suppression_status", "ignore_suppression_status", "source_ignore_suppression_status", "check-gitleaks-suppressions.sh", "--gitleaks-ignore-path=/dev/null"} {
		if !strings.Contains(string(data), required) {
			t.Errorf("Gitleaks suppression smoke test missing %q", required)
		}
	}
}

func TestGitleaksGuardRejectsUnmanagedSuppressionFiles(t *testing.T) {
	script := filepath.Join("..", "..", ".github", "scripts", "check-gitleaks-suppressions.sh")
	for _, name := range []string{".gitleaksignore", ".gitleaks.toml"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			candidate := filepath.Join(root, name)
			if err := os.WriteFile(candidate, []byte("synthetic suppression fixture\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", script, root)
			if output, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(output), "prohibited") {
				t.Fatalf("guard must reject %s; err=%v output=%s", name, err, output)
			}

			if err := os.Remove(candidate); err != nil {
				t.Fatal(err)
			}
			cmd = exec.Command("bash", script, root)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("guard must accept clean fixture; err=%v output=%s", err, output)
			}
		})
	}
}

func TestGitleaksGuardFailsClosedWhenDiscoveryFails(t *testing.T) {
	script := filepath.Join("..", "..", ".github", "scripts", "check-gitleaks-suppressions.sh")
	root := t.TempDir()
	shimDir := t.TempDir()
	tempDir := t.TempDir()
	findShim := filepath.Join(shimDir, "find")
	findMarker := filepath.Join(shimDir, "find-invoked")
	if err := os.WriteFile(findShim, []byte("#!/bin/sh\nprintf invoked > \"$FIND_SHIM_MARKER\"\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", script, root)
	cmd.Env = envWithOverrides(map[string]string{
		"FIND_SHIM_MARKER": findMarker,
		"PATH":             shimDir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"TMPDIR":           tempDir,
	})
	if output, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(output), "Failed to enumerate") {
		t.Fatalf("guard must report failed suppression discovery; err=%v output=%s", err, output)
	}
	if _, err := os.Stat(findMarker); err != nil {
		t.Fatalf("synthetic find shim was not invoked: %v", err)
	}
	if entries, err := os.ReadDir(tempDir); err != nil {
		t.Fatal(err)
	} else if len(entries) != 0 {
		t.Fatalf("guard left temporary discovery files behind: %v", entries)
	}
}

func TestLicenseSmokeUsesForbiddenTestOnlyDependency(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "scripts", "test-go-licenses.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"main_test.go", "fixtures/GPL-3.0.txt", "Not allowed license 'GPL-3\\.0'", "--include_tests", "--allowed_licenses=", "replace example.test/forbidden", ".security/allowed-go-licenses.txt", "grep -Ev", "paste -sd, -"} {
		if !strings.Contains(string(data), required) {
			t.Errorf("license smoke test missing %q", required)
		}
	}
	if strings.Contains(string(data), "allowed_licenses='Apache-2.0,BSD-3-Clause,MIT'") {
		t.Error("license smoke test must consume the checked-in production allowlist, not duplicate it")
	}
	license, err := os.ReadFile(filepath.Join("..", "..", ".github", "scripts", "fixtures", "GPL-3.0.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(license), "GNU GENERAL PUBLIC LICENSE") || !strings.Contains(string(license), "Version 3, 29 June 2007") {
		t.Fatal("forbidden-license fixture must contain the full GPL-3.0 license text")
	}
}

func validateWorkflow(w workflow) error {
	if strings.TrimSpace(w.Name) == "" || len(w.Jobs) == 0 {
		return fmt.Errorf("workflow name and jobs are required")
	}
	if len(w.Permissions) != 1 || w.Permissions["contents"] != "read" {
		return fmt.Errorf("workflow permissions must be exactly contents: read")
	}
	if strings.TrimSpace(w.Concurrency.Group) == "" || !w.Concurrency.CancelInProgress {
		return fmt.Errorf("workflow must define a concurrency group and cancel superseded runs")
	}
	if err := validateJobTimeouts(w); err != nil {
		return err
	}
	for name, configured := range w.Jobs {
		if strings.TrimSpace(configured.If) != "" {
			return fmt.Errorf("job %s must not be conditional", name)
		}
		if configured.ContinueOnError {
			return fmt.Errorf("job %s must not continue after failure", name)
		}
		if len(configured.Permissions) != 0 {
			return fmt.Errorf("job %s must not expand workflow permissions", name)
		}
	}
	return nil
}

func validateJobTimeouts(w workflow) error {
	for name, configured := range w.Jobs {
		if configured.TimeoutMinutes < 5 || configured.TimeoutMinutes > 30 {
			return fmt.Errorf("job %s timeout must be between 5 and 30 minutes", name)
		}
	}
	return nil
}

func validateActionPins(w workflow, raw string) error {
	annotation := regexp.MustCompile(`uses:\s+[^\s]+@[0-9a-f]{40}\s+#\s+v[0-9]`)
	for jobName, configured := range w.Jobs {
		for _, s := range configured.Steps {
			if s.Uses == "" {
				continue
			}
			parts := strings.SplitN(s.Uses, "@", 2)
			if len(parts) != 2 || !actionSHAPattern.MatchString(parts[1]) {
				return fmt.Errorf("job %s action %q is not pinned to a full commit SHA", jobName, s.Uses)
			}
			line := "uses: " + s.Uses
			found := false
			for _, candidate := range strings.Split(raw, "\n") {
				if strings.Contains(candidate, line) && annotation.MatchString(candidate) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("job %s action %q lacks a version comment", jobName, s.Uses)
			}
		}
	}
	return nil
}

func validateMainAndWeeklyTriggers(events map[string]any) error {
	if len(events) != 3 {
		return fmt.Errorf("expected push, pull_request, and schedule triggers")
	}
	for _, event := range []string{"push", "pull_request"} {
		config, ok := events[event].(map[string]any)
		if !ok {
			return fmt.Errorf("missing %s trigger", event)
		}
		branches, ok := config["branches"].([]any)
		if !ok || len(branches) != 1 || branches[0] != "main" {
			return fmt.Errorf("%s must target only main", event)
		}
	}
	schedules, ok := events["schedule"].([]any)
	if !ok || len(schedules) != 1 {
		return fmt.Errorf("expected one weekly schedule")
	}
	schedule, ok := schedules[0].(map[string]any)
	if !ok || schedule["cron"] != "17 5 * * 1" {
		return fmt.Errorf("weekly schedule must use the approved Monday 05:17 UTC cadence")
	}
	return nil
}

func validateScannerSteps(w workflow) error {
	var toolInstall, govuln, nestedCloudBootstrapGovuln, nestedGovuln, licenses, nestedLicenses, licenseFixture, actionlint, gitleaks, gitleaksInstall, gitleaksSmoke, gitleaksGuard bool
	for jobName, configured := range w.Jobs {
		for _, s := range configured.Steps {
			if expectedShell, pinned := pinnedScannerStepShell[s.Name]; pinned {
				if s.Shell != expectedShell || len(s.Env) != 0 {
					return fmt.Errorf("job %s scanner step %q must keep shell %q with no step-level env", jobName, s.Name, expectedShell)
				}
			}
			command := strings.TrimSpace(s.Run)
			if (jobName == "go-security" || jobName == "secrets") && strings.TrimSpace(s.If) != "" {
				return fmt.Errorf("job %s scanner step %q must not be conditional", jobName, s.Name)
			}
			if s.ContinueOnError || strings.Contains(command, "|| true") {
				return fmt.Errorf("job %s scanner step %q can hide failure", jobName, s.Name)
			}
			if strings.EqualFold(s.Name, "Check vulnerabilities including tests") {
				govuln = true
				if !strings.Contains(command, "govulncheck -test ./...") || strings.Contains(command, "-json") || strings.Contains(command, "-format sarif") || strings.Contains(command, "-format openvex") {
					return fmt.Errorf("govulncheck must scan test dependencies using exit-status text mode")
				}
			}
			if s.Name == "Check nested cloud bootstrap vulnerabilities including tests" {
				nestedCloudBootstrapGovuln = s.WorkingDirectory == "cmd/cloud-tools-bootstrap" && strings.TrimSpace(command) == "govulncheck -test ./..."
				if !nestedCloudBootstrapGovuln {
					return fmt.Errorf("nested cloud bootstrap module must receive a test-inclusive govulncheck scan")
				}
			}
			if s.Name == "Check nested adapter vulnerabilities including tests" {
				nestedGovuln = s.WorkingDirectory == "ci/pdfua2/adapter" && strings.TrimSpace(command) == "govulncheck -test ./..."
				if !nestedGovuln {
					return fmt.Errorf("nested adapter module must receive a test-inclusive govulncheck scan")
				}
			}
			if s.Name == "Install pinned security tools" {
				required := []string{
					"golang.org/x/vuln/cmd/govulncheck@v1.9.0",
					"github.com/google/go-licenses/v2@v2.0.1",
					"github.com/rhysd/actionlint/cmd/actionlint@v1.7.12",
				}
				toolInstall = true
				for _, pin := range required {
					if !strings.Contains(command, pin) {
						toolInstall = false
						return fmt.Errorf("security tool install must pin %s", pin)
					}
				}
			}
			if s.Name == "Check test-inclusive dependency licenses" {
				licenses = true
				required := []string{"go-licenses check --include_tests", ".security/allowed-go-licenses.txt", "--allowed_licenses", "test -n \"$allowed_licenses\""}
				for _, fragment := range required {
					if !strings.Contains(command, fragment) {
						return fmt.Errorf("license check must include test dependencies and use the explicit fail-closed allowlist")
					}
				}
				if strings.Contains(command, "|| true") || strings.Contains(command, "--allowed_licenses=*") {
					return fmt.Errorf("license check must not hide findings or use a wildcard allowlist")
				}
			}
			if s.Name == "Check nested adapter test-inclusive dependency licenses" {
				nestedLicenses = s.WorkingDirectory == "ci/pdfua2/adapter" &&
					strings.Contains(command, "go-licenses check --include_tests") &&
					strings.Contains(command, "../../../.security/allowed-go-licenses.txt") &&
					strings.Contains(command, "--allowed_licenses=\"$allowed_licenses\"") &&
					strings.Contains(command, "test -n \"$allowed_licenses\"")
				if !nestedLicenses {
					return fmt.Errorf("nested adapter module must receive test-inclusive license checks with the explicit fail-closed allowlist")
				}
			}
			if s.Name == "Verify license policy rejects a test-only forbidden dependency" {
				licenseFixture = command == "bash .github/scripts/test-go-licenses.sh"
				if !licenseFixture {
					return fmt.Errorf("license policy must test rejection of a forbidden test-only dependency")
				}
			}
			if s.Name == "Lint workflow files" {
				actionlint = command == "actionlint"
				if !actionlint {
					return fmt.Errorf("workflow linting must run actionlint directly and fail on error")
				}
			}
			if s.Name == "Install pinned Gitleaks" {
				gitleaksInstall = strings.Contains(command, "gitleaks_8.30.1_linux_x64.tar.gz") &&
					strings.Contains(command, "551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb") &&
					strings.Contains(command, "sha256sum --check") &&
					strings.Contains(command, "set -euo pipefail")
				if !gitleaksInstall {
					return fmt.Errorf("Gitleaks must use the pinned v8.30.1 Linux release with SHA-256 verification")
				}
			}
			if s.Name == "Verify Gitleaks clean, finding, and error behavior" {
				gitleaksSmoke = command == "bash .github/scripts/test-gitleaks.sh"
				if !gitleaksSmoke {
					return fmt.Errorf("Gitleaks must run the synthetic suppression and redaction smoke checks")
				}
			}
			if s.Name == "Reject unmanaged Gitleaks config and ignore files" {
				gitleaksGuard = command == "bash .github/scripts/check-gitleaks-suppressions.sh"
				if !gitleaksGuard {
					return fmt.Errorf("Gitleaks must reject unmanaged source-level config and suppression files")
				}
			}
			if strings.Contains(strings.ToLower(s.Name), "gitleaks full-history scan") {
				gitleaks = true
				if !strings.Contains(command, "gitleaks git") || !strings.Contains(command, "--log-opts=--all") || !strings.Contains(command, "--redact=100") || !strings.Contains(command, "--ignore-gitleaks-allow") || !strings.Contains(command, "--config=.security/gitleaks.toml") || !strings.Contains(command, "--gitleaks-ignore-path=/dev/null") || strings.Contains(command, "--report-path") || strings.Contains(command, "--report-format") || strings.Contains(command, "--exit-code=0") {
					return fmt.Errorf("Gitleaks must scan all history with full redaction and native finding/error exit status")
				}
			}
		}
	}
	if !toolInstall || !govuln || !nestedCloudBootstrapGovuln || !nestedGovuln || !licenses || !nestedLicenses || !licenseFixture || !actionlint || !gitleaks || !gitleaksInstall || !gitleaksSmoke || !gitleaksGuard {
		return fmt.Errorf("pinned Go security tools, root and both nested-module govulncheck, root and nested test-inclusive license checks and negative fixture, actionlint, pinned Gitleaks install, suppression guard and smoke, and full-history scan steps are required")
	}
	return nil
}

// requiredGoChecksModuleGates enumerates every root and nested-module gate the
// go-checks "checks" job must keep. validateScannerIsolation rejects any
// workflow missing one of them, naming the missing gate. dap-jo9.5.6:
// package-level so negative policy tests can prove removal rejection for every
// gate. Keep entries in lockstep with .github/workflows/go-checks.yml.
var requiredGoChecksModuleGates = []struct {
	name             string
	workingDirectory string
	command          string
}{
	{name: "Check module files", command: "go mod tidy -diff"},
	{name: "Test with race detector", command: "go test -race ./..."},
	{name: "Vet", command: "go vet ./..."},
	{name: "Test standalone cloud bootstrap module", workingDirectory: "cmd/cloud-tools-bootstrap", command: "go test -race ./..."},
	{name: "Vet standalone cloud bootstrap module", workingDirectory: "cmd/cloud-tools-bootstrap", command: "go vet ./..."},
	{name: "Check cloud bootstrap module files", workingDirectory: "cmd/cloud-tools-bootstrap", command: "go mod tidy -diff"},
	{name: "Test PDF/UA-2 adapter module", workingDirectory: "ci/pdfua2/adapter", command: "go test -race ./... -count=1"},
	{name: "Vet PDF/UA-2 adapter module", workingDirectory: "ci/pdfua2/adapter", command: "go vet ./..."},
	{name: "Check adapter module files", workingDirectory: "ci/pdfua2/adapter", command: "go mod tidy -diff"},
}

func validateScannerIsolation(security, project workflow) error {
	scannerJob, ok := security.Jobs["go-security"]
	if !ok {
		return fmt.Errorf("required go-security scanner context is missing")
	}
	projectJob, ok := project.Jobs["checks"]
	if !ok {
		return fmt.Errorf("project tests must remain in the Go checks job")
	}
	if strings.TrimSpace(projectJob.If) != "" || projectJob.ContinueOnError {
		return fmt.Errorf("project job %q must be unconditional and fail closed", "checks")
	}
	if len(projectJob.Env) != 0 {
		return fmt.Errorf("project job %q must not set job-level env", "checks")
	}
	if projectJob.Defaults != nil {
		return fmt.Errorf("project job %q must not set defaults", "checks")
	}
	if projectJob.Container != nil {
		return fmt.Errorf("project job %q must not run in a container", "checks")
	}
	for _, candidate := range projectJob.Steps {
		if candidate.Container != nil {
			return fmt.Errorf("project step %q must not run in a container", candidate.Name)
		}
	}
	for _, candidate := range scannerJob.Steps {
		fields := strings.Fields(candidate.Run)
		for i := 0; i+1 < len(fields); i++ {
			if fields[i] == "go" && (fields[i+1] == "test" || fields[i+1] == "vet" || fields[i+1] == "run") {
				return fmt.Errorf("go-security step %q executes project Go code; keep tests in the separate Go checks workflow", candidate.Name)
			}
		}
	}
	var setup step
	setupFound := false
	for _, candidate := range scannerJob.Steps {
		if candidate.Name == "Set up Go" {
			setup = candidate
			setupFound = true
			break
		}
	}
	if !setupFound || setup.With["cache"] != false {
		return fmt.Errorf("go-security must not restore or save a Go build cache shared with PR-controlled tests")
	}
	var install step
	installFound := false
	for _, candidate := range scannerJob.Steps {
		if candidate.Name == "Install pinned security tools" {
			install = candidate
			installFound = true
			break
		}
	}
	if !installFound ||
		!strings.Contains(install.Run, "tool_dir=\"$RUNNER_TEMP/security-tools\"") ||
		!strings.Contains(install.Run, "GOBIN=\"$tool_dir\" go install golang.org/x/vuln/cmd/govulncheck@v1.9.0") ||
		!strings.Contains(install.Run, "GOBIN=\"$tool_dir\" go install github.com/google/go-licenses/v2@v2.0.1") ||
		!strings.Contains(install.Run, "GOBIN=\"$tool_dir\" go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12") ||
		!strings.Contains(install.Run, "echo \"$tool_dir\" >> \"$GITHUB_PATH\"") {
		return fmt.Errorf("pinned scanners must install into a private RUNNER_TEMP directory and add only that directory to PATH")
	}
	for _, required := range requiredGoChecksModuleGates {
		found := false
		for _, candidate := range projectJob.Steps {
			if candidate.Name == required.name &&
				candidate.WorkingDirectory == required.workingDirectory &&
				strings.TrimSpace(candidate.Run) == required.command {
				if strings.TrimSpace(candidate.If) != "" || candidate.ContinueOnError {
					return fmt.Errorf("project step %q must be unconditional and fail closed", required.name)
				}
				if candidate.Shell != "" || len(candidate.Env) != 0 {
					return fmt.Errorf("project step %q must not set shell or step-level env", required.name)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("separate Go checks job must keep %q running %q", required.name, required.command)
		}
	}
	return nil
}

func validateGitleaksConfig(data []byte) error {
	const approved = "title = \"Repository secret scanning defaults\"\n\n[extend]\nuseDefault = true\n"
	if string(data) != approved {
		return fmt.Errorf("Gitleaks config must extend defaults without adding suppressions or custom rules")
	}
	return nil
}

func validateExceptions(config exceptionFile, now time.Time) error {
	if config.Version != 1 {
		return fmt.Errorf("exception config version must be 1")
	}
	seen := map[string]bool{}
	for _, e := range config.Exceptions {
		if strings.TrimSpace(e.ID) == "" || seen[e.ID] {
			return fmt.Errorf("exception IDs must be non-empty and unique")
		}
		seen[e.ID] = true
		if len(e.Paths) == 0 || len(e.Rules) == 0 || strings.TrimSpace(e.Reason) == "" {
			return fmt.Errorf("exception %s requires path, rule, and reason", e.ID)
		}
		for _, p := range e.Paths {
			clean := filepath.ToSlash(filepath.Clean(p))
			if clean != p || strings.ContainsAny(p, "*?[]") || !strings.HasPrefix(p, "internal/devsecops/testdata/gitleaks/") {
				return fmt.Errorf("exception %s path must name one synthetic scanner fixture", e.ID)
			}
		}
		for _, rule := range e.Rules {
			if strings.TrimSpace(rule) == "" {
				return fmt.Errorf("exception %s has an empty rule id", e.ID)
			}
		}
		if !ownerPattern.MatchString(e.Owner) {
			return fmt.Errorf("exception %s requires an owner", e.ID)
		}
		if !beadsIDPattern.MatchString(e.BeadsID) {
			return fmt.Errorf("exception %s requires a Beads issue id", e.ID)
		}
		expires, err := time.Parse("2006-01-02", e.Expires)
		if err != nil || !expires.After(now) {
			return fmt.Errorf("exception %s requires a future YYYY-MM-DD expiry", e.ID)
		}
	}
	return nil
}

func findStep(t *testing.T, w workflow, name string) step {
	t.Helper()
	for _, configured := range w.Jobs {
		for _, s := range configured.Steps {
			if s.Name == name {
				return s
			}
		}
	}
	t.Fatalf("step %q not found", name)
	return step{}
}

func mutateRun(w *workflow, name string, mutate func(*step)) {
	for jobName, configured := range w.Jobs {
		for i := range configured.Steps {
			if configured.Steps[i].Name == name {
				mutate(&configured.Steps[i])
				w.Jobs[jobName] = configured
				return
			}
		}
	}
	panic("scanner step not found: " + name)
}

func mutateUse(w *workflow, name, value string) {
	for jobName, configured := range w.Jobs {
		for i := range configured.Steps {
			if configured.Steps[i].Name == name {
				configured.Steps[i].Uses = value
				w.Jobs[jobName] = configured
				return
			}
		}
	}
	panic("step not found: " + name)
}

func cloneJobs(source map[string]job) map[string]job {
	clone := make(map[string]job, len(source))
	for name, configured := range source {
		configured.Steps = append([]step(nil), configured.Steps...)
		for i := range configured.Steps {
			if configured.Steps[i].With != nil {
				with := make(map[string]any, len(configured.Steps[i].With))
				for key, value := range configured.Steps[i].With {
					with[key] = value
				}
				configured.Steps[i].With = with
			}
			if configured.Steps[i].Env != nil {
				env := make(map[string]string, len(configured.Steps[i].Env))
				for key, value := range configured.Steps[i].Env {
					env[key] = value
				}
				configured.Steps[i].Env = env
			}
		}
		clone[name] = configured
	}
	return clone
}

func cloneEvents(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source))
	for name, value := range source {
		switch event := value.(type) {
		case map[string]any:
			copy := make(map[string]any, len(event))
			for key, setting := range event {
				if values, ok := setting.([]any); ok {
					copy[key] = append([]any(nil), values...)
				} else {
					copy[key] = setting
				}
			}
			clone[name] = copy
		case []any:
			copy := make([]any, len(event))
			for i, item := range event {
				if config, ok := item.(map[string]any); ok {
					configCopy := make(map[string]any, len(config))
					for key, setting := range config {
						configCopy[key] = setting
					}
					copy[i] = configCopy
				} else {
					copy[i] = item
				}
			}
			clone[name] = copy
		default:
			clone[name] = value
		}
	}
	return clone
}

func envWithOverrides(overrides map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			env = append(env, entry)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

func TestValidatorEvidenceRetentionFloor(t *testing.T) {
	w, _ := readWorkflow(t, "verapdf-integration.yml")
	const floor = 90
	found := false
	for jobName, j := range w.Jobs {
		for _, s := range j.Steps {
			name, _ := s.With["name"].(string)
			if name != "pdfua2-reports" || !strings.HasPrefix(s.Uses, "actions/upload-artifact@") {
				continue
			}
			found = true
			raw, ok := s.With["retention-days"]
			if !ok {
				t.Errorf("job %s: pdfua2-reports artifact step is missing retention-days", jobName)
				continue
			}
			var days int
			switch v := raw.(type) {
			case int:
				days = v
			case string:
				n, err := strconv.Atoi(strings.TrimSpace(v))
				if err != nil {
					t.Errorf("job %s: pdfua2-reports retention-days %q is not a number", jobName, v)
					continue
				}
				days = n
			default:
				t.Errorf("job %s: pdfua2-reports retention-days has unexpected type %T", jobName, raw)
				continue
			}
			if days < floor {
				t.Errorf("job %s: pdfua2-reports retention-days = %d, want >= %d (evidence must outlast the release-gate review window)", jobName, days, floor)
			}
		}
	}
	if !found {
		t.Error("no pdfua2-reports artifact upload step found in verapdf-integration.yml")
	}
}
