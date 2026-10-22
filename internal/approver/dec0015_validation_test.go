// Tests for the Decision 0015 validation rules (fixes 2-6, 10).
//
// All identities and keys are synthetic. No real person names, no real keys.
package approver

import (
	"crypto/ed25519"
	"errors"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Fix 3 — appended_at mandatory. A missing or null appended_at is
// malformed: there is no epoch fallback (Decision 0015: appended_at is
// source-authenticated).
// ---------------------------------------------------------------------------

func TestParseEntryMissingAppendedAtRejected(t *testing.T) {
	base := testGrantLine()
	variants := map[string]string{
		"absent": strings.Replace(base, `"appended_at":"2026-10-21T22:00:01Z",`, ``, 1),
		"null":   strings.Replace(base, `"appended_at":"2026-10-21T22:00:01Z"`, `"appended_at":null`, 1),
	}
	for name, line := range variants {
		if _, err := ParseEntry([]byte(line)); !errors.Is(err, ErrMalformed) {
			t.Fatalf("appended_at %s: got %v, want errors.Is ErrMalformed", name, err)
		}
	}
	// Revokes have the same rule: appended_at is the authenticated append
	// time of every entry type.
	revokeBase := testRevokeLine()
	revokeNull := strings.Replace(revokeBase, `"appended_at":"2026-10-22T00:00:01Z"`, `"appended_at":null`, 1)
	if _, err := ParseEntry([]byte(revokeNull)); !errors.Is(err, ErrMalformed) {
		t.Fatalf("revoke appended_at null: got %v, want errors.Is ErrMalformed", err)
	}
}

// ---------------------------------------------------------------------------
// Fix 10 — Strict timestamp spelling. Decision 0015 mandates RFC 3339 UTC Z
// with whole-second precision. The spelling check runs BEFORE time.Parse
// (which would otherwise accept offsets and fractional seconds). Applies to
// valid_from, valid_until, effective_at, and appended_at.
// ---------------------------------------------------------------------------

func TestParseEntryNoncanonicalTimeSpellingRejected(t *testing.T) {
	grantBase := testGrantLine()
	revokeBase := testRevokeLine()
	cases := map[string]string{
		"valid_from offset":     strings.Replace(grantBase, `"valid_from":"2026-10-21T22:00:01Z"`, `"valid_from":"2026-10-21T22:00:01+02:00"`, 1),
		"valid_from fraction":   strings.Replace(grantBase, `"valid_from":"2026-10-21T22:00:01Z"`, `"valid_from":"2026-10-21T22:00:01.5Z"`, 1),
		"valid_until offset":    strings.Replace(grantBase, `"valid_until":"2027-12-31T23:59:59Z"`, `"valid_until":"2027-12-31T23:59:59+02:00"`, 1),
		"valid_until fraction":  strings.Replace(grantBase, `"valid_until":"2027-12-31T23:59:59Z"`, `"valid_until":"2027-12-31T23:59:59.123Z"`, 1),
		"appended_at offset":    strings.Replace(grantBase, `"appended_at":"2026-10-21T22:00:01Z"`, `"appended_at":"2026-10-21T22:00:01+02:00"`, 1),
		"appended_at fraction":  strings.Replace(grantBase, `"appended_at":"2026-10-21T22:00:01Z"`, `"appended_at":"2026-10-21T22:00:01.5Z"`, 1),
		"effective_at offset":   strings.Replace(revokeBase, `"effective_at":"2026-10-22T00:00:01Z"`, `"effective_at":"2026-10-22T00:00:01+02:00"`, 1),
		"effective_at fraction": strings.Replace(revokeBase, `"effective_at":"2026-10-22T00:00:01Z"`, `"effective_at":"2026-10-22T00:00:01.5Z"`, 1),
	}
	for name, line := range cases {
		if _, err := ParseEntry([]byte(line)); !errors.Is(err, ErrMalformed) {
			t.Fatalf("%s: got %v, want errors.Is ErrMalformed", name, err)
		}
	}
	// Canonical whole-second Z spellings still parse.
	if _, err := ParseEntry([]byte(grantBase)); err != nil {
		t.Fatalf("canonical grant Z spelling: unexpected rejection: %v", err)
	}
	if _, err := ParseEntry([]byte(revokeBase)); err != nil {
		t.Fatalf("canonical revoke Z spelling: unexpected rejection: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Wire-time emission consistency. Decision 0015 mandates whole-second
// precision on the wire: MarshalEntry emits strict whole-second UTC Z and
// REJECTS (fail loud, no silent truncation) any entry carrying sub-second
// nanos.
// ---------------------------------------------------------------------------

func TestMarshalEntryRejectsSubSecondNanos(t *testing.T) {
	_, priv := mustTestKeys(t)
	sign := func(e Entry) Entry {
		e.Signature = ed25519.Sign(priv, e.CanonicalBytes())
		return e
	}
	// Sub-second ValidFrom: rejected, not truncated.
	g := sign(testGrantEntry())
	g.ValidFrom = g.ValidFrom + 123456789
	if _, err := MarshalEntry(g); !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with sub-second ValidFrom: got %v, want errors.Is ErrMalformed", err)
	}
	// Sub-second ValidUntil: rejected.
	g2 := sign(testGrantEntry())
	g2.ValidUntil = g2.ValidUntil + 500
	if _, err := MarshalEntry(g2); !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with sub-second ValidUntil: got %v, want errors.Is ErrMalformed", err)
	}
	// Sub-second AppendedAt: rejected.
	g3 := sign(testGrantEntry())
	g3.AppendedAt = g3.AppendedAt + 1
	if _, err := MarshalEntry(g3); !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with sub-second AppendedAt: got %v, want errors.Is ErrMalformed", err)
	}
	// Sub-second EffectiveAt on a revoke: rejected.
	r := sign(testRevokeEntry())
	r.EffectiveAt = r.EffectiveAt + 500
	if _, err := MarshalEntry(r); !errors.Is(err, ErrMalformed) {
		t.Fatalf("revoke with sub-second EffectiveAt: got %v, want errors.Is ErrMalformed", err)
	}
	// Whole-second entries emit the strict UTC Z spelling on the wire.
	g4 := sign(testGrantEntry())
	line, err := MarshalEntry(g4)
	if err != nil {
		t.Fatalf("MarshalEntry: %v", err)
	}
	for _, want := range []string{
		`"valid_from":"2026-10-21T22:00:01Z"`,
		`"valid_until":"2027-12-31T23:59:59Z"`,
		`"appended_at":"2026-10-21T22:00:01Z"`,
	} {
		if !strings.Contains(string(line), want) {
			t.Fatalf("grant line must contain %s: %s", want, line)
		}
	}
	r2 := sign(testRevokeEntry())
	rline, err := MarshalEntry(r2)
	if err != nil {
		t.Fatalf("MarshalEntry(revoke): %v", err)
	}
	for _, want := range []string{
		`"effective_at":"2026-10-22T00:00:01Z"`,
		`"appended_at":"2026-10-22T00:00:01Z"`,
	} {
		if !strings.Contains(string(rline), want) {
			t.Fatalf("revoke line must contain %s: %s", want, rline)
		}
	}
}

// ---------------------------------------------------------------------------
// Fix 5 — Revoke cross-field check. Decision 0015: effective_at may not
// precede the revocation's authenticated appended_at.
// ---------------------------------------------------------------------------

func TestParseEntryRevokeEffectiveAtBeforeAppendedAtRejected(t *testing.T) {
	base := testRevokeLine() // effective_at == appended_at == 2026-10-22T00:00:01Z
	// effective_at before appended_at: rejected.
	early := strings.Replace(base, `"effective_at":"2026-10-22T00:00:01Z"`, `"effective_at":"2026-10-21T23:59:59Z"`, 1)
	if _, err := ParseEntry([]byte(early)); !errors.Is(err, ErrMalformed) {
		t.Fatalf("revoke effective_at before appended_at: got %v, want errors.Is ErrMalformed", err)
	}
	// effective_at exactly equal to appended_at is not earlier: accepted.
	if _, err := ParseEntry([]byte(base)); err != nil {
		t.Fatalf("revoke effective_at == appended_at: unexpected rejection: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Fix 4 — valid_until mandatory on grants. Decision 0015 requires not_after
// on every grant: no open-ended grants. ParseEntry rejects an absent/null
// valid_until on a grant line; MarshalEntry rejects an Entry with
// ValidUntil == 0 instead of omitting the field.
// ---------------------------------------------------------------------------

func TestParseEntryGrantMissingValidUntilRejected(t *testing.T) {
	base := testGrantLine()
	variants := map[string]string{
		"absent": strings.Replace(base, `"valid_until":"2027-12-31T23:59:59Z",`, ``, 1),
		"null":   strings.Replace(base, `"valid_until":"2027-12-31T23:59:59Z"`, `"valid_until":null`, 1),
	}
	for name, line := range variants {
		if _, err := ParseEntry([]byte(line)); !errors.Is(err, ErrMalformed) {
			t.Fatalf("grant valid_until %s: got %v, want errors.Is ErrMalformed", name, err)
		}
	}
	// A grant line WITH valid_until parses.
	if _, err := ParseEntry([]byte(base)); err != nil {
		t.Fatalf("grant with valid_until: unexpected rejection: %v", err)
	}
}

func TestMarshalEntryGrantMissingValidUntilRejected(t *testing.T) {
	_, priv := mustTestKeys(t)
	g := testGrantEntry()
	g.ValidUntil = 0
	g.Signature = ed25519.Sign(priv, g.CanonicalBytes())
	if _, err := MarshalEntry(g); !errors.Is(err, ErrMalformed) {
		t.Fatalf("grant with ValidUntil==0: got %v, want errors.Is ErrMalformed", err)
	}
}

// ---------------------------------------------------------------------------
// Fix 2 — Retroactive grants fail closed. Decision 0015: not_before may not
// predate the grant's authenticated append time; this format implements no
// retroactive_grant_exception entry type.
// ---------------------------------------------------------------------------

func TestParseEntryGrantRetroactiveValidFromRejected(t *testing.T) {
	base := testGrantLine() // valid_from == appended_at == 2026-10-21T22:00:01Z
	// valid_from one hour before appended_at: retroactive. Rejected.
	retro := strings.Replace(base, `"valid_from":"2026-10-21T22:00:01Z"`, `"valid_from":"2026-10-21T21:00:01Z"`, 1)
	if _, err := ParseEntry([]byte(retro)); !errors.Is(err, ErrMalformed) {
		t.Fatalf("retroactive valid_from: got %v, want errors.Is ErrMalformed", err)
	}
	// valid_from exactly equal to appended_at is not retroactive: accepted.
	if _, err := ParseEntry([]byte(base)); err != nil {
		t.Fatalf("valid_from == appended_at: unexpected rejection: %v", err)
	}
}
