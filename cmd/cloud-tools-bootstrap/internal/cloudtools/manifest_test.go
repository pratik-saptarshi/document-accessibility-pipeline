package cloudtools

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestRepositoryManifestIsPinnedAndKeepsVeraPDFCIOnly(t *testing.T) {
	f, err := os.Open("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := DecodeManifest(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Tools) != 9 {
		t.Fatalf("tools = %d; want nine entries, including Go and CI-only veraPDF", len(m.Tools))
	}
	var got map[string]Tool
	got = make(map[string]Tool, len(m.Tools))
	for _, tool := range m.Tools {
		got[tool.Name] = tool
	}
	for name, version := range map[string]string{"go": "1.27.2", "bd": "1.3.1", "neurostrata": "1.9.0", "neuroplasticity": "2.1.0", "actionlint": "1.7.12", "govulncheck": "1.9.0", "go-licenses": "2.0.1", "gitleaks": "8.30.1", "verapdf": "1.30.3"} {
		if got[name].Version != version {
			t.Errorf("%s version = %q; want %q", name, got[name].Version, version)
		}
	}
	if got["verapdf"].Kind != "ci-only-external" || got["verapdf"].Scope != "CI-only; never install into application/runtime image" || got["verapdf"].SignerFingerprint != "13DD102B4DD69354D12DE5A83184863278B17FE7" {
		t.Fatalf("veraPDF boundary weakened: %#v", got["verapdf"])
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
	for _, check := range []struct{ key, pattern, want string }{
		{"version", `readonly go_version='([^']+)'`, got["go"].Version},
		{"checksum", `readonly go_sha256='([0-9a-f]+)'`, got["go"].SHA256},
		{"bootstrap revision", `readonly bootstrap_source_revision='([0-9a-f]{40})'`, m.BootstrapSource.Revision},
		{"bootstrap archive checksum", `readonly bootstrap_source_sha256='([0-9a-f]{64})'`, m.BootstrapSource.SHA256},
	} {
		match := regexp.MustCompile(check.pattern).FindSubmatch(script)
		if len(match) != 2 || string(match[1]) != check.want {
			t.Errorf("shell Go %s drifted from manifest: got %q, want %q", check.key, match, check.want)
		}
	}
}

func TestManifestRejectsMutableAndUnpinnedSources(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","url":"https://github.com/example/project/archive/0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"https://example.test/tool","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checksum_source":"official","kind":"raw","binary":"tool","version_source":"asset digest"}]}`
	if _, err := DecodeManifest(stringsReader(base)); err != nil {
		t.Fatal(err)
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
