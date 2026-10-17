// dap-jo9.1.6.7 spike harness (test-only).
// These tests exercise a manually supplied suggestion and a small operator gate.
// They do not read financial fixtures, parse PDFs, implement a suggester, or
// measure mapping accuracy. All suggestion data below is synthetic test input.

package financialfixtures

import "testing"

// spikeAssoc is a test-local cell-to-header pair. It is not fixture metadata.
type spikeAssoc struct {
	cell   string
	header string
}

// spikeSuggestion is a literal test input to the operator-gate harness.
type spikeSuggestion struct {
	assoc               spikeAssoc
	confidence          float64
	confirmedByOperator bool
	unsafe              bool
	approach            string
}

// spikeApplier models only the confirmation/safety gate, not mapping.
type spikeApplier struct {
	applied     []spikeAssoc
	ambiguous   []string
	conflicting []string
}

func (a *spikeApplier) apply(s spikeSuggestion) {
	if s.confirmedByOperator && !s.unsafe {
		a.applied = append(a.applied, s.assoc)
	}
}

func TestHarnessAppliesOnlySafeConfirmedSuggestion(t *testing.T) {
	for _, tc := range []struct {
		name        string
		suggestion  spikeSuggestion
		wantApplied bool
	}{
		{
			name: "safe but unconfirmed",
			suggestion: spikeSuggestion{
				assoc: spikeAssoc{cell: "test-cell", header: "test-header"},
			},
		},
		{
			name: "safe and confirmed",
			suggestion: spikeSuggestion{
				assoc:               spikeAssoc{cell: "test-cell", header: "test-header"},
				confirmedByOperator: true,
			},
			wantApplied: true,
		},
		{
			name: "confirmed but unsafe",
			suggestion: spikeSuggestion{
				assoc:               spikeAssoc{cell: "test-cell", header: "test-header"},
				confirmedByOperator: true,
				unsafe:              true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gate spikeApplier
			gate.apply(tc.suggestion)
			if got := len(gate.applied) == 1; got != tc.wantApplied {
				t.Fatalf("gate applied = %v, want %v", got, tc.wantApplied)
			}
			if tc.wantApplied && gate.applied[0] != tc.suggestion.assoc {
				t.Fatalf("gate applied %+v, want passed-in value %+v", gate.applied[0], tc.suggestion.assoc)
			}
		})
	}
}

func TestHarnessMissingStubReturnsNoSuggestion(t *testing.T) {
	// This closure represents only the missing-value case; it is not a lookup
	// against a fixture or a production suggester.
	stub := func(string) (spikeSuggestion, bool) { return spikeSuggestion{}, false }
	suggestion, ok := stub("test-cell")
	if ok {
		t.Fatal("missing stub returned a suggestion")
	}

	// A missing suggestion (the stub's zero value) must never reach the
	// gate's applied list, even when it is fed through the gate.
	var gate spikeApplier
	gate.apply(suggestion)
	if len(gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied for a missing suggestion", gate.applied)
	}
}

func TestHarnessAmbiguousStubSuggestionsRemainUnconfirmed(t *testing.T) {
	cell := "test-cell"
	stub := func(cell string) []spikeSuggestion {
		return []spikeSuggestion{
			{assoc: spikeAssoc{cell: cell, header: "header-a"}, confidence: 0.7},
			{assoc: spikeAssoc{cell: cell, header: "header-b"}, confidence: 0.7},
		}
	}
	candidates := stub(cell)
	if len(candidates) != 2 || candidates[0].confidence != candidates[1].confidence {
		t.Fatalf("ambiguous stub candidates = %+v, want two equal-confidence values", candidates)
	}

	var gate spikeApplier
	for _, candidate := range candidates {
		if candidate.confirmedByOperator {
			t.Fatal("ambiguous stub candidate must remain unconfirmed")
		}
		gate.apply(candidate)
	}
	gate.ambiguous = append(gate.ambiguous, cell)
	if len(gate.applied) != 0 || len(gate.ambiguous) != 1 || gate.ambiguous[0] != cell {
		t.Fatalf("gate state after ambiguous stub values = %+v, want none applied and cell recorded ambiguous", gate)
	}
}

func TestHarnessConflictingStubSuggestionsRemainUnapplied(t *testing.T) {
	cell := "test-cell"
	fromLookupStub := spikeSuggestion{assoc: spikeAssoc{cell: cell, header: "header-a"}}
	fromHeuristicStub := spikeSuggestion{assoc: spikeAssoc{cell: cell, header: "header-b"}}

	var gate spikeApplier
	if fromLookupStub.assoc.header != fromHeuristicStub.assoc.header {
		gate.conflicting = append(gate.conflicting, cell)
	} else {
		gate.apply(fromLookupStub)
	}
	if len(gate.applied) != 0 || len(gate.conflicting) != 1 || gate.conflicting[0] != cell {
		t.Fatalf("gate state after conflicting stub values = %+v, want none applied and cell recorded conflicting", gate)
	}
}

func TestHarnessUnsafeConfirmedSuggestionIsRejected(t *testing.T) {
	suggestion := spikeSuggestion{
		assoc:               spikeAssoc{cell: "test-cell", header: "test-header"},
		confirmedByOperator: true,
		unsafe:              true,
	}
	var gate spikeApplier
	gate.apply(suggestion)
	if len(gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want unsafe value rejected", gate.applied)
	}
}

func TestHarnessUnconfirmedStubValuesNeverApply(t *testing.T) {
	suggestions := []spikeSuggestion{
		{assoc: spikeAssoc{cell: "test-cell-a", header: "header-a"}, approach: "lookup-stub"},
		{assoc: spikeAssoc{cell: "test-cell-b", header: "header-b"}, confidence: 0.7, approach: "heuristic-stub"},
	}
	for _, suggestion := range suggestions {
		var gate spikeApplier
		gate.apply(suggestion)
		if len(gate.applied) != 0 {
			t.Fatalf("unconfirmed %s value applied %+v", suggestion.approach, gate.applied)
		}
	}
}
