// Command beads-export-reconcile creates and promotes reviewed, immutable
// evidence bundles for safe Beads issue-export updates.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/beadsreconcile"
	"github.com/pratik-saptarshi/document-accessibility-pipeline/internal/jsonstrict"
)

type commandRecord struct {
	Argv             []string `json:"argv"`
	WorkingDirectory string   `json:"working_directory"`
	EnvironmentSHA   string   `json:"environment_sha256"`
	ExitCode         int      `json:"exit_code"`
	StdoutSHA        string   `json:"stdout_sha256"`
	StderrSHA        string   `json:"stderr_sha256"`
}

const maxCommandOutputBytes = 1 << 20
const maxStderrExcerptBytes = 8 << 10

const (
	maxCanonicalBytes = 32 << 20
	maxPlanBytes      = 1 << 20
	maxBDBinaryBytes  = 256 << 20
	maxBundleBytes    = 64 << 20
)

type boundedCapture struct {
	digest hash.Hash
	data   bytes.Buffer
	limit  int
	total  int64
}

func newBoundedCapture(limit int) *boundedCapture {
	return &boundedCapture{digest: sha256.New(), limit: limit}
}

func (c *boundedCapture) Write(p []byte) (int, error) {
	if _, err := c.digest.Write(p); err != nil {
		return 0, err
	}
	remaining := c.limit - c.data.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		if _, err := c.data.Write(p[:remaining]); err != nil {
			return 0, err
		}
	}
	c.total += int64(len(p))
	return len(p), nil
}

func (c *boundedCapture) Bytes() []byte { return c.data.Bytes() }

func (c *boundedCapture) Digest() string { return hex.EncodeToString(c.digest.Sum(nil)) }

func (c *boundedCapture) Truncated() bool { return c.total > int64(c.limit) }

type manifest struct {
	Version       int                     `json:"version"`
	BundleID      string                  `json:"bundle_id"`
	BDVersion     string                  `json:"bd_version"`
	BDBinarySHA   string                  `json:"bd_binary_sha256"`
	MetricsStatus string                  `json:"metrics_status,omitempty"`
	Commands      []commandRecord         `json:"commands"`
	Evidence      beadsreconcile.Evidence `json:"evidence"`
	Validation    beadsreconcile.Evidence `json:"validation_evidence"`
}

func main() {
	if len(os.Args) < 2 {
		fail("usage: beads-export-reconcile create|promote [flags]")
	}
	switch os.Args[1] {
	case "create":
		if err := create(os.Args[2:]); err != nil && !errors.Is(err, flag.ErrHelp) {
			fail("create: %v", err)
		}
	case "promote":
		if err := promote(os.Args[2:]); err != nil && !errors.Is(err, flag.ErrHelp) {
			fail("promote: %v", err)
		}
	default:
		fail("unknown command %q", os.Args[1])
	}
}

func create(args []string) error {
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	canonical := flags.String("canonical", ".beads/issues.jsonl", "canonical source export")
	planPath := flags.String("plan", "", "JSON update/create evidence plan")
	bdPath := flags.String("bd", "", "pinned Beads executable")
	output := flags.String("output", "", "new evidence bundle ZIP path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *planPath == "" || *bdPath == "" || *output == "" {
		return errors.New("--plan, --bd, and --output are required")
	}
	canonicalBytes, canonicalInfo, canonicalAbs, err := readRegularBounded(*canonical, maxCanonicalBytes)
	if err != nil {
		return fmt.Errorf("read canonical export: %w", err)
	}
	planBytes, planInfo, planAbs, err := readRegularBounded(*planPath, maxPlanBytes)
	if err != nil {
		return fmt.Errorf("read update plan: %w", err)
	}
	if os.SameFile(canonicalInfo, planInfo) {
		return errors.New("plan aliases canonical export")
	}
	bdBytes, bdInfo, bdAbs, err := readRegularBounded(*bdPath, maxBDBinaryBytes)
	if err != nil {
		return fmt.Errorf("read bd binary: %w", err)
	}
	bdSHA := sha(bdBytes)
	if os.SameFile(canonicalInfo, bdInfo) || os.SameFile(canonicalInfo, planInfo) || os.SameFile(bdInfo, planInfo) {
		return errors.New("inputs alias the same file")
	}
	var plan beadsreconcile.Plan
	if err := strictJSON(planBytes, &plan); err != nil {
		return fmt.Errorf("decode update plan: %w", err)
	}
	if len(plan.Updates) == 0 && len(plan.Creates) == 0 && len(plan.Closes) == 0 {
		return errors.New("plan must contain at least one issue update, create, or close")
	}
	if err := beadsreconcile.ValidatePlan(canonicalBytes, plan); err != nil {
		return fmt.Errorf("invalid update plan: %w", err)
	}
	if len(plan.Closes) > 0 && bdSHA != "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653" {
		return errors.New("closure requires the exact authorized Beads 1.3.1 executable digest")
	}
	for _, input := range []string{canonicalAbs, planAbs, bdAbs} {
		if _, err := os.Lstat(input); err != nil {
			return err
		}
	}
	outputAbs, err := filepath.Abs(*output)
	if err != nil {
		return err
	}
	parent := filepath.Dir(outputAbs)
	if err := rejectSymlinkPath(parent); err != nil {
		return err
	}
	if _, err := os.Lstat(outputAbs); err == nil {
		return errors.New("output bundle already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	for _, input := range []string{canonicalAbs, planAbs, bdAbs} {
		if input == outputAbs {
			return errors.New("output aliases an input")
		}
	}

	work, err := os.MkdirTemp("", "dap-beads-reconcile-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	store := filepath.Join(work, "store")
	home := filepath.Join(work, "home")
	if err := os.MkdirAll(store, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0700); err != nil {
		return err
	}
	env := cleanEnv(home)
	verifyBD := func() error {
		current, info, _, err := readRegularBounded(bdAbs, maxBDBinaryBytes)
		if err != nil {
			return err
		}
		if !os.SameFile(info, bdInfo) || sha(current) != bdSHA {
			return errors.New("pinned bd executable changed during evidence capture")
		}
		return nil
	}
	if err := verifyBD(); err != nil {
		return err
	}
	commands := make([]commandRecord, 0)
	runAt := func(dir string, argv ...string) ([]byte, error) {
		if err := verifyBD(); err != nil {
			return nil, err
		}
		out, rec, err := runCommand(bdAbs, argv, dir, env)
		commands = append(commands, rec)
		if verifyErr := verifyBD(); verifyErr != nil {
			return nil, verifyErr
		}
		return out, err
	}
	run := func(argv ...string) ([]byte, error) { return runAt(store, argv...) }
	metricsStatus := ""
	if len(plan.Closes) > 0 {
		if _, err := run("metrics", "off"); err != nil {
			return fmt.Errorf("disable metrics in disposable Beads home: %w", err)
		}
		status, err := run("metrics")
		if err != nil {
			return fmt.Errorf("verify metrics in disposable Beads home: %w", err)
		}
		if !strings.Contains(strings.ToLower(string(status)), "off") {
			return errors.New("Beads metrics status did not confirm OFF")
		}
		metricsStatus = "OFF"
	}
	bdVersion, versionRecord, err := runBDVersion(bdAbs, env)
	if err != nil {
		return err
	}
	commands = append(commands, versionRecord)
	if err := verifyBD(); err != nil {
		return err
	}
	if len(plan.Closes) > 0 && bdVersion != "1.3.1" {
		return errors.New("closure requires Beads 1.3.1")
	}
	plan.BDVersion = bdVersion
	planCanonical, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	if _, err := run("init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"); err != nil {
		return err
	}
	if _, err := run("import", canonicalAbs); err != nil {
		return err
	}
	b0Path := filepath.Join(work, "baseline.jsonl")
	if _, err := run("export", "-o", b0Path); err != nil {
		return err
	}
	b0, err := readPathBounded(b0Path, maxCanonicalBytes)
	if err != nil {
		return err
	}
	if _, _, err := beadsreconcile.ReconcileWithBinary(canonicalBytes, b0, b0, beadsreconcile.Plan{BDVersion: bdVersion}, bdSHA); err != nil {
		return fmt.Errorf("source-to-baseline validation failed before updates: %w", err)
	}
	updates := append([]beadsreconcile.IssueUpdate(nil), plan.Updates...)
	sort.Slice(updates, func(i, j int) bool { return updates[i].ID < updates[j].ID })
	for _, update := range updates {
		fieldNames := make([]string, 0, len(update.Fields))
		for name := range update.Fields {
			fieldNames = append(fieldNames, name)
		}
		sort.Strings(fieldNames)
		argv := []string{"update", update.ID}
		for _, name := range fieldNames {
			if len(update.Fields[name]) > 64*1024 {
				return fmt.Errorf("plan value %s.%s exceeds 64 KiB", update.ID, name)
			}
			flagName := map[string]string{"title": "--title", "description": "--description", "acceptance_criteria": "--acceptance", "notes": "--notes"}[name]
			if flagName == "" {
				return fmt.Errorf("unsupported plan field %q", name)
			}
			if name == "description" && update.Fields[name] == "" {
				argv = append(argv, "--allow-empty-description")
			}
			argv = append(argv, flagName+"="+update.Fields[name])
		}
		if _, err := run(argv...); err != nil {
			return fmt.Errorf("apply plan update %s: %w", update.ID, err)
		}
	}
	for _, closure := range plan.Closes {
		if err := runClosure(run, closure); err != nil {
			return fmt.Errorf("apply planned close %s: %w", closure.ID, err)
		}
	}
	for _, planned := range plan.Creates {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(planned, &fields); err != nil {
			return fmt.Errorf("decode planned create row: %w", err)
		}
		rowBytes, err := json.Marshal(fields)
		if err != nil {
			return fmt.Errorf("encode planned create row: %w", err)
		}
		createPath := filepath.Join(work, "planned-create.jsonl")
		if err := os.WriteFile(createPath, append(rowBytes, '\n'), 0600); err != nil {
			return err
		}
		if _, err := run("import", createPath); err != nil {
			var id string
			_ = json.Unmarshal(fields["id"], &id)
			return fmt.Errorf("apply planned create %q: %w", id, err)
		}
	}
	b1Path := filepath.Join(work, "updated.jsonl")
	if _, err := run("export", "-o", b1Path); err != nil {
		return err
	}
	b1, err := readPathBounded(b1Path, maxCanonicalBytes)
	if err != nil {
		return err
	}
	candidate, evidence, err := beadsreconcile.ReconcileWithBinary(canonicalBytes, b0, b1, plan, bdSHA)
	if err != nil {
		return err
	}
	if evidence.SourceSHA256 != sha(canonicalBytes) {
		return errors.New("internal source digest mismatch")
	}
	if err := ensureUnchanged(canonicalAbs, evidence.SourceSHA256); err != nil {
		return fmt.Errorf("stale canonical before bundle creation: %w", err)
	}
	verifyStore := filepath.Join(work, "verification-store")
	if err := os.MkdirAll(verifyStore, 0700); err != nil {
		return err
	}
	candidatePath := filepath.Join(work, "candidate.jsonl")
	if err := os.WriteFile(candidatePath, candidate, 0600); err != nil {
		return err
	}
	if _, err := runAt(verifyStore, "init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"); err != nil {
		return err
	}
	if _, err := runAt(verifyStore, "import", candidatePath); err != nil {
		return err
	}
	roundTripPath := filepath.Join(work, "round-trip.jsonl")
	if _, err := runAt(verifyStore, "export", "-o", roundTripPath); err != nil {
		return err
	}
	roundTrip, err := readPathBounded(roundTripPath, maxCanonicalBytes)
	if err != nil {
		return err
	}
	var validation beadsreconcile.Evidence
	evidence, validation, err = beadsreconcile.ValidateCandidateRoundTrip(canonicalBytes, b0, b1, candidate, roundTrip, plan, bdSHA)
	if err != nil {
		return fmt.Errorf("second-store candidate validation failed: %w", err)
	}
	evidence.RoundTripSHA256 = sha(roundTrip)
	validation.RoundTripSHA256 = sha(roundTrip)
	if err := ensureUnchanged(canonicalAbs, evidence.SourceSHA256); err != nil {
		return fmt.Errorf("stale canonical before evidence bundle creation: %w", err)
	}
	manifestValue := manifest{Version: 2, BDVersion: bdVersion, BDBinarySHA: bdSHA, MetricsStatus: metricsStatus, Commands: commands, Evidence: evidence, Validation: validation}
	manifestValue.BundleID = bundleID(manifestValue)
	manifestBytes, err := json.MarshalIndent(manifestValue, "", "  ")
	if err != nil {
		return err
	}
	entries := map[string][]byte{"source.jsonl": canonicalBytes, "baseline.jsonl": b0, "updated.jsonl": b1, "plan.json": planCanonical, "candidate.jsonl": candidate, "round-trip.jsonl": roundTrip, "manifest.json": manifestBytes}
	if err := writeBundle(outputAbs, entries); err != nil {
		return err
	}
	bundleBytes, err := readPathBounded(outputAbs, maxBundleBytes)
	if err != nil {
		return err
	}
	fmt.Printf("created evidence bundle %s\nsource=%s candidate=%s bundle=%s\n", outputAbs, evidence.SourceSHA256, evidence.CandidateSHA256, sha(bundleBytes))
	return nil
}

func runClosure(run func(...string) ([]byte, error), closure beadsreconcile.IssueClosure) error {
	_, err := run("close", closure.ID, "--reason="+closure.Reason)
	return err
}

func promote(args []string) error {
	flags := flag.NewFlagSet("promote", flag.ContinueOnError)
	canonical := flags.String("canonical", ".beads/issues.jsonl", "canonical source export to update")
	bundlePath := flags.String("bundle", "", "reviewed evidence bundle ZIP")
	bundleSHA := flags.String("bundle-sha256", "", "bundle digest recorded during review")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *bundlePath == "" || *bundleSHA == "" {
		return errors.New("--bundle and --bundle-sha256 are required")
	}
	bundleBytes, bundleInfo, _, err := readRegularBounded(*bundlePath, maxBundleBytes)
	if err != nil {
		return err
	}
	if sha(bundleBytes) != strings.ToLower(*bundleSHA) {
		return errors.New("evidence bundle digest mismatch")
	}
	canonicalBytes, info, canonicalAbs, err := readRegularBounded(*canonical, maxCanonicalBytes)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&022 != 0 {
		return errors.New("canonical export is group or world writable")
	}
	if err := requireSingleLink(info); err != nil {
		return fmt.Errorf("canonical export: %w", err)
	}
	if os.SameFile(info, bundleInfo) {
		return errors.New("bundle aliases canonical export")
	}
	entries, err := readBundle(bundleBytes)
	if err != nil {
		return err
	}
	var m manifest
	if err := strictJSON(entries["manifest.json"], &m); err != nil {
		return err
	}
	if m.Version != 2 || m.Evidence.Version != 2 || m.Evidence.Outcome != "reconciled" || m.Validation.Outcome != "reconciled" {
		return errors.New("unsupported or incomplete evidence manifest")
	}
	if bundleID(m) != m.BundleID {
		return errors.New("evidence bundle identity mismatch")
	}
	if sha(entries["source.jsonl"]) != m.Evidence.SourceSHA256 || sha(entries["baseline.jsonl"]) != m.Evidence.BaselineSHA256 || sha(entries["updated.jsonl"]) != m.Evidence.UpdatedSHA256 || sha(entries["candidate.jsonl"]) != m.Evidence.CandidateSHA256 || sha(entries["plan.json"]) != m.Evidence.PlanSHA256 || sha(entries["round-trip.jsonl"]) != m.Evidence.RoundTripSHA256 {
		return errors.New("evidence snapshot or candidate digest mismatch")
	}
	if len(m.Commands) < 6 || m.BDVersion != "1.3.1" || len(m.BDBinarySHA) != 64 {
		return errors.New("command provenance is incomplete")
	}
	for _, command := range m.Commands {
		if command.ExitCode != 0 {
			return errors.New("command provenance contains a failed operation")
		}
	}
	var plan beadsreconcile.Plan
	if err := strictJSON(entries["plan.json"], &plan); err != nil {
		return err
	}
	if plan.BDVersion != m.BDVersion {
		return errors.New("plan and command provenance Beads versions differ")
	}
	if err := validateProvenance(m, plan); err != nil {
		return err
	}
	evidence, validation, err := beadsreconcile.ValidateCandidateRoundTrip(entries["source.jsonl"], entries["baseline.jsonl"], entries["updated.jsonl"], entries["candidate.jsonl"], entries["round-trip.jsonl"], plan, m.BDBinarySHA)
	if err != nil {
		return fmt.Errorf("revalidate candidate and second-store evidence: %w", err)
	}
	evidence.RoundTripSHA256 = sha(entries["round-trip.jsonl"])
	validation.RoundTripSHA256 = sha(entries["round-trip.jsonl"])
	evidenceBytes, _ := json.Marshal(evidence)
	manifestEvidenceBytes, _ := json.Marshal(m.Evidence)
	validationBytes, _ := json.Marshal(validation)
	manifestValidationBytes, _ := json.Marshal(m.Validation)
	if !bytes.Equal(evidenceBytes, manifestEvidenceBytes) || !bytes.Equal(validationBytes, manifestValidationBytes) {
		return errors.New("candidate or evidence does not match the three snapshots and plan")
	}
	if sha(canonicalBytes) != m.Evidence.SourceSHA256 {
		return errors.New("canonical export is stale; create a new evidence bundle")
	}
	if err := atomicReplace(canonicalAbs, info, entries["candidate.jsonl"], m.Evidence.SourceSHA256); err != nil {
		return err
	}
	fmt.Printf("promoted candidate %s to %s\n", m.Evidence.CandidateSHA256, canonicalAbs)
	return nil
}

func validateProvenance(m manifest, plan beadsreconcile.Plan) error {
	if m.BDVersion != "1.3.1" || len(m.BDBinarySHA) != 64 {
		return errors.New("unsupported Beads provenance")
	}
	if _, err := hex.DecodeString(m.BDBinarySHA); err != nil {
		return errors.New("invalid Beads executable digest")
	}
	updates := append([]beadsreconcile.IssueUpdate(nil), plan.Updates...)
	sort.Slice(updates, func(i, j int) bool { return updates[i].ID < updates[j].ID })
	telemetryPrefix := 0
	if len(plan.Closes) > 0 {
		if m.MetricsStatus != "OFF" || m.BDBinarySHA != "a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653" {
			return errors.New("closure provenance lacks the authorized binary pin or metrics-off confirmation")
		}
		telemetryPrefix = 2
	}
	wantCount := telemetryPrefix + len(updates) + len(plan.Creates) + len(plan.Closes) + 8
	if len(m.Commands) != wantCount {
		return errors.New("command provenance count does not match the plan")
	}
	binary := ""
	check := func(index int, args ...string) error {
		record := m.Commands[index]
		if record.ExitCode != 0 || len(record.Argv) != len(args)+1 {
			return errors.New("invalid command provenance record")
		}
		if index == 0 {
			binary = record.Argv[0]
		}
		if binary == "" || record.Argv[0] != binary {
			return errors.New("command provenance mixes executables")
		}
		for i, arg := range args {
			if record.Argv[i+1] != arg {
				return fmt.Errorf("command provenance mismatch at operation %d", index+1)
			}
		}
		if record.WorkingDirectory == "" || len(record.EnvironmentSHA) != 64 || len(record.StdoutSHA) != 64 || len(record.StderrSHA) != 64 {
			return errors.New("command output digests are missing")
		}
		if _, err := hex.DecodeString(record.EnvironmentSHA); err != nil {
			return err
		}
		if _, err := hex.DecodeString(record.StdoutSHA); err != nil {
			return err
		}
		if _, err := hex.DecodeString(record.StderrSHA); err != nil {
			return err
		}
		return nil
	}
	versionIndex := telemetryPrefix
	if telemetryPrefix == 2 {
		if err := check(0, "metrics", "off"); err != nil {
			return err
		}
		if err := check(1, "metrics"); err != nil {
			return err
		}
	}
	if err := check(versionIndex, "--version"); err != nil {
		return err
	}
	initIndex := versionIndex + 1
	if err := check(initIndex, "init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"); err != nil {
		return err
	}
	importIndex := initIndex + 1
	if len(m.Commands[importIndex].Argv) != 3 {
		return errors.New("invalid import command provenance")
	}
	if err := check(importIndex, "import", m.Commands[importIndex].Argv[2]); err != nil {
		return err
	}
	if len(m.Commands[importIndex].Argv[2]) == 0 {
		return errors.New("import command has no source path")
	}
	baselineIndex := importIndex + 1
	if len(m.Commands[baselineIndex].Argv) != 4 || m.Commands[baselineIndex].Argv[3] == "" {
		return errors.New("invalid baseline export provenance")
	}
	if err := check(baselineIndex, "export", "-o", m.Commands[baselineIndex].Argv[3]); err != nil {
		return err
	}
	operationStart := baselineIndex + 1
	for i, update := range updates {
		args := []string{"update", update.ID}
		fields := make([]string, 0, len(update.Fields))
		for field := range update.Fields {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		for _, field := range fields {
			flagName := map[string]string{"title": "--title", "description": "--description", "acceptance_criteria": "--acceptance", "notes": "--notes"}[field]
			if flagName == "" {
				return fmt.Errorf("unsupported provenance field %q", field)
			}
			if field == "description" && update.Fields[field] == "" {
				args = append(args, "--allow-empty-description")
			}
			args = append(args, flagName+"="+update.Fields[field])
		}
		if err := check(operationStart+i, args...); err != nil {
			return err
		}
	}
	closeStart := operationStart + len(updates) + len(plan.Creates)
	for i, closure := range plan.Closes {
		if err := check(closeStart+i, "close", closure.ID, "--reason="+closure.Reason); err != nil {
			return err
		}
	}
	for i, planned := range plan.Creates {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(planned, &fields); err != nil {
			return fmt.Errorf("invalid planned create provenance: %w", err)
		}
		createIndex := operationStart + len(updates) + i
		if len(m.Commands[createIndex].Argv) != 3 || filepath.Base(m.Commands[createIndex].Argv[2]) != "planned-create.jsonl" {
			return errors.New("invalid planned create import provenance")
		}
		if err := check(createIndex, "import", m.Commands[createIndex].Argv[2]); err != nil {
			return err
		}
	}
	b1Index := operationStart + len(updates) + len(plan.Creates) + len(plan.Closes)
	if len(m.Commands[b1Index].Argv) != 4 || m.Commands[b1Index].Argv[3] == "" {
		return errors.New("invalid updated export provenance")
	}
	if err := check(b1Index, "export", "-o", m.Commands[b1Index].Argv[3]); err != nil {
		return err
	}
	verifyInit := b1Index + 1
	if err := check(verifyInit, "init", "--non-interactive", "--skip-agents", "--skip-hooks", "--prefix", "dap"); err != nil {
		return err
	}
	verifyImport := verifyInit + 1
	if len(m.Commands[verifyImport].Argv) != 3 || m.Commands[verifyImport].Argv[2] == "" {
		return errors.New("invalid verification import provenance")
	}
	if err := check(verifyImport, "import", m.Commands[verifyImport].Argv[2]); err != nil {
		return err
	}
	verifyExport := verifyImport + 1
	if len(m.Commands[verifyExport].Argv) != 4 || m.Commands[verifyExport].Argv[3] == "" {
		return errors.New("invalid verification export provenance")
	}
	if err := check(verifyExport, "export", "-o", m.Commands[verifyExport].Argv[3]); err != nil {
		return err
	}
	return nil
}

func runBDVersion(path string, env []string) (string, commandRecord, error) {
	out, rec, err := runCommand(path, []string{"--version"}, "", env)
	if err != nil {
		return "", rec, err
	}
	match := regexp.MustCompile(`(?m)^bd version ([0-9]+\.[0-9]+\.[0-9]+) `).FindSubmatch(out)
	if len(match) != 2 {
		return "", rec, errors.New("could not parse bd --version output")
	}
	version := string(match[1])
	if version != "1.3.1" {
		return "", rec, fmt.Errorf("unsupported Beads version %s; expected 1.3.1", version)
	}
	return version, rec, nil
}

func runCommand(binary string, argv []string, dir string, env []string) ([]byte, commandRecord, error) {
	return runCommandWithTimeout(binary, argv, dir, env, 3*time.Minute)
}

func runCommandWithTimeout(binary string, argv []string, dir string, env []string, timeout time.Duration) ([]byte, commandRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, argv...)
	cmd.Dir = dir
	cmd.Env = env
	workingDirectory := dir
	if workingDirectory == "" {
		workingDirectory, _ = os.Getwd()
	}
	stdout, stderr := newBoundedCapture(maxCommandOutputBytes), newBoundedCapture(maxCommandOutputBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := runCommandProcess(ctx, cmd)
	code := 0
	if err != nil {
		code = 1
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		}
	}
	rec := commandRecord{Argv: append([]string{binary}, argv...), WorkingDirectory: workingDirectory, EnvironmentSHA: sha([]byte(strings.Join(env, "\x00"))), ExitCode: code, StdoutSHA: stdout.Digest(), StderrSHA: stderr.Digest()}
	if err != nil {
		return nil, rec, fmt.Errorf("command %q failed (exit %d): %s", argv[0], code, stderrExcerpt(stderr))
	}
	if stdout.Truncated() || stderr.Truncated() {
		streams := make([]string, 0, 2)
		if stdout.Truncated() {
			streams = append(streams, "stdout")
		}
		if stderr.Truncated() {
			streams = append(streams, "stderr")
		}
		return nil, rec, fmt.Errorf("command %q exceeded the %d-byte capture limit for %s", argv[0], maxCommandOutputBytes, strings.Join(streams, " and "))
	}
	return stdout.Bytes(), rec, nil
}

func stderrExcerpt(capture *boundedCapture) string {
	data := capture.Bytes()
	truncated := capture.Truncated()
	if len(data) > maxStderrExcerptBytes {
		data = data[:maxStderrExcerptBytes]
		truncated = true
	}
	excerpt := strings.TrimSpace(string(data))
	if truncated {
		return excerpt + " [stderr excerpt truncated; full stream digest recorded]"
	}
	return excerpt
}

func cleanEnv(home string) []string {
	path := os.Getenv("PATH")
	tmp := os.TempDir()
	return []string{"PATH=" + path, "TMPDIR=" + tmp, "HOME=" + home, "GIT_CONFIG_GLOBAL=/dev/null", "BD_NON_INTERACTIVE=1", "BD_NO_AUTO_IMPORT=1"}
}

func writeBundle(path string, entries map[string][]byte) error {
	return writeBundleWithStat(path, entries, os.Stat)
}

func writeBundleWithStat(path string, entries map[string][]byte, stat func(string) (os.FileInfo, error)) error {
	parent := filepath.Dir(path)
	tmp, err := os.CreateTemp(parent, ".beads-reconcile-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	w := zip.NewWriter(tmp)
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		header := &zip.FileHeader{Name: key, Method: zip.Deflate}
		header.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		file, err := w.CreateHeader(header)
		if err != nil {
			_ = w.Close()
			_ = tmp.Close()
			return err
		}
		if _, err := file.Write(entries[key]); err != nil {
			_ = w.Close()
			_ = tmp.Close()
			return err
		}
	}
	if err := w.Close(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	info, err := stat(tmpPath)
	if err != nil {
		return err
	}
	if err := validateGeneratedBundleSize(info.Size()); err != nil {
		return err
	}
	if err := os.Link(tmpPath, path); err != nil {
		return fmt.Errorf("publish bundle without overwrite: %w", err)
	}
	if dir, err := os.Open(parent); err == nil {
		syncErr := dir.Sync()
		_ = dir.Close()
		if syncErr != nil {
			return syncErr
		}
	}
	return nil
}

func validateGeneratedBundleSize(size int64) error {
	if size > maxBundleBytes {
		return fmt.Errorf("evidence bundle exceeds %d MiB", maxBundleBytes>>20)
	}
	return nil
}

func readBundle(data []byte) (map[string][]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	want := map[string]bool{"source.jsonl": true, "baseline.jsonl": true, "updated.jsonl": true, "plan.json": true, "candidate.jsonl": true, "round-trip.jsonl": true, "manifest.json": true}
	entries := map[string][]byte{}
	for _, file := range zr.File {
		if !want[file.Name] {
			return nil, fmt.Errorf("unexpected bundle entry %q", file.Name)
		}
		if _, ok := entries[file.Name]; ok {
			return nil, fmt.Errorf("duplicate bundle entry %q", file.Name)
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(r, 32<<20+1))
		closeErr := r.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(data) > 32<<20 {
			return nil, errors.New("bundle entry exceeds 32 MiB")
		}
		entries[file.Name] = data
	}
	if len(entries) != len(want) {
		return nil, errors.New("bundle is missing required entries")
	}
	return entries, nil
}

// atomicReplaceOps isolates failure-prone filesystem operations so tests can
// inject deterministic errors without changing the replacement sequence.
type atomicReplaceOps struct {
	writeTemp func(*os.File, []byte) (int, error)
	syncTemp  func(*os.File) error
	closeTemp func(*os.File) error
	rename    func(string, string) error
	syncDir   func(*os.File) error
}

func defaultAtomicReplaceOps() atomicReplaceOps {
	return atomicReplaceOps{
		writeTemp: func(file *os.File, data []byte) (int, error) { return file.Write(data) },
		syncTemp:  func(file *os.File) error { return file.Sync() },
		closeTemp: func(file *os.File) error { return file.Close() },
		rename:    os.Rename,
		syncDir:   func(dir *os.File) error { return dir.Sync() },
	}
}

func atomicReplace(path string, original os.FileInfo, candidate []byte, expectedSource string) error {
	return atomicReplaceWithOps(path, original, candidate, expectedSource, defaultAtomicReplaceOps())
}

func atomicReplaceWithOps(path string, original os.FileInfo, candidate []byte, expectedSource string, ops atomicReplaceOps) error {
	// This sibling lock serializes publishers that follow this protocol. It is
	// not a filesystem compare-and-swap: callers must ensure every writer of the
	// canonical export honors the same lock or is otherwise quiescent.
	lock := path + ".beads-reconcile.lock"
	lockFile, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("canonical promotion lock unavailable: %w", err)
	}
	_ = lockFile.Close()
	defer os.Remove(lock)
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !current.Mode().IsRegular() || current.Mode()&os.ModeSymlink != 0 {
		return errors.New("canonical path is not a regular non-symlink file")
	}
	currentBytes, err := readPathBounded(path, maxCanonicalBytes)
	if err != nil {
		return err
	}
	if sha(currentBytes) != expectedSource {
		return errors.New("canonical export changed during promotion")
	}
	if !os.SameFile(original, current) {
		return errors.New("canonical file identity changed or has hard-link aliases")
	}
	if err := requireSingleLink(current); err != nil {
		return fmt.Errorf("canonical file: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".beads-candidate-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := ops.writeTemp(tmp, candidate); err != nil {
		_ = ops.closeTemp(tmp)
		return err
	}
	if err := ops.syncTemp(tmp); err != nil {
		_ = ops.closeTemp(tmp)
		return err
	}
	if err := tmp.Chmod(current.Mode().Perm()); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := ops.closeTemp(tmp); err != nil {
		return err
	}
	latest, err := readPathBounded(path, maxCanonicalBytes)
	if err != nil {
		return err
	}
	if sha(latest) != expectedSource {
		return errors.New("canonical export became stale immediately before atomic replacement")
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	if err := ops.rename(tmpName, path); err != nil {
		return err
	}
	if err := ops.syncDir(dir); err != nil {
		return fmt.Errorf("canonical candidate was replaced but directory sync failed: %w", err)
	}
	return nil
}

func readRegular(path string) ([]byte, os.FileInfo, string, error) {
	return readRegularBounded(path, maxCanonicalBytes)
}

func readRegularBounded(path string, maxBytes int64) ([]byte, os.FileInfo, string, error) {
	if maxBytes < 0 {
		return nil, nil, "", errors.New("input size limit must not be negative")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, "", err
	}
	if err := rejectSymlinkPath(filepath.Dir(abs)); err != nil {
		return nil, nil, "", err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, nil, "", err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, "", errors.New("input must be a regular non-symlink file")
	}
	if err := requireSingleLink(info); err != nil {
		return nil, nil, "", fmt.Errorf("input file: %w", err)
	}
	if info.Size() > maxBytes {
		return nil, nil, "", fmt.Errorf("input exceeds %d-byte limit", maxBytes)
	}
	file, err := os.Open(abs)
	if err != nil {
		return nil, nil, "", err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, nil, "", err
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return nil, nil, "", errors.New("input file identity changed while opening")
	}
	if err := requireSingleLink(openedInfo); err != nil {
		return nil, nil, "", fmt.Errorf("input file: %w", err)
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, nil, "", err
	}
	if int64(len(data)) > maxBytes {
		return nil, nil, "", fmt.Errorf("input exceeds %d-byte limit", maxBytes)
	}
	return data, info, abs, nil
}

func readPathBounded(path string, maxBytes int64) ([]byte, error) {
	data, _, _, err := readRegularBounded(path, maxBytes)
	return data, err
}

func requireSingleLink(info os.FileInfo) error {
	count, supported := nlink(info)
	if !supported {
		return errors.New("cannot verify hard-link count on this platform")
	}
	if count != 1 {
		return errors.New("file has hard-link aliases")
	}
	return nil
}

func rejectSymlinkPath(path string) error {
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(path), current), current) {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("unsafe path component %s", current)
		}
	}
	return nil
}

func strictJSON(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	var raw any
	rawDecoder := json.NewDecoder(bytes.NewReader(data))
	if err := rawDecoder.Decode(&raw); err != nil {
		return err
	}
	if err := rejectNoncanonicalSchemaKeys(raw, reflect.TypeOf(target), ""); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("trailing JSON value")
	}
	return nil
}

func rejectNoncanonicalSchemaKeys(value any, target reflect.Type, path string) error {
	for target != nil && target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	if target == nil {
		return nil
	}
	switch target.Kind() {
	case reflect.String:
		if value == nil {
			return fmt.Errorf("JSON null is not allowed for string at %s", jsonPath(path))
		}
		if _, ok := value.(string); !ok {
			return fmt.Errorf("value at %s must be a JSON string", jsonPath(path))
		}
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil // The typed decoder reports a non-object value.
		}
		fields := make(map[string]reflect.Type, target.NumField())
		for i := 0; i < target.NumField(); i++ {
			field := target.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			fields[name] = field.Type
		}
		for key, child := range object {
			fieldType, exact := fields[key]
			if !exact {
				for canonical := range fields {
					if strings.EqualFold(key, canonical) {
						return fmt.Errorf("noncanonical JSON key %q at %s; use %q", key, jsonPath(path), canonical)
					}
				}
				continue // DisallowUnknownFields reports unrelated unknown keys.
			}
			if err := rejectNoncanonicalSchemaKeys(child, fieldType, path+"."+key); err != nil {
				return err
			}
		}
	case reflect.Map:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		for key, child := range object {
			if err := rejectNoncanonicalSchemaKeys(child, target.Elem(), path+"."+key); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		array, ok := value.([]any)
		if !ok {
			return nil
		}
		for i, child := range array {
			if err := rejectNoncanonicalSchemaKeys(child, target.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func jsonPath(path string) string {
	if path == "" {
		return "$"
	}
	return path
}

func rejectDuplicateKeys(data []byte) error {
	if err := jsonstrict.RejectDuplicateKeys(bytes.NewReader(data)); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	var value any
	if err := dec.Decode(&value); err != nil {
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

func ensureUnchanged(path, digest string) error {
	data, _, _, err := readRegularBounded(path, maxCanonicalBytes)
	if err != nil {
		return err
	}
	if sha(data) != digest {
		return errors.New("source export bytes changed")
	}
	return nil
}
func sha(data []byte) string            { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func bundleID(m manifest) string        { m.BundleID = ""; data, _ := json.Marshal(m); return sha(data) }
func fail(format string, values ...any) { fmt.Fprintf(os.Stderr, format+"\n", values...); os.Exit(2) }
