//go:build !linux

package pdfua

import (
	"errors"
	"os"
	"os/exec"
)

func requireProcessTreeSupport() error {
	return errors.New("validator process-tree cleanup is supported only on Linux")
}

func openRegularNoFollow(string) (*os.File, error) {
	return nil, errors.New("safe validator input opening is supported only on Linux")
}

func configureProcessTree(*exec.Cmd) (func() error, error) {
	return nil, errors.New("validator process-tree cleanup is supported only on Linux")
}
