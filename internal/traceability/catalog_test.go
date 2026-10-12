package traceability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (Catalog, map[string]Issue) {
	t.Helper()
	root := filepath.Join("..", "..")
	catalog, err := LoadCatalog(filepath.Join(root, "docs", "traceability", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	issues, err := LoadIssues(filepath.Join(root, ".beads", "issues.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return catalog, issues
}

func cloneCatalog(t *testing.T, input Catalog) Catalog {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var output Catalog
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}
	return output
}

func cloneIssues(input map[string]Issue) map[string]Issue {
	output := make(map[string]Issue, len(input))
	for id, issue := range input {
		issue.Labels = append([]string(nil), issue.Labels...)
		issue.Dependencies = append([]Dependency(nil), issue.Dependencies...)
		output[id] = issue
	}
	return output
}

func TestCatalogValidates(t *testing.T) {
	catalog, issues := fixture(t)
	if problems := Validate(catalog, issues); len(problems) != 0 {
		t.Fatalf("unexpected validation errors: %s", strings.Join(problems, "; "))
	}
}

func TestLoadIssuesRejectsDuplicateBeadsIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "issues.jsonl")
	data := "{\"id\":\"dap-1\"}\n{\"id\":\"dap-1\"}\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIssues(path); err == nil || !strings.Contains(err.Error(), "duplicate Beads issue ID") {
		t.Fatalf("expected duplicate-ID error, got %v", err)
	}
}

func TestRejectsInvalidPlanningData(t *testing.T) {
	catalog, issues := fixture(t)
	tests := []struct {
		name   string
		mutate func(*Catalog, map[string]Issue)
		want   string
	}{
		{
			name:   "duplicate stable ID",
			mutate: func(c *Catalog, _ map[string]Issue) { c.Tests = append(c.Tests, c.Tests[0]) },
			want:   "duplicate catalog ID",
		},
		{
			name: "broken requirement link",
			mutate: func(c *Catalog, _ map[string]Issue) {
				c.Stories[0].Requirements = append(c.Stories[0].Requirements, "REQ-MISSING")
			},
			want: "unresolved requirement",
		},
		{
			name: "task parent mismatch",
			mutate: func(c *Catalog, beads map[string]Issue) {
				id := c.Tasks[0].ID
				issue := beads[id]
				for i := range issue.Dependencies {
					if issue.Dependencies[i].Type == "parent-child" {
						issue.Dependencies[i].DependsOnID = "dap-jo9"
						break
					}
				}
				beads[id] = issue
			},
			want: "Beads parent does not match",
		},
		{
			name: "stale Beads acceptance ID",
			mutate: func(_ *Catalog, beads map[string]Issue) {
				issue := beads["dap-jo9.2.3"]
				issue.AcceptanceCriteria += " TC-NOT-IN-CATALOG-001"
				beads[issue.ID] = issue
			},
			want: "unknown acceptance case",
		},
		{
			name: "missing dependency endpoint",
			mutate: func(_ *Catalog, beads map[string]Issue) {
				issue := beads["dap-jo9.2"]
				issue.Dependencies = append(issue.Dependencies, Dependency{IssueID: issue.ID, DependsOnID: "dap-missing", Type: "blocks"})
				beads[issue.ID] = issue
			},
			want: "dependency endpoint missing",
		},
		{
			name: "missing family coverage",
			mutate: func(c *Catalog, _ map[string]Issue) {
				filtered := c.Tests[:0]
				for _, item := range c.Tests {
					if item.Lane != "generation" || item.Family != "reports" {
						filtered = append(filtered, item)
					}
				}
				c.Tests = filtered
			},
			want: "missing lane/family case coverage",
		},
		{
			name:   "missing Gherkin clause",
			mutate: func(c *Catalog, _ map[string]Issue) { c.Tests[0].Then = " " },
			want:   "missing THEN clause",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			copyCatalog, copyIssues := cloneCatalog(t, catalog), cloneIssues(issues)
			test.mutate(&copyCatalog, copyIssues)
			problems := Validate(copyCatalog, copyIssues)
			if !containsString(problems, test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, problems)
			}
		})
	}
}

func TestRejectsDependencyCycle(t *testing.T) {
	_, issues := fixture(t)
	issues = cloneIssues(issues)
	root := issues["dap-jo9"]
	root.Dependencies = append(root.Dependencies, Dependency{IssueID: root.ID, DependsOnID: "dap-jo9.2", Type: "blocks"})
	issues[root.ID] = root
	milestone := issues["dap-jo9.2"]
	milestone.Dependencies = append(milestone.Dependencies, Dependency{IssueID: milestone.ID, DependsOnID: "dap-jo9", Type: "blocks"})
	issues[milestone.ID] = milestone
	if problems := validateDependencies(issues); !containsString(problems, "dependency cycle") {
		t.Fatalf("expected cycle error, got %v", problems)
	}
}

func TestGeneratedViewsHavePlanningBoundaries(t *testing.T) {
	catalog, _ := fixture(t)
	views := GeneratedMarkdown(catalog)
	cases := string(views[filepath.Join("docs", "validation", "cases.md")])
	matrix := string(views[filepath.Join("docs", "traceability", "matrix.md")])
	if !strings.Contains(cases, "specifications only; none of these product cases have been executed") {
		t.Fatal("case view does not mark cases as specifications")
	}
	if !strings.Contains(matrix, "Definition of Ready") || !strings.Contains(matrix, "Definition of Done") {
		t.Fatal("matrix is missing readiness definitions")
	}
}

func TestGeneratedViewsMatchRepository(t *testing.T) {
	catalog, _ := fixture(t)
	for name, expected := range GeneratedMarkdown(catalog) {
		path := filepath.Join("..", "..", name)
		actual, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(actual) != string(expected) {
			t.Errorf("generated view %s is stale; run go run ./cmd/traceability-check --write", name)
		}
	}
}

func containsString(values []string, fragment string) bool {
	for _, value := range values {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}
