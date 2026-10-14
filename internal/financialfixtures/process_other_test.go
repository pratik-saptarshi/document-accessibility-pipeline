//go:build !linux

package financialfixtures

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestV2BoundedProcessFailsClosedOffLinux(t *testing.T) {
	err := runV2BoundedProcess(context.Background(), "must-not-run", nil, nil, nil, nil, 250*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "require Linux process-group cleanup") {
		t.Fatalf("unsupported-platform error = %v, want fail-closed Linux boundary", err)
	}
}
