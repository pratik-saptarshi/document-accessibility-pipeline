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

	"golang.org/x/sys/unix"
)

var errV2HelperReapTimeout = errors.New("financial fixture helper reap timed out")

// runV2BoundedProcess runs a fixed financial-fixture helper in its own Linux
// process group. It observes the leader's exit without reaping it, kills the
// group while the leader still owns its PID/PGID, and then reaps the leader.
// Output writers retain only their configured prefix while continuing to drain
// pipes. Process-group cleanup does not contain descendants that deliberately
// detach into another session or provide CPU, memory, filesystem, or network
// limits.
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
	if err := ctx.Err(); err != nil {
		return err
	}

	command := exec.Command(name, args...)
	if env != nil {
		command.Env = env
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = waitDelay
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		return err
	}

	pid := command.Process.Pid
	return finishV2StartedProcess(ctx, pid, waitDelay,
		waitForV2LeaderOrCancel,
		killV2ProcessGroup,
		command.Process.Kill,
		command.Wait,
	)
}

func finishV2StartedProcess(ctx context.Context, pid int, reapTimeout time.Duration, wait func(context.Context, int) error, killGroup func(int) error, killLeader func() error, reap func() error) error {
	waitErr := wait(ctx, pid)
	// The leader remains an unreaped zombie after waitid(WNOWAIT), preventing
	// PID/PGID reuse until this final group signal completes.
	killErr := killGroup(pid)
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		leaderErr := killLeader()
		if leaderErr != nil && !errors.Is(leaderErr, os.ErrProcessDone) {
			killErr = errors.Join(killErr, fmt.Errorf("terminate financial fixture helper leader: %w", leaderErr))
		}
	}
	reapErr := reapWithin(reap, reapTimeout)
	if waitErr != nil {
		waitErr = fmt.Errorf("observe financial fixture helper exit: %w", waitErr)
	}
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		killErr = fmt.Errorf("terminate financial fixture helper process group: %w", killErr)
	}
	return errors.Join(waitErr, killErr, reapedCommandError(reapErr, ctx))
}

func reapWithin(reap func() error, timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("financial fixture helper reap timeout must be positive")
	}
	done := make(chan error, 1)
	go func() { done <- reap() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		return fmt.Errorf("%w after %s waiting to reap financial fixture helper", errV2HelperReapTimeout, timeout)
	}
}

// waitForV2LeaderOrCancel uses WNOWAIT so the child remains unreaped, and its
// numeric process-group identity cannot be recycled, until cleanup signals it.
func waitForV2LeaderOrCancel(ctx context.Context, pid int) error {
	if pid <= 0 {
		return errors.New("financial fixture helper process ID is invalid")
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		exited, err := observeV2LeaderExit(pid)
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

func observeV2LeaderExit(pid int) (bool, error) {
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

func reapedCommandError(err error, ctx context.Context) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errV2HelperReapTimeout) {
		return err
	}
	// Cancellation is already represented by the caller's context and remains
	// distinguishable from the expected SIGKILL exit used for cleanup.
	if ctx.Err() != nil {
		return nil
	}
	return err
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
