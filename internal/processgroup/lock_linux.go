//go:build linux

package processgroup

import (
	"context"
	"errors"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// waitForLockChange blocks until the lock directory named name inside
// parent may have been removed, remaining elapses, or ctx is done. The
// wakeup is an inotify event on the parent; poll(2) timeouts only bound
// the wait and keep it responsive to ctx — the removal event itself
// wakes the poll immediately. If inotify cannot be set up, it falls
// back to a bounded poll, here inside the lock home only.
func waitForLockChange(ctx context.Context, parent, name string, remaining time.Duration) error {
	fd, err := unix.InotifyInit()
	if err == nil {
		_, err = unix.InotifyAddWatch(fd, parent, unix.IN_DELETE|unix.IN_MOVED_FROM)
	}
	if err != nil {
		return pollLockWait(ctx, remaining)
	}
	defer unix.Close(fd)
	deadline := time.Now().Add(remaining)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return nil
		}
		chunk := left
		if chunk > 50*time.Millisecond {
			chunk = 50 * time.Millisecond
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, perr := unix.Poll(fds, int(chunk/time.Millisecond))
		if perr != nil && !errors.Is(perr, syscall.EINTR) {
			return pollLockWait(ctx, left)
		}
		if n > 0 {
			// Some entry in the parent changed; the acquire loop
			// retries the mkdir and re-waits if it was not ours.
			return nil
		}
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
	}
}

// pollLockWait is the no-inotify fallback: a bounded poll in chunks,
// used only inside the lock homes.
func pollLockWait(ctx context.Context, remaining time.Duration) error {
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
