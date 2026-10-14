//go:build linux || darwin

package financialfixtures

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestV2BoundedFileReadRejectsSymlinkAndFifo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.pdf")
	if err := os.WriteFile(target, []byte("bounded test input"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.pdf")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readV2FileBoundedWithContext(t.Context(), link, v2PDFInputLimit); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("symlink bounded input error = %v, want regular-file rejection", err)
	}
	fifo := filepath.Join(dir, "input.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err := readV2FileBoundedWithContext(t.Context(), fifo, v2PDFInputLimit)
	elapsed := time.Since(start)
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("FIFO bounded input error = %v, want regular-file rejection", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("FIFO bounded read took %s; want prompt rejection, no open block", elapsed)
	}
}
