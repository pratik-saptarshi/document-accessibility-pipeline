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
			beforeFonts := fontMetricKeySnapshot(t, parsed)
			beforeMetadata := pdfMetadataSnapshot(parsed)

			staging, err := os.MkdirTemp("", "dap-financial-noop-")
			if err != nil {
				t.Fatal(err)
			}
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
			afterFonts := fontMetricKeySnapshot(t, fresh)
			added := addedFontMetricKeys(beforeFonts, afterFonts)
			sort.Strings(added)
			wantAdded := []string{"F1:FirstChar", "F1:FontDescriptor", "F1:LastChar", "F1:Widths"}
			if !reflect.DeepEqual(added, wantAdded) {
				t.Fatalf("positive no-op disposition changed: added Helvetica metric keys=%v, want the reviewed bounded no-go delta %v; input_sha256=%s output_sha256=%s staged_output=%s", added, wantAdded, fixture.digest, hex.EncodeToString(outputHash[:]), outputPath)
			}
			t.Logf("NO-GO: positive no-op added Helvetica metric keys %v; resulting metrics %s; table structure and fresh-read cell bindings are unchanged; Info metadata before=%#v after=%#v; input_sha256=%s output_sha256=%s input_bytes=%d output_bytes=%d staged_output=%s", added, fontMetricDetails(t, fresh), beforeMetadata, pdfMetadataSnapshot(fresh), fixture.digest, hex.EncodeToString(outputHash[:]), len(input), len(candidate), outputPath)
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
		MaxInputBytes: 64 << 10, MaxObjectBytes: 32 << 10,
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

func fontMetricKeySnapshot(t *testing.T, ctx *model.Context) map[string]map[string]bool {
	t.Helper()
	page, _, _, err := ctx.PageDict(context.Background(), 1, false)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := ctx.DereferenceDict(page["Resources"])
	if err != nil {
		t.Fatalf("dereference page resources: %v", err)
	}
	fonts, err := ctx.DereferenceDict(resources["Font"])
	if err != nil {
		t.Fatalf("dereference page fonts: %v", err)
	}
	result := make(map[string]map[string]bool, len(fonts))
	for resource, object := range fonts {
		font, err := ctx.DereferenceDict(object)
		if err != nil {
			t.Fatalf("dereference page font %q: %v", resource, err)
		}
		keys := make(map[string]bool)
		for _, key := range []string{"FirstChar", "LastChar", "Widths", "FontDescriptor"} {
			_, keys[key] = font[key]
		}
		result[fmt.Sprint(resource)] = keys
	}
	return result
}

func addedFontMetricKeys(before, after map[string]map[string]bool) []string {
	var added []string
	for resource, prior := range before {
		current, ok := after[resource]
		if !ok {
			added = append(added, resource+":<font-resource-removed>")
			continue
		}
		for key, existed := range prior {
			if !existed && current[key] {
				added = append(added, resource+":"+key)
			}
		}
	}
	return added
}

func fontMetricDetails(t *testing.T, ctx *model.Context) string {
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
	var details []string
	for resource, object := range fonts {
		font, err := ctx.DereferenceDict(object)
		if err != nil {
			t.Fatalf("dereference font %q: %v", resource, err)
		}
		widths, ok := font["Widths"].(types.Array)
		if !ok {
			t.Fatalf("font %q Widths = %T, want array", resource, font["Widths"])
		}
		widthsDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", widths)))
		descriptorRef := fmt.Sprintf("%v", font["FontDescriptor"])
		descriptor, err := ctx.DereferenceDict(font["FontDescriptor"])
		if err != nil {
			t.Fatalf("dereference font %q descriptor: %v", resource, err)
		}
		keys := make([]string, 0, len(descriptor))
		for key := range descriptor {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		descriptorDigest := sha256.Sum256([]byte(fmt.Sprintf("%#v", descriptor)))
		details = append(details, fmt.Sprintf("%s FirstChar=%v LastChar=%v Widths.count=%d Widths.sha256=%x FontDescriptor.ref=%s FontDescriptor.keys=%v FontDescriptor.sha256=%x", resource, font["FirstChar"], font["LastChar"], len(widths), widthsDigest, descriptorRef, keys, descriptorDigest))
	}
	sort.Strings(details)
	return strings.Join(details, "; ")
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
