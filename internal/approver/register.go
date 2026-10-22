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
type Entry struct {
	Version    uint8
	Seq        uint64
	PrevHash   [32]byte
	Type       uint8
	Identity   string
	Role       string
	ValidFrom  int64
	ValidUntil int64 // 0 = open-ended (grant only)
	RevokedAt  int64
	AppendedAt int64
	Signature  []byte
}

// CanonicalBytes returns the exact byte encoding signed by the trust anchor:
// Version(1B), Seq(8B BE), PrevHash(32B), Type(1B), Identity(uvarint len +
// bytes), Role(uvarint len + bytes), ValidFrom(8B BE), ValidUntil(8B BE),
// RevokedAt(8B BE), AppendedAt(8B BE). No JSON, no ambiguity; the Signature
// is excluded.
func (e Entry) CanonicalBytes() []byte {
	b := make([]byte, 0, 128)
	b = append(b, e.Version)
	b = binary.BigEndian.AppendUint64(b, e.Seq)
	b = append(b, e.PrevHash[:]...)
	b = append(b, e.Type)
	b = binary.AppendUvarint(b, uint64(len(e.Identity)))
	b = append(b, e.Identity...)
	b = binary.AppendUvarint(b, uint64(len(e.Role)))
	b = append(b, e.Role...)
	b = binary.BigEndian.AppendUint64(b, uint64(e.ValidFrom))
	b = binary.BigEndian.AppendUint64(b, uint64(e.ValidUntil))
	b = binary.BigEndian.AppendUint64(b, uint64(e.RevokedAt))
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
// timestamp; unset encodes as null.
type wireEntry struct {
	V          int     `json:"v"`
	Seq        uint64  `json:"seq"`
	Prev       string  `json:"prev"`
	Type       string  `json:"type"`
	Identity   string  `json:"identity"`
	Role       string  `json:"role"`
	ValidFrom  *string `json:"valid_from"`
	ValidUntil *string `json:"valid_until"`
	RevokedAt  *string `json:"revoked_at"`
	AppendedAt *string `json:"appended_at"`
	Sig        string  `json:"sig"`
}

// parseWireEntry differs from wireEntry only in Seq: a missing seq must be
// rejected, so absence is detected with a pointer.
type parseWireEntry struct {
	V          int     `json:"v"`
	Seq        *uint64 `json:"seq"`
	Prev       string  `json:"prev"`
	Type       string  `json:"type"`
	Identity   string  `json:"identity"`
	Role       string  `json:"role"`
	ValidFrom  *string `json:"valid_from"`
	ValidUntil *string `json:"valid_until"`
	RevokedAt  *string `json:"revoked_at"`
	AppendedAt *string `json:"appended_at"`
	Sig        string  `json:"sig"`
}

func wireTime(nanos int64) string {
	return time.Unix(0, nanos).UTC().Format(time.RFC3339Nano)
}

func strptr(s string) *string { return &s }

// MarshalEntry encodes e as one JSONL line in the pinned wire encoding.
// Time fields become RFC3339Nano UTC strings; unset fields encode as null.
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
	w := wireEntry{
		V:          int(e.Version),
		Seq:        e.Seq,
		Prev:       hex.EncodeToString(e.PrevHash[:]),
		Type:       typ,
		Identity:   e.Identity,
		Role:       e.Role,
		AppendedAt: strptr(wireTime(e.AppendedAt)),
		Sig:        hex.EncodeToString(e.Signature),
	}
	switch e.Type {
	case EntryGrant:
		w.ValidFrom = strptr(wireTime(e.ValidFrom))
		if e.ValidUntil != 0 {
			w.ValidUntil = strptr(wireTime(e.ValidUntil))
		}
	case EntryRevoke:
		w.RevokedAt = strptr(wireTime(e.RevokedAt))
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
	tm, err := time.Parse(time.RFC3339Nano, *s)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrMalformed, name, err)
	}
	return tm.UTC().UnixNano(), nil
}

// ParseEntry parses and validates one JSONL line into an Entry. It enforces
// the pinned wire rules: v must be 1, type exactly "grant"/"revoke", prev
// 64-char lowercase hex, sig 128-char lowercase hex, unknown fields
// rejected, and per-type required/forbidden fields. Any violation returns a
// non-nil error wrapping ErrMalformed.
func ParseEntry(line []byte) (Entry, error) {
	var e Entry
	fail := func(format string, args ...any) (Entry, error) {
		return e, fmt.Errorf("%w: %s", ErrMalformed, fmt.Sprintf(format, args...))
	}

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
	revokedAt, err := parseWireTime(w.RevokedAt, "revoked_at")
	if err != nil {
		return e, err
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
		if w.RevokedAt != nil {
			return fail("grant must not carry revoked_at")
		}
	case EntryRevoke:
		if w.RevokedAt == nil {
			return fail("revoke requires revoked_at")
		}
		if w.ValidFrom != nil {
			return fail("revoke must not carry valid_from")
		}
		if w.ValidUntil != nil {
			return fail("revoke must not carry valid_until")
		}
	}

	e = Entry{
		Version:    FormatVersion,
		Seq:        *w.Seq,
		Type:       typ,
		Identity:   w.Identity,
		Role:       w.Role,
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
		RevokedAt:  revokedAt,
		AppendedAt: appendedAt,
		Signature:  sig,
	}
	copy(e.PrevHash[:], prev)
	return e, nil
}
