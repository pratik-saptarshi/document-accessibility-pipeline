// Package beadsreconcile validates three Beads exports and builds a candidate
// interchange file by splicing only explicitly planned values into the source.
package beadsreconcile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const evidenceVersion = 1

var allowedFields = map[string]bool{
	"title": true, "description": true, "acceptance_criteria": true, "notes": true,
}

// Plan describes the exact issue fields an operator authorized Beads to update.
type Plan struct {
	Updates []IssueUpdate `json:"updates"`
	// BDVersion pins narrowly accepted importer behavior. The sole derived-field
	// exception is only recognized for Beads 1.3.1.
	BDVersion string `json:"bd_version,omitempty"`
}

type IssueUpdate struct {
	ID     string            `json:"id"`
	Fields map[string]string `json:"fields"`
}

// Evidence binds the candidate to all three immutable exports and the plan.
type Evidence struct {
	Version         int              `json:"version"`
	Outcome         string           `json:"outcome"`
	SourceSHA256    string           `json:"source_sha256"`
	BaselineSHA256  string           `json:"baseline_sha256"`
	UpdatedSHA256   string           `json:"updated_sha256"`
	PlanSHA256      string           `json:"plan_sha256"`
	CandidateSHA256 string           `json:"candidate_sha256"`
	RoundTripSHA256 string           `json:"round_trip_sha256,omitempty"`
	Changes         []EvidenceChange `json:"changes"`
	IssueIDs        []string         `json:"issue_ids"`
	AllIssueIDs     []string         `json:"all_issue_ids"`
}

type EvidenceChange struct {
	ID        string            `json:"id"`
	Fields    map[string]string `json:"fields"`
	UpdatedAt string            `json:"updated_at"`
}

type row struct {
	fields map[string]json.RawMessage
	line   []byte
	id     string
}

// Reconcile validates source C, pre-update export B0, post-update export B1,
// then returns candidate bytes derived from C. No export bytes are copied into
// the candidate except planned strings and Beads' updated_at values.
func Reconcile(source, baseline, updated []byte, plan Plan) ([]byte, Evidence, error) {
	c, err := parseExport("source", source)
	if err != nil {
		return nil, Evidence{}, err
	}
	b0, err := parseExport("baseline", baseline)
	if err != nil {
		return nil, Evidence{}, err
	}
	b1, err := parseExport("updated", updated)
	if err != nil {
		return nil, Evidence{}, err
	}
	if err := validateIDs(c, b0, b1); err != nil {
		return nil, Evidence{}, err
	}
	updates, err := validatePlan(plan, c)
	if err != nil {
		return nil, Evidence{}, err
	}

	for id, cr := range c {
		if err := compareRows(cr, b0[id], id, nil, false, plan.BDVersion); err != nil {
			return nil, Evidence{}, fmt.Errorf("source to baseline issue %s: %w", id, err)
		}
		var allowed map[string]bool
		if update, ok := updates[id]; ok {
			allowed = make(map[string]bool, len(update.Fields))
			for name := range update.Fields {
				allowed[name] = true
			}
		}
		if err := compareRows(b0[id], b1[id], id, allowed, true, plan.BDVersion); err != nil {
			return nil, Evidence{}, fmt.Errorf("baseline to updated issue %s: %w", id, err)
		}
		if update, ok := updates[id]; ok {
			for name, want := range update.Fields {
				got, err := stringField(b1[id], name)
				if err != nil || got != want {
					return nil, Evidence{}, fmt.Errorf("issue %s planned field %q does not match requested value", id, name)
				}
				before, beforeErr := stringField(b0[id], name)
				if beforeErr == nil && before == want {
					return nil, Evidence{}, fmt.Errorf("issue %s planned field %q was not changed", id, name)
				}
			}
			oldAt, oldErr := stringField(b0[id], "updated_at")
			newAt, newErr := stringField(b1[id], "updated_at")
			if oldErr != nil || newErr != nil || oldAt == newAt {
				return nil, Evidence{}, fmt.Errorf("issue %s updated_at was not changed", id)
			}
			if _, err := parseTimestamp(newAt); err != nil {
				return nil, Evidence{}, fmt.Errorf("issue %s updated_at is invalid: %w", id, err)
			}
		}
	}

	lines := splitLines(source)
	changes := make([]EvidenceChange, 0, len(updates))
	ids := make([]string, 0, len(updates))
	for index, original := range lines {
		id, err := rowID(original)
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("source line %d: %w", index+1, err)
		}
		u, ok := updates[id]
		if !ok {
			continue
		}
		changesForID := make(map[string]string, len(u.Fields))
		for name, value := range u.Fields {
			changesForID[name] = value
		}
		updatedAt, err := stringField(b1[id], "updated_at")
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("issue %s: %w", id, err)
		}
		changed, err := spliceFields(original, changesForID, "updated_at", updatedAt)
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("issue %s: %w", id, err)
		}
		lines[index] = changed
		changes = append(changes, EvidenceChange{ID: id, Fields: changesForID, UpdatedAt: updatedAt})
		ids = append(ids, id)
	}
	candidate := bytes.Join(lines, []byte("\n"))
	if len(source) > 0 && source[len(source)-1] == '\n' {
		candidate = append(candidate, '\n')
	}
	e := Evidence{
		Version: evidenceVersion, Outcome: "reconciled",
		SourceSHA256: digest(source), BaselineSHA256: digest(baseline), UpdatedSHA256: digest(updated),
		PlanSHA256: digestJSON(plan), CandidateSHA256: digest(candidate), Changes: changes, IssueIDs: ids, AllIssueIDs: sortedIDs(c),
	}
	return candidate, e, nil
}

// ValidatePlan rejects unknown, duplicate, or missing issue/field targets
// before a caller applies any plan through Beads.
func ValidatePlan(source []byte, plan Plan) error {
	rows, err := parseExport("source", source)
	if err != nil {
		return err
	}
	_, err = validatePlan(plan, rows)
	return err
}

func sortedIDs(rows map[string]row) []string {
	ids := make([]string, 0, len(rows))
	for id := range rows {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func parseExport(name string, data []byte) (map[string]row, error) {
	result := make(map[string]row)
	for i, line := range splitLines(data) {
		if len(bytes.TrimSpace(line)) == 0 {
			return nil, fmt.Errorf("%s line %d is empty", name, i+1)
		}
		if err := rejectDuplicateKeys(line); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", name, i+1, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil || fields == nil {
			if err == nil {
				err = errors.New("row must be a JSON object")
			}
			return nil, fmt.Errorf("%s line %d: %w", name, i+1, err)
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("%s line %d: missing string issue id", name, i+1)
		}
		if _, exists := result[id]; exists {
			return nil, fmt.Errorf("%s: duplicate issue id %q", name, id)
		}
		result[id] = row{fields: fields, line: append([]byte(nil), line...), id: id}
	}
	if err := validateGraph(result); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return result, nil
}

func validateGraph(rows map[string]row) error {
	for id, item := range rows {
		raw, ok := item.fields["dependencies"]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		var dependencies []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &dependencies); err != nil {
			return fmt.Errorf("issue %s dependencies must be an array", id)
		}
		seen := make(map[string]bool, len(dependencies))
		for _, dependency := range dependencies {
			if dependency == nil {
				return fmt.Errorf("issue %s has a non-object dependency", id)
			}
			edgeIssue, e1 := stringFieldMap(dependency, "issue_id")
			target, e2 := stringFieldMap(dependency, "depends_on_id")
			edgeType, e3 := stringFieldMap(dependency, "type")
			if e1 != nil || e2 != nil || e3 != nil || edgeIssue != id || strings.TrimSpace(edgeType) == "" {
				return fmt.Errorf("issue %s has an invalid dependency edge", id)
			}
			if _, exists := rows[target]; !exists {
				return fmt.Errorf("issue %s dependency target %q is missing", id, target)
			}
			canonical, _ := json.Marshal([3]string{edgeIssue, target, edgeType})
			if seen[string(canonical)] {
				return fmt.Errorf("issue %s has a duplicate dependency edge", id)
			}
			seen[string(canonical)] = true
		}
	}
	return nil
}

func splitLines(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}
	parts := bytes.Split(data, []byte("\n"))
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		return parts[:len(parts)-1]
	}
	return parts
}

func validateIDs(c, b0, b1 map[string]row) error {
	if len(c) != len(b0) || len(c) != len(b1) {
		return fmt.Errorf("issue ID set changed: source=%d baseline=%d updated=%d", len(c), len(b0), len(b1))
	}
	for id := range c {
		if _, ok := b0[id]; !ok {
			return fmt.Errorf("issue ID %q missing from baseline", id)
		}
		if _, ok := b1[id]; !ok {
			return fmt.Errorf("issue ID %q missing from updated export", id)
		}
	}
	return nil
}

func validatePlan(plan Plan, source map[string]row) (map[string]IssueUpdate, error) {
	updates := make(map[string]IssueUpdate, len(plan.Updates))
	for _, u := range plan.Updates {
		if _, exists := updates[u.ID]; exists {
			return nil, fmt.Errorf("duplicate plan issue id %q", u.ID)
		}
		if _, exists := source[u.ID]; !exists {
			return nil, fmt.Errorf("plan issue %q not present in source", u.ID)
		}
		if len(u.Fields) == 0 {
			return nil, fmt.Errorf("plan issue %q has no fields", u.ID)
		}
		for field := range u.Fields {
			if !allowedFields[field] {
				return nil, fmt.Errorf("plan field %q is not allowed", field)
			}
		}
		updates[u.ID] = u
	}
	return updates, nil
}

func compareRows(left, right row, id string, allowed map[string]bool, updated bool, bdVersion string) error {
	l := cloneFields(left.fields)
	r := cloneFields(right.fields)
	if updated {
		normalizeArrays(l)
		normalizeArrays(r)
		if len(allowed) == 0 {
			leftAt, leftHas := l["updated_at"]
			rightAt, rightHas := r["updated_at"]
			if leftHas != rightHas || (leftHas && !jsonEqual(leftAt, rightAt)) {
				return errors.New("unplanned updated_at change")
			}
		}
		delete(l, "updated_at")
		delete(r, "updated_at")
		for field := range allowed {
			delete(l, field)
			delete(r, field)
		}
	} else {
		normalizeImported(l)
		normalizeImported(r)
	}
	if id == "dap-jo9.5.5" && bdVersion == "1.3.1" && !updated {
		if _, sourceHas := left.fields["closed_at"]; !sourceHas {
			status, statusErr := stringField(left, "status")
			if statusErr != nil || status != "closed" {
				return errors.New("pinned closed_at normalization applies only to the closed .5.5 issue")
			}
			want, err := derivedClosedAt(left)
			if err != nil {
				return err
			}
			got, err := stringField(right, "closed_at")
			if err != nil || got != want {
				return errors.New("Beads 1.3.1 closed_at normalization does not match the pinned deterministic formula")
			}
			delete(l, "closed_at")
			delete(r, "closed_at")
		}
	}
	if !jsonEqualMaps(l, r) {
		return fmt.Errorf("unplanned semantic delta in fields %v", differingFields(l, r))
	}
	return nil
}

func differingFields(a, b map[string]json.RawMessage) []string {
	all := make(map[string]bool, len(a)+len(b))
	for key := range a {
		all[key] = true
	}
	for key := range b {
		all[key] = true
	}
	fields := make([]string, 0)
	for key := range all {
		av, aok := a[key]
		bv, bok := b[key]
		if !aok || !bok || !jsonEqual(av, bv) {
			fields = append(fields, key)
		}
	}
	sort.Strings(fields)
	return fields
}

func derivedClosedAt(r row) (string, error) {
	created, err := stringField(r, "created_at")
	if err != nil {
		return "", err
	}
	updated, err := stringField(r, "updated_at")
	if err != nil {
		return "", err
	}
	c, err := parseTimestamp(created)
	if err != nil {
		return "", err
	}
	u, err := parseTimestamp(updated)
	if err != nil {
		return "", err
	}
	if c.After(u) {
		u = c
	}
	return u.Add(time.Second).UTC().Format(time.RFC3339), nil
}

func normalizeImported(fields map[string]json.RawMessage) {
	defaults := map[string]json.RawMessage{"_type": json.RawMessage(`"issue"`), "comment_count": json.RawMessage(`0`)}
	for key, value := range defaults {
		if _, ok := fields[key]; !ok {
			fields[key] = value
		}
	}
	for _, key := range []string{"owner", "notes", "close_reason"} {
		if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte(`""`)) {
			delete(fields, key)
		}
	}
	normalizeArrays(fields)
}

func normalizeArrays(fields map[string]json.RawMessage) {
	for _, key := range []string{"labels", "dependencies"} {
		if raw, ok := fields[key]; ok {
			trimmed := bytes.TrimSpace(raw)
			if len(trimmed) == 0 || trimmed[0] != '[' {
				continue
			}
			var values []json.RawMessage
			if json.Unmarshal(raw, &values) == nil {
				for i := range values {
					if canonical, err := canonicalJSON(values[i]); err == nil {
						values[i] = canonical
					}
				}
				sort.Slice(values, func(i, j int) bool { return string(values[i]) < string(values[j]) })
				encoded, _ := json.Marshal(values)
				fields[key] = encoded
				if len(values) == 0 {
					delete(fields, key)
				}
			}
		}
	}
}

func cloneFields(src map[string]json.RawMessage) map[string]json.RawMessage {
	dst := make(map[string]json.RawMessage, len(src))
	for k, v := range src {
		dst[k] = append(json.RawMessage(nil), v...)
	}
	return dst
}

func jsonEqualMaps(a, b map[string]json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for key, av := range a {
		bv, ok := b[key]
		if !ok || !jsonEqual(av, bv) {
			return false
		}
	}
	return true
}

func jsonEqual(a, b json.RawMessage) bool {
	ca, errA := canonicalJSON(a)
	cb, errB := canonicalJSON(b)
	return errA == nil && errB == nil && bytes.Equal(ca, cb)
}

func canonicalJSON(raw json.RawMessage) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func stringField(r row, name string) (string, error) { return stringFieldMap(r.fields, name) }
func stringFieldMap(fields map[string]json.RawMessage, name string) (string, error) {
	raw, ok := fields[name]
	if !ok {
		return "", fmt.Errorf("missing field %q", name)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("field %q must be a string", name)
	}
	return value, nil
}

func rowID(line []byte) (string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(line, &fields); err != nil {
		return "", err
	}
	return stringFieldMap(fields, "id")
}

func spliceFields(line []byte, fields map[string]string, timestampField, timestamp string) ([]byte, error) {
	values := make(map[string]json.RawMessage, len(fields)+1)
	for key, value := range fields {
		encoded, _ := json.Marshal(value)
		values[key] = encoded
	}
	encoded, _ := json.Marshal(timestamp)
	values[timestampField] = encoded
	return spliceTopLevel(line, values)
}

func spliceTopLevel(line []byte, replacements map[string]json.RawMessage) ([]byte, error) {
	spans, _, err := topLevelSpans(line)
	if err != nil {
		return nil, err
	}
	result := append([]byte(nil), line...)
	type edit struct {
		start, end int
		value      []byte
	}
	edits := make([]edit, 0, len(replacements))
	for key, value := range replacements {
		if span, ok := spans[key]; ok {
			edits = append(edits, edit{span[0], span[1], value})
			delete(replacements, key)
		}
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		result = append(result[:e.start], append(append([]byte(nil), e.value...), result[e.end:]...)...)
	}
	if len(replacements) == 0 {
		return result, nil
	}
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var addition bytes.Buffer
	if len(spans) > 0 {
		addition.WriteByte(',')
	}
	for i, key := range keys {
		encodedKey, _ := json.Marshal(key)
		if i > 0 {
			addition.WriteByte(',')
		}
		addition.Write(encodedKey)
		addition.WriteByte(':')
		addition.Write(replacements[key])
	}
	// Preserve whitespace immediately before the closing brace.
	_, insertAt, err := topLevelSpans(result)
	if err != nil {
		return nil, err
	}
	result = append(result[:insertAt], append(append([]byte(nil), addition.Bytes()...), result[insertAt:]...)...)
	return result, nil
}

func topLevelSpans(line []byte) (map[string][2]int, int, error) {
	spans := make(map[string][2]int)
	i := skipSpace(line, 0)
	if i >= len(line) || line[i] != '{' {
		return nil, 0, errors.New("row must be a JSON object")
	}
	i++
	for {
		i = skipSpace(line, i)
		if i >= len(line) {
			return nil, 0, errors.New("unterminated JSON object")
		}
		if line[i] == '}' {
			return spans, i, nil
		}
		keyStart := i
		keyEnd, err := scanString(line, i)
		if err != nil {
			return nil, 0, err
		}
		var key string
		if err := json.Unmarshal(line[keyStart:keyEnd], &key); err != nil {
			return nil, 0, err
		}
		i = skipSpace(line, keyEnd)
		if i >= len(line) || line[i] != ':' {
			return nil, 0, errors.New("missing colon after JSON key")
		}
		valueStart := skipSpace(line, i+1)
		valueEnd, err := scanValue(line, valueStart)
		if err != nil {
			return nil, 0, err
		}
		spans[key] = [2]int{valueStart, valueEnd}
		i = skipSpace(line, valueEnd)
		if i >= len(line) {
			return nil, 0, errors.New("unterminated JSON object")
		}
		if line[i] == ',' {
			i++
			continue
		}
		if line[i] == '}' {
			return spans, i, nil
		}
		return nil, 0, errors.New("expected comma or object end")
	}
}

func scanString(data []byte, start int) (int, error) {
	if start >= len(data) || data[start] != '"' {
		return 0, errors.New("expected JSON string")
	}
	for i := start + 1; i < len(data); i++ {
		if data[i] == '\\' {
			i++
			continue
		}
		if data[i] == '"' {
			return i + 1, nil
		}
	}
	return 0, errors.New("unterminated JSON string")
}

func scanValue(data []byte, start int) (int, error) {
	if start >= len(data) {
		return 0, errors.New("missing JSON value")
	}
	if data[start] == '"' {
		return scanString(data, start)
	}
	if data[start] != '{' && data[start] != '[' {
		i := start
		for i < len(data) && data[i] != ',' && data[i] != '}' && data[i] != ']' {
			i++
		}
		end := i
		for end > start && (data[end-1] == ' ' || data[end-1] == '\t' || data[end-1] == '\r' || data[end-1] == '\n') {
			end--
		}
		return end, nil
	}
	stack := []byte{data[start]}
	inString, escaped := false, false
	for i := start + 1; i < len(data); i++ {
		b := data[i]
		if inString {
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}
		if b == '"' {
			inString = true
			continue
		}
		if b == '{' || b == '[' {
			stack = append(stack, b)
			continue
		}
		if b == '}' || b == ']' {
			if len(stack) == 0 || (b == '}' && stack[len(stack)-1] != '{') || (b == ']' && stack[len(stack)-1] != '[') {
				return 0, errors.New("mismatched JSON delimiters")
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return i + 1, nil
			}
		}
	}
	return 0, errors.New("unterminated JSON value")
}

func skipSpace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
		i++
	}
	return i
}

func digest(data []byte) string   { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func digestJSON(value any) string { data, _ := json.Marshal(value); return digest(data) }

// rejectDuplicateKeys walks JSON tokens so encoding/json cannot silently accept
// a repeated key before we construct maps from the row.
func rejectDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := consumeJSONValue(dec); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func consumeJSONValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = true
			if err := consumeJSONValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	case '[':
		for dec.More() {
			if err := consumeJSONValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	default:
		return errors.New("unexpected JSON delimiter")
	}
}

func parseTimestamp(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
