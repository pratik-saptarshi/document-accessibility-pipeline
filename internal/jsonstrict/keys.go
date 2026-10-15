// Package jsonstrict contains small checks shared by JSON configuration
// readers. It does not define any application's schema or unknown-field policy.
package jsonstrict

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RejectDuplicateKeys rejects JSON object keys that encoding/json would fold
// to the same name when matching struct fields.
// It inspects only the first JSON value; callers requiring exactly one value
// must validate EOF themselves.
func RejectDuplicateKeys(reader io.Reader) error {
	return checkValueKeys(json.NewDecoder(reader))
}

func checkValueKeys(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]string)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			folded := foldJSONName(key)
			if previous, exists := seen[folded]; exists {
				return fmt.Errorf("duplicate or case-variant key %q (also present as %q)", key, previous)
			}
			seen[folded] = key
			if err := checkValueKeys(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := checkValueKeys(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

// foldJSONName mirrors encoding/json.foldName: ASCII letters are folded to
// uppercase, while each non-ASCII rune is mapped to its Unicode simple-fold
// representative.
func foldJSONName(value string) string {
	var folded strings.Builder
	folded.Grow(len(value))
	for _, r := range value {
		if 'a' <= r && r <= 'z' {
			r -= 'a' - 'A'
		} else if r >= utf8.RuneSelf {
			for {
				next := unicode.SimpleFold(r)
				if next <= r {
					r = next
					break
				}
				r = next
			}
		}
		folded.WriteRune(r)
	}
	return folded.String()
}
