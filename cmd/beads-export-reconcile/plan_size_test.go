package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestReadPlanSizeAtLimitAndLimitPlusOne(t *testing.T) {
	plan := append([]byte(`{"updates":[]}`), bytes.Repeat([]byte{' '}, int(maxPlanBytes)-len(`{"updates":[]}`))...)
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, plan, 0600); err != nil {
		t.Fatal(err)
	}
	got, _, _, err := readRegularBounded(path, maxPlanBytes)
	if err != nil {
		t.Fatalf("plan at %d-byte limit rejected: %v", maxPlanBytes, err)
	}
	if len(got) != int(maxPlanBytes) || !bytes.Equal(got, plan) {
		t.Fatalf("read plan length = %d, want %d identical bytes", len(got), maxPlanBytes)
	}
	var decoded beadsreconcile.Plan
	if err := strictJSON(got, &decoded); err != nil {
		t.Fatalf("valid plan at size limit did not decode: %v", err)
	}
	if len(decoded.Updates) != 0 {
		t.Fatalf("decoded update count = %d, want 0", len(decoded.Updates))
	}

	plan = append(plan, ' ')
	if err := os.WriteFile(path, plan, 0600); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = readRegularBounded(path, maxPlanBytes)
	want := fmt.Sprintf("input exceeds %d-byte limit", maxPlanBytes)
	if err == nil || err.Error() != want {
		t.Fatalf("plan at limit+1 error = %v, want %q", err, want)
	}
}
