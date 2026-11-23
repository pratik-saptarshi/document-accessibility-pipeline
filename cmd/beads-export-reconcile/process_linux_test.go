//go:build linux

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestClassifyProcessStateReadError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    string
		wantErr bool
	}{
		{name: "missing proc entry", err: &os.PathError{Op: "read", Path: "/proc/1/stat", Err: syscall.ENOENT}, want: "gone"},
		{name: "process exited during read", err: &os.PathError{Op: "read", Path: "/proc/1/stat", Err: syscall.ESRCH}, want: "gone"},
		{name: "permission denied", err: &os.PathError{Op: "read", Path: "/proc/1/stat", Err: fs.ErrPermission}, wantErr: true},
		{name: "unrelated read error", err: &os.PathError{Op: "read", Path: "/proc/1/stat", Err: syscall.EIO}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := classifyProcessStateReadError(tc.err)
			if (err != nil) != tc.wantErr {
				t.Fatalf("classifyProcessStateReadError() error = %v, wantErr %t", err, tc.wantErr)
			}
			if tc.wantErr {
				if !errors.Is(err, tc.err) {
					t.Fatalf("classifyProcessStateReadError() error = %v, want wrapped %v", err, tc.err)
				}
				return
			}
			if got != tc.want {
				t.Fatalf("classifyProcessStateReadError() state = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunCommandTerminatesDescendantsAfterLeaderExit(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "descendant.pid")
	env := append(os.Environ(), "DAP_BRECONCILE_PROCESS_HELPER_PID_FILE="+pidPath, "DAP_BRECONCILE_PROCESS_HELPER_MODE=exit")
	started := time.Now()
	_, _, err := runCommand(os.Args[0], []string{"-test.run=^TestCommandProcessTreeHelper$"}, "", env)
	if err != nil {
		t.Fatalf("runCommand() error = %v, want successful leader exit", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("runCommand() waited %s for a descendant-held output pipe, want prompt process-group cleanup", elapsed)
	}
	assertHelperDescendantStopped(t, pidPath)
}

func TestRunCommandTerminatesDescendantsOnTimeout(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "descendant.pid")
	env := append(os.Environ(), "DAP_BRECONCILE_PROCESS_HELPER_PID_FILE="+pidPath, "DAP_BRECONCILE_PROCESS_HELPER_MODE=wait")
	// The deadline must leave room for the helper process tree to start and
	// record the descendant PID even under full-suite CPU contention: the
	// group SIGKILL lands at the deadline whether or not the helper's shell
	// has written the PID file yet, so a 100 ms deadline races startup
	// rather than testing cleanup. What is asserted is the deadline error,
	// the bounded cleanup after the deadline, and the descendant's death.
	const timeout = 2 * time.Second
	started := time.Now()
	_, _, err := runCommandWithTimeout(os.Args[0], []string{"-test.run=^TestCommandProcessTreeHelper$"}, "", env, timeout)
	if err == nil {
		t.Fatal("runCommandWithTimeout() error = nil, want deadline error")
	}
	// The return is pinned to [timeout, timeout + 1 s]: the context is
	// created after started and its deadline never fires early, and the
	// helper cannot exit on its own before it, so an earlier return can
	// only be a premature cancellation or an unrelated early failure -
	// neither is the deadline error this test exists to observe. The
	// adapter timeout test asserts the same lower bound.
	elapsed := time.Since(started)
	if elapsed < timeout {
		t.Fatalf("runCommandWithTimeout() returned after %s, before the %s deadline expired: premature cancellation is not a timeout", elapsed, timeout)
	}
	if elapsed > timeout+time.Second {
		t.Fatalf("runCommandWithTimeout() waited %s after deadline, want bounded process and pipe cleanup", elapsed)
	}
	assertHelperDescendantStopped(t, pidPath)
}

// descendantObservationBound is a failure bound for observing helper
// process setup and teardown, not a success criterion: on a healthy
// machine the PID file exists and the descendant is observed stopped
// almost immediately, but under full-suite CPU contention process
// startup, file writes, and SIGKILL state transitions all slow down, so
// the observations below poll until this bound instead of racing a
// fixed short window.
const descendantObservationBound = 15 * time.Second

func assertHelperDescendantStopped(t *testing.T, pidPath string) {
	t.Helper()
	pid := readDescendantPID(t, pidPath)
	// Failure-path cleanup: if the descendant is still alive when the
	// test ends - the failure this assertion exists to report - kill it
	// directly so a regression run does not leak a long-lived sleeper
	// into the rest of the suite. On a healthy run the descendant is
	// already dead and this is a no-op. PID reuse within one test's
	// lifetime is not a practical concern: it would require the kernel
	// to cycle the whole PID space before the test ends.
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	deadline := time.Now().Add(descendantObservationBound)
	for {
		if state := processState(t, pid); state != "live" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("descendant process %d remains alive after SIGKILL observation deadline", pid)
		}
		// SIGKILL delivery is asynchronous. Allow the kernel to transition the
		// descendant to an exited or zombie state before asserting cleanup.
		time.Sleep(5 * time.Millisecond)
	}
}

func readDescendantPID(t *testing.T, pidPath string) int {
	t.Helper()
	deadline := time.Now().Add(descendantObservationBound)
	for {
		pidBytes, err := os.ReadFile(pidPath)
		if err != nil {
			if time.Now().After(deadline) {
				t.Fatalf("read descendant PID: %v", err)
			}
		} else if pid, atoiErr := strconv.Atoi(strings.TrimSpace(string(pidBytes))); atoiErr == nil && pid > 0 {
			return pid
		} else if time.Now().After(deadline) {
			t.Fatalf("invalid descendant PID %q", pidBytes)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// The descendant's lifetime must exceed the scenario deadline plus the
// observation bound (2 s + 15 s): if it could exit naturally inside the
// observation window, a cleanup regression that left it alive would be
// indistinguishable from a successful process-group kill, because the
// assertion accepts any state other than "live". Thirty seconds matches
// the financialfixtures helpers and puts the descendant's natural exit
// far beyond the window's end; with correct cleanup it is killed at the
// deadline or leader exit, long before its lifetime matters.
func TestCommandProcessTreeHelper(t *testing.T) {
	pidPath := os.Getenv("DAP_BRECONCILE_PROCESS_HELPER_PID_FILE")
	if pidPath == "" {
		return
	}
	mode := os.Getenv("DAP_BRECONCILE_PROCESS_HELPER_MODE")
	command := exec.Command("/bin/sh", "-c", `sleep 30 & echo $! > "$1"; if [ "$2" = wait ]; then wait; else exit 0; fi`, "sh", pidPath, mode)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("start descendant: %v", err)
	}
}

func processState(t *testing.T, pid int) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		state, classifyErr := classifyProcessStateReadError(err)
		if classifyErr != nil {
			t.Fatalf("read descendant process state: %v", classifyErr)
		}
		return state
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		t.Fatalf("malformed /proc process state %q", data)
	}
	if fields[2] == "Z" || fields[2] == "X" {
		return "dead"
	}
	return "live"
}

func classifyProcessStateReadError(err error) (string, error) {
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ESRCH) {
		return "gone", nil
	}
	return "", fmt.Errorf("read process state: %w", err)
}
