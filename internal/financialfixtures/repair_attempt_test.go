package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

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

// snapshotStructTree serializes every structure element dict reachable from
// the StructTreeRoot, keyed by object number.
func snapshotStructTree(t *testing.T, ctx *model.Context) map[int]string {
	t.Helper()
	out := map[int]string{}
	var visit func(ref types.IndirectRef)
	visit = func(ref types.IndirectRef) {
		if _, seen := out[int(ref.ObjectNumber)]; seen {
			return
		}
		dict, err := ctx.DereferenceDict(ref)
		if err != nil {
			t.Fatalf("resolve structure element %d: %v", int(ref.ObjectNumber), err)
		}
		out[int(ref.ObjectNumber)] = dict.String()
		kids, ok := dict["K"].(types.Array)
		if !ok {
			return
		}
		for _, kid := range kids {
			kref, ok := kid.(types.IndirectRef)
			if !ok {
				t.Fatalf("structure element %d child = %T, want indirect reference", int(ref.ObjectNumber), kid)
			}
			visit(kref)
		}
	}
	rootRef, ok := ctx.RootDict["StructTreeRoot"].(types.IndirectRef)
	if !ok {
		t.Fatalf("StructTreeRoot = %T, want indirect reference", ctx.RootDict["StructTreeRoot"])
	}
	visit(rootRef)
	return out
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

// assertPageResourcesIdentical compares the page resource dictionaries
// (fonts and related resources) between source and candidate contexts.
func assertPageResourcesIdentical(t *testing.T, goCtx context.Context, src, candidate *model.Context) {
	t.Helper()
	srcPage, _, _, err := src.PageDict(goCtx, 1, false)
	if err != nil {
		t.Fatalf("source page dict: %v", err)
	}
	candPage, _, _, err := candidate.PageDict(goCtx, 1, false)
	if err != nil {
		t.Fatalf("candidate page dict: %v", err)
	}
	srcRes, err := src.DereferenceDict(srcPage["Resources"])
	if err != nil {
		t.Fatalf("source /Resources: %v", err)
	}
	candRes, err := candidate.DereferenceDict(candPage["Resources"])
	if err != nil {
		t.Fatalf("candidate /Resources: %v", err)
	}
	if srcRes.String() != candRes.String() {
		t.Fatalf("/Resources changed:\nbefore: %s\nafter:  %s", srcRes.String(), candRes.String())
	}
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

func TestV2Row2AmountHeaderRepair(t *testing.T) {
	for _, tc := range v2RepairAttemptCases {
		t.Run(tc.name, func(t *testing.T) {
			// One deadline for the whole repair evidence path (read,
			// validate, write, re-read) per the repo's bound-work rule.
			caseCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			_, ctx := readV2DamagedFixture(caseCtx, t, tc.relPath, tc.wantSHA256)

			// Characterize the damage: the row-2 amount cell carries the
			// wrong header ID and only that ID.
			target := locateV2Row2AmountCell(t, ctx)
			if len(target.headerIDs) != 1 || target.headerIDs[0] != tc.damagedHeaderID {
				t.Fatalf("damaged %s row-2 amount Headers = %q, want [%q]",
					tc.name, target.headerIDs, tc.damagedHeaderID)
			}
			before := snapshotStructTree(t, ctx)

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
			_, candidateBytes, err := writeV2RepairCandidate(caseCtx, ctx, t.TempDir(), tc.name+"-repaired-v2.pdf", 256<<10)
			if err != nil {
				t.Fatalf("write %s candidate: %v", tc.name, err)
			}
			candidateDigest := sha256.Sum256(candidateBytes)
			t.Logf("%s candidate SHA-256: %s", tc.name, hex.EncodeToString(candidateDigest[:]))

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
			assertPageResourcesIdentical(t, caseCtx, ctx, candidateCtx)

			// Every other structure element must be identical to the source.
			after := snapshotStructTree(t, candidateCtx)
			if len(after) != len(before) {
				t.Fatalf("structure element count: source %d, candidate %d", len(before), len(after))
			}
			for objNum, beforeDict := range before {
				afterDict, ok := after[objNum]
				if !ok {
					t.Fatalf("structure element %d missing from candidate", objNum)
				}
				if objNum == repairedObj {
					continue
				}
				if afterDict != beforeDict {
					t.Fatalf("structure element %d changed:\nbefore: %s\nafter:  %s", objNum, beforeDict, afterDict)
				}
			}
		})
	}
}
