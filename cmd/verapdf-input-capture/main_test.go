//go:build linux

// Command verapdf-input-capture copies a validator input file into a private
// snapshot, validating its type at open and bounding the captured bytes.
//
// These tests require the Linux open path (O_NOFOLLOW|O_NONBLOCK); off
// Linux the helper fails closed, covered by capture_other_test.go.
package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) = %v", path, err)
	}
	return path
}

func TestCaptureBoundedInputAcceptsRegularFileWithinLimit(t *testing.T) {
	src := writeFixture(t, "input.pdf", []byte("%PDF-1.7 fixture"))
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(src, dst, 64); err != nil {
		t.Fatalf("captureBoundedInput() = %v; want nil", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "%PDF-1.7 fixture" {
		t.Fatalf("snapshot = %q, %v; want fixture bytes", got, err)
	}
	info, err := os.Stat(dst)
	if err != nil || info.Mode().Perm() != 0o444 {
		t.Fatalf("snapshot mode = %v, %v; want 0444", info.Mode().Perm(), err)
	}
}

func TestCaptureBoundedInputAcceptsFileExactlyAtLimit(t *testing.T) {
	src := writeFixture(t, "input.pdf", []byte("12345678"))
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(src, dst, 8); err != nil {
		t.Fatalf("captureBoundedInput() = %v; want nil", err)
	}
}

func TestCaptureBoundedInputRejectsFileOverLimit(t *testing.T) {
	src := writeFixture(t, "input.pdf", []byte("123456789"))
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if !errors.Is(captureBoundedInput(src, dst, 8), errInputLimit) {
		t.Fatalf("captureBoundedInput() did not return errInputLimit")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("oversize capture left snapshot behind: %v", err)
	}
}

func TestCaptureBoundedInputRejectsMissingFile(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if err := captureBoundedInput(filepath.Join(t.TempDir(), "absent.pdf"), dst, 8); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error for missing input")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("failed capture left snapshot behind: %v", err)
	}
}

func TestCaptureBoundedInputRejectsDirectory(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "snapshot.pdf")
	if !errors.Is(captureBoundedInput(t.TempDir(), dst, 8), errInputNotRegular) {
		t.Fatal("captureBoundedInput() did not return errInputNotRegular for a directory")
	}
}

func TestCaptureBoundedInputRefusesToClobberExistingOutput(t *testing.T) {
	src := writeFixture(t, "input.pdf", []byte("data"))
	dst := writeFixture(t, "snapshot.pdf", []byte("existing"))
	if err := captureBoundedInput(src, dst, 8); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error when output already exists")
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "existing" {
		t.Fatalf("existing output = %q, %v; want untouched", got, err)
	}
}
