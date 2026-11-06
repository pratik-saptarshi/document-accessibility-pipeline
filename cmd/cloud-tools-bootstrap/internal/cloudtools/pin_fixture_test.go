package cloudtools

import (
	"os"
	"path/filepath"
	"testing"
)

// The fake-Go fixtures in this package must present the Go version the
// repository actually pins, so they read it from the root
// buildtools.manifest.json -- the single source of truth for the bootstrap
// pin set -- instead of embedding a copy that a re-pin would silently strand.

func pinnedGoVersion(t *testing.T) string {
	t.Helper()
	f, err := os.Open("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := DecodeManifest(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range m.Tools {
		if tool.Name == "go" {
			return tool.Version
		}
	}
	t.Fatal("root buildtools manifest has no go tool")
	return ""
}

func pinnedGoVersionOutput(t *testing.T) string {
	t.Helper()
	return "go version go" + pinnedGoVersion(t) + " linux/amd64"
}

func fakeGoToolchainBinary(t *testing.T, toolsRoot string) string {
	t.Helper()
	return filepath.Join(toolsRoot, "toolchains", "go", pinnedGoVersion(t), "bin", "go")
}
