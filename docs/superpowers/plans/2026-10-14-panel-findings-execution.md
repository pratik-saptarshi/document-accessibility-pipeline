# Adversarial Panel Findings Execution Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the four Beads items filed from the 2026-10-14 adversarial panel review, in priority order, with tests gating each change.

**Architecture:** Four independent workstreams, each a small vertical slice: workflow YAML change plus policy test (Tasks 1–2), parser hardening plus fuzz tests (Task 3), docs plus process-cleanup unification (Task 4). No task depends on another; order is by panel priority (P1s first).

**Tech Stack:** Go (root module plus `ci/pdfua2/adapter` and `cmd/cloud-tools-bootstrap` nested modules), GitHub Actions workflow YAML, `internal/devsecops` policy tests.

**Spec:** `docs/reviews/panel-full-repo-2026-10-14.md` and Beads items `dap-jo9.1.6.10`, `dap-jo9.5.6`, `dap-jo9.2.1.4`, `dap-jo9.5.7`.

## Global Constraints

- Conventional Commit subjects enforced by `scripts/source-publisher.sh`; publish only via that script.
- `go test -race` needs cgo and a C compiler; this does not make cgo a product runtime requirement.
- Root `./...` commands do not traverse the two nested Go modules; run their checks separately per AGENTS.md.
- Never bypass, disable, or weaken a failing CI check to get a change through.
- Beads edits go through the JSONL procedure in AGENTS.md (Go-compatible JSON, recompute counts); close items only with evidence each AC passed.
- Do not run the validator provisioner or real CLI locally; adapter tests run with `PDFUA_REAL_CLI_JAR=` unset.

## Review Focus

- A workflow YAML edit that is valid YAML but semantically wrong (e.g., retention set on the wrong artifact): the policy test must assert on the parsed `pdfua2-reports` artifact's `retention-days`, not just grep the file.
- A policy-test map entry whose step name drifts from the workflow's `name:` field: the test compares exact names, so any rename must update both.
- A fuzz seed corpus that never exercises the `isCompliant` neither-true-nor-false branch: add that exact input as a seed.
- A "strict" JSON decoder that rejects duplicate keys but still folds case variants: test `"PROVENANCE"` vs `"provenance"` explicitly.
- A process-cleanup unification that changes signal ordering on one platform but not the other: keep both `process_linux.go` files in the same task.

---

### Task 1: Durable validator evidence retention (`dap-jo9.1.6.10`, P1)

**Files:**
- Modify: `.github/workflows/verapdf-integration.yml` (artifact `pdfua2-reports`, currently `retention-days: 1`)
- Modify: `internal/devsecops/policy_test.go` (add retention-floor assertion)
- Test: `internal/devsecops/policy_test.go`

**Interfaces:**
- Consumes: nothing from other tasks.
- Produces: `retention-days: 90` on the `pdfua2-reports` artifact; policy test function asserting the floor.

- [ ] **Step 1: Write the failing policy test**

Add to `internal/devsecops/policy_test.go` a test `TestValidatorEvidenceRetentionFloor` that loads `.github/workflows/verapdf-integration.yml`, finds the `pdfua2-reports` artifact upload step, and asserts `retention-days >= 90`. Follow the existing workflow-parsing pattern in that file.

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -race ./internal/devsecops/ -run TestValidatorEvidenceRetentionFloor -v`
Expected: FAIL (current value is 1)

- [ ] **Step 3: Set `retention-days: 90` in `.github/workflows/verapdf-integration.yml`**

Change only the `pdfua2-reports` artifact's `retention-days` from `1` to `90` (GitHub maximum). Do not touch other artifacts or jobs.

- [ ] **Step 4: Run the policy test to verify it passes**

Run: `go test -race ./internal/devsecops/ -v`
Expected: PASS (whole package)

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/verapdf-integration.yml internal/devsecops/policy_test.go
git commit -m "fix(ci): keep validator evidence for 90 days with policy floor (dap-jo9.1.6.10)"
```

### Task 2: Nested-module CI coverage (`dap-jo9.5.6`, P1)

**Files:**
- Modify: `.github/workflows/go-checks.yml` (add 4 steps)
- Modify: `internal/devsecops/policy_test.go` (extend `requiredProjectSteps` map near line 1031)
- Test: `internal/devsecops/policy_test.go`; also add the untested `isCompliant` branch test in the adapter module's report tests

**Interfaces:**
- Consumes: nothing from other tasks.
- Produces: four new CI steps with exact names; extended `requiredProjectSteps` map entries.

- [ ] **Step 1: Write the failing policy test**

Extend the `requiredProjectSteps` map in `internal/devsecops/policy_test.go` with four entries:
- `"Test PDF/UA-2 adapter module"`: `{workingDirectory: "ci/pdfua2/adapter", command: "go test -race ./... -count=1"}` (adapter tests must run without the real CLI jar; the workflow must not set `PDFUA_REAL_CLI_JAR`)
- `"Vet PDF/UA-2 adapter module"`: `{workingDirectory: "ci/pdfua2/adapter", command: "go vet ./..."}`
- `"Check adapter module files"`: `{workingDirectory: "ci/pdfua2/adapter", command: "go mod tidy -diff"}`
- `"Check cloud bootstrap module files"`: `{workingDirectory: "cmd/cloud-tools-bootstrap", command: "go mod tidy -diff"}`

Run the policy test; expected FAIL (steps missing from workflow).

- [ ] **Step 2: Add the four steps to `.github/workflows/go-checks.yml`**

Insert after the existing "Vet standalone cloud bootstrap module" step, keeping the file's existing style (pinned action SHAs untouched). For the adapter test step, do not set `PDFUA_REAL_CLI_JAR` (process-contract tests only, per AGENTS.md).

- [ ] **Step 3: Cover the untested `isCompliant` branch**

In the adapter module's report tests, add a test with `isCompliant` set to a value that is neither `"true"` nor `"false"` (e.g., `"maybe"`), asserting the parse returns the `"validationReport isCompliant must be true or false"` error. Run: `(cd ci/pdfua2/adapter && go test -race ./... -count=1)`. Expected: PASS.

- [ ] **Step 4: Run the full policy test package**

Run: `go test -race ./internal/devsecops/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/go-checks.yml internal/devsecops/policy_test.go ci/pdfua2/adapter/
git commit -m "fix(ci): run nested-module checks in go-checks.yml (dap-jo9.5.6)"
```

### Task 3: Strict manifest decoding and parser fuzzing (`dap-jo9.2.1.4`, P2)

**Files:**
- Modify: `internal/corpus/manifest.go` (reject duplicate and case-variant keys)
- Modify: catalog loader dropping misspelled fields (find via `LoadCatalog`; add `DisallowUnknownFields` or equivalent)
- Test: new `FuzzParseReport` plus regression test for the `"PROVENANCE"`/`"provenance"` override

**Interfaces:**
- Consumes: nothing from other tasks.
- Produces: strict decoders; `FuzzParseReport` seed corpus.

- [ ] **Step 1: Write the failing regression test**

Add a test that loads a manifest containing both `"PROVENANCE":"synthetic-only"` and `"provenance":"private-customer-reports"` and asserts `Load` returns an error (duplicate/case-variant key). Run it; expected FAIL (currently the second silently overrides the first).

- [ ] **Step 2: Implement strict key handling in `internal/corpus/manifest.go`**

Decode into a `map[string]json.RawMessage` first (or use a decoder hook); reject if any two keys are equal under case-folding. Keep the existing `DisallowUnknownFields` behavior for struct fields.

- [ ] **Step 3: Make the catalog loader strict**

Find the `LoadCatalog` implementation; add unknown-field rejection so misspelled fields fail loudly instead of being dropped. Add a test with a misspelled field asserting an error.

- [ ] **Step 4: Add `FuzzParseReport`**

Fuzz target on the report parser (`internal/pdfua/report.go`): assert parsing never panics, and a report with inconsistent counts is never classed compliant. Seed the corpus with: a valid compliant report, a valid non-compliant report, and a report with `isCompliant="maybe"`. Run briefly: `go test -fuzz=FuzzParseReport -fuzztime=30s ./internal/pdfua/`. Expected: no failures.

- [ ] **Step 5: Run affected packages and commit**

Run: `go test -race ./internal/corpus/ ./internal/pdfua/ ./internal/traceability/`
Expected: PASS

```bash
git add internal/corpus/ internal/pdfua/ internal/traceability/
git commit -m "fix(parsing): strict manifest decoding and report fuzzing (dap-jo9.2.1.4)"
```

### Task 4: Doc contradictions and process-cleanup unification (`dap-jo9.5.7`, P2)

**Files:**
- Modify: `SECURITY.md:17`, `CONTRIBUTING.md:87`, `docs/validation/README.md:14`
- Modify: `ci/pdfua2/adapter/process_linux.go`, `cmd/verapdf-container-runner/process_linux.go`
- Test: existing suites for the touched packages; no new tests required except a regression pinning signal-before-reap ordering if the codebase has a harness for it

**Interfaces:**
- Consumes: nothing from other tasks.
- Produces: corrected docs; unified cleanup helper (or documented ordering invariant).

- [ ] **Step 1: Fix the three doc contradictions**

  - `SECURITY.md:17`: change "runs tests and `go vet`" to describe what the security-scan job actually runs (read the workflow first; do not invent).
  - `CONTRIBUTING.md:87`: list the actual direct dependencies from `go.mod` (three), or scope the pdfcpu claim correctly.
  - `docs/validation/README.md:14`: scope the boundary-test claim to what `internal/pdfua/boundary_test.go` actually scans, or extend the test — pick the smaller correct change and note the choice in the commit message.

- [ ] **Step 2: Unify process-group cleanup**

Read both `process_linux.go` files. Extract the correct signal-before-reap ordering into one shared helper (or duplicate the correct ordering with a comment pointing at the canonical copy — pick whichever matches the codebase's existing sharing pattern; these are separate modules so a shared package may not be appropriate). The invariant: signal the process group **before** the leader is reaped, closing the PID-reuse window.

- [ ] **Step 3: Run the affected checks**

Run: `go test -race ./ci/pdfua2/adapter/ ./cmd/verapdf-container-runner/ ./internal/pdfua/` plus `go vet` on the same. Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add SECURITY.md CONTRIBUTING.md docs/validation/README.md ci/pdfua2/adapter/process_linux.go cmd/verapdf-container-runner/process_linux.go
git commit -m "docs(fix): correct contradictions; unify process-group cleanup (dap-jo9.5.7)"
```

---

## Self-review

**1. Spec coverage:** Panel P1s → Tasks 1–2. Manifest/fuzz P2+P3 → Task 3. Doc contradictions + process cleanup P2+P3 → Task 4. The pdfcpu font-file P2 is accepted risk per analysis (documented in the panel report, no task). Fixture publish/rollback coverage and the smaller issues are explicitly untracked follow-ups (noted in the panel report, no task — deliberately, to keep this plan shippable).

**2. Step scan:** Each step names one action with a checkable result. Task 4 Step 1 contains three doc edits in one step — acceptable because each is a one-line correction and none is independently rejectable from the others; a reviewer would accept or reject the "docs are now accurate" unit together.

**3. Type consistency:** No cross-task interfaces beyond Beads item IDs; each task is self-contained. The plan references exact file paths and the exact `requiredProjectSteps` map.

**4. Review Focus:** Covered above; each line has its owning task (policy test asserts on parsed artifact in Task 1; name-drift in Task 2's map; fuzz seed in Task 3; case-fold test in Task 3; both platform files in Task 4).

**5. Proportion:** Four tasks for four Beads items; steps are actions, not code transcripts. The longest code-like content is the map entries in Task 2, which are exact values the implementer cannot invent.
