//go:build linux

package financialfixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestV2ProcessLifecycleSignalsBeforeReleasingLeader(t *testing.T) {
	for _, waitResult := range []struct {
		name string
		err  error
	}{
		{name: "leader exit"},
		{name: "cancellation", err: context.Canceled},
	} {
		t.Run(waitResult.name, func(t *testing.T) {
			var events []string
			err := finishV2StartedProcess(context.Background(), 42,
				func(_ context.Context, pid int) error {
					if pid != 42 {
						t.Fatalf("wait PID = %d, want 42", pid)
					}
					events = append(events, "observe")
					return waitResult.err
				},
				func(pid int) error {
					if pid != 42 {
						t.Fatalf("signal PGID = %d, want 42", pid)
					}
					events = append(events, "signal")
					return nil
				},
				func() error {
					t.Fatalf("leader fallback kill was called after successful group signal")
					return nil
				},
				func() error {
					events = append(events, "release ownership")
					return nil
				},
			)
			if waitResult.err != nil && !errors.Is(err, waitResult.err) {
				t.Fatalf("lifecycle error = %v, want %v", err, waitResult.err)
			}
			if got := strings.Join(events, ","); got != "observe,signal,release ownership" {
				t.Fatalf("lifecycle order = %s, want observe,signal,release ownership", got)
			}
		})
	}
}

func TestV2ObserveLeaderExitUsesArchitectureAwareSiginfo(t *testing.T) {
	command := exec.Command("/bin/sh", "-c", "exit 0")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	waitErr := waitForV2LeaderOrCancel(ctx, command.Process.Pid)
	reapErr := command.Wait()
	if waitErr != nil {
		t.Fatalf("observe leader exit: %v", waitErr)
	}
	if reapedCommandError(reapErr, ctx) != nil {
		t.Fatalf("reap fast-success leader: %v", reapErr)
	}
}

func TestV2BoundedProcessDoesNotStartWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runV2BoundedProcess(ctx, filepath.Join(t.TempDir(), "must-not-start"), nil, nil, nil, nil, 25*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled run error = %v, want context.Canceled before start", err)
	}
}

func TestV2ProcessLifecycleKillsLeaderIfGroupSignalFails(t *testing.T) {
	var events []string
	wantSignalErr := errors.New("group signal denied")
	err := finishV2StartedProcess(context.Background(), 42,
		func(context.Context, int) error { events = append(events, "observe"); return nil },
		func(int) error { events = append(events, "group signal"); return wantSignalErr },
		func() error { events = append(events, "leader signal"); return nil },
		func() error { events = append(events, "reap"); return nil },
	)
	if !errors.Is(err, wantSignalErr) {
		t.Fatalf("lifecycle error = %v, want group-signal failure", err)
	}
	if got := strings.Join(events, ","); got != "observe,group signal,leader signal,reap" {
		t.Fatalf("failure lifecycle order = %s, want observe,group signal,leader signal,reap", got)
	}
}

func TestV2ProcessGroupCleanupCoversTerminalPaths(t *testing.T) {
	tests := []struct {
		name   string
		script string
		ctx    func() (context.Context, context.CancelFunc)
		limit  int
		want   func(error, context.Context, *limitedV2Output) error
	}{
		{
			name: "deadline",
			script: `sleep 30 &
record_child
wait
`,
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 100*time.Millisecond)
			},
			want: func(err error, ctx context.Context, _ *limitedV2Output) error {
				if !errors.Is(ctx.Err(), context.DeadlineExceeded) || err == nil {
					return fmt.Errorf("run error = %v, context error = %v; want deadline cancellation", err, ctx.Err())
				}
				return nil
			},
		},
		{
			name: "output overflow",
			script: `sleep 30 >/dev/null 2>&1 &
record_child
i=0
while [ "$i" -lt 100 ]; do
  printf x
  printf y >&2
  i=$((i + 1))
done
`,
			limit: 16,
			want: func(_ error, _ context.Context, output *limitedV2Output) error {
				if !output.overflow {
					return errors.New("bounded output did not report overflow")
				}
				return nil
			},
		},
		{
			name: "parent exit",
			script: `sleep 30 >/dev/null 2>&1 &
record_child
printf 'ok\n'
exit 0
`,
			limit: 16,
			want: func(err error, _ context.Context, _ *limitedV2Output) error {
				if err != nil {
					return fmt.Errorf("run error = %v, want clean parent exit", err)
				}
				return nil
			},
		},
		{
			name: "parent exit with inherited pipe",
			script: `sleep 30 &
record_child
exit 0
`,
			want: func(err error, _ context.Context, _ *limitedV2Output) error {
				if err != nil {
					return fmt.Errorf("run error = %v, want group cleanup and drained inherited pipe", err)
				}
				return nil
			},
		},
		{
			name: "parent error",
			script: `sleep 30 >/dev/null 2>&1 &
record_child
exit 17
`,
			want: func(err error, _ context.Context, _ *limitedV2Output) error {
				if err == nil {
					return errors.New("nonzero parent exit unexpectedly succeeded")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			childPIDFile := filepath.Join(dir, "child.pid")
			scriptPath := filepath.Join(dir, "helper")
			script := "#!/bin/sh\nrecord_child() { child=$!; start=$(awk '{print $22}' \"/proc/$child/stat\"); printf '%s %s\\n' \"$child\" \"$start\" > \"$V2_TEST_CHILD_PID_FILE\"; }\n" + tt.script
			if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { killRecordedProcessSafely(childPIDFile) })
			ctx := context.Background()
			cancel := func() {}
			if tt.ctx != nil {
				ctx, cancel = tt.ctx()
			}
			defer cancel()
			stdout := &limitedV2Output{limit: tt.limit}
			stderrLimit := 4096
			if tt.name == "output overflow" {
				stderrLimit = 16
			}
			stderr := &limitedV2Output{limit: stderrLimit}
			env := append(os.Environ(),
				"V2_TEST_CHILD_PID_FILE="+childPIDFile,
			)
			err := runV2BoundedProcess(ctx, scriptPath, nil, env, stdout, stderr, 25*time.Millisecond)
			if checkErr := tt.want(err, ctx, stdout); checkErr != nil {
				t.Fatal(checkErr)
			}
			if tt.name == "output overflow" && !stderr.overflow {
				t.Fatal("bounded stderr did not report overflow")
			}
			childPID, readErr := readPIDFile(childPIDFile)
			if readErr != nil {
				t.Fatalf("read child PID: %v", readErr)
			}
			if !waitV2ProcessStopped(childPID, time.Second) {
				t.Fatalf("descendant process %d remained alive after %s path", childPID, tt.name)
			}
		})
	}
}

func TestV2BoundedProcessStartFailure(t *testing.T) {
	err := runV2BoundedProcess(context.Background(), filepath.Join(t.TempDir(), "missing-helper"), nil, nil, nil, nil, 25*time.Millisecond)
	if err == nil {
		t.Fatal("missing helper unexpectedly started")
	}
}

func TestV2BoundedProcessRejectsZeroWaitDelay(t *testing.T) {
	err := runV2BoundedProcess(context.Background(), "/must/not/start", nil, nil, nil, nil, 0)
	if err == nil || !strings.Contains(err.Error(), "wait delay must be positive") {
		t.Fatalf("zero wait-delay error = %v, want fail-closed policy error", err)
	}
}

func readPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return 0, fmt.Errorf("invalid process identity %q", strings.TrimSpace(string(data)))
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid process ID %q", fields[0])
	}
	return pid, nil
}

func killRecordedProcessSafely(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 {
		return
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil || pid <= 0 {
		return
	}
	wantStart, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return
	}
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return
	}
	defer unix.Close(fd)
	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return
	}
	end := strings.LastIndexByte(string(stat), ')')
	if end < 0 {
		return
	}
	statFields := strings.Fields(string(stat[end+1:]))
	// starttime is field 22; after stripping pid and comm, it is index 19.
	if len(statFields) <= 19 {
		return
	}
	gotStart, err := strconv.ParseUint(statFields[19], 10, 64)
	if err != nil || gotStart != wantStart {
		return
	}
	_ = unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0)
}

func waitV2ProcessStopped(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err := syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return true
		}
		if err == nil {
			data, readErr := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
			if readErr == nil {
				if end := strings.LastIndexByte(string(data), ')'); end >= 0 {
					fields := strings.Fields(string(data[end+1:]))
					if len(fields) > 0 && (fields[0] == "Z" || fields[0] == "X") {
						return true
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
