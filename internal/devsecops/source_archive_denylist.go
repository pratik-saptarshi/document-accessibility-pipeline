package devsecops

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
)

// maxArchiveMemberScanBytes bounds how much of a single archive member is
// buffered for content-pattern scanning. Members larger than this fail closed.
const maxArchiveMemberScanBytes = 4 << 20

const (
	maxArchiveMemberMetadataBytes = 4 << 10
	maxArchiveMetadataBytes       = 256 << 10
	maxArchiveEntryCount          = 512
	maxArchiveExpandedBytes       = 4 << 20
)

// ArchiveDenylist is a membership denylist for bootstrap source archives.
// Path prefixes are matched against the member path relative to the archive
// root (for example ".beads/"); content patterns are matched against regular
// file contents and emitted textual metadata, including member names. A match
// drops the member.
type ArchiveDenylist struct {
	DeniedPathPrefixes    []string
	DeniedContentPatterns []*regexp.Regexp
}

var (
	archiveRootNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	personalEmailPattern   = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	personalStreetPattern  = regexp.MustCompile(`\b[0-9]{1,6}\s+[A-Z][A-Za-z0-9.'-]*(?:\s+[A-Z][A-Za-z0-9.'-]*)*\s+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr|Court|Ct|Circle|Cir|Way|Terrace|Place|Parkway|Plaza)\b`)
)

// DefaultSourceArchiveDenylist returns the denylist applied to bootstrap
// source archives: issue-tracker state (".beads/") is never published, and
// members whose contents look like personal contact data are dropped.
func DefaultSourceArchiveDenylist() ArchiveDenylist {
	return ArchiveDenylist{
		DeniedPathPrefixes: []string{".beads/"},
		DeniedContentPatterns: []*regexp.Regexp{
			personalEmailPattern,
			personalStreetPattern,
		},
	}
}

// DeniedReason reports why the archive member at relPath (relative to the
// archive root) is denied, or "" when it is allowed. A nil content skips the
// content-pattern check.
func (list ArchiveDenylist) DeniedReason(relPath string, content []byte) string {
	rel := strings.TrimSuffix(relPath, "/")
	for _, prefix := range list.DeniedPathPrefixes {
		bare := strings.TrimSuffix(prefix, "/")
		if rel == bare || strings.HasPrefix(rel, prefix) {
			return fmt.Sprintf("path matches denylisted prefix %q", prefix)
		}
	}
	if content == nil {
		return ""
	}
	for _, pattern := range list.DeniedContentPatterns {
		if pattern.Match(content) {
			return "content matches a denylisted pattern"
		}
	}
	return ""
}

// metadataDenial inspects all textual metadata that archive/tar will emit for
// a member. It deliberately returns no matched value or pattern in diagnostics.
func (list ArchiveDenylist) metadataDenial(header *tar.Header) (bool, int64) {
	var size int64
	denied := false
	check := func(value string) {
		size += int64(len(value))
		for _, pattern := range list.DeniedContentPatterns {
			if pattern.MatchString(value) {
				denied = true
				return
			}
		}
	}
	check(header.Name)
	check(header.Linkname)
	check(header.Uname)
	check(header.Gname)
	for key, value := range header.PAXRecords {
		check(key)
		check(value)
	}
	for key, value := range header.Xattrs {
		check(key)
		check(value)
	}
	return denied, size
}

func (list ArchiveDenylist) archiveNameForDiagnostic(name string) string {
	for _, pattern := range list.DeniedContentPatterns {
		if pattern.MatchString(name) {
			return "[redacted archive member]"
		}
	}
	return name
}

// archiveRootRelative validates one member against the archive's common,
// single-component root and returns its clean root-relative path. Rejecting
// noncanonical paths before denylist matching prevents traversal and rootless
// tracker paths from being normalized into apparently safe names.
func archiveRootRelative(name, diagnosticName, expectedRoot string, directory bool) (string, string, error) {
	if name == "" || strings.Contains(name, `\`) || strings.HasPrefix(name, "/") {
		return "", "", fmt.Errorf("archive member path %q is empty, absolute, or uses backslashes", diagnosticName)
	}
	normalized := name
	if directory && strings.HasSuffix(normalized, "/") {
		normalized = strings.TrimSuffix(normalized, "/")
	}
	clean := path.Clean(normalized)
	if clean != normalized || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", "", fmt.Errorf("archive member path %q is not a clean relative path", diagnosticName)
	}
	parts := strings.Split(clean, "/")
	root := parts[0]
	if !archiveRootNamePattern.MatchString(root) {
		return "", "", fmt.Errorf("archive member path %q has an unsafe root component", diagnosticName)
	}
	if len(parts) == 1 && !directory {
		return "", "", fmt.Errorf("archive member path %q is not beneath an archive root", diagnosticName)
	}
	if expectedRoot != "" && root != expectedRoot {
		return "", "", fmt.Errorf("archive member path %q does not share the archive root", diagnosticName)
	}
	return root, strings.Join(parts[1:], "/"), nil
}

// FilterSourceArchive copies a gzip-compressed tar archive from r to w,
// dropping every member denied by the denylist. The output is deterministic:
// kept members are written in archive order with their original headers, and
// the gzip stream uses fixed settings. It returns dropped member names in
// archive order, redacting any name that matches a denied content pattern.
func FilterSourceArchive(r io.Reader, w io.Writer, list ArchiveDenylist) ([]string, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open source archive gzip stream: %w", err)
	}
	defer gzReader.Close()
	boundedGzip := &io.LimitedReader{R: gzReader, N: maxArchiveExpandedBytes + 1}
	tarReader := tar.NewReader(boundedGzip)

	gzWriter, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("open filtered archive gzip stream: %w", err)
	}
	tarWriter := tar.NewWriter(gzWriter)

	var dropped []string
	archiveRoot := ""
	entryCount := 0
	archiveMetadataBytes := int64(0)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read source archive entry: %w", err)
		}
		diagnosticName := list.archiveNameForDiagnostic(header.Name)
		entryCount++
		if entryCount > maxArchiveEntryCount {
			return nil, fmt.Errorf("source archive exceeds the %d entry limit", maxArchiveEntryCount)
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			if header.Name != "pax_global_header" || header.Size < 0 || header.Size > 256 {
				return nil, fmt.Errorf("unexpected global archive metadata %q (%d bytes)", diagnosticName, header.Size)
			}
			for key := range header.PAXRecords {
				if key != "comment" {
					return nil, fmt.Errorf("unexpected global archive metadata key")
				}
			}
			metadata, err := io.ReadAll(io.LimitReader(tarReader, 257))
			if err != nil || int64(len(metadata)) != header.Size {
				return nil, fmt.Errorf("invalid global archive metadata %q", diagnosticName)
			}
			metadataDenied, metadataSize := list.metadataDenial(header)
			metadataSize += int64(len(metadata))
			if metadataSize > maxArchiveMemberMetadataBytes || archiveMetadataBytes+metadataSize > maxArchiveMetadataBytes {
				return nil, fmt.Errorf("global archive metadata exceeds configured bounds")
			}
			archiveMetadataBytes += metadataSize
			if !metadataDenied {
				for _, pattern := range list.DeniedContentPatterns {
					if pattern.Match(metadata) {
						metadataDenied = true
						break
					}
				}
			}
			if metadataDenied {
				return nil, fmt.Errorf("global archive metadata matches a denylisted pattern")
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				return nil, fmt.Errorf("write global archive metadata: %w", err)
			}
			if _, err := tarWriter.Write(metadata); err != nil {
				return nil, fmt.Errorf("write global archive metadata body: %w", err)
			}
			continue
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if header.Size != 0 || header.Linkname != "" {
				return nil, fmt.Errorf("archive directory %q has unexpected size or link target", diagnosticName)
			}
		case tar.TypeReg, tar.TypeRegA:
			if header.Linkname != "" {
				return nil, fmt.Errorf("archive regular file %q has an unexpected link target", diagnosticName)
			}
		default:
			return nil, fmt.Errorf("unsupported archive member type %d for %q", header.Typeflag, diagnosticName)
		}
		metadataDenied, metadataSize := list.metadataDenial(header)
		if metadataSize > maxArchiveMemberMetadataBytes || archiveMetadataBytes+metadataSize > maxArchiveMetadataBytes {
			return nil, fmt.Errorf("archive member metadata exceeds configured bounds")
		}
		archiveMetadataBytes += metadataSize
		root, rel, err := archiveRootRelative(header.Name, diagnosticName, archiveRoot, header.Typeflag == tar.TypeDir)
		if err != nil {
			return nil, err
		}
		if archiveRoot == "" {
			archiveRoot = root
		}
		var content []byte
		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			if header.Size < 0 || header.Size > maxArchiveMemberScanBytes {
				return nil, fmt.Errorf("archive member %q size %d out of bounds", diagnosticName, header.Size)
			}
			content, err = io.ReadAll(io.LimitReader(tarReader, maxArchiveMemberScanBytes+1))
			if err != nil {
				return nil, fmt.Errorf("read archive member %q: %w", diagnosticName, err)
			}
			if int64(len(content)) != header.Size {
				return nil, fmt.Errorf("archive member %q truncated: %d of %d bytes", diagnosticName, len(content), header.Size)
			}
		} else {
			if _, err := io.Copy(io.Discard, tarReader); err != nil {
				return nil, fmt.Errorf("skip archive member %q: %w", diagnosticName, err)
			}
		}
		if metadataDenied || list.DeniedReason(rel, content) != "" {
			dropped = append(dropped, diagnosticName)
			continue
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("write archive member %q: %w", diagnosticName, err)
		}
		if len(content) > 0 {
			if _, err := tarWriter.Write(content); err != nil {
				return nil, fmt.Errorf("write archive member %q body: %w", diagnosticName, err)
			}
		}
	}
	if _, err := io.Copy(io.Discard, boundedGzip); err != nil {
		return nil, fmt.Errorf("finish source archive stream: %w", err)
	}
	if boundedGzip.N == 0 {
		return nil, fmt.Errorf("source archive exceeds the %d-byte expanded limit", maxArchiveExpandedBytes)
	}
	if archiveRoot == "" {
		return nil, fmt.Errorf("source archive has no rooted members")
	}
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("close filtered archive tar stream: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("close filtered archive gzip stream: %w", err)
	}
	if err := gzReader.Close(); err != nil {
		return nil, fmt.Errorf("close source archive gzip stream: %w", err)
	}
	return dropped, nil
}

// ArchiveDenials lists the member names in the gzip-compressed tar archive
// from r that the denylist would drop, redacting names that match denied
// content patterns and without modifying anything.
func ArchiveDenials(r io.Reader, list ArchiveDenylist) ([]string, error) {
	return FilterSourceArchive(r, io.Discard, list)
}
