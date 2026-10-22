// Tests for the approver-role register entry format (Task 1).
//
// All identities and keys are synthetic. No real person names, no real keys.
package approver

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

// testGrantEntry returns a deterministic synthetic grant entry.
func testGrantEntry() Entry {
	return Entry{
		Version:    FormatVersion,
		Seq:        0,
		PrevHash:   [32]byte{},
		Type:       EntryGrant,
		Identity:   "synthetic-approver-001",
		Role:       "approver",
		ValidFrom:  time.Date(2026, 10, 21, 22, 0, 0, 0, time.UTC).UnixNano(),
		ValidUntil: 0,
		AppendedAt: time.Date(2026, 10, 21, 22, 0, 1, 0, time.UTC).UnixNano(),
	}
}

// testGrantLine returns a deterministic synthetic grant JSONL line for parse tests.
func testGrantLine() string {
	return `{"v":1,"seq":0,"prev":"` + strings.Repeat("0", 64) + `",` +
		`"type":"grant","identity":"synthetic-approver-001","role":"approver",` +
		`"valid_from":"2026-10-21T22:00:00Z","valid_until":null,"revoked_at":null,` +
		`"appended_at":"2026-10-21T22:00:01Z","sig":"` + strings.Repeat("0", 128) + `"}`
}

func TestCanonicalBytesDeterministic(t *testing.T) {
	e := testGrantEntry()
	a, b := e.CanonicalBytes(), e.CanonicalBytes()
	if !bytes.Equal(a, b) {
		t.Fatal("same entry produced different canonical bytes")
	}
}

func TestCanonicalBytesFieldOrderSensitivity(t *testing.T) {
	// Fixed order: Version(1B), Seq(8B BE), PrevHash(32B), Type(1B), ...
	e := Entry{Version: FormatVersion, Seq: 0x0102030405060708, Type: EntryGrant}
	cb := e.CanonicalBytes()
	if cb[0] != 0x01 {
		t.Fatalf("byte 0: got %#02x, want Version 0x01", cb[0])
	}
	if want := []byte{1, 2, 3, 4, 5, 6, 7, 8}; !bytes.Equal(cb[1:9], want) {
		t.Fatalf("bytes 1..8: got %x, want big-endian Seq %x", cb[1:9], want)
	}
	if want := 1 + 8 + 32 + 1 + 1 + 1 + 8 + 8 + 8 + 8; len(cb) != want {
		t.Fatalf("empty Identity/Role entry: got %d bytes, want %d", len(cb), want)
	}
	// Identity and Role occupy ordered, distinct slots: swapping them changes bytes.
	a := testGrantEntry()
	a.Identity, a.Role = "aaa", "bbb"
	b := testGrantEntry()
	b.Identity, b.Role = "bbb", "aaa"
	if bytes.Equal(a.CanonicalBytes(), b.CanonicalBytes()) {
		t.Fatal("swapping Identity/Role must change canonical bytes")
	}
}

func TestChainHashCommitsToPrevious(t *testing.T) {
	canonical := []byte("synthetic-canonical-bytes")
	h1 := ChainHash([32]byte{1}, canonical)
	if ChainHash([32]byte{1}, canonical) != h1 {
		t.Fatal("ChainHash is not deterministic")
	}
	if h1 == ChainHash([32]byte{2}, canonical) {
		t.Fatal("ChainHash does not commit to the previous hash")
	}
	if h1 == ChainHash([32]byte{1}, []byte("other-canonical-bytes")) {
		t.Fatal("ChainHash does not commit to the canonical bytes")
	}
	var zero [32]byte
	if ChainHash(zero, canonical) == h1 {
		t.Fatal("different previous hashes must yield different chain hashes")
	}
}

func TestParseEntryGenesisPrevHashZeros(t *testing.T) {
	e, err := ParseEntry([]byte(testGrantLine()))
	if err != nil {
		t.Fatalf("ParseEntry: %v", err)
	}
	if e.PrevHash != [32]byte{} {
		t.Fatalf("genesis PrevHash: got %x, want all zeros", e.PrevHash)
	}
	if e.Version != FormatVersion || e.Seq != 0 || e.Type != EntryGrant {
		t.Fatalf("unexpected entry fields: %+v", e)
	}
	if e.Identity != "synthetic-approver-001" || e.Role != "approver" {
		t.Fatalf("unexpected identity/role: %q/%q", e.Identity, e.Role)
	}
	wantVF := time.Date(2026, 10, 21, 22, 0, 0, 0, time.UTC).UnixNano()
	if e.ValidFrom != wantVF {
		t.Fatalf("ValidFrom: got %d, want %d", e.ValidFrom, wantVF)
	}
}

func TestParseEntryRejectsUnknownType(t *testing.T) {
	for _, typ := range []string{"admin", "", "GRANT", "grantx"} {
		line := strings.Replace(testGrantLine(), `"type":"grant"`, `"type":"`+typ+`"`, 1)
		if _, err := ParseEntry([]byte(line)); err == nil {
			t.Fatalf("type %q: expected rejection, got nil error", typ)
		}
	}
}

func TestParseEntryRejectsVersion2(t *testing.T) {
	for _, v := range []string{`{"v":2`, `{"v":0`, `{"v":-1`} {
		line := v + testGrantLine()[len(`{"v":1`):]
		if _, err := ParseEntry([]byte(line)); err == nil {
			t.Fatalf("version in %q: expected rejection, got nil error", v)
		}
	}
}

func TestParseEntryGrantMissingIdentity(t *testing.T) {
	line := strings.Replace(testGrantLine(), `"identity":"synthetic-approver-001"`, `"identity":""`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("grant with empty Identity: expected rejection, got nil error")
	}
}

func TestParseEntryGrantMissingValidFrom(t *testing.T) {
	for _, variant := range []string{
		strings.Replace(testGrantLine(), `"valid_from":"2026-10-21T22:00:00Z"`, `"valid_from":null`, 1),
		strings.Replace(testGrantLine(), `"valid_from":"2026-10-21T22:00:00Z",`, ``, 1),
	} {
		if _, err := ParseEntry([]byte(variant)); err == nil {
			t.Fatal("grant missing valid_from: expected rejection, got nil error")
		}
	}
}

func TestParseEntryRevokeMissingRevokedAt(t *testing.T) {
	line := strings.Replace(testGrantLine(), `"type":"grant"`, `"type":"revoke"`, 1)
	line = strings.Replace(line, `"valid_from":"2026-10-21T22:00:00Z"`, `"valid_from":null`, 1)
	line = strings.Replace(line, `"revoked_at":null`, `"revoked_at":"2026-10-22T00:00:00Z"`, 1)
	// Drop the revoked_at timestamp entirely (null + absent both must be rejected).
	for _, variant := range []string{
		strings.Replace(line, `"revoked_at":"2026-10-22T00:00:00Z"`, `"revoked_at":null`, 1),
		strings.Replace(line, `"revoked_at":"2026-10-22T00:00:00Z",`, ``, 1),
	} {
		if _, err := ParseEntry([]byte(variant)); err == nil {
			t.Fatal("revoke missing revoked_at: expected rejection, got nil error")
		}
	}
}

func TestParseEntryGrantWithRevokedAtRejected(t *testing.T) {
	line := strings.Replace(testGrantLine(), `"revoked_at":null`, `"revoked_at":"2026-10-22T00:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("grant carrying revoked_at: expected rejection, got nil error")
	}
}

func TestParseEntryRevokeWithValidFromRejected(t *testing.T) {
	line := strings.Replace(testGrantLine(), `"type":"grant"`, `"type":"revoke"`, 1)
	line = strings.Replace(line, `"revoked_at":null`, `"revoked_at":"2026-10-22T00:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("revoke carrying valid_from: expected rejection, got nil error")
	}
	line2 := strings.Replace(line, `"valid_until":null`, `"valid_until":"2026-10-23T00:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line2)); err == nil {
		t.Fatal("revoke carrying valid_until: expected rejection, got nil error")
	}
}

func TestParseEntryUnknownFieldRejected(t *testing.T) {
	line := strings.Replace(testGrantLine(), `"sig":"`, `"extra":1,"sig":"`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("line with unknown field: expected rejection, got nil error")
	}
}

func TestParseEntryMalformedHexRejected(t *testing.T) {
	base := testGrantLine()
	zeros64 := strings.Repeat("0", 64)
	cases := map[string]string{
		"prev uppercase": strings.Replace(base, `"prev":"`+zeros64+`"`, `"prev":"`+strings.Repeat("A", 64)+`"`, 1),
		"prev too short": strings.Replace(base, `"prev":"`+zeros64+`"`, `"prev":"`+zeros64[:63]+`"`, 1),
		"prev too long":  strings.Replace(base, `"prev":"`+zeros64+`"`, `"prev":"`+zeros64+`0"`, 1),
		"prev non-hex":   strings.Replace(base, `"prev":"`+zeros64+`"`, `"prev":"`+strings.Repeat("z", 64)+`"`, 1),
		"sig uppercase":  strings.Replace(base, `"sig":"`+strings.Repeat("0", 128)+`"`, `"sig":"`+strings.Repeat("B", 128)+`"`, 1),
		"sig too short":  strings.Replace(base, `"sig":"`+strings.Repeat("0", 128)+`"`, `"sig":"`+strings.Repeat("0", 127)+`"`, 1),
		"sig non-hex":    strings.Replace(base, `"sig":"`+strings.Repeat("0", 128)+`"`, `"sig":"`+strings.Repeat("q", 128)+`"`, 1),
	}
	for name, line := range cases {
		if _, err := ParseEntry([]byte(line)); err == nil {
			t.Fatalf("%s: expected rejection, got nil error", name)
		}
	}
}

func TestEntryJSONLRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	_ = pub // the public key is injected by the caller in later tasks, never parsed
	grant := testGrantEntry()
	grant.Signature = ed25519.Sign(priv, grant.CanonicalBytes())

	line, err := MarshalEntry(grant)
	if err != nil {
		t.Fatalf("MarshalEntry: %v", err)
	}
	parsed, err := ParseEntry(line)
	if err != nil {
		t.Fatalf("ParseEntry: %v", err)
	}
	if !reflect.DeepEqual(grant, parsed) {
		t.Fatalf("round-trip mismatch:\n got %+v\nwant %+v", parsed, grant)
	}

	revoke := Entry{
		Version:    FormatVersion,
		Seq:        1,
		PrevHash:   ChainHash(grant.PrevHash, grant.CanonicalBytes()),
		Type:       EntryRevoke,
		Identity:   "synthetic-approver-001",
		Role:       "approver",
		RevokedAt:  time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC).UnixNano(),
		AppendedAt: time.Date(2026, 10, 22, 0, 0, 1, 0, time.UTC).UnixNano(),
	}
	revoke.Signature = ed25519.Sign(priv, revoke.CanonicalBytes())
	rline, err := MarshalEntry(revoke)
	if err != nil {
		t.Fatalf("MarshalEntry(revoke): %v", err)
	}
	rparsed, err := ParseEntry(rline)
	if err != nil {
		t.Fatalf("ParseEntry(revoke): %v", err)
	}
	if !reflect.DeepEqual(revoke, rparsed) {
		t.Fatalf("revoke round-trip mismatch:\n got %+v\nwant %+v", rparsed, revoke)
	}
	// Revoke valid_from must be null on the wire (round-trip preserves that).
	if strings.Contains(string(rline), `"valid_from":"`) {
		t.Fatalf("revoke line must not carry valid_from: %s", rline)
	}
}
