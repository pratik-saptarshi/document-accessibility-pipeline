// Package mappingoptions is an EXPLORATION harness for beads dap-jo9.1.6.7
// (RALPH-022 spike): "Explore financial-table mapping options".
//
// It models candidate mapping suggestions and the operator-confirmation
// guardrail policy against synthetic controls. It implements NO inference,
// NO automatic mapping, NO production writer, and NO semantic/AT reviewer
// authority. Per Decision 0003
// (docs/decisions/0003-operator-confirmed-financial-table-mapping.md), the
// only way any association may be applied is explicit operator confirmation
// of the exact intended association, and even that cannot authorize an
// unsafe change (exact displayed values must be preserved).
package mappingoptions

import "errors"

// Evaluation statuses for synthetic controls.
const (
	// StatusSuggested: a single candidate matches the fixture-authored
	// expected association. It is surfaced as a suggestion only.
	StatusSuggested = "suggested"
	// StatusMissing: no table or header data; no suggestion exists.
	StatusMissing = "missing"
	// StatusAmbiguous: multiple candidates match; no silent pick is made.
	StatusAmbiguous = "ambiguous"
	// StatusConflicting: signals disagree; no suggestion is surfaced.
	StatusConflicting = "conflicting"
	// StatusUnsafeBlocked: the association is unsafe; it can never be applied.
	StatusUnsafeBlocked = "unsafe-blocked"
)

// Association is a single synthetic mapping suggestion: the fixture cell
// that an option's signal would associate with a header. Unsafe marks an
// association that would alter displayed text or reorder rows.
type Association struct {
	Family          string
	Cell            string
	SuggestedHeader string
	Signal          string
	Unsafe          bool
}

// Evaluation is the harness outcome for a set of candidate suggestions.
// Applied is true only after a valid operator confirmation; the harness
// never applies a suggestion on its own.
type Evaluation struct {
	Status     string
	Suggestion Association
	Applied    bool
	Detail     string
}

// OperatorConfirmation records an operator naming the exact intended
// association before any change. Per Decision 0003 it is the only
// confirmation boundary, and it grants no semantic/AT or release authority.
type OperatorConfirmation struct {
	Family string
	Cell   string
	Header string
}

// valid reports whether the confirmation names the evaluation's suggested
// association exactly.
func (c OperatorConfirmation) validFor(e Evaluation) bool {
	if c.Family == "" || c.Cell == "" || c.Header == "" {
		return false
	}
	return c.Family == e.Suggestion.Family &&
		c.Cell == e.Suggestion.Cell &&
		c.Header == e.Suggestion.SuggestedHeader
}

// Evaluate classifies a set of suggestion candidates from one option's
// signals. It surfaces at most one suggestion and never applies anything.
func Evaluate(candidates []Association) Evaluation {
	if len(candidates) == 0 {
		return Evaluation{Status: StatusMissing, Detail: "no table or header data; no suggestion exists"}
	}
	for _, c := range candidates {
		if c.Unsafe {
			return Evaluation{Status: StatusUnsafeBlocked, Detail: "association would alter displayed text or reorder rows; blocked"}
		}
	}
	if len(candidates) > 1 {
		distinct := map[string]bool{}
		for _, c := range candidates {
			distinct[c.SuggestedHeader] = true
		}
		if len(distinct) > 1 {
			return Evaluation{Status: StatusAmbiguous, Detail: "multiple candidate associations; no silent pick"}
		}
	}
	return Evaluation{Status: StatusSuggested, Suggestion: candidates[0], Detail: "single candidate suggestion; unconfirmed"}
}

// EvaluateConflicting classifies candidates whose signals disagree (each
// signal proposes a different header). No suggestion is surfaced and there
// is nothing an operator confirmation can apply.
func EvaluateConflicting(candidates []Association) Evaluation {
	if len(candidates) == 0 {
		return Evaluation{Status: StatusMissing, Detail: "no table or header data; no suggestion exists"}
	}
	return Evaluation{Status: StatusConflicting, Detail: "signals disagree; no suggestion surfaced"}
}

// ResolveWithConfirmation resolves an ambiguous candidate set by applying
// the exact candidate the operator confirmed. It returns an error unless
// the confirmation names one of the candidates exactly.
func ResolveWithConfirmation(candidates []Association, confirmation OperatorConfirmation) (Evaluation, error) {
	for _, c := range candidates {
		if c.Unsafe {
			return Evaluation{}, errors.New("mappingoptions: unsafe candidate cannot be confirmed")
		}
		if c.Family == confirmation.Family && c.Cell == confirmation.Cell && c.SuggestedHeader == confirmation.Header && confirmation.Header != "" {
			return Evaluation{Status: StatusSuggested, Suggestion: c, Applied: true, Detail: "operator confirmed the exact intended association"}, nil
		}
	}
	return Evaluation{}, errors.New("mappingoptions: confirmation does not name an exact candidate")
}

// Apply is the guardrail: a suggestion becomes applied only when the
// evaluation holds a suggestion and the operator confirmation names that
// exact association. Unsafe, missing, and conflicting evaluations can never
// be applied, even with a confirmation.
func Apply(e Evaluation, confirmation OperatorConfirmation) error {
	switch e.Status {
	case StatusSuggested:
		if !confirmation.validFor(e) {
			return errors.New("mappingoptions: no valid operator confirmation; unconfirmed suggestion not applied")
		}
		return nil
	case StatusUnsafeBlocked:
		return errors.New("mappingoptions: unsafe association blocked; confirmation cannot authorize it")
	default:
		return errors.New("mappingoptions: evaluation status " + e.Status + " has no applicable suggestion")
	}
}
