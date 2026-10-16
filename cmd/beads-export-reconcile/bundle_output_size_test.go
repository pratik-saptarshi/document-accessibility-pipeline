package main

import (
	"fmt"
	"testing"
)

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
