//go:build linux

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	procgroup "github.com/pratik-saptarshi/document-accessibility-pipeline/internal/procgroup"
)

const commandWaitDelay = 250 * time.Millisecond

// commandTerminationGrace follows the unified process-group termination
// policy: SIGTERM, a short grace, then SIGKILL.
const commandTerminationGrace = procgroup.DefaultTerminationGrace

func runCommandProcess(ctx context.Context, cmd *exec.Cmd) error {
	procgroup.SetPrivateGroup(cmd)
	cmd.WaitDelay = commandWaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return procgroup.TerminateGroup(cmd.Process.Pid, commandTerminationGrace)
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	observeErr := procgroup.WaitForLeaderExitNoReap(ctx, pid)
	killErr := procgroup.TerminateGroup(pid, commandTerminationGrace)
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return errors.Join(ctx.Err(), killErr, waitErr)
	}
	return errors.Join(observeErr, killErr, waitErr)
}
