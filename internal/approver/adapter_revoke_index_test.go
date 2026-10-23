// Revocation-index tests (Decision 0015, Codex P2 thread on PR #183).
//
// VerifyRoleAt replaced its O(GxR) per-grant revocation scan with a
// seq->earliest-EffectiveAt index built once per call. These tests pin the
// semantics the index must preserve exactly:
//
//   - a grant is revoked at `at` iff a revoke names its Seq with
//     EffectiveAt <= at (inclusive, prospective);
//   - with several revokes naming the same seq, the EARLIEST EffectiveAt
//     governs;
//   - revokes naming an unknown seq (or a non-grant seq) are inert;
//   - denial precedence is unchanged: revoked > expired > not-yet-valid;
//   - any clean covering grant authorizes.
//
// The behavioral tests below were written against the pre-index code and
// passed there; they must keep passing after the refactor, proving no
// semantic change. The TestRevokeIndex* unit tests cover the unexported
// index builder directly and fail until the index exists.
package approver

import (
	"testing"
	"time"
)

// multiRevokeFixture builds a chain exercising the index's edge cases:
//
//	seq0: grant [T0, T0+10h] for synthetic-approver-mr (appended T0)
//	seq1: revoke naming seq0, effective T0+8h (appended T0+1h)
//	seq2: revoke naming seq0, effective T0+3h (appended T0+2h) — earlier
//	      than seq1 despite being appended later (both satisfy Decision
//	      0015: effective_at >= appended_at, appended_at non-decreasing)
//	seq3: grant [T0+2h, T0+10h] for synthetic-approver-mr (appended T0+2h)
//	seq4: revoke naming seq3, effective T0+5h (appended T0+5h)
//	seq5: revoke naming seq 999 (unknown seq; must be inert, appended T0+6h)
func multiRevokeFixture(t *testing.T, b *chainBuilder) [32]byte {
	t.Helper()
	b.addGrant("synthetic-approver-mr", "approver", baseTime, baseTime.Add(10*time.Hour), baseTime)
	b.addRevoke(0, baseTime.Add(8*time.Hour), baseTime.Add(time.Hour))
	b.addRevoke(0, baseTime.Add(3*time.Hour), baseTime.Add(2*time.Hour))
	b.addGrant("synthetic-approver-mr", "approver", baseTime.Add(2*time.Hour), baseTime.Add(10*time.Hour), baseTime.Add(2*time.Hour))
	b.addRevoke(3, baseTime.Add(5*time.Hour), baseTime.Add(5*time.Hour))
	b.addRevoke(999, baseTime.Add(6*time.Hour), baseTime.Add(6*time.Hour))
	return b.head()
}

// TestVerifyRoleAtMultipleRevokesEarliestGoverns is the behavioral
// characterization of the index semantics. Every verdict below holds under
// the pre-index O(GxR) scan; the index must reproduce each of them.
func TestVerifyRoleAtMultipleRevokesEarliestGoverns(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	head := multiRevokeFixture(t, b)
	for _, tc := range []struct {
		name  string
		at    time.Time
		want  error // nil means authorized
		check func(t *testing.T, err error)
	}{
		// Before the earliest revoke for seq0 (effective T0+3h): seq0 is
		// clean and covers -> authorize. The unknown-seq revoke (seq5)
		// must not interfere.
		{"before earliest revoke", baseTime.Add(30 * time.Minute), nil, nil},
		// At exactly T0+3h the earliest revoke for seq0 fires (inclusive
		// boundary): seq0 is revoked, but seq3 covers and is clean ->
		// authorize. The later revoke for seq0 (T0+8h) is not yet
		// effective; the verdict must not depend on it.
		{"earliest effective fires, other grant clean", baseTime.Add(3 * time.Hour), nil, nil},
		// T0+5h: seq3's revoke fires as well; every covering grant is now
		// revoked -> ErrRevoked.
		{"all covering grants revoked", baseTime.Add(5 * time.Hour), ErrRevoked, nil},
		// T0+9h: the LATER revoke for seq0 (T0+8h) is effective too;
		// verdict unchanged, proving the earliest governed all along.
		{"later revoke also effective", baseTime.Add(9 * time.Hour), ErrRevoked, nil},
		// T0+11h: both grant windows ended; nothing covers. Revocations
		// cannot create coverage, so the expired grants classify the
		// denial (the revoked>expired precedence needs a covering grant).
		{"post-window, no covering grant", baseTime.Add(11 * time.Hour), ErrExpired, nil},
		// T0-1h: every grant for the identity starts after the query.
		{"pre-window", baseTime.Add(-1 * time.Hour), ErrNotYetValid, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyRoleAt("synthetic-approver-mr", "approver", tc.at, b.snapshot(pub), head)
			if tc.want == nil {
				if err != nil {
					t.Fatalf("got %v, want nil (authorized)", err)
				}
				return
			}
			assertErrIs(t, err, tc.want, tc.name)
		})
	}
}

// TestVerifyRoleAtRevokeNamingRevokeSeqIsInert: a revoke naming the seq of
// another revoke entry (a non-grant seq) is inert — no grant carries that
// seq, so the lookup never fires and the grant stays authorized.
func TestVerifyRoleAtRevokeNamingRevokeSeqIsInert(t *testing.T) {
	pub, priv := mustTestKeys(t)
	b := newChainBuilder(t, priv)
	b.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(1000*time.Hour), baseTime)
	b.addRevoke(999, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))
	// seq2 names seq1, which is itself a revoke entry, not a grant.
	b.addRevoke(1, baseTime.Add(3*time.Hour), baseTime.Add(3*time.Hour))
	head := b.head()
	if err := VerifyRoleAt("synthetic-approver-001", "approver", baseTime.Add(4*time.Hour), b.snapshot(pub), head); err != nil {
		t.Fatalf("revoke naming a non-grant seq must be inert: %v", err)
	}
}

// ---------------------------------------------------------------------------
// revokeIndex unit tests (Codex P2 thread on PR #183). The builder is the
// heart of the O(GxR) -> O(G+R) change: these tests pin its contract. If
// the builder is deleted, the package fails to build — a loud,
// fail-closed guard rather than a silent slowdown.
// ---------------------------------------------------------------------------

// TestRevokeIndexEarliestEffectiveAtWins: several revokes naming the same
// seq collapse to the earliest EffectiveAt, interleaved grants and an
// unknown-seq revoke do not disturb the map, and grant entries never enter
// the index even when they carry revoke-only fields (programmatic
// Snapshots; the wire format forbids this).
func TestRevokeIndexEarliestEffectiveAtWins(t *testing.T) {
	entries := []Entry{
		{Type: EntryRevoke, RevokesSequence: 7, EffectiveAt: 300},
		{Type: EntryGrant, RevokesSequence: 7, EffectiveAt: 10}, // must not pollute the index
		{Type: EntryRevoke, RevokesSequence: 7, EffectiveAt: 100},
		{Type: EntryRevoke, RevokesSequence: 42, EffectiveAt: 500},
		{Type: EntryRevoke, RevokesSequence: 0, EffectiveAt: 900}, // zero-value seq key
		{Type: EntryRevoke, RevokesSequence: 7, EffectiveAt: 200},
		{Type: EntryRevoke, RevokesSequence: 9999, EffectiveAt: 50}, // unknown seq; key present, inert
	}
	index := revokeIndex(entries)
	if got, want := index[7], int64(100); got != want {
		t.Fatalf("index[7] = %d, want %d (earliest EffectiveAt wins)", got, want)
	}
	if got, want := index[42], int64(500); got != want {
		t.Fatalf("index[42] = %d, want %d", got, want)
	}
	if got, want := index[0], int64(900); got != want {
		t.Fatalf("index[0] = %d, want %d (zero-value seq key)", got, want)
	}
	if got, want := index[9999], int64(50); got != want {
		t.Fatalf("index[9999] = %d, want %d (unknown seq still indexed)", got, want)
	}
	if len(index) != 4 {
		t.Fatalf("len(index) = %d, want 4 (grants must not contribute)", len(index))
	}
}

// TestRevokeIndexEmpty: no entries (or no revokes) yields an empty index —
// the temporal loop then treats every covering grant as clean.
func TestRevokeIndexEmpty(t *testing.T) {
	if index := revokeIndex(nil); len(index) != 0 {
		t.Fatalf("revokeIndex(nil) has %d entries, want 0", len(index))
	}
	grantsOnly := []Entry{
		{Type: EntryGrant, Seq: 0},
		{Type: EntryGrant, Seq: 1},
	}
	if index := revokeIndex(grantsOnly); len(index) != 0 {
		t.Fatalf("revokeIndex(grants-only) has %d entries, want 0", len(index))
	}
}

// TestRevokeIndexMatchesScanSemantics: for a crafted mixed entry list, the
// index lookup (earliest EffectiveAt <= at) agrees with the old O(GxR)
// scan's existential check (some revoke naming the seq with EffectiveAt <=
// at) at every interesting query time. This is the semantic-consistency
// bridge between the old scan and the new index.
func TestRevokeIndexMatchesScanSemantics(t *testing.T) {
	entries := []Entry{
		{Type: EntryGrant, Seq: 0},
		{Type: EntryRevoke, Seq: 1, RevokesSequence: 0, EffectiveAt: 800},
		{Type: EntryRevoke, Seq: 2, RevokesSequence: 0, EffectiveAt: 300},
		{Type: EntryRevoke, Seq: 3, RevokesSequence: 5, EffectiveAt: 100},
	}
	index := revokeIndex(entries)
	oldScanRevoked := func(seq uint64, at int64) bool {
		for _, r := range entries {
			if r.Type == EntryRevoke && r.RevokesSequence == seq && r.EffectiveAt <= at {
				return true
			}
		}
		return false
	}
	for _, at := range []int64{0, 100, 299, 300, 500, 799, 800, 1000} {
		for _, seq := range []uint64{0, 1, 5, 42} {
			effAt, ok := index[seq]
			indexRevoked := ok && effAt <= at
			if indexRevoked != oldScanRevoked(seq, at) {
				t.Fatalf("seq %d at %d: index says revoked=%v, scan says %v",
					seq, at, indexRevoked, oldScanRevoked(seq, at))
			}
		}
	}
}
