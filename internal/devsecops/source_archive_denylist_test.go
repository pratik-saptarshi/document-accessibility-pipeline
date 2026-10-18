package devsecops

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
		name  string
		typ   byte
		link  string
		major int
		minor int
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

const (
	fixtureMetadataSecret         = "secret-metadata-token"
	fixtureMaxMemberMetadataBytes = 4 << 10
	fixtureMaxArchiveEntryCount   = 512
	fixtureMaxExpandedBytes       = 4 << 20
)

func buildMetadataArchive(t *testing.T, global map[string]string, members ...*tar.Header) []byte {
	t.Helper()
	var rawTar bytes.Buffer
	if len(global) > 0 {
		writeGlobalPAXHeader(t, &rawTar, global)
	}
	tw := tar.NewWriter(&rawTar)
	for _, member := range members {
		if err := tw.WriteHeader(member); err != nil {
			t.Fatalf("write metadata test member: %v", err)
		}
		if member.Size > 0 {
			if _, err := io.CopyN(tw, strings.NewReader(strings.Repeat("x", int(member.Size))), member.Size); err != nil {
				t.Fatalf("write metadata test member body: %v", err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := io.Copy(gz, &rawTar); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeGlobalPAXHeader(t *testing.T, output *bytes.Buffer, records map[string]string) {
	t.Helper()
	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var payload bytes.Buffer
	for _, key := range keys {
		value := records[key]
		length := len(key) + len(value) + 3
		for {
			record := fmt.Sprintf("%d %s=%s\n", length, key, value)
			if len(record) == length {
				payload.WriteString(record)
				break
			}
			length = len(record)
		}
	}
	var block [512]byte
	copy(block[0:100], "pax_global_header")
	copy(block[100:108], "0000644\x00")
	copy(block[108:116], "0000000\x00")
	copy(block[116:124], "0000000\x00")
	copy(block[124:136], fmt.Sprintf("%011o\x00", payload.Len()))
	copy(block[136:148], "00000000000\x00")
	for i := 148; i < 156; i++ {
		block[i] = ' '
	}
	block[156] = tar.TypeXGlobalHeader
	copy(block[257:263], "ustar\x00")
	copy(block[263:265], "00")
	var checksum int
	for _, value := range block {
		checksum += int(value)
	}
	copy(block[148:156], fmt.Sprintf("%06o\x00 ", checksum))
	output.Write(block[:])
	output.Write(payload.Bytes())
	if padding := (512 - payload.Len()%512) % 512; padding > 0 {
		output.Write(make([]byte, padding))
	}
}

func TestFilterSourceArchiveChecksEmittedTextMetadata(t *testing.T) {
	list := ArchiveDenylist{DeniedContentPatterns: []*regexp.Regexp{regexp.MustCompile(`secret-metadata-token`)}}
	tests := []struct {
		name   string
		global map[string]string
		member *tar.Header
	}{
		{name: "global pax value", global: map[string]string{"comment": fixtureMetadataSecret}},
		{name: "per-entry pax value", member: &tar.Header{Name: "testroot/file", Typeflag: tar.TypeReg, Format: tar.FormatPAX, PAXRecords: map[string]string{"comment": fixtureMetadataSecret}}},
		{name: "xattr value", member: &tar.Header{Name: "testroot/file", Typeflag: tar.TypeReg, Format: tar.FormatPAX, Xattrs: map[string]string{"user.note": fixtureMetadataSecret}}},
		{name: "file owner", member: &tar.Header{Name: "testroot/file", Typeflag: tar.TypeReg, Uname: fixtureMetadataSecret}},
		{name: "file group", member: &tar.Header{Name: "testroot/file", Typeflag: tar.TypeReg, Gname: fixtureMetadataSecret}},
		{name: "directory owner", member: &tar.Header{Name: "testroot/dir", Typeflag: tar.TypeDir, Uname: fixtureMetadataSecret}},
		{name: "directory group", member: &tar.Header{Name: "testroot/dir", Typeflag: tar.TypeDir, Gname: fixtureMetadataSecret}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var members []*tar.Header
			if tc.member != nil {
				members = append(members, tc.member)
			}
			fixture := buildMetadataArchive(t, tc.global, members...)
			var out bytes.Buffer
			dropped, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list)
			if tc.global != nil {
				if err == nil {
					t.Fatal("global PAX metadata with a denied value was emitted")
				}
				if strings.Contains(err.Error(), fixtureMetadataSecret) {
					t.Fatalf("diagnostic exposed denied metadata: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("filter archive: %v", err)
			}
			if len(dropped) != 1 || dropped[0] != tc.member.Name {
				t.Fatalf("dropped entries = %q; want %q", dropped, []string{tc.member.Name})
			}
			for _, header := range readArchiveHeaders(t, out.Bytes()) {
				if header.Uname == fixtureMetadataSecret || header.Gname == fixtureMetadataSecret {
					t.Fatalf("filtered archive retained denied owner/group metadata in %q", header.Name)
				}
				for key, value := range header.PAXRecords {
					if strings.Contains(key, fixtureMetadataSecret) || strings.Contains(value, fixtureMetadataSecret) {
						t.Fatalf("filtered archive retained denied PAX metadata in %q", header.Name)
					}
				}
				for key, value := range header.Xattrs {
					if strings.Contains(key, fixtureMetadataSecret) || strings.Contains(value, fixtureMetadataSecret) {
						t.Fatalf("filtered archive retained denied xattr metadata in %q", header.Name)
					}
				}
			}
		})
	}
}

func TestFilterSourceArchiveDeniesSensitiveUSTARMemberNames(t *testing.T) {
	const sensitiveName = "testroot/jane.doe@example.com"
	list := ArchiveDenylist{DeniedContentPatterns: []*regexp.Regexp{regexp.MustCompile(`jane\.doe@example\.com`)}}
	fixture := buildMetadataArchive(t, nil, &tar.Header{
		Name: sensitiveName, Mode: 0644, Size: 4, Typeflag: tar.TypeReg, Format: tar.FormatUSTAR,
	})
	inputHeaders := readArchiveHeaders(t, fixture)
	if len(inputHeaders) != 1 || inputHeaders[0].Format != tar.FormatUSTAR {
		t.Fatalf("fixture was not a short USTAR member: %#v", inputHeaders)
	}
	var out bytes.Buffer
	dropped, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list)
	if err != nil {
		t.Fatalf("filter archive: %v", err)
	}
	if strings.Contains(strings.Join(dropped, " "), "jane.doe@example.com") {
		t.Fatalf("sensitive member name was returned in diagnostics: %q", dropped)
	}
	if len(dropped) != 1 || dropped[0] != "[redacted archive member]" {
		t.Fatalf("dropped entries = %q; want one redacted member", dropped)
	}
	if headers := readArchiveHeaders(t, out.Bytes()); len(headers) != 0 {
		t.Fatalf("sensitive-name member was emitted: %#v", headers)
	}
}

func TestFilterSourceArchiveRedactsSensitiveNamesFromErrors(t *testing.T) {
	const sensitive = "jane.doe@example.com"
	list := ArchiveDenylist{DeniedContentPatterns: []*regexp.Regexp{regexp.MustCompile(`jane\.doe@example\.com`)}}
	tests := []struct {
		name   string
		member *tar.Header
	}{
		{
			name:   "unsafe path",
			member: &tar.Header{Name: "testroot/../" + sensitive, Mode: 0644, Typeflag: tar.TypeReg, Format: tar.FormatUSTAR},
		},
		{
			name:   "unsupported type",
			member: &tar.Header{Name: "testroot/" + sensitive, Mode: 0644, Typeflag: tar.TypeSymlink, Linkname: "target", Format: tar.FormatUSTAR},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := buildArchiveWithMember(t, tc.member)
			var out bytes.Buffer
			_, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list)
			if err == nil {
				t.Fatal("invalid archive member was accepted")
			}
			if strings.Contains(err.Error(), sensitive) {
				t.Fatalf("diagnostic exposed sensitive member name: %v", err)
			}
		})
	}
}

func readArchiveHeaders(t *testing.T, data []byte) []*tar.Header {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	var headers []*tar.Header
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return headers
		}
		if err != nil {
			t.Fatal(err)
		}
		copyHeader := *header
		headers = append(headers, &copyHeader)
		if _, err := io.Copy(io.Discard, tr); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFilterSourceArchivePreservesBenignMetadataDeterministically(t *testing.T) {
	fixture := buildMetadataArchive(t, map[string]string{"comment": "safe global note"}, &tar.Header{
		Name: "testroot/file", Mode: 0644, Size: 1, Typeflag: tar.TypeReg, Format: tar.FormatPAX,
		Uname: "build-user", Gname: "build-group", PAXRecords: map[string]string{"comment": "safe entry note"},
		Xattrs: map[string]string{"user.note": "safe xattr"},
	})
	var first, second bytes.Buffer
	for _, output := range []*bytes.Buffer{&first, &second} {
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), output, ArchiveDenylist{DeniedContentPatterns: []*regexp.Regexp{regexp.MustCompile(`secret-metadata-token`)}}); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("metadata-preserving filter output is not byte-deterministic")
	}
	seenGlobal := false
	for _, header := range readArchiveHeaders(t, first.Bytes()) {
		if header.Typeflag == tar.TypeXGlobalHeader {
			seenGlobal = header.PAXRecords["comment"] == "safe global note"
		}
		if header.Name == "testroot/file" {
			if header.Uname != "build-user" || header.Gname != "build-group" || header.PAXRecords["comment"] != "safe entry note" || header.Xattrs["user.note"] != "safe xattr" {
				t.Fatalf("benign metadata changed in output header: %#v", header)
			}
		}
	}
	if !seenGlobal {
		t.Fatal("benign global PAX metadata was not preserved")
	}
}

func TestFilterSourceArchiveBoundsMetadataEntriesAndExpandedSize(t *testing.T) {
	list := ArchiveDenylist{}
	t.Run("metadata boundary accepted", func(t *testing.T) {
		fixture := buildMetadataArchive(t, nil, &tar.Header{
			Name: "testroot/file", Typeflag: tar.TypeReg, Format: tar.FormatPAX,
			PAXRecords: map[string]string{"x": strings.Repeat("a", fixtureMaxMemberMetadataBytes-len("testroot/file")-1)},
		})
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err != nil {
			t.Fatalf("metadata at the configured boundary was rejected: %v", err)
		}
	})
	t.Run("metadata over boundary rejected", func(t *testing.T) {
		fixture := buildMetadataArchive(t, nil, &tar.Header{
			Name: "testroot/file", Typeflag: tar.TypeReg, Format: tar.FormatPAX,
			PAXRecords: map[string]string{"x": strings.Repeat("a", fixtureMaxMemberMetadataBytes-len("testroot/file"))},
		})
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err == nil {
			t.Fatal("oversized PAX metadata was accepted")
		}
	})
	t.Run("entry count boundary rejected", func(t *testing.T) {
		members := make([]*tar.Header, fixtureMaxArchiveEntryCount+1)
		for i := range members {
			members[i] = &tar.Header{Name: fmt.Sprintf("testroot/dir-%03d", i), Typeflag: tar.TypeDir}
		}
		fixture := buildMetadataArchive(t, nil, members...)
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err == nil {
			t.Fatal("archive over the entry-count boundary was accepted")
		}
	})
	t.Run("entry count at boundary accepted", func(t *testing.T) {
		members := make([]*tar.Header, fixtureMaxArchiveEntryCount)
		for i := range members {
			members[i] = &tar.Header{Name: fmt.Sprintf("testroot/dir-%03d", i), Typeflag: tar.TypeDir}
		}
		fixture := buildMetadataArchive(t, nil, members...)
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err != nil {
			t.Fatalf("archive at the entry-count boundary was rejected: %v", err)
		}
	})
	t.Run("expanded size at boundary accepted", func(t *testing.T) {
		partSize := int64((fixtureMaxExpandedBytes - 4*512) / 2)
		fixture := buildMetadataArchive(t, nil,
			&tar.Header{Name: "testroot/a", Typeflag: tar.TypeReg, Size: partSize},
			&tar.Header{Name: "testroot/b", Typeflag: tar.TypeReg, Size: partSize},
		)
		gz, err := gzip.NewReader(bytes.NewReader(fixture))
		if err != nil {
			t.Fatal(err)
		}
		expandedSize, err := io.Copy(io.Discard, gz)
		if err != nil {
			t.Fatal(err)
		}
		if expandedSize != fixtureMaxExpandedBytes {
			t.Fatalf("test archive expands to %d bytes; want exact boundary %d", expandedSize, fixtureMaxExpandedBytes)
		}
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err != nil {
			t.Fatalf("archive at the expanded-size boundary was rejected: %v", err)
		}
	})
	t.Run("expanded size boundary rejected", func(t *testing.T) {
		partSize := int64(fixtureMaxExpandedBytes/2 + 1)
		fixture := buildMetadataArchive(t, nil,
			&tar.Header{Name: "testroot/a", Typeflag: tar.TypeReg, Size: partSize},
			&tar.Header{Name: "testroot/b", Typeflag: tar.TypeReg, Size: partSize},
		)
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader(fixture), &out, list); err == nil {
			t.Fatal("archive over the expanded-size boundary was accepted")
		}
	})
	t.Run("malformed gzip rejected", func(t *testing.T) {
		var out bytes.Buffer
		if _, err := FilterSourceArchive(bytes.NewReader([]byte("not gzip")), &out, list); err == nil {
			t.Fatal("malformed archive was accepted")
		}
	})
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
