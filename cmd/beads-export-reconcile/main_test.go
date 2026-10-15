package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestReadRegularRejectsSymlinksAndHardLinks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	if err := os.WriteFile(path, []byte("safe"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "symbolic")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readRegular(link); err == nil {
		t.Fatal("symlink input was accepted")
	}
	hard := filepath.Join(dir, "hard")
	if err := os.Link(path, hard); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readRegular(path); err == nil {
		t.Fatal("hard-linked input was accepted")
	}
}

func TestRequireSingleLinkFailsClosedWithoutPlatformLinkMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	if err := os.WriteFile(path, []byte("safe"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := requireSingleLink(noLinkMetadata{FileInfo: info}); err == nil || !strings.Contains(err.Error(), "cannot verify hard-link count") {
		t.Fatalf("requireSingleLink() error = %v, want fail-closed metadata error", err)
	}
}

type noLinkMetadata struct{ os.FileInfo }

func (noLinkMetadata) Sys() any { return struct{}{} }

func TestStrictJSONRejectsDuplicateKeys(t *testing.T) {
	var value map[string]string
	if err := strictJSON([]byte(`{"title":"old","title":"new"}`), &value); err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
		t.Fatalf("duplicate-key error = %v", err)
	}
}

func TestStrictJSONRejectsNoncanonicalSchemaKeys(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		target any
	}{
		{"plan top-level alias", `{"UPDATES":[]}`, new(beadsreconcile.Plan)},
		{"plan alias collision", `{"updates":[],"UPDATES":[]}`, new(beadsreconcile.Plan)},
		{"plan nested ID alias", `{"updates":[{"ID":"dap-1","fields":{}}]}`, new(beadsreconcile.Plan)},
		{"plan nested fields alias", `{"updates":[{"id":"dap-1","FIELDS":{}}]}`, new(beadsreconcile.Plan)},
		{"manifest alias", `{"VERSION":1}`, new(manifest)},
		{"manifest alias collision", `{"version":1,"VERSION":2}`, new(manifest)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := strictJSON([]byte(tc.data), tc.target); err == nil || !strings.Contains(err.Error(), "noncanonical JSON key") {
				t.Fatalf("strictJSON() error = %v, want noncanonical-key error", err)
			}
		})
	}
}

func TestStrictJSONRejectsNullStringValuesButAllowsExplicitEmpty(t *testing.T) {
	var plan beadsreconcile.Plan
	if err := strictJSON([]byte(`{"updates":[{"id":"dap-1","fields":{"description":null}}]}`), &plan); err == nil || !strings.Contains(err.Error(), "null is not allowed for string") {
		t.Fatalf("null description error = %v, want null-string rejection", err)
	}
	if err := strictJSON([]byte(`{"updates":[{"id":"dap-1","fields":{"description":17}}]}`), &plan); err == nil || !strings.Contains(err.Error(), "must be a JSON string") {
		t.Fatalf("numeric description error = %v, want non-string rejection", err)
	}
	if err := strictJSON([]byte(`{"updates":[{"id":"dap-1","fields":{"description":""}}]}`), &plan); err != nil {
		t.Fatalf("explicit empty description rejected: %v", err)
	}
	if got := plan.Updates[0].Fields["description"]; got != "" {
		t.Fatalf("explicit empty description = %q, want empty string", got)
	}
}

func TestPromoteRequiresRecordedBundleDigest(t *testing.T) {
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	bundle := filepath.Join(dir, "evidence.zip")
	if err := os.WriteFile(canonical, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundle, []byte("bundle"), 0600); err != nil {
		t.Fatal(err)
	}
	err := promote([]string{"--canonical", canonical, "--bundle", bundle, "--bundle-sha256", strings.Repeat("0", 64)})
	if err == nil || !strings.Contains(err.Error(), "bundle digest mismatch") {
		t.Fatalf("replaced bundle error = %v", err)
	}
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "source" {
		t.Fatal("digest mismatch modified canonical export")
	}
}

func TestAtomicReplaceRejectsConcurrentPublisherLock(t *testing.T) {
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	lock := canonical + ".beads-reconcile.lock"
	originalBytes := []byte("source")
	if err := os.WriteFile(canonical, originalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	original, err := os.Stat(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock, []byte("another publisher holds the lock"), 0600); err != nil {
		t.Fatal(err)
	}

	err = atomicReplace(canonical, original, []byte("candidate"), sha(originalBytes))
	if err == nil || !strings.Contains(err.Error(), "lock unavailable") {
		t.Fatalf("atomicReplace() error = %v, want lock contention error", err)
	}
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(originalBytes) {
		t.Fatalf("canonical after lock contention = %q, want unchanged source", got)
	}
}

func TestCreateAndPromoteWithPinnedBeads(t *testing.T) {
	bd := os.Getenv("DAP_BD_BINARY")
	if bd == "" {
		t.Skip("set DAP_BD_BINARY to run the isolated Beads integration test")
	}
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	planPath := filepath.Join(dir, "plan.json")
	bundle := filepath.Join(dir, "evidence.zip")
	initial := `{"id":"dap-smoke","title":"old","status":"open","priority":2,"issue_type":"task","created_at":"2026-10-14T00:00:00Z","updated_at":"2026-10-14T00:00:00Z","dependency_count":0,"dependent_count":0}` + "\n"
	if err := os.WriteFile(canonical, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte(`{"updates":[{"id":"dap-smoke","fields":{"title":"new"}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := create([]string{"--canonical", canonical, "--plan", planPath, "--bd", bd, "--output", bundle}); err != nil {
		t.Fatalf("create: %v", err)
	}
	archive, err := os.Open(bundle)
	if err != nil {
		t.Fatal(err)
	}
	info, err := archive.Stat()
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(archive, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	_ = archive.Close()
	if len(zr.File) != 7 {
		t.Fatalf("bundle entries = %d, want 7", len(zr.File))
	}
	bundleBytes, err := os.ReadFile(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", canonical, "--bundle", bundle, "--bundle-sha256", sha(bundleBytes)}); err != nil {
		t.Fatalf("promote: %v", err)
	}
	data, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"title":"new"`) || strings.Contains(string(data), `"title":"old"`) {
		t.Fatalf("promoted bytes = %s", data)
	}
}

func TestPromoteRejectsStaleCanonical(t *testing.T) {
	bd := os.Getenv("DAP_BD_BINARY")
	if bd == "" {
		t.Skip("set DAP_BD_BINARY to run the isolated Beads integration test")
	}
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	planPath := filepath.Join(dir, "plan.json")
	bundle := filepath.Join(dir, "evidence.zip")
	initial := `{"id":"dap-smoke","title":"old","status":"open","priority":2,"issue_type":"task","created_at":"2026-10-14T00:00:00Z","updated_at":"2026-10-14T00:00:00Z","dependency_count":0,"dependent_count":0}` + "\n"
	if err := os.WriteFile(canonical, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte(`{"updates":[{"id":"dap-smoke","fields":{"title":"new"}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := create([]string{"--canonical", canonical, "--plan", planPath, "--bd", bd, "--output", bundle}); err != nil {
		t.Fatalf("create: %v", err)
	}
	changed := strings.Replace(initial, `"title":"old"`, `"title":"raced"`, 1)
	if err := os.WriteFile(canonical, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	bundleBytes, err := os.ReadFile(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", canonical, "--bundle", bundle, "--bundle-sha256", sha(bundleBytes)}); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale promote error = %v", err)
	}
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != changed {
		t.Fatal("stale promotion changed the canonical file")
	}
}
