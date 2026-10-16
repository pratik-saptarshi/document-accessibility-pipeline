// Package beadsreconcile validates three Beads exports and builds a candidate
// interchange file by splicing only explicitly planned values into the source.
package beadsreconcile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const evidenceVersion = 2

const (
	approvedNullLossHistoricSourceSHA256      = "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890"
	approvedNullLossCurrentSourceSHA256       = "768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de"
	approvedNullLossIssueCreationSourceSHA256 = "5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465"
	approvedNullLossBDSHA256                  = "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653"
	approvedNullLossVersion                   = "1.3.1"
	approvedClosureSourceSHA256               = "5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465"
	approvedClosureBDSHA256                   = "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653"
	approvedClosureReason                     = "Acceptance evidence: merged PRs #89, #90, #92, #93, #94, and #96. This bundle records the pinned Beads close event, source snapshots, command provenance, and candidate."
)

var approvedClosureEvidenceRefs = []string{"#89", "#90", "#92", "#93", "#94", "#96"}

// These are the only source snapshots approved for the five explicit-null
// losses below. Adding another digest requires separate review and approval.
var approvedNullLossSources = map[string]bool{
	approvedNullLossHistoricSourceSHA256:      true,
	approvedNullLossCurrentSourceSHA256:       true,
	approvedNullLossIssueCreationSourceSHA256: true,
}

var approvedNullLossPairs = []fieldPair{
	{issueID: "dap-jo9.5.3", field: "dependencies"},
	{issueID: "dap-jo9.1.6.10", field: "closed_at"},
	{issueID: "dap-jo9.5.6", field: "closed_at"},
	{issueID: "dap-jo9.2.1.4", field: "closed_at"},
	{issueID: "dap-jo9.5.7", field: "closed_at"},
}

type fieldPair struct {
	issueID string
	field   string
}

var allowedFields = map[string]bool{
	"title": true, "description": true, "acceptance_criteria": true, "notes": true,
}

// Plan describes exact existing-issue updates and bounded complete issue rows
// imported into the disposable store for evidence capture.
type Plan struct {
	Updates []IssueUpdate `json:"updates"`
	// Creates lists complete, canonical rows that the importer added to B1.
	// Creation is currently bounded to one synthetic direct child of dap-jo9.
	Creates []json.RawMessage `json:"creates,omitempty"`
	// Closes is a single target- and source-bound operation authorized by
	// Decision 0008. It is intentionally separate from generic field updates.
	Closes []IssueClosure `json:"closes,omitempty"`
	// BDVersion pins narrowly accepted importer behavior. The sole derived-field
	// exception is only recognized for Beads 1.3.1.
	BDVersion string `json:"bd_version,omitempty"`
}

type IssueClosure struct {
	ID                   string   `json:"id"`
	ExpectedSourceSHA256 string   `json:"expected_source_sha256"`
	ExpectedStatus       string   `json:"expected_status"`
	DecisionID           string   `json:"decision_id"`
	Reason               string   `json:"reason"`
	EvidenceRefs         []string `json:"evidence_refs"`
}

type IssueUpdate struct {
	ID     string            `json:"id"`
	Fields map[string]string `json:"fields"`
}

// Evidence binds the candidate to all three immutable exports and the plan.
type Evidence struct {
	Version              int                  `json:"version"`
	Outcome              string               `json:"outcome"`
	SourceSHA256         string               `json:"source_sha256"`
	BDBinarySHA256       string               `json:"bd_binary_sha256,omitempty"`
	BaselineSHA256       string               `json:"baseline_sha256"`
	UpdatedSHA256        string               `json:"updated_sha256"`
	PlanSHA256           string               `json:"plan_sha256"`
	CandidateSHA256      string               `json:"candidate_sha256"`
	RoundTripSHA256      string               `json:"round_trip_sha256,omitempty"`
	Changes              []EvidenceChange     `json:"changes"`
	IssueIDs             []string             `json:"issue_ids"`
	CreatedIDs           []string             `json:"created_ids,omitempty"`
	Closures             []EvidenceClosure    `json:"closures,omitempty"`
	AllIssueIDs          []string             `json:"all_issue_ids"`
	RepresentationLosses []RepresentationLoss `json:"representation_losses,omitempty"`
	CandidateNullSpans   []NullSpan           `json:"candidate_null_spans,omitempty"`
}

type EvidenceClosure struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	ClosedAt    string `json:"closed_at"`
	CloseReason string `json:"close_reason"`
	UpdatedAt   string `json:"updated_at"`
}

// RepresentationLoss records only the approved explicit-null importer delta.
// The source and candidate digests bind the values; snapshot digests bind the
// baseline, updated export, and second-store round trip.
type RepresentationLoss struct {
	IssueID            string `json:"issue_id"`
	Field              string `json:"field"`
	CanonicalValue     string `json:"canonical_value"`
	BaselineAbsent     bool   `json:"baseline_absent"`
	UpdatedAbsent      bool   `json:"updated_absent"`
	CandidateValue     string `json:"candidate_value"`
	RoundTripAbsent    bool   `json:"round_trip_absent"`
	CanonicalSourceSHA string `json:"canonical_source_sha256"`
}

// NullSpan identifies the exact JSON null value in a derived candidate line.
// It is recomputed from candidate bytes during second-store validation.
type NullSpan struct {
	IssueID string `json:"issue_id"`
	Field   string `json:"field"`
	Line    int    `json:"line"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Raw     string `json:"raw"`
}

type EvidenceChange struct {
	ID        string            `json:"id"`
	Fields    map[string]string `json:"fields"`
	UpdatedAt string            `json:"updated_at"`
}

type row struct {
	fields map[string]json.RawMessage
	line   []byte
	id     string
}

// Reconcile validates source C, pre-update export B0, post-update export B1,
// then returns candidate bytes derived from C. No export bytes are copied into
// the candidate except planned strings and Beads' updated_at values.
func Reconcile(source, baseline, updated []byte, plan Plan) ([]byte, Evidence, error) {
	return reconcile(source, baseline, updated, plan, "", nil, nil)
}

// ReconcileWithBinary applies the reviewed explicit-null compatibility
// exception only when the canonical source is one of the exact approved
// snapshots and the verified Beads binary matches its pinned digest. All other
// importer deltas remain rejected.
func ReconcileWithBinary(source, baseline, updated []byte, plan Plan, bdBinarySHA256 string) ([]byte, Evidence, error) {
	if len(plan.Closes) > 0 && bdBinarySHA256 != approvedClosureBDSHA256 {
		return nil, Evidence{}, errors.New("closure requires the exact Beads 1.3.1 executable pin")
	}
	allowed, losses, err := authorizeNullLoss(source, baseline, updated, plan.BDVersion, bdBinarySHA256)
	if err != nil {
		return nil, Evidence{}, err
	}
	return reconcile(source, baseline, updated, plan, bdBinarySHA256, allowed, losses)
}

// ValidateCandidateRoundTrip independently derives the candidate from C/B0/B1
// and the plan before applying the null-loss exception to the second-store
// round trip. Callers cannot authorize arbitrary candidate digests or spans.
func ValidateCandidateRoundTrip(source, baseline, updated, candidate, roundTrip []byte, plan Plan, bdBinarySHA256 string) (Evidence, Evidence, error) {
	derived, evidence, err := ReconcileWithBinary(source, baseline, updated, plan, bdBinarySHA256)
	if err != nil {
		return Evidence{}, Evidence{}, err
	}
	if !bytes.Equal(derived, candidate) {
		return Evidence{}, Evidence{}, errors.New("second-store candidate is not the derived candidate")
	}
	allowed := lossMap(evidence.RepresentationLosses)
	validationCandidate, validation, err := reconcile(candidate, roundTrip, roundTrip, Plan{BDVersion: plan.BDVersion}, bdBinarySHA256, allowed, evidence.RepresentationLosses)
	if err != nil {
		return Evidence{}, Evidence{}, err
	}
	if !bytes.Equal(validationCandidate, candidate) {
		return Evidence{}, Evidence{}, errors.New("second-store validation candidate changed bytes")
	}
	spans, err := candidateNullSpans(candidate, evidence.RepresentationLosses)
	if err != nil {
		return Evidence{}, Evidence{}, err
	}
	for i := range evidence.RepresentationLosses {
		evidence.RepresentationLosses[i].RoundTripAbsent = true
	}
	validation.RepresentationLosses = append([]RepresentationLoss(nil), evidence.RepresentationLosses...)
	validation.Closures = append([]EvidenceClosure(nil), evidence.Closures...)
	validation.CandidateNullSpans = spans
	validation.SourceSHA256 = digest(candidate)
	validation.CandidateSHA256 = digest(candidate)
	return evidence, validation, nil
}

func reconcile(source, baseline, updated []byte, plan Plan, bdBinarySHA256 string, nullLosses map[string]map[string]bool, lossEvidence []RepresentationLoss) ([]byte, Evidence, error) {
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
	updates, err := validatePlan(plan, c)
	if err != nil {
		return nil, Evidence{}, err
	}
	if len(plan.Closes) > 0 && digest(source) != approvedClosureSourceSHA256 {
		return nil, Evidence{}, errors.New("closure source bytes do not match the exact authorized snapshot")
	}
	if err := validateIDs(c, b0, b1, plan); err != nil {
		return nil, Evidence{}, err
	}

	for id, cr := range c {
		if err := compareRows(cr, b0[id], id, nil, false, plan.BDVersion, nullLosses[id]); err != nil {
			return nil, Evidence{}, fmt.Errorf("source to baseline issue %s: %w", id, err)
		}
		var allowed map[string]bool
		if update, ok := updates[id]; ok {
			allowed = make(map[string]bool, len(update.Fields))
			for name := range update.Fields {
				allowed[name] = true
			}
		}
		if len(plan.Closes) == 1 && plan.Closes[0].ID == id {
			allowed = map[string]bool{"status": true, "closed_at": true, "close_reason": true}
		}
		if err := compareRows(b0[id], b1[id], id, allowed, true, plan.BDVersion, nil); err != nil {
			return nil, Evidence{}, fmt.Errorf("baseline to updated issue %s: %w", id, err)
		}
		if update, ok := updates[id]; ok {
			for name, want := range update.Fields {
				got, err := stringField(b1[id], name)
				if err != nil || got != want {
					return nil, Evidence{}, fmt.Errorf("issue %s planned field %q does not match requested value", id, name)
				}
				before, beforeErr := stringField(b0[id], name)
				if beforeErr == nil && before == want {
					return nil, Evidence{}, fmt.Errorf("issue %s planned field %q was not changed", id, name)
				}
			}
			oldAt, oldErr := stringField(b0[id], "updated_at")
			newAt, newErr := stringField(b1[id], "updated_at")
			if oldErr != nil || newErr != nil || oldAt == newAt {
				return nil, Evidence{}, fmt.Errorf("issue %s updated_at was not changed", id)
			}
			if _, err := parseTimestamp(newAt); err != nil {
				return nil, Evidence{}, fmt.Errorf("issue %s updated_at is invalid: %w", id, err)
			}
		}
		if len(plan.Closes) == 1 && plan.Closes[0].ID == id {
			if err := validateClosedTransition(plan.Closes[0], b0[id], b1[id]); err != nil {
				return nil, Evidence{}, err
			}
		}
	}
	creates, err := validateCreates(plan, c, b1)
	if err != nil {
		return nil, Evidence{}, err
	}

	lines := splitLines(source)
	changes := make([]EvidenceChange, 0, len(updates))
	ids := make([]string, 0, len(updates))
	for index, original := range lines {
		id, err := rowID(original)
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("source line %d: %w", index+1, err)
		}
		u, ok := updates[id]
		if !ok {
			continue
		}
		changesForID := make(map[string]string, len(u.Fields))
		for name, value := range u.Fields {
			changesForID[name] = value
		}
		updatedAt, err := stringField(b1[id], "updated_at")
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("issue %s: %w", id, err)
		}
		changed, err := spliceFields(original, changesForID, "updated_at", updatedAt)
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("issue %s: %w", id, err)
		}
		lines[index] = changed
		changes = append(changes, EvidenceChange{ID: id, Fields: changesForID, UpdatedAt: updatedAt})
		ids = append(ids, id)
	}
	closures := make([]EvidenceClosure, 0, len(plan.Closes))
	for _, closure := range plan.Closes {
		b1row := b1[closure.ID]
		status, _ := stringField(b1row, "status")
		closedAt, _ := stringField(b1row, "closed_at")
		reason, _ := stringField(b1row, "close_reason")
		updatedAt, _ := stringField(b1row, "updated_at")
		closedLine, err := spliceFields(c[closure.ID].line, map[string]string{
			"status": status, "closed_at": closedAt, "close_reason": reason,
		}, "updated_at", updatedAt)
		if err != nil {
			return nil, Evidence{}, fmt.Errorf("issue %s: %w", closure.ID, err)
		}
		for index, original := range lines {
			id, err := rowID(original)
			if err != nil {
				return nil, Evidence{}, err
			}
			if id == closure.ID {
				lines[index] = closedLine
				break
			}
		}
		closures = append(closures, EvidenceClosure{ID: closure.ID, Status: status, ClosedAt: closedAt, CloseReason: reason, UpdatedAt: updatedAt})
	}
	createdIDs := make([]string, 0, len(creates))
	for _, created := range creates {
		lines = append(lines, append([]byte(nil), created.line...))
		createdIDs = append(createdIDs, created.id)
	}
	candidate := bytes.Join(lines, []byte("\n"))
	if len(source) > 0 && source[len(source)-1] == '\n' {
		candidate = append(candidate, '\n')
	}
	e := Evidence{
		Version: evidenceVersion, Outcome: "reconciled",
		SourceSHA256: digest(source), BaselineSHA256: digest(baseline), UpdatedSHA256: digest(updated),
		PlanSHA256: digestJSON(plan), CandidateSHA256: digest(candidate), Changes: changes, IssueIDs: ids, CreatedIDs: createdIDs,
		AllIssueIDs: sortedIDs(mergeRows(c, creates)), Closures: closures,
		BDBinarySHA256:       bdBinarySHA256,
		RepresentationLosses: append([]RepresentationLoss(nil), lossEvidence...),
	}
	if len(e.RepresentationLosses) > 0 {
		spans, err := candidateNullSpans(candidate, e.RepresentationLosses)
		if err != nil {
			return nil, Evidence{}, err
		}
		e.CandidateNullSpans = spans
	}
	return candidate, e, nil
}

// ValidatePlan rejects unknown, duplicate, or missing issue/field targets
// before a caller applies any plan through Beads.
func ValidatePlan(source []byte, plan Plan) error {
	if len(plan.Closes) > 0 && digest(source) != approvedClosureSourceSHA256 {
		return errors.New("closure source bytes do not match the exact authorized snapshot")
	}
	rows, err := parseExport("source", source)
	if err != nil {
		return err
	}
	_, err = validatePlan(plan, rows)
	return err
}

func sortedIDs(rows map[string]row) []string {
	ids := make([]string, 0, len(rows))
	for id := range rows {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func parseExport(name string, data []byte) (map[string]row, error) {
	result := make(map[string]row)
	for i, line := range splitLines(data) {
		if len(bytes.TrimSpace(line)) == 0 {
			return nil, fmt.Errorf("%s line %d is empty", name, i+1)
		}
		if err := rejectDuplicateKeys(line); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", name, i+1, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil || fields == nil {
			if err == nil {
				err = errors.New("row must be a JSON object")
			}
			return nil, fmt.Errorf("%s line %d: %w", name, i+1, err)
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil || strings.TrimSpace(id) == "" {
			return nil, fmt.Errorf("%s line %d: missing string issue id", name, i+1)
		}
		if _, exists := result[id]; exists {
			return nil, fmt.Errorf("%s: duplicate issue id %q", name, id)
		}
		result[id] = row{fields: fields, line: append([]byte(nil), line...), id: id}
	}
	if err := validateGraph(result); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return result, nil
}

func validateGraph(rows map[string]row) error {
	for id, item := range rows {
		raw, ok := item.fields["dependencies"]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		var dependencies []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &dependencies); err != nil {
			return fmt.Errorf("issue %s dependencies must be an array", id)
		}
		seen := make(map[string]bool, len(dependencies))
		for _, dependency := range dependencies {
			if dependency == nil {
				return fmt.Errorf("issue %s has a non-object dependency", id)
			}
			edgeIssue, e1 := stringFieldMap(dependency, "issue_id")
			target, e2 := stringFieldMap(dependency, "depends_on_id")
			edgeType, e3 := stringFieldMap(dependency, "type")
			if e1 != nil || e2 != nil || e3 != nil || edgeIssue != id || strings.TrimSpace(edgeType) == "" {
				return fmt.Errorf("issue %s has an invalid dependency edge", id)
			}
			if _, exists := rows[target]; !exists {
				return fmt.Errorf("issue %s dependency target %q is missing", id, target)
			}
			canonical, _ := json.Marshal([3]string{edgeIssue, target, edgeType})
			if seen[string(canonical)] {
				return fmt.Errorf("issue %s has a duplicate dependency edge", id)
			}
			seen[string(canonical)] = true
		}
	}
	return nil
}

func splitLines(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}
	parts := bytes.Split(data, []byte("\n"))
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		return parts[:len(parts)-1]
	}
	return parts
}

func validateIDs(c, b0, b1 map[string]row, plan Plan) error {
	if len(c) != len(b0) || len(b1) != len(c)+len(plan.Creates) {
		return fmt.Errorf("issue ID set changed: source=%d baseline=%d updated=%d planned creates=%d", len(c), len(b0), len(b1), len(plan.Creates))
	}
	for id := range c {
		if _, ok := b0[id]; !ok {
			return fmt.Errorf("issue ID %q missing from baseline", id)
		}
		if _, ok := b1[id]; !ok {
			return fmt.Errorf("issue ID %q missing from updated export", id)
		}
	}
	for _, raw := range plan.Creates {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
			return errors.New("planned create must be a complete issue object")
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil {
			return errors.New("planned create is missing its string id")
		}
		if _, exists := c[id]; exists {
			return fmt.Errorf("planned create id %q collides with source", id)
		}
		if _, exists := b1[id]; !exists {
			return fmt.Errorf("planned create id %q missing from updated export", id)
		}
	}
	return nil
}

func validatePlan(plan Plan, source map[string]row) (map[string]IssueUpdate, error) {
	if len(plan.Closes) > 1 {
		return nil, errors.New("at most one planned close is supported")
	}
	if len(plan.Closes) > 0 {
		if len(plan.Updates) > 0 || len(plan.Creates) > 0 {
			return nil, errors.New("closure cannot be combined with updates or creates")
		}
		if err := validateClosure(plan, source); err != nil {
			return nil, err
		}
	}
	if len(plan.Creates) > 1 {
		return nil, errors.New("at most one planned create is supported")
	}
	if len(plan.Creates) > 0 && len(plan.Updates) > 0 {
		return nil, errors.New("creation and updates cannot be combined in one plan")
	}
	seenCreates := make(map[string]bool, len(plan.Creates))
	for _, raw := range plan.Creates {
		if err := rejectDuplicateKeys(raw); err != nil {
			return nil, fmt.Errorf("planned create: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
			return nil, errors.New("planned create must be a complete issue object")
		}
		id, err := stringFieldMap(fields, "id")
		if err != nil {
			return nil, errors.New("planned create is missing its string id")
		}
		if seenCreates[id] {
			return nil, fmt.Errorf("duplicate planned create id %q", id)
		}
		seenCreates[id] = true
		if _, exists := source[id]; exists {
			return nil, fmt.Errorf("planned create id %q collides with source", id)
		}
		if err := validateCreateRecord(id, fields); err != nil {
			return nil, err
		}
		if _, exists := source["dap-jo9"]; !exists {
			return nil, errors.New("planned create parent dap-jo9 is not present in source")
		}
	}
	updates := make(map[string]IssueUpdate, len(plan.Updates))
	for _, u := range plan.Updates {
		if _, exists := updates[u.ID]; exists {
			return nil, fmt.Errorf("duplicate plan issue id %q", u.ID)
		}
		if _, exists := source[u.ID]; !exists {
			return nil, fmt.Errorf("plan issue %q not present in source", u.ID)
		}
		if len(u.Fields) == 0 {
			return nil, fmt.Errorf("plan issue %q has no fields", u.ID)
		}
		for field := range u.Fields {
			if !allowedFields[field] {
				return nil, fmt.Errorf("plan field %q is not allowed", field)
			}
		}
		updates[u.ID] = u
	}
	return updates, nil
}

func validateClosure(plan Plan, source map[string]row) error {
	c := plan.Closes[0]
	if c.ID != "dap-jo9.18" || c.ExpectedSourceSHA256 != approvedClosureSourceSHA256 ||
		plan.BDVersion != approvedNullLossVersion || c.DecisionID != "0008" || c.ExpectedStatus != "open" {
		return errors.New("closure does not match the exact target, source, Beads version, and Decision 0008")
	}
	if c.Reason != approvedClosureReason || len([]byte(c.Reason)) > 64*1024 || !utf8.ValidString(c.Reason) {
		return errors.New("closure reason does not match the approved evidence-bearing reason")
	}
	if len(c.EvidenceRefs) != len(approvedClosureEvidenceRefs) {
		return errors.New("closure must cite the exact approved acceptance evidence references")
	}
	for i := range approvedClosureEvidenceRefs {
		if c.EvidenceRefs[i] != approvedClosureEvidenceRefs[i] {
			return errors.New("closure must cite the exact approved acceptance evidence references")
		}
	}
	r, ok := source[c.ID]
	if !ok {
		return errors.New("closure target is not present in source")
	}
	status, err := stringField(r, "status")
	if err != nil || status != c.ExpectedStatus {
		return errors.New("closure target source status is not open")
	}
	for _, field := range []string{"closed_at", "close_reason"} {
		if _, exists := r.fields[field]; exists {
			return fmt.Errorf("closure target source already has %s", field)
		}
	}
	return nil
}

func validateClosedTransition(closure IssueClosure, before, after row) error {
	status, statusErr := stringField(after, "status")
	reason, reasonErr := stringField(after, "close_reason")
	closedAt, closedAtErr := stringField(after, "closed_at")
	oldAt, oldAtErr := stringField(before, "updated_at")
	newAt, newAtErr := stringField(after, "updated_at")
	if statusErr != nil || status != "closed" || reasonErr != nil || reason != closure.Reason {
		return errors.New("Beads close result does not match the exact planned status and reason")
	}
	if _, exists := before.fields["closed_at"]; exists {
		return errors.New("closure baseline already has closed_at")
	}
	if _, exists := before.fields["close_reason"]; exists {
		return errors.New("closure baseline already has close_reason")
	}
	if closedAtErr != nil {
		return errors.New("Beads close result is missing its closed_at timestamp")
	}
	closedTime, err := parseTimestamp(closedAt)
	if err != nil || !strings.HasSuffix(closedAt, "Z") || closedTime.Format(time.RFC3339Nano) != closedAt {
		return errors.New("Beads close result closed_at must be a canonical UTC timestamp")
	}
	if oldAtErr != nil || newAtErr != nil || oldAt == newAt {
		return errors.New("Beads close result did not change updated_at")
	}
	updatedTime, err := parseTimestamp(newAt)
	if err != nil || !strings.HasSuffix(newAt, "Z") || updatedTime.Format(time.RFC3339Nano) != newAt {
		return errors.New("Beads close result updated_at must be a canonical UTC timestamp")
	}
	return nil
}

var createFields = map[string]bool{
	"_type": true, "id": true, "title": true, "description": true, "acceptance_criteria": true,
	"notes": true, "status": true, "priority": true, "issue_type": true, "created_at": true,
	"created_by": true, "updated_at": true, "labels": true, "dependencies": true,
	"dependency_count": true, "dependent_count": true, "comment_count": true,
}

func validateCreates(plan Plan, source, updated map[string]row) ([]row, error) {
	if len(plan.Creates) == 0 {
		return nil, nil
	}
	created := make([]row, 0, len(plan.Creates))
	seen := make(map[string]bool, len(plan.Creates))
	for _, raw := range plan.Creates {
		if err := rejectDuplicateKeys(raw); err != nil {
			return nil, fmt.Errorf("planned create: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
			return nil, errors.New("planned create must be a complete issue object")
		}
		for field := range fields {
			if !createFields[field] {
				return nil, fmt.Errorf("planned create has unsupported field %q", field)
			}
		}
		id, err := stringFieldMap(fields, "id")
		suffix := strings.TrimPrefix(id, "dap-jo9.")
		validSuffix := suffix != id && suffix != ""
		for _, r := range suffix {
			if r < '0' || r > '9' {
				validSuffix = false
			}
		}
		if err != nil || !validSuffix {
			return nil, errors.New("planned create id must be a direct dap-jo9 child")
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate planned create id %q", id)
		}
		seen[id] = true
		if _, exists := source[id]; exists {
			return nil, fmt.Errorf("planned create id %q collides with source", id)
		}
		if _, exists := updated[id]; !exists {
			return nil, fmt.Errorf("planned create id %q missing from updated export", id)
		}
		if err := validateCreateRecord(id, fields); err != nil {
			return nil, err
		}
		actual := updated[id]
		if !jsonEqualMaps(normalizeCreateFields(fields), normalizeCreateFields(actual.fields)) {
			return nil, fmt.Errorf("created issue %s differs from its exact planned record", id)
		}
		created = append(created, actual)
	}
	sort.Slice(created, func(i, j int) bool { return created[i].id < created[j].id })
	return created, nil
}

func validateCreateRecord(id string, fields map[string]json.RawMessage) error {
	for field := range fields {
		if !createFields[field] {
			return fmt.Errorf("planned create has unsupported field %q", field)
		}
	}
	for field := range createFields {
		if field == "labels" {
			continue
		}
		if _, ok := fields[field]; !ok {
			return fmt.Errorf("planned create is missing required canonical field %q", field)
		}
	}
	suffix := strings.TrimPrefix(id, "dap-jo9.")
	validSuffix := suffix != id && suffix != ""
	for _, r := range suffix {
		if r < '0' || r > '9' {
			validSuffix = false
		}
	}
	if !validSuffix {
		return errors.New("planned create id must be a direct dap-jo9 child")
	}
	for _, key := range []string{"_type", "title", "description", "acceptance_criteria", "notes", "status", "issue_type", "created_at", "created_by", "updated_at"} {
		value, err := stringFieldMap(fields, key)
		if err != nil || (key != "description" && key != "acceptance_criteria" && key != "notes" && strings.TrimSpace(value) == "") {
			return fmt.Errorf("planned create %s has invalid %s", id, key)
		}
	}
	if value, err := stringFieldMap(fields, "_type"); err != nil || value != "issue" {
		return fmt.Errorf("planned create %s must have _type issue", id)
	}
	for _, key := range []string{"created_at", "updated_at"} {
		value, _ := stringFieldMap(fields, key)
		if _, err := parseTimestamp(value); err != nil {
			return fmt.Errorf("planned create %s has invalid %s timestamp", id, key)
		}
	}
	for _, key := range []string{"priority", "dependency_count", "dependent_count", "comment_count"} {
		raw := bytes.TrimSpace(fields[key])
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			return fmt.Errorf("planned create %s has invalid %s", id, key)
		}
		var value int
		if err := json.Unmarshal(raw, &value); err != nil || value < 0 || (key == "priority" && value > 4) || (key != "priority" && value != 0) {
			return fmt.Errorf("planned create %s has invalid %s", id, key)
		}
	}
	if raw, ok := fields["labels"]; ok {
		var labels []string
		if err := json.Unmarshal(raw, &labels); err != nil || labels == nil {
			return fmt.Errorf("planned create %s labels must be an array", id)
		}
	}
	var edges []map[string]json.RawMessage
	if err := json.Unmarshal(fields["dependencies"], &edges); err != nil || len(edges) != 1 {
		return fmt.Errorf("planned create %s must have exactly one parent-child edge", id)
	}
	edge := edges[0]
	issue, e1 := stringFieldMap(edge, "issue_id")
	parent, e2 := stringFieldMap(edge, "depends_on_id")
	typ, e3 := stringFieldMap(edge, "type")
	if e1 != nil || e2 != nil || e3 != nil || issue != id || parent != "dap-jo9" || typ != "parent-child" || len(edge) != 6 {
		return fmt.Errorf("planned create %s must have its exact direct parent-child edge", id)
	}
	for _, key := range []string{"created_at", "created_by", "metadata"} {
		if _, ok := edge[key]; !ok {
			return fmt.Errorf("planned create %s edge is missing %s", id, key)
		}
	}
	for _, key := range []string{"created_at", "created_by", "metadata"} {
		value, err := stringFieldMap(edge, key)
		if err != nil || strings.TrimSpace(value) == "" {
			return fmt.Errorf("planned create %s edge has invalid %s", id, key)
		}
		if key == "created_at" {
			if _, err := parseTimestamp(value); err != nil {
				return fmt.Errorf("planned create %s edge has invalid created_at timestamp", id)
			}
		}
		if key == "metadata" && !json.Valid([]byte(value)) {
			return fmt.Errorf("planned create %s edge metadata is invalid JSON", id)
		}
	}
	return nil
}

func normalizeCreateFields(fields map[string]json.RawMessage) map[string]json.RawMessage {
	result := cloneFields(fields)
	if raw, ok := result["labels"]; ok {
		trimmed := bytes.TrimSpace(raw)
		var labels []json.RawMessage
		if len(trimmed) > 0 && trimmed[0] == '[' && json.Unmarshal(trimmed, &labels) == nil && len(labels) == 0 {
			delete(result, "labels")
		}
	}
	normalizeArrays(result)
	return result
}

func mergeRows(source map[string]row, creates []row) map[string]row {
	result := make(map[string]row, len(source)+len(creates))
	for id, item := range source {
		result[id] = item
	}
	for _, item := range creates {
		result[item.id] = item
	}
	return result
}

func compareRows(left, right row, id string, allowed map[string]bool, updated bool, bdVersion string, nullLosses map[string]bool) error {
	l := cloneFields(left.fields)
	r := cloneFields(right.fields)
	if updated {
		normalizeArrays(l)
		normalizeArrays(r)
		if len(allowed) == 0 {
			leftAt, leftHas := l["updated_at"]
			rightAt, rightHas := r["updated_at"]
			if leftHas != rightHas || (leftHas && !jsonEqual(leftAt, rightAt)) {
				return errors.New("unplanned updated_at change")
			}
		}
		delete(l, "updated_at")
		delete(r, "updated_at")
		for field := range allowed {
			delete(l, field)
			delete(r, field)
		}
	} else {
		normalizeImported(l)
		normalizeImported(r)
	}
	if id == "dap-jo9.5.5" && bdVersion == "1.3.1" && !updated {
		if _, sourceHas := left.fields["closed_at"]; !sourceHas {
			status, statusErr := stringField(left, "status")
			if statusErr != nil || status != "closed" {
				return errors.New("pinned closed_at normalization applies only to the closed .5.5 issue")
			}
			want, err := derivedClosedAt(left)
			if err != nil {
				return err
			}
			got, err := stringField(right, "closed_at")
			if err != nil || got != want {
				return errors.New("Beads 1.3.1 closed_at normalization does not match the pinned deterministic formula")
			}
			delete(l, "closed_at")
			delete(r, "closed_at")
		}
	}
	for field := range nullLosses {
		leftValue, leftPresent := left.fields[field]
		_, rightPresent := right.fields[field]
		if !leftPresent || string(bytes.TrimSpace(leftValue)) != "null" || rightPresent {
			return fmt.Errorf("approved explicit-null loss does not match issue %s field %s", id, field)
		}
		delete(l, field)
		delete(r, field)
	}
	if !jsonEqualMaps(l, r) {
		return fmt.Errorf("unplanned semantic delta in fields %v", differingFields(l, r))
	}
	return nil
}

func authorizeNullLoss(source, baseline, updated []byte, bdVersion, bdBinarySHA256 string) (map[string]map[string]bool, []RepresentationLoss, error) {
	sourceSHA256 := digest(source)
	if !approvedNullLossSources[sourceSHA256] || bdVersion != approvedNullLossVersion || bdBinarySHA256 != approvedNullLossBDSHA256 {
		return nil, nil, nil
	}
	c, err := parseExport("source", source)
	if err != nil {
		return nil, nil, err
	}
	b0, err := parseExport("baseline", baseline)
	if err != nil {
		return nil, nil, err
	}
	b1, err := parseExport("updated", updated)
	if err != nil {
		return nil, nil, err
	}
	allowed := make(map[string]map[string]bool, len(approvedNullLossPairs))
	losses := make([]RepresentationLoss, 0, len(approvedNullLossPairs))
	for _, pair := range approvedNullLossPairs {
		canonical, ok := c[pair.issueID]
		if !ok {
			return nil, nil, fmt.Errorf("approved null-loss issue %s is missing from canonical source", pair.issueID)
		}
		value, ok := canonical.fields[pair.field]
		if !ok || string(bytes.TrimSpace(value)) != "null" {
			return nil, nil, fmt.Errorf("approved null-loss source %s.%s is not explicit null", pair.issueID, pair.field)
		}
		if _, ok := b0[pair.issueID]; !ok {
			return nil, nil, fmt.Errorf("approved null-loss issue %s is missing from baseline", pair.issueID)
		}
		if _, ok := b1[pair.issueID]; !ok {
			return nil, nil, fmt.Errorf("approved null-loss issue %s is missing from updated export", pair.issueID)
		}
		if _, present := b0[pair.issueID].fields[pair.field]; present {
			return nil, nil, fmt.Errorf("approved null-loss baseline %s.%s must be absent", pair.issueID, pair.field)
		}
		if _, present := b1[pair.issueID].fields[pair.field]; present {
			return nil, nil, fmt.Errorf("approved null-loss updated export %s.%s must remain absent", pair.issueID, pair.field)
		}
		if allowed[pair.issueID] == nil {
			allowed[pair.issueID] = make(map[string]bool)
		}
		allowed[pair.issueID][pair.field] = true
		losses = append(losses, RepresentationLoss{
			IssueID: pair.issueID, Field: pair.field, CanonicalValue: "null",
			BaselineAbsent: true, UpdatedAbsent: true, CandidateValue: "null",
			CanonicalSourceSHA: sourceSHA256,
		})
	}
	return allowed, losses, nil
}

func lossMap(losses []RepresentationLoss) map[string]map[string]bool {
	allowed := make(map[string]map[string]bool, len(losses))
	for _, loss := range losses {
		if allowed[loss.IssueID] == nil {
			allowed[loss.IssueID] = make(map[string]bool)
		}
		allowed[loss.IssueID][loss.Field] = true
	}
	return allowed
}

func candidateNullSpans(candidate []byte, losses []RepresentationLoss) ([]NullSpan, error) {
	if len(losses) == 0 {
		return nil, nil
	}
	wanted := lossMap(losses)
	spans := make([]NullSpan, 0, len(losses))
	for lineIndex, line := range splitLines(candidate) {
		id, err := rowID(line)
		if err != nil {
			return nil, fmt.Errorf("candidate line %d: %w", lineIndex+1, err)
		}
		fields := wanted[id]
		if len(fields) == 0 {
			continue
		}
		values, _, err := topLevelSpans(line)
		if err != nil {
			return nil, fmt.Errorf("candidate line %d: %w", lineIndex+1, err)
		}
		for field := range fields {
			span, ok := values[field]
			if !ok || string(bytes.TrimSpace(line[span[0]:span[1]])) != "null" {
				return nil, fmt.Errorf("derived candidate %s.%s is not explicit null", id, field)
			}
			spans = append(spans, NullSpan{IssueID: id, Field: field, Line: lineIndex + 1, Start: span[0], End: span[1], Raw: string(line[span[0]:span[1]])})
		}
	}
	if len(spans) != len(losses) {
		return nil, errors.New("derived candidate is missing an approved explicit-null field")
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].IssueID == spans[j].IssueID {
			return spans[i].Field < spans[j].Field
		}
		return spans[i].IssueID < spans[j].IssueID
	})
	return spans, nil
}

func differingFields(a, b map[string]json.RawMessage) []string {
	all := make(map[string]bool)
	for key := range a {
		all[key] = true
	}
	for key := range b {
		all[key] = true
	}
	fields := make([]string, 0)
	for key := range all {
		av, aok := a[key]
		bv, bok := b[key]
		if !aok || !bok || !jsonEqual(av, bv) {
			fields = append(fields, key)
		}
	}
	sort.Strings(fields)
	return fields
}

func derivedClosedAt(r row) (string, error) {
	created, err := stringField(r, "created_at")
	if err != nil {
		return "", err
	}
	updated, err := stringField(r, "updated_at")
	if err != nil {
		return "", err
	}
	c, err := parseTimestamp(created)
	if err != nil {
		return "", err
	}
	u, err := parseTimestamp(updated)
	if err != nil {
		return "", err
	}
	if c.After(u) {
		u = c
	}
	return u.Add(time.Second).UTC().Format(time.RFC3339), nil
}

func normalizeImported(fields map[string]json.RawMessage) {
	defaults := map[string]json.RawMessage{"_type": json.RawMessage(`"issue"`), "comment_count": json.RawMessage(`0`)}
	for key, value := range defaults {
		if _, ok := fields[key]; !ok {
			fields[key] = value
		}
	}
	for _, key := range []string{"owner", "notes", "close_reason"} {
		if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte(`""`)) {
			delete(fields, key)
		}
	}
	normalizeArrays(fields)
}

func normalizeArrays(fields map[string]json.RawMessage) {
	for _, key := range []string{"labels", "dependencies"} {
		if raw, ok := fields[key]; ok {
			trimmed := bytes.TrimSpace(raw)
			if len(trimmed) == 0 || trimmed[0] != '[' {
				continue
			}
			var values []json.RawMessage
			if json.Unmarshal(raw, &values) == nil {
				for i := range values {
					if canonical, err := canonicalJSON(values[i]); err == nil {
						values[i] = canonical
					}
				}
				sort.Slice(values, func(i, j int) bool { return string(values[i]) < string(values[j]) })
				encoded, _ := json.Marshal(values)
				fields[key] = encoded
				if len(values) == 0 {
					delete(fields, key)
				}
			}
		}
	}
}

func cloneFields(src map[string]json.RawMessage) map[string]json.RawMessage {
	dst := make(map[string]json.RawMessage, len(src))
	for k, v := range src {
		dst[k] = append(json.RawMessage(nil), v...)
	}
	return dst
}

func jsonEqualMaps(a, b map[string]json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for key, av := range a {
		bv, ok := b[key]
		if !ok || !jsonEqual(av, bv) {
			return false
		}
	}
	return true
}

func jsonEqual(a, b json.RawMessage) bool {
	ca, errA := canonicalJSON(a)
	cb, errB := canonicalJSON(b)
	return errA == nil && errB == nil && bytes.Equal(ca, cb)
}

func canonicalJSON(raw json.RawMessage) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func stringField(r row, name string) (string, error) { return stringFieldMap(r.fields, name) }
func stringFieldMap(fields map[string]json.RawMessage, name string) (string, error) {
	raw, ok := fields[name]
	if !ok {
		return "", fmt.Errorf("missing field %q", name)
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return "", fmt.Errorf("field %q must be a string", name)
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return "", fmt.Errorf("field %q must be a string", name)
	}
	return value, nil
}

func rowID(line []byte) (string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(line, &fields); err != nil {
		return "", err
	}
	return stringFieldMap(fields, "id")
}

func spliceFields(line []byte, fields map[string]string, timestampField, timestamp string) ([]byte, error) {
	values := make(map[string]json.RawMessage, len(fields)+1)
	for key, value := range fields {
		encoded, _ := json.Marshal(value)
		values[key] = encoded
	}
	encoded, _ := json.Marshal(timestamp)
	values[timestampField] = encoded
	return spliceTopLevel(line, values)
}

func spliceTopLevel(line []byte, replacements map[string]json.RawMessage) ([]byte, error) {
	spans, _, err := topLevelSpans(line)
	if err != nil {
		return nil, err
	}
	result := append([]byte(nil), line...)
	type edit struct {
		start, end int
		value      []byte
	}
	edits := make([]edit, 0, len(replacements))
	for key, value := range replacements {
		if span, ok := spans[key]; ok {
			edits = append(edits, edit{span[0], span[1], value})
			delete(replacements, key)
		}
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		result = append(result[:e.start], append(append([]byte(nil), e.value...), result[e.end:]...)...)
	}
	if len(replacements) == 0 {
		return result, nil
	}
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var addition bytes.Buffer
	if len(spans) > 0 {
		addition.WriteByte(',')
	}
	for i, key := range keys {
		encodedKey, _ := json.Marshal(key)
		if i > 0 {
			addition.WriteByte(',')
		}
		addition.Write(encodedKey)
		addition.WriteByte(':')
		addition.Write(replacements[key])
	}
	// Preserve whitespace immediately before the closing brace.
	_, insertAt, err := topLevelSpans(result)
	if err != nil {
		return nil, err
	}
	result = append(result[:insertAt], append(append([]byte(nil), addition.Bytes()...), result[insertAt:]...)...)
	return result, nil
}

func topLevelSpans(line []byte) (map[string][2]int, int, error) {
	spans := make(map[string][2]int)
	i := skipSpace(line, 0)
	if i >= len(line) || line[i] != '{' {
		return nil, 0, errors.New("row must be a JSON object")
	}
	i++
	for {
		i = skipSpace(line, i)
		if i >= len(line) {
			return nil, 0, errors.New("unterminated JSON object")
		}
		if line[i] == '}' {
			return spans, i, nil
		}
		keyStart := i
		keyEnd, err := scanString(line, i)
		if err != nil {
			return nil, 0, err
		}
		var key string
		if err := json.Unmarshal(line[keyStart:keyEnd], &key); err != nil {
			return nil, 0, err
		}
		i = skipSpace(line, keyEnd)
		if i >= len(line) || line[i] != ':' {
			return nil, 0, errors.New("missing colon after JSON key")
		}
		valueStart := skipSpace(line, i+1)
		valueEnd, err := scanValue(line, valueStart)
		if err != nil {
			return nil, 0, err
		}
		spans[key] = [2]int{valueStart, valueEnd}
		i = skipSpace(line, valueEnd)
		if i >= len(line) {
			return nil, 0, errors.New("unterminated JSON object")
		}
		if line[i] == ',' {
			i++
			continue
		}
		if line[i] == '}' {
			return spans, i, nil
		}
		return nil, 0, errors.New("expected comma or object end")
	}
}

func scanString(data []byte, start int) (int, error) {
	if start >= len(data) || data[start] != '"' {
		return 0, errors.New("expected JSON string")
	}
	for i := start + 1; i < len(data); i++ {
		if data[i] == '\\' {
			i++
			continue
		}
		if data[i] == '"' {
			return i + 1, nil
		}
	}
	return 0, errors.New("unterminated JSON string")
}

func scanValue(data []byte, start int) (int, error) {
	if start >= len(data) {
		return 0, errors.New("missing JSON value")
	}
	if data[start] == '"' {
		return scanString(data, start)
	}
	if data[start] != '{' && data[start] != '[' {
		i := start
		for i < len(data) && data[i] != ',' && data[i] != '}' && data[i] != ']' {
			i++
		}
		end := i
		for end > start && (data[end-1] == ' ' || data[end-1] == '\t' || data[end-1] == '\r' || data[end-1] == '\n') {
			end--
		}
		return end, nil
	}
	stack := []byte{data[start]}
	inString, escaped := false, false
	for i := start + 1; i < len(data); i++ {
		b := data[i]
		if inString {
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}
		if b == '"' {
			inString = true
			continue
		}
		if b == '{' || b == '[' {
			stack = append(stack, b)
			continue
		}
		if b == '}' || b == ']' {
			if len(stack) == 0 || (b == '}' && stack[len(stack)-1] != '{') || (b == ']' && stack[len(stack)-1] != '[') {
				return 0, errors.New("mismatched JSON delimiters")
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return i + 1, nil
			}
		}
	}
	return 0, errors.New("unterminated JSON value")
}

func skipSpace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
		i++
	}
	return i
}

func digest(data []byte) string   { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func digestJSON(value any) string { data, _ := json.Marshal(value); return digest(data) }

// rejectDuplicateKeys walks JSON tokens so encoding/json cannot silently accept
// a repeated key before we construct maps from the row.
func rejectDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := consumeJSONValue(dec); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func consumeJSONValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = true
			if err := consumeJSONValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	case '[':
		for dec.More() {
			if err := consumeJSONValue(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	default:
		return errors.New("unexpected JSON delimiter")
	}
}

func parseTimestamp(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
