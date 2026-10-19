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
	"slices"
	"strings"
	"testing"
)

const (
	bootstrapSourceArchivePath         = "../../buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz"
	bootstrapSourceArchiveSize         = 388215
	bootstrapSourceArchiveSHA          = "f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d"
	filteredBootstrapSourceArchivePath = "../../buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz"
	bootstrapSourceArchiveRoot         = "document-accessibility-pipeline-ef4c35629c467887b6749f9e31af111cbe8f8913"
)

func TestPinnedBootstrapSourceArchive(t *testing.T) {
	attributes, err := os.ReadFile(filepath.Join("..", "..", ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(attributes, []byte("buildtools/source-archives/** export-ignore")) {
		t.Fatal("future Git archives must exclude stored source archives")
	}
	if !bytes.Contains(attributes, []byte(".beads/** export-ignore")) {
		t.Fatal("future Git archives must exclude issue-tracker state")
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

func TestPinnedBootstrapSourceArchivePairs(t *testing.T) {
	tests := []struct {
		name         string
		original     string
		originalSize int64
		originalSHA  string
		filtered     string
		filteredSize int64
		filteredSHA  string
	}{
		{
			name:         "historical-631260",
			original:     "cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz",
			originalSize: 306734,
			originalSHA:  "6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5",
			filtered:     "cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a-filtered-afeb27cdf5179516.tar.gz",
			filteredSize: 259355,
			filteredSHA:  "afeb27cdf51795168e7e606cd736d20988aecfba85949f5d36af671d8b8eeaed",
		},
		{
			name:         "pinned-ef4c",
			original:     "cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz",
			originalSize: 469726,
			originalSHA:  "16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473",
			filtered:     "cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz",
			filteredSize: 388215,
			filteredSHA:  "f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			archiveDir := filepath.Join("..", "..", "buildtools", "source-archives")
			original, err := os.ReadFile(filepath.Join(archiveDir, tc.original))
			if err != nil {
				t.Fatal(err)
			}
			if int64(len(original)) != tc.originalSize {
				t.Fatalf("original archive size = %d; want %d", len(original), tc.originalSize)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(original)); got != tc.originalSHA {
				t.Fatalf("original archive SHA-256 = %s; want %s", got, tc.originalSHA)
			}
			originalDenied, err := ArchiveDenials(bytes.NewReader(original), DefaultSourceArchiveDenylist())
			if err != nil {
				t.Fatalf("inspect original archive denied members: %v", err)
			}
			if len(originalDenied) == 0 {
				t.Fatal("original archive contains no members selected by the denylist")
			}

			filtered, err := os.ReadFile(filepath.Join(archiveDir, tc.filtered))
			if err != nil {
				t.Fatal(err)
			}
			if int64(len(filtered)) != tc.filteredSize {
				t.Fatalf("filtered archive size = %d; want %d", len(filtered), tc.filteredSize)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(filtered)); got != tc.filteredSHA {
				t.Fatalf("filtered archive SHA-256 = %s; want %s", got, tc.filteredSHA)
			}
			filteredDenied, err := ArchiveDenials(bytes.NewReader(filtered), DefaultSourceArchiveDenylist())
			if err != nil {
				t.Fatalf("inspect filtered archive denied members: %v", err)
			}
			if len(filteredDenied) != 0 {
				t.Fatalf("filtered archive still contains denied members: %q", filteredDenied)
			}

			var first, second bytes.Buffer
			firstDenied, err := FilterSourceArchive(bytes.NewReader(original), &first, DefaultSourceArchiveDenylist())
			if err != nil {
				t.Fatalf("regenerate filtered archive: %v", err)
			}
			secondDenied, err := FilterSourceArchive(bytes.NewReader(original), &second, DefaultSourceArchiveDenylist())
			if err != nil {
				t.Fatalf("repeat filtered archive regeneration: %v", err)
			}
			if len(firstDenied) == 0 || !slices.Equal(firstDenied, originalDenied) || !slices.Equal(secondDenied, firstDenied) {
				t.Fatalf("denylisted members are not stable across filtering: archive=%q first=%q second=%q", originalDenied, firstDenied, secondDenied)
			}
			if !bytes.Equal(first.Bytes(), filtered) || !bytes.Equal(second.Bytes(), filtered) {
				t.Fatal("filtered archive is not the deterministic output of the pinned original bytes")
			}
		})
	}
}
