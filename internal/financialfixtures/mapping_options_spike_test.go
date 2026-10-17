// Copyright 2026
// SPDX-License-Identifier: Apache-2.0

// dap-jo9.1.6.7 spike harness (discovery only; TEST-ONLY).
// Explores mapping-association suggestion accuracy with an operator-gated
// applier: a suggestion is applied only when confirmedByOperator && !unsafe.
// All symbols here are unexported and test-local; no production code changes.

package financialfixtures

import "testing"

// spikeAssoc is a cell→header association proposed by a suggester.
type spikeAssoc struct {
	cell   string
	header string
}

// spikeSuggestion is one suggester's proposal for a cell's header mapping.
type spikeSuggestion struct {
	assoc               spikeAssoc
	confidence          float64
	confirmedByOperator bool
	unsafe              bool
	approach            string
}

// spikeApplier accumulates applied associations; ambiguous or conflicting
// proposals are recorded by the caller instead of being applied.
type spikeApplier struct {
	applied     []spikeAssoc
	ambiguous   []string
	conflicting []string
}

// apply appends s.assoc to applied ONLY if the suggestion was confirmed by
// the operator and is not marked unsafe (the accuracy gate under test).
func (a *spikeApplier) apply(s spikeSuggestion) {
	if s.confirmedByOperator && !s.unsafe {
		a.applied = append(a.applied, s.assoc)
	}
}

// TestMappingOptionsAccuracy: fixture-authored expected association for the
// synthetic statement row-2 amount cell → amount header. A stub suggester
// returns that association with confirmedByOperator=true and unsafe=false;
// feeding it to a fresh spikeApplier must apply exactly that association.
func TestMappingOptionsAccuracy(t *testing.T) {
	// Fixture-authored expected association.
	want := spikeAssoc{cell: "row-2-amount", header: "amount"}

	// stubFixtureLookup: test-local stub suggester returning the expected
	// association as an operator-confirmed, safe fixture-lookup suggestion.
	stubFixtureLookup := func(cell string) (spikeSuggestion, bool) {
		if cell != want.cell {
			return spikeSuggestion{}, false
		}
		return spikeSuggestion{
			assoc:               want,
			confidence:          1.0,
			confirmedByOperator: true,
			unsafe:              false,
			approach:            "fixture-lookup",
		}, true
	}

	// stubHeaderTextHeuristic: secondary test-local stub; not consulted in
	// this accuracy control but kept for later harness tasks.
	stubHeaderTextHeuristic := func(cell string) (spikeSuggestion, bool) {
		return spikeSuggestion{}, false
	}
	_ = stubHeaderTextHeuristic

	s, ok := stubFixtureLookup(want.cell)
	if !ok {
		t.Fatal("stubFixtureLookup returned no suggestion for row-2 amount cell")
	}

	var applier spikeApplier
	applier.apply(s)

	if len(applier.applied) != 1 {
		t.Fatalf("applied %d associations, want exactly 1", len(applier.applied))
	}
	if got := applier.applied[0]; got != want {
		t.Fatalf("applied association = %+v, want %+v", got, want)
	}
	if applier.applied[0].header != "amount" {
		t.Fatalf("applied header = %q, want %q", applier.applied[0].header, "amount")
	}
}
