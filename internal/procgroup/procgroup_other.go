//go:build !linux

package procgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

// errNoProcessGroups reports that the unified process-group cleanup policy
// needs Linux process groups.
var errNoProcessGroups = errors.New("procgroup: process-group cleanup requires Linux")

// ConfigureTree keeps the package buildable where process groups are
// unavailable: it bounds pipe waits and kills only the direct child.
func ConfigureTree(cmd *exec.Cmd, waitDelay, _ time.Duration) (cleanup func() error) {
	cmd.WaitDelay = waitDelay
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

// SetPrivateGroup has no effect where process groups are unavailable.
func SetPrivateGroup(_ *exec.Cmd) {}

// WaitForLeaderExitNoReap is unreachable where process groups are
// unavailable.
func WaitForLeaderExitNoReap(_ context.Context, _ int) error {
	return errNoProcessGroups
}

// TerminateGroup is unreachable where process groups are unavailable.
func TerminateGroup(_ int, _ time.Duration) error {
	return errNoProcessGroups
}
