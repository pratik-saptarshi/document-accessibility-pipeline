// Tests for the approver-role verification adapter (Task 2).
//
// All identities and keys are synthetic. No real person names, no real keys.
// Times are UTC throughout.
package approver

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"
)

// baseTime anchors the synthetic chain timestamps.
var baseTime = time.Date(2026, 10, 21, 22, 0, 0, 0, time.UTC)

// chainBuilder assembles a valid, internally consistent chain in-test: it
// assigns seq 0,1,2,..., links PrevHash via ChainHash, and signs every entry
// with the test private key. MarshalEntry (Task 1) is used for JSONL lines
// where needed.
type chainBuilder struct {
	t        *testing.T
	priv     ed25519.PrivateKey
	entries  []Entry
	prev     [32]byte
	nextSeq  uint64
	appended time.Time
}

func newChainBuilder(t *testing.T, priv ed25519.PrivateKey) *chainBuilder {
	t.Helper()
	return &chainBuilder{t: t, priv: priv, appended: baseTime}
}

func (b *chainBuilder) append(e Entry) {
	e.Version = FormatVersion
	e.Seq = b.nextSeq
	e.PrevHash = b.prev
	e.Signature = ed25519.Sign(b.priv, e.CanonicalBytes())
	b.entries = append(b.entries, e)
	b.prev = ChainHash(e.PrevHash, e.CanonicalBytes())
	b.nextSeq++
}

func (b *chainBuilder) addGrant(identity, role string, validFrom, validUntil, appendedAt time.Time) {
	// Decision 0015: not_after is mandatory on every grant — no open-ended
	// grants. The helper always sets ValidUntil; callers must pass a
	// concrete time.
	e := Entry{
		Type:       EntryGrant,
		Identity:   identity,
		Role:       role,
		ValidFrom:  validFrom.UTC().UnixNano(),
		ValidUntil: validUntil.UTC().UnixNano(),
		AppendedAt: appendedAt.UTC().UnixNano(),
	}
	b.append(e)
}

func (b *chainBuilder) addRevoke(revokesSeq uint64, effectiveAt, appendedAt time.Time) {
	e := Entry{
		Type:            EntryRevoke,
		RevokesSequence: revokesSeq,
		EffectiveAt:     effectiveAt.UTC().UnixNano(),
		AppendedAt:      appendedAt.UTC().UnixNano(),
	}
	b.append(e)
}

func (b *chainBuilder) snapshot(pub ed25519.PublicKey) Snapshot {
	entries := make([]Entry, len(b.entries))
	copy(entries, b.entries)
	return Snapshot{Entries: entries, PubKey: pub}
}

func (b *chainBuilder) head() [32]byte { return b.prev }

func mustTestKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	return pub, priv
}

// validFixture builds the canonical three-entry chain used by many tests:
// grant(seq0, approver-001) -> grant(seq1, approver-002) ->
// revoke(seq2 naming seq1, effective T0+2h). The returned head is the chain
// hash of the last entry.
func validFixture(t *testing.T) (*chainBuilder, ed25519.PublicKey, [32]byte) {
	t.Helper()
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	b.addGrant("synthetic-approver-002", "approver", baseTime.Add(time.Hour), baseTime.Add(48*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(1, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	return b, pub, b.head()
}

func assertErrIs(t *testing.T, err, target error, name string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("%s: got error %v, want errors.Is(%v)", name, err, target)
	}
}

// ---------------------------------------------------------------------------
// Positive cases
// ---------------------------------------------------------------------------

func TestVerifyRoleAtValidGrantInsideWindow(t *testing.T) {
	b, pub, head := validFixture(t)
	at := baseTime.Add(3 * time.Hour) // inside 24h window of approver-001's grant
	if err := VerifyRoleAt("synthetic-approver-001", "approver", at, b.snapshot(pub), head); err != nil {
		t.Fatalf("valid grant inside window: %v", err)
	}
	// Long-lived grant (approver-002's grant was revoked, so use a fresh chain).
	pub2, priv2 := mustTestKeys(t)
	b2 := newChainBuilder(t, priv2)
	b2.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	head2 := b2.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(720*time.Hour), b2.snapshot(pub2), head2); err != nil {
		t.Fatalf("long-lived grant far in the future: %v", err)
	}
}

func TestVerifyRoleAtRegrantAfterRevocationVerifies(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addRevoke(0, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	b.addGrant("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), baseTime.Add(1000*time.Hour), baseTime.Add(3*time.Hour))
	head := b.head()
	// Query after the re-grant's ValidFrom: the revoke names grant seq 0 only,
	// so the re-grant (seq 2) stands.
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(4*time.Hour), b.snapshot(pub), head); err != nil {
		t.Fatalf("query after re-grant ValidFrom: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Adversarial cases
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Decision 0015: not_before <= reviewed_at <= not_after, closed at both
// boundaries. A query at exactly ValidUntil still authorizes; one
// nanosecond later the grant is expired.
// ---------------------------------------------------------------------------

func TestVerifyRoleAtExpiryBoundaryInclusive(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	end := baseTime.Add(24 * time.Hour)
	b.addGrant("synthetic-approver-001", "approver", baseTime, end, baseTime)
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver", end, b.snapshot(pub), head); err != nil {
		t.Fatalf("query at exactly ValidUntil: got %v, want nil (inclusive boundary)", err)
	}
	err := VerifyRoleAt("synthetic-approver-001", "approver", end.Add(time.Nanosecond), b.snapshot(pub), head)
	assertErrIs(t, err, ErrExpired, "query one nanosecond after ValidUntil")
}

func TestVerifyRoleAtOpenEndedGrantRejected(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	// Hand-built programmatic Snapshot: the wire format cannot produce
	// ValidUntil == 0 (Fix 4 rejects it at parse), so construct it
	// directly. Must fail closed, not authorize indefinitely.
	e := Entry{
		Type:       EntryGrant,
		Identity:   "synthetic-approver-001",
		Role:       "approver",
		ValidFrom:  baseTime.UTC().UnixNano(),
		ValidUntil: 0,
		AppendedAt: baseTime.UTC().UnixNano(),
	}
	b.append(e)
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrMalformed, "open-ended grant (ValidUntil == 0)")
}

func TestVerifyRoleAtPreWindowQueryNotYetValid(t *testing.T) {
	b, pub, head := validFixture(t)
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(-time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrNotYetValid, "pre-window query")
}

func TestVerifyRoleAtPostExpiry(t *testing.T) {
	b, pub, head := validFixture(t)
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(25*time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrExpired, "post-expiry query")
}

func TestVerifyRoleAtRevokedNoRegrant(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addRevoke(0, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrRevoked, "post-revocation query with no re-grant")
}

func TestVerifyRoleAtQueryBetweenRevokeAndRegrantDenied(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addRevoke(0, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	b.addGrant("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), baseTime.Add(1000*time.Hour), baseTime.Add(3*time.Hour))
	head := b.head()
	// Between revoke and re-grant there is no covering grant: the latest
	// grant with ValidFrom <= at is seq 0, which the revoke names.
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(2*time.Hour+30*time.Minute), b.snapshot(pub), head)
	if err == nil {
		t.Fatal("query between revoke and re-grant: expected deny, got nil")
	}
	assertErrIs(t, err, ErrRevoked, "query between revoke and re-grant")
}

// ---------------------------------------------------------------------------
// Fix 9: overlapping grants — authorize via ANY covering grant (Decision
// 0015: "A grant authorizes a review iff not_before <= reviewed_at <=
// not_after"). The old code selected only the latest grant by ValidFrom, so
// with grants [T0,T10] + [T5,T6] a query at T7 denied (ErrExpired) even
// though [T0,T10] covers T7.
// ---------------------------------------------------------------------------

// overlappingFixture builds: grant seq0 ("ov1"/"approver", T0..T0+10h),
// grant seq1 ("ov1"/"approver", T0+5h..T0+6h). Under the old latest-only
// rule, a query at T0+7h picked seq1 and denied ErrExpired; the correct
// verdict is nil (seq0 covers T0+7h).
func overlappingFixture(t *testing.T, priv ed25519.PrivateKey) *chainBuilder {
	t.Helper()
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-ov1", "approver", baseTime, baseTime.Add(10*time.Hour), baseTime)
	b.addGrant("synthetic-approver-ov1", "approver", baseTime.Add(5*time.Hour), baseTime.Add(6*time.Hour), baseTime.Add(time.Hour))
	return b
}

func TestVerifyRoleAtOverlappingGrantsAuthorizeViaEarlier(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := overlappingFixture(t, priv)
	head := b.head()
	// T0+7h: inside grant A [T0,T10] only. A clean covering grant exists →
	// authorize, even though the later grant B [T5,T6] has expired.
	if err := VerifyRoleAt("synthetic-approver-ov1", "approver", baseTime.Add(7*time.Hour), b.snapshot(pub), head); err != nil {
		t.Fatalf("query inside earlier grant A only: got %v, want nil", err)
	}
}

func TestVerifyRoleAtOverlappingCoveredGrantRevokedOtherClean(t *testing.T) {
	// Grant A [T0,T10] revoked effective T0+5h; grant B [T5,T6] clean.
	// Query at T0+5.5h: A covers but is revoked, B covers and is clean →
	// authorize (a revocation names one grant by sequence, not the
	// identity/role).
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-ov2", "approver", baseTime, baseTime.Add(10*time.Hour), baseTime)
	b.addGrant("synthetic-approver-ov2", "approver", baseTime.Add(5*time.Hour), baseTime.Add(6*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(0, baseTime.Add(5*time.Hour), baseTime.Add(5*time.Hour))
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-ov2", "approver", baseTime.Add(5*time.Hour+30*time.Minute), b.snapshot(pub), head); err != nil {
		t.Fatalf("query where clean covering grant B exists: got %v, want nil", err)
	}
}

func TestVerifyRoleAtOverlappingAllCoveringRevoked(t *testing.T) {
	// Grants A [T0,T10] and B [T5,T6] both revoked before the query.
	// Query at T0+5.5h: covered by both, both revoked → ErrRevoked
	// (denial precedence: revoked > expired > not-yet-valid).
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-ov3", "approver", baseTime, baseTime.Add(10*time.Hour), baseTime)
	b.addGrant("synthetic-approver-ov3", "approver", baseTime.Add(5*time.Hour), baseTime.Add(6*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(0, baseTime.Add(4*time.Hour), baseTime.Add(4*time.Hour))
	b.addRevoke(1, baseTime.Add(4*time.Hour), baseTime.Add(4*time.Hour))
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-ov3", "approver", baseTime.Add(5*time.Hour+30*time.Minute), b.snapshot(pub), head)
	assertErrIs(t, err, ErrRevoked, "all covering grants revoked")
}

func TestVerifyRoleAtDenialPrecedenceRevokedBeatsExpired(t *testing.T) {
	// Grant A [T0,T10] revoked effective T0+5h; grant B [T5,T6] expired.
	// Query at T0+7h: A covers-but-revoked, B expired → ErrRevoked wins.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-ov4", "approver", baseTime, baseTime.Add(10*time.Hour), baseTime)
	b.addGrant("synthetic-approver-ov4", "approver", baseTime.Add(5*time.Hour), baseTime.Add(6*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(0, baseTime.Add(5*time.Hour), baseTime.Add(5*time.Hour))
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-ov4", "approver", baseTime.Add(7*time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrRevoked, "revoked-beats-expired precedence")
}

func TestVerifyRoleAtDenialPrecedenceExpiredBeatsNotYetValid(t *testing.T) {
	// Grant A [T0,T1] expired; grant B [T5,T6] starts after the query.
	// Query at T0+3h: no covering grant, nothing revoked → ErrExpired wins.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-ov5", "approver", baseTime, baseTime.Add(time.Hour), baseTime)
	b.addGrant("synthetic-approver-ov5", "approver", baseTime.Add(5*time.Hour), baseTime.Add(6*time.Hour), baseTime.Add(2*time.Hour))
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-ov5", "approver", baseTime.Add(3*time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrExpired, "expired-beats-not-yet-valid precedence")
}

func TestVerifyRoleAtOverlappingAllExpired(t *testing.T) {
	// Grants A [T0,T10] and B [T5,T6], query after both windows → ErrExpired.
	pub, priv := mustTestKeys(t)
	b := overlappingFixture(t, priv)
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-ov1", "approver", baseTime.Add(11*time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrExpired, "all overlapping grants expired")
}

// gapChain builds entries [seq0, seq2] whose PrevHash links are consistent:
// only the sequence numbering is broken, isolating ErrGap.
func gapChain(t *testing.T, priv ed25519.PrivateKey) []Entry {
	t.Helper()
	e0 := Entry{Version: FormatVersion, Seq: 0, PrevHash: [32]byte{}, Type: EntryGrant,
		Identity: "synthetic-approver-001", Role: "approver",
		ValidFrom: baseTime.UTC().UnixNano(), ValidUntil: baseTime.Add(1000 * time.Hour).UTC().UnixNano(),
		AppendedAt: baseTime.UTC().UnixNano()}
	e0.Signature = ed25519.Sign(priv, e0.CanonicalBytes())
	e1 := Entry{Version: FormatVersion, Seq: 2, PrevHash: ChainHash(e0.PrevHash, e0.CanonicalBytes()), Type: EntryGrant,
		Identity: "synthetic-approver-001", Role: "approver",
		ValidFrom: baseTime.Add(time.Hour).UTC().UnixNano(), ValidUntil: baseTime.Add(1000 * time.Hour).UTC().UnixNano(),
		AppendedAt: baseTime.Add(time.Hour).UTC().UnixNano()}
	e1.Signature = ed25519.Sign(priv, e1.CanonicalBytes())
	return []Entry{e0, e1}
}

func TestVerifyRoleAtSequenceGap(t *testing.T) {
	pub, priv := mustTestKeys(t)
	entries := gapChain(t, priv)
	last := entries[len(entries)-1]
	head := ChainHash(last.PrevHash, last.CanonicalBytes())
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), Snapshot{Entries: entries, PubKey: pub}, head)
	assertErrIs(t, err, ErrGap, "sequence gap")
}

// dupChain builds entries [seq0, seq0] whose PrevHash links are consistent:
// only the duplicate sequence number is broken, isolating ErrGap.
func dupChain(t *testing.T, priv ed25519.PrivateKey) []Entry {
	t.Helper()
	entries := gapChain(t, priv)
	entries[1].Seq = 0
	entries[1].Signature = ed25519.Sign(priv, entries[1].CanonicalBytes())
	return entries
}

func TestVerifyRoleAtDuplicateSeq(t *testing.T) {
	pub, priv := mustTestKeys(t)
	entries := dupChain(t, priv)
	last := entries[len(entries)-1]
	head := ChainHash(last.PrevHash, last.CanonicalBytes())
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), Snapshot{Entries: entries, PubKey: pub}, head)
	assertErrIs(t, err, ErrGap, "duplicate seq")
}

func TestVerifyRoleAtTamperedPayloadStaleSignature(t *testing.T) {
	b, pub, head := validFixture(t)
	snap := b.snapshot(pub)
	// Tamper the revoke's target sequence without re-signing.
	snap.Entries[2].RevokesSequence = 42
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), snap, head)
	assertErrIs(t, err, ErrBadSignature, "tampered payload, stale signature")
}

func TestVerifyRoleAtTamperedPayloadAttackerKey(t *testing.T) {
	b, pub, head := validFixture(t)
	_, attackerPriv := mustTestKeys(t)
	snap := b.snapshot(pub)
	// Attacker tampers the payload and re-signs with their own key. The
	// injected key is the honest one, so verification must still fail.
	snap.Entries[2].RevokesSequence = 42
	snap.Entries[2].Signature = ed25519.Sign(attackerPriv, snap.Entries[2].CanonicalBytes())
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), snap, head)
	assertErrIs(t, err, ErrBadSignature, "tampered payload, attacker-key signature")
}

func TestVerifyRoleAtReorderedEntries(t *testing.T) {
	b, pub, _ := validFixture(t)
	snap := b.snapshot(pub)
	// Whole entries reordered: signatures remain valid, but the PrevHash
	// links no longer chain in order.
	snap.Entries[1], snap.Entries[2] = snap.Entries[2], snap.Entries[1]
	last := snap.Entries[len(snap.Entries)-1]
	head := ChainHash(last.PrevHash, last.CanonicalBytes())
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), snap, head)
	assertErrIs(t, err, ErrInvalidChain, "reordered entries")
}

func TestVerifyRoleAtWrongPinnedHead(t *testing.T) {
	b, pub, _ := validFixture(t)
	var wrong [32]byte
	if _, err := rand.Read(wrong[:]); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), b.snapshot(pub), wrong)
	assertErrIs(t, err, ErrHeadMismatch, "wrong pinned head")
}

func TestVerifyRoleAtTruncatedRegisterTrueHead(t *testing.T) {
	b, pub, fullHead := validFixture(t)
	// Truncation attack: drop the tail entries (which include the revoke of
	// approver-002) and pin the TRUE head of the full register. The terminal
	// hash of the truncated snapshot cannot match.
	truncated := b.entries[:2]
	snap := Snapshot{Entries: append([]Entry(nil), truncated...), PubKey: pub}
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), snap, fullHead)
	assertErrIs(t, err, ErrHeadMismatch, "truncated register with true head pinned")
	// Control: the same two entries with their own head must verify.
	last := snap.Entries[len(snap.Entries)-1]
	ownHead := ChainHash(last.PrevHash, last.CanonicalBytes())
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), snap, ownHead); err != nil {
		t.Fatalf("truncated control with matching head: %v", err)
	}
}

func TestVerifyRoleAtWrongKeySignatures(t *testing.T) {
	pub, _ := mustTestKeys(t) // honest key, injected below
	_, attackerPriv := mustTestKeys(t)
	b := newChainBuilder(t, attackerPriv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	head := b.head()
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(time.Hour), b.snapshot(pub), head)
	assertErrIs(t, err, ErrBadSignature, "wrong-key signatures")
}

// Empty snapshots carry no grant. This test pins the documented choice:
// ErrNoGrant (not ErrMalformed) for len(Entries)==0.
func TestVerifyRoleAtEmptyEntries(t *testing.T) {
	pub, _ := mustTestKeys(t)
	for name, entries := range map[string][]Entry{"nil": nil, "empty": {}} {
		err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime, Snapshot{Entries: entries, PubKey: pub}, [32]byte{})
		assertErrIs(t, err, ErrNoGrant, "empty entries ("+name+")")
	}
}

// ---------------------------------------------------------------------------
// Revocation model per Decision 0015: a revocation names exactly one grant
// by revokes_sequence and applies prospectively from effective_at, inclusive.
// ---------------------------------------------------------------------------

func TestVerifyRoleAtRevokeNamesOneGrantBySequence(t *testing.T) {
	// Two grants for the SAME identity/role; the revoke names only the first
	// grant's sequence. The second grant is untouched by it.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addGrant("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), baseTime.Add(1000*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(0, baseTime.Add(2*time.Hour), baseTime.Add(time.Hour))
	head := b.head()
	// Query while only grant seq 0 covers the identity: the revoke names seq 0
	// and its effective_at has passed → revoked.
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(150*time.Minute), b.snapshot(pub), head)
	assertErrIs(t, err, ErrRevoked, "revoke names grant seq 0")
	// Query after grant seq 1's ValidFrom: the latest covering grant is seq 1,
	// which no revoke names → allowed.
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(4*time.Hour), b.snapshot(pub), head); err != nil {
		t.Fatalf("revoke of seq 0 must not affect grant seq 1: %v", err)
	}
}

func TestVerifyRoleAtRevokeEffectiveAtBoundary(t *testing.T) {
	// A revocation applies prospectively from effective_at, inclusive:
	// queries with at < effective_at are unaffected.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	eff := baseTime.Add(2 * time.Hour)
	b.addRevoke(0, eff, eff)
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver", eff.Add(-time.Nanosecond), b.snapshot(pub), head); err != nil {
		t.Fatalf("query before effective_at must not be revoked: %v", err)
	}
	err := VerifyRoleAt("synthetic-approver-001", "approver", eff, b.snapshot(pub), head)
	assertErrIs(t, err, ErrRevoked, "query at effective_at")
}

func TestVerifyRoleAtRevokeUnknownSeqIsInert(t *testing.T) {
	// A revoke naming a sequence that is not a grant is inert: the chain
	// still verifies end to end and the grant is unaffected.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addRevoke(999, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), b.snapshot(pub), head); err != nil {
		t.Fatalf("revoke for a nonexistent seq must be inert: %v", err)
	}
}

func TestVerifyRoleAtAppendedAtBackward(t *testing.T) {
	pub, priv := mustTestKeys(t)
	e0 := Entry{Version: FormatVersion, Seq: 0, PrevHash: [32]byte{}, Type: EntryGrant,
		Identity: "synthetic-approver-001", Role: "approver",
		ValidFrom: baseTime.UTC().UnixNano(), ValidUntil: baseTime.Add(1000 * time.Hour).UTC().UnixNano(),
		AppendedAt: baseTime.UTC().UnixNano()}
	e0.Signature = ed25519.Sign(priv, e0.CanonicalBytes())
	e1 := Entry{Version: FormatVersion, Seq: 1, PrevHash: ChainHash(e0.PrevHash, e0.CanonicalBytes()), Type: EntryGrant,
		Identity: "synthetic-approver-001", Role: "approver",
		ValidFrom: baseTime.Add(2 * time.Hour).UTC().UnixNano(), ValidUntil: baseTime.Add(1000 * time.Hour).UTC().UnixNano(),
		AppendedAt: baseTime.Add(-time.Hour).UTC().UnixNano()} // backward (and still >= ValidFrom)
	e1.Signature = ed25519.Sign(priv, e1.CanonicalBytes())
	entries := []Entry{e0, e1}
	head := ChainHash(e1.PrevHash, e1.CanonicalBytes())
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour), Snapshot{Entries: entries, PubKey: pub}, head)
	assertErrIs(t, err, ErrInvalidChain, "AppendedAt going backward")
}

func TestVerifyRoleAtIdentityRoleExactMatch(t *testing.T) {
	b, pub, head := validFixture(t)
	snap := b.snapshot(pub)
	at := baseTime.Add(3 * time.Hour)
	// Identity/Role comparison is exact byte equality: no case folding.
	for _, tc := range []struct{ identity, role string }{
		{"synthetic-approver-001", "Approver"},
		{"SYNTHETIC-APPROVER-001", "approver"},
		{"synthetic-approver-001 ", "approver"},
		{"synthetic-approver-001", "approver "},
		{"unknown-identity", "approver"},
	} {
		if err := VerifyRoleAt(tc.identity, tc.role, at, snap, head); err == nil {
			t.Fatalf("identity %q role %q: expected deny, got nil", tc.identity, tc.role)
		}
	}
}

// ---------------------------------------------------------------------------
// Malformed input reaches the adapter through ParseEntry (Task 1 boundary):
// structural failures are ErrMalformed before verification ever runs.
// ---------------------------------------------------------------------------

func mustSignedLine(t *testing.T, priv ed25519.PrivateKey) string {
	t.Helper()
	e := Entry{Version: FormatVersion, Seq: 0, PrevHash: [32]byte{}, Type: EntryGrant,
		Identity: "synthetic-approver-001", Role: "approver",
		ValidFrom: baseTime.UTC().UnixNano(), ValidUntil: baseTime.Add(24 * time.Hour).UTC().UnixNano(),
		AppendedAt: baseTime.UTC().UnixNano()}
	e.Signature = ed25519.Sign(priv, e.CanonicalBytes())
	line, err := MarshalEntry(e)
	if err != nil {
		t.Fatalf("MarshalEntry: %v", err)
	}
	return string(line)
}

func assertParseMalformed(t *testing.T, line, name string) {
	t.Helper()
	_, err := ParseEntry([]byte(line))
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: got error %v, want errors.Is(ErrMalformed)", name, err)
	}
}

func TestParseEntryVersion2Rejected(t *testing.T) {
	_, priv := mustTestKeys(t)
	line := strings.Replace(mustSignedLine(t, priv), `"v":1`, `"v":2`, 1)
	assertParseMalformed(t, line, "v != 1 entry")
}

func TestParseEntryUnknownTypeRejected(t *testing.T) {
	_, priv := mustTestKeys(t)
	line := strings.Replace(mustSignedLine(t, priv), `"type":"grant"`, `"type":"admin"`, 1)
	assertParseMalformed(t, line, "unknown type string")
}

func TestAdapterParseEntryUnknownFieldRejected(t *testing.T) {
	_, priv := mustTestKeys(t)
	line := strings.Replace(mustSignedLine(t, priv), `"sig":"`, `"unknown_field":1,"sig":"`, 1)
	assertParseMalformed(t, line, "JSONL line with unknown field")
}

// ---------------------------------------------------------------------------
// Per-sequence revocation semantics (Decision 0015): revocations name one
// grant by sequence, never an identity/role.
// ---------------------------------------------------------------------------

func TestVerifyRoleAtRevokeNamesSequenceNotIdentityRole(t *testing.T) {
	// A revocation names exactly one grant by sequence. A later grant for
	// the same identity/role is a distinct successor entry, unaffected by
	// revocations that name earlier grants.
	//
	// Scenario: grant seq 0 ("r1"/"approver", T0..T0+10h), revoke seq 1
	// naming seq 0 effective T0+2h, re-grant seq 2 for the same
	// identity/role with ValidFrom T0+1h (<= the revoke's effective time).
	// Under the OLD rule (revoke kills every grant with
	// ValidFrom <= effective time) the re-grant would be DENIED at T0+4h;
	// under the NEW per-sequence rule it is allowed.
	t0 := baseTime
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("r1", "approver", t0, t0.Add(10*time.Hour), t0)
	b.addRevoke(0, t0.Add(2*time.Hour), t0)
	b.addGrant("r1", "approver", t0.Add(time.Hour), t0.Add(10*time.Hour), t0.Add(time.Hour))

	at := t0.Add(4 * time.Hour)
	if err := VerifyRoleAt("r1", "approver", at, b.snapshot(pub), b.head()); err != nil {
		t.Fatalf("re-grant at T0+4h under per-sequence revocation: got %v, want nil (allowed)", err)
	}

	// Sanity: the revocation still kills the grant it names. Evaluate the
	// original grant alone (entries 0 and 1): at T0+3h the revoke's
	// EffectiveAt (T0+2h) is <= the query time, so it must be ErrRevoked.
	orig := newChainBuilder(t, priv)
	orig.addGrant("r1", "approver", t0, t0.Add(10*time.Hour), t0)
	orig.addRevoke(0, t0.Add(2*time.Hour), t0.Add(2*time.Hour))
	if err := VerifyRoleAt("r1", "approver", t0.Add(3*time.Hour), orig.snapshot(pub), orig.head()); !errors.Is(err, ErrRevoked) {
		t.Fatalf("original grant at T0+3h: got %v, want ErrRevoked", err)
	}
}

// ---------------------------------------------------------------------------
// Structural non-retroactivity on programmatic Snapshots (thread 2/8):
// VerifyRoleAt must enforce ValidFrom >= AppendedAt (grants) and
// EffectiveAt >= AppendedAt (revokes) itself — ParseEntry is not the only
// entry point. A signed backdated grant fed directly as a Snapshot must be
// DENIED, not evaluated.
// ---------------------------------------------------------------------------

func TestVerifyRoleAtBackdatedGrantDenied(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	// Signed grant whose ValidFrom predates its own AppendedAt by 2h. The
	// window would otherwise cover the query time (old code: authorized).
	b.addGrant("synthetic-approver-001", "approver",
		baseTime.Add(-2*time.Hour), baseTime.Add(1000*time.Hour), baseTime)
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour),
		b.snapshot(pub), b.head())
	assertErrIs(t, err, ErrMalformed, "backdated grant (ValidFrom < AppendedAt)")
}

func TestVerifyRoleAtBackdatedRevokeDenied(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver",
		baseTime, baseTime.Add(1000*time.Hour), baseTime)
	// Revoke whose EffectiveAt predates its own AppendedAt. The old code
	// evaluated it and denied with ErrRevoked; structural validation must
	// reject it as malformed first.
	b.addRevoke(0, baseTime.Add(-time.Hour), baseTime)
	err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(3*time.Hour),
		b.snapshot(pub), b.head())
	assertErrIs(t, err, ErrMalformed, "backdated revoke (EffectiveAt < AppendedAt)")
}

// ---------------------------------------------------------------------------
// Query-time UnixNano range (thread 11): VerifyRoleAt REJECTS (never
// clamps) an `at` outside the representable unix-nanosecond range, where
// time.Time.UnixNano would wrap and destroy time ordering.
// ---------------------------------------------------------------------------

func TestVerifyRoleAtQueryTimeOutOfRangeDenied(t *testing.T) {
	b, pub, head := validFixture(t)
	snap := b.snapshot(pub)
	for _, at := range []time.Time{
		time.Date(2600, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1500, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2262, 4, 11, 23, 47, 17, 0, time.UTC),
		time.Date(1677, 9, 21, 0, 12, 43, 0, time.UTC),
	} {
		err := VerifyRoleAt("synthetic-approver-001", "approver", at, snap, head)
		assertErrIs(t, err, ErrMalformed, "out-of-range query time "+at.Format(time.RFC3339))
	}
}

func TestVerifyRoleAtQueryTimeAtUnixNanoBounds(t *testing.T) {
	// The extreme in-range query times are accepted end-to-end: a grant
	// covering them authorizes. Bounds verified empirically:
	// 9223372036 unix-sec = 2262-04-11T23:47:16Z,
	// -9223372036 unix-sec = 1677-09-21T00:12:44Z.
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	// NOTE: the min bound is 1677-09-21T00:12:44Z, not 00:00:00Z — an
	// earlier midnight would itself wrap on .UnixNano(), which is exactly
	// the trap this fix closes.
	b.addGrant("synthetic-approver-001", "approver",
		time.Date(1677, 9, 21, 0, 12, 44, 0, time.UTC),
		time.Date(1677, 9, 21, 1, 0, 0, 0, time.UTC),
		time.Date(1677, 9, 21, 0, 12, 44, 0, time.UTC))
	b.addGrant("synthetic-approver-001", "approver",
		time.Date(2262, 4, 11, 23, 0, 0, 0, time.UTC),
		time.Date(2262, 4, 11, 23, 47, 16, 0, time.UTC),
		time.Date(2262, 4, 11, 23, 0, 0, 0, time.UTC))
	snap := b.snapshot(pub)
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver",
		time.Date(2262, 4, 11, 23, 47, 16, 0, time.UTC), snap, head); err != nil {
		t.Fatalf("query at max bound: got %v, want nil (authorized)", err)
	}
	if err := VerifyRoleAt("synthetic-approver-001", "approver",
		time.Date(1677, 9, 21, 0, 12, 44, 0, time.UTC), snap, head); err != nil {
		t.Fatalf("query at min bound: got %v, want nil (authorized)", err)
	}
}
