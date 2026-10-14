package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRunPropagatesCommandCancellationToHelper(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("financial fixture helper process-tree cleanup is supported only on Linux")
	}
	for _, test := range []struct {
		action string
		tool   string
	}{
		{action: "author", tool: "go"},
		{action: "observe", tool: "pdftotext"},
	} {
		t.Run(test.action, func(t *testing.T) {
			directory := t.TempDir()
			marker := filepath.Join(directory, "started")
			tool := filepath.Join(directory, test.tool)
			script := "#!/bin/sh\n: > \"$COMMAND_MARKER\"\nexec sleep 1\n"
			if err := os.WriteFile(tool, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("COMMAND_MARKER", marker)
			if test.action == "author" {
				t.Setenv("GOMODCACHE", "")
			}
			t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))

			root, err := filepath.Abs(filepath.Join("..", ".."))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := make(chan error, 1)
			go func() {
				result <- run(ctx, []string{test.action, "--repo-root", root})
			}()

			deadline := time.Now().Add(3 * time.Second)
			for {
				if _, err := os.Stat(marker); err == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("helper did not start before the command cancellation test timed out")
				}
				time.Sleep(5 * time.Millisecond)
			}
			cancel()
			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("run error = %v, want context.Canceled", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("command did not return after its context was canceled")
			}
		})
	}
}

func TestRunRequiresCommandContext(t *testing.T) {
	if err := run(nil, []string{"observe"}); err == nil {
		t.Fatal("run accepted a nil command context")
	}
}
