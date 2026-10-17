package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// v2RenderPNGOutputLimit bounds the rendered PNG read, matching the
// subprocess output cap used for pdftoppm.
const v2RenderPNGOutputLimit = 4 << 20

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
	png, err := readV2FileBounded(pngPrefix+".png", v2RenderPNGOutputLimit)
	if err != nil {
		t.Fatalf("read rendered PNG for %s: %v", pdfPath, err)
	}
	return png
}

// TestV2CandidateRenderIdentical renders each damaged v2 source and its
// repaired candidate at 72 dpi and asserts the two PNGs are byte-identical,
// i.e. the repair changed only non-rendered structure (/Headers) and not a
// single pixel. Candidate digests come from the shared v2RepairAttemptCases
// table.
func TestV2CandidateRenderIdentical(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			damagedData := readV2PinnedFileBounded(t, tc.relPath, tc.wantSHA256)
			candidateData := readV2PinnedFileBounded(t, tc.candidateRelPath, tc.candidateSHA256)
			t.Logf("%s damaged SHA-256: %x (%d bytes)", tc.name, sha256.Sum256(damagedData), len(damagedData))
			t.Logf("%s candidate SHA-256: %s (%d bytes)", tc.name, tc.candidateSHA256, len(candidateData))

			staging := t.TempDir()
			damagedPNG := renderV2PDFToPNG(t, ctx, tc.relPath, filepath.Join(staging, tc.name+"-damaged"))
			candidatePNG := renderV2PDFToPNG(t, ctx, tc.candidateRelPath, filepath.Join(staging, tc.name+"-repaired"))
			damagedRender := sha256.Sum256(damagedPNG)
			candidateRender := sha256.Sum256(candidatePNG)
			t.Logf("%s damaged render SHA-256: %x (%d bytes)", tc.name, damagedRender, len(damagedPNG))
			t.Logf("%s candidate render SHA-256: %x (%d bytes)", tc.name, candidateRender, len(candidatePNG))

			if len(damagedPNG) == 0 {
				t.Fatalf("%s damaged render produced no PNG bytes; pdftoppm output is missing", tc.name)
			}
			if len(candidatePNG) == 0 {
				t.Fatalf("%s candidate render produced no PNG bytes; pdftoppm output is missing", tc.name)
			}

			if !bytes.Equal(damagedPNG, candidatePNG) {
				t.Fatalf("%s render differs: damaged render SHA-256 = %x, candidate render SHA-256 = %x", tc.name, damagedRender, candidateRender)
			}
		})
	}
}
