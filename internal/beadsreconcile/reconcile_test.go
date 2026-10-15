package beadsreconcile

import (
	"bytes"
	"strings"
	"testing"
)

func TestReconcileChangesOnlyPlannedFields(t *testing.T) {
	source := []byte("{ \"id\" : \"dap-1\", \"title\" : \"old\", \"updated_at\" : \"2026-01-01T00:00:00Z\", \"status\":\"open\", \"labels\":[\"one\"] }\n")
	baseline := []byte("{ \"id\" : \"dap-1\", \"title\" : \"old\", \"updated_at\" : \"2026-01-01T00:00:00Z\", \"status\":\"open\", \"labels\":[\"one\"] }\n")
	updated := []byte("{ \"id\" : \"dap-1\", \"title\" : \"new\", \"updated_at\" : \"2026-01-02T00:00:00Z\", \"status\":\"open\", \"labels\":[\"one\"] }\n")
	plan := Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"title": "new"}}}}

	got, evidence, err := Reconcile(source, baseline, updated, plan)
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	want := []byte("{ \"id\" : \"dap-1\", \"title\" : \"new\", \"updated_at\" : \"2026-01-02T00:00:00Z\", \"status\":\"open\", \"labels\":[\"one\"] }\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("candidate bytes differ\n got: %q\nwant: %q", got, want)
	}
	if evidence.SourceSHA256 == "" || evidence.BaselineSHA256 == "" || evidence.UpdatedSHA256 == "" || evidence.CandidateSHA256 == "" {
		t.Fatal("evidence is missing a required snapshot digest")
	}
}

func TestReconcileSplicesMultipleFieldsWithoutOffsetDrift(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"title\":\"x\",\"notes\":\"notes-old\",\"updated_at\":\"2026-01-01T00:00:00Z\",\"status\":\"open\"}\n")
	updated := []byte("{\"id\":\"dap-1\",\"title\":\"a much longer title\",\"notes\":\"n\",\"updated_at\":\"2026-01-02T00:00:00Z\",\"status\":\"open\"}\n")
	plan := Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"title": "a much longer title", "notes": "n"}}}}
	got, _, err := Reconcile(source, source, updated, plan)
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	want := []byte("{\"id\":\"dap-1\",\"title\":\"a much longer title\",\"notes\":\"n\",\"updated_at\":\"2026-01-02T00:00:00Z\",\"status\":\"open\"}\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("candidate bytes differ\n got: %q\nwant: %q", got, want)
	}
}

func TestReconcileAddsPlannedFieldWithoutReformattingSourceFields(t *testing.T) {
	source := []byte("{ \"id\": \"dap-1\", \"title\": \"old\", \"updated_at\": \"2026-01-01T00:00:00Z\", \"status\": \"open\" }\n")
	updated := []byte("{ \"id\": \"dap-1\", \"title\": \"old\", \"updated_at\": \"2026-01-02T00:00:00Z\", \"status\": \"open\", \"acceptance_criteria\": \"accepted\" }\n")
	got, _, err := Reconcile(source, source, updated, Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"acceptance_criteria": "accepted"}}}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	want := []byte("{ \"id\": \"dap-1\", \"title\": \"old\", \"updated_at\": \"2026-01-02T00:00:00Z\", \"status\": \"open\" ,\"acceptance_criteria\":\"accepted\"}\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("candidate bytes differ\n got: %q\nwant: %q", got, want)
	}
}

func TestReconcileFailsClosedForUnplannedMachineChange(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"title\":\"old\",\"updated_at\":\"2026-01-01T00:00:00Z\",\"status\":\"open\"}\n")
	updated := []byte("{\"id\":\"dap-1\",\"title\":\"new\",\"updated_at\":\"2026-01-02T00:00:00Z\",\"status\":\"closed\"}\n")
	_, _, err := Reconcile(source, source, updated, Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"title": "new"}}}})
	if err == nil || !strings.Contains(err.Error(), "unplanned") {
		t.Fatalf("Reconcile() error = %v, want unplanned-change error", err)
	}
}

func TestReconcileRejectsUnplannedUpdatedAtChange(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"title\":\"old\",\"updated_at\":\"2026-01-01T00:00:00Z\"}\n{\"id\":\"dap-2\",\"title\":\"unchanged\",\"updated_at\":\"2026-01-01T00:00:00Z\"}\n")
	updated := []byte("{\"id\":\"dap-1\",\"title\":\"new\",\"updated_at\":\"2026-01-02T00:00:00Z\"}\n{\"id\":\"dap-2\",\"title\":\"unchanged\",\"updated_at\":\"2026-01-02T00:00:00Z\"}\n")
	plan := Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"title": "new"}}}}

	if _, _, err := Reconcile(source, source, updated, plan); err == nil || !strings.Contains(err.Error(), "unplanned updated_at") {
		t.Fatalf("Reconcile() error = %v, want unplanned updated_at error", err)
	}
}

func TestReconcileRejectsDuplicateIDsAndMismatchedPlan(t *testing.T) {
	row := []byte("{\"id\":\"dap-1\",\"title\":\"same\",\"updated_at\":\"2026-01-01T00:00:00Z\"}\n")
	_, _, err := Reconcile(append(row, row...), append(row, row...), append(row, row...), Plan{})
	if err == nil || !strings.Contains(err.Error(), "duplicate issue id") {
		t.Fatalf("duplicate-ID error = %v", err)
	}
	_, _, err = Reconcile(row, row, row, Plan{Updates: []IssueUpdate{{ID: "dap-404", Fields: map[string]string{"title": "new"}}}})
	if err == nil || !strings.Contains(err.Error(), "not present") {
		t.Fatalf("mismatched-plan error = %v", err)
	}
}

func TestReconcileRejectsDuplicateNestedKeysAndDuplicateEdges(t *testing.T) {
	duplicateKey := []byte("{\"id\":\"dap-1\",\"title\":\"a\",\"title\":\"b\"}\n")
	if _, _, err := Reconcile(duplicateKey, duplicateKey, duplicateKey, Plan{}); err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
		t.Fatalf("duplicate-key error = %v", err)
	}
	rows := []byte("{\"id\":\"dap-1\",\"dependencies\":[{\"issue_id\":\"dap-1\",\"depends_on_id\":\"dap-2\",\"type\":\"blocks\"},{\"issue_id\":\"dap-1\",\"depends_on_id\":\"dap-2\",\"type\":\"blocks\"}]}\n{\"id\":\"dap-2\"}\n")
	if _, _, err := Reconcile(rows, rows, rows, Plan{}); err == nil || !strings.Contains(err.Error(), "duplicate dependency edge") {
		t.Fatalf("duplicate-edge error = %v", err)
	}
}

func TestReconcileAllowsPinnedImporterNormalizationWithoutCopyingDerivedCloseTime(t *testing.T) {
	source := []byte("{\"id\":\"dap-jo9.5.5\",\"status\":\"closed\",\"created_at\":\"2026-01-01T00:00:00Z\",\"updated_at\":\"2026-01-01T00:00:01Z\"}\n")
	baseline := []byte("{\"_type\":\"issue\",\"id\":\"dap-jo9.5.5\",\"status\":\"closed\",\"created_at\":\"2026-01-01T00:00:00Z\",\"updated_at\":\"2026-01-01T00:00:01Z\",\"closed_at\":\"2026-01-01T00:00:02Z\",\"comment_count\":0}\n")
	plan := Plan{BDVersion: "1.3.1"}
	got, _, err := Reconcile(source, baseline, baseline, plan)
	if err != nil {
		t.Fatalf("pinned normalization rejected: %v", err)
	}
	if !bytes.Equal(got, source) {
		t.Fatalf("derived timestamp leaked into candidate: %q", got)
	}
	changed := bytes.Replace(baseline, []byte("00:00:02Z"), []byte("00:00:03Z"), 1)
	if _, _, err := Reconcile(source, changed, changed, plan); err == nil || !strings.Contains(err.Error(), "deterministic formula") {
		t.Fatalf("changed derived timestamp error = %v", err)
	}
}

func TestReconcileRejectsChangedDependencyGraph(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"dependencies\":[{\"issue_id\":\"dap-1\",\"depends_on_id\":\"dap-2\",\"type\":\"blocks\"}]}\n{\"id\":\"dap-2\"}\n")
	changed := []byte("{\"id\":\"dap-1\",\"dependencies\":[]}\n{\"id\":\"dap-2\"}\n")
	if _, _, err := Reconcile(source, source, changed, Plan{}); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("changed dependency graph error = %v", err)
	}
}

func TestImporterDefaultsAreScopedToFreshImport(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"title\":\"same\"}\n")
	baseline := []byte("{\"id\":\"dap-1\",\"_type\":\"issue\",\"title\":\"same\",\"comment_count\":0}\n")
	if _, _, err := Reconcile(source, baseline, baseline, Plan{}); err != nil {
		t.Fatalf("known source-to-import defaults rejected: %v", err)
	}
	updated := []byte("{\"id\":\"dap-1\",\"_type\":\"issue\",\"title\":\"same\"}\n")
	if _, _, err := Reconcile(source, baseline, updated, Plan{}); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("post-import default removal error = %v", err)
	}
}

func TestImporterCannotEraseExplicitNullValues(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"closed_at\":null}\n")
	imported := []byte("{\"id\":\"dap-1\"}\n")
	if _, _, err := Reconcile(source, imported, imported, Plan{}); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("null-to-absent normalization error = %v", err)
	}
	source = []byte("{\"id\":\"dap-1\",\"dependencies\":null}\n")
	imported = []byte("{\"id\":\"dap-1\"}\n")
	if _, _, err := Reconcile(source, imported, imported, Plan{}); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("null dependency normalization error = %v", err)
	}
}
