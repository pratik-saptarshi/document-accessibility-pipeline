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

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/processgroup"
)

const processTerminationGrace = 200 * time.Millisecond

func configureProcessTree(cmd *exec.Cmd) func() error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = commandPipeWaitDelay
	var cleanupMu sync.Mutex
	terminate := func(pid int) error {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		return processgroup.TerminateGroup(pid, processTerminationGrace)
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

// awaitLeaderExit waits for the command leader's exit without reaping it and
// reports false, so the caller reaps with cmd.Wait after the drain loop's
// process-group signals (closing the PID-reuse window). A wait failure still
// falls through to those signals, which are the safety net; cancellation is
// reported by the caller's context checks. The no-reap observation itself is
// processgroup.WaitForLeaderExitNoReap: observe the exit, signal the group
// while the leader still owns its PID, then reap.
func awaitLeaderExit(ctx context.Context, cmd *exec.Cmd) (bool, error) {
	_ = processgroup.WaitForLeaderExitNoReap(ctx, cmd.Process.Pid)
	return false, nil
}
