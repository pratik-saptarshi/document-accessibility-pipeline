package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const syntheticParent = "{\"_type\":\"issue\",\"id\":\"dap-jo9\",\"title\":\"parent\",\"status\":\"open\",\"dependencies\":[]}\n{\"_type\":\"issue\",\"id\":\"dap-other\",\"title\":\"other target\",\"status\":\"open\",\"dependencies\":[]}\n"

func syntheticCreate(t *testing.T) json.RawMessage {
	t.Helper()
	row := map[string]any{
		"_type": "issue", "id": "dap-jo9.99", "title": "synthetic child",
		"description": "synthetic", "acceptance_criteria": "exact addition", "notes": "test only",
		"status": "open", "priority": 2, "issue_type": "task",
		"created_at": "2026-10-16T00:00:00Z", "created_by": "synthetic-fixture",
		"updated_at": "2026-10-16T00:00:00Z", "labels": []string{},
		"dependencies": []any{map[string]any{
			"issue_id": "dap-jo9.99", "depends_on_id": "dap-jo9", "type": "parent-child",
			"created_at": "2026-10-16T00:00:00Z", "created_by": "synthetic-fixture", "metadata": "{}",
		}},
		"dependency_count": 0, "dependent_count": 0, "comment_count": 0,
	}
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func syntheticPlan(t *testing.T) (Plan, []byte, []byte) {
	t.Helper()
	parent := []byte(syntheticParent)
	child := syntheticCreate(t)
	updated := append(bytes.TrimSuffix(append([]byte(nil), parent...), []byte("\n")), '\n')
	updated = append(updated, child...)
	updated = append(updated, '\n')
	return Plan{Creates: []json.RawMessage{child}}, parent, updated
}

func TestReconcileCreationRejectsUnplannedChanges(t *testing.T) {
	plan, source, updated := syntheticPlan(t)
	tests := []struct {
		name    string
		plan    Plan
		updated []byte
		want    string
	}{
		{name: "extra record", plan: plan, updated: append(append([]byte(nil), updated...), []byte("{\"id\":\"dap-extra\"}\n")...), want: "ID set changed"},
		{name: "extra field", plan: plan, updated: bytes.Replace(updated, []byte("\"notes\":\"test only\""), []byte("\"notes\":\"test only\",\"owner\":\"unexpected\""), 1), want: "exact planned record"},
		{name: "changed edge", plan: plan, updated: bytes.Replace(updated, []byte("\"depends_on_id\":\"dap-jo9\""), []byte("\"depends_on_id\":\"dap-missing\""), 1), want: "missing"},
		{name: "extra edge to existing issue", plan: plan, updated: bytes.Replace(updated,
			[]byte(`],"dependency_count"`),
			[]byte(`,{"issue_id":"dap-jo9.99","depends_on_id":"dap-other","type":"blocks","created_at":"2026-10-16T00:00:00Z","created_by":"synthetic-fixture","metadata":"{}"}],"dependency_count"`), 1),
			want: "exact planned record"},
		{name: "changed planned field", plan: plan, updated: bytes.Replace(updated, []byte("synthetic child"), []byte("unplanned child"), 1), want: "exact planned record"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := Reconcile(source, source, tt.updated, tt.plan); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Reconcile() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestReconcileCreationRejectsDuplicateOrCollidingPlanIDs(t *testing.T) {
	plan, source, updated := syntheticPlan(t)
	duplicate := plan
	duplicate.Creates = append(append([]json.RawMessage(nil), plan.Creates...), plan.Creates[0])
	if _, _, err := Reconcile(source, source, updated, duplicate); err == nil || !strings.Contains(err.Error(), "at most one") {
		t.Fatalf("duplicate create plan error = %v, want bounded one-create rejection", err)
	}
	collision := bytes.Replace(plan.Creates[0], []byte("\"id\":\"dap-jo9.99\""), []byte("\"id\":\"dap-jo9\""), 1)
	collision = bytes.Replace(collision, []byte("\"issue_id\":\"dap-jo9.99\""), []byte("\"issue_id\":\"dap-jo9\""), 1)
	plan.Creates = []json.RawMessage{collision}
	if _, _, err := Reconcile(source, source, updated, plan); err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("colliding create plan error = %v, want collision rejection", err)
	}
}

func TestValidatePlanRejectsIncompleteAndUnsupportedCreateRows(t *testing.T) {
	plan, source, _ := syntheticPlan(t)
	tests := []struct {
		name string
		edit func(json.RawMessage) json.RawMessage
	}{
		{name: "missing required field", edit: func(raw json.RawMessage) json.RawMessage {
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(raw, &fields)
			delete(fields, "title")
			got, _ := json.Marshal(fields)
			return got
		}},
		{name: "unplanned field", edit: func(raw json.RawMessage) json.RawMessage {
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(raw, &fields)
			fields["owner"] = json.RawMessage(`"unexpected"`)
			got, _ := json.Marshal(fields)
			return got
		}},
		{name: "nonnumeric child ID", edit: func(raw json.RawMessage) json.RawMessage {
			return json.RawMessage(bytes.Replace(raw, []byte("dap-jo9.99"), []byte("dap-jo9.child"), 2))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := plan
			candidate.Creates = []json.RawMessage{tt.edit(plan.Creates[0])}
			if err := ValidatePlan(source, candidate); err == nil {
				t.Fatal("ValidatePlan() accepted incomplete, unsupported, or malformed creation")
			}
		})
	}
}

func TestValidateCandidateRoundTripRetainsCreatedRecordAndProvenance(t *testing.T) {
	plan, source, updated := syntheticPlan(t)
	candidate, evidence, err := Reconcile(source, source, updated, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence.CreatedIDs) != 1 || evidence.CreatedIDs[0] != "dap-jo9.99" || len(evidence.AllIssueIDs) != 3 {
		t.Fatalf("creation evidence = %+v", evidence)
	}
	creationRow := bytes.TrimPrefix(candidate, source)
	first, second, err := ValidateCandidateRoundTrip(source, source, updated, candidate, candidate, plan, "")
	if err != nil {
		t.Fatalf("ValidateCandidateRoundTrip() rejected exact synthetic bundle: %v", err)
	}
	if len(first.CreatedIDs) != 1 || first.CreatedIDs[0] != "dap-jo9.99" || len(second.AllIssueIDs) != 3 || !bytes.Equal(creationRow, bytes.TrimPrefix(candidate, source)) {
		t.Fatalf("round-trip creation provenance = first %+v second %+v", first, second)
	}
}
