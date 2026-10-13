//go:build linux

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap/internal/cloudtools"
)

func TestMainSignalCancelsSeparateHelperGroupAndReleasesLock(t *testing.T) {
	if os.Getenv("CLOUDTOOLS_MAIN_SIGNAL_HELPER") == "1" {
		os.Args = []string{
			"cloud-tools-bootstrap",
			"--manifest", os.Getenv("CLOUDTOOLS_SIGNAL_MANIFEST"),
			"--root", os.Getenv("CLOUDTOOLS_SIGNAL_ROOT"),
			"--cache", os.Getenv("CLOUDTOOLS_SIGNAL_CACHE"),
			"--go", os.Getenv("CLOUDTOOLS_SIGNAL_GO"),
		}
		os.Exit(run())
	}

	work := t.TempDir()
	root := filepath.Join(work, "tools")
	cache := filepath.Join(work, "cache")
	manifest := filepath.Join(work, "manifest.json")
	childPIDFile := filepath.Join(work, "helper-pid")
	goFixture := filepath.Join(work, "fake-go")
	if err := os.WriteFile(manifest, []byte(`{
		"schema_version":1,
		"platform":"linux/amd64",
		"bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","url":"https://github.com/example/project/archive/0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
		"tools":[{"name":"fixture","version":"1.2.3","license":"MIT","source":"go install example.com/fixture@v1.2.3","checksum_source":"fixture","kind":"go-install","module":"example.com/fixture@v1.2.3","binary":"fixture","version_args":["--version"],"version_pattern":"fixture 1\\.2\\.3"}]
	}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goFixture, []byte("#!/bin/bash\nprintf '%s\\n' \"$$\" > \"$CLOUDTOOLS_SIGNAL_HELPER_PID\"\ntrap '' INT TERM\nwhile :; do sleep 0.1; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestMainSignalCancelsSeparateHelperGroupAndReleasesLock$")
	cmd.Env = append(os.Environ(),
		"CLOUDTOOLS_MAIN_SIGNAL_HELPER=1",
		"CLOUDTOOLS_SIGNAL_MANIFEST="+manifest,
		"CLOUDTOOLS_SIGNAL_ROOT="+root,
		"CLOUDTOOLS_SIGNAL_CACHE="+cache,
		"CLOUDTOOLS_SIGNAL_GO="+goFixture,
		"CLOUDTOOLS_SIGNAL_HELPER_PID="+childPIDFile,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var helperPID, helperPGID int
	completed := false
	defer func() {
		if helperPGID > 0 {
			_ = syscall.Kill(-helperPGID, syscall.SIGKILL)
		}
		if helperPID > 0 {
			_ = syscall.Kill(helperPID, syscall.SIGKILL)
		}
		if completed {
			return
		}
		_ = cmd.Process.Kill()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Errorf("signal helper process did not stop during fixture teardown")
		}
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(childPIDFile); err == nil {
			break
		}
		select {
		case err := <-done:
			completed = true
			t.Fatalf("orchestrator exited before helper started: %v", err)
		case <-time.After(10 * time.Millisecond):
		}
	}
	data, err := os.ReadFile(childPIDFile)
	if err != nil {
		t.Fatalf("helper did not start within bound: %v", err)
	}
	if helperPID, err = strconv.Atoi(strings.TrimSpace(string(data))); err != nil || helperPID <= 0 {
		t.Fatalf("invalid helper PID %q: %v", data, err)
	}
	if helperPGID, err = syscall.Getpgid(helperPID); err != nil {
		t.Fatalf("get helper process group: %v", err)
	}
	orchestratorPGID, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil || helperPGID == orchestratorPGID {
		t.Fatalf("helper was not isolated in its own process group: helper=%d orchestrator=%d err=%v", helperPGID, orchestratorPGID, err)
	}
	lock := filepath.Join(root, ".bootstrap.lock")
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("orchestrator lock missing while helper is active: %v", err)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		completed = true
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("orchestrator exit = %v; want bootstrap failure after cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("orchestrator did not finish cancellation and cleanup within bound")
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("inner bootstrap lock remains after cancellation: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "activate.sh")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap proceeded to activation after cancellation: %v", err)
	}
	if processAlive(helperPID) {
		t.Fatalf("helper PID %d survived orchestrator cancellation", helperPID)
	}
}

func TestActivationUsesSelectedVerifiedGoBinary(t *testing.T) {
	for _, tc := range []struct {
		name      string
		version   string
		defaultGo bool
	}{
		{name: "alternate pinned toolchain at path with spaces", version: "1.28.3"},
		{name: "default toolchain compatibility", version: "1.27.2", defaultGo: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			work := t.TempDir()
			root := filepath.Join(work, "tools root with spaces")
			cache := filepath.Join(work, "cache root with spaces")
			if err := os.MkdirAll(filepath.Join(root, "bin"), 0700); err != nil {
				t.Fatal(err)
			}
			goBinary := filepath.Join(work, "alternate Go's path", "bin dir", "go")
			if tc.defaultGo {
				goBinary = filepath.Join(root, "toolchains", "go", tc.version, "bin", "go")
			}
			if err := os.MkdirAll(filepath.Dir(goBinary), 0700); err != nil {
				t.Fatal(err)
			}
			writeFakeGo(t, goBinary, tc.version)
			manifestPath := writeToolchainManifest(t, work, tc.version)
			if code := runBootstrapCLI(t, manifestPath, root, cache, goBinary); code != 0 {
				t.Fatalf("bootstrap with selected verified Go returned %d", code)
			}
			activation := filepath.Join(root, "activate.sh")
			cmd := exec.Command("bash", "-c", `source "$1" || exit 50
resolved="$(command -v go)"
[[ "$resolved" == "$EXPECTED_GO" ]] || exit 51
[[ "$(go version)" == "$EXPECTED_VERSION" ]] || exit 52`, "bash", activation)
			cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin", "EXPECTED_GO="+goBinary, "EXPECTED_VERSION=go version go"+tc.version+" linux/amd64")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("activation did not select the validated Go binary: output=%q err=%v", output, err)
			}
		})
	}
}

func TestBootstrapRejectsInvalidSelectedGoBeforeActivation(t *testing.T) {
	work := t.TempDir()
	root := filepath.Join(work, "tools root")
	cache := filepath.Join(work, "cache root")
	goBinary := filepath.Join(work, "alternate go", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBinary), 0700); err != nil {
		t.Fatal(err)
	}
	writeFakeGo(t, goBinary, "1.27.2")
	manifestPath := writeToolchainManifest(t, work, "1.28.3")
	if code := runBootstrapCLI(t, manifestPath, root, cache, goBinary); code == 0 {
		t.Fatal("bootstrap accepted a selected Go binary outside the manifest version pin")
	}
	if _, err := os.Stat(filepath.Join(root, "activate.sh")); !os.IsNotExist(err) {
		t.Fatalf("activation was written for invalid selected Go: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap lock remained after invalid selected Go: %v", err)
	}
}

func runBootstrapCLI(t *testing.T, manifest, root, cache, goBinary string) int {
	t.Helper()
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"cloud-tools-bootstrap", "--manifest", manifest, "--root", root, "--cache", cache, "--go", goBinary}
	return run()
}

func writeFakeGo(t *testing.T, path, version string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nif [ \"${1-}\" = version ]; then printf 'go version go"+version+" linux/amd64\\n'; else exit 1; fi\n"), 0700); err != nil {
		t.Fatal(err)
	}
}

func writeToolchainManifest(t *testing.T, dir, version string) string {
	t.Helper()
	revision := "0123456789abcdef0123456789abcdef01234567"
	manifest := cloudtools.Manifest{
		SchemaVersion: 1,
		Platform:      "linux/amd64",
		BootstrapSource: cloudtools.BootstrapSource{
			Revision: revision,
			URL:      "https://github.com/example/project/archive/" + revision + ".tar.gz",
			SHA256:   strings.Repeat("a", 64),
		},
		Tools: []cloudtools.Tool{{
			Name:           "go",
			Version:        version,
			License:        "BSD-3-Clause",
			Source:         "https://go.dev/dl/go" + version + ".linux-amd64.tar.gz",
			SHA256:         strings.Repeat("b", 64),
			ChecksumSource: "go.dev download checksums",
			Kind:           "bootstrap-toolchain",
			Binary:         "go",
			VersionArgs:    []string{"version"},
			VersionPattern: `go version go` + strings.ReplaceAll(version, ".", `\.`) + ` linux/amd64`,
		}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "toolchain manifest.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
