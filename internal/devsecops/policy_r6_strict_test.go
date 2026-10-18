package devsecops

import (
	"strings"
	"testing"
)

// R6-01 (DAP-R6-01) strict-parsing regression tests.
//
// The pre-fix security-scans parse path lowercased every structural key before
// decoding, so case-variant spellings of execution-altering keys (DEFAULTS:,
// ENV:) were normalized into their modeled lowercase forms and accepted, and
// unmodeled keys were silently dropped. The validator then saw a different
// workflow than the one GitHub would execute. The security-scans path now
// decodes strictly with known fields, so any misspelled or unmodeled key
// fails closed at parse time.
//
// NOTE (TDD seam): these tests first exercised parseWorkflow(data, false),
// the pre-fix security-scans parse entry, and watched each bypass get
// accepted (RED). They now target parseSecurityScansWorkflow, the strict
// production entry point wired into readWorkflow for security-scans.yml.

// TestSecurityScansStrictParseRejectsCaseVariantExecutionKeys: DEFAULTS:/ENV:
// (workflow level) and Defaults:/Env: (job level) must not be normalized into
// their modeled lowercase forms.
func TestSecurityScansStrictParseRejectsCaseVariantExecutionKeys(t *testing.T) {
	raw := securityScansRaw(t)
	cases := []struct {
		name   string
		anchor string
		insert string
		key    string
	}{
		{
			name:   "workflow-level DEFAULTS",
			anchor: "cancel-in-progress: true\n",
			insert: "\nDEFAULTS:\n  run:\n    shell: \"true {0}\"\n",
			key:    "DEFAULTS",
		},
		{
			name:   "workflow-level ENV",
			anchor: "permissions:\n  contents: read\n",
			insert: "\nENV:\n  BASH_ENV: /tmp/evil-bash-env\n",
			key:    "ENV",
		},
		{
			name:   "job-level Defaults",
			anchor: "  go-security:\n    runs-on: ubuntu-latest\n    timeout-minutes: 25\n",
			insert: "    Defaults:\n      run:\n        shell: \"true {0}\"\n",
			key:    "Defaults",
		},
		{
			name:   "job-level Env",
			anchor: "  go-security:\n    runs-on: ubuntu-latest\n    timeout-minutes: 25\n",
			insert: "    Env:\n      BASH_ENV: /tmp/evil-bash-env\n",
			key:    "Env",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(raw, tc.anchor, tc.anchor+tc.insert, 1)
			if mutated == raw {
				t.Fatal("mutation anchor not found")
			}
			if _, err := parseSecurityScansWorkflow([]byte(mutated)); err == nil {
				t.Fatalf("case-variant key %q was accepted by the security-scans parse path", tc.key)
			} else if !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("rejection = %v, want error naming %q", err, tc.key)
			}
		})
	}
}

// TestSecurityScansStrictParseRejectsUnknownFields: unmodeled keys at
// workflow, job, and step level must fail closed instead of being silently
// dropped (the validator would otherwise reason about a different workflow
// than the one that executes).
func TestSecurityScansStrictParseRejectsUnknownFields(t *testing.T) {
	raw := securityScansRaw(t)
	cases := []struct {
		name   string
		anchor string
		insert string
		field  string
	}{
		{
			name:   "workflow field",
			anchor: "name: Security scans\n",
			insert: "unmodeled_security_field: true\n",
			field:  "unmodeled_security_field",
		},
		{
			name:   "job field",
			anchor: "  go-security:\n    runs-on: ubuntu-latest\n",
			insert: "    unmodeled_job_field: true\n",
			field:  "unmodeled_job_field",
		},
		{
			name:   "step field",
			anchor: "      - name: Lint workflow files\n",
			insert: "        unmodeled_step_field: true\n",
			field:  "unmodeled_step_field",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(raw, tc.anchor, tc.anchor+tc.insert, 1)
			if mutated == raw {
				t.Fatal("mutation anchor not found")
			}
			if _, err := parseSecurityScansWorkflow([]byte(mutated)); err == nil {
				t.Fatalf("unknown field %q was silently dropped by the security-scans parse path", tc.field)
			} else if !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("rejection = %v, want error naming %q", err, tc.field)
			}
		})
	}
}
