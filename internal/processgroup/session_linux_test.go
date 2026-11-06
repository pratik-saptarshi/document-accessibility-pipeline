//go:build linux

package processgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// sessionGrandchildHelper starts a leader whose shell spawns a sleeping
// grandchild in its own process group and records the grandchild's PID in
// pidFile. leaderTail is appended to the shell script (e.g. "exit 0" for a
// leader that exits at once, "wait" for one that stays alive).
func startSessionWithGrandchild(t *testing.T, leaderTail string) (*Session, int) {
	t.Helper()
	pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
	script := `sleep 30 & echo $! > "$1"; ` + leaderTail
	cmd := exec.Command("/bin/sh", "-c", script, "sh", pidFile)
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Kill(-s.PID(), syscall.SIGKILL) })
	return s, readRecordedPID(t, pidFile)
}

func readRecordedPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, convErr := strconv.Atoi(strings.TrimSpace(string(data)))
			if convErr == nil && pid > 0 {
				return pid
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("grandchild PID file %s never became readable", path)
	return 0
}

// processStopped reports whether pid is gone or a zombie, following the
// probe pattern used by financialfixtures' process tests.
func processStopped(pid int) bool {
	err := syscall.Kill(pid, 0)
	if errors.Is(err, syscall.ESRCH) {
		return true
	}
	if err != nil {
		return false
	}
	data, readErr := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if readErr != nil {
		return false
	}
	end := strings.LastIndexByte(string(data), ')')
	if end < 0 {
		return false
	}
	fields := strings.Fields(string(data[end+1:]))
	return len(fields) > 0 && (fields[0] == "Z" || fields[0] == "X")
}

func waitProcessStopped(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if processStopped(pid) {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return processStopped(pid)
}

// assertProcessAliveFor fails the test if pid is observed stopped at
// any sample across the window. A single aliveness sample taken
// immediately after an observe operation is not discriminating: a
// wrongly-signalled process takes a nonzero time to observably die,
// so the death can land after a one-shot check has already passed
// (review finding P2-1). Sampling continuously across a window makes
// the check catch any death the operation caused, while a genuinely
// live process stays alive for the whole window.
func assertProcessAliveFor(t *testing.T, pid int, window time.Duration) {
	t.Helper()
	deadline := time.Now().Add(window)
	for {
		if processStopped(pid) {
			t.Fatalf("process %d stopped inside the %s aliveness window; want it alive throughout", pid, window)
		}
		if !time.Now().Before(deadline) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// sessionOrderingScenario is plan scenario (i): observe the leader's exit
// without reaping, prove nothing was signalled, then let Reap perform the
// signal-then-reap sequence.
func sessionOrderingScenario(t *testing.T) {
	t.Helper()
	s, gpid := startSessionWithGrandchild(t, "exit 0")
	if s.PID() <= 0 {
		t.Fatalf("PID = %d after Start, want positive", s.PID())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.WaitExit(ctx); err != nil {
		t.Fatalf("WaitExit: %v", err)
	}
	// WaitExit is repeatable.
	if err := s.WaitExit(ctx); err != nil {
		t.Fatalf("second WaitExit: %v", err)
	}
	exited, err := leaderExitedNoReap(s.PID())
	if err != nil || !exited {
		t.Fatalf("leaderExitedNoReap = %v, %v; want true, nil for the unreaped leader", exited, err)
	}
	// The zombie still owns its PID, so the group ID cannot be recycled.
	if err := syscall.Kill(s.PID(), 0); err != nil {
		t.Fatalf("unreaped leader lost its PID: %v", err)
	}
	// Observing must not have signalled the group. The aliveness check
	// is windowed, not one-shot: a one-shot sample microseconds after
	// WaitExit can precede the observable death of a wrongly-signalled
	// grandchild and let a signal-in-WaitExit regression pass
	// (review finding P2-1).
	assertProcessAliveFor(t, gpid, 250*time.Millisecond)
	// Reap signals first, then reaps: the leader exited 0, so the wait
	// result is nil, and the grandchild dies from Reap's signal.
	if err := s.Reap(); err != nil {
		t.Fatalf("Reap: %v", err)
	}
	if !waitProcessStopped(gpid, 5*time.Second) {
		t.Fatal("grandchild survived Reap")
	}
}

func TestSessionOrderingObserveThenReapSignalsGroup(t *testing.T) {
	sessionOrderingScenario(t)
}

func TestSessionReapUnsignalledKillsGroupFirst(t *testing.T) {
	s, gpid := startSessionWithGrandchild(t, "wait")
	if processStopped(s.PID()) || processStopped(gpid) {
		t.Fatal("leader or grandchild not running after Start")
	}
	// No WaitExit, no signal: Reap must kill the group before reaping.
	err1 := s.Reap()
	if err1 == nil {
		t.Fatal("Reap of a SIGKILLed leader returned nil, want the wait error")
	}
	if !waitProcessStopped(gpid, 5*time.Second) {
		t.Fatal("grandchild survived Reap of an unsignalled group")
	}
	// Reap is idempotent and returns the stored result.
	if err2 := s.Reap(); err2 != err1 {
		t.Fatalf("second Reap = %v, want the stored result %v", err2, err1)
	}
}

func TestSessionEntryPointsRejectUnstartedSession(t *testing.T) {
	s := NewSession(exec.Command("sleep", "30"))
	if got := s.PID(); got != 0 {
		t.Fatalf("PID before Start = %d, want 0", got)
	}
	if err := s.WaitExit(context.Background()); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("WaitExit error = %v, want ErrInvalidPID", err)
	}
	if err := s.Terminate(time.Second); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("Terminate error = %v, want ErrInvalidPID", err)
	}
	if err := s.Kill(); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("Kill error = %v, want ErrInvalidPID", err)
	}
	if err := s.Reap(); !errors.Is(err, ErrInvalidPID) {
		t.Fatalf("Reap error = %v, want ErrInvalidPID", err)
	}
}

func TestSessionStartRejectsConflictingSysProcAttr(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := NewSession(cmd).Start(); err == nil {
		t.Fatal("Start accepted a SysProcAttr without Setpgid")
	}
	preset := exec.Command("sleep", "30")
	preset.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s := NewSession(preset)
	if err := s.Start(); err != nil {
		t.Fatalf("Start with caller-preset Setpgid: %v", err)
	}
	if err := s.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a SIGKILLed leader returned nil, want the wait error")
	}
	if err := s.Start(); err == nil {
		t.Fatal("second Start accepted")
	}
	// A preset Setpgid is accepted only with Pgid 0 — the leader then
	// owns the new group whose ID is its own PID, which is the group
	// every Session signal targets. A nonzero preset Pgid would join
	// an existing foreign group instead: the child would live in that
	// group while signals go to -leaderPID, so Kill would report a
	// vanished group while the child remains alive (Codex finding
	// F2). Preset this test process's own group as the foreign one;
	// if Start wrongly accepts it, kill the joined child directly by
	// PID (never by group) before failing.
	foreign := exec.Command("sleep", "30")
	foreign.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: syscall.Getpgrp()}
	if err := NewSession(foreign).Start(); err == nil {
		_ = foreign.Process.Kill()
		_ = foreign.Wait()
		t.Fatal("Start accepted a preset nonzero Pgid (foreign process group)")
	}
	if foreign.Process != nil {
		t.Fatal("Start with a preset nonzero Pgid still started the command")
	}
}

func TestSessionSignalsVanishedGroupReportProcessDone(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Vanish the group behind the Session's back: kill the only member
	// and reap it directly, so no signal can land anywhere.
	if err := syscall.Kill(s.PID(), syscall.SIGKILL); err != nil {
		t.Fatalf("kill leader: %v", err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("direct Wait of a SIGKILLed leader returned nil")
	}
	if err := s.Terminate(time.Second); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Terminate on a vanished group = %v, want os.ErrProcessDone", err)
	}
	if err := s.Kill(); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Kill on a vanished group = %v, want os.ErrProcessDone", err)
	}
}

// recordedGroupSignal is one groupSignalFn call observed by the spy
// installed by installGroupSignalSpy.
type recordedGroupSignal struct {
	pid int
	sig syscall.Signal
}

// installGroupSignalSpy replaces groupSignalFn with a recorder that
// delegates to the real kill(2), so behavior is unchanged while every
// signal attempt becomes observable. Tests are serial in this package,
// and Sessions signal from the caller's goroutine, so a plain slice is
// a safe record.
func installGroupSignalSpy(t *testing.T) *[]recordedGroupSignal {
	t.Helper()
	var calls []recordedGroupSignal
	orig := groupSignalFn
	groupSignalFn = func(pid int, sig syscall.Signal) error {
		calls = append(calls, recordedGroupSignal{pid: pid, sig: sig})
		return orig(pid, sig)
	}
	t.Cleanup(func() { groupSignalFn = orig })
	return &calls
}

// Codex finding F1: after Reap the leader PID is free for reuse, so a
// repeated Terminate/Kill must not signal -pid again — the numeric
// group ID may by then belong to an unrelated process group. These
// tests assert the signal *attempts* via the spy, because the return
// values alone do not discriminate: the pre-fix code also returns nil
// or os.ErrProcessDone here, via an ESRCH from the very signal attempt
// that must not happen.
func TestSessionSignalsAfterReapStaySilent(t *testing.T) {
	calls := installGroupSignalSpy(t)
	s, gpid := startSessionWithGrandchild(t, "wait")
	// Reap performs the kill itself: exactly one group signal (the
	// SIGKILL from Reap's signal step) is expected by now.
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a live group returned nil, want the wait error")
	}
	if !waitProcessStopped(gpid, 5*time.Second) {
		t.Fatal("grandchild survived Reap")
	}
	if len(*calls) != 1 || (*calls)[0].sig != syscall.SIGKILL {
		t.Fatalf("signals through Reap = %v, want exactly one SIGKILL", *calls)
	}
	// The group had been signalled before the reap, so post-reap
	// signal calls report the reached goal state (nil) — silently.
	if err := s.Terminate(time.Second); err != nil {
		t.Fatalf("Terminate after Reap = %v, want nil (group was signalled)", err)
	}
	if err := s.Kill(); err != nil {
		t.Fatalf("Kill after Reap = %v, want nil (group was signalled)", err)
	}
	if len(*calls) != 1 {
		t.Fatalf("signals after Reap = %v, want no signal attempted after Reap (still exactly the one from Reap)", *calls)
	}
}

func TestSessionSignalsAfterReapOfVanishedGroupStaySilent(t *testing.T) {
	calls := installGroupSignalSpy(t)
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Vanish the group behind the Session's back — directly, not via
	// the spy — so no Session signal ever lands: kill the only member
	// and reap it with a direct Wait.
	if err := syscall.Kill(s.PID(), syscall.SIGKILL); err != nil {
		t.Fatalf("kill leader: %v", err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("direct Wait of a SIGKILLed leader returned nil")
	}
	// Reap now finds the group already gone: its own kill attempt
	// reports ESRCH and is tolerated; the stored result is the second
	// Wait's error.
	if err := s.Reap(); err == nil {
		t.Fatal("Reap after a direct Wait returned nil, want the stored wait error")
	}
	throughReap := len(*calls)
	// The group vanished before any signal landed, so post-reap signal
	// calls report os.ErrProcessDone — silently.
	if err := s.Terminate(time.Second); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Terminate after Reap of a vanished group = %v, want os.ErrProcessDone", err)
	}
	if err := s.Kill(); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("Kill after Reap of a vanished group = %v, want os.ErrProcessDone", err)
	}
	if len(*calls) != throughReap {
		t.Fatalf("signals after Reap = %v, want no signal attempted after Reap (still %d through Reap)", *calls, throughReap)
	}
}

// sessionTerminateScenario is plan scenario (iii): a group that dies of
// its TERM makes Terminate return nil, promptly — the grace is an event
// wait, not a sleep.
func sessionTerminateScenario(t *testing.T) {
	t.Helper()
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	start := time.Now()
	if err := s.Terminate(10 * time.Second); err != nil {
		t.Fatalf("Terminate: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 5*time.Second {
		t.Fatalf("Terminate took %v with a 10s grace; the grace wait is not event-based", elapsed)
	}
	if !waitProcessStopped(s.PID(), 5*time.Second) {
		t.Fatal("leader survived Terminate")
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a terminated leader returned nil, want the wait error")
	}
}

func TestSessionTerminateExitsPromptlyOnTerm(t *testing.T) {
	sessionTerminateScenario(t)
}

// TestSessionFallbackParity forces the pidfd seam to fail and re-runs the
// ordering and terminate scenarios against the WNOWAIT probe path: the
// fallback must behave identically.
func TestSessionFallbackParity(t *testing.T) {
	orig := pidfdOpenFn
	pidfdOpenFn = func(int, int) (int, error) {
		return -1, errors.New("pidfd forced unavailable (test seam)")
	}
	defer func() { pidfdOpenFn = orig }()
	sessionOrderingScenario(t)
	sessionTerminateScenario(t)
}

func TestSessionWaitExitHonorsCancellation(t *testing.T) {
	before := runtime.NumGoroutine()
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := s.WaitExit(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitExit error = %v, want context deadline", err)
	}
	exited, err := leaderExitedNoReap(s.PID())
	if err != nil || exited {
		t.Fatalf("leaderExitedNoReap = %v, %v; want false, nil for a running leader", exited, err)
	}
	if err := s.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a SIGKILLed leader returned nil, want the wait error")
	}
	// The pidfd poll goroutine must not outlive the session past Reap.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && runtime.NumGoroutine() > before {
		time.Sleep(10 * time.Millisecond)
	}
	if got := runtime.NumGoroutine(); got > before {
		t.Fatalf("goroutines after Reap = %d, want <= %d (poll goroutine leaked)", got, before)
	}
}

// The Session half of Codex finding F3: repeated timed-out WaitExit
// calls against a still-running leader must not accumulate goroutines
// while the leader is alive and unreaped — the assertion runs before
// Kill/Reap on purpose, because a kill would release pre-fix parked
// goroutines and mask the leak.
func TestSessionWaitExitCancelledWaitsDoNotLeakGoroutines(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	baseline := runtime.NumGoroutine()
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		err := s.WaitExit(ctx)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("cancelled WaitExit %d error = %v, want context deadline", i, err)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && runtime.NumGoroutine() > baseline {
		time.Sleep(10 * time.Millisecond)
	}
	if got := runtime.NumGoroutine(); got > baseline {
		t.Fatalf("goroutines with the leader alive = %d, want <= baseline %d (cancelled WaitExit leaked goroutines)", got, baseline)
	}
	if err := s.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a SIGKILLed leader returned nil, want the wait error")
	}
}

// installCountingPidfdOpen wraps pidfdOpenFn with an attempt counter
// that still delegates to the real pidfd_open, so behavior is unchanged
// while every attempt becomes observable. Tests are serial in this
// package, so a plain int is a safe record.
func installCountingPidfdOpen(t *testing.T) *int {
	t.Helper()
	opens := 0
	orig := pidfdOpenFn
	pidfdOpenFn = func(pid int, flags int) (int, error) {
		opens++
		return orig(pid, flags)
	}
	t.Cleanup(func() { pidfdOpenFn = orig })
	return &opens
}

// installFailingCountingPidfdOpen forces the WNOWAIT fallback (the
// TestSessionFallbackParity seam) and counts the attempts: a fallback
// Session's WaitExit reaches pidfdOpenFn through waitLeaderExit on
// every call, so the counter shows whether a call tried to open a
// pidfd on the leader PID at all.
func installFailingCountingPidfdOpen(t *testing.T) *int {
	t.Helper()
	opens := 0
	orig := pidfdOpenFn
	pidfdOpenFn = func(int, int) (int, error) {
		opens++
		return -1, errors.New("pidfd forced unavailable (test seam)")
	}
	t.Cleanup(func() { pidfdOpenFn = orig })
	return &opens
}

// Codex fix-round-4 finding: WaitExit is documented as safe to run
// concurrently with Reap, but an observer whose registration lands
// after Reap completed fell into waitLeaderExit with the stored PID
// and fd -1 — reopening a pidfd on a PID the session no longer owns.
// Normally that surfaces as an ECHILD from the probe instead of the
// completed exit the caller was promised; if the kernel has recycled
// the PID, the observer instead waits on an unrelated process. (This
// is the scenario fix round 2's review graded a P3 residual for PR-2
// on the premise that post-Reap WaitExit is out-of-contract use; the
// concurrent interleaving defeats that premise, so the fix lands
// here.) The recorded state already proves the leader exited: after
// Reap, WaitExit must return nil without any pidfd attempt against
// the stored PID.
func TestSessionWaitExitAfterReapReturnsRecordedExit(t *testing.T) {
	opens := installCountingPidfdOpen(t)
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if *opens != 1 {
		t.Fatalf("pidfd attempts through Start = %d, want 1 (the session pidfd)", *opens)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a live group returned nil, want the wait error")
	}
	*opens = 0
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	if err := s.WaitExit(ctx); err != nil {
		t.Fatalf("WaitExit after Reap = %v, want nil (Reap's stored result already proves the leader exited)", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("WaitExit after Reap took %v, want a prompt return from recorded state", elapsed)
	}
	if *opens != 0 {
		t.Fatalf("pidfd attempts after Reap = %d, want 0 (the released PID may have been recycled)", *opens)
	}
	// Repeatable: a second post-Reap WaitExit answers the same way,
	// still without touching the stored PID.
	if err := s.WaitExit(ctx); err != nil {
		t.Fatalf("second WaitExit after Reap = %v, want nil", err)
	}
	if *opens != 0 {
		t.Fatalf("pidfd attempts after repeated WaitExit = %d, want 0", *opens)
	}
}

// The fallback half of the same finding: with the pidfd seam forced
// to fail, the session runs on the WNOWAIT probe. Fallback observer
// registration after Reap must be gated identically — no pidfd
// attempt and no probe against a PID whose ownership has ended.
func TestSessionWaitExitAfterReapFallbackReturnsRecordedExit(t *testing.T) {
	opens := installFailingCountingPidfdOpen(t)
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// The fallback is engaged before Reap: a live observation runs
	// through waitLeaderExit, which attempts the open and probes.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := s.WaitExit(ctx); !errors.Is(err, context.DeadlineExceeded) {
		cancel()
		t.Fatalf("live fallback WaitExit = %v, want context deadline", err)
	}
	cancel()
	if *opens == 0 {
		t.Fatal("live fallback WaitExit made no pidfd attempt; the fallback path was not engaged")
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a live group returned nil, want the wait error")
	}
	*opens = 0
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := s.WaitExit(ctx2); err != nil {
		t.Fatalf("fallback WaitExit after Reap = %v, want nil (Reap's stored result already proves the leader exited)", err)
	}
	if *opens != 0 {
		t.Fatalf("pidfd attempts after Reap = %d, want 0 (fallback registration must not reopen a PID whose ownership has ended)", *opens)
	}
}

// The concurrent interleaving from the finding, exercised directly:
// wave A observers register before Reap (the round-3 duplicated-
// descriptor path); wave B observers launch only after Reap has
// returned, so their registration can only land on the reaped state.
// Every observer must report the leader's exit (nil), promptly —
// including wave B, which the pre-fix code answered with an ECHILD
// from the probe or a wait on a recycled PID.
func TestSessionWaitExitRacingReapRegistration(t *testing.T) {
	for i := 0; i < 25; i++ {
		cmd := exec.Command("sleep", "30")
		s := NewSession(cmd)
		if err := s.Start(); err != nil {
			t.Fatalf("iteration %d: Start: %v", i, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		errs := make(chan error, 4)
		for k := 0; k < 2; k++ {
			go func() { errs <- s.WaitExit(ctx) }()
		}
		// Let wave A enter its wait on the live leader, then reap.
		time.Sleep(20 * time.Millisecond)
		if err := s.Reap(); err == nil {
			cancel()
			t.Fatalf("iteration %d: Reap of a live group returned nil, want the wait error", i)
		}
		// Wave B: registration can only happen after Reap completed.
		for k := 0; k < 2; k++ {
			go func() { errs <- s.WaitExit(ctx) }()
		}
		for k := 0; k < 4; k++ {
			select {
			case err := <-errs:
				if err != nil {
					cancel()
					t.Fatalf("iteration %d: observer %d = %v, want nil (every registration ordering must observe the completed exit)", i, k, err)
				}
			case <-time.After(5 * time.Second):
				cancel()
				t.Fatalf("iteration %d: observer %d did not return after Reap completed", i, k)
			}
		}
		cancel()
	}
}

// Codex fix-round-3 finding N1: WaitExit is documented as safe to run
// concurrently with Reap, but it read s.pidfd without synchronization
// while Reap closed and overwrote it. An observer could copy the
// descriptor number, be suspended before polling it, and — once Reap
// closed the descriptor and another operation reused the number —
// poll an unrelated resource instead of the leader's pidfd; and the
// unsynchronized read/write pair is a data race. The observer below is
// parked inside its wait on the live leader before Reap runs, so each
// iteration forces the overlap: under -race the detector reports the
// pidfd race on the pre-fix code, and the observer must still return
// the leader's exit (nil), promptly, on its own duplicated descriptor.
func TestSessionWaitExitConcurrentWithReap(t *testing.T) {
	for i := 0; i < 25; i++ {
		cmd := exec.Command("sleep", "30")
		s := NewSession(cmd)
		if err := s.Start(); err != nil {
			t.Fatalf("iteration %d: Start: %v", i, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		waitErr := make(chan error, 1)
		go func() { waitErr <- s.WaitExit(ctx) }()
		// Let the observer enter its wait on the live leader before
		// Reap closes the session pidfd underneath it.
		time.Sleep(20 * time.Millisecond)
		if err := s.Reap(); err == nil {
			cancel()
			t.Fatalf("iteration %d: Reap of a live group returned nil, want the wait error", i)
		}
		select {
		case err := <-waitErr:
			if err != nil {
				cancel()
				t.Fatalf("iteration %d: WaitExit racing Reap = %v, want nil (it must observe the leader's exit on a descriptor Reap cannot close)", i, err)
			}
		case <-time.After(5 * time.Second):
			cancel()
			t.Fatalf("iteration %d: WaitExit did not return after Reap completed (left polling a descriptor Reap closed)", i)
		}
		cancel()
	}
}

// Codex fix-round-4 re-review, §7: the round-4 registration gate
// protects observers that register after Reap, and the pidfd route
// is secured at registration by the duplicated descriptor — but an
// observer already in flight on the fd -1 route when Reap completes
// was still answered with the probe's ECHILD. On that route — every
// WaitExit of a fallback session (pidfd -1 from Start) and any call
// whose duplication fails — registration secured nothing: the
// observer unlocked and waitLeaderExit worked against the bare
// stored PID, and a WNOWAIT probe structurally cannot observe an
// already-reaped leader. The reviewer reproduced ECHILD in 100/100
// probe iterations of exactly this shape: fallback forced, observer
// parked inside its probe wait on a live leader, then Reap. The
// in-flight wait must resolve to the recorded exit: once Reap has
// run, the leader's exit is a fact in the session's stored state,
// and WaitExit reports it (nil) instead of the observation artifact.
func TestSessionWaitExitFallbackInFlightRacingReap(t *testing.T) {
	installFailingCountingPidfdOpen(t)
	for i := 0; i < 25; i++ {
		cmd := exec.Command("sleep", "30")
		s := NewSession(cmd)
		if err := s.Start(); err != nil {
			t.Fatalf("iteration %d: Start: %v", i, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		waitErr := make(chan error, 1)
		go func() { waitErr <- s.WaitExit(ctx) }()
		// Park the observer inside its probe wait on the live
		// leader (the probe ticks every 5 ms), then reap out from
		// under it.
		time.Sleep(20 * time.Millisecond)
		if err := s.Reap(); err == nil {
			cancel()
			t.Fatalf("iteration %d: Reap of a live group returned nil, want the wait error", i)
		}
		select {
		case err := <-waitErr:
			if err != nil {
				cancel()
				t.Fatalf("iteration %d: in-flight fallback WaitExit racing Reap = %v, want nil (Reap recorded the exit while the probe was in flight)", i, err)
			}
		case <-time.After(5 * time.Second):
			cancel()
			t.Fatalf("iteration %d: in-flight fallback WaitExit did not return after Reap completed", i)
		}
		cancel()
	}
}

// The transient-pidfd sub-route of the same §7 gap: a session whose
// Start could not open a pidfd (stored pidfd -1) on a machine where
// pidfd is in fact available by WaitExit time — and, by the same
// construction, a pidfd session whose descriptor duplication fails —
// reaches WaitExit holding no descriptor. Registration must bind
// the transient pidfd while ownership is still held (inside the
// registration lock), so a Reap landing after registration cannot
// leave the observer opening a released PID afterwards. The seam
// makes the interleaving deterministic: Start's open fails; the
// WaitExit-time open signals its entry and blocks; Reap is launched
// while the open is blocked. Pre-fix, WaitExit holds no lock during
// that open, so Reap runs to completion and the released open fails
// ESRCH, leaving the probe to answer ECHILD. Post-fix, the open
// happens inside the registration lock, so Reap is still parked on
// the mutex when the open is released and the open binds the live
// leader; the wait then observes the exit Reap performs.
func TestSessionWaitExitTransientBindRacingReap(t *testing.T) {
	orig := pidfdOpenFn
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	blocking := false
	pidfdOpenFn = func(pid int, flags int) (int, error) {
		if !blocking {
			return -1, errors.New("pidfd unavailable at Start (test seam)")
		}
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
		return orig(pid, flags)
	}
	t.Cleanup(func() { pidfdOpenFn = orig })

	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if s.pidfd != -1 {
		t.Fatalf("session pidfd = %d after failed Start open, want -1 (fallback state)", s.pidfd)
	}
	blocking = true
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	waitErr := make(chan error, 1)
	go func() { waitErr <- s.WaitExit(ctx) }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("WaitExit never attempted the transient pidfd open")
	}
	reapErr := make(chan error, 1)
	go func() { reapErr <- s.Reap() }()
	// Give Reap its chance to run: pre-fix it completes while the
	// open is blocked; post-fix it stays parked on the registration
	// mutex — which is exactly the ownership guarantee under test.
	time.Sleep(50 * time.Millisecond)
	close(release)
	select {
	case err := <-waitErr:
		if err != nil {
			t.Fatalf("WaitExit with transient bind racing Reap = %v, want nil (the transient pidfd must be bound while the session still owns the PID)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WaitExit did not return after the transient open was released")
	}
	select {
	case err := <-reapErr:
		if err == nil {
			t.Fatal("Reap of a live group returned nil, want the wait error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Reap did not return after the transient open was released")
	}
}

// The boundary of the §7 resolution: an in-flight fallback observer
// whose context is cancelled while the session is NOT reaped must
// still get the context error. The recorded exit can replace a
// wait's error only when Reap actually ran during the wait; a wait
// error on a never-reaped session propagates unchanged. (This pin
// holds before and after the fix — it fences the conversion.)
func TestSessionWaitExitFallbackCancelInFlightWithoutReap(t *testing.T) {
	installFailingCountingPidfdOpen(t)
	cmd := exec.Command("sleep", "30")
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	waitErr := make(chan error, 1)
	go func() { waitErr <- s.WaitExit(ctx) }()
	// Park the observer inside its probe wait, then cancel without
	// any Reap: the context error is the wait's own outcome.
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-waitErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled in-flight fallback WaitExit = %v, want context.Canceled (no Reap ran, so no recorded exit can replace the wait's own error)", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled in-flight fallback WaitExit did not return")
	}
	if err := s.Kill(); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of a SIGKILLed leader returned nil, want the wait error")
	}
}

// waitSessionMutexHeld blocks until the session mutex is observed
// held by another goroutine (a lifecycle method), probing with
// TryLock. It makes the mutex-contention tests deterministic: the
// observer under test is only launched once the holder is provably
// inside its critical section.
func waitSessionMutexHeld(t *testing.T, s *Session) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if s.mu.TryLock() {
			s.mu.Unlock()
		} else {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("session mutex never became held by the lifecycle method")
		}
		runtime.Gosched()
	}
}

// termIgnoringSession starts a session whose leader ignores SIGTERM:
// the trap installs SIG_IGN, which survives the exec into sleep (only
// caught handlers reset across exec), so only SIGKILL stops the
// leader. A Terminate against it spends its whole grace in its
// internal wait — holding the session mutex for exactly that long.
// The ready file is written after the trap is installed and before
// the exec: without that handshake a Terminate launched immediately
// after Start can land its SIGTERM in the window before the trap
// exists and kill the leader outright, collapsing the scenario.
func termIgnoringSession(t *testing.T) *Session {
	t.Helper()
	readyFile := filepath.Join(t.TempDir(), "leader.ready")
	script := `trap '' TERM; echo ready > "$1"; exec sleep 30`
	cmd := exec.Command("/bin/sh", "-c", script, "sh", readyFile)
	s := NewSession(cmd)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = syscall.Kill(-s.PID(), syscall.SIGKILL) })
	deadline := time.Now().Add(5 * time.Second)
	for {
		if data, err := os.ReadFile(readyFile); err == nil && strings.TrimSpace(string(data)) == "ready" {
			return s
		}
		if time.Now().After(deadline) {
			t.Fatal("TERM-ignoring leader never signalled readiness")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// Codex fix-round-6 finding G1 (registration site): WaitExit takes
// the session mutex to register, and Terminate holds that mutex for
// its entire grace while Reap holds it for its entire cmd.Wait —
// so a plain Lock made the observer's context powerless for the
// holder's whole duration. Codex's reproduction: a TERM-ignoring
// leader and a 600 ms termination grace made a 20 ms WaitExit
// deadline return after ~600 ms with nil. Registration must wait
// for the mutex cancellation-awarely: if the context ends first,
// WaitExit returns the context error at the deadline, without ever
// registering (and so without touching the leader PID).
func TestSessionWaitExitRegistrationHonorsContextWhileMutexHeld(t *testing.T) {
	s := termIgnoringSession(t)
	termErr := make(chan error, 1)
	go func() { termErr <- s.Terminate(600 * time.Millisecond) }()
	waitSessionMutexHeld(t, s)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := s.WaitExit(ctx)
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitExit with a 20ms deadline against a mutex held for a 600ms grace = %v after %v, want context.DeadlineExceeded at the deadline (registration blocked on the mutex past the context)", err, elapsed)
	}
	if elapsed >= 300*time.Millisecond {
		t.Fatalf("WaitExit returned after %v, want the ~20ms deadline, well under half the 600ms grace", elapsed)
	}
	if err := <-termErr; err != nil {
		t.Fatalf("Terminate: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of the SIGKILLed leader returned nil, want the wait error")
	}
}

// Codex fix-round-6 finding G1 (completion-state re-check site):
// after its wait errors, WaitExit re-takes the session mutex to
// check whether Reap completed during the wait (the round-5
// recorded-exit resolution). That second acquisition had the same
// defect: an observer whose context expires while a lifecycle
// method holds the mutex stays blocked past its deadline even
// though its wait has already produced the context error. Here the
// observer registers and parks in its wait first; Terminate then
// holds the mutex across the observer's 250 ms deadline.
func TestSessionWaitExitRecheckHonorsContextWhileMutexHeld(t *testing.T) {
	s := termIgnoringSession(t)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	waitErr := make(chan error, 1)
	start := time.Now()
	go func() { waitErr <- s.WaitExit(ctx) }()
	// Let the observer register (the mutex is free; registration
	// is brief) and park inside its wait on the live leader before
	// Terminate takes the mutex for its grace.
	time.Sleep(50 * time.Millisecond)
	termErr := make(chan error, 1)
	go func() { termErr <- s.Terminate(600 * time.Millisecond) }()
	select {
	case err := <-waitErr:
		elapsed := time.Since(start)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("in-flight WaitExit whose deadline passed while Terminate held the mutex = %v after %v, want context.DeadlineExceeded", err, elapsed)
		}
		// The mutex frees at ~650 ms (50 ms offset + 600 ms
		// grace); the re-check must answer at the 250 ms
		// deadline instead. 450 ms leaves scheduling slack on
		// both sides of the discrimination.
		if elapsed >= 450*time.Millisecond {
			t.Fatalf("WaitExit returned after %v, want the ~250ms deadline; the post-error state re-check blocked on the held mutex", elapsed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("WaitExit did not return")
	}
	if err := <-termErr; err != nil {
		t.Fatalf("Terminate: %v", err)
	}
	if err := s.Reap(); err == nil {
		t.Fatal("Reap of the SIGKILLed leader returned nil, want the wait error")
	}
}
