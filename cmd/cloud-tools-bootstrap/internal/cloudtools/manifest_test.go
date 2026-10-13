package cloudtools

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestRepositoryManifestKeepsEightToolsAndNoValidatorArtifacts(t *testing.T) {
	f, err := os.Open("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := DecodeManifest(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Tools) != 8 {
		t.Fatalf("tools = %d; want eight shared bootstrap tools", len(m.Tools))
	}
	var got map[string]Tool
	got = make(map[string]Tool, len(m.Tools))
	for _, tool := range m.Tools {
		got[tool.Name] = tool
	}
	for name, version := range map[string]string{"go": "1.27.2", "bd": "1.3.1", "neurostrata": "1.9.0", "neuroplasticity": "2.1.0", "actionlint": "1.7.12", "govulncheck": "1.9.0", "go-licenses": "2.0.1", "gitleaks": "8.30.1"} {
		if got[name].Version != version {
			t.Errorf("%s version = %q; want %q", name, got[name].Version, version)
		}
	}
	if _, ok := got["verapdf"]; ok {
		t.Fatal("veraPDF artifact must not be part of the shared Go tool bootstrap")
	}
	if got["neurostrata"].BinarySHA256 != "9c029685bb881252b67de021d2ce5265e0acda0098532592d000a1586226166e" || got["neuroplasticity"].SHA256 != "85a476b1b13f0ddebeb9928946004ea2f55b04bee27038f71913456759db3634" {
		t.Fatal("Neuro tool binary provenance hashes changed")
	}
	for name, digest := range map[string]string{
		"bd":         "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653",
		"actionlint": "c872d6db8c6bf83a8eaa704fc93999f027d55dffbc63b8a6abdccb47df5f4cd4",
		"gitleaks":   "88f91962aa2f93ac6ab281d553b9e125f5197bbbce38f9f2437f7299c32e5509",
	} {
		if got[name].BinarySHA256 != digest {
			t.Errorf("%s installed binary digest = %q; want pinned digest", name, got[name].BinarySHA256)
		}
	}
	rootManifest, err := os.ReadFile("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	embeddedManifest, err := os.ReadFile("../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	rootDecoded, err := DecodeManifest(strings.NewReader(string(rootManifest)))
	if err != nil {
		t.Fatal(err)
	}
	embeddedDecoded, err := DecodeManifest(strings.NewReader(string(embeddedManifest)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rootDecoded, embeddedDecoded) {
		t.Fatal("embedded and repository tool manifests differ")
	}
	script, err := os.ReadFile("../../../../scripts/bootstrap-cloud-build-tools.sh")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(script), "gpg") || strings.Contains(string(script), "verapdf") || strings.Contains(string(script), "VERAPDF_") {
		t.Fatal("shared Go bootstrap must not require or provision veraPDF/GPG")
	}
	for _, check := range []struct{ key, pattern, want string }{
		{"version", `readonly go_version='([^']+)'`, got["go"].Version},
		{"checksum", `readonly go_sha256='([0-9a-f]+)'`, got["go"].SHA256},
		{"bootstrap revision", `readonly bootstrap_source_revision='([0-9a-f]{40})'`, m.BootstrapSource.Revision},
		{"bootstrap archive checksum", `readonly bootstrap_source_sha256='([0-9a-f]{64})'`, m.BootstrapSource.SHA256},
		{"bootstrap archive size", `readonly bootstrap_source_size_bytes='([0-9]+)'`, fmt.Sprint(m.BootstrapSource.SizeBytes)},
	} {
		match := regexp.MustCompile(check.pattern).FindSubmatch(script)
		if len(match) != 2 || string(match[1]) != check.want {
			t.Errorf("shell Go %s drifted from manifest: got %q, want %q", check.key, match, check.want)
		}
	}
}

func TestManifestRejectsMutableAndUnpinnedSources(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","url":"https://github.com/example/project/archive/0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size_bytes":42},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"https://example.test/tool","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checksum_source":"official","kind":"raw","binary":"tool","version_source":"asset digest"}]}`
	if _, err := DecodeManifest(stringsReader(base)); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		name string
		json string
	}{
		{name: "missing size", json: strings.Replace(base, `,"size_bytes":42`, "", 1)},
		{name: "zero size", json: strings.Replace(base, `"size_bytes":42`, `"size_bytes":0`, 1)},
		{name: "negative size", json: strings.Replace(base, `"size_bytes":42`, `"size_bytes":-1`, 1)},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			if _, err := DecodeManifest(stringsReader(invalid.json)); err == nil {
				t.Fatal("accepted bootstrap source without a positive pinned size")
			}
		})
	}
	bad := strings.Replace(base, "https://example.test/tool", "http://example.test/tool", 1)
	if _, err := DecodeManifest(stringsReader(bad)); err == nil {
		t.Fatal("accepted non-HTTPS source")
	}
	bad = strings.Replace(base, `"kind":"raw"`, `"kind":"go-install","module":"example/tool@latest"`, 1)
	if _, err := DecodeManifest(stringsReader(bad)); err == nil {
		t.Fatal("accepted an unpinned Go module")
	}
}

// stringsReader keeps the manifest tests independent from filesystem state.
func stringsReader(s string) *strings.Reader { return strings.NewReader(s) }
