//go:build !linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Off Linux the helper fails closed: every capture errors and no snapshot
// is created.
func TestCaptureBoundedInputFailsClosedOffLinux(t *testing.T) {
	src := filepath.Join(t.TempDir(), "input.pdf")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) = %v", src, err)
	}
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(src, dst, 64); err == nil {
		t.Fatal("captureBoundedInput() = nil; want fail-closed error off Linux")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("failed capture left snapshot behind: %v", err)
	}
}
