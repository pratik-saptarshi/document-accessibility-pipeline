package financialfixtures

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthorFinancialTableV2ReproducesTrackedBytesWithoutGOMODCACHE(t *testing.T) {
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		output, err := exec.Command("go", "env", "GOMODCACHE").Output()
		if err != nil {
			t.Fatal(err)
		}
		moduleCache = strings.TrimSpace(string(output))
	}
	t.Setenv("GOMODCACHE", "")
	afmPath, err := ResolvePinnedHelveticaAFM(moduleCache)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.ToSlash(afmPath), "github.com/pdfcpu/pdfcpu@v0.16.2/internal/corefont/Core14_AFMs/Helvetica.afm") {
		t.Fatalf("resolved AFM path = %s", afmPath)
	}
	root := filepath.Join("..", "..")
	before := snapshotV2AuthorArtifacts(t, root)
	if err := AuthorFinancialTableV2(root, afmPath); err != nil {
		t.Fatalf("reproduce metric-complete v2 fixtures: %v", err)
	}
	after := snapshotV2AuthorArtifacts(t, root)
	if !equalByteSnapshots(before, after) {
		t.Fatal("deterministic authoring changed already-correct v2 fixture or manifest bytes")
	}
}

func TestResolvePinnedHelveticaAFMUsesGoEnvWhenGOMODCACHEUnset(t *testing.T) {
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		output, err := exec.Command("go", "env", "GOMODCACHE").Output()
		if err != nil {
			t.Fatal(err)
		}
		moduleCache = strings.TrimSpace(string(output))
	}
	source := filepath.Join(moduleCache, "github.com/pdfcpu/pdfcpu@v0.16.2/internal/corefont/Core14_AFMs/Helvetica.afm")
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	cache := filepath.Join(temp, "module-cache")
	afmpath := filepath.Join(cache, "github.com/pdfcpu/pdfcpu@v0.16.2/internal/corefont/Core14_AFMs/Helvetica.afm")
	if err := os.MkdirAll(filepath.Dir(afmpath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(afmpath, afmdat, 0o600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(temp, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	goTool := filepath.Join(bin, "go")
	if err := os.WriteFile(goTool, []byte("#!/bin/sh\nprintf '%s\\n' '"+cache+"'\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOMODCACHE", "")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	resolved, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	if resolved != afmpath {
		t.Fatalf("resolved AFM path = %s, want go env path %s", resolved, afmpath)
	}
}

func TestAuthorFinancialTableV2RejectsStaleManifestWithoutClobber(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	tempRoot := t.TempDir()
	copyV2AuthorInputs(t, repoRoot, tempRoot)
	manifestPath := filepath.Join(tempRoot, "docs/fixtures/remediation/financial-table-experiment.v2.json")
	originalManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	staleManifest := bytes.Replace(originalManifest, []byte("metric-complete PDF 2.0 inputs"), []byte("stale PDF 2.0 inputs"), 1)
	if bytes.Equal(staleManifest, originalManifest) {
		t.Fatal("failed to create stale manifest control")
	}
	if err := os.WriteFile(manifestPath, staleManifest, 0o644); err != nil {
		t.Fatal(err)
	}
	before := snapshotV2AuthorArtifacts(t, tempRoot)
	afmPath, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	if err := AuthorFinancialTableV2(tempRoot, afmPath); err == nil || !strings.Contains(err.Error(), "manifest pins differ") {
		t.Fatalf("stale manifest failure = %v, want preflight mismatch", err)
	}
	after := snapshotV2AuthorArtifacts(t, tempRoot)
	if !equalByteSnapshots(before, after) {
		t.Fatal("stale manifest preflight modified a v2 fixture or manifest")
	}
}

func TestMetricCompleteV2PDFBuilderIsDeterministicAndNarrow(t *testing.T) {
	manifestBytes, err := os.ReadFile(evidenceV2Path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence v2FixtureEvidence
	if err := json.Unmarshal(manifestBytes, &evidence); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range evidence.Cases {
		for _, item := range []struct{ input, output, digest string }{
			{baseFixturePath(t, fixture.Family, "positive"), filepath.Join("..", "..", fixture.PositivePath), fixture.PositiveSHA},
			{baseFixturePath(t, fixture.Family, "damaged"), filepath.Join("..", "..", fixture.DamagedPath), fixture.DamagedSHA},
		} {
			input, err := os.ReadFile(filepath.Join("..", "..", item.input))
			if err != nil {
				t.Fatal(err)
			}
			first, err := buildMetricCompleteV2PDF(input, evidence.Metrics.Widths)
			if err != nil {
				t.Fatal(err)
			}
			second, err := buildMetricCompleteV2PDF(input, evidence.Metrics.Widths)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(item.output)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) || !bytes.Equal(first, want) || sha256Hex(first) != item.digest {
				t.Fatalf("deterministic v2 author output differs for %s", item.output)
			}
		}
	}
	if _, err := buildMetricCompleteV2PDF([]byte("%PDF-2.0\n"), evidence.Metrics.Widths); err == nil || !strings.Contains(err.Error(), "object count") {
		t.Fatalf("unexpected minimal input error = %v, want bounded object inventory rejection", err)
	}
	if _, err := buildMetricCompleteV2PDF([]byte("%PDF-2.0\n"), evidence.Metrics.Widths[:255]); err == nil || !strings.Contains(err.Error(), "width count") {
		t.Fatalf("unexpected width-vector error = %v, want exact width count rejection", err)
	}
}

func TestHelveticaAFMMappingAndObservationManifestAreDeterministic(t *testing.T) {
	afmpath := filepath.Join(t.TempDir(), "Helvetica.afm")
	afmdat := []byte("StartFontMetrics 4.1\nC 32 ; WX 278 ; N space ; B 0 0 0 0 ;\nC 65 ; WX 667 ; N A ; B 0 0 0 0 ;\nC -1 ; WX 111 ; N ignored ; B 0 0 0 0 ;\nEndCharMetrics\n")
	if err := os.WriteFile(afmpath, afmdat, 0o600); err != nil {
		t.Fatal(err)
	}
	widths, digest, err := parseHelveticaAFM(afmpath)
	if err != nil {
		t.Fatal(err)
	}
	if len(widths) != 256 || widths[32] != 278 || widths[65] != 667 || widths[66] != 0 || digest != sha256Hex(afmdat) {
		t.Fatalf("AFM mapping sample mismatch: len=%d width[32]=%d width[65]=%d width[66]=%d digest=%s", len(widths), widths[32], widths[65], widths[66], digest)
	}

	const hashA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const hashB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const hashC = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	records := []FinancialTableV2ObservationInput{
		{Name: "statement-positive-v2", PDFSHA256: hashA, TextSHA256: hashB, RenderSHA256: hashC},
		{Name: "statement-damaged-v2", PDFSHA256: hashA, TextSHA256: hashB, RenderSHA256: hashC},
		{Name: "invoice-positive-v2", PDFSHA256: hashA, TextSHA256: hashB, RenderSHA256: hashC},
		{Name: "invoice-damaged-v2", PDFSHA256: hashA, TextSHA256: hashB, RenderSHA256: hashC},
	}
	var first, second bytes.Buffer
	if err := WriteFinancialTableV2ObservationManifest(records, "pdftotext version test", "10.05.1", &first); err != nil {
		t.Fatal(err)
	}
	if err := WriteFinancialTableV2ObservationManifest(records, "pdftotext version test", "10.05.1", &second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("observation manifest JSON is not deterministic")
	}
	var manifest v2ObservationManifest
	if err := json.Unmarshal(first.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != "2.0.0" || len(manifest.Observations) != 4 || manifest.Observations[0].TextExitStatus != 0 || !strings.Contains(manifest.Observations[0].RenderTool, "Ghostscript 10.05.1") {
		t.Fatalf("emitted v2 observation manifest = %#v", manifest)
	}
	var rejected bytes.Buffer
	reversed := append([]FinancialTableV2ObservationInput(nil), records...)
	reversed[0], reversed[1] = reversed[1], reversed[0]
	if err := WriteFinancialTableV2ObservationManifest(reversed, "pdftotext version test", "10.05.1", &rejected); err == nil || rejected.Len() != 0 {
		t.Fatalf("out-of-order records error=%v output_bytes=%d, want pre-serialization rejection", err, rejected.Len())
	}
}

func snapshotV2AuthorArtifacts(t *testing.T, root string) map[string][]byte {
	t.Helper()
	paths := []string{
		"docs/fixtures/remediation/financial-table-experiment.v2.json",
		"docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
	}
	result := make(map[string][]byte, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		result[path] = data
	}
	return result
}

func copyV2AuthorInputs(t *testing.T, sourceRoot, destinationRoot string) {
	t.Helper()
	paths := []string{
		"docs/fixtures/remediation/financial-table-experiment.v1.json",
		"docs/fixtures/remediation/financial-table-experiment.v2.json",
		"docs/fixtures/remediation/financial-tables/statement-positive.pdf",
		"docs/fixtures/remediation/financial-tables/statement-damaged.pdf",
		"docs/fixtures/remediation/financial-tables/invoice-positive.pdf",
		"docs/fixtures/remediation/financial-tables/invoice-damaged.pdf",
		"docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
		"docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
	}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(sourceRoot, path))
		if err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(destinationRoot, path)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func equalByteSnapshots(left, right map[string][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for name, data := range left {
		if !bytes.Equal(data, right[name]) {
			return false
		}
	}
	return true
}

func baseFixturePath(t *testing.T, family, state string) string {
	t.Helper()
	switch family {
	case "statements":
		return "docs/fixtures/remediation/financial-tables/statement-" + state + ".pdf"
	case "invoices":
		return "docs/fixtures/remediation/financial-tables/invoice-" + state + ".pdf"
	default:
		t.Fatalf("unknown family %s", family)
		return ""
	}
}

func hexSHA(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
