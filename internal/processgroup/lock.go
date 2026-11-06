package processgroup

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// A DirLock is a held directory lock: the lock is the existence of the
// directory at the lock path, taken with a single mkdir. The holder
// writes its pid into the directory's "pid" file for diagnostics.
//
// The wait for a held lock is event-based where the platform allows
// (Linux: inotify on the parent directory, watching for the lock
// directory's removal). The synchronization event is the holder
// REMOVING the directory at Release — never the holder's death: a dead
// holder's directory stays in place and waiters time out with guidance
// text. That stale-lock operator contract ("inspect and remove it only
// if the holder is not running") is deliberate and shared by every
// implementation of this contract.
type DirLock struct {
	path string

	mu       sync.Mutex
	released bool
}

// AcquireDirLock takes the mkdir-exclusive lock at path, waiting up to
// maxWait while another holder holds it; maxWait == 0 makes a single
// attempt (fail-fast). Waiting is bounded by maxWait and by ctx,
// whichever ends first. On success the holder's pid is recorded in
// path/pid; Release removes the pid file and the directory.
func AcquireDirLock(ctx context.Context, path string, maxWait time.Duration) (*DirLock, error) {
	deadline := time.Now().Add(maxWait)
	for {
		err := os.Mkdir(path, 0o755)
		if err == nil {
			l := &DirLock{path: path}
			pidPath := filepath.Join(path, "pid")
			content := strconv.Itoa(os.Getpid()) + "\n"
			if werr := os.WriteFile(pidPath, []byte(content), 0o644); werr != nil {
				_ = os.Remove(path)
				return nil, fmt.Errorf("processgroup: record lock holder pid in %s: %w", pidPath, werr)
			}
			return l, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("processgroup: directory lock %s still held after %s; inspect and remove it only if the holder is not running", path, maxWait)
		}
		if cerr := ctx.Err(); cerr != nil {
			return nil, cerr
		}
		if werr := waitForLockChange(ctx, filepath.Dir(path), filepath.Base(path), remaining); werr != nil {
			return nil, werr
		}
	}
}

// Release removes the lock's pid file and directory. It is idempotent.
func (l *DirLock) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return nil
	}
	l.released = true
	pidErr := os.Remove(filepath.Join(l.path, "pid"))
	dirErr := os.Remove(l.path)
	if dirErr != nil && !errors.Is(dirErr, fs.ErrNotExist) {
		return dirErr
	}
	if pidErr != nil && !errors.Is(pidErr, fs.ErrNotExist) {
		return pidErr
	}
	return nil
}

// sleepLockChunk waits for d or ctx cancellation, whichever is first.
// It is the bounded-poll building block used only inside the lock
// homes: as the fallback when no event source is available, and to
// keep waits responsive to ctx.
func sleepLockChunk(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
