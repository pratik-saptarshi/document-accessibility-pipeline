package cloudtools

import (
	"bytes"
	"fmt"
	"io"
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
	for name, version := range map[string]string{"go": "1.27.3", "bd": "1.3.1", "neurostrata": "1.9.0", "neuroplasticity": "2.1.0", "actionlint": "1.7.12", "govulncheck": "1.9.0", "go-licenses": "2.0.1", "gitleaks": "8.30.1"} {
		if got[name].Version != version {
			t.Errorf("%s version = %q; want %q", name, got[name].Version, version)
		}
	}
	if got["go"].SHA256 != "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40" {
		t.Errorf("go archive SHA256 = %q; want official Go 1.27.3 linux/amd64 checksum", got["go"].SHA256)
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
		{"bootstrap storage revision", `readonly bootstrap_source_storage_revision='([0-9a-f]{40})'`, m.BootstrapSource.StorageRevision},
		{"bootstrap source path", `readonly bootstrap_source_path='([^']+)'`, m.BootstrapSource.Path},
		{"bootstrap archive checksum", `readonly bootstrap_source_sha256='([0-9a-f]{64})'`, m.BootstrapSource.SHA256},
		{"bootstrap archive size", `readonly bootstrap_source_size_bytes='([0-9]+)'`, fmt.Sprint(m.BootstrapSource.SizeBytes)},
	} {
		match := regexp.MustCompile(check.pattern).FindSubmatch(script)
		if len(match) != 2 || string(match[1]) != check.want {
			t.Errorf("shell Go %s drifted from manifest: got %q, want %q", check.key, match, check.want)
		}
	}
	if !strings.Contains(string(script), `readonly bootstrap_source_url="https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/${bootstrap_source_storage_revision}/${bootstrap_source_path}"`) {
		t.Fatal("shell bootstrap source URL must use the pinned storage revision and exact artifact path")
	}
}

func TestManifestRejectsMutableAndUnpinnedSources(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","storage_revision":"abcdef0123456789abcdef0123456789abcdef01","path":"buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","url":"https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/abcdef0123456789abcdef0123456789abcdef01/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size_bytes":42},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"https://example.test/tool","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checksum_source":"official","kind":"raw","binary":"tool","version_source":"asset digest"}]}`
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
		{name: "missing storage revision", json: strings.Replace(base, `"storage_revision":"abcdef0123456789abcdef0123456789abcdef01",`, "", 1)},
		{name: "abbreviated storage revision", json: strings.ReplaceAll(base, "abcdef0123456789abcdef0123456789abcdef01", "abcdef01234567")},
		{name: "same content and storage revision", json: strings.Replace(base, "abcdef0123456789abcdef0123456789abcdef01", "0123456789abcdef0123456789abcdef01234567", 1)},
		{name: "mutable branch URL", json: strings.Replace(base, "/abcdef0123456789abcdef0123456789abcdef01/buildtools/", "/main/buildtools/", 1)},
		{name: "mutable tag URL", json: strings.Replace(base, "/abcdef0123456789abcdef0123456789abcdef01/buildtools/", "/v1.2.3/buildtools/", 1)},
		{name: "generated archive URL", json: strings.Replace(base, "https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/abcdef0123456789abcdef0123456789abcdef01/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/0123456789abcdef0123456789abcdef01234567.tar.gz", 1)},
		{name: "wrong repository", json: strings.Replace(base, "raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline", "raw.githubusercontent.com/example/other", 1)},
		{name: "wrong manifest path", json: strings.Replace(base, `"path":"buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz"`, `"path":"buildtools/source-archives/other.tar.gz"`, 1)},
		{name: "wrong URL path", json: strings.Replace(base, "/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "/buildtools/source-archives/other.tar.gz", 1)},
		{name: "storage revision mismatch", json: strings.Replace(base, "/abcdef0123456789abcdef0123456789abcdef01/buildtools/", "/0123456789abcdef0123456789abcdef01234567/buildtools/", 1)},
		{name: "query string", json: strings.Replace(base, `.tar.gz","sha256"`, `.tar.gz?download=1","sha256"`, 1)},
		{name: "filtered suffix with non-hex characters", json: strings.ReplaceAll(base, "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567-filtered-zzzzzzzzzzzzzzzz.tar.gz")},
		{name: "filtered suffix too short", json: strings.ReplaceAll(base, "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567-filtered-0123456789abcde.tar.gz")},
		{name: "filtered suffix too long", json: strings.ReplaceAll(base, "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567-filtered-0123456789abcdef0.tar.gz")},
		{name: "filtered path embeds different revision", json: strings.ReplaceAll(base, "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "cloud-tools-bootstrap-1111111111111111111111111111111111111111-filtered-0123456789abcdef.tar.gz")},
		{name: "filtered path missing archive extension", json: strings.ReplaceAll(base, "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz", "cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567-filtered-0123456789abcdef")},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			if _, err := DecodeManifest(stringsReader(invalid.json)); err == nil {
				t.Fatalf("accepted invalid bootstrap source: %s", invalid.name)
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

func TestManifestRequiresExactImmutableGoModuleVersion(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","storage_revision":"abcdef0123456789abcdef0123456789abcdef01","path":"buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","url":"https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/abcdef0123456789abcdef0123456789abcdef01/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size_bytes":42},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"go install example.com/tool@v1.2.3","checksum_source":"Go module checksum database","kind":"go-install","module":"example.com/tool@v1.2.3","binary":"tool","version_args":["version"],"version_pattern":"tool 1\\.2\\.3"}]}`
	if _, err := DecodeManifest(stringsReader(base)); err != nil {
		t.Fatalf("rejected exact version pin: %v", err)
	}
	for _, tc := range []struct {
		name, version, target string
	}{
		{name: "branch query", version: "main", target: "example.com/tool@vmain"},
		{name: "latest query", version: "latest", target: "example.com/tool@vlatest"},
		{name: "revision query", version: "0123456789abcdef0123456789abcdef01234567", target: "example.com/tool@v0123456789abcdef0123456789abcdef01234567"},
		{name: "longer patch version", version: "1.2.3", target: "example.com/tool@v1.2.30"},
		{name: "version suffix", version: "1.2.3", target: "example.com/tool@v1.2.3suffix"},
		{name: "leading zero version", version: "01.2.3", target: "example.com/tool@v01.2.3"},
		{name: "leading zero prerelease identifier", version: "1.2.3-01", target: "example.com/tool@v1.2.3-01"},
		{name: "malformed module path", version: "1.2.3", target: "../tool@v1.2.3"},
		{name: "reserved Windows path element", version: "1.2.3", target: "example.com/tool/CON/cmd@v1.2.3"},
		{name: "Windows short-name path element", version: "1.2.3", target: "example.com/tool/pkg~1@v1.2.3"},
		{name: "uppercase module domain", version: "1.2.3", target: "Example.com/tool@v1.2.3"},
		{name: "gopkg.in v2 missing required suffix", version: "2.0.0", target: "gopkg.in/yaml@v2.0.0"},
		{name: "gopkg.in major suffix mismatch", version: "2.0.0", target: "gopkg.in/yaml.v3@v2.0.0"},
		{name: "noncanonical build metadata", version: "1.2.3+metadata", target: "example.com/tool@v1.2.3+metadata"},
		{name: "invalid module path punctuation", version: "1.2.3", target: "example.com/foo!/tool@v1.2.3"},
		{name: "nested major suffix mismatch", version: "2.0.0", target: "example.com/tool/v3/cmd/tool@v2.0.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(base, `"version":"1.2.3"`, `"version":"`+tc.version+`"`, 1)
			mutated = strings.Replace(mutated, "example.com/tool@v1.2.3", tc.target, 2)
			if _, err := DecodeManifest(stringsReader(mutated)); err == nil {
				t.Fatalf("accepted Go install target %q for tool version %q", tc.target, tc.version)
			}
		})
	}
	for _, tc := range []struct{ version, packagePath string }{
		{version: "1.2.3-rc.1", packagePath: "example.com/tool"},
		{version: "0.0.0-20240101123456-0123456789ab", packagePath: "example.com/tool"},
		{version: "0.2.3+incompatible", packagePath: "example.com/tool"},
		{version: "1.2.3+incompatible", packagePath: "example.com/tool"},
		{version: "2.0.0+incompatible", packagePath: "example.com/tool"},
		{version: "2.0.0+incompatible", packagePath: "example.com/tool/v2"},
		{version: "1.2.3", packagePath: "example.com/tool/cmd/v1"},
		{version: "1.2.3", packagePath: "example.com/tool/cmd/g++"},
		{version: "2.0.0+incompatible", packagePath: "example.com/tool/cmd/v2"},
		{version: "2.0.0", packagePath: "example.com/tool/v2"},
		{version: "2.0.0", packagePath: "example.com/tool/v2/cmd/tool"},
		{version: "1.2.3", packagePath: "example.com/neighbor"},
		{version: "1.0.0", packagePath: "gopkg.in/yaml.v1"},
		{version: "2.0.0", packagePath: "gopkg.in/yaml.v2"},
		{version: "2.0.0", packagePath: "gopkg.in/yaml.v2-unstable"},
		{version: "2.0.0", packagePath: "gopkg.in/foo.v3/bar.v2/cmd/tool"},
	} {
		t.Run("valid "+tc.version+" "+tc.packagePath, func(t *testing.T) {
			mutated := strings.Replace(base, `"version":"1.2.3"`, `"version":"`+tc.version+`"`, 1)
			mutated = strings.Replace(mutated, "example.com/tool@v1.2.3", tc.packagePath+"@v"+tc.version, 2)
			if _, err := DecodeManifest(stringsReader(mutated)); err != nil {
				t.Fatalf("rejected immutable Go version %q: %v", tc.version, err)
			}
		})
	}
}

func TestDecodeManifestRejectsStandaloneAliasesAtTypedLevels(t *testing.T) {
	data, err := os.ReadFile("../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		from string
		to   string
	}{
		{
			name: "manifest root",
			from: `"schema_version": 1`,
			to:   `"SCHEMA_VERSION": 1`,
		},
		{
			name: "bootstrap source",
			from: `"revision": "ef4c35629c467887b6749f9e31af111cbe8f8913"`,
			to:   `"REVISION": "ef4c35629c467887b6749f9e31af111cbe8f8913"`,
		},
		{
			name: "tool entry",
			from: `"name":"go"`,
			to:   `"NAME":"go"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(string(data), tc.from, tc.to, 1)
			if mutated == string(data) {
				t.Fatalf("test mutation did not find %s", tc.from)
			}
			if _, err := DecodeManifest(stringsReader(mutated)); err == nil {
				t.Fatalf("DecodeManifest accepted noncanonical key %s", tc.to)
			}
		})
	}
}

func TestDecodeManifestRejectsExactDuplicateKeysAtTypedLevels(t *testing.T) {
	data, err := os.ReadFile("../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		from string
		to   string
	}{
		{
			name: "manifest root",
			from: `"schema_version": 1`,
			to:   `"schema_version": 1, "schema_version": 1`,
		},
		{
			name: "bootstrap source",
			from: `"revision": "ef4c35629c467887b6749f9e31af111cbe8f8913"`,
			to:   `"revision": "ef4c35629c467887b6749f9e31af111cbe8f8913", "revision": "ef4c35629c467887b6749f9e31af111cbe8f8913"`,
		},
		{
			name: "tool entry",
			from: `"name":"go"`,
			to:   `"name":"go","name":"go"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutated := strings.Replace(string(data), tc.from, tc.to, 1)
			if mutated == string(data) {
				t.Fatalf("test mutation did not find %s", tc.from)
			}
			if _, err := DecodeManifest(stringsReader(mutated)); err == nil {
				t.Fatalf("DecodeManifest accepted duplicate key %s", tc.from)
			}
		})
	}
}

func TestDecodeManifestRejectsOversizedInput(t *testing.T) {
	data, err := os.ReadFile("../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) >= maxManifestBytes {
		t.Fatalf("checked-in manifest size %d is not below test limit %d", len(data), maxManifestBytes)
	}
	data = append(data, bytes.Repeat([]byte(" "), maxManifestBytes-len(data))...)
	if _, err := DecodeManifest(bytes.NewReader(data)); err != nil {
		t.Fatalf("DecodeManifest rejected valid input exactly at the %d-byte limit: %v", maxManifestBytes, err)
	}
	limitPlusOne := append(append([]byte(nil), data...), ' ')
	_, err = DecodeManifest(bytes.NewReader(limitPlusOne))
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("exceeds %d-byte limit", maxManifestBytes)) {
		t.Fatalf("DecodeManifest error = %v, want size-specific error", err)
	}

	oversized := append(append([]byte(nil), data...), bytes.Repeat([]byte(" "), 32*1024)...)
	reader := &countingReader{reader: bytes.NewReader(oversized)}
	_, err = DecodeManifest(reader)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("exceeds %d-byte limit", maxManifestBytes)) {
		t.Fatalf("DecodeManifest error for oversized reader = %v, want size-specific error", err)
	}
	if reader.read > maxManifestBytes+1 {
		t.Fatalf("DecodeManifest consumed %d bytes, want no more than limit+1 (%d)", reader.read, maxManifestBytes+1)
	}
}

type countingReader struct {
	reader io.Reader
	read   int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read += n
	return n, err
}

// stringsReader keeps the manifest tests independent from filesystem state.
func stringsReader(s string) *strings.Reader { return strings.NewReader(s) }
