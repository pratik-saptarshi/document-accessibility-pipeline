// Package traceability validates the repository's planning catalog and renders
// its generated Markdown views. It does not implement or assess PDF behavior.
package traceability

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const specificationOnly = "specification-only, not executed"

var acceptanceIDPattern = regexp.MustCompile(`TC-[A-Z0-9-]+`)

type Catalog struct {
	Status        string        `json:"status"`
	Capabilities  []Capability  `json:"capabilities"`
	Requirements  []Requirement `json:"requirements"`
	Stories       []Story       `json:"stories"`
	Functions     []Function    `json:"functions"`
	Groups        []TaskGroup   `json:"groups"`
	Tasks         []Task        `json:"tasks"`
	Tests         []Case        `json:"tests"`
	OpenDecisions []string      `json:"open_decisions"`
}

type Capability struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Lane  string `json:"lane"`
}

type Requirement struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Text       string `json:"text"`
}

type Story struct {
	ID            string   `json:"id"`
	Beads         string   `json:"beads"`
	Capability    string   `json:"capability"`
	Requirements  []string `json:"requirements"`
	Text          string   `json:"text"`
	PersonaStatus string   `json:"persona_status"`
	Tests         []string `json:"tests"`
}

type Function struct {
	ID      string `json:"id"`
	Story   string `json:"story"`
	Inputs  string `json:"inputs"`
	Outputs string `json:"outputs"`
	Rules   string `json:"rules"`
	Errors  string `json:"errors"`
}

type TaskGroup struct {
	ID      string `json:"id"`
	Story   string `json:"story"`
	Purpose string `json:"purpose"`
}

type Task struct {
	ID     string   `json:"id"`
	Story  string   `json:"story"`
	Group  string   `json:"group,omitempty"`
	Status string   `json:"status"`
	Tests  []string `json:"tests"`
}

type Case struct {
	ID     string `json:"id"`
	Lane   string `json:"lane"`
	Family string `json:"family"`
	Beads  string `json:"beads"`
	Given  string `json:"given"`
	When   string `json:"when"`
	Then   string `json:"then"`
	Status string `json:"status"`
}

type Issue struct {
	ID                 string       `json:"id"`
	Description        string       `json:"description"`
	AcceptanceCriteria string       `json:"acceptance_criteria"`
	Status             string       `json:"status"`
	Labels             []string     `json:"labels"`
	Dependencies       []Dependency `json:"dependencies"`
}

type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
}

func LoadCatalog(path string) (Catalog, error) {
	var catalog Catalog
	data, err := os.ReadFile(path)
	if err != nil {
		return catalog, fmt.Errorf("read catalog %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		return catalog, fmt.Errorf("decode catalog %s: %w", path, err)
	}
	return catalog, nil
}

func LoadIssues(path string) (map[string]Issue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open Beads export %s: %w", path, err)
	}
	defer file.Close()

	issues := make(map[string]Issue)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var issue Issue
		if err := json.Unmarshal(scanner.Bytes(), &issue); err != nil {
			return nil, fmt.Errorf("decode Beads export line %d: %w", line, err)
		}
		if issue.ID == "" {
			return nil, fmt.Errorf("Beads export line %d has no issue ID", line)
		}
		if _, exists := issues[issue.ID]; exists {
			return nil, fmt.Errorf("duplicate Beads issue ID %s at line %d", issue.ID, line)
		}
		issues[issue.ID] = issue
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Beads export %s: %w", path, err)
	}
	return issues, nil
}

// Validate checks planning IDs and trace links, the Beads export, and its
// prerequisite graph. It does not validate product behavior or execute cases.
func Validate(catalog Catalog, issues map[string]Issue) []string {
	var problems []string
	allIDs := make(map[string]string)
	addID := func(id, kind string) {
		if previous, exists := allIDs[id]; exists {
			problems = append(problems, fmt.Sprintf("duplicate catalog ID %s (%s and %s)", id, previous, kind))
			return
		}
		allIDs[id] = kind
	}
	for _, item := range catalog.Capabilities {
		addID(item.ID, "capability")
	}
	for _, item := range catalog.Requirements {
		addID(item.ID, "requirement")
	}
	for _, item := range catalog.Stories {
		addID(item.ID, "story")
	}
	for _, item := range catalog.Functions {
		addID(item.ID, "function")
	}
	for _, item := range catalog.Tasks {
		addID(item.ID, "task")
	}
	for _, item := range catalog.Tests {
		addID(item.ID, "test")
	}

	capabilities := make(map[string]Capability, len(catalog.Capabilities))
	requirements := make(map[string]Requirement, len(catalog.Requirements))
	stories := make(map[string]Story, len(catalog.Stories))
	tasks := make(map[string]Task, len(catalog.Tasks))
	cases := make(map[string]Case, len(catalog.Tests))
	groups := make(map[string]TaskGroup, len(catalog.Groups))
	for _, item := range catalog.Capabilities {
		capabilities[item.ID] = item
	}
	for _, item := range catalog.Requirements {
		requirements[item.ID] = item
	}
	for _, item := range catalog.Stories {
		stories[item.ID] = item
	}
	for _, item := range catalog.Tasks {
		tasks[item.ID] = item
	}
	for _, item := range catalog.Tests {
		cases[item.ID] = item
	}
	for _, item := range catalog.Groups {
		groups[item.ID] = item
	}

	for _, item := range catalog.Requirements {
		if _, ok := capabilities[item.Capability]; !ok {
			problems = append(problems, fmt.Sprintf("%s: unresolved capability %s", item.ID, item.Capability))
		}
	}
	for _, story := range catalog.Stories {
		if _, ok := capabilities[story.Capability]; !ok {
			problems = append(problems, fmt.Sprintf("%s: unresolved capability %s", story.ID, story.Capability))
		}
		if len(story.Requirements) == 0 {
			problems = append(problems, fmt.Sprintf("%s: no requirement links", story.ID))
		}
		for _, id := range story.Requirements {
			if _, ok := requirements[id]; !ok {
				problems = append(problems, fmt.Sprintf("%s: unresolved requirement %s", story.ID, id))
			}
		}
		for _, id := range story.Tests {
			if _, ok := cases[id]; !ok {
				problems = append(problems, fmt.Sprintf("%s: unresolved acceptance case %s", story.ID, id))
			}
		}
		if _, ok := issues[story.Beads]; !ok {
			problems = append(problems, fmt.Sprintf("%s: missing Beads story %s", story.ID, story.Beads))
		}
	}
	for _, group := range catalog.Groups {
		if _, ok := stories[group.Story]; !ok {
			problems = append(problems, fmt.Sprintf("%s: unresolved task-group story %s", group.ID, group.Story))
		}
		if _, ok := issues[group.ID]; !ok {
			problems = append(problems, fmt.Sprintf("missing Beads task group %s", group.ID))
		}
	}
	for _, fn := range catalog.Functions {
		if _, ok := stories[fn.Story]; !ok {
			problems = append(problems, fmt.Sprintf("%s: unresolved story %s", fn.ID, fn.Story))
		}
		for label, value := range map[string]string{"inputs": fn.Inputs, "outputs": fn.Outputs, "rules": fn.Rules, "errors": fn.Errors} {
			if strings.TrimSpace(value) == "" {
				problems = append(problems, fmt.Sprintf("%s: missing %s contract", fn.ID, label))
			}
		}
	}
	for _, task := range catalog.Tasks {
		story, storyExists := stories[task.Story]
		if !storyExists {
			problems = append(problems, fmt.Sprintf("%s: unresolved story %s", task.ID, task.Story))
		}
		issue, issueExists := issues[task.ID]
		if !issueExists {
			problems = append(problems, fmt.Sprintf("missing Beads task %s", task.ID))
		} else {
			parent := story.Beads
			if task.Group != "" {
				group, ok := groups[task.Group]
				if !ok {
					problems = append(problems, fmt.Sprintf("%s: unresolved task group %s", task.ID, task.Group))
				} else if group.Story != task.Story {
					problems = append(problems, fmt.Sprintf("%s: task group %s maps to a different story", task.ID, task.Group))
				}
				parent = task.Group
			}
			if !hasParentEdge(issue, parent) {
				problems = append(problems, fmt.Sprintf("%s: Beads parent does not match expected story/group %s", task.ID, parent))
			}
			if !contains(issue.Labels, task.Status) {
				problems = append(problems, fmt.Sprintf("%s: readiness label %s is missing from Beads", task.ID, task.Status))
			}
		}
		if task.Status != "refinement-ready" && task.Status != "decision-blocked" {
			problems = append(problems, fmt.Sprintf("%s: invalid readiness status %s", task.ID, task.Status))
		}
		if len(task.Tests) == 0 {
			problems = append(problems, fmt.Sprintf("%s: no acceptance case links", task.ID))
		}
		for _, id := range task.Tests {
			if _, ok := cases[id]; !ok {
				problems = append(problems, fmt.Sprintf("%s: unresolved acceptance case %s", task.ID, id))
			}
		}
	}
	for _, test := range catalog.Tests {
		for label, value := range map[string]string{"GIVEN": test.Given, "WHEN": test.When, "THEN": test.Then} {
			if strings.TrimSpace(value) == "" {
				problems = append(problems, fmt.Sprintf("%s: missing %s clause", test.ID, label))
			}
		}
		if test.Status != specificationOnly {
			problems = append(problems, fmt.Sprintf("%s: status must be %q", test.ID, specificationOnly))
		}
		if _, ok := issues[test.Beads]; !ok {
			problems = append(problems, fmt.Sprintf("%s: missing Beads test-spec task %s", test.ID, test.Beads))
		} else if _, ok := tasks[test.Beads]; !ok {
			problems = append(problems, fmt.Sprintf("%s: Beads test-spec issue is not a catalog task", test.ID))
		}
	}
	for _, issue := range issues {
		text := issue.Description + " " + issue.AcceptanceCriteria
		for _, id := range acceptanceIDPattern.FindAllString(text, -1) {
			if _, ok := cases[id]; !ok {
				problems = append(problems, fmt.Sprintf("%s: Beads references unknown acceptance case %s", issue.ID, id))
			}
		}
	}

	for _, req := range catalog.Requirements {
		linkedStories := 0
		for _, story := range catalog.Stories {
			if !contains(story.Requirements, req.ID) {
				continue
			}
			linkedStories++
			functionCount, taskCount, linkedCases := 0, 0, false
			for _, fn := range catalog.Functions {
				if fn.Story == story.ID {
					functionCount++
				}
			}
			for _, task := range catalog.Tasks {
				if task.Story != story.ID {
					continue
				}
				taskCount++
				for _, caseID := range task.Tests {
					if _, ok := cases[caseID]; ok {
						linkedCases = true
					}
				}
			}
			if functionCount == 0 {
				problems = append(problems, fmt.Sprintf("%s: story %s has no function contract", req.ID, story.ID))
			}
			if taskCount == 0 {
				problems = append(problems, fmt.Sprintf("%s: story %s has no actionable task", req.ID, story.ID))
			}
			if !linkedCases {
				problems = append(problems, fmt.Sprintf("%s: story %s has no task-linked acceptance case", req.ID, story.ID))
			}
		}
		if linkedStories == 0 {
			problems = append(problems, fmt.Sprintf("%s: no user story", req.ID))
		}
	}

	expectedFamilies := []string{"statements", "invoices", "correspondence", "reports"}
	for _, lane := range []string{"remediation", "generation"} {
		for _, family := range expectedFamilies {
			if !hasLaneFamily(catalog.Tests, lane, family) {
				problems = append(problems, fmt.Sprintf("missing lane/family case coverage: %s / %s", lane, family))
			}
		}
	}
	problems = append(problems, validateDependencies(issues)...)
	sort.Strings(problems)
	return problems
}

func hasParentEdge(issue Issue, parent string) bool {
	for _, edge := range issue.Dependencies {
		if edge.Type == "parent-child" && edge.DependsOnID == parent {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasLaneFamily(cases []Case, lane, family string) bool {
	for _, item := range cases {
		if item.Lane == lane && item.Family == family {
			return true
		}
	}
	return false
}

func validateDependencies(issues map[string]Issue) []string {
	var problems []string
	graph := make(map[string][]string)
	for _, issue := range issues {
		parents := 0
		for _, edge := range issue.Dependencies {
			if _, ok := issues[edge.DependsOnID]; !ok {
				problems = append(problems, fmt.Sprintf("%s: dependency endpoint missing %s", issue.ID, edge.DependsOnID))
				continue
			}
			if edge.Type == "parent-child" {
				parents++
				continue
			}
			if edge.Type == "blocks" {
				graph[issue.ID] = append(graph[issue.ID], edge.DependsOnID)
			}
		}
		if parents > 1 {
			problems = append(problems, fmt.Sprintf("%s: multiple Beads parent-child edges", issue.ID))
		}
	}
	state := make(map[string]uint8, len(issues))
	var visit func(string)
	visit = func(node string) {
		if state[node] == 1 {
			problems = append(problems, fmt.Sprintf("dependency cycle through %s", node))
			return
		}
		if state[node] == 2 {
			return
		}
		state[node] = 1
		for _, next := range graph[node] {
			visit(next)
		}
		state[node] = 2
	}
	for id := range graph {
		visit(id)
	}
	return problems
}

func GeneratedMarkdown(catalog Catalog) map[string][]byte {
	return map[string][]byte{
		filepath.Join("docs", "requirements", "catalog.md"):              markdownBytes(renderRequirements(catalog)),
		filepath.Join("docs", "requirements", "functional-contracts.md"): markdownBytes(renderFunctions(catalog)),
		filepath.Join("docs", "validation", "cases.md"):                  markdownBytes(renderCases(catalog)),
		filepath.Join("docs", "traceability", "matrix.md"):               markdownBytes(renderMatrix(catalog)),
	}
}

func markdownBytes(content string) []byte {
	return []byte(strings.TrimRight(content, "\n") + "\n")
}

func renderRequirements(c Catalog) string {
	capabilities := make(map[string]Capability, len(c.Capabilities))
	for _, item := range c.Capabilities {
		capabilities[item.ID] = item
	}
	var b strings.Builder
	b.WriteString("# Requirements catalog\n\n**Status: Proposed.** Requirement statements are planning contracts. They do not establish compliance, workload, or approved implementation.\n\n")
	for _, req := range c.Requirements {
		fmt.Fprintf(&b, "## %s — %s\n\n%s\n\n", req.ID, capabilities[req.Capability].Title, req.Text)
	}
	b.WriteString("## Open decisions\n\n")
	for _, decision := range c.OpenDecisions {
		fmt.Fprintf(&b, "- %s\n", decision)
	}
	b.WriteString("\n")
	return b.String()
}

func renderFunctions(c Catalog) string {
	var b strings.Builder
	b.WriteString("# Proposed function contracts\n\nThese are behavior contracts for planning only. They do not select a language library, engine, service boundary, or production design.\n\n")
	for _, fn := range c.Functions {
		fmt.Fprintf(&b, "## %s (%s)\n\n- **Inputs:** %s\n- **Outputs:** %s\n- **Rules:** %s\n- **Error states:** %s\n\n", fn.ID, fn.Story, fn.Inputs, fn.Outputs, fn.Rules, fn.Errors)
	}
	return b.String()
}

func renderCases(c Catalog) string {
	var b strings.Builder
	b.WriteString("# Acceptance case specifications\n\n**Status: specifications only; none of these product cases have been executed.** Cases use synthetic/deidentified fixtures. Open thresholds, profiles, and reviewer authority must be resolved before implementation readiness.\n\n")
	for _, item := range c.Tests {
		fmt.Fprintf(&b, "## %s — %s / %s\n\n- **Beads:** `%s`\n- **Status:** %s\n- **Given:** %s\n- **When:** %s\n- **Then:** %s\n\n", item.ID, item.Lane, item.Family, item.Beads, item.Status, item.Given, item.When, item.Then)
	}
	return b.String()
}

func renderMatrix(c Catalog) string {
	var b strings.Builder
	b.WriteString("# Traceability matrix\n\nThe JSON catalog is the planning source of truth for stable IDs and explicit links. The checker validates these links against the tracked Beads JSONL export and generates this view.\n\n```text\nCapability → requirement → proposed persona/story → function contract → Beads task → acceptance case specification\n```\n\n## Stories and linked work\n\n")
	for _, story := range c.Stories {
		fmt.Fprintf(&b, "### %s — %s\n\n%s\n\n- **Capability:** `%s`\n- **Requirements:** %s\n- **Persona:** %s\n", story.ID, story.Beads, story.Text, story.Capability, strings.Join(story.Requirements, ", "), story.PersonaStatus)
		var functionIDs, taskIDs []string
		for _, fn := range c.Functions {
			if fn.Story == story.ID {
				functionIDs = append(functionIDs, fn.ID)
			}
		}
		for _, task := range c.Tasks {
			if task.Story != story.ID {
				continue
			}
			label := task.Status
			if task.Group != "" {
				label += ", group " + task.Group
			}
			taskIDs = append(taskIDs, fmt.Sprintf("`%s` (%s)", task.ID, label))
		}
		var caseIDs []string
		for _, id := range story.Tests {
			caseIDs = append(caseIDs, fmt.Sprintf("`%s`", id))
		}
		fmt.Fprintf(&b, "\n- **Functions:** %s\n- **Tasks:** %s\n- **Cases:** %s\n\n", strings.Join(functionIDs, ", "), strings.Join(taskIDs, ", "), strings.Join(caseIDs, ", "))
	}
	b.WriteString("## Readiness rules\n\n- **Refinement-ready** means a planning/specification task can be detailed using current information; it does not mean implementation is authorized or product tests pass.\n- **Decision-blocked** means one or more profile, feasibility, fidelity, tool, privacy, or authority decisions must be recorded before implementation.\n- **Definition of Ready:** a specific story, proposed persona, linked requirement, function contract, actionable Beads task, synthetic case IDs with Given/When/Then, and explicit unresolved decisions.\n- **Definition of Done for planning:** IDs and edges validate; each requirement traces through a story, function, task, and case; family coverage exists in both lanes; test states say specification-only; and open decisions are visible.\n- **Definition of Done for future product work:** write failing tests first, implement the smallest slice, and record passing results; demonstrate preservation/failure behavior; verify licenses and versions; independently record machine, human semantic, and assistive-technology review. A planning case is not a passing test.\n")
	return b.String()
}
