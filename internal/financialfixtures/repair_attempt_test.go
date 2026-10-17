package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// v2RepairSpec describes one fixture-authored row-2 amount-cell /Headers
// repair: the single header ID to replace and its replacement. This helper is
// kept in a test file so pdfcpu remains outside ordinary package importers.
type v2RepairSpec struct {
	DamagedHeaderID  string
	RepairedHeaderID string
}

// repairV2Row2AmountHeaders rewrites the row-2 amount TD's /A /Headers name
// array to the repaired header ID and returns the repaired structure
// element's object number. It fails closed unless the current /Headers is
// exactly the single damaged header ID: an already-repaired or differently
// damaged cell is never rewritten. Only the target /Headers array changes.
func repairV2Row2AmountHeaders(ctx *model.Context, spec v2RepairSpec) (int, error) {
	if spec.DamagedHeaderID == "" || spec.RepairedHeaderID == "" {
		return 0, fmt.Errorf("repair spec needs damaged and repaired header IDs")
	}
	if spec.DamagedHeaderID == spec.RepairedHeaderID {
		return 0, fmt.Errorf("damaged and repaired header IDs are identical: %q", spec.DamagedHeaderID)
	}
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return 0, fmt.Errorf("resolve StructTreeRoot: %w", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return 0, fmt.Errorf("resolve Document element: %w", err)
	}
	if doc["S"] != types.Name("Document") {
		return 0, fmt.Errorf("structure root K is %v, want Document", doc["S"])
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil {
		return 0, fmt.Errorf("resolve Table element: %w", err)
	}
	if table["S"] != types.Name("Table") {
		return 0, fmt.Errorf("Document K is %v, want Table", table["S"])
	}
	rowRefs, ok := table["K"].(types.Array)
	if !ok || len(rowRefs) != 3 {
		return 0, fmt.Errorf("Table K has %d rows, want one header row and two data rows", len(rowRefs))
	}
	rowRef, ok := rowRefs[2].(types.IndirectRef)
	if !ok {
		return 0, fmt.Errorf("row 2 ref is %T, want indirect reference", rowRefs[2])
	}
	row, err := ctx.DereferenceDict(rowRef)
	if err != nil {
		return 0, fmt.Errorf("resolve row 2: %w", err)
	}
	if row["S"] != types.Name("TR") {
		return 0, fmt.Errorf("row 2 is %v, want TR", row["S"])
	}
	cellRefs, ok := row["K"].(types.Array)
	if !ok || len(cellRefs) != 3 {
		return 0, fmt.Errorf("row 2 has %d cells, want three", len(cellRefs))
	}
	for _, cellRef := range cellRefs {
		cref, ok := cellRef.(types.IndirectRef)
		if !ok {
			return 0, fmt.Errorf("row 2 cell ref is %T, want indirect reference", cellRef)
		}
		cell, err := ctx.DereferenceDict(cref)
		if err != nil {
			return 0, fmt.Errorf("resolve row 2 cell %d: %w", cref.ObjectNumber, err)
		}
		text, ok := cell["T"].(types.StringLiteral)
		if !ok || !v2AmountCellText.MatchString(string(text)) {
			continue
		}
		if cell["S"] != types.Name("TD") {
			return 0, fmt.Errorf("amount cell tag is %v, want TD", cell["S"])
		}
		attrs, ok := cell["A"].(types.Dict)
		if !ok {
			return 0, fmt.Errorf("amount cell A is %T, want attribute dictionary", cell["A"])
		}
		if attrs["O"] != types.Name("Table") {
			return 0, fmt.Errorf("amount cell attribute owner is %v, want Table", attrs["O"])
		}
		raw, ok := attrs["Headers"].(types.Array)
		if !ok || len(raw) != 1 {
			return 0, fmt.Errorf("amount cell Headers has %d entries, want exactly one", len(raw))
		}
		current, ok := raw[0].(types.StringLiteral)
		if !ok {
			return 0, fmt.Errorf("amount cell Headers member is %T, want string ID", raw[0])
		}
		if string(current) != spec.DamagedHeaderID {
			return 0, fmt.Errorf("amount cell Headers = [%s], want damaged [%s]; refusing to rewrite",
				string(current), spec.DamagedHeaderID)
		}
		attrs["Headers"] = types.Array{types.StringLiteral(spec.RepairedHeaderID)}
		return int(cref.ObjectNumber), nil
	}
	return 0, fmt.Errorf("row 2 has no amount cell matching %q", v2AmountCellText)
}

// Fixture-authored damage characterization for the bounded v2 repair
// attempt (dap-jo9.1.6.3). Each damaged fixture has exactly one wrong
// row-2 amount-cell /Headers association; everything else is correct.
// This is the single canonical table pinning both damaged and repaired
// candidate digests; the text and render preservation tests reference it.
var v2RepairAttemptCases = []struct {
	name             string
	relPath          string
	wantSHA256       string
	damagedHeaderID  string
	repairedHeaderID string
	candidateRelPath string
	candidateSHA256  string
}{
	{
		name:             "statement",
		relPath:          "../../docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
		wantSHA256:       "73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836",
		damagedHeaderID:  "descriptionHeader",
		repairedHeaderID: "amountHeader",
		candidateRelPath: "../../docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf",
		candidateSHA256:  "8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52",
	},
	{
		name:             "invoice",
		relPath:          "../../docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
		wantSHA256:       "d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad",
		damagedHeaderID:  "quantityHeader",
		repairedHeaderID: "unitAmountHeader",
		candidateRelPath: "../../docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf",
		candidateSHA256:  "31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448",
	},
}

var v2AmountCellText = regexp.MustCompile(`^\$[0-9]`)

// repairTarget locates the row-2 amount TD of a v2 fixture: the second data
// row's TD whose /T text matches the amount pattern.
type repairTarget struct {
	objectNumber int
	cell         types.Dict
	headerIDs    []string
}

func locateV2Row2AmountCell(t *testing.T, ctx *model.Context) repairTarget {
	t.Helper()
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatalf("resolve StructTreeRoot: %v", err)
	}
	doc, err := ctx.DereferenceDict(root["K"])
	if err != nil || doc["S"] != types.Name("Document") {
		t.Fatalf("structure root K = %v (err %v), want Document", root["K"], err)
	}
	table, err := ctx.DereferenceDict(doc["K"])
	if err != nil || table["S"] != types.Name("Table") {
		t.Fatalf("Document K = %v (err %v), want Table", doc["K"], err)
	}
	rowRefs, ok := table["K"].(types.Array)
	if !ok || len(rowRefs) != 3 {
		t.Fatalf("Table K has %d rows, want one header row and two data rows", len(rowRefs))
	}
	rowRef, ok := rowRefs[2].(types.IndirectRef)
	if !ok {
		t.Fatalf("row 2 ref = %T, want indirect reference", rowRefs[2])
	}
	row, err := ctx.DereferenceDict(rowRef)
	if err != nil || row["S"] != types.Name("TR") {
		t.Fatalf("row 2 = %v (err %v), want TR", rowRefs[2], err)
	}
	cellRefs, ok := row["K"].(types.Array)
	if !ok || len(cellRefs) != 3 {
		t.Fatalf("row 2 has %d cells, want three", len(cellRefs))
	}
	for _, cellRef := range cellRefs {
		cref, ok := cellRef.(types.IndirectRef)
		if !ok {
			t.Fatalf("row 2 cell ref = %T, want indirect reference", cellRef)
		}
		cell, err := ctx.DereferenceDict(cref)
		if err != nil {
			t.Fatalf("resolve row 2 cell %d: %v", cref.ObjectNumber, err)
		}
		text, ok := cell["T"].(types.StringLiteral)
		if !ok || !v2AmountCellText.MatchString(string(text)) {
			continue
		}
		if cell["S"] != types.Name("TD") {
			t.Fatalf("amount cell tag = %v, want TD", cell["S"])
		}
		attrs, ok := cell["A"].(types.Dict)
		if !ok {
			t.Fatalf("amount cell A = %T, want attribute dictionary", cell["A"])
		}
		raw, ok := attrs["Headers"].(types.Array)
		if !ok {
			t.Fatalf("amount cell Headers = %T, want array", attrs["Headers"])
		}
		var ids []string
		for _, h := range raw {
			id, ok := h.(types.StringLiteral)
			if !ok {
				t.Fatalf("amount cell Headers member = %T, want string", h)
			}
			ids = append(ids, string(id))
		}
		return repairTarget{objectNumber: int(cref.ObjectNumber), cell: cell, headerIDs: ids}
	}
	t.Fatalf("row 2 has no amount cell matching %q", v2AmountCellText)
	return repairTarget{}
}

const (
	maxStructSnapshotDepth  = 64
	maxStructSnapshotValues = 1024
	maxStructSnapshotArray  = 256
	maxStructSnapshotNodes  = 256
)

// snapshotStructTree serializes the indirect structure dictionaries reachable
// from StructTreeRoot, keyed by object number. It walks both singular and
// array-valued /K entries. The small limits are specific to these bounded v2
// fixtures; malformed, cyclic, shared, or unexpectedly large trees fail
// closed instead of producing an incomplete preservation snapshot.
func snapshotStructTree(ctx *model.Context) (map[int]string, error) {
	rootRef, ok := ctx.RootDict["StructTreeRoot"].(types.IndirectRef)
	if !ok {
		return nil, fmt.Errorf("StructTreeRoot = %T, want indirect reference", ctx.RootDict["StructTreeRoot"])
	}
	out := map[int]string{}
	active := map[int]bool{}
	visited := map[int]bool{}
	valuesSeen := 0

	var walkValue func(types.Object, int, int) error
	var walkRef func(types.IndirectRef, int, bool) error
	walkRef = func(ref types.IndirectRef, depth int, root bool) error {
		objectNumber := int(ref.ObjectNumber)
		if depth > maxStructSnapshotDepth {
			return fmt.Errorf("structure tree exceeds %d-level depth limit at object %d", maxStructSnapshotDepth, objectNumber)
		}
		if active[objectNumber] {
			return fmt.Errorf("structure tree contains a /K cycle at object %d", objectNumber)
		}
		if visited[objectNumber] {
			return fmt.Errorf("structure tree references object %d more than once", objectNumber)
		}
		if len(visited) >= maxStructSnapshotNodes {
			return fmt.Errorf("structure tree exceeds %d indirect-object limit", maxStructSnapshotNodes)
		}
		visited[objectNumber] = true
		active[objectNumber] = true
		defer delete(active, objectNumber)

		value, err := ctx.Dereference(ref)
		if err != nil {
			return fmt.Errorf("dereference structure child %d: %w", objectNumber, err)
		}
		switch value := value.(type) {
		case types.Array:
			return walkValue(value, depth+1, objectNumber)
		case types.Integer:
			if value < 0 {
				return fmt.Errorf("indirect /K MCID object %d is negative", objectNumber)
			}
			return nil
		case types.Dict:
			if !root {
				if value["Type"] == types.Name("MCR") || value["Type"] == types.Name("OBJR") {
					return nil
				}
				if _, direct := value["S"]; direct {
					if _, ok := value["S"].(types.Name); !ok {
						return fmt.Errorf("structure child object %d has malformed /S type %T", objectNumber, value["S"])
					}
				} else {
					return fmt.Errorf("indirect /K object %d is neither a structure element nor an MCR/OBJR", objectNumber)
				}
			}
			out[objectNumber] = value.String()
			if kid, exists := value["K"]; exists {
				if kid == nil {
					return fmt.Errorf("structure object %d has a nil /K value", objectNumber)
				}
				return walkValue(kid, depth+1, objectNumber)
			}
			return nil
		default:
			return fmt.Errorf("indirect /K object %d has unsupported type %T", objectNumber, value)
		}
	}
	walkValue = func(value types.Object, depth, owner int) error {
		valuesSeen++
		if valuesSeen > maxStructSnapshotValues {
			return fmt.Errorf("structure tree exceeds %d visited-value limit", maxStructSnapshotValues)
		}
		if depth > maxStructSnapshotDepth {
			return fmt.Errorf("structure tree exceeds %d-level depth limit below object %d", maxStructSnapshotDepth, owner)
		}
		switch kid := value.(type) {
		case types.IndirectRef:
			return walkRef(kid, depth, false)
		case types.Array:
			if len(kid) > maxStructSnapshotArray {
				return fmt.Errorf("structure object %d /K array has %d entries; limit is %d", owner, len(kid), maxStructSnapshotArray)
			}
			for _, child := range kid {
				if child == nil {
					return fmt.Errorf("structure object %d /K array contains a nil child", owner)
				}
				if err := walkValue(child, depth+1, owner); err != nil {
					return err
				}
			}
			return nil
		case types.Integer:
			if kid < 0 {
				return fmt.Errorf("structure object %d /K contains negative MCID %d", owner, kid)
			}
			return nil
		case types.Dict:
			if kid["Type"] == types.Name("MCR") || kid["Type"] == types.Name("OBJR") {
				return nil
			}
			return fmt.Errorf("structure object %d /K contains an unsupported direct dictionary", owner)
		default:
			return fmt.Errorf("structure object %d /K contains unsupported child type %T", owner, value)
		}
	}
	if err := walkRef(rootRef, 0, true); err != nil {
		return nil, err
	}
	return out, nil
}

func requireSnapshotRef(t *testing.T, value types.Object, where string) types.IndirectRef {
	t.Helper()
	ref, ok := value.(types.IndirectRef)
	if !ok {
		t.Fatalf("%s = %T, want indirect structure reference", where, value)
	}
	return ref
}

func TestSnapshotStructTreeVisitsSingularAndArrayKids(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, pdfCtx := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)

			rootRef := requireSnapshotRef(t, pdfCtx.RootDict["StructTreeRoot"], "catalog StructTreeRoot")
			root, err := pdfCtx.DereferenceDict(rootRef)
			if err != nil {
				t.Fatalf("resolve StructTreeRoot: %v", err)
			}
			docRef := requireSnapshotRef(t, root["K"], "StructTreeRoot K")
			doc, err := pdfCtx.DereferenceDict(docRef)
			if err != nil {
				t.Fatalf("resolve Document: %v", err)
			}
			tableRef := requireSnapshotRef(t, doc["K"], "Document K")
			table, err := pdfCtx.DereferenceDict(tableRef)
			if err != nil {
				t.Fatalf("resolve Table: %v", err)
			}
			rowRefs, ok := table["K"].(types.Array)
			if !ok || len(rowRefs) != 3 {
				t.Fatalf("Table K = %T with length %d, want three-row array", table["K"], len(rowRefs))
			}

			wantRefs := []types.IndirectRef{rootRef, docRef, tableRef}
			for rowIndex, rowValue := range rowRefs {
				rowRef := requireSnapshotRef(t, rowValue, fmt.Sprintf("Table K[%d]", rowIndex))
				wantRefs = append(wantRefs, rowRef)
				row, err := pdfCtx.DereferenceDict(rowRef)
				if err != nil {
					t.Fatalf("resolve row %d: %v", rowIndex, err)
				}
				cellRefs, ok := row["K"].(types.Array)
				if !ok || len(cellRefs) != 3 {
					t.Fatalf("row %d K = %T with length %d, want three-cell array", rowIndex, row["K"], len(cellRefs))
				}
				for cellIndex, cellValue := range cellRefs {
					wantRefs = append(wantRefs, requireSnapshotRef(t, cellValue, fmt.Sprintf("row %d K[%d]", rowIndex, cellIndex)))
				}
			}

			got, err := snapshotStructTree(pdfCtx)
			if err != nil {
				t.Fatalf("snapshot structure tree: %v", err)
			}
			for _, ref := range wantRefs {
				if _, ok := got[int(ref.ObjectNumber)]; !ok {
					t.Errorf("structure object %d was not visited", ref.ObjectNumber)
				}
			}
		})
	}
}

func TestSnapshotStructTreeDetectsNonTargetDocumentTableAndCellChanges(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, pdfCtx := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)

			root, err := pdfCtx.DereferenceDict(pdfCtx.RootDict["StructTreeRoot"])
			if err != nil {
				t.Fatalf("resolve StructTreeRoot: %v", err)
			}
			docRef := requireSnapshotRef(t, root["K"], "StructTreeRoot K")
			doc, err := pdfCtx.DereferenceDict(docRef)
			if err != nil {
				t.Fatalf("resolve Document: %v", err)
			}
			tableRef := requireSnapshotRef(t, doc["K"], "Document K")
			table, err := pdfCtx.DereferenceDict(tableRef)
			if err != nil {
				t.Fatalf("resolve Table: %v", err)
			}
			rowRefs, ok := table["K"].(types.Array)
			if !ok || len(rowRefs) == 0 {
				t.Fatalf("Table K = %T, want non-empty row array", table["K"])
			}
			firstRowRef := requireSnapshotRef(t, rowRefs[0], "Table K[0]")
			firstRow, err := pdfCtx.DereferenceDict(firstRowRef)
			if err != nil {
				t.Fatalf("resolve first row: %v", err)
			}
			cellRefs, ok := firstRow["K"].(types.Array)
			if !ok || len(cellRefs) == 0 {
				t.Fatalf("first row K = %T, want non-empty cell array", firstRow["K"])
			}
			cellRef := requireSnapshotRef(t, cellRefs[0], "first row K[0]")

			mutations := []struct {
				name string
				ref  types.IndirectRef
			}{
				{name: "Document", ref: docRef},
				{name: "Table", ref: tableRef},
				{name: "non-target cell", ref: cellRef},
			}
			for _, mutation := range mutations {
				dict, err := pdfCtx.DereferenceDict(mutation.ref)
				if err != nil {
					t.Fatalf("resolve %s for mutation: %v", mutation.name, err)
				}
				before, err := snapshotStructTree(pdfCtx)
				if err != nil {
					t.Fatalf("snapshot before %s mutation: %v", mutation.name, err)
				}
				dict["SnapshotMutation"] = types.StringLiteral(mutation.name)
				after, err := snapshotStructTree(pdfCtx)
				if err != nil {
					t.Fatalf("snapshot after %s mutation: %v", mutation.name, err)
				}
				objectNumber := int(mutation.ref.ObjectNumber)
				if before[objectNumber] == after[objectNumber] {
					t.Errorf("snapshot did not detect mutation to %s structure object %d", mutation.name, objectNumber)
				}
			}
		})
	}
}

func TestSnapshotStructTreeRejectsMalformedCyclicAndOversizedKids(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, control := range []string{"malformed", "cycle", "depth", "array-limit"} {
				t.Run(control, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					_, pdfCtx := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
					rootRef := requireSnapshotRef(t, pdfCtx.RootDict["StructTreeRoot"], "catalog StructTreeRoot")
					root, err := pdfCtx.DereferenceDict(rootRef)
					if err != nil {
						t.Fatalf("resolve StructTreeRoot: %v", err)
					}
					switch control {
					case "malformed":
						root["K"] = types.Name("unexpected-child")
					case "cycle":
						docRef := requireSnapshotRef(t, root["K"], "StructTreeRoot K")
						doc, err := pdfCtx.DereferenceDict(docRef)
						if err != nil {
							t.Fatalf("resolve Document: %v", err)
						}
						doc["K"] = rootRef
					case "depth":
						var kid types.Object = types.Integer(0)
						for i := 0; i <= maxStructSnapshotDepth; i++ {
							kid = types.Array{kid}
						}
						root["K"] = kid
					case "array-limit":
						kids := make(types.Array, maxStructSnapshotArray+1)
						for i := range kids {
							kids[i] = types.Integer(i)
						}
						root["K"] = kids
					}
					if _, err := snapshotStructTree(pdfCtx); err == nil {
						t.Fatalf("snapshot accepted %s structure tree", control)
					}
				})
			}
		})
	}
}

func assertV2FreshCandidateTextAndRender(t *testing.T, ctx context.Context, tc struct {
	name             string
	relPath          string
	wantSHA256       string
	damagedHeaderID  string
	repairedHeaderID string
	candidateRelPath string
	candidateSHA256  string
}, candidatePath string, candidateBytes []byte) {
	t.Helper()
	stagedBytes, err := readV2FileBoundedWithContext(ctx, candidatePath, 256<<10)
	if err != nil {
		t.Fatalf("read freshly generated %s candidate: %v", tc.name, err)
	}
	if !bytes.Equal(stagedBytes, candidateBytes) {
		t.Fatalf("fresh %s candidate path bytes differ from the bytes freshly parsed", tc.name)
	}

	damagedText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", tc.relPath, "-")
	if err != nil {
		t.Fatalf("extract damaged %s text: %v", tc.name, err)
	}
	candidateText, err := runV2NoopCommand(ctx, v2CandidateTextLimit, "pdftotext", "-layout", candidatePath, "-")
	if err != nil {
		t.Fatalf("extract fresh %s candidate text: %v", tc.name, err)
	}
	if len(damagedText) == 0 || len(candidateText) == 0 || !bytes.Equal(damagedText, candidateText) {
		t.Fatalf("fresh %s candidate text differs from damaged input or is empty (%d vs %d bytes)", tc.name, len(candidateText), len(damagedText))
	}
	textDigest := sha256.Sum256(candidateText)
	t.Logf("%s fresh candidate text SHA-256: %x (%d bytes)", tc.name, textDigest, len(candidateText))

	staging := t.TempDir()
	damagedPNG := renderV2PDFToPNG(t, ctx, tc.relPath, filepath.Join(staging, tc.name+"-damaged"))
	candidatePNG := renderV2PDFToPNG(t, ctx, candidatePath, filepath.Join(staging, tc.name+"-fresh-candidate"))
	if len(damagedPNG) == 0 || len(candidatePNG) == 0 || !bytes.Equal(damagedPNG, candidatePNG) {
		damagedDigest := sha256.Sum256(damagedPNG)
		candidateDigest := sha256.Sum256(candidatePNG)
		t.Fatalf("fresh %s candidate render differs from damaged input or is empty (damaged=%x, candidate=%x)", tc.name, damagedDigest, candidateDigest)
	}
	renderDigest := sha256.Sum256(candidatePNG)
	t.Logf("%s fresh candidate render SHA-256: %x (%d bytes)", tc.name, renderDigest, len(candidatePNG))
}

// assertNamedTreeIdentical compares a StructTreeRoot name tree (IDTree,
// ParentTree) between source and candidate contexts.
func assertNamedTreeIdentical(t *testing.T, src, candidate *model.Context, name string) {
	t.Helper()
	srcRoot, err := src.DereferenceDict(src.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatalf("source StructTreeRoot: %v", err)
	}
	candRoot, err := candidate.DereferenceDict(candidate.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatalf("candidate StructTreeRoot: %v", err)
	}
	srcTree, err := src.DereferenceDict(srcRoot[name])
	if err != nil {
		t.Fatalf("source /%s: %v", name, err)
	}
	candTree, err := candidate.DereferenceDict(candRoot[name])
	if err != nil {
		t.Fatalf("candidate /%s: %v", name, err)
	}
	if srcTree.String() != candTree.String() {
		t.Fatalf("/%s changed:\nbefore: %s\nafter:  %s", name, srcTree.String(), candTree.String())
	}
}

func compareV2RepairedTargetCell(before, after types.Dict, repairedHeaderID string) error {
	expected := before.Clone().(types.Dict)
	attrs, ok := expected["A"].(types.Dict)
	if !ok {
		return fmt.Errorf("original cell A is %T, want dictionary", expected["A"])
	}
	attrs = attrs.Clone().(types.Dict)
	if headers, ok := attrs["Headers"].(types.Array); !ok || len(headers) != 1 {
		return fmt.Errorf("original cell Headers = %v, want one-element array", attrs["Headers"])
	}
	attrs["Headers"] = types.Array{types.StringLiteral(repairedHeaderID)}
	expected["A"] = attrs
	if !reflect.DeepEqual(expected, after) {
		return fmt.Errorf("repaired target cell differs beyond the expected /A /Headers replacement: expected %s; got %s", expected, after)
	}
	return nil
}

func TestV2RepairedTargetCellComparisonRejectsOtherMutations(t *testing.T) {
	tc := v2RepairAttemptCases[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, pdfCtx := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
	original := locateV2Row2AmountCell(t, pdfCtx).cell.Clone().(types.Dict)
	expected := original.Clone().(types.Dict)
	attrs := expected["A"].(types.Dict).Clone().(types.Dict)
	attrs["Headers"] = types.Array{types.StringLiteral(tc.repairedHeaderID)}
	expected["A"] = attrs

	mutations := []struct {
		name   string
		mutate func(types.Dict)
	}{
		{name: "/P", mutate: func(cell types.Dict) { cell["P"] = types.StringLiteral("unexpected") }},
		{name: "/Pg", mutate: func(cell types.Dict) { cell["Pg"] = types.StringLiteral("unexpected") }},
		{name: "/K", mutate: func(cell types.Dict) { cell["K"] = types.StringLiteral("unexpected") }},
		{name: "/T", mutate: func(cell types.Dict) { cell["T"] = types.StringLiteral("unexpected") }},
		{name: "extra attribute", mutate: func(cell types.Dict) {
			mutatedAttrs := cell["A"].(types.Dict).Clone().(types.Dict)
			mutatedAttrs["ReviewMutation"] = types.Boolean(true)
			cell["A"] = mutatedAttrs
		}},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := expected.Clone().(types.Dict)
			mutation.mutate(mutated)
			if err := compareV2RepairedTargetCell(original, mutated, tc.repairedHeaderID); err == nil {
				t.Fatalf("comparison accepted unrelated %s mutation", mutation.name)
			}
		})
	}
	if err := compareV2RepairedTargetCell(original, expected, tc.repairedHeaderID); err != nil {
		t.Fatalf("comparison rejected the expected Headers-only change: %v", err)
	}
}

type pageResourcesSnapshot struct {
	resourceDictionary string
	fonts              fontResourcesSnapshot
}

func snapshotPageResources(goCtx context.Context, document *model.Context) (pageResourcesSnapshot, error) {
	page, _, _, err := document.PageDict(goCtx, 1, false)
	if err != nil {
		return pageResourcesSnapshot{}, fmt.Errorf("read page dict: %w", err)
	}
	resources, err := document.DereferenceDict(page["Resources"])
	if err != nil {
		return pageResourcesSnapshot{}, fmt.Errorf("dereference page resources: %w", err)
	}
	fonts, err := snapshotPageFonts(goCtx, document)
	if err != nil {
		return pageResourcesSnapshot{}, fmt.Errorf("snapshot page fonts: %w", err)
	}
	return pageResourcesSnapshot{resourceDictionary: resources.String(), fonts: fonts}, nil
}

func assertPageResourcesSnapshotIdentical(goCtx context.Context, expected pageResourcesSnapshot, candidate *model.Context) error {
	candidatePage, _, _, err := candidate.PageDict(goCtx, 1, false)
	if err != nil {
		return fmt.Errorf("candidate page dict: %w", err)
	}
	candidateResources, err := candidate.DereferenceDict(candidatePage["Resources"])
	if err != nil {
		return fmt.Errorf("candidate /Resources: %w", err)
	}
	if expected.resourceDictionary != candidateResources.String() {
		return fmt.Errorf("/Resources changed: before %s; after %s", expected.resourceDictionary, candidateResources.String())
	}
	candidateFonts, err := snapshotPageFonts(goCtx, candidate)
	if err != nil {
		return fmt.Errorf("snapshot candidate page fonts: %w", err)
	}
	if !reflect.DeepEqual(expected.fonts, candidateFonts) {
		return fmt.Errorf("dereferenced page font or descriptor dictionaries changed: before %#v; after %#v", expected.fonts, candidateFonts)
	}
	return nil
}

// assertPageResourcesIdentical compares the page resources in two contexts.
// The repair evidence test uses an earlier immutable snapshot instead, so
// changes already made to the source context before serialization are caught.
func assertPageResourcesIdentical(goCtx context.Context, src, candidate *model.Context) error {
	expected, err := snapshotPageResources(goCtx, src)
	if err != nil {
		return fmt.Errorf("snapshot source page resources: %w", err)
	}
	return assertPageResourcesSnapshotIdentical(goCtx, expected, candidate)
}

func TestAssertPageResourcesIdenticalDetectsReferencedDescriptorMutation(t *testing.T) {
	tc := v2RepairAttemptCases[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, source := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
	_, candidate := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
	if err := assertPageResourcesIdentical(ctx, source, candidate); err != nil {
		t.Fatalf("identical resources rejected: %v", err)
	}
	mutateF1FontDescriptor(t, candidate)
	if err := assertPageResourcesIdentical(ctx, source, candidate); err == nil {
		t.Fatal("resource comparison accepted changed referenced font descriptor with unchanged /F1 reference")
	}
}

func TestPageResourceComparisonDetectsOriginalContextMutationBeforeWrite(t *testing.T) {
	tc := v2RepairAttemptCases[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, source := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
	expected, err := snapshotPageResources(ctx, source)
	if err != nil {
		t.Fatalf("snapshot source page resources before mutation: %v", err)
	}
	mutateF1FontDescriptor(t, source)

	_, candidateBytes, err := writeV2RepairCandidate(ctx, source, t.TempDir(), "mutated-source.pdf", 256<<10)
	if err != nil {
		t.Fatalf("write candidate from mutated source context: %v", err)
	}
	candidate, err := api.ReadContext(ctx, bytes.NewReader(candidateBytes), boundedStatelessConfig(t))
	if err != nil {
		t.Fatalf("read candidate from mutated source context: %v", err)
	}
	if err := api.ValidateContext(ctx, candidate); err != nil {
		t.Fatalf("validate candidate from mutated source context: %v", err)
	}
	if err := assertPageResourcesSnapshotIdentical(ctx, expected, candidate); err == nil {
		t.Fatal("resource comparison accepted a descriptor mutation made to the original context before serialization")
	}
}

func mutateF1FontDescriptor(t *testing.T, ctx *model.Context) {
	t.Helper()
	fonts, err := pageFontResources(ctx)
	if err != nil {
		t.Fatalf("read font resources: %v", err)
	}
	font, err := ctx.DereferenceDict(fonts["F1"])
	if err != nil {
		t.Fatalf("dereference /F1: %v", err)
	}
	descriptorRef, ok := font["FontDescriptor"].(types.IndirectRef)
	if !ok {
		t.Fatalf("/F1 FontDescriptor = %T, want indirect reference", font["FontDescriptor"])
	}
	descriptor, err := ctx.DereferenceDict(descriptorRef)
	if err != nil {
		t.Fatalf("dereference /F1 FontDescriptor: %v", err)
	}
	descriptor["ReviewMutation"] = types.Boolean(true)
}

func readV2DamagedFixture(ctx context.Context, t *testing.T, relPath, wantSHA256 string) ([]byte, *model.Context) {
	t.Helper()
	data, err := readV2FileBoundedWithContext(ctx, relPath, v2PDFInputLimit)
	if err != nil {
		t.Fatalf("read damaged fixture %s: %v", relPath, err)
	}
	digest := sha256.Sum256(data)
	if got := hex.EncodeToString(digest[:]); got != wantSHA256 {
		t.Fatalf("damaged fixture %s SHA-256 = %s, want %s", relPath, got, wantSHA256)
	}
	// The caller's case context carries the per-case deadline (bound-work rule).
	pdfCtx, err := api.ReadContext(ctx, bytes.NewReader(data), boundedStatelessConfig(t))
	if err != nil {
		t.Fatalf("pdfcpu read of %s: %v", relPath, err)
	}
	if err := api.ValidateContext(ctx, pdfCtx); err != nil {
		t.Fatalf("pdfcpu validate %s: %v", relPath, err)
	}
	return data, pdfCtx
}

func TestWriteV2RepairCandidateOutputCapCleansStagingFile(t *testing.T) {
	tc := v2RepairAttemptCases[0]
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, document := readV2DamagedFixture(ctx, t, tc.relPath, tc.wantSHA256)
	directory := t.TempDir()
	if _, _, err := writeV2RepairCandidate(ctx, document, directory, "over-limit.pdf", 64); err == nil || !strings.Contains(err.Error(), "staged output exceeds 64-byte experiment limit") {
		t.Fatalf("bounded repair write error = %v, want the 64-byte ceiling error", err)
	}
	if entries, err := os.ReadDir(directory); err != nil || len(entries) != 0 {
		t.Fatalf("failed bounded repair write left staged bytes: entries=%v err=%v", entries, err)
	}
}

func TestWriteV2RepairCandidateHonorsExpiredDeadline(t *testing.T) {
	tc := v2RepairAttemptCases[0]
	readCtx, cancelRead := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelRead()
	_, document := readV2DamagedFixture(readCtx, t, tc.relPath, tc.wantSHA256)

	deadlineCtx, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancelDeadline()
	directory := t.TempDir()
	if _, _, err := writeV2RepairCandidate(deadlineCtx, document, directory, "expired.pdf", 256<<10); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expired-deadline repair write error = %v, want context deadline exceeded", err)
	}
	if entries, err := os.ReadDir(directory); err != nil || len(entries) != 0 {
		t.Fatalf("expired-deadline write left staged bytes: entries=%v err=%v", entries, err)
	}
}

func TestV2Row2AmountHeaderRepair(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			// One deadline for the whole repair evidence path (read,
			// validate, write, re-read) per the repo's bound-work rule.
			caseCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			_, ctx := readV2DamagedFixture(caseCtx, t, tc.relPath, tc.wantSHA256)
			beforePageResources, err := snapshotPageResources(caseCtx, ctx)
			if err != nil {
				t.Fatalf("snapshot original page resources: %v", err)
			}

			// Characterize the damage: the row-2 amount cell carries the
			// wrong header ID and only that ID.
			target := locateV2Row2AmountCell(t, ctx)
			if len(target.headerIDs) != 1 || target.headerIDs[0] != tc.damagedHeaderID {
				t.Fatalf("damaged %s row-2 amount Headers = %q, want [%q]",
					tc.name, target.headerIDs, tc.damagedHeaderID)
			}
			beforeTargetCell := target.cell.Clone().(types.Dict)
			before, err := snapshotStructTree(ctx)
			if err != nil {
				t.Fatalf("snapshot damaged structure tree: %v", err)
			}

			repairedObj, err := repairV2Row2AmountHeaders(ctx, v2RepairSpec{
				DamagedHeaderID:  tc.damagedHeaderID,
				RepairedHeaderID: tc.repairedHeaderID,
			})
			if err != nil {
				t.Fatalf("repair %s: %v", tc.name, err)
			}
			if repairedObj != target.objectNumber {
				t.Fatalf("repaired object %d, want row-2 amount cell %d", repairedObj, target.objectNumber)
			}

			// Write the candidate to private staging; never into the fixture tree.
			// The write and the read-back are both bounded at the 256 KiB
			// experiment ceiling, so a pdfcpu regression cannot exhaust
			// disk or memory here.
			candidatePath, candidateBytes, err := writeV2RepairCandidate(caseCtx, ctx, t.TempDir(), tc.name+"-repaired-v2.pdf", 256<<10)
			if err != nil {
				t.Fatalf("write %s candidate: %v", tc.name, err)
			}
			candidateDigest := sha256.Sum256(candidateBytes)
			t.Logf("%s candidate SHA-256: %s", tc.name, hex.EncodeToString(candidateDigest[:]))
			assertV2FreshCandidateTextAndRender(t, caseCtx, tc, candidatePath, candidateBytes)

			// Re-read the candidate and verify the fix.
			candidateCtx, err := api.ReadContext(caseCtx, bytes.NewReader(candidateBytes), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("pdfcpu re-read of %s candidate: %v", tc.name, err)
			}
			if err := api.ValidateContext(caseCtx, candidateCtx); err != nil {
				t.Fatalf("pdfcpu validate %s candidate: %v", tc.name, err)
			}
			fixed := locateV2Row2AmountCell(t, candidateCtx)
			if len(fixed.headerIDs) != 1 || fixed.headerIDs[0] != tc.repairedHeaderID {
				t.Fatalf("repaired %s row-2 amount Headers = %q, want [%q]",
					tc.name, fixed.headerIDs, tc.repairedHeaderID)
			}

			// IDTree, ParentTree, and page resources (fonts) must be identical.
			assertNamedTreeIdentical(t, ctx, candidateCtx, "IDTree")
			assertNamedTreeIdentical(t, ctx, candidateCtx, "ParentTree")
			if err := assertPageResourcesSnapshotIdentical(caseCtx, beforePageResources, candidateCtx); err != nil {
				t.Fatalf("page resources changed: %v", err)
			}

			// Every other structure element must be identical to the source.
			after, err := snapshotStructTree(candidateCtx)
			if err != nil {
				t.Fatalf("snapshot repaired structure tree: %v", err)
			}
			if len(after) != len(before) {
				t.Fatalf("structure element count: source %d, candidate %d", len(before), len(after))
			}
			for objNum, beforeDict := range before {
				afterDict, ok := after[objNum]
				if !ok {
					t.Fatalf("structure element %d missing from candidate", objNum)
				}
				if objNum == repairedObj {
					if err := compareV2RepairedTargetCell(beforeTargetCell, fixed.cell, tc.repairedHeaderID); err != nil {
						t.Fatalf("repaired target cell changed beyond /A /Headers: %v", err)
					}
					continue
				}
				if afterDict != beforeDict {
					t.Fatalf("structure element %d changed:\nbefore: %s\nafter:  %s", objNum, beforeDict, afterDict)
				}
			}
		})
	}
}
