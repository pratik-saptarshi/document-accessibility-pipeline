package beadsreconcile

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
	if err := ValidateDirectEditTransition(source, live); err != nil {
		t.Fatalf("live .beads/issues.jsonl failed source-to-candidate guard: %v", err)
	}
}

// TestDirectEditGuardAcceptsMinimalWellFormed ensures a minimal well-formed
// record passes (guards against over-strict rejection).
func TestDirectEditGuardAcceptsMinimalWellFormed(t *testing.T) {
	good := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\"}\n")
	if err := ValidateDirectEdit(good); err != nil {
		t.Fatalf("ValidateDirectEdit rejected well-formed record: %v", err)
	}
}
