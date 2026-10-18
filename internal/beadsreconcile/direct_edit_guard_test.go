package beadsreconcile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const approvedD2CorrectionReceiptCanonicalSHA256 = "ecdfc1b38860cb50b17b1a83714c463e3f043137430d4644980c1e9e72fc92ef"

// TestDirectEditGuardRejectsInvalidJSON feeds malformed JSONL and expects
// fail-closed rejection.
func TestDirectEditGuardRejectsInvalidJSON(t *testing.T) {
	bad := []byte("{\"id\": \"x\"}\nnot-json\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted invalid JSONL, want error")
	}
}

// TestDirectEditGuardRejectsMissingID feeds a record without the required id
// field and expects rejection.
func TestDirectEditGuardRejectsMissingID(t *testing.T) {
	bad := []byte("{\"status\": \"open\"}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted record without id, want error")
	}
}

// TestDirectEditGuardRejectsDuplicateIDs feeds two records with the same id
// and expects rejection.
func TestDirectEditGuardRejectsDuplicateIDs(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"status\": \"open\"}\n{\"id\": \"a\", \"status\": \"closed\"}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted duplicate ids, want error")
	}
}

// TestDirectEditGuardRejectsNullLoss feeds a record where a field that must
// not be null is null, and expects rejection.
func TestDirectEditGuardRejectsNullLoss(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": null}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted null title, want error")
	}
}

// TestDirectEditGuardRejectsBadDependencyCount feeds a record whose
// dependency_count disagrees with its dependencies list length.
func TestDirectEditGuardRejectsBadDependencyCount(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\", \"dependencies\": [{\"issue_id\": \"a\", \"depends_on_id\": \"b\", \"type\": \"blocks\"}], \"dependency_count\": 5}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted inconsistent dependency_count, want error")
	}
}

// TestDirectEditGuardRejectsStaleDependencyCount feeds a record with one
// blocks edge but dependency_count 0: the stale count must be rejected.
// (Codex P2: the old guard only rejected inflated counts.)
func TestDirectEditGuardRejectsStaleDependencyCount(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\", \"dependencies\": [{\"issue_id\": \"a\", \"depends_on_id\": \"b\", \"type\": \"blocks\"}], \"dependency_count\": 0}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted stale dependency_count 0 with one blocks edge, want error")
	}
}

// TestDirectEditGuardRejectsNegativeDependencyCount feeds a negative count,
// which must be rejected (counts are exact, never negative).
func TestDirectEditGuardRejectsNegativeDependencyCount(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\", \"dependency_count\": -1}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted negative dependency_count, want error")
	}
}

func TestDirectEditGuardRejectsNullDependencyCount(t *testing.T) {
	bad := []byte("{\"id\":\"a\",\"title\":\"t\",\"dependency_count\":null}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted null dependency_count, want error")
	}
}

// TestDirectEditGuardAcceptsMixedDependencyTypes ensures only blocks edges
// count: one blocks + one parent-child edge with count 1 must pass.
func TestDirectEditGuardAcceptsMixedDependencyTypes(t *testing.T) {
	good := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\", \"dependencies\": [{\"issue_id\": \"a\", \"depends_on_id\": \"b\", \"type\": \"blocks\"}, {\"issue_id\": \"a\", \"depends_on_id\": \"c\", \"type\": \"parent-child\"}], \"dependency_count\": 1}\n")
	if err := ValidateDirectEdit(good); err != nil {
		t.Fatalf("ValidateDirectEdit rejected exact blocks-edge count: %v", err)
	}
}

// TestDirectEditGuardRejectsDuplicateKeys feeds a record repeating a key.
// encoding/json would silently keep the last value; the guard must fail.
func TestDirectEditGuardRejectsDuplicateKeys(t *testing.T) {
	bad := []byte("{\"id\": \"expected\", \"id\": \"other\"}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted duplicate JSON keys, want error")
	}
}

// TestDirectEditGuardRejectsDuplicateKeyHidingNullTitle feeds a record where
// a duplicate key masks a null title from the null-loss check.
func TestDirectEditGuardRejectsDuplicateKeyHidingNullTitle(t *testing.T) {
	bad := []byte("{\"id\": \"a\", \"title\": null, \"title\": \"t\"}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted duplicate key masking null title, want error")
	}
}

// TestDirectEditGuardAcceptsLiveFile validates the actual .beads/issues.jsonl
// on disk. This is the CI gate for the direct-edit path.
func TestDirectEditGuardAcceptsLiveFile(t *testing.T) {
	// Walk up from the package dir to the repo root.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var live []byte
	for i := 0; i < 6; i++ {
		p := filepath.Join(dir, ".beads", "issues.jsonl")
		if b, err := os.ReadFile(p); err == nil {
			live = b
			break
		}
		dir = filepath.Dir(dir)
	}
	if live == nil {
		t.Skip("live .beads/issues.jsonl not found from test working dir")
	}
	repoRoot := dir
	// Compare the candidate against the base revision's trusted export. The
	// workflow supplies the PR base SHA or push-before SHA; local runs default
	// to HEAD^ and can override it with DAP_BEADS_SOURCE_REF.
	sourceRef := os.Getenv("DAP_BEADS_SOURCE_REF")
	if sourceRef == "" {
		sourceRef = "HEAD^"
	}
	cmd := exec.Command("git", "show", sourceRef+":.beads/issues.jsonl")
	source, err := cmd.Output()
	if err != nil {
		t.Fatalf("read trusted .beads source at %q: %v", sourceRef, err)
	}
	trustedD2Receipt, err := trustedHistoryContainsD2CorrectionReceipt(t, repoRoot, sourceRef)
	if err != nil {
		t.Fatalf("inspect trusted .beads history at %q: %v", sourceRef, err)
	}
	if err := validateLiveExportForTrustedSource(source, live, trustedD2Receipt); err != nil {
		t.Fatalf("live .beads/issues.jsonl failed source-to-candidate guard: %v", err)
	}
}

func TestD2ReceiptIsRecognizedOnlyInTrustedSource(t *testing.T) {
	uncorrected := d2SourceRow(t)
	corrected := withD2Reopen(t, uncorrected)
	if hasExactD2CorrectionReceipt(uncorrected) {
		t.Fatal("unreopened trusted row was treated as containing the D2 correction receipt")
	}
	if !hasExactD2CorrectionReceipt(corrected) {
		t.Fatal("exact corrected row was not recognized as the D2 correction receipt")
	}
	mutatedReceipt := setDirectEditField(t, corrected, approvedD2IssueID, "notes", []byte(`"different note"`))
	if hasExactD2CorrectionReceipt(mutatedReceipt) {
		t.Fatal("mutated correction receipt was treated as the exact trusted receipt")
	}
}

func TestLiveExportGuardRetiresD2TransitionOnlyAfterTrustedReceipt(t *testing.T) {
	uncorrected := append(d2SourceRow(t), canonicalIssueForTest(t, "dap-jo9.1.6")...)
	corrected := withD2Reopen(t, uncorrected)

	// A correction present only in the candidate does not retire the exact
	// source-to-candidate transition check.
	candidateOnlyReceipt := append(append([]byte(nil), corrected...), []byte(`{"id":"later-update","title":"valid later issue"}`+"\n")...)
	if err := validateLiveExportForTrustedSource(uncorrected, candidateOnlyReceipt, false); err == nil {
		t.Fatal("candidate-controlled D2 receipt disabled the one-off transition guard")
	}

	// The first later tracker update edits D2 itself. The exact reopen row is
	// still present in trusted history, even though the current source row has
	// legitimately changed.
	editedD2 := setDirectEditField(t, corrected, approvedD2IssueID, "title", []byte(`"accepted D2 follow-up"`))
	if err := validateLiveExportForTrustedSource(corrected, editedD2, true); err != nil {
		t.Fatalf("valid D2 follow-up was rejected after trusted receipt: %v", err)
	}

	// The next export is unrelated. The durable trusted-history receipt keeps
	// retiring the one-off transition check after the D2 row itself changes.
	laterExport := append(append([]byte(nil), editedD2...), canonicalIssueForTest(t, "later-update")...)
	if err := validateLiveExportForTrustedSource(editedD2, laterExport, true); err != nil {
		t.Fatalf("valid unrelated tracker update was rejected after D2 follow-up: %v", err)
	}
	invalidLaterExport := append(append([]byte(nil), corrected...), []byte("not-json\n")...)
	if err := validateLiveExportForTrustedSource(corrected, invalidLaterExport, true); err == nil {
		t.Fatal("general export validator accepted malformed JSONL after D2 receipt")
	}
}

func TestCanonicalExportRequiresMandatoryFieldsAndTypes(t *testing.T) {
	valid := canonicalIssueForTest(t, "dap-example")
	if err := validateCanonicalExport(valid); err != nil {
		t.Fatalf("valid canonical export rejected: %v", err)
	}
	if err := validateCanonicalExport([]byte(`{"id":"dap-example"}` + "\n")); err == nil {
		t.Fatal("ID-only issue record passed canonical export validation")
	}

	fields := map[string]any{
		"id":                  "dap-example",
		"title":               "Example title",
		"description":         "Example description",
		"acceptance_criteria": "Example criteria",
		"status":              "open",
		"priority":            2,
		"issue_type":          "task",
		"created_at":          "2026-10-18T00:00:00Z",
		"updated_at":          "2026-10-18T00:00:00Z",
		"dependency_count":    0,
		"dependent_count":     0,
	}
	for _, field := range []string{"title", "description", "acceptance_criteria", "status", "issue_type", "created_at", "updated_at", "priority", "dependency_count", "dependent_count"} {
		t.Run("missing_"+field, func(t *testing.T) {
			candidate := cloneCanonicalIssueFields(fields)
			delete(candidate, field)
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateCanonicalExport(append(data, '\n')); err == nil {
				t.Fatalf("canonical export accepted missing mandatory field %q", field)
			}
		})
	}
	for field, wrong := range map[string]any{
		"title":               3,
		"description":         false,
		"acceptance_criteria": []string{"wrong"},
		"status":              2,
		"issue_type":          false,
		"created_at":          123,
		"updated_at":          nil,
		"priority":            "2",
		"dependency_count":    0.5,
		"dependent_count":     "0",
	} {
		t.Run("wrong_type_"+field, func(t *testing.T) {
			candidate := cloneCanonicalIssueFields(fields)
			candidate[field] = wrong
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateCanonicalExport(append(data, '\n')); err == nil {
				t.Fatalf("canonical export accepted wrong type for mandatory field %q", field)
			}
		})
	}
	for _, field := range []string{"title", "description", "acceptance_criteria", "status", "issue_type", "created_at", "updated_at", "priority", "dependency_count", "dependent_count"} {
		t.Run("null_"+field, func(t *testing.T) {
			candidate := cloneCanonicalIssueFields(fields)
			candidate[field] = nil
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateCanonicalExport(append(data, '\n')); err == nil {
				t.Fatalf("canonical export accepted null mandatory field %q", field)
			}
		})
	}
	for _, priority := range []int{-1, 5, 99} {
		t.Run(fmt.Sprintf("priority_out_of_range_%d", priority), func(t *testing.T) {
			candidate := cloneCanonicalIssueFields(fields)
			candidate["priority"] = priority
			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateCanonicalExport(append(data, '\n')); err == nil {
				t.Fatalf("canonical export accepted priority %d outside the Beads 0-4 range", priority)
			}
		})
	}
}

func TestCanonicalExportPostReceiptValidatesDependencyGraphAndIncomingCounts(t *testing.T) {
	source := canonicalGraphExportForTest(t)
	if err := validateLiveExportForTrustedSource(source, source, true); err != nil {
		t.Fatalf("valid historical-style graph rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*testing.T, []byte) []byte
	}{
		{"malformed edge", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[null]`))
		}},
		{"missing edge type", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"dap-b"}]`))
		}},
		{"empty edge source", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"","depends_on_id":"dap-b","type":"blocks"}]`))
		}},
		{"empty edge target", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"","type":"blocks"}]`))
		}},
		{"empty edge type", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"dap-b","type":" "}]`))
		}},
		{"edge ownership", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-d","depends_on_id":"dap-b","type":"blocks"}]`))
		}},
		{"missing target", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"dap-missing","type":"blocks"}]`))
		}},
		{"duplicate edge", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-a", "dependencies", json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"dap-b","type":"blocks"},{"issue_id":"dap-a","depends_on_id":"dap-b","type":"blocks"}]`))
		}},
		{"stale high incoming count", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-b", "dependent_count", json.RawMessage(`2`))
		}},
		{"stale low incoming count", func(t *testing.T, b []byte) []byte {
			return setDirectEditField(t, b, "dap-b", "dependent_count", json.RawMessage(`0`))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := tt.mutate(t, append([]byte(nil), source...))
			if err := validateLiveExportForTrustedSource(source, candidate, true); err == nil {
				t.Fatalf("post-receipt validation accepted %s", tt.name)
			}
		})
	}
}

func TestCanonicalExportIncomingCountsIgnoreNonBlockingEdgesAndOrdering(t *testing.T) {
	data := canonicalGraphExportForTest(t)
	if err := validateCanonicalExport(data); err != nil {
		t.Fatalf("one incoming block plus parent-child/tracks edges should count as one: %v", err)
	}
	// Reverse row order and edge order. Counts must depend on graph contents,
	// not the serialization order.
	reordered := canonicalReorderedGraphExportForTest(t, data)
	if err := validateCanonicalExport(reordered); err != nil {
		t.Fatalf("reordered graph rejected: %v", err)
	}

	// Two incoming blocks edges count as two. Parent-child and tracks edges to
	// the same target remain excluded.
	multiple := append(data, canonicalIssueWithDependenciesForTest(t, "dap-e", 1, 0, []map[string]any{
		{"issue_id": "dap-e", "depends_on_id": "dap-b", "type": "blocks"},
	})...)
	multiple = setDirectEditField(t, multiple, "dap-b", "dependent_count", json.RawMessage(`2`))
	if err := validateCanonicalExport(multiple); err != nil {
		t.Fatalf("two incoming blocks edges should count as two: %v", err)
	}
	zero := canonicalIssueWithDependenciesForTest(t, "dap-x", 0, 0, []map[string]any{
		{"issue_id": "dap-x", "depends_on_id": "dap-y", "type": "parent-child"},
		{"issue_id": "dap-x", "depends_on_id": "dap-y", "type": "locally-supported-edge-type"},
	})
	zero = append(zero, canonicalIssueForTest(t, "dap-y")...)
	if err := validateCanonicalExport(zero); err != nil {
		t.Fatalf("parent-child edge should not count as a blocker: %v", err)
	}
}

func TestCanonicalExportAcceptsOrdinaryPostReceiptUpdateAndAddition(t *testing.T) {
	source := canonicalGraphExportForTest(t)
	candidate := setDirectEditField(t, source, "dap-a", "title", json.RawMessage(`"updated title"`))
	candidate = append(candidate, canonicalIssueWithDependenciesForTest(t, "dap-new", 1, 0, []map[string]any{
		{"issue_id": "dap-new", "depends_on_id": "dap-b", "type": "blocks"},
	})...)
	candidate = setDirectEditField(t, candidate, "dap-b", "dependent_count", json.RawMessage(`2`))
	if err := validateLiveExportForTrustedSource(source, candidate, true); err != nil {
		t.Fatalf("valid ordinary update and addition rejected: %v", err)
	}
}

func TestCanonicalExportChecksSupportedOptionalFieldTypes(t *testing.T) {
	wrongTypes := map[string]any{
		"_type":            false,
		"assignee":         3,
		"close_reason":     false,
		"closed_at":        42,
		"comment_count":    "0",
		"created_by":       4,
		"dependencies":     []any{nil},
		"heartbeat_at":     false,
		"labels":           []any{"ok", nil},
		"lease_expires_at": 4,
		"notes":            []string{"wrong"},
		"owner":            false,
		"started_at":       42,
	}
	for field, wrong := range wrongTypes {
		t.Run("wrong_type_"+field, func(t *testing.T) {
			candidate := setDirectEditField(t, canonicalIssueForTest(t, "dap-example"), "dap-example", field, mustRawJSONForTest(t, wrong))
			if err := validateLiveExportForTrustedSource(canonicalIssueForTest(t, "dap-example"), candidate, true); err == nil {
				t.Fatalf("post-receipt validator accepted wrong type for optional field %q", field)
			}
		})
	}

	for _, field := range []string{"closed_at", "dependencies"} {
		t.Run("documented_null_"+field, func(t *testing.T) {
			candidate := setDirectEditField(t, canonicalIssueForTest(t, "dap-example"), "dap-example", field, json.RawMessage(`null`))
			if err := validateCanonicalExport(candidate); err != nil {
				t.Fatalf("documented null for %s rejected: %v", field, err)
			}
		})
	}
	for _, field := range []string{"_type", "assignee", "close_reason", "comment_count", "created_by", "heartbeat_at", "labels", "lease_expires_at", "notes", "owner", "started_at"} {
		t.Run("unsupported_null_"+field, func(t *testing.T) {
			candidate := setDirectEditField(t, canonicalIssueForTest(t, "dap-example"), "dap-example", field, json.RawMessage(`null`))
			if err := validateCanonicalExport(candidate); err == nil {
				t.Fatalf("unsupported null for optional field %s accepted", field)
			}
		})
	}
	for _, field := range []string{"_type", "assignee", "close_reason", "closed_at", "comment_count", "created_by", "dependencies", "heartbeat_at", "labels", "lease_expires_at", "notes", "owner", "started_at"} {
		t.Run("omitted_"+field, func(t *testing.T) {
			candidate := removeDirectEditField(t, canonicalIssueForTest(t, "dap-example"), "dap-example", field)
			if err := validateCanonicalExport(candidate); err != nil {
				t.Fatalf("omitted optional field %s rejected: %v", field, err)
			}
		})
	}
}

func TestCanonicalExportAcceptsCurrentAndApprovedHistoricalExports(t *testing.T) {
	paths := []string{
		"../../.beads/issues.jsonl",
		"testdata/approved-current-source-canonical.jsonl",
		"testdata/approved-latest-null-loss-source.jsonl",
		"testdata/approved-null-loss-canonical.jsonl",
		"testdata/approved-issue-creation-source.jsonl",
		"testdata/approved-pr112-corrected-source.jsonl",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateCanonicalExport(data); err != nil {
				t.Fatalf("approved export rejected: %v", err)
			}
		})
	}
}

func canonicalGraphExportForTest(t *testing.T) []byte {
	t.Helper()
	data := canonicalIssueWithDependenciesForTest(t, "dap-a", 1, 0, []map[string]any{
		{"issue_id": "dap-a", "depends_on_id": "dap-b", "type": "blocks"},
		{"issue_id": "dap-a", "depends_on_id": "dap-c", "type": "tracks"},
	})
	data = append(data, canonicalIssueWithDependenciesForTest(t, "dap-d", 0, 0, []map[string]any{
		{"issue_id": "dap-d", "depends_on_id": "dap-b", "type": "parent-child"},
		{"issue_id": "dap-d", "depends_on_id": "dap-c", "type": "tracks"},
	})...)
	data = append(data, canonicalIssueForTest(t, "dap-b")...)
	data = setDirectEditField(t, data, "dap-b", "dependent_count", json.RawMessage(`1`))
	data = append(data, canonicalIssueWithDependenciesForTest(t, "dap-c", 0, 0, nil)...)
	data = setDirectEditField(t, data, "dap-c", "dependencies", json.RawMessage(`null`))
	return data
}

func canonicalIssueWithDependenciesForTest(t *testing.T, id string, dependencyCount, dependentCount int, edges []map[string]any) []byte {
	t.Helper()
	fields := directEditFieldsForTest(t, canonicalIssueForTest(t, id), id)
	fields["dependency_count"] = mustRawJSONForTest(t, dependencyCount)
	fields["dependent_count"] = mustRawJSONForTest(t, dependentCount)
	if edges != nil {
		fields["dependencies"] = mustRawJSONForTest(t, edges)
	}
	line, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return append(line, '\n')
}

func canonicalReorderedGraphExportForTest(t *testing.T, data []byte) []byte {
	t.Helper()
	rows := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	var out []byte
	for i := len(rows) - 1; i >= 0; i-- {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rows[i], &fields); err != nil {
			t.Fatal(err)
		}
		if fields["id"] != nil && string(fields["id"]) == `"dap-a"` {
			fields["dependencies"] = json.RawMessage(`[{"issue_id":"dap-a","depends_on_id":"dap-c","type":"tracks"},{"issue_id":"dap-a","depends_on_id":"dap-b","type":"blocks"}]`)
		}
		line, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out
}

func mustRawJSONForTest(t *testing.T, value any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func TestLiveExportGuardRejectsStrippedCanonicalRecordAfterTrustedReceipt(t *testing.T) {
	source := canonicalIssueForTest(t, "dap-example")
	stripped := []byte(`{"id":"dap-example"}` + "\n")
	if err := validateLiveExportForTrustedSource(source, stripped, true); err == nil {
		t.Fatal("trusted-history path accepted a tracker export stripped to IDs")
	}
	malformed := []byte(`{"id":"dap-example","title":3}` + "\n")
	if err := validateLiveExportForTrustedSource(source, malformed, true); err == nil {
		t.Fatal("trusted-history path accepted a tracker record with malformed canonical field types")
	}
}

func TestLiveExportGuardPreservesSourceIssueIDsAfterTrustedReceipt(t *testing.T) {
	source := append(canonicalIssueForTest(t, "dap-example"), canonicalIssueForTest(t, "dap-other")...)
	candidateMissingOne := canonicalIssueForTest(t, "dap-example")
	if err := validateLiveExportForTrustedSource(source, candidateMissingOne, true); err == nil {
		t.Fatal("trusted-history path accepted a canonical export that dropped a source issue")
	}

	candidateWithAddition := append(append([]byte(nil), source...), canonicalIssueForTest(t, "dap-new")...)
	if err := validateLiveExportForTrustedSource(source, candidateWithAddition, true); err != nil {
		t.Fatalf("trusted-history path rejected an export that preserved source issues and added one: %v", err)
	}
}

func canonicalIssueForTest(t *testing.T, id string) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"id":                  id,
		"title":               "Example title",
		"description":         "Example description",
		"acceptance_criteria": "Example criteria",
		"status":              "open",
		"priority":            2,
		"issue_type":          "task",
		"created_at":          "2026-10-18T00:00:00Z",
		"updated_at":          "2026-10-18T00:00:00Z",
		"dependency_count":    0,
		"dependent_count":     0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func cloneCanonicalIssueFields(fields map[string]any) map[string]any {
	clone := make(map[string]any, len(fields))
	for key, value := range fields {
		clone[key] = value
	}
	return clone
}

func TestTrustedHistoryReceiptSurvivesTwoSuccessiveTrackerExports(t *testing.T) {
	repo := t.TempDir()
	gitForTest(t, repo, "init", "-q")
	gitForTest(t, repo, "config", "user.name", "Test Author")
	gitForTest(t, repo, "config", "user.email", "test@example.invalid")
	if err := os.MkdirAll(filepath.Join(repo, ".beads"), 0o755); err != nil {
		t.Fatal(err)
	}

	uncorrected := append(d2SourceRow(t), canonicalIssueForTest(t, "dap-jo9.1.6")...)
	writeTrackerAndCommit(t, repo, uncorrected, "base tracker")
	corrected := withD2Reopen(t, uncorrected)
	if err := validateLiveExportForTrustedSource(uncorrected, corrected, false); err != nil {
		t.Fatalf("initial exact D2 correction rejected before receipt entered trusted history: %v", err)
	}
	writeTrackerAndCommit(t, repo, corrected, "authorized D2 correction")

	trustedReceipt, err := trustedHistoryContainsD2CorrectionReceipt(t, repo, "HEAD")
	if err != nil || !trustedReceipt {
		t.Fatalf("trusted history did not retain correction receipt: found=%t err=%v", trustedReceipt, err)
	}
	editedD2 := setDirectEditField(t, corrected, approvedD2IssueID, "title", []byte(`"accepted D2 follow-up"`))
	if err := validateLiveExportForTrustedSource(corrected, editedD2, trustedReceipt); err != nil {
		t.Fatalf("first legitimate export editing D2 was rejected: %v", err)
	}
	writeTrackerAndCommit(t, repo, editedD2, "later D2 tracker update")

	trustedReceipt, err = trustedHistoryContainsD2CorrectionReceipt(t, repo, "HEAD")
	if err != nil || !trustedReceipt {
		t.Fatalf("trusted history lost correction receipt after D2 edit: found=%t err=%v", trustedReceipt, err)
	}
	unrelated := append(append([]byte(nil), editedD2...), canonicalIssueForTest(t, "later-update")...)
	if err := validateLiveExportForTrustedSource(editedD2, unrelated, trustedReceipt); err != nil {
		t.Fatalf("second legitimate unrelated export was rejected: %v", err)
	}
}

func validateLiveExportForTrustedSource(source, candidate []byte, trustedHistoryHasD2Receipt bool) error {
	if err := ValidateDirectEdit(source); err != nil {
		return err
	}
	if err := ValidateDirectEdit(candidate); err != nil {
		return err
	}
	if trustedHistoryHasD2Receipt {
		if err := validateCanonicalExport(source); err != nil {
			return fmt.Errorf("trusted source export is not canonical: %w", err)
		}
		if err := validateCanonicalExport(candidate); err != nil {
			return err
		}
		return validateSourceIssueIDsPreserved(source, candidate)
	}
	return ValidateDirectEditTransition(source, candidate)
}

// validateSourceIssueIDsPreserved allows ordinary post-receipt tracker edits
// and additions while preventing an export from silently dropping existing
// source records.
func validateSourceIssueIDsPreserved(source, candidate []byte) error {
	sourceRows, err := parseDirectEditRows(source)
	if err != nil {
		return fmt.Errorf("parse trusted source issue IDs: %w", err)
	}
	candidateRows, err := parseDirectEditRows(candidate)
	if err != nil {
		return fmt.Errorf("parse candidate issue IDs: %w", err)
	}
	for id := range sourceRows {
		if _, ok := candidateRows[id]; !ok {
			return fmt.Errorf("candidate export removed existing issue %q", id)
		}
	}
	return nil
}

func trustedHistoryContainsD2CorrectionReceipt(t *testing.T, repoDir, sourceRef string) (bool, error) {
	t.Helper()
	logCmd := exec.Command("git", "-C", repoDir, "log", "--full-history", "--format=%H", sourceRef, "--", ".beads/issues.jsonl")
	commits, err := logCmd.Output()
	if err != nil {
		return false, err
	}
	for _, commit := range strings.Fields(string(commits)) {
		showCmd := exec.Command("git", "-C", repoDir, "show", commit+":.beads/issues.jsonl")
		blob, err := showCmd.Output()
		if err != nil {
			return false, fmt.Errorf("read .beads/issues.jsonl at trusted commit %s: %w", commit, err)
		}
		if hasExactD2CorrectionReceipt(blob) {
			return true, nil
		}
	}
	return false, nil
}

func gitForTest(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
}

func writeTrackerAndCommit(t *testing.T, repo string, tracker []byte, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, ".beads", "issues.jsonl"), tracker, 0o600); err != nil {
		t.Fatal(err)
	}
	gitForTest(t, repo, "add", ".beads/issues.jsonl")
	gitForTest(t, repo, "commit", "-q", "-m", message)
}

func hasExactD2CorrectionReceipt(source []byte) bool {
	rows, err := parseDirectEditRows(source)
	if err != nil {
		return false
	}
	row, ok := rows[approvedD2IssueID]
	if !ok {
		return false
	}
	canonicalRow, err := json.Marshal(row.fields)
	if err != nil {
		return false
	}
	digest := sha256.Sum256(canonicalRow)
	return hex.EncodeToString(digest[:]) == approvedD2CorrectionReceiptCanonicalSHA256
}

// TestDirectEditGuardAcceptsMinimalWellFormed ensures a minimal well-formed
// record passes (guards against over-strict rejection).
func TestDirectEditGuardAcceptsMinimalWellFormed(t *testing.T) {
	good := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\"}\n")
	if err := ValidateDirectEdit(good); err != nil {
		t.Fatalf("ValidateDirectEdit rejected well-formed record: %v", err)
	}
}
