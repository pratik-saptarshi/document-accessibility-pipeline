// Package afp is the DAP-side anchor for the adopted go-afp-line dependency
// (Decision 0022, dap-jo9.4.2). Phase 1 pins the exact reviewed revision;
// later progressive-enhancement phases wire the AFP integration (see
// docs/process/plans/afp-adoption-progressive-phases.md).
package afp

import (
	"fmt"
	"runtime/debug"
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

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("debug.ReadBuildInfo unavailable")
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/pratik-saptarshi/go-afp-line" {
			if err := verifyPinnedDep(d); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatal("github.com/pratik-saptarshi/go-afp-line not found in build list")
}

// pinnedModuleVersion is the exact reviewed go-afp-line pseudo-version
// pinned by Decision 0022 condition 1 (commit 492c4c7, never the beta).
const pinnedModuleVersion = "v0.1.0-beta.1.0.20261022163835-492c4c7d35b9"

// verifyPinnedDep enforces the pin against one build-info dependency
// entry. The complete pseudo-version must match: a version that merely
// ends in the same commit fragment (for example a prerelease tag) does
// not prove the reviewed source is in use. A non-nil replacement is
// rejected outright: debug.Module.Version keeps the originally required
// version while Replace records the source actually compiled, so a
// replaced module could satisfy the version check with arbitrary code.
func verifyPinnedDep(dep *debug.Module) error {
	if dep.Replace != nil {
		return fmt.Errorf("go-afp-line replaced by %s %s: the pin gate requires the unmodified reviewed module", dep.Replace.Path, dep.Replace.Version)
	}
	if dep.Version != pinnedModuleVersion {
		return fmt.Errorf("go-afp-line version = %q, want exact pinned %q", dep.Version, pinnedModuleVersion)
	}
	return nil
}

// TestVerifyPinnedDep exercises the pin gate against synthetic build-info
// entries: the gate must accept only the exact reviewed pseudo-version
// with no module replacement (Decision 0022 condition 1; PR #187 review).
func TestVerifyPinnedDep(t *testing.T) {
	const modulePath = "github.com/pratik-saptarshi/go-afp-line"
	tests := []struct {
		name    string
		dep     *debug.Module
		wantErr bool
	}{
		{
			name: "exact pinned revision, no replacement",
			dep:  &debug.Module{Path: modulePath, Version: pinnedModuleVersion},
		},
		{
			name:    "suffix-matching tag is not the pin",
			dep:     &debug.Module{Path: modulePath, Version: "v0.1.0-492c4c7d35b9"},
			wantErr: true,
		},
		{
			name:    "beta release is not the pin",
			dep:     &debug.Module{Path: modulePath, Version: "v0.1.0-beta.1"},
			wantErr: true,
		},
		{
			name: "module replacement hides the compiled source",
			dep: &debug.Module{
				Path:    modulePath,
				Version: pinnedModuleVersion,
				Replace: &debug.Module{Path: modulePath, Version: "v0.0.0-20260101000000-abcdefabcdef"},
			},
			wantErr: true,
		},
		{
			name: "local path replacement hides the compiled source",
			dep: &debug.Module{
				Path:    modulePath,
				Version: pinnedModuleVersion,
				Replace: &debug.Module{Path: "../go-afp-line"},
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyPinnedDep(tc.dep)
			if tc.wantErr && err == nil {
				t.Fatalf("verifyPinnedDep(%+v) = nil, want error", tc.dep)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("verifyPinnedDep(%+v) = %v, want nil", tc.dep, err)
			}
		})
	}
}
