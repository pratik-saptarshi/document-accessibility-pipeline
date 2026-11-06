//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// A Session owns one started process-group leader from Start to Reap.
// It makes the observe → signal → reap ordering unrepresentable-when-
// wrong: there is no method that reaps before the group has been
// signalled, because Reap performs the signal itself when the caller
// has not. The leader stays unreaped — a zombie holding its PID, and
// therefore its process-group ID — from Start until Reap, so a group
// signal can never land on a recycled PID.
//
// Start must return before any other method is called. Terminate, Kill,
// and Reap are serialized against each other and are idempotent;
// WaitExit is pure observation and may run concurrently with them.
type Session struct {
	cmd *exec.Cmd

	mu        sync.Mutex // serializes Start, Terminate, Kill, Reap; WaitExit takes it briefly to register its pidfd copy
	pid       int        // leader PID; set by Start, never changes after
	pidfd     int        // leader pidfd opened by Start, or -1 (fallback)
	started   bool
	signalled bool // a SIGTERM or SIGKILL has landed on the group
	reaped    bool
	reapErr   error // stored result of the single cmd.Wait, from Reap
}

// groupSignalFn sends a signal to a process group: syscall.Kill with a
// negative pid. It is held in a variable so tests can record signal
// attempts — in particular, that no signal is attempted after Reap,
// when the leader PID may have been recycled by an unrelated process
// (the same seam pattern as pidfdOpenFn and inotifyAddWatchFn).
var groupSignalFn = syscall.Kill

// NewSession wraps a caller-built command. The caller keeps its own
// command construction (arguments, environment, output wiring); the
// Session owns only the process-group lifecycle.
func NewSession(cmd *exec.Cmd) *Session {
	return &Session{cmd: cmd, pidfd: -1}
}

// Start sets SysProcAttr{Setpgid: true} on the command and starts it. A
// caller that preset Setpgid with Pgid 0 keeps its SysProcAttr
// untouched — the leader then owns the new group whose ID is its own
// PID, which is the group every Session signal targets. A caller that
// preset a SysProcAttr without Setpgid, or with a nonzero Pgid (which
// would join an existing foreign group while signals target the
// leader's own PID), is rejected, because the Session's guarantees
// require the leader to own its process group.
func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return errors.New("processgroup: session already started")
	}
	if s.cmd.Process != nil {
		return errors.New("processgroup: command already started")
	}
	if s.cmd.SysProcAttr != nil && !s.cmd.SysProcAttr.Setpgid {
		return errors.New("processgroup: command SysProcAttr does not set Setpgid")
	}
	if s.cmd.SysProcAttr != nil && s.cmd.SysProcAttr.Pgid != 0 {
		return errors.New("processgroup: command SysProcAttr presets a process-group ID; the session leader must own its group (Pgid must be 0)")
	}
	if s.cmd.SysProcAttr == nil {
		s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	if err := s.cmd.Start(); err != nil {
		return err
	}
	s.pid = s.cmd.Process.Pid
	s.started = true
	// A pidfd failure is not fatal: WaitExit falls back to the WNOWAIT
	// probe (see waitLeaderExit).
	if fd, err := pidfdOpenFn(s.pid, 0); err == nil {
		s.pidfd = fd
	}
	return nil
}

// PID returns the leader's PID, or 0 before Start.
func (s *Session) PID() int {
	return s.pid
}

// waitExitMuPollInterval is the cadence at which WaitExit retries
// the session mutex while a lifecycle method holds it (see lockMuCtx).
const waitExitMuPollInterval = time.Millisecond

// lockMuCtx acquires s.mu, returning false if ctx ends first. The
// lifecycle methods hold s.mu for as long as they take — Terminate
// for its whole grace, Reap for its whole cmd.Wait — so a plain Lock
// would make an observer's context powerless for the holder's entire
// duration: a WaitExit whose deadline passes while Terminate sits in
// its grace would still be parked on the mutex when the holder
// finally releases it. Polling TryLock keeps the context
// authoritative while waiting, at a bounded cadence, without
// spawning a goroutine. The poll changes only how the lock is
// waited for: every decision WaitExit makes is still taken with the
// lock held, atomically against the lifecycle methods, exactly as
// with a plain Lock — no registration state is read or descriptor
// bound outside the critical section.
func (s *Session) lockMuCtx(ctx context.Context) bool {
	if s.mu.TryLock() {
		return true
	}
	timer := time.NewTimer(waitExitMuPollInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return false
		case <-timer.C:
			if s.mu.TryLock() {
				return true
			}
			timer.Reset(waitExitMuPollInterval)
		}
	}
}

// WaitExit blocks until the leader exits (an event wait) or ctx is
// done. It never reaps and never signals: the leader stays a zombie
// holding its PID/PGID until Reap. It is safe to call more than once,
// and safe concurrently with the serialized methods: registration
// under the session mutex binds the descriptor this call polls —
// a duplicate of the session pidfd, or a transient pidfd opened
// while the lock still holds ownership of the PID — so a concurrent
// Reap cannot pull the descriptor from under the poll, release the
// PID before it is bound, or hand this wait a descriptor number
// Reap released for reuse. If no pidfd can be bound (the fallback),
// the WNOWAIT probe watches the bare PID instead. Once Reap has
// run, WaitExit instead returns nil from recorded state: Reap's
// stored result proves the leader exited, and its released PID —
// possibly recycled by then — is never reopened or probed. An
// observer that began concurrently with Reap but registers only
// after it therefore still observes the completed exit, and so
// does an observer whose probe was already in flight when Reap
// completed: the probe cannot see an already-reaped leader and
// reports ECHILD, which WaitExit resolves to the recorded exit.
// A wait that fails while the session is unreaped — a context
// cancellation, say — returns its own error unchanged. Both
// mutex acquisitions — registration, and the state re-check after
// a failed wait — wait cancellation-awarely: the lifecycle
// methods hold the mutex for their whole duration (Terminate
// across its grace, Reap across cmd.Wait), so a context that
// ends while WaitExit is queued behind one returns the context
// error at its own deadline instead of the holder's.
func (s *Session) WaitExit(ctx context.Context) error {
	if !s.lockMuCtx(ctx) {
		return ctx.Err()
	}
	if !s.started || s.pid <= 0 {
		s.mu.Unlock()
		return ErrInvalidPID
	}
	if s.reaped {
		// Registration after Reap: the exit is a recorded fact, and
		// the leader PID has been released for reuse. Falling into
		// the wait with the stored PID would reopen (or probe) a
		// PID this session no longer owns — normally surfacing as
		// an ECHILD from the probe instead of the completed exit,
		// or, on a recycled PID, waiting on an unrelated process.
		// The gate sits at registration, before both descriptor
		// bindings below and the probe, so no ordering can reopen
		// the PID.
		s.mu.Unlock()
		return nil
	}
	pid := s.pid
	fd := -1
	if s.pidfd >= 0 {
		if dup, err := unix.FcntlInt(uintptr(s.pidfd), unix.F_DUPFD_CLOEXEC, 0); err == nil {
			fd = int(dup)
		}
	}
	if fd < 0 {
		// No descriptor yet: the session is on the fallback
		// (Start could not open a pidfd) or the duplication
		// failed. Bind a transient pidfd now, inside the
		// registration lock, while the session still owns the
		// PID — a Reap landing after registration then cannot
		// leave this wait opening a released PID. If the open
		// fails (pidfd unavailable), fd stays -1 and the wait
		// below is the WNOWAIT probe, entered directly: the
		// open must not be re-attempted after the lock is
		// dropped, or the released-PID window reopens.
		if opened, err := pidfdOpenFn(pid, 0); err == nil {
			fd = opened
		}
	}
	s.mu.Unlock()
	var err error
	if fd >= 0 {
		defer func() { _ = unix.Close(fd) }()
		err = waitLeaderExit(ctx, pid, fd)
	} else {
		err = waitLeaderProbe(ctx, pid)
	}
	if err == nil {
		return nil
	}
	// The wait may have raced a completing Reap. The probe cannot
	// observe an already-reaped leader and reports ECHILD, and any
	// other wait error that coincides with a completed reap is an
	// artifact of watching a PID whose ownership ended mid-wait.
	// If Reap ran, the leader's exit is a recorded fact — report
	// it instead of the artifact. A wait that fails while the
	// session is unreaped returns its own error unchanged. The
	// re-check waits for the mutex cancellation-awarely, like
	// registration: Reap holds the mutex for its whole cmd.Wait,
	// so if the context ends while the check is still queued
	// behind a holder, the context error — usually the very error
	// the wait just returned — is the answer, at the deadline
	// rather than at the holder's convenience.
	if !s.lockMuCtx(ctx) {
		return ctx.Err()
	}
	reaped := s.reaped
	s.mu.Unlock()
	if reaped {
		return nil
	}
	return err
}

// Terminate sends SIGTERM to the group, waits event-based up to grace
// for the leader's exit, then sends SIGKILL. A group that died of the
// TERM is Terminate's goal state, not an error; a group that had already
// vanished before any signal landed reports os.ErrProcessDone.
//
// After Reap, Terminate issues no signal at all: the reap released the
// leader PID, which the kernel may have recycled for an unrelated
// process group. It reports the recorded outcome instead — nil if a
// signal had landed before the reap (the goal state was reached),
// os.ErrProcessDone if the group vanished unsignalled.
func (s *Session) Terminate(grace time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started || s.pid <= 0 {
		return ErrInvalidPID
	}
	if s.reaped {
		return s.afterReapSignalResult()
	}
	if err := groupSignalFn(-s.pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			if s.signalled {
				return nil
			}
			return os.ErrProcessDone
		}
		return err
	}
	s.signalled = true
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	_ = waitLeaderExit(ctx, s.pid, s.pidfd)
	cancel()
	// Group members may outlive the leader; the KILL goes out whether
	// the leader exited early or the grace expired. A vanished group
	// here died of the TERM above: success, per the rule above.
	if err := groupSignalFn(-s.pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	return nil
}

// Kill sends SIGKILL to the group. A group that had already vanished
// before any signal landed reports os.ErrProcessDone; once a signal has
// landed, a vanished group is the goal state and reports nil. After
// Reap, Kill issues no signal (the leader PID may have been recycled)
// and reports the recorded outcome, exactly as Terminate does.
func (s *Session) Kill() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reaped {
		return s.afterReapSignalResult()
	}
	return s.killLocked()
}

// afterReapSignalResult is the answer Terminate and Kill give once the
// session has been reaped. No signal may be attempted then — Reap's
// cmd.Wait released the leader PID for reuse — so the result replays
// Decision 3's vanished-group rule from the recorded state instead of
// probing the group ID with a real signal: a session that had signalled
// reached the goal state (nil); a session whose group vanished before
// any signal landed reports os.ErrProcessDone.
func (s *Session) afterReapSignalResult() error {
	if s.signalled {
		return nil
	}
	return os.ErrProcessDone
}

func (s *Session) killLocked() error {
	if !s.started || s.pid <= 0 {
		return ErrInvalidPID
	}
	if err := groupSignalFn(-s.pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			if s.signalled {
				return nil
			}
			return os.ErrProcessDone
		}
		return err
	}
	s.signalled = true
	return nil
}

// Reap is the only reap path. If the group has not been signalled yet,
// Reap kills it first (a vanished group is tolerated) — no method
// sequence can reap an unsignalled group into a recycled-PID hazard.
// Reap runs the command's Wait exactly once and returns its stored
// result on every subsequent call.
func (s *Session) Reap() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started || s.pid <= 0 {
		return ErrInvalidPID
	}
	if s.reaped {
		return s.reapErr
	}
	s.reaped = true
	var killErr error
	if !s.signalled {
		if err := s.killLocked(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			killErr = err
		}
	}
	waitErr := s.cmd.Wait()
	if s.pidfd >= 0 {
		_ = unix.Close(s.pidfd)
		s.pidfd = -1
	}
	if killErr != nil {
		s.reapErr = errors.Join(killErr, waitErr)
	} else {
		s.reapErr = waitErr
	}
	return s.reapErr
}
