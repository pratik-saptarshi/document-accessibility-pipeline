package financialfixtures

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var syntheticV2PublicationPaths = []string{
	"docs/fixtures/remediation/financial-tables-v2/statement-positive-v2.pdf",
	"docs/fixtures/remediation/financial-tables-v2/statement-damaged-v2.pdf",
	"docs/fixtures/remediation/financial-tables-v2/invoice-positive-v2.pdf",
	"docs/fixtures/remediation/financial-tables-v2/invoice-damaged-v2.pdf",
}

func syntheticV2PublicationOutputs() map[string][]byte {
	return map[string][]byte{
		syntheticV2PublicationPaths[0]: []byte("synthetic statement positive bytes\x00\x01"),
		syntheticV2PublicationPaths[1]: []byte("synthetic statement damaged bytes\x02\x03"),
		syntheticV2PublicationPaths[2]: []byte("synthetic invoice positive bytes\x04\x05"),
		syntheticV2PublicationPaths[3]: []byte("synthetic invoice damaged bytes\x06\x07"),
	}
}

func syntheticV2PublicationRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, rel := range syntheticV2PublicationPaths {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func assertV2PublicationBytes(t *testing.T, root string, outputs map[string][]byte) {
	t.Helper()
	for _, rel := range syntheticV2PublicationPaths {
		got, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if !bytes.Equal(got, outputs[rel]) {
			t.Errorf("%s bytes = %q, want %q", rel, got, outputs[rel])
		}
		if gotSHA, wantSHA := sha256Hex(got), sha256Hex(outputs[rel]); gotSHA != wantSHA {
			t.Errorf("%s sha256 = %s, want %s", rel, gotSHA, wantSHA)
		}
	}
}

func assertNoV2PublicationTemps(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2/.financial-table-v2-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Errorf("staged temporary files remain: %v", matches)
	}
}

func assertNoV2StagedFiles(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2/.financial-table-v2-*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range matches {
		info, err := os.Lstat(match)
		if err != nil {
			t.Errorf("inspect publication temporary %s: %v", match, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("staged temporary file remains: %s", match)
		}
	}
}

// TestPublishV2ArtifactsRepeatIdenticalPublicationIsNoOp pins the F4
// remediation: re-publishing identical targets must be accepted as a no-op.
// The repeat exercises the conflict preflight's identical-target path; any
// rewrite (re-staging) of an identical target is a regression.
func TestPublishV2ArtifactsRepeatIdenticalPublicationIsNoOp(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	if err := publishV2Artifacts(root, outputs); err != nil {
		t.Fatalf("initial publication: %v", err)
	}
	beforeInfo := make(map[string]os.FileInfo, len(syntheticV2PublicationPaths))
	for _, rel := range syntheticV2PublicationPaths {
		info, err := os.Lstat(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("stat published %s: %v", rel, err)
		}
		beforeInfo[rel] = info
	}
	if err := publishV2Artifacts(root, outputs); err != nil {
		t.Fatalf("repeat identical publication: %v", err)
	}
	assertV2PublicationBytes(t, root, outputs)
	for _, rel := range syntheticV2PublicationPaths {
		after, err := os.Lstat(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("stat republished %s: %v", rel, err)
		}
		if !os.SameFile(beforeInfo[rel], after) {
			t.Errorf("repeat publication rewrote identical target %s: re-staging identical targets is not a no-op", rel)
		}
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsSyntheticOutputsAreExact(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	readBackCalls := 0
	ops := v2ArtifactOps{readBack: func(path string) ([]byte, error) {
		readBackCalls++
		return readV2FileBounded(path, 64<<10)
	}}
	if err := publishV2ArtifactsWithOps(root, outputs, ops); err != nil {
		t.Fatalf("publish synthetic v2 outputs: %v", err)
	}
	if readBackCalls != len(syntheticV2PublicationPaths) {
		t.Fatalf("read-back calls = %d, want one for each of %d outputs", readBackCalls, len(syntheticV2PublicationPaths))
	}
	assertV2PublicationBytes(t, root, outputs)
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsPreservesConflictingExistingTarget(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	path := filepath.Join(root, syntheticV2PublicationPaths[0])
	knownGood := []byte("preexisting known-good bytes")
	if err := os.WriteFile(path, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := publishV2Artifacts(root, outputs); err == nil {
		t.Fatal("conflicting preexisting target was accepted")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, knownGood) {
		t.Fatalf("known-good target changed: bytes=%q err=%v", got, err)
	}
	for _, rel := range syntheticV2PublicationPaths[1:] {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("unexpected output %s exists or could not be checked: %v", rel, err)
		}
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsStagingFailureCleansEarlierStages(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	createCalls := 0
	stageErr := errors.New("injected stage failure")
	ops := v2ArtifactOps{createTemp: func(dir, pattern string) (*os.File, error) {
		createCalls++
		if createCalls == 2 {
			return nil, stageErr
		}
		return os.CreateTemp(dir, pattern)
	}}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, stageErr) {
		t.Fatalf("staging error = %v, want %v", err, stageErr)
	}
	for _, rel := range syntheticV2PublicationPaths {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("output %s exists after staging failure: %v", rel, err)
		}
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsReportsStagedFileCleanupFailure(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	createCalls := 0
	var firstTemporary string
	stageErr := errors.New("injected stage failure")
	cleanupErr := errors.New("injected temporary cleanup failure")
	ops := v2ArtifactOps{
		createTemp: func(dir, pattern string) (*os.File, error) {
			createCalls++
			if createCalls == 2 {
				return nil, stageErr
			}
			file, err := os.CreateTemp(dir, pattern)
			if createCalls == 1 && err == nil {
				firstTemporary = file.Name()
			}
			return file, err
		},
		remove: func(path string) error {
			if path == firstTemporary {
				return cleanupErr
			}
			return os.Remove(path)
		},
	}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, stageErr) || !strings.Contains(err.Error(), cleanupErr.Error()) {
		t.Fatalf("error = %v, want both staging and cleanup failures", err)
	}
	if _, err := os.Stat(firstTemporary); err != nil {
		t.Fatalf("expected residual temporary to be inspectable after reported cleanup failure: %v", err)
	}
	for _, rel := range syntheticV2PublicationPaths {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("output %s exists after staging failure: %v", rel, err)
		}
	}
}

func TestPublishV2ArtifactsSecondPublicationFailureRollsBackOnlyNewFiles(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
	knownGood := outputs[syntheticV2PublicationPaths[3]]
	if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	linkCalls := 0
	publicationErr := errors.New("injected second link failure")
	ops := v2ArtifactOps{link: func(oldPath, newPath string) error {
		linkCalls++
		if linkCalls == 2 {
			return publicationErr
		}
		return os.Link(oldPath, newPath)
	}}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, publicationErr) {
		t.Fatalf("publication error = %v, want %v", err, publicationErr)
	}
	if _, err := os.Lstat(filepath.Join(root, syntheticV2PublicationPaths[0])); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("first new target not rolled back: %v", err)
	}
	got, err := os.ReadFile(knownPath)
	if err != nil || !bytes.Equal(got, knownGood) {
		t.Errorf("preexisting known-good target changed: bytes=%q err=%v", got, err)
	}
	for _, rel := range syntheticV2PublicationPaths[1:3] {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("unexpected output %s exists after rollback: %v", rel, err)
		}
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsReportsRollbackFailureAndResidualFile(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
	knownGood := outputs[syntheticV2PublicationPaths[3]]
	if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	linkCalls := 0
	publicationErr := errors.New("injected second link failure")
	rollbackErr := errors.New("injected rollback failure")
	firstTarget := filepath.Join(root, syntheticV2PublicationPaths[0])
	ops := v2ArtifactOps{
		link: func(oldPath, newPath string) error {
			linkCalls++
			if linkCalls == 2 {
				return publicationErr
			}
			return os.Link(oldPath, newPath)
		},
		remove: func(path string) error {
			if filepath.Base(path) == filepath.Base(firstTarget) {
				return rollbackErr
			}
			return os.Remove(path)
		},
	}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, publicationErr) || !strings.Contains(err.Error(), rollbackErr.Error()) {
		t.Fatalf("error = %v, want publication and rollback failures", err)
	}
	got, readErr := os.ReadFile(firstTarget)
	if readErr != nil || !bytes.Equal(got, outputs[syntheticV2PublicationPaths[0]]) {
		t.Fatalf("residual output not retained/reported as expected: bytes=%q err=%v", got, readErr)
	}
	got, readErr = os.ReadFile(knownPath)
	if readErr != nil || !bytes.Equal(got, knownGood) {
		t.Errorf("preexisting known-good target changed: bytes=%q err=%v", got, readErr)
	}
	assertNoV2StagedFiles(t, root)
	recoveryDirs, err := filepath.Glob(filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2/.financial-table-v2-rollback-*"))
	if err != nil || len(recoveryDirs) != 1 {
		t.Fatalf("recovery directories = %v, err=%v; want one reported residual recovery path", recoveryDirs, err)
	}
	recovered, err := os.ReadFile(filepath.Join(recoveryDirs[0], filepath.Base(firstTarget)))
	if err != nil || !bytes.Equal(recovered, outputs[syntheticV2PublicationPaths[0]]) {
		t.Fatalf("rollback residual recovery file = %q, err=%v", recovered, err)
	}
}

func TestPublishV2ArtifactsReportsRollbackAndRestoreFailure(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
	knownGood := outputs[syntheticV2PublicationPaths[3]]
	if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	linkCalls := 0
	publicationErr := errors.New("injected second link failure")
	restoreErr := errors.New("injected rollback restore failure")
	rollbackErr := errors.New("injected rollback removal failure")
	firstTarget := filepath.Join(root, syntheticV2PublicationPaths[0])
	ops := v2ArtifactOps{
		link: func(oldPath, newPath string) error {
			linkCalls++
			if linkCalls == 2 {
				return publicationErr
			}
			if linkCalls == 3 {
				return restoreErr
			}
			return os.Link(oldPath, newPath)
		},
		remove: func(path string) error {
			if filepath.Base(path) == filepath.Base(firstTarget) {
				return rollbackErr
			}
			return os.Remove(path)
		},
	}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, publicationErr) || !strings.Contains(err.Error(), rollbackErr.Error()) || !strings.Contains(err.Error(), restoreErr.Error()) {
		t.Fatalf("error = %v, want publication, rollback, and restore failures", err)
	}
	if _, statErr := os.Lstat(firstTarget); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("target unexpectedly exists after both rollback actions failed: %v", statErr)
	}
	got, readErr := os.ReadFile(knownPath)
	if readErr != nil || !bytes.Equal(got, knownGood) {
		t.Errorf("preexisting known-good target changed: bytes=%q err=%v", got, readErr)
	}
	assertNoV2StagedFiles(t, root)
	recoveryDirs, globErr := filepath.Glob(filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2/.financial-table-v2-rollback-*"))
	if globErr != nil || len(recoveryDirs) != 1 {
		t.Fatalf("recovery directories = %v, err=%v; want one reported residual recovery path", recoveryDirs, globErr)
	}
	recoveryPath := filepath.Join(recoveryDirs[0], filepath.Base(firstTarget))
	if !strings.Contains(err.Error(), recoveryPath) {
		t.Fatalf("error = %v, want residual recovery path %q", err, recoveryPath)
	}
	recovered, readErr := os.ReadFile(recoveryPath)
	if readErr != nil || !bytes.Equal(recovered, outputs[syntheticV2PublicationPaths[0]]) {
		t.Fatalf("rollback residual recovery file = %q, err=%v", recovered, readErr)
	}
}

func TestPublishV2ArtifactsDoesNotOverwriteRaceCreatedTarget(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
	knownGood := outputs[syntheticV2PublicationPaths[3]]
	if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	raceBytes := []byte("concurrent writer bytes")
	beforeLinkCalls := 0
	ops := v2ArtifactOps{beforeLink: func(oldPath, newPath string) error {
		beforeLinkCalls++
		if beforeLinkCalls != 1 {
			return nil
		}
		return os.WriteFile(newPath, raceBytes, 0o644)
	}}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("publication error = %v, want actual no-clobber link EEXIST", err)
	}
	got, err := os.ReadFile(filepath.Join(root, syntheticV2PublicationPaths[0]))
	if err != nil || !bytes.Equal(got, raceBytes) {
		t.Fatalf("race-created target was overwritten or removed: bytes=%q err=%v", got, err)
	}
	for _, rel := range syntheticV2PublicationPaths[1:3] {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("unexpected output %s exists after no-clobber failure: %v", rel, err)
		}
	}
	got, err = os.ReadFile(knownPath)
	if err != nil || !bytes.Equal(got, knownGood) {
		t.Errorf("preexisting target changed: bytes=%q err=%v", got, err)
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsReadBackFailureIsNotSuccess(t *testing.T) {
	for _, failure := range []string{"read error", "mismatch"} {
		t.Run(failure, func(t *testing.T) {
			root := syntheticV2PublicationRoot(t)
			outputs := syntheticV2PublicationOutputs()
			knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
			knownGood := outputs[syntheticV2PublicationPaths[3]]
			if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
				t.Fatal(err)
			}
			calls := 0
			readErr := errors.New("injected read-back error")
			ops := v2ArtifactOps{readBack: func(path string) ([]byte, error) {
				calls++
				if calls == 2 {
					if failure == "read error" {
						return nil, readErr
					}
					return []byte("wrong read-back bytes"), nil
				}
				return readV2FileBounded(path, 64<<10)
			}}
			err := publishV2ArtifactsWithOps(root, outputs, ops)
			if err == nil {
				t.Fatal("publisher reported success without verifying read-back")
			}
			if failure == "read error" && !errors.Is(err, readErr) {
				t.Fatalf("read-back error = %v, want %v", err, readErr)
			}
			if calls < 2 {
				t.Fatalf("read-back calls = %d, want at least 2 before failure", calls)
			}
			for _, rel := range syntheticV2PublicationPaths[:3] {
				if _, statErr := os.Lstat(filepath.Join(root, rel)); !errors.Is(statErr, os.ErrNotExist) {
					t.Errorf("new output %s remains after failed verification: %v", rel, statErr)
				}
			}
			got, readErr := os.ReadFile(knownPath)
			if readErr != nil || !bytes.Equal(got, knownGood) {
				t.Errorf("preexisting known-good target changed: bytes=%q err=%v", got, readErr)
			}
			assertNoV2PublicationTemps(t, root)
		})
	}
}

func TestPublishV2ArtifactsRollbackDoesNotRemoveRaceReplacement(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	firstTarget := filepath.Join(root, syntheticV2PublicationPaths[0])
	replacement := []byte("replacement writer bytes")
	readCalls := 0
	readErr := errors.New("injected read-back error")
	ops := v2ArtifactOps{readBack: func(path string) ([]byte, error) {
		readCalls++
		if readCalls == 1 {
			return nil, readErr
		}
		return readV2FileBounded(path, 64<<10)
	}, beforeRollback: func(path string) error {
		if path != firstTarget {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove published output for race: %w", err)
		}
		if err := os.WriteFile(path, replacement, 0o644); err != nil {
			return fmt.Errorf("create race replacement: %w", err)
		}
		return nil
	}}
	publicationErr := publishV2ArtifactsWithOps(root, outputs, ops)
	if publicationErr == nil {
		t.Fatal("publisher reported success after read-back error")
	}
	got, err := os.ReadFile(firstTarget)
	if err != nil || !bytes.Equal(got, replacement) {
		t.Fatalf("rollback removed a replacement it did not create: bytes=%q err=%v", got, err)
	}
	if !strings.Contains(publicationErr.Error(), "rollback") || !strings.Contains(publicationErr.Error(), "target changed") || !strings.Contains(publicationErr.Error(), "replacement restored") {
		t.Fatalf("error does not report residual rollback state: %v", publicationErr)
	}
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsStagingFailurePreservesIdenticalPreexisting(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	// Pre-create two of the four targets with exactly the output bytes; the
	// preflight accepts identical preexisting targets and stages nothing for them.
	preexisting := syntheticV2PublicationPaths[:2]
	for _, rel := range preexisting {
		if err := os.WriteFile(filepath.Join(root, rel), outputs[rel], 0o644); err != nil {
			t.Fatal(err)
		}
	}
	createCalls := 0
	stageErr := errors.New("injected stage failure")
	ops := v2ArtifactOps{createTemp: func(dir, pattern string) (*os.File, error) {
		createCalls++
		if createCalls == 2 {
			return nil, stageErr
		}
		return os.CreateTemp(dir, pattern)
	}}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, stageErr) {
		t.Fatalf("staging error = %v, want %v", err, stageErr)
	}
	if createCalls != 2 {
		t.Fatalf("staging calls = %d, want exactly 2 (one per missing target, none for preexisting)", createCalls)
	}
	for _, rel := range preexisting {
		got, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Errorf("read preexisting target %s: %v", rel, err)
			continue
		}
		if !bytes.Equal(got, outputs[rel]) {
			t.Errorf("preexisting target %s changed: bytes=%q", rel, got)
		}
		if gotSHA, wantSHA := sha256Hex(got), sha256Hex(outputs[rel]); gotSHA != wantSHA {
			t.Errorf("preexisting target %s sha256 = %s, want %s", rel, gotSHA, wantSHA)
		}
	}
	for _, rel := range syntheticV2PublicationPaths[2:] {
		if _, err := os.Lstat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("missing target %s exists after staging failure: %v", rel, err)
		}
	}
	assertNoV2PublicationTemps(t, root)
}

// trackedBaselineFileLimit bounds every baseline fixture read: the four
// statement/invoice v1/v2 PDFs are a few kilobytes each, so 16 MiB is a
// generous documented ceiling that keeps a malformed checkout from
// exhausting test memory before hashing.
const trackedBaselineFileLimit = 16 << 20

// trackedBaselineMaxEntries bounds the total number of directory entries the
// baseline scan will visit across both fixture directories: every entry
// returned by a directory read counts, including non-regular entries that
// are skipped rather than hashed. The per-file ceiling above does not bound
// how many files a malformed or polluted checkout can hold; the real fixture
// trees hold ~30 entries, so 4096 leaves more than two orders of magnitude
// of headroom while keeping total scan work bounded.
const trackedBaselineMaxEntries = 4096

// trackedBaselineMaxBytes bounds the total bytes hashed by the baseline
// scan. The real baselines total under 200 KiB; 64 MiB is generous headroom
// that keeps a polluted checkout from exhausting CI memory.
const trackedBaselineMaxBytes = 64 << 20

// trackedBaselineReadBatchSize bounds how many directory entries are
// materialized per read while scanning fixture baselines. The scan streams
// entries in batches of this size so the entry budget above is enforced
// incrementally: a polluted directory with millions of entries is rejected
// after the first few batches instead of being slurped into memory by a
// single unbounded read.
const trackedBaselineReadBatchSize = 512

// fixtureDirStreamer is the minimal surface of *os.File the baseline scan
// needs to stream a directory's entries. It is satisfied by *os.File and is
// injected in tests so the batch-streaming behavior is observable.
type fixtureDirStreamer interface {
	ReadDir(n int) ([]fs.DirEntry, error)
	Close() error
}

// openFixtureDir opens a fixture directory for streaming entry reads. It is
// a variable so tests can substitute a probe that observes batch sizes and
// entry counts; production code always opens the real directory.
var openFixtureDir = func(path string) (fixtureDirStreamer, error) {
	return os.Open(path)
}

func scanTrackedFixtureBaselines(repoRoot string) (map[string]string, error) {
	digests := make(map[string]string)
	var entries int64
	var totalBytes int64
	countEntry := func() error {
		entries++
		if entries > trackedBaselineMaxEntries {
			return fmt.Errorf("fixture baseline scan exceeded entry budget of %d: malformed or polluted checkout", trackedBaselineMaxEntries)
		}
		return nil
	}
	for _, dir := range []string{
		"docs/fixtures/remediation/financial-tables",
		"docs/fixtures/remediation/financial-tables-v2",
	} {
		// Explicit stack instead of filepath.WalkDir: recursion depth stays
		// constant no matter how deeply nested the checkout is.
		stack := []string{filepath.Join(repoRoot, dir)}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := countEntry(); err != nil {
				return nil, err
			}
			// Stream the listing in bounded batches: every entry is counted
			// against the budget before it is retained or descended into,
			// so a polluted directory is rejected mid-stream instead of
			// after one unbounded in-memory read.
			stream, err := openFixtureDir(current)
			if err != nil {
				return nil, fmt.Errorf("scan fixture baseline directory %s: %w", current, err)
			}
			err = func() error {
				defer stream.Close()
				for {
					batch, err := stream.ReadDir(trackedBaselineReadBatchSize)
					if err == io.EOF {
						return nil
					}
					if err != nil {
						return fmt.Errorf("scan fixture baseline directory %s: %w", current, err)
					}
					for _, child := range batch {
						if err := countEntry(); err != nil {
							return err
						}
						path := filepath.Join(current, child.Name())
						// child.Type() never follows symlinks: symlinked directories
						// are not descended into, so cyclic links cannot loop the scan.
						switch {
						case child.IsDir():
							stack = append(stack, path)
						case child.Type().IsRegular():
							data, err := readV2FileBounded(path, trackedBaselineFileLimit)
							if err != nil {
								return err
							}
							totalBytes += int64(len(data))
							if totalBytes > trackedBaselineMaxBytes {
								return fmt.Errorf("fixture baseline scan exceeded byte budget of %d: malformed or polluted checkout", trackedBaselineMaxBytes)
							}
							rel, err := filepath.Rel(repoRoot, path)
							if err != nil {
								return err
							}
							digests[filepath.ToSlash(rel)] = sha256Hex(data)
						default:
							// Skip symlinks, devices, and other non-regular entries,
							// mirroring the previous WalkDir behavior.
						}
					}
				}
			}()
			if err != nil {
				return nil, err
			}
		}
	}
	return digests, nil
}

func hashTrackedFixtureBaselines(t *testing.T, repoRoot string) map[string]string {
	t.Helper()
	digests, err := scanTrackedFixtureBaselines(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	return digests
}

func TestPublishV2ArtifactsLeavesTrackedFixtureBaselinesUnchanged(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	before := hashTrackedFixtureBaselines(t, repoRoot)
	if len(before) == 0 {
		t.Fatal("no tracked fixture baseline files hashed before the publish cycle")
	}
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	if err := publishV2Artifacts(root, outputs); err != nil {
		t.Fatalf("publish against temp root: %v", err)
	}
	assertV2PublicationBytes(t, root, outputs)
	assertNoV2PublicationTemps(t, root)
	after := hashTrackedFixtureBaselines(t, repoRoot)
	for path, beforeSum := range before {
		afterSum, ok := after[path]
		switch {
		case !ok:
			t.Errorf("tracked fixture removed by publish cycle: %s", path)
		case afterSum != beforeSum:
			t.Errorf("tracked fixture modified by publish cycle: %s (before %s, after %s)", path, beforeSum, afterSum)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("tracked fixture added by publish cycle: %s", path)
		}
	}
}

func TestScanTrackedFixtureBaselinesRejectsEntryBudgetOverflow(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		"docs/fixtures/remediation/financial-tables",
		"docs/fixtures/remediation/financial-tables-v2",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	polluted := filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2")
	for i := 0; i < trackedBaselineMaxEntries; i++ {
		if err := os.WriteFile(filepath.Join(polluted, fmt.Sprintf("polluted-%05d.bin", i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := scanTrackedFixtureBaselines(root); err == nil || !strings.Contains(err.Error(), "entry budget") {
		t.Fatalf("scan error = %v, want entry-budget exhaustion", err)
	}
}

func TestScanTrackedFixtureBaselinesRejectsByteBudgetOverflow(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		"docs/fixtures/remediation/financial-tables",
		"docs/fixtures/remediation/financial-tables-v2",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Five 15 MiB files: each is under the 16 MiB per-file ceiling, but the
	// 75 MiB total exceeds the aggregate byte budget.
	chunk := bytes.Repeat([]byte("0123456789abcdef"), 15<<20/16)
	polluted := filepath.Join(root, "docs/fixtures/remediation/financial-tables-v2")
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(filepath.Join(polluted, fmt.Sprintf("large-%d.bin", i)), chunk, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := scanTrackedFixtureBaselines(root); err == nil || !strings.Contains(err.Error(), "byte budget") {
		t.Fatalf("scan error = %v, want byte-budget exhaustion", err)
	}
}

// probeDirEntry is a synthetic non-regular directory entry: it is counted
// against the entry budget like any other entry but is never retained,
// descended into, or hashed, so the probe below needs no files on disk.
type probeDirEntry struct{ name string }

func (e probeDirEntry) Name() string      { return e.name }
func (e probeDirEntry) IsDir() bool       { return false }
func (e probeDirEntry) Type() fs.FileMode { return fs.ModeSymlink }
func (e probeDirEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("probe entry has no FileInfo")
}

// probeDirStreamer presents a huge directory without touching the disk:
// ReadDir returns at most n synthetic entries per call until the entries
// are exhausted, then io.EOF. It records how many batches were served and
// the largest batch requested so the test can verify entries were streamed
// in bounded batches rather than materialized in one read.
type probeDirStreamer struct {
	remaining     int
	batchesServed int
	largestBatch  int
	closed        bool
}

func (s *probeDirStreamer) ReadDir(n int) ([]fs.DirEntry, error) {
	if n > s.largestBatch {
		s.largestBatch = n
	}
	if s.remaining <= 0 {
		return nil, io.EOF
	}
	if n > s.remaining {
		n = s.remaining
	}
	batch := make([]fs.DirEntry, n)
	for i := range batch {
		batch[i] = probeDirEntry{name: fmt.Sprintf("ghost-%08d", s.remaining)}
		s.remaining--
	}
	s.batchesServed++
	return batch, nil
}

func (s *probeDirStreamer) Close() error {
	s.closed = true
	return nil
}

func TestScanTrackedFixtureBaselinesStreamsEntriesBeforeApplyingEntryBudget(t *testing.T) {
	// The entry budget must be enforced incrementally as directory entries
	// stream in, not after the whole listing is materialized: a polluted
	// checkout with millions of entries must be rejected mid-stream, before
	// a single read can consume unbounded memory. The probe presents a
	// directory with twice the budget's worth of entries without creating
	// any files; the scan must reject it while never asking for more than
	// one bounded batch per read.
	oldOpen := openFixtureDir
	defer func() { openFixtureDir = oldOpen }()

	streamer := &probeDirStreamer{remaining: 2 * trackedBaselineMaxEntries}
	openFixtureDir = func(path string) (fixtureDirStreamer, error) {
		return streamer, nil
	}

	root := t.TempDir()
	for _, dir := range []string{
		"docs/fixtures/remediation/financial-tables",
		"docs/fixtures/remediation/financial-tables-v2",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	_, err := scanTrackedFixtureBaselines(root)
	if err == nil || !strings.Contains(err.Error(), "entry budget") {
		t.Fatalf("scan error = %v, want entry-budget exhaustion", err)
	}
	if streamer.batchesServed < 2 {
		t.Errorf("batches served = %d, want at least 2 to prove entries were streamed incrementally", streamer.batchesServed)
	}
	if streamer.largestBatch > trackedBaselineReadBatchSize {
		t.Errorf("largest batch requested = %d, want at most the batch size %d", streamer.largestBatch, trackedBaselineReadBatchSize)
	}
	if !streamer.closed {
		t.Error("scan did not close the directory streamer")
	}
}

func TestScanTrackedFixtureBaselinesHashesSmallTree(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		"docs/fixtures/remediation/financial-tables",
		"docs/fixtures/remediation/financial-tables-v2",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	payload := map[string][]byte{
		"docs/fixtures/remediation/financial-tables/a.pdf":    []byte("aaa"),
		"docs/fixtures/remediation/financial-tables-v2/b.pdf": []byte("bbbb"),
	}
	for rel, data := range payload {
		if err := os.WriteFile(filepath.Join(root, rel), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	digests, err := scanTrackedFixtureBaselines(root)
	if err != nil {
		t.Fatalf("scan small tree: %v", err)
	}
	if len(digests) != len(payload) {
		t.Fatalf("digests = %d entries, want %d", len(digests), len(payload))
	}
	for rel, data := range payload {
		if got := digests[filepath.ToSlash(rel)]; got != sha256Hex(data) {
			t.Errorf("digest %s = %s, want %s", rel, got, sha256Hex(data))
		}
	}
}
