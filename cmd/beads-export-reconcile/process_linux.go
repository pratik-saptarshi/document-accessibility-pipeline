//go:build linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const commandWaitDelay = 250 * time.Millisecond

func runCommandProcess(ctx context.Context, cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = commandWaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return killCommandProcessGroup(cmd.Process.Pid)
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	observeErr := waitCommandLeaderExit(ctx, pid)
	killErr := killCommandProcessGroup(pid)
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return errors.Join(ctx.Err(), killCommandError(killErr), waitErr)
	}
	return errors.Join(observeErr, killCommandError(killErr), waitErr)
}

// waitCommandLeaderExit observes the leader without reaping it. Keeping the
// zombie unreaped prevents its PID, and therefore its process-group ID, from
// being reused before the process group has been terminated.
func waitCommandLeaderExit(ctx context.Context, pid int) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var info unix.Siginfo
		err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOHANG|unix.WNOWAIT, nil)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Signo != 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func killCommandProcessGroup(pid int) error {
	if pid <= 0 {
		return errors.New("invalid command process ID")
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}

func killCommandError(err error) error {
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}
