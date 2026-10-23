// Loader parses a JSONL register file into a Snapshot (Task 3).
//
// The register file is untrusted input. The public key is a caller
// parameter and is never read from the file — the wire format carries no
// key field (ParseEntry rejects one as unknown), so a key can never be
// substituted from attacker-controlled bytes.
//
// The loader also does NOT compute or set any head: Snapshot carries no
// head at all. The caller supplies the trusted pinnedHead to VerifyRoleAt
// from gate configuration or review evidence, independent of the file.
// Without a caller-pinned head the adapter proves internal consistency
// only; freshness is the caller's responsibility.
package approver

import (
	"bufio"
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
)

// maxRegisterLineBytes bounds a single JSONL line. Lines larger than this
// are rejected before parsing so a hostile file cannot force unbounded
// per-line allocation.
const maxRegisterLineBytes = 1 << 20 // 1 MiB

// maxRegisterBytes bounds the total size of the register file (64 MiB).
// The file is streamed line by line — never held whole in memory — but an
// unbounded byte count still means unbounded parse and verification work
// (hash-chain verification is linear in the entry count). A genuine
// approver register holds a handful of grant/revoke entries for named
// individuals (Decision 0015: small teams, explicit owner sign-off), so 64
// MiB — 64x the per-line cap — leaves orders of magnitude of headroom
// while bounding hostile-input CPU and I/O.
const maxRegisterBytes = 64 << 20 // 64 MiB

// maxRegisterEntries bounds the number of entries in the register file
// (1,000,000). Together with maxRegisterBytes this caps hostile-input
// work: even the smallest conceivable line cannot push verification past
// this many chain/hash/signature operations, and the byte cap stops huge
// lines far earlier. For a register that in practice holds a handful of
// entries, this is headroom by several orders of magnitude.
const maxRegisterEntries = 1_000_000

// registerLimits groups the hostile-input bounds enforced by the loader.
// LoadSnapshot uses the production constants above; tests inject small
// limits through loadSnapshotWithLimits to exercise the bounds without
// writing large fixtures to disk.
type registerLimits struct {
	maxLineBytes  int
	maxTotalBytes int64
	maxEntries    int
}

var defaultRegisterLimits = registerLimits{
	maxLineBytes:  maxRegisterLineBytes,
	maxTotalBytes: maxRegisterBytes,
	maxEntries:    maxRegisterEntries,
}

// LoadSnapshot parses path as newline-delimited JSONL and returns
// Snapshot{Entries, PubKey: pubKey}. Every line goes through ParseEntry,
// so all fail-closed parse rules (duplicate-key rejection, version check,
// type check, unknown-field rejection, per-type required/forbidden fields)
// apply.
//
// The file is streamed through a bufio.Scanner whose buffer is hard-capped
// at maxLineBytes+1, so the file is never read whole into memory and no
// single line can force unbounded allocation.
//
// Loader-level rules, all documented here:
//   - line size bound: a line longer than 1 MiB is rejected;
//   - total size bound: a file larger than 64 MiB is rejected;
//   - entry count bound: a file with more than 1,000,000 entries is
//     rejected;
//   - empty file (or a file with no entries) is rejected;
//   - the file must end with a newline: a final partial line without a
//     trailing newline is trailing garbage in a newline-delimited format
//     and the whole file is rejected.
//
// Sequence ordering is deliberately NOT enforced here: that is the
// verifier's job (VerifyRoleAt returns ErrGap). The loader is thin:
// stream + bound + parse + accumulate.
func LoadSnapshot(path string, pubKey ed25519.PublicKey) (Snapshot, error) {
	return loadSnapshotWithLimits(path, pubKey, defaultRegisterLimits)
}

func loadSnapshotWithLimits(path string, pubKey ed25519.PublicKey, lim registerLimits) (Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("approver: read register file: %w", err)
	}
	defer f.Close()

	// The scanner buffer is hard-capped: a line longer than the cap makes
	// Scan fail with bufio.ErrTooLong instead of growing the buffer
	// without bound.
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), lim.maxLineBytes+1)

	snap := Snapshot{PubKey: pubKey}
	var totalBytes int64
	for sc.Scan() {
		line := sc.Bytes()
		// +1 accounts for the terminating newline the scanner consumed.
		// A final line without one is rejected below as trailing garbage,
		// so the 1-byte overcount in that case is harmless.
		totalBytes += int64(len(line)) + 1
		if totalBytes > lim.maxTotalBytes {
			return Snapshot{}, fmt.Errorf(
				"%w: register file exceeds the %d-byte total size limit", ErrMalformed, lim.maxTotalBytes)
		}
		if len(snap.Entries) >= lim.maxEntries {
			return Snapshot{}, fmt.Errorf(
				"%w: register file exceeds the %d-entry limit", ErrMalformed, lim.maxEntries)
		}
		e, err := ParseEntry(line)
		if err != nil {
			return Snapshot{}, fmt.Errorf("line %d: %w", len(snap.Entries), err)
		}
		snap.Entries = append(snap.Entries, e)
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return Snapshot{}, fmt.Errorf(
				"%w: line %d exceeds the %d-byte line limit", ErrMalformed, len(snap.Entries), lim.maxLineBytes)
		}
		return Snapshot{}, fmt.Errorf("approver: read register file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		return Snapshot{}, fmt.Errorf("approver: read register file: %w", err)
	}
	if info.Size() == 0 {
		return Snapshot{}, fmt.Errorf("%w: empty register file", ErrMalformed)
	}
	// Documented choice: the wire format is newline-delimited JSONL, so the
	// file must end with a newline. A final partial line without one is
	// trailing garbage — reject the file rather than silently dropping or
	// half-parsing the last line.
	last := make([]byte, 1)
	if _, err := f.ReadAt(last, info.Size()-1); err != nil {
		return Snapshot{}, fmt.Errorf("approver: read register file: %w", err)
	}
	if last[0] != '\n' {
		return Snapshot{}, fmt.Errorf(
			"%w: final line is missing its terminating newline (trailing garbage)", ErrMalformed)
	}
	if len(snap.Entries) == 0 {
		// Defensive: with the trailing-newline rule a successful scan
		// implies at least one line, and every line is parsed above.
		return Snapshot{}, fmt.Errorf("%w: register file contains no entries", ErrMalformed)
	}
	return snap, nil
}
