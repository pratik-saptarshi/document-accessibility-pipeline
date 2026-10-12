package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRunBoundedCommandPreservesValidatorFindingExitAndStreams(t *testing.T) {
	stdout, stderr := filepath.Join(t.TempDir(), "stdout.xml"), filepath.Join(t.TempDir(), "stderr.txt")
	status, err := runBoundedCommand(context.Background(), "sh", []string{"-c", "printf report; printf warning >&2; exit 1"}, stdout, stderr, 64)
	if err != nil || status != 1 {
		t.Fatalf("runBoundedCommand() = %d, %v; want finding exit 1", status, err)
	}
	gotStdout, err := os.ReadFile(stdout)
	if err != nil || string(gotStdout) != "report" {
		t.Fatalf("stdout = %q, %v", gotStdout, err)
	}
	gotStderr, err := os.ReadFile(stderr)
	if err != nil || string(gotStderr) != "warning" {
		t.Fatalf("stderr = %q, %v", gotStderr, err)
	}
}

func TestRunBoundedCommandRejectsOutputAtLimit(t *testing.T) {
	stdout, stderr := filepath.Join(t.TempDir(), "stdout.xml"), filepath.Join(t.TempDir(), "stderr.txt")
	status, err := runBoundedCommand(context.Background(), "sh", []string{"-c", "printf 12345678; exit 1"}, stdout, stderr, 8)
	if !errors.Is(err, errOutputLimit) || status != 1 {
		t.Fatalf("runBoundedCommand() = %d, %v; want exit 1 plus output-limit error", status, err)
	}
	got, readErr := os.ReadFile(stdout)
	if readErr != nil || len(got) != 8 {
		t.Fatalf("bounded stdout size = %d, %v; want exactly 8 bytes", len(got), readErr)
	}
}

func TestRunBoundedCommandClassifiesStdoutSinkFailure(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full is unavailable")
	}
	stderr := filepath.Join(t.TempDir(), "stderr.txt")
	status, err := runBoundedCommand(context.Background(), "sh", []string{"-c", "printf report; exit 1"}, "/dev/full", stderr, 64)
	if !errors.Is(err, errOutputSink) {
		t.Fatalf("runBoundedCommand() = %d, %v; want an output-sink error", status, err)
	}
}

func TestRunBoundedCommandCapsCombinedStreams(t *testing.T) {
	stdout, stderr := filepath.Join(t.TempDir(), "stdout.xml"), filepath.Join(t.TempDir(), "stderr.txt")
	status, err := runBoundedCommand(context.Background(), "sh", []string{"-c", "printf abc; printf def >&2; exit 1"}, stdout, stderr, 6)
	if !errors.Is(err, errOutputLimit) || status != 1 {
		t.Fatalf("runBoundedCommand() = %d, %v; want combined output-limit error", status, err)
	}
	gotStdout, stdoutErr := os.ReadFile(stdout)
	gotStderr, stderrErr := os.ReadFile(stderr)
	if stdoutErr != nil || stderrErr != nil || len(gotStdout)+len(gotStderr) != 6 {
		t.Fatalf("captured sizes = %d + %d, errors %v and %v", len(gotStdout), len(gotStderr), stdoutErr, stderrErr)
	}
}
