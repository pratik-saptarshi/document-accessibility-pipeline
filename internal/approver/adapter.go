// Adapter implements the Decision 0019 verification contract for the
// release gate: VerifyRoleAt proves that an identity held a role at a given
// time, based on a signed, hash-chained register snapshot plus a
// caller-pinned head hash.
//
// Threat model: the register is untrusted input. The attacker can write
// arbitrary bytes to the file, including a fully self-consistent forged
// chain. The public key comes from the release-gate configuration — never
// from the register. Without a caller-pinned head the adapter proves only
// internal consistency, not completeness: freshness is the caller's
// responsibility, which is why pinnedHead is a mandatory separate parameter
// and never part of the Snapshot.
package approver

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"
)

// Verification failure sentinels. VerifyRoleAt returns nil iff every check
// passes; any non-nil error means DENY — the release gate must not
// distinguish soft paths. Callers use errors.Is to classify failures.
var (
	// ErrInvalidChain: the PrevHash links or AppendedAt monotonicity are broken.
	ErrInvalidChain = errors.New("approver: register chain integrity failure")
	// ErrGap: sequence numbers are not exactly 0,1,2,... (gap or duplicate).
	ErrGap = errors.New("approver: sequence gap or duplicate in register")
	// ErrBadSignature: an entry's Ed25519 signature does not verify against
	// the injected public key.
	ErrBadSignature = errors.New("approver: entry signature verification failed")
	// ErrHeadMismatch: the terminal chain hash does not equal the pinned head
	// (wrong head, or the register was truncated).
	ErrHeadMismatch = errors.New("approver: register head does not match pinned head")
	// ErrNoGrant: no grant covers (identity, role) at the query time.
	// Also used for an empty snapshot, which can contain no grant.
	ErrNoGrant = errors.New("approver: no grant for identity/role")
	// ErrExpired: no grant for (identity, role) covers the query time, but at
	// least one grant had ValidFrom <= the query time with the query time
	// strictly after its ValidUntil. Decision 0015 closes the interval at
	// both boundaries (not_before <= reviewed_at <= not_after), so a query
	// at exactly ValidUntil still authorizes.
	ErrExpired = errors.New("approver: grant window expired")
	// ErrNotYetValid: grants for (identity, role) exist but none covers the
	// query time, none is covered-but-revoked or expired, and at least one
	// starts after the query time.
	ErrNotYetValid = errors.New("approver: grant not yet valid")
	// ErrRevoked: at least one grant for (identity, role) covers the query
	// time, but every covering grant is named by a revoke entry whose
	// effective_at is at or before the query time.
	ErrRevoked = errors.New("approver: grant revoked")
)

// Snapshot is the loader output consumed by VerifyRoleAt. It carries NO
// head: verification cannot run without the caller's explicit pinnedHead.
type Snapshot struct {
	Entries []Entry
	PubKey  ed25519.PublicKey
}

// VerifyRoleAt returns nil iff ALL of the following hold:
//
//  1. len(snap.Entries) > 0.
//  2. Every entry's Version is 1 and its Type is grant/revoke, every
//     entry's signature verifies against snap.PubKey (ed25519.Verify called
//     directly; no verifier interface), grants carry a non-zero ValidUntil
//     (no open-ended grants), and — per Decision 0015 non-retroactivity —
//     grants have ValidFrom >= AppendedAt and revokes have EffectiveAt >=
//     AppendedAt.
//  3. Chain integrity: entry 0 has the zero genesis PrevHash, every other
//     entry's PrevHash equals the chain hash of the previous entry, and
//     AppendedAt is non-decreasing along the chain.
//  4. Sequence numbers are exactly 0,1,2,..., with no gaps or duplicates.
//  5. The chain hash of the last entry equals pinnedHead.
//  6. Temporal evaluation at `at` (UTC), per Decision 0015 "Validity at
//     reviewed_at": a grant authorizes a review iff ValidFrom <= at <=
//     ValidUntil (the closed interval applies at both exact boundaries)
//     and no revoke entry names that grant's Seq (RevokesSequence ==
//     grant.Seq) with EffectiveAt <= at. ANY clean covering grant
//     authorizes: overlapping grants are evaluated independently, and each
//     revocation targets exactly one grant by sequence — a revocation of
//     one grant never affects another covering grant, and a new grant is a
//     distinct successor entry that does not erase a revocation. A revoke
//     naming a sequence that is not a grant is inert. When no grant
//     authorizes, the denial is classified with deterministic, fail-closed
//     precedence: ErrRevoked (at least one grant covered `at` but every
//     covering grant was revoked) > ErrExpired (no grant covered `at`,
//     but at least one grant's window ended before `at`) > ErrNotYetValid
//     (every grant starts after `at`). ValidUntil is always set (no
//     open-ended grants); a grant with ValidUntil == 0 is rejected as
//     malformed during structural validation.
//  7. Identity and Role are compared by exact byte equality (no case folding,
//
// no trimming). Any failure returns a typed sentinel (any non-nil = deny).
// pinnedHead must come from a trusted source (release-gate configuration /
// review evidence), independent of the register file.
func VerifyRoleAt(identity, role string, at time.Time, snap Snapshot, pinnedHead [32]byte) error {
	entries := snap.Entries
	if len(entries) == 0 {
		// Documented choice per plan: an empty snapshot contains no grant.
		return ErrNoGrant
	}
	atUnix := at.UTC().UnixNano()

	// Structural checks plus signature verification, fail-closed. Signatures
	// are checked before chain structure so that any tampering reads as an
	// authentication failure.
	for i := range entries {
		e := &entries[i]
		if e.Version != FormatVersion {
			return fmt.Errorf("%w: entry %d has version %d", ErrMalformed, i, e.Version)
		}
		if e.Type != EntryGrant && e.Type != EntryRevoke {
			return fmt.Errorf("%w: entry %d has unknown type %#02x", ErrMalformed, i, e.Type)
		}
		if e.Type == EntryGrant && e.ValidUntil == 0 {
			return fmt.Errorf("%w: entry %d grant has no ValidUntil (open-ended grants are not supported)", ErrMalformed, i)
		}
		// Non-retroactivity (Decision 0015): not_before may not predate the
		// entry's authenticated append time, and a revocation's effective_at
		// may not predate its own append time. ParseEntry enforces this on
		// the wire path; VerifyRoleAt re-enforces it here so programmatic
		// Snapshots cannot bypass it.
		if e.Type == EntryGrant && e.ValidFrom < e.AppendedAt {
			return fmt.Errorf("%w: entry %d grant valid_from predates appended_at (no retroactive grants)", ErrMalformed, i)
		}
		if e.Type == EntryRevoke && e.EffectiveAt < e.AppendedAt {
			return fmt.Errorf("%w: entry %d revoke effective_at predates appended_at", ErrMalformed, i)
		}
		if len(snap.PubKey) != ed25519.PublicKeySize ||
			len(e.Signature) != ed25519.SignatureSize ||
			!ed25519.Verify(snap.PubKey, e.CanonicalBytes(), e.Signature) {
			return fmt.Errorf("%w: entry %d", ErrBadSignature, i)
		}
	}

	// Chain integrity: genesis rule, PrevHash links, AppendedAt monotonicity.
	for i := range entries {
		e := &entries[i]
		if i == 0 {
			if e.PrevHash != [32]byte{} {
				return fmt.Errorf("%w: entry 0 PrevHash is not the genesis zero hash", ErrInvalidChain)
			}
			continue
		}
		prev := &entries[i-1]
		if want := ChainHash(prev.PrevHash, prev.CanonicalBytes()); e.PrevHash != want {
			return fmt.Errorf("%w: entry %d PrevHash does not match the chain hash of entry %d", ErrInvalidChain, i, i-1)
		}
		if e.AppendedAt < prev.AppendedAt {
			return fmt.Errorf("%w: entry %d AppendedAt goes backward", ErrInvalidChain, i)
		}
	}

	// Sequence numbers exactly 0,1,2,...
	for i := range entries {
		if entries[i].Seq != uint64(i) {
			return fmt.Errorf("%w: entry %d has seq %d", ErrGap, i, entries[i].Seq)
		}
	}

	// The pinned head is mandatory: it detects truncation and proves the
	// snapshot is the register the caller expects, not just a consistent one.
	last := &entries[len(entries)-1]
	if head := ChainHash(last.PrevHash, last.CanonicalBytes()); head != pinnedHead {
		return ErrHeadMismatch
	}

	// Temporal evaluation (Decision 0015 "Validity at reviewed_at"): a
	// review is authorized iff ANY grant for (identity, role) covers the
	// query time — ValidFrom <= at <= ValidUntil — and is not revoked at
	// that time. Overlapping grants are evaluated independently. When no
	// grant authorizes, the denial follows deterministic, fail-closed
	// precedence: revoked > expired > not-yet-valid.
	var (
		anyGrant       bool
		coveredRevoked bool
		expired        bool
		notYetValid    bool
	)
	for i := range entries {
		g := &entries[i]
		if g.Type != EntryGrant || g.Identity != identity || g.Role != role {
			continue
		}
		anyGrant = true
		// ValidUntil is always set: the wire format rejects open-ended
		// grants (Fix 4) and the structural check above rejects
		// ValidUntil == 0 on programmatic Snapshots. The closed interval
		// ValidFrom <= at <= ValidUntil applies at both boundaries.
		if g.ValidFrom > atUnix || atUnix > g.ValidUntil {
			if g.ValidFrom > atUnix {
				notYetValid = true
			} else {
				expired = true
			}
			continue
		}
		// The grant covers the query time. Decision 0015: the revoke
		// names exactly one grant by sequence. No identity/role matching:
		// a revocation cannot touch grants it does not name.
		revoked := false
		for j := range entries {
			r := &entries[j]
			if r.Type != EntryRevoke {
				continue
			}
			if r.RevokesSequence == g.Seq && r.EffectiveAt <= atUnix {
				revoked = true
				break
			}
		}
		if !revoked {
			return nil // authorized by a clean covering grant
		}
		coveredRevoked = true
	}
	if !anyGrant {
		return ErrNoGrant
	}
	if coveredRevoked {
		return ErrRevoked
	}
	if expired {
		return ErrExpired
	}
	if notYetValid {
		return ErrNotYetValid
	}
	// Unreachable: every grant for (identity, role) is classified above.
	return ErrNoGrant
}
