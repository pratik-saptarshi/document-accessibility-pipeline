package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const authorizedD2ReopenNote = "\n\n2026-10-18 REOPENED by owner decision D2: the 09:00Z close preceded AC1/overall acceptance resolution (AC2f accepted 10:52Z, after close). The premature close does not stand; issue returns to open pending proper acceptance."

func TestValidateDirectEditTransitionAcceptsOnlyAuthorizedD2Reopen(t *testing.T) {
	source := d2SourceRow(t)
	candidate := withD2Reopen(t, source)
	if err := ValidateDirectEditTransition(source, candidate); err != nil {
		t.Fatalf("ValidateDirectEditTransition() rejected the bounded D2 reopen: %v", err)
	}
}

func TestValidateDirectEditTransitionRejectsUnapprovedSourceFieldChanges(t *testing.T) {
	d2 := d2SourceRow(t)
	source := append(append([]byte(nil), d2...),
		[]byte("{\"id\":\"dap-jo9.5.3\",\"title\":\"preserve null\",\"dependencies\":null,\"dependency_count\":0}\n")...)
	valid := withD2Reopen(t, source)
	tests := []struct {
		name string
		edit func([]byte) []byte
	}{
		{"remove explicit null", func(b []byte) []byte { return removeDirectEditField(t, b, "dap-jo9.5.3", "dependencies") }},
		{"restore derived closed timestamp", func(b []byte) []byte {
			return setDirectEditField(t, b, "dap-jo9.5.3", "closed_at", json.RawMessage(`"2026-10-14T04:25:39Z"`))
		}},
		{"mutate unrelated title", func(b []byte) []byte {
			return setDirectEditField(t, b, "dap-jo9.5.3", "title", json.RawMessage(`"changed"`))
		}},
		{"close unrelated issue", func(b []byte) []byte {
			return setDirectEditField(t, b, "dap-jo9.5.3", "status", json.RawMessage(`"closed"`))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateDirectEditTransition(source, tc.edit(valid)); err == nil {
				t.Fatal("ValidateDirectEditTransition() accepted an unapproved source-to-candidate change")
			}
		})
	}
}

func TestValidateDirectEditTransitionRejectsMutatedD2LifecycleFields(t *testing.T) {
	source := d2SourceRow(t)
	valid := withD2Reopen(t, source)
	sourceFields := directEditFieldsForTest(t, source, "dap-jo9.1.6.7")
	tests := []struct {
		name  string
		field string
		value json.RawMessage
	}{
		{"wrong status", "status", json.RawMessage(`"closed"`)},
		{"non-null closed_at", "closed_at", json.RawMessage(`"2026-10-18T15:55:00Z"`)},
		{"restored source closed_at", "closed_at", sourceFields["closed_at"]},
		{"fabricated close reason", "close_reason", json.RawMessage(`"new reason"`)},
		{"restored source close reason", "close_reason", sourceFields["close_reason"]},
		{"wrong D2 note", "notes", json.RawMessage(`"D2"`)},
		{"timestamp at or before source", "updated_at", json.RawMessage(`"2026-10-18T09:00:32Z"`)},
		{"non-UTC timestamp", "updated_at", json.RawMessage(`"2026-10-18T15:55:00+01:00"`)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := setDirectEditField(t, valid, "dap-jo9.1.6.7", tc.field, tc.value)
			if err := ValidateDirectEditTransition(source, candidate); err == nil {
				t.Fatal("ValidateDirectEditTransition() accepted a mutated D2 transition")
			}
		})
	}
}

func d2SourceRow(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "approved-d2-reopen-source-row.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func withD2Reopen(t *testing.T, source []byte) []byte {
	t.Helper()
	fields := directEditFieldsForTest(t, source, "dap-jo9.1.6.7")
	fields["status"] = json.RawMessage(`"open"`)
	fields["closed_at"] = json.RawMessage(`null`)
	fields["close_reason"] = json.RawMessage(`""`)
	fields["updated_at"] = json.RawMessage(`"2026-10-18T15:55:00Z"`)
	fields["notes"], _ = json.Marshal(authorizedD2ReopenNote)
	return replaceDirectEditRow(t, source, "dap-jo9.1.6.7", fields)
}

func directEditFieldsForTest(t *testing.T, data []byte, id string) map[string]json.RawMessage {
	t.Helper()
	for _, line := range bytes.Split(data, []byte("\n")) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			t.Fatal(err)
		}
		var got string
		_ = json.Unmarshal(fields["id"], &got)
		if got == id {
			return fields
		}
	}
	t.Fatalf("issue %q not found", id)
	return nil
}

func replaceDirectEditRow(t *testing.T, data []byte, id string, replacement map[string]json.RawMessage) []byte {
	t.Helper()
	var out bytes.Buffer
	found := false
	for _, line := range bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n")) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			t.Fatal(err)
		}
		var got string
		_ = json.Unmarshal(fields["id"], &got)
		if got == id {
			line, _ = json.Marshal(replacement)
			found = true
		}
		out.Write(line)
		out.WriteByte('\n')
	}
	if !found {
		t.Fatalf("issue %q not found", id)
	}
	return out.Bytes()
}

func setDirectEditField(t *testing.T, data []byte, id, field string, value json.RawMessage) []byte {
	t.Helper()
	fields := directEditFieldsForTest(t, data, id)
	fields[field] = append(json.RawMessage(nil), value...)
	return replaceDirectEditRow(t, data, id, fields)
}

func removeDirectEditField(t *testing.T, data []byte, id, field string) []byte {
	t.Helper()
	fields := directEditFieldsForTest(t, data, id)
	delete(fields, field)
	return replaceDirectEditRow(t, data, id, fields)
}
