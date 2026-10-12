//go:build !linux

package main

import (
	"os"
	"os/exec"
)

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
