package processgroup

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"hash"
	"io"
	"sync"
)

// ErrOutputLimitExceeded is returned by a Capture created with
// ErrorOnExceed when a write crosses the budget limit, and by every
// later write once the budget is exhausted.
var ErrOutputLimitExceeded = errors.New("output limit exceeded")

// A Budget is the shared byte allowance for one or more Captures. A
// budget serving a single Capture gives per-writer limiting; a budget
// shared by several Captures limits their combined retained output.
// The zero value is not usable; construct with NewBudget.
type Budget struct {
	mu       sync.Mutex
	limit    int64
	retained int64 // bytes retained across all captures on this budget
	seen     int64 // bytes presented across all captures on this budget
	exact    bool  // exact-limit consumption counts as exceeded
	exceeded bool
}

// A BudgetOption adjusts a Budget at construction.
type BudgetOption func(*Budget)

// ExactLimitExceeds makes a budget report exceeded when its retained
// allowance is consumed exactly, not only when output passes the limit
// (the container runner's contract: a write that exactly fills the
// remaining budget is over-limit). A default budget reports exceeded
// only past the limit.
func ExactLimitExceeds() BudgetOption {
	return func(b *Budget) { b.exact = true }
}

// NewBudget returns a Budget retaining at most limit bytes across all
// its Captures. A negative limit is treated as zero.
func NewBudget(limit int64, opts ...BudgetOption) *Budget {
	if limit < 0 {
		limit = 0
	}
	b := &Budget{limit: limit}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// consume accounts for a presented write of n bytes and returns how
// many of them may be retained under the budget.
func (b *Budget) consume(n int64) (keep int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seen += n
	if b.exact {
		if b.seen >= b.limit && b.seen > 0 {
			b.exceeded = true
		}
	} else if b.seen > b.limit {
		b.exceeded = true
	}
	room := b.limit - b.retained
	if room < 0 {
		room = 0
	}
	keep = n
	if keep > room {
		keep = room
	}
	b.retained += keep
	return keep
}

// exhausted reports whether the budget's retained allowance is full.
func (b *Budget) exhausted() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.retained >= b.limit
}

// Exceeded reports whether output presented on this budget passed the
// limit — or exactly consumed it, for an ExactLimitExceeds budget.
func (b *Budget) Exceeded() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.exceeded
}

// A Capture is a bounded io.Writer for one subprocess output stream.
// It retains up to its Budget's allowance and silently drains the rest,
// unless created with ErrorOnExceed. With no sink option the retained
// bytes accumulate in memory; with WithSink they are written through to
// the sink (and Bytes/String then report nothing — the sink owns the
// retained output). A Capture is safe for concurrent use; Captures
// sharing a Budget coordinate through the Budget.
type Capture struct {
	budget *Budget

	mu            sync.Mutex
	buf           bytes.Buffer
	sink          io.Writer
	sinkFailed    bool
	digest        hash.Hash
	total         int64 // bytes presented to this Capture
	kept          int64 // bytes retained by this Capture
	errorOnExceed bool
	frozen        bool
}

// A CaptureOption adjusts a Capture at construction.
type CaptureOption func(*Capture)

// WithSink writes retained bytes through to w instead of accumulating
// them in memory. Sink errors are swallowed into the SinkFailed flag;
// Write still reports success, because a subprocess must not be killed
// by its evidence sink.
func WithSink(w io.Writer) CaptureOption {
	return func(c *Capture) { c.sink = w }
}

// WithDigest makes the Capture keep a SHA-256 of the full presented
// stream, including bytes past the retention cap.
func WithDigest() CaptureOption {
	return func(c *Capture) { c.digest = sha256.New() }
}

// ErrorOnExceed makes a write that crosses the budget fail with a
// partial count and ErrOutputLimitExceeded, and every write after the
// budget is exhausted fail too — even an empty one. Rejected writes
// are still accounted (see Write): Total, the digest, and Exceeded
// describe the full presented stream, not only the retained prefix.
// The default is to drain: Write always consumes the whole input and
// reports success.
func ErrorOnExceed() CaptureOption {
	return func(c *Capture) { c.errorOnExceed = true }
}

// NewCapture returns a Capture retaining bytes against budget b.
func NewCapture(b *Budget, opts ...CaptureOption) *Capture {
	c := &Capture{budget: b}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Write consumes p per the Capture's budget and mode. In the default
// drain mode it always returns len(p), nil. See ErrorOnExceed for the
// erroring mode and Freeze for the post-freeze behavior. Every write
// that is not frozen is accounted — Total, the digest, and the
// budget's seen/exceeded state describe the full presented stream —
// including a write ErrorOnExceed rejects: rejection changes the
// return value, never the accounting, so a caller inspecting the
// capture cannot mistake rejected output for a complete stream.
func (c *Capture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.frozen {
		return len(p), nil
	}
	if c.errorOnExceed && c.budget.exhausted() {
		c.accountLocked(p)
		return 0, ErrOutputLimitExceeded
	}
	keep := c.accountLocked(p)
	c.retainLocked(p[:keep])
	if c.errorOnExceed && keep < int64(len(p)) {
		return int(keep), ErrOutputLimitExceeded
	}
	return len(p), nil
}

// accountLocked records p as presented: the budget accounts it
// (driving the shared seen/exceeded state), this Capture's total
// grows, and the digest covers it. It returns how many of p's bytes
// the budget retains. On an exhausted budget the retained count is
// zero, but the accounting still happens — that is the point of
// calling it on the ErrorOnExceed rejection path.
func (c *Capture) accountLocked(p []byte) int64 {
	keep := c.budget.consume(int64(len(p)))
	c.total += int64(len(p))
	if c.digest != nil {
		_, _ = c.digest.Write(p)
	}
	return keep
}

// retainLocked records q as retained output: to the sink if one is set,
// else to the in-memory buffer.
func (c *Capture) retainLocked(q []byte) {
	if len(q) == 0 {
		return
	}
	c.kept += int64(len(q))
	if c.sink != nil {
		if _, err := c.sink.Write(q); err != nil {
			c.sinkFailed = true
		}
		return
	}
	_, _ = c.buf.Write(q)
}

// Freeze stops all further mutation: later writes are drained and
// discarded and every read surface keeps its current value. It is the
// post-reap-timeout stop a caller applies when output may still arrive
// but must no longer be recorded.
func (c *Capture) Freeze() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frozen = true
}

// Bytes returns a copy of the retained output (empty for sink-backed
// Captures; the sink owns those bytes).
func (c *Capture) Bytes() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]byte(nil), c.buf.Bytes()...)
}

// String returns the retained output as a string.
func (c *Capture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

// Total returns the number of bytes presented to this Capture,
// including bytes drained past the cap.
func (c *Capture) Total() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total
}

// Truncated reports whether this Capture dropped bytes: its retained
// count is below its presented count.
func (c *Capture) Truncated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.kept < c.total
}

// Exceeded reports the Budget's exceeded state (shared by all Captures
// on the budget).
func (c *Capture) Exceeded() bool {
	return c.budget.Exceeded()
}

// SinkFailed reports whether any write to the sink has failed.
func (c *Capture) SinkFailed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sinkFailed
}

// Digest returns the SHA-256 of the full presented stream when the
// Capture was created WithDigest, and nil otherwise.
func (c *Capture) Digest() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.digest == nil {
		return nil
	}
	return c.digest.Sum(nil)
}

var _ io.Writer = (*Capture)(nil)
