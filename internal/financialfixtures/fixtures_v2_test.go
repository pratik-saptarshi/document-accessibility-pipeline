package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const evidenceV2Path = "../../docs/fixtures/remediation/financial-table-experiment.v2.json"

type v2FixtureEvidence struct {
	SchemaVersion       string          `json:"schema_version"`
	Scope               string          `json:"scope"`
	Metrics             v2FontMetrics   `json:"font_metrics"`
	Cases               []v2FixtureCase `json:"cases"`
	Execution           string          `json:"execution_status"`
	ObservationManifest string          `json:"external_observation_manifest"`
}

type v2FontMetrics struct {
	FontName      string         `json:"font_name"`
	Encoding      string         `json:"encoding"`
	FirstChar     int            `json:"first_char"`
	LastChar      int            `json:"last_char"`
	Widths        []int          `json:"widths_1000_em"`
	Descriptor    map[string]any `json:"descriptor"`
	WidthsSHA256  string         `json:"widths_sha256"`
	DescriptorSHA string         `json:"descriptor_sha256"`
	SourceVersion string         `json:"source_version"`
	SourceCommit  string         `json:"source_commit"`
	AFMPath       string         `json:"afm_path"`
	AFMSHA256     string         `json:"afm_sha256"`
	License       string         `json:"license"`
	LicenseNotice string         `json:"license_notice"`
	Mapping       string         `json:"mapping_algorithm"`
}

type v2FixtureCase struct {
	Family       string   `json:"family"`
	PositivePath string   `json:"positive_path"`
	PositiveSHA  string   `json:"positive_sha256"`
	DamagedPath  string   `json:"damaged_path"`
	DamagedSHA   string   `json:"damaged_sha256"`
	TargetHeader string   `json:"target_header"`
	TargetCells  []string `json:"target_cells"`
	WrongHeader  string   `json:"damaged_wrong_header"`
	ReadingOrder []string `json:"reading_order"`
}

type v2ObservationManifest struct {
	SchemaVersion string                `json:"schema_version"`
	Scope         string                `json:"scope"`
	Observations  []externalObservation `json:"observations"`
}

// TestFinancialTableV2MetricCompleteFixtures locks the v2 source bytes,
// complete Helvetica metric resources, and the same semantic cases as v1.
// It is characterization only: this test never invokes a PDF writer or repair.
func TestFinancialTableV2MetricCompleteFixtures(t *testing.T) {
	manifestBytes, err := os.ReadFile(evidenceV2Path)
	if err != nil {
		t.Fatalf("read v2 fixture evidence: %v", err)
	}
	var evidence v2FixtureEvidence
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatalf("decode v2 fixture evidence: %v", err)
	}
	if evidence.SchemaVersion != "2.0.0" || evidence.Scope != "synthetic-fixtures-only; metric-complete PDF 2.0 inputs; no repair or conformance claim" {
		t.Fatalf("v2 schema/scope = %q/%q", evidence.SchemaVersion, evidence.Scope)
	}
	if evidence.Execution != "fixture authoring and independent observations only; no pdfcpu writer, no repair, and no validator run" {
		t.Fatalf("v2 execution boundary = %q", evidence.Execution)
	}
	if evidence.ObservationManifest != "docs/fixtures/remediation/financial-tables-v2/observations/observations.v2.json" {
		t.Fatalf("v2 external observation manifest = %q", evidence.ObservationManifest)
	}
	if err := assertV2MetricManifest(evidence.Metrics); err != nil {
		t.Fatal(err)
	}
	notice, err := os.ReadFile("../../" + evidence.Metrics.LicenseNotice)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notice), "all copyright notices are retained") || !strings.Contains(string(notice), "that this paragraph is not modified") || !strings.Contains(string(notice), "derived Helvetica metric values only, not the AFM file") {
		t.Fatalf("AFM permission notice incomplete: %s", notice)
	}
	if len(evidence.Cases) != 2 {
		t.Fatalf("v2 case count = %d, want equal-priority statement and invoice", len(evidence.Cases))
	}

	baseBytes, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	var base fixtureEvidence
	if err := json.Unmarshal(baseBytes, &base); err != nil {
		t.Fatal(err)
	}
	baseCases := make(map[string]fixtureRecord, len(base.Cases))
	for _, c := range base.Cases {
		baseCases[c.Family] = c
	}
	observationsBytes, err := os.ReadFile("../../" + evidence.ObservationManifest)
	if err != nil {
		t.Fatal(err)
	}
	var observations v2ObservationManifest
	if err := json.Unmarshal(observationsBytes, &observations); err != nil {
		t.Fatal(err)
	}
	if observations.SchemaVersion != "2.0.0" || observations.Scope != "independent text extraction and rendering observations for synthetic metric-complete v2 fixtures" || len(observations.Observations) != 4 {
		t.Fatalf("v2 observations version/scope/count = %q/%q/%d", observations.SchemaVersion, observations.Scope, len(observations.Observations))
	}
	observationByFixture := make(map[string]externalObservation, len(observations.Observations))
	for _, observation := range observations.Observations {
		observationByFixture[observation.Fixture] = observation
	}
	seen := map[string]bool{}
	for _, fixture := range evidence.Cases {
		fixture := fixture
		t.Run(fixture.Family, func(t *testing.T) {
			if fixture.Family != "statements" && fixture.Family != "invoices" || seen[fixture.Family] {
				t.Fatalf("unexpected or duplicate v2 family %q", fixture.Family)
			}
			seen[fixture.Family] = true
			v1, ok := baseCases[fixture.Family]
			if !ok || fixture.TargetHeader != v1.TargetHeader || !reflect.DeepEqual(fixture.TargetCells, v1.TargetCells) || fixture.WrongHeader != v1.DamagedWrongHeader || !reflect.DeepEqual(fixture.ReadingOrder, v1.ReadingOrder) {
				t.Fatalf("v2 semantics differ from accepted equal-priority v1 case: v2=%#v v1=%#v", fixture, v1)
			}
			positivePath := filepath.Join("..", "..", fixture.PositivePath)
			damagedPath := filepath.Join("..", "..", fixture.DamagedPath)
			assertFixtureXRef(t, positivePath)
			assertFixtureXRef(t, damagedPath)
			for _, path := range []string{positivePath, damagedPath} {
				pdfBytes, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.HasPrefix(pdfBytes, []byte("%PDF-2.0\n")) {
					t.Fatalf("v2 fixture %s does not declare PDF 2.0", path)
				}
			}
			positive := readPDFWithDigest(t, positivePath, fixture.PositiveSHA)
			positiveBytes, err := os.ReadFile(positivePath)
			if err != nil {
				t.Fatal(err)
			}
			damaged := readPDFWithDigest(t, damagedPath, fixture.DamagedSHA)
			damagedBytes, err := os.ReadFile(damagedPath)
			if err != nil {
				t.Fatal(err)
			}
			positive, err = api.ReadContext(t.Context(), bytes.NewReader(positiveBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("read v2 positive fixture: %v", err)
			}
			if err := api.ValidateContext(t.Context(), positive); err != nil {
				t.Fatalf("validate v2 positive fixture: %v", err)
			}
			damaged, err = api.ReadContext(t.Context(), bytes.NewReader(damagedBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("read v2 damaged fixture: %v", err)
			}
			if err := api.ValidateContext(t.Context(), damaged); err != nil {
				t.Fatalf("validate v2 damaged fixture: %v", err)
			}
			assertV2HelveticaMetrics(t, positive, evidence.Metrics)
			assertV2HelveticaMetrics(t, damaged, evidence.Metrics)
			assertReadingOrder(t, positive, fixture.ReadingOrder)
			assertReadingOrder(t, damaged, fixture.ReadingOrder)
			if err := assertCellContentBindings(positive); err != nil {
				t.Fatalf("positive content/ParentTree bindings: %v", err)
			}
			if err := assertCellContentBindings(damaged); err != nil {
				t.Fatalf("damaged content/ParentTree bindings: %v", err)
			}
			assertTableAssociations(t, positive, fixture.TargetHeader, fixture.TargetCells, "", fixture.WrongHeader, false)
			assertTableAssociations(t, damaged, fixture.TargetHeader, fixture.TargetCells, v1.DamagedTargetCell, fixture.WrongHeader, true)
			assertOnlyDeclaredAssociationChanged(t, positive, damaged, fixture.TargetHeader, v1.DamagedTargetCell)
			assertOnlyObject20Changed(t, positivePath, damagedPath)
			for _, item := range []struct{ path, digest string }{{fixture.PositivePath, fixture.PositiveSHA}, {fixture.DamagedPath, fixture.DamagedSHA}} {
				obs, ok := observationByFixture[filepath.Base(item.path)]
				if !ok || obs.PDFSHA256 != item.digest || obs.TextExitStatus != 0 || obs.RenderExit != 0 || obs.TextTool == "" || obs.RenderTool == "" {
					t.Fatalf("missing or invalid independent observation for %s: %#v", item.path, obs)
				}
				textPath := filepath.Join("..", "..", obs.TextOutputPath)
				textBytes, err := os.ReadFile(textPath)
				if err != nil {
					t.Fatal(err)
				}
				textHash := sha256.Sum256(textBytes)
				if hex.EncodeToString(textHash[:]) != obs.TextOutputSHA || string(textBytes) != v1.ExtractedText {
					t.Fatalf("independent text output differs for %s: hash=%x expected=%s", item.path, textHash, obs.TextOutputSHA)
				}
				renderBytes, err := os.ReadFile(filepath.Join("..", "..", obs.RenderOutput))
				if err != nil {
					t.Fatal(err)
				}
				renderHash := sha256.Sum256(renderBytes)
				if hex.EncodeToString(renderHash[:]) != obs.RenderOutputSHA {
					t.Fatalf("independent render hash mismatch for %s: got %x want %s", item.path, renderHash, obs.RenderOutputSHA)
				}
			}
			positiveObs := observationByFixture[filepath.Base(fixture.PositivePath)]
			damagedObs := observationByFixture[filepath.Base(fixture.DamagedPath)]
			if positiveObs.TextOutputSHA != damagedObs.TextOutputSHA || positiveObs.RenderOutputSHA != damagedObs.RenderOutputSHA {
				t.Fatalf("v2 positive/damaged external text or render differs: %#v != %#v", positiveObs, damagedObs)
			}
		})
	}
	if !seen["statements"] || !seen["invoices"] {
		t.Fatalf("v2 families = %#v", seen)
	}
}

func assertV2MetricManifest(metrics v2FontMetrics) error {
	if metrics.FontName != "Helvetica" || metrics.Encoding != "StandardEncoding" || metrics.FirstChar != 0 || metrics.LastChar != 255 || len(metrics.Widths) != 256 {
		return fmt.Errorf("incomplete v2 metric vector: font=%q encoding=%q range=%d..%d width_count=%d", metrics.FontName, metrics.Encoding, metrics.FirstChar, metrics.LastChar, len(metrics.Widths))
	}
	canonical := make([]string, len(metrics.Widths))
	for i, width := range metrics.Widths {
		canonical[i] = fmt.Sprint(width)
	}
	widthBytes := []byte(strings.Join(canonical, ","))
	widthHash := sha256.Sum256(widthBytes)
	if hex.EncodeToString(widthHash[:]) != metrics.WidthsSHA256 {
		return fmt.Errorf("v2 width vector digest mismatch: got %x want %s", widthHash, metrics.WidthsSHA256)
	}
	descriptorBytes, err := json.Marshal(metrics.Descriptor)
	if err != nil {
		return fmt.Errorf("marshal v2 descriptor: %w", err)
	}
	descriptorHash := sha256.Sum256(descriptorBytes)
	if hex.EncodeToString(descriptorHash[:]) != metrics.DescriptorSHA {
		return fmt.Errorf("v2 descriptor digest mismatch: got %x want %s", descriptorHash, metrics.DescriptorSHA)
	}
	if metrics.SourceVersion != "github.com/pdfcpu/pdfcpu v0.16.2" || metrics.SourceCommit != "b41a41d5323f6119ff6a183ad3f287eb4a20a617" || metrics.AFMSHA256 != "db772f2830fb6d000907791d8d26a12524d96943a9a739e520ee855c6b25c96f" || metrics.AFMPath != "internal/corefont/Core14_AFMs/Helvetica.afm" || !strings.Contains(metrics.License, "Adobe Core 14 AFM permission") || metrics.LicenseNotice != "docs/fixtures/remediation/financial-tables-v2/AFM-NOTICE.txt" || !strings.Contains(metrics.Mapping, "codes 0..255") {
		return fmt.Errorf("v2 pinned source, AFM provenance, or notice is incomplete: %#v", metrics)
	}
	for key, value := range map[string]any{"Type": "FontDescriptor", "FontName": "Helvetica", "Flags": float64(32), "FontBBox": []any{float64(-166), float64(-225), float64(1000), float64(931)}, "ItalicAngle": float64(0), "Ascent": float64(718), "Descent": float64(-207), "CapHeight": float64(718), "StemV": float64(88)} {
		if got, ok := metrics.Descriptor[key]; !ok || !reflect.DeepEqual(got, value) {
			return fmt.Errorf("v2 Helvetica descriptor field %s = %#v, want %#v", key, got, value)
		}
	}
	return nil
}

func assertV2HelveticaMetrics(t *testing.T, ctx *model.Context, want v2FontMetrics) {
	t.Helper()
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := ctx.DereferenceDict(page["Resources"])
	if err != nil {
		t.Fatal(err)
	}
	fonts, err := ctx.DereferenceDict(resources["Font"])
	if err != nil {
		t.Fatal(err)
	}
	if len(fonts) != 1 {
		t.Fatalf("v2 page has %d font resources, want only the declared Helvetica resource", len(fonts))
	}
	font, err := ctx.DereferenceDict(fonts["F1"])
	if err != nil {
		t.Fatal(err)
	}
	if font["Subtype"] != types.Name("Type1") || font["BaseFont"] != types.Name("Helvetica") || font["Encoding"] != types.Name("StandardEncoding") {
		t.Fatalf("v2 font identity/encoding = %v/%v/%v", font["Subtype"], font["BaseFont"], font["Encoding"])
	}
	if font["FirstChar"] != types.Integer(want.FirstChar) || font["LastChar"] != types.Integer(want.LastChar) {
		t.Fatalf("v2 font code range = %v..%v, want %d..%d", font["FirstChar"], font["LastChar"], want.FirstChar, want.LastChar)
	}
	widths, err := ctx.DereferenceArray(font["Widths"])
	if err != nil {
		t.Fatal(err)
	}
	if len(widths) != len(want.Widths) {
		t.Fatalf("v2 widths length = %d, want %d", len(widths), len(want.Widths))
	}
	for i, width := range widths {
		if width != types.Integer(want.Widths[i]) {
			t.Fatalf("v2 width[%d] = %v, want %d", i, width, want.Widths[i])
		}
	}
	descriptor, err := ctx.DereferenceDict(font["FontDescriptor"])
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range want.Descriptor {
		if got := fmt.Sprint(descriptor[key]); got != fmt.Sprint(value) {
			t.Fatalf("v2 descriptor /%s = %s, want %v", key, got, value)
		}
	}
	for _, key := range []string{"FontFile", "FontFile2", "FontFile3"} {
		if _, present := descriptor[key]; present {
			t.Fatalf("unexpected embedded font program %s", key)
		}
	}
}
