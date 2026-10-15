//go:build linux

package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// openInput opens input without following symlinks and without blocking on
// FIFOs. A symlink fails here with ELOOP; a FIFO opens immediately because
// of O_NONBLOCK and is rejected by the caller's regular-file fstat check.
func openInput(input string) (*os.File, error) {
	f, err := os.OpenFile(input, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return nil, errInputNotRegular
		}
		return nil, fmt.Errorf("%w: open: %w", errCaptureFailed, err)
	}
	return f, nil
}
