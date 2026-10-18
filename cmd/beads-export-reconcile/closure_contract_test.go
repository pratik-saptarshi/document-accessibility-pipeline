package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

func TestCreateCommandCapturesPinnedScopedClosureWithMetricsOff(t *testing.T) {
	bd := os.Getenv("DAP_BD_BINARY")
	if bd == "" {
		t.Skip("set DAP_BD_BINARY to the exact Beads 1.3.1 executable for the isolated closure integration test")
	}
	canonical := filepath.Join("..", "..", "internal", "beadsreconcile", "testdata", "approved-issue-creation-source.jsonl")
	source, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	const sourceSHA = "5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465"
	if got := sha(source); got != sourceSHA {
		t.Fatalf("closure source digest = %s, want %s", got, sourceSHA)
	}
	plan := beadsreconcile.Plan{BDVersion: "1.3.1", Closes: []beadsreconcile.IssueClosure{{
		ID: "dap-jo9.18", ExpectedSourceSHA256: sourceSHA, ExpectedStatus: "open", DecisionID: "0008",
		Reason:       "Acceptance evidence: merged PRs #89, #90, #92, #93, #94, and #96. This bundle records the pinned Beads close event, source snapshots, command provenance, and candidate.",
		EvidenceRefs: []string{"#89", "#90", "#92", "#93", "#94", "#96"},
	}}}
	planBytes, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	planPath := filepath.Join(dir, "plan.json")
	bundlePath := filepath.Join(dir, "evidence.zip")
	if err := os.WriteFile(planPath, planBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := create(context.Background(), []string{"--canonical", canonical, "--plan", planPath, "--bd", bd, "--output", bundlePath}); err != nil {
		t.Fatalf("create() failed in disposable stores: %v", err)
	}
	bundleBytes, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := readBundle(bundleBytes)
	if err != nil {
		t.Fatal(err)
	}
	var got manifest
	if err := strictJSON(entries["manifest.json"], &got); err != nil {
		t.Fatal(err)
	}
	if got.MetricsStatus != "OFF" || got.BDVersion != "1.3.1" || got.BDBinarySHA != "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653" {
		t.Fatalf("manifest pin/metrics status = %s/%s/%s", got.MetricsStatus, got.BDVersion, got.BDBinarySHA)
	}
	if len(got.Commands) != 11 || got.Commands[0].Argv[1] != "metrics" || got.Commands[0].Argv[2] != "off" || got.Commands[1].Argv[1] != "metrics" || got.Commands[2].Argv[1] != "--version" {
		t.Fatalf("command sequence did not start with metrics off/status/version: %+v", got.Commands)
	}
	if err := validateProvenance(got, plan); err != nil {
		t.Fatalf("validateProvenance() rejected exact command history: %v", err)
	}
	for _, mutate := range []struct {
		name string
		fn   func(*manifest)
	}{
		{"metrics order", func(m *manifest) { m.Commands[0], m.Commands[1] = m.Commands[1], m.Commands[0] }},
		{"metrics status", func(m *manifest) { m.MetricsStatus = "ON" }},
		{"close command", func(m *manifest) { m.Commands[6].Argv[1] = "update" }},
		{"close reason", func(m *manifest) { m.Commands[6].Argv[3] += " altered" }},
		{"force flag", func(m *manifest) { m.Commands[6].Argv = append(m.Commands[6].Argv, "--force") }},
		{"duplicate close", func(m *manifest) {
			m.Commands = append(m.Commands[:7], append([]commandRecord{m.Commands[6]}, m.Commands[7:]...)...)
		}},
	} {
		t.Run("provenance rejects "+mutate.name, func(t *testing.T) {
			bad := got
			bad.Commands = append([]commandRecord(nil), got.Commands...)
			for i := range bad.Commands {
				bad.Commands[i].Argv = append([]string(nil), got.Commands[i].Argv...)
			}
			mutate.fn(&bad)
			if err := validateProvenance(bad, plan); err == nil {
				t.Fatal("validateProvenance() accepted modified command history")
			}
		})
	}
	closeRecord := got.Commands[6]
	if len(closeRecord.Argv) != 4 || closeRecord.Argv[1] != "close" || closeRecord.Argv[2] != "dap-jo9.18" || closeRecord.Argv[3] != "--reason="+plan.Closes[0].Reason || closeRecord.ExitCode != 0 {
		t.Fatalf("close command provenance = %+v", closeRecord)
	}
	if len(got.Evidence.Closures) != 1 || got.Evidence.Closures[0].ID != "dap-jo9.18" {
		t.Fatalf("closure evidence = %+v", got.Evidence.Closures)
	}
	var sourceLines, candidateLines [][]byte
	for _, line := range bytes.Split(entries["source.jsonl"], []byte("\n")) {
		if len(line) > 0 {
			sourceLines = append(sourceLines, line)
		}
	}
	for _, line := range bytes.Split(entries["candidate.jsonl"], []byte("\n")) {
		if len(line) > 0 {
			candidateLines = append(candidateLines, line)
		}
	}
	if len(sourceLines) != len(candidateLines) {
		t.Fatalf("candidate rows = %d, source rows = %d", len(candidateLines), len(sourceLines))
	}
	changed := 0
	for i := range sourceLines {
		var before, after map[string]json.RawMessage
		if err := json.Unmarshal(sourceLines[i], &before); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(candidateLines[i], &after); err != nil {
			t.Fatal(err)
		}
		var id string
		_ = json.Unmarshal(before["id"], &id)
		if !bytes.Equal(sourceLines[i], candidateLines[i]) {
			changed++
			if id != "dap-jo9.18" {
				t.Fatalf("unrelated source row %s changed", id)
			}
			for key := range before {
				if key != "status" && key != "updated_at" && key != "closed_at" && key != "close_reason" && !bytes.Equal(before[key], after[key]) {
					t.Fatalf("unapproved target field %s changed", key)
				}
			}
		}
	}
	if changed != 1 {
		t.Fatalf("changed source row count = %d, want exactly one", changed)
	}
}

func TestRunClosureReturnsBeadsRefusalWithoutRetry(t *testing.T) {
	closure := beadsreconcile.IssueClosure{ID: "dap-jo9.18", Reason: "approved reason"}
	called := 0
	var got []string
	wantErr := errors.New("close refused")
	err := runClosure(func(args ...string) ([]byte, error) {
		called++
		got = append([]string(nil), args...)
		return nil, wantErr
	}, closure)
	if !errors.Is(err, wantErr) || called != 1 {
		t.Fatalf("runClosure() error/calls = %v/%d, want refusal and one invocation", err, called)
	}
	want := []string{"close", "dap-jo9.18", "--reason=approved reason"}
	if !bytes.Equal([]byte(strings.Join(got, "\x00")), []byte(strings.Join(want, "\x00"))) {
		t.Fatalf("close argv = %q, want %q", got, want)
	}
}
