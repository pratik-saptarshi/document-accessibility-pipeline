// Tests for the approver-role register entry format (Task 1).
//
// All identities and keys are synthetic. No real person names, no real keys.
package approver

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
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
		ValidFrom:  time.Date(2026, 10, 21, 22, 0, 1, 0, time.UTC).UnixNano(),
		ValidUntil: time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC).UnixNano(),
		AppendedAt: time.Date(2026, 10, 21, 22, 0, 1, 0, time.UTC).UnixNano(),
	}
}

// testGrantLine returns a deterministic synthetic grant JSONL line for parse tests.
func testGrantLine() string {
	return `{"v":1,"seq":0,"prev":"` + strings.Repeat("0", 64) + `",` +
		`"type":"grant","identity":"synthetic-approver-001","role":"approver",` +
		`"valid_from":"2026-10-21T22:00:01Z","valid_until":"2027-12-31T23:59:59Z",` +
		`"revokes_sequence":null,"effective_at":null,` +
		`"appended_at":"2026-10-21T22:00:01Z","sig":"` + strings.Repeat("0", 128) + `"}`
}

// testRevokeEntry returns a deterministic synthetic revoke entry naming
// grant sequence 0. Revoke entries carry no identity, role, or validity
// window per Decision 0015.
func testRevokeEntry() Entry {
	return Entry{
		Version:         FormatVersion,
		Seq:             1,
		PrevHash:        [32]byte{},
		Type:            EntryRevoke,
		RevokesSequence: 0,
		EffectiveAt:     time.Date(2026, 10, 22, 0, 0, 1, 0, time.UTC).UnixNano(),
		AppendedAt:      time.Date(2026, 10, 22, 0, 0, 1, 0, time.UTC).UnixNano(),
	}
}

// testRevokeLine returns a deterministic synthetic revoke JSONL line for parse tests.
func testRevokeLine() string {
	return `{"v":1,"seq":1,"prev":"` + strings.Repeat("0", 64) + `",` +
		`"type":"revoke","identity":"","role":"",` +
		`"valid_from":null,"valid_until":null,` +
		`"revokes_sequence":0,"effective_at":"2026-10-22T00:00:01Z",` +
		`"appended_at":"2026-10-22T00:00:01Z","sig":"` + strings.Repeat("0", 128) + `"}`
}

func TestCanonicalBytesDeterministic(t *testing.T) {
	e := testGrantEntry()
	a, b := e.CanonicalBytes(), e.CanonicalBytes()
	if !bytes.Equal(a, b) {
		t.Fatal("same entry produced different canonical bytes")
	}
}

func TestCanonicalBytesFieldOrderSensitivity(t *testing.T) {
	// Fixed order for grants: Version(1B), Seq(8B BE), PrevHash(32B), Type(1B),
	// Identity(uvarint+bytes), Role(uvarint+bytes), ValidFrom(8B BE),
	// ValidUntil(8B BE), AppendedAt(8B BE). The old RevokedAt field is gone.
	e := Entry{Version: FormatVersion, Seq: 0x0102030405060708, Type: EntryGrant}
	cb := e.CanonicalBytes()
	if cb[0] != 0x01 {
		t.Fatalf("byte 0: got %#02x, want Version 0x01", cb[0])
	}
	if want := []byte{1, 2, 3, 4, 5, 6, 7, 8}; !bytes.Equal(cb[1:9], want) {
		t.Fatalf("bytes 1..8: got %x, want big-endian Seq %x", cb[1:9], want)
	}
	if want := 1 + 8 + 32 + 1 + 1 + 1 + 8 + 8 + 8; len(cb) != want {
		t.Fatalf("empty Identity/Role grant entry: got %d bytes, want %d", len(cb), want)
	}
	// Identity and Role occupy ordered, distinct slots: swapping them changes bytes.
	a := testGrantEntry()
	a.Identity, a.Role = "aaa", "bbb"
	b := testGrantEntry()
	b.Identity, b.Role = "bbb", "aaa"
	if bytes.Equal(a.CanonicalBytes(), b.CanonicalBytes()) {
		t.Fatal("swapping Identity/Role must change canonical bytes")
	}
	// Grant canonical bytes must not be influenced by the revoke-only fields.
	c := testGrantEntry()
	c.RevokesSequence, c.EffectiveAt = 42, 123456789
	if !bytes.Equal(c.CanonicalBytes(), testGrantEntry().CanonicalBytes()) {
		t.Fatal("revoke-only fields must not affect grant canonical bytes")
	}
}

func TestCanonicalBytesRevokeLayout(t *testing.T) {
	// Fixed order for revokes: Version(1B), Seq(8B BE), PrevHash(32B), Type(1B),
	// RevokesSequence(8B BE), EffectiveAt(8B BE), AppendedAt(8B BE).
	// No identity, role, or validity window bytes.
	e := Entry{
		Version:         FormatVersion,
		Seq:             0x0102030405060708,
		Type:            EntryRevoke,
		RevokesSequence: 0x1112131415161718,
		EffectiveAt:     0x2122232425262728,
		AppendedAt:      0x3132333435363738,
	}
	cb := e.CanonicalBytes()
	if want := 1 + 8 + 32 + 1 + 8 + 8 + 8; len(cb) != want {
		t.Fatalf("revoke entry: got %d bytes, want %d", len(cb), want)
	}
	if cb[0] != 0x01 {
		t.Fatalf("byte 0: got %#02x, want Version 0x01", cb[0])
	}
	if want := []byte{1, 2, 3, 4, 5, 6, 7, 8}; !bytes.Equal(cb[1:9], want) {
		t.Fatalf("bytes 1..8: got %x, want big-endian Seq %x", cb[1:9], want)
	}
	if want := [32]byte{}; !bytes.Equal(cb[9:41], want[:]) {
		t.Fatalf("bytes 9..40: got %x, want zero PrevHash", cb[9:41])
	}
	if cb[41] != EntryRevoke {
		t.Fatalf("byte 41: got %#02x, want Type 0x02", cb[41])
	}
	if want := []byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}; !bytes.Equal(cb[42:50], want) {
		t.Fatalf("bytes 42..49: got %x, want big-endian RevokesSequence %x", cb[42:50], want)
	}
	if want := []byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28}; !bytes.Equal(cb[50:58], want) {
		t.Fatalf("bytes 50..57: got %x, want big-endian EffectiveAt %x", cb[50:58], want)
	}
	if want := []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}; !bytes.Equal(cb[58:66], want) {
		t.Fatalf("bytes 58..65: got %x, want big-endian AppendedAt %x", cb[58:66], want)
	}
	// Revoke canonical bytes must not be influenced by identity/role text.
	f := e
	f.Identity, f.Role = "synthetic-approver-001", "approver"
	if !bytes.Equal(f.CanonicalBytes(), e.CanonicalBytes()) {
		t.Fatal("identity/role must not affect revoke canonical bytes")
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
	wantVF := time.Date(2026, 10, 21, 22, 0, 1, 0, time.UTC).UnixNano()
	if e.ValidFrom != wantVF {
		t.Fatalf("ValidFrom: got %d, want %d", e.ValidFrom, wantVF)
	}
	wantVU := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC).UnixNano()
	if e.ValidUntil != wantVU {
		t.Fatalf("ValidUntil: got %d, want %d", e.ValidUntil, wantVU)
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
		strings.Replace(testGrantLine(), `"valid_from":"2026-10-21T22:00:01Z"`, `"valid_from":null`, 1),
		strings.Replace(testGrantLine(), `"valid_from":"2026-10-21T22:00:01Z",`, ``, 1),
	} {
		if _, err := ParseEntry([]byte(variant)); err == nil {
			t.Fatal("grant missing valid_from: expected rejection, got nil error")
		}
	}
}

func TestParseEntryRevokeMissingRevokesSequenceOrEffectiveAt(t *testing.T) {
	base := testRevokeLine()
	// Absent and explicit-null both must be rejected for each required field.
	variants := []string{
		strings.Replace(base, `"revokes_sequence":0,`, `"revokes_sequence":null,`, 1),
		strings.Replace(base, `"revokes_sequence":0,`, ``, 1),
		strings.Replace(base, `"effective_at":"2026-10-22T00:00:01Z"`, `"effective_at":null`, 1),
		strings.Replace(base, `"effective_at":"2026-10-22T00:00:01Z",`, ``, 1),
	}
	for i, variant := range variants {
		if _, err := ParseEntry([]byte(variant)); err == nil {
			t.Fatalf("variant %d: revoke missing revokes_sequence/effective_at: expected rejection, got nil error", i)
		}
	}
	// A well-formed revoke line parses.
	e, err := ParseEntry([]byte(base))
	if err != nil {
		t.Fatalf("ParseEntry(revoke): %v", err)
	}
	if e.Type != EntryRevoke || e.RevokesSequence != 0 {
		t.Fatalf("unexpected revoke fields: %+v", e)
	}
	wantEff := time.Date(2026, 10, 22, 0, 0, 1, 0, time.UTC).UnixNano()
	if e.EffectiveAt != wantEff {
		t.Fatalf("EffectiveAt: got %d, want %d", e.EffectiveAt, wantEff)
	}
	if e.Identity != "" || e.Role != "" {
		t.Fatalf("revoke identity/role must be empty: %q/%q", e.Identity, e.Role)
	}
}

func TestParseEntryRevokeWithIdentityOrRoleRejected(t *testing.T) {
	base := testRevokeLine()
	for name, line := range map[string]string{
		"identity": strings.Replace(base, `"identity":""`, `"identity":"synthetic-approver-001"`, 1),
		"role":     strings.Replace(base, `"role":""`, `"role":"approver"`, 1),
	} {
		if _, err := ParseEntry([]byte(line)); err == nil {
			t.Fatalf("revoke carrying %s: expected rejection, got nil error", name)
		}
	}
}

func TestParseEntryGrantWithRevokeFieldsRejected(t *testing.T) {
	base := testGrantLine()
	for name, line := range map[string]string{
		"revokes_sequence": strings.Replace(base, `"revokes_sequence":null`, `"revokes_sequence":0`, 1),
		"effective_at":     strings.Replace(base, `"effective_at":null`, `"effective_at":"2026-10-22T00:00:00Z"`, 1),
	} {
		if _, err := ParseEntry([]byte(line)); err == nil {
			t.Fatalf("grant carrying %s: expected rejection, got nil error", name)
		}
	}
}

func TestParseEntryRevokeWithValidFromRejected(t *testing.T) {
	base := testRevokeLine()
	line := strings.Replace(base, `"valid_from":null`, `"valid_from":"2026-10-21T22:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("revoke carrying valid_from: expected rejection, got nil error")
	}
	line2 := strings.Replace(base, `"valid_until":null`, `"valid_until":"2026-10-23T00:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line2)); err == nil {
		t.Fatal("revoke carrying valid_until: expected rejection, got nil error")
	}
}

func TestParseEntryRevokeWithLegacyRevokedAtRejected(t *testing.T) {
	// The old revoked_at wire field is gone: it must be rejected as unknown,
	// not silently ignored.
	line := strings.Replace(testRevokeLine(), `"revokes_sequence":0`, `"revokes_sequence":0,"revoked_at":"2026-10-22T00:00:00Z"`, 1)
	if _, err := ParseEntry([]byte(line)); err == nil {
		t.Fatal("revoke carrying revoked_at: expected rejection, got nil error")
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

	revoke := testRevokeEntry()
	revoke.PrevHash = ChainHash(grant.PrevHash, grant.CanonicalBytes())
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
	// Revoke valid_from/valid_until must be null on the wire; the revoke
	// carries revokes_sequence and effective_at instead of revoked_at.
	for _, want := range []string{`"revokes_sequence":0`, `"effective_at":"2026-10-22T00:00:01Z"`, `"valid_from":null`} {
		if !strings.Contains(string(rline), want) {
			t.Fatalf("revoke line must contain %s: %s", want, rline)
		}
	}
	for _, forbid := range []string{`"valid_from":"`, `"valid_until":"`, "revoked_at"} {
		if strings.Contains(string(rline), forbid) {
			t.Fatalf("revoke line must not contain %s: %s", forbid, rline)
		}
	}
}

func TestMarshalEntryRevokeValidation(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	_ = pub
	sign := func(e Entry) Entry {
		e.Signature = ed25519.Sign(priv, e.CanonicalBytes())
		return e
	}
	// Zero effective_at is a missing required field: reject.
	if _, err := MarshalEntry(sign(testRevokeEntryWithoutEffective())); err == nil {
		t.Fatal("revoke with zero effective_at: expected rejection, got nil error")
	}
	// Revokes must not carry identity/role: reject.
	e := sign(testRevokeEntry())
	e.Identity, e.Role = "synthetic-approver-001", "approver"
	if _, err := MarshalEntry(e); err == nil {
		t.Fatal("revoke with identity/role: expected rejection, got nil error")
	}
}

// testRevokeEntryWithoutEffective returns a revoke missing effective_at.
func testRevokeEntryWithoutEffective() Entry {
	e := testRevokeEntry()
	e.EffectiveAt = 0
	return e
}

// ---------------------------------------------------------------------------
// MarshalEntry must fail loud on revoke-only fields carried by a grant,
// symmetric with ParseEntry which rejects those fields on grant lines.
// ---------------------------------------------------------------------------

func TestMarshalEntryGrantWithRevokeFieldsRejected(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sign := func(e Entry) Entry {
		e.Signature = ed25519.Sign(priv, e.CanonicalBytes())
		return e
	}
	// A grant carrying a revoke-only field must be rejected, not silently
	// dropped.
	g := sign(testGrantEntry())
	g.RevokesSequence = 3
	if _, err := MarshalEntry(g); err == nil {
		t.Fatal("grant with RevokesSequence set: expected rejection, got nil error")
	} else if !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with RevokesSequence set: error %v does not wrap ErrMalformed", err)
	}
	g2 := sign(testGrantEntry())
	g2.EffectiveAt = g2.ValidFrom
	if _, err := MarshalEntry(g2); err == nil {
		t.Fatal("grant with EffectiveAt set: expected rejection, got nil error")
	} else if !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with EffectiveAt set: error %v does not wrap ErrMalformed", err)
	}
	// Revoke entries still marshal normally.
	revoke := sign(testRevokeEntry())
	if _, err := MarshalEntry(revoke); err != nil {
		t.Fatalf("well-formed revoke: unexpected rejection: %v", err)
	}
}
