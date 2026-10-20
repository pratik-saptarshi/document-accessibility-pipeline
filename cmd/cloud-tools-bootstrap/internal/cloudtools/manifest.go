package cloudtools

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
)

type Manifest struct {
	SchemaVersion   int             `json:"schema_version"`
	Platform        string          `json:"platform"`
	BootstrapSource BootstrapSource `json:"bootstrap_source"`
	Tools           []Tool          `json:"tools"`
}

type BootstrapSource struct {
	Revision        string `json:"revision"`
	StorageRevision string `json:"storage_revision"`
	Path            string `json:"path"`
	URL             string `json:"url"`
	SHA256          string `json:"sha256"`
	SizeBytes       int64  `json:"size_bytes"`
}

type Tool struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	License         string   `json:"license"`
	Source          string   `json:"source"`
	SHA256          string   `json:"sha256"`
	BinarySHA256    string   `json:"binary_sha256"`
	ChecksumSource  string   `json:"checksum_source"`
	Kind            string   `json:"kind"`
	ArchiveMember   string   `json:"archive_member"`
	Module          string   `json:"module"`
	Binary          string   `json:"binary"`
	VersionArgs     []string `json:"version_args"`
	VersionPattern  string   `json:"version_pattern"`
	VersionSource   string   `json:"version_source"`
	VersionFallback string   `json:"version_fallback"`
}

const sourceRepoPath = "pratik-saptarshi/document-accessibility-pipeline"

var (
	safeName  = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	commitSHA = regexp.MustCompile(`^[a-f0-9]{40}$`)
	// Go module queries require canonical semantic versions. Go's module
	// query syntax preserves only +incompatible as build metadata.
	canonicalGoVersion = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+incompatible)?$`)
	goModulePath       = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]*(/[A-Za-z0-9][A-Za-z0-9._~-]*)*$`)
	// The denylist-filtered bootstrap archive naming documented in
	// buildtools/source-archives/README.md: the original archive name for a
	// contained source revision with a "-filtered-<16 hex>" suffix inserted
	// before the ".tar.gz" extension.
	filteredArchiveSuffix = regexp.MustCompile(`^[a-f0-9]{16}$`)
)

const maxManifestBytes = 1 << 20

func validGoModulePath(packagePath, version string) bool {
	versionParts := canonicalGoVersion.FindStringSubmatch("v" + version)
	if versionParts == nil || !goModulePath.MatchString(packagePath) {
		return false
	}
	parts := strings.Split(packagePath, "/")
	if parts[0] != "localhost" && !strings.Contains(parts[0], ".") {
		return false
	}
	for _, part := range parts {
		if part == "." || part == ".." || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".") {
			return false
		}
	}
	major := versionParts[1]
	if strings.HasPrefix(packagePath, "gopkg.in/") {
		for _, part := range parts[1:] {
			if suffix := strings.LastIndex(part, ".v"); suffix >= 0 && suffix+2 < len(part) {
				candidate := part[suffix+2:]
				if strings.Trim(candidate, "0123456789") == "" {
					return candidate == major && !strings.HasSuffix(version, "+incompatible")
				}
			}
		}
		return false
	}
	for i, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 && strings.Trim(part[1:], "0123456789") == "" {
			candidate := part[1:]
			if candidate == major && major != "0" && major != "1" && !strings.HasSuffix(version, "+incompatible") {
				return true
			}
			// A final vN element is the major-version suffix, so reject an
			// absent or mismatched suffix. Earlier elements may be package
			// subpaths; the install target does not expose the module root.
			if i == len(parts)-1 {
				return false
			}
		}
	}
	if strings.HasSuffix(version, "+incompatible") {
		return major != "0" && major != "1"
	}
	return major == "0" || major == "1"
}

func DecodeManifest(r io.Reader) (Manifest, error) {
	var m Manifest
	data, err := io.ReadAll(io.LimitReader(r, int64(maxManifestBytes+1)))
	if err != nil {
		return m, fmt.Errorf("read build-tool manifest: %w", err)
	}
	if len(data) > maxManifestBytes {
		return m, fmt.Errorf("build-tool manifest exceeds %d-byte limit", maxManifestBytes)
	}
	if err := validateManifestJSONKeys(data); err != nil {
		return m, fmt.Errorf("decode build-tool manifest: %w", err)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, fmt.Errorf("decode build-tool manifest: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return m, fmt.Errorf("manifest must contain exactly one JSON value")
	}
	if err := m.Validate(); err != nil {
		return m, err
	}
	return m, nil
}

type manifestJSONContext uint8

const (
	manifestJSONGeneric manifestJSONContext = iota
	manifestJSONRoot
	manifestJSONBootstrapSource
	manifestJSONTools
	manifestJSONTool
)

var manifestJSONFields = map[manifestJSONContext]map[string]struct{}{
	manifestJSONRoot: {
		"schema_version": {}, "platform": {}, "bootstrap_source": {}, "tools": {},
	},
	manifestJSONBootstrapSource: {
		"revision": {}, "storage_revision": {}, "path": {}, "url": {}, "sha256": {}, "size_bytes": {},
	},
	manifestJSONTool: {
		"name": {}, "version": {}, "license": {}, "source": {}, "sha256": {},
		"binary_sha256": {}, "checksum_source": {}, "kind": {}, "archive_member": {},
		"module": {}, "binary": {}, "version_args": {}, "version_pattern": {},
		"version_source": {}, "version_fallback": {},
	},
}

func validateManifestJSONKeys(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := scanManifestJSONValue(d, manifestJSONRoot); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("manifest must contain exactly one JSON value")
	}
	return nil
}

func scanManifestJSONValue(d *json.Decoder, context manifestJSONContext) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	switch token {
	case json.Delim('{'):
		seen := make(map[string]struct{})
		for d.More() {
			keyToken, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("manifest object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("manifest contains duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if fields := manifestJSONFields[context]; fields != nil {
				if _, canonical := fields[key]; !canonical {
					for field := range fields {
						if strings.EqualFold(key, field) {
							return fmt.Errorf("manifest object key %q is not canonical; use %q", key, field)
						}
					}
				}
			}
			if err := scanManifestJSONValue(d, manifestChildContext(context, key)); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	case json.Delim('['):
		child := manifestJSONGeneric
		if context == manifestJSONTools {
			child = manifestJSONTool
		}
		for d.More() {
			if err := scanManifestJSONValue(d, child); err != nil {
				return err
			}
		}
		_, err = d.Token()
		return err
	default:
		return nil
	}
}

func manifestChildContext(parent manifestJSONContext, key string) manifestJSONContext {
	if parent == manifestJSONRoot && key == "bootstrap_source" {
		return manifestJSONBootstrapSource
	}
	if parent == manifestJSONRoot && key == "tools" {
		return manifestJSONTools
	}
	return manifestJSONGeneric
}

func (m Manifest) Validate() error {
	if m.SchemaVersion != 1 || m.Platform != "linux/amd64" {
		return fmt.Errorf("unsupported manifest schema/platform: %d/%q", m.SchemaVersion, m.Platform)
	}
	source := m.BootstrapSource
	originalPath := "buildtools/source-archives/cloud-tools-bootstrap-" + source.Revision + ".tar.gz"
	filteredPrefix := strings.TrimSuffix(originalPath, ".tar.gz") + "-filtered-"
	pathOK := source.Path == originalPath
	if !pathOK && strings.HasPrefix(source.Path, filteredPrefix) && strings.HasSuffix(source.Path, ".tar.gz") {
		suffix := strings.TrimSuffix(strings.TrimPrefix(source.Path, filteredPrefix), ".tar.gz")
		pathOK = filteredArchiveSuffix.MatchString(suffix)
	}
	expectedURL := "https://raw.githubusercontent.com/" + sourceRepoPath + "/" + source.StorageRevision + "/" + source.Path
	sourceDigest, digestErr := hex.DecodeString(source.SHA256)
	if !commitSHA.MatchString(source.Revision) || !commitSHA.MatchString(source.StorageRevision) || source.StorageRevision == source.Revision || !pathOK || source.URL != expectedURL || digestErr != nil || len(sourceDigest) != 32 || source.SizeBytes <= 0 {
		return fmt.Errorf("bootstrap source must pin the expected raw repository file, distinct full source/storage revisions, SHA-256, and positive size")
	}
	seen := map[string]bool{}
	for _, t := range m.Tools {
		if !safeName.MatchString(t.Name) || !safeName.MatchString(t.Binary) || t.Version == "" || t.License == "" {
			return fmt.Errorf("tool entry has invalid name, version, binary, or license: %q", t.Name)
		}
		if seen[t.Name] {
			return fmt.Errorf("duplicate tool %q", t.Name)
		}
		seen[t.Name] = true
		if t.Kind == "bootstrap-toolchain" {
			u, err := url.Parse(t.Source)
			digest, digestErr := hex.DecodeString(t.SHA256)
			_, versionErr := regexp.Compile(t.VersionPattern)
			if t.Name != "go" || err != nil || u.Scheme != "https" || u.Host == "" || digestErr != nil || len(digest) != 32 || t.ChecksumSource == "" || len(t.VersionArgs) == 0 || versionErr != nil {
				return fmt.Errorf("bootstrap toolchain %q needs an official HTTPS source and SHA-256", t.Name)
			}
			continue
		}
		if t.Kind != "archive" && t.Kind != "raw" && t.Kind != "go-install" {
			return fmt.Errorf("tool %q has unsupported kind %q", t.Name, t.Kind)
		}
		if t.Kind == "go-install" {
			moduleVersionAt := strings.LastIndex(t.Module, "@")
			// Go module queries also accept branches and revisions. Require a
			// canonical Go semantic version so an exactly matching query cannot
			// still resolve to mutable content (for example, @vmain).
			if moduleVersionAt <= 0 || strings.Contains(t.Module[:moduleVersionAt], "@") || t.Module[moduleVersionAt:] != "@v"+t.Version || !validGoModulePath(t.Module[:moduleVersionAt], t.Version) {
				return fmt.Errorf("tool %q must use a version-pinned Go module", t.Name)
			}
			if t.Source != "go install "+t.Module {
				return fmt.Errorf("tool %q source must match its pinned go install module", t.Name)
			}
		} else {
			u, err := url.Parse(t.Source)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				return fmt.Errorf("tool %q source must be an absolute HTTPS URL", t.Name)
			}
			digest, err := hex.DecodeString(t.SHA256)
			if err != nil || len(digest) != 32 || t.ChecksumSource == "" {
				return fmt.Errorf("tool %q must have a SHA-256 and checksum source", t.Name)
			}
		}
		if t.Kind == "archive" && t.ArchiveMember == "" {
			return fmt.Errorf("tool %q archive member is missing", t.Name)
		}
		if t.BinarySHA256 != "" {
			binaryDigest, err := hex.DecodeString(t.BinarySHA256)
			if err != nil || len(binaryDigest) != 32 {
				return fmt.Errorf("tool %q has an invalid installed-binary SHA-256", t.Name)
			}
		}
		if t.VersionSource == "" && (len(t.VersionArgs) == 0 || t.VersionPattern == "") {
			return fmt.Errorf("tool %q needs a runtime version check or digest provenance", t.Name)
		}
		if t.VersionPattern != "" {
			if _, err := regexp.Compile(t.VersionPattern); err != nil {
				return fmt.Errorf("tool %q has invalid version pattern: %w", t.Name, err)
			}
		}
	}
	return nil
}
