package beadsreconcile

import (
	"strings"
	"testing"
)

func TestUpdateOnlyReconcileRejectsAddedImporterRecords(t *testing.T) {
	source := []byte("{\"id\":\"dap-jo9\",\"title\":\"parent\"}\n")
	extra := []byte("{\"id\":\"dap-jo9.99\",\"title\":\"unplanned child\"}\n")

	tests := []struct {
		name     string
		baseline []byte
		updated  []byte
	}{
		{name: "extra record in baseline", baseline: append(append([]byte(nil), source...), extra...), updated: append(append([]byte(nil), source...), extra...)},
		{name: "extra record after update", baseline: source, updated: append(append([]byte(nil), source...), extra...)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Reconcile(source, tc.baseline, tc.updated, Plan{}); err == nil || !strings.Contains(err.Error(), "issue ID set changed") {
				t.Fatalf("Reconcile() error = %v, want exact source/import ID-set rejection", err)
			}
		})
	}
}
