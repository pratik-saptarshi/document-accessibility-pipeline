//go:build linux

package main

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

const processTerminationGrace = 200 * time.Millisecond

func configureProcessTree(cmd *exec.Cmd) func() error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = commandPipeWaitDelay
	var cleanupMu sync.Mutex
	terminate := func(pid int) error {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		return terminateProcessGroup(pid)
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := terminate(cmd.Process.Pid)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		return terminate(cmd.Process.Pid)
	}
}

func terminateProcessGroup(pid int) error {
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
	time.Sleep(processTerminationGrace)
	err := syscall.Kill(groupID, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

// awaitLeaderExit waits for the command leader's exit without reaping it and
// reports false, so the caller reaps with cmd.Wait after the drain loop's
// process-group signals (closing the PID-reuse window). A wait failure still
// falls through to those signals, which are the safety net; cancellation is
// reported by the caller's context checks.
func awaitLeaderExit(ctx context.Context, cmd *exec.Cmd) (bool, error) {
	_ = waitForLeaderExitNoReap(ctx, cmd.Process.Pid)
	return false, nil
}

// waitForLeaderExitNoReap waits until pid has exited (or ctx is done) without
// reaping it. The unreaped zombie keeps its PID — and therefore the process
// group ID, since a Setpgid leader's PGID equals its PID — allocated, so a
// later process-group signal cannot hit a recycled PID. This is the same
// ordering as internal/financialfixtures/process_linux.go: observe the exit,
// signal the group while the leader still owns its PID, then reap.
func waitForLeaderExitNoReap(ctx context.Context, pid int) error {
	if pid <= 0 {
		return errors.New("invalid process ID")
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
