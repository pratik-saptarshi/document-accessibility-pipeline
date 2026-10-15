package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestStrictJSONRejectsDuplicateKeys(t *testing.T) {
	var value map[string]string
	if err := strictJSON([]byte(`{"title":"old","title":"new"}`), &value); err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
		t.Fatalf("duplicate-key error = %v", err)
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
