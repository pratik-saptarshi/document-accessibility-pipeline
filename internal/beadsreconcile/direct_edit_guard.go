package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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
		// Reject duplicate or case-variant keys before unmarshalling:
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
