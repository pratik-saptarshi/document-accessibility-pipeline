package cloudtools

import (
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
)

func DecodeManifest(r io.Reader) (Manifest, error) {
	var m Manifest
	d := json.NewDecoder(r)
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

func (m Manifest) Validate() error {
	if m.SchemaVersion != 1 || m.Platform != "linux/amd64" {
		return fmt.Errorf("unsupported manifest schema/platform: %d/%q", m.SchemaVersion, m.Platform)
	}
	source := m.BootstrapSource
	expectedPath := "buildtools/source-archives/cloud-tools-bootstrap-" + source.Revision + ".tar.gz"
	expectedURL := "https://raw.githubusercontent.com/" + sourceRepoPath + "/" + source.StorageRevision + "/" + source.Path
	sourceDigest, digestErr := hex.DecodeString(source.SHA256)
	if !commitSHA.MatchString(source.Revision) || !commitSHA.MatchString(source.StorageRevision) || source.StorageRevision == source.Revision || source.Path != expectedPath || source.URL != expectedURL || digestErr != nil || len(sourceDigest) != 32 || source.SizeBytes <= 0 {
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
			if t.Module == "" || !strings.Contains(t.Module, "@v"+t.Version) {
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
