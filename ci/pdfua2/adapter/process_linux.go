//go:build linux

package pdfua

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

const (
	validatorWaitDelay      = 500 * time.Millisecond
	processTerminationGrace = 200 * time.Millisecond
)

func requireProcessTreeSupport() error { return nil }

func openRegularNoFollow(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), "validated-input")
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, errors.New("not a regular file")
	}
	return file, nil
}

func configureProcessTree(cmd *exec.Cmd) (func() error, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = validatorWaitDelay
	var cleanupMu sync.Mutex
	terminateGroup := func(pid int) error {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		return terminateProcessGroup(pid)
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		if err := terminateGroup(cmd.Process.Pid); errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		} else {
			return err
		}
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		return terminateGroup(cmd.Process.Pid)
	}, nil
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
