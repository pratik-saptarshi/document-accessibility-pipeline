package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

// v2CandidateTextCases pairs each damaged v2 fixture with its repaired
// candidate for independent text validation (dap-jo9.1.6.3). Damaged digests
// are pinned by v2RepairAttemptCases; candidate digests were computed from
// the committed candidate files on 2026-10-16.
var v2CandidateTextCases = []struct {
	name          string
	damagedPath   string
	damagedSHA256 string
	candidatePath string
	candidateSHA  string
}{
	{
		name:          "statement",
		damagedPath:   "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
		damagedSHA256: "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836",
		candidatePath: "../../docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf",
		candidateSHA:  "8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52",
	},
	{
		name:          "invoice",
		damagedPath:   "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
		damagedSHA256: "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad",
		candidatePath: "../../docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf",
		candidateSHA:  "31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448",
	},
}

// v2CandidateTextLimit bounds the pdftotext output per file, matching the
// 256<<10 convention used by the noop text experiments for these families.
const v2CandidateTextLimit = 256 << 10

// TestV2CandidateTextPreserved independently validates that each repaired v2
// candidate preserves the damaged source's extracted text: it pins both file
// digests, then runs pdftotext -layout on both and asserts the outputs are
// byte-identical. A missing pdftotext binary fails the test; it is installed
// (poppler 24.02.0) in the repair-attempt environment.
func TestV2CandidateTextPreserved(t *testing.T) {
	for _, tc := range v2CandidateTextCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			for _, f := range []struct {
				path string
				want string
			}{
				{tc.damagedPath, tc.damagedSHA256},
				{tc.candidatePath, tc.candidateSHA},
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

			damagedText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", tc.damagedPath, "-")
			if err != nil {
				t.Fatalf("extract damaged text: %v", err)
			}
			candidateText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", tc.candidatePath, "-")
			if err != nil {
				t.Fatalf("extract candidate text: %v", err)
			}
			if !bytes.Equal(damagedText, candidateText) {
				t.Fatalf("candidate text differs from damaged source: %d vs %d bytes", len(candidateText), len(damagedText))
			}
			textDigest := sha256.Sum256(candidateText)
			t.Logf("%s: text preserved, %d bytes, text_sha256=%x", tc.name, len(candidateText), textDigest)
		})
	}
}
