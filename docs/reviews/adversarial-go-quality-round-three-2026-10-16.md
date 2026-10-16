# Adversarial Go quality review — round three

**Current status (2026-10-16; main `76f14f69a3283715f5d5278983363121168a2f09`):** PR #84 merged as `4938882da31eca44c22346db784c9ad5b7fa5b4e`. Its root parser regressions, adapter timeout/early-exit controls, and three issue-field refinements are in main. `dap-jo9.18` was subsequently created by PR #87 and closed in the canonical issue export by the scoped workflow in PR #97. See the [closure review](dap-jo9.18-closure-candidate-2026-10-16.md) and [Decision 0008](../decisions/0008-scoped-dap-jo9-18-closure.md). This records an export closure only; the live embedded Beads database was not changed or verified. The PR #84 publication-state statements below describe the review-time snapshot, not current main.

**Reviewed baseline:** `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac` (`origin/main` on 2026-10-16). **PR #84 code candidate at review time:** `cfac3a8a5baf2fa7313ccdcb87cb44caa8af64bc`; later commits on `codex/round-three-review` updated review documentation only. The branch was published through the guarded source publisher.

**Scope:** validator report parser mutations and tests; existing `beads-export-reconcile` workflow, evidence bundle, promotion, and binary execution trust boundary; planned-field null/empty handling.

**Method:** read-only source review, temporary one-at-a-time source mutations with explicit acceptance probes, root and nested focused tests, disposable Beads stores, and existing reconciler integration tests. No product source was changed. Mutated source was restored byte-for-byte before finishing.

## Findings

| Disposition | Finding | Evidence / tracking |
|---|---|---|
| **Baseline test gap; now covered by PR #84 and tracked by `dap-jo9.2.1.4`** | At the reviewed baseline, the root parser suite did not reject a second `<validationReports>` element or a summary count other than `1`; disabling either predicate survived both root tests and the nested adapter tests. | Mutants 3 and 5 below. The assertions belong to `.2.1.4`; `.5.6` retains module-gate and policy coverage. |
| **Real root-only gap; nested adapter already covers it** | Root tests do not reject `failedJobs="1"`; disabling the check survives root tests but fails nested adapter cases for the batch summary, missing attribute, and empty numeric counter. | Mutant 4 below. Do not add duplicate adapter coverage. |
| **Root-only gaps; nested adapter already covers them** | Root tests do not reject non-whitespace before or after the root element. Both mutants survive root tests and are rejected by nested adapter leading/trailing cases. | Mutants 1–2 below. |
| **Existing tracked work** | The current workflow runs root, `cmd/cloud-tools-bootstrap`, and `ci/pdfua2/adapter` tests/vet/tidy checks, and `internal/devsecops/policy_test.go` asserts those steps. | `dap-jo9.5.6`; current module coverage is present. Automatic discovery of future modules would be maintenance hardening, not a missing current gate. |
| **Existing docs task, corrected here** | The remediation inventory incorrectly called pdfcpu the only direct module in root `go.mod`. The actual direct modules are pdfcpu, go.yaml.in/yaml/v3, and golang.org/x/sys. | `dap-jo9.5.7`; corrected in `docs/research/remediation-component-inventory.md`. Current `SECURITY.md`, `CONTRIBUTING.md`, and validation-boundary text match their code/workflow boundaries. |
| **Optional hardening / test coverage** | `runCommand` verifies the selected Beads path and SHA before and after each command but executes by pathname with `exec.CommandContext`. A concurrent writer able to replace that path between verification and process start could cause another binary to run before the post-check detects the replacement. | Conditional local filesystem race, not demonstrated as an exploit. Consider running a verified immutable copy and adding a deterministic replacement test. |
| **Optional test coverage** | Bundle code rejects duplicate, missing, unexpected, and oversized ZIP entries and validates command provenance, but direct malformed-ZIP tests and a mutate-then-recompute-outer-digest case are absent. | `readBundle`, `validateManifest`, and promotion checks enforce these conditions by inspection; this review did not find a demonstrated acceptance bypass. |
| **Rejected claim** | “Six mutants” is inaccurate: the six review steps comprised five actual mutants plus one restoration/control step. | Explicitly reported as five mutants below. |
| **Rejected claim** | The alleged planned-string empty-value failure is not present on this revision. JSON `null` is rejected and explicit `""` is accepted. | `internal/beadsreconcile/reconcile_test.go:80` and `cmd/beads-export-reconcile/main_test.go:138`. |

The parser assertions in the first four rows are tests for the existing contract; this review found no reason to change the parser implementation. The five mutation probes tested one predicate at a time. Each probe used a temporary test that confirmed the mutated parser accepted the invalid report; the temporary test and source mutation were removed before completion. The final `report.go` SHA-256 matched the pristine revision: `6e8d5524532a3aeb6ddb3d37939fff43edbb100491571e94b9c076ebd990687f`.

## Parser mutation results

All commands used Go 1.27.2 from the isolated cloud toolchain. `go test ./internal/pdfua -count=1` passed on baseline. The nested parser-focused suite passed on baseline:

```sh
PDFUA_REAL_CLI_JAR= go test -count=1 -run '^(TestValidateClassifiesUA2Reports|TestParseReportRejectsMissingSummaryAttributes|TestParseReportRejectsDuplicateClassificationAttributes|TestParseReportRejectsNamespaceShadowing|TestParseReportRejectsEmptyRequiredNumericCounters|TestParseReportErrorsRedactMalformedTokens|TestParseReportRejectsNonWhitespaceOutsideRoot)$' ./...
```

| # | Predicate disabled | Root suite | Nested focused suite | Result |
|---:|---|---|---|---|
| 1 | Reject non-whitespace before the root | survived | failed `TestParseReportRejectsNonWhitespaceOutsideRoot/leading` | Adapter catches; root coverage gap. |
| 2 | Reject non-whitespace after the root | survived | failed `TestParseReportRejectsNonWhitespaceOutsideRoot/trailing` | Adapter catches; root coverage gap. |
| 3 | Require exactly one `validationReports` element | survived | survived | True gap across both suites. A temporary acceptance probe confirmed a duplicate valid summary was accepted. |
| 4 | Require `failedJobs == "0"` | survived | failed the batch-summary, missing-attribute, and empty-counter cases | Adapter catches; root coverage gap. |
| 5 | Require the validation report count to equal `1` | survived | survived | True gap across both suites. A temporary acceptance probe confirmed content `2` was accepted. |

The baseline nested adapter full suite exposed a deterministic test-setup mismatch in `TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles`: its leader exited before the 300 ms timeout, and the adapter correctly returned the malformed-XML early-exit result. The test-only fix now holds the leader until group termination for the true-timeout case and separately asserts pre-deadline malformed-report behavior for early exit. Both paths retain descendant and temporary-file cleanup assertions. The corrected focused and full adapter suites pass, with and without `-race`, on the exact local candidate; results are recorded below.

## Reconciler and trust-boundary review

The reconciler's `create` and `promote` integration tests were run with the pinned Beads 1.3.1 executable and temporary `t.TempDir` stores only:

```sh
DAP_BD_BINARY=/workspace/.codex-tools/cloud-build-tools/bin/bd \
  go test ./cmd/beads-export-reconcile \
    -run '^(TestCreateAndPromoteWithPinnedBeads|TestPromoteRejectsStaleCanonical)$' \
    -count=1 -v
```

Both passed. The first created a seven-entry evidence bundle and promoted its reviewed candidate to a temporary JSONL file. The second detected a changed canonical source and left that file unchanged. Separately, the complete package tests passed:

```sh
go test ./cmd/beads-export-reconcile ./internal/beadsreconcile -count=1
```

At the time of PR #84, the three existing-issue refinements were regenerated from the exact approved source and promoted through the reconciler in an isolated clean-main worktree. Source SHA-256: `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`; candidate SHA-256: `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`; evidence ZIP SHA-256: `7123c9925acfa0579cb75f5637425a6d2a459a0376c160f955bdad1083db4acc`. The export was included in PR #84 as a normal linear commit and its digest matched that candidate's `.beads/issues.jsonl`. PR #84 preserved the three issue statuses; PR #97 later closed `.18` in the canonical export.

Source and tests show these controls:

- `readRegularBounded` accepts only regular, non-symlink, single-link files and enforces 32 MiB canonical, 1 MiB plan, and 256 MiB Beads binary caps. Command output capture retains at most 1 MiB while hashing the full stdout/stderr streams. Evidence bundles are capped at 64 MiB with 32 MiB per entry.
- The manifest records the binary SHA, Beads version, exact command arguments, exit status, working directory, environment digest, and full-stream stdout/stderr digests. Validation requires the expected command count and argument sequence with one executable identity. The path identity and digest are checked before and after each command.
- `readBundle` checks the seven expected entry names, duplicate/missing/unexpected entries, and each decompressed entry cap. Promotion verifies the externally recorded bundle digest, manifest/snapshot/candidate digests and semantics, round-trip evidence, current source hash, and current file identity.
- `atomicReplace` acquires an exclusive sibling lock; rechecks source hash and file identity; writes, syncs, and closes a sibling temporary file; rechecks source bytes immediately before rename; then syncs the containing directory. The code documents that the lock is cooperative, not a filesystem compare-and-swap. Existing tests cover stale-source rejection, lock contention without source change, and digest mismatch. Positive create/promote integration ran only against disposable files.

The remaining executable trust-boundary concern is conditional: the executable is hash-checked by path, then launched by that same path. If an untrusted concurrent writer can replace the executable between those operations, the replacement could run before the post-execution hash check notices. This review did not demonstrate a race. Treat immutable-copy execution and a deterministic replacement test as optional hardening unless the threat model grants an attacker write access to that path or its parent.

The existing ZIP negative controls were inspected rather than extended: no direct tests cover duplicate/missing/unexpected entries, decompressed-size limits, or a modified ZIP paired with a newly supplied external digest. These are coverage opportunities; source inspection found the corresponding validation checks. Do not describe them as a confirmed bypass.

No Beads commands were run against the live Dolt store. `bd prime`, `bd ready`, and `bd show` were successfully used after importing the tracked export into a fresh disposable workspace rooted under `/tmp`; the initial scratch checkout's read-only-home database initialization failure was left untouched. Pinned reconciler integration also used its own disposable stores. The review inspection changed no tracker data; the two issue text updates in this PR candidate were later produced and promoted to the isolated worktree export by the guarded reconciler. No issue status was changed, and no `.5.5.closed_at` value was fabricated.

## Tracker and scope disposition

- `dap-jo9.2.1.4` is the appropriate parser-assurance task for the two direct root `ParseReport` tests, alongside its strict loader/fuzz requirements. It should retain decoder-specific requirements from round-two Task 1; PR #83's shared `jsonstrict` helper tests are already merged and need no duplicate task.
- **PR #84 result:** the reviewed three-issue refinement is merged. `.5.6` retains root/bootstrap/adapter CI-gate responsibility; `.2.1.4` owns direct parser cardinality assurance; `.5.7` retains process lifecycle and documentation scope. PR #84 refined fields and preserved their statuses and dependencies. On current main, `.5.6`, `.2.1.4`, and `.5.7` remain open; `.5.7`'s timeout controls are implemented, and the canonical export records no acceptance or closure.
- `.5.6` now retains its CI/policy requirement for root, bootstrap, and adapter module checks; the two parser assertions moved to `.2.1.4` in this candidate.
- `.5.7` remains the existing scope for process lifecycle test invariants and documentation. The adapter's `waitForLeaderExitNoReap` observes leader exit with `WNOWAIT`, signals the process group, then reaps the leader; the historical PID-reuse concern is not a demonstrated current adapter defect. Separate true-timeout and early-leader-exit controls are implemented, including descendant and temporary-file cleanup checks, without claiming a production timeout regression.
- `dap-jo9.5.5`'s historical importer timestamp exception remains unchanged. No missing closure time may be invented.
- The other round-two loader boundaries (`author_v2`, observer decoding, and `LoadIssues`) remain separate from map-key semantics and direct `jsonstrict` tests. PR #83's direct Unicode-folding helper tests are already in current main and are not a new task here.
- Product semantic and assistive-technology acceptance gates remain open. This parser/reconciler review supplies no product or PDF/UA acceptance evidence.

The PR #84 issue export linked `dap-jo9.5.6`, `dap-jo9.2.1.4`, and `dap-jo9.5.7` to this assessment. Issue statuses were unchanged by that PR; no close operation was included in PR #84.

An independent read-only review of the exact candidate diff against `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac` found no material factual or workflow defects. It confirmed the issue-field scope, preserved `.5.5` metadata, parser mutation dispositions, and inventory/workflow claims.

## Publication-candidate verification

The Go code candidate was `d68d13991d26d69162b1dc935abac01ff2a22ee1`; the published report/plan refresh at `cfac3a8a5baf2fa7313ccdcb87cb44caa8af64bc` changes documentation only. Root and bootstrap validations below ran on the code candidate; focused adapter commands were rerun with verbose output on the published head:

- Root `go test ./... -count=1`, `go vet ./...`, and `go mod tidy -diff` passed.
- `cmd/cloud-tools-bootstrap` tests, vet, and tidy-diff passed.
- `ci/pdfua2/adapter`: the focused timeout and early leader-exit tests passed with verbose test-name output on `cfac3a8`. Exact commands from that module, using Go 1.27.2 and `PDFUA_REAL_CLI_JAR=`, were:

  ```sh
  go test ./... -run '^(TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles|TestValidateCleansProcessGroupAfterLeaderExits)$' -count=1 -v
  go test -race ./... -run '^(TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles|TestValidateCleansProcessGroupAfterLeaderExits)$' -count=1 -v
  ```

  The focused tests passed in 1.120s and 4.155s respectively. The full adapter suite also passed both without `-race` and with `-race` on `d68d139`; that test code is unchanged in `cfac3a8`. Commands: `go test ./... -count=1` (5.073s) and `go test -race ./... -count=1` (23.424s).
- `go run ./cmd/traceability-check` passed with 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases.
- Fresh disposable Beads import and `bd show` confirmed `.5.6`/`.5.7` remain open with their parents intact. Artifact checks passed. The conventions check reported one pre-existing warning on unrelated `dap-jo9.4.2` (missing `Rationale` and `Alternatives Considered`); no new warning is attributable to these two issue updates.
- `git diff --check` passed. The source-publisher commit identity/subject checks and guarded push passed for `cfac3a8`; the publisher fetched and verified the remote PR ref at the exact object. At this validation snapshot PR #84 was Draft; it later merged as `4938882da31eca44c22346db784c9ad5b7fa5b4e`.

### Hosted race-check follow-up

The first hosted run on the initial PR head failed in `TestRunCommandTerminatesDescendantsAfterLeaderExit`: the test checked immediately after `SIGKILL` and observed a still-live descendant. The failure reproduced in two of ten local race-enabled repetitions. Inspection shortly afterward showed the processes had transitioned to zombies, indicating an observation race rather than a production cleanup failure. The Linux test helper now polls for up to 500 ms at 5 ms intervals, accepting only a gone, exited, or zombie process; it still fails if the descendant remains live through that deadline. No production code changed. An independent read-only review found no issue with this bounded test synchronization.

Validation after that earlier test-only adjustment: the focused process case passed 20 race-enabled repetitions, and `go test -race ./... -count=1` passed across the root modules. The independent reviewer also passed root, bootstrap, and adapter race suites on `0646db6`; the earlier hosted failure was on `fc9f1b4`, and all eight hosted jobs succeeded on `0646db612b9d115123a0a80e4c62f9bc43cb0f6d`. All nine required hosted checks succeeded on code-bearing head `cfac3a8`: Go checks, go-security, secrets, validator-container, pdfua2, dependency review, CodeQL actions, CodeQL Go, and the required checks gate. That result does not certify later documentation-only commits; verify every required check on the final PR head before merge. The baseline adapter timeout test setup mismatch is corrected by test-only changes; no production adapter change was needed.

## Follow-up tracking plan and dispositions

The following table records the round-three disposition; its implementation references are now merged through PR #84. On current main, `.2.1.4` remains open for loader/parser criteria beyond the two cardinality tests; `.5.6` remains open as the required module-gate preservation obligation; and `.5.7` remains open with no acceptance or closure recorded in the canonical export. Bounded issue creation later merged in PR #85, and `.18` was created in PR #87 and closed in the canonical export by PR #97. The separate round-two Tasks 3, 4, 5, and 7 still have no matching issue IDs in the current canonical export; do not create or infer IDs from this plan. The live embedded database remains unchanged and unverified.

| Priority / issue | Actionable work and disposition |
|---|---|
| P2 — `dap-jo9.2.1.4` (existing) | Direct `ParseReport` negative tests for duplicate `validationReports` and count other than one are implemented in this candidate, each killed by its corresponding predicate mutation with compliant/finding positives. Remaining strict manifest/loader, fuzz, and round-two Task 1 cases for corpus, `author_v2`, observer decoding, and `LoadIssues` remain open with decoder-specific compatibility rules. Shared helper behavior already tested in merged PR #83 is not repeated. |
| P1 — `dap-jo9.5.6` (existing) | Parser acceptance was moved to `.2.1.4`; `.5.6` retains the `go-checks.yml` and policy requirements for tests, vet, and tidy-diff across root, `cmd/cloud-tools-bootstrap`, and `ci/pdfua2/adapter`. Existing CI acceptance was not weakened. |
| P2 — `dap-jo9.5.7` (existing) | The adapter leader observation, `WNOWAIT` ordering, and descendant cleanup remain covered. The timeout test now has separate leader-held-past-deadline and early-leader-exit cases with outcome, timing, and temp cleanup assertions. Do not record a PID-reuse or production timeout defect without contrary evidence. Remaining issue criteria stay open. |
| P2 — `dap-jo9.18` (now closed in the canonical export) | This row describes its original proposed scope: adversarial tests for bundle parsing, provenance binding, tampering, atomic replacement, cooperative-lock behavior, and documented size bounds. The reviewed coverage landed across PRs #89, #90, #92, #93, #94, and #96. PR #97 separately added the exact-source typed closure workflow. See the current [closure review](dap-jo9.18-closure-candidate-2026-10-16.md). |
| Round-two Task 2 — completed slice plus residual above | PR #80's counter tests and current exact-head verification cover the five summary counters. Round three found only the two root `validationReports` cardinality gaps; track those in `.2.1.4`. No parser implementation change is proposed. |
| Round-two Task 3 — open; no issue ID present in current export | Bootstrap module/version identity, prerelease policy, archive `VersionSource`, and raw-binary positive controls remain outside this review. Preserve the round-two exact-coordinate tests; do not infer a shipped wrong-version install from source-only acceptance. |
| Round-two Task 4 — open; no issue ID present in current export | Policy mutation tests for skipped required jobs/steps, `continue-on-error`, ignored failures, and effective-shell behavior remain outstanding. `.5.6` covers required module gate presence only; it does not substitute for these Task 4 controls. Do not broaden `.5.6` without an accepted tracker update. |
| Round-two Task 5 — open; no issue ID present in current export | Dependabot coverage for the adapter module and a rule for bootstrap's currently empty external dependency set remain outstanding. Existing vulnerability and license checks are separate from dependency-update automation. |
| Round-two Task 6 — optional P3 decision | Bounded scheduled/manual fuzz CI remains an optional decision after useful fuzz targets and triage/corpus retention are defined. No generated corpus alone is treated as a defect. |
| Round-two Task 7 — open; no issue ID present in current export | Failure-injected financial fixture pair rollback and bootstrap `run()` CLI tests remain outstanding. This review did not exercise those paths or make rollback claims. |
| Round-two Task 8 — optional P3 assessment | Process helper consolidation remains optional. Preserve and test behavior (leader exit ordering, bounded reaping, inherited-pipe handling, descendant cleanup, lock lifetime); helper count is not an acceptance criterion. The adapter currently uses `WNOWAIT` before group signaling. |

### Beads issue-creation blocker

No new issue ID was assigned. The `beads-export-reconcile` plan schema accepts field updates to existing issue IDs; its `create` subcommand creates a reviewable evidence bundle, not a Beads issue. The earlier attempt to promote an old bundle over changed source bytes correctly failed closed; the three existing-issue updates were then regenerated from the authorized source, promoted in a clean worktree based on that exact source, and added as a normal PR commit. No source-lineage migration is needed while `origin/main` remains at the authorized source snapshot.

The candidate contains `.beads/issues.jsonl` SHA-256 `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`; its parent source SHA-256 is the approved `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`; the pinned Beads 1.3.1 binary SHA-256 is `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. All five allowlisted explicit-null field pairs remain unchanged. At the time of this review, this new source digest was not covered by the exception, so another reconciliation was deferred pending a separately reviewed continuation. Decision 0006 later recorded approval for that digest. PR #87 then added `dap-jo9.18`, yielding exact source digest `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`; the repository owner separately approved this one continuation in Decision 0007. Further digests remain unauthorized.

The complete three-issue refinement was promoted via the supported reconciler path from the unchanged approved source, preserving all five null tokens and all issue IDs and dependency edges. No attempt broadened or re-keyed the exception, directly edited canonical JSONL, or used the live Dolt store. New records still require the separately reviewed reconciler issue-creation capability; keep unassigned work plan-only until that exists.
