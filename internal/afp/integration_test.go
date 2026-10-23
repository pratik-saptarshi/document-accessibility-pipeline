package afp

// DAP-side integration test for the adopted go-afp-line dependency.
// Required by Decision 0021 §k, recorded as Decision 0022 condition 6.
// It exercises the reviewed revision through the intended DAP call path
// (afp.Fields → afptoline.Reduce → linetopdf.Render) so the §k(a) evidence
// no longer rests on a /tmp scratch driver.
//
// Fixtures (internal/afp/testdata):
//   - statement-001.afp: hand-authored synthetic AFP (224 bytes), copied
//     from go-afp-line's fixture corpus (Apache-2.0, same owner).
//   - 01_Health_Coverage.afp: the Decision 0020 accepted AFP reference
//     (SHA-256 553a336d…b73a, synthetic-looking PII); used only to assert
//     the owner-accepted 0x80 rejection stays loud.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	goafp "github.com/pratik-saptarshi/go-afp-line/afp"
	"github.com/pratik-saptarshi/go-afp-line/afptoline"
	"github.com/pratik-saptarshi/go-afp-line/linetopdf"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if len(b) == 0 {
		t.Fatalf("fixture %s is empty", name)
	}
	return b
}

// The adopted revision must decode every structured field of the synthetic
// fixture, reduce it with zero exclusions, and render a valid PDF.
func TestIntegrationControlFixtureEndToEnd(t *testing.T) {
	raw := readFixture(t, "statement-001.afp")

	nFields, nErrs := 0, 0
	for _, err := range goafp.Fields(bytes.NewReader(raw)) {
		nFields++
		if err != nil {
			nErrs++
		}
	}
	if nFields == 0 {
		t.Fatal("no structured fields decoded")
	}
	if nErrs != 0 {
		t.Fatalf("%d/%d field decode errors (registry regression?)", nErrs, nFields)
	}

	doc, excl, err := afptoline.Reduce(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Reduce: %v", err)
	}
	if got := len(doc.Pages); got != 2 {
		t.Fatalf("pages = %d, want 2", got)
	}
	if len(excl) != 0 {
		t.Fatalf("exclusions = %d, want 0: %+v", len(excl), excl)
	}

	var out bytes.Buffer
	if err := linetopdf.Render(doc, &out); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(out.Bytes(), []byte("%PDF-1.7")) {
		t.Fatalf("missing %%PDF-1.7 magic, got %q", out.Bytes()[:8])
	}
	if out.Len() < 512 {
		t.Fatalf("suspiciously small PDF: %d bytes", out.Len())
	}
	// Magic and size alone cannot catch a truncated-but-long malformed
	// buffer: validate the rendered bytes as a real PDF document.
	if err := api.Validate(context.Background(), bytes.NewReader(out.Bytes()), model.NewDefaultConfiguration(), nil); err != nil {
		t.Fatalf("pdfcpu validation of rendered PDF: %v", err)
	}
}

// The 0020 profile file must abort Reduce with a loud typed diagnostic
// naming the afplib-only PTOCA code 0x80 — the owner-accepted expected
// behavior (owner ruling 2026-10-22). It must never become a silent skip.
func TestIntegrationProfileRejectsAfplibPTOCA(t *testing.T) {
	raw := readFixture(t, "01_Health_Coverage.afp")

	_, _, err := afptoline.Reduce(bytes.NewReader(raw))
	if err == nil {
		t.Fatal("expected 0x80 abort, got nil error")
	}
	var de goafp.DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("error is %T, want afp.DecodeError: %v", err, err)
	}
	if de.Kind != goafp.KindValidation {
		t.Errorf("kind = %q, want %q", de.Kind, goafp.KindValidation)
	}
	if !strings.Contains(de.Reason, "0x80") {
		t.Errorf("reason %q does not name code 0x80", de.Reason)
	}
	if de.FieldID == "" {
		t.Error("FieldID empty: diagnostic must name the failing field")
	}
}
