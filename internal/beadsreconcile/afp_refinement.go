package beadsreconcile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// AFPCommandWindow binds one Beads-created dependency edge timestamp to the
// recorded start and finish of the exact disposable-store command.
type AFPCommandWindow struct {
	Edge       AFPEdge
	StartedAt  time.Time
	FinishedAt time.Time
}

// ApplyAFPRefinement runs only the literal operations compiled into this
// source-bound refinement. The callback must be the pinned disposable Beads
// runner; IDs are accepted only from its --silent create output.
func ApplyAFPRefinement(run func(...string) ([]byte, error)) (map[string]string, error) {
	for _, update := range afpTextUpdates {
		args := []string{"update", update.id}
		fields := make([]string, 0, len(update.fields))
		for field := range update.fields {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		for _, field := range fields {
			flag := map[string]string{"title": "--title", "description": "--description", "acceptance_criteria": "--acceptance", "notes": "--notes"}[field]
			args = append(args, flag+"="+update.fields[field])
		}
		if priority, ok := afpPriorities[update.id]; ok {
			args = append(args, "--priority="+fmt.Sprint(priority))
		}
		if _, err := run(args...); err != nil {
			return nil, fmt.Errorf("apply AFP update %s: %w", update.id, err)
		}
	}
	roles := make(map[string]string, len(afpTemplates))
	for _, t := range afpTemplates {
		args := afpCreateArgs(t, roles)
		out, err := run(args...)
		if err != nil {
			return nil, fmt.Errorf("create AFP %s issue: %w", t.role, err)
		}
		id := strings.TrimSpace(string(out))
		if !regexp.MustCompile(`^dap-[a-zA-Z0-9][a-zA-Z0-9._-]{1,63}$`).MatchString(id) {
			return nil, fmt.Errorf("Beads returned an invalid ID for AFP %s issue", t.role)
		}
		roles[t.role] = id
	}
	for _, edge := range afpAdditionalEdges {
		from, to := edge.IssueID, resolveAFPRole(edge.DependsOnID, roles)
		if _, err := run("dep", "add", from, to, "--type=blocks"); err != nil {
			return nil, fmt.Errorf("add AFP dependency %s -> %s: %w", from, to, err)
		}
	}
	return roles, nil
}

// AFPRefinementCommandSequence returns the deterministic command arguments
// used to validate bundle provenance after role IDs have been resolved.
func AFPRefinementCommandSequence(roles map[string]string) ([][]string, error) {
	if len(roles) != len(afpTemplates) {
		return nil, errors.New("AFP role ID set is incomplete")
	}
	for _, t := range afpTemplates {
		if roles[t.role] == "" {
			return nil, fmt.Errorf("AFP role %s has no generated ID", t.role)
		}
	}
	commands := make([][]string, 0, len(afpTextUpdates)+len(afpTemplates)+len(afpAdditionalEdges))
	for _, u := range afpTextUpdates {
		args := []string{"update", u.id}
		fields := make([]string, 0, len(u.fields))
		for f := range u.fields {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		for _, f := range fields {
			flag := map[string]string{"title": "--title", "description": "--description", "acceptance_criteria": "--acceptance", "notes": "--notes"}[f]
			args = append(args, flag+"="+u.fields[f])
		}
		if p, ok := afpPriorities[u.id]; ok {
			args = append(args, "--priority="+fmt.Sprint(p))
		}
		commands = append(commands, args)
	}
	for _, t := range afpTemplates {
		commands = append(commands, afpCreateArgs(t, roles))
	}
	for _, e := range afpAdditionalEdges {
		commands = append(commands, []string{"dep", "add", e.IssueID, resolveAFPRole(e.DependsOnID, roles), "--type=blocks"})
	}
	return commands, nil
}

// AFPRefinementEdgeCommandIndices binds each generated edge to the exact
// create or dep-add operation that produced it. Indexes are zero-based within
// AFPRefinementCommandSequence's operation list.
func AFPRefinementEdgeCommandIndices(roles map[string]string) (map[AFPEdge]int, error) {
	if _, err := AFPRefinementCommandSequence(roles); err != nil {
		return nil, err
	}
	result := make(map[AFPEdge]int, len(afpAdditionalEdges)+len(afpTemplates)+2)
	createOffset := len(afpTextUpdates)
	for i, template := range afpTemplates {
		id := roles[template.role]
		result[AFPEdge{IssueID: id, DependsOnID: "dap-jo9.4", Type: "parent-child"}] = createOffset + i
		if template.dependsOnRole != "" {
			dependsOn := template.dependsOnRole
			if dependsOn == "decision" {
				dependsOn = "dap-jo9.4.2"
			} else {
				dependsOn = roles[dependsOn]
			}
			result[AFPEdge{IssueID: id, DependsOnID: dependsOn, Type: "blocks"}] = createOffset + i
		}
	}
	dependencyOffset := createOffset + len(afpTemplates)
	for i, edge := range afpAdditionalEdges {
		edge.DependsOnID = resolveAFPRole(edge.DependsOnID, roles)
		result[edge] = dependencyOffset + i
	}
	return result, nil
}

func afpCreateArgs(t afpIssueTemplate, roles map[string]string) []string {
	args := []string{"create", "--title=" + t.title, "--description=" + t.description, "--acceptance=" + t.acceptance, "--notes=", "--status=open", "--priority=P" + fmt.Sprint(t.priority), "--type=" + t.issueType, "--labels=" + strings.Join(t.labels, ","), "--no-inherit-labels", "--parent=dap-jo9.4"}
	if t.dependsOnRole != "" {
		dep := t.dependsOnRole
		if dep == "decision" {
			dep = "dap-jo9.4.2"
		} else {
			dep = roles[dep]
		}
		args = append(args, "--deps="+dep)
	}
	return append(args, "--silent")
}

type afpIssueTemplate struct {
	role, title, description, acceptance, issueType string
	priority                                        int
	labels                                          []string
	dependsOnRole                                   string
}

type afpTextUpdate struct {
	id     string
	fields map[string]string
}

var afpTextUpdates = []afpTextUpdate{
	{id: "dap-jo9.4", fields: map[string]string{
		"description":         "Evaluate the first owner-agreed AFP source profile as an in-product feasibility workstream in this pipeline, using the separately maintained go-afp-line candidate library intended for this pipeline. This work does not approve dependency adoption or converter implementation.",
		"acceptance_criteria": "The workstream records (1) an owner-agreed AFP source profile, PDF target, authorized fixtures, trusted reference observations, and acceptor; (2) a bounded feasibility assessment of the pinned candidate library against that profile, including independently checked text, page order, and geometry; (3) a separate dependency review; and (4) an explicit owner go/no-go/unknown or deferral before delivery refinement. Do not infer fidelity from parser round-trip, treat a zero-exclusion count as sufficient, authorize implementation, or make AFP a prerequisite for unrelated PDF lanes.",
	}},
	{id: "dap-jo9.4.1", fields: map[string]string{
		"title":       "Record AFP organizational decision and preserve feasibility traceability",
		"description": "The owner decided AFP remains within this pipeline and is not a separate project. go-afp-line remains separately maintained and is only a candidate dependency intended for this pipeline; adoption is unapproved. This issue records the organizational boundary. Its existing feasibility acceptance criteria and case links remain in place until a separately reviewed traceability reassignment; this issue stays open and does not authorize implementation.",
		"notes":       "\n\n2026-10-14 panel (deferred): link exact catalog rows (docs/requirements/catalog.md) instead of bare CAP-/REQ-/TC- IDs in story AC.\n\nDecision 0018 records that AFP remains in this pipeline, is not a separate project, and go-afp-line remains separately maintained. This organizational decision does not satisfy or close the existing feasibility acceptance criteria. Preserve feasibility traceability until its separately reviewed reassignment.",
	}},
	{id: "dap-jo9.4.1.2", fields: map[string]string{
		"title":               "Assess go-afp-line against the owner-agreed AFP profile",
		"description":         "## Goal\n\nAssess the exact go-afp-line revision against the source profile, PDF target, fixtures, trusted references, and feature/error boundaries recorded by the owner profile/reference decision. Return evidence-backed proceed, reject, or unknown. Do not implement a converter or adopt the library.\n\n## Evidence\n\nIndependently compare extracted text, page order, and page geometry with the authorized trusted references. Preflight semantic features and classify unsupported, malformed, resource, cancellation, encoding, and font failures. Record bounded-resource and cancellation behavior, no-successful-partial-output behavior, original preservation, input/output/version provenance, privacy-safe diagnostics, and the untagged/non-accessibility boundary. Use CI positive and negative fixtures and human acceptance against references. Do not invent numeric tolerances, budgets, fixture counts, or accepted personas.\n\n## Test cases\n\nTC-AFP-FEASIBILITY-001\n\n## Traceability\n\nTC-AFP-FEASIBILITY-001 remains a specification, not executed evidence. Preserve the existing parent and dap-jo9.1.1 prerequisite; this assessment does not close the profile decision or authorize either PDF lane.\n\n## Case participation and evidence boundary\n\n- `TC-AFP-FEASIBILITY-001` — implements; canonical owner `dap-jo9.4.1.1`; scope: bounded AFP feasibility assessment and recommendation only; no conversion implementation.\n\nThese are task participation links only; the canonical case owner remains authoritative and no case execution is implied.",
		"acceptance_criteria": "The owner profile/reference decision and dap-jo9.1.1 prerequisite are satisfied. Record the exact candidate revision and evidence for the agreed profile: independent text/order/geometry comparison to trusted references; semantic preflight; classified unsupported, malformed, resource, cancellation, encoding, and font failures; approved resource bounds; no successful partial output; original preservation; input/output/version provenance; privacy-safe diagnostics; and explicit untagged/non-accessibility limits. Include positive and negative CI fixture results and human reference acceptance. Return proceed, reject, or unknown with unresolved gaps. TC-AFP-FEASIBILITY-001 remains a specification link, not executed evidence. No converter implementation, inferred thresholds, or closure of the profile/organizational issues.",
	}},
	{id: "dap-jo9.4.2", fields: map[string]string{
		"description":         "## Goal\n\nRecord the owner's AFP go/no-go/unknown decision or explicit deferral only after the agreed profile, feasibility assessment, and separately maintained library dependency review have produced their evidence. A go requires an explicit adoption decision and authorized delivery scope.\n\n## Context\n\nAFP remains in this pipeline and is not a separate project; this organizational decision does not determine feasibility, dependency adoption, or delivery.",
		"acceptance_criteria": "1. Record go/no-go/unknown, the named decision owner, input profile and criteria, evidence reviewed, and date. A GO explicitly approves adoption of the exact separately maintained go-afp-line revision and names the authorized next scope. A NO GO records non-adoption and rationale. UNKNOWN or deferral records unresolved evidence and the trigger that re-opens the decision. Require completed dap-jo9.4.1.2 and dependency-review evidence; this decision does not itself implement conversion.",
	}},
}

var afpPriorities = map[string]int{"dap-jo9.4.1.2": 1, "dap-jo9.4.2": 1}

var afpTemplates = []afpIssueTemplate{
	{role: "profile", title: "Agree first AFP source profile and trusted references", issueType: "decision", priority: 1,
		description: "Record the owner-approved first useful AFP input profile and PDF target before feasibility or dependency review starts. Identify the profile's source characteristics, authorized synthetic or deidentified fixtures, trusted expected text/page-order/geometry observations, relevant semantic features and failure boundaries, and the people authorized to accept the comparison. Do not invent a customer, persona, sample count, tolerance, resource budget, or delivery scope.",
		acceptance:  "Record the named owner and date; AFP source profile and PDF target; authorized fixture identities and handling constraints; trusted reference identities and how text, page order, and geometry will be checked; semantic feature and unsupported/error boundaries; and the authorized reference acceptor. Mark every unresolved item explicitly. A feasibility recommendation cannot pass this gate while required profile/reference inputs remain unresolved. This decision does not approve dependency adoption or implementation.",
		labels:      []string{"proposed", "lane-afp", "phase-na"}},
	{role: "dependency_review", title: "Review go-afp-line dependency for the agreed AFP profile", issueType: "task", priority: 1,
		description: "Review the separately maintained go-afp-line revision proposed for use by this pipeline, against the owner-approved AFP profile. This is dependency due diligence and an adoption recommendation, not an approval or integration.",
		acceptance:  "The owner profile/reference decision is complete. Record the exact source revision; direct and transitive license evidence; maintenance and security posture; release/versioning evidence; supported and unsupported profile features; integration behavior; operational fit; and material unknowns. Return adopt, reject, or defer with evidence and explicit limitations. A zero-exclusion count alone is not a fidelity or adoption result. Do not add the dependency to the product or implement conversion in this issue.",
		labels:      []string{"proposed", "lane-afp", "phase-na"}, dependsOnRole: "profile"},
	{role: "delivery", title: "Convert supported AFP documents into reviewable PDFs", issueType: "story", priority: 2,
		description: "Implement an owner-authorized workflow that converts AFP from the agreed source profile into a reviewable PDF with explicit diagnostics. Internal document-operations review remains a hypothesis until the owner confirms audience and scope. This is a future implementation story, currently blocked pending feasibility, dependency review, and explicit owner GO/adoption/scope authorization. This refinement task does not authorize implementation.",
		acceptance:  "This is implementation acceptance, not completion of the current planning or refinement task. Do not begin implementation until the owner confirms profile, audience, and delivery scope; the profile/reference decision is complete; feasibility recommends proceed; dependency review recommends adopting the exact library revision; and dap-jo9.4.2 records GO and adoption approval. Completion requires the implemented behavior and executed evidence below:\n1. Independently compare output text, page order, and page geometry with the authorized trusted references; record method, reference identity, result, and differences without inventing tolerances.\n2. Implement semantic-feature preflight that reports supported, unsupported, and not-evaluated features. Required unsupported features, malformed input, resource exhaustion, cancellation, encoding errors, or unavailable required fonts produce classified failures and no successful partial result. The owner decides whether any declared exclusion is permissible.\n3. Enforce approved bounds for input, memory, time, output, and concurrency. Executed cancellation/failure checks show no output is presented as successful; complete output is published only after required checks pass.\n4. Preserve the original. Record input/output identities and hashes, exact library and pipeline versions, profile identity, and processing outcome.\n5. Keep ordinary diagnostics privacy-safe; do not include source text or document content. Retention, access, and deletion follow owner/security decisions.\n6. State that the output PDF is untagged and is not an accessible or PDF/UA-conformant deliverable. Make no accessibility claim.\n7. Run CI positive and negative fixtures for the agreed profile and explicit exclusions, record their results, and obtain human acceptance from the authorized reviewer against trusted references. Tests do not replace human acceptance.\nDo not assign requirement or case IDs until traceability ownership and scope are approved. A plan, template, or unexecuted test specification does not satisfy this story.",
		labels:      []string{"proposed", "user-story", "lane-afp", "phase-na", "decision-blocked"}, dependsOnRole: "decision"},
}

var afpAdditionalEdges = []AFPEdge{
	{IssueID: "dap-jo9.4.1.2", DependsOnID: "@profile", Type: "blocks"},
	{IssueID: "dap-jo9.4.2", DependsOnID: "dap-jo9.4.1.2", Type: "blocks"},
	{IssueID: "dap-jo9.4.2", DependsOnID: "@dependency_review", Type: "blocks"},
}

var afpImportTimestampPairs = map[string]map[string]string{
	"dap-jo9.1.6.10": {"closed_at": "2026-10-17T05:50:49.684035Z", "updated_at": "2026-10-17T05:50:49.684035Z"},
	"dap-jo9.19":     {"closed_at": "2026-10-17T08:56:03.948214+00:00", "updated_at": "2026-10-17T08:56:03.948214+00:00"},
}

var afpImportedTimestampValues = map[string]map[string]string{
	"dap-jo9.1.6.10": {"closed_at": "2026-10-17T05:50:50Z", "updated_at": "2026-10-17T05:50:50Z"},
	"dap-jo9.19":     {"closed_at": "2026-10-17T08:56:04Z", "updated_at": "2026-10-17T08:56:04Z"},
}

// ValidateAFPBaseline checks the exact pinned import result before the batch
// applies any updates. It also verifies the one source-bound null loss.
func ValidateAFPBaseline(source, baseline []byte, version, binary string) error {
	if version != approvedNullLossVersion || binary != approvedNullLossBDSHA256 {
		return errors.New("AFP baseline requires the pinned Beads 1.3.1 executable")
	}
	allowed, _, err := authorizeNullLoss(source, baseline, baseline, version, binary, true)
	if err != nil {
		return err
	}
	c, err := parseExport("source", source)
	if err != nil {
		return err
	}
	b, err := parseExport("baseline", baseline)
	if err != nil {
		return err
	}
	if len(c) != len(b) {
		return errors.New("AFP baseline changed the issue ID set")
	}
	for id, left := range c {
		right, ok := b[id]
		if !ok {
			return fmt.Errorf("AFP baseline is missing issue %s", id)
		}
		ln, rn, _, err := normalizeAFPImportRows(left, right)
		if err != nil {
			return err
		}
		if err := compareRows(ln, rn, id, nil, false, version, allowed[id]); err != nil {
			return fmt.Errorf("AFP source to baseline issue %s: %w", id, err)
		}
	}
	return nil
}

func validateAFPPlan(source []byte, plan Plan) error {
	if digest(source) != approvedAFPRefinementSourceSHA256 {
		return errors.New("AFP refinement requires the exact approved canonical source digest")
	}
	if plan.BDVersion != approvedNullLossVersion || !plan.AFPRefinement || len(plan.Updates) != 0 || len(plan.Creates) != 0 || len(plan.Closes) != 0 {
		return errors.New("AFP refinement plan must select only the pinned AFP batch and Beads 1.3.1")
	}
	return nil
}

func reconcileAFP(source, baseline, updated []byte, plan Plan, binary string, nullLosses map[string]map[string]bool, lossEvidence []RepresentationLoss, windows []AFPCommandWindow) ([]byte, Evidence, error) {
	if err := validateAFPPlan(source, plan); err != nil {
		return nil, Evidence{}, err
	}
	if binary != approvedNullLossBDSHA256 {
		return nil, Evidence{}, errors.New("AFP refinement requires the exact pinned Beads executable")
	}
	c, err := parseExport("source", source)
	if err != nil {
		return nil, Evidence{}, err
	}
	b0, err := parseExport("baseline", baseline)
	if err != nil {
		return nil, Evidence{}, err
	}
	b1, err := parseExport("updated", updated)
	if err != nil {
		return nil, Evidence{}, err
	}
	if len(c) != len(b0) || len(b1) != len(c)+len(afpTemplates) {
		return nil, Evidence{}, fmt.Errorf("AFP batch IDs changed unexpectedly: source=%d baseline=%d updated=%d", len(c), len(b0), len(b1))
	}
	for id, cr := range c {
		br, ok := b0[id]
		if !ok {
			return nil, Evidence{}, fmt.Errorf("AFP baseline is missing source issue %s", id)
		}
		cn, bn, _, normErr := normalizeAFPImportRows(cr, br)
		if normErr != nil {
			return nil, Evidence{}, normErr
		}
		if err := compareRows(cn, bn, id, nil, false, plan.BDVersion, nullLosses[id]); err != nil {
			return nil, Evidence{}, fmt.Errorf("source to AFP baseline issue %s: %w", id, err)
		}
		if _, ok := b1[id]; !ok {
			return nil, Evidence{}, fmt.Errorf("AFP updated export is missing source issue %s", id)
		}
	}
	created, roleIDs, err := identifyAFPCreates(c, b1)
	if err != nil {
		return nil, Evidence{}, err
	}
	expectedEdges := append([]AFPEdge(nil), afpAdditionalEdges...)
	for i := range expectedEdges {
		expectedEdges[i].DependsOnID = resolveAFPRole(expectedEdges[i].DependsOnID, roleIDs)
	}
	for _, t := range afpTemplates {
		id := roleIDs[t.role]
		expectedEdges = append(expectedEdges, AFPEdge{IssueID: id, DependsOnID: "dap-jo9.4", Type: "parent-child"})
		if t.dependsOnRole != "" {
			dep := t.dependsOnRole
			if dep == "decision" {
				dep = "dap-jo9.4.2"
			} else {
				dep = roleIDs[dep]
			}
			expectedEdges = append(expectedEdges, AFPEdge{IssueID: id, DependsOnID: dep, Type: "blocks"})
		}
	}
	if err := validateAFPGraph(c, b0, b1, expectedEdges, roleIDs, windows); err != nil {
		return nil, Evidence{}, err
	}
	if err := validateAFPExistingRows(b0, b1); err != nil {
		return nil, Evidence{}, err
	}
	if err := validateAFPCreatedRows(roleIDs, b1); err != nil {
		return nil, Evidence{}, err
	}
	if err := validateAFPCounters(c, b1); err != nil {
		return nil, Evidence{}, err
	}

	lines := splitLines(source)
	changes := make([]EvidenceChange, 0)
	for i, line := range lines {
		id, err := rowID(line)
		if err != nil {
			return nil, Evidence{}, err
		}
		before, after := b0[id], b1[id]
		replacements := make(map[string]json.RawMessage)
		changedFields := make(map[string]string)
		for _, u := range afpTextUpdates {
			if u.id != id {
				continue
			}
			for field, value := range u.fields {
				got, e := stringField(after, field)
				if e != nil || got != value {
					return nil, Evidence{}, fmt.Errorf("AFP update %s.%s differs from approved text", id, field)
				}
				raw, _ := json.Marshal(value)
				if !jsonEqual(before.fields[field], raw) {
					replacements[field] = raw
					changedFields[field] = value
				}
			}
		}
		if priority, ok := afpPriorities[id]; ok {
			var got int
			if err := json.Unmarshal(after.fields["priority"], &got); err != nil || got != priority {
				return nil, Evidence{}, fmt.Errorf("AFP priority for %s differs from approved P1", id)
			}
			if !jsonEqual(before.fields["priority"], after.fields["priority"]) {
				replacements["priority"] = append(json.RawMessage(nil), after.fields["priority"]...)
				changedFields["priority"] = fmt.Sprint(priority)
			}
		}
		for field := range map[string]bool{"dependencies": true, "dependency_count": true, "dependent_count": true} {
			if !afpEquivalentField(before, after, field) {
				if field == "dependencies" {
					replacements[field], err = candidateAFPDependencies(c[id], after, expectedEdges)
					if err != nil {
						return nil, Evidence{}, fmt.Errorf("preserve AFP source dependency records for %s: %w", id, err)
					}
				} else {
					replacements[field] = append(json.RawMessage(nil), after.fields[field]...)
				}
				changedFields[field] = string(after.fields[field])
			}
		}
		if len(replacements) > 0 {
			oldAt, e1 := stringField(before, "updated_at")
			newAt, e2 := stringField(after, "updated_at")
			if e1 != nil || e2 != nil || oldAt == newAt {
				return nil, Evidence{}, fmt.Errorf("AFP changed issue %s has no Beads updated_at change", id)
			}
			replacements["updated_at"], _ = json.Marshal(newAt)
			changed, err := spliceTopLevel(line, replacements)
			if err != nil {
				return nil, Evidence{}, err
			}
			lines[i] = changed
			changes = append(changes, EvidenceChange{ID: id, Fields: changedFields, UpdatedAt: newAt})
		}
	}
	createdIDs := make([]string, 0, len(created))
	for _, row := range created {
		createdLine := append([]byte(nil), row.line...)
		if _, present := row.fields["notes"]; !present {
			createdLine, err = spliceTopLevel(createdLine, map[string]json.RawMessage{"notes": json.RawMessage(`""`)})
			if err != nil {
				return nil, Evidence{}, fmt.Errorf("add approved empty notes to AFP child %s: %w", row.id, err)
			}
		}
		lines = append(lines, createdLine)
		createdIDs = append(createdIDs, row.id)
	}
	sort.Strings(createdIDs)
	candidate := bytes.Join(lines, []byte("\n"))
	if len(source) > 0 && source[len(source)-1] == '\n' {
		candidate = append(candidate, '\n')
	}
	spans, err := candidateNullSpans(candidate, lossEvidence)
	if err != nil {
		return nil, Evidence{}, err
	}
	counters := afpCounterEvidence(c, b1)
	all := sortedIDs(mergeRows(c, created))
	importNormalizations := approvedAFPImportNormalizations()
	e := Evidence{Version: evidenceVersion, Outcome: "reconciled", SourceSHA256: digest(source), BDBinarySHA256: binary, BaselineSHA256: digest(baseline), UpdatedSHA256: digest(updated), PlanSHA256: digestJSON(plan), CandidateSHA256: digest(candidate), Changes: changes, CreatedIDs: createdIDs, AllIssueIDs: all, RepresentationLosses: append([]RepresentationLoss(nil), lossEvidence...), CandidateNullSpans: spans, AFPRefinement: true, AFPRoles: roleIDs, AFPAddedEdges: expectedEdges, AFPCounterChanges: counters, AFPImportNormalizations: importNormalizations}
	return candidate, e, nil
}

func afpEquivalentField(before, after row, field string) bool {
	left, right := cloneFields(before.fields), cloneFields(after.fields)
	normalizeArrays(left)
	normalizeArrays(right)
	lv, lok := left[field]
	rv, rok := right[field]
	return lok == rok && (!lok || jsonEqual(lv, rv))
}

func resolveAFPRole(id string, roles map[string]string) string {
	if strings.HasPrefix(id, "@") {
		return roles[strings.TrimPrefix(id, "@")]
	}
	return id
}

func identifyAFPCreates(source, updated map[string]row) ([]row, map[string]string, error) {
	added := make([]row, 0, len(afpTemplates))
	roles := make(map[string]string, len(afpTemplates))
	byTitle := make(map[string]afpIssueTemplate, len(afpTemplates))
	for _, t := range afpTemplates {
		byTitle[t.title] = t
	}
	for id, r := range updated {
		if _, exists := source[id]; exists {
			continue
		}
		title, err := stringField(r, "title")
		if err != nil {
			return nil, nil, err
		}
		t, ok := byTitle[title]
		if !ok {
			return nil, nil, fmt.Errorf("unexpected AFP created issue %s title %q", id, title)
		}
		if _, dup := roles[t.role]; dup {
			return nil, nil, fmt.Errorf("duplicate AFP role %s", t.role)
		}
		roles[t.role] = id
		added = append(added, r)
	}
	if len(added) != len(afpTemplates) {
		return nil, nil, fmt.Errorf("AFP batch created %d issues, want exactly %d", len(added), len(afpTemplates))
	}
	for _, t := range afpTemplates {
		if roles[t.role] == "" {
			return nil, nil, fmt.Errorf("AFP batch is missing role %s", t.role)
		}
	}
	sort.Slice(added, func(i, j int) bool { return added[i].id < added[j].id })
	return added, roles, nil
}

type afpDependencyRecord struct {
	edge   AFPEdge
	raw    json.RawMessage
	fields map[string]json.RawMessage
}

func validateAFPGraph(source, baseline, updated map[string]row, additions []AFPEdge, roles map[string]string, windows []AFPCommandWindow) error {
	sourceRecords, err := allAFPDependencyRecords(source)
	if err != nil {
		return err
	}
	before, err := allAFPDependencyRecords(baseline)
	if err != nil {
		return err
	}
	after, err := allAFPDependencyRecords(updated)
	if err != nil {
		return err
	}
	sourceKeys := dependencyKeys(sourceRecords)
	beforeKeys := dependencyKeys(before)
	if !sameKeySet(sourceKeys, beforeKeys) {
		return errors.New("AFP source and baseline dependency graph differ")
	}
	planned := make(map[string]AFPEdge, len(additions))
	for _, edge := range additions {
		key := afpEdgeKey(edge)
		if _, exists := before[key]; exists {
			return fmt.Errorf("AFP planned edge already exists: %s", key)
		}
		if _, exists := planned[key]; exists {
			return fmt.Errorf("AFP planned edge is duplicated: %s", key)
		}
		planned[key] = edge
	}
	windowByEdge := make(map[string]AFPCommandWindow, len(windows))
	for _, window := range windows {
		key := afpEdgeKey(window.Edge)
		if _, exists := planned[key]; !exists || window.StartedAt.IsZero() || window.FinishedAt.Before(window.StartedAt) {
			return fmt.Errorf("invalid AFP Beads command window for added edge %s", key)
		}
		if _, duplicate := windowByEdge[key]; duplicate {
			return fmt.Errorf("duplicate AFP Beads command window for edge %s", key)
		}
		windowByEdge[key] = window
	}
	if len(windowByEdge) != len(planned) {
		return fmt.Errorf("AFP evidence has %d edge command windows, expected %d", len(windowByEdge), len(planned))
	}
	for key, old := range before {
		current, exists := after[key]
		if !exists {
			return fmt.Errorf("AFP dependency graph is missing edge %s", key)
		}
		if !jsonEqualMaps(old.fields, current.fields) {
			return fmt.Errorf("AFP existing dependency edge metadata changed for %s", key)
		}
	}
	for key, current := range after {
		if _, exists := before[key]; exists {
			continue
		}
		edge, ok := planned[key]
		if !ok {
			return fmt.Errorf("AFP dependency graph has unplanned edge %s", key)
		}
		if err := validateAFPAddedDependency(current, edge, windowByEdge[key]); err != nil {
			return fmt.Errorf("AFP added edge %s: %w", key, err)
		}
		for _, createdID := range roles {
			if edge.IssueID != createdID {
				continue
			}
			createdAtText, fieldErr := stringField(updated[createdID], "created_at")
			createdAt, parseErr := parseTimestamp(createdAtText)
			if fieldErr != nil || parseErr != nil || !timestampWithinAFPWindow(createdAt, windowByEdge[key]) {
				return fmt.Errorf("AFP created issue %s timestamp falls outside its Beads create command", createdID)
			}
		}
		delete(planned, key)
	}
	if len(planned) != 0 {
		for key := range planned {
			return fmt.Errorf("AFP dependency graph is missing planned edge %s", key)
		}
	}
	if len(after) != len(before)+len(additions) {
		return fmt.Errorf("AFP dependency graph delta has %d edges, expected %d", len(after)-len(before), len(additions))
	}
	return nil
}

func allAFPDependencyRecords(rows map[string]row) (map[string]afpDependencyRecord, error) {
	result := make(map[string]afpDependencyRecord)
	for _, item := range rows {
		records, err := afpDependencyRecords(item)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			key := afpEdgeKey(record.edge)
			if _, exists := result[key]; exists {
				return nil, fmt.Errorf("duplicate AFP dependency edge %s", key)
			}
			result[key] = record
		}
	}
	return result, nil
}

func afpDependencyRecords(item row) ([]afpDependencyRecord, error) {
	raw, ok := item.fields["dependencies"]
	if !ok || string(bytes.TrimSpace(raw)) == "null" {
		return nil, nil
	}
	var dependencies []json.RawMessage
	if err := json.Unmarshal(raw, &dependencies); err != nil {
		return nil, fmt.Errorf("issue %s dependencies: %w", item.id, err)
	}
	records := make([]afpDependencyRecord, 0, len(dependencies))
	for _, dependency := range dependencies {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(dependency, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("issue %s has malformed dependency record", item.id)
		}
		issue, e1 := stringFieldMap(fields, "issue_id")
		target, e2 := stringFieldMap(fields, "depends_on_id")
		typ, e3 := stringFieldMap(fields, "type")
		if e1 != nil || e2 != nil || e3 != nil || issue != item.id {
			return nil, fmt.Errorf("issue %s has malformed dependency identity", item.id)
		}
		records = append(records, afpDependencyRecord{edge: AFPEdge{IssueID: issue, DependsOnID: target, Type: typ}, raw: append(json.RawMessage(nil), dependency...), fields: fields})
	}
	return records, nil
}

func dependencyKeys(records map[string]afpDependencyRecord) map[string]bool {
	keys := make(map[string]bool, len(records))
	for key := range records {
		keys[key] = true
	}
	return keys
}

func sameKeySet(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if !right[key] {
			return false
		}
	}
	return true
}

func validateAFPAddedDependency(record afpDependencyRecord, expected AFPEdge, window AFPCommandWindow) error {
	if record.edge != expected {
		return errors.New("dependency identity differs from the approved edge")
	}
	if len(record.fields) != 6 {
		return fmt.Errorf("new dependency has %d metadata fields, expected exactly 6", len(record.fields))
	}
	createdBy, err := stringFieldMap(record.fields, "created_by")
	if err != nil || createdBy != "unknown" {
		return errors.New("new dependency created_by differs from pinned Beads output")
	}
	metadata, err := stringFieldMap(record.fields, "metadata")
	if err != nil || metadata != "{}" {
		return errors.New("new dependency metadata differs from pinned empty metadata")
	}
	createdAt, err := stringFieldMap(record.fields, "created_at")
	if err != nil {
		return errors.New("new dependency is missing created_at")
	}
	created, err := parseTimestamp(createdAt)
	if err != nil {
		return errors.New("new dependency created_at is not a valid timestamp")
	}
	// Beads 1.3.1 emits dependency timestamps at whole-second precision using
	// nearest-second rounding. Round the recorded command endpoints with the
	// same rule; flooring the start would incorrectly allow the previous second.
	if !timestampWithinAFPWindow(created, window) {
		return errors.New("new dependency created_at falls outside its recorded Beads command")
	}
	return nil
}

func timestampWithinAFPWindow(timestamp time.Time, window AFPCommandWindow) bool {
	if timestamp.Nanosecond() != 0 {
		return false
	}
	start := window.StartedAt.UTC().Round(time.Second)
	end := window.FinishedAt.UTC().Round(time.Second)
	timestamp = timestamp.UTC().Round(time.Second)
	return !timestamp.Before(start) && !timestamp.After(end)
}

func candidateAFPDependencies(source, updated row, additions []AFPEdge) (json.RawMessage, error) {
	sourceRecords, err := afpDependencyRecords(source)
	if err != nil {
		return nil, err
	}
	updatedRecords, err := afpDependencyRecords(updated)
	if err != nil {
		return nil, err
	}
	original, ok := source.fields["dependencies"]
	if !ok || string(bytes.TrimSpace(original)) == "null" {
		return nil, errors.New("updated AFP issue has no source dependency array")
	}
	var existing []json.RawMessage
	if err := json.Unmarshal(original, &existing); err != nil {
		return nil, err
	}
	capacity, err := checkedAFPDependencyCapacity(len(existing), len(additions))
	if err != nil {
		return nil, err
	}
	all := make([]json.RawMessage, 0, capacity)
	for _, record := range sourceRecords {
		all = append(all, record.raw)
	}
	added := make(map[string]afpDependencyRecord, len(updatedRecords))
	for _, record := range updatedRecords {
		added[afpEdgeKey(record.edge)] = record
	}
	for _, edge := range additions {
		if edge.IssueID != source.id {
			continue
		}
		key := afpEdgeKey(edge)
		record, ok := added[key]
		if !ok {
			return nil, fmt.Errorf("updated Beads export is missing new edge %s", key)
		}
		all = append(all, record.raw)
	}
	var encoded bytes.Buffer
	encoded.WriteByte('[')
	for i, record := range all {
		if i != 0 {
			encoded.WriteByte(',')
		}
		encoded.Write(record)
	}
	encoded.WriteByte(']')
	return encoded.Bytes(), nil
}

func checkedAFPDependencyCapacity(existing, additions int) (int, error) {
	if existing < 0 || additions < 0 {
		return 0, errors.New("AFP dependency capacity cannot be negative")
	}
	maxInt := int(^uint(0) >> 1)
	if additions > maxInt-existing {
		return 0, errors.New("AFP dependency capacity overflows int")
	}
	return existing + additions, nil
}

func afpEdges(rows map[string]row) ([]AFPEdge, error) {
	records, err := allAFPDependencyRecords(rows)
	if err != nil {
		return nil, err
	}
	edges := make([]AFPEdge, 0, len(records))
	for _, record := range records {
		edges = append(edges, record.edge)
	}
	return edges, nil
}

func afpEdgeKey(e AFPEdge) string { return e.IssueID + "\x00" + e.DependsOnID + "\x00" + e.Type }
func edgeSet(edges []AFPEdge) map[string]bool {
	out := make(map[string]bool, len(edges))
	for _, e := range edges {
		out[afpEdgeKey(e)] = true
	}
	return out
}

func validateAFPExistingRows(before, after map[string]row) error {
	allowed := map[string]map[string]bool{
		"dap-jo9.4":     {"description": true, "acceptance_criteria": true},
		"dap-jo9.4.1":   {"title": true, "description": true, "notes": true},
		"dap-jo9.4.1.2": {"title": true, "description": true, "acceptance_criteria": true, "priority": true, "dependencies": true, "dependency_count": true, "dependent_count": true},
		"dap-jo9.4.2":   {"description": true, "acceptance_criteria": true, "priority": true, "dependencies": true, "dependency_count": true, "dependent_count": true},
	}
	for id, b := range before {
		a := after[id]
		permit := allowed[id]
		if permit == nil {
			permit = map[string]bool{}
		}
		if err := compareRows(b, a, id, permit, true, approvedNullLossVersion, nil); err != nil {
			return fmt.Errorf("AFP update issue %s: %w", id, err)
		}
		changed := false
		for f := range permit {
			if !jsonEqual(b.fields[f], a.fields[f]) {
				changed = true
			}
		}
		oldAt, e1 := stringField(b, "updated_at")
		newAt, e2 := stringField(a, "updated_at")
		if e1 != nil || e2 != nil {
			return fmt.Errorf("AFP issue %s is missing updated_at", id)
		}
		if changed && oldAt == newAt {
			return fmt.Errorf("AFP issue %s changed without updated_at", id)
		}
		if !changed && oldAt != newAt {
			return fmt.Errorf("AFP issue %s updated_at changed without an approved field change", id)
		}
	}
	for _, u := range afpTextUpdates {
		for f, want := range u.fields {
			got, err := stringField(after[u.id], f)
			if err != nil || got != want {
				return fmt.Errorf("AFP issue %s.%s does not match the exact approved text", u.id, f)
			}
		}
	}
	for id, want := range afpPriorities {
		var got int
		if err := json.Unmarshal(after[id].fields["priority"], &got); err != nil || got != want {
			return fmt.Errorf("AFP issue %s priority does not match P1", id)
		}
	}
	return nil
}

func validateAFPCreatedRows(roles map[string]string, updated map[string]row) error {
	for _, t := range afpTemplates {
		r := updated[roles[t.role]]
		f := r.fields
		for k := range f {
			if !createFields[k] {
				return fmt.Errorf("AFP created %s has unsupported field %q", t.role, k)
			}
		}
		for _, field := range []string{"_type", "id", "title", "description", "acceptance_criteria", "status", "priority", "issue_type", "created_at", "created_by", "updated_at", "labels", "dependencies", "dependency_count", "dependent_count", "comment_count"} {
			if _, ok := f[field]; !ok {
				return fmt.Errorf("AFP created %s is missing %s", t.role, field)
			}
		}
		if r.id == "" || !strings.HasPrefix(r.id, "dap-") {
			return fmt.Errorf("AFP created %s has invalid Beads ID", t.role)
		}
		for field, want := range map[string]string{"_type": "issue", "title": t.title, "description": t.description, "acceptance_criteria": t.acceptance, "status": "open", "issue_type": t.issueType} {
			got, err := stringField(r, field)
			if err != nil || got != want {
				return fmt.Errorf("AFP created %s.%s differs from approved template", t.role, field)
			}
		}
		if notes, present := f["notes"]; present {
			var got string
			if err := json.Unmarshal(notes, &got); err != nil || got != "" {
				return fmt.Errorf("AFP created %s notes must be empty", t.role)
			}
		}
		var priority int
		if err := json.Unmarshal(f["priority"], &priority); err != nil || priority != t.priority {
			return fmt.Errorf("AFP created %s priority differs from template", t.role)
		}
		var labels []string
		if err := json.Unmarshal(f["labels"], &labels); err != nil || !equalLabelSets(labels, t.labels) {
			return fmt.Errorf("AFP created %s labels differ from template", t.role)
		}
		for _, key := range []string{"created_at", "updated_at"} {
			v, _ := stringField(r, key)
			if _, err := parseTimestamp(v); err != nil {
				return fmt.Errorf("AFP created %s has invalid %s", t.role, key)
			}
		}
		if _, err := stringField(r, "created_by"); err != nil {
			return fmt.Errorf("AFP created %s has invalid created_by", t.role)
		}
		for _, key := range []string{"dependency_count", "dependent_count", "comment_count"} {
			var n int
			if err := json.Unmarshal(f[key], &n); err != nil || n < 0 {
				return fmt.Errorf("AFP created %s has invalid %s", t.role, key)
			}
		}
		var comments int
		if err := json.Unmarshal(f["comment_count"], &comments); err != nil || comments != 0 {
			return fmt.Errorf("AFP created %s has unexpected comments", t.role)
		}
		if err := validateAFPCreateEdges(r, t, roles); err != nil {
			return err
		}
	}
	return nil
}

func validateAFPCreateEdges(r row, t afpIssueTemplate, roles map[string]string) error {
	var edges []map[string]json.RawMessage
	if err := json.Unmarshal(r.fields["dependencies"], &edges); err != nil {
		return fmt.Errorf("AFP created %s has invalid dependencies", t.role)
	}
	want := map[string]string{r.id + "\x00dap-jo9.4": "parent-child"}
	if t.dependsOnRole != "" {
		dep := t.dependsOnRole
		if dep == "decision" {
			dep = "dap-jo9.4.2"
		} else {
			dep = roles[dep]
		}
		want[r.id+"\x00"+dep] = "blocks"
	}
	if len(edges) != len(want) {
		return fmt.Errorf("AFP created %s has unexpected dependency edges", t.role)
	}
	for _, edge := range edges {
		if len(edge) != 6 {
			return fmt.Errorf("AFP created %s edge has unexpected metadata fields", t.role)
		}
		issue, e1 := stringFieldMap(edge, "issue_id")
		target, e2 := stringFieldMap(edge, "depends_on_id")
		typ, e3 := stringFieldMap(edge, "type")
		if e1 != nil || e2 != nil || e3 != nil || issue != r.id || want[issue+"\x00"+target] != typ {
			return fmt.Errorf("AFP created %s has an unplanned edge", t.role)
		}
		for _, field := range []string{"created_at", "created_by", "metadata"} {
			v, e := stringFieldMap(edge, field)
			if e != nil || strings.TrimSpace(v) == "" {
				return fmt.Errorf("AFP created %s edge has invalid %s", t.role, field)
			}
			if field == "metadata" && v != "{}" {
				return fmt.Errorf("AFP created %s edge metadata differs from the expected empty object", t.role)
			}
			if field == "created_at" {
				if _, e := parseTimestamp(v); e != nil {
					return fmt.Errorf("AFP created %s edge timestamp is invalid", t.role)
				}
			}
		}
	}
	return nil
}

func equalLabelSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func validateAFPCounters(source, updated map[string]row) error {
	counts := make(map[string][2]int, len(updated))
	edges, err := afpEdges(updated)
	if err != nil {
		return err
	}
	for id := range updated {
		counts[id] = [2]int{}
	}
	for _, e := range edges {
		if e.Type != "blocks" {
			continue
		}
		from := counts[e.IssueID]
		from[0]++
		counts[e.IssueID] = from
		to := counts[e.DependsOnID]
		to[1]++
		counts[e.DependsOnID] = to
	}
	for id, r := range updated {
		want := counts[id]
		var dep, dependent int
		if err := json.Unmarshal(r.fields["dependency_count"], &dep); err != nil {
			return fmt.Errorf("AFP issue %s has invalid dependency_count", id)
		}
		if err := json.Unmarshal(r.fields["dependent_count"], &dependent); err != nil {
			return fmt.Errorf("AFP issue %s has invalid dependent_count", id)
		}
		if dep != want[0] || dependent != want[1] {
			return fmt.Errorf("AFP issue %s counters %d/%d disagree with graph %d/%d", id, dep, dependent, want[0], want[1])
		}
	}
	for id, b := range source {
		a := updated[id]
		var bd, be, ad, ae int
		_ = json.Unmarshal(b.fields["dependency_count"], &bd)
		_ = json.Unmarshal(b.fields["dependent_count"], &be)
		_ = json.Unmarshal(a.fields["dependency_count"], &ad)
		_ = json.Unmarshal(a.fields["dependent_count"], &ae)
		wantDep, wantDependent := bd, be
		switch id {
		case "dap-jo9.4.1.2":
			wantDep++
			wantDependent++
		case "dap-jo9.4.2":
			wantDep += 2
			wantDependent++
		}
		if ad != wantDep || ae != wantDependent {
			return fmt.Errorf("AFP existing issue %s counter delta %d/%d -> %d/%d is unplanned", id, bd, be, ad, ae)
		}
	}
	return nil
}

func afpCounterEvidence(source, updated map[string]row) []AFPCounterChange {
	var out []AFPCounterChange
	for id, b := range source {
		a := updated[id]
		for _, field := range []string{"dependency_count", "dependent_count"} {
			var before, after int
			_ = json.Unmarshal(b.fields[field], &before)
			_ = json.Unmarshal(a.fields[field], &after)
			if before != after {
				out = append(out, AFPCounterChange{IssueID: id, Field: field, Before: before, After: after})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IssueID == out[j].IssueID {
			return out[i].Field < out[j].Field
		}
		return out[i].IssueID < out[j].IssueID
	})
	return out
}

func validateAFPRoundTrip(candidate, roundTrip []byte, evidence Evidence) (Evidence, error) {
	c, err := parseExport("AFP candidate", candidate)
	if err != nil {
		return Evidence{}, err
	}
	rt, err := parseExport("AFP round-trip", roundTrip)
	if err != nil {
		return Evidence{}, err
	}
	if len(c) != len(rt) {
		return Evidence{}, errors.New("AFP second-store round-trip changed the issue ID set")
	}
	allowed := lossMap(evidence.RepresentationLosses)
	for id, cr := range c {
		rr, ok := rt[id]
		if !ok {
			return Evidence{}, fmt.Errorf("AFP round-trip is missing issue %s", id)
		}
		cn, rn, _, normErr := normalizeAFPImportRows(cr, rr)
		if normErr != nil {
			return Evidence{}, normErr
		}
		if err := compareRows(cn, rn, id, nil, false, approvedNullLossVersion, allowed[id]); err != nil {
			return Evidence{}, fmt.Errorf("AFP candidate round-trip issue %s: %w", id, err)
		}
	}
	ce, err := afpEdges(c)
	if err != nil {
		return Evidence{}, err
	}
	re, err := afpEdges(rt)
	if err != nil {
		return Evidence{}, err
	}
	cs, rs := edgeSet(ce), edgeSet(re)
	if len(cs) != len(rs) {
		return Evidence{}, errors.New("AFP round-trip dependency graph changed")
	}
	for k := range cs {
		if !rs[k] {
			return Evidence{}, fmt.Errorf("AFP round-trip dependency graph is missing edge %s", k)
		}
	}
	spans, err := candidateNullSpans(candidate, evidence.RepresentationLosses)
	if err != nil {
		return Evidence{}, err
	}
	return Evidence{Version: evidenceVersion, Outcome: "validated", AllIssueIDs: sortedIDs(c), RepresentationLosses: append([]RepresentationLoss(nil), evidence.RepresentationLosses...), CandidateNullSpans: spans, AFPRefinement: true, AFPAddedEdges: append([]AFPEdge(nil), evidence.AFPAddedEdges...), AFPCounterChanges: append([]AFPCounterChange(nil), evidence.AFPCounterChanges...), AFPImportNormalizations: append([]AFPImportNormalization(nil), evidence.AFPImportNormalizations...)}, nil
}

func normalizeAFPImportRows(source, imported row) (row, row, []AFPImportNormalization, error) {
	left := source
	left.fields = cloneFields(source.fields)
	right := imported
	right.fields = cloneFields(imported.fields)
	pairs, ok := afpImportTimestampPairs[source.id]
	if !ok {
		return left, right, nil, nil
	}
	var normalizations []AFPImportNormalization
	for field, sourceWant := range pairs {
		leftValue, e1 := stringField(source, field)
		rightValue, e2 := stringField(imported, field)
		_, parseErr := parseTimestamp(sourceWant)
		if e1 != nil || e2 != nil || parseErr != nil || leftValue != sourceWant {
			return row{}, row{}, nil, fmt.Errorf("AFP source %s.%s does not match the approved timestamp spelling", source.id, field)
		}
		want := afpImportedTimestampValues[source.id][field]
		if rightValue != want {
			return row{}, row{}, nil, fmt.Errorf("Beads AFP import %s.%s = %q, want exact normalized value %q", source.id, field, rightValue, want)
		}
		encoded, _ := json.Marshal(want)
		left.fields[field] = encoded
		normalizations = append(normalizations, AFPImportNormalization{IssueID: source.id, Field: field, SourceValue: sourceWant, ImportedValue: want})
	}
	sort.Slice(normalizations, func(i, j int) bool { return normalizations[i].Field < normalizations[j].Field })
	return left, right, normalizations, nil
}

func approvedAFPImportNormalizations() []AFPImportNormalization {
	var result []AFPImportNormalization
	for id, fields := range afpImportTimestampPairs {
		for field, source := range fields {
			result = append(result, AFPImportNormalization{IssueID: id, Field: field, SourceValue: source, ImportedValue: afpImportedTimestampValues[id][field]})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IssueID == result[j].IssueID {
			return result[i].Field < result[j].Field
		}
		return result[i].IssueID < result[j].IssueID
	})
	return result
}
