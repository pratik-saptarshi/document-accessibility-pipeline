package beadsreconcile

import (
	"os"
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
	bad := []byte("{\"id\": \"a\", \"status\": \"open\", \"title\": \"t\", \"dependencies\": [{\"issue_id\": \"a\", \"depends_on_id\": \"b\"}], \"dependency_count\": 5}\n")
	if err := ValidateDirectEdit(bad); err == nil {
		t.Fatal("ValidateDirectEdit accepted inconsistent dependency_count, want error")
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
	// The live file is expected to pass only if it is well-formed. If it
	// fails, the failure message names the problem.
	if err := ValidateDirectEdit(live); err != nil {
		// Distinguish "live file has a real problem" from "not applicable":
		// report the error so CI surfaces it.
		t.Fatalf("live .beads/issues.jsonl failed direct-edit guard: %v", err)
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
