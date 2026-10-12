package cloudtools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func TestActivationCanBeSourcedRepeatedlyWithoutChangingPATHAgain(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"bin", "toolchains/go/1.27.2/bin"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	activation := filepath.Join(root, "activate.sh")
	if err := writeActivation(activation); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", `source "$1"; first=$PATH; source "$1"; test "$PATH" = "$first"`, "bash", activation)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("activation failed or duplicated PATH: %s %v", out, err)
	}
}

func TestBootstrapKeepsVeraPDFOutsidePersistentToolRoot(t *testing.T) {
	m := Manifest{SchemaVersion: 1, Platform: "linux/amd64", Tools: []Tool{{Name: "verapdf", Version: "1.30.3", License: "MPL-2.0", Source: "https://software.verapdf.org/releases/1.30/installer.zip", SHA256: strings.Repeat("a", 64), ChecksumSource: "CI provisioner", InstallerSHA256: strings.Repeat("b", 64), SignerFingerprint: "13DD102B4DD69354D12DE5A83184863278B17FE7", Kind: "ci-only-external", Binary: "verapdf", Scope: "CI-only; never install into application/runtime image", Status: "verified"}}}
	root, cache := filepath.Join(t.TempDir(), "tools"), filepath.Join(t.TempDir(), "cache")
	var log bytes.Buffer
	if err := Bootstrap(context.Background(), m, Options{Root: root, Cache: cache, GoBinary: os.Args[0], Output: &log}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(log.String(), "blocked verapdf") || !strings.Contains(log.String(), "external verapdf") {
		t.Fatalf("unexpected CI tool status: %s", log.String())
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "verapdf")); !os.IsNotExist(err) {
		t.Fatalf("veraPDF was installed into persistent tool root: %v", err)
	}
}

func testManifest(url, sha, kind string) Manifest {
	return Manifest{SchemaVersion: 1, Platform: "linux/amd64", Tools: []Tool{{Name: "fixture", Version: "1.2.3", License: "MIT", Source: url, SHA256: sha, ChecksumSource: "fixture server", Kind: kind, Binary: "fixture", VersionArgs: []string{"--version"}, VersionPattern: `testtool 1\.2\.3`}}}
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

func TestConfiguredVersionPatternIsExactEnough(t *testing.T) {
	p := regexp.MustCompile(`(?:^|[^0-9])1\.2\.3(?:[^0-9]|$)`)
	if p.MatchString("testtool 11.2.30") || !p.MatchString("testtool v1.2.3") {
		t.Fatal("version pattern must accept only the exact numeric version")
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
