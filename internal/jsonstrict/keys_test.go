package jsonstrict

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRejectDuplicateKeys(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "nested objects",
			input: `{"outer":{"left":1,"right":2}}`,
		},
		{
			name:    "nested exact duplicate",
			input:   `{"outer":{"key":1,"key":2}}`,
			wantErr: true,
		},
		{
			name:    "ASCII case-fold collision",
			input:   `{"key":1,"KEY":2}`,
			wantErr: true,
		},
		{
			name:    "Unicode simple-fold collision",
			input:   `{"ſ":1,"s":2}`,
			wantErr: true,
		},
		{
			name:  "same keys in separate array objects",
			input: `[{"key":1},{"key":2}]`,
		},
		{
			name:  "standalone case variant is schema policy",
			input: `{"PROVENANCE":"synthetic-only"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := RejectDuplicateKeys(strings.NewReader(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("RejectDuplicateKeys(%s) error = %v, wantErr %t", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestRejectDuplicateKeysMatchesEncodingJSONStructFolding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		parse func([]byte) (int, error)
	}{
		{
			name:  "ASCII field match",
			input: `{"key":1,"KEY":2}`,
			parse: func(data []byte) (int, error) {
				var target struct {
					Key int `json:"key"`
				}
				err := json.Unmarshal(data, &target)
				return target.Key, err
			},
		},
		{
			name:  "Unicode simple-fold field match",
			input: `{"ſ":1,"s":2}`,
			parse: func(data []byte) (int, error) {
				var target struct {
					S int `json:"s"`
				}
				err := json.Unmarshal(data, &target)
				return target.S, err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse([]byte(tc.input))
			if err != nil || got != 2 {
				t.Fatalf("encoding/json parsed %s as (%d, %v), want (2, nil)", tc.input, got, err)
			}
			if err := RejectDuplicateKeys(strings.NewReader(tc.input)); err == nil {
				t.Fatalf("RejectDuplicateKeys(%s) accepted keys encoding/json folds together", tc.input)
			}
		})
	}
}

func TestRejectDuplicateKeysTraversesArraysAndScalars(t *testing.T) {
	for _, input := range []string{`null`, `true`, `17`, `"text"`, `[]`, `[1,null,{"key":1}]`} {
		t.Run(input, func(t *testing.T) {
			if err := RejectDuplicateKeys(strings.NewReader(input)); err != nil {
				t.Fatalf("RejectDuplicateKeys(%s) error = %v, want nil", input, err)
			}
		})
	}
	if err := RejectDuplicateKeys(strings.NewReader(`[{"nested":{"x":1,"x":2}}]`)); err == nil {
		t.Fatal("RejectDuplicateKeys accepted duplicate keys nested inside an array")
	}
}

func TestRejectDuplicateKeysRejectsMalformedJSON(t *testing.T) {
	for _, input := range []string{`{"key":1`, `[1,`, `"\x"`} {
		t.Run(input, func(t *testing.T) {
			if err := RejectDuplicateKeys(strings.NewReader(input)); err == nil {
				t.Fatalf("RejectDuplicateKeys(%s) error = nil, want malformed JSON error", input)
			}
		})
	}
}

func TestRejectDuplicateKeysChecksOnlyFirstJSONValue(t *testing.T) {
	input := `{"first":1} {"later":1,"later":2}`
	if err := RejectDuplicateKeys(strings.NewReader(input)); err != nil {
		t.Fatalf("RejectDuplicateKeys(%s) error = %v, want nil for first value; caller owns EOF validation", input, err)
	}
}
