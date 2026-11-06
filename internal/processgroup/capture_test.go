package processgroup

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"testing"
)

func TestCaptureDrainsPastLimit(t *testing.T) {
	c := NewCapture(NewBudget(8))
	n, err := c.Write([]byte("0123456789"))
	if n != 10 || err != nil {
		t.Fatalf("Write = %d, %v; want 10, nil (drain mode never errors)", n, err)
	}
	if got := c.String(); got != "01234567" {
		t.Fatalf("String = %q, want the retained prefix %q", got, "01234567")
	}
	if got := c.Total(); got != 10 {
		t.Fatalf("Total = %d, want 10 (all bytes presented)", got)
	}
	if !c.Truncated() {
		t.Fatal("Truncated = false after writing past the limit, want true")
	}
	if !c.Exceeded() {
		t.Fatal("Exceeded = false after writing past the limit, want true")
	}
}

func TestCaptureExactLimitIsNotTruncated(t *testing.T) {
	c := NewCapture(NewBudget(4))
	if _, err := c.Write([]byte("abcd")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if c.Truncated() {
		t.Fatal("Truncated = true at exactly the limit, want false (total > limit is the rule)")
	}
	if c.Exceeded() {
		t.Fatal("Exceeded = true at exactly the limit on a default budget, want false")
	}
	if _, err := c.Write([]byte("e")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !c.Truncated() || !c.Exceeded() {
		t.Fatal("one byte past the limit must report Truncated and Exceeded")
	}
	if got := c.String(); got != "abcd" {
		t.Fatalf("String = %q, want %q", got, "abcd")
	}
}

func TestCaptureSharedBudgetExhaustsJointly(t *testing.T) {
	b := NewBudget(6)
	c1 := NewCapture(b)
	c2 := NewCapture(b)
	if _, err := c1.Write([]byte("abcd")); err != nil {
		t.Fatalf("c1 Write: %v", err)
	}
	if _, err := c2.Write([]byte("wxyz")); err != nil {
		t.Fatalf("c2 Write: %v", err)
	}
	if got := c1.String(); got != "abcd" {
		t.Fatalf("c1 String = %q, want %q", got, "abcd")
	}
	if got := c2.String(); got != "wx" {
		t.Fatalf("c2 String = %q, want %q (the shared budget had 2 bytes left)", got, "wx")
	}
	if !c2.Truncated() {
		t.Fatal("c2 Truncated = false, want true (it dropped bytes)")
	}
	if c1.Truncated() {
		t.Fatal("c1 Truncated = true, want false (it dropped nothing)")
	}
	if !c1.Exceeded() || !c2.Exceeded() {
		t.Fatal("Exceeded is a budget-level fact; both captures must see it")
	}
}

func TestCaptureDigestCoversFullStream(t *testing.T) {
	data := []byte("0123456789")
	c := NewCapture(NewBudget(4), WithDigest())
	if _, err := c.Write(data); err != nil {
		t.Fatalf("Write: %v", err)
	}
	want := sha256.Sum256(data)
	if got := c.Digest(); !bytes.Equal(got, want[:]) {
		t.Fatalf("Digest = %x, want sha256 of the FULL stream %x (including bytes past the cap)", got, want)
	}
	if got := c.String(); got != "0123" {
		t.Fatalf("String = %q, want the retained prefix %q", got, "0123")
	}
}

func TestCaptureWithoutDigestHasNoDigest(t *testing.T) {
	c := NewCapture(NewBudget(4))
	if _, err := c.Write([]byte("ab")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := c.Digest(); got != nil {
		t.Fatalf("Digest = %x, want nil without WithDigest", got)
	}
}

type failingSink struct{}

func (failingSink) Write([]byte) (int, error) { return 0, errors.New("sink unavailable") }

func TestCaptureSinkFailureIsFlaggedNotFatal(t *testing.T) {
	c := NewCapture(NewBudget(8), WithSink(failingSink{}))
	n, err := c.Write([]byte("abc"))
	if n != 3 || err != nil {
		t.Fatalf("Write = %d, %v; want 3, nil (sink errors are swallowed into the flag)", n, err)
	}
	if !c.SinkFailed() {
		t.Fatal("SinkFailed = false after a failing sink write, want true")
	}
}

func TestCaptureSinkReceivesRetainedPrefix(t *testing.T) {
	var sink bytes.Buffer
	c := NewCapture(NewBudget(3), WithSink(&sink))
	if _, err := c.Write([]byte("abcdef")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := sink.String(); got != "abc" {
		t.Fatalf("sink received %q, want the retained prefix %q", got, "abc")
	}
	if c.SinkFailed() {
		t.Fatal("SinkFailed = true with a healthy sink, want false")
	}
}

func TestBudgetExactLimitExceeds(t *testing.T) {
	exact := NewCapture(NewBudget(4, ExactLimitExceeds()))
	if _, err := exact.Write([]byte("abcd")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !exact.Exceeded() {
		t.Fatal("Exceeded = false after exactly consuming an ExactLimitExceeds budget, want true")
	}
	def := NewCapture(NewBudget(4))
	if _, err := def.Write([]byte("abcd")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if def.Exceeded() {
		t.Fatal("Exceeded = true after exactly consuming a default budget, want false")
	}
}

func TestCaptureErrorOnExceed(t *testing.T) {
	c := NewCapture(NewBudget(5), ErrorOnExceed())
	if n, err := c.Write([]byte("abc")); n != 3 || err != nil {
		t.Fatalf("Write = %d, %v; want 3, nil", n, err)
	}
	n, err := c.Write([]byte("defg"))
	if n != 2 || !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("crossing Write = %d, %v; want partial count 2 and ErrOutputLimitExceeded", n, err)
	}
	if got := c.String(); got != "abcde" {
		t.Fatalf("String = %q, want %q (the fitting prefix is retained)", got, "abcde")
	}
	if n, err := c.Write([]byte("x")); n != 0 || !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("Write once exhausted = %d, %v; want 0, ErrOutputLimitExceeded", n, err)
	}
	if n, err := c.Write(nil); n != 0 || !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("empty Write once exhausted = %d, %v; want 0, ErrOutputLimitExceeded", n, err)
	}
}

func TestCaptureFreezeStopsMutation(t *testing.T) {
	c := NewCapture(NewBudget(16))
	if _, err := c.Write([]byte("abc")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	c.Freeze()
	n, err := c.Write([]byte("defgh"))
	if n != 5 || err != nil {
		t.Fatalf("Write after Freeze = %d, %v; want 5, nil (drained and discarded)", n, err)
	}
	if got := c.String(); got != "abc" {
		t.Fatalf("String after Freeze = %q, want %q", got, "abc")
	}
	if got := c.Total(); got != 3 {
		t.Fatalf("Total after Freeze = %d, want 3 (frozen writes change no state)", got)
	}
}

func TestCaptureBytesReturnsACopy(t *testing.T) {
	c := NewCapture(NewBudget(8))
	if _, err := c.Write([]byte("abc")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := c.Bytes()
	got[0] = 'X'
	if c.String() != "abc" {
		t.Fatal("mutating the Bytes result mutated the Capture; want a copy")
	}
	var _ io.Writer = c
}
