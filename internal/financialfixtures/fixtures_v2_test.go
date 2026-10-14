package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const evidenceV2Path = "../../docs/fixtures/remediation/financial-table-experiment.v2.json"
const v2PDFInputLimit = 64 << 10

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
	manifestBytes, err := readV2FileBounded(evidenceV2Path, 64<<10)
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
			positive, positiveBytes := readV2PDFBounded(t, positivePath, fixture.PositiveSHA)
			damaged, damagedBytes := readV2PDFBounded(t, damagedPath, fixture.DamagedSHA)
			assertFixtureXRefBytes(t, positivePath, positiveBytes)
			assertFixtureXRefBytes(t, damagedPath, damagedBytes)
			for _, item := range []struct {
				path string
				data []byte
			}{{positivePath, positiveBytes}, {damagedPath, damagedBytes}} {
				if !bytes.HasPrefix(item.data, []byte("%PDF-2.0\n")) {
					t.Fatalf("v2 fixture %s does not declare PDF 2.0", item.path)
				}
			}
			assertV2HelveticaMetrics(context.Background(), t, positive, evidence.Metrics)
			assertV2HelveticaMetrics(context.Background(), t, damaged, evidence.Metrics)
			assertReadingOrder(t, positive, fixture.ReadingOrder)
			assertReadingOrder(t, damaged, fixture.ReadingOrder)
			if err := assertCellContentBindings(context.Background(), positive); err != nil {
				t.Fatalf("positive content/ParentTree bindings: %v", err)
			}
			if err := assertCellContentBindings(context.Background(), damaged); err != nil {
				t.Fatalf("damaged content/ParentTree bindings: %v", err)
			}
			assertTableAssociations(t, positive, fixture.TargetHeader, fixture.TargetCells, "", fixture.WrongHeader, false)
			assertTableAssociations(t, damaged, fixture.TargetHeader, fixture.TargetCells, v1.DamagedTargetCell, fixture.WrongHeader, true)
			assertOnlyDeclaredAssociationChanged(t, positive, damaged, fixture.TargetHeader, v1.DamagedTargetCell)
			assertOnlyObject20ChangedBytes(t, positivePath, positiveBytes, damagedPath, damagedBytes)
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

func assertV2HelveticaMetrics(operationContext context.Context, t *testing.T, ctx *model.Context, want v2FontMetrics) {
	t.Helper()
	if err := checkV2HelveticaMetrics(operationContext, ctx, want); err != nil {
		t.Fatal(err)
	}
}

func checkV2HelveticaMetrics(operationContext context.Context, ctx *model.Context, want v2FontMetrics) error {
	if operationContext == nil {
		return fmt.Errorf("v2 metric check requires an operation context")
	}
	if err := operationContext.Err(); err != nil {
		return fmt.Errorf("v2 metric check context: %w", err)
	}
	page, _, _, err := ctx.PageDict(operationContext, 1, false)
	if err != nil {
		return fmt.Errorf("read v2 page: %w", err)
	}
	resources, err := ctx.DereferenceDict(page["Resources"])
	if err != nil {
		return fmt.Errorf("read v2 page resources: %w", err)
	}
	if !v2DictHasKeys(resources, "Font") {
		return fmt.Errorf("v2 page resource keys differ from exact {Font} set")
	}
	fonts, err := ctx.DereferenceDict(resources["Font"])
	if err != nil {
		return fmt.Errorf("read v2 font resources: %w", err)
	}
	if !v2DictHasKeys(fonts, "F1") {
		return fmt.Errorf("v2 font resource names differ from exact {F1} set")
	}
	fontReference, ok := fonts["F1"].(types.IndirectRef)
	if !ok || fontReference.ObjectNumber != 5 || fontReference.GenerationNumber != 0 {
		return fmt.Errorf("v2 page /F1 reference = %v, want object 5 0 R", fonts["F1"])
	}
	font, err := ctx.DereferenceDict(fonts["F1"])
	if err != nil {
		return fmt.Errorf("dereference v2 Helvetica font: %w", err)
	}
	if !v2DictHasKeys(font, "Type", "Subtype", "BaseFont", "Encoding", "FirstChar", "LastChar", "Widths", "FontDescriptor") {
		return fmt.Errorf("v2 Helvetica font dictionary keys differ from exact metric-complete set")
	}
	if font["Type"] != types.Name("Font") {
		return fmt.Errorf("v2 Helvetica font /Type = %v, want PDF name /Font", font["Type"])
	}
	if font["Subtype"] != types.Name("Type1") || font["BaseFont"] != types.Name("Helvetica") || font["Encoding"] != types.Name("StandardEncoding") {
		return fmt.Errorf("v2 font identity/encoding = %v/%v/%v", font["Subtype"], font["BaseFont"], font["Encoding"])
	}
	if font["FirstChar"] != types.Integer(want.FirstChar) || font["LastChar"] != types.Integer(want.LastChar) {
		return fmt.Errorf("v2 font code range = %v..%v, want %d..%d", font["FirstChar"], font["LastChar"], want.FirstChar, want.LastChar)
	}
	widths, err := ctx.DereferenceArray(font["Widths"])
	if err != nil {
		return fmt.Errorf("dereference v2 widths: %w", err)
	}
	if len(widths) != len(want.Widths) {
		return fmt.Errorf("v2 widths length = %d, want %d", len(widths), len(want.Widths))
	}
	for i, width := range widths {
		if width != types.Integer(want.Widths[i]) {
			return fmt.Errorf("v2 width[%d] = %v, want %d", i, width, want.Widths[i])
		}
	}
	descriptorRef, ok := font["FontDescriptor"].(types.IndirectRef)
	if !ok || descriptorRef.ObjectNumber != 23 || descriptorRef.GenerationNumber != 0 {
		return fmt.Errorf("v2 /FontDescriptor reference = %v, want object 23 0 R", font["FontDescriptor"])
	}
	descriptor, err := ctx.DereferenceDict(descriptorRef)
	if err != nil {
		return fmt.Errorf("dereference v2 FontDescriptor: %w", err)
	}
	if !v2DictHasKeys(descriptor, "Type", "FontName", "Flags", "FontBBox", "ItalicAngle", "Ascent", "Descent", "CapHeight", "StemV") {
		return fmt.Errorf("v2 FontDescriptor keys differ from exact complete descriptor set")
	}
	if descriptor["Type"] != types.Name("FontDescriptor") {
		return fmt.Errorf("v2 FontDescriptor /Type = %v, want PDF name /FontDescriptor", descriptor["Type"])
	}
	if descriptor["FontName"] != types.Name("Helvetica") {
		return fmt.Errorf("v2 FontDescriptor /FontName = %v, want PDF name /Helvetica", descriptor["FontName"])
	}
	for key, wantPDFType := range map[string]string{
		"Flags": "integer", "ItalicAngle": "integer", "Ascent": "integer",
		"Descent": "integer", "CapHeight": "integer", "StemV": "integer",
	} {
		wantValue, ok := want.Descriptor[key].(float64)
		if !ok || wantValue != float64(int(wantValue)) {
			return fmt.Errorf("v2 manifest FontDescriptor /%s is not an integer", key)
		}
		got, ok := descriptor[key].(types.Integer)
		if !ok || int(got) != int(wantValue) {
			return fmt.Errorf("v2 FontDescriptor /%s = %v (%T), want PDF %s %d", key, descriptor[key], descriptor[key], wantPDFType, int(wantValue))
		}
	}
	wantBBox, ok := want.Descriptor["FontBBox"].([]any)
	if !ok || len(wantBBox) != 4 {
		return fmt.Errorf("v2 manifest FontDescriptor /FontBBox is not a four-integer array")
	}
	gotBBox, ok := descriptor["FontBBox"].(types.Array)
	if !ok || len(gotBBox) != len(wantBBox) {
		return fmt.Errorf("v2 FontDescriptor /FontBBox = %v (%T), want a four-integer PDF array", descriptor["FontBBox"], descriptor["FontBBox"])
	}
	for i, expected := range wantBBox {
		wantInteger, ok := expected.(float64)
		if !ok || wantInteger != float64(int(wantInteger)) {
			return fmt.Errorf("v2 manifest FontDescriptor /FontBBox[%d] is not an integer", i)
		}
		gotInteger, ok := gotBBox[i].(types.Integer)
		if !ok || int(gotInteger) != int(wantInteger) {
			return fmt.Errorf("v2 FontDescriptor /FontBBox[%d] = %v (%T), want PDF integer %d", i, gotBBox[i], gotBBox[i], int(wantInteger))
		}
	}
	return nil
}

func TestV2PreservedDocumentStateRejectsMetadataAndVersionChanges(t *testing.T) {
	manifestBytes, err := os.ReadFile(evidenceV2Path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence v2FixtureEvidence
	if err := json.Unmarshal(manifestBytes, &evidence); err != nil {
		t.Fatal(err)
	}
	fixture := evidence.Cases[0]
	newContext := func() *model.Context {
		ctx, _ := readV2PDFBounded(t, filepath.Join("..", "..", fixture.PositivePath), fixture.PositiveSHA)
		if err := checkV2PreservedDocumentState(ctx); err != nil {
			t.Fatalf("baseline document state rejected: %v", err)
		}
		return ctx
	}
	for name, mutate := range map[string]func(*model.Context){
		"header version downgrade": func(ctx *model.Context) { v := model.V17; ctx.HeaderVersion = &v },
		"catalog XMP metadata":     func(ctx *model.Context) { ctx.RootDict["Metadata"] = types.IndirectRef{ObjectNumber: 4} },
		"Info dictionary":          func(ctx *model.Context) { ref := types.IndirectRef{ObjectNumber: 4}; ctx.Info = &ref },
		"catalog language":         func(ctx *model.Context) { ctx.RootDict["Lang"] = types.Name("en-US") },
		"marked flag": func(ctx *model.Context) {
			ctx.RootDict["MarkInfo"] = types.Dict{"Marked": types.Boolean(false)}
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := newContext()
			mutate(ctx)
			if err := checkV2PreservedDocumentState(ctx); err == nil {
				t.Fatal("document state mutation unexpectedly passed preservation check")
			}
		})
	}
}

func v2DictHasKeys(dict types.Dict, expected ...string) bool {
	if len(dict) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := dict[key]; !ok {
			return false
		}
	}
	return true
}

func readV2PDFBounded(t *testing.T, path, wantSHA256 string) (*model.Context, []byte) {
	t.Helper()
	return readV2PDFBoundedWithContext(t, t.Context(), path, wantSHA256)
}

func readV2PDFBoundedWithContext(t *testing.T, ctx context.Context, path, wantSHA256 string) (*model.Context, []byte) {
	t.Helper()
	data, err := readV2FileBoundedWithContext(ctx, path, v2PDFInputLimit)
	if err != nil {
		t.Fatalf("read v2 PDF %s: %v", path, err)
	}
	digest := sha256.Sum256(data)
	if got := hex.EncodeToString(digest[:]); got != wantSHA256 {
		t.Fatalf("v2 PDF %s SHA-256 = %s, want %s", path, got, wantSHA256)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-2.0\n")) {
		t.Fatalf("v2 PDF %s does not declare PDF 2.0", path)
	}
	parsed, err := api.ReadContext(ctx, bytes.NewReader(data), boundedStatelessConfig(t))
	if err != nil {
		t.Fatalf("bounded stateless read of v2 PDF %s: %v", path, err)
	}
	if err := api.ValidateContext(ctx, parsed); err != nil {
		t.Fatalf("validate v2 PDF %s: %v", path, err)
	}
	return parsed, data
}

func TestV2PDFInputLimitRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.pdf")
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), v2PDFInputLimit+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readV2FileBounded(path, v2PDFInputLimit); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("exceeds %d-byte limit", v2PDFInputLimit)) {
		t.Fatalf("oversized v2 fixture error = %v, want %d-byte limit rejection", err, v2PDFInputLimit)
	}
}

func TestV2BoundedFileReadChecksContextAndRequiresRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.pdf")
	want := []byte("bounded test input")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := readV2FileBoundedWithContext(cancelled, path, v2PDFInputLimit); !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-cancelled bounded read error = %v, want context.Canceled", err)
	}
	if _, err := readV2FileBoundedWithContext(t.Context(), t.TempDir(), v2PDFInputLimit); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("non-regular bounded input error = %v, want regular-file rejection", err)
	}
	got, err := readV2FileBoundedWithContext(t.Context(), path, v2PDFInputLimit)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("bounded read = %q, want %q", got, want)
	}
}

func TestV2BoundedFileReadRejectsSymlinkAndFifo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.pdf")
	if err := os.WriteFile(target, []byte("bounded test input"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.pdf")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readV2FileBoundedWithContext(t.Context(), link, v2PDFInputLimit); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("symlink bounded input error = %v, want regular-file rejection", err)
	}
	fifo := filepath.Join(dir, "input.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err := readV2FileBoundedWithContext(t.Context(), fifo, v2PDFInputLimit)
	elapsed := time.Since(start)
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("FIFO bounded input error = %v, want regular-file rejection", err)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("FIFO bounded read took %s; want prompt rejection, no open block", elapsed)
	}
}

func TestFinancialTableV2MutationControlsRejectUnexpectedChanges(t *testing.T) {
	manifestBytes, err := os.ReadFile(evidenceV2Path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence v2FixtureEvidence
	if err := json.Unmarshal(manifestBytes, &evidence); err != nil {
		t.Fatal(err)
	}
	fixture := evidence.Cases[0]
	fixturePath := filepath.Join("..", "..", fixture.PositivePath)
	newContext := func() *model.Context {
		ctx, _ := readV2PDFBounded(t, fixturePath, fixture.PositiveSHA)
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err != nil {
			t.Fatalf("positive v2 mutation baseline rejected before mutation: %v", err)
		}
		if !v2AllTableAssociationsMatch(t, ctx) {
			t.Fatal("positive v2 mutation baseline has an unexpected table association")
		}
		if err := assertCellContentBindings(context.Background(), ctx); err != nil {
			t.Fatalf("positive v2 mutation baseline content binding: %v", err)
		}
		return ctx
	}

	t.Run("width value", func(t *testing.T) {
		ctx := newContext()
		_, font, _ := v2FontDictionaries(t, ctx)
		widths, err := ctx.DereferenceArray(font["Widths"])
		if err != nil {
			t.Fatal(err)
		}
		mutated := append(types.Array(nil), widths...)
		mutated[65] = types.Integer(evidence.Metrics.Widths[65] + 1)
		font["Widths"] = mutated
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "width[65]") {
			t.Fatalf("width mutation rejection = %v, want width[65] mismatch", err)
		}
	})
	t.Run("font Type value", func(t *testing.T) {
		ctx := newContext()
		_, font, _ := v2FontDictionaries(t, ctx)
		font["Type"] = types.Name("NotFont")
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "font /Type") {
			t.Fatalf("wrong font Type rejection = %v, want /Type /Font failure", err)
		}
	})
	t.Run("descriptor Type uses PDF string", func(t *testing.T) {
		ctx := newContext()
		_, _, descriptor := v2FontDictionaries(t, ctx)
		descriptor["Type"] = types.StringLiteral("FontDescriptor")
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "FontDescriptor /Type") {
			t.Fatalf("wrong PDF type rejection = %v, want /Type name-object failure", err)
		}
	})
	t.Run("descriptor numeric text uses PDF name", func(t *testing.T) {
		ctx := newContext()
		_, _, descriptor := v2FontDictionaries(t, ctx)
		descriptor["Ascent"] = types.Name("718")
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "FontDescriptor /Ascent") {
			t.Fatalf("same-text wrong PDF type rejection = %v, want integer-object failure", err)
		}
	})
	t.Run("descriptor Flags numeric text uses PDF name", func(t *testing.T) {
		ctx := newContext()
		_, _, descriptor := v2FontDictionaries(t, ctx)
		descriptor["Flags"] = types.Name("32")
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "FontDescriptor /Flags") {
			t.Fatalf("same-text Flags type rejection = %v, want integer-object failure", err)
		}
	})
	t.Run("extra font dictionary key", func(t *testing.T) {
		ctx := newContext()
		_, font, _ := v2FontDictionaries(t, ctx)
		font["Unexpected"] = types.Integer(1)
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "font dictionary keys") {
			t.Fatalf("extra font key rejection = %v, want exact font dictionary key-set failure", err)
		}
	})
	t.Run("extra descriptor key", func(t *testing.T) {
		ctx := newContext()
		_, _, descriptor := v2FontDictionaries(t, ctx)
		descriptor["Unexpected"] = types.Integer(1)
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "FontDescriptor keys") {
			t.Fatalf("extra descriptor key rejection = %v, want exact descriptor key-set failure", err)
		}
	})
	t.Run("extra page font resource", func(t *testing.T) {
		ctx := newContext()
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
		fonts["F2"] = fonts["F1"]
		if err := checkV2HelveticaMetrics(context.Background(), ctx, evidence.Metrics); err == nil || !strings.Contains(err.Error(), "font resource names") {
			t.Fatalf("extra font resource rejection = %v, want exact resource-name failure", err)
		}
	})
	t.Run("association change", func(t *testing.T) {
		ctx := newContext()
		mutateV2TargetAssociation(t, ctx, fixture.TargetHeader, fixture.WrongHeader)
		if v2AllTableAssociationsMatch(t, ctx) {
			t.Fatal("row-2 target association mutation was not rejected")
		}
	})
	t.Run("ParentTree content binding", func(t *testing.T) {
		ctx := newContext()
		root := mustStructTreeRoot(t, ctx)
		parentTree, err := ctx.DereferenceDict(root["ParentTree"])
		if err != nil {
			t.Fatal(err)
		}
		nums, err := resolvedArray(ctx, parentTree["Nums"])
		if err != nil || len(nums) != 2 {
			t.Fatalf("v2 ParentTree nums = %v err=%v", nums, err)
		}
		parents, err := resolvedArray(ctx, nums[1])
		if err != nil || len(parents) != 9 {
			t.Fatalf("v2 ParentTree array = %v err=%v", parents, err)
		}
		parents[8] = parents[7]
		nums[1] = parents
		parentTree["Nums"] = nums
		if err := assertCellContentBindings(context.Background(), ctx); err == nil || !strings.Contains(err.Error(), "ParentTree MCID 8 points") {
			t.Fatalf("ParentTree binding mutation rejection = %v, want MCID 8 reverse-binding failure", err)
		}
	})
}

func v2FontDictionaries(t *testing.T, ctx *model.Context) (types.Dict, types.Dict, types.Dict) {
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
	font, err := ctx.DereferenceDict(fonts["F1"])
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := ctx.DereferenceDict(font["FontDescriptor"])
	if err != nil {
		t.Fatal(err)
	}
	return resources, font, descriptor
}

func v2AllTableAssociationsMatch(t *testing.T, ctx *model.Context) bool {
	t.Helper()
	headers, rows := tableObservations(t, ctx)
	ids := make(map[int]string, len(headers))
	for _, header := range headers {
		ids[header.index] = header.id
	}
	for _, row := range rows {
		for column, cell := range row {
			if !reflect.DeepEqual(cell.headerIDs, []string{ids[column]}) {
				return false
			}
		}
	}
	return true
}

func mutateV2TargetAssociation(t *testing.T, ctx *model.Context, targetHeader, wrongHeader string) {
	t.Helper()
	headers, _ := tableObservations(t, ctx)
	targetColumn, wrongID := -1, ""
	for _, header := range headers {
		if header.title == targetHeader {
			targetColumn = header.index
		}
		if header.title == wrongHeader {
			wrongID = header.id
		}
	}
	if targetColumn < 0 || wrongID == "" {
		t.Fatalf("target or wrong header missing for mutation: %q/%q", targetHeader, wrongHeader)
	}
	root := mustStructTreeRoot(t, ctx)
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		t.Fatal(err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		t.Fatal(err)
	}
	rows := table["K"].(types.Array)
	dataRow, err := ctx.DereferenceDict(rows[2])
	if err != nil {
		t.Fatal(err)
	}
	cells := dataRow["K"].(types.Array)
	cell, err := ctx.DereferenceDict(cells[targetColumn])
	if err != nil {
		t.Fatal(err)
	}
	attributes, ok := cell["A"].(types.Dict)
	if !ok {
		t.Fatalf("target cell attributes are %T", cell["A"])
	}
	attributes["Headers"] = types.Array{types.StringLiteral(wrongID)}
}
