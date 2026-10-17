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

// TestMappingOptionsMissing: a cell with no fixture-authored association.
// The stub lookup returns ok=false; with no suggestion, the applier's
// applied list must stay empty.
func TestMappingOptionsMissing(t *testing.T) {
	// stubFixtureLookup: test-local stub with no authored association for
	// this cell; returns ok=false.
	stubFixtureLookup := func(cell string) (spikeSuggestion, bool) {
		return spikeSuggestion{}, false
	}

	_, ok := stubFixtureLookup("row-9-missing")
	if ok {
		t.Fatal("stubFixtureLookup returned a suggestion for a cell with no association")
	}

	var applier spikeApplier
	if len(applier.applied) != 0 {
		t.Fatalf("applied %d associations, want 0 (nothing to apply)", len(applier.applied))
	}
}

// TestMappingOptionsMultiple: the heuristic stub proposes two
// equal-confidence headers for one cell. The caller treats this as
// ambiguous: the suggestion stays unconfirmed (confirmedByOperator=false),
// the cell id is recorded in applier.ambiguous, and applied stays empty.
func TestMappingOptionsMultiple(t *testing.T) {
	cell := "row-3-amount"

	// stubHeaderTextHeuristic: test-local stub proposing two equal-confidence
	// headers for one cell. The caller must not auto-confirm either.
	stubHeaderTextHeuristic := func(cell string) []spikeSuggestion {
		return []spikeSuggestion{
			{assoc: spikeAssoc{cell: cell, header: "amount"}, confidence: 0.7, approach: "heuristic"},
			{assoc: spikeAssoc{cell: cell, header: "total"}, confidence: 0.7, approach: "heuristic"},
		}
	}

	candidates := stubHeaderTextHeuristic(cell)
	if len(candidates) != 2 {
		t.Fatalf("heuristic returned %d candidates, want 2", len(candidates))
	}
	if candidates[0].confidence != candidates[1].confidence {
		t.Fatal("candidates do not have equal confidence")
	}

	var applier spikeApplier
	for _, s := range candidates {
		if s.confirmedByOperator {
			t.Fatalf("candidate %+v must not be operator-confirmed", s.assoc)
		}
		applier.apply(s)
	}
	// Caller records the ambiguous cell instead of applying a guess.
	applier.ambiguous = append(applier.ambiguous, cell)

	if len(applier.applied) != 0 {
		t.Fatalf("applied %d associations, want 0 (ambiguous must not apply)", len(applier.applied))
	}
	if len(applier.ambiguous) != 1 || applier.ambiguous[0] != cell {
		t.Fatalf("ambiguous = %v, want [%q]", applier.ambiguous, cell)
	}
}

// TestMappingOptionsConflicting: the fixture lookup says header H1 while the
// heuristic stub says H2. The caller records the conflict in
// applier.conflicting and applies neither; applied stays empty.
func TestMappingOptionsConflicting(t *testing.T) {
	cell := "row-4-amount"

	// stubFixtureLookup: fixture-authored association points at H1,
	// operator-confirmed and safe.
	stubFixtureLookup := func(cell string) (spikeSuggestion, bool) {
		return spikeSuggestion{
			assoc:               spikeAssoc{cell: cell, header: "amount"},
			confidence:          1.0,
			confirmedByOperator: true,
			unsafe:              false,
			approach:            "fixture-lookup",
		}, true
	}

	// stubHeaderTextHeuristic: secondary stub proposes H2 for the same cell.
	stubHeaderTextHeuristic := func(cell string) (spikeSuggestion, bool) {
		return spikeSuggestion{
			assoc:      spikeAssoc{cell: cell, header: "total"},
			confidence: 0.6,
			approach:   "heuristic",
		}, true
	}

	fromFixture, ok := stubFixtureLookup(cell)
	if !ok {
		t.Fatal("stubFixtureLookup returned no suggestion")
	}
	fromHeuristic, ok := stubHeaderTextHeuristic(cell)
	if !ok {
		t.Fatal("stubHeaderTextHeuristic returned no suggestion")
	}

	var applier spikeApplier
	if fromFixture.assoc.header != fromHeuristic.assoc.header {
		// Disagreement: caller records the conflicting cell instead of
		// applying either suggestion.
		applier.conflicting = append(applier.conflicting, cell)
	} else {
		applier.apply(fromFixture)
	}

	if len(applier.applied) != 0 {
		t.Fatalf("applied %d associations, want 0 (conflict must not apply)", len(applier.applied))
	}
	if len(applier.conflicting) != 1 || applier.conflicting[0] != cell {
		t.Fatalf("conflicting = %v, want [%q]", applier.conflicting, cell)
	}
}
