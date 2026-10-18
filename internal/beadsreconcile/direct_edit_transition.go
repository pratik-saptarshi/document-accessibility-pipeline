package beadsreconcile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	approvedD2SourceRowSHA256 = "818b70f260299049572a14d707e2de2d766d060fe95915917dd3fdde60ceafee"
	approvedD2IssueID         = "dap-jo9.1.6.7"
	approvedD2UpdatedAt       = "2026-10-18T15:55:00Z"
	approvedD2ReopenNote      = "\n\n2026-10-18 REOPENED by owner decision D2: the 09:00Z close preceded AC1/overall acceptance resolution (AC2f accepted 10:52Z, after close). The premature close does not stand; issue returns to open pending proper acceptance."
)

// ValidateDirectEditTransition verifies a JSONL candidate against trusted
// source bytes. All records, including absent-versus-null fields, must be
// preserved exactly except for the single source-row-pinned D2 reopen.
// This is not a general lifecycle transition API.
func ValidateDirectEditTransition(source, candidate []byte) error {
	if err := ValidateDirectEdit(source); err != nil {
		return fmt.Errorf("invalid trusted source: %w", err)
	}
	if err := ValidateDirectEdit(candidate); err != nil {
		return fmt.Errorf("invalid candidate: %w", err)
	}
	sourceRows, err := parseDirectEditRows(source)
	if err != nil {
		return fmt.Errorf("parse trusted source: %w", err)
	}
	candidateRows, err := parseDirectEditRows(candidate)
	if err != nil {
		return fmt.Errorf("parse candidate: %w", err)
	}
	if len(sourceRows) != len(candidateRows) {
		return errors.New("direct edit changed the issue ID set")
	}
	for id, before := range sourceRows {
		after, ok := candidateRows[id]
		if !ok {
			return fmt.Errorf("direct edit removed issue %q", id)
		}
		if id == approvedD2IssueID && !bytes.Equal(before.line, after.line) {
			if err := validateD2Reopen(before, after); err != nil {
				return err
			}
			continue
		}
		if err := compareDirectEditFields(id, before.fields, after.fields); err != nil {
			return err
		}
	}
	for id := range candidateRows {
		if _, ok := sourceRows[id]; !ok {
			return fmt.Errorf("direct edit added issue %q", id)
		}
	}
	return nil
}

type directEditRow struct {
	fields map[string]json.RawMessage
	line   []byte
}

func parseDirectEditRows(data []byte) (map[string]directEditRow, error) {
	rows := make(map[string]directEditRow)
	for i, line := range bytes.Split(data, []byte("\n")) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		if err := rejectDuplicateKeys(trimmed); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("line %d: expected a JSON object", i+1)
		}
		id, err := directEditStringField(fields, "id")
		if err != nil || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("line %d: missing string issue id", i+1)
		}
		if _, exists := rows[id]; exists {
			return nil, fmt.Errorf("line %d: duplicate issue id %q", i+1, id)
		}
		rows[id] = directEditRow{fields: fields, line: append([]byte(nil), trimmed...)}
	}
	if len(rows) == 0 {
		return nil, errors.New("no issue records found")
	}
	return rows, nil
}

func validateD2Reopen(before, after directEditRow) error {
	digest := sha256.Sum256(before.line)
	if got := hex.EncodeToString(digest[:]); got != approvedD2SourceRowSHA256 {
		return errors.New("dap-jo9.1.6.7 source row does not match the exact D2-authorized close")
	}
	allowed := map[string]bool{"status": true, "closed_at": true, "close_reason": true, "updated_at": true, "notes": true}
	for field, value := range before.fields {
		if allowed[field] {
			continue
		}
		if !sameDirectEditValue(value, after.fields[field]) {
			return fmt.Errorf("D2 reopen changed unrelated field %q on %s", field, approvedD2IssueID)
		}
	}
	for field := range after.fields {
		if _, existed := before.fields[field]; !existed && !allowed[field] {
			return fmt.Errorf("D2 reopen added unrelated field %q on %s", field, approvedD2IssueID)
		}
	}
	status, statusErr := directEditStringField(after.fields, "status")
	closedAt, closedAtErr := directEditStringField(before.fields, "closed_at")
	closeReason, reasonErr := directEditStringField(before.fields, "close_reason")
	updatedAt, updatedErr := directEditStringField(after.fields, "updated_at")
	note, noteErr := directEditStringField(after.fields, "notes")
	if statusErr != nil || status != "open" || closedAtErr != nil || closedAt != "2026-10-18T09:00:32Z" ||
		reasonErr != nil || closeReason == "" || updatedErr != nil || updatedAt != approvedD2UpdatedAt || noteErr != nil || note != approvedD2ReopenNote {
		return errors.New("dap-jo9.1.6.7 candidate does not match the exact D2 reopen record")
	}
	if !bytes.Equal(bytes.TrimSpace(after.fields["closed_at"]), []byte("null")) {
		return errors.New("D2 reopen must retain an explicit null closed_at value")
	}
	if !bytes.Equal(bytes.TrimSpace(after.fields["close_reason"]), []byte(`""`)) {
		return errors.New("D2 reopen must clear close_reason to the explicit empty string")
	}
	if oldAt, err := directEditStringField(before.fields, "updated_at"); err != nil || oldAt != "2026-10-18T09:00:32Z" {
		return errors.New("D2 source updated_at does not match the recorded close event")
	}
	return nil
}

func compareDirectEditFields(id string, before, after map[string]json.RawMessage) error {
	for field, value := range before {
		candidate, exists := after[field]
		if !exists || !sameDirectEditValue(value, candidate) {
			return fmt.Errorf("direct edit changed or removed field %q on %s", field, id)
		}
	}
	for field := range after {
		if _, exists := before[field]; !exists {
			return fmt.Errorf("direct edit added field %q on %s", field, id)
		}
	}
	return nil
}

func sameDirectEditValue(left, right json.RawMessage) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return bytes.Equal(bytes.TrimSpace(left), bytes.TrimSpace(right))
}

func directEditStringField(fields map[string]json.RawMessage, name string) (string, error) {
	raw, ok := fields[name]
	if !ok {
		return "", fmt.Errorf("field %q is absent", name)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("field %q is not a string", name)
	}
	return value, nil
}
