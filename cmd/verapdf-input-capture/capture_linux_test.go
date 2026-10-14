//go:build linux

package main

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestCaptureBoundedInputRejectsSymlinkEvenToExpectedBytes(t *testing.T) {
	dir := t.TempDir()
	target := writeFixture(t, "target.pdf", []byte("%PDF-1.7 fixture"))
	// The symlink points at a file whose bytes are exactly the expected ones;
	// type validation must still reject it because the pathname was not a
	// regular file at open.
	link := filepath.Join(dir, "link.pdf")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Symlink() = %v", err)
	}
	dst := filepath.Join(dir, "snapshot.pdf")
	if err := captureBoundedInput(link, dst, 64); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error for symlink input")
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("symlink capture left snapshot behind: %v", err)
	}
}

func TestCaptureBoundedInputRejectsSymlinkToMissingTarget(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "dangling.pdf")
	if err := os.Symlink(filepath.Join(dir, "absent.pdf"), link); err != nil {
		t.Fatalf("Symlink() = %v", err)
	}
	dst := filepath.Join(dir, "snapshot.pdf")
	if err := captureBoundedInput(link, dst, 64); err == nil {
		t.Fatal("captureBoundedInput() = nil; want error for dangling symlink")
	}
}

func TestCaptureBoundedInputRejectsFifoWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "input.fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatalf("Mkfifo() = %v", err)
	}
	dst := filepath.Join(dir, "snapshot.pdf")
	start := time.Now()
	err := captureBoundedInput(fifo, dst, 64)
	elapsed := time.Since(start)
	if !errors.Is(err, errInputNotRegular) {
		t.Fatalf("captureBoundedInput() = %v; want errInputNotRegular for FIFO", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("FIFO capture took %s; want prompt rejection, no open block", elapsed)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("FIFO capture left snapshot behind: %v", err)
	}
}
