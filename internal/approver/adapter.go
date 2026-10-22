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
	// ErrExpired: the latest covering grant's window ended at or before the
	// query time.
	ErrExpired = errors.New("approver: grant window expired")
	// ErrNotYetValid: a grant for (identity, role) exists but its ValidFrom
	// is after the query time.
	ErrNotYetValid = errors.New("approver: grant not yet valid")
	// ErrRevoked: a revoke entry names the evaluated grant's sequence and its
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
//  2. Every entry's Version is 1 and its Type is grant/revoke, and every
//     entry's signature verifies against snap.PubKey (ed25519.Verify called
//     directly; no verifier interface).
//  3. Chain integrity: entry 0 has the zero genesis PrevHash, every other
//     entry's PrevHash equals the chain hash of the previous entry, and
//     AppendedAt is non-decreasing along the chain.
//  4. Sequence numbers are exactly 0,1,2,..., with no gaps or duplicates.
//  5. The chain hash of the last entry equals pinnedHead.
//  6. Temporal evaluation at `at` (UTC): the LATEST grant for
//     (identity, role) with ValidFrom <= at satisfies
//     ValidUntil == 0 || at < ValidUntil, and no revoke entry names that
//     grant's Seq (RevokesSequence == grant.Seq) with EffectiveAt <= at.
//     Each revocation targets exactly one grant by sequence and applies
//     prospectively from effective_at, inclusive (Decision 0015): a new
//     grant is a distinct successor entry, unaffected by revocations that
//     name earlier grants. A revoke naming a sequence that is not a grant
//     is inert.
//
// Identity and Role are compared by exact byte equality (no case folding,
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

	// Temporal evaluation. grant ends up the latest grant for
	// (identity, role) with ValidFrom <= at (ties broken by later seq).
	var grant *Entry
	for i := range entries {
		e := &entries[i]
		if e.Type != EntryGrant || e.Identity != identity || e.Role != role {
			continue
		}
		if e.ValidFrom > atUnix {
			continue
		}
		if grant == nil || e.ValidFrom > grant.ValidFrom ||
			(e.ValidFrom == grant.ValidFrom && e.Seq > grant.Seq) {
			grant = e
		}
	}
	if grant == nil {
		for i := range entries {
			e := &entries[i]
			if e.Type == EntryGrant && e.Identity == identity && e.Role == role && e.ValidFrom > atUnix {
				return ErrNotYetValid
			}
		}
		return ErrNoGrant
	}
	if grant.ValidUntil != 0 && atUnix >= grant.ValidUntil {
		return ErrExpired
	}
	for i := range entries {
		e := &entries[i]
		if e.Type != EntryRevoke {
			continue
		}
		// Decision 0015: the revoke names exactly one grant by sequence.
		// No identity/role matching: a revocation cannot touch grants it
		// does not name.
		if e.RevokesSequence == grant.Seq && e.EffectiveAt <= atUnix {
			return ErrRevoked
		}
	}
	return nil
}
