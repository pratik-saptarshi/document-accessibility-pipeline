package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func approvedAFPSourceFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "approved-afp-refinement-source.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if got := digest(b); got != approvedAFPRefinementSourceSHA256 {
		t.Fatalf("AFP source fixture digest = %s, want %s", got, approvedAFPRefinementSourceSHA256)
	}
	return b
}

func TestCheckedAFPDependencyCapacityRejectsIntegerOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if got, err := checkedAFPDependencyCapacity(maxInt-1, 1); err != nil || got != maxInt {
		t.Fatalf("largest representable capacity = (%d, %v), want (%d, nil)", got, err, maxInt)
	}
	for _, tc := range []struct {
		name      string
		existing  int
		additions int
	}{
		{name: "positive overflow", existing: maxInt, additions: 1},
		{name: "negative existing", existing: -1, additions: 1},
		{name: "negative additions", existing: 1, additions: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := checkedAFPDependencyCapacity(tc.existing, tc.additions); err == nil {
				t.Fatalf("capacity = %d, want overflow/invalid-size error", got)
			}
		})
	}
}

func TestAFPNullLossIsUnavailableToGenericPlans(t *testing.T) {
	source := approvedAFPSourceFixture(t)
	imported := bytes.Replace(source, []byte(`,"dependencies":null`), nil, 1)
	if bytes.Equal(imported, source) {
		t.Fatal("fixture no longer contains the pinned explicit-null loss")
	}
	plan := Plan{BDVersion: approvedNullLossVersion}
	if _, _, err := ReconcileWithBinary(source, imported, imported, plan, approvedNullLossBDSHA256); err == nil {
		t.Fatal("generic plan used the AFP-only explicit-null exception")
	}
}

func TestValidateAFPRefinementPlanIsBoundToExactSource(t *testing.T) {
	source := approvedAFPSourceFixture(t)
	if err := ValidatePlan(source, Plan{AFPRefinement: true, BDVersion: approvedNullLossVersion}); err != nil {
		t.Fatalf("exact AFP source was rejected: %v", err)
	}

	mutated := append(append([]byte(nil), source...), ' ')
	if err := ValidatePlan(mutated, Plan{AFPRefinement: true, BDVersion: approvedNullLossVersion}); err == nil {
		t.Fatal("AFP batch accepted a different source digest")
	}
}

func TestAFPSourceDigestDocumentationMatchesExactFixture(t *testing.T) {
	const prefix = "dd048534"
	for _, name := range []string{"../../docs/planning/afp-refinement-2026-10-19.md", "../../docs/tools/beads-export-reconcile.md"} {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		claims := regexp.MustCompile(prefix+`[0-9a-f]*`).FindAllString(string(content), -1)
		if len(claims) == 0 {
			t.Fatalf("%s contains no AFP source digest claim", name)
		}
		for _, claim := range claims {
			if len(claim) != 64 || claim != approvedAFPRefinementSourceSHA256 {
				t.Errorf("%s contains malformed or mismatched AFP source SHA-256 (%d chars): %s", name, len(claim), claim)
			}
		}
		if strings.Contains(string(content), prefix+"e6e007841e0750d5370a25398178f2b0c70778eabe1f4b1508409") {
			t.Errorf("%s retains the truncated AFP source digest transcription", name)
		}
	}
}

func TestValidateAFPRefinementRejectsMixedGenericOperations(t *testing.T) {
	source := approvedAFPSourceFixture(t)
	tests := []Plan{
		{AFPRefinement: true, Updates: []IssueUpdate{{ID: "dap-jo9.4", Fields: map[string]string{"title": "unplanned"}}}},
		{AFPRefinement: true, Creates: []json.RawMessage{json.RawMessage(`{"id":"dap-jo9.99"}`)}},
		{AFPRefinement: true, Closes: []IssueClosure{{ID: "dap-jo9.4.1"}}},
	}
	for _, plan := range tests {
		plan.BDVersion = approvedNullLossVersion
		if err := ValidatePlan(source, plan); err == nil {
			t.Fatalf("ValidatePlan accepted mixed AFP/generic operations: %+v", plan)
		}
	}
}
