package cloudtools

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ContextWithSignals cancels bootstrap work when the process receives an
// interrupt or termination request, allowing owned subprocess groups and locks
// to be cleaned up before the command exits.
func ContextWithSignals(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}
