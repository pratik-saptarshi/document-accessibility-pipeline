// Package pdfpreserve contains a bounded characterization harness for the
// proposed PDF writer. The tests do not implement accessibility remediation.
package pdfpreserve

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestFullWriteRoundTripPreservesSyntheticPDFFeatures(t *testing.T) {
	for _, family := range []string{"statement", "invoice", "correspondence", "report"} {
		t.Run(family, func(t *testing.T) {
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.pdf"), filepath.Join(dir, "output.pdf")
			if err := os.WriteFile(input, syntheticPDF(family), 0o600); err != nil {
				t.Fatal(err)
			}
			before, err := api.ReadContextFile(context.Background(), input)
			if err != nil {
				t.Fatalf("read synthetic input: %v", err)
			}
			if err := api.WriteContextFile(context.Background(), before, output); err != nil {
				t.Fatalf("full write: %v", err)
			}
			after, err := api.ReadContextFile(context.Background(), output)
			if err != nil {
				t.Fatalf("read round-trip output: %v", err)
			}
			if before.PageCount != 1 || after.PageCount != before.PageCount {
				t.Fatalf("page count changed: %d -> %d", before.PageCount, after.PageCount)
			}
			if before.Title != after.Title {
				t.Fatalf("title changed: %q -> %q", before.Title, after.Title)
			}
			for _, objectNumber := range []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 16} {
				if !reflect.DeepEqual(before.Table[objectNumber].Object, after.Table[objectNumber].Object) {
					t.Errorf("PDF object %d changed (%T)", objectNumber, before.Table[objectNumber].Object)
				}
			}
			for _, objectNumber := range []int{13, 14} {
				inStream, ok1 := before.Table[objectNumber].Object.(types.StreamDict)
				outStream, ok2 := after.Table[objectNumber].Object.(types.StreamDict)
				if !ok1 || !ok2 {
					t.Fatalf("object %d is not a stream", objectNumber)
				}
				if !reflect.DeepEqual(inStream.Dict, outStream.Dict) {
					t.Errorf("stream %d dictionary changed", objectNumber)
				}
				if !bytes.Equal(inStream.Content, outStream.Content) {
					t.Errorf("stream %d decoded content changed", objectNumber)
				}
			}
			assertSparseParentTree(t, after)
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
