//go:build linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

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
	started := time.Now()
	_, _, err := runCommandWithTimeout(os.Args[0], []string{"-test.run=^TestCommandProcessTreeHelper$"}, "", env, 100*time.Millisecond)
	if err == nil {
		t.Fatal("runCommandWithTimeout() error = nil, want deadline error")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("runCommandWithTimeout() waited %s after deadline, want bounded process and pipe cleanup", elapsed)
	}
	assertHelperDescendantStopped(t, pidPath)
}

func assertHelperDescendantStopped(t *testing.T, pidPath string) {
	t.Helper()
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("read descendant PID: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid descendant PID %q", pidBytes)
	}
	deadline := time.Now().Add(500 * time.Millisecond)
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

func TestCommandProcessTreeHelper(t *testing.T) {
	pidPath := os.Getenv("DAP_BRECONCILE_PROCESS_HELPER_PID_FILE")
	if pidPath == "" {
		return
	}
	mode := os.Getenv("DAP_BRECONCILE_PROCESS_HELPER_MODE")
	command := exec.Command("/bin/sh", "-c", `sleep 5 & echo $! > "$1"; if [ "$2" = wait ]; then wait; else exit 0; fi`, "sh", pidPath, mode)
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
		if os.IsNotExist(err) {
			return "gone"
		}
		t.Fatalf("read descendant process state: %v", err)
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
