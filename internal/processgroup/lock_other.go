//go:build !linux

package processgroup

import (
	"context"
	"time"
)

// waitForLockChange is the off-Linux DirLock wait: no inotify exists
// here, so every wait is the shared one-chunk fallback poll from
// lock.go — the acquire loop retries its mkdir between chunks, so a
// release is observed at chunk scale rather than at maxWait scale.
// Locking has no Linux-specific safety property, so — unlike the
// lifecycle helpers — it does not fail closed off Linux.
func waitForLockChange(ctx context.Context, parent, name string, remaining time.Duration) error {
	return pollLockWait(ctx, remaining)
}
