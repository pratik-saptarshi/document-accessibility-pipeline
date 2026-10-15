//go:build linux || darwin

package financialfixtures

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// openRegularInput opens path without following symlinks and without
// blocking on FIFOs, then requires the opened file to be regular.
// A symlink fails at open with ELOOP; a FIFO opens immediately because of
// O_NONBLOCK and is rejected by the regular-file check.
func openRegularInput(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		if errors.Is(err, syscall.ELOOP) {
			return nil, fmt.Errorf("input %s is not a regular file", path)
		}
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, fmt.Errorf("input %s is not a regular file", path)
	}
	return file, nil
}
