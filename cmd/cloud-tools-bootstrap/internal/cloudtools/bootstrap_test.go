package cloudtools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBootstrapChecksChecksumBeforeInstallAndCleansTemporaryFiles(t *testing.T) {
	body := []byte("#!/bin/sh\necho testtool 1.2.3\n")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer server.Close()
	m := testManifest(server.URL+"/tool", strings.Repeat("0", 64), "raw")
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	err := Bootstrap(context.Background(), m, Options{Root: root, Cache: cache, GoBinary: os.Args[0], Client: server.Client(), LockWait: time.Second})
	if err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("Bootstrap error = %v; want checksum failure", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "fixture")); !os.IsNotExist(err) {
		t.Fatalf("binary exists after failed checksum: %v", err)
	}
	if entries, _ := filepath.Glob(filepath.Join(cache, "downloads", ".download-*")); len(entries) != 0 {
		t.Fatalf("temporary downloads remain: %v", entries)
	}
}

func TestBootstrapIsIdempotentAndVerifiesVersion(t *testing.T) {
	body := []byte("#!/bin/sh\necho testtool 1.2.3\n")
	digest := sha256.Sum256(body)
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); _, _ = w.Write(body) }))
	defer server.Close()
	m := testManifest(server.URL+"/tool", hex.EncodeToString(digest[:]), "raw")
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	o := Options{Root: root, Cache: cache, GoBinary: os.Args[0], Client: server.Client(), LockWait: time.Second}
	for i := 0; i < 2; i++ {
		if err := Bootstrap(context.Background(), m, o); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("asset requests = %d; want one for two runs", requests.Load())
	}
	if _, err := os.Stat(filepath.Join(root, "activate.sh")); err != nil {
		t.Fatalf("activation file missing: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "bin", "fixture"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("installed bytes changed: %q", got)
	}
}

func TestBootstrapRejectsWrongCLIVersionEvenWithCorrectChecksum(t *testing.T) {
	body := []byte("#!/bin/sh\necho testtool 9.9.9\n")
	digest := sha256.Sum256(body)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer server.Close()
	m := testManifest(server.URL+"/tool", hex.EncodeToString(digest[:]), "raw")
	root := filepath.Join(t.TempDir(), "tools")
	err := Bootstrap(context.Background(), m, Options{Root: root, Cache: filepath.Join(t.TempDir(), "cache"), GoBinary: os.Args[0], Client: server.Client(), LockWait: time.Second})
	if err == nil || !strings.Contains(err.Error(), "version validation") {
		t.Fatalf("Bootstrap error = %v; want version check failure", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "fixture")); !os.IsNotExist(err) {
		t.Fatalf("wrong-version binary was published: %v", err)
	}
}

func TestBlockedRepairDoesNotReportPreexistingWrongVersionAsVerified(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer server.Close()
	root := filepath.Join(t.TempDir(), "tools")
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "bin", "fixture")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho testtool 9.9.9\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(path, "--version").CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "testtool 9.9.9" {
		t.Fatalf("wrong-version fixture did not execute as intended: output=%q err=%v", out, err)
	}
	m := testManifest(server.URL+"/fixture", strings.Repeat("a", 64), "raw")
	o := Options{Root: root, Cache: filepath.Join(t.TempDir(), "cache"), GoBinary: os.Args[0], Client: server.Client(), LockWait: time.Second}
	if err := Bootstrap(context.Background(), m, o); err == nil {
		t.Fatal("expected blocked repair download")
	}
	results, err := VerifyInstalled(context.Background(), m, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].OK || results[0].Path != path || !strings.Contains(results[0].Status, "invalid") {
		t.Fatalf("wrong-version executable was reported ready: %#v", results)
	}
}

func TestTarExtractionRequiresSafeRegularMember(t *testing.T) {
	makeArchive := func(name string, typ byte) string {
		path := filepath.Join(t.TempDir(), "asset.tgz")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		gz := gzip.NewWriter(f)
		tw := tar.NewWriter(gz)
		size := int64(0)
		if typ == tar.TypeReg {
			size = 3
		}
		h := &tar.Header{Name: name, Mode: 0755, Size: size, Typeflag: typ}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			_, _ = tw.Write([]byte("abc"))
		}
		_ = tw.Close()
		_ = gz.Close()
		_ = f.Close()
		return path
	}
	if err := extractTarBinary(makeArchive("../escape", tar.TypeReg), "escape", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("accepted traversal path")
	}
	if err := extractTarBinary(makeArchive("link", tar.TypeSymlink), "link", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Fatal("accepted symlink member")
	}
	got := filepath.Join(t.TempDir(), "out")
	if err := extractTarBinary(makeArchive("sub/tool", tar.TypeReg), "tool", got); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(got)
	if err != nil || string(contents) != "abc" {
		t.Fatalf("extracted contents = %q, %v", contents, err)
	}
}

func TestConcurrentBootstrapTimesOutOnOwnedLock(t *testing.T) {
	root := t.TempDir()
	lock := filepath.Join(root, ".bootstrap.lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	_, err := acquireLock(context.Background(), lock, 50*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("lock error = %v", err)
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatal("lock wait returned before timeout")
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatal("waiter removed lock owned by another process")
	}
}

func TestCommandCancellationKillsDescendantHoldingCombinedOutputPipe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := filepath.Join(t.TempDir(), "child-started")
	cmd := commandContext(ctx, "sh", "-c", `sleep 1 & echo ready > "$1"; wait`, "sh", ready)
	done := make(chan error, 1)
	go func() {
		_, err := cmd.CombinedOutput()
		done <- err
	}()
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("descendant did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	started := time.Now()
	cancel()
	select {
	case <-done:
		if time.Since(started) > 500*time.Millisecond {
			t.Fatalf("cancellation waited for descendant-held pipe: %s", time.Since(started))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("command did not return promptly after cancellation")
	}
}

func TestActivationCanBeSourcedRepeatedlyWithoutChangingPATHAgain(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"bin", "toolchains/go/1.27.2/bin"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	activation := filepath.Join(root, "activate.sh")
	if err := writeActivation(activation, nil); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", `source "$1"; first=$PATH; source "$1"; test "$PATH" = "$first"`, "bash", activation)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("activation failed or duplicated PATH: %s %v", out, err)
	}
}

func TestBootstrapCachesVeraPDFInstallerWithoutRuntimeOrValidation(t *testing.T) {
	archive := []byte("synthetic authenticated installer fixture")
	archiveHash := sha256.Sum256(archive)
	m := testVeraPDFManifest(hex.EncodeToString(archiveHash[:]))
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	provisioner := filepath.Join(t.TempDir(), "provisioner.sh")
	callLog := filepath.Join(t.TempDir(), "calls")
	script := "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$1\" >> \"$CALL_LOG\"\n[ \"$1\" = --cache-only ] || exit 41\nmkdir -p \"$(dirname -- \"$2\")\"\nprintf '%s' \"$ARCHIVE_CONTENT\" > \"$2\"\n"
	if err := os.WriteFile(provisioner, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CALL_LOG", callLog)
	t.Setenv("ARCHIVE_CONTENT", string(archive))
	o := Options{Root: root, Cache: cache, GoBinary: os.Args[0], VeraPDFProvisioner: provisioner, LockWait: time.Second}
	var output bytes.Buffer
	o.Output = &output
	for i := 0; i < 2; i++ {
		if err := Bootstrap(context.Background(), m, o); err != nil {
			t.Fatalf("bootstrap run %d: %v", i+1, err)
		}
	}
	artifactPath := veraPDFArtifactPath(root, m.Tools[0])
	if got, err := os.ReadFile(artifactPath); err != nil || !bytes.Equal(got, archive) {
		t.Fatalf("cached installer = %q, %v; want verified archive", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "verapdf")); !os.IsNotExist(err) {
		t.Fatalf("veraPDF executable was installed in PATH: %v", err)
	}
	activation, err := os.ReadFile(filepath.Join(root, "activate.sh"))
	if err != nil || !strings.Contains(string(activation), "export VERAPDF_INSTALLER_ARCHIVE=") || strings.Contains(string(activation), "bin/verapdf") {
		t.Fatalf("activation does not expose only the persistent installer artifact: %s, %v", activation, err)
	}
	callData, err := os.ReadFile(callLog)
	if err != nil || strings.TrimSpace(string(callData)) != "--cache-only" {
		t.Fatalf("provisioner invocations = %q, %v; want one cache-only call", callData, err)
	}
	results, err := VerifyInstalled(context.Background(), m, o)
	if err != nil || len(results) != 1 || !results[0].OK || !results[0].ArtifactIntegrityVerified || results[0].RuntimeReady || results[0].ValidationExecuted {
		t.Fatalf("artifact/runtime/execution statuses were conflated: %#v, %v", results, err)
	}
	if !strings.Contains(results[0].Status, "runtime not checked") || !strings.Contains(results[0].Status, "validation not run") {
		t.Fatalf("status does not distinguish readiness from artifact integrity: %#v", results[0])
	}
	if !strings.Contains(string(activation), filepath.Base(artifactPath)) || !strings.Contains(string(activation), m.Tools[0].InstallerSHA256) {
		t.Fatalf("activation did not expose the pinned archive path and digest: %s", activation)
	}
	if strings.Contains(output.String(), "validated PDF") || strings.Contains(output.String(), "runtime ready") {
		t.Fatalf("bootstrap claimed execution or runtime readiness: %s", output.String())
	}

	if err := os.WriteFile(artifactPath, []byte("tampered installer"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err = VerifyInstalled(context.Background(), m, o)
	if err != nil || len(results) != 1 || results[0].OK || results[0].ArtifactIntegrityVerified || !strings.Contains(results[0].Status, "missing or invalid") {
		t.Fatalf("tampered installer was not reported NOT VERIFIED: %#v, %v", results, err)
	}
	if err := Bootstrap(context.Background(), m, o); err != nil {
		t.Fatalf("repair tampered installer: %v", err)
	}
	callData, err = os.ReadFile(callLog)
	if err != nil || strings.Count(string(callData), "--cache-only") != 2 {
		t.Fatalf("tampered archive did not trigger exactly one repair: %q, %v", callData, err)
	}
}

func TestBootstrapFailsClosedWhenVeraPDFProvisionerFails(t *testing.T) {
	archiveHash := sha256.Sum256([]byte("expected installer"))
	m := testVeraPDFManifest(hex.EncodeToString(archiveHash[:]))
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	provisioner := filepath.Join(t.TempDir(), "provisioner.sh")
	if err := os.WriteFile(provisioner, []byte("#!/bin/sh\necho synthetic-secret >&2\nexit 17\n"), 0700); err != nil {
		t.Fatal(err)
	}
	o := Options{Root: root, Cache: cache, GoBinary: os.Args[0], VeraPDFProvisioner: provisioner, LockWait: time.Second}
	err := Bootstrap(context.Background(), m, o)
	if err == nil || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatalf("provisioning failure was not safely reported: %v", err)
	}
	artifactPath := veraPDFArtifactPath(root, m.Tools[0])
	if _, statErr := os.Lstat(artifactPath); !os.IsNotExist(statErr) {
		t.Fatalf("failed provisioner left an accepted artifact: %v", statErr)
	}
	results, verifyErr := VerifyInstalled(context.Background(), m, o)
	if verifyErr != nil || len(results) != 1 || results[0].OK || results[0].ArtifactIntegrityVerified || !strings.Contains(results[0].Status, "missing or invalid") {
		t.Fatalf("failed provisioner was reported as verified: %#v, %v", results, verifyErr)
	}
}

func TestVeraPDFCacheOnlyRejectsDirectoryTargetsBeforeDownloading(t *testing.T) {
	temp := t.TempDir()
	archiveName := "verapdf-greenfield-1.30.3-installer.zip"
	provisioner, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".github", "scripts", "provision-verapdf.sh"))
	if err != nil {
		t.Fatal(err)
	}
	fakeBin := filepath.Join(temp, "bin")
	if err := os.Mkdir(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	curlMarker := filepath.Join(temp, "curl-was-called")
	fakeCurl := "#!/bin/sh\nprintf called > \"$CURL_MARKER\"\nexit 1\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "curl"), []byte(fakeCurl), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURL_MARKER", curlMarker)
	t.Setenv("PATH", fakeBin+":"+os.Getenv("PATH"))

	directDirectory := filepath.Join(temp, archiveName)
	if err := os.Mkdir(directDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	linkedDirectory := filepath.Join(temp, "linked-directory")
	if err := os.Mkdir(linkedDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	symlinkTarget := filepath.Join(temp, "symlink-to-directory", archiveName)
	if err := os.Mkdir(filepath.Dir(symlinkTarget), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(linkedDirectory, symlinkTarget); err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{directDirectory, symlinkTarget} {
		t.Run(filepath.Base(filepath.Dir(target)), func(t *testing.T) {
			cmd := exec.Command("bash", provisioner, "--cache-only", target)
			if out, runErr := cmd.CombinedOutput(); runErr == nil {
				t.Fatalf("directory target was accepted: %s", out)
			}
			if _, err := os.Stat(curlMarker); !os.IsNotExist(err) {
				t.Fatalf("provisioner attempted a download before rejecting directory target: %v", err)
			}
		})
	}
	for _, path := range []string{filepath.Join(directDirectory, archiveName), filepath.Join(linkedDirectory, archiveName)} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("installer archive was moved inside a rejected directory target %s: %v", path, err)
		}
	}
}

func TestCloudBootstrapEntryPointRejectsUnexpectedArgsBeforeWrites(t *testing.T) {
	temp := t.TempDir()
	entrypoint, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "scripts", "bootstrap-cloud-build-tools.sh"))
	if err != nil {
		t.Fatal(err)
	}
	toolsRoot := filepath.Join(temp, "canonical-tools-root")
	cacheRoot := filepath.Join(temp, "canonical-cache-root")
	fakeBin := filepath.Join(temp, "bin")
	if err := os.Mkdir(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	sideEffectMarker := filepath.Join(temp, "filesystem-command-called")
	fakeMkdir := "#!/bin/sh\nprintf called > \"$SIDE_EFFECT_MARKER\"\nexit 1\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "mkdir"), []byte(fakeMkdir), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", entrypoint, "--root", filepath.Join(temp, "requested-root"))
	cmd.Env = append(os.Environ(),
		"CODEX_BUILD_TOOLS_ROOT="+toolsRoot,
		"CODEX_BUILD_CACHE_ROOT="+cacheRoot,
		"SIDE_EFFECT_MARKER="+sideEffectMarker,
		"PATH="+fakeBin+":"+os.Getenv("PATH"),
	)
	out, runErr := cmd.CombinedOutput()
	if runErr == nil || !strings.Contains(string(out), "does not accept command-line arguments") {
		t.Fatalf("unsupported CLI arguments were not rejected immediately: output=%q err=%v", out, runErr)
	}
	if _, err := os.Stat(sideEffectMarker); !os.IsNotExist(err) {
		t.Fatalf("entrypoint ran a filesystem command before rejecting arguments: %v", err)
	}
	for _, path := range []string{toolsRoot, cacheRoot, filepath.Join(temp, "requested-root")} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("unsupported argument changed or created root %s: %v", path, err)
		}
	}
}

func testVeraPDFManifest(installerSHA256 string) Manifest {
	return Manifest{SchemaVersion: 1, Platform: "linux/amd64", BootstrapSource: testBootstrapSource(), Tools: []Tool{{
		Name: "verapdf", Version: "1.30.3", License: "GPL-3.0-or-later OR MPL-2.0 (CI/CD-only use boundary retained)",
		Source: "https://software.verapdf.org/releases/1.30/verapdf-greenfield-1.30.3-installer.zip",
		SHA256: "19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70", ChecksumSource: "official detached signature in .github/scripts/provision-verapdf.sh",
		InstallerSHA256: installerSHA256, SignerFingerprint: "13DD102B4DD69354D12DE5A83184863278B17FE7",
		Kind: "ci-only-artifact", Binary: "verapdf", Scope: "CI/CD-only execution; cache signed installer only; never application/runtime",
		Status: "installer integrity is distinct from runtime readiness and CI validation execution",
	}}}
}

func testManifest(url, sha, kind string) Manifest {
	return Manifest{SchemaVersion: 1, Platform: "linux/amd64", BootstrapSource: testBootstrapSource(), Tools: []Tool{{Name: "fixture", Version: "1.2.3", License: "MIT", Source: url, SHA256: sha, ChecksumSource: "fixture server", Kind: kind, Binary: "fixture", VersionArgs: []string{"--version"}, VersionPattern: `testtool 1\.2\.3`}}}
}

func testBootstrapSource() BootstrapSource {
	return BootstrapSource{Revision: "0123456789abcdef0123456789abcdef01234567", URL: "https://github.com/example/project/archive/0123456789abcdef0123456789abcdef01234567.tar.gz", SHA256: strings.Repeat("c", 64)}
}

func TestManifestErrorTextDoesNotIncludeURLQuery(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer server.Close()
	m := testManifest(server.URL+"/tool?credential=synthetic", strings.Repeat("a", 64), "raw")
	err := Bootstrap(context.Background(), m, Options{Root: filepath.Join(t.TempDir(), "tools"), Cache: filepath.Join(t.TempDir(), "cache"), GoBinary: os.Args[0], Client: server.Client(), LockWait: time.Second})
	if err == nil || strings.Contains(err.Error(), "credential") {
		t.Fatalf("unexpected error leakage: %v", err)
	}
}

func TestEveryConfiguredVersionPatternRejectsLongerNumericPrefix(t *testing.T) {
	f, err := os.Open("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := DecodeManifest(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range m.Tools {
		if tool.VersionPattern == "" {
			continue
		}
		p := regexp.MustCompile(tool.VersionPattern)
		good := "tool version " + tool.Version
		bad := "tool 1" + tool.Version
		if tool.Name == "go" {
			good = "go" + tool.Version
			bad = "go1" + tool.Version
		}
		if !p.MatchString(good) {
			t.Errorf("%s configured pattern rejects expected version output %q", tool.Name, good)
		}
		if p.MatchString(bad) {
			t.Errorf("%s configured pattern accepts longer numeric prefix %q", tool.Name, bad)
		}
	}
}

func TestGoBuildInfoVersionFallbackRequiresExactModuleAndVersion(t *testing.T) {
	good := "path github.com/google/go-licenses/v2\nmod github.com/google/go-licenses/v2 v2.0.1 h1:synthetic\n"
	if !matchesModuleVersion(good, "github.com/google/go-licenses/v2", "2.0.1") {
		t.Fatal("did not recognize exact module version")
	}
	for _, bad := range []string{
		"mod github.com/google/go-licenses/v2 v2.0.10 h1:synthetic\n",
		"mod github.com/other/tool v2.0.1 h1:synthetic\n",
		"dep github.com/google/go-licenses/v2 v2.0.1 h1:synthetic\n",
	} {
		if matchesModuleVersion(bad, "github.com/google/go-licenses/v2", "2.0.1") {
			t.Errorf("accepted build metadata %q", bad)
		}
	}
}

func TestInstallGoModuleUsesOfficialProxyAndChecksumDatabaseWithoutFallback(t *testing.T) {
	capture := filepath.Join(t.TempDir(), "go-env")
	fakeGo := filepath.Join(t.TempDir(), "go")
	script := "#!/bin/sh\nprintf '%s\\n' \"$GOPROXY\" \"$GONOPROXY\" \"$GOSUMDB\" \"$GOPRIVATE\" \"$GONOSUMDB\" \"$GOINSECURE\" > \"$CAPTURE_ENV\"\n"
	if err := os.WriteFile(fakeGo, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CAPTURE_ENV", capture)
	t.Setenv("GOPROXY", "direct")
	t.Setenv("GONOPROXY", "*")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOPRIVATE", "*")
	t.Setenv("GONOSUMDB", "*")
	t.Setenv("GOINSECURE", "*")
	cache := filepath.Join(t.TempDir(), "cache")
	if err := installGoModule(context.Background(), Tool{Module: "example.invalid/tool@v1.2.3"}, t.TempDir(), Options{Cache: cache, GoBinary: fakeGo}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://proxy.golang.org\n\nsum.golang.org\n\n\n\n"
	if string(got) != want {
		t.Fatalf("module fetch policy environment = %q, want %q", got, want)
	}
}

func TestInstallGoModuleIgnoresPersistedGoEnvProxyBypass(t *testing.T) {
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(goBinary); err != nil {
		t.Fatalf("locate current Go toolchain: %v", err)
	}
	persisted := filepath.Join(t.TempDir(), "goenv")
	content := "GONOPROXY=*\nGONOSUMDB=*\nGOPRIVATE=*\nGOINSECURE=*\n"
	if err := os.WriteFile(persisted, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(t.TempDir(), "effective-go-env.json")
	rawGoEnv := filepath.Join(t.TempDir(), "raw-goenv")
	wrapper := filepath.Join(t.TempDir(), "go-wrapper")
	script := "#!/bin/sh\nprintf '%s\\n' \"$GOENV\" > \"$CAPTURE_GOENV\"\n\"$REAL_GO\" env -json GOPROXY GONOPROXY GOSUMDB GOPRIVATE GONOSUMDB GOINSECURE > \"$CAPTURE_ENV\"\n"
	if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOENV", persisted)
	t.Setenv("REAL_GO", goBinary)
	t.Setenv("CAPTURE_ENV", capture)
	t.Setenv("CAPTURE_GOENV", rawGoEnv)
	t.Setenv("GOPROXY", "direct")
	t.Setenv("GONOPROXY", "*")
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOPRIVATE", "*")
	t.Setenv("GONOSUMDB", "*")
	t.Setenv("GOINSECURE", "*")
	if err := installGoModule(context.Background(), Tool{Module: "example.invalid/tool@v1.2.3"}, t.TempDir(), Options{Cache: filepath.Join(t.TempDir(), "cache"), GoBinary: wrapper}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		GOPROXY    string
		GONOPROXY  string
		GOSUMDB    string
		GOPRIVATE  string
		GONOSUMDB  string
		GOINSECURE string
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode real Go environment: %v; data=%q", err, data)
	}
	want := struct {
		GOPROXY    string
		GONOPROXY  string
		GOSUMDB    string
		GOPRIVATE  string
		GONOSUMDB  string
		GOINSECURE string
	}{GOPROXY: "https://proxy.golang.org", GOSUMDB: "sum.golang.org"}
	if got != want {
		t.Fatalf("real go env = %#v, want %#v", got, want)
	}
	raw, err := os.ReadFile(rawGoEnv)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "off\n" {
		t.Fatalf("install subprocess GOENV = %q, want off", raw)
	}
}
