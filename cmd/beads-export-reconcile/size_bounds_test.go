package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestCanonicalInputSizeAtLimitAndLimitPlusOne(t *testing.T) {
	const canonicalLimit = 32 << 20
	if maxCanonicalBytes != canonicalLimit {
		t.Fatalf("configured canonical limit = %d, want %d", maxCanonicalBytes, canonicalLimit)
	}
	const prefix = `{"id":"dap-size","title":"before","notes":"`
	const suffix = `"}` + "\n"
	canonical := []byte(prefix + strings.Repeat("x", canonicalLimit-len(prefix)-len(suffix)) + suffix)
	path := filepath.Join(t.TempDir(), "canonical.jsonl")
	if err := os.WriteFile(path, canonical, 0600); err != nil {
		t.Fatal(err)
	}
	got, _, _, err := readRegularBounded(path, maxCanonicalBytes)
	if err != nil {
		t.Fatalf("canonical source at %d-byte limit rejected: %v", canonicalLimit, err)
	}
	if len(got) != canonicalLimit || !bytes.Equal(got, canonical) {
		t.Fatalf("canonical bytes read = %d, want %d identical bytes", len(got), canonicalLimit)
	}
	var plan beadsreconcile.Plan
	if err := strictJSON([]byte(`{"updates":[{"id":"dap-size","fields":{"title":"after"}}]}`), &plan); err != nil {
		t.Fatal(err)
	}
	if err := beadsreconcile.ValidatePlan(got, plan); err != nil {
		t.Fatalf("valid exact-limit canonical source rejected: %v", err)
	}

	canonical = append(canonical, ' ')
	if err := os.WriteFile(path, canonical, 0600); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(filepath.Dir(path), "plan.json")
	if err := os.WriteFile(planPath, []byte(`{"updates":[{"id":"dap-size","fields":{"title":"after"}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	bdPath := filepath.Join(filepath.Dir(path), "bd")
	invocationMarker := writeBDInvocationMarker(t, bdPath)
	outputPath := filepath.Join(filepath.Dir(path), "bundle.zip")
	err = create([]string{"--canonical", path, "--plan", planPath, "--bd", bdPath, "--output", outputPath})
	want := fmt.Sprintf("input exceeds %d-byte limit", canonicalLimit)
	want = "read canonical export: " + want
	if err == nil || err.Error() != want {
		t.Fatalf("create() canonical source at limit+1 error = %v, want %q", err, want)
	}
	if _, err := os.Stat(invocationMarker); !os.IsNotExist(err) {
		t.Fatalf("Beads executable invocation marker stat error = %v; canonical preflight must precede execution", err)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output bundle stat error = %v; rejected canonical source must not publish output", err)
	}
}

func TestBDBinarySizeAtLimitAndLimitPlusOne(t *testing.T) {
	const binaryLimit = int64(256 << 20)
	if maxBDBinaryBytes != binaryLimit {
		t.Fatalf("configured bd binary limit = %d, want %d", maxBDBinaryBytes, binaryLimit)
	}
	path := filepath.Join(t.TempDir(), "bd")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(binaryLimit); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	got, _, _, err := readRegularBounded(path, maxBDBinaryBytes)
	if err != nil {
		t.Fatalf("bd binary at %d-byte limit rejected: %v", binaryLimit, err)
	}
	if int64(len(got)) != binaryLimit || got[0] != 0 || got[len(got)-1] != 0 {
		t.Fatalf("bd binary read length/content = %d/%d/%d, want %d zero-filled bytes", len(got), got[0], got[len(got)-1], binaryLimit)
	}
	got = nil

	file, err = os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(binaryLimit + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(path)
	canonicalPath := filepath.Join(dir, "canonical.jsonl")
	canonical := []byte(`{"id":"dap-size","title":"before","dependencies":[]}` + "\n")
	if err := os.WriteFile(canonicalPath, canonical, 0600); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(planPath, []byte(`{"updates":[{"id":"dap-size","fields":{"title":"after"}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "bundle.zip")
	err = create([]string{"--canonical", canonicalPath, "--plan", planPath, "--bd", path, "--output", outputPath})
	want := fmt.Sprintf("read bd binary: input exceeds %d-byte limit", binaryLimit)
	if err == nil || err.Error() != want {
		t.Fatalf("create() bd binary at limit+1 error = %v, want %q", err, want)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output bundle stat error = %v; rejected binary must not publish output", err)
	}
}

func writeBDInvocationMarker(t *testing.T, path string) string {
	t.Helper()
	marker := path + ".invoked"
	stub := []byte("#!/bin/sh\nprintf invoked > \"$0.invoked\"\n")
	if err := os.WriteFile(path, stub, 0700); err != nil {
		t.Fatal(err)
	}
	return marker
}
