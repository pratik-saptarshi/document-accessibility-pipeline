package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	// errInputNotRegular reports an input that is not a regular file at open:
	// a symlink (rejected by O_NOFOLLOW), FIFO, directory, or device.
	errInputNotRegular = errors.New("validator input is not a regular file")
	// errInputLimit reports an input whose captured bytes exceed the limit.
	errInputLimit = errors.New("input exceeds the declared validator-input limit")
	// errCaptureFailed reports any other open, copy, or staging failure.
	errCaptureFailed = errors.New("failed to capture bounded validator input snapshot")
)

// captureBoundedInput copies input into output, validating the input's type
// at open and bounding the captured bytes. The output is created exclusive
// with 0600, chmodded to 0444 on success, and removed on any failure, so a
// failed capture never leaves a partial snapshot behind.
func captureBoundedInput(input, output string, maxBytes int64) error {
	if maxBytes < 0 {
		return fmt.Errorf("%w: negative limit %d", errCaptureFailed, maxBytes)
	}
	src, err := openInput(input)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("%w: stat: %w", errCaptureFailed, err)
	}
	if !info.Mode().IsRegular() {
		return errInputNotRegular
	}

	dst, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("%w: create output: %w", errCaptureFailed, err)
	}
	failed := true
	defer func() {
		dst.Close()
		if failed {
			os.Remove(output)
		}
	}()

	// Read at most maxBytes+1 so growth during capture is detected instead
	// of trusted from a pre-copy size check.
	written, err := io.CopyN(dst, src, maxBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: copy: %w", errCaptureFailed, err)
	}
	if written > maxBytes {
		return errInputLimit
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("%w: close output: %w", errCaptureFailed, err)
	}
	if err := os.Chmod(output, 0o444); err != nil {
		return fmt.Errorf("%w: chmod output: %w", errCaptureFailed, err)
	}
	failed = false
	return nil
}
