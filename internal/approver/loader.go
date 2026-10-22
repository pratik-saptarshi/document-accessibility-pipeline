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
	"bytes"
	"crypto/ed25519"
	"fmt"
	"os"
)

// maxRegisterLineBytes bounds a single JSONL line. Lines larger than this
// are rejected before parsing so a hostile file cannot force unbounded
// per-line allocation.
const maxRegisterLineBytes = 1 << 20 // 1 MiB

// LoadSnapshot parses path as newline-delimited JSONL and returns
// Snapshot{Entries, PubKey: pubKey}. Every line goes through ParseEntry,
// so all fail-closed parse rules (version check, type check, unknown-field
// rejection, per-type required/forbidden fields) apply.
//
// Loader-level rules, all documented here:
//   - line size bound: a line longer than 1 MiB is rejected;
//   - empty file (or a file with no entries) is rejected;
//   - the file must end with a newline: a final partial line without a
//     trailing newline is trailing garbage in a newline-delimited format
//     and the whole file is rejected.
//
// Sequence ordering is deliberately NOT enforced here: that is the
// verifier's job (VerifyRoleAt returns ErrGap). The loader is thin:
// parse + accumulate.
func LoadSnapshot(path string, pubKey ed25519.PublicKey) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("approver: read register file: %w", err)
	}
	if len(data) == 0 {
		return Snapshot{}, fmt.Errorf("%w: empty register file", ErrMalformed)
	}
	// Documented choice: the wire format is newline-delimited JSONL, so the
	// file must end with a newline. A final partial line without one is
	// trailing garbage — reject the file rather than silently dropping or
	// half-parsing the last line.
	if !bytes.HasSuffix(data, []byte("\n")) {
		return Snapshot{}, fmt.Errorf(
			"%w: final line is missing its terminating newline (trailing garbage)", ErrMalformed)
	}
	lines := bytes.Split(data[:len(data)-1], []byte("\n"))
	snap := Snapshot{PubKey: pubKey}
	for i, line := range lines {
		if len(line) > maxRegisterLineBytes {
			return Snapshot{}, fmt.Errorf(
				"%w: line %d exceeds the %d-byte line limit", ErrMalformed, i, maxRegisterLineBytes)
		}
		e, err := ParseEntry(line)
		if err != nil {
			return Snapshot{}, fmt.Errorf("line %d: %w", i, err)
		}
		snap.Entries = append(snap.Entries, e)
	}
	if len(snap.Entries) == 0 {
		// A file that is just "\n" parses to zero entries.
		return Snapshot{}, fmt.Errorf("%w: register file contains no entries", ErrMalformed)
	}
	return snap, nil
}
