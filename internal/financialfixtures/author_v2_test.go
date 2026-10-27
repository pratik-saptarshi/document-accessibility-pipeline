package financialfixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAuthorFinancialTableV2ReproducesTrackedBytesWithoutGOMODCACHE(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		moduleCache = testV2ModuleCache(t)
	}
	afmPath, err := ResolvePinnedHelveticaAFM(moduleCache)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOMODCACHE", "")
	resolvedVersion, err := v2ResolvedPDFCPUVersion()
	if err != nil {
		t.Fatal(err)
	}
	wantSegment := "github.com/pdfcpu/pdfcpu@" + resolvedVersion + "/internal/corefont/Core14_AFMs/Helvetica.afm"
	if !strings.Contains(filepath.ToSlash(afmPath), wantSegment) {
		t.Fatalf("resolved AFM path = %s, want segment %s", afmPath, wantSegment)
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

func TestResolvePinnedHelveticaAFMUsesCustomGOMODCACHE(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	source, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	customCache := t.TempDir()
	stubV2ResolvedPDFCPUVersion(t, "v0.0.0-testcache")
	customAFM := filepath.Join(customCache, "github.com/pdfcpu/pdfcpu@v0.0.0-testcache/internal/corefont/Core14_AFMs/Helvetica.afm")
	if err := os.MkdirAll(filepath.Dir(customAFM), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(customAFM, afmdat, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOMODCACHE", customCache)
	resolved, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	if resolved != customAFM {
		t.Fatalf("resolved AFM path = %s, want custom cache path %s", resolved, customAFM)
	}
}

func TestResolvePinnedHelveticaAFMUsesGoEnvWhenGOMODCACHEUnset(t *testing.T) {
	requireV2ProcessGroupSupport(t)
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
	source, err := ResolvePinnedHelveticaAFM(moduleCache)
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	cache := filepath.Join(temp, "module-cache")
	stubV2ResolvedPDFCPUVersion(t, "v0.0.0-testcache")
	afmpath := filepath.Join(cache, "github.com/pdfcpu/pdfcpu@v0.0.0-testcache/internal/corefont/Core14_AFMs/Helvetica.afm")
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

// stubV2ResolvedPDFCPUVersion pins the resolver's version source for the
// duration of a test, so fake module-cache trees exercise resolution
// deterministically whatever pdfcpu version the test binary was built with.
func stubV2ResolvedPDFCPUVersion(t *testing.T, version string) {
	t.Helper()
	stubV2ResolvedPDFCPUVersionFunc(t, func() (string, error) { return version, nil })
}

func stubV2ResolvedPDFCPUVersionFunc(t *testing.T, fn func() (string, error)) {
	t.Helper()
	original := v2ResolvedPDFCPUVersion
	v2ResolvedPDFCPUVersion = fn
	t.Cleanup(func() { v2ResolvedPDFCPUVersion = original })
}

func writeV2FakeCachedAFM(t *testing.T, cache, version string, afmdat []byte) string {
	t.Helper()
	afmPath := filepath.Join(cache, "github.com/pdfcpu/pdfcpu@"+version+"/internal/corefont/Core14_AFMs/Helvetica.afm")
	if err := os.MkdirAll(filepath.Dir(afmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(afmPath, afmdat, 0o600); err != nil {
		t.Fatal(err)
	}
	return afmPath
}

func TestPDFCPUVersionFromBuildInfoResult(t *testing.T) {
	tests := []struct {
		name    string
		info    *debug.BuildInfo
		ok      bool
		want    string
		wantErr string
	}{
		{
			name: "resolved version from dependency list",
			info: &debug.BuildInfo{Deps: []*debug.Module{
				{Path: "example.com/other", Version: "v1.0.0"},
				{Path: v2PDFCPUModulePath, Version: "v9.9.9"},
			}},
			ok:   true,
			want: "v9.9.9",
		},
		{
			name:    "build info unavailable",
			ok:      false,
			wantErr: "build info is unavailable",
		},
		{
			name:    "module absent from dependency list",
			info:    &debug.BuildInfo{Deps: []*debug.Module{{Path: "example.com/other", Version: "v1.0.0"}}},
			ok:      true,
			wantErr: "not in the build dependency list",
		},
		{
			name: "replaced module is rejected",
			info: &debug.BuildInfo{Deps: []*debug.Module{{
				Path:    v2PDFCPUModulePath,
				Version: "v9.9.9",
				Replace: &debug.Module{Path: "../pdfcpu", Version: ""},
			}}},
			ok:      true,
			wantErr: "replaced by",
		},
		{
			name:    "empty version is rejected",
			info:    &debug.BuildInfo{Deps: []*debug.Module{{Path: v2PDFCPUModulePath}}},
			ok:      true,
			wantErr: "empty version",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := pdfcpuVersionFromBuildInfoResult(test.info, test.ok)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("version error = %v, want substring %q", err, test.wantErr)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("version = %q, err = %v, want %q", got, err, test.want)
			}
		})
	}
	// The function the resolver actually uses must resolve this test
	// binary's pdfcpu version from its real build info.
	if version, err := pdfcpuVersionFromBuildInfo(); err != nil || version == "" {
		t.Fatalf("pdfcpuVersionFromBuildInfo = %q, %v; want a resolved version", version, err)
	}
}

func TestResolvePinnedHelveticaAFMFallsBackToUniqueCachedVersion(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	source, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	stubV2ResolvedPDFCPUVersionFunc(t, func() (string, error) {
		return "", errors.New("build info is unavailable")
	})
	cache := t.TempDir()
	want := writeV2FakeCachedAFM(t, cache, "v8.8.8-fallback", afmdat)
	resolved, err := ResolvePinnedHelveticaAFM(cache)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolved AFM path = %s, want unique cached fallback %s", resolved, want)
	}
}

func TestResolvePinnedHelveticaAFMFallbackAcceptsMultipleIdenticalCachedVersions(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	source, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	stubV2ResolvedPDFCPUVersionFunc(t, func() (string, error) {
		return "", errors.New("build info is unavailable")
	})
	cache := t.TempDir()
	want := writeV2FakeCachedAFM(t, cache, "v0.16.2", afmdat)
	writeV2FakeCachedAFM(t, cache, "v0.16.3", afmdat)
	resolved, err := ResolvePinnedHelveticaAFM(cache)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolved AFM path = %s, want lexically first matching cached AFM %s", resolved, want)
	}
}

func TestResolvePinnedHelveticaAFMFallbackSkipsHashMismatchingAFM(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	source, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	// Same AFM shape and widths, different bytes: the hash must exclude it.
	mismatched := append([]byte(nil), afmdat...)
	mismatched[0] = 'X'
	if sha256Hex(mismatched) == v2AFMSHA256 {
		t.Fatal("mismatched control unexpectedly matches the pinned AFM hash")
	}
	stubV2ResolvedPDFCPUVersionFunc(t, func() (string, error) {
		return "", errors.New("build info is unavailable")
	})
	cache := t.TempDir()
	writeV2FakeCachedAFM(t, cache, "v0.16.2", mismatched)
	want := writeV2FakeCachedAFM(t, cache, "v0.16.3", afmdat)
	resolved, err := ResolvePinnedHelveticaAFM(cache)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolved AFM path = %s, want the hash-matching cached AFM %s", resolved, want)
	}
}

func TestResolvePinnedHelveticaAFMFallbackRejectsOnlyMismatchingAFMs(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	source, err := ResolvePinnedHelveticaAFM("")
	if err != nil {
		t.Fatal(err)
	}
	afmdat, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	mismatched := append([]byte(nil), afmdat...)
	mismatched[0] = 'X'
	stubV2ResolvedPDFCPUVersionFunc(t, func() (string, error) {
		return "", errors.New("build info is unavailable")
	})
	cache := t.TempDir()
	writeV2FakeCachedAFM(t, cache, "v0.16.2", mismatched)
	writeV2FakeCachedAFM(t, cache, "v0.16.3", mismatched)
	if _, err := ResolvePinnedHelveticaAFM(cache); err == nil || !strings.Contains(err.Error(), "none matching pinned SHA-256") {
		t.Fatalf("only-mismatching-cache error = %v, want pinned-content fallback rejection", err)
	}
}

func TestReadV2GoModuleCacheRejectsTimedOutGoEnv(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	installV2FakeGo(t, "#!/bin/sh\nexec sleep 30\n")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	if _, err := readV2GoModuleCache(ctx); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("stalled go env error = %v, want timeout classification", err)
	}
}

func TestReadV2GoModuleCacheBoundsInheritedPipesWithWaitDelay(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	script := "#!/bin/sh\nsleep 1 &\nexit 0\n"
	installV2FakeGo(t, script)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	started := time.Now()
	if _, err := readV2GoModuleCache(ctx); err == nil || !strings.Contains(err.Error(), "empty path") {
		t.Fatalf("inherited go env pipe error = %v, want drained pipe and empty-path classification", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("inherited go env pipe held for %s, want Cmd.WaitDelay bound", elapsed)
	}
	// The short-lived helper descendant exits naturally, so this test never
	// signals a stale numeric PID after the process-group runner has reaped it.
	if remaining := time.Until(started.Add(time.Second)); remaining > 0 {
		time.Sleep(remaining)
	}
}

func TestReadV2GoModuleCacheBoundsStdoutAndStderr(t *testing.T) {
	requireV2ProcessGroupSupport(t)
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
	requireV2ProcessGroupSupport(t)
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
	requireV2ProcessGroupSupport(t)
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
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		moduleCache = testV2ModuleCache(t)
	}
	afmPath, err := ResolvePinnedHelveticaAFM(moduleCache)
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

func TestAuthorFinancialTableV2RejectsAmbiguousBaseManifestKeys(t *testing.T) {
	requireV2ProcessGroupSupport(t)
	repoRoot := filepath.Join("..", "..")
	moduleCache := os.Getenv("GOMODCACHE")
	if moduleCache == "" {
		moduleCache = testV2ModuleCache(t)
	}
	afmpath, err := ResolvePinnedHelveticaAFM(moduleCache)
	if err != nil {
		t.Fatal(err)
	}

	fields := []string{
		"cases", "family", "positive_path", "positive_sha256", "damaged_path",
		"damaged_sha256", "target_header", "target_cells", "damaged_wrong_header", "reading_order",
	}
	tests := make([]struct {
		name      string
		wantError string
		mutate    func([]byte) []byte
	}, 0, len(fields)+1)
	for _, field := range fields {
		field := field
		alias := strings.ToUpper(field)
		tests = append(tests, struct {
			name      string
			wantError string
			mutate    func([]byte) []byte
		}{
			name:      "standalone case-variant " + field,
			wantError: fmt.Sprintf("does not exactly match schema field %q", field),
			mutate: func(data []byte) []byte {
				return bytes.Replace(data, []byte(`"`+field+`":`), []byte(`"`+alias+`":`), 1)
			},
		})
	}
	tests = append(tests, struct {
		name      string
		wantError string
		mutate    func([]byte) []byte
	}{
		name:      "duplicate exact schema key",
		wantError: `duplicate or case-variant key "family"`,
		mutate: func(data []byte) []byte {
			return bytes.Replace(data, []byte(`"family": "statements"`), []byte(`"family": "statements", "family": "statements"`), 1)
		},
	})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tempRoot := t.TempDir()
			copyV2AuthorInputs(t, repoRoot, tempRoot)
			manifestPath := filepath.Join(tempRoot, "docs/fixtures/remediation/financial-table-experiment.v1.json")
			original, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			malformed := test.mutate(original)
			if bytes.Equal(malformed, original) {
				t.Fatal("test mutation did not match the v1 manifest")
			}
			if err := os.WriteFile(manifestPath, malformed, 0o644); err != nil {
				t.Fatal(err)
			}
			before := snapshotV2AuthorArtifacts(t, tempRoot)
			if err := AuthorFinancialTableV2(tempRoot, afmpath); err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("authoring error = %v, want %q", err, test.wantError)
			}
			after := snapshotV2AuthorArtifacts(t, tempRoot)
			if !equalByteSnapshots(before, after) {
				t.Fatal("ambiguous v1 manifest modified v2 fixture or manifest bytes")
			}
			got, err := os.ReadFile(manifestPath)
			if err != nil || !bytes.Equal(got, malformed) {
				t.Fatalf("v1 manifest bytes changed after rejection: err=%v", err)
			}
		})
	}
}

func testV2ModuleCache(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), v2GoEnvTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "env", "GOMODCACHE")
	command.Env = v2GoEnvEnvironment()
	output, err := runBoundedTestV2Command(ctx, command)
	if err != nil {
		t.Fatalf("resolve test Go module cache: %v", err)
	}
	moduleCache := strings.TrimSpace(output)
	if moduleCache == "" {
		t.Fatal("test Go module cache path is empty")
	}
	return moduleCache
}

func runBoundedTestV2Command(ctx context.Context, command *exec.Cmd) (string, error) {
	stdout := &limitedV2Output{limit: v2GoEnvOutputLimit}
	stderr := &limitedV2Output{limit: v2GoEnvOutputLimit}
	command.WaitDelay = v2GoEnvWaitDelay
	command.Stdout = stdout
	command.Stderr = stderr
	runErr := command.Run()
	if ctx.Err() != nil {
		return stdout.String(), fmt.Errorf("test Go command timed out or was canceled: %w", ctx.Err())
	}
	if stdout.Overflow() {
		return stdout.String(), fmt.Errorf("test Go command stdout exceeded %d bytes", v2GoEnvOutputLimit)
	}
	if stderr.Overflow() {
		return stdout.String(), fmt.Errorf("test Go command stderr exceeded %d bytes", v2GoEnvOutputLimit)
	}
	if runErr != nil {
		return stdout.String(), fmt.Errorf("test Go command failed: %w", runErr)
	}
	return stdout.String(), nil
}

func TestRunBoundedTestV2CommandCapsOutputAndInheritedPipes(t *testing.T) {
	for _, mode := range []string{"stdout", "stderr", "inherited-pipe"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestV2TestCommandHelper$")
			command.Env = append(os.Environ(), "V2_TEST_COMMAND_HELPER_MODE="+mode)
			started := time.Now()
			_, err := runBoundedTestV2Command(ctx, command)
			if err == nil {
				t.Fatal("bounded test Go command unexpectedly succeeded")
			}
			expected := map[string]string{
				"stdout":         "stdout exceeded",
				"stderr":         "stderr exceeded",
				"inherited-pipe": "WaitDelay",
			}[mode]
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("bounded test Go command error = %v, want %q", err, expected)
			}
			if elapsed := time.Since(started); elapsed > 2500*time.Millisecond {
				t.Fatalf("bounded test Go command took %s, want output/WaitDelay limit", elapsed)
			}
			if mode == "inherited-pipe" {
				// The holder exits naturally; no raw PID is signaled after reap.
				if remaining := time.Until(started.Add(3 * time.Second)); remaining > 0 {
					time.Sleep(remaining)
				}
			}
		})
	}
}

func TestV2TestCommandHelper(t *testing.T) {
	switch os.Getenv("V2_TEST_COMMAND_HELPER_MODE") {
	case "":
		return
	case "stdout":
		_, _ = os.Stdout.Write([]byte(strings.Repeat("x", v2GoEnvOutputLimit+1)))
	case "stderr":
		_, _ = os.Stderr.Write([]byte(strings.Repeat("x", v2GoEnvOutputLimit+1)))
	case "inherited-pipe":
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		child := exec.Command(executable, "-test.run=^TestV2TestCommandHelper$")
		child.Env = append(os.Environ(), "V2_TEST_COMMAND_HELPER_MODE=inherited-pipe-holder")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			t.Fatal(err)
		}
	case "inherited-pipe-holder":
		time.Sleep(1500 * time.Millisecond)
	default:
		t.Fatalf("unknown helper mode %q", os.Getenv("V2_TEST_COMMAND_HELPER_MODE"))
	}
}

func TestIncompleteReapDoesNotInspectOutputStillOwnedByWait(t *testing.T) {
	tests := []struct {
		name string
		run  func(chan struct{}, v2BoundedProcessRunner) error
	}{
		{
			name: "go module cache",
			run: func(done chan struct{}, runner v2BoundedProcessRunner) error {
				_, err := readV2GoModuleCacheWithRunner(context.Background(), runner)
				<-done
				return err
			},
		},
		{
			name: "observation tool",
			run: func(done chan struct{}, runner v2BoundedProcessRunner) error {
				_, _, err := runV2ObservationCommandWithRunner(context.Background(), "helper", nil, 1, runner)
				<-done
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			err := tt.run(done, func(_ context.Context, _ string, _ []string, _ []string, stdout, stderr io.Writer, _ time.Duration) error {
				go func() {
					defer close(done)
					_, _ = stdout.Write([]byte("xx")) // overflow before returning the failure
					for i := 0; i < 100_000; i++ {
						_, _ = stdout.Write([]byte("y"))
						if i%16 == 0 {
							runtime.Gosched()
						}
					}
				}()
				// Wait until the first overflow is recorded, then return while the
				// output-copy goroutine still owns the writer.
				for !stdout.(*limitedV2Output).Overflow() {
					runtime.Gosched()
				}
				return errV2HelperReapTimeout
			})
			if !errors.Is(err, errV2HelperReapTimeout) {
				t.Fatalf("incomplete-reap error = %v, want reap-timeout sentinel", err)
			}
			if !strings.Contains(err.Error(), "may still be active") || !strings.Contains(err.Error(), "incomplete") {
				t.Fatalf("incomplete-reap error = %v, want honest active-process/output detail", err)
			}
		})
	}
}

func TestLimitedV2OutputSupportsConcurrentSnapshots(t *testing.T) {
	output := &limitedV2Output{limit: 64}
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 100_000; i++ {
			_, _ = output.Write([]byte("x"))
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 100_000; i++ {
			_ = output.String()
			_ = output.Overflow()
		}
	}()
	close(start)
	workers.Wait()
}

func TestFreezeV2ProcessOutputStopsMutationAfterReapTimeout(t *testing.T) {
	stdout := &limitedV2Output{limit: 16}
	stderr := &limitedV2Output{limit: 16}
	_, _ = stdout.Write([]byte("stdout"))
	_, _ = stderr.Write([]byte("stderr"))
	if stdout.Overflow() || stderr.Overflow() {
		t.Fatal("initial output unexpectedly exceeded its limit")
	}
	if err := freezeV2ProcessOutputAfterReapTimeout(stdout, stderr, errV2HelperReapTimeout); !errors.Is(err, errV2HelperReapTimeout) {
		t.Fatalf("reap error = %v, want reap-timeout sentinel", err)
	}
	stdoutBefore, stderrBefore := stdout.String(), stderr.String()
	stdoutOverflowBefore, stderrOverflowBefore := stdout.Overflow(), stderr.Overflow()
	var writers sync.WaitGroup
	writers.Add(2)
	go func() {
		defer writers.Done()
		for i := 0; i < 1000; i++ {
			_, _ = stdout.Write([]byte("x"))
		}
	}()
	go func() {
		defer writers.Done()
		for i := 0; i < 1000; i++ {
			_, _ = stderr.Write([]byte("x"))
		}
	}()
	writers.Wait()
	if stdout.String() != stdoutBefore || stderr.String() != stderrBefore {
		t.Fatal("frozen output accepted bytes after bounded reap timeout")
	}
	if stdout.Overflow() != stdoutOverflowBefore || stderr.Overflow() != stderrOverflowBefore {
		t.Fatal("frozen output changed after bounded reap timeout")
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
