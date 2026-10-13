package financialfixtures

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const evidencePath = "../../docs/fixtures/remediation/financial-table-experiment.v1.json"

type fixtureEvidence struct {
	SchemaVersion       string          `json:"schema_version"`
	Scope               string          `json:"scope"`
	ObservationManifest string          `json:"external_observation_manifest"`
	ControlsStatus      string          `json:"controls_status"`
	ValidatorLimit      string          `json:"validator_limit"`
	Cases               []fixtureRecord `json:"cases"`
	Controls            []string        `json:"controls"`
}

type fixtureRecord struct {
	Family             string   `json:"family"`
	Provenance         string   `json:"provenance"`
	Topology           string   `json:"topology"`
	PDFProfile         string   `json:"pdf_profile"`
	PositivePath       string   `json:"positive_path"`
	PositiveSHA256     string   `json:"positive_sha256"`
	DamagedPath        string   `json:"damaged_path"`
	DamagedSHA256      string   `json:"damaged_sha256"`
	DerivedFromSHA256  string   `json:"derived_from_positive_sha256"`
	TargetHeader       string   `json:"target_header"`
	TargetRelation     string   `json:"target_relation"`
	TargetCells        []string `json:"target_cells"`
	DamagedTargetCell  string   `json:"damaged_target_cell"`
	DamagedWrongHeader string   `json:"damaged_wrong_header"`
	DisplayedStrings   []string `json:"displayed_strings"`
	ReadingOrder       []string `json:"reading_order"`
	SupportBoundaries  []string `json:"support_boundaries"`
	Exclusions         []string `json:"exclusions"`
	PositiveRelation   string   `json:"positive_relation"`
	DamagedChange      string   `json:"damaged_change"`
	ExtractedText      string   `json:"extracted_text"`
	TextSHA256         string   `json:"text_sha256"`
	PositiveTextSHA    string   `json:"positive_text_sha256"`
	DamagedTextSHA     string   `json:"damaged_text_sha256"`
	TextTool           string   `json:"text_tool"`
	RenderTool         string   `json:"render_tool"`
	PositiveRenderSHA  string   `json:"positive_render_sha256"`
	DamagedRenderSHA   string   `json:"damaged_render_sha256"`
	StructureSource    string   `json:"structure_source"`
	StructureResult    string   `json:"structure_result"`
}

type observationManifest struct {
	SchemaVersion string                `json:"schema_version"`
	Scope         string                `json:"scope"`
	Observations  []externalObservation `json:"observations"`
}

type externalObservation struct {
	Fixture         string `json:"fixture"`
	PDFSHA256       string `json:"pdf_sha256"`
	TextCommand     string `json:"text_command"`
	TextTool        string `json:"text_tool"`
	TextExitStatus  int    `json:"text_exit_status"`
	TextOutputPath  string `json:"text_output_path"`
	TextOutputSHA   string `json:"text_output_sha256"`
	RenderCommand   string `json:"render_command"`
	RenderTool      string `json:"render_tool"`
	RenderExit      int    `json:"render_exit_status"`
	RenderOutput    string `json:"render_output_path"`
	RenderOutputSHA string `json:"render_output_sha256"`
}

func TestFinancialTableFixtureBaselines(t *testing.T) {
	evidenceBytes, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("read fixture evidence: %v", err)
	}
	var evidence fixtureEvidence
	decoder := json.NewDecoder(strings.NewReader(string(evidenceBytes)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		t.Fatalf("decode fixture evidence: %v", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("fixture evidence has trailing JSON: %v", err)
	}
	if evidence.SchemaVersion != "1.0.0" || evidence.Scope != "synthetic-fixtures-only" {
		t.Fatalf("evidence schema/scope = %q/%q", evidence.SchemaVersion, evidence.Scope)
	}
	if evidence.ObservationManifest != "docs/fixtures/remediation/financial-tables/observations/observations.v1.json" {
		t.Fatalf("external observation manifest path = %q", evidence.ObservationManifest)
	}
	if !strings.Contains(evidence.ControlsStatus, "classic-xref controls executed in Go tests") || !strings.Contains(evidence.ValidatorLimit, "no table-header association result is claimed") {
		t.Fatalf("validator/control boundary is not explicit: %#v", evidence)
	}
	if len(evidence.Cases) != 2 {
		t.Fatalf("fixture count = %d, want one statement and one invoice", len(evidence.Cases))
	}
	manifestBytes, err := os.ReadFile("../../docs/fixtures/remediation/financial-tables/observations/observations.v1.json")
	if err != nil {
		t.Fatalf("read external observation manifest: %v", err)
	}
	var manifest observationManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode external observation manifest: %v", err)
	}
	if manifest.SchemaVersion != "1.0.0" || manifest.Scope != "reproducible external text and rendering observations for synthetic fixtures" || len(manifest.Observations) != 4 {
		t.Fatalf("external observation manifest version/scope/count = %q/%q/%d", manifest.SchemaVersion, manifest.Scope, len(manifest.Observations))
	}
	observationByFixture := make(map[string]externalObservation, len(manifest.Observations))
	for _, observation := range manifest.Observations {
		observationByFixture[observation.Fixture] = observation
	}
	if !reflect.DeepEqual(evidence.Controls, []string{
		"positive-headers-array-associations",
		"target-data-cell-misassociation",
		"displayed-amount-change",
		"malformed-pdf",
		"unsupported-table-topology",
		"ambiguous-target-map",
		"positive-identifier-name-tree",
		"mcid-to-displayed-content-binding",
		"structparents-parenttree-key",
		"exactly-one-tj-per-cell",
		"single-classic-xref-offsets",
	}) {
		t.Fatalf("pre-repair controls = %#v", evidence.Controls)
	}

	seenFamilies := map[string]bool{}
	for _, fixture := range evidence.Cases {
		fixture := fixture
		t.Run(fixture.Family, func(t *testing.T) {
			if fixture.Family != "statements" && fixture.Family != "invoices" {
				t.Fatalf("unsupported family %q", fixture.Family)
			}
			if seenFamilies[fixture.Family] {
				t.Fatalf("duplicate family %q", fixture.Family)
			}
			seenFamilies[fixture.Family] = true
			if fixture.TargetHeader == "" || fixture.TargetRelation == "" || len(fixture.TargetCells) != 2 || len(fixture.DisplayedStrings) == 0 {
				t.Fatalf("incomplete expected table mapping: %#v", fixture)
			}
			if fixture.Provenance != "authored synthetic data; no source document" || fixture.Topology != "one page; one table; one header row; two data rows; three columns; no spans or merges" || fixture.PDFProfile != "PDF 2.0; English; Helvetica" {
				t.Fatalf("missing declared synthetic assumptions: %#v", fixture)
			}
			if fixture.PositiveRelation != "each TD /Headers array references its associated TH /ID" || fixture.DamagedChange != fmt.Sprintf("derived from the normative positive by changing only the second-row target amount TD /Headers ID to the %s TH /ID", fixture.DamagedWrongHeader) || fixture.DamagedTargetCell == "" || fixture.DamagedWrongHeader == "" {
				t.Fatalf("positive/damaged relationship representation is not explicit: %#v", fixture)
			}
			if fixture.DerivedFromSHA256 != fixture.PositiveSHA256 {
				t.Fatalf("damaged input provenance hash = %s, want positive source hash %s", fixture.DerivedFromSHA256, fixture.PositiveSHA256)
			}
			if len(fixture.ReadingOrder) != 9 || len(fixture.SupportBoundaries) == 0 || len(fixture.Exclusions) == 0 {
				t.Fatalf("missing expected order or support/exclusion boundary: %#v", fixture)
			}
			textHash := sha256.Sum256([]byte(fixture.ExtractedText))
			if fixture.TextSHA256 != hex.EncodeToString(textHash[:]) || fixture.PositiveTextSHA != fixture.TextSHA256 || fixture.DamagedTextSHA != fixture.TextSHA256 {
				t.Fatalf("independent text observation SHA-256 = %s, want %s", hex.EncodeToString(textHash[:]), fixture.TextSHA256)
			}
			if fixture.TextTool == "" || !strings.Contains(fixture.RenderTool, "Ghostscript") || fixture.StructureSource == "" {
				t.Fatalf("missing independent observation provenance: %#v", fixture)
			}
			for _, value := range fixture.DisplayedStrings {
				if !strings.Contains(fixture.ExtractedText, value) {
					t.Fatalf("exact displayed string %q is absent from independent text observation", value)
				}
			}
			if !strings.Contains(fixture.StructureResult, "positive IDTree resolves to all three exact TH elements") || !strings.Contains(fixture.StructureResult, "damaged row-2 target amount points to "+fixture.DamagedWrongHeader) || !strings.Contains(fixture.StructureResult, "actual MCID text") {
				t.Fatalf("structure observation = %q", fixture.StructureResult)
			}
			if fixture.PositiveRenderSHA == "" || fixture.PositiveRenderSHA != fixture.DamagedRenderSHA {
				t.Fatalf("positive/damaged render observations differ: %q != %q", fixture.PositiveRenderSHA, fixture.DamagedRenderSHA)
			}

			positivePath := filepath.Join("..", "..", fixture.PositivePath)
			damagedPath := filepath.Join("..", "..", fixture.DamagedPath)
			assertFixtureXRef(t, positivePath)
			assertFixtureXRef(t, damagedPath)
			positive := readPDFWithDigest(t, positivePath, fixture.PositiveSHA256)
			damaged := readPDFWithDigest(t, damagedPath, fixture.DamagedSHA256)
			for _, item := range []struct {
				path, digest string
			}{
				{fixture.PositivePath, fixture.PositiveSHA256},
				{fixture.DamagedPath, fixture.DamagedSHA256},
			} {
				checkExternalObservation(t, item.path, item.digest, fixture.ExtractedText, fixture.TextTool, fixture.RenderTool, fixture.PositiveRenderSHA, observationByFixture[filepath.Base(item.path)])
			}
			assertReadingOrder(t, positive, fixture.ReadingOrder)
			if err := assertCellContentBindings(positive); err != nil {
				t.Fatalf("positive MCID/content/ParentTree binding: %v", err)
			}
			assertReadingOrder(t, damaged, fixture.ReadingOrder)
			if err := assertCellContentBindings(damaged); err != nil {
				t.Fatalf("damaged MCID/content/ParentTree binding: %v", err)
			}
			assertTableAssociations(t, positive, fixture.TargetHeader, fixture.TargetCells, fixture.DamagedTargetCell, fixture.DamagedWrongHeader, false)
			assertTableAssociations(t, damaged, fixture.TargetHeader, fixture.TargetCells, fixture.DamagedTargetCell, fixture.DamagedWrongHeader, true)
			assertOnlyDeclaredAssociationChanged(t, positive, damaged, fixture.TargetHeader, fixture.DamagedTargetCell)
			assertOnlyObject20Changed(t, positivePath, damagedPath)
		})
	}
	if !seenFamilies["statements"] || !seenFamilies["invoices"] {
		t.Fatalf("fixture families = %#v, want statements and invoices", seenFamilies)
	}
}

// assertFixtureXRef checks the deliberately simple, uncompressed classic
// cross-reference layout used by these synthetic fixtures. It is not a general
// PDF parser; it guards against stale xref blocks and incorrect byte offsets.
func assertFixtureXRef(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture xref %s: %v", path, err)
	}
	xrefMatches := regexp.MustCompile(`(?m)^xref\r?$`).FindAllIndex(data, -1)
	if len(xrefMatches) != 1 {
		t.Fatalf("fixture %s has %d classic xref sections, want exactly one", path, len(xrefMatches))
	}
	xrefOffset := xrefMatches[0][0]
	scanner := bufio.NewScanner(bytes.NewReader(data[xrefOffset:]))
	if !scanner.Scan() || scanner.Text() != "xref" {
		t.Fatalf("fixture %s does not begin its xref section at %d", path, xrefOffset)
	}
	var firstObject, size int
	if !scanner.Scan() {
		t.Fatalf("fixture %s has no xref subsection", path)
	}
	if _, err := fmt.Sscanf(scanner.Text(), "%d %d", &firstObject, &size); err != nil || firstObject != 0 || size < 2 {
		t.Fatalf("fixture %s xref subsection = %q, want 0 <size>", path, scanner.Text())
	}
	for objectNumber := 0; objectNumber < size; objectNumber++ {
		if !scanner.Scan() {
			t.Fatalf("fixture %s xref ends before object %d entry", path, objectNumber)
		}
		var offset, generation int
		var state byte
		if _, err := fmt.Sscanf(scanner.Text(), "%d %d %c", &offset, &generation, &state); err != nil {
			t.Fatalf("fixture %s xref entry %d = %q: %v", path, objectNumber, scanner.Text(), err)
		}
		if objectNumber == 0 {
			if offset != 0 || generation != 65535 || state != 'f' {
				t.Fatalf("fixture %s object-zero xref entry is malformed: %q", path, scanner.Text())
			}
			continue
		}
		wantHeader := []byte(fmt.Sprintf("%d %d obj\n", objectNumber, generation))
		if state != 'n' || offset < 0 || offset >= xrefOffset || !bytes.HasPrefix(data[offset:], wantHeader) {
			t.Fatalf("fixture %s xref object %d offset=%d generation=%d state=%q does not point to its object header", path, objectNumber, offset, generation, state)
		}
	}
	if !scanner.Scan() || scanner.Text() != "trailer" {
		t.Fatalf("fixture %s xref table is not followed by a trailer", path)
	}
	if !scanner.Scan() {
		t.Fatalf("fixture %s has no trailer dictionary", path)
	}
	var trailerSize int
	if _, err := fmt.Sscanf(scanner.Text(), "<< /Size %d", &trailerSize); err != nil || trailerSize != size {
		t.Fatalf("fixture %s trailer size line = %q, want /Size %d", path, scanner.Text(), size)
	}
	if !scanner.Scan() || scanner.Text() != "startxref" || !scanner.Scan() {
		t.Fatalf("fixture %s trailer lacks startxref", path)
	}
	var recordedOffset int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &recordedOffset); err != nil || recordedOffset != xrefOffset {
		t.Fatalf("fixture %s startxref = %q, want byte offset %d", path, scanner.Text(), xrefOffset)
	}
	if !scanner.Scan() || scanner.Text() != "%%EOF" {
		t.Fatalf("fixture %s has no final %%EOF marker", path)
	}
	if scanner.Scan() {
		t.Fatalf("fixture %s has trailing content after %%EOF", path)
	}
}

func assertOnlyObject20Changed(t *testing.T, positivePath, damagedPath string) {
	t.Helper()
	readObjects := func(path string) map[string][]byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read PDF objects from %s: %v", path, err)
		}
		matches := regexp.MustCompile(`(?ms)^(\d+) 0 obj\n(.*?)\nendobj\n`).FindAllSubmatch(data, -1)
		objects := make(map[string][]byte, len(matches))
		for _, match := range matches {
			objects[string(match[1])] = append([]byte(nil), match[2]...)
		}
		return objects
	}
	positive, damaged := readObjects(positivePath), readObjects(damagedPath)
	if len(positive) != len(damaged) || positive["20"] == nil || damaged["20"] == nil {
		t.Fatalf("positive/damaged object inventories differ or lack target cell object 20")
	}
	for number, body := range positive {
		if number == "20" {
			if bytes.Equal(body, damaged[number]) {
				t.Fatal("damaged target cell object 20 has no association change")
			}
			continue
		}
		if !bytes.Equal(body, damaged[number]) {
			t.Fatalf("non-target PDF object %s changed between positive and damaged fixtures", number)
		}
	}
}

func checkExternalObservation(t *testing.T, fixturePath, fixtureSHA, expectedText, expectedTextTool, expectedRenderTool, expectedRenderSHA string, observation externalObservation) {
	t.Helper()
	if observation.Fixture != filepath.Base(fixturePath) || observation.PDFSHA256 != fixtureSHA || observation.TextExitStatus != 0 || observation.RenderExit != 0 {
		t.Fatalf("external observation does not bind successful commands to %s: %#v", fixturePath, observation)
	}
	textVersion := strings.TrimPrefix(observation.TextTool, "pdftotext version ")
	renderVersion := strings.TrimPrefix(strings.Split(observation.RenderTool, ";")[0], "Ghostscript ")
	if observation.TextCommand == "" || observation.RenderCommand == "" || !strings.Contains(expectedTextTool, "pdftotext "+textVersion) || !strings.Contains(expectedRenderTool, renderVersion) {
		t.Fatalf("external tool provenance for %s = %#v", fixturePath, observation)
	}
	if !strings.Contains(observation.TextCommand, fixturePath) || !strings.Contains(observation.RenderCommand, fixturePath) || filepath.Base(observation.TextOutputPath) != strings.TrimSuffix(filepath.Base(fixturePath), ".pdf")+".txt" || filepath.Base(observation.RenderOutput) != strings.TrimSuffix(filepath.Base(fixturePath), ".pdf")+".png" {
		t.Fatalf("external commands/outputs are not bound to fixture path %s: %#v", fixturePath, observation)
	}
	textBytes, err := os.ReadFile(filepath.Join("..", "..", observation.TextOutputPath))
	if err != nil {
		t.Fatalf("read retained extracted text %s: %v", observation.TextOutputPath, err)
	}
	textHash := sha256.Sum256(textBytes)
	if string(textBytes) != expectedText || hex.EncodeToString(textHash[:]) != observation.TextOutputSHA {
		t.Fatalf("retained text observation for %s differs from the recorded extracted text/hash", fixturePath)
	}
	renderPath := filepath.Join("..", "..", observation.RenderOutput)
	renderBytes, err := os.ReadFile(renderPath)
	if err != nil {
		t.Fatalf("read retained rendering %s: %v", observation.RenderOutput, err)
	}
	renderHash := sha256.Sum256(renderBytes)
	if hex.EncodeToString(renderHash[:]) != observation.RenderOutputSHA || observation.RenderOutputSHA != expectedRenderSHA {
		t.Fatalf("retained render observation for %s has SHA-256 %x, want %s", fixturePath, renderHash, expectedRenderSHA)
	}
}

func TestStructureBindingNegativeControls(t *testing.T) {
	ctx := readPDFWithDigest(t, filepath.Join("..", "..", "docs/fixtures/remediation/financial-tables/statement-positive.pdf"), "14c0456e3b8ec0146bc2a25fa9af51d0c5559f8a53e23c4e163a3c422e66cfed")
	beforeHeaders, beforeRows := tableObservations(t, ctx)
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatal(err)
	}
	if err := assertIDTree(ctx, root); err != nil {
		t.Fatalf("valid normative IDTree rejected: %v", err)
	}

	missing := types.Dict{}
	for key, value := range root {
		missing[key] = value
	}
	delete(missing, "IDTree")
	if err := assertIDTree(ctx, missing); err == nil || !strings.Contains(err.Error(), "IDTree") {
		t.Fatalf("missing IDTree error = %v, want explicit identifier-tree failure", err)
	}

	wrong := types.Dict{}
	for key, value := range root {
		wrong[key] = value
	}
	tree, err := ctx.DereferenceDict(root["IDTree"])
	if err != nil {
		t.Fatal(err)
	}
	wrongTree := types.Dict{}
	for key, value := range tree {
		wrongTree[key] = value
	}
	names, err := resolvedArray(ctx, wrongTree["Names"])
	if err != nil {
		t.Fatal(err)
	}
	wrongNames := append(types.Array(nil), names...)
	wrongNames[1] = types.IndirectRef{ObjectNumber: 11}
	wrongTree["Names"] = wrongNames
	wrong["IDTree"] = wrongTree
	if err := assertIDTree(ctx, wrong); err == nil || !strings.Contains(err.Error(), "element ID") {
		t.Fatalf("wrong IDTree target error = %v, want target-ID mismatch", err)
	}

	if err := assertCellContentBindings(ctx); err != nil {
		t.Fatalf("valid positive binding rejected: %v", err)
	}
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := assertStructParentsMatch(page, types.Integer(0)); err != nil {
		t.Fatalf("valid StructParents binding rejected: %v", err)
	}
	missingStructParents := types.Dict{}
	for key, value := range page {
		missingStructParents[key] = value
	}
	delete(missingStructParents, "StructParents")
	if err := assertStructParentsMatch(missingStructParents, types.Integer(0)); err == nil || !strings.Contains(err.Error(), "/StructParents") {
		t.Fatalf("missing StructParents error = %v, want explicit page-key failure", err)
	}
	wrongStructParents := types.Dict{}
	for key, value := range page {
		wrongStructParents[key] = value
	}
	wrongStructParents["StructParents"] = types.Integer(1)
	if err := assertStructParentsMatch(wrongStructParents, types.Integer(0)); err == nil || !strings.Contains(err.Error(), "does not match ParentTree key") {
		t.Fatalf("wrong StructParents error = %v, want mismatched ParentTree-key failure", err)
	}

	// A second simple Tj in a cell sequence must not be silently ignored.
	textCtx := readPDFWithDigest(t, filepath.Join("..", "..", "docs/fixtures/remediation/financial-tables/statement-positive.pdf"), "14c0456e3b8ec0146bc2a25fa9af51d0c5559f8a53e23c4e163a3c422e66cfed")
	textPage, _, _, err := textCtx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatal(err)
	}
	textStream, _, err := textCtx.DereferenceStreamDict(textPage["Contents"])
	if err != nil {
		t.Fatal(err)
	}
	if err := textStream.Decode(); err != nil {
		t.Fatal(err)
	}
	textStream.Content = bytes.Replace(textStream.Content, []byte("(Date) Tj ET EMC"), []byte("(Date) Tj (Unexpected) Tj ET EMC"), 1)
	textRef := textPage["Contents"].(types.IndirectRef)
	textCtx.Table[textRef.ObjectNumber.Value()].Object = *textStream
	if err := assertCellContentBindings(textCtx); err == nil || !strings.Contains(err.Error(), "exactly one simple literal Tj string") {
		t.Fatalf("second-Tj mutation error = %v, want exact-one-string failure", err)
	}

	stream, _, err := ctx.DereferenceStreamDict(page["Contents"])
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Decode(); err != nil {
		t.Fatal(err)
	}
	// Swap which text is enclosed by MCIDs 7 and 8. /T and every /Headers
	// array stay untouched; the semantic content binding must still reject it.
	stream.Content = bytes.ReplaceAll(stream.Content, []byte("/MCID 7 >>"), []byte("/MCID temporary >>"))
	stream.Content = bytes.ReplaceAll(stream.Content, []byte("/MCID 8 >>"), []byte("/MCID 7 >>"))
	stream.Content = bytes.ReplaceAll(stream.Content, []byte("/MCID temporary >>"), []byte("/MCID 8 >>"))
	ref := page["Contents"].(types.IndirectRef)
	ctx.Table[ref.ObjectNumber.Value()].Object = *stream
	afterHeaders, afterRows := tableObservations(t, ctx)
	if !reflect.DeepEqual(beforeHeaders, afterHeaders) || !reflect.DeepEqual(beforeRows, afterRows) {
		t.Fatal("MCID content mutation changed structure titles or /Headers arrays")
	}
	if err := assertCellContentBindings(ctx); err == nil || !strings.Contains(err.Error(), "binds displayed content") {
		t.Fatalf("MCID/content mutation error = %v, want displayed-content binding failure", err)
	}
}

func assertReadingOrder(t *testing.T, ctx *model.Context, want []string) {
	t.Helper()
	headers, rows := tableObservations(t, ctx)
	got := make([]string, 0, len(want))
	for _, header := range headers {
		got = append(got, header.title)
	}
	for _, row := range rows {
		for _, cell := range row {
			got = append(got, cell.title)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("structure reading order = %#v, want %#v", got, want)
	}
}

func readPDFWithDigest(t *testing.T, path, wantSHA256 string) *model.Context {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read PDF %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != wantSHA256 {
		t.Fatalf("PDF %s SHA-256 = %s, want %s", path, got, wantSHA256)
	}
	if !strings.HasPrefix(string(data), "%PDF-2.0") {
		t.Fatalf("PDF %s does not declare the protocol's PDF 2.0 input assumption", path)
	}
	ctx, err := api.ReadContextFile(context.Background(), path)
	if err != nil {
		t.Fatalf("parse PDF %s: %v", path, err)
	}
	return ctx
}

type headerObservation struct {
	title string
	id    string
	scope string
	index int
}

type cellObservation struct {
	title     string
	role      string
	headerIDs []string
	index     int
}

func tableObservations(t *testing.T, ctx *model.Context) ([]headerObservation, [][]cellObservation) {
	t.Helper()
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatalf("resolve StructTreeRoot: %v", err)
	}
	if err := assertIDTree(ctx, root); err != nil {
		t.Fatalf("StructTreeRoot IDTree: %v", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil || doc["S"] != types.Name("Document") {
		t.Fatalf("structure root K = %v (err %v), want Document element", root["K"], err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil || table["S"] != types.Name("Table") {
		t.Fatalf("Document K = %v (err %v), want Table element", doc["K"], err)
	}
	rowRefs, ok := table["K"].(types.Array)
	if !ok || len(rowRefs) != 3 {
		t.Fatalf("Table K = %v, want one header row and two data rows", table["K"])
	}
	var headers []headerObservation
	var dataRows [][]cellObservation
	for rowIndex, rowRef := range rowRefs {
		row, err := ctx.DereferenceDict(rowRef)
		if err != nil || row["S"] != types.Name("TR") {
			t.Fatalf("Table row %d = %v (err %v), want TR", rowIndex, rowRef, err)
		}
		cellRefs, ok := row["K"].(types.Array)
		if !ok || len(cellRefs) != 3 {
			t.Fatalf("Table row %d cells = %v, want three ordered cells", rowIndex, row["K"])
		}
		var cells []cellObservation
		for cellIndex, cellRef := range cellRefs {
			cell, err := ctx.DereferenceDict(cellRef)
			if err != nil {
				t.Fatalf("resolve row %d cell %d: %v", rowIndex, cellIndex, err)
			}
			text, ok := cell["T"].(types.StringLiteral)
			if !ok {
				t.Fatalf("row %d cell %d title = %T, want string", rowIndex, cellIndex, cell["T"])
			}
			observation := cellObservation{title: string(text), index: cellIndex}
			if rowIndex == 0 {
				if cell["S"] != types.Name("TH") {
					t.Fatalf("header cell %d tag = %v, want TH", cellIndex, cell["S"])
				}
				attributes, ok := cell["A"].(types.Dict)
				if !ok {
					t.Fatalf("header cell %d A = %T, want attribute dictionary", cellIndex, cell["A"])
				}
				if attributes["O"] != types.Name("Table") {
					t.Fatalf("header cell %d attribute owner = %v, want Table", cellIndex, attributes["O"])
				}
				id, ok := cell["ID"].(types.StringLiteral)
				if !ok {
					t.Fatalf("header cell %d ID = %T, want string", cellIndex, cell["ID"])
				}
				scope, ok := attributes["Scope"].(types.Name)
				if !ok {
					t.Fatalf("header cell %d Scope = %T, want name", cellIndex, attributes["Scope"])
				}
				headers = append(headers, headerObservation{title: string(text), id: string(id), scope: string(scope), index: cellIndex})
			} else if cell["S"] != types.Name("TD") {
				t.Fatalf("data cell row %d col %d tag = %v, want TD", rowIndex, cellIndex, cell["S"])
			} else {
				observation.role = "TD"
				attributes, ok := cell["A"].(types.Dict)
				if !ok || attributes["O"] != types.Name("Table") {
					t.Fatalf("data cell row %d col %d table attributes = %v, want Table owner", rowIndex, cellIndex, cell["A"])
				}
				refs, ok := attributes["Headers"].(types.Array)
				if !ok {
					t.Fatalf("data cell row %d col %d Headers = %T, want array", rowIndex, cellIndex, attributes["Headers"])
				}
				for _, ref := range refs {
					headerID, ok := ref.(types.StringLiteral)
					if !ok {
						t.Fatalf("data cell row %d col %d Headers member = %T, want string ID", rowIndex, cellIndex, ref)
					}
					observation.headerIDs = append(observation.headerIDs, string(headerID))
				}
			}
			cells = append(cells, observation)
		}
		if rowIndex > 0 {
			dataRows = append(dataRows, cells)
		}
	}
	return headers, dataRows
}

func assertIDTree(ctx *model.Context, root types.Dict) error {
	nameTree, err := ctx.DereferenceDict(root["IDTree"])
	if err != nil {
		return fmt.Errorf("resolve /IDTree: %w", err)
	}
	names, err := resolvedArray(ctx, nameTree["Names"])
	if err != nil || len(names) == 0 || len(names)%2 != 0 {
		return fmt.Errorf("/IDTree /Names = %v (err %v), want nonempty name/value pairs", names, err)
	}
	seen := make(map[string]types.IndirectRef, len(names)/2)
	previous := ""
	for i := 0; i < len(names); i += 2 {
		name, ok := names[i].(types.StringLiteral)
		if !ok {
			return fmt.Errorf("/IDTree key %d has type %T, want byte string", i/2, names[i])
		}
		id := string(name)
		if id <= previous {
			return fmt.Errorf("/IDTree keys are not unique and sorted at %q", id)
		}
		previous = id
		ref, ok := names[i+1].(types.IndirectRef)
		if !ok {
			return fmt.Errorf("/IDTree value for %q has type %T, want indirect reference", id, names[i+1])
		}
		cell, err := ctx.DereferenceDict(ref)
		if err != nil || cell["S"] != types.Name("TH") {
			return fmt.Errorf("/IDTree value for %q does not resolve to TH: %v (err %v)", id, cell, err)
		}
		cellID, ok := cell["ID"].(types.StringLiteral)
		if !ok || string(cellID) != id {
			return fmt.Errorf("/IDTree key %q resolves to element ID %v", id, cell["ID"])
		}
		seen[id] = ref
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return fmt.Errorf("resolve document for IDTree coverage: %w", err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return fmt.Errorf("resolve table for IDTree coverage: %w", err)
	}
	rows, ok := table["K"].(types.Array)
	if !ok || len(rows) == 0 {
		return fmt.Errorf("table has no header row for IDTree coverage")
	}
	headerRow, err := ctx.DereferenceDict(rows[0])
	if err != nil {
		return fmt.Errorf("resolve header row: %w", err)
	}
	headers, ok := headerRow["K"].(types.Array)
	if !ok || len(headers) != len(seen) {
		return fmt.Errorf("IDTree has %d entries, table has %d header cells", len(seen), len(headers))
	}
	for _, headerRef := range headers {
		header, err := ctx.DereferenceDict(headerRef)
		if err != nil || header["S"] != types.Name("TH") {
			return fmt.Errorf("header reference %v does not resolve to TH: %v (err %v)", headerRef, header, err)
		}
		id, ok := header["ID"].(types.StringLiteral)
		if !ok || seen[string(id)] != headerRef {
			return fmt.Errorf("actual TH %v identifier %v does not map back through IDTree", headerRef, header["ID"])
		}
	}
	return nil
}

func resolvedArray(ctx *model.Context, obj types.Object) (types.Array, error) {
	if array, ok := obj.(types.Array); ok {
		return array, nil
	}
	return ctx.DereferenceArray(obj)
}

// assertCellContentBindings checks this deliberately narrow fixture syntax: one
// page content stream, one /MCID marked-content sequence per table cell, and one
// literal Tj string in each sequence. pdfcpu decodes the PDF stream; this small
// matcher is not a general PDF content parser.
func assertCellContentBindings(ctx *model.Context) error {
	page, pageRef, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil || pageRef == nil {
		return fmt.Errorf("resolve page 1: %v", err)
	}
	contents, ok := page["Contents"].(types.IndirectRef)
	if !ok {
		return fmt.Errorf("page /Contents has type %T, want one indirect stream", page["Contents"])
	}
	stream, _, err := ctx.DereferenceStreamDict(contents)
	if err != nil {
		return fmt.Errorf("resolve page content stream: %w", err)
	}
	if err := stream.Decode(); err != nil {
		return fmt.Errorf("decode page content stream: %w", err)
	}
	sequence := regexp.MustCompile(`(?s)/Span\s*<<\s*/MCID\s+(\d+)\s*>>\s*BDC(.*?)EMC`)
	text := regexp.MustCompile(`\(([^()]*)\)\s*Tj`)
	actual := make(map[int]string)
	for _, match := range sequence.FindAllSubmatch(stream.Content, -1) {
		var mcid int
		if _, err := fmt.Sscan(string(match[1]), &mcid); err != nil {
			return fmt.Errorf("parse fixture MCID: %w", err)
		}
		stringsFound := text.FindAllSubmatch(match[2], -1)
		if len(stringsFound) != 1 || len(stringsFound[0]) != 2 {
			return fmt.Errorf("MCID %d does not contain exactly one simple literal Tj string", mcid)
		}
		if _, exists := actual[mcid]; exists {
			return fmt.Errorf("duplicate marked-content MCID %d", mcid)
		}
		actual[mcid] = string(stringsFound[0][1])
	}
	if len(actual) != 9 {
		return fmt.Errorf("found %d marked table cells, want 9", len(actual))
	}

	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return fmt.Errorf("resolve StructTreeRoot: %w", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return fmt.Errorf("resolve document element: %w", err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return fmt.Errorf("resolve table element: %w", err)
	}
	rows, ok := table["K"].(types.Array)
	if !ok || len(rows) != 3 {
		return fmt.Errorf("table rows have shape %T/%d, want three rows", table["K"], len(rows))
	}
	parentTree, err := ctx.DereferenceDict(root["ParentTree"])
	if err != nil {
		return fmt.Errorf("resolve ParentTree: %w", err)
	}
	nums, err := resolvedArray(ctx, parentTree["Nums"])
	if err != nil || len(nums) != 2 {
		return fmt.Errorf("ParentTree /Nums = %v (err %v), want one page key/value pair", nums, err)
	}
	if err := assertStructParentsMatch(page, nums[0]); err != nil {
		return err
	}
	parents, err := resolvedArray(ctx, nums[1])
	if err != nil || len(parents) != 9 {
		return fmt.Errorf("ParentTree page array = %v (err %v), want 9 MCID bindings", parents, err)
	}
	for _, rowRef := range rows {
		row, err := ctx.DereferenceDict(rowRef)
		if err != nil {
			return fmt.Errorf("resolve row: %w", err)
		}
		cells, ok := row["K"].(types.Array)
		if !ok || len(cells) != 3 {
			return fmt.Errorf("row cells have shape %T/%d, want three", row["K"], len(cells))
		}
		for _, cellRef := range cells {
			cell, err := ctx.DereferenceDict(cellRef)
			if err != nil {
				return fmt.Errorf("resolve cell: %w", err)
			}
			mcid, ok := cell["K"].(types.Integer)
			if !ok || int(mcid) < 0 || int(mcid) >= len(parents) {
				return fmt.Errorf("cell /K = %v, want MCID 0 through 8", cell["K"])
			}
			if parents[int(mcid)] != cellRef {
				return fmt.Errorf("ParentTree MCID %d points to %v, want cell %v", mcid, parents[int(mcid)], cellRef)
			}
			if cell["P"] != rowRef {
				return fmt.Errorf("cell MCID %d /P = %v, want containing row %v", mcid, cell["P"], rowRef)
			}
			if cell["Pg"] != *pageRef {
				return fmt.Errorf("cell MCID %d /Pg = %v, want page %v", mcid, cell["Pg"], *pageRef)
			}
			if actual[int(mcid)] != string(cell["T"].(types.StringLiteral)) {
				return fmt.Errorf("cell MCID %d binds displayed content %q to structure title %q", mcid, actual[int(mcid)], cell["T"])
			}
		}
	}
	return nil
}

func assertStructParentsMatch(page types.Dict, parentTreeKey types.Object) error {
	structParents, ok := page["StructParents"].(types.Integer)
	if !ok {
		return fmt.Errorf("page /StructParents = %v, want integer matching its ParentTree key", page["StructParents"])
	}
	if parentTreeKey != structParents {
		return fmt.Errorf("page /StructParents %d does not match ParentTree key %v", structParents, parentTreeKey)
	}
	return nil
}

func assertTableAssociations(t *testing.T, ctx *model.Context, targetHeader string, targetCells []string, damagedTargetCell, damagedWrongHeader string, damaged bool) {
	t.Helper()
	headers, rows := tableObservations(t, ctx)
	targetIndex := -1
	headerIDByColumn := make([]string, len(headers))
	headerIDByTitle := make(map[string]string, len(headers))
	for _, header := range headers {
		if header.scope != "Column" {
			t.Fatalf("header %q scope = %q, want preserved Column", header.title, header.scope)
		}
		headerIDByColumn[header.index] = header.id
		headerIDByTitle[header.title] = header.id
		if header.title == targetHeader {
			if targetIndex != -1 {
				t.Fatalf("target header %q is ambiguous", targetHeader)
			}
			targetIndex = header.index
		}
	}
	if targetIndex < 0 {
		t.Fatalf("target header %q missing", targetHeader)
	}
	if headerIDByTitle[damagedWrongHeader] == "" {
		t.Fatalf("damaged relationship's wrong header %q missing", damagedWrongHeader)
	}
	var gotCells []string
	associationDefects := 0
	for rowIndex, row := range rows {
		if len(row) != len(headerIDByColumn) {
			t.Fatalf("data row %d width = %d, want %d", rowIndex+1, len(row), len(headerIDByColumn))
		}
		for colIndex, cell := range row {
			wantHeaderIDs := []string{headerIDByColumn[colIndex]}
			isDeclaredDefect := damaged && rowIndex == 1 && colIndex == targetIndex && cell.title == damagedTargetCell
			if isDeclaredDefect {
				wantHeaderIDs = []string{headerIDByTitle[damagedWrongHeader]}
				associationDefects++
			}
			if !reflect.DeepEqual(cell.headerIDs, wantHeaderIDs) {
				t.Fatalf("data row %d cell %q header IDs = %#v, want %#v", rowIndex+1, cell.title, cell.headerIDs, wantHeaderIDs)
			}
		}
		gotCells = append(gotCells, row[targetIndex].title)
	}
	if associationDefects != boolCount(damaged) {
		t.Fatalf("declared association defects = %d, want %d", associationDefects, boolCount(damaged))
	}
	if !reflect.DeepEqual(gotCells, targetCells) {
		t.Fatalf("amount cells in target column = %#v, want %#v", gotCells, targetCells)
	}
}

func assertOnlyDeclaredAssociationChanged(t *testing.T, positive, damaged *model.Context, targetHeader, damagedTargetCell string) {
	t.Helper()
	positiveHeaders, positiveRows := tableObservations(t, positive)
	damagedHeaders, damagedRows := tableObservations(t, damaged)
	if !reflect.DeepEqual(positiveHeaders, damagedHeaders) {
		t.Fatalf("header identity/order/scope changed: %#v != %#v", positiveHeaders, damagedHeaders)
	}
	if len(positiveRows) != len(damagedRows) {
		t.Fatalf("data row count changed: %d != %d", len(positiveRows), len(damagedRows))
	}
	targetIndex := -1
	for _, header := range positiveHeaders {
		if header.title == targetHeader {
			targetIndex = header.index
		}
	}
	if targetIndex < 0 {
		t.Fatalf("target header %q missing while comparing fixtures", targetHeader)
	}
	for rowIndex := range positiveRows {
		if len(positiveRows[rowIndex]) != len(damagedRows[rowIndex]) {
			t.Fatalf("row %d cell count changed", rowIndex+1)
		}
		for colIndex := range positiveRows[rowIndex] {
			before, after := positiveRows[rowIndex][colIndex], damagedRows[rowIndex][colIndex]
			if before.title != after.title || before.role != after.role || before.index != after.index {
				t.Fatalf("non-association cell structure changed at row %d col %d: %#v != %#v", rowIndex+1, colIndex, before, after)
			}
			isTarget := rowIndex == 1 && colIndex == targetIndex && before.title == damagedTargetCell
			if !isTarget && !reflect.DeepEqual(before.headerIDs, after.headerIDs) {
				t.Fatalf("non-target association changed at row %d col %d: %#v != %#v", rowIndex+1, colIndex, before.headerIDs, after.headerIDs)
			}
		}
	}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestFixtureEvidencePathsStayWithinSyntheticCorpus(t *testing.T) {
	data, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatalf("read fixture evidence: %v", err)
	}
	var evidence fixtureEvidence
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range evidence.Cases {
		for _, name := range []string{fixture.PositivePath, fixture.DamagedPath} {
			if filepath.IsAbs(name) || filepath.Clean(name) != name || strings.HasPrefix(name, "..") {
				t.Fatalf("fixture path escapes corpus: %q", name)
			}
			if !strings.HasPrefix(name, "docs/fixtures/remediation/financial-tables/") || !strings.HasSuffix(name, ".pdf") {
				t.Fatalf("fixture path outside financial table package: %q", name)
			}
		}
	}
}
