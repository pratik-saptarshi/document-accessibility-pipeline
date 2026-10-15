//go:build !linux

package main

import (
	"context"
	"os/exec"
	"time"
)

const commandWaitDelay = 250 * time.Millisecond

func runCommandProcess(_ context.Context, cmd *exec.Cmd) error {
	cmd.WaitDelay = commandWaitDelay
	return cmd.Run()
}
