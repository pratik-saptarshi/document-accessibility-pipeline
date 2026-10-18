//go:build linux

package main

import (
	"context"
	"os/exec"

	procgroup "github.com/pratik-saptarshi/document-accessibility-pipeline/internal/procgroup"
)

// processTerminationGrace follows the unified process-group termination
// policy: SIGTERM, a short grace, then SIGKILL.
const processTerminationGrace = procgroup.DefaultTerminationGrace

func configureProcessTree(cmd *exec.Cmd) func() error {
	return procgroup.ConfigureTree(cmd, commandPipeWaitDelay, processTerminationGrace)
}

// awaitLeaderExit waits for the command leader's exit without reaping it and
// reports false, so the caller reaps with cmd.Wait after the drain loop's
// process-group signals (closing the PID-reuse window). A wait failure still
// falls through to those signals, which are the safety net; cancellation is
// reported by the caller's context checks.
func awaitLeaderExit(ctx context.Context, cmd *exec.Cmd) (bool, error) {
	_ = procgroup.WaitForLeaderExitNoReap(ctx, cmd.Process.Pid)
	return false, nil
}
