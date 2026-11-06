package cloudtools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
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
	// Tool names are identity, not pin values: the shared bootstrap set is
	// exactly these eight tools. Versions and digests are deliberately NOT
	// asserted here -- the root manifest is their single source of truth, and
	// the surface checks below fail only when a *partial* change leaves one
	// surface disagreeing with it.
	for _, name := range []string{"actionlint", "bd", "gitleaks", "go", "go-licenses", "govulncheck", "neuroplasticity", "neurostrata"} {
		if _, ok := got[name]; !ok {
			t.Errorf("shared bootstrap tool %q missing from manifest", name)
		}
	}
	if _, ok := got["verapdf"]; ok {
		t.Fatal("veraPDF artifact must not be part of the shared Go tool bootstrap")
	}
	hexDigest := regexp.MustCompile(`^[0-9a-f]{64}$`)
	for _, tool := range m.Tools {
		if tool.Version == "" {
			t.Errorf("%s version is empty", tool.Name)
		}
		if tool.License == "" || tool.Kind == "" || tool.Binary == "" {
			t.Errorf("%s is missing license, kind, or binary", tool.Name)
		}
		if tool.Source == "" || !strings.Contains(tool.Source, tool.Version) {
			t.Errorf("%s source %q does not pin its own version %q", tool.Name, tool.Source, tool.Version)
		}
		switch tool.Kind {
		case "archive", "raw", "bootstrap-toolchain":
			if !hexDigest.MatchString(tool.SHA256) {
				t.Errorf("%s %s SHA256 = %q; want 64 lowercase hex digits", tool.Name, tool.Kind, tool.SHA256)
			}
		}
		if tool.SHA256 != "" && !hexDigest.MatchString(tool.SHA256) {
			t.Errorf("%s SHA256 = %q; want 64 lowercase hex digits", tool.Name, tool.SHA256)
		}
		if tool.BinarySHA256 != "" && !hexDigest.MatchString(tool.BinarySHA256) {
			t.Errorf("%s binary SHA256 = %q; want 64 lowercase hex digits", tool.Name, tool.BinarySHA256)
		}
		if tool.VersionSource == "" {
			if tool.VersionPattern == "" {
				t.Errorf("%s has neither a version pattern nor a version source", tool.Name)
			} else if _, err := regexp.Compile(tool.VersionPattern); err != nil {
				t.Errorf("%s version pattern does not compile: %v", tool.Name, err)
			}
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

	// The source-archives README's "Current bootstrap pin" section is a
	// checked surface: every value it states must agree with the manifest's
	// bootstrap_source fields. The historical sections below it are records,
	// not surfaces, and are deliberately not extracted.
	readme, err := os.ReadFile("../../../../buildtools/source-archives/README.md")
	if err != nil {
		t.Fatal(err)
	}
	currentPin := string(readme)
	if start := strings.Index(currentPin, "## Current bootstrap pin"); start >= 0 {
		currentPin = currentPin[start:]
	} else {
		t.Fatal("source-archives README has no Current bootstrap pin section")
	}
	if end := strings.Index(currentPin, "## Previous bootstrap source"); end >= 0 {
		currentPin = currentPin[:end]
	}
	for _, check := range []struct{ key, pattern, want string }{
		{"contained source revision", "Contained source revision: `([0-9a-f]{40})`", m.BootstrapSource.Revision},
		{"filtered artifact filename", "Filtered artifact \\(the active no-checkout pin\\):\\s*`([^`]+)`", path.Base(m.BootstrapSource.Path)},
		{"protected-main storage commit", "Protected-main storage commit: `([0-9a-f]{40})`", m.BootstrapSource.StorageRevision},
	} {
		match := regexp.MustCompile(check.pattern).FindStringSubmatch(currentPin)
		if len(match) != 2 || match[1] != check.want {
			t.Errorf("README current-pin %s drifted from manifest: got %q, want %q", check.key, match, check.want)
		}
	}
	sizeAndSHA := regexp.MustCompile("Filtered size and SHA-256: ([0-9,]+) bytes;\\s*`([0-9a-f]{64})`").FindStringSubmatch(currentPin)
	if len(sizeAndSHA) != 3 || strings.ReplaceAll(sizeAndSHA[1], ",", "") != fmt.Sprint(m.BootstrapSource.SizeBytes) || sizeAndSHA[2] != m.BootstrapSource.SHA256 {
		t.Errorf("README current-pin size/SHA-256 drifted from manifest: got %q, want size %d sha256 %q", sizeAndSHA, m.BootstrapSource.SizeBytes, m.BootstrapSource.SHA256)
	}

	// Every module's effective Go version directive (toolchain line when
	// present, otherwise the go line) must equal the manifest's pinned Go
	// tool version.
	for _, goModPath := range []string{
		"../../../../go.mod",
		"../../../../cmd/cloud-tools-bootstrap/go.mod",
		"../../../../ci/pdfua2/adapter/go.mod",
	} {
		goMod, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatal(err)
		}
		directive := regexp.MustCompile(`(?m)^toolchain go([0-9]+\.[0-9]+\.[0-9]+)$`).FindSubmatch(goMod)
		if directive == nil {
			directive = regexp.MustCompile(`(?m)^go ([0-9]+\.[0-9]+\.[0-9]+)$`).FindSubmatch(goMod)
		}
		if directive == nil || string(directive[1]) != got["go"].Version {
			t.Errorf("%s Go version directive = %q; want manifest Go pin %q", goModPath, directive, got["go"].Version)
		}
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

func TestManifestRejectsTraversalVersionsAndAcceptsPinnedVersions(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","storage_revision":"abcdef0123456789abcdef0123456789abcdef01","path":"buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","url":"https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/abcdef0123456789abcdef0123456789abcdef01/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size_bytes":42},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"https://example.test/tool","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checksum_source":"official","kind":"raw","binary":"tool","version_source":"asset digest"}]}`
	withVersion := func(version string) string {
		return strings.Replace(base, `"version":"1.2.3"`, `"version":`+fmt.Sprintf("%q", version), 1)
	}
	// bootstrap.go joins Name-Version-filename into a cache path, so a
	// Version carrying separators or traversal must never validate.
	for _, version := range []string{
		"../../../escape",
		"9.9.9/../../x",
		"a/b",
		"..",
		".",
		"../x",
		"x/..",
		"1..2",
		"/abs/1.2.3",
		`a\b`,
		`1.2.3\..\..\x`,
		`..\escape`,
		`C:\escape`,
	} {
		t.Run("reject "+version, func(t *testing.T) {
			_, err := DecodeManifest(stringsReader(withVersion(version)))
			if err == nil {
				t.Fatalf("accepted traversal version %q", version)
			}
			// Exact-message assertion, with no substring fallback: a
			// rejection for any other reason (or a generic message that
			// any invalid entry could produce) does not satisfy this
			// test, so narrowing validToolVersion to a subset of these
			// shapes cannot hide behind another check's error.
			want := fmt.Sprintf("tool %q version %q must not contain path separators or traversal components", "tool", version)
			if err.Error() != want {
				t.Fatalf("rejection of version %q = %q; want exactly %q", version, err.Error(), want)
			}
		})
	}
	// Accept-cases: every version pinned in buildtools.manifest.json (read
	// from the manifest itself, so a re-pin extends this coverage instead of
	// silently losing it), plus the canonical Go module version shapes the
	// go-install checks accept.
	manifestFile, err := os.Open("../../../../buildtools.manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	defer manifestFile.Close()
	repositoryManifest, err := DecodeManifest(manifestFile)
	if err != nil {
		t.Fatal(err)
	}
	acceptVersions := []string{"1.2.3-rc.1", "1.2.3+incompatible", "0.0.0-20240101123456-0123456789ab"}
	seenVersion := make(map[string]bool)
	for _, tool := range repositoryManifest.Tools {
		if !seenVersion[tool.Version] {
			seenVersion[tool.Version] = true
			acceptVersions = append(acceptVersions, tool.Version)
		}
	}
	for _, version := range acceptVersions {
		t.Run("accept "+version, func(t *testing.T) {
			if _, err := DecodeManifest(stringsReader(withVersion(version))); err != nil {
				t.Fatalf("rejected legitimate version %q: %v", version, err)
			}
		})
	}
}

// The version gate is kind-independent, so traversal versions are rejected
// for every tool kind with the same exact message. There is deliberately no
// accepted separator case for non-archive kinds: verifiedAsset embeds
// Version in the cache path it builds (Name-Version-filename under the
// downloads directory) for both archive and raw tools, go-install versions
// are pinned into the module query, and the bootstrap-toolchain version is
// matched against the verified Go toolchain — no kind treats Version as a
// free-form label. Symmetric coverage therefore means every kind rejects
// both separator shapes, including a raw backslash such as `9.9.9\vendor`.
func TestManifestRejectsTraversalVersionsForEveryToolKind(t *testing.T) {
	base := `{"schema_version":1,"platform":"linux/amd64","bootstrap_source":{"revision":"0123456789abcdef0123456789abcdef01234567","storage_revision":"abcdef0123456789abcdef0123456789abcdef01","path":"buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","url":"https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/abcdef0123456789abcdef0123456789abcdef01/buildtools/source-archives/cloud-tools-bootstrap-0123456789abcdef0123456789abcdef01234567.tar.gz","sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","size_bytes":42},"tools":[{"name":"tool","version":"1.2.3","license":"MIT","source":"https://example.test/tool","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","checksum_source":"official","kind":"raw","binary":"tool","version_source":"asset digest"}]}`
	for _, kind := range []string{"archive", "raw", "go-install", "bootstrap-toolchain"} {
		for _, version := range []string{"9.9.9/vendor", `9.9.9\vendor`, `9.9.9\build\tool`, "../escape", `..\escape`} {
			t.Run(kind+" rejects "+version, func(t *testing.T) {
				manifest := strings.Replace(base, `"kind":"raw"`, `"kind":`+fmt.Sprintf("%q", kind), 1)
				manifest = strings.Replace(manifest, `"version":"1.2.3"`, `"version":`+fmt.Sprintf("%q", version), 1)
				_, err := DecodeManifest(stringsReader(manifest))
				if err == nil {
					t.Fatalf("kind %q accepted traversal version %q", kind, version)
				}
				want := fmt.Sprintf("tool %q version %q must not contain path separators or traversal components", "tool", version)
				if err.Error() != want {
					t.Fatalf("rejection of kind %q version %q = %q; want exactly %q", kind, version, err.Error(), want)
				}
			})
		}
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
	embedded, err := DecodeManifest(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	schemaSnippet := fmt.Sprintf(`"schema_version": %d`, embedded.SchemaVersion)
	revisionSnippet := fmt.Sprintf(`"revision": %q`, embedded.BootstrapSource.Revision)
	for _, tc := range []struct {
		name string
		from string
		to   string
	}{
		{
			name: "manifest root",
			from: schemaSnippet,
			to:   fmt.Sprintf(`"SCHEMA_VERSION": %d`, embedded.SchemaVersion),
		},
		{
			name: "bootstrap source",
			from: revisionSnippet,
			to:   fmt.Sprintf(`"REVISION": %q`, embedded.BootstrapSource.Revision),
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
	embedded, err := DecodeManifest(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	schemaSnippet := fmt.Sprintf(`"schema_version": %d`, embedded.SchemaVersion)
	revisionSnippet := fmt.Sprintf(`"revision": %q`, embedded.BootstrapSource.Revision)
	for _, tc := range []struct {
		name string
		from string
		to   string
	}{
		{
			name: "manifest root",
			from: schemaSnippet,
			to:   schemaSnippet + ", " + schemaSnippet,
		},
		{
			name: "bootstrap source",
			from: revisionSnippet,
			to:   revisionSnippet + ", " + revisionSnippet,
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
