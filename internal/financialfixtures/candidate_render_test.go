package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// v2RepairedRenderCandidates pins the repaired v2 candidates produced by the
// bounded repair attempt (dap-jo9.1.6.3). Digests were computed from the
// committed candidate files on 2026-10-16.
var v2RepairedRenderCandidates = map[string]struct {
	relPath    string
	wantSHA256 string
}{
	"statement": {
		relPath:    "../../docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf",
		wantSHA256: "8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52",
	},
	"invoice": {
		relPath:    "../../docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf",
		wantSHA256: "31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448",
	},
}

// readV2PinnedFileBounded reads a fixture path and asserts its SHA-256
// against a pinned digest.
func readV2PinnedFileBounded(t *testing.T, path, wantSHA256 string) []byte {
	t.Helper()
	data, err := readV2FileBounded(path, v2PDFInputLimit)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	digest := sha256.Sum256(data)
	if got := hex.EncodeToString(digest[:]); got != wantSHA256 {
		t.Fatalf("SHA-256 of %s = %s, want pinned %s", path, got, wantSHA256)
	}
	return data
}

// renderV2PDFToPNG renders page 1 of a PDF to a PNG file via pdftoppm at 72
// dpi using the bounded-subprocess pattern shared with the v2 experiment.
func renderV2PDFToPNG(t *testing.T, ctx context.Context, pdfPath, pngPrefix string) []byte {
	t.Helper()
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Fatalf("pdftoppm is required for render validation: %v", err)
	}
	if _, err := runV2NoopCommand(ctx, 4<<20, "pdftoppm", "-png", "-r", "72", "-singlefile", pdfPath, pngPrefix); err != nil {
		t.Fatalf("render %s: %v", pdfPath, err)
	}
	png, err := os.ReadFile(pngPrefix + ".png")
	if err != nil {
		t.Fatalf("read rendered PNG for %s: %v", pdfPath, err)
	}
	return png
}

// TestV2CandidateRenderIdentical renders each damaged v2 source and its
// repaired candidate at 72 dpi and asserts the two PNGs are byte-identical,
// i.e. the repair changed only non-rendered structure (/Headers) and not a
// single pixel. Damaged-source digests are grounded in v2RepairAttemptCases.
func TestV2CandidateRenderIdentical(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			candidate, ok := v2RepairedRenderCandidates[tc.name]
			if !ok {
				t.Fatalf("no pinned repaired candidate for %q", tc.name)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			damagedData := readV2PinnedFileBounded(t, tc.relPath, tc.wantSHA256)
			candidateData := readV2PinnedFileBounded(t, candidate.relPath, candidate.wantSHA256)
			t.Logf("%s damaged SHA-256: %x (%d bytes)", tc.name, sha256.Sum256(damagedData), len(damagedData))
			t.Logf("%s candidate SHA-256: %s (%d bytes)", tc.name, candidate.wantSHA256, len(candidateData))

			staging := t.TempDir()
			damagedPNG := renderV2PDFToPNG(t, ctx, tc.relPath, filepath.Join(staging, tc.name+"-damaged"))
			candidatePNG := renderV2PDFToPNG(t, ctx, candidate.relPath, filepath.Join(staging, tc.name+"-repaired"))
			damagedRender := sha256.Sum256(damagedPNG)
			candidateRender := sha256.Sum256(candidatePNG)
			t.Logf("%s damaged render SHA-256: %x (%d bytes)", tc.name, damagedRender, len(damagedPNG))
			t.Logf("%s candidate render SHA-256: %x (%d bytes)", tc.name, candidateRender, len(candidatePNG))

			if !bytes.Equal(damagedPNG, candidatePNG) {
				t.Fatalf("%s render differs: damaged render SHA-256 = %x, candidate render SHA-256 = %x", tc.name, damagedRender, candidateRender)
			}
		})
	}
}
