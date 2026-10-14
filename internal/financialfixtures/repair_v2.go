package financialfixtures

import (
	"fmt"
	"regexp"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// v2RepairSpec describes one fixture-authored row-2 amount-cell /Headers
// repair: the single header ID to replace and its replacement.
type v2RepairSpec struct {
	DamagedHeaderID  string
	RepairedHeaderID string
}

var v2RepairAmountText = regexp.MustCompile(`^\$[0-9]`)

// repairV2Row2AmountHeaders rewrites the row-2 amount TD's /A /Headers name
// array to the repaired header ID and returns the repaired structure
// element's object number. It fails closed unless the current /Headers is
// exactly the single damaged header ID: an already-repaired or differently
// damaged cell is never rewritten. Only the target /Headers array changes;
// all other structure elements, displayed strings, and resources are left
// untouched.
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
		if !ok || !v2RepairAmountText.MatchString(string(text)) {
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
	return 0, fmt.Errorf("row 2 has no amount cell matching %q", v2RepairAmountText)
}
