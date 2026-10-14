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

func TestStandaloneTaskCanRemainStoryLinkedWithoutParentEdge(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.9"
	found := false
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "Early candidate review proceeds independently of the later pipeline story."
			if catalog.Tasks[i].Story != "US-REM-002" {
				t.Fatalf("standalone task story = %q, want US-REM-002", catalog.Tasks[i].Story)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("task %s not found", taskID)
	}
	issue := issues[taskID]
	filtered := issue.Dependencies[:0]
	for _, edge := range issue.Dependencies {
		if edge.Type != "parent-child" {
			filtered = append(filtered, edge)
		}
	}
	issue.Dependencies = filtered
	issues[taskID] = issue
	if problems := Validate(catalog, issues); len(problems) != 0 {
		t.Fatalf("standalone story-linked task should validate without a Beads parent edge: %s", strings.Join(problems, "; "))
	}
}

func TestStandaloneSpecificationCanRemainStoryLinkedWithoutParentEdge(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.2"
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "Review specifications can be prepared separately from actual candidate review acceptance."
		}
	}
	issue := issues[taskID]
	filtered := issue.Dependencies[:0]
	for _, edge := range issue.Dependencies {
		if edge.Type != "parent-child" {
			filtered = append(filtered, edge)
		}
	}
	issue.Dependencies = filtered
	issues[taskID] = issue
	if problems := Validate(catalog, issues); len(problems) != 0 {
		t.Fatalf("review specification task should validate as the specifically approved standalone mapping: %s", strings.Join(problems, "; "))
	}
}

func TestStandaloneTaskRequiresReason(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.9"
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "  "
		}
	}
	if problems := Validate(catalog, issues); !containsString(problems, taskID+": standalone reason must not be blank") {
		t.Fatalf("expected standalone reason diagnostic, got %v", problems)
	}
}

func TestStandaloneTaskRejectsBeadsParentEdge(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.9"
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "Early candidate review is a separately gated story increment."
		}
	}
	issue := issues[taskID]
	issue.Dependencies = append(issue.Dependencies, Dependency{IssueID: taskID, DependsOnID: "dap-jo9.2.5", Type: "parent-child"})
	issues[taskID] = issue
	if problems := Validate(catalog, issues); !containsString(problems, taskID+": standalone task must not have a Beads parent edge") {
		t.Fatalf("expected standalone parent-edge diagnostic, got %v", problems)
	}
}

func TestStandaloneReviewTaskRequiresExactBlockingPrerequisites(t *testing.T) {
	const taskID = "dap-jo9.2.5.9"
	const diagnostic = taskID + ": blocking dependencies must be exactly dap-jo9.1.1.4, dap-jo9.1.6.4, and dap-jo9.2.5.2"
	tests := []struct {
		name   string
		mutate func(*Issue)
	}{
		{
			name: "missing human and AT authority gate",
			mutate: func(issue *Issue) {
				issue.Dependencies = removeBlockingDependency(issue.Dependencies, "dap-jo9.1.1.4")
			},
		},
		{
			name: "missing candidate feasibility gate",
			mutate: func(issue *Issue) {
				issue.Dependencies = removeBlockingDependency(issue.Dependencies, "dap-jo9.1.6.4")
			},
		},
		{
			name: "missing specification gate",
			mutate: func(issue *Issue) {
				issue.Dependencies = removeBlockingDependency(issue.Dependencies, "dap-jo9.2.5.2")
			},
		},
		{
			name: "added full remediation task blocker",
			mutate: func(issue *Issue) {
				issue.Dependencies = append(issue.Dependencies, Dependency{IssueID: taskID, DependsOnID: "dap-jo9.2.3.2", Type: "blocks"})
			},
		},
		{
			name: "added redundant administrative blocker",
			mutate: func(issue *Issue) {
				issue.Dependencies = append(issue.Dependencies, Dependency{IssueID: taskID, DependsOnID: "dap-jo9.2.5.10", Type: "blocks"})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog, issues := fixture(t)
			issue := issues[taskID]
			test.mutate(&issue)
			issues[taskID] = issue
			if problems := Validate(catalog, issues); !containsString(problems, diagnostic) {
				t.Fatalf("expected exact blocking-prerequisite diagnostic %q, got %v", diagnostic, problems)
			}
		})
	}
}

func removeBlockingDependency(dependencies []Dependency, prerequisite string) []Dependency {
	filtered := dependencies[:0]
	for _, dependency := range dependencies {
		if dependency.Type != "blocks" || dependency.DependsOnID != prerequisite {
			filtered = append(filtered, dependency)
		}
	}
	return filtered
}

func TestStandaloneReasonCannotBypassUnrelatedTaskParent(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.4"
	found := false
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "A non-review task must not opt out of its story/group parent."
			found = true
		}
	}
	if !found {
		t.Fatalf("task %s not found", taskID)
	}
	issue := issues[taskID]
	filtered := issue.Dependencies[:0]
	for _, edge := range issue.Dependencies {
		if edge.Type != "parent-child" {
			filtered = append(filtered, edge)
		}
	}
	issue.Dependencies = filtered
	issues[taskID] = issue
	if problems := Validate(catalog, issues); !containsString(problems, taskID+": standalone traceability is reserved for dap-jo9.2.5.2 or dap-jo9.2.5.9") {
		t.Fatalf("expected unrelated standalone opt-out diagnostic, got %v", problems)
	}
}

func TestStandaloneReviewSpecificationRequiresExactCaseParticipation(t *testing.T) {
	catalog, issues := fixture(t)
	const taskID = "dap-jo9.2.5.2"
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].StandaloneReason = "Review specifications can be prepared separately from actual candidate review acceptance."
			catalog.Tasks[i].CaseRoles = nil
		}
	}
	issue := issues[taskID]
	filtered := issue.Dependencies[:0]
	for _, edge := range issue.Dependencies {
		if edge.Type != "parent-child" {
			filtered = append(filtered, edge)
		}
	}
	issue.Dependencies = filtered
	issues[taskID] = issue
	if problems := Validate(catalog, issues); !containsString(problems, taskID+": standalone task mapping differs from the reviewed US-REM-002 review-gate mapping") {
		t.Fatalf("expected standalone case participation diagnostic, got %v", problems)
	}
}

func TestStandaloneReviewSpecificationRejectsMappingChanges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Catalog)
	}{
		{
			name: "story",
			mutate: func(c *Catalog) {
				for i := range c.Tasks {
					if c.Tasks[i].ID == "dap-jo9.2.5.2" {
						c.Tasks[i].Story = "US-REM-001"
					}
				}
			},
		},
		{
			name: "group",
			mutate: func(c *Catalog) {
				for i := range c.Tasks {
					if c.Tasks[i].ID == "dap-jo9.2.5.2" {
						c.Tasks[i].Group = "dap-jo9.2.4"
					}
				}
			},
		},
		{
			name: "canonical generation case owner",
			mutate: func(c *Catalog) {
				for i := range c.Tests {
					if c.Tests[i].ID == "TC-GEN-VALIDATION-001" {
						c.Tests[i].Beads = "dap-jo9.2.5.2"
					}
				}
			},
		},
		{
			name: "case set",
			mutate: func(c *Catalog) {
				for i := range c.Tasks {
					if c.Tasks[i].ID == "dap-jo9.2.5.2" {
						c.Tasks[i].Tests = []string{"TC-REM-HUMAN-001", "TC-GEN-VALIDATION-001"}
					}
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog, issues := fixture(t)
			for i := range catalog.Tasks {
				if catalog.Tasks[i].ID == "dap-jo9.2.5.2" {
					catalog.Tasks[i].StandaloneReason = "Review specifications can be prepared separately from actual candidate review acceptance."
					break
				}
			}
			test.mutate(&catalog)
			issue := issues["dap-jo9.2.5.2"]
			filtered := issue.Dependencies[:0]
			for _, edge := range issue.Dependencies {
				if edge.Type != "parent-child" {
					filtered = append(filtered, edge)
				}
			}
			issue.Dependencies = filtered
			issues["dap-jo9.2.5.2"] = issue
			if problems := Validate(catalog, issues); !containsString(problems, "dap-jo9.2.5.2: standalone task mapping differs from the reviewed US-REM-002 review-gate mapping") {
				t.Fatalf("expected exact standalone mapping diagnostic, got %v", problems)
			}
		})
	}
}

func TestAggregateReleaseGateRequiresCandidateReviewDependency(t *testing.T) {
	catalog, issues := fixture(t)
	releaseGate := issues["dap-jo9.2.5"]
	filtered := releaseGate.Dependencies[:0]
	for _, edge := range releaseGate.Dependencies {
		if edge.DependsOnID != "dap-jo9.2.5.9" {
			filtered = append(filtered, edge)
		}
	}
	releaseGate.Dependencies = filtered
	issues[releaseGate.ID] = releaseGate
	const diagnostic = "dap-jo9.2.5: aggregate review/release gate must block on dap-jo9.2.5.9"
	if problems := Validate(catalog, issues); !containsString(problems, diagnostic) {
		t.Fatalf("expected aggregate review-gate dependency diagnostic, got %v", problems)
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

func TestRejectsDuplicateTaskGroupIDs(t *testing.T) {
	catalog, issues := fixture(t)
	cases := []struct {
		name string
		want string
		add  func(*Catalog)
	}{
		{
			name: "identical duplicate groups",
			want: "duplicate catalog ID dap-jo9.2.5 (task group and task group)",
			add: func(c *Catalog) {
				c.Groups = append(c.Groups, c.Groups[0])
			},
		},
		{
			name: "conflicting duplicate ownership",
			want: "duplicate catalog ID dap-jo9 (task group and task group)",
			add: func(c *Catalog) {
				c.Groups = append(c.Groups,
					TaskGroup{ID: "dap-jo9", Story: "US-REM-001", Purpose: "first owner"},
					TaskGroup{ID: "dap-jo9", Story: "US-GEN-001", Purpose: "last owner"},
				)
			},
		},
		{
			name: "task and group ID collision",
			want: "duplicate catalog ID dap-jo9.2.3.1 (task group and task)",
			add: func(c *Catalog) {
				task := c.Tasks[0]
				c.Groups = append(c.Groups, TaskGroup{ID: task.ID, Story: task.Story, Purpose: "cross-kind collision"})
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			copy := cloneCatalog(t, catalog)
			test.add(&copy)
			problems := Validate(copy, cloneIssues(issues))
			if len(problems) != 1 || problems[0] != test.want {
				t.Fatalf("expected only %q diagnostic, got %v", test.want, problems)
			}
		})
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
			name: "known acceptance case missing from reviewed task links",
			mutate: func(_ *Catalog, beads map[string]Issue) {
				issue := beads["dap-jo9.2.3.2"]
				issue.AcceptanceCriteria += " Also TC-GEN-INVOICE-001."
				beads[issue.ID] = issue
			},
			want: "acceptance case TC-GEN-INVOICE-001 is not in reviewed task-case links",
		},
		{
			name: "catalog and Beads acceptance sets diverge",
			mutate: func(c *Catalog, _ map[string]Issue) {
				id := "dap-jo9.3.4.1"
				for i := range c.Tasks {
					if c.Tasks[i].ID == id {
						c.Tasks[i].Tests = c.Tasks[i].Tests[:len(c.Tasks[i].Tests)-1]
					}
				}
			},
			want: "catalog cases and Beads task-case links differ",
		},
		{
			name: "duplicate task acceptance case",
			mutate: func(c *Catalog, beads map[string]Issue) {
				id := "dap-jo9.2.3.2"
				caseID := "TC-REM-STATEMENT-001"
				for i := range c.Tasks {
					if c.Tasks[i].ID == id {
						c.Tasks[i].Tests = append(c.Tasks[i].Tests, caseID)
					}
				}
				issue := beads[id]
				issue.Description = strings.Replace(issue.Description, caseID, caseID+", "+caseID, 1)
				beads[id] = issue
			},
			want: "catalog cases and Beads task-case links differ",
		},
		{
			name: "canonical owner lacks inverse link",
			mutate: func(c *Catalog, _ map[string]Issue) {
				for i := range c.Tasks {
					if c.Tasks[i].ID == "dap-jo9.3.4.1" {
						c.Tasks[i].Tests = c.Tasks[i].Tests[:len(c.Tasks[i].Tests)-1]
					}
				}
			},
			want: "canonical owner task dap-jo9.3.4.1 lacks inverse case link",
		},
		{
			name: "role annotation does not imply owner",
			mutate: func(c *Catalog, _ map[string]Issue) {
				c.Tasks[0].CaseRoles = append(c.Tasks[0].CaseRoles, TaskCaseRole{Case: "TC-REM-STATEMENT-001", Role: "implements", Scope: "owner must not be labeled as consumer"})
			},
			want: "owner case TC-REM-STATEMENT-001 must not have a participation role",
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

func TestGeneratedStoryMatrixIncludesLinkedTaskCases(t *testing.T) {
	catalog, _ := fixture(t)
	matrix := string(GeneratedMarkdown(catalog)[filepath.Join("docs", "traceability", "matrix.md")])
	start := strings.Index(matrix, "### US-REM-002 —")
	if start < 0 {
		t.Fatal("matrix is missing US-REM-002")
	}
	section := matrix[start:]
	if next := strings.Index(section[1:], "\n### "); next >= 0 {
		section = section[:next+1]
	}
	for _, id := range []string{"TC-REM-HUMAN-001", "TC-REM-ASSISTIVE-TECH-001"} {
		if !strings.Contains(section, "`"+id+"`") {
			t.Errorf("US-REM-002 matrix row is missing linked task case %s", id)
		}
	}
}

func TestGeneratedMatrixExplainsParticipationAndSpecificationStatus(t *testing.T) {
	catalog, _ := fixture(t)
	matrix := string(GeneratedMarkdown(catalog)[filepath.Join("docs", "traceability", "matrix.md")])
	for _, text := range []string{
		"Task-case links describe contractual participation; all cases remain specification-only, not executed results.",
		"**consumes** (canonical owner `dap-jo9.2.5.2`) — human/AT review contract and eventual results; does not perform human or AT review",
		"**implements** (canonical owner `dap-jo9.3.3.1`) — invoice output; no arithmetic policy implied",
	} {
		if !strings.Contains(matrix, text) {
			t.Errorf("matrix is missing participation boundary %q", text)
		}
	}
}

func TestUnreviewedNonOwnerLinkRequiresRoleOutsideReviewedFixture(t *testing.T) {
	catalog, issues, taskID, caseID, ownerID := catalogWithUnreviewedNonOwnerLink(t)
	problems := Validate(catalog, issues)
	want := taskID + ": non-owner case " + caseID + " requires a participation role (canonical owner " + ownerID + ")"
	if len(problems) != 1 || problems[0] != want {
		t.Fatalf("expected only missing-role error %q for an unreviewed case, got %v", want, problems)
	}
}

func TestRepeatedAcceptanceCaseMentionIsCheckedOnce(t *testing.T) {
	catalog, issues := fixture(t)
	issue := issues["dap-jo9.2.3.2"]
	issue.AcceptanceCriteria += " Repeated reference: TC-REM-STATEMENT-001."
	issues[issue.ID] = issue
	if problems := Validate(catalog, issues); len(problems) != 0 {
		t.Fatalf("a repeated case mention already linked by the task should remain valid, got %v", problems)
	}
}

func TestScopedRoleMakesSyntheticTaskCaseValid(t *testing.T) {
	catalog, issues, taskID, caseID, _ := catalogWithUnreviewedNonOwnerLink(t)
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].CaseRoles = append(catalog.Tasks[i].CaseRoles, TaskCaseRole{
				Case: caseID, Role: "implements", Scope: "synthetic bounded contribution",
			})
		}
	}
	if problems := Validate(catalog, issues); len(problems) != 0 {
		t.Fatalf("scoped role should make the synthetic task case valid, got %v", problems)
	}
}

func TestExistingNonOwnerLinksRenderTheirActualCanonicalOwner(t *testing.T) {
	catalog, _ := fixture(t)
	matrix := string(GeneratedMarkdown(catalog)[filepath.Join("docs", "traceability", "matrix.md")])
	want := []string{
		"`dap-jo9.5.2` → `TC-CROSS-HOSTILE-001`: **implements** (canonical owner `dap-jo9.5.1`)",
		"`dap-jo9.5.2` → `TC-CROSS-TIMEOUT-001`: **implements** (canonical owner `dap-jo9.5.1`)",
		"`dap-jo9.5.2` → `TC-CROSS-PRIVACY-001`: **implements** (canonical owner `dap-jo9.5.1`)",
		"`dap-jo9.5.2` → `TC-CROSS-AUDIT-001`: **implements** (canonical owner `dap-jo9.5.1`)",
		"`dap-jo9.4.1.2` → `TC-AFP-FEASIBILITY-001`: **implements** (canonical owner `dap-jo9.4.1.1`)",
	}
	for _, text := range want {
		if !strings.Contains(matrix, text) {
			t.Errorf("matrix does not preserve explicit non-owner link: %s", text)
		}
	}
	if strings.Contains(matrix, "`dap-jo9.5.2` → `TC-CROSS-HOSTILE-001`: **specification owner**") ||
		strings.Contains(matrix, "`dap-jo9.4.1.2` → `TC-AFP-FEASIBILITY-001`: **specification owner**") {
		t.Fatal("matrix rendered a non-owner link as a specification owner")
	}
}

func TestGeneratedMatrixDoesNotInferOwnershipFromMissingRole(t *testing.T) {
	catalog, _, taskID, caseID, ownerID := catalogWithUnreviewedNonOwnerLink(t)
	matrix := string(GeneratedMarkdown(catalog)[filepath.Join("docs", "traceability", "matrix.md")])
	want := "- `" + taskID + "` → `" + caseID + "`: **participation role missing; canonical owner `" + ownerID + "`**"
	if !strings.Contains(matrix, want) {
		t.Fatalf("matrix must identify the canonical owner and missing role, not infer ownership; want %q", want)
	}
}

func catalogWithUnreviewedNonOwnerLink(t *testing.T) (Catalog, map[string]Issue, string, string, string) {
	t.Helper()
	catalog, issues := fixture(t)
	catalog, issues = cloneCatalog(t, catalog), cloneIssues(issues)
	ownerID, taskID, caseID := "dap-jo9.2.3.1", "dap-jo9.2.3.2", "TC-MAPPING-EDGE-001"
	catalog.Tests = append(catalog.Tests, Case{
		ID: caseID, Lane: "remediation", Family: "cross-family", Beads: ownerID,
		Given:  "a synthetic case outside the reviewed mapping fixture",
		When:   "it is linked to an implementing task without a role",
		Then:   "the canonical owner remains explicit and the missing role is rejected",
		Status: specificationOnly,
	})
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == ownerID || catalog.Tasks[i].ID == taskID {
			catalog.Tasks[i].Tests = append(catalog.Tasks[i].Tests, caseID)
		}
	}
	for _, id := range []string{ownerID, taskID} {
		issue := issues[id]
		issue.AcceptanceCriteria += " Acceptance case: " + caseID + "."
		match := taskDescriptionCasesPattern.FindStringSubmatchIndex(issue.Description)
		if len(match) != 6 {
			t.Fatalf("%s has no parseable Test cases block", id)
		}
		block := strings.TrimSpace(issue.Description[match[2]:match[3]])
		issue.Description = issue.Description[:match[2]] + block + ", " + caseID + "\n\n" + issue.Description[match[3]:]
		issues[id] = issue
	}
	return catalog, issues, taskID, caseID, ownerID
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
