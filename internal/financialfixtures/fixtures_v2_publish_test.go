package financialfixtures

import (
	"bytes"
	"errors"
	"fmt"
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

func TestPublishV2ArtifactsSyntheticOutputsAreExactAndRepeatable(t *testing.T) {
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
	if err := publishV2Artifacts(root, outputs); err != nil {
		t.Fatalf("repeat identical publication: %v", err)
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
			if path == firstTarget {
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
	assertNoV2PublicationTemps(t, root)
}

func TestPublishV2ArtifactsDoesNotOverwriteRaceCreatedTarget(t *testing.T) {
	root := syntheticV2PublicationRoot(t)
	outputs := syntheticV2PublicationOutputs()
	knownPath := filepath.Join(root, syntheticV2PublicationPaths[3])
	knownGood := outputs[syntheticV2PublicationPaths[3]]
	if err := os.WriteFile(knownPath, knownGood, 0o644); err != nil {
		t.Fatal(err)
	}
	linkCalls := 0
	raceBytes := []byte("concurrent writer bytes")
	publicationErr := errors.New("destination appeared after preflight")
	ops := v2ArtifactOps{link: func(oldPath, newPath string) error {
		linkCalls++
		if linkCalls == 2 {
			if err := os.WriteFile(newPath, raceBytes, 0o644); err != nil {
				return err
			}
			return publicationErr
		}
		return os.Link(oldPath, newPath)
	}}
	err := publishV2ArtifactsWithOps(root, outputs, ops)
	if !errors.Is(err, publicationErr) {
		t.Fatalf("publication error = %v, want %v", err, publicationErr)
	}
	got, err := os.ReadFile(filepath.Join(root, syntheticV2PublicationPaths[1]))
	if err != nil || !bytes.Equal(got, raceBytes) {
		t.Fatalf("race-created target was overwritten or removed: bytes=%q err=%v", got, err)
	}
	if _, err := os.Lstat(filepath.Join(root, syntheticV2PublicationPaths[0])); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("first new output was not rolled back: %v", err)
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
			if err := os.Remove(path); err != nil {
				return nil, fmt.Errorf("replace published output: %w", err)
			}
			if err := os.WriteFile(path, replacement, 0o644); err != nil {
				return nil, err
			}
			return nil, readErr
		}
		return readV2FileBounded(path, 64<<10)
	}}
	publicationErr := publishV2ArtifactsWithOps(root, outputs, ops)
	if publicationErr == nil {
		t.Fatal("publisher reported success after read-back error")
	}
	got, err := os.ReadFile(firstTarget)
	if err != nil || !bytes.Equal(got, replacement) {
		t.Fatalf("rollback removed a replacement it did not create: bytes=%q err=%v", got, err)
	}
	if !strings.Contains(publicationErr.Error(), "rollback") {
		t.Fatalf("error does not report residual rollback state: %v", publicationErr)
	}
	assertNoV2PublicationTemps(t, root)
}
