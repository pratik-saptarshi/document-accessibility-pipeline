//go:build linux

// Package procgroup provides the repository's single Linux process-group
// cleanup policy for child processes: each child starts in a private process
// group, the leader's exit is observed without reaping it (so the zombie
// keeps its PID — and therefore the process-group ID — allocated), the group
// is signaled while the leader still owns its PID, and only then is the
// leader reaped. Termination is SIGTERM, a grace period, then SIGKILL; a
// missing group (ESRCH) is treated as success, making cleanup idempotent.
//
// The policy intentionally does not cover descendants that deliberately
// detach into another session, nor CPU, memory, filesystem, or network
// limits.
package procgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// DefaultTerminationGrace is the SIGTERM grace period before SIGKILL in the
// unified termination policy.
const DefaultTerminationGrace = 200 * time.Millisecond

// SetPrivateGroup puts cmd in a new private process group when it starts, so
// group signals reach the child and its descendants but nothing else.
func SetPrivateGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// ConfigureTree puts cmd in a private process group and installs a
// mutex-guarded Cancel plus a cleanup function. Both terminate the group
// with SIGTERM, wait the grace period, then SIGKILL; an already-gone group
// is success. Callers observe the leader's exit with WaitForLeaderExitNoReap
// and reap with cmd.Wait after the group signal.
func ConfigureTree(cmd *exec.Cmd, waitDelay, grace time.Duration) (cleanup func() error) {
	SetPrivateGroup(cmd)
	cmd.WaitDelay = waitDelay
	var mu sync.Mutex
	terminate := func(pid int) error {
		mu.Lock()
		defer mu.Unlock()
		return TerminateGroup(pid, grace)
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return terminate(cmd.Process.Pid)
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		return terminate(cmd.Process.Pid)
	}
}

// WaitForLeaderExitNoReap blocks until pid exits (or ctx ends) without
// reaping it. The unreaped zombie keeps its PID — and therefore the process
// group ID, since a Setpgid leader's PGID equals its PID — allocated, so a
// later process-group signal cannot hit a recycled PID.
func WaitForLeaderExitNoReap(ctx context.Context, pid int) error {
	if pid <= 0 {
		return errors.New("procgroup: invalid process ID")
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		exited, err := leaderExitedNoReap(pid)
		if err != nil {
			return err
		}
		if exited {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// TerminateGroup sends SIGTERM to the process group -pid, waits the grace
// period, then sends SIGKILL. A missing group (ESRCH) is not an error.
func TerminateGroup(pid int, grace time.Duration) error {
	if pid <= 0 {
		return nil
	}
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	if grace > 0 {
		time.Sleep(grace)
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	return nil
}

func leaderExitedNoReap(pid int) (bool, error) {
	var info unix.Siginfo
	for {
		err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOHANG|unix.WNOWAIT, nil)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return false, err
		}
		return info.Signo != 0, nil
	}
}
