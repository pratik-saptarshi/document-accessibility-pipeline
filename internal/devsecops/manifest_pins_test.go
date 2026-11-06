package devsecops

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the devsecops test package's single reader of the root
// buildtools manifest. Pin assertions in this package compare surfaces
// (workflows, scripts, archives) against values decoded here, so a re-pin
// edits the manifest alone instead of every test that restated the value.

type testManifestTool struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Source       string `json:"source"`
	BinarySHA256 string `json:"binary_sha256"`
}

type testBootstrapSource struct {
	Revision        string `json:"revision"`
	StorageRevision string `json:"storage_revision"`
	Path            string `json:"path"`
	URL             string `json:"url"`
	SHA256          string `json:"sha256"`
	SizeBytes       int64  `json:"size_bytes"`
}

type testBuildtoolsManifest struct {
	Tools           []testManifestTool  `json:"tools"`
	BootstrapSource testBootstrapSource `json:"bootstrap_source"`
}

func (m testBuildtoolsManifest) tool(name string) (testManifestTool, bool) {
	for _, tool := range m.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return testManifestTool{}, false
}

func decodeBuildtoolsManifest(raw []byte) (testBuildtoolsManifest, error) {
	var manifest testBuildtoolsManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return testBuildtoolsManifest{}, fmt.Errorf("decode repository tool manifest: %w", err)
	}
	return manifest, nil
}

func loadBuildtoolsManifest(t *testing.T) testBuildtoolsManifest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "buildtools.manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeBuildtoolsManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func manifestTool(t *testing.T, name string) testManifestTool {
	t.Helper()
	tool, ok := loadBuildtoolsManifest(t).tool(name)
	if !ok {
		t.Fatalf("buildtools manifest has no %q tool entry", name)
	}
	return tool
}

// testManifest is decoded at package initialization so package-level pin
// fixtures (the archive path/size/digest vars and the go-checks step
// inventory) can derive from it. A malformed manifest fails every test in
// the package at startup, which is the intended fail-closed behavior.
var testManifest = mustLoadBuildtoolsManifest()

func mustLoadBuildtoolsManifest() testBuildtoolsManifest {
	raw, err := os.ReadFile(filepath.Join("..", "..", "buildtools.manifest.json"))
	if err != nil {
		panic(err)
	}
	manifest, err := decodeBuildtoolsManifest(raw)
	if err != nil {
		panic(err)
	}
	return manifest
}

func mustManifestTool(name string) testManifestTool {
	tool, ok := testManifest.tool(name)
	if !ok {
		panic("buildtools manifest has no " + name + " tool entry")
	}
	return tool
}

// testPinnedBDDigest is the manifest's bd binary digest, for pin fixtures
// that must state the digest inside a larger expected string.
var testPinnedBDDigest = mustManifestTool("bd").BinarySHA256

// firstURLIn returns the first whitespace-delimited https URL in text, or
// "" if there is none. It is the small extractor the workflow-URL
// assertions use instead of restating a URL literal.
func firstURLIn(text string) string {
	for _, field := range strings.Fields(text) {
		if strings.HasPrefix(field, "https://") {
			return strings.TrimSuffix(field, "\\")
		}
	}
	return ""
}
