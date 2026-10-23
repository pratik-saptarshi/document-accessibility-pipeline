// Package approver implements the approver-role register entry format from
// Decision 0019: a versioned, hash-chained JSONL entry format. Each entry
// carries a sequence number, the hash of the previous entry, a grant or
// revoke payload, append timestamps, and an Ed25519 signature over the
// canonical bytes.
//
// The register file is untrusted input. Signatures authenticate the issuer;
// the public key is always a caller-provided parameter and is never parsed
// from the register file.
package approver

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"
)

// Entry types.
const (
	EntryGrant  uint8 = 0x01
	EntryRevoke uint8 = 0x02
)

// FormatVersion is the only accepted register format version.
const FormatVersion uint8 = 1

// ErrMalformed reports a structurally invalid register entry. All parse
// failures wrap this sentinel so callers can fail closed on any bad line.
var ErrMalformed = errors.New("approver: malformed register entry")

// Entry is one register record. Time fields are int64 unix nanoseconds
// derived from UTC timestamps; Signature is excluded from CanonicalBytes.
// Identity/Role/ValidFrom/ValidUntil are grant-only fields. ValidUntil is
// mandatory on grants — there are no open-ended grants (Decision 0015), so
// ValidUntil == 0 is rejected by both MarshalEntry and ParseEntry.
// RevokesSequence and EffectiveAt are revoke-only fields: a revocation
// names exactly one grant by its sequence number and applies prospectively
// from its effective_at, inclusive (Decision 0015). Revoke entries must
// carry empty Identity/Role.
type Entry struct {
	Version         uint8
	Seq             uint64
	PrevHash        [32]byte
	Type            uint8
	Identity        string
	Role            string
	ValidFrom       int64
	ValidUntil      int64 // mandatory on grants; no open-ended grants (Decision 0015)
	RevokesSequence uint64
	EffectiveAt     int64
	AppendedAt      int64
	Signature       []byte
}

// CanonicalBytes returns the exact byte encoding signed by the trust anchor.
// Grants: Version(1B), Seq(8B BE), PrevHash(32B), Type(1B),
// Identity(uvarint len + bytes), Role(uvarint len + bytes), ValidFrom(8B BE),
// ValidUntil(8B BE), AppendedAt(8B BE). Revokes: Version(1B), Seq(8B BE),
// PrevHash(32B), Type(1B), RevokesSequence(8B BE), EffectiveAt(8B BE),
// AppendedAt(8B BE) — no identity, role, or validity window. No JSON, no
// ambiguity; the Signature is excluded.
func (e Entry) CanonicalBytes() []byte {
	b := make([]byte, 0, 128)
	b = append(b, e.Version)
	b = binary.BigEndian.AppendUint64(b, e.Seq)
	b = append(b, e.PrevHash[:]...)
	b = append(b, e.Type)
	if e.Type == EntryRevoke {
		// Revocations name their target grant by sequence alone; identity
		// and role are not part of the signed bytes. Unknown entry types are
		// rejected before signing/verification, so any non-revoke takes the
		// grant layout.
		b = binary.BigEndian.AppendUint64(b, e.RevokesSequence)
		b = binary.BigEndian.AppendUint64(b, uint64(e.EffectiveAt))
		b = binary.BigEndian.AppendUint64(b, uint64(e.AppendedAt))
		return b
	}
	b = binary.AppendUvarint(b, uint64(len(e.Identity)))
	b = append(b, e.Identity...)
	b = binary.AppendUvarint(b, uint64(len(e.Role)))
	b = append(b, e.Role...)
	b = binary.BigEndian.AppendUint64(b, uint64(e.ValidFrom))
	b = binary.BigEndian.AppendUint64(b, uint64(e.ValidUntil))
	b = binary.BigEndian.AppendUint64(b, uint64(e.AppendedAt))
	return b
}

// ChainHash computes the hash identifying an entry in the chain: SHA-256
// over the previous entry's chain hash concatenated with this entry's
// canonical bytes. The genesis entry uses an all-zero PrevHash.
func ChainHash(prev [32]byte, canonical []byte) [32]byte {
	buf := make([]byte, 0, len(prev)+len(canonical))
	buf = append(buf, prev[:]...)
	buf = append(buf, canonical...)
	return sha256.Sum256(buf)
}

// wireEntry is the JSONL object encoding, with fields in wire order.
// Time fields are *string so JSON null/absent is distinguishable from a
// timestamp; unset encodes as null (though ParseEntry rejects null/absent
// for fields the entry type requires). Grant lines carry identity, role,
// valid_from, and valid_until. Revoke lines carry revokes_sequence and
// effective_at instead; their identity, role, valid_from, valid_until
// fields encode as empty/null and must be empty/null.
type wireEntry struct {
	V               int     `json:"v"`
	Seq             uint64  `json:"seq"`
	Prev            string  `json:"prev"`
	Type            string  `json:"type"`
	Identity        string  `json:"identity"`
	Role            string  `json:"role"`
	ValidFrom       *string `json:"valid_from"`
	ValidUntil      *string `json:"valid_until"`
	RevokesSequence *uint64 `json:"revokes_sequence"`
	EffectiveAt     *string `json:"effective_at"`
	AppendedAt      *string `json:"appended_at"`
	Sig             string  `json:"sig"`
}

// parseWireEntry differs from wireEntry only in Seq: a missing seq must be
// rejected, so absence is detected with a pointer. RevokesSequence uses a
// pointer here too so a missing revokes_sequence is distinguishable from
// the legitimate target sequence 0.
type parseWireEntry struct {
	V               int     `json:"v"`
	Seq             *uint64 `json:"seq"`
	Prev            string  `json:"prev"`
	Type            string  `json:"type"`
	Identity        string  `json:"identity"`
	Role            string  `json:"role"`
	ValidFrom       *string `json:"valid_from"`
	ValidUntil      *string `json:"valid_until"`
	RevokesSequence *uint64 `json:"revokes_sequence"`
	EffectiveAt     *string `json:"effective_at"`
	AppendedAt      *string `json:"appended_at"`
	Sig             string  `json:"sig"`
}

// hasDuplicateTopLevelKeys reports whether the raw JSON object in line
// repeats a top-level key. encoding/json silently keeps one value for a
// repeated key and DisallowUnknownFields does not catch repeats, so this
// lightweight token scan runs before the struct decode. Decision 0015
// explicitly requires rejecting duplicate keys.
//
// Only top-level keys are scanned: each value is consumed as a RawMessage,
// so a duplicate inside a nested object value cannot be seen here. The
// wire format has no nested objects, so a nested object is rejected by the
// unknown-field rule anyway. Key comparison uses the decoded string, so an
// escaped spelling of the same key ("\u0073eq" vs "seq") is a duplicate.
//
// Non-object or invalid input returns false; the struct decode in
// ParseEntry reports those errors in its own terms.
func hasDuplicateTopLevelKeys(line []byte) bool {
	dec := json.NewDecoder(bytes.NewReader(line))
	tok, err := dec.Token()
	if err != nil {
		return false
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return false
	}
	seen := make(map[string]struct{})
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return false
		}
		key, ok := keyTok.(string)
		if !ok {
			return false
		}
		if _, dup := seen[key]; dup {
			return true
		}
		seen[key] = struct{}{}
		// Consume the value as raw bytes without inspecting it: nested
		// objects are out of scope for this scan.
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return false
		}
	}
	return false
}

// UnixNano representability bounds. time.Time.UnixNano is undefined when
// the nanosecond value cannot fit in an int64 — past that point it wraps
// (empirically confirmed in TestUnixNanoWrapsOutsideRange), destroying
// time ordering. Wire and query timestamps outside the extreme whole
// seconds whose nanosecond value is exactly representable are REJECTED,
// never clamped. Bounds verified: 9223372036  = 2262-04-11T23:47:16Z (max),
// -9223372036 = 1677-09-21T00:12:44Z (min).
const (
	maxUnixNanoSec = 9223372036
	minUnixNanoSec = -9223372036
)

// strictWireTime is the only permitted wire timestamp spelling per Decision
// 0015: RFC 3339 UTC "Z" with whole-second precision. This is checked before
// time.Parse, which would otherwise accept numeric offsets and fractional
// seconds.
var strictWireTime = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)

// wireTime renders nanos in the strict wire spelling: whole-second UTC Z
// per Decision 0015. It returns an ErrMalformed-wrapped error — never
// silently truncating — when nanos carries sub-second precision.
//
// No UnixNano-range check is needed here (deliberate scoping): wireTime
// performs no UnixNano conversion — it renders time.Unix(nanos/1e9, 0) —
// and truncated int64 division always lands within ±9223372036 whole
// seconds, so it can never emit wire bytes the fixed parser rejects.
// parseWireTime remains the single enforcement point for the range.
func wireTime(nanos int64, name string) (string, error) {
	if nanos%1e9 != 0 {
		return "", fmt.Errorf("%w: %s carries sub-second precision; wire time is whole-second UTC Z", ErrMalformed, name)
	}
	return time.Unix(nanos/1e9, 0).UTC().Format(time.RFC3339), nil
}

func strptr(s string) *string { return &s }

// MarshalEntry encodes e as one JSONL line in the pinned wire encoding.
// Time fields become strict whole-second UTC "Z" strings per Decision 0015;
// an entry carrying sub-second nanos is rejected with an
// ErrMalformed-wrapped error (fail loud, no silent truncation). Unset fields
// encode as null, but grants must always carry valid_until: a grant with
// ValidUntil == 0 is rejected instead of emitting an open-ended line.
// Revoke entries emit revokes_sequence and effective_at; they must have a
// non-zero effective_at and empty identity/role, or marshaling fails.
// Grant entries must not carry revokes_sequence or effective_at, or
// marshaling fails with an ErrMalformed-wrapped error — symmetric with
// ParseEntry, which rejects those fields on grant lines.
func MarshalEntry(e Entry) ([]byte, error) {
	if e.Version != FormatVersion {
		return nil, fmt.Errorf("%w: unsupported version %d", ErrMalformed, e.Version)
	}
	var typ string
	switch e.Type {
	case EntryGrant:
		typ = "grant"
	case EntryRevoke:
		typ = "revoke"
	default:
		return nil, fmt.Errorf("%w: unknown entry type %#02x", ErrMalformed, e.Type)
	}
	if len(e.Signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("%w: signature must be %d bytes, got %d",
			ErrMalformed, ed25519.SignatureSize, len(e.Signature))
	}
	appendedAt, err := wireTime(e.AppendedAt, "appended_at")
	if err != nil {
		return nil, err
	}
	w := wireEntry{
		V:          int(e.Version),
		Seq:        e.Seq,
		Prev:       hex.EncodeToString(e.PrevHash[:]),
		Type:       typ,
		Identity:   e.Identity,
		Role:       e.Role,
		AppendedAt: strptr(appendedAt),
		Sig:        hex.EncodeToString(e.Signature),
	}
	switch e.Type {
	case EntryGrant:
		// A grant carrying revoke-only fields is malformed; reject it
		// rather than silently dropping the fields (which would desync
		// the Entry from its wire encoding).
		if e.RevokesSequence != 0 {
			return nil, fmt.Errorf("%w: grant must not carry revokes_sequence", ErrMalformed)
		}
		if e.EffectiveAt != 0 {
			return nil, fmt.Errorf("%w: grant must not carry effective_at", ErrMalformed)
		}
		validFrom, err := wireTime(e.ValidFrom, "valid_from")
		if err != nil {
			return nil, err
		}
		w.ValidFrom = strptr(validFrom)
		// Decision 0015: not_after is mandatory on every grant. Reject a
		// zero ValidUntil rather than omitting the field (fail loud; no
		// open-ended grants on the wire).
		if e.ValidUntil == 0 {
			return nil, fmt.Errorf("%w: grant requires non-zero valid_until", ErrMalformed)
		}
		validUntil, err := wireTime(e.ValidUntil, "valid_until")
		if err != nil {
			return nil, err
		}
		w.ValidUntil = strptr(validUntil)
	case EntryRevoke:
		if e.EffectiveAt == 0 {
			return nil, fmt.Errorf("%w: revoke requires non-zero effective_at", ErrMalformed)
		}
		if e.Identity != "" || e.Role != "" {
			return nil, fmt.Errorf("%w: revoke must not carry identity/role", ErrMalformed)
		}
		effectiveAt, err := wireTime(e.EffectiveAt, "effective_at")
		if err != nil {
			return nil, err
		}
		w.RevokesSequence = &e.RevokesSequence
		w.EffectiveAt = strptr(effectiveAt)
	}
	return json.Marshal(w)
}

// isLowerHex reports whether s consists solely of lowercase hex digits.
func isLowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' {
			continue
		}
		return false
	}
	return true
}

func decodeHashField(s string, wantLen int, name string) ([]byte, error) {
	if len(s) != 2*wantLen || !isLowerHex(s) {
		return nil, fmt.Errorf("%w: %s must be %d lowercase hex chars",
			ErrMalformed, name, 2*wantLen)
	}
	raw, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrMalformed, name, err)
	}
	return raw, nil
}

func parseWireTime(s *string, name string) (int64, error) {
	if s == nil {
		return 0, nil
	}
	// The spelling check runs before time.Parse: offsets and fractions are
	// rejected even though Go's parser would accept them.
	if !strictWireTime.MatchString(*s) {
		return 0, fmt.Errorf("%w: %s %q is not strict whole-second UTC Z", ErrMalformed, name, *s)
	}
	tm, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrMalformed, name, err)
	}
	// UnixNano would wrap outside ±9223372036 whole seconds, silently
	// destroying time ordering. Reject (never clamp) so hostile timestamps
	// cannot smuggle a wrapped value past the temporal checks.
	if sec := tm.UTC().Unix(); sec < minUnixNanoSec || sec > maxUnixNanoSec {
		return 0, fmt.Errorf("%w: %s %q is outside the representable unix-nanosecond range", ErrMalformed, name, *s)
	}
	return tm.UTC().UnixNano(), nil
}

// ParseEntry parses and validates one JSONL line into an Entry. It enforces
// the pinned wire rules: no duplicate JSON object keys, v must be 1, type
// exactly "grant"/"revoke", prev 64-char lowercase hex, sig 128-char
// lowercase hex, unknown fields rejected, per-type required/forbidden
// fields, and the Decision 0015 temporal rules — strict whole-second UTC
// "Z" timestamp spelling on all four wire time fields, wire timestamps
// within the representable unix-nanosecond range (±9223372036 whole
// seconds; out-of-range times are rejected, never clamped), mandatory
// appended_at on every entry, mandatory valid_until on grants (no
// open-ended grants), grant valid_from not predating appended_at (no
// retroactive grants; no exception entry type exists in this format), and
// revoke effective_at not predating appended_at, and grant valid_from not
// after valid_until (an inverted window can never cover any query time).
// Any violation returns a non-nil error wrapping ErrMalformed.
func ParseEntry(line []byte) (Entry, error) {
	var e Entry
	fail := func(format string, args ...any) (Entry, error) {
		return e, fmt.Errorf("%w: %s", ErrMalformed, fmt.Sprintf(format, args...))
	}

	// Decision 0015 rejects duplicate JSON object keys. encoding/json
	// silently keeps one value for a repeated key and DisallowUnknownFields
	// does not catch repeats, so the scan runs on the raw line bytes before
	// the struct decode below.
	if hasDuplicateTopLevelKeys(line) {
		return fail("duplicate JSON object key")
	}

	// Decision 0024 (accepted 2026-10-23): case-variant JSON field names
	// are accepted and normalized by encoding/json into the same signed
	// field (liberal parsing). Duplicate exact-spelling keys are still
	// rejected above; this is about case variants only. Revisit on testing
	// evidence of interop or hostile-input harm.
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.DisallowUnknownFields()
	var w parseWireEntry
	if err := dec.Decode(&w); err != nil {
		return fail("invalid JSON: %v", err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return fail("trailing data after JSON object")
	}

	if w.V != int(FormatVersion) {
		return fail("unsupported version %d", w.V)
	}
	if w.Seq == nil {
		return fail("missing seq")
	}
	var typ uint8
	switch w.Type {
	case "grant":
		typ = EntryGrant
	case "revoke":
		typ = EntryRevoke
	default:
		return fail("unknown type %q", w.Type)
	}
	prev, err := decodeHashField(w.Prev, 32, "prev")
	if err != nil {
		return e, err
	}
	sig, err := decodeHashField(w.Sig, ed25519.SignatureSize, "sig")
	if err != nil {
		return e, err
	}
	validFrom, err := parseWireTime(w.ValidFrom, "valid_from")
	if err != nil {
		return e, err
	}
	validUntil, err := parseWireTime(w.ValidUntil, "valid_until")
	if err != nil {
		return e, err
	}
	effectiveAt, err := parseWireTime(w.EffectiveAt, "effective_at")
	if err != nil {
		return e, err
	}
	// Decision 0015: appended_at is source-authenticated and mandatory on
	// every entry. A missing/null value is malformed — there is no epoch
	// fallback.
	if w.AppendedAt == nil {
		return fail("missing appended_at")
	}
	appendedAt, err := parseWireTime(w.AppendedAt, "appended_at")
	if err != nil {
		return e, err
	}

	switch typ {
	case EntryGrant:
		if w.Identity == "" {
			return fail("grant requires non-empty identity")
		}
		if w.ValidFrom == nil {
			return fail("grant requires valid_from")
		}
		// Decision 0015 requires not_after on every grant: no open-ended
		// grants.
		if w.ValidUntil == nil {
			return fail("grant requires valid_until")
		}
		if w.RevokesSequence != nil {
			return fail("grant must not carry revokes_sequence")
		}
		if w.EffectiveAt != nil {
			return fail("grant must not carry effective_at")
		}
		// Decision 0015: retroactive grants are not permitted by default —
		// not_before may not predate the grant's authenticated append time.
		// This format implements no retroactive_grant_exception entry type.
		if validFrom < appendedAt {
			return fail("grant valid_from predates appended_at: retroactive grants not permitted")
		}
		// An inverted validity window can never cover any query time;
		// reject it as malformed rather than silently treating the grant
		// as never-valid. The closed interval ValidFrom <= at <=
		// ValidUntil authorizes at the exact boundary, so equality here
		// is well-formed.
		if validFrom > validUntil {
			return fail("grant valid_from is after valid_until")
		}
	case EntryRevoke:
		if w.RevokesSequence == nil {
			return fail("revoke requires revokes_sequence")
		}
		if w.EffectiveAt == nil {
			return fail("revoke requires effective_at")
		}
		if w.Identity != "" {
			return fail("revoke must not carry identity")
		}
		if w.Role != "" {
			return fail("revoke must not carry role")
		}
		if w.ValidFrom != nil {
			return fail("revoke must not carry valid_from")
		}
		if w.ValidUntil != nil {
			return fail("revoke must not carry valid_until")
		}
		// Decision 0015: effective_at may not precede the revocation's
		// authenticated appended_at.
		if effectiveAt < appendedAt {
			return fail("revoke effective_at predates appended_at")
		}
	}

	var revokesSequence uint64
	if w.RevokesSequence != nil {
		revokesSequence = *w.RevokesSequence
	}
	e = Entry{
		Version:         FormatVersion,
		Seq:             *w.Seq,
		Type:            typ,
		Identity:        w.Identity,
		Role:            w.Role,
		ValidFrom:       validFrom,
		ValidUntil:      validUntil,
		RevokesSequence: revokesSequence,
		EffectiveAt:     effectiveAt,
		AppendedAt:      appendedAt,
		Signature:       sig,
	}
	copy(e.PrevHash[:], prev)
	return e, nil
}
