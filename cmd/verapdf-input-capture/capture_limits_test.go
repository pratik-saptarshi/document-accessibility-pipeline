package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Limit validation runs before the platform-specific open, so these tests
// are platform-independent.
func TestCaptureBoundedInputRejectsNegativeLimit(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(filepath.Join(t.TempDir(), "input.pdf"), dst, -1); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error for negative limit")
	}
}

func TestCaptureBoundedInputRejectsMaxInt64Limit(t *testing.T) {
	src := filepath.Join(t.TempDir(), "input.pdf")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) = %v", src, err)
	}
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(src, dst, math.MaxInt64); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error for MaxInt64 limit")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("rejected capture left snapshot behind: %v", err)
	}
}
