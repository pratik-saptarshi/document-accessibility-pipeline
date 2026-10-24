// Fuzz target for the approver register parser (R8-02).
//
// The fuzz oracle is CRASH-ONLY: this target may assert "no panic / no
// crash" only. It must NOT assert that case-variant keys are rejected:
// Decision 0024 accepted liberal parsing, and fuzzer evidence of
// byte-distinct entries normalizing to the same signed field is a
// DECISION EVENT to report, not a test failure. Do not tighten the
// parser or add rejection assertions here on the basis of such
// evidence; report it for an owner decision instead.
package approver

import (
	"strings"
	"testing"
)

// FuzzParseEntry feeds hostile bytes to ParseEntry, the register
// parser boundary for untrusted JSONL input. Any error is an acceptable
// outcome; a panic is not.
func FuzzParseEntry(f *testing.F) {
	// Seed with well-formed grant and revoke lines so the fuzzer starts
	// from inputs that reach deep into the parser's per-type checks.
	f.Add([]byte(testGrantLine()))
	f.Add([]byte(testRevokeLine()))
	// Case-variant key spelling (Decision 0024 liberal parsing): a
	// seed, not an expected rejection — see the file comment above.
	f.Add([]byte(strings.Replace(testGrantLine(), `"seq"`, `"Seq"`, 1)))
	// Truncated, empty, and non-JSON seeds cover the early-exit paths.
	f.Add([]byte(testGrantLine()[:len(testGrantLine())/2]))
	f.Add([]byte(""))
	f.Add([]byte("not json"))
	f.Add([]byte(`{"v":1,"seq":0,"seq":0}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		// Crash-only oracle: ParseEntry must return (with or without an
		// error) for every input; it must never panic.
		_, _ = ParseEntry(data)
	})
}
