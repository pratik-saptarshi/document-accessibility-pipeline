package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAuthorFinancialTableV2ReproducesTrackedBytesWithoutGOMODCACHE(t *testing.T) {
	t.Setenv("GOMODCACHE", "")
	afmPath, err := ResolvePinnedHelveticaAFM("")
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
		ctx, cancel := context.WithTimeout(context.Background(), v2GoEnvTimeout)
		defer cancel()
		resolved, err := readV2GoModuleCache(ctx)
		if err != nil {
			t.Fatal(err)
		}
		moduleCache = resolved
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

func TestReadV2GoModuleCacheRejectsTimedOutGoEnv(t *testing.T) {
	installV2FakeGo(t, "#!/bin/sh\nexec sleep 30\n")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, err := readV2GoModuleCache(ctx); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("stalled go env error = %v, want timeout classification", err)
	}
}

func TestReadV2GoModuleCacheBoundsInheritedPipesWithWaitDelay(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "child.pid")
	t.Cleanup(func() {
		pidBytes, err := os.ReadFile(pidPath)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			return
		}
		child, err := os.FindProcess(pid)
		if err == nil {
			_ = child.Kill()
		}
	})
	script := "#!/bin/sh\nsleep 30 &\necho $! > \"$V2_GO_ENV_CHILD_PID_FILE\"\nexit 0\n"
	installV2FakeGo(t, script)
	t.Setenv("V2_GO_ENV_CHILD_PID_FILE", pidPath)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	started := time.Now()
	if _, err := readV2GoModuleCache(ctx); err == nil || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("inherited go env pipe error = %v, want bounded wait failure", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("inherited go env pipe held for %s, want Cmd.WaitDelay bound", elapsed)
	}
}

func TestReadV2GoModuleCacheBoundsStdoutAndStderr(t *testing.T) {
	for _, stream := range []string{"stdout", "stderr"} {
		t.Run(stream, func(t *testing.T) {
			redirection := ""
			if stream == "stderr" {
				redirection = " >&2"
			}
			script := fmt.Sprintf("#!/bin/sh\ni=0\nwhile [ $i -lt %d ]; do printf x%s; i=$((i+1)); done\n", v2GoEnvOutputLimit+1, redirection)
			installV2FakeGo(t, script)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			output, err := readV2GoModuleCache(ctx)
			if err == nil || !strings.Contains(err.Error(), stream+" exceeded") {
				t.Fatalf("oversized go env %s error = %v, output bytes=%d, want bounded-output classification", stream, err, len(output))
			}
		})
	}
}

func TestAuthorFinancialTableV2GoEnvOverflowLeavesFixturesUntouched(t *testing.T) {
	root := filepath.Join("..", "..")
	before := snapshotV2AuthorArtifacts(t, root)
	t.Setenv("GOMODCACHE", "")
	script := fmt.Sprintf("#!/bin/sh\ni=0\nwhile [ $i -lt %d ]; do printf x; i=$((i+1)); done\n", v2GoEnvOutputLimit+1)
	installV2FakeGo(t, script)
	if err := AuthorFinancialTableV2(root, ""); err == nil || !strings.Contains(err.Error(), "stdout exceeded") {
		t.Fatalf("overflowing go env error = %v, want bounded stdout rejection", err)
	}
	after := snapshotV2AuthorArtifacts(t, root)
	if !equalByteSnapshots(before, after) {
		t.Fatal("go env output rejection changed an existing v2 fixture or manifest")
	}
}

func installV2FakeGo(t *testing.T, script string) {
	t.Helper()
	directory := t.TempDir()
	goPath := filepath.Join(directory, "go")
	if err := os.WriteFile(goPath, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
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
