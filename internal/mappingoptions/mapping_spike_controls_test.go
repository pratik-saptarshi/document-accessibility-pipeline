// Synthetic exploration controls for beads dap-jo9.1.6.7 (RALPH-022 spike).
//
// These tests assert the SPIKE's guardrails for future financial-table
// mapping options, NOT mapping inference. No production writer, no
// automatic mapping, and no semantic/AT authority is implemented or
// exercised here. Every control asserts that an unconfirmed suggestion is
// never applied, and that unsafe associations are never applied at all.
package mappingoptions

import "testing"

// fixtureExpected mirrors the fixture-authored row-two /Headers associations
// recorded in Decision 0005 (docs/decisions/0005-*.md): the statement's
// "$42.10" cell expects amountHeader, the invoice's "$8.00" cell expects
// unitAmountHeader.
func fixtureExpected(t *testing.T, family string) Association {
	t.Helper()
	switch family {
	case "statement":
		return Association{Family: "statement", Cell: "$42.10", SuggestedHeader: "amountHeader", Signal: "fixture-authored"}
	case "invoice":
		return Association{Family: "invoice", Cell: "$8.00", SuggestedHeader: "unitAmountHeader", Signal: "fixture-authored"}
	default:
		t.Fatalf("unknown synthetic family %q", family)
		return Association{}
	}
}

// Control 1 (accuracy): a suggestion matching the fixture-authored expected
// association is recognized as a suggestion — and is NOT applied.
func TestSpikeAccuracyExpectedAssociationRecognizedNotApplied(t *testing.T) {
	for _, family := range []string{"statement", "invoice"} {
		expected := fixtureExpected(t, family)
		got := Evaluate([]Association{expected})
		if got.Status != StatusSuggested {
			t.Fatalf("family %q: expected status %q, got %q (%s)", family, StatusSuggested, got.Status, got.Detail)
		}
		if got.Suggestion.SuggestedHeader != expected.SuggestedHeader {
			t.Fatalf("family %q: suggestion header = %q, want %q", family, got.Suggestion.SuggestedHeader, expected.SuggestedHeader)
		}
		if got.Applied {
			t.Fatalf("family %q: unconfirmed suggestion was applied", family)
		}
		if err := Apply(got, OperatorConfirmation{}); err == nil {
			t.Fatalf("family %q: Apply without operator confirmation succeeded", family)
		}
	}
}

// Control 2 (missing): absent header data produces no suggestion at all.
func TestSpikeMissingTableDataYieldsNoSuggestion(t *testing.T) {
	got := Evaluate(nil)
	if got.Status != StatusMissing {
		t.Fatalf("expected status %q, got %q (%s)", StatusMissing, got.Status, got.Detail)
	}
	if got.Applied {
		t.Fatal("missing-input control was applied")
	}
	if err := Apply(got, OperatorConfirmation{Family: "statement", Cell: "$42.10", Header: "amountHeader"}); err == nil {
		t.Fatal("Apply on missing input succeeded")
	}
}

// Control 3 (multiple): multiple equally-matching candidates surface
// ambiguity instead of a silent pick; the ambiguous suggestion is not applied
// until an operator confirms one exact candidate.
func TestSpikeMultipleCandidatesSurfaceAmbiguityNotSilentPick(t *testing.T) {
	candidates := []Association{
		{Family: "statement", Cell: "$42.10", SuggestedHeader: "amountHeader", Signal: "exact-header-string"},
		{Family: "statement", Cell: "$42.10", SuggestedHeader: "totalHeader", Signal: "column-geometry"},
	}
	got := Evaluate(candidates)
	if got.Status != StatusAmbiguous {
		t.Fatalf("expected status %q, got %q (%s)", StatusAmbiguous, got.Status, got.Detail)
	}
	if got.Applied {
		t.Fatal("ambiguous control was applied")
	}
	// Unconfirmed: no single suggestion is surfaced and Apply refuses.
	if got.Suggestion.SuggestedHeader != "" {
		t.Fatalf("ambiguous control surfaced a silent pick: %q", got.Suggestion.SuggestedHeader)
	}
	if err := Apply(got, OperatorConfirmation{}); err == nil {
		t.Fatal("Apply on ambiguous input without confirmation succeeded")
	}
	// Operator-confirmed baseline (Decision 0003): an operator confirming one
	// exact candidate resolves the ambiguity.
	confirmed := OperatorConfirmation{Family: "statement", Cell: "$42.10", Header: "amountHeader"}
	resolved, err := ResolveWithConfirmation(candidates, confirmed)
	if err != nil {
		t.Fatalf("operator confirmation of an exact candidate failed: %v", err)
	}
	if !resolved.Applied || resolved.Suggestion.SuggestedHeader != "amountHeader" {
		t.Fatalf("operator-confirmed resolution did not apply the confirmed association: %+v", resolved)
	}
}

// Control 4 (conflicting): signals that disagree (e.g. header string says
// one header, structure-tree reuse says another) surface conflict and no
// suggestion is applied; there is nothing to confirm.
func TestSpikeConflictingSignalsYieldNoSuggestion(t *testing.T) {
	candidates := []Association{
		{Family: "invoice", Cell: "$8.00", SuggestedHeader: "unitAmountHeader", Signal: "exact-header-string"},
		{Family: "invoice", Cell: "$8.00", SuggestedHeader: "quantityHeader", Signal: "structure-tree-reuse"},
	}
	got := EvaluateConflicting(candidates)
	if got.Status != StatusConflicting {
		t.Fatalf("expected status %q, got %q (%s)", StatusConflicting, got.Status, got.Detail)
	}
	if got.Suggestion.SuggestedHeader != "" {
		t.Fatalf("conflicting control surfaced a suggestion: %q", got.Suggestion.SuggestedHeader)
	}
	if got.Applied {
		t.Fatal("conflicting control was applied")
	}
	if err := Apply(got, OperatorConfirmation{Family: "invoice", Cell: "$8.00", Header: "unitAmountHeader"}); err == nil {
		t.Fatal("Apply on conflicting input succeeded despite operator confirmation")
	}
}

// Control 5 (unsafe): an association flagged unsafe (it would alter displayed
// text or reorder rows) is blocked; even operator confirmation cannot apply
// it, because the operator-confirmed workflow (Decision 0003) preserves exact
// displayed values.
func TestSpikeUnsafeAssociationNeverApplied(t *testing.T) {
	unsafe := Association{Family: "statement", Cell: "$42.10", SuggestedHeader: "amountHeader", Signal: "column-geometry", Unsafe: true}
	got := Evaluate([]Association{unsafe})
	if got.Status != StatusUnsafeBlocked {
		t.Fatalf("expected status %q, got %q (%s)", StatusUnsafeBlocked, got.Status, got.Detail)
	}
	if got.Applied {
		t.Fatal("unsafe control was applied")
	}
	if err := Apply(got, OperatorConfirmation{Family: "statement", Cell: "$42.10", Header: "amountHeader"}); err == nil {
		t.Fatal("Apply on unsafe association succeeded despite operator confirmation")
	}
}

// Guardrail: an unconfirmed suggested association can never be applied,
// regardless of which family or signal produced it.
func TestSpikeNoUnconfirmedSuggestionEverApplied(t *testing.T) {
	cases := []struct {
		name  string
		eval  Evaluation
		apply OperatorConfirmation
	}{
		{"suggested no confirmation", Evaluate([]Association{fixtureExpected(t, "statement")}), OperatorConfirmation{}},
		{"suggested mismatched confirmation", Evaluate([]Association{fixtureExpected(t, "invoice")}), OperatorConfirmation{Family: "invoice", Cell: "$8.00", Header: "wrongHeader"}},
		{"missing", Evaluate(nil), OperatorConfirmation{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.eval.Applied {
				t.Fatal("evaluation reported Applied without confirmation")
			}
			if err := Apply(tc.eval, tc.apply); err == nil {
				t.Fatal("Apply succeeded without a valid operator confirmation")
			}
		})
	}
}
