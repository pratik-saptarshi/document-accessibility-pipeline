//go:build !linux

package cloudtools

import (
	"context"
	"os/exec"
	"time"
)

// The manifest only supports Linux amd64. This fallback keeps the package
// buildable for source inspection on other hosts, with bounded pipe cleanup.
func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	return cmd
}
