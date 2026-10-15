//go:build !linux

package main

import (
	"context"
	"os"
	"os/exec"
)

// awaitLeaderExit preserves the previous reap-then-drain order where process
// groups are unavailable: it reaps with cmd.Wait and reports true with the
// result, so the caller does not wait again.
func awaitLeaderExit(_ context.Context, cmd *exec.Cmd) (bool, error) {
	return true, cmd.Wait()
}

// The validator container job runs on Linux. This fallback still bounds
// command and pipe waits on other hosts, where process-group cleanup is not
// available through the standard library's cross-platform process API.
func configureProcessTree(cmd *exec.Cmd) func() error {
	cmd.WaitDelay = commandPipeWaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return cmd.Process.Kill()
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		return cmd.Process.Kill()
	}
}
