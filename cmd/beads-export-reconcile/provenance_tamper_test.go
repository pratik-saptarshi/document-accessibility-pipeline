package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

// A realistic regression here is accepting a reviewed evidence archive after
// its command history, source snapshot, or candidate has been altered. Rebuild
// the outer ZIP digest in each negative case so rejection must come from the
// archive's provenance and derivation checks.
func TestPromoteRejectsRehashedProvenanceAndSnapshotTampering(t *testing.T) {
	source, entries := syntheticPromotionEntries(t)

	t.Run("positive control promotes exact evidence", func(t *testing.T) {
		canonical := filepath.Join(t.TempDir(), "issues.jsonl")
		bundle := filepath.Join(t.TempDir(), "evidence.zip")
		if err := os.WriteFile(canonical, source, 0600); err != nil {
			t.Fatal(err)
		}
		if err := writeBundle(bundle, entries); err != nil {
			t.Fatal(err)
		}
		bundleBytes, err := os.ReadFile(bundle)
		if err != nil {
			t.Fatal(err)
		}
		if err := promote([]string{"--canonical", canonical, "--bundle", bundle, "--bundle-sha256", sha(bundleBytes)}); err != nil {
			t.Fatalf("valid synthetic evidence bundle rejected: %v", err)
		}
		got, err := os.ReadFile(canonical)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, entries["candidate.jsonl"]) {
			t.Fatalf("promoted bytes differ from verified candidate: got %q want %q", got, entries["candidate.jsonl"])
		}
	})

	tests := []struct {
		name string
		want string
		edit func(map[string][]byte) error
	}{
		{name: "manifest identity", want: "evidence bundle identity mismatch", edit: func(entries map[string][]byte) error {
			var m manifest
			if err := strictJSON(entries["manifest.json"], &m); err != nil {
				return err
			}
			m.BundleID = strings.Repeat("0", 64)
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			if bytes.Equal(data, entries["manifest.json"]) {
				return os.ErrInvalid
			}
			entries["manifest.json"] = data
			return nil
		}},
		{name: "command provenance", want: "command provenance mismatch", edit: func(entries map[string][]byte) error {
			var m manifest
			if err := strictJSON(entries["manifest.json"], &m); err != nil {
				return err
			}
			m.Commands[1].Argv[1] = "--unsafe-hooks"
			return rewriteSyntheticManifest(entries, m)
		}},
		{name: "source snapshot", want: "evidence snapshot or candidate digest mismatch", edit: func(entries map[string][]byte) error {
			return replaceSyntheticBundleText(entries, "source.jsonl", `"title":"parent"`, `"title":"tampered parent"`)
		}},
		{name: "baseline snapshot", want: "evidence snapshot or candidate digest mismatch", edit: func(entries map[string][]byte) error {
			return replaceSyntheticBundleText(entries, "baseline.jsonl", `"title":"parent"`, `"title":"tampered parent"`)
		}},
		{name: "updated snapshot", want: "evidence snapshot or candidate digest mismatch", edit: func(entries map[string][]byte) error {
			return replaceSyntheticBundleText(entries, "updated.jsonl", `"title":"parent"`, `"title":"tampered parent"`)
		}},
		{name: "round-trip snapshot", want: "evidence snapshot or candidate digest mismatch", edit: func(entries map[string][]byte) error {
			return replaceSyntheticBundleText(entries, "round-trip.jsonl", `"title":"parent"`, `"title":"tampered parent"`)
		}},
		{name: "candidate with refreshed manifest digests", want: "second-store candidate is not the derived candidate", edit: func(entries map[string][]byte) error {
			var m manifest
			if err := strictJSON(entries["manifest.json"], &m); err != nil {
				return err
			}
			original := entries["candidate.jsonl"]
			entries["candidate.jsonl"] = bytes.Replace(original, []byte("Review source-freshness continuation"), []byte("Tampered source-freshness continuation"), 1)
			if bytes.Equal(entries["candidate.jsonl"], original) {
				return os.ErrInvalid
			}
			m.Evidence.CandidateSHA256 = sha(entries["candidate.jsonl"])
			m.Validation.CandidateSHA256 = sha(entries["candidate.jsonl"])
			return rewriteSyntheticManifest(entries, m)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			caseEntries := cloneBundleEntries(entries)
			if err := tc.edit(caseEntries); err != nil {
				t.Fatal(err)
			}
			canonical := filepath.Join(t.TempDir(), "issues.jsonl")
			bundle := filepath.Join(t.TempDir(), "evidence.zip")
			if err := os.WriteFile(canonical, source, 0600); err != nil {
				t.Fatal(err)
			}
			if err := writeBundle(bundle, caseEntries); err != nil {
				t.Fatal(err)
			}
			bundleBytes, err := os.ReadFile(bundle)
			if err != nil {
				t.Fatal(err)
			}
			if err := promote([]string{"--canonical", canonical, "--bundle", bundle, "--bundle-sha256", sha(bundleBytes)}); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("promote() error = %v, want rejection containing %q", err, tc.want)
			}
			got, err := os.ReadFile(canonical)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, source) {
				t.Fatalf("rejected tampered bundle changed canonical bytes: got %q want %q", got, source)
			}
		})
	}
}

func syntheticPromotionEntries(t *testing.T) ([]byte, map[string][]byte) {
	t.Helper()
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	if err := strictJSON(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	plan.BDVersion = "1.3.1"
	source := []byte(`{"_type":"issue","id":"dap-jo9","title":"parent","status":"open","dependencies":[]}` + "\n")
	var createFields map[string]json.RawMessage
	if err := json.Unmarshal(plan.Creates[0], &createFields); err != nil {
		t.Fatal(err)
	}
	created, err := json.Marshal(createFields)
	if err != nil {
		t.Fatal(err)
	}
	updated := append(append([]byte(nil), source...), created...)
	updated = append(updated, '\n')
	planBytes, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	candidate, _, err := beadsreconcile.Reconcile(source, source, updated, plan)
	if err != nil {
		t.Fatal(err)
	}
	const binarySHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	evidence, validation, err := beadsreconcile.ValidateCandidateRoundTrip(source, source, updated, candidate, candidate, plan, binarySHA)
	if err != nil {
		t.Fatal(err)
	}
	evidence.RoundTripSHA256 = sha(candidate)
	validation.RoundTripSHA256 = sha(candidate)
	manifestValue := manifest{
		Version: 2, BDVersion: plan.BDVersion, BDBinarySHA: binarySHA,
		Commands: syntheticPromotionCommands(), Evidence: evidence, Validation: validation,
	}
	manifestValue.BundleID = bundleID(manifestValue)
	manifestBytes, err := json.Marshal(manifestValue)
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string][]byte{
		"source.jsonl":     source,
		"baseline.jsonl":   source,
		"updated.jsonl":    updated,
		"plan.json":        planBytes,
		"candidate.jsonl":  candidate,
		"round-trip.jsonl": candidate,
		"manifest.json":    manifestBytes,
	}
	return source, entries
}

func syntheticPromotionCommands() []commandRecord {
	const (
		binary = "/synthetic/bd"
		hash   = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	makeRecord := func(args ...string) commandRecord {
		return commandRecord{Argv: append([]string{binary}, args...), WorkingDirectory: "/synthetic/store", EnvironmentSHA: hash, StdoutSHA: hash, StderrSHA: hash}
	}
	return []commandRecord{
		makeRecord("--version"),
		makeRecord("init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"),
		makeRecord("import", "/synthetic/source.jsonl"),
		makeRecord("export", "-o", "/synthetic/baseline.jsonl"),
		makeRecord("import", "/synthetic/planned-create.jsonl"),
		makeRecord("export", "-o", "/synthetic/updated.jsonl"),
		makeRecord("init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"),
		makeRecord("import", "/synthetic/candidate.jsonl"),
		makeRecord("export", "-o", "/synthetic/round-trip.jsonl"),
	}
}

func rewriteSyntheticManifest(entries map[string][]byte, m manifest) error {
	m.BundleID = bundleID(m)
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	entries["manifest.json"] = data
	return nil
}

func replaceSyntheticBundleText(entries map[string][]byte, name, old, replacement string) error {
	original := entries[name]
	updated := bytes.Replace(original, []byte(old), []byte(replacement), 1)
	if bytes.Equal(original, updated) {
		return os.ErrInvalid
	}
	entries[name] = updated
	return nil
}

func cloneBundleEntries(entries map[string][]byte) map[string][]byte {
	cloned := make(map[string][]byte, len(entries))
	for name, data := range entries {
		cloned[name] = append([]byte(nil), data...)
	}
	return cloned
}
