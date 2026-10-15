package devsecops

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

const (
	bootstrapSourceArchivePath = "../../buildtools/source-archives/cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz"
	bootstrapSourceArchiveSize = 306734
	bootstrapSourceArchiveSHA  = "6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5"
	bootstrapSourceArchiveRoot = "document-accessibility-pipeline-63126024666fd71b48c25df55d1e98924e5eb09a"
)

func TestPinnedBootstrapSourceArchive(t *testing.T) {
	attributes, err := os.ReadFile(filepath.Join("..", "..", ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(attributes, []byte("buildtools/source-archives/** export-ignore")) {
		t.Fatal("future Git archives must exclude stored source archives")
	}

	archivePath := filepath.Clean(bootstrapSourceArchivePath)
	archive, err := os.Open(archivePath)
	if errors.Is(err, os.ErrNotExist) {
		if _, gitErr := os.Stat(filepath.Join("..", "..", ".git")); errors.Is(gitErr, os.ErrNotExist) {
			t.Skip("source archives omit export-ignored source-artifact blobs")
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	info, err := archive.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != bootstrapSourceArchiveSize {
		t.Fatalf("archive size = %d; want %d", info.Size(), bootstrapSourceArchiveSize)
	}
	data, err := io.ReadAll(io.LimitReader(archive, bootstrapSourceArchiveSize+1))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != bootstrapSourceArchiveSize {
		t.Fatalf("read %d archive bytes; want %d", len(data), bootstrapSourceArchiveSize)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != bootstrapSourceArchiveSHA {
		t.Fatalf("archive SHA-256 = %s; want %s", got, bootstrapSourceArchiveSHA)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open archive gzip stream: %v", err)
	}
	defer gzipReader.Close()
	limited := &io.LimitedReader{R: gzipReader, N: 4 << 20}
	tarReader := tar.NewReader(limited)
	seen := make(map[string]bool)
	regularFiles := 0
	for entries := 0; ; entries++ {
		if entries >= 512 {
			t.Fatal("archive exceeds 512 entries")
		}
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read archive entry: %v", err)
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			if header.Name != "pax_global_header" || header.Size > 256 {
				t.Fatalf("unexpected global archive metadata: %q (%d bytes)", header.Name, header.Size)
			}
			continue
		}
		name := strings.TrimSuffix(header.Name, "/")
		if name != bootstrapSourceArchiveRoot && !strings.HasPrefix(name, bootstrapSourceArchiveRoot+"/") {
			t.Fatalf("archive entry outside pinned root: %q", header.Name)
		}
		relative := strings.TrimPrefix(name, bootstrapSourceArchiveRoot)
		relative = strings.TrimPrefix(relative, "/")
		clean := path.Clean(relative)
		if path.IsAbs(relative) || (relative != "" && clean != relative) || clean == ".." || strings.HasPrefix(clean, "../") {
			t.Fatalf("unsafe archive path: %q", header.Name)
		}
		if seen[name] {
			t.Fatalf("duplicate archive path: %q", header.Name)
		}
		seen[name] = true
		if header.Size < 0 || header.Size > 2<<20 {
			t.Fatalf("archive entry size out of bounds: %q (%d bytes)", header.Name, header.Size)
		}
		switch header.Typeflag {
		case tar.TypeDir:
		case tar.TypeReg, tar.TypeRegA:
			regularFiles++
			if strings.HasPrefix(relative, "buildtools/source-archives/") {
				t.Fatalf("archive recursively contains a stored source archive: %q", header.Name)
			}
		default:
			t.Fatalf("unsupported archive entry type %d: %q", header.Typeflag, header.Name)
		}
	}
	remaining, err := io.Copy(io.Discard, limited)
	if err != nil {
		t.Fatalf("finish archive gzip stream: %v", err)
	}
	if limited.N == 0 || remaining > 4<<20 {
		t.Fatal("archive expands beyond 4 MiB bound")
	}
	if !seen[bootstrapSourceArchiveRoot+"/LICENSE"] || !seen[bootstrapSourceArchiveRoot+"/buildtools.manifest.json"] || regularFiles < 100 {
		t.Fatalf("archive is missing expected source files (regular files: %d)", regularFiles)
	}
	if err := gzipReader.Close(); err != nil {
		t.Fatalf("close archive gzip stream: %v", err)
	}
}
