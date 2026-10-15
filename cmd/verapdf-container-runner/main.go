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

const commandPipeWaitDelay = 500 * time.Millisecond

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
	if w.budget.sinkError {
		return len(p), nil
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
			return len(p), nil
		}
		if n != writeSize {
			w.budget.sinkError = true
			return len(p), nil
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
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		return -1, errOutputSink
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return -1, errOutputSink
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	cleanupProcessTree := configureProcessTree(cmd)
	if err := cmd.Start(); err != nil {
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
		_ = stderrReader.Close()
		_ = stderrWriter.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return -1, errExecution
	}
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	drainResults := make(chan error, 2)
	go func() {
		_, err := io.Copy(&boundedOutput{file: stdout, budget: budget}, stdoutReader)
		drainResults <- err
	}()
	go func() {
		_, err := io.Copy(&boundedOutput{file: stderr, budget: budget}, stderrReader)
		drainResults <- err
	}()
	// On Linux, wait for the leader's exit without reaping it, so the drain
	// loop's process-group signals below happen BEFORE the leader is reaped,
	// closing the PID-reuse window. The pipe-holding descendants must stay
	// alive until the drain deadline so an incomplete drain is still detected.
	// Where process groups are unavailable the previous reap-then-drain order
	// is kept.
	reaped, runErr := awaitLeaderExit(ctx, cmd)

	drainIncomplete := false
	var drainError error
	drainDeadline := time.NewTimer(commandPipeWaitDelay)
	defer drainDeadline.Stop()
	drainTimeout := drainDeadline.C
	remainingDrains := 2
	contextDone := ctx.Done()
	for remainingDrains > 0 {
		select {
		case err := <-drainResults:
			remainingDrains--
			if err != nil {
				drainError = err
			}
		case <-contextDone:
			contextDone = nil
			_ = cleanupProcessTree()
		case <-drainTimeout:
			drainIncomplete = true
			_ = cleanupProcessTree()
			_ = stdoutReader.Close()
			_ = stderrReader.Close()
			drainTimeout = nil
		}
	}
	_ = cleanupProcessTree()
	_ = stdoutReader.Close()
	_ = stderrReader.Close()
	stdoutCloseErr := stdout.Close()
	stderrCloseErr := stderr.Close()

	// Reap the leader only after the process group has been signaled above,
	// so the group signals cannot hit a recycled PID.
	if !reaped {
		waitResult := make(chan error, 1)
		go func() {
			waitResult <- cmd.Wait()
		}()
		runErr = <-waitResult
	}

	budget.mu.Lock()
	sinkFailed := budget.sinkError || stdoutCloseErr != nil || stderrCloseErr != nil
	outputExceeded := budget.exceeded
	budget.mu.Unlock()
	status := commandExitCode(runErr)
	if drainIncomplete || drainError != nil {
		return status, errExecution
	}
	if sinkFailed {
		return status, errOutputSink
	}
	if ctx.Err() != nil {
		return status, errExecution
	}
	if outputExceeded {
		return status, errOutputLimit
	}
	if runErr == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) && exitErr.ExitCode() >= 0 {
		return exitErr.ExitCode(), nil
	}
	return status, errExecution
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
