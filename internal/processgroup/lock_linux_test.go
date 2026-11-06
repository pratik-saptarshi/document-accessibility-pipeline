//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDirLockAcquireCreatesDirAndPIDFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	l, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		t.Fatalf("lock path stat = %v, %v; want an existing directory", info, err)
	}
	data, err := os.ReadFile(filepath.Join(path, "pid"))
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != strconv.Itoa(os.Getpid()) {
		t.Fatalf("pid file = %q, want the holder pid %d", got, os.Getpid())
	}
	// A second acquire with maxWait=0 fails immediately.
	start := time.Now()
	if _, err := AcquireDirLock(context.Background(), path, 0); err == nil {
		t.Fatal("second AcquireDirLock succeeded while the lock is held")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("fail-fast acquire took %v, want immediate", elapsed)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

func TestDirLockWaiterAcquiresPromptlyAfterRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = holder.Release()
	}()
	start := time.Now()
	waiter, err := AcquireDirLock(context.Background(), path, 30*time.Second)
	if err != nil {
		t.Fatalf("waiter AcquireDirLock: %v", err)
	}
	// The wakeup is an inotify event on the holder's removal, not a poll
	// interval: acquisition must land well inside any coarse bound.
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("waiter took %v to acquire after a 50ms hold; the wait is not event-based", elapsed)
	}
	if err := waiter.Release(); err != nil {
		t.Fatalf("waiter Release: %v", err)
	}
}

func TestDirLockReleaseRacingWatchSetupIsObserved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	// Deterministically place the holder's release inside the race
	// window of review finding P2-2: the waiter's mkdir has already
	// failed (the lock is held), and the release lands after that
	// failure but before the waiter's inotify watch is armed, so no
	// inotify event can ever report it. The waiter must still observe
	// the release promptly instead of sleeping out its whole maxWait.
	orig := inotifyAddWatchFn
	released := false
	inotifyAddWatchFn = func(fd int, pathname string, mask uint32) (int, error) {
		if !released {
			released = true
			_ = holder.Release()
		}
		return orig(fd, pathname, mask)
	}
	defer func() { inotifyAddWatchFn = orig }()
	start := time.Now()
	waiter, err := AcquireDirLock(context.Background(), path, 3*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("waiter AcquireDirLock: %v", err)
	}
	if !released {
		t.Fatal("the watch-setup seam never ran; the race window was not exercised")
	}
	if elapsed > 2*time.Second {
		t.Fatalf("waiter took %v against a 3s maxWait; the release that raced watch setup was missed and the waiter slept out its wait", elapsed)
	}
	if err := waiter.Release(); err != nil {
		t.Fatalf("waiter Release: %v", err)
	}
}

func TestDirLockStaleLockIsNotStolen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	// A lock left by a dead holder: directory + pid file, no live
	// holder goroutine will ever Release it. The waiter's only correct
	// outcome is a timeout carrying the operator guidance — holder
	// death is not an acquisition event (the stale-lock contract).
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "pid"), []byte("999999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := AcquireDirLock(context.Background(), path, 300*time.Millisecond)
	if err == nil {
		t.Fatal("AcquireDirLock stole a stale lock; want a timeout error")
	}
	if !strings.Contains(err.Error(), "inspect and remove it only if the holder is not running") {
		t.Fatalf("timeout error = %q, want the operator guidance text", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatal("the stale lock directory was removed by the waiter")
	}
}

func TestDirLockWaitHonorsContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	defer func() { _ = holder.Release() }()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = AcquireDirLock(ctx, path, 30*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AcquireDirLock error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("context-bound wait took %v, want prompt cancellation", elapsed)
	}
}

func TestDirLockReleaseRemovesAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	l, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lock path after Release: stat error = %v, want not-exist", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("second Release = %v, want nil (idempotent)", err)
	}
	// The path is free for the next holder.
	next, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock after Release: %v", err)
	}
	if err := next.Release(); err != nil {
		t.Fatalf("next Release: %v", err)
	}
}
