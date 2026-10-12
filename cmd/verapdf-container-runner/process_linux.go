//go:build linux

package main

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const processTerminationGrace = 200 * time.Millisecond

func configureProcessTree(cmd *exec.Cmd) func() error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = commandPipeWaitDelay
	var cleanupMu sync.Mutex
	terminate := func(pid int) error {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		return terminateProcessGroup(pid)
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := terminate(cmd.Process.Pid)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		return terminate(cmd.Process.Pid)
	}
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
