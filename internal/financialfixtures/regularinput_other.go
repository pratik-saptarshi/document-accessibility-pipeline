//go:build !linux

package financialfixtures

import (
	"fmt"
	"os"
)

// openRegularInput opens path and requires the opened file to be regular.
// Off Linux this is best-effort: without O_NOFOLLOW|O_NONBLOCK, a symlink
// is followed and a FIFO open may block. The experiment runs on Linux CI,
// where the Linux implementation applies.
func openRegularInput(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
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
