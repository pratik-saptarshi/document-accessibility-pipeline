package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
