package traceability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type mappingExpectations struct {
	CaseOwners        map[string]string         `json:"case_owners"`
	TaskRoles         map[string][]TaskCaseRole `json:"task_roles"`
	StoryRequirements map[string][]string       `json:"story_requirements"`
	LegacyCrosswalk   []legacyCrosswalkEntry    `json:"legacy_crosswalk"`
}

type legacyCrosswalkEntry struct {
	ID      string   `json:"id"`
	Stories []string `json:"stories"`
	Targets []string `json:"targets"`
}

func TestReviewedTaskCaseObligationsMatchCatalogAndBeads(t *testing.T) {
	catalog, issues := fixture(t)
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "internal", "traceability", "testdata", "task-case-obligations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string][]string
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}

	tasks := make(map[string]Task, len(catalog.Tasks))
	for _, task := range catalog.Tasks {
		tasks[task.ID] = task
	}
	for id, want := range expected {
		t.Run(id, func(t *testing.T) {
			task, ok := tasks[id]
			if !ok {
				t.Fatalf("reviewed task %s is missing from the catalog", id)
			}
			if !sameIDSet(task.Tests, want) {
				t.Errorf("catalog cases = %v, want reviewed obligations %v", sortedIDs(task.Tests), sortedIDs(want))
			}
			issue, ok := issues[id]
			if !ok {
				t.Fatalf("reviewed task %s is missing from the Beads export", id)
			}
			gotFromBeads := issueTaskCaseLinks(issue.Description)
			if !sameIDSet(gotFromBeads, want) {
				t.Errorf("Beads task-case links = %v, want reviewed obligations %v", sortedIDs(gotFromBeads), sortedIDs(want))
			}
		})
	}
}

func TestMappingOwnersRolesAndStoryRequirementsMatchReviewFixture(t *testing.T) {
	catalog, issues := fixture(t)
	expect := loadMappingExpectations(t)
	actualOwners := make(map[string]string, len(catalog.Tests))
	for _, item := range catalog.Tests {
		actualOwners[item.ID] = item.Beads
	}
	if !reflect.DeepEqual(actualOwners, expect.CaseOwners) {
		t.Fatalf("canonical case owners differ from reviewed mapping fixture")
	}
	tasks := make(map[string]Task, len(catalog.Tasks))
	for _, task := range catalog.Tasks {
		tasks[task.ID] = task
	}
	for id, want := range expect.TaskRoles {
		if got := tasks[id].CaseRoles; !reflect.DeepEqual(got, want) {
			t.Errorf("%s roles = %#v, want %#v", id, got, want)
		}
		issue, ok := issues[id]
		if !ok {
			t.Errorf("%s participation issue is missing from Beads", id)
			continue
		}
		for _, role := range want {
			owner := expect.CaseOwners[role.Case]
			link := fmt.Sprintf("`%s` — %s; canonical owner `%s`; scope: %s.", role.Case, role.Role, owner, role.Scope)
			if !strings.Contains(issue.Description, link) {
				t.Errorf("%s Beads description is missing reviewed role link %q", id, link)
			}
		}
	}
	for id, task := range tasks {
		if _, reviewed := expect.TaskRoles[id]; !reviewed && len(task.CaseRoles) != 0 {
			t.Errorf("%s has unreviewed participation roles: %#v", id, task.CaseRoles)
		}
	}
	for id, want := range expect.StoryRequirements {
		var story *Story
		for i := range catalog.Stories {
			if catalog.Stories[i].Beads == id {
				story = &catalog.Stories[i]
				break
			}
		}
		if story == nil {
			t.Errorf("missing catalog story for %s", id)
			continue
		}
		if !sameIDSet(story.Requirements, want) {
			t.Errorf("%s catalog requirements = %v, want %v", id, sortedIDs(story.Requirements), sortedIDs(want))
		}
		issue := issues[id]
		got := requirementIDPattern.FindAllString(issue.AcceptanceCriteria, -1)
		if !sameIDSet(got, want) {
			t.Errorf("%s Beads requirements = %v, want %v", id, sortedIDs(got), sortedIDs(want))
		}
	}
}

func TestOmittedMappingFailsEvenWhenCatalogAndBeadsBothOmitIt(t *testing.T) {
	catalog, issues := fixture(t)
	want := loadTaskCaseObligations(t)
	taskID := "dap-jo9.3.4.1"
	caseID := "TC-GEN-ENGINE-GATE-001"
	for i := range catalog.Tasks {
		if catalog.Tasks[i].ID == taskID {
			filtered := catalog.Tasks[i].Tests[:0]
			for _, id := range catalog.Tasks[i].Tests {
				if id != caseID {
					filtered = append(filtered, id)
				}
			}
			catalog.Tasks[i].Tests = filtered
		}
	}
	issue := issues[taskID]
	issue.Description = strings.ReplaceAll(issue.Description, caseID, "")
	issues[taskID] = issue
	var got []string
	for _, task := range catalog.Tasks {
		if task.ID == taskID {
			got = task.Tests
		}
	}
	if sameIDSet(got, want[taskID]) {
		t.Fatal("mutated catalog unexpectedly matches the independent reviewed mapping")
	}
	if sameIDSet(issueTaskCaseLinks(issues[taskID].Description), want[taskID]) {
		t.Fatal("mutated Beads acceptance unexpectedly matches the independent reviewed mapping")
	}
}

func TestLegacyCrosswalkIsDocumentedAsProposedContinuity(t *testing.T) {
	expect := loadMappingExpectations(t)
	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "traceability", "legacy-requirement-crosswalk-2026-10-13.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	if !strings.Contains(text, "proposed continuity") || !strings.Contains(text, "not proven semantic equivalence") {
		t.Fatal("legacy crosswalk must label its mappings as proposed continuity, not proven equivalence")
	}
	rows := legacyCrosswalkRows(text)
	for _, entry := range expect.LegacyCrosswalk {
		if !crosswalkRowMatches(rows[entry.ID], entry) {
			t.Errorf("crosswalk row for %s does not retain its expected stories and targets: %v", entry.ID, rows[entry.ID])
		}
	}
}

func TestLegacyCrosswalkRowOracleRejectsSwappedTargetsAndDeletedStory(t *testing.T) {
	expect := loadMappingExpectations(t)
	byID := make(map[string]legacyCrosswalkEntry, len(expect.LegacyCrosswalk))
	for _, entry := range expect.LegacyCrosswalk {
		byID[entry.ID] = entry
	}
	doc, err := os.ReadFile(filepath.Join("..", "..", "docs", "traceability", "legacy-requirement-crosswalk-2026-10-13.md"))
	if err != nil {
		t.Fatal(err)
	}
	rows := legacyCrosswalkRows(string(doc))
	swappedTargets := append([]string(nil), rows["REQ-GEN-OUTPUT-001"]...)
	sourceTargets := strings.Join(byID["REQ-GEN-SOURCE-001"].Targets, ", ")
	swappedTargets[2] = " `" + strings.ReplaceAll(sourceTargets, ", ", "`, `") + "` "
	if crosswalkRowMatches(swappedTargets, byID["REQ-GEN-OUTPUT-001"]) {
		t.Fatal("row oracle accepted targets swapped from REQ-GEN-SOURCE-001")
	}
	deletedStory := append([]string(nil), rows["REQ-SCOPE-001"]...)
	deletedStory[1] = " `dap-jo9.2.3` "
	if crosswalkRowMatches(deletedStory, byID["REQ-SCOPE-001"]) {
		t.Fatal("row oracle accepted a deleted story link for REQ-SCOPE-001")
	}
}

func legacyCrosswalkRows(text string) map[string][]string {
	rows := make(map[string][]string)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "| `REQ-") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 6 {
			continue
		}
		id := strings.Trim(strings.TrimSpace(cells[1]), "`")
		rows[id] = cells
	}
	return rows
}

func crosswalkRowMatches(cells []string, entry legacyCrosswalkEntry) bool {
	if len(cells) < 6 || strings.Trim(strings.TrimSpace(cells[1]), "`") != entry.ID {
		return false
	}
	for _, story := range entry.Stories {
		if !strings.Contains(cells[2], "`"+story+"`") {
			return false
		}
	}
	for _, target := range entry.Targets {
		if !strings.Contains(cells[3], "`"+target+"`") {
			return false
		}
	}
	return true
}

func loadMappingExpectations(t *testing.T) mappingExpectations {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "mapping-expectations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expect mappingExpectations
	if err := json.Unmarshal(data, &expect); err != nil {
		t.Fatal(err)
	}
	return expect
}

func loadTaskCaseObligations(t *testing.T) map[string][]string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "task-case-obligations.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string][]string
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	return expected
}

func sameIDSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, id := range left {
		counts[id]++
		if counts[id] > 1 {
			return false
		}
	}
	for _, id := range right {
		counts[id]--
		if counts[id] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

var requirementIDPattern = regexp.MustCompile(`REQ-[A-Z0-9-]+`)

func sortedIDs(ids []string) []string {
	result := append([]string(nil), ids...)
	sort.Strings(result)
	return result
}
