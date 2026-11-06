// Package modulepin is the single home of the build-info pin checks that
// were previously implemented per consumer (the go-afp-line pin test and
// the pdfcpu version resolution in financialfixtures): an exact-version
// requirement with outright rejection of replaced modules. It is a leaf
// package (stdlib only) so both test-only and production callers can use
// it without import cycles.
package modulepin

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// CheckPinned enforces an exact pin against one build-info dependency
// entry: the version must equal wantVersion, and a replacement is
// rejected outright. debug.Module.Version keeps the originally required
// version while Replace records the source actually compiled, so a
// replaced module could satisfy the version check with arbitrary code.
func CheckPinned(dep *debug.Module, wantVersion string) error {
	if dep.Replace != nil {
		return fmt.Errorf("%s is replaced by %s %s: the pin gate requires the unmodified reviewed module", dep.Path, dep.Replace.Path, dep.Replace.Version)
	}
	if dep.Version != wantVersion {
		return fmt.Errorf("%s version = %q, want exact pinned %q", dep.Path, dep.Version, wantVersion)
	}
	return nil
}

// Version returns the resolved version of the dependency at path in the
// build info, rejecting a replaced entry and an entry with an empty
// version. ok reports whether build info was available at all, in the
// shape debug.ReadBuildInfo returns it.
func Version(info *debug.BuildInfo, ok bool, path string) (string, error) {
	if !ok || info == nil {
		return "", errors.New("build info is unavailable")
	}
	for _, dep := range info.Deps {
		if dep.Path != path {
			continue
		}
		if dep.Replace != nil {
			return "", fmt.Errorf("%s is replaced by %s %s: the pin gate requires the unmodified reviewed module", path, dep.Replace.Path, dep.Replace.Version)
		}
		if dep.Version == "" {
			return "", fmt.Errorf("%s build-info entry has an empty version", path)
		}
		return dep.Version, nil
	}
	return "", fmt.Errorf("%s is not in the build dependency list", path)
}
