//go:build !linux

package processgroup

import (
	"context"
	"errors"
	"time"
)

// errUnsupported fails closed off Linux: these helpers' guarantees
// require Linux process groups, and direct-process signalling is not an
// equivalent fallback.
var errUnsupported = errors.New("processgroup: Linux process groups required")

func LeaderExitedNoReap(int) (bool, error) { return false, errUnsupported }

func WaitForLeaderExitNoReap(context.Context, int) error { return errUnsupported }

func KillGroup(int) error { return errUnsupported }

func TerminateGroup(int, time.Duration) error { return errUnsupported }
