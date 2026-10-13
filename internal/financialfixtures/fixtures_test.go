package financialfixtures

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const evidencePath = "../../docs/fixtures/remediation/financial-table-experiment.v1.json"

type fixtureEvidence struct {
	SchemaVersion  string          `json:"schema_version"`
	Scope          string          `json:"scope"`
	ControlsStatus string          `json:"controls_status"`
	ValidatorLimit string          `json:"validator_limit"`
	Cases          []fixtureRecord `json:"cases"`
	Controls       []string        `json:"controls"`
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
	if evidence.ControlsStatus != "defined-before-repair; amount-change and failure controls not executed" || !strings.Contains(evidence.ValidatorLimit, "no table-header association result is claimed") {
		t.Fatalf("validator/control boundary is not explicit: %#v", evidence)
	}
	if len(evidence.Cases) != 2 {
		t.Fatalf("fixture count = %d, want one statement and one invoice", len(evidence.Cases))
	}
	if !reflect.DeepEqual(evidence.Controls, []string{
		"positive-headers-array-associations",
		"target-data-cell-misassociation",
		"displayed-amount-change",
		"malformed-pdf",
		"unsupported-table-topology",
		"ambiguous-target-map",
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
			if fixture.StructureResult != fmt.Sprintf("positive exact header-ID mapping; damaged row-2 target amount maps to %s; remaining cell-header IDs and order preserved", fixture.DamagedWrongHeader) {
				t.Fatalf("structure observation = %q", fixture.StructureResult)
			}
			if fixture.PositiveRenderSHA == "" || fixture.PositiveRenderSHA != fixture.DamagedRenderSHA {
				t.Fatalf("positive/damaged render observations differ: %q != %q", fixture.PositiveRenderSHA, fixture.DamagedRenderSHA)
			}

			positivePath := filepath.Join("..", "..", fixture.PositivePath)
			damagedPath := filepath.Join("..", "..", fixture.DamagedPath)
			positive := readPDFWithDigest(t, positivePath, fixture.PositiveSHA256)
			damaged := readPDFWithDigest(t, damagedPath, fixture.DamagedSHA256)
			assertReadingOrder(t, positive, fixture.ReadingOrder)
			assertReadingOrder(t, damaged, fixture.ReadingOrder)
			assertTableAssociations(t, positive, fixture.TargetHeader, fixture.TargetCells, fixture.DamagedTargetCell, fixture.DamagedWrongHeader, false)
			assertTableAssociations(t, damaged, fixture.TargetHeader, fixture.TargetCells, fixture.DamagedTargetCell, fixture.DamagedWrongHeader, true)
			assertOnlyDeclaredAssociationChanged(t, positive, damaged, fixture.TargetHeader, fixture.DamagedTargetCell)
		})
	}
	if !seenFamilies["statements"] || !seenFamilies["invoices"] {
		t.Fatalf("fixture families = %#v, want statements and invoices", seenFamilies)
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
