package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestAFPRefinementPinnedBeadsBundleAndAdversarialDeltas(t *testing.T) {
	bdPath := os.Getenv("DAP_BD_BINARY")
	if bdPath == "" {
		t.Skip("set DAP_BD_BINARY to the exact Beads 1.3.1 executable for AFP batch integration")
	}
	bdBytes, _, bdAbs, err := readRegularBounded(bdPath, maxBDBinaryBytes)
	if err != nil {
		t.Fatal(err)
	}
	const pinned = "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653"
	if got := sha(bdBytes); got != pinned {
		t.Fatalf("Beads executable SHA-256 = %s, want %s", got, pinned)
	}
	sourcePath := filepath.Join("..", "..", "internal", "beadsreconcile", "testdata", "approved-afp-refinement-source.jsonl")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	const sourceSHA = "dd048534e6e007841e0750d5370a25398178f2b0c70750078eabe1f4b1508409"
	if got := sha(source); got != sourceSHA {
		t.Fatalf("AFP fixture SHA-256 = %s, want %s", got, sourceSHA)
	}
	temp := t.TempDir()
	planPath := filepath.Join(temp, "plan.json")
	bundlePath := filepath.Join(temp, "bundle.zip")
	if err := os.WriteFile(planPath, []byte(`{"bd_version":"1.3.1","afp_refinement":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := create([]string{"--canonical", sourcePath, "--plan", planPath, "--bd", bdAbs, "--output", bundlePath}); err != nil {
		t.Fatalf("create AFP batch bundle: %v", err)
	}
	bundleBytes, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := readBundle(bundleBytes)
	if err != nil {
		t.Fatal(err)
	}
	var plan beadsreconcile.Plan
	if err := strictJSON(entries["plan.json"], &plan); err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := strictJSON(entries["manifest.json"], &m); err != nil {
		t.Fatal(err)
	}
	evidence, validation, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], entries["candidate.jsonl"], entries["round-trip.jsonl"], plan, pinned)
	if err != nil {
		t.Fatalf("valid AFP candidate rejected: %v", err)
	}
	if !bytes.Equal(marshalJSON(t, evidence), marshalJSON(t, m.Evidence)) || !bytes.Equal(marshalJSON(t, validation), marshalJSON(t, m.Validation)) {
		t.Fatalf("bundle AFP evidence differs from rederived evidence: evidence fields %v; validation fields %v", jsonDifferenceKeys(t, evidence, m.Evidence), jsonDifferenceKeys(t, validation, m.Validation))
	}
	if len(evidence.CreatedIDs) != 3 || len(evidence.AFPAddedEdges) != 8 || len(evidence.AFPCounterChanges) != 4 {
		t.Fatalf("AFP graph evidence counts = IDs %d edges %d counters %d", len(evidence.CreatedIDs), len(evidence.AFPAddedEdges), len(evidence.AFPCounterChanges))
	}
	if len(evidence.RepresentationLosses) != 1 || !evidence.RepresentationLosses[0].RoundTripAbsent {
		t.Fatalf("explicit-null round-trip evidence = %+v", evidence.RepresentationLosses)
	}
	if len(evidence.CandidateNullSpans) != 1 || evidence.CandidateNullSpans[0].IssueID != "dap-jo9.5.3" || evidence.CandidateNullSpans[0].Raw != "null" {
		t.Fatalf("candidate null span = %+v", evidence.CandidateNullSpans)
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{name: "unplanned feasibility acceptance edit", mutate: func(b []byte) []byte {
			return mutateIssue(t, b, "dap-jo9.4.1.2", func(f map[string]json.RawMessage) { f["acceptance_criteria"] = json.RawMessage(`"unplanned"`) })
		}},
		{name: "unexpected existing priority", mutate: func(b []byte) []byte {
			return mutateIssue(t, b, "dap-jo9.4.1", func(f map[string]json.RawMessage) { f["priority"] = json.RawMessage(`1`) })
		}},
		{name: "unplanned existing graph count", mutate: func(b []byte) []byte {
			return mutateIssue(t, b, "dap-jo9.4.2", func(f map[string]json.RawMessage) { f["dependency_count"] = json.RawMessage(`4`) })
		}},
		{name: "extra dependency edge", mutate: func(b []byte) []byte {
			return mutateIssue(t, b, "dap-jo9.4.2", func(f map[string]json.RawMessage) {
				var d []json.RawMessage
				_ = json.Unmarshal(f["dependencies"], &d)
				d = append(d, json.RawMessage(`{"issue_id":"dap-jo9.4.2","depends_on_id":"dap-jo9.1","type":"blocks","created_at":"2026-10-19T00:00:00Z","created_by":"fixture","metadata":"{}"}`))
				f["dependencies"], _ = json.Marshal(d)
			})
		}},
		{name: "unplanned created field", mutate: func(b []byte) []byte {
			return mutateIssue(t, b, evidence.AFPRoles["profile"], func(f map[string]json.RawMessage) { f["owner"] = json.RawMessage(`"unexpected"`) })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], tc.mutate(entries["updated.jsonl"]), plan, pinned); err == nil {
				t.Fatal("AFP reconciler accepted an unplanned delta")
			}
		})
	}
	badCandidate := bytes.Replace(entries["candidate.jsonl"], []byte(`"dependencies":null`), []byte(`"dependencies":[]`), 1)
	if _, _, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], badCandidate, entries["round-trip.jsonl"], plan, pinned); err == nil {
		t.Fatal("AFP second-store validation accepted a candidate that erased the canonical null")
	}
	badRoundTrip := mutateIssue(t, entries["round-trip.jsonl"], "dap-jo9.4.2", func(f map[string]json.RawMessage) { f["dependency_count"] = json.RawMessage(`99`) })
	if _, _, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], entries["candidate.jsonl"], badRoundTrip, plan, pinned); err == nil {
		t.Fatal("AFP second-store validation accepted a changed graph counter")
	}
}

func marshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func jsonDifferenceKeys(t *testing.T, a, b any) []string {
	t.Helper()
	decode := func(v any) map[string]json.RawMessage {
		var out map[string]json.RawMessage
		if err := json.Unmarshal(marshalJSON(t, v), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	aa, bb := decode(a), decode(b)
	var keys []string
	for k, av := range aa {
		if bv, ok := bb[k]; !ok || !bytes.Equal(av, bv) {
			keys = append(keys, k)
		}
	}
	for k := range bb {
		if _, ok := aa[k]; !ok {
			keys = append(keys, k)
		}
	}
	return keys
}

func mutateIssue(t *testing.T, data []byte, id string, mutate func(map[string]json.RawMessage)) []byte {
	t.Helper()
	lines := bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n"))
	found := false
	for i, line := range lines {
		var f map[string]json.RawMessage
		if err := json.Unmarshal(line, &f); err != nil {
			t.Fatal(err)
		}
		var got string
		if err := json.Unmarshal(f["id"], &got); err != nil {
			t.Fatal(err)
		}
		if got != id {
			continue
		}
		mutate(f)
		updated, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		lines[i] = updated
		found = true
		break
	}
	if !found {
		t.Fatalf("issue %s missing from mutation input", id)
	}
	return append(bytes.Join(lines, []byte("\n")), '\n')
}
