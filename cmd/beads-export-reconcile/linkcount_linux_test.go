//go:build linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxLinkCountReportsActualHardLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(path, []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path, path+"-alias"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	count, supported := nlink(info)
	if !supported || count != 2 {
		t.Fatalf("nlink() = (%d, %t), want actual link count (2, true)", count, supported)
	}
}
