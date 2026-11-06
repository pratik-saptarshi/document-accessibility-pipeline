//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// pidfdOpenFn is the pidfd_open seam. Production uses unix.PidfdOpen;
// tests substitute a failing stub to force the WNOWAIT fallback in
// waitLeaderExit and prove it behaves identically.
var pidfdOpenFn = unix.PidfdOpen

// leaderExitedNoReap reports whether pid has exited, without reaping it.
// The unreaped zombie keeps its PID allocated, so a later process-group
// signal cannot hit a recycled PID.
func leaderExitedNoReap(pid int) (bool, error) {
	var info unix.Siginfo
	for {
		err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOHANG|unix.WNOWAIT, nil)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return false, err
		}
		return info.Signo != 0, nil
	}
}

// waitLeaderExit blocks until pid has exited — without reaping it — or
// ctx is done. fd is an already-open pidfd for pid owned by the caller,
// or -1 to open one here (closed before returning). When no pidfd can be
// opened (pre-5.3 kernel, restrictive seccomp) it falls back to
// waitLeaderProbe; the fallback lives only in this file.
func waitLeaderExit(ctx context.Context, pid int, fd int) error {
	if fd < 0 {
		opened, err := pidfdOpenFn(pid, 0)
		if err == nil {
			defer unix.Close(opened)
			fd = opened
		}
	}
	if fd >= 0 {
		return waitPidfd(ctx, pid, fd)
	}
	return waitLeaderProbe(ctx, pid)
}

// waitPidfd blocks in the kernel until the pidfd reports the leader's
// exit or ctx is done. poll(2) takes no context, so an eventfd stands
// in for ctx inside the poll set: a watcher goroutine writes to the
// eventfd when ctx fires, and the poll itself runs in the caller's
// goroutine — still zero intermediate wakeups. The watcher is joined
// before the eventfd is closed and exits as soon as the wait returns,
// so a cancelled wait leaves nothing behind: however often a caller
// observes with a deadline (a Session's WaitExit, or the ownerless
// WaitForLeaderExitNoReap), no goroutine accumulates. (The previous
// shape ran the poll itself in a goroutine that cancellation could
// not stop; it stayed parked until the leader exited.)
func waitPidfd(ctx context.Context, pid int, fd int) error {
	efd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		// Without an eventfd the kernel wait cannot be made
		// cancellation-aware; degrade to the probe, which selects
		// on ctx directly.
		return waitLeaderProbe(ctx, pid)
	}
	done := make(chan struct{})
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			// Wake the poll. The write can only fail if the
			// eventfd filled or was closed, and the main wait
			// is already returning in both cases.
			_, _ = unix.Write(efd, []byte{1, 0, 0, 0, 0, 0, 0, 0})
		case <-done:
		}
	}()
	defer func() {
		close(done)
		<-watcherDone
		_ = unix.Close(efd)
	}()
	fds := []unix.PollFd{
		{Fd: int32(fd), Events: unix.POLLIN},
		{Fd: int32(efd), Events: unix.POLLIN},
	}
	for {
		_, err := unix.Poll(fds, -1)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if fds[0].Revents != 0 {
			// Any pidfd event (readable, error, hangup) means
			// the leader's state changed terminally: it exited.
			return nil
		}
		if fds[1].Revents != 0 {
			return ctx.Err()
		}
	}
}

// waitLeaderProbe is the fallback leader-exit wait for environments
// without pidfd: the WNOWAIT probe on a ticker. It is the one ticker in
// the package, retained so the fallback behaves exactly like the
// pre-pidfd implementation.
func waitLeaderProbe(ctx context.Context, pid int) error {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		exited, err := leaderExitedNoReap(pid)
		if err != nil {
			return err
		}
		if exited {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// WaitForLeaderExitNoReap waits until pid has exited (or ctx is done)
// without reaping it: observe the exit, let the caller signal the group
// while the leader still owns its PID, then reap.
func WaitForLeaderExitNoReap(ctx context.Context, pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}
	return waitLeaderExit(ctx, pid, -1)
}

// KillGroup sends SIGKILL to the process group led by pid. A group that
// no longer exists is reported as os.ErrProcessDone.
func KillGroup(pid int) error {
	if pid <= 0 {
		return ErrInvalidPID
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}

// TerminateGroup sends SIGTERM to the process group led by pid, waits
// event-based up to grace for the leader's exit, then sends SIGKILL. A
// non-positive pid, or a group that no longer exists, is a no-op.
func TerminateGroup(pid int, grace time.Duration) error {
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
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	_ = waitLeaderExit(ctx, pid, -1)
	cancel()
	err := syscall.Kill(groupID, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
