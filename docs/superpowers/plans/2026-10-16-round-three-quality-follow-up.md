# Round-three adversarial Go quality follow-up

**Status:** Proposed execution plan; no implementation is authorized by this document. The PR remains draft. No issue-creation operation or additional tracker update has been promoted. Beads work is blocked until the source-lineage/reconciliation path below is approved and supported.

**Date and evidence cutoff:** 2026-10-16. Round-three review baseline: `6ae3f03bf8272007e94b1bb9f9e2317f3d5c5fac`. PR #84 documentation head at planning time: `fff070ee8a5296157296d2318cb4cd3a73f418f3`, based on current main `6ae3f03`. The updated plan and report are not implementation results.

**Repository guidance used:** [`dap-test-driven-development`](../../../.agents/skills/dap-test-driven-development/SKILL.md), [`dap-systematic-debugging`](../../../.agents/skills/dap-systematic-debugging/SKILL.md), [`dap-verification-before-completion`](../../../.agents/skills/dap-verification-before-completion/SKILL.md), and [`dap-plan-review-integrator`](../../../.agents/skills/dap-plan-review-integrator/SKILL.md). The first three are local adaptations of imported Superpowers v7.0.0 skills; the review-to-plan integrator is the repository's local roundtable adaptation. The imported-skill inventory/provenance was checked. It does not include the upstream Superpowers `writing-plans` skill, so that skill is not claimed as used; this plan follows the existing repository plan format and the actual local TDD/debugging/review guidance.

## Decision summary

Round three found **test assurance gaps, not a demonstrated current parser bypass**. Two one-predicate mutations—allowing multiple `validationReports` elements and accepting summary content whose count is not exactly one—survived both root and nested focused suites. Other mutations for `failedJobs`, leading text, and trailing text survived root tests but were rejected by the nested adapter. The shared `jsonstrict` Unicode duplicate-key regression from PR #83 is already merged; do not duplicate it.

The nested adapter's full timeout test has a reproducible **test setup mismatch**: its `SPAWN_PIPE_HOLDER` leader exits after descendant readiness, before the configured 300 ms deadline. The adapter returns the early-exit malformed-XML result after its process-group cleanup grace (about 0.21 s). This does not demonstrate a production timeout defect. Split the scenarios so the timeout leader stays alive past the deadline and an early-exit leader is separately expected to fail before the deadline.

Stale claims in existing issue text need to be reconciled against present work and source lineage. In particular, the current exported `.5.6` acceptance criteria still include parser tests; the current `.2.1.4` criteria do not yet contain the two direct `ParseReport` cardinality cases. **Intended refinement:** put parser assurance under `.2.1.4` and reserve `.5.6` for root/nested CI and policy gate coverage. This intended reassignment has not been promoted to current Beads state.

## Severity and execution priority

Priority below is an execution recommendation; it is not a claim that the issue has that priority in Beads. Current production risk, test gaps, stale claims, and optional hardening are separated:

| Class | Evidence | Execution priority |
|---|---|---|
| Current required-suite test failure | Adapter timeout test's helper exits before its own deadline and gets an early-exit error; exact reproduction and cause are documented in the round-three report. | **P1 sequencing gate:** split the cases first so the required nested suite has a truthful deterministic contract. Beads `.5.7` remains recorded P2 until a safe tracker update is accepted. |
| Regression assurance gaps | Two XML parser predicate mutations survive root and nested focused tests. The running implementation still contains the predicates; this is not proof of an input bypass. | **P2:** add root-owned tests and mutation controls under `.2.1.4`. |
| Existing boundary contract gaps | Remaining round-two decoder-specific boundaries are bootstrap manifest, `author_v2`/observer decoding, and `LoadIssues`. Corpus/catalog work and PR #83 helper tests are already complete. | **P2:** test only remaining consumer contracts; do not impose blanket strict decoding. |
| CI/policy/dependency/transaction assurance gaps | Round-two Tasks 3, 4, 5, and 7 remain open and currently lack active Beads IDs verified for their specific controls. | **P2:** create/reconcile valid issue ownership before implementation; plan-only until then. |
| Stale documentation/tracker scope | `.5.6` currently mixes CI ownership with parser tests; `.2.1.4` lacks the direct cardinality requirements. | **P2:** reconcile the two issue texts as one reviewed complete candidate from an authorized source. |
| Optional hardening | Scheduled/manual fuzz CI and helper consolidation are not current defects. | **P3 / decision only:** do not implement without a demonstrated added safety benefit. |
| Conditional executable-path race | `runCommand` verifies a Beads executable path before and after invocation but does not execute an immutable copy; this review did not reproduce replacement. | **Defer / no tracker item:** revisit only if an untrusted concurrent writer can replace the executable or its parent-directory entry. |

No finding establishes a financial-output defect, PDF/UA nonconformance, or active security bypass. Preserve the original statement and invoice fixtures, equal family priority, human semantic review, and assistive-technology gates. Machine tests and CI are not substitutes for those gates.

## Tracker and source-lineage state

### Existing issue ownership: current versus intended

| Issue | Current canonical export in PR #84 | Intended refined assignment | State |
|---|---|---|---|
| `dap-jo9.5.6` (Beads P1) | Its current acceptance criteria include root parser assertions and the root/bootstrap/adapter Go gates. | Keep only CI and policy coverage for all three current modules; move parser predicates to `.2.1.4`. Keep future module discovery optional. | Open. The intended complete update was generated only against the original approved source and was not promoted against current canonical bytes. |
| `dap-jo9.2.1.4` (Beads P2) | Existing criteria cover corpus/catalog key strictness and `FuzzParseReport`, but not direct root tests for multiple `validationReports` elements or count other than one. | Own decoder-specific loader requirements plus direct `ParseReport` regressions for both surviving mutations; preserve valid compliant and finding controls. | Open. The intended update was not promoted against current canonical bytes. |
| `dap-jo9.5.7` (Beads P2) | Current issue text already corrects several stale documentation claims and records process behavior; its adapter timeout test still has the setup mismatch. | Keep the corrected inventory and process-group invariants, distinguish early leader exit from actual timeout, and test both cleanup paths. | Open. Intended update not promoted against current canonical bytes. |
| Round-two Tasks 3, 4, 5, 7 | No active issue ID was verified as owning these specific work items. `.5.6` owns required module gate presence, not policy fail-open controls. | Bootstrap pin identity; workflow-policy skip/fail-open/shell controls; adapter/bootstrap Dependabot coverage; financial fixture rollback and bootstrap CLI tests. | **Plan-only / unassigned.** Do not describe as fully tracker-mapped or completed. |
| Reconciler trust/atomicity negative tests and safe issue-creation/source migration | No active issue ID was created. The existing proposed child of `.5` has no ID. | Add the smallest test-first prerequisite described below, then create/refine separate issue records through the supported reconciler. | **Plan-only / unassigned.** The reconciler's `create` command creates an evidence bundle for updates; it does not create a Beads issue. |

### Exact Beads reconciliation evidence

The pinned importer is Beads 1.3.1, executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. The explicit-null exception is authorized only for source SHA-256 `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`, and exactly these pairs: `dap-jo9.5.3.dependencies`, `dap-jo9.1.6.10.closed_at`, `dap-jo9.5.6.closed_at`, `dap-jo9.2.1.4.closed_at`, and `dap-jo9.5.7.closed_at`.

Current canonical `.beads/issues.jsonl` SHA-256 is `785e8713c1d4175f6371898410a3f22db2f1b95e4a3e70a7185e65264bfd7ec0`; the original approved source is the export at `6ae3f03` with SHA-256 `7ce7e649…1125890`. A disposable import/export with the pinned binary confirmed the source has explicit `null` and B0 omits the field for **all five** listed pairs. The previous attempt on the changed current source stopped at `dap-jo9.5.7.closed_at` with `source-to-baseline validation failed before updates ... unplanned semantic delta in fields [closed_at]`. That is the first reported pair, not the complete set of losses: the changed source digest invalidates the full allowlist, so all five import losses are unapproved for that source.

The complete three-issue update plan (`.2.1.4`, `.5.6`, `.5.7`) was regenerated from the original approved source using the unchanged pinned exception. `create` succeeded with candidate SHA-256 `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de` and evidence bundle SHA-256 `7123c9925acfa0579cb75f5637425a6d2a459a0376c160f955bdad1083db4acc`. `promote` succeeded on a disposable copy of that exact original source. The same bundle failed against a disposable copy of current source with `promote: canonical export is stale; create a new evidence bundle`; its before/after digest remained `785e8713…bfd7ec0`. No live Dolt store or tracked canonical JSONL was changed by these experiments.

**Conclusion:** recomputing from the original approved source can preserve all three intended existing-issue refinements and the five null tokens under the current exception. That candidate cannot safely be applied to the current PR source: the promoter correctly rejects stale source bytes. Do not re-key or broaden the null-loss allowlist. A separate, reviewed source-lineage migration path must validate current-to-approved-source changes, preserve each exact null token and field pair, and reject all other semantic or graph deltas before promotion.

### Minimal test-first prerequisite for issue creation and source migration

The current reconciler plan accepts only `updates` to IDs already present in canonical source. Its `create` subcommand means “create a reviewable evidence bundle,” not “create a Beads issue.” While Beads `bd import` itself supports upsert/new records, bypassing the reconciler would evade its graph, provenance, snapshot, and atomic-promotion guarantees.

Before creating or assigning new Beads records, add a separately reviewed reconciler feature, with the first tests written against the current fail-closed interface:

1. **Red: issue creation contract.** Add plan fixtures that request exactly one new issue and an exact parent dependency; prove the current schema rejects unknown create fields. Specify required canonical fields, valid IDs, planned dependency shape, and how source/baseline/candidate issue counts differ. Add duplicate-ID and collision-with-existing-ID negatives.
2. **Green: scoped creation.** Implement only explicit planned records; require exact planned additions, forbid modifications/deletions to unlisted records or edges, validate full graph invariants, and verify round-trip export in the second disposable store. Positive controls preserve all existing issue bytes and dependency records.
3. **Red: source-lineage migration.** Show a bundle from source `7ce7e649…` cannot promote over `785e8713…`; show a changed or missing null token, extra null pair, altered dependency, unplanned issue edit, or stale current digest cannot pass. No changed digest may inherit the current exception merely by configuration.
4. **Green: explicit migration evidence.** If product/security review authorizes a lineage migration design, bind both exact source snapshots, pinned Beads binary, all five pair values/spans, exact approved edits, and current-source delta to a separately versioned migration authorization. Re-derive candidate and round-trip data independently; require exclusive writer coordination and atomic promotion. Until that design and tests pass, do not update canonical JSONL or the allowlist.
5. **Refactor and failure injection.** Inject malformed bundle, archive digest mismatch, recomputed outer-digest tampering, missing/duplicate/unexpected ZIP entries, altered manifest/provenance/snapshot/candidate, source race, lock refusal, size max/max+1, and pre-rename write/sync/close/rename failures. Assert unchanged source, temp cleanup, and lock release before rename; test post-rename directory-sync failure without claiming rollback. Keep command output capture bounded and fail closed at cap+1.

This is a proposed technical prerequisite, not a completed feature or a Beads issue. Keep issue-creation support separate from the issue-text update refinement. Do not hand-edit canonical JSONL.

## Workstreams and dependencies

Workstreams may proceed in parallel only after their corresponding Beads issue exists or an owner explicitly accepts a different safe tracker path. No code stream should claim issue completion based on this plan.

| Workstream | Priority and tracker | Can parallelize after | Scope |
|---|---|---|---|
| A. Adapter timeout contract | P1 execution gate; existing `.5.7` is P2 | Safe tracker update for `.5.7` | Split true timeout from early leader exit; no production code change unless a focused failing test demonstrates one. |
| B. Parser and loader assurance | P2; intended `.2.1.4` | Safe tracker update that removes parser acceptance from `.5.6` | Root mutation-killing parser tests plus consumer-specific strict JSON contracts and useful fuzz oracle properties. |
| C. Bootstrap pins and identity | P2; round-two Task 3, unassigned | Issue creation/reconciliation prerequisite | Bootstrap module/version/archive provenance controls. |
| D. CI policy and dependency updates | `.5.6` is an existing Beads P1 for current module gates; Tasks 4/5 remain proposed P2 with no IDs | Distinguish existing `.5.6` acceptance from the extra controls and safely assign Tasks 4/5 | Workflow fail-open/skip/shell mutation tests, adapter Dependabot, and conditional bootstrap module coverage. |
| E. Fixture transaction and bootstrap CLI | P2; round-two Task 7, unassigned | Issue creation/reconciliation prerequisite; preserve fixture gate | Failure-injected paired statement/invoice publication and bootstrap `run()` CLI boundaries. |
| F. Reconciler assurance and safe tracker mutation | P1 execution prerequisite; proposed child of `.5`, no Beads ID/priority assigned | Product review of source-lineage design before exception or migration behavior changes | Existing negative tests plus the scoped issue-creation/source-migration prerequisite above. |
| G. Optional fuzz CI / helper assessment | P3 decision only; Tasks 6/8 unassigned optional | A/B establish useful fuzz targets and behavioral evidence | Bounded scheduled/manual fuzzing only if useful; helper consolidation only after lifecycle tests prove need. |
| H. Conditional executable-path race | Optional hardening; no active tracker item | Revisit only if an untrusted concurrent writer to the executable or its parent is in scope | No replacement race was demonstrated; write a deterministic replacement test before any code change. |

**Critical path:** resolve safe tracker ownership and source lineage → A timeout test split to restore a truthful full-suite contract → B parser/loader regression suite → required local module gates and independent review → fresh protected CI. C–E are independent of parser implementation after ownership is established. F precedes adding new issue records or changing canonical exports. G is not a release gate unless a separate review justifies it.

## Test-first implementation plans

### A. Split the adapter timeout and early-exit tests

**Known test state:** a focused adapter timeout test expected timeout but received malformed XML around 0.21 s. `SPAWN_PIPE_HOLDER` terminates its leader immediately after child readiness. The production adapter waits on the unreaped leader, sends a group signal, waits its 200 ms SIGTERM grace, then reaps; observed duration aligns with early leader exit. This is a test setup diagnosis, not proof a timeout path is broken.

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

Use the numbered red/green/migration/failure-injection sequence under “Minimal test-first prerequisite” above. Add cases for duplicate/missing/unexpected ZIP entries; altered manifest, command provenance, snapshots and candidate including recomputed outer digest; every entry/canonical/plan/binary/output cap and cap+1; lock refusal and cleanup; source races; pre-rename write/sync/close/rename failures; and post-rename directory-sync failure. Assert behavior/state invariants and exact source preservation, not arbitrary coverage thresholds. Never copy the current canonical export by hand or expand the five-pair null exception to accommodate a new source digest.

### G. Optional fuzz CI and helper consolidation (round-two Tasks 6/8)

After A/B establish a useful fuzz oracle, decide whether short scheduled/manual fuzz jobs add unique defect detection; pin module and command, retain crash artifacts, and fail on crashes. Consider process helper consolidation only after tests prove common lifecycle invariants (leader exit ordering, bounded reaping, inherited-pipe draining, descendants, lock lifetime) and a minimal shared design respects module/API boundaries. Neither absence of scheduled fuzz CI nor helper count is itself a defect.

## Verification gates and external-check boundaries

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
| Round-three XML parser mutants for `validationReports` count/cardinality survive both suites | Add root parser tests and one-at-a-time mutation proof | P2; intended `.2.1.4`; current `.5.6` acceptance must be corrected | Regression-test gap; no production bypass shown |
| `failedJobs`, leading text, trailing text mutants survive root but nested adapter catches them | Preserve nested contract; avoid duplicate assertions | Existing `.2.1.4`/`.5.6` refinement; do not invent new defect | Nested tests already reject those mutations |
| Adapter timeout test helper exits leader before 300 ms deadline | Split true timeout and early-exit contracts | Execution gate P1; existing `.5.7` Beads priority remains P2 | Reproducible test setup mismatch, not a proven production timeout bug |
| Current `.5.6` mixes parser and module-gate acceptance; `.2.1.4` lacks direct cardinality acceptance | Reconcile full issue fields from source lineage that preserves five nulls | Existing `.2.1.4`, `.5.6`; not promoted against current source | Complete candidate from original source is verified, current source promotion rejected stale |
| Plan update aborts on `.5.7.closed_at` | Preserve exact source/binary/pair allowlist; design separate migration tests | Reconciler child under `.5` is proposed but unassigned | All five null pairs lost on import from changed digest; only `.5.7.closed_at` was first reported |
| Need new tracker issues for Tasks 3/4/5/7 and reconciler prerequisite | Add scoped reconciler issue-creation feature with tests before tracker edits | Plan-only / unassigned | Current reconciler handles field updates only; `create` makes evidence bundles |
| Round-two strict JSON tasks | Exercise only remaining decoder-specific contracts; preserve already-complete corpus/catalog/helper work | `.2.1.4` plus existing consumer-owned work | PR #82 corpus, PR #76 `LoadCatalog`, and PR #83 `jsonstrict` helper contracts are complete; no duplication |
| Bootstrap module/version/archive identity Task 3 | Coordinate/version/provenance red-green suite | P2, plan-only / unassigned | Source-verified assertions require runtime regression proof |
| CI policy skip/fail-open/shell Task 4 | Add narrow policy mutants and controls | P2, plan-only / unassigned; `.5.6` only owns gate presence | Existing scanner controls must not be duplicated or overclaimed |
| Dependabot Task 5 | Add adapter entry and bootstrap empty-dependency guard | P2, plan-only / unassigned | Existing vulnerability/license CI are not update automation |
| Fixture rollback/bootstrap CLI Task 7 | Failure-injected pair tests and subprocess controls | P2, plan-only / unassigned | Synthetic statement/invoice fixtures remain unchanged/equal priority |
| Fuzz CI Task 6 and helper consolidation Task 8 | Decide only after unique benefit is established | P3, optional | Not current defects; no coverage quota |
| Conditional `runCommand` executable-path replacement race | Defer unless the threat model includes a concurrent untrusted writer; if included, add deterministic replacement control first | Optional; no tracker item | Not demonstrated in this review |

## Review and revision gates

1. Reconcile current versus intended Beads fields through a supported source-lineage path; no direct canonical edit or issue creation by `bd import`.
2. **Independent plan review completed before publication:** a separate read-only reviewer found three issues in the first draft: it repeated completed corpus/catalog/helper tests, omitted a disposition for the conditional executable-path race, and left the early-exit expectation ambiguous. This revision removes the duplicate test work, records the race as deferred/no tracker item, and requires the empty-report fixture's malformed-report parse error before the deadline. The independent reviewer rechecked the corrected revision and reported no remaining findings.
3. Keep the implementation plan distinct from actual issue state. When safe tracker support is available, promote scoped issue updates and create only reviewed new issue records, then refresh the plan snapshot.
4. For each implemented task, obtain independent exact-head code review and fresh required hosted CI. No merge is authorized here; PR #84 remains draft.
