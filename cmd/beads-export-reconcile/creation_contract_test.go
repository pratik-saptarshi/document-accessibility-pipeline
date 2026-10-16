package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
)

type creationContractFixture struct {
	CurrentSourceSHA256       string          `json:"current_source_sha256"`
	PreviousNullLossSourceSHA string          `json:"previous_null_loss_source_sha256"`
	ProposedPlan              json.RawMessage `json:"proposed_plan"`
	SnapshotCounts            struct {
		Source    int `json:"source"`
		Baseline  int `json:"baseline"`
		Updated   int `json:"updated"`
		Candidate int `json:"candidate"`
		RoundTrip int `json:"round_trip"`
	} `json:"snapshot_counts"`
	LineageDesign struct {
		Status               string   `json:"status"`
		ParentSourceSHA256   string   `json:"parent_source_sha256"`
		ApprovedSourceSHA256 string   `json:"approved_source_sha256"`
		BDVersion            string   `json:"bd_version"`
		BDBinarySHA256       string   `json:"bd_binary_sha256"`
		ApprovedPairs        []string `json:"approved_pairs"`
		Alternatives         []string `json:"alternatives"`
		Invariant            string   `json:"invariant"`
	} `json:"lineage_design"`
	SourceFreshness struct {
		OldBundleSourceSHA256   string `json:"old_bundle_source_sha256"`
		CurrentSourceSHA256     string `json:"current_source_sha256"`
		OldBundleAgainstCurrent string `json:"old_bundle_against_current_source"`
		FreshBundleAfterReview  string `json:"fresh_bundle_rebuilt_from_current_source_after_reviewed_lineage_policy"`
	} `json:"source_freshness_continuation"`
	Cases []struct {
		ID       string `json:"id"`
		Expected string `json:"expected"`
		Mutation string `json:"mutation"`
	} `json:"cases"`
}

func readCreationContractFixture(t *testing.T) creationContractFixture {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "issue-creation-contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture creationContractFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode creation contract fixture: %v", err)
	}
	return fixture
}

func TestPlannedCreationPlanIsAcceptedByStrictSchema(t *testing.T) {
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	err := strictJSON(fixture.ProposedPlan, &plan)
	if err != nil {
		t.Fatalf("strictJSON() rejected the proposed explicit creation plan: %v", err)
	}
	if len(plan.Creates) != 1 || len(plan.Updates) != 0 {
		t.Fatalf("decoded plan has %d creates and %d updates, want one create only", len(plan.Creates), len(plan.Updates))
	}
}

func TestReconcileAcceptsExactSyntheticCreate(t *testing.T) {
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	if err := strictJSON(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	parent := []byte(`{"_type":"issue","id":"dap-jo9","title":"parent","status":"open","dependencies":[]}` + "\n")
	var createFields map[string]json.RawMessage
	if err := json.Unmarshal(plan.Creates[0], &createFields); err != nil {
		t.Fatal(err)
	}
	create, err := json.Marshal(createFields)
	if err != nil {
		t.Fatal(err)
	}
	updated := append(append([]byte(nil), parent[:len(parent)-1]...), '\n')
	updated = append(updated, create...)
	updated = append(updated, '\n')
	candidate, evidence, err := beadsreconcile.Reconcile(parent, parent, updated, plan)
	if err != nil {
		t.Fatalf("Reconcile() rejected the exact synthetic create: %v", err)
	}
	if len(evidence.IssueIDs) != 0 || len(evidence.CreatedIDs) != 1 || evidence.CreatedIDs[0] != "dap-jo9.99" || len(evidence.AllIssueIDs) != 2 {
		t.Fatalf("creation provenance = updates %v, creates %v, all %v", evidence.IssueIDs, evidence.CreatedIDs, evidence.AllIssueIDs)
	}
	want := append(append([]byte(nil), parent...), create...)
	want = append(want, '\n')
	if !bytes.Equal(candidate, want) {
		t.Fatalf("candidate = %s, want exact source row plus one planned row %s", candidate, want)
	}
}

func TestSyntheticCreationSurvivesPinnedBeadsSecondStore(t *testing.T) {
	bdPath := os.Getenv("DAP_BD_BINARY")
	if bdPath == "" {
		t.Skip("set DAP_BD_BINARY to the verified Beads 1.3.1 binary for the synthetic second-store integration check")
	}
	bdBytes, _, bdAbs, err := readRegularBounded(bdPath, maxBDBinaryBytes)
	if err != nil {
		t.Fatal(err)
	}
	const pinnedDigest = "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653"
	if got := sha(bdBytes); got != pinnedDigest {
		t.Fatalf("DAP_BD_BINARY SHA-256 = %s, want pinned synthetic-test binary %s", got, pinnedDigest)
	}
	work, err := os.MkdirTemp("", "dap-create-roundtrip-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(work) })
	store, home := filepath.Join(work, "store"), filepath.Join(work, "home")
	if err := os.MkdirAll(store, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	env := cleanEnv(home)
	version, _, err := runBDVersion(bdAbs, env)
	if err != nil || version != "1.3.1" {
		t.Fatalf("pinned Beads version = %q, err = %v", version, err)
	}
	if _, _, err := runCommand(bdAbs, []string{"init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"}, store, env); err != nil {
		t.Fatal(err)
	}
	fixture := readCreationContractFixture(t)
	var plan beadsreconcile.Plan
	if err := strictJSON(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	source := []byte("{\"_type\":\"issue\",\"id\":\"dap-jo9\",\"title\":\"parent\",\"status\":\"open\",\"priority\":0,\"issue_type\":\"task\",\"created_at\":\"2026-10-16T00:00:00Z\",\"updated_at\":\"2026-10-16T00:00:00Z\",\"dependency_count\":0,\"dependent_count\":0,\"comment_count\":0}\n")
	var createFields map[string]json.RawMessage
	if err := json.Unmarshal(plan.Creates[0], &createFields); err != nil {
		t.Fatal(err)
	}
	createRow, err := json.Marshal(createFields)
	if err != nil {
		t.Fatal(err)
	}
	updated := append(append([]byte(nil), source...), createRow...)
	updated = append(updated, '\n')
	candidate, _, err := beadsreconcile.Reconcile(source, source, updated, plan)
	if err != nil {
		t.Fatalf("synthetic candidate derivation failed: %v", err)
	}
	candidatePath := filepath.Join(work, "candidate.jsonl")
	if err := os.WriteFile(candidatePath, candidate, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCommand(bdAbs, []string{"import", candidatePath}, store, env); err != nil {
		t.Fatalf("import synthetic candidate into second store: %v", err)
	}
	roundTripPath := filepath.Join(work, "round-trip.jsonl")
	if _, _, err := runCommand(bdAbs, []string{"export", "-o", roundTripPath}, store, env); err != nil {
		t.Fatalf("export synthetic second-store round trip: %v", err)
	}
	roundTrip, err := readPathBounded(roundTripPath, maxCanonicalBytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := beadsreconcile.ValidateCandidateRoundTrip(source, source, updated, candidate, roundTrip, plan, pinnedDigest); err != nil {
		t.Fatalf("second-store round-trip validation failed: %v", err)
	}
}

func TestCreateCommandBuildsSyntheticCreationEvidenceBundle(t *testing.T) {
	bdPath := os.Getenv("DAP_BD_BINARY")
	if bdPath == "" {
		t.Skip("set DAP_BD_BINARY to the verified Beads 1.3.1 binary for synthetic CLI integration")
	}
	bdBytes, _, _, err := readRegularBounded(bdPath, maxBDBinaryBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha(bdBytes); got != "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653" {
		t.Fatalf("DAP_BD_BINARY SHA-256 = %s, want the pinned Beads 1.3.1 binary", got)
	}
	fixture := readCreationContractFixture(t)
	planBytes, err := json.Marshal(fixture.ProposedPlan)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	canonicalPath := filepath.Join(dir, "canonical.jsonl")
	planPath := filepath.Join(dir, "plan.json")
	outputPath := filepath.Join(dir, "evidence.zip")
	canonical := []byte("{\"_type\":\"issue\",\"id\":\"dap-jo9\",\"title\":\"parent\",\"status\":\"open\",\"priority\":0,\"issue_type\":\"task\",\"created_at\":\"2026-10-16T00:00:00Z\",\"updated_at\":\"2026-10-16T00:00:00Z\",\"dependencies\":[],\"dependency_count\":0,\"dependent_count\":0,\"comment_count\":0}\n")
	if err := os.WriteFile(canonicalPath, canonical, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, planBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := create([]string{"--canonical", canonicalPath, "--plan", planPath, "--bd", bdPath, "--output", outputPath}); err != nil {
		t.Fatalf("create() with a synthetic one-record plan failed: %v", err)
	}
	archive, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	info, err := archive.Stat()
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(archive, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	entries := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		opened, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(opened)
		closeErr := opened.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read bundle entry %s: read=%v close=%v", file.Name, readErr, closeErr)
		}
		entries[file.Name] = data
	}
	var result manifest
	if err := json.Unmarshal(entries["manifest.json"], &result); err != nil {
		t.Fatalf("decode generated manifest: %v", err)
	}
	var plan beadsreconcile.Plan
	if err := strictJSON(planBytes, &plan); err != nil {
		t.Fatal(err)
	}
	if err := validateProvenance(result, plan); err != nil {
		t.Fatalf("create command provenance validation failed: %v", err)
	}
	if len(result.Evidence.CreatedIDs) != 1 || result.Evidence.CreatedIDs[0] != "dap-jo9.99" || len(result.Evidence.AllIssueIDs) != 2 {
		t.Fatalf("bundle creation provenance = %+v", result.Evidence)
	}
	if len(entries["candidate.jsonl"]) == 0 || len(entries["round-trip.jsonl"]) == 0 || result.Evidence.RoundTripSHA256 == "" {
		t.Fatal("bundle is missing candidate or second-store round-trip evidence")
	}
	bundleBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", canonicalPath, "--bundle", outputPath, "--bundle-sha256", sha(bundleBytes)}); err != nil {
		t.Fatalf("promote synthetic evidence bundle into its disposable canonical fixture: %v", err)
	}
	promoted, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(promoted, []byte(`"id":"dap-jo9.99"`)) {
		t.Fatalf("synthetic promotion omitted the planned child: %s", promoted)
	}
}

func TestCreateCommandReconcilesApprovedCurrentSourceInDisposableStores(t *testing.T) {
	bdPath := os.Getenv("DAP_BD_BINARY")
	if bdPath == "" {
		t.Skip("set DAP_BD_BINARY to the verified Beads 1.3.1 binary for current-source integration")
	}
	const approvedSourceSHA256 = "768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de"
	canonicalPath := filepath.Join("..", "..", ".beads", "issues.jsonl")
	source, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha(source); got != approvedSourceSHA256 {
		t.Fatalf("current canonical source SHA-256 = %s, want exact approval %s", got, approvedSourceSHA256)
	}
	fixture := readCreationContractFixture(t)
	planBytes, err := json.Marshal(fixture.ProposedPlan)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	planPath := filepath.Join(dir, "plan.json")
	outputPath := filepath.Join(dir, "evidence.zip")
	if err := os.WriteFile(planPath, planBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := create([]string{"--canonical", canonicalPath, "--plan", planPath, "--bd", bdPath, "--output", outputPath}); err != nil {
		t.Fatalf("create() rejected the exact approved source during disposable import/export: %v", err)
	}
	archive, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	info, err := archive.Stat()
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(archive, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	var result manifest
	var candidateBytes []byte
	for _, file := range reader.File {
		opened, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(opened)
		closeErr := opened.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read bundle entry %s: read=%v close=%v", file.Name, readErr, closeErr)
		}
		switch file.Name {
		case "manifest.json":
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("decode manifest: %v", err)
			}
		case "candidate.jsonl":
			candidateBytes = data
		}
	}
	if len(candidateBytes) == 0 || sha(candidateBytes) != result.Evidence.CandidateSHA256 {
		t.Fatal("current-source bundle candidate is missing or does not match its evidence digest")
	}
	wantLosses := map[string]bool{
		"dap-jo9.5.3.dependencies": true,
		"dap-jo9.1.6.10.closed_at": true,
		"dap-jo9.5.6.closed_at":    true,
		"dap-jo9.2.1.4.closed_at":  true,
		"dap-jo9.5.7.closed_at":    true,
	}
	if result.Evidence.SourceSHA256 != approvedSourceSHA256 || len(result.Evidence.RepresentationLosses) != len(wantLosses) {
		t.Fatalf("current-source evidence source/loss count = %s/%d, want %s/%d", result.Evidence.SourceSHA256, len(result.Evidence.RepresentationLosses), approvedSourceSHA256, len(wantLosses))
	}
	for _, loss := range result.Evidence.RepresentationLosses {
		pair := loss.IssueID + "." + loss.Field
		if !wantLosses[pair] {
			t.Fatalf("current-source evidence contains unapproved pair %s", pair)
		}
		delete(wantLosses, pair)
		if loss.CanonicalSourceSHA != approvedSourceSHA256 || !loss.BaselineAbsent || !loss.UpdatedAbsent || loss.CanonicalValue != "null" {
			t.Fatalf("current-source evidence has unexpected null scope: %+v", loss)
		}
	}
	if len(wantLosses) != 0 {
		t.Fatalf("current-source evidence omitted approved pairs: %v", wantLosses)
	}
	promotionPath := filepath.Join(dir, "disposable-canonical.jsonl")
	if err := os.WriteFile(promotionPath, source, 0600); err != nil {
		t.Fatal(err)
	}
	bundleBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := promote([]string{"--canonical", promotionPath, "--bundle", outputPath, "--bundle-sha256", sha(bundleBytes)}); err != nil {
		t.Fatalf("promote current-source evidence into disposable canonical fixture: %v", err)
	}
	promoted, err := os.ReadFile(promotionPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(promoted, candidateBytes) {
		t.Fatal("disposable current-source promotion does not equal the fully verified candidate")
	}
	unchangedSource, err := os.ReadFile(canonicalPath)
	if err != nil || sha(unchangedSource) != approvedSourceSHA256 {
		t.Fatalf("repository canonical source changed during disposable promotion: sha=%s err=%v", sha(unchangedSource), err)
	}
}

// This fixture pins the exact approved continuation and remaining acceptance cases.
func TestCreationContractFixtureDocumentsStaticAcceptanceCases(t *testing.T) {
	fixture := readCreationContractFixture(t)
	if fixture.CurrentSourceSHA256 != "768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de" ||
		fixture.PreviousNullLossSourceSHA != "7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890" {
		t.Fatalf("fixture source lineage = current %q, previous %q", fixture.CurrentSourceSHA256, fixture.PreviousNullLossSourceSHA)
	}
	if fixture.CurrentSourceSHA256 == fixture.PreviousNullLossSourceSHA {
		t.Fatal("fixture must distinguish the new canonical source from the source covered by the historic null-loss exception")
	}
	wantPairs := []string{
		"dap-jo9.5.3.dependencies", "dap-jo9.1.6.10.closed_at", "dap-jo9.5.6.closed_at",
		"dap-jo9.2.1.4.closed_at", "dap-jo9.5.7.closed_at",
	}
	if fixture.LineageDesign.Status != "approved-exact-continuation" ||
		fixture.LineageDesign.ParentSourceSHA256 != fixture.PreviousNullLossSourceSHA ||
		fixture.LineageDesign.ApprovedSourceSHA256 != fixture.CurrentSourceSHA256 ||
		fixture.LineageDesign.BDVersion != "1.3.1" || fixture.LineageDesign.BDBinarySHA256 != "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653" ||
		strings.Join(fixture.LineageDesign.ApprovedPairs, ",") != strings.Join(wantPairs, ",") ||
		len(fixture.LineageDesign.Alternatives) != 2 || !strings.Contains(fixture.LineageDesign.Invariant, "generic null-equals-absence rule is forbidden") {
		t.Fatalf("lineage continuation must stay bound to its exact approval and fail-closed invariant: %+v", fixture.LineageDesign)
	}
	if fixture.SourceFreshness.OldBundleSourceSHA256 != fixture.PreviousNullLossSourceSHA ||
		fixture.SourceFreshness.CurrentSourceSHA256 != fixture.CurrentSourceSHA256 ||
		fixture.SourceFreshness.OldBundleAgainstCurrent != "reject" || fixture.SourceFreshness.FreshBundleAfterReview != "approved-after-tests-and-review" {
		t.Fatalf("source-freshness continuation contract = %+v", fixture.SourceFreshness)
	}
	if fixture.SnapshotCounts.Source != 1 || fixture.SnapshotCounts.Baseline != 1 ||
		fixture.SnapshotCounts.Updated != 2 || fixture.SnapshotCounts.Candidate != 2 || fixture.SnapshotCounts.RoundTrip != 2 {
		t.Fatalf("source/baseline/updated/candidate/round-trip counts = %+v", fixture.SnapshotCounts)
	}

	var plan struct {
		Creates []struct {
			ID           string `json:"id"`
			Dependencies []struct {
				IssueID     string `json:"issue_id"`
				DependsOnID string `json:"depends_on_id"`
				Type        string `json:"type"`
			} `json:"dependencies"`
		} `json:"creates"`
	}
	if err := json.Unmarshal(fixture.ProposedPlan, &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Creates) != 1 || plan.Creates[0].ID != "dap-jo9.99" || len(plan.Creates[0].Dependencies) != 1 {
		t.Fatalf("proposed plan must contain exactly one child and edge: %+v", plan)
	}
	edge := plan.Creates[0].Dependencies[0]
	if edge.IssueID != "dap-jo9.99" || edge.DependsOnID != "dap-jo9" || edge.Type != "parent-child" {
		t.Fatalf("planned dependency edge = %+v, want child-to-existing-parent parent-child edge", edge)
	}
	var rawPlan map[string]json.RawMessage
	if err := json.Unmarshal(fixture.ProposedPlan, &rawPlan); err != nil {
		t.Fatal(err)
	}
	var records []map[string]json.RawMessage
	if err := json.Unmarshal(rawPlan["creates"], &records); err != nil || len(records) != 1 {
		t.Fatalf("planned create records = %s, err = %v; want one complete canonical issue row", rawPlan["creates"], err)
	}
	wantRecordFields := []string{
		"_type", "id", "title", "description", "acceptance_criteria", "notes", "status", "priority",
		"issue_type", "created_at", "created_by", "updated_at", "labels", "dependencies",
		"dependency_count", "dependent_count", "comment_count",
	}
	if len(records[0]) != len(wantRecordFields) {
		t.Fatalf("planned record fields = %v, want exactly %v", records[0], wantRecordFields)
	}
	for _, field := range wantRecordFields {
		if _, ok := records[0][field]; !ok {
			t.Errorf("planned canonical issue row is missing %q", field)
		}
	}
	var edgeRecords []map[string]json.RawMessage
	if err := json.Unmarshal(records[0]["dependencies"], &edgeRecords); err != nil || len(edgeRecords) != 1 {
		t.Fatalf("planned dependency records = %s, err = %v; want one complete edge", records[0]["dependencies"], err)
	}
	for _, field := range []string{"issue_id", "depends_on_id", "type", "created_at", "created_by", "metadata"} {
		if _, ok := edgeRecords[0][field]; !ok {
			t.Errorf("planned parent-child edge is missing %q", field)
		}
	}

	wantCases := []string{
		"exact_planned_record_and_parent_edge",
		"fresh_bundle_after_reviewed_lineage_continuation",
		"unplanned_record",
		"unplanned_dependency_edge",
		"unplanned_record_field",
		"duplicate_planned_id",
		"collision_with_existing_id",
		"stale_source_or_bundle",
		"unauthorized_source_lineage",
		"null_loss_without_exact_authorization",
	}
	if len(fixture.Cases) != len(wantCases) {
		t.Fatalf("acceptance case count = %d, want %d", len(fixture.Cases), len(wantCases))
	}
	for i, wantID := range wantCases {
		if fixture.Cases[i].ID != wantID {
			t.Errorf("acceptance case %d = %q, want %q", i, fixture.Cases[i].ID, wantID)
		}
		wantResult := "reject"
		if i == 0 || i == 1 {
			wantResult = "accept"
		}
		if fixture.Cases[i].Expected != wantResult {
			t.Errorf("case %s expected %q, want %q", fixture.Cases[i].ID, fixture.Cases[i].Expected, wantResult)
		}
		if strings.TrimSpace(fixture.Cases[i].Mutation) == "" {
			t.Errorf("case %s is missing its concrete acceptance/rejection mutation", fixture.Cases[i].ID)
		}
	}
	if bytes.Contains(fixture.ProposedPlan, []byte(`"updates"`)) {
		t.Fatal("creation contract fixture must remain isolated from unrelated issue updates")
	}
}
