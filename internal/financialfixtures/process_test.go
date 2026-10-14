package financialfixtures

import (
	"runtime"
	"testing"
)

func requireV2ProcessGroupSupport(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("financial fixture helper process-tree cleanup is supported only on Linux")
	}
}
