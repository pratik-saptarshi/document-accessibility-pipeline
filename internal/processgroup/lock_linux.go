//go:build linux

package processgroup

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// inotifyAddWatchFn is the watch-arming call, held in a variable so a
// test can interpose on watch setup (the same seam pattern as
// pidfdOpenFn in the session code).
var inotifyAddWatchFn = unix.InotifyAddWatch

// waitForLockChange blocks until the lock directory named name inside
// parent may have been removed, remaining elapses, or ctx is done. The
// wakeup is an inotify event on the parent; poll(2) timeouts only bound
// the wait and keep it responsive to ctx — the removal event itself
// wakes the poll immediately. If inotify cannot be set up, it falls
// back to a bounded poll, here inside the lock home only.
//
// The caller reaches this wait only after its mkdir failed, and a
// release landing between that failure and the watch being armed
// produces no event (review finding P2-2). The existence re-check
// right after arming closes that window: if the directory is already
// gone, return at once and let the caller's loop retry its mkdir
// immediately. A removal after arming is queued as an event, so no
// window remains on either side of the check.
func waitForLockChange(ctx context.Context, parent, name string, remaining time.Duration) error {
	fd, err := unix.InotifyInit()
	if err == nil {
		_, err = inotifyAddWatchFn(fd, parent, unix.IN_DELETE|unix.IN_MOVED_FROM)
	}
	if err != nil {
		return pollLockWait(ctx, remaining)
	}
	defer unix.Close(fd)
	if _, serr := os.Stat(filepath.Join(parent, name)); errors.Is(serr, fs.ErrNotExist) {
		return nil
	}
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

