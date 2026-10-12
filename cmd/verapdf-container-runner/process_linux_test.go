//go:build linux

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func TestRunBoundedCommandCancellationKillsPipeHoldingDescendant(t *testing.T) {
	setChildSubreaper(t)
	pidPath := filepath.Join(t.TempDir(), "descendant.pid")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	stdout, stderr := filepath.Join(t.TempDir(), "stdout"), filepath.Join(t.TempDir(), "stderr")
	started := time.Now()
	status, err := runBoundedCommand(ctx, "sh", []string{"-c", "sleep 30 & child=$!; echo $child > \"$1\"; wait", "sh", pidPath}, stdout, stderr, 1024)
	if !errors.Is(err, errExecution) {
		t.Fatalf("runBoundedCommand() = %d, %v; want cancellation execution error", status, err)
	}
	if elapsed := time.Since(started); elapsed > 1500*time.Millisecond {
		t.Fatalf("cancellation took %s; want bounded process and pipe cleanup", elapsed)
	}
	assertProcessStopped(t, pidPath)
}

func TestRunBoundedCommandDoesNotAcceptFindingWithIncompletePipes(t *testing.T) {
	setChildSubreaper(t)
	pidPath := filepath.Join(t.TempDir(), "descendant.pid")
	stdout, stderr := filepath.Join(t.TempDir(), "stdout"), filepath.Join(t.TempDir(), "stderr")
	started := time.Now()
	status, err := runBoundedCommand(context.Background(), "sh", []string{"-c", "sleep 30 & child=$!; echo report; echo $child > \"$1\"; exit 1", "sh", pidPath}, stdout, stderr, 1024)
	if !errors.Is(err, errExecution) {
		t.Fatalf("runBoundedCommand() = %d, %v; incomplete output with child exit 1 must be an execution error", status, err)
	}
	if status != 1 {
		t.Fatalf("runBoundedCommand() status = %d; expected child exit status to remain observable", status)
	}
	if elapsed := time.Since(started); elapsed > 1500*time.Millisecond {
		t.Fatalf("pipe cleanup took %s; want bounded drain and process cleanup", elapsed)
	}
	assertProcessStopped(t, pidPath)
}

func assertProcessStopped(t *testing.T, pidPath string) {
	t.Helper()
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("descendant did not start: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		t.Fatalf("invalid descendant pid %q: %v", pidBytes, err)
	}
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		var status syscall.WaitStatus
		waited, waitErr := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
		if waitErr == nil && waited == pid {
			return
		}
		if waitErr != nil && errors.Is(waitErr, syscall.ECHILD) {
			if killErr := syscall.Kill(pid, 0); errors.Is(killErr, syscall.ESRCH) {
				return
			}
		}
		if time.Now().After(deadline) {
			if waitErr == nil {
				_ = syscall.Kill(pid, syscall.SIGKILL)
				_, _ = syscall.Wait4(pid, &status, 0, nil)
			}
			t.Fatalf("pipe-holding descendant %d survived command cleanup: %v", pid, waitErr)
		}
		time.Sleep(10 * time.Millisecond)
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
