package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicReplaceFailurePathsPreserveSourceAndCleanArtifacts(t *testing.T) {
	for _, stage := range []string{"write", "sync", "close", "rename"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			canonical := filepath.Join(dir, "issues.jsonl")
			originalBytes := []byte("original canonical bytes\n")
			candidate := []byte("candidate canonical bytes\n")
			if err := os.WriteFile(canonical, originalBytes, 0600); err != nil {
				t.Fatal(err)
			}
			original, err := os.Stat(canonical)
			if err != nil {
				t.Fatal(err)
			}
			injected := fmt.Errorf("injected %s failure", stage)
			ops := defaultAtomicReplaceOps()
			switch stage {
			case "write":
				ops.writeTemp = func(file *os.File, data []byte) (int, error) {
					_, _ = file.Write(data[:3])
					return 3, injected
				}
			case "sync":
				ops.syncTemp = func(*os.File) error { return injected }
			case "close":
				ops.closeTemp = func(file *os.File) error {
					_ = file.Close()
					return injected
				}
			case "rename":
				ops.rename = func(string, string) error { return injected }
			}

			err = atomicReplaceWithOps(canonical, original, candidate, sha(originalBytes), ops)
			if !errors.Is(err, injected) {
				t.Fatalf("atomicReplaceWithOps() error = %v, want injected failure %v", err, injected)
			}
			got, err := os.ReadFile(canonical)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, originalBytes) {
				t.Fatalf("canonical after pre-rename %s failure = %q, want original %q", stage, got, originalBytes)
			}
			assertNoReplacementArtifacts(t, dir, canonical)
		})
	}
}

func TestAtomicReplaceDirectorySyncFailureReportsAlreadyReplacedSource(t *testing.T) {
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	originalBytes := []byte("original canonical bytes\n")
	candidate := []byte("candidate canonical bytes\n")
	if err := os.WriteFile(canonical, originalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	original, err := os.Stat(canonical)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected directory sync failure")
	ops := defaultAtomicReplaceOps()
	ops.syncDir = func(*os.File) error { return injected }

	err = atomicReplaceWithOps(canonical, original, candidate, sha(originalBytes), ops)
	if !errors.Is(err, injected) || !strings.Contains(err.Error(), "canonical candidate was replaced but directory sync failed") {
		t.Fatalf("directory sync failure = %v, want truthful post-rename error wrapping %v", err, injected)
	}
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, candidate) {
		t.Fatalf("canonical after directory sync failure = %q, want published candidate %q", got, candidate)
	}
	assertNoReplacementArtifacts(t, dir, canonical)
}

func TestAtomicReplaceSuccessReleasesLockAndTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	canonical := filepath.Join(dir, "issues.jsonl")
	originalBytes := []byte("original canonical bytes\n")
	candidate := []byte("candidate canonical bytes\n")
	if err := os.WriteFile(canonical, originalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	original, err := os.Stat(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if err := atomicReplace(canonical, original, candidate, sha(originalBytes)); err != nil {
		t.Fatalf("successful atomicReplace() failed: %v", err)
	}
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, candidate) {
		t.Fatalf("canonical after successful replacement = %q, want %q", got, candidate)
	}
	assertNoReplacementArtifacts(t, dir, canonical)
}

func assertNoReplacementArtifacts(t *testing.T, dir, canonical string) {
	t.Helper()
	lock := canonical + ".beads-reconcile.lock"
	if _, err := os.Lstat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock cleanup stat error = %v; lock should be absent", err)
	}
	assertNoCandidateTemporaryFiles(t, dir)
}

func assertNoCandidateTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".beads-candidate-") {
			t.Errorf("temporary replacement file remains after return: %q", entry.Name())
		}
	}
}
