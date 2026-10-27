package main

import (
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

// Mutant-killing tests for closure command provenance.
//
// Decision 0008's closure path is only trustworthy if validateProvenance
// pins the exact command history: metrics disabled and confirmed before
// anything else, the exact authorized Beads 1.3.1 binary digest, an
// exact command count, and the exact close invocation. The pre-existing
// coverage of these checks sits behind the DAP_BD_BINARY integration
// gate, so weakened checks survived the ungated suite. Each tamper case
// below was verified to FAIL by assertion against a production mutation
// that removes the named check, and the positive control passes against
// unmutated code.

const closureProvenanceReason = "Acceptance evidence: merged PRs #89, #90, #92, #93, #94, and #96. This bundle records the pinned Beads close event, source snapshots, command provenance, and candidate."

func syntheticClosureProvenance(t *testing.T) (manifest, beadsreconcile.Plan) {
	t.Helper()
	plan := beadsreconcile.Plan{BDVersion: "1.3.1", Closes: []beadsreconcile.IssueClosure{{
		ID: "dap-jo9.18", ExpectedStatus: "open", DecisionID: "0008", Reason: closureProvenanceReason,
	}}}
	const (
		binary = "/synthetic/bd"
		digest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	record := func(args ...string) commandRecord {
		return commandRecord{
			Argv: append([]string{binary}, args...), WorkingDirectory: "/synthetic/store",
			EnvironmentSHA: digest, StdoutSHA: digest, StderrSHA: digest,
		}
	}
	m := manifest{
		Version: 2, BDVersion: "1.3.1",
		BDBinarySHA:   "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653",
		MetricsStatus: "OFF",
		Commands: []commandRecord{
			record("metrics", "off"),
			record("metrics"),
			record("--version"),
			record("init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"),
			record("import", "/synthetic/source.jsonl"),
			record("export", "-o", "/synthetic/baseline.jsonl"),
			record("close", "dap-jo9.18", "--reason="+closureProvenanceReason),
			record("export", "-o", "/synthetic/updated.jsonl"),
			record("init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"),
			record("import", "/synthetic/candidate.jsonl"),
			record("export", "-o", "/synthetic/round-trip.jsonl"),
		},
	}
	return m, plan
}

func cloneManifest(m manifest) manifest {
	cloned := m
	cloned.Commands = append([]commandRecord(nil), m.Commands...)
	for i := range cloned.Commands {
		cloned.Commands[i].Argv = append([]string(nil), m.Commands[i].Argv...)
	}
	return cloned
}

func TestValidateProvenanceClosureMutantsRejected(t *testing.T) {
	base, plan := syntheticClosureProvenance(t)
	if err := validateProvenance(base, plan); err != nil {
		t.Fatalf("positive control: validateProvenance() rejected the exact closure history: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*manifest)
	}{
		{name: "metrics status not OFF", mutate: func(m *manifest) { m.MetricsStatus = "ON" }},
		{name: "closure binary digest not the authorized pin", mutate: func(m *manifest) {
			m.BDBinarySHA = strings.Repeat("c", 64)
		}},
		{name: "telemetry prefix commands missing", mutate: func(m *manifest) {
			m.Commands = m.Commands[2:]
		}},
		{name: "metrics off command altered", mutate: func(m *manifest) {
			m.Commands[0].Argv[2] = "on"
		}},
		{name: "metrics confirmation command altered", mutate: func(m *manifest) {
			m.Commands[1].Argv = append(m.Commands[1].Argv, "off")
		}},
		{name: "close target altered", mutate: func(m *manifest) {
			m.Commands[6].Argv[2] = "dap-jo9.19"
		}},
		{name: "close reason altered", mutate: func(m *manifest) {
			m.Commands[6].Argv[3] += " altered"
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bad := cloneManifest(base)
			tc.mutate(&bad)
			if err := validateProvenance(bad, plan); err == nil {
				t.Fatal("validateProvenance() accepted tampered closure provenance")
			}
		})
	}
}
