//go:build linux

package pdfua

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func init() {
	if os.Getenv("PDFUA_FAKE_PIPE_HOLDER") == "1" {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM)
		if path := os.Getenv("PDFUA_FAKE_DESCENDANT_PID_FILE"); path != "" {
			_ = os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o600)
		}
		<-stop
		os.Exit(0)
	}
	if os.Getenv("PDFUA_FAKE_SPAWN_PIPE_HOLDER") == "1" {
		helper := exec.Command(os.Args[0], "-test.run=^$")
		helper.Env = append(os.Environ(), "PDFUA_FAKE_PIPE_HOLDER=1")
		helper.Stdout, helper.Stderr = os.Stdout, os.Stderr
		if err := helper.Start(); err != nil {
			os.Exit(32)
		}
		readyDeadline := time.Now().Add(3 * time.Second)
		for {
			if _, err := os.Stat(os.Getenv("PDFUA_FAKE_DESCENDANT_PID_FILE")); err == nil {
				break
			}
			if time.Now().After(readyDeadline) {
				os.Exit(33)
			}
			time.Sleep(time.Millisecond)
		}
		if os.Getenv("PDFUA_FAKE_HOLD_LEADER") == "1" {
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM)
			<-stop
		}
		os.Exit(0)
	}
	if os.Getenv("PDFUA_FAKE_SPAWN_CLOSED_PIPE_HOLDER") == "1" {
		helper := exec.Command(os.Args[0], "-test.run=^$")
		helper.Env = append(os.Environ(), "PDFUA_FAKE_PIPE_HOLDER=1")
		// Nil child streams are redirected to the null device by os/exec, so
		// this process can outlive its parent without keeping the report pipes open.
		if err := helper.Start(); err != nil {
			os.Exit(34)
		}
		readyDeadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(os.Getenv("PDFUA_FAKE_DESCENDANT_PID_FILE")); err == nil {
				break
			}
			if time.Now().After(readyDeadline) {
				os.Exit(35)
			}
			time.Sleep(time.Millisecond)
		}
		_, _ = os.Stdout.WriteString(os.Getenv("PDFUA_FAKE_REPORT"))
		os.Exit(0)
	}
}

func TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles(t *testing.T) {
	setChildSubreaper(t)
	marker := filepath.Join(t.TempDir(), "descendant.pid")
	t.Setenv("PDFUA_FAKE_SPAWN_PIPE_HOLDER", "1")
	t.Setenv("PDFUA_FAKE_HOLD_LEADER", "1")
	t.Setenv("PDFUA_FAKE_DESCENDANT_PID_FILE", marker)
	options := fakeOptions(t, fakeArtifact(t))
	options.Timeout = 300 * time.Millisecond
	options.TempDir = t.TempDir()
	started := time.Now()
	_, err := Validate(context.Background(), options, fakePDF(t))
	elapsed := time.Since(started)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected validator timeout, got %v", err)
	}
	maximumDuration := options.Timeout + validatorWaitDelay + 200*time.Millisecond
	if elapsed > maximumDuration {
		t.Fatalf("cancellation took %s, timeout+wait bound is %s", elapsed, maximumDuration)
	}
	pidBytes, readErr := os.ReadFile(marker)
	if readErr != nil {
		t.Fatalf("descendant was not started: %v", readErr)
	}
	pid, parseErr := strconv.Atoi(string(pidBytes))
	if parseErr != nil {
		t.Fatalf("invalid descendant pid: %v", parseErr)
	}
	waitAndReapDescendant(t, pid, 500*time.Millisecond)
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("descendant process %d survived cancellation: %v", pid, err)
	}
	assertValidationTempDirEmpty(t, options.TempDir)
}

func TestValidateCleansProcessGroupAfterLeaderExits(t *testing.T) {
	setChildSubreaper(t)
	t.Run("inherited pipes outlive wait delay", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "descendant.pid")
		t.Setenv("PDFUA_FAKE_SPAWN_PIPE_HOLDER", "1")
		t.Setenv("PDFUA_FAKE_DESCENDANT_PID_FILE", marker)
		options := fakeOptions(t, fakeArtifact(t))
		options.Timeout = 5 * time.Second
		started := time.Now()
		_, err := Validate(context.Background(), options, fakePDF(t))
		if elapsed := time.Since(started); elapsed >= options.Timeout {
			t.Fatalf("early leader exit took %s, want less than timeout %s", elapsed, options.Timeout)
		}
		if err == nil || !strings.Contains(err.Error(), "invalid validator XML report") || strings.Contains(err.Error(), "timeout") {
			t.Fatalf("expected a malformed-report error without timeout after the leader exited early, got %v", err)
		}
		assertDescendantTerminated(t, marker)
		assertValidationTempDirEmpty(t, options.TempDir)
	})

	t.Run("valid report with descendant output closed", func(t *testing.T) {
		marker := filepath.Join(t.TempDir(), "descendant.pid")
		t.Setenv("PDFUA_FAKE_SPAWN_CLOSED_PIPE_HOLDER", "1")
		t.Setenv("PDFUA_FAKE_DESCENDANT_PID_FILE", marker)
		t.Setenv("PDFUA_FAKE_REPORT", report("true"))
		options := fakeOptions(t, fakeArtifact(t))
		options.Timeout = 5 * time.Second
		result, err := Validate(context.Background(), options, fakePDF(t))
		if err != nil || result.Outcome != OutcomePass {
			t.Fatalf("expected a valid report before group cleanup, result=%#v err=%v", result, err)
		}
		assertDescendantTerminated(t, marker)
	})
}

func TestWaitForLeaderExitNoReapPreservesZombieUntilGroupSignal(t *testing.T) {
	setChildSubreaper(t)
	marker := filepath.Join(t.TempDir(), "descendant.pid")
	t.Setenv("PDFUA_FAKE_SPAWN_PIPE_HOLDER", "1")
	t.Setenv("PDFUA_FAKE_DESCENDANT_PID_FILE", marker)
	// PDFUA_FAKE_HOLD_LEADER intentionally unset: the leader exits
	// immediately after descendant readiness.
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start leader: %v", err)
	}
	leader := cmd.Process.Pid
	cleaned := false
	t.Cleanup(func() {
		// Safety net for failure paths: if the no-reap wait reaps the leader
		// (the regression this test pins) or /proc inspection fails, the
		// test would return before terminating the process group, leaving
		// the pipe-holder descendant alive with the test binary's
		// stdout/stderr. Idempotent: skipped once the test reaps below.
		if cleaned {
			return
		}
		_ = terminateProcessGroup(leader)
		_, _ = syscall.Wait4(leader, nil, syscall.WNOHANG, nil)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 1. Observe leader exit WITHOUT reaping.
	if err := waitForLeaderExitNoReap(ctx, leader); err != nil {
		t.Fatalf("waitForLeaderExitNoReap: %v", err)
	}
	// 2. The leader must still be unreaped: it is a zombie (state Z in
	// /proc/<pid>/stat), observable without reaping. If a future change
	// reaps inside the wait, this assertion fails.
	if state := procStatState(t, leader); state != "Z" {
		t.Fatalf("leader %d not a zombie after no-reap wait: state %q", leader, state)
	}
	// 3. Signal the process group while the zombie leader still owns its
	// PID (PGID == leader PID for a Setpgid leader), so the signal cannot
	// hit a recycled PID.
	if err := terminateProcessGroup(leader); err != nil {
		t.Fatalf("terminateProcessGroup: %v", err)
	}
	// 4. Descendant cleanup verified.
	assertDescendantTerminated(t, marker)
	// 5. Reap the leader (test's own cleanup; production callers reap after signaling).
	var status syscall.WaitStatus
	if _, err := syscall.Wait4(leader, &status, 0, nil); err != nil {
		t.Fatalf("reap leader: %v", err)
	}
	cleaned = true
}

// procStatState returns the process state of pid from /proc/<pid>/stat:
// the first field after the LAST ')', which terminates the parenthesized
// command name (comm may itself contain ')' or spaces). It fails the test if
// /proc is unavailable — this is linux-only code, so a silent skip would hide
// a broken environment.
func procStatState(t *testing.T, pid int) string {
	t.Helper()
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		t.Fatalf("read /proc/%d/stat: %v", pid, err)
	}
	rest := string(data)
	if i := strings.LastIndex(rest, ")"); i >= 0 {
		rest = rest[i+1:]
	} else {
		t.Fatalf("/proc/%d/stat has no closing paren: %q", pid, rest)
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		t.Fatalf("/proc/%d/stat has no state field: %q", pid, string(data))
	}
	return fields[0]
}

func assertValidationTempDirEmpty(t *testing.T, tempDir string) {
	t.Helper()
	entries, err := os.ReadDir(tempDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("validation temp directory not cleaned: entries=%v err=%v", entries, err)
	}
}

func assertDescendantTerminated(t *testing.T, marker string) {
	t.Helper()
	pidBytes, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("descendant was not started: %v", err)
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		t.Fatalf("invalid descendant pid: %v", err)
	}
	waitAndReapDescendant(t, pid, 500*time.Millisecond)
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("descendant process %d survived command completion: %v", pid, err)
	}
}

func setChildSubreaper(t *testing.T) {
	t.Helper()
	var previous uintptr
	if _, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, 37, uintptr(unsafe.Pointer(&previous)), 0, 0, 0, 0); errno != 0 {
		t.Fatalf("read subreaper state: %v", errno)
	}
	if _, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, 36, 1, 0, 0, 0, 0); errno != 0 {
		t.Fatalf("enable test subreaper: %v", errno)
	}
	t.Cleanup(func() {
		_, _, _ = syscall.Syscall6(syscall.SYS_PRCTL, 36, previous, 0, 0, 0, 0)
	})
}

func waitAndReapDescendant(t *testing.T, pid int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		var status syscall.WaitStatus
		waited, err := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
		if err == nil && waited == pid {
			return
		}
		if err != nil && !errors.Is(err, syscall.EINTR) {
			t.Fatalf("reap validator descendant: %v", err)
		}
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			_, _ = syscall.Wait4(pid, &status, 0, nil)
			t.Fatalf("validator descendant %d did not exit and reap", pid)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestValidateRejectsFIFOAndSymlinkBeforeOpening(t *testing.T) {
	options := fakeOptions(t, fakeArtifact(t))
	root := t.TempDir()
	fifo := filepath.Join(root, "input.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "private-target.pdf")
	if err := os.WriteFile(target, []byte("private target content"), 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "input-link.pdf")
	if err := os.Symlink(target, symlink); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{fifo, symlink} {
		started := time.Now()
		if _, err := Validate(context.Background(), options, input); err == nil {
			t.Fatalf("accepted unsupported input type %s", filepath.Base(input))
		}
		if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
			t.Fatalf("opening unsupported input took %s", elapsed)
		}
	}
	artifactFIFO := filepath.Join(root, "validator.fifo")
	if err := syscall.Mkfifo(artifactFIFO, 0o600); err != nil {
		t.Fatal(err)
	}
	options.Artifact = artifactFIFO
	started := time.Now()
	if _, err := Validate(context.Background(), options, fakePDF(t)); err == nil {
		t.Fatal("accepted FIFO validator artifact")
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("opening unsupported artifact took %s", elapsed)
	}
	options.Artifact = symlink
	if _, err := Validate(context.Background(), options, fakePDF(t)); err == nil {
		t.Fatal("followed a symlink for validator artifact")
	}
}

func TestValidateCanceledBeforePreparationDoesNotOpenFIFO(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "input.fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	options := fakeOptions(t, fakeArtifact(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if _, err := Validate(ctx, options, fifo); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("expected preparation cancellation, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("pre-canceled validation blocked for %s", elapsed)
	}
}
