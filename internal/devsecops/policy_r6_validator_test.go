package devsecops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// R6-01 (DAP-R6-01) validator regression tests.
//
// A workflow- or job-level `defaults.run.shell: "true {0}"` or
// `env: {BASH_ENV: ...}` turns every scan step into a no-op that still
// "passes", and the pre-fix security-scans validator accepted all of them:
// validateWorkflow never inspected the workflow/job Defaults or Env fields.
// Each test below constructs the bypass at YAML level (the way an attacker
// would submit it), parses it the way the validator consumes it, and requires
// a fail-closed rejection.

func securityScansRaw(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "security-scans.yml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mustParseSecurityWorkflowLoose(t *testing.T, data string) workflow {
	t.Helper()
	var w workflow
	if err := yaml.Unmarshal([]byte(data), &w); err != nil {
		t.Fatalf("parse mutated security-scans workflow: %v", err)
	}
	return w
}

// TestSecurityScansRejectsWorkflowLevelDefaultsShell: a workflow-level
// defaults.run.shell replaces the shell of every run step that does not set
// its own shell, turning every scan into a no-op that still "passes".
func TestSecurityScansRejectsWorkflowLevelDefaultsShell(t *testing.T) {
	raw := securityScansRaw(t)
	anchor := "cancel-in-progress: true\n"
	mutated := strings.Replace(raw, anchor, anchor+"\ndefaults:\n  run:\n    shell: \"true {0}\"\n", 1)
	if mutated == raw {
		t.Fatal("defaults anchor not found")
	}
	w := mustParseSecurityWorkflowLoose(t, mutated)
	if w.Defaults == nil {
		t.Fatal("mutated workflow-level defaults was not parsed; test is vacuous")
	}
	if err := validateWorkflow(w); err == nil {
		t.Fatal("workflow-level defaults.run.shell bypass was accepted")
	} else if !strings.Contains(err.Error(), "defaults") {
		t.Fatalf("rejection = %v, want error naming defaults", err)
	}
}

// TestSecurityScansRejectsJobLevelDefaultsShell: the same shell-replacement
// bypass scoped to the go-security job.
func TestSecurityScansRejectsJobLevelDefaultsShell(t *testing.T) {
	raw := securityScansRaw(t)
	anchor := "  go-security:\n    runs-on: ubuntu-latest\n    timeout-minutes: 25\n"
	mutated := strings.Replace(raw, anchor, anchor+"    defaults:\n      run:\n        shell: \"true {0}\"\n", 1)
	if mutated == raw {
		t.Fatal("job defaults anchor not found")
	}
	w := mustParseSecurityWorkflowLoose(t, mutated)
	if w.Jobs["go-security"].Defaults == nil {
		t.Fatal("mutated job-level defaults was not parsed; test is vacuous")
	}
	if err := validateWorkflow(w); err == nil {
		t.Fatal("job-level defaults.run.shell bypass was accepted")
	} else if !strings.Contains(err.Error(), "go-security") || !strings.Contains(err.Error(), "defaults") {
		t.Fatalf("rejection = %v, want error naming the go-security job defaults", err)
	}
}

// TestSecurityScansRejectsWorkflowLevelEnvBashEnv: non-interactive bash
// sources BASH_ENV, so a workflow-level env: {BASH_ENV: ...} alters every scan
// step without tripping any body pin.
func TestSecurityScansRejectsWorkflowLevelEnvBashEnv(t *testing.T) {
	raw := securityScansRaw(t)
	anchor := "permissions:\n  contents: read\n"
	mutated := strings.Replace(raw, anchor, anchor+"\nenv:\n  BASH_ENV: /tmp/security-scans-bash-env\n", 1)
	if mutated == raw {
		t.Fatal("env anchor not found")
	}
	w := mustParseSecurityWorkflowLoose(t, mutated)
	if len(w.Env) == 0 {
		t.Fatal("mutated workflow-level env was not parsed; test is vacuous")
	}
	if err := validateWorkflow(w); err == nil {
		t.Fatal("workflow-level env BASH_ENV bypass was accepted")
	} else if !strings.Contains(err.Error(), "env") {
		t.Fatalf("rejection = %v, want error naming env", err)
	}
}

// TestSecurityScansRejectsJobLevelEnvBashEnv: the same BASH_ENV bypass scoped
// to the go-security job.
func TestSecurityScansRejectsJobLevelEnvBashEnv(t *testing.T) {
	raw := securityScansRaw(t)
	anchor := "  go-security:\n    runs-on: ubuntu-latest\n    timeout-minutes: 25\n"
	mutated := strings.Replace(raw, anchor, anchor+"    env:\n      BASH_ENV: /tmp/security-scans-bash-env\n", 1)
	if mutated == raw {
		t.Fatal("job env anchor not found")
	}
	w := mustParseSecurityWorkflowLoose(t, mutated)
	if len(w.Jobs["go-security"].Env) == 0 {
		t.Fatal("mutated job-level env was not parsed; test is vacuous")
	}
	if err := validateWorkflow(w); err == nil {
		t.Fatal("job-level env BASH_ENV bypass was accepted")
	} else if !strings.Contains(err.Error(), "go-security") || !strings.Contains(err.Error(), "env") {
		t.Fatalf("rejection = %v, want error naming the go-security job env", err)
	}
}

// TestSecurityScansRejectsMultiDocumentYAML: a single Decode/Unmarshal reads
// only the first YAML document and silently drops the rest, so trailing
// attacker-controlled documents would ride along unvalidated. The strict
// decoder must fail closed on any input after the first document.
func TestSecurityScansRejectsMultiDocumentYAML(t *testing.T) {
	raw := securityScansRaw(t)
	mutated := raw + "\n---\n# trailing document dropped by loose decoders\njobs: {}\n"
	if _, err := parseWorkflow([]byte(mutated), true); err == nil {
		t.Fatal("multi-document security-scans YAML was accepted")
	}
}
