package beadsreconcile

import (
	"encoding/json"
	"os"
	"testing"
)

// TestPinnedBDMatchesBuildtoolsManifest is the cross-check that keeps the
// code-side bd pin (PinnedBDVersion / PinnedBDSHA256, the constants the
// null-loss and closure gates compare against) and the bootstrap manifest's
// bd tool entry from drifting apart: a re-pin must move both homes, and
// this test fails naming the disagreement if only one moved. It reads both
// surfaces; it carries no pin literal of its own.
func TestPinnedBDMatchesBuildtoolsManifest(t *testing.T) {
	raw, err := os.ReadFile("../../buildtools.manifest.json")
	if err != nil {
		t.Fatalf("read buildtools manifest: %v", err)
	}
	var manifest struct {
		Tools []struct {
			Name         string `json:"name"`
			Version      string `json:"version"`
			BinarySHA256 string `json:"binary_sha256"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("decode buildtools manifest: %v", err)
	}
	for _, tool := range manifest.Tools {
		if tool.Name != "bd" {
			continue
		}
		if tool.Version != PinnedBDVersion {
			t.Errorf("manifest bd version %q != PinnedBDVersion %q: re-pin both homes together", tool.Version, PinnedBDVersion)
		}
		if tool.BinarySHA256 != PinnedBDSHA256 {
			t.Errorf("manifest bd binary_sha256 %q != PinnedBDSHA256 %q: re-pin both homes together", tool.BinarySHA256, PinnedBDSHA256)
		}
		return
	}
	t.Fatal("buildtools manifest has no bd tool entry")
}
