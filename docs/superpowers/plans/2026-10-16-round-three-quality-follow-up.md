# Round-three adversarial Go quality follow-up

**Status:** PR #84's parser-test and adapter-test setup fixes are merged, and the three existing-issue field refinements were promoted from the authorized source. Synthetic planned-creation evidence capture is implemented and independently reviewed. The exact current-source continuation is now approved for the same five null pairs and pinned Beads 1.3.1 binary; code and disposable-store regressions cover it. Any live tracker promotion still waits for the complete behavioral suite and fresh independent reviews.

**Date and evidence cutoff:** 2026-10-16. Round-three review baseline: `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac`. Local PR #84 candidate before this document refresh: `d68d13991d26d69162b1dc935abac01ff2a22ee1`, based on then-current main `6ae3f03`. The parser tests, timeout-test contract correction, and synthetic creation evidence path are implementation results. Current-source promotion and the other unassigned workstreams remain open or planned as described below.

**Repository guidance used:** [`dap-test-driven-development`](../../../.agents/skills/dap-test-driven-development/SKILL.md), [`dap-systematic-debugging`](../../../.agents/skills/dap-systematic-debugging/SKILL.md), [`dap-verification-before-completion`](../../../.agents/skills/dap-verification-before-completion/SKILL.md), and [`dap-plan-review-integrator`](../../../.agents/skills/dap-plan-review-integrator/SKILL.md). The first three are local adaptations of imported Superpowers v7.0.0 skills; the review-to-plan integrator is the repository's local roundtable adaptation. The imported-skill inventory/provenance was checked. It does not include the upstream Superpowers `writing-plans` skill, so that skill is not claimed as used; this plan follows the existing repository plan format and the actual local TDD/debugging/review guidance.

## Decision summary

Round three found **test assurance gaps, not a demonstrated current parser bypass**. Two one-predicate mutations—allowing multiple `validationReports` elements and accepting summary content whose count is not exactly one—survived both root and nested focused suites. Other mutations for `failedJobs`, leading text, and trailing text survived root tests but were rejected by the nested adapter. The shared `jsonstrict` Unicode duplicate-key regression from PR #83 is already merged; do not duplicate it.

The baseline nested adapter's full timeout test had a reproducible **test setup mismatch**: its `SPAWN_PIPE_HOLDER` leader exited after descendant readiness, before the configured 300 ms deadline. The adapter returned the early-exit malformed-XML result after its process-group cleanup grace (about 0.21 s). This did not demonstrate a production timeout defect. The candidate now keeps the leader alive through the true timeout and separately expects the early-exit malformed-report result before the deadline.

Stale claims in existing issue text have been reconciled against present work and source lineage. Parser assurance is assigned to `.2.1.4`, while `.5.6` retains root/nested CI and policy gate coverage. The reviewed fields were promoted from the authorized source; issue statuses and dependencies remain unchanged.

## Severity and execution priority

Priority below is an execution recommendation; it is not a claim that the issue has that priority in Beads. Current production risk, test gaps, stale claims, and optional hardening are separated:

| Class | Evidence | Execution priority |
|---|---|---|
| Current required-suite test failure | Baseline adapter timeout helper exited before its deadline and yielded the early-exit error. | **Resolved test setup:** split true-timeout and early-exit cases; focused and full adapter suites pass normal and race modes on the candidate. `.5.7` remains open. |
| Regression assurance gaps | Two XML parser predicate mutations survived root and nested focused tests; the parser still contained both predicates, so this was not a demonstrated bypass. | **Added in candidate:** root-owned tests under `.2.1.4`, each killed by its corresponding mutation. `.2.1.4` remains open. |
| Existing boundary contract gaps | Remaining round-two decoder-specific boundaries are bootstrap manifest, `author_v2`/observer decoding, and `LoadIssues`. Corpus/catalog work and PR #83 helper tests are already complete. | **P2:** test only remaining consumer contracts; do not impose blanket strict decoding. |
| CI/policy/dependency/transaction assurance gaps | Round-two Tasks 3, 4, 5, and 7 remain open and currently lack active Beads IDs verified for their specific controls. | **P2:** create/reconcile valid issue ownership before implementation; plan-only until then. |
| Stale documentation/tracker scope | Earlier `.5.6` mixed CI ownership with parser tests; `.2.1.4` lacked direct cardinality requirements. | **Resolved in candidate:** the existing-issue fields were reconciled through the supported workflow from the authorized source. |
| Optional hardening | Scheduled/manual fuzz CI and helper consolidation are not current defects. | **P3 / decision only:** do not implement without a demonstrated added safety benefit. |
| Conditional executable-path race | `runCommand` verifies a Beads executable path before and after invocation but does not execute an immutable copy; this review did not reproduce replacement. | **Defer / no tracker item:** revisit only if an untrusted concurrent writer can replace the executable or its parent-directory entry. |

No finding establishes a financial-output defect, PDF/UA nonconformance, or active security bypass. Preserve the original statement and invoice fixtures, equal family priority, human semantic review, and assistive-technology gates. Machine tests and CI are not substitutes for those gates.

## Tracker and source-lineage state

### Existing issue ownership: current versus intended

| Issue | Current canonical export in PR #84 | Intended refined assignment | State |
|---|---|---|---|
| `dap-jo9.5.6` (Beads P1) | Earlier acceptance criteria mixed root parser assertions with root/bootstrap/adapter Go gates. | Keep CI and policy coverage for all three current modules; parser predicates move to `.2.1.4`. Keep future module discovery optional. | Open; refined fields are present in the candidate. |
| `dap-jo9.2.1.4` (Beads P2) | Earlier criteria covered corpus/catalog key strictness and `FuzzParseReport`, but not direct root tests for multiple `validationReports` elements or count other than one. | Own decoder-specific loader requirements plus direct `ParseReport` regressions for both surviving mutations; preserve valid compliant and finding controls. | Open; refined fields and both added parser tests are present in the candidate. |
| `dap-jo9.5.7` (Beads P2) | Its text corrected stale documentation claims and records process behavior; the adapter timeout test had the setup mismatch. | Keep the corrected inventory and process-group invariants, distinguish early leader exit from actual timeout, and test both cleanup paths. | Open; refined fields are present in the candidate and both paths now pass adapter tests. |
| Round-two Tasks 3, 4, 5, 7 | No active issue ID was verified as owning these specific work items. `.5.6` owns required module gate presence, not policy fail-open controls. | Bootstrap pin identity; workflow-policy skip/fail-open/shell controls; adapter/bootstrap Dependabot coverage; financial fixture rollback and bootstrap CLI tests. | **Plan-only / unassigned.** Do not describe as fully tracker-mapped or completed. |
| Reconciler trust/atomicity negative tests and safe issue creation | No active issue ID was created. | Keep creation bounded to one explicit planned record; validate source lineage, exact edges, second-store round trip, and promotion evidence. | **Synthetic-only implementation; live promotion pending tests and review.** `create` captures the planned addition in disposable Beads stores and an evidence bundle. The exact current source digest `768d13…` has an approved continuation for the same five null pairs and pinned importer; no live Beads record or canonical export has been changed in this slice. |

### Exact Beads reconciliation evidence

The pinned importer is Beads 1.3.1, executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. Explicit-null compatibility is authorized only for the historic source SHA-256 `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890` and its exact approved continuation `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`. Both snapshots are bound to exactly these pairs: `dap-jo9.5.3.dependencies`, `dap-jo9.1.6.10.closed_at`, `dap-jo9.5.6.closed_at`, `dap-jo9.2.1.4.closed_at`, and `dap-jo9.5.7.closed_at`. No further source digest, binary, version, field pair, or graph delta is authorized.

The PR candidate `.beads/issues.jsonl` SHA-256 is `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`; its authorized source at `6ae3f03` has SHA-256 `7ce7e649…1125890`. A disposable import/export with the pinned binary confirmed the source has explicit `null` and B0 omits the field for all five listed pairs. An earlier attempt to use a stale bundle against changed source bytes failed closed before mutation. The complete update was then regenerated and promoted from the authorized source; no exception expansion or source migration was used.

The complete three-issue update plan (`.2.1.4`, `.5.6`, `.5.7`) was regenerated from the original approved source using the unchanged pinned exception. `create` succeeded with candidate SHA-256 `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de` and evidence bundle SHA-256 `7123c9925acfa0579cb75f5637425a6d2a459a0376c160f955bdad1083db4acc`. `promote` succeeded on a disposable copy and then in a clean worktree based on that exact authorized source. The resulting canonical export is included in the local PR candidate through a normal linear commit. No live Dolt store or direct canonical JSONL edit was used.

**Conclusion:** the three existing-issue refinements remain in the candidate and the five null tokens are preserved. The user has approved one exact continuation from the historic source digest to `768d13…`, retaining the same Beads binary/version and five pairs. The reconciler now recognizes only those two exact source snapshots. A fresh current-source disposable import/export and second-store regression covers that authorization; the historic bundle remains stale against current bytes. Do not extend the allowlist again without a separate approval. Any live tracker or canonical export promotion waits until full behavioral checks and independent reviews complete.

### Initial test-first prerequisite for issue creation (historical status)

At the time this prerequisite was written, the plan accepted only `updates` to IDs already present in canonical source, and `create` only produced an evidence bundle. Workstream F below implements the bounded synthetic `creates` plan and evidence flow. It imports the planned record only into a disposable Beads store; live tracker changes remain outside this implementation.

The original test-first acceptance criteria were:

1. **Creation contract.** Add plan fixtures for one new issue and its exact parent dependency. Specify required canonical fields, valid IDs, planned dependency shape, and source/baseline/candidate counts. Cover duplicate IDs and collisions.
2. **Scoped creation.** Accept only explicit planned records; forbid changes or deletions to unlisted records or edges; validate graph invariants and verify second-store round-trip. Preserve existing issue bytes and dependency records.
3. **Source freshness control:** retain the existing stale-source test and prove a bundle from a prior source cannot promote against a changed source. If source lineage ever changes, require a separately reviewed migration design that binds both snapshots and exact null pairs; never re-key the allowlist by configuration alone.
4. **Refactor and failure injection.** Inject malformed bundle, archive digest mismatch, recomputed outer-digest tampering, missing/duplicate/unexpected ZIP entries, altered manifest/provenance/snapshot/candidate, source race, lock refusal, size max/max+1, and pre-rename write/sync/close/rename failures. Assert unchanged source, temp cleanup, and lock release before rename; test post-rename directory-sync failure without claiming rollback. Keep command output capture bounded and fail closed at cap+1.

The synthetic evidence flow for these criteria is implemented in Workstream F. Keep it separate from the completed issue-text refinement. Do not hand-edit canonical JSONL; live promotion waits for the exact-source behavioral suite and independent review.

### Workstream F design: explicit creation and source continuation

**Status: Synthetic-only creation and evidence capture implemented; exact current-source continuation approved and under behavioral verification.** No live issue, live Beads state, or repository canonical JSONL has been changed in this slice. Temporary Beads writes and promotion tests use disposable stores/files. The current source SHA-256 is `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`; its approved parent snapshot is `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`. Approval binds both exact digests to Beads 1.3.1 binary SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653` and the unchanged five pairs. No third digest, alternate binary/version, or additional pair is accepted; plan content and caller-provided digests cannot authorize another lineage.

The creation contract is a separate explicit `creates` set in the evidence plan, distinct from `updates`. Each create entry carries one complete proposed record and its exact dependency edges. The evidence pipeline proves that B0 has exactly the source IDs; B1 has exactly source IDs plus planned create IDs; and the candidate contains exactly those records and edges while existing records and edges remain unchanged. It rejects empty/incomplete records, unsupported fields, malformed IDs, duplicate planned IDs, collisions with canonical IDs, missing targets, duplicate edges, extra records, unplanned field/edge changes, and re-key or unplanned deletion. The synthetic test child is `dap-jo9.99`, a direct child of `dap-jo9`; it is test data only and does not reserve an ID for future work.

The selected source-continuation option is the exact snapshot-specific authorization below. Live creation/promotion remains gated on tests and review:

| Option | Required proof | Boundary / decision |
|---|---|---|
| **A. Source-specific lineage authorization** | The approval names the exact current source digest, parent source digest, pinned binary/version, and unchanged five issue/field pairs. Regression tests re-import and export the exact source in disposable stores, retain source null tokens in the candidate, and validate a second-store round trip. | **Selected and approved for these two snapshots only.** This does not authorize any additional source digest; a further continuation requires separate approval. |
| **B. Preserve null presence before or during import** | Test a supported pinned Beads import/export path on a disposable source copy; any sidecar would need exact source and path binding and separate review. | Not selected. No generic null-preserving or absent-equals-null behavior is added. |

The exact continuation is approved, and the current-source path is under test. Keep stale-source rejection and exact bundle/candidate digest checks. A stale bundle, changed canonical source, or newly unauthorized lineage must leave canonical bytes untouched. No live Beads store is used in design or test development; live promotion remains gated on the behavioral suite and independent review.

The synthetic reconciler accepts one complete canonical issue row and requires the B1 addition to match its planned semantic fields and exact parent edge; an empty labels array may be omitted by the pinned importer, while JSON `null` is still distinct and rejected. The CLI imports the planned row only into its disposable evidence store, then derives the candidate and verifies it through a second Beads store. Tests cover exact creation, record counts, duplicate/colliding IDs, changed fields and edges, CLI evidence provenance and promotion against a disposable synthetic canonical file, and the actual pinned import/export round trip from the exact approved current source. No live tracker record or canonical export has been promoted. Any further source, binary, version, or pair-scope change requires separate written approval naming its exact scope.

## Workstreams and dependencies

Workstreams may proceed in parallel only after their corresponding Beads issue exists or an owner explicitly accepts a different safe tracker path. No code stream should claim issue completion based on this plan.

| Workstream | Priority and tracker | Can parallelize after | Scope |
|---|---|---|---|
| A. Adapter timeout contract | P1 execution gate; existing `.5.7` is P2 | Safe tracker update for `.5.7` | Split true timeout from early leader exit; no production code change unless a focused failing test demonstrates one. |
| B. Parser and loader assurance | P2; intended `.2.1.4` | Safe tracker update that removes parser acceptance from `.5.6` | Root mutation-killing parser tests plus consumer-specific strict JSON contracts and useful fuzz oracle properties. |
| C. Bootstrap pins and identity | P2; round-two Task 3, unassigned | Issue creation/reconciliation prerequisite | Bootstrap module/version/archive provenance controls. |
| D. CI policy and dependency updates | `.5.6` is an existing Beads P1 for current module gates; Tasks 4/5 remain proposed P2 with no IDs | Distinguish existing `.5.6` acceptance from the extra controls and safely assign Tasks 4/5 | Workflow fail-open/skip/shell mutation tests, adapter Dependabot, and conditional bootstrap module coverage. |
| E. Fixture transaction and bootstrap CLI | P2; round-two Task 7, unassigned | Issue creation/reconciliation prerequisite; preserve fixture gate | Failure-injected paired statement/invoice publication and bootstrap `run()` CLI boundaries. |
| F. Reconciler assurance and safe tracker mutation | P1 execution prerequisite; no Beads issue created. `dap-jo9.18` remains an unreserved candidate ID from a disposable probe. | Use a direct child of `dap-jo9` to avoid unrelated product blockers; exact source continuation is approved, while live promotion awaits behavioral tests and independent review | Existing negative tests plus the scoped issue-creation/source-freshness validation above. |
| G. Optional fuzz CI / helper assessment | P3 decision only; Tasks 6/8 unassigned optional | A/B establish useful fuzz targets and behavioral evidence | Bounded scheduled/manual fuzzing only if useful; helper consolidation only after lifecycle tests prove need. |
| H. Conditional executable-path race | Optional hardening; no active tracker item | Revisit only if an untrusted concurrent writer to the executable or its parent is in scope | No replacement race was demonstrated; write a deterministic replacement test before any code change. |

**Critical path:** A and B are implemented in the candidate and passed their required tests → independent plan/code review → guarded publication → fresh protected CI. C–E remain unassigned until a reviewed issue-creation capability exists. F precedes adding new issue records. After this candidate becomes main, do not reconcile Beads again until the five-null source-lineage continuation is reviewed. G is not a release gate unless a separate review justifies it.

## Test-first implementation plans

### A. Split the adapter timeout and early-exit tests

**Baseline test state:** the focused adapter timeout test expected timeout but received malformed XML around 0.21 s. `SPAWN_PIPE_HOLDER` terminated its leader immediately after child readiness. The production adapter waits on the unreaped leader, sends a group signal, waits its 200 ms SIGTERM grace, then reaps; observed duration aligned with early leader exit. This was a test setup mismatch, not proof a timeout path was broken; the candidate separates both cases and passes focused/full tests in normal and race modes.

**Red → green → refactor:**

1. Keep the current implementation unchanged. Add one helper mode where the leader remains alive beyond the 300 ms deadline while a descendant holds the inherited pipe. Assert timeout classification, prompt process-group termination/reaping, no leaked temp files, and the documented time bound. Confirm it fails if timeout handling is removed.
2. Add a distinct early-leader-exit case using the existing empty-output `SPAWN_PIPE_HOLDER` fixture: the leader exits after readiness and before the deadline while the descendant holds the pipe. Assert the malformed-report parsing error (and explicitly assert it is not a timeout), assert completion before the configured deadline, and verify descendant/temp cleanup. Keep the valid-output/descendant control separate.
3. Add positive valid-report control and injected termination control as applicable. Keep bounded wait loops, avoid arbitrary fixed sleeps, and preserve the `WNOWAIT`/group-signal-before-reap behavior.
4. Run each case repeatedly under `-race`, then the entire adapter module test/vet/tidy-diff checks. Change production code only if the true-timeout regression remains red after helper timing is corrected; demonstrate the production cause first.

### B. Parser cardinality and decoder-specific JSON contracts

**Round-three red evidence:** removing exactly-one `validationReports` validation or the count-equals-one predicate survives root and nested focused tests. Mutants for `failedJobs`, leading text, and trailing text survive root tests but nested adapter tests reject them. Do not add duplicate adapter tests for cases its contract already rejects.

1. In `internal/pdfua/report_test.go`, add direct pass and finding report controls with multiple `<validationReports>` elements; add valid-noncompliant and compliant controls for the summary count other than exactly one. Run each focused test against the unmodified parser.
2. Remove one predicate at a time in disposable source and prove the corresponding root test fails for the required report classes; restore byte-identical source and rerun root+nested focused tests. No mutation is retained.
3. Do not repeat corpus or `LoadCatalog` tests: the round-two current-main update records both as complete, including exact-spelling, duplicate/fold, trailing-value, and canonical-fixture controls. Do not repeat the `jsonstrict` helper tests already merged in PR #83; existing `keys_test.go` covers nested exact duplicates, Unicode folding, arrays/scalars, malformed JSON, and the helper's one-value behavior.
4. For remaining consumer-owned boundaries, add only current-source gaps in the bootstrap manifest, `author_v2`/observer decoding, and `internal/traceability.LoadIssues`. Test exact duplicates and case aliases where those formats have a closed key contract, trailing values where one JSON value is required, and malformed shapes. Preserve forward-compatible ignored-field behavior in the Beads JSONL projection. Compare literal inputs to `encoding/json` behavior where compatibility matters; do not infer helper-level EOF guarantees.
5. Inspect the existing `FuzzParseReport` oracle after PR #80; retain its completed five-counter assertions and add a cardinality property only if it is genuinely absent and independent of parser assumptions. Do not recreate counter fuzz coverage or require generated corpus volume or a coverage percentage.

### C. Bootstrap module/version and archive provenance (round-two Task 3)

1. Write a regression for each documented malformed module/version coordinate currently accepted by substring matching; capture exact current outcome.
2. Define supported Go module path and parsed-version forms before implementation. Cover exact pins, `1.9` vs `1.9.0`, prerelease vs release, allowed prerelease pins, pseudo-versions, `+incompatible`, alternate module paths, and malicious prefix/suffix controls. Validate runtime `version_pattern` against complete output.
3. Reject archive `VersionSource` unless an explicit installed-binary digest contract is separately adopted. Add a positive raw-binary version-source control where installed bytes match the pinned release-asset digest and no version command runs. Preserve the NeuroPlasticity digest-only path and do not invoke its unsafe version flag.
4. Refactor only after the tests pass. No downloads or installed tools are required for unit tests.

### D. Required CI policy and Dependabot (round-two Tasks 4/5)

1. Add red policy negatives for required workflow job/step skipped by `if`, `continue-on-error`, ignored failures, and effective shell overrides that can suppress required checks. Retain safe positives for current matrix/scanner conditions, inherited shell behavior, and explicit `bash` with `set -euo pipefail`. Do not blanket-ban valid conditions or shell forms.
2. Verify scanner-specific negative controls already cover scanner conditions and `continue-on-error`; extend only missing project-check paths. Policy mutation tests must show the required step's name, working directory, test/vet/tidy command and failure behavior are all protected.
3. Add `ci/pdfua2/adapter` Dependabot Go coverage. For `cmd/cloud-tools-bootstrap`, add its module entry or a discovery guard that verifies today's empty external dependency set and requires update coverage if that changes. Keep root, scanner, vulnerability, and license check scopes distinct.
4. Validate YAML and run policy tests. `.5.6` remains the three-module workflow/policy-presence issue; Tasks 4/5 need separate reviewed issue ownership and acceptance.

### E. Financial fixture pair rollback and bootstrap CLI (round-two Task 7)

1. Keep original synthetic statement and invoice fixtures byte-identical as preservation baselines. Give both families equal priority and use authored bytes only.
2. Write failure-injection tests for staging, second-file publication, rollback failure, read-back mismatch, and pre-existing-target preservation using narrow injectable filesystem/process seams. Assert no mixed or partial pair is reported as success; retain exact known-good bytes whenever rollback promises it.
3. Add bootstrap `run()` process-level tests for usage error, malformed manifest, verify-only status, and operational failure without network/download access.
4. Keep runtime dependencies and generic storage abstraction out of scope. Do not change human semantic or AT acceptance gates. Tests cannot promote a no-op or validator pass into accessibility approval.

### F. Reconciler trust and atomicity

Use the numbered red/green/source-freshness/failure-injection sequence under “Minimal test-first prerequisite” above. Add cases for duplicate/missing/unexpected ZIP entries; altered manifest, command provenance, snapshots and candidate including recomputed outer digest; every entry/canonical/plan/binary/output cap and cap+1; lock refusal and cleanup; source races; pre-rename write/sync/close/rename failures; and post-rename directory-sync failure. Assert behavior/state invariants and exact source preservation, not arbitrary coverage thresholds. Never copy the current canonical export by hand or expand the five-pair null exception to accommodate a new source digest.

### G. Optional fuzz CI and helper consolidation (round-two Tasks 6/8)

After A/B establish a useful fuzz oracle, decide whether short scheduled/manual fuzz jobs add unique defect detection; pin module and command, retain crash artifacts, and fail on crashes. Consider process helper consolidation only after tests prove common lifecycle invariants (leader exit ordering, bounded reaping, inherited-pipe draining, descendants, lock lifetime) and a minimal shared design respects module/API boundaries. Neither absence of scheduled fuzz CI nor helper count is itself a defect.

## Verification gates and external-check boundaries

**Completed before final independent plan review:** on local PR candidate `d68d13991d26d69162b1dc935abac01ff2a22ee1`, the adapter's focused parser/process tests and full suite passed both without `-race` and with `-race` (Go 1.27.2, `PDFUA_REAL_CLI_JAR=`). The companion review report records commands and timings. The four checks were rerun after the test-only timeout contract fix and after combining both test streams.

After each stream, record exact commit and commands. For changed Go code:

- Root module: `gofmt`, `go mod tidy -diff`, `go vet ./...`, `go test -race ./... -count=1`.
- `cmd/cloud-tools-bootstrap`: its own gofmt/tidy-diff/vet/race suite.
- `ci/pdfua2/adapter`: its own gofmt/tidy-diff/vet/race suite; do not rely on root module commands to traverse nested `go.mod` boundaries.
- Run focused test-first regressions and disposable one-predicate mutations before broader suites; verify pristine source restoration.
- Run traceability and shell/source-publisher checks when their files or workflows are touched.
- Keep vulnerability scans and Java/veraPDF validator jobs in isolated hosted CI only. Never run Java/veraPDF locally. A clean local gate is not a substitute for all required exact-head hosted checks.
- Preserve `ua2` as the validator invocation. Continue requiring candidate-bound machine evidence plus independent human semantic and assistive-technology review for any output claim.

Plan completion requires each issue to have an exact Beads ID, agreed acceptance criteria, TDD red/green evidence, and fresh local/hosted evidence at the implementation head. Until the reconciler path and issue ownership are safely resolved, report work as plan-only/unassigned. The PR remains draft; do not merge from this plan.

## Finding-to-action traceability

| Finding / source | Action | Priority / Beads mapping | Evidence boundary |
|---|---|---|---|
| Round-three XML parser mutants for `validationReports` count/cardinality survive both suites | Added root parser tests and one-at-a-time mutation proof | P2; `.2.1.4` acceptance refined in candidate | Regression-test gap; no production bypass shown |
| `failedJobs`, leading text, trailing text mutants survive root but nested adapter catches them | Preserve nested contract; avoid duplicate assertions | Existing `.2.1.4`/`.5.6` refinement; do not invent new defect | Nested tests already reject those mutations |
| Adapter timeout test helper exits leader before 300 ms deadline | Split true timeout and early-exit contracts | Test setup fix implemented; `.5.7` remains open | Reproducible test setup mismatch, not a proven production timeout bug |
| Earlier `.5.6` mixed parser and module-gate acceptance; `.2.1.4` lacked direct cardinality acceptance | Reconciled complete issue fields from the authorized source | Existing `.2.1.4`, `.5.6` | Candidate generated and promoted from exact approved source; five null tokens preserved |
| Stale plan bundle rejected against changed source | Preserve exact source/binary/pair allowlist; no migration needed on current base | Reconciler child for broader assurance remains unassigned | Stale bundle failed closed; final candidate was regenerated from authorized source |
| Need new tracker issues for Tasks 3/4/5/7 and reconciler prerequisite | Add scoped reconciler issue-creation feature with tests before tracker edits | Plan-only / unassigned | Current reconciler supports existing-field updates and one bounded planned creation through disposable evidence stores; additional workstreams still lack active issue IDs |
| Round-two strict JSON tasks | Exercise only remaining decoder-specific contracts; preserve already-complete corpus/catalog/helper work | `.2.1.4` plus existing consumer-owned work | PR #82 corpus, PR #76 `LoadCatalog`, and PR #83 `jsonstrict` helper contracts are complete; no duplication |
| Bootstrap module/version/archive identity Task 3 | Coordinate/version/provenance red-green suite | P2, plan-only / unassigned | Source-verified assertions require runtime regression proof |
| CI policy skip/fail-open/shell Task 4 | Add narrow policy mutants and controls | P2, plan-only / unassigned; `.5.6` only owns gate presence | Existing scanner controls must not be duplicated or overclaimed |
| Dependabot Task 5 | Add adapter entry and bootstrap empty-dependency guard | P2, plan-only / unassigned | Existing vulnerability/license CI are not update automation |
| Fixture rollback/bootstrap CLI Task 7 | Failure-injected pair tests and subprocess controls | P2, plan-only / unassigned | Synthetic statement/invoice fixtures remain unchanged/equal priority |
| Fuzz CI Task 6 and helper consolidation Task 8 | Decide only after unique benefit is established | P3, optional | Not current defects; no coverage quota |
| Conditional `runCommand` executable-path replacement race | Defer unless the threat model includes a concurrent untrusted writer; if included, add deterministic replacement control first | Optional; no tracker item | Not demonstrated in this review |

## Review and revision gates

1. **Completed:** reconcile the three existing-issue fields through the supported promoter from the authorized source; no direct canonical edit or issue creation by `bd import`.
2. **Independent plan review completed before publication:** a separate read-only reviewer found three issues in the first draft: it repeated completed corpus/catalog/helper tests, omitted a disposition for the conditional executable-path race, and left the early-exit expectation ambiguous. This revision removes the duplicate test work, records the race as deferred/no tracker item, and requires the empty-report fixture's malformed-report parse error before the deadline. The independent reviewer rechecked the corrected revision and reported no remaining findings.
3. Keep the implementation plan distinct from actual issue state. When safe tracker issue-creation support is implemented and reviewed, create only reviewed new issue records, then refresh the plan snapshot.
4. For each implemented task, obtain independent exact-head code review and fresh required hosted CI. No merge is authorized here; PR #84 remains draft.
