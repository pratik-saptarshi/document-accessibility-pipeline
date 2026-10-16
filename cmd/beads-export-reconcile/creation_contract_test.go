package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

type creationContractFixture struct {
	CurrentSourceSHA256       string          `json:"current_source_sha256"`
	PreviousNullLossSourceSHA string          `json:"previous_null_loss_source_sha256"`
	ProposedPlan              json.RawMessage `json:"proposed_plan"`
	SnapshotCounts            struct {
		Source    int `json:"source"`
		Baseline  int `json:"baseline"`
		Updated   int `json:"updated"`
		Candidate int `json:"candidate"`
		RoundTrip int `json:"round_trip"`
	} `json:"snapshot_counts"`
	LineageDesign struct {
		Status       string   `json:"status"`
		Alternatives []string `json:"alternatives"`
		Invariant    string   `json:"invariant"`
	} `json:"lineage_design"`
	SourceFreshness struct {
		OldBundleSourceSHA256   string `json:"old_bundle_source_sha256"`
		CurrentSourceSHA256     string `json:"current_source_sha256"`
		OldBundleAgainstCurrent string `json:"old_bundle_against_current_source"`
		FreshBundleAfterReview  string `json:"fresh_bundle_rebuilt_from_current_source_after_reviewed_lineage_policy"`
	} `json:"source_freshness_continuation"`
	Cases []struct {
		ID       string `json:"id"`
		Expected string `json:"expected"`
		Mutation string `json:"mutation"`
	} `json:"cases"`
}

func readCreationContractFixture(t *testing.T) creationContractFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "issue-creation-contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture creationContractFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode creation contract fixture: %v", err)
	}
	return fixture
}

func TestPlannedCreationPlanIsAcceptedByStrictSchema(t *testing.T) {
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	err := strictJSON(fixture.ProposedPlan, &plan)
	if err != nil {
		t.Fatalf("strictJSON() rejected the proposed explicit creation plan: %v", err)
	}
	if len(plan.Creates) != 1 || len(plan.Updates) != 0 {
		t.Fatalf("decoded plan has %d creates and %d updates, want one create only", len(plan.Creates), len(plan.Updates))
	}
}

func TestReconcileAcceptsExactSyntheticCreate(t *testing.T) {
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	if err := strictJSON(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	parent := []byte(`{"_type":"issue","id":"dap-jo9","title":"parent","status":"open","dependencies":[]}` + "\n")
	var createFields map[string]json.RawMessage
	if err := json.Unmarshal(plan.Creates[0], &createFields); err != nil {
		t.Fatal(err)
	}
	create, err := json.Marshal(createFields)
	if err != nil {
		t.Fatal(err)
	}
	updated := append(append([]byte(nil), parent[:len(parent)-1]...), '\n')
	updated = append(updated, create...)
	updated = append(updated, '\n')
	candidate, evidence, err := beadsreconcile.Reconcile(parent, parent, updated, plan)
	if err != nil {
		t.Fatalf("Reconcile() rejected the exact synthetic create: %v", err)
	}
	if len(evidence.IssueIDs) != 0 || len(evidence.CreatedIDs) != 1 || evidence.CreatedIDs[0] != "dap-jo9.99" || len(evidence.AllIssueIDs) != 2 {
		t.Fatalf("creation provenance = updates %v, creates %v, all %v", evidence.IssueIDs, evidence.CreatedIDs, evidence.AllIssueIDs)
	}
	want := append(append([]byte(nil), parent...), create...)
	want = append(want, '\n')
	if !bytes.Equal(candidate, want) {
		t.Fatalf("candidate = %s, want exact source row plus one planned row %s", candidate, want)
	}
}

// This fixture pins the open lineage decision and remaining acceptance cases.
func TestCreationContractFixtureDocumentsStaticAcceptanceCases(t *testing.T) {
	fixture := readCreationContractFixture(t)
	if fixture.CurrentSourceSHA256 != "768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de" ||
		fixture.PreviousNullLossSourceSHA != "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890" {
		t.Fatalf("fixture source lineage = current %q, previous %q", fixture.CurrentSourceSHA256, fixture.PreviousNullLossSourceSHA)
	}
	if fixture.CurrentSourceSHA256 == fixture.PreviousNullLossSourceSHA {
		t.Fatal("fixture must distinguish the new canonical source from the source covered by the historic null-loss exception")
	}
	if fixture.LineageDesign.Status != "open" || len(fixture.LineageDesign.Alternatives) != 2 ||
		!strings.Contains(fixture.LineageDesign.Invariant, "generic null-equals-absence rule is forbidden") {
		t.Fatalf("lineage design must preserve the unresolved alternatives and fail-closed null invariant: %+v", fixture.LineageDesign)
	}
	if fixture.SourceFreshness.OldBundleSourceSHA256 != fixture.PreviousNullLossSourceSHA ||
		fixture.SourceFreshness.CurrentSourceSHA256 != fixture.CurrentSourceSHA256 ||
		fixture.SourceFreshness.OldBundleAgainstCurrent != "reject" || fixture.SourceFreshness.FreshBundleAfterReview != "conditional" {
		t.Fatalf("source-freshness continuation contract = %+v", fixture.SourceFreshness)
	}
	if fixture.SnapshotCounts.Source != 1 || fixture.SnapshotCounts.Baseline != 1 ||
		fixture.SnapshotCounts.Updated != 2 || fixture.SnapshotCounts.Candidate != 2 || fixture.SnapshotCounts.RoundTrip != 2 {
		t.Fatalf("source/baseline/updated/candidate/round-trip counts = %+v", fixture.SnapshotCounts)
	}

	var plan struct {
		Creates []struct {
			ID           string `json:"id"`
			Dependencies []struct {
				IssueID     string `json:"issue_id"`
				DependsOnID string `json:"depends_on_id"`
				Type        string `json:"type"`
			} `json:"dependencies"`
		} `json:"creates"`
	}
	if err := json.Unmarshal(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Creates) != 1 || plan.Creates[0].ID != "dap-jo9.99" || len(plan.Creates[0].Dependencies) != 1 {
		t.Fatalf("proposed plan must contain exactly one child and edge: %+v", plan)
	}
	edge := plan.Creates[0].Dependencies[0]
	if edge.IssueID != "dap-jo9.99" || edge.DependsOnID != "dap-jo9" || edge.Type != "parent-child" {
		t.Fatalf("planned dependency edge = %+v, want child-to-existing-parent parent-child edge", edge)
	}
	var rawPlan map[string]json.RawMessage
	if err := json.Unmarshal(fixture.ProposedPlan, &rawPlan); err != nil {
		t.Fatal(err)
	}
	var records []map[string]json.RawMessage
	if err := json.Unmarshal(rawPlan["creates"], &records); err != nil || len(records) != 1 {
		t.Fatalf("planned create records = %s, err = %v; want one complete canonical issue row", rawPlan["creates"], err)
	}
	wantRecordFields := []string{
		"_type", "id", "title", "description", "acceptance_criteria", "notes", "status", "priority",
		"issue_type", "created_at", "created_by", "updated_at", "labels", "dependencies",
		"dependency_count", "dependent_count", "comment_count",
	}
	if len(records[0]) != len(wantRecordFields) {
		t.Fatalf("planned record fields = %v, want exactly %v", records[0], wantRecordFields)
	}
	for _, field := range wantRecordFields {
		if _, ok := records[0][field]; !ok {
			t.Errorf("planned canonical issue row is missing %q", field)
		}
	}
	var edgeRecords []map[string]json.RawMessage
	if err := json.Unmarshal(records[0]["dependencies"], &edgeRecords); err != nil || len(edgeRecords) != 1 {
		t.Fatalf("planned dependency records = %s, err = %v; want one complete edge", records[0]["dependencies"], err)
	}
	for _, field := range []string{"issue_id", "depends_on_id", "type", "created_at", "created_by", "metadata"} {
		if _, ok := edgeRecords[0][field]; !ok {
			t.Errorf("planned parent-child edge is missing %q", field)
		}
	}

	wantCases := []string{
		"exact_planned_record_and_parent_edge",
		"fresh_bundle_after_reviewed_lineage_continuation",
		"unplanned_record",
		"unplanned_dependency_edge",
		"unplanned_record_field",
		"duplicate_planned_id",
		"collision_with_existing_id",
		"stale_source_or_bundle",
		"unauthorized_source_lineage",
		"null_loss_without_exact_authorization",
	}
	if len(fixture.Cases) != len(wantCases) {
		t.Fatalf("acceptance case count = %d, want %d", len(fixture.Cases), len(wantCases))
	}
	for i, wantID := range wantCases {
		if fixture.Cases[i].ID != wantID {
			t.Errorf("acceptance case %d = %q, want %q", i, fixture.Cases[i].ID, wantID)
		}
		wantResult := "reject"
		if i == 0 {
			wantResult = "accept"
		} else if i == 1 {
			wantResult = "conditional"
		}
		if fixture.Cases[i].Expected != wantResult {
			t.Errorf("case %s expected %q, want %q", fixture.Cases[i].ID, fixture.Cases[i].Expected, wantResult)
		}
		if strings.TrimSpace(fixture.Cases[i].Mutation) == "" {
			t.Errorf("case %s is missing its concrete acceptance/rejection mutation", fixture.Cases[i].ID)
		}
	}
	if bytes.Contains(fixture.ProposedPlan, []byte(`"updates"`)) {
		t.Fatal("creation contract fixture must remain isolated from unrelated issue updates")
	}
}
