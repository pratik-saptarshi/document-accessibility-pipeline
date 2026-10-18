//go:build linux

package pdfua

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	procgroup "github.com/pratik-saptarshi/document-accessibility-pipeline/internal/procgroup"
)

const (
	validatorWaitDelay      = 500 * time.Millisecond
	processTerminationGrace = procgroup.DefaultTerminationGrace
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
	return procgroup.ConfigureTree(cmd, validatorWaitDelay, processTerminationGrace), nil
}

// waitForLeaderExitNoReap waits until pid has exited (or ctx is done) without
// reaping it. The unreaped zombie keeps its PID — and therefore the process
// group ID, since a Setpgid leader's PGID equals its PID — allocated, so a
// later process-group signal cannot hit a recycled PID. This is the same
// ordering as internal/financialfixtures/process_linux.go: observe the exit,
// signal the group while the leader still owns its PID, then reap.
func waitForLeaderExitNoReap(ctx context.Context, pid int) error {
	return procgroup.WaitForLeaderExitNoReap(ctx, pid)
}

func terminateProcessGroup(pid int) error {
	return procgroup.TerminateGroup(pid, processTerminationGrace)
}
