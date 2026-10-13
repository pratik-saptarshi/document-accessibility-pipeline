package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// TestPositiveNoopWriterControlRecordsNoGo is the first gate in the bounded
// writer experiment. It starts with each immutable positive fixture, full-writes
// it in private temporary staging, and freshly reads the result. No damaged
// input is opened and no repair candidate is produced here.
func TestPositiveNoopWriterControlRecordsNoGo(t *testing.T) {
	fixtures := []struct {
		family string
		path   string
		digest string
	}{
		{"statement", "../../docs/fixtures/remediation/financial-tables/statement-positive.pdf", "14c0456e3b8ec0146bc2a25fa9af51d0c5559f8a53e23c4e163a3c422e66cfed"},
		{"invoice", "../../docs/fixtures/remediation/financial-tables/invoice-positive.pdf", "9600519a8b3c64e7c48796dabb6f09cbe0c432bfe2daa4d2f6764937d3f2695e"},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.family, func(t *testing.T) {
			input, err := os.ReadFile(fixture.path)
			if err != nil {
				t.Fatal(err)
			}
			inputHash := sha256.Sum256(input)
			if hex.EncodeToString(inputHash[:]) != fixture.digest {
				t.Fatalf("immutable input SHA-256 = %x, want %s", inputHash, fixture.digest)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			conf := boundedStatelessConfig(t)
			parsed, err := api.ReadContext(ctx, bytes.NewReader(input), conf)
			if err != nil {
				t.Fatalf("read positive fixture: %v", err)
			}
			if err := api.ValidateContext(ctx, parsed); err != nil {
				t.Fatalf("validate positive fixture: %v", err)
			}
			beforeHeaders, beforeRows := tableObservations(t, parsed)
			beforeFonts, err := snapshotPageFonts(parsed)
			if err != nil {
				t.Fatalf("snapshot input font resources: %v", err)
			}
			beforeMetadata := pdfMetadataSnapshot(parsed)

			staging := t.TempDir()
			outputPath := filepath.Join(staging, fixture.family+"-positive-noop.pdf")
			output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			bounded := &maxBytesWriter{writer: output, remaining: 256 << 10}
			writeErr := api.WriteContext(ctx, parsed, bounded)
			closeErr := output.Close()
			if writeErr != nil {
				t.Fatalf("positive no-op full-write: %v", writeErr)
			}
			if closeErr != nil {
				t.Fatalf("close private staged output: %v", closeErr)
			}

			candidate, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			outputHash := sha256.Sum256(candidate)
			fresh, err := api.ReadContext(ctx, bytes.NewReader(candidate), boundedStatelessConfig(t))
			if err != nil {
				t.Fatalf("fresh-read full-write output: %v", err)
			}
			if err := api.ValidateContext(ctx, fresh); err != nil {
				t.Fatalf("fresh-read validation: %v", err)
			}
			afterHeaders, afterRows := tableObservations(t, fresh)
			if !reflect.DeepEqual(beforeHeaders, afterHeaders) || !reflect.DeepEqual(beforeRows, afterRows) {
				t.Fatalf("positive no-op changed declared table structure: before headers=%#v rows=%#v; after headers=%#v rows=%#v", beforeHeaders, beforeRows, afterHeaders, afterRows)
			}
			if err := assertIDTree(fresh, mustStructTreeRoot(t, fresh)); err != nil {
				t.Fatalf("fresh-read identifier tree: %v", err)
			}
			if err := assertCellContentBindings(fresh); err != nil {
				t.Fatalf("fresh-read cell/content/ParentTree bindings: %v", err)
			}
			afterFonts, err := snapshotPageFonts(fresh)
			if err != nil {
				t.Fatalf("snapshot output font resources: %v", err)
			}
			if err := comparePositiveNoopFontDelta(beforeFonts, afterFonts, fresh); err != nil {
				t.Fatalf("positive no-op font delta differs from the exact reviewed no-go: %v; input_sha256=%s output_sha256=%s staged_output=%s", err, fixture.digest, hex.EncodeToString(outputHash[:]), outputPath)
			}
			t.Logf("NO-GO: positive no-op added only the exact reviewed Helvetica metrics; resulting metrics %s; table structure and fresh-read cell bindings are unchanged; Info metadata before=%#v after=%#v; input_sha256=%s output_sha256=%s input_bytes=%d output_bytes=%d", fontMetricDetails(t, fresh), beforeMetadata, pdfMetadataSnapshot(fresh), fixture.digest, hex.EncodeToString(outputHash[:]), len(input), len(candidate))
			assertFontOracleRejectsUnreviewedMutations(t, candidate, beforeFonts)
		})
	}
}

func boundedStatelessConfig(t *testing.T) *model.Configuration {
	t.Helper()
	conf, err := api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})
	if err != nil {
		t.Fatalf("load stateless pdfcpu configuration: %v", err)
	}
	conf.Offline = true
	conf.Limits = model.ResourceLimits{
		MaxInputBytes: v2PDFInputLimit, MaxObjectBytes: 32 << 10,
		MaxStreamBytes: 32 << 10, MaxDecodeBytes: 64 << 10,
		MaxImagePixels: 1_000_000, MaxImageBytes: 4 << 20,
		MaxObjectCount: 256, MaxObjectStreamCount: 64,
		MaxObjectStreamFirst: 16 << 10, MaxXRefEntries: 256,
		MaxRecursionDepth: 32,
	}
	return conf
}

func mustStructTreeRoot(t *testing.T, ctx *model.Context) types.Dict {
	t.Helper()
	root, err := ctx.DereferenceDict(ctx.RootDict["StructTreeRoot"])
	if err != nil {
		t.Fatal(err)
	}
	return root
}

type fontResourceSnapshot struct {
	reference  string
	dictionary types.Dict
	indirect   map[string]types.Object
}

type fontResourcesSnapshot map[string]fontResourceSnapshot

func snapshotPageFonts(ctx *model.Context) (fontResourcesSnapshot, error) {
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		return nil, fmt.Errorf("read page dictionary: %w", err)
	}
	resources, err := ctx.DereferenceDict(page["Resources"])
	if err != nil {
		return nil, fmt.Errorf("dereference page resources: %w", err)
	}
	fonts, err := ctx.DereferenceDict(resources["Font"])
	if err != nil {
		return nil, fmt.Errorf("dereference page font resources: %w", err)
	}
	result := make(fontResourcesSnapshot, len(fonts))
	for resource, object := range fonts {
		name := string(resource)
		font, err := ctx.DereferenceDict(object)
		if err != nil {
			return nil, fmt.Errorf("dereference font resource /%s: %w", name, err)
		}
		reference, err := fontReferenceIdentity(object)
		if err != nil {
			return nil, fmt.Errorf("snapshot font resource /%s reference: %w", name, err)
		}
		indirect := make(map[string]types.Object, len(font))
		for key, value := range font {
			if reference, ok := value.(types.IndirectRef); ok {
				resolved, err := ctx.Dereference(reference)
				if err != nil {
					return nil, fmt.Errorf("snapshot font resource /%s entry /%s (%s): %w", name, key, reference.PDFString(), err)
				}
				indirect[key] = resolved.Clone()
			}
		}
		result[name] = fontResourceSnapshot{reference: reference, dictionary: font.Clone().(types.Dict), indirect: indirect}
	}
	return result, nil
}

func fontReferenceIdentity(object types.Object) (string, error) {
	if reference, ok := object.(types.IndirectRef); ok {
		return fmt.Sprintf("%d %d R", reference.ObjectNumber, reference.GenerationNumber), nil
	}
	return "direct", nil
}

func comparePositiveNoopFontDelta(before, after fontResourcesSnapshot, output *model.Context) error {
	var addedResources, removedResources []string
	for name := range before {
		if _, ok := after[name]; !ok {
			removedResources = append(removedResources, name)
		}
	}
	for name := range after {
		if _, ok := before[name]; !ok {
			addedResources = append(addedResources, name)
		}
	}
	sort.Strings(addedResources)
	sort.Strings(removedResources)
	if len(addedResources) != 0 || len(removedResources) != 0 {
		return fmt.Errorf("font resource names changed: added=%v removed=%v", addedResources, removedResources)
	}

	var added []string
	for name, original := range before {
		written := after[name]
		if original.reference != written.reference {
			return fmt.Errorf("font resource /%s reference changed from %s to %s", name, original.reference, written.reference)
		}
		for key, oldValue := range original.dictionary {
			newValue, ok := written.dictionary[key]
			if !ok {
				return fmt.Errorf("font resource /%s entry /%s was removed", name, key)
			}
			if !reflect.DeepEqual(oldValue, newValue) {
				return fmt.Errorf("font resource /%s entry /%s changed", name, key)
			}
		}
		for key := range written.dictionary {
			if _, ok := original.dictionary[key]; !ok {
				added = append(added, name+":"+key)
			}
		}
		for key, oldValue := range original.indirect {
			newValue, ok := written.indirect[key]
			if !ok || !reflect.DeepEqual(oldValue, newValue) {
				return fmt.Errorf("font resource /%s indirect entry /%s target changed", name, key)
			}
		}
	}
	sort.Strings(added)
	wantAdded := []string{"F1:FirstChar", "F1:FontDescriptor", "F1:LastChar", "F1:Widths"}
	if !reflect.DeepEqual(added, wantAdded) {
		return fmt.Errorf("added font entries %v, want exactly %v", added, wantAdded)
	}
	return assertReviewedHelveticaMetrics(output)
}

func assertReviewedHelveticaMetrics(ctx *model.Context) error {
	resources, err := pageFontResources(ctx)
	if err != nil {
		return err
	}
	if len(resources) != 1 {
		return fmt.Errorf("reviewed metric result has %d font resources, want only /F1", len(resources))
	}
	font, err := ctx.DereferenceDict(resources["F1"])
	if err != nil {
		return fmt.Errorf("dereference /F1: %w", err)
	}
	for key, want := range map[string]types.Name{"Type": "Font", "Subtype": "Type1", "BaseFont": "Helvetica"} {
		if got, ok := font[key].(types.Name); !ok || got != want {
			return fmt.Errorf("/F1 /%s = %v, want /%s", key, font[key], want)
		}
	}
	first, ok := font["FirstChar"].(types.Integer)
	if !ok || first != 0 {
		return fmt.Errorf("/F1 /FirstChar = %v, want integer 0", font["FirstChar"])
	}
	last, ok := font["LastChar"].(types.Integer)
	if !ok || last != 255 {
		return fmt.Errorf("/F1 /LastChar = %v, want integer 255", font["LastChar"])
	}
	widths, err := resolveFontObject(ctx, font["Widths"])
	if err != nil {
		return fmt.Errorf("resolve /F1 /Widths: %w", err)
	}
	widthArray, ok := widths.(types.Array)
	if !ok || len(widthArray) != 256 {
		return fmt.Errorf("/F1 /Widths resolves to %T with length %d, want 256 numeric entries", widths, arrayLen(widths))
	}
	for i, value := range widthArray {
		if _, ok := value.(types.Integer); !ok {
			return fmt.Errorf("/F1 /Widths[%d] = %T, want integer", i, value)
		}
	}
	widthDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", widthArray)))
	if got := hex.EncodeToString(widthDigest[:]); got != reviewedWidthsSHA256 {
		return fmt.Errorf("/F1 /Widths SHA-256 = %s, want reviewed value %s", got, reviewedWidthsSHA256)
	}
	if _, direct := font["Widths"].(types.Array); !direct {
		return fmt.Errorf("/F1 /Widths is indirect, want the reviewed direct array representation")
	}
	descriptorRef, ok := font["FontDescriptor"].(types.IndirectRef)
	if !ok {
		return fmt.Errorf("/F1 /FontDescriptor = %T, want reviewed indirect reference", font["FontDescriptor"])
	}
	if got := descriptorRef.PDFString(); got != "23 0 R" {
		return fmt.Errorf("/F1 /FontDescriptor reference = %s, want reviewed identity 23 0 R", got)
	}
	descriptor, err := ctx.DereferenceDict(descriptorRef)
	if err != nil {
		return fmt.Errorf("dereference /F1 /FontDescriptor: %w", err)
	}
	keys := make([]string, 0, len(descriptor))
	for key := range descriptor {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	wantKeys := []string{"Ascent", "CapHeight", "Descent", "Flags", "FontBBox", "FontName", "ItalicAngle", "StemV", "Type"}
	if !reflect.DeepEqual(keys, wantKeys) {
		return fmt.Errorf("/F1 /FontDescriptor keys = %v, want %v", keys, wantKeys)
	}
	descriptorDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", descriptor)))
	if got := hex.EncodeToString(descriptorDigest[:]); got != reviewedDescriptorSHA256 {
		return fmt.Errorf("/F1 /FontDescriptor SHA-256 = %s, want reviewed value %s", got, reviewedDescriptorSHA256)
	}
	return nil
}

const (
	reviewedWidthsSHA256     = "6cb5449864ad91580d3eb294e278eaca115f354a611fb37e3b92056349874e65"
	reviewedDescriptorSHA256 = "92a177916a70f32f53ffa12561ce3b05210feb36199ffc1b9eab1279c98cc108"
)

func resolveFontObject(ctx *model.Context, object types.Object) (types.Object, error) {
	if _, ok := object.(types.IndirectRef); ok {
		return ctx.Dereference(object)
	}
	return object, nil
}

func arrayLen(object types.Object) int {
	if array, ok := object.(types.Array); ok {
		return len(array)
	}
	return 0
}

func pageFontResources(ctx *model.Context) (types.Dict, error) {
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		return nil, err
	}
	resources, err := ctx.DereferenceDict(page["Resources"])
	if err != nil {
		return nil, err
	}
	return ctx.DereferenceDict(resources["Font"])
}

func fontMetricDetails(t *testing.T, ctx *model.Context) string {
	t.Helper()
	resources, err := pageFontResources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	font, err := ctx.DereferenceDict(resources["F1"])
	if err != nil {
		t.Fatal(err)
	}
	widths, err := resolveFontObject(ctx, font["Widths"])
	if err != nil {
		t.Fatal(err)
	}
	widthDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", widths)))
	descriptorRef := font["FontDescriptor"].(types.IndirectRef)
	descriptor, err := ctx.DereferenceDict(descriptorRef)
	if err != nil {
		t.Fatal(err)
	}
	descriptorDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", descriptor)))
	return fmt.Sprintf("/F1 type=%v subtype=%v basefont=%v resource_ref=5 0 R FirstChar=%v LastChar=%v Widths.count=%d Widths.sha256=%x FontDescriptor.ref=%s FontDescriptor.sha256=%x", font["Type"], font["Subtype"], font["BaseFont"], font["FirstChar"], font["LastChar"], arrayLen(widths), widthDigest, descriptorRef.PDFString(), descriptorDigest)
}

func assertFontOracleRejectsUnreviewedMutations(t *testing.T, output []byte, before fontResourcesSnapshot) {
	t.Helper()
	tests := []struct {
		name   string
		want   string
		mutate func(*testing.T, *model.Context)
	}{
		{"changed BaseFont", "entry /BaseFont changed", func(t *testing.T, ctx *model.Context) {
			font := fontDictForMutation(t, ctx)
			font["BaseFont"] = types.Name("Courier")
		}},
		{"removed BaseFont", "entry /BaseFont was removed", func(t *testing.T, ctx *model.Context) {
			delete(fontDictForMutation(t, ctx), "BaseFont")
		}},
		{"added font resource", "added=[F2]", func(t *testing.T, ctx *model.Context) {
			fonts, err := pageFontResources(ctx)
			if err != nil {
				t.Fatal(err)
			}
			fonts["F2"] = fonts["F1"]
		}},
		{"changed font resource reference", "font resource /F1 reference changed", func(t *testing.T, ctx *model.Context) {
			fonts, err := pageFontResources(ctx)
			if err != nil {
				t.Fatal(err)
			}
			font, err := ctx.DereferenceDict(fonts["F1"])
			if err != nil {
				t.Fatal(err)
			}
			objectNumber := ctx.XRefTable.InsertNew(model.XRefTableEntry{Object: font.Clone(), Valid: true})
			fonts["F1"] = *types.NewIndirectRef(objectNumber, 0)
		}},
		{"changed width", "/Widths SHA-256", func(t *testing.T, ctx *model.Context) {
			font := fontDictForMutation(t, ctx)
			widths, err := resolveFontObject(ctx, font["Widths"])
			if err != nil {
				t.Fatal(err)
			}
			array := append(types.Array(nil), widths.(types.Array)...)
			array[0] = types.Integer(int(array[0].(types.Integer)) + 1)
			font["Widths"] = array
		}},
		{"indirect changed width", "/Widths SHA-256", func(t *testing.T, ctx *model.Context) {
			font := fontDictForMutation(t, ctx)
			widths, err := resolveFontObject(ctx, font["Widths"])
			if err != nil {
				t.Fatal(err)
			}
			array := append(types.Array(nil), widths.(types.Array)...)
			array[0] = types.Integer(int(array[0].(types.Integer)) + 1)
			objectNumber := ctx.XRefTable.InsertNew(model.XRefTableEntry{Object: array, Valid: true})
			font["Widths"] = *types.NewIndirectRef(objectNumber, 0)
		}},
		{"indirect unchanged width", "/Widths is indirect", func(t *testing.T, ctx *model.Context) {
			font := fontDictForMutation(t, ctx)
			widths, err := resolveFontObject(ctx, font["Widths"])
			if err != nil {
				t.Fatal(err)
			}
			array := append(types.Array(nil), widths.(types.Array)...)
			objectNumber := ctx.XRefTable.InsertNew(model.XRefTableEntry{Object: array, Valid: true})
			font["Widths"] = *types.NewIndirectRef(objectNumber, 0)
		}},
		{"changed indirect descriptor", "/FontDescriptor SHA-256", func(t *testing.T, ctx *model.Context) {
			font := fontDictForMutation(t, ctx)
			descriptor, err := ctx.DereferenceDict(font["FontDescriptor"])
			if err != nil {
				t.Fatal(err)
			}
			descriptor["Ascent"] = types.Integer(1)
		}},
		{"changed allowed metric value", "/FirstChar", func(t *testing.T, ctx *model.Context) {
			fontDictForMutation(t, ctx)["FirstChar"] = types.Integer(1)
		}},
	}
	for _, test := range tests {
		t.Run("rejects "+test.name, func(t *testing.T) {
			readCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			ctx, err := api.ReadContext(readCtx, bytes.NewReader(output), boundedStatelessConfig(t))
			if err != nil {
				t.Fatal(err)
			}
			if err := api.ValidateContext(readCtx, ctx); err != nil {
				t.Fatalf("validate mutation source: %v", err)
			}
			test.mutate(t, ctx)
			after, err := snapshotPageFonts(ctx)
			if err != nil {
				t.Fatal(err)
			}
			err = comparePositiveNoopFontDelta(before, after, ctx)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("mutated font delta error = %v, want error containing %q", err, test.want)
			}
		})
	}
}

func fontDictForMutation(t *testing.T, ctx *model.Context) types.Dict {
	t.Helper()
	fonts, err := pageFontResources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	font, err := ctx.DereferenceDict(fonts["F1"])
	if err != nil {
		t.Fatal(err)
	}
	return font
}

type pdfMetadata struct {
	Title        string
	Subject      string
	Author       string
	Creator      string
	Producer     string
	CreationDate string
	ModDate      string
	Keywords     string
}

func pdfMetadataSnapshot(ctx *model.Context) pdfMetadata {
	return pdfMetadata{
		Title: ctx.XRefTable.Title, Subject: ctx.XRefTable.Subject, Author: ctx.XRefTable.Author,
		Creator: ctx.XRefTable.Creator, Producer: ctx.XRefTable.Producer,
		CreationDate: ctx.XRefTable.CreationDate, ModDate: ctx.XRefTable.ModDate,
		Keywords: ctx.XRefTable.Keywords,
	}
}

type maxBytesWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *maxBytesWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > w.remaining {
		return 0, fmt.Errorf("staged output exceeds 256 KiB experiment limit")
	}
	n, err := w.writer.Write(p)
	w.remaining -= int64(n)
	if err == nil && n != len(p) {
		return n, io.ErrShortWrite
	}
	return n, err
}
