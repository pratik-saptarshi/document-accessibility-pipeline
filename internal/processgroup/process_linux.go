//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// LeaderExitedNoReap reports whether pid has exited, without reaping it.
// The unreaped zombie keeps its PID allocated, so a later process-group
// signal cannot hit a recycled PID.
func LeaderExitedNoReap(pid int) (bool, error) {
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

// WaitForLeaderExitNoReap waits until pid has exited (or ctx is done)
// without reaping it: observe the exit, let the caller signal the group
// while the leader still owns its PID, then reap.
func WaitForLeaderExitNoReap(ctx context.Context, pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		exited, err := LeaderExitedNoReap(pid)
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

// KillGroup sends SIGKILL to the process group led by pid. A group that
// no longer exists is reported as os.ErrProcessDone.
func KillGroup(pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}

// TerminateGroup sends SIGTERM to the process group led by pid, waits
// grace, then sends SIGKILL. A non-positive pid, or a group that no
// longer exists, is a no-op.
func TerminateGroup(pid int, grace time.Duration) error {
	if pid <= 0 {
		return nil
	}
	groupID := -pid
	if err := syscall.Kill(groupID, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	time.Sleep(grace)
	err := syscall.Kill(groupID, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
