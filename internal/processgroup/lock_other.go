//go:build !linux

package processgroup

import (
	"context"
	"time"
)

// waitForLockChange is the off-Linux DirLock wait: no inotify exists
// here, so it uses the bounded poll kept inside the lock home. Locking
// has no Linux-specific safety property, so — unlike the lifecycle
// helpers — it does not fail closed off Linux.
func waitForLockChange(ctx context.Context, parent, name string, remaining time.Duration) error {
	deadline := time.Now().Add(remaining)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return nil
		}
		chunk := left
		if chunk > 25*time.Millisecond {
			chunk = 25 * time.Millisecond
		}
		if err := sleepLockChunk(ctx, chunk); err != nil {
			return err
		}
	}
}
