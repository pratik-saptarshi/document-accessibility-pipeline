//go:build !linux

package main

import (
	"fmt"
	"os"
)

// openInput fails closed off Linux: descriptor-level type validation relies
// on O_NOFOLLOW|O_NONBLOCK, which this helper does not emulate elsewhere.
// The validator job runs on Linux.
func openInput(_ string) (*os.File, error) {
	return nil, fmt.Errorf("%w: input capture requires Linux", errCaptureFailed)
}
