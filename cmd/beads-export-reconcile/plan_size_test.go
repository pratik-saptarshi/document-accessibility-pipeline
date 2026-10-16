package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestCreatePlanSizeAtLimitAndLimitPlusOne(t *testing.T) {
	const source = `{"id":"dap-size","title":"before","dependencies":[]}` + "\n"
	const compactPlan = `{"updates":[{"id":"dap-size","fields":{"title":"after"}}]}`
	plan := append([]byte(compactPlan), bytes.Repeat([]byte{' '}, int(maxPlanBytes)-len(compactPlan))...)
	dir := t.TempDir()
	planPath := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(planPath, plan, 0600); err != nil {
		t.Fatal(err)
	}
	got, _, _, err := readRegularBounded(planPath, maxPlanBytes)
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
	if len(decoded.Updates) != 1 || decoded.Updates[0].ID != "dap-size" || decoded.Updates[0].Fields["title"] != "after" {
		t.Fatalf("decoded update plan = %#v, want one valid dap-size title update", decoded)
	}
	if err := beadsreconcile.ValidatePlan([]byte(source), decoded); err != nil {
		t.Fatalf("valid plan at size limit failed canonical-source validation: %v", err)
	}

	plan = append(plan, ' ')
	if err := os.WriteFile(planPath, plan, 0600); err != nil {
		t.Fatal(err)
	}
	canonicalPath := filepath.Join(dir, "canonical.jsonl")
	if err := os.WriteFile(canonicalPath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	bdPath := filepath.Join(dir, "bd")
	invocationMarker := bdPath + ".invoked"
	bdStub := []byte("#!/bin/sh\nprintf invoked > \"$0.invoked\"\n")
	if err := os.WriteFile(bdPath, bdStub, 0700); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(dir, "bundle.zip")
	err = create([]string{"--canonical", canonicalPath, "--plan", planPath, "--bd", bdPath, "--output", outputPath})
	want := fmt.Sprintf("read update plan: input exceeds %d-byte limit", maxPlanBytes)
	if err == nil || err.Error() != want {
		t.Fatalf("create() limit+1 error = %v, want %q", err, want)
	}
	if _, err := os.Stat(invocationMarker); !os.IsNotExist(err) {
		t.Fatalf("Beads executable invocation marker stat error = %v; executable must not run", err)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output bundle stat error = %v; rejected plan must not publish output", err)
	}
}
