//go:build linux

package cloudtools

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const bootstrapScript = "../../../../scripts/bootstrap-cloud-build-tools.sh"

func TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock(t *testing.T) {
	for _, tc := range []struct {
		name   string
		signal syscall.Signal
		status int
	}{
		{name: "interrupt", signal: syscall.SIGINT, status: 130},
		{name: "terminate", signal: syscall.SIGTERM, status: 143},
	} {
		t.Run(tc.name, func(t *testing.T) {
			work := t.TempDir()
			toolsRoot := filepath.Join(work, "tools")
			cacheRoot := filepath.Join(work, "cache")
			goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
			if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
				t.Fatal(err)
			}
			started := filepath.Join(work, "go-child-pid")
			calls := filepath.Join(work, "go-calls")
			stub := "#!/bin/bash\nprintf '%s\\n' \"$*\" >> \"$GO_STUB_CALLS\"\n" +
				"if [[ \"${1-}\" == version ]]; then echo 'go version go1.27.2 linux/amd64'; exit 0; fi\n" +
				"printf '%s\\n' \"$$\" > \"$GO_STUB_STARTED\"\ntrap '' INT TERM\nwhile :; do sleep 0.1; done\n"
			if err := os.WriteFile(goBin, []byte(stub), 0700); err != nil {
				t.Fatal(err)
			}
			proc := startBootstrap(t, toolsRoot, cacheRoot, nil,
				"GO_STUB_STARTED="+started,
				"GO_STUB_CALLS="+calls,
			)
			var childPID, childPGID int
			defer func() { stopBootstrapFixture(proc, &childPID, &childPGID, started) }()
			waitForFile(t, started, proc)
			childPID = readPID(t, started)
			childPGID, err := syscall.Getpgid(childPID)
			if err != nil {
				t.Fatalf("get fake Go process group: %v", err)
			}
			bootstrapPGID, err := syscall.Getpgid(proc.cmd.Process.Pid)
			if err != nil || childPGID == bootstrapPGID {
				t.Fatalf("fixture did not create the separate child group: child=%d bootstrap=%d err=%v", childPGID, bootstrapPGID, err)
			}

			lock := filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")
			if _, err := os.Stat(lock); err != nil {
				t.Fatalf("bootstrap lock missing while fixture runs: %v", err)
			}
			if err := proc.cmd.Process.Signal(tc.signal); err != nil {
				t.Fatal(err)
			}
			time.Sleep(200 * time.Millisecond)
			if !processAlive(childPID) {
				t.Fatal("uncooperative child stopped before bounded cleanup")
			}
			if _, err := os.Stat(lock); err != nil {
				t.Fatalf("bootstrap released lock before child cleanup completed: %v", err)
			}
			if err := waitBootstrap(t, proc, 5*time.Second, tc.status); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(lock); !os.IsNotExist(err) {
				t.Fatalf("bootstrap lock remains after cleanup: %v", err)
			}
			if processAlive(childPID) {
				t.Fatalf("fake child PID %d survived bootstrap cleanup", childPID)
			}
			if data, err := os.ReadFile(calls); err != nil || strings.Count(string(data), "--manifest") != 1 || strings.Contains(string(data), "--verify") {
				t.Fatalf("later verification proceeded after signal: calls=%q err=%v", data, err)
			}
			if _, err := os.Stat(filepath.Join(toolsRoot, "bin", "cloud-tools-bootstrap")); !os.IsNotExist(err) {
				t.Fatalf("bootstrap installation proceeded after signal: %v", err)
			}
		})
	}
}

func TestBootstrapEntrypointStopsSignalDuringChildStartup(t *testing.T) {
	work := t.TempDir()
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	fakeBin := filepath.Join(work, "bin")
	if err := os.MkdirAll(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
		t.Fatal(err)
	}
	started := filepath.Join(work, "setsid-child-pid")
	if err := os.WriteFile(filepath.Join(fakeBin, "setsid"), []byte("#!/bin/bash\nprintf '%s\\n' \"$$\" > \"$SETSID_STUB_STARTED\"\ntrap '' INT TERM\nwhile :; do :; done\nexec /usr/bin/setsid \"$@\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goBin, []byte("#!/bin/bash\necho 'go version go1.27.2 linux/amd64'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(work, "go-calls")
	proc := startBootstrap(t, toolsRoot, cacheRoot, []string{"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")},
		"SETSID_STUB_STARTED="+started,
		"GO_STUB_CALLS="+calls,
	)
	var launcherPID, launcherPGID int
	defer func() { stopBootstrapFixture(proc, &launcherPID, &launcherPGID, started) }()
	waitForFile(t, started, proc)
	launcherPID = readPID(t, started)
	launcherPGID, _ = syscall.Getpgid(launcherPID)
	if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := waitBootstrap(t, proc, 4*time.Second, 143); err != nil {
		t.Fatal(err)
	}
	if processAlive(launcherPID) {
		t.Fatalf("setsid launcher PID %d survived startup cancellation", launcherPID)
	}
	if _, err := os.Stat(filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap lock remains after startup cancellation: %v", err)
	}
	if _, err := os.Stat(calls); !os.IsNotExist(err) {
		t.Fatalf("bootstrap continued into Go work after startup cancellation: %v", err)
	}
}

func TestBootstrapEntrypointSignalAfterReadinessBeforeAcknowledgement(t *testing.T) {
	work := t.TempDir()
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	fakeBin := filepath.Join(work, "bin")
	if err := os.MkdirAll(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
		t.Fatal(err)
	}
	started := filepath.Join(work, "go-started")
	release := filepath.Join(work, "release-mv")
	mvReady := filepath.Join(work, "mv-ready")
	calls := filepath.Join(work, "go-calls")
	goStub := "#!/bin/bash\nprintf '%s\\n' \"$*\" >> \"$GO_STUB_CALLS\"\n" +
		"if [[ \"${1-}\" == version ]]; then printf '%s\\n' \"$$\" > \"$GO_STUB_STARTED\"; trap '' INT TERM; while :; do :; done; fi\n"
	if err := os.WriteFile(goBin, []byte(goStub), 0700); err != nil {
		t.Fatal(err)
	}
	mvStub := `#!/bin/bash
[[ "${1-}" == -- ]] && shift
source=$1
destination=$2
if [[ "$source" == */.bootstrap-child.*/ack.pending ]]; then
  : > "$MV_STUB_READY"
  while [[ ! -e "$MV_STUB_RELEASE" ]]; do sleep 0.01; done
fi
exec /bin/mv "$source" "$destination"
`
	if err := os.WriteFile(filepath.Join(fakeBin, "mv"), []byte(mvStub), 0700); err != nil {
		t.Fatal(err)
	}
	proc := startBootstrap(t, toolsRoot, cacheRoot, []string{"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")},
		"GO_STUB_STARTED="+started,
		"GO_STUB_CALLS="+calls,
		"MV_STUB_READY="+mvReady,
		"MV_STUB_RELEASE="+release,
	)
	var childPID, childPGID int
	defer func() {
		for _, path := range []string{started} {
			if childPID == 0 {
				if data, err := os.ReadFile(path); err == nil {
					childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
				}
			}
		}
		if files, _ := filepath.Glob(filepath.Join(cacheRoot, ".bootstrap-child.*", "ready")); len(files) > 0 {
			if data, err := os.ReadFile(files[0]); err == nil {
				childPGID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			}
		}
		stopBootstrapFixture(proc, &childPID, &childPGID, started)
	}()
	waitForFile(t, mvReady, proc)
	readyFiles, err := filepath.Glob(filepath.Join(cacheRoot, ".bootstrap-child.*", "ready"))
	if err != nil || len(readyFiles) != 1 {
		t.Fatalf("expected published child readiness marker, files=%v err=%v", readyFiles, err)
	}
	childPGID = readPID(t, readyFiles[0])
	ackPath := filepath.Join(filepath.Dir(readyFiles[0]), "ack")
	if _, err := os.Stat(ackPath); !os.IsNotExist(err) {
		t.Fatalf("child was acknowledged before gated parent release: %v", err)
	}
	if _, err := os.Stat(started); !os.IsNotExist(err) {
		t.Fatalf("Go tool began before parent acknowledgement: %v", err)
	}
	lock := filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")
	if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("lock released while parent acknowledgement was blocked: %v", err)
	}
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := waitBootstrap(t, proc, 5*time.Second, 143); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(started); err == nil {
		childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if childPID > 0 && processRunning(childPID) {
		t.Fatalf("Go child PID %d survived cancellation after readiness publication", childPID)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock remains after cancellation cleanup: %v", err)
	}
	if data, err := os.ReadFile(calls); err == nil && strings.Contains(string(data), "--verify") {
		t.Fatalf("verification proceeded after cancellation: %q", data)
	} else if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestBootstrapEntrypointSignalDuringVersionCheckKillsForegroundTool(t *testing.T) {
	work := t.TempDir()
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
		t.Fatal(err)
	}
	started := filepath.Join(work, "go-version-pid")
	goStub := "#!/bin/bash\nif [[ \"${1-}\" == version ]]; then printf '%s\\n' \"$$\" > \"$GO_STUB_STARTED\"; trap '' INT TERM; while :; do :; done; fi\n"
	if err := os.WriteFile(goBin, []byte(goStub), 0700); err != nil {
		t.Fatal(err)
	}
	proc := startBootstrap(t, toolsRoot, cacheRoot, nil, "GO_STUB_STARTED="+started)
	var childPID, childPGID int
	defer func() { stopBootstrapFixture(proc, &childPID, &childPGID, started) }()
	waitForFile(t, started, proc)
	childPID = readPID(t, started)
	childPGID, err := syscall.Getpgid(childPID)
	if err != nil {
		t.Fatal(err)
	}
	bootstrapPGID, err := syscall.Getpgid(proc.cmd.Process.Pid)
	if err != nil || childPGID == bootstrapPGID {
		t.Fatalf("version checker is not in the managed process group: child=%d bootstrap=%d err=%v", childPGID, bootstrapPGID, err)
	}
	lock := filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")
	if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if !processAlive(childPID) {
		t.Fatal("uncooperative version checker stopped before bounded cleanup")
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("lock released before foreground version tool cleanup: %v", err)
	}
	if err := waitBootstrap(t, proc, 5*time.Second, 143); err != nil {
		t.Fatal(err)
	}
	if processRunning(childPID) {
		t.Fatalf("foreground version checker PID %d survived cleanup", childPID)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock remains after version-check cleanup: %v", err)
	}
}

func TestBootstrapEntrypointSourceBuildDeadlineCleansDescendantBeforeUnlock(t *testing.T) {
	work := t.TempDir()
	bootstrapRoot := filepath.Join(work, "bootstrap")
	scriptDir := filepath.Join(bootstrapRoot, "scripts")
	if err := os.MkdirAll(scriptDir, 0700); err != nil {
		t.Fatal(err)
	}
	originalScript, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "scripts", "bootstrap-cloud-build-tools.sh"))
	if err != nil {
		t.Fatal(err)
	}
	testScript := strings.Replace(string(originalScript), "timeout --foreground 8m", "timeout --foreground 1s", 1)
	if testScript == string(originalScript) {
		t.Fatal("source-build timeout fixture did not find the pinned build deadline")
	}
	const revisionMarker = "readonly bootstrap_source_revision='"
	revisionAt := strings.Index(testScript, revisionMarker)
	if revisionAt < 0 {
		t.Fatal("source-build timeout fixture could not find the pinned source revision")
	}
	revision := strings.SplitN(testScript[revisionAt+len(revisionMarker):], "'", 2)[0]
	entrypoint := filepath.Join(scriptDir, "bootstrap-cloud-build-tools.sh")
	if err := os.WriteFile(entrypoint, []byte(testScript), 0700); err != nil {
		t.Fatal(err)
	}
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	fakeBin := filepath.Join(work, "bin")
	if err := os.MkdirAll(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
		t.Fatal(err)
	}
	descendantFile := filepath.Join(work, "build-descendant-pid")
	goStub := `#!/bin/bash
if [[ "${1-}" == version ]]; then
  echo 'go version go1.27.2 linux/amd64'
  exit 0
fi
if [[ "${1-}" == build ]]; then
  bash -c 'trap "" TERM; printf "%s\\n" "$$" > "$BUILD_DESCENDANT_PID"; while :; do sleep 0.05; done' &
  child=$!
  wait "$child"
fi
exit 0
`
	if err := os.WriteFile(goBin, []byte(goStub), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeBin, "sha256sum"), []byte("#!/bin/bash\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	tarStub := `#!/bin/bash
destination=''
while (($#)); do
  if [[ "$1" == -C ]]; then shift; destination=$1; fi
  shift
done
mkdir -p "$destination/cmd/cloud-tools-bootstrap"
printf 'module fixture\\n' > "$destination/cmd/cloud-tools-bootstrap/go.mod"
printf 'package main\\n' > "$destination/cmd/cloud-tools-bootstrap/main.go"
`
	if err := os.WriteFile(filepath.Join(fakeBin, "tar"), []byte(tarStub), 0700); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(cacheRoot, "downloads", fmt.Sprintf("cloud-tools-bootstrap-%s.tar.gz", revision))
	if err := os.MkdirAll(filepath.Dir(archive), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, []byte("synthetic archive fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	proc := startBootstrapAt(t, entrypoint, toolsRoot, cacheRoot,
		[]string{"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")},
		"BUILD_DESCENDANT_PID="+descendantFile,
	)
	var childPID, childPGID int
	defer func() {
		if childPID == 0 {
			if data, err := os.ReadFile(descendantFile); err == nil {
				childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			}
		}
		if files, _ := filepath.Glob(filepath.Join(cacheRoot, ".bootstrap-child.*", "ready")); len(files) > 0 {
			if data, err := os.ReadFile(files[0]); err == nil {
				childPGID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
			}
		}
		stopBootstrapFixture(proc, &childPID, &childPGID, descendantFile)
	}()
	waitForFile(t, descendantFile, proc)
	childPID = readPID(t, descendantFile)
	childPGID, err = syscall.Getpgid(childPID)
	if err != nil {
		t.Fatal(err)
	}
	readyFiles, err := filepath.Glob(filepath.Join(cacheRoot, ".bootstrap-child.*", "ready"))
	if err != nil || len(readyFiles) != 1 {
		t.Fatalf("expected one managed child process group, files=%v err=%v", readyFiles, err)
	}
	managedPGID := readPID(t, readyFiles[0])
	if childPGID != managedPGID {
		t.Fatalf("build descendant escaped managed process group: descendant=%d managed=%d", childPGID, managedPGID)
	}
	lock := filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")
	time.Sleep(1200 * time.Millisecond)
	if !processRunning(childPID) {
		t.Fatal("uncooperative build descendant stopped before bounded cleanup completed")
	}
	if _, err := os.Stat(lock); err != nil {
		t.Fatalf("bootstrap released lock before the source-build descendant was cleaned: %v", err)
	}
	if err := waitBootstrap(t, proc, 4*time.Second, 1); err != nil {
		t.Fatalf("source-build deadline did not fail bootstrap cleanly: %v; output=%s", err, proc.output())
	}
	if processRunning(childPID) {
		t.Fatalf("source-build descendant PID %d survived timeout cleanup", childPID)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("bootstrap lock remains after timed-out build cleanup: %v", err)
	}
	if _, err := os.Stat(filepath.Join(toolsRoot, "bin", "cloud-tools-bootstrap")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap installed after source-build timeout: %v", err)
	}
}

func TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput(t *testing.T) {
	for _, tc := range []struct {
		name       string
		versionErr string
		timeout    bool
	}{
		{name: "nonzero-version", versionErr: "23"},
		{name: "timeout-after-output", timeout: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			work := t.TempDir()
			toolsRoot := filepath.Join(work, "tools")
			cacheRoot := filepath.Join(work, "cache")
			fakeBin := filepath.Join(work, "bin")
			if err := os.MkdirAll(fakeBin, 0700); err != nil {
				t.Fatal(err)
			}
			goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
			if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
				t.Fatal(err)
			}
			calls := filepath.Join(work, "go-calls")
			stub := "#!/bin/bash\nprintf '%s\\n' \"$*\" >> \"$GO_STUB_CALLS\"\n" +
				"if [[ \"${1-}\" == version ]]; then echo 'go version go1.27.2 linux/amd64'; exit \"${GO_STUB_VERSION_STATUS:-0}\"; fi\nexit 0\n"
			if err := os.WriteFile(goBin, []byte(stub), 0700); err != nil {
				t.Fatal(err)
			}
			if tc.timeout {
				timeoutStub := "#!/bin/bash\n[[ \"${1-}\" == --foreground ]] && shift\nshift\n\"$@\"\nexit 124\n"
				if err := os.WriteFile(filepath.Join(fakeBin, "timeout"), []byte(timeoutStub), 0700); err != nil {
					t.Fatal(err)
				}
			}
			proc := startBootstrap(t, toolsRoot, cacheRoot, []string{"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")},
				"GO_STUB_CALLS="+calls,
				"GO_STUB_VERSION_STATUS="+tc.versionErr,
			)
			if err := waitBootstrap(t, proc, 5*time.Second, 1); err != nil {
				t.Fatalf("bootstrap failed unexpectedly: %v; output=%s", err, proc.output())
			}
			output := proc.output()
			if !strings.Contains(string(output), "Existing Go path has an unexpected version") {
				t.Fatalf("bootstrap did not reject the failed version pipeline: %s", output)
			}
			if data, err := os.ReadFile(calls); err != nil || strings.TrimSpace(string(data)) != "version" {
				t.Fatalf("bootstrap proceeded after version check failed: calls=%q err=%v", data, err)
			}
			if _, err := os.Stat(filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")); !os.IsNotExist(err) {
				t.Fatalf("bootstrap lock remains after failed version check: %v", err)
			}
		})
	}
}

func TestBootstrapEntrypointRejectsUnsuccessfulDownloadedToolchainVersion(t *testing.T) {
	work := t.TempDir()
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	fakeBin := filepath.Join(work, "bin")
	if err := os.MkdirAll(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]string{
		"curl": `#!/bin/bash
while (($#)); do
  if [[ "$1" == --output ]]; then shift; : > "$1"; fi
  shift
done
`,
		"sha256sum": "#!/bin/bash\nexit 0\n",
		"tar": `#!/bin/bash
destination=''
while (($#)); do
  if [[ "$1" == -C ]]; then shift; destination="$1"; fi
  shift
done
mkdir -p "$destination/bin"
cat > "$destination/bin/go" <<'GO_FIXTURE'
#!/bin/bash
if [[ "${1-}" == version ]]; then
  echo 'go version go1.27.2 linux/amd64'
  exit 23
fi
printf '%s\n' "$*" >> "$GO_STUB_CALLS"
exit 0
GO_FIXTURE
chmod 0755 "$destination/bin/go"
`,
	}
	for name, body := range fixtures {
		if err := os.WriteFile(filepath.Join(fakeBin, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	calls := filepath.Join(work, "later-go-calls")
	proc := startBootstrap(t, toolsRoot, cacheRoot,
		[]string{"PATH=" + fakeBin + string(os.PathListSeparator) + os.Getenv("PATH")},
		"GO_STUB_CALLS="+calls,
	)
	defer func() { stopBootstrapFixture(proc, new(int), new(int)) }()
	if err := waitBootstrap(t, proc, 5*time.Second, 1); err != nil {
		t.Fatalf("bootstrap failed unexpectedly: %v; output=%s", err, proc.output())
	}
	if !strings.Contains(proc.output(), "Go archive version did not match pinned version") {
		t.Fatalf("bootstrap did not reject failed archived-tool version check: %s", proc.output())
	}
	if _, err := os.Stat(calls); !os.IsNotExist(err) {
		t.Fatalf("bootstrap continued after archived-tool version failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(toolsRoot, "toolchains", "go", "1.27.2")); !os.IsNotExist(err) {
		t.Fatalf("unverified toolchain was installed after failed version check: %v", err)
	}
	if _, err := os.Stat(filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap lock remains after failed version check: %v", err)
	}
}

func TestBootstrapEntrypointPreservesSuccessfulInstallAndVerifyFlow(t *testing.T) {
	work := t.TempDir()
	toolsRoot := filepath.Join(work, "tools")
	cacheRoot := filepath.Join(work, "cache")
	goBin := filepath.Join(toolsRoot, "toolchains", "go", "1.27.2", "bin", "go")
	if err := os.MkdirAll(filepath.Dir(goBin), 0700); err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(work, "go-calls")
	stub := "#!/bin/bash\nprintf '%s\\n' \"$*\" >> \"$GO_STUB_CALLS\"\n" +
		"if [[ \"${1-}\" == version ]]; then echo 'go version go1.27.2 linux/amd64'; fi\nexit 0\n"
	if err := os.WriteFile(goBin, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	proc := startBootstrap(t, toolsRoot, cacheRoot, nil, "GO_STUB_CALLS="+calls)
	if err := waitBootstrap(t, proc, 5*time.Second, 0); err != nil {
		t.Fatalf("successful fixture bootstrap failed: %s: %v", proc.output(), err)
	}
	data, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 || lines[0] != "version" || !strings.Contains(lines[1], "--manifest") || !strings.Contains(lines[2], "--verify") {
		t.Fatalf("successful bootstrap calls = %q; want version, install, verify", lines)
	}
	if _, err := os.Stat(filepath.Join(toolsRoot, "toolchains", "go", ".bootstrap.lock")); !os.IsNotExist(err) {
		t.Fatalf("bootstrap lock remains after successful completion: %v", err)
	}
	if readyFiles, err := filepath.Glob(filepath.Join(cacheRoot, ".bootstrap-child-ready.*")); err != nil || len(readyFiles) != 0 {
		t.Fatalf("successful bootstrap left child readiness files: %v err=%v", readyFiles, err)
	}
}

type bootstrapProcess struct {
	cmd      *exec.Cmd
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	done     chan error
	finished bool
	waitErr  error
}

func (p *bootstrapProcess) output() string {
	return p.stdout.String() + p.stderr.String()
}

func startBootstrap(t *testing.T, toolsRoot, cacheRoot string, envOverrides []string, extraEnv ...string) *bootstrapProcess {
	t.Helper()
	entrypoint, err := filepath.Abs(bootstrapScript)
	if err != nil {
		t.Fatal(err)
	}
	return startBootstrapAt(t, entrypoint, toolsRoot, cacheRoot, envOverrides, extraEnv...)
}

func startBootstrapAt(t *testing.T, entrypoint, toolsRoot, cacheRoot string, envOverrides []string, extraEnv ...string) *bootstrapProcess {
	t.Helper()
	cmd := exec.Command("bash", entrypoint)
	cmd.Env = append(os.Environ(),
		"CODEX_BUILD_TOOLS_ROOT="+toolsRoot,
		"CODEX_BUILD_CACHE_ROOT="+cacheRoot,
	)
	cmd.Env = append(cmd.Env, envOverrides...)
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	proc := &bootstrapProcess{cmd: cmd, done: make(chan error, 1)}
	cmd.Stdout = &proc.stdout
	cmd.Stderr = &proc.stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() { proc.done <- cmd.Wait() }()
	return proc
}

func waitForFile(t *testing.T, path string, proc *bootstrapProcess) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		select {
		case proc.waitErr = <-proc.done:
			proc.finished = true
			t.Fatalf("bootstrap exited before fixture readiness: %v; output=%s", proc.waitErr, proc.output())
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fixture did not start within bound: %s", path)
}

func waitBootstrap(t *testing.T, proc *bootstrapProcess, timeout time.Duration, status int) error {
	t.Helper()
	if proc.finished {
		return checkBootstrapExit(proc.waitErr, status)
	}
	select {
	case proc.waitErr = <-proc.done:
		proc.finished = true
		return checkBootstrapExit(proc.waitErr, status)
	case <-time.After(timeout):
		return fmt.Errorf("bootstrap did not exit within %s", timeout)
	}
}

func checkBootstrapExit(err error, status int) error {
	if status == 0 {
		if err == nil {
			return nil
		}
		return fmt.Errorf("bootstrap exit = %v; want status 0", err)
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != status {
		return fmt.Errorf("bootstrap exit = %v; want status %d", err, status)
	}
	return nil
}

func stopBootstrapFixture(proc *bootstrapProcess, childPID, childPGID *int, pidFiles ...string) {
	if *childPID <= 0 {
		for _, path := range pidFiles {
			if _, err := os.Stat(path); err == nil {
				if data, readErr := os.ReadFile(path); readErr == nil {
					*childPID, _ = strconv.Atoi(strings.TrimSpace(string(data)))
					break
				}
			}
		}
	}
	if *childPID > 0 && *childPGID <= 0 {
		*childPGID, _ = syscall.Getpgid(*childPID)
	}
	if *childPGID > 0 {
		_ = syscall.Kill(-*childPGID, syscall.SIGKILL)
	}
	if *childPID > 0 {
		_ = syscall.Kill(*childPID, syscall.SIGKILL)
	}
	if proc.finished {
		return
	}
	_ = syscall.Kill(-proc.cmd.Process.Pid, syscall.SIGKILL)
	_ = proc.cmd.Process.Kill()
	select {
	case proc.waitErr = <-proc.done:
		proc.finished = true
	case <-time.After(2 * time.Second):
	}
}

func readPID(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid PID fixture %q: %v", data, err)
	}
	return pid
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func processRunning(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return processAlive(pid)
	}
	fields := strings.Fields(string(data))
	return len(fields) < 3 || fields[2] != "Z"
}
