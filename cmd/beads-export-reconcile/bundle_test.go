package main

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

type testBundleEntry struct {
	name string
	data []byte
}

func TestReadBundleRequiresExactEntrySet(t *testing.T) {
	valid := requiredTestBundleEntries()
	if _, err := readBundle(zipTestBytes(t, valid)); err != nil {
		t.Fatalf("positive control: valid bundle rejected: %v", err)
	}

	tests := []struct {
		name    string
		entries []testBundleEntry
		want    string
	}{
		{name: "missing required entry", entries: withoutTestBundleEntry(valid, "manifest.json"), want: "missing required entries"},
		{name: "duplicate entry", entries: append(append([]testBundleEntry(nil), valid...), testBundleEntry{name: "source.jsonl", data: []byte("duplicate")}), want: `duplicate bundle entry "source.jsonl"`},
		{name: "unexpected entry", entries: append(withoutTestBundleEntry(valid, "manifest.json"), testBundleEntry{name: "extra.txt", data: []byte("unexpected")}), want: `unexpected bundle entry "extra.txt"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := readBundle(zipTestBytes(t, tc.entries)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("readBundle() error = %v, want rejection containing %q", err, tc.want)
			}
		})
	}
}

func TestReadBundleRejectsCorruptArchiveBytes(t *testing.T) {
	valid := zipTestBytes(t, requiredTestBundleEntries())
	truncated := append([]byte(nil), valid[:len(valid)/2]...)
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "non-ZIP bytes", data: []byte("not a ZIP archive")},
		{name: "truncated ZIP central directory", data: truncated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := readBundle(tc.data); err == nil {
				t.Fatal("readBundle() accepted corrupt archive bytes")
			}
		})
	}
}

func TestReadBundleEntrySizeAtLimitAndLimitPlusOne(t *testing.T) {
	const entryLimit = 32 << 20
	base := requiredTestBundleEntries()
	for i := range base {
		if base[i].name == "source.jsonl" {
			base[i].data = bytes.Repeat([]byte{'x'}, entryLimit)
		}
	}
	entries, err := readBundle(zipTestBytes(t, base))
	if err != nil {
		t.Fatalf("entry at %d-byte limit rejected: %v", entryLimit, err)
	}
	if got := len(entries["source.jsonl"]); got != entryLimit {
		t.Fatalf("decoded entry size = %d, want %d", got, entryLimit)
	}

	for i := range base {
		if base[i].name == "source.jsonl" {
			base[i].data = bytes.Repeat([]byte{'x'}, entryLimit+1)
		}
	}
	if _, err := readBundle(zipTestBytes(t, base)); err == nil || !strings.Contains(err.Error(), "bundle entry exceeds 32 MiB") {
		t.Fatalf("entry at limit+1 error = %v, want size rejection", err)
	}
}

func requiredTestBundleEntries() []testBundleEntry {
	return []testBundleEntry{
		{name: "source.jsonl"},
		{name: "baseline.jsonl"},
		{name: "updated.jsonl"},
		{name: "plan.json"},
		{name: "candidate.jsonl"},
		{name: "round-trip.jsonl"},
		{name: "manifest.json"},
	}
}

func withoutTestBundleEntry(entries []testBundleEntry, omitted string) []testBundleEntry {
	result := make([]testBundleEntry, 0, len(entries)-1)
	for _, entry := range entries {
		if entry.name != omitted {
			result = append(result, entry)
		}
	}
	return result
}

func zipTestBytes(t *testing.T, entries []testBundleEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		file, err := writer.Create(entry.name)
		if err != nil {
			t.Fatalf("create zip entry %q: %v", entry.name, err)
		}
		if _, err := file.Write(entry.data); err != nil {
			t.Fatalf("write zip entry %q: %v", entry.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close test zip: %v", err)
	}
	return buffer.Bytes()
}
