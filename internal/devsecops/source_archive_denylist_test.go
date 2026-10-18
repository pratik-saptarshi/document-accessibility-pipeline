package devsecops

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Synthetic personal data for fixtures. Never real personal data.
const (
	fixtureFakeEmail   = "jane.doe.1984@example.com"
	fixtureFakeAddress = "742 Evergreen Terrace"
)

func buildDenylistFixtureArchive(t *testing.T) []byte {
	t.Helper()
	files := map[string]string{
		"testroot/.beads/issues.jsonl": "{\"owner\":\"" + fixtureFakeEmail + "\"}\n",
		"testroot/.beads/config.yaml":  "version: 1\n",
		"testroot/contacts.txt":        "Jane Doe\n" + fixtureFakeAddress + "\nSpringfield\n",
		"testroot/cmd/app/main.go":     "package main\n\nfunc main() {}\n",
		"testroot/README.md":           "# Fixture\n\nTeam-maintained documentation.\n",
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, name := range names {
		body := files[name]
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     0644,
			Size:     int64(len(body)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildNamedArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, name := range names {
		body := files[name]
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildArchiveWithMember(t *testing.T, member *tar.Header) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "testroot/README.md", Mode: 0644, Size: 5, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(tw, "safe\n"); err != nil {
		t.Fatal(err)
	}
	if err := tw.WriteHeader(member); err != nil {
		t.Fatalf("write test member of type %d: %v", member.Typeflag, err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func readArchiveMembers(t *testing.T, data []byte) map[string]string {
	t.Helper()
	members := make(map[string]string)
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		members[header.Name] = string(body)
	}
	return members
}

func TestFilterSourceArchiveDropsDenylistedMembers(t *testing.T) {
	fixture := buildDenylistFixtureArchive(t)
	var out bytes.Buffer
	dropped, err := FilterSourceArchive(bytes.NewReader(fixture), &out, DefaultSourceArchiveDenylist())
	if err != nil {
		t.Fatalf("FilterSourceArchive: %v", err)
	}
	members := readArchiveMembers(t, out.Bytes())
	for name := range members {
		rel := strings.TrimPrefix(name, "testroot/")
		if rel == ".beads" || strings.HasPrefix(rel, ".beads/") {
			t.Errorf("filtered archive still contains denylisted path %q", name)
		}
		if strings.Contains(members[name], fixtureFakeEmail) || strings.Contains(members[name], fixtureFakeAddress) {
			t.Errorf("filtered archive still contains personal data in %q", name)
		}
	}
	if _, ok := members["testroot/contacts.txt"]; ok {
		t.Error("filtered archive still contains the file with the fake personal address")
	}
	for _, want := range []string{"testroot/cmd/app/main.go", "testroot/README.md"} {
		if _, ok := members[want]; !ok {
			t.Errorf("filtered archive dropped clean file %q", want)
		}
	}
	if got := members["testroot/cmd/app/main.go"]; got != "package main\n\nfunc main() {}\n" {
		t.Errorf("clean file content altered: %q", got)
	}
	wantDropped := []string{
		"testroot/.beads/config.yaml",
		"testroot/.beads/issues.jsonl",
		"testroot/contacts.txt",
	}
	if len(dropped) != len(wantDropped) {
		t.Fatalf("dropped %d members %q; want %d %q", len(dropped), dropped, len(wantDropped), wantDropped)
	}
	for i, want := range wantDropped {
		if dropped[i] != want {
			t.Errorf("dropped[%d] = %q; want %q", i, dropped[i], want)
		}
	}
}

func TestFilterSourceArchiveIsDeterministic(t *testing.T) {
	fixture := buildDenylistFixtureArchive(t)
	var first, second bytes.Buffer
	if _, err := FilterSourceArchive(bytes.NewReader(fixture), &first, DefaultSourceArchiveDenylist()); err != nil {
		t.Fatal(err)
	}
	if _, err := FilterSourceArchive(bytes.NewReader(fixture), &second, DefaultSourceArchiveDenylist()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("filter output is not byte-deterministic")
	}
}

func TestDefaultSourceArchiveDenylistPathRules(t *testing.T) {
	list := DefaultSourceArchiveDenylist()
	for _, denied := range []string{".beads", ".beads/", ".beads/issues.jsonl", ".beads/nested/deep.txt"} {
		if reason := list.DeniedReason(denied, nil); reason == "" {
			t.Errorf("expected denylist to deny path %q", denied)
		}
	}
	for _, allowed := range []string{"cmd/app/main.go", ".beadsx/notes.txt", "docs/.beads-mention.txt"} {
		if reason := list.DeniedReason(allowed, []byte("clean content\n")); reason != "" {
			t.Errorf("expected denylist to allow path %q (reason %q)", allowed, reason)
		}
	}
}

func TestFilterSourceArchiveRejectsUnsafeMemberLayouts(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		{"rootless tracker path", map[string]string{".beads/config.yaml": "version: 1\n"}},
		{"traversal tracker path", map[string]string{"repo/x/../.beads/config.yaml": "version: 1\n"}},
		{"absolute path", map[string]string{"/repo/.beads/config.yaml": "version: 1\n"}},
		{"backslash path", map[string]string{"repo\\..\\repo\\.beads\\config.yaml": "version: 1\n"}},
		{"multiple roots", map[string]string{"repo/README.md": "safe\n", "other/.beads/config.yaml": "version: 1\n"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := buildNamedArchive(t, tc.files)
			var filtered bytes.Buffer
			if _, err := FilterSourceArchive(bytes.NewReader(fixture), &filtered, DefaultSourceArchiveDenylist()); err == nil {
				t.Fatal("unsafe archive member layout was accepted")
			}
		})
	}
}

func TestFilterSourceArchiveRejectsLinksAndSpecialMembers(t *testing.T) {
	tests := []struct {
		name   string
		typ    byte
		link   string
		major  int
		minor  int
	}{
		{name: "symbolic link", typ: tar.TypeSymlink, link: "/outside/file.go"},
		{name: "hard link", typ: tar.TypeLink, link: "../../outside/file.go"},
		{name: "fifo", typ: tar.TypeFifo},
		{name: "character device", typ: tar.TypeChar, major: 1, minor: 3},
		{name: "block device", typ: tar.TypeBlock, major: 1, minor: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := buildArchiveWithMember(t, &tar.Header{
				Name:     "testroot/sub/member",
				Mode:     0644,
				Typeflag: tc.typ,
				Linkname: tc.link,
				Devmajor: int64(tc.major),
				Devminor: int64(tc.minor),
			})
			var filtered bytes.Buffer
			if _, err := FilterSourceArchive(bytes.NewReader(fixture), &filtered, DefaultSourceArchiveDenylist()); err == nil {
				t.Fatalf("archive member type %d was accepted", tc.typ)
			}
		})
	}
}

// TestFilteredBootstrapSourceArchiveHasNoDenylistedMembers guards the
// replacement artifact: it must already be a fixed point of the membership
// denylist, so re-applying the filter changes nothing.
func TestFilteredBootstrapSourceArchiveHasNoDenylistedMembers(t *testing.T) {
	archivePath := filepath.Clean(filteredBootstrapSourceArchivePath)
	data, err := os.ReadFile(archivePath)
	if errors.Is(err, os.ErrNotExist) {
		if _, gitErr := os.Stat(filepath.Join("..", "..", ".git")); errors.Is(gitErr, os.ErrNotExist) {
			t.Skip("source archives omit export-ignored source-artifact blobs")
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	denials, err := ArchiveDenials(bytes.NewReader(data), DefaultSourceArchiveDenylist())
	if err != nil {
		t.Fatalf("ArchiveDenials: %v", err)
	}
	if len(denials) != 0 {
		t.Fatalf("pinned archive contains denylisted members: %q", denials)
	}
}
