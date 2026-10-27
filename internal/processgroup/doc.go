// Package processgroup provides the small Linux process-group primitives
// shared by the root module's bounded subprocess runners: observing a
// process-group leader's exit without reaping it, and signalling the
// group. Keeping the leader unreaped until the group has been signalled
// prevents its PID — and therefore its process-group ID, since a Setpgid
// leader's PGID equals its PID — from being recycled before cleanup.
//
// The primitives are Linux-only. On other platforms they fail closed;
// callers that require process-group cleanup must not fall back to
// direct-process cancellation as an equivalent.
package processgroup

import "errors"

// ErrInvalidPID is returned by helpers that require a positive process
// ID when pid is zero or negative. Signalling helpers never signal on an
// invalid PID: in particular, pid 0 would address the caller's own
// process group if it reached kill(2). Callers may translate this error
// into a domain-specific message at their boundary.
var ErrInvalidPID = errors.New("invalid process ID")
