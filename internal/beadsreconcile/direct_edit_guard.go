package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ValidateDirectEdit validates a single .beads/issues.jsonl export. It fails
// closed: the bytes must parse as JSONL, every record must be well-formed
// with required fields, titles may not suffer null loss, and dependency
// counts must match the number of blocks edges. It does not compare a
// candidate to trusted source bytes; ValidateDirectEditTransition does that
// and permits only the separately pinned D2 reopen.
func ValidateDirectEdit(source []byte) error {
	seen := make(map[string]int) // id -> line number
	lines := bytes.Split(source, []byte("\n"))
	for i, line := range lines {
		lineno := i + 1
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		// Reject duplicate keys before unmarshalling:
		// encoding/json would silently keep the last value, letting a
		// duplicate key mask a null title from the null-loss check below.
		if err := rejectDuplicateKeys(trimmed); err != nil {
			return fmt.Errorf("line %d: %w", lineno, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &fields); err != nil {
			return fmt.Errorf("line %d: invalid JSON: %w", lineno, err)
		}
		// Required: id must be a non-empty string.
		rawID, ok := fields["id"]
		if !ok {
			return fmt.Errorf("line %d: missing required field \"id\"", lineno)
		}
		var id string
		if err := json.Unmarshal(rawID, &id); err != nil || id == "" {
			return fmt.Errorf("line %d: field \"id\" must be a non-empty string", lineno)
		}
		if prev, dup := seen[id]; dup {
			return fmt.Errorf("line %d: duplicate id %q (first seen on line %d)", lineno, id, prev)
		}
		seen[id] = lineno
		// Null-loss: title must not be null when present.
		if raw, ok := fields["title"]; ok && string(bytes.TrimSpace(raw)) == "null" {
			return fmt.Errorf("line %d: field \"title\" must not be null (id %q)", lineno, id)
		}
		// Dependency count consistency: dependency_count must equal the number
		// of "blocks" edges exactly. The export convention lists all edges in
		// dependencies but counts only blocks edges (parent-child and tracks
		// edges are excluded); a stale or negative count is a malformed edit.
		if rawCount, ok := fields["dependency_count"]; ok {
			if string(bytes.TrimSpace(rawCount)) == "null" {
				return fmt.Errorf("line %d: field \"dependency_count\" must be an integer (id %q)", lineno, id)
			}
			var count int
			if err := json.Unmarshal(rawCount, &count); err != nil {
				return fmt.Errorf("line %d: field \"dependency_count\" must be an integer (id %q)", lineno, id)
			}
			blocks := 0
			if rawDeps, ok := fields["dependencies"]; ok && string(bytes.TrimSpace(rawDeps)) != "null" {
				var deps []map[string]json.RawMessage
				if err := json.Unmarshal(rawDeps, &deps); err != nil {
					return fmt.Errorf("line %d: field \"dependencies\" must be an array of objects (id %q)", lineno, id)
				}
				for j, dep := range deps {
					if dep == nil {
						return fmt.Errorf("line %d: dependency %d is not an object (id %q)", lineno, j, id)
					}
					rawType, ok := dep["type"]
					if !ok {
						return fmt.Errorf("line %d: dependency %d missing \"type\" (id %q)", lineno, j, id)
					}
					var typ string
					if err := json.Unmarshal(rawType, &typ); err != nil || strings.TrimSpace(typ) == "" {
						return fmt.Errorf("line %d: dependency %d has invalid \"type\" (id %q)", lineno, j, id)
					}
					if typ == "blocks" {
						blocks++
					}
				}
			}
			if count != blocks {
				return fmt.Errorf("line %d: dependency_count %d != blocks edge count %d (id %q)", lineno, count, blocks, id)
			}
		}
	}
	if len(seen) == 0 {
		return fmt.Errorf("no records found")
	}
	return nil
}

// ValidateCanonicalExport checks the mandatory fields emitted in every
// tracked Beads issue row. A few legacy rows omit newer optional fields such
// as _type, created_by, comment_count, labels, or dependencies, so those are
// not promoted to mandatory here. Explicit-null dependencies remain
// compatible with the repository's existing source-specific approvals.
func validateCanonicalExport(source []byte) error {
	if err := ValidateDirectEdit(source); err != nil {
		return err
	}
	rows, err := parseDirectEditRows(source)
	if err != nil {
		return err
	}
	if err := validateCanonicalOptionalFields(rows); err != nil {
		return err
	}
	for id, row := range rows {
		for _, field := range []string{"title", "description", "acceptance_criteria", "status", "issue_type", "created_at", "updated_at"} {
			raw, ok := row.fields[field]
			if !ok {
				return fmt.Errorf("issue %q missing canonical field %q", id, field)
			}
			if string(bytes.TrimSpace(raw)) == "null" {
				return fmt.Errorf("issue %q canonical field %q must not be null", id, field)
			}
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("issue %q canonical field %q must be a string", id, field)
			}
			if value == "" && (field == "title" || field == "status" || field == "issue_type" || field == "created_at" || field == "updated_at") {
				return fmt.Errorf("issue %q canonical field %q must be non-empty", id, field)
			}
			if field == "created_at" || field == "updated_at" {
				if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
					return fmt.Errorf("issue %q canonical field %q must be an RFC3339 timestamp", id, field)
				}
			}
		}
		for _, field := range []string{"priority", "dependency_count", "dependent_count"} {
			raw, ok := row.fields[field]
			if !ok {
				return fmt.Errorf("issue %q missing canonical field %q", id, field)
			}
			if string(bytes.TrimSpace(raw)) == "null" {
				return fmt.Errorf("issue %q canonical field %q must not be null", id, field)
			}
			var value int
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("issue %q canonical field %q must be an integer", id, field)
			}
			if field == "priority" && (value < 0 || value > 4) {
				return fmt.Errorf("issue %q canonical field %q must be in the Beads range 0-4", id, field)
			}
			if field != "priority" && value < 0 {
				return fmt.Errorf("issue %q canonical field %q must not be negative", id, field)
			}
		}
	}
	return validateCanonicalDependencyGraph(rows)
}

// validateCanonicalOptionalFields checks the supported optional fields emitted
// by the pinned Beads 1.3.1 export contract when they are present. Omission is
// valid. Historical explicit nulls are supported only for closed_at and
// dependencies; other optional fields retain their observed JSON types.
func validateCanonicalOptionalFields(rows map[string]directEditRow) error {
	stringFields := []string{
		"_type", "assignee", "close_reason", "created_by", "heartbeat_at",
		"lease_expires_at", "notes", "owner", "started_at",
	}
	for id, row := range rows {
		for _, field := range stringFields {
			if raw, ok := row.fields[field]; ok {
				var value string
				if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
					return fmt.Errorf("issue %q optional field %q must be a string", id, field)
				}
			}
		}
		if raw, ok := row.fields["closed_at"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("issue %q optional field \"closed_at\" must be a string or null", id)
			}
		}
		if raw, ok := row.fields["comment_count"]; ok {
			var value int
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
				return fmt.Errorf("issue %q optional field \"comment_count\" must be an integer", id)
			}
		}
		if raw, ok := row.fields["dependencies"]; ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			var value []json.RawMessage
			if err := json.Unmarshal(raw, &value); err != nil || value == nil {
				return fmt.Errorf("issue %q optional field \"dependencies\" must be an array or null", id)
			}
		}
		if raw, ok := row.fields["labels"]; ok {
			var labels []json.RawMessage
			if err := json.Unmarshal(raw, &labels); err != nil || labels == nil {
				return fmt.Errorf("issue %q optional field \"labels\" must be an array of strings", id)
			}
			for index, rawLabel := range labels {
				var label string
				if bytes.Equal(bytes.TrimSpace(rawLabel), []byte("null")) || json.Unmarshal(rawLabel, &label) != nil {
					return fmt.Errorf("issue %q optional field \"labels\" element %d must be a string", id, index)
				}
			}
		}
	}
	return nil
}

// validateCanonicalDependencyGraph reuses validateGraph for edge structure,
// ownership, targets, and uniqueness, then verifies the incoming blocks count
// recorded on every row. Other edge types do not contribute to dependent_count.
func validateCanonicalDependencyGraph(rows map[string]directEditRow) error {
	graph := make(map[string]row, len(rows))
	for id, item := range rows {
		graph[id] = row{fields: item.fields, line: item.line, id: id}
	}
	if err := validateGraph(graph); err != nil {
		return fmt.Errorf("canonical dependency graph: %w", err)
	}
	incomingBlocks := make(map[string]int, len(rows))
	for id, item := range rows {
		raw, ok := item.fields["dependencies"]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		var dependencies []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &dependencies); err != nil {
			return fmt.Errorf("issue %q dependencies must be an array", id)
		}
		for _, dependency := range dependencies {
			typ, _ := stringFieldMap(dependency, "type")
			if typ != "blocks" {
				continue
			}
			target, _ := stringFieldMap(dependency, "depends_on_id")
			incomingBlocks[target]++
		}
	}
	for id, item := range rows {
		var count int
		if err := json.Unmarshal(item.fields["dependent_count"], &count); err != nil {
			return fmt.Errorf("issue %q canonical field \"dependent_count\" must be an integer", id)
		}
		if want := incomingBlocks[id]; count != want {
			return fmt.Errorf("issue %q dependent_count %d != incoming blocks edge count %d", id, count, want)
		}
	}
	return nil
}
