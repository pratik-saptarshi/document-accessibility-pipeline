//go:build linux

package financialfixtures

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// runV2BoundedProcess runs a fixed financial-fixture helper in its own Linux
// process group. It kills remaining group members after every exit and on
// cancellation, including when a descendant inherited a captured pipe.
// Process-group cleanup does not contain descendants that deliberately detach
// into another session or provide CPU, memory, filesystem, or network limits.
func runV2BoundedProcess(ctx context.Context, name string, args, env []string, stdout, stderr io.Writer, waitDelay time.Duration) error {
	if ctx == nil {
		return errors.New("financial fixture helper requires a bounded context")
	}
	if name == "" {
		return errors.New("financial fixture helper name is required")
	}
	if waitDelay <= 0 {
		return errors.New("financial fixture helper wait delay must be positive")
	}

	command := exec.CommandContext(ctx, name, args...)
	if env != nil {
		command.Env = env
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = waitDelay
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		return killV2ProcessGroup(command.Process.Pid)
	}
	command.Stdout = stdout
	command.Stderr = stderr

	runErr := command.Run()
	if command.Process != nil {
		if err := killV2ProcessGroup(command.Process.Pid); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return errors.Join(runErr, fmt.Errorf("terminate financial fixture helper process group: %w", err))
		}
	}
	return runErr
}

func killV2ProcessGroup(pid int) error {
	if pid <= 0 {
		return errors.New("financial fixture helper process ID is invalid")
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}
