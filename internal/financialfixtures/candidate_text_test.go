package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

// v2CandidateTextLimit bounds the pdftotext output per file, matching the
// 256<<10 convention used by the noop text experiments for these families.
const v2CandidateTextLimit = 256 << 10

// TestV2CandidateTextPreserved independently validates that each repaired v2
// candidate preserves the damaged source's extracted text: it pins both file
// digests from the shared v2RepairAttemptCases table, then runs pdftotext
// -layout on both and asserts the outputs are byte-identical. A missing
// pdftotext binary fails the test; it is installed (poppler 24.02.0) in the
// repair-attempt environment.
func TestV2CandidateTextPreserved(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			for _, f := range []struct {
				path string
				want string
			}{
				{tc.relPath, tc.wantSHA256},
				{tc.candidateRelPath, tc.candidateSHA256},
			} {
				data, err := readV2FileBoundedWithContext(ctx, f.path, v2PDFInputLimit)
				if err != nil {
					t.Fatalf("read %s: %v", f.path, err)
				}
				digest := sha256.Sum256(data)
				if got := hex.EncodeToString(digest[:]); got != f.want {
					t.Fatalf("%s SHA-256 = %s, want %s", f.path, got, f.want)
				}
			}

			damagedText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", tc.relPath, "-")
			if err != nil {
				t.Fatalf("extract damaged text: %v", err)
			}
			candidateText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", tc.candidateRelPath, "-")
			if err != nil {
				t.Fatalf("extract candidate text: %v", err)
			}
			if len(damagedText) == 0 {
				t.Fatalf("%s damaged text extraction returned no bytes; pdftotext output is missing", tc.name)
			}
			if len(candidateText) == 0 {
				t.Fatalf("%s candidate text extraction returned no bytes; pdftotext output is missing", tc.name)
			}
			if !bytes.Equal(damagedText, candidateText) {
				t.Fatalf("candidate text differs from damaged source: %d vs %d bytes", len(candidateText), len(damagedText))
			}
			textDigest := sha256.Sum256(candidateText)
			t.Logf("%s: text preserved, %d bytes, text_sha256=%x", tc.name, len(candidateText), textDigest)
		})
	}
}
