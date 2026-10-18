package devsecops

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// maxArchiveMemberScanBytes bounds how much of a single archive member is
// buffered for content-pattern scanning. Members larger than this fail closed.
const maxArchiveMemberScanBytes = 4 << 20

// ArchiveDenylist is a membership denylist for bootstrap source archives.
// Path prefixes are matched against the member path relative to the archive
// root (for example ".beads/"); content patterns are matched against regular
// file contents. Both are fail-closed: a match drops the member.
type ArchiveDenylist struct {
	DeniedPathPrefixes    []string
	DeniedContentPatterns []*regexp.Regexp
}

var (
	personalEmailPattern  = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	personalStreetPattern = regexp.MustCompile(`\b[0-9]{1,6}\s+[A-Z][A-Za-z0-9.'-]*(?:\s+[A-Z][A-Za-z0-9.'-]*)*\s+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr|Court|Ct|Circle|Cir|Way|Terrace|Place|Parkway|Plaza)\b`)
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
			return fmt.Sprintf("content matches denylisted pattern %q", pattern.String())
		}
	}
	return ""
}

// archiveRootRelative strips the "<root>/" prefix from a tar member name so
// denylist rules apply relative to the source root.
func archiveRootRelative(name string) string {
	if i := strings.Index(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// FilterSourceArchive copies a gzip-compressed tar archive from r to w,
// dropping every member denied by the denylist. The output is deterministic:
// kept members are written in archive order with their original headers, and
// the gzip stream uses fixed settings. It returns the dropped member names in
// archive order.
func FilterSourceArchive(r io.Reader, w io.Writer, list ArchiveDenylist) ([]string, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open source archive gzip stream: %w", err)
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)

	gzWriter, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("open filtered archive gzip stream: %w", err)
	}
	tarWriter := tar.NewWriter(gzWriter)

	var dropped []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read source archive entry: %w", err)
		}
		rel := archiveRootRelative(header.Name)
		var content []byte
		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			if header.Size < 0 || header.Size > maxArchiveMemberScanBytes {
				return nil, fmt.Errorf("archive member %q size %d out of bounds", header.Name, header.Size)
			}
			content, err = io.ReadAll(io.LimitReader(tarReader, maxArchiveMemberScanBytes+1))
			if err != nil {
				return nil, fmt.Errorf("read archive member %q: %w", header.Name, err)
			}
			if int64(len(content)) != header.Size {
				return nil, fmt.Errorf("archive member %q truncated: %d of %d bytes", header.Name, len(content), header.Size)
			}
		} else {
			if _, err := io.Copy(io.Discard, tarReader); err != nil {
				return nil, fmt.Errorf("skip archive member %q: %w", header.Name, err)
			}
		}
		if reason := list.DeniedReason(rel, content); reason != "" {
			dropped = append(dropped, header.Name)
			continue
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("write archive member %q: %w", header.Name, err)
		}
		if len(content) > 0 {
			if _, err := tarWriter.Write(content); err != nil {
				return nil, fmt.Errorf("write archive member %q body: %w", header.Name, err)
			}
		}
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
// from r that the denylist would drop, without modifying anything.
func ArchiveDenials(r io.Reader, list ArchiveDenylist) ([]string, error) {
	return FilterSourceArchive(r, io.Discard, list)
}
