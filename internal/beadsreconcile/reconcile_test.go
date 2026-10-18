package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var approvedNullLosses = []struct {
	id    string
	field string
}{
	{"dap-jo9.5.3", "dependencies"},
	{"dap-jo9.1.6.10", "closed_at"},
	{"dap-jo9.5.6", "closed_at"},
	{"dap-jo9.2.1.4", "closed_at"},
	{"dap-jo9.5.7", "closed_at"},
}

const approvedNullLossBDBinarySHA256 = "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653"

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

func TestReconcileRejectsNullForPlannedStringAndAllowsExplicitEmptyString(t *testing.T) {
	source := []byte("{\"id\":\"dap-1\",\"description\":\"old\",\"updated_at\":\"2026-10-14T00:00:00Z\"}\n")
	baseline := source
	plan := Plan{Updates: []IssueUpdate{{ID: "dap-1", Fields: map[string]string{"description": ""}}}}

	t.Run("null is not an empty string", func(t *testing.T) {
		updated := []byte("{\"id\":\"dap-1\",\"description\":null,\"updated_at\":\"2026-10-15T00:00:00Z\"}\n")
		if _, _, err := Reconcile(source, baseline, updated, plan); err == nil {
			t.Fatal("Reconcile() accepted JSON null as the planned empty description string")
		}
	})

	t.Run("explicit empty string is accepted", func(t *testing.T) {
		updated := []byte("{\"id\":\"dap-1\",\"description\":\"\",\"updated_at\":\"2026-10-15T00:00:00Z\"}\n")
		candidate, _, err := Reconcile(source, baseline, updated, plan)
		if err != nil {
			t.Fatalf("Reconcile() rejected an explicit empty string: %v", err)
		}
		want := []byte("{\"id\":\"dap-1\",\"description\":\"\",\"updated_at\":\"2026-10-15T00:00:00Z\"}\n")
		if !bytes.Equal(candidate, want) {
			t.Fatalf("candidate = %q, want %q", candidate, want)
		}
	})
}

func TestDifferingFieldsWithOverlappingAndDisjointKeys(t *testing.T) {
	t.Run("overlapping keys", func(t *testing.T) {
		left := map[string]json.RawMessage{
			"same":    json.RawMessage(`"value"`),
			"changed": json.RawMessage(`1`),
			"left":    json.RawMessage(`true`),
		}
		right := map[string]json.RawMessage{
			"same":    json.RawMessage(`"value"`),
			"changed": json.RawMessage(`2`),
			"right":   json.RawMessage(`false`),
		}
		want := []string{"changed", "left", "right"}
		if got := differingFields(left, right); !equalStrings(got, want) {
			t.Fatalf("differingFields() = %v, want %v", got, want)
		}
	})

	t.Run("disjoint keys", func(t *testing.T) {
		left := map[string]json.RawMessage{"left": json.RawMessage(`1`)}
		right := map[string]json.RawMessage{"right": json.RawMessage(`2`)}
		want := []string{"left", "right"}
		if got := differingFields(left, right); !equalStrings(got, want) {
			t.Fatalf("differingFields() = %v, want %v", got, want)
		}
	})
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
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
	if _, _, err := Reconcile(duplicateKey, duplicateKey, duplicateKey, Plan{}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate-key error = %v", err)
	}
	rows := []byte("{\"id\":\"dap-1\",\"dependencies\":[{\"issue_id\":\"dap-1\",\"depends_on_id\":\"dap-2\",\"type\":\"blocks\"},{\"issue_id\":\"dap-1\",\"depends_on_id\":\"dap-2\",\"type\":\"blocks\"}]}\n{\"id\":\"dap-2\"}\n")
	if _, _, err := Reconcile(rows, rows, rows, Plan{}); err == nil || !strings.Contains(err.Error(), "duplicate dependency edge") {
		t.Fatalf("duplicate-edge error = %v", err)
	}
}

// TestReconcileRejectsCaseVariantDuplicateKeys pins the jsonstrict
// consolidation: keys that fold to the same encoding/json field name are
// rejected, not just byte-identical duplicates.
func TestReconcileRejectsCaseVariantDuplicateKeys(t *testing.T) {
	variant := []byte("{\"id\":\"dap-1\",\"Title\":\"a\",\"title\":\"b\"}\n")
	if _, _, err := Reconcile(variant, variant, variant, Plan{}); err == nil || !strings.Contains(err.Error(), "case-variant") {
		t.Fatalf("case-variant-key error = %v", err)
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

func TestReconcileScopedNullLossPreservesCanonicalBytesOnUpdatedRecord(t *testing.T) {
	source, baseline, updated, plan := approvedNullLossFixture(t)
	candidate, evidence, err := ReconcileWithBinary(source, baseline, updated, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("Reconcile() rejected the reviewed Beads null-loss fixture: %v", err)
	}
	if evidence.SourceSHA256 != "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890" {
		t.Fatalf("source digest = %s, want the approved canonical source digest", evidence.SourceSHA256)
	}
	for _, pair := range approvedNullLosses {
		row := findTestRow(t, candidate, pair.id)
		if raw, ok := row[pair.field]; !ok || string(bytes.TrimSpace(raw)) != "null" {
			t.Fatalf("candidate %s.%s = %s, present=%t; want original explicit null", pair.id, pair.field, raw, ok)
		}
	}
	updatedRow := findTestRow(t, candidate, "dap-jo9.5.3")
	var title string
	if err := json.Unmarshal(updatedRow["title"], &title); err != nil || title != "Scoped null-loss regression title" {
		t.Fatalf("candidate title = %q, error=%v", title, err)
	}
	if len(evidence.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("evidence loss count = %d, want %d", len(evidence.RepresentationLosses), len(approvedNullLosses))
	}
	if len(evidence.CandidateNullSpans) != len(approvedNullLosses) {
		t.Fatalf("candidate null span count = %d, want %d", len(evidence.CandidateNullSpans), len(approvedNullLosses))
	}
	for _, span := range evidence.CandidateNullSpans {
		lines := splitLines(candidate)
		if span.Line < 1 || span.Line > len(lines) || span.Start < 0 || span.End > len(lines[span.Line-1]) || string(lines[span.Line-1][span.Start:span.End]) != "null" || span.Raw != "null" {
			t.Fatalf("candidate span does not identify the retained null token: %+v", span)
		}
	}
}

func TestApprovedCurrentSourceContinuationPreservesNulls(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "approved-current-source-canonical.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	const approvedSourceSHA256 = "768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de"
	if got := digest(source); got != approvedSourceSHA256 {
		t.Fatalf("current authorized source digest = %s, want %s", got, approvedSourceSHA256)
	}
	baseline := importerNullLossFixture(t, source)
	plan := Plan{BDVersion: approvedNullLossVersion}

	candidate, evidence, err := ReconcileWithBinary(source, baseline, baseline, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact approved continuation source was rejected: %v", err)
	}
	if !bytes.Equal(candidate, source) {
		t.Fatal("continuation candidate changed canonical bytes without a planned update")
	}
	if evidence.SourceSHA256 != approvedSourceSHA256 || len(evidence.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("continuation evidence source/losses = %s/%d, want %s/%d", evidence.SourceSHA256, len(evidence.RepresentationLosses), approvedSourceSHA256, len(approvedNullLosses))
	}
	for _, loss := range evidence.RepresentationLosses {
		if loss.CanonicalSourceSHA != approvedSourceSHA256 || !loss.BaselineAbsent || !loss.UpdatedAbsent || loss.CanonicalValue != "null" {
			t.Fatalf("continuation evidence broadened or weakened null-loss scope: %+v", loss)
		}
	}

	secondStore := importerNullLossFixture(t, candidate)
	first, second, err := ValidateCandidateRoundTrip(source, baseline, baseline, candidate, secondStore, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact approved continuation failed second-store validation: %v", err)
	}
	if len(first.RepresentationLosses) != len(approvedNullLosses) || len(second.CandidateNullSpans) != len(approvedNullLosses) {
		t.Fatalf("round-trip continuation evidence has %d losses and %d null spans", len(first.RepresentationLosses), len(second.CandidateNullSpans))
	}

	if _, _, err := ReconcileWithBinary(source, baseline, baseline, plan, strings.Repeat("0", 64)); err == nil {
		t.Fatal("continuation accepted a different Beads binary digest")
	}
	wrongVersion := plan
	wrongVersion.BDVersion = "1.3.2"
	if _, _, err := ReconcileWithBinary(source, baseline, baseline, wrongVersion, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("continuation accepted a different Beads version")
	}
	changedSource := replaceTestField(t, source, "dap-jo9.5.3", "title", json.RawMessage(`"unapproved source edit"`))
	changedBaseline := importerNullLossFixture(t, changedSource)
	if _, _, err := ReconcileWithBinary(changedSource, changedBaseline, changedBaseline, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("continuation accepted a source digest outside the exact approvals")
	}
}

func TestApprovedIssueCreationSourceContinuationPreservesNulls(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "approved-issue-creation-source.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	const approvedSourceSHA256 = "5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465"
	if got := digest(source); got != approvedSourceSHA256 {
		t.Fatalf("issue-creation continuation digest = %s, want %s", got, approvedSourceSHA256)
	}
	baseline := importerNullLossFixture(t, source)
	plan := Plan{BDVersion: approvedNullLossVersion}

	candidate, evidence, err := ReconcileWithBinary(source, baseline, baseline, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact approved issue-creation continuation was rejected: %v", err)
	}
	if !bytes.Equal(candidate, source) {
		t.Fatal("continuation candidate changed canonical bytes without a planned update")
	}
	if evidence.SourceSHA256 != approvedSourceSHA256 || len(evidence.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("continuation evidence source/losses = %s/%d, want %s/%d", evidence.SourceSHA256, len(evidence.RepresentationLosses), approvedSourceSHA256, len(approvedNullLosses))
	}
	wantPairs := make(map[fieldPair]bool, len(approvedNullLosses))
	for _, pair := range approvedNullLosses {
		wantPairs[fieldPair{issueID: pair.id, field: pair.field}] = true
	}
	gotPairs := make(map[fieldPair]bool, len(evidence.RepresentationLosses))
	for _, loss := range evidence.RepresentationLosses {
		if loss.CanonicalSourceSHA != approvedSourceSHA256 || !loss.BaselineAbsent || !loss.UpdatedAbsent || loss.CanonicalValue != "null" {
			t.Fatalf("issue-creation continuation broadened or weakened null-loss scope: %+v", loss)
		}
		pair := fieldPair{issueID: loss.IssueID, field: loss.Field}
		if !wantPairs[pair] || gotPairs[pair] {
			t.Fatalf("issue-creation continuation reported an unapproved or duplicate null-loss pair: %+v", pair)
		}
		gotPairs[pair] = true
	}
	if len(gotPairs) != len(wantPairs) {
		t.Fatalf("issue-creation continuation evidence pairs = %v, want exactly %v", gotPairs, wantPairs)
	}
	for _, pair := range approvedNullLosses {
		row := findTestRow(t, candidate, pair.id)
		if raw, ok := row[pair.field]; !ok || string(bytes.TrimSpace(raw)) != "null" {
			t.Fatalf("candidate %s.%s = %s, present=%t; want original explicit null", pair.id, pair.field, raw, ok)
		}
	}

	secondStore := importerNullLossFixture(t, candidate)
	first, second, err := ValidateCandidateRoundTrip(source, baseline, baseline, candidate, secondStore, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact approved issue-creation continuation failed second-store validation: %v", err)
	}
	if len(first.RepresentationLosses) != len(approvedNullLosses) || len(second.CandidateNullSpans) != len(approvedNullLosses) {
		t.Fatalf("round-trip continuation evidence has %d losses and %d null spans", len(first.RepresentationLosses), len(second.CandidateNullSpans))
	}

	if _, _, err := ReconcileWithBinary(source, baseline, baseline, plan, strings.Repeat("0", 64)); err == nil {
		t.Fatal("issue-creation continuation accepted a different Beads binary digest")
	}
	wrongVersion := plan
	wrongVersion.BDVersion = "1.3.2"
	if _, _, err := ReconcileWithBinary(source, baseline, baseline, wrongVersion, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("issue-creation continuation accepted a different Beads version")
	}
	changedSource := replaceTestField(t, source, "dap-jo9.5.3", "title", json.RawMessage(`"unapproved issue-creation source edit"`))
	changedBaseline := importerNullLossFixture(t, changedSource)
	if _, _, err := ReconcileWithBinary(changedSource, changedBaseline, changedBaseline, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("issue-creation continuation accepted a source digest outside the exact approvals")
	}
	changedBaseline = replaceTestField(t, baseline, "dap-jo9.5.3", "dependencies", json.RawMessage(`null`))
	if _, _, err := ReconcileWithBinary(source, changedBaseline, baseline, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("issue-creation continuation accepted a non-absent baseline value for an approved field")
	}
	changedUpdated := replaceTestField(t, baseline, "dap-jo9.2.3.2", "dependencies", json.RawMessage(`[]`))
	if _, _, err := ReconcileWithBinary(source, baseline, changedUpdated, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("issue-creation continuation accepted an unexpected dependency change")
	}
}

func TestApprovedLatestCanonicalSourceContinuationPreservesOnlyFiveNulls(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "approved-latest-null-loss-source.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	const approvedSourceSHA256 = "69d6079754284f5b81fea74cda717594851f498848cda69c5809044711a754cb"
	if got := digest(source); got != approvedSourceSHA256 {
		t.Fatalf("latest authorized source digest = %s, want %s", got, approvedSourceSHA256)
	}
	baseline := importerNullLossFixture(t, source)
	const updatedTitle = "Scoped current-source null-preservation regression"
	updated := replaceTestField(t, baseline, "dap-jo9.5.3", "title", json.RawMessage(`"`+updatedTitle+`"`))
	updated = replaceTestField(t, updated, "dap-jo9.5.3", "updated_at", json.RawMessage(`"2026-10-17T00:00:00Z"`))
	plan := Plan{BDVersion: approvedNullLossVersion, Updates: []IssueUpdate{{
		ID: "dap-jo9.5.3", Fields: map[string]string{"title": updatedTitle},
	}}}

	candidate, evidence, err := ReconcileWithBinary(source, baseline, updated, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact latest source continuation was rejected: %v", err)
	}
	if evidence.SourceSHA256 != approvedSourceSHA256 || len(evidence.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("continuation evidence source/losses = %s/%d, want %s/%d", evidence.SourceSHA256, len(evidence.RepresentationLosses), approvedSourceSHA256, len(approvedNullLosses))
	}
	wantPairs := make(map[fieldPair]bool, len(approvedNullLosses))
	for _, pair := range approvedNullLosses {
		wantPairs[fieldPair{issueID: pair.id, field: pair.field}] = true
		row := findTestRow(t, candidate, pair.id)
		if raw, ok := row[pair.field]; !ok || string(bytes.TrimSpace(raw)) != "null" {
			t.Fatalf("candidate %s.%s = %s, present=%t; want source null token", pair.id, pair.field, raw, ok)
		}
	}
	gotPairs := make(map[fieldPair]bool, len(evidence.RepresentationLosses))
	for _, loss := range evidence.RepresentationLosses {
		pair := fieldPair{issueID: loss.IssueID, field: loss.Field}
		if loss.CanonicalSourceSHA != approvedSourceSHA256 || !loss.BaselineAbsent || !loss.UpdatedAbsent || loss.CanonicalValue != "null" || !wantPairs[pair] || gotPairs[pair] {
			t.Fatalf("continuation evidence broadened or weakened null-loss scope: %+v", loss)
		}
		gotPairs[pair] = true
	}
	if len(gotPairs) != len(wantPairs) {
		t.Fatalf("continuation evidence pairs = %v, want exactly %v", gotPairs, wantPairs)
	}

	// The existing closed issue's evidence remains byte-for-byte sourced from C.
	for _, field := range []string{"status", "closed_at", "close_reason", "updated_at"} {
		before, beforeOK := findTestRow(t, source, "dap-jo9.18")[field]
		after, afterOK := findTestRow(t, candidate, "dap-jo9.18")[field]
		if beforeOK != afterOK || !bytes.Equal(before, after) {
			t.Fatalf("candidate changed dap-jo9.18.%s: source=%s candidate=%s", field, before, after)
		}
	}
	if _, present := findTestRow(t, source, "dap-jo9.5.5")["closed_at"]; present {
		t.Fatal("test source unexpectedly contains dap-jo9.5.5.closed_at")
	}
	if _, present := findTestRow(t, baseline, "dap-jo9.5.5")["closed_at"]; !present {
		t.Fatal("importer fixture did not generate the known derived closed_at")
	}
	if _, present := findTestRow(t, candidate, "dap-jo9.5.5")["closed_at"]; present {
		t.Fatal("candidate promoted importer-generated dap-jo9.5.5.closed_at")
	}
	if got := string(findTestRow(t, candidate, "dap-jo9.5.3")["title"]); got != `"`+updatedTitle+`"` {
		t.Fatalf("planned title = %s, want %q", got, updatedTitle)
	}

	roundTrip := importerNullLossFixture(t, candidate)
	first, second, err := ValidateCandidateRoundTrip(source, baseline, updated, candidate, roundTrip, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact latest source failed second-store validation: %v", err)
	}
	if len(first.RepresentationLosses) != len(approvedNullLosses) || len(second.CandidateNullSpans) != len(approvedNullLosses) {
		t.Fatalf("round-trip evidence has %d losses and %d spans, want %d each", len(first.RepresentationLosses), len(second.CandidateNullSpans), len(approvedNullLosses))
	}

	if _, _, err := ReconcileWithBinary(source, baseline, updated, plan, strings.Repeat("0", 64)); err == nil {
		t.Fatal("latest source continuation accepted a different Beads binary digest")
	}
	changedSource := replaceTestField(t, source, "dap-jo9.5.6", "title", json.RawMessage(`"changed canonical source"`))
	changedBaseline := importerNullLossFixture(t, changedSource)
	changedUpdated := replaceTestField(t, changedBaseline, "dap-jo9.5.3", "title", json.RawMessage(`"`+updatedTitle+`"`))
	changedUpdated = replaceTestField(t, changedUpdated, "dap-jo9.5.3", "updated_at", json.RawMessage(`"2026-10-17T00:00:00Z"`))
	if _, _, err := ReconcileWithBinary(changedSource, changedBaseline, changedUpdated, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("latest source continuation accepted a modified source digest")
	}
	changedUnlistedField := replaceTestField(t, updated, "dap-jo9.18", "close_reason", json.RawMessage(`"unapproved importer field change"`))
	if _, _, err := ReconcileWithBinary(source, baseline, changedUnlistedField, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("latest source continuation accepted an unlisted field change")
	}
	changedBaseline = replaceTestField(t, baseline, "dap-jo9.5.6", "closed_at", json.RawMessage(`null`))
	if _, _, err := ReconcileWithBinary(source, changedBaseline, updated, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("latest source continuation accepted a non-absent baseline value for an approved null field")
	}
}

func TestApprovedPR112CorrectedSourcePreservesOnlyFourCurrentNulls(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("testdata", "approved-pr112-corrected-source.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	const approvedSourceSHA256 = "6d2993b1248396fb1b180423551b67346e4f6a576060d487dd7d4cb768448459"
	if got := digest(source); got != approvedSourceSHA256 {
		t.Fatalf("PR112-corrected source digest = %s, want %s", got, approvedSourceSHA256)
	}
	if _, present := findTestRow(t, source, "dap-jo9.1.6.7")["notes"]; present {
		t.Fatal("corrected canonical source unexpectedly contains dap-jo9.1.6.7.notes")
	}
	if got := string(findTestRow(t, source, "dap-jo9.1.6.10")["status"]); got != `"closed"` {
		t.Fatalf("PR #114 closure status = %s, want closed", got)
	}

	baseline := importerNullLossFixture(t, source)
	const updatedTitle = "Bounded PR112 correction source continuation regression"
	updated := replaceTestField(t, baseline, "dap-jo9.5.3", "title", json.RawMessage(`"`+updatedTitle+`"`))
	updated = replaceTestField(t, updated, "dap-jo9.5.3", "updated_at", json.RawMessage(`"2026-10-17T00:00:00Z"`))
	plan := Plan{BDVersion: approvedNullLossVersion, Updates: []IssueUpdate{{
		ID: "dap-jo9.5.3", Fields: map[string]string{"title": updatedTitle},
	}}}

	candidate, evidence, err := ReconcileWithBinary(source, baseline, updated, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact PR112-corrected source continuation was rejected: %v", err)
	}
	if evidence.SourceSHA256 != approvedSourceSHA256 {
		t.Fatalf("continuation evidence source = %s, want %s", evidence.SourceSHA256, approvedSourceSHA256)
	}
	wantPairs := []fieldPair{
		{issueID: "dap-jo9.5.3", field: "dependencies"},
		{issueID: "dap-jo9.5.6", field: "closed_at"},
		{issueID: "dap-jo9.2.1.4", field: "closed_at"},
		{issueID: "dap-jo9.5.7", field: "closed_at"},
	}
	if len(evidence.RepresentationLosses) != len(wantPairs) {
		t.Fatalf("continuation losses = %d, want exact source-specific set of %d", len(evidence.RepresentationLosses), len(wantPairs))
	}
	want := make(map[fieldPair]bool, len(wantPairs))
	for _, pair := range wantPairs {
		want[pair] = true
		row := findTestRow(t, candidate, pair.issueID)
		if got, ok := row[pair.field]; !ok || string(bytes.TrimSpace(got)) != "null" {
			t.Fatalf("candidate %s.%s = %s, present=%t; want preserved source null", pair.issueID, pair.field, got, ok)
		}
	}
	for _, loss := range evidence.RepresentationLosses {
		pair := fieldPair{issueID: loss.IssueID, field: loss.Field}
		if loss.CanonicalSourceSHA != approvedSourceSHA256 || !loss.BaselineAbsent || !loss.UpdatedAbsent || loss.CanonicalValue != "null" || !want[pair] {
			t.Fatalf("continuation loss is outside exact source-specific approval: %+v", loss)
		}
		delete(want, pair)
	}
	if len(want) != 0 {
		t.Fatalf("continuation omitted approved null-loss pairs: %v", want)
	}
	if _, present := findTestRow(t, candidate, "dap-jo9.1.6.7")["notes"]; present {
		t.Fatal("candidate reintroduced dap-jo9.1.6.7.notes")
	}
	for _, field := range []string{"status", "closed_at", "updated_at"} {
		before, beforeOK := findTestRow(t, source, "dap-jo9.1.6.10")[field]
		after, afterOK := findTestRow(t, candidate, "dap-jo9.1.6.10")[field]
		if beforeOK != afterOK || !bytes.Equal(before, after) {
			t.Fatalf("candidate changed PR #114 closure field %s: source=%s candidate=%s", field, before, after)
		}
	}

	roundTrip := importerNullLossFixture(t, candidate)
	first, second, err := ValidateCandidateRoundTrip(source, baseline, updated, candidate, roundTrip, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("exact PR112-corrected source failed second-store validation: %v", err)
	}
	if len(first.RepresentationLosses) != len(wantPairs) || len(second.CandidateNullSpans) != len(wantPairs) {
		t.Fatalf("round-trip evidence has %d losses and %d spans, want %d each", len(first.RepresentationLosses), len(second.CandidateNullSpans), len(wantPairs))
	}

	changedSource := replaceTestField(t, source, "dap-jo9.1.6.7", "description", json.RawMessage(`"changed canonical source"`))
	changedBaseline := importerNullLossFixture(t, changedSource)
	changedUpdated := replaceTestField(t, changedBaseline, "dap-jo9.5.3", "title", json.RawMessage(`"`+updatedTitle+`"`))
	changedUpdated = replaceTestField(t, changedUpdated, "dap-jo9.5.3", "updated_at", json.RawMessage(`"2026-10-17T00:00:00Z"`))
	if _, _, err := ReconcileWithBinary(changedSource, changedBaseline, changedUpdated, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("PR112-corrected continuation accepted a modified canonical source digest")
	}
}

func TestApprovedNullLossRegressionUsesHistoricFixtureAndRejectsLaterCanonicalEdit(t *testing.T) {
	source, baseline, updated, plan := approvedNullLossFixture(t)
	fixture, err := os.ReadFile(filepath.Join("testdata", "approved-null-loss-canonical.jsonl"))
	if err != nil {
		t.Fatalf("read immutable approved source fixture: %v", err)
	}
	if !bytes.Equal(source, fixture) {
		t.Fatal("approved null-loss source must come from the immutable historic fixture")
	}
	if digest(source) != "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890" {
		t.Fatalf("historic fixture digest = %s, want the exact approved source digest", digest(source))
	}

	if _, _, err := ReconcileWithBinary(source, baseline, updated, plan, approvedNullLossBDBinarySHA256); err != nil {
		t.Fatalf("historic fixture regression was invalidated by later canonical edits: %v", err)
	}
	laterCanonical := replaceTestField(t, fixture, "dap-jo9.5.3", "title", json.RawMessage(`"unrelated later canonical edit"`))
	if digest(laterCanonical) == digest(fixture) {
		t.Fatal("test setup did not create a later canonical source digest")
	}
	laterBaseline := replaceTestField(t, baseline, "dap-jo9.5.3", "title", json.RawMessage(`"unrelated later canonical edit"`))
	if _, _, err := ReconcileWithBinary(laterCanonical, laterBaseline, updated, plan, approvedNullLossBDBinarySHA256); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("production null-loss exception must reject only because this otherwise-consistent source digest is unapproved: %v", err)
	}
}

func TestReconcileScopedNullLossRejectsWrongBinaryAndChangedNullField(t *testing.T) {
	source, baseline, updated, plan := approvedNullLossFixture(t)
	if _, _, err := ReconcileWithBinary(source, baseline, updated, plan, strings.Repeat("0", 64)); err == nil {
		t.Fatal("null-loss exception accepted an unapproved Beads binary digest")
	}
	changedSource := replaceTestField(t, source, "dap-jo9.5.3", "title", json.RawMessage(`"changed source"`))
	if _, _, err := ReconcileWithBinary(changedSource, baseline, updated, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("null-loss exception accepted a changed canonical source digest")
	}
	changed := replaceTestField(t, updated, "dap-jo9.5.3", "dependencies", json.RawMessage(`[{"issue_id":"dap-jo9.5.3","depends_on_id":"dap-jo9.5.4","type":"blocks"}]`))
	if _, _, err := ReconcileWithBinary(source, baseline, changed, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("null-loss exception accepted a changed dependency field or graph")
	}
	changedBaseline := replaceTestField(t, baseline, "dap-jo9.5.6", "closed_at", json.RawMessage(`null`))
	if _, _, err := ReconcileWithBinary(source, changedBaseline, updated, plan, approvedNullLossBDBinarySHA256); err == nil || !strings.Contains(err.Error(), "baseline") {
		t.Fatalf("null-loss exception accepted a non-absent B0 value: %v", err)
	}
	changedGraph := replaceTestField(t, updated, "dap-jo9.2.5", "dependencies", json.RawMessage(`[]`))
	if _, _, err := ReconcileWithBinary(source, baseline, changedGraph, plan, approvedNullLossBDBinarySHA256); err == nil || !strings.Contains(err.Error(), "unplanned semantic delta") {
		t.Fatalf("null-loss exception accepted an unrelated dependency graph change: %v", err)
	}
}

func TestSecondStoreNullLossRequiresDerivedCandidateAndNullSpans(t *testing.T) {
	source, baseline, updated, plan := approvedNullLossFixture(t)
	candidate, _, err := ReconcileWithBinary(source, baseline, updated, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip := importerNullLossFixture(t, candidate)
	evidence, validation, err := ValidateCandidateRoundTrip(source, baseline, updated, candidate, roundTrip, plan, approvedNullLossBDBinarySHA256)
	if err != nil {
		t.Fatalf("valid second-store round trip rejected: %v", err)
	}
	if len(evidence.RepresentationLosses) != len(approvedNullLosses) || len(validation.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("loss evidence counts = %d/%d, want %d", len(evidence.RepresentationLosses), len(validation.RepresentationLosses), len(approvedNullLosses))
	}
	if len(validation.CandidateNullSpans) != len(approvedNullLosses) {
		t.Fatalf("verified candidate span count = %d, want %d", len(validation.CandidateNullSpans), len(approvedNullLosses))
	}
	for _, loss := range evidence.RepresentationLosses {
		if !loss.BaselineAbsent || !loss.UpdatedAbsent || !loss.RoundTripAbsent || loss.CanonicalValue != "null" || loss.CandidateValue != "null" {
			t.Fatalf("incomplete loss audit record: %+v", loss)
		}
	}
	changedCandidate := bytes.Replace(candidate, []byte("Scoped null-loss regression title"), []byte("Forged candidate content"), 1)
	if _, _, err := ValidateCandidateRoundTrip(source, baseline, updated, changedCandidate, roundTrip, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("second-store null-loss exception accepted a non-derived candidate")
	}
	if _, _, err := ValidateCandidateRoundTrip(source, baseline, updated, candidate, roundTrip, plan, strings.Repeat("0", 64)); err == nil {
		t.Fatal("second-store null-loss exception accepted an unapproved binary digest")
	}
	changedRoundTrip := replaceTestField(t, roundTrip, "dap-jo9.5.4", "title", json.RawMessage(`"unplanned second-store change"`))
	if _, _, err := ValidateCandidateRoundTrip(source, baseline, updated, candidate, changedRoundTrip, plan, approvedNullLossBDBinarySHA256); err == nil {
		t.Fatal("second-store null-loss exception accepted an unrelated round-trip change")
	}
}

func approvedNullLossFixture(t *testing.T) ([]byte, []byte, []byte, Plan) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "approved-null-loss-canonical.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if digest(source) != "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890" {
		t.Fatalf("historic approved source fixture digest changed: %s", digest(source))
	}
	baseline := make([]byte, 0, len(source))
	for _, line := range splitLines(source) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range approvedNullLosses {
			if pair.id == id {
				if raw, ok := fields[pair.field]; !ok || string(bytes.TrimSpace(raw)) != "null" {
					t.Fatalf("source %s.%s is not explicit null", id, pair.field)
				}
				delete(fields, pair.field)
			}
		}
		if id == "dap-jo9.5.5" {
			closedAt, err := derivedClosedAt(row{fields: fields})
			if err != nil {
				t.Fatal(err)
			}
			fields["closed_at"], _ = json.Marshal(closedAt)
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		baseline = append(baseline, encoded...)
		baseline = append(baseline, '\n')
	}
	updated := replaceTestField(t, baseline, "dap-jo9.5.3", "title", json.RawMessage(`"Scoped null-loss regression title"`))
	updated = replaceTestField(t, updated, "dap-jo9.5.3", "updated_at", json.RawMessage(`"2026-10-16T00:00:00Z"`))
	plan := Plan{BDVersion: "1.3.1", Updates: []IssueUpdate{{ID: "dap-jo9.5.3", Fields: map[string]string{"title": "Scoped null-loss regression title"}}}}
	return source, baseline, updated, plan
}

func importerNullLossFixture(t *testing.T, candidate []byte) []byte {
	t.Helper()
	var out []byte
	for _, line := range splitLines(candidate) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range approvedNullLosses {
			if pair.id == id {
				// Model only Beads' loss of source JSON nulls. A factual
				// non-null value (such as PR #114's closed_at) survives import.
				if raw, ok := fields[pair.field]; ok && string(bytes.TrimSpace(raw)) == "null" {
					delete(fields, pair.field)
				}
			}
		}
		if id == "dap-jo9.5.5" {
			closedAt, err := derivedClosedAt(row{fields: fields})
			if err != nil {
				t.Fatal(err)
			}
			fields["closed_at"], _ = json.Marshal(closedAt)
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, encoded...)
		out = append(out, '\n')
	}
	return out
}

func findTestRow(t *testing.T, data []byte, id string) map[string]json.RawMessage {
	t.Helper()
	for _, line := range splitLines(data) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		got, err := stringFieldMap(fields, "id")
		if err != nil {
			t.Fatal(err)
		}
		if got == id {
			return fields
		}
	}
	t.Fatalf("test issue %s not found", id)
	return nil
}

func replaceTestField(t *testing.T, data []byte, id, field string, value json.RawMessage) []byte {
	t.Helper()
	lines := splitLines(data)
	for index, line := range lines {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		got, err := stringFieldMap(fields, "id")
		if err != nil {
			t.Fatal(err)
		}
		if got == id {
			fields[field] = value
			lines[index], err = json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			return append(bytes.Join(lines, []byte("\n")), '\n')
		}
	}
	t.Fatalf("test issue %s not found", id)
	return nil
}
