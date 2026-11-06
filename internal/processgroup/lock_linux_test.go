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

	"golang.org/x/sys/unix"
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

// forceLockFallback makes every inotify watch setup fail, so
// waitForLockChange degrades to pollLockWait — the fallback whose
// chunking Codex finding F4 is about.
func forceLockFallback(t *testing.T) {
	t.Helper()
	orig := inotifyAddWatchFn
	inotifyAddWatchFn = func(int, string, uint32) (int, error) {
		return 0, errors.New("inotify watch forced unavailable (test seam)")
	}
	t.Cleanup(func() { inotifyAddWatchFn = orig })
}

// One fallback wait must be one bounded chunk: pollLockWait returns
// after ~25ms even when seconds of wait remain, so AcquireDirLock's
// loop — not the fallback — owns the retry cadence. The pre-fix
// fallback looped internally until the whole remainder elapsed.
func TestPollLockWaitReturnsAfterOneChunk(t *testing.T) {
	start := time.Now()
	if err := pollLockWait(context.Background(), 3*time.Second); err != nil {
		t.Fatalf("pollLockWait: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("pollLockWait(3s) returned after %v, want one bounded chunk (the acquire loop must retry the mkdir between chunks)", elapsed)
	}
}

// The fallback twin of TestDirLockWaiterAcquiresPromptlyAfterRelease:
// with inotify unavailable, a release at 30ms must still wake the
// waiter at chunk scale — the pre-fix fallback slept out the whole
// maxWait and then reported the lock as still held.
func TestDirLockFallbackWaiterObservesReleasePromptly(t *testing.T) {
	forceLockFallback(t)
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = holder.Release()
	}()
	start := time.Now()
	waiter, err := AcquireDirLock(context.Background(), path, 3*time.Second)
	if err != nil {
		t.Fatalf("waiter AcquireDirLock: %v (the release at 30ms was not observed between fallback chunks)", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("waiter took %v to acquire after a 30ms hold on the fallback path; want chunk scale, not maxWait scale", elapsed)
	}
	if err := waiter.Release(); err != nil {
		t.Fatalf("waiter Release: %v", err)
	}
}

// Codex finding F4's reproduced numbers: holder releases at 30ms, the
// waiter's context ends at 150ms. The waiter must win the lock in
// between — the pre-fix fallback never retried the mkdir inside that
// window, so the cancellation failed an acquisition whose lock had
// long been free.
func TestDirLockFallbackReleaseBeforeCancelIsNotStranded(t *testing.T) {
	forceLockFallback(t)
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		_ = holder.Release()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	waiter, err := AcquireDirLock(ctx, path, 30*time.Second)
	if err != nil {
		t.Fatalf("waiter AcquireDirLock = %v, want success: the lock freed at 30ms, before the 150ms context deadline", err)
	}
	if err := waiter.Release(); err != nil {
		t.Fatalf("waiter Release: %v", err)
	}
}

// Codex finding F5: when InotifyInit succeeds but adding the watch
// fails, the inotify descriptor must still be closed. The pre-fix
// code registered its deferred close only after the watch attempt
// succeeded, so every contended wait whose watch setup failed leaked
// one descriptor and one inotify instance.
func TestDirLockWatchSetupFailureDoesNotLeakDescriptors(t *testing.T) {
	orig := inotifyAddWatchFn
	inotifyAddWatchFn = func(int, string, uint32) (int, error) {
		return 0, errors.New("inotify watch forced to fail (test seam)")
	}
	t.Cleanup(func() { inotifyAddWatchFn = orig })
	parent := t.TempDir()
	countFDs := func() int {
		entries, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatalf("read /proc/self/fd: %v", err)
		}
		return len(entries)
	}
	before := countFDs()
	for i := 0; i < 25; i++ {
		if err := waitForLockChange(context.Background(), parent, "lock", 5*time.Millisecond); err != nil {
			t.Fatalf("waitForLockChange: %v", err)
		}
	}
	if grew := countFDs() - before; grew > 2 {
		t.Fatalf("descriptors grew by %d across 25 failed watch setups, want no leak (slack 2)", grew)
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

// Codex fix-round-6 finding G2: waitForLockChange created its
// inotify descriptor with unix.InotifyInit, which does not set
// FD_CLOEXEC. A subprocess started while a contended acquisition
// waits — by any spawn path that does not itself scrub descriptors —
// inherits the descriptor, and the inotify instance and its watch
// then outlive the parent's close until every inheriting child has
// exited. The descriptor must be created close-on-exec. The
// watch-arming seam receives the very descriptor production created,
// so the test reads its file-status flags with F_GETFD in the middle
// of a genuine contended wait.
func TestDirLockInotifyDescriptorIsCloseOnExec(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	holder, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("holder AcquireDirLock: %v", err)
	}
	orig := inotifyAddWatchFn
	flagsCh := make(chan int, 1)
	inotifyAddWatchFn = func(fd int, pathname string, mask uint32) (int, error) {
		f, ferr := unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
		if ferr != nil {
			t.Errorf("F_GETFD on the inotify descriptor: %v", ferr)
			f = -2
		}
		select {
		case flagsCh <- f:
		default:
		}
		return orig(fd, pathname, mask)
	}
	t.Cleanup(func() { inotifyAddWatchFn = orig })
	waiterErr := make(chan error, 1)
	var waiter *DirLock
	go func() {
		w, aerr := AcquireDirLock(context.Background(), path, 30*time.Second)
		waiter = w
		waiterErr <- aerr
	}()
	var flags int
	select {
	case flags = <-flagsCh:
	case <-time.After(5 * time.Second):
		t.Fatal("the watch-arming seam never ran; the waiter never reached its inotify wait")
	}
	if err := holder.Release(); err != nil {
		t.Fatalf("holder Release: %v", err)
	}
	if err := <-waiterErr; err != nil {
		t.Fatalf("waiter AcquireDirLock: %v", err)
	}
	if err := waiter.Release(); err != nil {
		t.Fatalf("waiter Release: %v", err)
	}
	if flags&unix.FD_CLOEXEC == 0 {
		t.Fatalf("inotify descriptor flags = %#x, want FD_CLOEXEC set (a subprocess spawned during the contended wait would inherit the descriptor and pin the inotify instance past the parent's close)", flags)
	}
}

// Codex round-5 finding H1: Release set its released flag before
// attempting the directory removal, so a failed removal (here: an
// unexpected file inside the lock directory makes the directory
// non-empty) was permanent — the error returned, but every later
// Release was a nil no-op that never retried the cleanup, leaving
// the lock held and subsequent acquisitions timing out. The released
// state may be marked only once the removal succeeds or confirms
// the directory absent; a failed Release leaves the lock held and
// retryable.
func TestDirLockReleaseRetriesRemovalAfterFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	l, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock: %v", err)
	}
	obstruction := filepath.Join(path, "unexpected")
	if err := os.WriteFile(obstruction, []byte("leftover\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err == nil {
		t.Fatal("Release with an obstructed lock directory = nil, want the removal error")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("lock directory after failed Release: stat error = %v, want the directory still present (lock still held)", err)
	}
	if err := os.Remove(obstruction); err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("retry Release after clearing the obstruction: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lock directory after retry Release: stat error = %v, want not-exist (the retry must actually remove it)", err)
	}
	// The user-visible harm, closed: the next acquisition succeeds
	// promptly instead of waiting out its timeout on a lock whose
	// holder believes it released.
	start := time.Now()
	next, err := AcquireDirLock(context.Background(), path, 3*time.Second)
	if err != nil {
		t.Fatalf("AcquireDirLock after retried Release: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("acquisition after retried Release took %v, want prompt", elapsed)
	}
	if err := next.Release(); err != nil {
		t.Fatalf("next Release: %v", err)
	}
}

// The released state is also marked when Release finds the removal
// target already absent (e.g. an operator removed the lock per the
// stale-lock contract): Release returns nil, and — because the lock
// is released — a later Release must not remove a new directory a
// different holder has since created at the same path.
func TestDirLockReleaseWhenDirectoryAlreadyAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	l, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock: %v", err)
	}
	if err := os.Remove(filepath.Join(path, "pid")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("Release with the lock directory already absent = %v, want nil", err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "pid")
	if err := os.WriteFile(marker, []byte("999999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("second Release = %v, want nil (idempotent)", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("the new holder's directory was removed by a stale Release: stat error = %v", err)
	}
}

// The acquire rollback must undo the initialization in reverse
// order. The pid write can fail after it has created the pid file
// (a filesystem filling mid-write, a zero-byte RLIMIT_FSIZE): the
// rollback originally removed only the lock directory, which is
// then non-empty, so that removal silently failed — AcquireDirLock
// returned no handle, yet the lock directory and pid file stayed
// behind, and every later acquisition waited out its timeout on a
// lock nobody held, even after the write fault cleared. Removing
// the potentially created pid file first makes the directory
// removal able to succeed, so a failed acquisition leaves no lock
// behind at all.
func TestDirLockAcquireRollsBackWhenPIDWriteFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock")
	errPIDWrite := errors.New("injected pid write failure")
	orig := writeFileFn
	writeFileFn = func(name string, data []byte, perm os.FileMode) error {
		// The write partially completes: the pid file exists
		// with truncated content, then the write fails.
		if err := os.WriteFile(name, data[:1], perm); err != nil {
			t.Fatalf("fault injection: create pid file: %v", err)
		}
		return errPIDWrite
	}
	t.Cleanup(func() { writeFileFn = orig })

	l, err := AcquireDirLock(context.Background(), path, 0)
	if err == nil {
		t.Fatal("AcquireDirLock with a failing pid write = nil error, want the write error")
	}
	if !errors.Is(err, errPIDWrite) {
		t.Fatalf("AcquireDirLock error = %v, want it to wrap the pid write error", err)
	}
	if l != nil {
		t.Fatalf("AcquireDirLock returned handle %v with the error, want nil", l)
	}
	if _, serr := os.Stat(filepath.Join(path, "pid")); !errors.Is(serr, os.ErrNotExist) {
		t.Fatalf("pid file after failed acquire: stat error = %v, want not-exist (the rollback must remove the file it may have created)", serr)
	}
	if _, serr := os.Stat(path); !errors.Is(serr, os.ErrNotExist) {
		t.Fatalf("lock directory after failed acquire: stat error = %v, want not-exist (a leftover directory leaves the lock held with no handle)", serr)
	}

	// The user-visible harm, closed: once the write fault clears,
	// the next acquisition succeeds immediately instead of waiting
	// out its timeout on the stranded lock.
	writeFileFn = orig
	next, err := AcquireDirLock(context.Background(), path, 0)
	if err != nil {
		t.Fatalf("AcquireDirLock after the write fault cleared: %v", err)
	}
	if err := next.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}
