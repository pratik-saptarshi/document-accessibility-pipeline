package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	commandWindows, err := afpCommandWindows(m.Commands, m.Evidence.AFPRoles)
	if err != nil {
		t.Fatalf("derive AFP command timestamp windows: %v", err)
	}
	evidence, validation, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], entries["candidate.jsonl"], entries["round-trip.jsonl"], plan, pinned, commandWindows)
	if err != nil {
		t.Fatalf("valid AFP candidate rejected: %v", err)
	}
	if !bytes.Equal(marshalJSON(t, evidence), marshalJSON(t, m.Evidence)) || !bytes.Equal(marshalJSON(t, validation), marshalJSON(t, m.Validation)) {
		t.Fatalf("bundle AFP evidence differs from rederived evidence: evidence fields %v; validation fields %v", jsonDifferenceKeys(t, evidence, m.Evidence), jsonDifferenceKeys(t, validation, m.Validation))
	}
	canonicalPath := filepath.Join(temp, "canonical.jsonl")
	if err := os.WriteFile(canonicalPath, source, 0600); err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", canonicalPath, "--bundle", bundlePath, "--bundle-sha256", sha(bundleBytes)}); err != nil {
		t.Fatalf("promote valid AFP candidate to disposable canonical copy: %v", err)
	}
	promoted, err := os.ReadFile(canonicalPath)
	if err != nil || !bytes.Equal(promoted, entries["candidate.jsonl"]) {
		t.Fatalf("promoted AFP bytes differ from reviewed candidate: read error=%v", err)
	}
	stalePath := filepath.Join(temp, "stale.jsonl")
	stale := append(append([]byte(nil), source...), ' ')
	if err := os.WriteFile(stalePath, stale, 0600); err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", stalePath, "--bundle", bundlePath, "--bundle-sha256", sha(bundleBytes)}); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("promote stale AFP source error = %v, want stale-source rejection", err)
	}
	afterStale, err := os.ReadFile(stalePath)
	if err != nil || !bytes.Equal(afterStale, stale) {
		t.Fatalf("stale AFP promotion changed canonical bytes: read error=%v", err)
	}
	tamperedEntries := make([]testBundleEntry, 0, len(entries))
	for name, data := range entries {
		if name == "candidate.jsonl" {
			data = bytes.Replace(data, []byte("dap-jo9.4.3"), []byte("dap-jo9.4.9"), 1)
		}
		tamperedEntries = append(tamperedEntries, testBundleEntry{name: name, data: data})
	}
	tamperedBundle := zipTestBytes(t, tamperedEntries)
	tamperedPath := filepath.Join(temp, "tampered.zip")
	if err := os.WriteFile(tamperedPath, tamperedBundle, 0600); err != nil {
		t.Fatal(err)
	}
	tamperedCanonical := filepath.Join(temp, "tampered-canonical.jsonl")
	if err := os.WriteFile(tamperedCanonical, source, 0600); err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", tamperedCanonical, "--bundle", tamperedPath, "--bundle-sha256", sha(tamperedBundle)}); err == nil {
		t.Fatal("AFP promotion accepted a recompressed bundle with a tampered candidate")
	}
	afterTamper, err := os.ReadFile(tamperedCanonical)
	if err != nil || !bytes.Equal(afterTamper, source) {
		t.Fatalf("tampered AFP promotion changed canonical bytes: read error=%v", err)
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
		{name: "existing edge created_by changed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", "dap-jo9.1.1", "created_by", json.RawMessage(`"unplanned-review-mutation"`), false)
		}},
		{name: "existing edge created_at changed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", "dap-jo9.1.1", "created_at", json.RawMessage(`"2026-10-18T00:00:00Z"`), false)
		}},
		{name: "existing edge metadata key added", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", "dap-jo9.1.1", "reviewer", json.RawMessage(`"unexpected"`), false)
		}},
		{name: "existing edge metadata key removed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", "dap-jo9.1.1", "created_by", nil, true)
		}},
		{name: "new edge created_by changed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", evidence.AFPRoles["profile"], "created_by", json.RawMessage(`"other"`), false)
		}},
		{name: "new edge created_at changed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", evidence.AFPRoles["profile"], "created_at", json.RawMessage(`"2026-10-18T00:00:00Z"`), false)
		}},
		{name: "new edge metadata key added", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", evidence.AFPRoles["profile"], "reviewer", json.RawMessage(`"unexpected"`), false)
		}},
		{name: "new edge metadata key removed", mutate: func(b []byte) []byte {
			return mutateDependency(t, b, "dap-jo9.4.1.2", evidence.AFPRoles["profile"], "metadata", nil, true)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], tc.mutate(entries["updated.jsonl"]), plan, pinned, commandWindows); err == nil {
				t.Fatal("AFP reconciler accepted an unplanned delta")
			}
		})
	}
	for _, edge := range []beadsreconcile.AFPEdge{
		{IssueID: evidence.AFPRoles["profile"], DependsOnID: "dap-jo9.4", Type: "parent-child"},
		{IssueID: evidence.AFPRoles["delivery"], DependsOnID: "dap-jo9.4.2", Type: "blocks"},
		{IssueID: "dap-jo9.4.1.2", DependsOnID: evidence.AFPRoles["profile"], Type: "blocks"},
	} {
		t.Run("edge timestamp before rounded command start "+edge.Type+" "+edge.IssueID, func(t *testing.T) {
			var window *beadsreconcile.AFPCommandWindow
			for i := range commandWindows {
				if commandWindows[i].Edge == edge {
					window = &commandWindows[i]
					break
				}
			}
			if window == nil {
				t.Fatalf("missing command window for %+v", edge)
			}
			before := window.StartedAt.UTC().Round(time.Second).Add(-time.Second).Format(time.RFC3339)
			mutated := mutateIssueAndDependencyTimestamp(t, entries["updated.jsonl"], edge, before)
			if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], mutated, plan, pinned, commandWindows); err == nil {
				t.Fatalf("AFP reconciler accepted created_at %s before command window for %+v", before, edge)
			}
		})
	}
	profileEdge := beadsreconcile.AFPEdge{IssueID: evidence.AFPRoles["profile"], DependsOnID: "dap-jo9.4", Type: "parent-child"}
	for i := range commandWindows {
		if commandWindows[i].Edge != profileEdge {
			continue
		}
		t.Run("created issue timestamp before rounded create start", func(t *testing.T) {
			before := commandWindows[i].StartedAt.UTC().Round(time.Second).Add(-time.Second).Format(time.RFC3339)
			mutated := mutateIssue(t, entries["updated.jsonl"], evidence.AFPRoles["profile"], func(fields map[string]json.RawMessage) {
				fields["created_at"], _ = json.Marshal(before)
			})
			if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], mutated, plan, pinned, commandWindows); err == nil {
				t.Fatalf("AFP reconciler accepted created issue timestamp %s before command window", before)
			}
		})
		break
	}
	for _, edge := range []beadsreconcile.AFPEdge{
		{IssueID: "dap-jo9.4.1.2", DependsOnID: evidence.AFPRoles["profile"], Type: "blocks"},
	} {
		for i := range commandWindows {
			if commandWindows[i].Edge != edge {
				continue
			}
			t.Run("fractional added edge timestamp", func(t *testing.T) {
				fractional := commandWindows[i].StartedAt.UTC().Round(time.Second).Add(500 * time.Millisecond).Format(time.RFC3339Nano)
				mutated := mutateDependency(t, entries["updated.jsonl"], edge.IssueID, edge.DependsOnID, "created_at", json.RawMessage(marshalJSON(t, fractional)), false)
				if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], mutated, plan, pinned, commandWindows); err == nil {
					t.Fatalf("AFP reconciler accepted fractional edge timestamp %s", fractional)
				}
			})
			break
		}
	}
	for i := range commandWindows {
		if commandWindows[i].Edge != profileEdge {
			continue
		}
		t.Run("fractional created issue timestamp", func(t *testing.T) {
			fractional := commandWindows[i].StartedAt.UTC().Round(time.Second).Add(500 * time.Millisecond).Format(time.RFC3339Nano)
			mutated := mutateIssue(t, entries["updated.jsonl"], evidence.AFPRoles["profile"], func(fields map[string]json.RawMessage) {
				fields["created_at"], _ = json.Marshal(fractional)
			})
			if _, _, err := beadsreconcile.ReconcileWithBinary(entries["source.jsonl"], entries["baseline.jsonl"], mutated, plan, pinned, commandWindows); err == nil {
				t.Fatalf("AFP reconciler accepted fractional issue timestamp %s", fractional)
			}
		})
		break
	}
	badCandidate := bytes.Replace(entries["candidate.jsonl"], []byte(`"dependencies":null`), []byte(`"dependencies":[]`), 1)
	if _, _, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], badCandidate, entries["round-trip.jsonl"], plan, pinned, commandWindows); err == nil {
		t.Fatal("AFP second-store validation accepted a candidate that erased the canonical null")
	}
	badRoundTrip := mutateIssue(t, entries["round-trip.jsonl"], "dap-jo9.4.2", func(f map[string]json.RawMessage) { f["dependency_count"] = json.RawMessage(`99`) })
	if _, _, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], entries["candidate.jsonl"], badRoundTrip, plan, pinned, commandWindows); err == nil {
		t.Fatal("AFP second-store validation accepted a changed graph counter")
	}
}

func mutateIssueAndDependencyTimestamp(t *testing.T, data []byte, edge beadsreconcile.AFPEdge, timestamp string) []byte {
	t.Helper()
	return mutateIssue(t, data, edge.IssueID, func(fields map[string]json.RawMessage) {
		fields["created_at"], _ = json.Marshal(timestamp)
		var dependencies []map[string]json.RawMessage
		if err := json.Unmarshal(fields["dependencies"], &dependencies); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, dependency := range dependencies {
			var id, target, typ string
			_ = json.Unmarshal(dependency["issue_id"], &id)
			_ = json.Unmarshal(dependency["depends_on_id"], &target)
			_ = json.Unmarshal(dependency["type"], &typ)
			if id == edge.IssueID && target == edge.DependsOnID && typ == edge.Type {
				dependency["created_at"], _ = json.Marshal(timestamp)
				found = true
			}
		}
		if !found {
			t.Fatalf("dependency edge not found: %+v", edge)
		}
		fields["dependencies"], _ = json.Marshal(dependencies)
	})
}

func mutateDependency(t *testing.T, data []byte, issueID, dependsOnID, field string, value json.RawMessage, remove bool) []byte {
	t.Helper()
	var found bool
	return mutateIssue(t, data, issueID, func(f map[string]json.RawMessage) {
		var dependencies []map[string]json.RawMessage
		if err := json.Unmarshal(f["dependencies"], &dependencies); err != nil {
			t.Fatal(err)
		}
		for _, dependency := range dependencies {
			var got string
			if err := json.Unmarshal(dependency["depends_on_id"], &got); err != nil {
				t.Fatal(err)
			}
			if got != dependsOnID {
				continue
			}
			if remove {
				delete(dependency, field)
			} else {
				dependency[field] = value
			}
			found = true
		}
		if !found {
			t.Fatalf("dependency %s -> %s is missing", issueID, dependsOnID)
		}
		encoded, err := json.Marshal(dependencies)
		if err != nil {
			t.Fatal(err)
		}
		f["dependencies"] = encoded
	})
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
