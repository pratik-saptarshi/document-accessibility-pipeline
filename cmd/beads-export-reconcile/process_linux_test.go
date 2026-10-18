//go:build linux

package main

import (
	"context"
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
	_, _, err := runCommand(context.Background(), os.Args[0], []string{"-test.run=^TestCommandProcessTreeHelper$"}, "", env)
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
	_, _, err := runCommandWithTimeout(context.Background(), os.Args[0], []string{"-test.run=^TestCommandProcessTreeHelper$"}, "", env, 100*time.Millisecond)
	if err == nil {
		t.Fatal("runCommandWithTimeout() error = nil, want deadline error")
	}
	if !errors.Is(err, errCommandTimeout) {
		t.Fatalf("runCommandWithTimeout() error = %v, want errCommandTimeout", err)
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
