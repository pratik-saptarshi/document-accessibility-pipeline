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
)

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
echo "$!" > "$V2_TEST_CHILD_PID_FILE"
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
echo "$!" > "$V2_TEST_CHILD_PID_FILE"
i=0
while [ "$i" -lt 100 ]; do
  printf x
  i=$((i + 1))
done
`,
			limit: 16,
			want: func(_ error, _ context.Context, output *limitedV2Output) error {
				if !output.overflow {
					return errors.New("bounded stdout did not report overflow")
				}
				return nil
			},
		},
		{
			name: "parent exit",
			script: `sleep 30 >/dev/null 2>&1 &
echo "$!" > "$V2_TEST_CHILD_PID_FILE"
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
echo "$!" > "$V2_TEST_CHILD_PID_FILE"
exit 0
`,
			want: func(err error, _ context.Context, _ *limitedV2Output) error {
				if !errors.Is(err, exec.ErrWaitDelay) {
					return fmt.Errorf("run error = %v, want inherited-pipe wait-delay evidence", err)
				}
				return nil
			},
		},
		{
			name: "parent error",
			script: `sleep 30 >/dev/null 2>&1 &
echo "$!" > "$V2_TEST_CHILD_PID_FILE"
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
			parentPIDFile := filepath.Join(dir, "parent.pid")
			scriptPath := filepath.Join(dir, "helper")
			script := "#!/bin/sh\necho \"$$\" > \"$V2_TEST_PARENT_PID_FILE\"\n" + tt.script
			if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				killRecordedProcessGroup(parentPIDFile)
				killRecordedProcess(childPIDFile)
			})

			ctx := context.Background()
			cancel := func() {}
			if tt.ctx != nil {
				ctx, cancel = tt.ctx()
			}
			defer cancel()
			stdout := &limitedV2Output{limit: tt.limit}
			stderr := &limitedV2Output{limit: 4096}
			env := append(os.Environ(),
				"V2_TEST_CHILD_PID_FILE="+childPIDFile,
				"V2_TEST_PARENT_PID_FILE="+parentPIDFile,
			)
			err := runV2BoundedProcess(ctx, scriptPath, nil, env, stdout, stderr, 25*time.Millisecond)
			if checkErr := tt.want(err, ctx, stdout); checkErr != nil {
				t.Fatal(checkErr)
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
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, fmt.Errorf("invalid process ID %q", strings.TrimSpace(string(data)))
	}
	return pid, nil
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

func killRecordedProcess(path string) {
	pid, err := readPIDFile(path)
	if err == nil {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
}

func killRecordedProcessGroup(path string) {
	pid, err := readPIDFile(path)
	if err == nil {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
}
