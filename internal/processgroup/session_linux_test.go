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
