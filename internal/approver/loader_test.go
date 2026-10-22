// Tests for the JSONL snapshot loader (Task 3).
//
// The loader is a hostile-input boundary: the register file is untrusted
// bytes and the public key is an injected caller parameter — never read from
// the file. All identities and keys are synthetic. All fixtures are written
// to t.TempDir() in-test; no fixture files are committed.
package approver

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// writeFixture writes content to a temp file and returns its path.
func writeFixture(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "register.jsonl")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

// marshalFixtureChain builds a valid chain and returns its JSONL bytes.
func marshalFixtureChain(t *testing.T, b *chainBuilder) []byte {
	t.Helper()
	var buf bytes.Buffer
	for _, e := range b.entries {
		line, err := MarshalEntry(e)
		if err != nil {
			t.Fatalf("MarshalEntry: %v", err)
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// lineAt returns the JSONL line at index i of content (0-based).
func lineAt(t *testing.T, content []byte, i int) []byte {
	t.Helper()
	lines := bytes.Split(bytes.TrimSuffix(content, []byte("\n")), []byte("\n"))
	if i >= len(lines) {
		t.Fatalf("lineAt: only %d lines", len(lines))
	}
	return lines[i]
}

// mutateLine decodes a JSONL line into a map, applies mutate, and re-encodes it.
func mutateLine(t *testing.T, line []byte, mutate func(m map[string]any)) []byte {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		t.Fatalf("mutateLine decode: %v", err)
	}
	mutate(m)
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("mutateLine encode: %v", err)
	}
	return out
}

func TestLoadMalformedLine(t *testing.T) {
	pub, _ := mustTestKeys(t)
	path := writeFixture(t, []byte("this is not json at all\n"))
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for malformed JSONL line, got nil")
	}
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("expected error wrapping ErrMalformed, got %v", err)
	}
}

func TestLoadOversizedLine(t *testing.T) {
	pub, _ := mustTestKeys(t)
	// A single line well above the 1 MiB bound. It need not be valid JSON:
	// the size bound applies before parsing.
	huge := bytes.Repeat([]byte("x"), 2*1024*1024)
	path := writeFixture(t, append(append([]byte{}, huge...), '\n'))
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for oversized line, got nil")
	}
}

func TestLoadOversizedValidEntry(t *testing.T) {
	// Even a cryptographically valid entry must be rejected if its line
	// exceeds the size bound.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	identity := "synthetic-" + strings.Repeat("a", 2*1024*1024)
	b.addGrant(identity, "approver", baseTime, time.Time{}, baseTime)
	line, err := MarshalEntry(b.entries[0])
	if err != nil {
		t.Fatalf("MarshalEntry: %v", err)
	}
	path := writeFixture(t, append(append([]byte{}, line...), '\n'))
	if _, err := LoadSnapshot(path, pub); err == nil {
		t.Fatal("expected error for oversized valid entry, got nil")
	}
}

func TestLoadTrailingGarbage(t *testing.T) {
	// The wire format is newline-delimited JSONL: a final partial line with
	// no trailing newline is rejected, not silently accepted or skipped.
	pub, _ := mustTestKeys(t)
	path := writeFixture(t, []byte("{\"v\":1}\npartial-final-line-without-newline"))
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for trailing garbage without final newline, got nil")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	pub, _ := mustTestKeys(t)
	path := writeFixture(t, []byte{})
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestLoadUnknownField(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, time.Time{}, baseTime)
	content := marshalFixtureChain(t, b)
	bad := mutateLine(t, lineAt(t, content, 0), func(m map[string]any) {
		m["injected_field"] = "attack"
	})
	path := writeFixture(t, append(bad, '\n'))
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for line with unknown field, got nil")
	}
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("expected error wrapping ErrMalformed, got %v", err)
	}
}

func TestLoadVersionTwo(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, time.Time{}, baseTime)
	content := marshalFixtureChain(t, b)
	bad := mutateLine(t, lineAt(t, content, 0), func(m map[string]any) {
		m["v"] = float64(2)
	})
	path := writeFixture(t, append(bad, '\n'))
	_, err := LoadSnapshot(path, pub)
	if err == nil {
		t.Fatal("expected error for v:2 line, got nil")
	}
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("expected error wrapping ErrMalformed, got %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	pub, _ := mustTestKeys(t)
	_, err := LoadSnapshot(filepath.Join(t.TempDir(), "does-not-exist.jsonl"), pub)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadValidSnapshot(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	b.addGrant("synthetic-approver-002", "approver", baseTime.Add(time.Hour), time.Time{}, baseTime.Add(time.Hour))
	content := marshalFixtureChain(t, b)
	path := writeFixture(t, content)

	snap, err := LoadSnapshot(path, pub)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if len(snap.Entries) != len(b.entries) {
		t.Fatalf("got %d entries, want %d", len(snap.Entries), len(b.entries))
	}
	for i := range snap.Entries {
		got, want := snap.Entries[i], b.entries[i]
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("entry %d mismatch:\n got %+v\nwant %+v", i, got, want)
		}
	}
	// The public key is the injected caller parameter. The wire format has
	// no key field at all (ParseEntry would reject it as unknown), so the
	// file can never supply a key to substitute.
	if !bytes.Equal(snap.PubKey, pub) {
		t.Fatal("Snapshot.PubKey is not the injected key")
	}
}

func TestLoadSnapshotVerifiesWithIndependentHead(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	b.addGrant("synthetic-approver-002", "approver", baseTime.Add(time.Hour), time.Time{}, baseTime.Add(time.Hour))
	content := marshalFixtureChain(t, b)
	path := writeFixture(t, content)

	snap, err := LoadSnapshot(path, pub)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	// Compute the head independently from the loaded entries: fold ChainHash
	// over CanonicalBytes, starting from the genesis zero hash.
	var head [32]byte
	for i := range snap.Entries {
		head = ChainHash(snap.Entries[i].PrevHash, snap.Entries[i].CanonicalBytes())
	}

	if err := VerifyRoleAt("synthetic-approver-001", "approver",
		baseTime.Add(12*time.Hour), snap, head); err != nil {
		t.Fatalf("VerifyRoleAt with independently computed head: %v", err)
	}
}

func TestLoadSnapshotVerifiesWrongHead(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	b.addGrant("synthetic-approver-002", "approver", baseTime.Add(time.Hour), time.Time{}, baseTime.Add(time.Hour))
	content := marshalFixtureChain(t, b)
	path := writeFixture(t, content)

	snap, err := LoadSnapshot(path, pub)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	var wrong [32]byte
	wrong[0] = 0x01
	if err := VerifyRoleAt("synthetic-approver-001", "approver",
		baseTime.Add(12*time.Hour), snap, wrong); !errors.Is(err, ErrHeadMismatch) {
		t.Fatalf("expected ErrHeadMismatch, got %v", err)
	}
}
