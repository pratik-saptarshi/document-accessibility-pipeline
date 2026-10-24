//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestHelpersRejectNonPositivePIDs(t *testing.T) {
	for _, pid := range []int{0, -1, -42} {
		if err := WaitForLeaderExitNoReap(context.Background(), pid); !errors.Is(err, ErrInvalidPID) {
			t.Fatalf("WaitForLeaderExitNoReap(%d) error = %v, want ErrInvalidPID", pid, err)
		}
		if err := KillGroup(pid); !errors.Is(err, ErrInvalidPID) {
			t.Fatalf("KillGroup(%d) error = %v, want ErrInvalidPID", pid, err)
		}
		// TerminateGroup is the graceful-cleanup path: a non-positive PID
		// is a no-op there, and must never reach kill(2) either (pid 0
		// would address the caller's own process group).
		if err := TerminateGroup(pid, time.Millisecond); err != nil {
			t.Fatalf("TerminateGroup(%d) error = %v, want nil no-op", pid, err)
		}
	}
}

func TestWaitForLeaderExitNoReapObservesExitWithoutReaping(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := WaitForLeaderExitNoReap(ctx, cmd.Process.Pid); err != nil {
		t.Fatalf("WaitForLeaderExitNoReap: %v", err)
	}
	exited, err := LeaderExitedNoReap(cmd.Process.Pid)
	if err != nil || !exited {
		t.Fatalf("LeaderExitedNoReap = %v, %v; want true, nil for the unreaped leader", exited, err)
	}
	// The leader is still ours to reap: Wait must succeed exactly once.
	if err := cmd.Wait(); err != nil {
		t.Fatalf("reap observed leader: %v", err)
	}
}

func TestWaitForLeaderExitNoReapHonorsCancellation(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = KillGroup(cmd.Process.Pid)
		_ = cmd.Wait()
	}()
	exited, err := LeaderExitedNoReap(cmd.Process.Pid)
	if err != nil || exited {
		t.Fatalf("LeaderExitedNoReap = %v, %v; want false, nil for a running leader", exited, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := WaitForLeaderExitNoReap(ctx, cmd.Process.Pid); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitForLeaderExitNoReap error = %v, want context deadline", err)
	}
}

func TestKillGroupTerminatesLeaderGroup(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := KillGroup(cmd.Process.Pid); err != nil {
		t.Fatalf("KillGroup: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("killed leader exited without an error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("leader survived KillGroup")
	}
	// After the reap the group is gone; a repeat signal reports the
	// process as done rather than succeeding or failing obscurely.
	if err := KillGroup(cmd.Process.Pid); err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("KillGroup after reap = %v, want nil or os.ErrProcessDone", err)
	}
}

func TestTerminateGroupTerminatesLeaderGroup(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := TerminateGroup(cmd.Process.Pid, 10*time.Millisecond); err != nil {
		t.Fatalf("TerminateGroup: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("leader survived TerminateGroup")
	}
}
