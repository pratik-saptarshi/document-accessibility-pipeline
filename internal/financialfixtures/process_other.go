//go:build !linux

package financialfixtures

import (
	"context"
	"errors"
	"io"
	"time"
)

// runV2BoundedProcess fails closed off Linux because this helper's cancellation
// guarantee requires process-group termination; direct-process cancellation
// is not an equivalent fallback.
func runV2BoundedProcess(_ context.Context, _ string, _ []string, _ []string, _ io.Writer, _ io.Writer, _ time.Duration) error {
	return errors.New("financial fixture helpers require Linux process-group cleanup")
}
