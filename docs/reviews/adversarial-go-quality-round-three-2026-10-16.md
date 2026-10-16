# Adversarial Go quality review — round three

**Reviewed revision:** `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac` (`origin/main` on 2026-10-16).

**Scope:** validator report parser mutations and tests; existing `beads-export-reconcile` workflow, evidence bundle, promotion, and binary execution trust boundary; planned-field null/empty handling.

**Method:** read-only source review, temporary one-at-a-time source mutations with explicit acceptance probes, root and nested focused tests, disposable Beads stores, and existing reconciler integration tests. No product source was changed. Mutated source was restored byte-for-byte before finishing.

## Findings

| Disposition | Finding | Evidence / tracking |
|---|---|---|
| **Real test gap; tracked by `dap-jo9.5.6`** | The root parser suite does not reject a second `<validationReports>` element or a summary count other than `1`; disabling either predicate survives both root tests and the nested adapter tests. | Mutants 3 and 5 below. The two assertions belong in the root parser tests. |
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

The nested adapter full suite has a separate reproducible baseline failure: `TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles` expects a timeout but receives `invalid validator XML report: malformed validator XML report` after about 0.21 seconds. This was also reproduced with that test alone. It is not caused by any mutation. No fix was attempted in this review.

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

The issue updates in this review candidate were also generated and promoted through the reconciler. Bundle record: source SHA-256 `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`; candidate SHA-256 `785e8713c1d4175f6371898410a3f22db2f1b95e4a3e70a7185e65264bfd7ec0`; evidence ZIP SHA-256 `a90b74ef20bd0b026987cbd95ed19a3269f51cbd73c902c5fad706ead47afee4`. The ZIP is retained in the review workspace at `/tmp/dap-round3-issues.zip`; the candidate digest matches the `.beads/issues.jsonl` included in this branch.

Source and tests show these controls:

- `readRegularBounded` accepts only regular, non-symlink, single-link files and enforces 32 MiB canonical, 1 MiB plan, and 256 MiB Beads binary caps. Command output capture retains at most 1 MiB while hashing the full stdout/stderr streams. Evidence bundles are capped at 64 MiB with 32 MiB per entry.
- The manifest records the binary SHA, Beads version, exact command arguments, exit status, working directory, environment digest, and full-stream stdout/stderr digests. Validation requires the expected command count and argument sequence with one executable identity. The path identity and digest are checked before and after each command.
- `readBundle` checks the seven expected entry names, duplicate/missing/unexpected entries, and each decompressed entry cap. Promotion verifies the externally recorded bundle digest, manifest/snapshot/candidate digests and semantics, round-trip evidence, current source hash, and current file identity.
- `atomicReplace` acquires an exclusive sibling lock; rechecks source hash and file identity; writes, syncs, and closes a sibling temporary file; rechecks source bytes immediately before rename; then syncs the containing directory. The code documents that the lock is cooperative, not a filesystem compare-and-swap. Existing tests cover stale-source rejection, lock contention without source change, and digest mismatch. Positive create/promote integration ran only against disposable files.

The remaining executable trust-boundary concern is conditional: the executable is hash-checked by path, then launched by that same path. If an untrusted concurrent writer can replace the executable between those operations, the replacement could run before the post-execution hash check notices. This review did not demonstrate a race. Treat immutable-copy execution and a deterministic replacement test as optional hardening unless the threat model grants an attacker write access to that path or its parent.

The existing ZIP negative controls were inspected rather than extended: no direct tests cover duplicate/missing/unexpected entries, decompressed-size limits, or a modified ZIP paired with a newly supplied external digest. These are coverage opportunities; source inspection found the corresponding validation checks. Do not describe them as a confirmed bypass.

No Beads commands were run against the live Dolt store. `bd prime`, `bd ready`, and `bd show` were successfully used after importing the tracked export into a fresh disposable workspace rooted under `/tmp`; the initial scratch checkout's read-only-home database initialization failure was left untouched. Pinned reconciler integration also used its own disposable stores. The review inspection changed no tracker data; the two issue text updates in this PR candidate were later produced and promoted to the isolated worktree export by the guarded reconciler. No issue status was changed, and no `.5.5.closed_at` value was fabricated.

## Tracker and scope disposition

- `dap-jo9.5.6` remains the appropriate existing task for current nested-module checks and the two surviving root parser assertions. Do not create a duplicate module-coverage issue. Its current issue text has stale “modules are not checked” and `isCompliant` claims; update it to the two remaining assertions and note current module gates.
- `dap-jo9.5.7` remains the existing home for correcting the inventory's direct-module statement and any separately demonstrated process-group behavior. Do not treat helper-file count or consolidation by itself as a defect; retain only behaviorally demonstrated acceptance criteria.
- `dap-jo9.5.5`'s historical importer timestamp exception remains unchanged. No missing closure time may be invented.
- The other round-two loader boundaries (`author_v2`, observer decoding, and `LoadIssues`) remain separate from map-key semantics and direct `jsonstrict` tests. PR #83's direct Unicode-folding helper tests are already in current main and are not a new task here.
- Product semantic and assistive-technology acceptance gates remain open. This parser/reconciler review supplies no product or PDF/UA acceptance evidence.

The issue export changes in this candidate link `dap-jo9.5.6` and `dap-jo9.5.7` to this assessment. Issue statuses remain open; no close operation is included.

An independent read-only review of the exact candidate diff against `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac` found no material factual or workflow defects. It confirmed the issue-field scope, preserved `.5.5` metadata, parser mutation dispositions, and inventory/workflow claims.

## Publication-candidate verification

On the clean candidate worktree based on the reviewed main SHA:

- Root `go test ./... -count=1`, `go vet ./...`, and `go mod tidy -diff` passed.
- `cmd/cloud-tools-bootstrap` tests, vet, and tidy-diff passed.
- `ci/pdfua2/adapter` vet and tidy-diff passed. Its full test suite fails only at the separately documented baseline timeout test.
- `go run ./cmd/traceability-check` passed with 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases.
- Fresh disposable Beads import and `bd show` confirmed `.5.6`/`.5.7` remain open with their parents intact. Artifact checks passed. The conventions check reported one pre-existing warning on unrelated `dap-jo9.4.2` (missing `Rationale` and `Alternatives Considered`); no new warning is attributable to these two issue updates.
- `git diff --check` passed. The source-publisher script's local safety tests passed. No GitHub PR or external branch update has been made by this review.

### Hosted race-check follow-up

The first hosted run on the initial PR head failed in `TestRunCommandTerminatesDescendantsAfterLeaderExit`: the test checked immediately after `SIGKILL` and observed a still-live descendant. The failure reproduced in two of ten local race-enabled repetitions. Inspection shortly afterward showed the processes had transitioned to zombies, indicating an observation race rather than a production cleanup failure. The Linux test helper now polls for up to 500 ms at 5 ms intervals, accepting only a gone, exited, or zombie process; it still fails if the descendant remains live through that deadline. No production code changed. An independent read-only review found no issue with this bounded test synchronization.

Validation after that test-only adjustment: the focused case passed 20 race-enabled repetitions, and `go test -race ./... -count=1` passed across the root modules. The original hosted failure was on the earlier PR head; the adjusted commit requires fresh exact-head hosted checks. The nested adapter's separate `TestValidateTimeoutKillsDescendantAndCleansTemporaryFiles` baseline failure remains: it returns malformed XML around 0.21 seconds where the test expects a timeout. No adapter production or test code was changed in this review.
