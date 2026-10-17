// Package spikemeasure is a synthetic measurement spike for beads dap-5hy.3.
//
// It measures whether the separate `cloud-tools-bootstrap --verify` pass adds a
// meaningful cold/warm setup cost on top of bootstrap's own internal
// verification, in checkout (`go run .`) and no-checkout (prebuilt binary)
// flows. All fixtures are synthetic: stub shell-script tools served from a
// local TLS test server, a synthetic manifest, and a stub Go toolchain. No
// real tool downloads and no production manifests are touched.
//
// Test-first controls (always run):
//   - corrupted checksum still fails bootstrap (exit 1)
//   - wrong tool version still fails bootstrap and --verify (exit 1)
//   - current failure exit codes are locked in (bootstrap failure 1,
//     verify failure 1, flag misuse 2, clean verify 0)
//
// Timings run only with SPIKE_MEASURE_TIMINGS=1 (and never under -short);
// results are written as JSON to SPIKE_RESULTS_PATH for the research doc.
//
// Pre-registered decision rule (see docs/research/bootstrap-verify-cost-2026-10-17.md):
// the extra --verify pass justifies deduplication work only if its mean cost
// is >= 2s or >= 25% of the flow's total setup time in any measured cell.
package spikemeasure

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap/internal/cloudtools"
)

// stubSpec describes one synthetic tool. Sizes are chosen to approximate the
// verification workload of the real pinned tool set (several small CLIs plus
// a few multi-MB binaries whose SHA-256 must be hashed on every pass).
type stubSpec struct {
	name string
	size int
}

var spikeStubTools = []stubSpec{
	{name: "synth-a", size: 1024},
	{name: "synth-b", size: 2048},
	{name: "synth-c", size: 65536},
	{name: "synth-d", size: 1 << 20},
	{name: "synth-e", size: 2 << 20},
	{name: "synth-f", size: 4 << 20},
	{name: "synth-g", size: 8 << 20},
	{name: "synth-h", size: 4096},
}

const (
	spikeVersion   = "1.2.3"
	spikeGoVersion = "1.99.0"
)

// spikeFixture bundles the synthetic server, manifest, and stub toolchain.
type spikeFixture struct {
	server       *httptest.Server
	bodies       map[string][]byte // path -> served bytes
	manifest     cloudtools.Manifest
	manifestPath string
	stubGo       string
}

func stubScript(name, version string, size int) []byte {
	var b bytes.Buffer
	b.WriteString("#!/bin/sh\n")
	fmt.Fprintf(&b, "echo \"%s %s\"\n", name, version)
	line := "# synthetic padding 0123456789abcdef0123456789abcdef01234567\n"
	for b.Len() < size {
		b.WriteString(line)
	}
	out := b.Bytes()
	return out[:size]
}

func shaHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// newSpikeFixture builds a fully synthetic environment. corruptSHA, when
// non-empty, names the tool whose manifest SHA-256 is corrupted (negative
// control). wrongVersion, when non-empty, names the tool whose served bytes
// report a different version than the manifest pins (negative control); its
// manifest checksum is computed over the wrong-version bytes so the checksum
// gate passes and the version gate is what rejects it.
func newSpikeFixture(t *testing.T, corruptSHA, wrongVersion string) *spikeFixture {
	t.Helper()
	work := t.TempDir()
	fx := &spikeFixture{bodies: map[string][]byte{}}

	stubGo := filepath.Join(work, "stub-go")
	stubGoScript := "#!/bin/sh\nif [ \"${1-}\" = version ]; then printf 'go version go" + spikeGoVersion + " linux/amd64\\n'; else exit 1; fi\n"
	if err := os.WriteFile(stubGo, []byte(stubGoScript), 0700); err != nil {
		t.Fatal(err)
	}
	fx.stubGo = stubGo

	revision := "0123456789abcdef0123456789abcdef01234567"
	storageRevision := "abcdef0123456789abcdef0123456789abcdef01"
	m := cloudtools.Manifest{
		SchemaVersion: 1,
		Platform:      "linux/amd64",
		BootstrapSource: cloudtools.BootstrapSource{
			Revision:        revision,
			StorageRevision: storageRevision,
			Path:            "buildtools/source-archives/cloud-tools-bootstrap-" + revision + ".tar.gz",
			URL:             "https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/" + storageRevision + "/buildtools/source-archives/cloud-tools-bootstrap-" + revision + ".tar.gz",
			SHA256:          strings.Repeat("c", 64),
			SizeBytes:       42,
		},
		Tools: []cloudtools.Tool{{
			Name:           "go",
			Version:        spikeGoVersion,
			License:        "BSD-3-Clause",
			Source:         "https://go.dev/dl/go" + spikeGoVersion + ".linux-amd64.tar.gz",
			SHA256:         strings.Repeat("b", 64),
			ChecksumSource: "synthetic spike fixture",
			Kind:           "bootstrap-toolchain",
			Binary:         "go",
			VersionArgs:    []string{"version"},
			VersionPattern: `go version go1\.99\.0 linux/amd64`,
		}},
	}
	for _, s := range spikeStubTools {
		version := spikeVersion
		if s.name == wrongVersion {
			version = "9.9.9"
		}
		body := stubScript(s.name, version, s.size)
		path := "/stub-" + s.name
		fx.bodies[path] = body
		toolSHA := shaHex(body)
		if s.name == corruptSHA {
			toolSHA = strings.Repeat("0", 64)
		}
		tool := cloudtools.Tool{
			Name:            s.name,
			Version:         spikeVersion,
			License:         "MIT",
			Source:          "https://spike.invalid" + path,
			SHA256:          toolSHA,
			BinarySHA256:    shaHex(body),
			ChecksumSource:  "synthetic spike fixture",
			Kind:            "raw",
			Binary:          s.name,
			VersionArgs:     []string{"--version"},
			VersionPattern:  s.name + ` 1\.2\.3(?:[^0-9]|$)`,
			VersionFallback: "",
		}
		m.Tools = append(m.Tools, tool)
	}
	fx.manifest = m
	// The real URLs are replaced with the local test server once it starts.
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(work, "spike-manifest.json")
	if err := os.WriteFile(manifestPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	fx.manifestPath = manifestPath

	fx.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := fx.bodies[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(fx.server.Close)

	// Rewrite the manifest sources to the test server and persist.
	for i := range fx.manifest.Tools {
		tool := &fx.manifest.Tools[i]
		if tool.Kind == "raw" {
			tool.Source = fx.server.URL + "/stub-" + tool.Name
		}
	}
	data, err = json.Marshal(fx.manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	return fx
}

func (fx *spikeFixture) options(root, cache string) cloudtools.Options {
	return cloudtools.Options{
		Root:     root,
		Cache:    cache,
		GoBinary: fx.stubGo,
		Client:   fx.server.Client(),
		LockWait: 10 * time.Second,
		Output:   new(bytes.Buffer),
	}
}

// realGo locates a Go toolchain to build/run the bootstrap module itself.
// This builds the harness's own binary; the measured flows use only synthetic
// stub tools plus the stub Go toolchain passed as --go.
func realGo(t *testing.T) string {
	t.Helper()
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	p := filepath.Join(os.Getenv("HOME"), "workspace", "toolchains", "go", "bin", "go")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	t.Fatal("no Go toolchain found")
	return ""
}

func moduleDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		t.Fatalf("spike harness must run from cmd/cloud-tools-bootstrap/spikemeasure: %v", err)
	}
	return dir
}

func buildBootstrapBinary(t *testing.T, dir string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "cloud-tools-bootstrap")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, realGo(t), "build", "-o", out, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=-mod=readonly")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build bootstrap binary: %v\n%s", err, out)
	}
	return out
}

// runBootstrapCLI executes the compiled bootstrap binary like the no-checkout
// shell path does and returns its exit code and combined output. extraEnv is
// appended to the child environment (used to trust the synthetic TLS server).
func runBootstrapCLI(t *testing.T, binary string, extraEnv []string, args ...string) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	err := cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run bootstrap CLI: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return code, combined.String()
}

// trustServerEnv writes the synthetic TLS server's certificate to a PEM file
// and returns env vars so CLI subprocesses trust it for HTTPS downloads.
func trustServerEnv(t *testing.T, fx *spikeFixture) []string {
	t.Helper()
	pemPath := filepath.Join(t.TempDir(), "spike-ca.pem")
	f, err := os.Create(pemPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := pemEncodeCert(f, fx.server.Certificate()); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return []string{"SSL_CERT_FILE=" + pemPath}
}

func pemEncodeCert(w io.Writer, cert *x509.Certificate) error {
	return pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

// --- Timings: criterion 1 (reproducible cold/warm timings) ---

// timingCell holds wall-clock milliseconds for one flow/state combination.
type timingCell struct {
	Flow        string    `json:"flow"`
	State       string    `json:"state"`
	BootstrapMs []float64 `json:"bootstrap_ms"`
	VerifyMs    []float64 `json:"verify_ms"`
}

type timingReport struct {
	Timestamp string       `json:"timestamp"`
	GoVersion string       `json:"go_version"`
	Platform  string       `json:"platform"`
	Cells     []timingCell `json:"cells"`
}

func timeRun(t *testing.T, name string, dir string, extraEnv []string, argv ...string) float64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", name, err, combined.String())
	}
	return float64(elapsed) / float64(time.Millisecond)
}

// TestSpikeVerifyCost_MeasureColdWarm records reproducible cold/warm timings
// for bootstrap vs bootstrap+--verify in checkout and no-checkout flows.
// Runs only with SPIKE_MEASURE_TIMINGS=1 (never under -short); results JSON
// goes to SPIKE_RESULTS_PATH for the research doc.
func TestSpikeVerifyCost_MeasureColdWarm(t *testing.T) {
	if testing.Short() {
		t.Skip("timing measurement skipped under -short")
	}
	if os.Getenv("SPIKE_MEASURE_TIMINGS") != "1" {
		t.Skip("timing measurement runs only with SPIKE_MEASURE_TIMINGS=1")
	}
	resultsPath := os.Getenv("SPIKE_RESULTS_PATH")
	if resultsPath == "" {
		t.Fatal("SPIKE_RESULTS_PATH must name the JSON results file")
	}
	fx := newSpikeFixture(t, "", "")
	trustEnv := trustServerEnv(t, fx)
	dir := moduleDir(t)
	goBin := realGo(t)
	bootstrapBin := buildBootstrapBinary(t, dir)

	out, err := exec.Command(goBin, "version").Output()
	if err != nil {
		t.Fatal(err)
	}

	const coldIters = 3
	const warmIters = 5
	var cells []timingCell
	for _, flow := range []string{"checkout", "nocheckout"} {
		for _, state := range []string{"cold", "warm"} {
			cell := timingCell{Flow: flow, State: state}
			iters := coldIters
			if state == "warm" {
				iters = warmIters
			}
			var root, cache string
			for i := 0; i < iters; i++ {
				iterWork := t.TempDir()
				if state == "cold" || i == 0 {
					root = filepath.Join(iterWork, "tools")
					cache = filepath.Join(iterWork, "cache")
				}
				manifestArgs := []string{"--manifest", fx.manifestPath, "--root", root, "--cache", cache, "--go", fx.stubGo}
				var bMs, vMs float64
				if flow == "checkout" {
					goEnv := append([]string{
						"GOTOOLCHAIN=local",
						"GOWORK=off",
						"GOFLAGS=-mod=readonly",
						"GOMODCACHE=" + filepath.Join(cache, "go", "mod"),
						"GOCACHE=" + filepath.Join(cache, "go", "build"),
					}, trustEnv...)
					bMs = timeRun(t, fmt.Sprintf("%s/%s bootstrap #%d", flow, state, i+1), dir, goEnv,
						append([]string{goBin, "run", "."}, manifestArgs...)...)
					vMs = timeRun(t, fmt.Sprintf("%s/%s verify #%d", flow, state, i+1), dir, goEnv,
						append([]string{goBin, "run", "."}, append(manifestArgs, "--verify")...)...)
				} else {
					bMs = timeRun(t, fmt.Sprintf("%s/%s bootstrap #%d", flow, state, i+1), "", trustEnv,
						append([]string{bootstrapBin}, manifestArgs...)...)
					vMs = timeRun(t, fmt.Sprintf("%s/%s verify #%d", flow, state, i+1), "", trustEnv,
						append([]string{bootstrapBin}, append(manifestArgs, "--verify")...)...)
				}
				cell.BootstrapMs = append(cell.BootstrapMs, bMs)
				cell.VerifyMs = append(cell.VerifyMs, vMs)
			}
			cells = append(cells, cell)
		}
	}

	report := timingReport{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		GoVersion: strings.TrimSpace(string(out)),
		Platform:  "linux/amd64",
		Cells:     cells,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resultsPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	var summary strings.Builder
	fmt.Fprintf(&summary, "%-12s %-6s %3s | %12s %12s | %10s %10s | %s\n",
		"flow", "state", "n", "bootstrap ms", "verify ms", "verify shr", "verdict*", "")
	for _, c := range cells {
		bMean, bMax := meanMax(c.BootstrapMs)
		vMean, vMax := meanMax(c.VerifyMs)
		share := vMean / (bMean + vMean) * 100
		verdict := "not meaningful"
		if vMean >= 2000 || share >= 25 {
			verdict = "MEANINGFUL?"
		}
		fmt.Fprintf(&summary, "%-12s %-6s %3d | %6.0f/%6.0f %6.0f/%6.0f | %9.1f%% %10s\n",
			c.Flow, c.State, len(c.BootstrapMs), bMean, bMax, vMean, vMax, share, verdict)
	}
	summary.WriteString("* meaningful = verify mean >= 2000ms or >= 25% of flow total (pre-registered rule)\n")
	t.Log("\n" + summary.String())
}

func meanMax(xs []float64) (mean, max float64) {
	for _, x := range xs {
		mean += x
		if x > max {
			max = x
		}
	}
	if len(xs) > 0 {
		mean /= float64(len(xs))
	}
	return mean, max
}

// --- Test-first controls: criterion 2 (checksum/version validation preserved) ---

func TestSpikeVerifyCost_CorruptedChecksumFailsBootstrap(t *testing.T) {
	fx := newSpikeFixture(t, "synth-d", "")
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")

	// Library-level: Bootstrap must reject the corrupted entry.
	err := cloudtools.Bootstrap(context.Background(), fx.manifest, fx.options(root, cache))
	if err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("Bootstrap error = %v; want SHA-256 mismatch rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "bin", "synth-d")); !os.IsNotExist(statErr) {
		t.Fatalf("corrupted tool was installed: %v", statErr)
	}

	// CLI-level: failure exit code must be 1 (unchanged contract).
	binary := buildBootstrapBinary(t, moduleDir(t))
	code, _ := runBootstrapCLI(t, binary, trustServerEnv(t, fx),
		"--manifest", fx.manifestPath, "--root", root, "--cache", cache, "--go", fx.stubGo)
	if code != 1 {
		t.Fatalf("bootstrap CLI exit = %d; want 1 on checksum failure", code)
	}
}

func TestSpikeVerifyCost_WrongVersionFailsBootstrapAndVerify(t *testing.T) {
	fx := newSpikeFixture(t, "", "synth-c")
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")

	// Library-level: checksum passes, version validation must reject.
	err := cloudtools.Bootstrap(context.Background(), fx.manifest, fx.options(root, cache))
	if err == nil || !strings.Contains(err.Error(), "version validation") {
		t.Fatalf("Bootstrap error = %v; want version validation rejection", err)
	}

	// CLI-level: bootstrap a good install, corrupt one installed binary's
	// version output, and confirm --verify still fails with exit 1.
	good := newSpikeFixture(t, "", "")
	binary := buildBootstrapBinary(t, moduleDir(t))
	env := trustServerEnv(t, good)
	code, out := runBootstrapCLI(t, binary, env,
		"--manifest", good.manifestPath, "--root", root, "--cache", cache, "--go", good.stubGo)
	if code != 0 {
		t.Fatalf("good bootstrap CLI exit = %d, want 0\n%s", code, out)
	}
	badBody := stubScript("synth-a", "9.9.9", 1024)
	if err := os.WriteFile(filepath.Join(root, "bin", "synth-a"), badBody, 0755); err != nil {
		t.Fatal(err)
	}
	code, out = runBootstrapCLI(t, binary, env,
		"--manifest", good.manifestPath, "--root", root, "--cache", cache, "--go", good.stubGo, "--verify")
	if code != 1 {
		t.Fatalf("--verify CLI exit = %d; want 1 on wrong installed version\n%s", code, out)
	}
	if !strings.Contains(out, "NOT VERIFIED") {
		t.Fatalf("--verify output missing NOT VERIFIED marker:\n%s", out)
	}
}

func TestSpikeVerifyCost_ExitCodesPreserved(t *testing.T) {
	fx := newSpikeFixture(t, "", "")
	binary := buildBootstrapBinary(t, moduleDir(t))
	env := trustServerEnv(t, fx)
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	args := []string{"--manifest", fx.manifestPath, "--root", root, "--cache", cache, "--go", fx.stubGo}

	if code, out := runBootstrapCLI(t, binary, env, args...); code != 0 {
		t.Fatalf("clean bootstrap exit = %d; want 0\n%s", code, out)
	}
	if code, out := runBootstrapCLI(t, binary, env, append(args, "--verify")...); code != 0 {
		t.Fatalf("clean --verify exit = %d; want 0\n%s", code, out)
	}
	// Flag misuse (missing required --go) must stay exit 2.
	misuse := []string{"--manifest", fx.manifestPath, "--root", root, "--cache", cache}
	if code, _ := runBootstrapCLI(t, binary, env, misuse...); code != 2 {
		t.Fatalf("flag-misuse exit = %d; want 2", code)
	}
	// --verify against an empty root must fail with exit 1.
	emptyRoot := filepath.Join(t.TempDir(), "empty-tools")
	emptyArgs := []string{"--manifest", fx.manifestPath, "--root", emptyRoot, "--cache", cache, "--go", fx.stubGo, "--verify"}
	if code, _ := runBootstrapCLI(t, binary, env, emptyArgs...); code != 1 {
		t.Fatalf("--verify on empty root exit = %d; want 1", code)
	}
}
