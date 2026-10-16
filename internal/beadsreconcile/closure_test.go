package beadsreconcile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const closureSourceSHA256 = "5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465"

func closurePlan() Plan {
	return Plan{BDVersion: "1.3.1", Closes: []IssueClosure{{
		ID: "dap-jo9.18", ExpectedSourceSHA256: closureSourceSHA256,
		ExpectedStatus: "open", DecisionID: "0008",
		Reason:       approvedClosureReason,
		EvidenceRefs: []string{"#89", "#90", "#92", "#93", "#94", "#96"},
	}}}
}

func closureSource(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "approved-issue-creation-source.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestValidatePlanAcceptsOnlyScopedClosure(t *testing.T) {
	source := closureSource(t)
	if err := ValidatePlan(source, closurePlan()); err != nil {
		t.Fatalf("ValidatePlan() rejected exact authorized closure: %v", err)
	}
}

func TestValidatePlanRejectsSourceDriftForClosure(t *testing.T) {
	source := closureSource(t)
	drift := append([]byte(" "), source...)
	if err := ValidatePlan(drift, closurePlan()); err == nil {
		t.Fatal("ValidatePlan() accepted a plan pin for different source bytes")
	}
}

func TestValidatePlanRejectsClosureScopeAndReasonMutations(t *testing.T) {
	source := closureSource(t)
	tests := []struct {
		name string
		edit func(*Plan)
	}{
		{"wrong target", func(p *Plan) { p.Closes[0].ID = "dap-jo9.17" }},
		{"wrong source", func(p *Plan) { p.Closes[0].ExpectedSourceSHA256 = strings.Repeat("0", 64) }},
		{"wrong decision", func(p *Plan) { p.Closes[0].DecisionID = "0007" }},
		{"wrong source status", func(p *Plan) { p.Closes[0].ExpectedStatus = "closed" }},
		{"missing evidence", func(p *Plan) { p.Closes[0].EvidenceRefs = nil }},
		{"changed reason", func(p *Plan) { p.Closes[0].Reason += " altered" }},
		{"invalid UTF-8 reason", func(p *Plan) { p.Closes[0].Reason = string([]byte{0xff}) }},
		{"oversized reason", func(p *Plan) { p.Closes[0].Reason = strings.Repeat("x", 64*1024+1) }},
		{"duplicate close", func(p *Plan) { p.Closes = append(p.Closes, p.Closes[0]) }},
		{"mixed update", func(p *Plan) {
			p.Updates = []IssueUpdate{{ID: "dap-jo9.18", Fields: map[string]string{"title": "changed"}}}
		}},
		{"mixed create", func(p *Plan) { p.Creates = []json.RawMessage{json.RawMessage(`{}`)} }},
		{"wrong Beads version", func(p *Plan) { p.BDVersion = "1.3.2" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := closurePlan()
			tt.edit(&p)
			if err := ValidatePlan(source, p); err == nil {
				t.Fatal("ValidatePlan() accepted invalid closure plan")
			}
		})
	}
}

func TestReconcileBuildsExactScopedClosureCandidate(t *testing.T) {
	source := closureSource(t)
	baseline := importerNullLossFixture(t, source)
	updated := closureExport(t, baseline, closurePlan().Closes[0].Reason)
	candidate, evidence, err := ReconcileWithBinary(source, baseline, updated, closurePlan(), approvedClosureBDSHA256)
	if err != nil {
		t.Fatalf("ReconcileWithBinary() rejected real closure transition: %v", err)
	}
	if got := digest(candidate); got != evidence.CandidateSHA256 {
		t.Fatalf("candidate digest = %s, evidence says %s", got, evidence.CandidateSHA256)
	}
	if len(evidence.Closures) != 1 || evidence.Closures[0].Status != "closed" || evidence.Closures[0].CloseReason != closurePlan().Closes[0].Reason {
		t.Fatalf("closure evidence = %+v", evidence.Closures)
	}
	closed := findTestRow(t, candidate, "dap-jo9.18")
	for field, want := range map[string]string{"status": `"closed"`, "closed_at": `"2026-10-16T12:00:00Z"`, "close_reason": "\"" + closurePlan().Closes[0].Reason + "\"", "updated_at": `"2026-10-16T12:00:00Z"`} {
		if got := string(closed[field]); got != want {
			t.Errorf("candidate %s = %s, want %s", field, got, want)
		}
	}
	if len(evidence.RepresentationLosses) != len(approvedNullLosses) {
		t.Fatalf("historical null-loss count = %d, want %d", len(evidence.RepresentationLosses), len(approvedNullLosses))
	}
	for _, pair := range approvedNullLosses {
		if got := findTestRow(t, candidate, pair.id)[pair.field]; string(got) != "null" {
			t.Errorf("approved null %s.%s changed to %s", pair.id, pair.field, got)
		}
	}
	if _, _, err := ReconcileWithBinary(source, baseline, updated, closurePlan(), strings.Repeat("0", 64)); err == nil {
		t.Fatal("closure accepted an unpinned Beads binary")
	}
}

func TestReconcileRejectsClosureWithoutVerifiedBinaryPin(t *testing.T) {
	source := closureSource(t)
	baseline := importerNullLossFixture(t, source)
	for _, pair := range approvedNullLosses {
		baseline = replaceTestField(t, baseline, pair.id, pair.field, json.RawMessage("null"))
	}
	if got := findTestRow(t, baseline, "dap-jo9.5.5")["closed_at"]; len(got) == 0 {
		t.Fatal("baseline fixture must include the normalized historical dap-jo9.5.5 closed_at")
	}
	for _, pair := range approvedNullLosses {
		if got := findTestRow(t, baseline, pair.id)[pair.field]; string(got) != "null" {
			t.Fatalf("baseline fixture did not retain historical null %s.%s: %s", pair.id, pair.field, got)
		}
	}
	updated := closureExport(t, baseline, closurePlan().Closes[0].Reason)
	if _, _, err := Reconcile(source, baseline, updated, closurePlan()); err == nil {
		t.Fatal("Reconcile() accepted a closure without a verified Beads executable pin")
	}
}

func TestReconcileRejectsInvalidClosureTransitions(t *testing.T) {
	source := closureSource(t)
	baseline := importerNullLossFixture(t, source)
	valid := closureExport(t, baseline, closurePlan().Closes[0].Reason)
	tests := []struct {
		name string
		edit func([]byte) []byte
	}{
		{"wrong status", func(b []byte) []byte {
			return replaceTestField(t, b, "dap-jo9.18", "status", json.RawMessage(`"open"`))
		}},
		{"wrong reason", func(b []byte) []byte {
			return replaceTestField(t, b, "dap-jo9.18", "close_reason", json.RawMessage(`"other"`))
		}},
		{"non-UTC timestamp", func(b []byte) []byte {
			return replaceTestField(t, b, "dap-jo9.18", "closed_at", json.RawMessage(`"2026-10-16T12:00:00+01:00"`))
		}},
		{"missing close time", func(b []byte) []byte { return removeTestField(t, b, "dap-jo9.18", "closed_at") }},
		{"unplanned metadata", func(b []byte) []byte {
			return replaceTestField(t, b, "dap-jo9.18", "title", json.RawMessage(`"mutated"`))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := ReconcileWithBinary(source, baseline, tt.edit(valid), closurePlan(), approvedClosureBDSHA256); err == nil {
				t.Fatal("ReconcileWithBinary() accepted invalid close transition")
			}
		})
	}
}

func closureExport(t *testing.T, baseline []byte, reason string) []byte {
	t.Helper()
	updated := replaceTestField(t, baseline, "dap-jo9.18", "status", json.RawMessage(`"closed"`))
	updated = replaceTestField(t, updated, "dap-jo9.18", "closed_at", json.RawMessage(`"2026-10-16T12:00:00Z"`))
	encodedReason, err := json.Marshal(reason)
	if err != nil {
		t.Fatal(err)
	}
	updated = replaceTestField(t, updated, "dap-jo9.18", "close_reason", encodedReason)
	return replaceTestField(t, updated, "dap-jo9.18", "updated_at", json.RawMessage(`"2026-10-16T12:00:00Z"`))
}

func removeTestField(t *testing.T, data []byte, id, field string) []byte {
	t.Helper()
	var out []byte
	for _, line := range splitLines(data) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		got, _ := stringFieldMap(fields, "id")
		if got == id {
			delete(fields, field)
			line, _ = json.Marshal(fields)
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out
}
