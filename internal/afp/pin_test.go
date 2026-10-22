// Package afp is the DAP-side anchor for the adopted go-afp-line dependency
// (Decision 0022, dap-jo9.4.2). Phase 1 pins the exact reviewed revision;
// later progressive-enhancement phases wire the AFP integration (see
// docs/process/plans/afp-adoption-progressive-phases.md).
package afp

import (
	"runtime/debug"
	"strings"
	"testing"

	goafp "github.com/pratik-saptarshi/go-afp-line/afp"
)

// TestPinnedRevision is the verification gate for Decision 0022 condition 1:
// the adopted go-afp-line revision must be exactly the reviewed commit
// 492c4c7 (never the beta). The import also anchors the pin so
// `go mod tidy` keeps the require directive until Phase 3 wires the
// integration.
func TestPinnedRevision(t *testing.T) {
	// Compile-time anchor: DecodeError is the typed error Phase 3 will
	// propagate for 0x80-class aborts.
	var _ = goafp.DecodeError{}

	const wantCommitSuffix = "-492c4c7d35b9"
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("debug.ReadBuildInfo unavailable")
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/pratik-saptarshi/go-afp-line" {
			if !strings.HasSuffix(d.Version, wantCommitSuffix) {
				t.Fatalf("go-afp-line version = %q, want suffix %q", d.Version, wantCommitSuffix)
			}
			return
		}
	}
	t.Fatal("github.com/pratik-saptarshi/go-afp-line not found in build list")
}
