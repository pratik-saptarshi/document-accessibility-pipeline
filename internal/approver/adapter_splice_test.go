// Splice (link-check) test for the approver register chain (R8-02).
//
// The Round 8 review found that removing the PrevHash chain check in
// VerifyRoleAt survived every existing test: the existing adversarial
// tests either break a signature, reorder whole entries (which also
// breaks sequence numbering / the pinned head), or truncate the register
// (which the pinned head catches). None of them isolates the PrevHash
// link itself.
//
// This test isolates it. Both chains are signed with the SAME key, share
// the same timestamps, and their seq-2 revoke entries carry identical
// fields except PrevHash — so the spliced entry is validly signed,
// structurally well-formed, in sequence, monotonic in AppendedAt, and
// (with the spliced head pinned) head-consistent. The ONLY check that
// can reject the spliced snapshot is the PrevHash link check in
// adapter.go: the spliced entry's PrevHash is the chain hash of the
// second chain's seq-1 entry, not of the first chain's seq-1 entry.
//
// Mutation check (performed during development, reported in the PR):
// bypassing the PrevHash comparison in VerifyRoleAt makes this test
// fail (verification returns nil, because the spliced revoke still
// names seq 1 and the queried grant at seq 0 is unaffected), while
// every other check still passes. Restoring the check makes it pass.
package approver

import (
	"testing"
	"time"
)

func TestVerifyRoleAtSplicedChainEntryRejected(t *testing.T) {
	pub, priv := mustTestKeys(t)

	// First chain: grant(seq0, approver-001) -> grant(seq1,
	// approver-002) -> revoke(seq2 naming seq1).
	first := newChainBuilder(t, priv)
	first.addGrant("synthetic-approver-001", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	first.addGrant("synthetic-approver-002", "approver", baseTime.Add(time.Hour), baseTime.Add(48*time.Hour), baseTime.Add(time.Hour))
	first.addRevoke(1, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))

	// Second chain, same signing key and same timestamps, but different
	// grants — so its chain hashes diverge from the first chain's from
	// seq 0 onward. Its seq-2 entry is a revoke with fields identical
	// to the first chain's seq-2 revoke except for PrevHash (and hence
	// its signature over those canonical bytes).
	second := newChainBuilder(t, priv)
	second.addGrant("synthetic-approver-101", "approver", baseTime, baseTime.Add(24*time.Hour), baseTime)
	second.addGrant("synthetic-approver-102", "approver", baseTime.Add(time.Hour), baseTime.Add(48*time.Hour), baseTime.Add(time.Hour))
	second.addRevoke(1, baseTime.Add(2*time.Hour), baseTime.Add(2*time.Hour))

	at := baseTime.Add(3 * time.Hour)

	// Controls: each chain verifies on its own. The first chain
	// authorizes approver-001 (its grant at seq 0 is not the revoked
	// one); the second chain authorizes approver-101 likewise. This
	// proves the spliced entry is validly signed in its own chain.
	if err := VerifyRoleAt("synthetic-approver-001", "approver", at, first.snapshot(pub), first.head()); err != nil {
		t.Fatalf("control: first chain alone: %v", err)
	}
	if err := VerifyRoleAt("synthetic-approver-101", "approver", at, second.snapshot(pub), second.head()); err != nil {
		t.Fatalf("control: second chain alone: %v", err)
	}
	if first.entries[2].PrevHash == second.entries[2].PrevHash {
		t.Fatal("test setup: seq-2 PrevHash values must differ between chains for the splice to be meaningful")
	}

	// Splice: replace the first chain's seq-2 revoke with the second
	// chain's seq-2 revoke. Sequence, timestamps, revoke target, and
	// signature validity are all preserved; only the PrevHash link to
	// the first chain's seq-1 entry is broken.
	spliced := first.snapshot(pub)
	spliced.Entries[2] = second.entries[2]

	// Pin the head of the spliced snapshot itself, so the pinned-head
	// check cannot be what rejects it: the chain-link check is the
	// only remaining line of defence.
	last := spliced.Entries[len(spliced.Entries)-1]
	splicedHead := ChainHash(last.PrevHash, last.CanonicalBytes())

	err := VerifyRoleAt("synthetic-approver-001", "approver", at, spliced, splicedHead)
	assertErrIs(t, err, ErrInvalidChain, "spliced seq-2 entry from a second chain")
}
