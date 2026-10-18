// Test-only harness: gate behavior against fixture-authored associations.
// Scope: no PDF parsing, no suggester implementation, no automatic mapping, no
// mapping-accuracy measurement. These tests verify the existing gate semantics
// (confirmedByOperator && !unsafe → applied, reusing spikeApplier from
// mapping_options_spike_test.go) against hand-authored fixture associations.
// Ambiguity/conflict/unsafe negative controls in this file are constructed
// test controls, not fixture evidence; they are marked as such at each site.

package financialfixtures

import "testing"

// fixtureAssoc is a hand-authored cell-to-header association. It is test data,
// not fixture metadata, not parsed output, and not an inference.
type fixtureAssoc struct {
	fixture    string
	cell       string
	header     string
	confidence float64
	unsafe     bool
}

// fixtureAuthoredManifest holds every hand-authored association for these
// tests. The entries below were typed by hand from the human-readable
// observation records; they were NOT parsed from PDFs and NOT inferred by any
// suggester. Each entry cites its evidence source in a comment.
var fixtureAuthoredManifest = []fixtureAssoc{
	// Evidence: docs/fixtures/remediation/financial-tables-v2/observations/statement-positive-v2.txt
	// (synthetic statement sample; headers Date | Description | Amount; data
	// row 2 = 2026-09-04 / City Transit / $42.10).
	{fixture: "statement", cell: "stmt-row2-amount", header: "Amount", confidence: 0.95},
	// Evidence: docs/fixtures/remediation/financial-tables-v2/observations/invoice-positive-v2.txt
	// (synthetic invoice sample; headers Item | Quantity | Unit amount; data
	// row 2 = Envelopes / 1 / $8.00).
	{fixture: "invoice", cell: "inv-row2-unit-amount", header: "Unit amount", confidence: 0.95},
}

// fixtureAssocsFor returns every manifest entry whose (fixture, cell) key
// exactly equals the requested key. Exact match only: no normalization, no
// prefix matching, no fallthrough.
func fixtureAssocsFor(fixture, cell string) []fixtureAssoc {
	var out []fixtureAssoc
	for _, entry := range fixtureAuthoredManifest {
		if entry.fixture == fixture && entry.cell == cell {
			out = append(out, entry)
		}
	}
	return out
}

// asUnconfirmedSuggestion converts an authored entry into a gate input that
// no operator has confirmed.
func (e fixtureAssoc) asUnconfirmedSuggestion() spikeSuggestion {
	return spikeSuggestion{
		assoc:      spikeAssoc{cell: e.cell, header: e.header},
		confidence: e.confidence,
		unsafe:     e.unsafe,
		approach:   "fixture-authored",
	}
}

// authoredHeadersDisagree reports whether the given authored entries for one
// cell assign different headers.
func authoredHeadersDisagree(entries []fixtureAssoc) bool {
	for _, entry := range entries[1:] {
		if entry.header != entries[0].header {
			return true
		}
	}
	return false
}

// equalConfidence reports whether every candidate carries the same confidence.
func equalConfidence(candidates []spikeSuggestion) bool {
	for _, c := range candidates[1:] {
		if c.confidence != candidates[0].confidence {
			return false
		}
	}
	return true
}

// fixtureDriver feeds authored entries and candidate suggestions through the
// existing gate semantics of spikeApplier. It never confirms a suggestion
// itself; confirmation always comes from the test as a simulated operator.
type fixtureDriver struct {
	gate           spikeApplier
	missing        []string // cells with no authored entry; nothing reached the gate
	contradictions []string // cells where a confirmed suggestion contradicted the authored entry
}

// resolveAuthored drives the manifest path for one cell: missing,
// contradictory authored entries, or a single authored entry against an
// optional confirmed suggestion.
func (d *fixtureDriver) resolveAuthored(cell string, entries []fixtureAssoc, confirmed spikeSuggestion, haveConfirmed bool) {
	switch {
	case len(entries) == 0:
		// No authored entry: no suggestion is constructed and the gate is
		// never touched for this cell.
		d.missing = append(d.missing, cell)
	case authoredHeadersDisagree(entries):
		// Authored entries contradict each other: flag the cell, apply nothing.
		d.gate.conflicting = append(d.gate.conflicting, cell)
	default:
		authored := entries[0]
		if !haveConfirmed {
			d.gate.apply(authored.asUnconfirmedSuggestion())
			return
		}
		if confirmed.assoc.header != authored.header {
			// A confirmed suggestion that contradicts the authored entry is
			// recorded as a contradiction and never applied.
			d.contradictions = append(d.contradictions, cell)
			return
		}
		suggestion := confirmed
		// An authored entry flagged unsafe marks the cell as unsafe: a
		// confirmed suggestion for that cell is rejected by the gate.
		suggestion.unsafe = suggestion.unsafe || authored.unsafe
		d.gate.apply(suggestion)
	}
}

// resolveCell is the manifest-backed convenience wrapper around
// resolveAuthored.
func (d *fixtureDriver) resolveCell(fixture, cell string, confirmed spikeSuggestion, haveConfirmed bool) {
	d.resolveAuthored(cell, fixtureAssocsFor(fixture, cell), confirmed, haveConfirmed)
}

// resolveCandidates drives the candidate path: multiple equal-confidence
// candidates are held unconfirmed and the cell is flagged ambiguous; nothing
// is applied.
func (d *fixtureDriver) resolveCandidates(cell string, candidates []spikeSuggestion) {
	for _, candidate := range candidates {
		if candidate.confirmedByOperator {
			panic("fixtureDriver never receives confirmed candidates; the driver confirms nothing itself")
		}
		d.gate.apply(candidate)
	}
	if len(candidates) > 1 && equalConfidence(candidates) {
		d.gate.ambiguous = append(d.gate.ambiguous, cell)
	}
}

func TestFixtureAuthoredAccuracyMatchApplies(t *testing.T) {
	var d fixtureDriver
	d.resolveCell("statement", "stmt-row2-amount",
		spikeSuggestion{
			assoc:               spikeAssoc{cell: "stmt-row2-amount", header: "Amount"},
			confidence:          0.95,
			confirmedByOperator: true,
			approach:            "fixture-authored",
		}, true)
	if len(d.gate.applied) != 1 {
		t.Fatalf("gate applied %d, want 1", len(d.gate.applied))
	}
	want := spikeAssoc{cell: "stmt-row2-amount", header: "Amount"}
	if d.gate.applied[0] != want {
		t.Fatalf("gate applied %+v, want %+v", d.gate.applied[0], want)
	}
}

func TestFixtureAuthoredContradictionNotApplied(t *testing.T) {
	var d fixtureDriver
	d.resolveCell("statement", "stmt-row2-amount",
		spikeSuggestion{
			assoc:               spikeAssoc{cell: "stmt-row2-amount", header: "Description"},
			confidence:          0.9,
			confirmedByOperator: true,
			approach:            "fixture-authored",
		}, true)
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied for a contradiction", d.gate.applied)
	}
	if len(d.contradictions) != 1 || d.contradictions[0] != "stmt-row2-amount" {
		t.Fatalf("contradictions = %v, want [stmt-row2-amount]", d.contradictions)
	}
}

func TestFixtureAuthoredMissingCellNoSuggestion(t *testing.T) {
	var d fixtureDriver
	d.resolveCell("statement", "stmt-row99-amount", spikeSuggestion{}, false)
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied for a missing cell", d.gate.applied)
	}
	if len(d.missing) != 1 || d.missing[0] != "stmt-row99-amount" {
		t.Fatalf("missing = %v, want [stmt-row99-amount]", d.missing)
	}
}

func TestFixtureAuthoredAmbiguousCandidatesHeld(t *testing.T) {
	// Constructed control: not fixture evidence. Two equal-confidence
	// candidates for a synthetic cell, kept deliberately unconfirmed.
	cell := "ctrl-ambiguous-cell"
	candidates := []spikeSuggestion{
		{assoc: spikeAssoc{cell: cell, header: "Amount"}, confidence: 0.7, approach: "control"},
		{assoc: spikeAssoc{cell: cell, header: "Description"}, confidence: 0.7, approach: "control"},
	}
	var d fixtureDriver
	d.resolveCandidates(cell, candidates)
	for _, candidate := range candidates {
		if candidate.confirmedByOperator {
			t.Fatal("ambiguous candidate must stay unconfirmed")
		}
	}
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied for ambiguous candidates", d.gate.applied)
	}
	if len(d.gate.ambiguous) != 1 || d.gate.ambiguous[0] != cell {
		t.Fatalf("ambiguous = %v, want [%s]", d.gate.ambiguous, cell)
	}
}

func TestFixtureAuthoredConflictingEntriesUnapplied(t *testing.T) {
	// Constructed control: not fixture evidence. Two authored entries for one
	// synthetic cell that assign different headers with different confidence.
	entries := []fixtureAssoc{
		{fixture: "control", cell: "ctrl-conflict-cell", header: "Item", confidence: 0.9},
		{fixture: "control", cell: "ctrl-conflict-cell", header: "Quantity", confidence: 0.6},
	}
	var d fixtureDriver
	d.resolveAuthored("ctrl-conflict-cell", entries, spikeSuggestion{}, false)
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied for conflicting authored entries", d.gate.applied)
	}
	if len(d.gate.conflicting) != 1 || d.gate.conflicting[0] != "ctrl-conflict-cell" {
		t.Fatalf("conflicting = %v, want [ctrl-conflict-cell]", d.gate.conflicting)
	}
}

func TestFixtureAuthoredUnsafeEntryRejected(t *testing.T) {
	// Constructed control: not fixture evidence. The authored entry flags the
	// cell unsafe, so even a confirmed matching suggestion must be rejected.
	entries := []fixtureAssoc{
		{fixture: "control", cell: "ctrl-unsafe-cell", header: "Amount", confidence: 0.9, unsafe: true},
	}
	var d fixtureDriver
	d.resolveAuthored("ctrl-unsafe-cell", entries,
		spikeSuggestion{
			assoc:               spikeAssoc{cell: "ctrl-unsafe-cell", header: "Amount"},
			confidence:          0.9,
			confirmedByOperator: true,
			approach:            "control",
		}, true)
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want unsafe entry rejected even when confirmed", d.gate.applied)
	}
}

func TestFixtureAuthoredUnconfirmedNeverApplies(t *testing.T) {
	var d fixtureDriver
	d.resolveCell("statement", "stmt-row2-amount", spikeSuggestion{}, false)
	d.resolveCell("invoice", "inv-row2-unit-amount", spikeSuggestion{}, false)
	if len(d.gate.applied) != 0 {
		t.Fatalf("gate applied %+v, want nothing applied without operator confirmation", d.gate.applied)
	}
}
