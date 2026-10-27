package beadsreconcile

import (
	"encoding/json"
	"strings"
	"testing"
)

// Mutant-killing tests for the create-record validation contract.
//
// The normative contract (Plan documentation in reconcile.go and the
// creation acceptance criteria) bounds creation to exactly one complete
// canonical row for a synthetic direct child of dap-jo9: it cannot be
// combined with updates, its derived counters must be zero, its priority
// must be in 0..4, and it must carry exactly one six-field parent-child
// edge to dap-jo9. Each test below was verified to FAIL by assertion
// against a production mutation that weakens the named check, and to
// PASS against the unmutated code. Earlier panel "kills" of these
// mutants were compile errors (unused variables), not assertion
// failures, so the mutants survived.

func mutantCreateFields(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(syntheticCreate(t), &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}

func remarshalCreate(t *testing.T, fields map[string]json.RawMessage) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func editCreateEdges(t *testing.T, fields map[string]json.RawMessage, edit func([]map[string]json.RawMessage) []map[string]json.RawMessage) {
	t.Helper()
	var edges []map[string]json.RawMessage
	if err := json.Unmarshal(fields["dependencies"], &edges); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(edit(edges))
	if err != nil {
		t.Fatal(err)
	}
	fields["dependencies"] = data
}

// TestMutantCreateCombinedWithUpdatesRejected kills the mutant that
// removes the creation/update combination ban from validatePlan.
func TestMutantCreateCombinedWithUpdatesRejected(t *testing.T) {
	plan, source, _ := syntheticPlan(t)
	if err := ValidatePlan(source, plan); err != nil {
		t.Fatalf("create-only control plan rejected: %v", err)
	}
	combined := plan
	combined.Updates = []IssueUpdate{{ID: "dap-jo9", Fields: map[string]string{"title": "also update"}}}
	if err := ValidatePlan(source, combined); err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("ValidatePlan(create+update) error = %v, want combination rejection", err)
	}
}

// TestMutantCreateNonZeroCountersRejected kills the mutant that drops
// the requirement that a created row's dependency/dependent/comment
// counters are all zero (a fresh row has no graph or comment history).
func TestMutantCreateNonZeroCountersRejected(t *testing.T) {
	for _, field := range []string{"dependency_count", "dependent_count", "comment_count"} {
		t.Run(field, func(t *testing.T) {
			fields := mutantCreateFields(t)
			fields[field] = json.RawMessage(`1`)
			if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), field) {
				t.Fatalf("validateCreateRecord() error = %v, want invalid %s rejection", err, field)
			}
			plan := Plan{Creates: []json.RawMessage{remarshalCreate(t, fields)}}
			if err := ValidatePlan([]byte(syntheticParent), plan); err == nil {
				t.Fatal("ValidatePlan() accepted a create with a non-zero derived counter")
			}
		})
	}
}

// TestMutantCreateOutOfRangePriorityRejected kills the mutant that
// widens the priority bound beyond the canonical 0..4 range.
func TestMutantCreateOutOfRangePriorityRejected(t *testing.T) {
	fields := mutantCreateFields(t)
	fields["priority"] = json.RawMessage(`5`)
	if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "priority") {
		t.Fatalf("validateCreateRecord() error = %v, want invalid priority rejection", err)
	}
	for _, okValue := range []string{"0", "4"} {
		control := mutantCreateFields(t)
		control["priority"] = json.RawMessage(okValue)
		if err := validateCreateRecord("dap-jo9.99", control); err != nil {
			t.Fatalf("boundary priority %s rejected: %v", okValue, err)
		}
	}
}

// TestMutantCreateEdgeShapeRejected kills the mutants that weaken the
// dependency-edge length checks: exactly one edge, with exactly the six
// canonical edge fields.
func TestMutantCreateEdgeShapeRejected(t *testing.T) {
	t.Run("two edges", func(t *testing.T) {
		fields := mutantCreateFields(t)
		editCreateEdges(t, fields, func(edges []map[string]json.RawMessage) []map[string]json.RawMessage {
			return append(edges, edges[0])
		})
		if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "exactly one") {
			t.Fatalf("validateCreateRecord() error = %v, want exactly-one-edge rejection", err)
		}
	})
	t.Run("extra edge field", func(t *testing.T) {
		fields := mutantCreateFields(t)
		editCreateEdges(t, fields, func(edges []map[string]json.RawMessage) []map[string]json.RawMessage {
			edges[0]["unexpected"] = json.RawMessage(`"field"`)
			return edges
		})
		if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "exact direct parent-child edge") {
			t.Fatalf("validateCreateRecord() error = %v, want exact-edge rejection", err)
		}
	})
}

// TestMutantCreateWrongParentRejected kills the mutant that drops the
// requirement that the create's sole edge points at dap-jo9 itself.
func TestMutantCreateWrongParentRejected(t *testing.T) {
	fields := mutantCreateFields(t)
	editCreateEdges(t, fields, func(edges []map[string]json.RawMessage) []map[string]json.RawMessage {
		edges[0]["depends_on_id"] = json.RawMessage(`"dap-other"`)
		return edges
	})
	if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "exact direct parent-child edge") {
		t.Fatalf("validateCreateRecord() error = %v, want exact-parent rejection", err)
	}
	plan := Plan{Creates: []json.RawMessage{remarshalCreate(t, fields)}}
	if err := ValidatePlan([]byte(syntheticParent), plan); err == nil {
		t.Fatal("ValidatePlan() accepted a create parented outside dap-jo9")
	}
}

// TestMutantCreateWrongEdgeTypeRejected kills the mutant that drops
// the parent-child edge-type requirement.
func TestMutantCreateWrongEdgeTypeRejected(t *testing.T) {
	fields := mutantCreateFields(t)
	editCreateEdges(t, fields, func(edges []map[string]json.RawMessage) []map[string]json.RawMessage {
		edges[0]["type"] = json.RawMessage(`"blocks"`)
		return edges
	})
	if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "exact direct parent-child edge") {
		t.Fatalf("validateCreateRecord() error = %v, want parent-child type rejection", err)
	}
}

// TestMutantCreateWrongRecordTypeRejected kills the mutant that drops
// the requirement that a created row is an issue record (_type issue).
func TestMutantCreateWrongRecordTypeRejected(t *testing.T) {
	fields := mutantCreateFields(t)
	fields["_type"] = json.RawMessage(`"epic"`)
	if err := validateCreateRecord("dap-jo9.99", fields); err == nil || !strings.Contains(err.Error(), "_type") {
		t.Fatalf("validateCreateRecord() error = %v, want _type issue rejection", err)
	}
}
