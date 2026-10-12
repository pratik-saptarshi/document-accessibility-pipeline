package devsecops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
}

type step struct {
	Name             string         `yaml:"name"`
	Uses             string         `yaml:"uses"`
	Run              string         `yaml:"run"`
	WorkingDirectory string         `yaml:"working-directory"`
	With             map[string]any `yaml:"with"`
	ContinueOnError  bool           `yaml:"continue-on-error"`
	If               string         `yaml:"if"`
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
	actionSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	beadsIDPattern   = regexp.MustCompile(`^dap-[a-z0-9]+(?:\.[a-z0-9]+)*$`)
	ownerPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{1,38}$`)
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
	var toolInstall, govuln, nestedGovuln, licenses, nestedLicenses, licenseFixture, actionlint, gitleaks, gitleaksInstall, gitleaksSmoke, gitleaksGuard bool
	for jobName, configured := range w.Jobs {
		for _, s := range configured.Steps {
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
	if !toolInstall || !govuln || !nestedGovuln || !licenses || !nestedLicenses || !licenseFixture || !actionlint || !gitleaks || !gitleaksInstall || !gitleaksSmoke || !gitleaksGuard {
		return fmt.Errorf("pinned Go security tools, root and nested-module govulncheck, root and nested test-inclusive license checks and negative fixture, actionlint, pinned Gitleaks install, suppression guard and smoke, and full-history scan steps are required")
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
