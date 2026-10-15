package corpus

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemediationManifestIsSyntheticVersionedAndMapped(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	manifest, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if problems := Validate(manifest); len(problems) != 0 {
		t.Fatalf("invalid remediation manifest: %v", problems)
	}
	if manifest.SchemaVersion != "1.0.0" || manifest.Status != "proposed" || manifest.Lane != "remediation" {
		t.Fatalf("manifest scope changed: %#v", manifest)
	}
	if len(manifest.Cases) != 4 {
		t.Fatalf("case count = %d, want one planned baseline for each family", len(manifest.Cases))
	}
	for _, c := range manifest.Cases {
		path := filepath.Join("..", "..", filepath.FromSlash(c.ProfileExample))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("referenced synthetic profile example does not exist: %s", c.ProfileExample)
		}
	}
}

func cloneManifest(m Manifest) Manifest {
	clone := m
	clone.RequirementIDs = append([]string(nil), m.RequirementIDs...)
	clone.BeadsIssueIDs = append([]string(nil), m.BeadsIssueIDs...)
	clone.Cases = append([]Case(nil), m.Cases...)
	for i := range clone.Cases {
		clone.Cases[i].Dimensions = append([]string(nil), m.Cases[i].Dimensions...)
	}
	return clone
}

func TestValidateRejectsIncompleteOrUnsafeManifest(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	base, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name            string
		mutate          func(*Manifest)
		wantProblemPart string
	}{
		{
			name:            "unsupported version",
			mutate:          func(m *Manifest) { m.SchemaVersion = "2.0.0" },
			wantProblemPart: "schema_version",
		},
		{
			name:            "unapproved requirement mapping",
			mutate:          func(m *Manifest) { m.RequirementIDs = []string{"REQ-GEN-001"} },
			wantProblemPart: "requirement_ids",
		},
		{
			name: "extra requirement mapping",
			mutate: func(m *Manifest) {
				m.RequirementIDs = append(m.RequirementIDs, "REQ-GEN-001")
			},
			wantProblemPart: "requirement_ids contains unsupported ID REQ-GEN-001",
		},
		{
			name: "extra Beads mapping",
			mutate: func(m *Manifest) {
				m.BeadsIssueIDs = append(m.BeadsIssueIDs, "dap-jo9.99")
			},
			wantProblemPart: "beads_issue_ids contains unsupported ID dap-jo9.99",
		},
		{
			name:            "non-synthetic provenance",
			mutate:          func(m *Manifest) { m.Provenance = "customer-sourced" },
			wantProblemPart: "provenance",
		},
		{
			name: "missing family",
			mutate: func(m *Manifest) {
				m.Cases[0].Family = "unknown"
			},
			wantProblemPart: "family",
		},
		{
			name: "duplicate case ID",
			mutate: func(m *Manifest) {
				m.Cases[1].ID = m.Cases[0].ID
			},
			wantProblemPart: "case ID must be unique",
		},
		{
			name: "non-synthetic case",
			mutate: func(m *Manifest) {
				m.Cases[0].Synthetic = false
			},
			wantProblemPart: "synthetic",
		},
		{
			name: "unsafe profile path",
			mutate: func(m *Manifest) {
				m.Cases[0].ProfileExample = "docs/profiles/examples/remediation/../../../../private/report.json"
			},
			wantProblemPart: "profile_example",
		},
		{
			name: "artifact already created",
			mutate: func(m *Manifest) {
				m.Cases[0].FixtureStatus = "created"
			},
			wantProblemPart: "fixture_status",
		},
		{
			name: "missing preservation dimension",
			mutate: func(m *Manifest) {
				m.Cases[0].Dimensions = []string{"visible_text"}
			},
			wantProblemPart: "dimensions",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := cloneManifest(base)
			tc.mutate(&m)
			problems := strings.Join(Validate(m), "\n")
			if !strings.Contains(problems, tc.wantProblemPart) {
				t.Fatalf("expected a validation problem containing %q, got %q", tc.wantProblemPart, problems)
			}
		})
	}
}

func TestLoadRejectsDuplicateOrCaseVariantKeys(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// encoding/json binds "PROVENANCE" to the Provenance field case-insensitively,
	// so the shadowed key silently overrides "provenance" instead of failing.
	shadowed := bytes.Replace(data, []byte(`"provenance": "synthetic-only"`), []byte(`"PROVENANCE": "synthetic-only", "provenance": "private-customer-reports"`), 1)
	if bytes.Equal(shadowed, data) {
		t.Fatal("test mutation did not find the provenance field")
	}
	shadowedPath := filepath.Join(t.TempDir(), "shadowed-manifest.json")
	if err := os.WriteFile(shadowedPath, shadowed, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(shadowedPath); err == nil || !strings.Contains(err.Error(), "case-variant") {
		t.Fatalf("Load error = %v, want duplicate/case-variant key rejection", err)
	}

	unicodeShadowed := bytes.Replace(data, []byte(`"synthetic": true`), []byte(`"ſynthetic": false, "synthetic": true`), 1)
	if bytes.Equal(unicodeShadowed, data) {
		t.Fatal("Unicode test mutation did not find the synthetic field")
	}
	unicodePath := filepath.Join(t.TempDir(), "unicode-shadowed-manifest.json")
	if err := os.WriteFile(unicodePath, unicodeShadowed, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(unicodePath); err == nil {
		t.Fatal("Load accepted keys that encoding/json folds to the same name")
	}
}

func TestLoadRejectsStandaloneNonCanonicalKeySpellings(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		from []byte
		to   []byte
	}{
		{
			name: "manifest key",
			from: []byte(`"provenance": "synthetic-only"`),
			to:   []byte(`"PROVENANCE": "synthetic-only"`),
		},
		{
			name: "nested case key",
			from: []byte(`"family": "statements"`),
			to:   []byte(`"FAMILY": "statements"`),
		},
		{
			name: "Unicode-folded nested case key",
			from: []byte(`"synthetic": true`),
			to:   []byte(`"ſynthetic": true`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mutated := bytes.Replace(data, tc.from, tc.to, 1)
			if bytes.Equal(mutated, data) {
				t.Fatalf("test mutation did not find %s", tc.from)
			}
			filename := filepath.Join(t.TempDir(), "noncanonical-manifest.json")
			if err := os.WriteFile(filename, mutated, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(filename); err == nil {
				t.Fatalf("Load accepted noncanonical key spelling %s", tc.to)
			}
		})
	}
}

func TestLoadRejectsExactDuplicateKeys(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Replace(data, []byte(`"provenance": "synthetic-only"`), []byte(`"provenance": "customer-sourced", "provenance": "synthetic-only"`), 1)
	if bytes.Equal(duplicate, data) {
		t.Fatal("test mutation did not find the provenance field")
	}
	filename := filepath.Join(t.TempDir(), "duplicate-manifest.json")
	if err := os.WriteFile(filename, duplicate, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(filename); err == nil {
		t.Fatal("Load accepted an exact duplicate key")
	}
}

func TestLoadRejectsTrailingJSONValue(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "trailing-manifest.json")
	if err := os.WriteFile(filename, append(data, []byte(` {}`)...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(filename); err == nil {
		t.Fatal("Load accepted multiple top-level JSON values")
	}
}

func TestLoadBoundsManifestIncludingTrailingWhitespace(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(t.TempDir(), "oversized-manifest.json")
	data = append(data, bytes.Repeat([]byte(" "), 1<<20)...)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load accepted a manifest larger than its bounded input limit")
	}
}

func TestLoadRejectsUnknownArtifactPathFields(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "fixtures", "remediation", "manifest.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	cases := raw["cases"].([]any)
	cases[0].(map[string]any)["artifact_path"] = "../../private/report.pdf"
	unsafeData, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	unsafePath := filepath.Join(t.TempDir(), "unsafe-manifest.json")
	if err := os.WriteFile(unsafePath, unsafeData, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(unsafePath); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Load error = %v, want unknown-field rejection", err)
	}
}
