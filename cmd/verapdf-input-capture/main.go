// Command verapdf-input-capture copies a validator input file into a private
// snapshot, validating its type at open and bounding the captured bytes.
//
// Usage: verapdf-input-capture <input> <output> <max-bytes>
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: verapdf-input-capture <input> <output> <max-bytes>")
		os.Exit(2)
	}
	maxBytes, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil || maxBytes < 0 {
		fmt.Fprintln(os.Stderr, "max-bytes must be a non-negative integer")
		os.Exit(2)
	}
	err = captureBoundedInput(os.Args[1], os.Args[2], maxBytes)
	if err == nil {
		return
	}
	// The script's tests match on these exact messages.
	switch {
	case errors.Is(err, errInputNotRegular):
		fmt.Fprintln(os.Stderr, "input must be an absolute regular non-symlink file")
	case errors.Is(err, errInputLimit):
		fmt.Fprintln(os.Stderr, "input exceeds the declared validator-input limit")
	default:
		fmt.Fprintln(os.Stderr, "failed to capture bounded validator input snapshot")
	}
	os.Exit(1)
}
