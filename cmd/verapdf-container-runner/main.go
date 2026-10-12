// Command verapdf-container-runner captures bounded output from the isolated
// CI validator container while preserving its exit status separately from
// host-side output and execution failures.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errOutputLimit = errors.New("validator container output limit exceeded")
	errOutputSink  = errors.New("validator container output sink failed")
	errExecution   = errors.New("validator container execution failed")
)

type outputBudget struct {
	mu        sync.Mutex
	remaining int64
	exceeded  bool
	sinkError bool
}

type boundedOutput struct {
	file   *os.File
	budget *outputBudget
}

func (w *boundedOutput) Write(p []byte) (int, error) {
	w.budget.mu.Lock()
	defer w.budget.mu.Unlock()
	if len(p) == 0 {
		return 0, nil
	}
	writeSize := len(p)
	if int64(writeSize) >= w.budget.remaining {
		writeSize = int(w.budget.remaining)
		w.budget.remaining = 0
		w.budget.exceeded = true
	} else {
		w.budget.remaining -= int64(writeSize)
	}
	if writeSize > 0 {
		n, err := w.file.Write(p[:writeSize])
		if err != nil {
			w.budget.sinkError = true
			return n, err
		}
		if n != writeSize {
			w.budget.sinkError = true
			return n, io.ErrShortWrite
		}
	}
	// Discard bytes beyond the shared budget while continuing to drain both
	// pipes. Exact-limit output is rejected as well as output above the limit.
	return len(p), nil
}

func runBoundedCommand(ctx context.Context, executable string, args []string, stdoutPath, stderrPath string, maxBytes int64) (int, error) {
	if maxBytes <= 0 || executable == "" || stdoutPath == "" || stderrPath == "" {
		return -1, errExecution
	}
	stdout, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return -1, errOutputSink
	}
	stderr, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		_ = stdout.Close()
		return -1, errOutputSink
	}
	budget := &outputBudget{remaining: maxBytes}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout = &boundedOutput{file: stdout, budget: budget}
	cmd.Stderr = &boundedOutput{file: stderr, budget: budget}
	runErr := cmd.Run()
	stdoutCloseErr := stdout.Close()
	stderrCloseErr := stderr.Close()

	budget.mu.Lock()
	sinkFailed := budget.sinkError || stdoutCloseErr != nil || stderrCloseErr != nil
	outputExceeded := budget.exceeded
	budget.mu.Unlock()
	if sinkFailed {
		return commandExitCode(runErr), errOutputSink
	}
	if ctx.Err() != nil {
		return commandExitCode(runErr), errExecution
	}
	if outputExceeded {
		return commandExitCode(runErr), errOutputLimit
	}
	if runErr == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) && exitErr.ExitCode() >= 0 {
		return exitErr.ExitCode(), nil
	}
	return commandExitCode(runErr), errExecution
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func runCLI(args []string) int {
	if len(args) < 6 || args[4] != "--" {
		fmt.Fprintln(os.Stderr, "usage: verapdf-container-runner <timeout> <max-output-bytes> <stdout-file> <stderr-file> -- <command> [args...]")
		return 2
	}
	timeout, err := time.ParseDuration(args[0])
	if err != nil || timeout <= 0 {
		fmt.Fprintln(os.Stderr, "invalid validator timeout")
		return 2
	}
	maxBytes, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || maxBytes <= 0 {
		fmt.Fprintln(os.Stderr, "invalid validator output limit")
		return 2
	}
	if strings.TrimSpace(args[5]) == "" {
		fmt.Fprintln(os.Stderr, "validator command is required")
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	status, err := runBoundedCommand(ctx, args[5], args[6:], args[2], args[3], maxBytes)
	if err != nil {
		switch {
		case errors.Is(err, errOutputLimit):
			fmt.Fprintln(os.Stderr, "validator container output limit exceeded")
		case errors.Is(err, errOutputSink):
			fmt.Fprintln(os.Stderr, "validator container output sink failed")
		default:
			fmt.Fprintln(os.Stderr, "validator container execution failed")
		}
		return 2
	}
	if status < 0 || status > 255 {
		fmt.Fprintln(os.Stderr, "validator container returned an invalid exit status")
		return 2
	}
	return status
}

func main() {
	os.Exit(runCLI(os.Args[1:]))
}
