package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type bundleSizeOverrideInfo struct {
	os.FileInfo
	size int64
}

func (i bundleSizeOverrideInfo) Size() int64 { return i.size }

func TestGeneratedBundleSizeAtLimitAndLimitPlusOne(t *testing.T) {
	const outputLimit = int64(64 << 20)
	if maxBundleBytes != outputLimit {
		t.Fatalf("configured evidence bundle limit = %d, want %d", maxBundleBytes, outputLimit)
	}
	if err := validateGeneratedBundleSize(outputLimit); err != nil {
		t.Fatalf("generated bundle exactly at %d-byte limit rejected: %v", outputLimit, err)
	}
	want := fmt.Sprintf("evidence bundle exceeds %d MiB", outputLimit>>20)
	if err := validateGeneratedBundleSize(outputLimit + 1); err == nil || err.Error() != want {
		t.Fatalf("generated bundle at limit+1 error = %v, want %q", err, want)
	}
}

func TestWriteBundleEnforcesMeasuredOutputSize(t *testing.T) {
	const outputLimit = int64(64 << 20)
	wantTooLarge := fmt.Sprintf("evidence bundle exceeds %d MiB", outputLimit>>20)

	for _, tc := range []struct {
		name      string
		measured  int64
		wantError string
	}{
		{name: "exact limit", measured: outputLimit},
		{name: "limit plus one", measured: outputLimit + 1, wantError: wantTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "bundle.zip")
			stat := func(name string) (os.FileInfo, error) {
				info, err := os.Stat(name)
				if err != nil {
					return nil, err
				}
				return bundleSizeOverrideInfo{FileInfo: info, size: tc.measured}, nil
			}

			err := writeBundleWithStat(output, map[string][]byte{"small.txt": []byte("payload")}, stat)
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("writeBundleWithStat at exact limit: %v", err)
				}
				if _, err := os.Stat(output); err != nil {
					t.Fatalf("exact-limit bundle was not published: %v", err)
				}
			} else {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("writeBundleWithStat at limit+1 error = %v, want %q", err, tc.wantError)
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Fatalf("over-limit bundle exists or stat failed unexpectedly: %v", err)
				}
			}
			matches, err := filepath.Glob(filepath.Join(dir, ".beads-reconcile-*.zip"))
			if err != nil {
				t.Fatalf("glob temporary bundles: %v", err)
			}
			if len(matches) != 0 {
				t.Fatalf("temporary bundles remain: %v", matches)
			}
		})
	}
}
