package beadsreconcile

import (
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
	uncorrected := d2SourceRow(t)
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

	uncorrected := d2SourceRow(t)
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
		return validateCanonicalExport(candidate)
	}
	return ValidateDirectEditTransition(source, candidate)
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
