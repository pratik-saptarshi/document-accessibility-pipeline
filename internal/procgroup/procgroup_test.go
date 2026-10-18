//go:build linux

package procgroup_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	procgroup "github.com/pratik-saptarshi/document-accessibility-pipeline/internal/procgroup"
)

// TestWaitForLeaderExitNoReapKeepsZombieUntilGroupSignal pins the unified
// ordering invariant: observe the leader's exit without reaping it, signal
// the process group while the zombie still owns its PID/PGID, then reap.
func TestWaitForLeaderExitNoReapKeepsZombieUntilGroupSignal(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "sh", "-c", "exit 0")
	cleanup := procgroup.ConfigureTree(cmd, time.Second, 50*time.Millisecond)
	defer func() {
		// Idempotent: the group is already gone after the test's own cleanup.
		_ = cleanup()
	}()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start leader: %v", err)
	}
	leader := cmd.Process.Pid
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := procgroup.WaitForLeaderExitNoReap(ctx, leader); err != nil {
		t.Fatalf("WaitForLeaderExitNoReap: %v", err)
	}
	if state := procStatState(t, leader); state != "Z" {
		t.Fatalf("leader %d not a zombie after no-reap wait: state %q", leader, state)
	}
	if err := procgroup.TerminateGroup(leader, 50*time.Millisecond); err != nil {
		t.Fatalf("TerminateGroup: %v", err)
	}
	var status syscall.WaitStatus
	if _, err := syscall.Wait4(leader, &status, 0, nil); err != nil {
		t.Fatalf("reap leader: %v", err)
	}
}

// TestTerminateGroupKillsLiveProcessGroup verifies SIGTERM-then-SIGKILL
// cleanup of a running private group.
func TestTerminateGroupKillsLiveProcessGroup(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "sleep", "60")
	procgroup.ConfigureTree(cmd, time.Second, 50*time.Millisecond)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	leader := cmd.Process.Pid
	if err := procgroup.TerminateGroup(leader, 50*time.Millisecond); err != nil {
		t.Fatalf("TerminateGroup: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("leader survived process-group termination")
	}
	if err := syscall.Kill(leader, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("leader pid %d still exists after termination", leader)
	}
}

// TestTerminateGroupIsIdempotent verifies a missing group is not an error.
func TestTerminateGroupIsIdempotent(t *testing.T) {
	if err := procgroup.TerminateGroup(0, time.Second); err != nil {
		t.Fatalf("TerminateGroup(0) = %v, want nil", err)
	}
	// An unallocated PID reports ESRCH, which cleanup treats as success.
	if err := procgroup.TerminateGroup(1<<22, time.Second); err != nil {
		t.Fatalf("TerminateGroup(unallocated) = %v, want nil", err)
	}
}

// TestWaitForLeaderExitNoReapRejectsInvalidPID verifies fail-closed input
// validation.
func TestWaitForLeaderExitNoReapRejectsInvalidPID(t *testing.T) {
	if err := procgroup.WaitForLeaderExitNoReap(context.Background(), 0); err == nil {
		t.Fatal("WaitForLeaderExitNoReap(0) = nil, want an error")
	}
}

func procStatState(t *testing.T, pid int) string {
	t.Helper()
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		t.Fatalf("read /proc/%d/stat: %v", pid, err)
	}
	rest := string(data)
	i := strings.LastIndex(rest, ")")
	if i < 0 {
		t.Fatalf("/proc/%d/stat has no closing paren: %q", pid, string(data))
	}
	fields := strings.Fields(rest[i+1:])
	if len(fields) == 0 {
		t.Fatalf("/proc/%d/stat has no state field", pid)
	}
	return fields[0]
}
