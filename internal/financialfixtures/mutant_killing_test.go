package financialfixtures

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mutant-killing tests for financialfixtures validation.
//
// Each test below targets a validation check that a mutation sweep
// (2026-10-23) showed surviving the pre-existing suite: the mutation
// was applied to production code, the package tests still passed, and
// the new test was then verified to FAIL by assertion against the
// mutated code and PASS against the unmutated code. Mutants that the
// pre-existing suite already kills by assertion (width-vector length,
// fixture object count, bounded-read limit, observation exit status,
// preflight input digest) are not repeated here.

const (
	mutantHashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mutantHashB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	mutantHashC = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func mutantValidRecords() []FinancialTableV2ObservationInput {
	return []FinancialTableV2ObservationInput{
		{Name: "statement-positive-v2", PDFSHA256: mutantHashA, TextSHA256: mutantHashB, RenderSHA256: mutantHashC},
		{Name: "statement-damaged-v2", PDFSHA256: mutantHashA, TextSHA256: mutantHashB, RenderSHA256: mutantHashC},
		{Name: "invoice-positive-v2", PDFSHA256: mutantHashA, TextSHA256: mutantHashB, RenderSHA256: mutantHashC},
		{Name: "invoice-damaged-v2", PDFSHA256: mutantHashA, TextSHA256: mutantHashB, RenderSHA256: mutantHashC},
	}
}

// Kills the mutant that drops the exact four-record requirement from
// WriteFinancialTableV2ObservationManifest: the manifest documents a
// fixed four-fixture observation set, in a fixed order.
func TestMutantObservationManifestRecordCountRejected(t *testing.T) {
	var out bytes.Buffer
	if err := WriteFinancialTableV2ObservationManifest(mutantValidRecords()[:3], "pdftotext version test", "10.05.1", &out); err == nil || out.Len() != 0 {
		t.Fatalf("three-record manifest error = %v output = %d bytes, want pre-serialization rejection", err, out.Len())
	}
}

// Kills the mutant that drops the non-empty tool-version requirement:
// an observation manifest without the tools that produced it is not
// provenance.
func TestMutantObservationManifestEmptyToolVersionsRejected(t *testing.T) {
	for _, versions := range [][2]string{{"", "10.05.1"}, {"pdftotext version test", ""}} {
		var out bytes.Buffer
		if err := WriteFinancialTableV2ObservationManifest(mutantValidRecords(), versions[0], versions[1], &out); err == nil || out.Len() != 0 {
			t.Fatalf("empty tool version (%q/%q) error = %v output = %d bytes, want rejection", versions[0], versions[1], err, out.Len())
		}
	}
}

// Kills the mutants that weaken the SHA-256 format checks on
// observation records: digests must be exactly 64 hex characters, not
// merely non-empty (weakened check) and not merely valid hex of any
// length (length check removed from isSHA256).
func TestMutantObservationManifestMalformedDigestsRejected(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "non-hex short digest", value: "not-a-digest"},
		{name: "non-hex 64 characters", value: strings.Repeat("z", 64)},
		{name: "hex digest of 66 characters", value: mutantHashA + "aa"},
		{name: "hex digest of 62 characters", value: mutantHashA[:62]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			records := mutantValidRecords()
			records[1].TextSHA256 = tc.value
			var out bytes.Buffer
			if err := WriteFinancialTableV2ObservationManifest(records, "pdftotext version test", "10.05.1", &out); err == nil || out.Len() != 0 {
				t.Fatalf("malformed digest error = %v output = %d bytes, want rejection", err, out.Len())
			}
		})
	}
}

func mutantBasePDFInput(t *testing.T) []byte {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("..", "..", "docs", "fixtures", "remediation", "financial-tables", "statement-positive.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func mutantWidths() []int { return make([]int, 256) }

// Kills the buildMetricCompleteV2PDF mutants that drop the PDF 2.0
// header check, the 64 KiB input bound, the exact object-5 Helvetica
// dictionary, the duplicate-object rejection, or the missing-object
// rejection. The builder is a narrow transformer for reviewed fixture
// inputs; anything else must be refused, not silently rebuilt.
func TestMutantMetricCompletePDFBuilderRejections(t *testing.T) {
	valid := mutantBasePDFInput(t)
	if _, err := buildMetricCompleteV2PDF(valid, mutantWidths()); err != nil {
		t.Fatalf("positive control rejected: %v", err)
	}
	trailerAt := bytes.Index(valid, []byte("trailer\n"))
	if trailerAt < 0 {
		t.Fatal("base fixture has no trailer section")
	}
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "not a PDF 2.0 header", input: bytes.Replace(valid, []byte("%PDF-2.0\n"), []byte("%PDF-1.7\n"), 1)},
		{name: "input over 64 KiB", input: append(append([]byte(nil), valid[:trailerAt]...),
			append([]byte("%"+strings.Repeat("x", 70<<10)+"\n"), valid[trailerAt:]...)...)},
		{name: "object 5 is not the reviewed Helvetica dictionary", input: bytes.Replace(valid,
			[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"),
			[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>"), 1)},
		{name: "duplicate object number", input: append(append([]byte(nil), valid[:trailerAt]...),
			append([]byte("1 0 obj\n<< /Type /Catalog >>\nendobj\n"), valid[trailerAt:]...)...)},
		{name: "object 22 renumbered away", input: bytes.Replace(valid, []byte("22 0 obj\n"), []byte("23 0 obj\n"), 1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Equal(tc.input, valid) {
				t.Fatal("mutation did not change the base fixture")
			}
			if _, err := buildMetricCompleteV2PDF(tc.input, mutantWidths()); err == nil {
				t.Fatal("buildMetricCompleteV2PDF() accepted an unreviewed fixture input")
			}
		})
	}
}

// Kills the mutant that widens the per-glyph width bounds: widths are
// AFM thousandths of an em and must stay within 0..10000.
func TestMutantMetricCompletePDFWidthBoundsRejected(t *testing.T) {
	valid := mutantBasePDFInput(t)
	for _, width := range []int{-1, 10001} {
		widths := mutantWidths()
		widths[10] = width
		if _, err := buildMetricCompleteV2PDF(valid, widths); err == nil || !strings.Contains(err.Error(), "outside bounds") {
			t.Fatalf("width %d error = %v, want outside-bounds rejection", width, err)
		}
	}
}

// Kills the parseHelveticaAFM mutants that drop the StandardEncoding
// code range (0..255), the non-negative width requirement, or the
// 2 MiB AFM size cap.
func TestMutantParseHelveticaAFMRejections(t *testing.T) {
	write := func(t *testing.T, data []byte) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "Helvetica.afm")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	t.Run("code outside 0..255", func(t *testing.T) {
		path := write(t, []byte("C 256 ; WX 100 ; N beyond ; B 0 0 0 0 ;\n"))
		if _, _, err := parseHelveticaAFM(path); err == nil || !strings.Contains(err.Error(), "outside 0..255") {
			t.Fatalf("parseHelveticaAFM() error = %v, want code-range rejection", err)
		}
	})
	t.Run("negative width", func(t *testing.T) {
		path := write(t, []byte("C 65 ; WX -5 ; N A ; B 0 0 0 0 ;\n"))
		if _, _, err := parseHelveticaAFM(path); err == nil || !strings.Contains(err.Error(), "invalid AFM width") {
			t.Fatalf("parseHelveticaAFM() error = %v, want width rejection", err)
		}
	})
	t.Run("exactly 2 MiB", func(t *testing.T) {
		path := write(t, bytes.Repeat([]byte("x\n"), 1<<20))
		if _, _, err := parseHelveticaAFM(path); err == nil || !strings.Contains(err.Error(), "exceeds 2 MiB") {
			t.Fatalf("parseHelveticaAFM() error = %v, want size-cap rejection", err)
		}
	})
}

// Kills the decodeStrictJSON mutant that drops the trailing-data
// check: a manifest boundary must consume exactly one JSON value.
func TestMutantDecodeStrictJSONTrailingDataRejected(t *testing.T) {
	var target struct {
		Name string `json:"name"`
	}
	if err := decodeStrictJSON([]byte("{\"name\":\"one\"}\n{\"name\":\"two\"}\n"), &target); err == nil {
		t.Fatal("decodeStrictJSON() accepted trailing JSON data")
	}
	if err := decodeStrictJSON([]byte("{\"name\":\"one\"}\n"), &target); err != nil || target.Name != "one" {
		t.Fatalf("decodeStrictJSON() positive control = %v (%q)", err, target.Name)
	}
}

func editMutantJSONFile(t *testing.T, root, relative string, edit func(map[string]any)) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	edit(document)
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
}

// Kills the preflightV2Observations mutants that drop the exact
// four-row observation scope, the fixed family order, the fixed tool
// identity across rows, or the positive/damaged observation equality
// that the damaged-fixture design depends on.
func TestMutantPreflightV2ObservationsRejections(t *testing.T) {
	t.Run("fifth observation row", func(t *testing.T) {
		root := copyV2ObservationPackage(t)
		editMutantJSONFile(t, root, v2ObservationsManifestRelPath, func(document map[string]any) {
			rows := document["observations"].([]any)
			document["observations"] = append(rows, rows[0])
		})
		if _, _, err := preflightV2Observations(root); err == nil || !strings.Contains(err.Error(), "row count") {
			t.Fatalf("preflightV2Observations() error = %v, want row-count rejection", err)
		}
	})
	t.Run("first family renamed", func(t *testing.T) {
		root := copyV2ObservationPackage(t)
		editMutantJSONFile(t, root, v2FixtureManifestRelPath, func(document map[string]any) {
			document["cases"].([]any)[0].(map[string]any)["family"] = "invoices"
		})
		if _, _, err := preflightV2Observations(root); err == nil || !strings.Contains(err.Error(), "family") {
			t.Fatalf("preflightV2Observations() error = %v, want family-order rejection", err)
		}
	})
	t.Run("tool identity differs across rows", func(t *testing.T) {
		root := copyV2ObservationPackage(t)
		editMutantJSONFile(t, root, v2ObservationsManifestRelPath, func(document map[string]any) {
			document["observations"].([]any)[2].(map[string]any)["text_tool"] = "pdftotext version 0.0.0"
		})
		if _, _, err := preflightV2Observations(root); err == nil || !strings.Contains(err.Error(), "tool identity") {
			t.Fatalf("preflightV2Observations() error = %v, want tool-identity rejection", err)
		}
	})
	t.Run("damaged text observation differs from positive", func(t *testing.T) {
		root := copyV2ObservationPackage(t)
		textPath := filepath.Join(root, "docs", "fixtures", "remediation", "financial-tables-v2", "observations", "statement-damaged-v2.txt")
		data, err := os.ReadFile(textPath)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, 'x')
		if err := os.WriteFile(textPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256Hex(data)
		editMutantJSONFile(t, root, v2ObservationsManifestRelPath, func(document map[string]any) {
			document["observations"].([]any)[1].(map[string]any)["text_output_sha256"] = digest
		})
		if _, _, err := preflightV2Observations(root); err == nil || !strings.Contains(err.Error(), "differ") {
			t.Fatalf("preflightV2Observations() error = %v, want positive/damaged difference rejection", err)
		}
	})
}
