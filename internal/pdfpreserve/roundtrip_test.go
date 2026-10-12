// Package pdfpreserve contains a bounded characterization harness for the
// proposed PDF writer. The tests do not implement accessibility remediation.
package pdfpreserve

import (
	"bytes"
	"context"
	"crypto/sha256"
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

const syntheticTwoPageInheritedFixtureSHA256 = "4a39c3f22b9c0f682c9b2816003d9167b24db7481b122e4c8fe6a0285dd5320d"
const syntheticNestedPageTreeFixtureSHA256 = "1b5060daf940c0591deaf91ecab8d8d704c1bfc3d02fdaebfd1630d430de170d"

func TestFullWriteRoundTripPreservesSyntheticPDFFeatures(t *testing.T) {
	for _, family := range []string{"statement", "invoice", "correspondence", "report"} {
		t.Run(family, func(t *testing.T) {
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
			if err := os.WriteFile(input, syntheticPDF(family), 0o600); err != nil {
				t.Fatal(err)
			}
			baseline, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read immutable synthetic baseline: %v", err)
			}
			writerCtx, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read writer context: %v", err)
			}
			if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
				t.Fatalf("full write: %v", err)
			}
			after, err := api.ReadContextFile(context.Background(), output)
			if err != nil {
				t.Fatalf("read round-trip output: %v", err)
			}
			if err := compareRoundTrip(baseline, after); err != nil {
				t.Error(err)
			}
			assertSparseParentTree(t, after)
		})
	}
}

func TestFullWriteRoundTripPreservesTwoPageInheritedAttributes(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
	inputBytes := syntheticTwoPageInheritedPDF()
	inputDigest := sha256.Sum256(inputBytes)
	if fmt.Sprintf("%x", inputDigest) != syntheticTwoPageInheritedFixtureSHA256 {
		t.Fatalf("synthetic two-page fixture changed: SHA-256 %x, want %s", inputDigest, syntheticTwoPageInheritedFixtureSHA256)
	}
	if err := os.WriteFile(input, inputBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	baseline, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read immutable synthetic baseline: %v", err)
	}
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read writer context: %v", err)
	}
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
		t.Fatalf("full write: %v", err)
	}
	after, err := api.ReadContextFile(context.Background(), output)
	if err != nil {
		t.Fatalf("read round-trip output: %v", err)
	}
	if err := compareTwoPageInheritedFixture(baseline, after); err != nil {
		t.Fatal(err)
	}
	t.Logf("synthetic source SHA-256: %x", inputDigest)
	t.Logf("source info dictionary: %#v", baseline.Table[13].Object)
	t.Logf("round-trip info dictionary: %#v", after.Table[13].Object)
	if got := sha256.Sum256(mustReadFile(t, input)); got != inputDigest {
		t.Fatalf("full write changed immutable source hash: got %x, want %x", got, inputDigest)
	}
}

func TestTwoPageInheritedRoundTripRejectsInheritedGeometryMutation(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "mutated.pdf")
	inputBytes := syntheticTwoPageInheritedPDF()
	if err := os.WriteFile(input, inputBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	baseline, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read immutable synthetic baseline: %v", err)
	}
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read writer context: %v", err)
	}
	pageTree, ok := writerCtx.Table[2].Object.(types.Dict)
	if !ok {
		t.Fatalf("page-tree root is %T", writerCtx.Table[2].Object)
	}
	pageTree["MediaBox"] = types.Array{types.Integer(0), types.Integer(0), types.Integer(612), types.Integer(600)}
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
		t.Fatalf("write intentionally mutated synthetic PDF: %v", err)
	}
	mutated, err := api.ReadContextFile(context.Background(), output)
	if err != nil {
		t.Fatalf("read intentionally mutated PDF: %v", err)
	}
	if err := compareTwoPageInheritedFixture(baseline, mutated); err == nil || !strings.Contains(err.Error(), "effective media box changed") {
		t.Fatalf("two-page fidelity assertions did not reject inherited geometry mutation: %v", err)
	}
}

func TestTwoPageInheritedRoundTripRejectsReferencedFontMutation(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "mutated-font.pdf")
	if err := os.WriteFile(input, syntheticTwoPageInheritedPDF(), 0o600); err != nil {
		t.Fatal(err)
	}
	baseline, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read immutable synthetic baseline: %v", err)
	}
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read writer context: %v", err)
	}
	font, ok := writerCtx.Table[6].Object.(types.Dict)
	if !ok {
		t.Fatalf("fixture font object is %T, want a dictionary", writerCtx.Table[6].Object)
	}
	font["BaseFont"] = types.Name("Courier")
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
		t.Fatalf("write intentionally mutated synthetic PDF: %v", err)
	}
	mutated, err := api.ReadContextFile(context.Background(), output)
	if err != nil {
		t.Fatalf("read intentionally mutated PDF: %v", err)
	}
	if err := compareTwoPageInheritedFixture(baseline, mutated); err == nil || !strings.Contains(err.Error(), "font changed") {
		t.Fatalf("two-page fidelity assertions did not reject referenced font mutation: %v", err)
	}
}

func TestFullWriteRoundTripPreservesNestedPageTreeInheritance(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
	inputBytes := syntheticNestedPageTreePDF()
	inputDigest := sha256.Sum256(inputBytes)
	if got := fmt.Sprintf("%x", inputDigest); got != syntheticNestedPageTreeFixtureSHA256 {
		t.Fatalf("synthetic nested-page-tree fixture changed: SHA-256 %s, want %s", got, syntheticNestedPageTreeFixtureSHA256)
	}
	if err := os.WriteFile(input, inputBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	baseline, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read immutable synthetic baseline: %v", err)
	}
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read writer context: %v", err)
	}
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
		t.Fatalf("full write: %v", err)
	}
	after, err := api.ReadContextFile(context.Background(), output)
	if err != nil {
		t.Fatalf("read round-trip output: %v", err)
	}
	if err := compareNestedPageTreeFixture(baseline, after); err != nil {
		t.Fatal(err)
	}
	t.Logf("synthetic nested-page-tree source SHA-256: %x", inputDigest)
	if got := sha256.Sum256(mustReadFile(t, input)); got != inputDigest {
		t.Fatalf("full write changed immutable source hash: got %x, want %x", got, inputDigest)
	}
}

func TestNestedPageTreeRoundTripRejectsInheritedCropBoxAndRotateMutations(t *testing.T) {
	for _, mutation := range []struct {
		name   string
		change func(types.Dict)
		want   string
	}{
		{
			name: "crop-box",
			change: func(pageTree types.Dict) {
				pageTree["CropBox"] = types.Array{types.Integer(30), types.Integer(24), types.Integer(594), types.Integer(768)}
			},
			want: "effective crop box changed",
		},
		{
			name: "rotate",
			change: func(pageTree types.Dict) {
				pageTree["Rotate"] = types.Integer(180)
			},
			want: "effective rotation changed",
		},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "mutated.pdf")
			if err := os.WriteFile(input, syntheticNestedPageTreePDF(), 0o600); err != nil {
				t.Fatal(err)
			}
			baseline, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read immutable synthetic baseline: %v", err)
			}
			writerCtx, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read writer context: %v", err)
			}
			pageTree, ok := writerCtx.Table[3].Object.(types.Dict)
			if !ok {
				t.Fatalf("nested page-tree node is %T, want a dictionary", writerCtx.Table[3].Object)
			}
			mutation.change(pageTree)
			if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
				t.Fatalf("write intentionally mutated synthetic PDF: %v", err)
			}
			mutated, err := api.ReadContextFile(context.Background(), output)
			if err != nil {
				t.Fatalf("read intentionally mutated PDF: %v", err)
			}
			if err := compareNestedPageTreeFixture(baseline, mutated); err == nil || !strings.Contains(err.Error(), mutation.want) {
				t.Fatalf("nested page-tree fidelity assertions did not reject inherited %s mutation: %v", mutation.name, err)
			}
		})
	}
}

func compareNestedPageTreeFixture(baseline, after *model.Context) error {
	if baseline.PageCount != 2 || after.PageCount != baseline.PageCount {
		return fmt.Errorf("page count changed: %d -> %d; want two pages", baseline.PageCount, after.PageCount)
	}
	nestedPageTree, ok := baseline.Table[3].Object.(types.Dict)
	if !ok || nestedPageTree["Type"] != types.Name("Pages") || nestedPageTree["Parent"] != (types.IndirectRef{ObjectNumber: 2}) {
		return fmt.Errorf("fixture object 3 is not a nested /Pages node beneath object 2")
	}
	wantResources, ok := nestedPageTree["Resources"].(types.Dict)
	if !ok || len(wantResources) != 0 {
		return fmt.Errorf("nested /Pages node must declare an explicit empty /Resources dictionary")
	}
	wantMediaBox := &types.Rectangle{LL: types.NewPoint(0, 0), UR: types.NewPoint(612, 792)}
	wantCropBox := &types.Rectangle{LL: types.NewPoint(18, 24), UR: types.NewPoint(594, 768)}
	for pageNr := 1; pageNr <= baseline.PageCount; pageNr++ {
		beforePage, beforeRef, beforeAttrs, err := baseline.PageDict(context.Background(), pageNr, false)
		if err != nil {
			return fmt.Errorf("read baseline page %d: %w", pageNr, err)
		}
		afterPage, _, afterAttrs, err := after.PageDict(context.Background(), pageNr, false)
		if err != nil {
			return fmt.Errorf("read round-trip page %d: %w", pageNr, err)
		}
		if beforeRef == nil || beforePage["Parent"] != (types.IndirectRef{ObjectNumber: 3}) {
			return fmt.Errorf("baseline page %d does not descend from nested page-tree node 3", pageNr)
		}
		beforeResources, beforeResourceOwner, err := nearestPageResources(baseline, beforePage)
		if err != nil {
			return fmt.Errorf("read baseline page %d resources: %w", pageNr, err)
		}
		if beforeResourceOwner != (types.IndirectRef{ObjectNumber: 3}) || len(beforeResources) != 0 {
			return fmt.Errorf("baseline page %d does not inherit an empty /Resources dictionary from nested node 3", pageNr)
		}
		afterResources, _, err := nearestPageResources(after, afterPage)
		if err != nil {
			return fmt.Errorf("read round-trip page %d resources: %w", pageNr, err)
		}
		if len(afterResources) != 0 {
			return fmt.Errorf("page %d effective resources changed: want empty dictionary, got %v", pageNr, afterResources)
		}
		for _, key := range []string{"MediaBox", "CropBox", "Rotate", "Resources"} {
			if _, found := beforePage[key]; found {
				return fmt.Errorf("baseline page %d unexpectedly declares inherited %s", pageNr, key)
			}
		}
		if beforeAttrs.MediaBox == nil || *beforeAttrs.MediaBox != *wantMediaBox {
			return fmt.Errorf("page %d effective media box = %v, want [0 0 612 792]", pageNr, beforeAttrs.MediaBox)
		}
		if beforeAttrs.CropBox == nil || *beforeAttrs.CropBox != *wantCropBox {
			return fmt.Errorf("page %d effective crop box = %v, want [18 24 594 768]", pageNr, beforeAttrs.CropBox)
		}
		if beforeAttrs.Rotate != 90 {
			return fmt.Errorf("page %d effective rotation = %d, want 90", pageNr, beforeAttrs.Rotate)
		}
		if afterAttrs.MediaBox == nil || *afterAttrs.MediaBox != *beforeAttrs.MediaBox {
			return fmt.Errorf("page %d effective media box changed: %v -> %v", pageNr, beforeAttrs.MediaBox, afterAttrs.MediaBox)
		}
		if afterAttrs.CropBox == nil || *afterAttrs.CropBox != *beforeAttrs.CropBox {
			return fmt.Errorf("page %d effective crop box changed: %v -> %v", pageNr, beforeAttrs.CropBox, afterAttrs.CropBox)
		}
		if afterAttrs.Rotate != beforeAttrs.Rotate {
			return fmt.Errorf("page %d effective rotation changed: %d -> %d", pageNr, beforeAttrs.Rotate, afterAttrs.Rotate)
		}
		beforeContent, err := baseline.PageContent(beforePage, pageNr)
		if err != nil {
			return fmt.Errorf("read baseline page %d content: %w", pageNr, err)
		}
		afterContent, err := after.PageContent(afterPage, pageNr)
		if err != nil {
			return fmt.Errorf("read round-trip page %d content: %w", pageNr, err)
		}
		if !bytes.Equal(beforeContent, afterContent) {
			return fmt.Errorf("page %d content changed", pageNr)
		}
		wantContent := []byte([]string{"q 1 0 0 rg 24 24 40 20 re f Q\n", "q 0 0 1 rg 80 80 20 40 re f Q\n"}[pageNr-1])
		if !bytes.Equal(beforeContent, wantContent) {
			return fmt.Errorf("page %d baseline content = %q, want %q", pageNr, beforeContent, wantContent)
		}
	}
	return nil
}

func nearestPageResources(ctx *model.Context, page types.Dict) (types.Dict, types.IndirectRef, error) {
	node := page
	var nodeRef types.IndirectRef
	for {
		if obj, found := node.Find("Resources"); found {
			if resources, ok := obj.(types.Dict); ok {
				return resources, nodeRef, nil
			}
			return nil, types.IndirectRef{}, fmt.Errorf("/Resources is %T, want a dictionary", obj)
		}
		parent, ok := node["Parent"].(types.IndirectRef)
		if !ok {
			return nil, types.IndirectRef{}, fmt.Errorf("page tree has no inherited /Resources dictionary")
		}
		entry, exists := ctx.Table[parent.ObjectNumber.Value()]
		if !exists {
			return nil, types.IndirectRef{}, fmt.Errorf("page-tree parent %s is absent", parent)
		}
		obj, ok := entry.Object.(types.Dict)
		if !ok {
			return nil, types.IndirectRef{}, fmt.Errorf("page-tree parent %s is %T, want a dictionary", parent, entry.Object)
		}
		node = obj
		nodeRef = parent
	}
}

func syntheticNestedPageTreePDF() []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 2 /MediaBox [0 0 612 792] /Rotate 270 >>",
		"<< /Type /Pages /Parent 2 0 R /Kids [4 0 R 5 0 R] /CropBox [18 24 594 768] /Rotate 90 /Resources <<>> >>",
		"<< /Type /Page /Parent 3 0 R /Contents 6 0 R >>",
		"<< /Type /Page /Parent 3 0 R /Contents 7 0 R >>",
		streamObject("q 1 0 0 rg 24 24 40 20 re f Q\n"),
		streamObject("q 0 0 1 rg 80 80 20 40 re f Q\n"),
		"<< /Producer (DAP nested page-tree fixture) >>",
	}
	return assembleSyntheticPDF(objects)
}

func compareTwoPageInheritedFixture(baseline, after *model.Context) error {
	if baseline.PageCount != 2 || after.PageCount != baseline.PageCount {
		return fmt.Errorf("page count changed: %d -> %d; want two pages", baseline.PageCount, after.PageCount)
	}
	for pageNr := 1; pageNr <= baseline.PageCount; pageNr++ {
		beforePage, _, beforeAttrs, err := baseline.PageDict(context.Background(), pageNr, false)
		if err != nil {
			return fmt.Errorf("read baseline page %d: %w", pageNr, err)
		}
		afterPage, _, afterAttrs, err := after.PageDict(context.Background(), pageNr, false)
		if err != nil {
			return fmt.Errorf("read round-trip page %d: %w", pageNr, err)
		}
		if _, ok := beforePage["MediaBox"]; ok {
			return fmt.Errorf("fixture page %d unexpectedly declares its own MediaBox", pageNr)
		}
		if _, ok := beforePage["Resources"]; ok {
			return fmt.Errorf("fixture page %d unexpectedly declares its own Resources", pageNr)
		}
		if beforeAttrs.MediaBox == nil || *beforeAttrs.MediaBox != (types.Rectangle{LL: types.NewPoint(0, 0), UR: types.NewPoint(612, 792)}) {
			return fmt.Errorf("page %d effective media box = %v, want inherited [0 0 612 792]", pageNr, beforeAttrs.MediaBox)
		}
		fontResources, ok := beforeAttrs.Resources["Font"].(types.Dict)
		if !ok {
			return fmt.Errorf("page %d inherited Font resources are %T, want a dictionary", pageNr, beforeAttrs.Resources["Font"])
		}
		fontRef, ok := fontResources["F1"].(types.IndirectRef)
		if !ok {
			return fmt.Errorf("page %d inherited /F1 resource is %T, want indirect font reference", pageNr, fontResources["F1"])
		}
		font, err := baseline.DereferenceDict(fontRef)
		if err != nil || font["BaseFont"] != types.Name("Helvetica") {
			return fmt.Errorf("page %d inherited /F1 font = %v (err %v), want /Helvetica", pageNr, font, err)
		}
		afterFontResources, ok := afterAttrs.Resources["Font"].(types.Dict)
		if !ok {
			return fmt.Errorf("page %d round-trip Font resources are %T, want a dictionary", pageNr, afterAttrs.Resources["Font"])
		}
		afterFontRef, ok := afterFontResources["F1"].(types.IndirectRef)
		if !ok {
			return fmt.Errorf("page %d round-trip /F1 resource is %T, want indirect font reference", pageNr, afterFontResources["F1"])
		}
		afterFont, err := after.DereferenceDict(afterFontRef)
		if err != nil || afterFont["BaseFont"] != types.Name("Helvetica") {
			return fmt.Errorf("page %d round-trip inherited /F1 font changed: %v (err %v), want /Helvetica", pageNr, afterFont, err)
		}
		if !reflect.DeepEqual(font, afterFont) {
			return fmt.Errorf("page %d round-trip inherited /F1 font changed: %v -> %v", pageNr, font, afterFont)
		}
		if !reflect.DeepEqual(beforeAttrs.MediaBox, afterAttrs.MediaBox) {
			return fmt.Errorf("page %d effective media box changed: %v -> %v", pageNr, beforeAttrs.MediaBox, afterAttrs.MediaBox)
		}
		if !reflect.DeepEqual(beforeAttrs.Resources, afterAttrs.Resources) {
			return fmt.Errorf("page %d effective resources changed", pageNr)
		}
		beforeContent, err := baseline.PageContent(beforePage, pageNr)
		if err != nil {
			return fmt.Errorf("read baseline page %d content: %w", pageNr, err)
		}
		afterContent, err := after.PageContent(afterPage, pageNr)
		if err != nil {
			return fmt.Errorf("read round-trip page %d content: %w", pageNr, err)
		}
		if !bytes.Equal(beforeContent, afterContent) {
			return fmt.Errorf("page %d content changed", pageNr)
		}
		wantText := []byte(fmt.Sprintf("Synthetic page %s", []string{"one", "two"}[pageNr-1]))
		if !bytes.Contains(beforeContent, wantText) {
			return fmt.Errorf("page %d baseline content lacks expected marker %q", pageNr, wantText)
		}
	}
	if err := assertTwoPageTagRelationships(baseline); err != nil {
		return fmt.Errorf("baseline tagged-page relationships: %w", err)
	}
	if err := assertTwoPageTagRelationships(after); err != nil {
		return fmt.Errorf("round-trip tagged-page relationships: %w", err)
	}
	return nil
}

func assertTwoPageTagRelationships(ctx *model.Context) error {
	pageRefs := make([]types.IndirectRef, 2)
	for pageNr := range pageRefs {
		pageDict, pageRef, _, err := ctx.PageDict(context.Background(), pageNr+1, false)
		if err != nil {
			return fmt.Errorf("resolve page %d: %w", pageNr+1, err)
		}
		pageRefs[pageNr] = *pageRef
		structParents, err := ctx.DereferenceInteger(pageDict["StructParents"])
		if err != nil || structParents == nil || int(*structParents) != pageNr {
			return fmt.Errorf("page %d StructParents = %v, want %d (err %v)", pageNr+1, structParents, pageNr, err)
		}
	}
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		return fmt.Errorf("resolve StructTreeRoot: %w", err)
	}
	documentRef, ok := root["K"].(types.IndirectRef)
	if !ok {
		return fmt.Errorf("StructTreeRoot /K is %T, want indirect reference", root["K"])
	}
	document, err := ctx.DereferenceDict(root["K"])
	if err != nil {
		return fmt.Errorf("resolve document structure element: %w", err)
	}
	kids, err := ctx.DereferenceArray(document["K"])
	if err != nil || len(kids) != 2 {
		return fmt.Errorf("document structure kids = %v (err %v), want two page elements", kids, err)
	}
	parentTree, err := ctx.DereferenceDict(root["ParentTree"])
	if err != nil {
		return fmt.Errorf("resolve ParentTree: %w", err)
	}
	nums, err := ctx.DereferenceArray(parentTree["Nums"])
	if err != nil || len(nums) != 4 {
		return fmt.Errorf("ParentTree /Nums = %v (err %v), want two key/value pairs", nums, err)
	}
	for pageNr, kid := range kids {
		structureElement, err := ctx.DereferenceDict(kid)
		if err != nil {
			return fmt.Errorf("resolve structure element for page %d: %w", pageNr+1, err)
		}
		if role, ok := structureElement["S"].(types.Name); !ok || role != types.Name("P") {
			return fmt.Errorf("page %d structure role is %v, want /P", pageNr+1, structureElement["S"])
		}
		if parentRef, ok := structureElement["P"].(types.IndirectRef); !ok || parentRef != documentRef {
			return fmt.Errorf("page %d structure element parent is %v, want document element %v", pageNr+1, structureElement["P"], documentRef)
		}
		mcr, err := ctx.DereferenceDict(structureElement["K"])
		if err != nil {
			return fmt.Errorf("resolve MCR for page %d: %w", pageNr+1, err)
		}
		pageRef, ok := mcr["Pg"].(types.IndirectRef)
		if !ok || pageRef != pageRefs[pageNr] {
			return fmt.Errorf("page %d MCR points to %v, want page ref %v", pageNr+1, mcr["Pg"], pageRefs[pageNr])
		}
		mcid, err := ctx.DereferenceInteger(mcr["MCID"])
		if err != nil || mcid == nil || *mcid != 0 {
			return fmt.Errorf("page %d MCR MCID = %v (err %v), want 0", pageNr+1, mcid, err)
		}
		parentTreeKey, err := ctx.DereferenceInteger(nums[pageNr*2])
		if err != nil || parentTreeKey == nil || int(*parentTreeKey) != pageNr {
			return fmt.Errorf("ParentTree key for page %d = %v (err %v), want %d", pageNr+1, parentTreeKey, err, pageNr)
		}
		parentArray, err := ctx.DereferenceArray(nums[pageNr*2+1])
		if err != nil || len(parentArray) != 1 {
			return fmt.Errorf("ParentTree entry for page %d = %v (err %v), want one parent", pageNr+1, parentArray, err)
		}
		parentRef, ok := parentArray[0].(types.IndirectRef)
		kidRef, kidIsRef := kid.(types.IndirectRef)
		if !ok || !kidIsRef || parentRef != kidRef {
			return fmt.Errorf("ParentTree entry for page %d is %v, want structure element %v", pageNr+1, parentArray[0], kid)
		}
	}
	return nil
}

func syntheticTwoPageInheritedPDF() []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R /Lang (en-US) /MarkInfo << /Marked true >> /StructTreeRoot 3 0 R >>",
		"<< /Type /Pages /Kids [4 0 R 5 0 R] /Count 2 /MediaBox [0 0 612 792] /Resources << /Font << /F1 6 0 R >> >> >>",
		"<< /Type /StructTreeRoot /K 7 0 R /ParentTree 8 0 R /ParentTreeNextKey 2 >>",
		"<< /Type /Page /Parent 2 0 R /Contents 9 0 R /StructParents 0 >>",
		"<< /Type /Page /Parent 2 0 R /Contents 10 0 R /StructParents 1 >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Type /StructElem /S /Document /P 3 0 R /K [11 0 R 12 0 R] >>",
		"<< /Nums [0 [11 0 R] 1 [12 0 R]] /Limits [0 1] >>",
		streamObject("/P <</MCID 0>> BDC BT /F1 12 Tf 72 740 Td (Synthetic page one) Tj ET EMC\n"),
		streamObject("/P <</MCID 0>> BDC BT /F1 12 Tf 72 740 Td (Synthetic page two) Tj ET EMC\n"),
		"<< /Type /StructElem /S /P /P 7 0 R /K << /Type /MCR /Pg 4 0 R /MCID 0 >> >>",
		"<< /Type /StructElem /S /P /P 7 0 R /K << /Type /MCR /Pg 5 0 R /MCID 0 >> >>",
		"<< /Title (Synthetic two-page inherited attributes fixture) /Producer (DAP synthetic fixture generator) >>",
	}
	return assembleSyntheticPDF(objects)
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func assembleSyntheticPDF(objects []string) []byte {
	size := len(objects)
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, size+1)
	for i, object := range objects {
		offsets[i+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", size+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R /Info %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", size+1, size, xref)
	return pdf.Bytes()
}

func compareRoundTrip(baseline, after *model.Context) error {
	if baseline.PageCount != 1 || after.PageCount != baseline.PageCount {
		return fmt.Errorf("page count changed: %d -> %d", baseline.PageCount, after.PageCount)
	}
	if baseline.Title != after.Title {
		return fmt.Errorf("title changed: %q -> %q", baseline.Title, after.Title)
	}
	for _, objectNumber := range []int{1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 16} {
		beforeObject, afterObject := baseline.Table[objectNumber], after.Table[objectNumber]
		if beforeObject == nil || afterObject == nil {
			return fmt.Errorf("PDF object %d is missing from a round-trip context", objectNumber)
		}
		if !reflect.DeepEqual(beforeObject.Object, afterObject.Object) {
			return fmt.Errorf("PDF object %d changed (%T)", objectNumber, beforeObject.Object)
		}
	}
	for _, objectNumber := range []int{13, 14} {
		beforeObject, afterObject := baseline.Table[objectNumber], after.Table[objectNumber]
		if beforeObject == nil || afterObject == nil {
			return fmt.Errorf("stream object %d is missing from a round-trip context", objectNumber)
		}
		inStream, ok1 := beforeObject.Object.(types.StreamDict)
		outStream, ok2 := afterObject.Object.(types.StreamDict)
		if !ok1 || !ok2 {
			return fmt.Errorf("object %d is not a stream", objectNumber)
		}
		if !reflect.DeepEqual(inStream.Dict, outStream.Dict) {
			return fmt.Errorf("stream %d dictionary changed", objectNumber)
		}
		if !bytes.Equal(inStream.Raw, outStream.Raw) {
			return fmt.Errorf("stream %d encoded bytes changed", objectNumber)
		}
		if !bytes.Equal(inStream.Content, outStream.Content) {
			return fmt.Errorf("stream %d decoded content changed", objectNumber)
		}
	}
	return nil
}

func TestRoundTripComparatorDetectsPreWriteProtectedMutation(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
	if err := os.WriteFile(input, syntheticPDF("statement"), 0o600); err != nil {
		t.Fatal(err)
	}
	baseline, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read immutable baseline: %v", err)
	}
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil {
		t.Fatalf("read writer context: %v", err)
	}
	page, ok := writerCtx.Table[3].Object.(types.Dict)
	if !ok {
		t.Fatalf("page object is %T", writerCtx.Table[3].Object)
	}
	page["StructParents"] = types.Integer(99)
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
		t.Fatalf("write mutated synthetic context: %v", err)
	}
	after, err := api.ReadContextFile(context.Background(), output)
	if err != nil {
		t.Fatalf("read mutated round-trip output: %v", err)
	}
	if err := compareRoundTrip(baseline, after); err == nil {
		t.Fatal("independent baseline comparison accepted changed protected page object")
	}
}

func TestRoundTripComparatorDetectsEncodedStreamMutation(t *testing.T) {
	for _, objectNumber := range []int{13, 14} {
		t.Run(fmt.Sprintf("object-%d", objectNumber), func(t *testing.T) {
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
			if err := os.WriteFile(input, syntheticPDF("statement"), 0o600); err != nil {
				t.Fatal(err)
			}
			baseline, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read immutable baseline: %v", err)
			}
			writerCtx, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read writer context: %v", err)
			}
			object := writerCtx.Table[objectNumber]
			if object == nil {
				t.Fatalf("stream object %d is missing", objectNumber)
			}
			stream, ok := object.Object.(types.StreamDict)
			if !ok || len(stream.Raw) == 0 {
				t.Fatalf("object %d raw stream is %T with %d bytes", objectNumber, object.Object, len(stream.Raw))
			}
			stream.Raw = append([]byte(nil), stream.Raw...)
			stream.Raw[len(stream.Raw)-1] ^= 1
			object.Object = stream
			if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
				t.Fatalf("write mutated stream object %d: %v", objectNumber, err)
			}
			after, err := api.ReadContextFile(context.Background(), output)
			if err != nil {
				t.Fatalf("read mutated output: %v", err)
			}
			if err := compareRoundTrip(baseline, after); err == nil || !strings.Contains(err.Error(), "encoded bytes changed") {
				t.Fatalf("independent baseline comparison did not identify changed Raw bytes in stream object %d: %v", objectNumber, err)
			}
		})
	}
}

func TestRoundTripComparatorDetectsCatalogKeyMutation(t *testing.T) {
	for _, key := range []string{"Lang", "MarkInfo", "StructTreeRoot", "AcroForm"} {
		t.Run(key, func(t *testing.T) {
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
			if err := os.WriteFile(input, syntheticPDF("statement"), 0o600); err != nil {
				t.Fatal(err)
			}
			baseline, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read immutable baseline: %v", err)
			}
			writerCtx, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read writer context: %v", err)
			}
			catalog, ok := writerCtx.Table[1].Object.(types.Dict)
			if !ok {
				t.Fatalf("catalog object is %T", writerCtx.Table[1].Object)
			}
			delete(catalog, key)
			if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil {
				t.Fatalf("write mutated synthetic catalog: %v", err)
			}
			after, err := api.ReadContextFile(context.Background(), output)
			if err != nil {
				t.Fatalf("read mutated output: %v", err)
			}
			if err := compareRoundTrip(baseline, after); err == nil {
				t.Fatalf("independent baseline comparison accepted removed /%s", key)
			}
		})
	}
}

func assertSparseParentTree(t *testing.T, ctx *model.Context) {
	t.Helper()
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatal(err)
	}
	parentTree, err := ctx.DereferenceDict(root["ParentTree"])
	if err != nil {
		t.Fatal(err)
	}
	nums, err := ctx.DereferenceArray(parentTree["Nums"])
	if err != nil {
		t.Fatal(err)
	}
	pageParents, err := ctx.DereferenceArray(nums[1])
	if err != nil {
		t.Fatal(err)
	}
	if len(pageParents) != 3 || pageParents[1] != nil {
		t.Fatalf("expected MCID 1 null hole, got %#v", pageParents)
	}
	if _, ok := pageParents[0].(types.IndirectRef); !ok {
		t.Fatalf("MCID 0 parent is %T, want indirect reference", pageParents[0])
	}
	if _, ok := pageParents[2].(types.IndirectRef); !ok {
		t.Fatalf("MCID 2 parent is %T, want indirect reference", pageParents[2])
	}
}

func syntheticPDF(family string) []byte {
	line1 := "Synthetic " + family + " example"
	line2 := map[string]string{"statement": "Balance 42.00", "invoice": "Balance 42.00", "correspondence": "Dear reader", "report": "Results summary"}[family]
	stream := "/P <</MCID 2>> BDC BT /F1 12 Tf 72 740 Td (" + line1 + ") Tj ET EMC\n" +
		"/P <</MCID 0>> BDC BT /F1 12 Tf 72 720 Td (" + line2 + ") Tj ET EMC\n"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R /Lang (en-US) /MarkInfo << /Marked true >> /StructTreeRoot 6 0 R /AcroForm << /Fields [8 0 R] /DA (/F1 10 Tf 0 g) /DR << /Font << /F1 9 0 R >> >> >> >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 9 0 R >> /XObject << /Fm 14 0 R >> >> /Contents 13 0 R /Annots [7 0 R 8 0 R] /StructParents 0 >>",
		"<< /Type /StructElem /S /Document /P 6 0 R /K [10 0 R 5 0 R 11 0 R 12 0 R 15 0 R] >>",
		"<< /Type /StructElem /S /P /P 4 0 R /K << /Type /MCR /Pg 3 0 R /MCID 2 >> >>",
		"<< /Type /StructTreeRoot /K 4 0 R /ParentTree 16 0 R /ParentTreeNextKey 4 >>",
		"<< /Type /Annot /Subtype /Link /Rect [72 700 160 720] /Border [0 0 0] /A << /S /URI /URI (https://example.invalid/) >> /StructParent 1 >>",
		"<< /Type /Annot /Subtype /Widget /FT /Tx /T (amount) /V (42.00) /Rect [72 650 172 670] /P 3 0 R /StructParent 2 >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Type /StructElem /S /P /P 4 0 R /K << /Type /MCR /Pg 3 0 R /MCID 0 >> >>",
		"<< /Type /StructElem /S /Link /P 4 0 R /K << /Type /OBJR /Pg 3 0 R /Obj 7 0 R >> >>",
		"<< /Type /StructElem /S /Form /P 4 0 R /K << /Type /OBJR /Pg 3 0 R /Obj 8 0 R >> >>",
		streamObject(stream),
		streamObject("/P <</MCID 0>> BDC EMC\n", "/Type /XObject /Subtype /Form /BBox [0 0 10 10] /Resources << >> /StructParents 3"),
		"<< /Type /StructElem /S /Span /P 4 0 R /K << /Type /MCR /Pg 3 0 R /Stm 14 0 R /MCID 0 >> >>",
		"<< /Nums [0 [10 0 R null 5 0 R] 1 11 0 R 2 12 0 R 3 [15 0 R]] /Limits [0 3] >>",
		"<< /Title (Synthetic " + family + " fixture) /Producer (DAP synthetic fixture generator) >>",
	}
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R /Info 17 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return pdf.Bytes()
}

func streamObject(content string, entries ...string) string {
	additional := ""
	if len(entries) > 0 {
		additional = " " + entries[0]
	}
	return fmt.Sprintf("<<%s /Length %d >>\nstream\n%s\nendstream", additional, len(content), content)
}
