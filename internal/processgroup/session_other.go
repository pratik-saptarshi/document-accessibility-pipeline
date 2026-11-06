//go:build !linux

package processgroup

import (
	"context"
	"os/exec"
	"time"
)

// Session is the Linux process-group lifecycle owner. Off Linux it
// fails closed, like the package's other lifecycle helpers: no
// direct-process fallback is an equivalent of its guarantees.
type Session struct {
	cmd *exec.Cmd
}

func NewSession(cmd *exec.Cmd) *Session { return &Session{cmd: cmd} }

func (s *Session) PID() int { return 0 }

func (s *Session) Start() error { return errUnsupported }

func (s *Session) WaitExit(context.Context) error { return errUnsupported }

func (s *Session) Terminate(time.Duration) error { return errUnsupported }

func (s *Session) Kill() error { return errUnsupported }

func (s *Session) Reap() error { return errUnsupported }
