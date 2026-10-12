//go:build linux

package pdfua

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
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

func configureProcessTree(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = validatorWaitDelay
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		groupID := -cmd.Process.Pid
		if err := syscall.Kill(groupID, syscall.SIGTERM); err != nil {
			if errors.Is(err, syscall.ESRCH) {
				return os.ErrProcessDone
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
	return nil
}
