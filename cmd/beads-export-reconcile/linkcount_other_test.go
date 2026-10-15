//go:build !linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnsupportedPlatformHardLinkCheckFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(path, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := requireSingleLink(info); err == nil || !strings.Contains(err.Error(), "cannot verify hard-link count") {
		t.Fatalf("requireSingleLink() error = %v, want explicit unsupported-platform failure", err)
	}
	if _, _, _, err := readRegular(path); err == nil || !strings.Contains(err.Error(), "cannot verify hard-link count") {
		t.Fatalf("readRegular() error = %v, want explicit unsupported-platform failure", err)
	}
}
