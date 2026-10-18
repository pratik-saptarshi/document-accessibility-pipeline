# PR #133: bounded mapping evidence and acceptance plan

**Status:** bounded, test-only evidence is implemented. The owner accepted the bounded AC2f interpretation on 2026-10-18 at 10:52:44 UTC. At the implementation base, the original Beads issue was open pending its separate closure audit and a supported guarded transition. PR #143 later changed the tracker status on `main`; this plan and the evidence supplement do not assess or treat that separate transition as acceptance evidence. Prepared 2026-10-18 UTC. This preserves the substantive reviewed design transferred from Library; it does not claim byte identity with Library file `libfile_3457ca7bb07c819188a8bf8f26ad7e82` (`pr133-mapping-acceptance-evidence-plan-2026-10-18.md`, reported digest `5549e4c6360e4f872f94939da072fd9180483d9dad2f156f3bb1020870775569`). Execution results and the accepted bounded clause interpretation are in [the dated evidence supplement](../../research/financial-table-mapping-evidence-2026-10-18.md).

## Goal and boundaries

Add fixture-backed evidence for the then-open `dap-jo9.1.6.7` discovery task without treating a literal-input gate test as mapping accuracy. Keep evidence implementation, its acceptance, and any tracker transition separate. The task is limited to synthetic statement and invoice tables. No automatic mapping, inference implementation, unconfirmed application, production writer, general-document support, semantic/AT authority, conformance claim, or tracker edit is authorized.

The execution refresh used `origin/main` `6d42691d2c95ac0b369408f527d5994fb187d5d3`; the earlier plan/source inspection used `7ba21bcf49edde780eccb0e98cdfe38fba0d8c42`. At that refresh, main and the exported `.beads/issues.jsonl` record showed the issue open. PR #143 later changed tracker status in commit `e219913f3cc25774e68eb597b33f37aa3da68a45`; this branch did not edit the tracker and this plan does not assess that transition. Work remains on branch `test/financialfixtures-mapping-evidence`. `bd` was unavailable during the initial inspection.

At plan preparation, PR #133's latest observed head was `472c1c88be7507d85b75664657a8ae99181aa3f8`, and PR #141's latest observed head was `c64f4d96ff`. A live GitHub REST read at `2026-10-18 11:24:55 UTC` of [`GET /repos/pratik-saptarshi/document-accessibility-pipeline/pulls/133`](https://api.github.com/repos/pratik-saptarshi/document-accessibility-pipeline/pulls/133) and [`GET /repos/pratik-saptarshi/document-accessibility-pipeline/pulls/141`](https://api.github.com/repos/pratik-saptarshi/document-accessibility-pipeline/pulls/141) showed both open. The response had `auto_merge: null` for #133, consistent with the owner's auto-merge stop, and an enabled auto-merge record for #141. Neither is an implementation baseline. Do not change either PR or canonical tracker state. PR #132's separately owned paths remain out of scope: `docs/research/financial-table-mapping-options-2026-10-17.md`, `cmd/cloud-tools-bootstrap/spikemeasure`, `ci/pdfua2/adapter/process_linux_test.go`, and `internal/mappingoptions`. Generic tracker/governance work is frozen; R5-02 remains undecided.

## Unchanged criteria and accepted bounded interpretation

The original issue requires: **AC1** an options comparison with expected associations, assumptions, ambiguity behavior, and limits for every candidate; **AC2** accuracy plus missing, multiple, conflicting, and unsafe synthetic controls that verify no unconfirmed suggestion is applied; **AC3** a recommendation that makes further implementation a separate decision; and **AC4** no automatic mapping or general-support authorization.

The bounded accuracy interpretation is fixture-authored correct-versus-wrong-valid-association discrimination on the named statement/invoice cases, with distinct failure controls and no extrapolation to option performance, population accuracy, or user performance.

**Owner acceptance, 2026-10-18 10:52:44 UTC:** AC2f is satisfied for this bounded discovery task by two separate test-only evidence strands: fixture-association discrimination and the existing literal-input confirmation gate's accepted-list behavior. This acceptance does not assert that a fixture-derived proposal traverses that gate, that a combined or causal application path exists, or that the tracker issue is ready to close. The approval record is preserved in the dated evidence supplement and the owner's task conversation.

The stronger reading would require a fixture-derived unconfirmed proposal to traverse a gate bound to the exact source, cell, and header. That would need a test-only synthetic confirmation-to-accepted-list composition. Causal in-memory repair through `repairV2Row2AmountHeaders` is a distinct stronger control and remains conditional on a separate explicit decision. A runtime application requirement is outside this plan. Simulated confirmation is not authenticated operator authority or semantic/AT review.

No dependent confirmation or application code has been added because the accepted interpretation does not require that stronger composition. Independent review, exact-head checks, and protected PR integration remain required. This bounded AC2f acceptance does not imply overall issue closure or authorize tracker mutation.

## Existing evidence and source pins

At the inspected implementation base, the task was open in `.beads/issues.jsonl`. The corrected paper comparison (`docs/research/financial-table-mapping-options-2026-10-16.md`) expressly says it did not open or parse PDFs and treats associations as hypotheses. `internal/financialfixtures/mapping_options_spike_test.go` reads no fixtures and only appends a supplied pair when booleans allow. PR #126 fixed a missing-stub call-site/vacuity issue, not fixture accuracy. PR #143 later changed the tracker record; that subsequent status change does not establish the evidence or acceptance described here. Preserve these distinctions; do not rewrite historical statements in the earlier report.

The current v2 PDF SHA-256 values, independently recomputed in the refreshed worktree, are:

| Family | Positive PDF | Damaged PDF | Authored target | Damaged relation |
|---|---|---|---|---|
| Statement | `4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec` | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | Second data row, third column, `$42.10`; `Amount` / `amountHeader` | `Description` / `descriptionHeader` |
| Invoice | `ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72` | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | Second data row, third column, `$8.00`; `Unit amount` / `unitAmountHeader` | `Quantity` / `quantityHeader` |

Paths: `docs/fixtures/remediation/financial-tables-v2/`. The v1/v2 manifests and historical repair records are the authored oracle sources. Fixed topology: one page/table, one header row, two data rows, three columns, no spans/merges, ASCII English. Position identifies these pinned test cells; it is not a general mapping algorithm. Re-verify all four hashes, target and non-target facts, actual PDF objects, and content bindings in this branch. Preserve original fixtures/manifests byte-for-byte.

Keep three independent evidence channels: (1) frozen literal oracle from authored manifests; (2) observation of verified PDF bytes, parsed `/Headers`, TH `/ID`/IDTree, target TD/cell identity, page/MCID/displayed text, `/ParentTree`, `/StructParents`, `/Pg`, `/P`, and related bindings; (3) assertions comparing observation to the independent oracle. A fresh raw-object review must independently verify these bindings and the input pins. Never derive `want` from the observer, repair helper, or candidate. Do not add a second general PDF parser or public runtime API. Existing `pdfcpu` use remains test-only. Negative structure cases start from fresh parses of verified bytes and mutate disposable in-memory state only; describe them as parsed-state mutations, not new PDF fixtures.

## Control matrix

Run applicable cases for both families. Assert explicit status/reason, values, alternatives, and exact state. One error path cannot count as several controls.

| Control | Required evidence |
|---|---|
| Source binding | Exact positive/damaged hashes parse; changed byte, unknown digest, wrong family/path, oversize, malformed input fail at distinct identity/acquisition boundaries. |
| Actual association | Observe the target's actual relation; distinguish positive correct from damaged wrong-valid relation; independently assert non-target relationships. |
| Correct vs wrong-valid proposal | Correct pair is valid but unconfirmed; a proposal to another real/resolvable TH mismatches oracle. Always-accept and always-reject mutants fail. |
| Missing relation / absent proposal | Fresh mutation removes only target `/Headers`, or proposal is absent; explicit distinct status/reason; prove mutation and valid unrelated preconditions. Empty accepted-list alone is insufficient. |
| Multiple relations | Target contains correct and wrong-valid IDs, plus duplicate variant; report all and hold without first/last choice; reverse array order. |
| Multiple proposals | Surface two alternatives for the same source/cell and hold without winner or blanket-confirmation selection. |
| Conflict | Two explicit records disagree for the same source/cell; both orders return conflict/no winner. This differs from the correct proposal differing from current damaged relation. |
| Unknown/dangling header | Nonexistent ID, mismatched IDTree entry, or missing TH reaches reference validation and fails explicitly. |
| Wrong cell/family | Existing header proposed for neighboring cell or cross-family pair is rejected despite confirmation. |
| Content binding | Mutate MCID-bound displayed content with valid titles still present; reject identity mismatch. |
| Forbidden delta | Allow only named `/Headers` relation; reject requested value/order/other-association/structure/resource changes based on actual delta, not an `unsafe` flag. |
| Separate gate evidence | Existing literal-input harness tests call `spikeApplier.apply` and assert accepted-list outcomes. Attribute this only to test-local boolean-gate behavior; do not claim fixture-specific source/cell/header binding or a product application path. |
| Optional confirmation binding | Only if the owner requires tighter linkage, compose the validated fixture proposal with a synthetic confirmation bound to exact source/cell/header, then feed the result to the existing accepted-list gate. Keep this test-only. |
| Preservation | Read-only trials preserve original bytes/digests and parsed state. If causal scope is accepted, assert actual in-memory positive change and no change in every rejected arm. |
| Unsupported shape/bounds | Reject malformed refs, cycles, over-depth/node, spans/multirow headers at a named boundary. Use existing limits; do not expand topology or promise filesystem deadlines. |

For missing-input anti-vacuity, assert exact classification/reason. Since `spikeApplier.apply` returns no value, deleting its call on an already rejected missing input can still leave an empty list; do not claim that deletion is mutation-killed. Verify the call site by source review. Separate direct gate characterization from checker/composition evidence.

The A–E supplement can state common observed fixture facts but must distinguish option-specific evidence: A gets proposal validation/simulated gate characterization only; B has exact case keys but no runtime lookup; C has observed headers but no score/threshold calibration; D has authored topology but no inference; E has explicit trial records, not a file import schema. Do not say all five options were tested or report an aggregate accuracy score. Recommend that selection/build/use remain a separate decision and preserve operator-confirmed mapping.

## Conditional causal application scope

Only if the owner separately accepts causal state-change evidence, add the minimal test-only wrapper around existing `repairV2Row2AmountHeaders`. For each family, parse exact damaged bytes anew. A valid explicit proposal and clearly labeled synthetic confirmation bound to exact source SHA and cell/header must change only the target in-memory `/Headers`; independently compare every other structure/content/resource fact. Reparse for unconfirmed, wrong-source, wrong-cell, wrong-header, missing, multiple, conflicting, and unsafe arms; invoke the same wrapper even when rejected and prove state unchanged. Guard-removal must fail by causing an unexpected state change; always-reject must fail the positive. Never serialize or publish a PDF. Exact source/cell/header confirmation-to-gate composition without this mutation is a separate, smaller optional scope.

Call this only **test-only integration of a simulated confirmation gate with an existing bounded in-memory repair helper**. It proves no product application path, authenticated identity, byte-output preservation, candidate validation, semantic/AT review, or conformance. This causal scope was not separately accepted or implemented. The accepted bounded AC2f interpretation relies on the separate fixture-association and literal-input gate evidence described above; it does not claim causal state-change evidence.

## Test-first slices and reviews

Use the local adapted `dap-test-driven-development` and `writing-good-tests`, systematic-debugging, verification-before-completion, requesting/receiving-review, bounded-parallel-dispatch, and feature-to-PR guides. Provenance: `docs/planning/imported-agent-skills.md`. No standalone upstream writing-plans or subagent-driven-development skill is installed. No installer, plugin, helper executable, framework, or dependency is needed.

For each behavior change: identify regression, add a focused compiling test, observe behavioral RED, implement the smallest predicate, observe GREEN, then refactor while green. Compile failure from a missing helper is setup, not RED. Preserve preexisting files. Use one implementer per slice, then a fresh spec-compliance reviewer and separate quality/security reviewer; reviewers inspect exact code, commands, and outputs. Coordinator integrates. Final branch receives independent evidence/claims review.

1. **Freeze oracle/source identity** (independent of AC2f): test-only reader/oracle characterization, four hashes and family facts, wrong-source controls. Existing correct behavior is characterized without fabricated RED; use a targeted observer mutant and restore it. Commit `test(financialfixtures): pin mapping evidence oracle`.
2. **Association discrimination:** actual relations; correct vs wrong-valid, wrong-cell/family, binding; smallest private test helper. Kill always-accept/reject/normative, wrong-target, skipped-binding mutants. Commit `test(financialfixtures): verify bounded association discrimination`.
3. **Failure controls:** fresh parsed-state missing, multiple, duplicate, conflicting, unknown/dangling, forbidden-change cases per family; prove mutation preconditions, reasons, alternatives, exact state. Commit `test(financialfixtures): cover mapping evidence failure cases`.
4. **Acceptance boundary:** preserve existing literal gate characterization and attribute its evidence separately from fixture-specific classification. Add a source/cell/header confirmation-to-gate composition only if the owner requests that linkage; causal helper application requires its own scope decision. Preserve the PR #126 call site and do not claim an empty-list-only mutant kill.
5. **Mutants/report:** one predicate at a time in isolated worktree; record exact test/exit and restored green result. Targets: source/read binding, ignored relation, wrong-valid acceptance, first/last choice, ignored conflict, skipped safety/target/binding, incorrect missing classification, always-reject, comparator omission, family swap, and removal of the existing gate's boolean-confirmation guard. Attribute the final mutant only to direct existing-gate characterization; it does not prove source/cell/header binding by the new checker. Compiler error, unrelated parser failure, timeout, or setup is not a kill. No mutation framework or coverage percentage. Commit `docs(research): record bounded mapping association evidence`.
6. **Integrate/review:** compare final diff to exact merge base; verify no fixture/manifests edit, runtime API/dependency/inference/writer, tracker/governance change, PR #132 overlap, or overclaim of old gate evidence. Review all #122/#133 findings irrespective of resolved labels. Report accepted, partial, or unresolved; tests alone do not establish closure/readiness.

Stop dependent work for unaccepted interpretation, source/hash/oracle mismatch, unexpected path ownership, unreachable predicate, surviving required mutant, preservation failure, inaccessible durable evidence, blocked required checks, or unsupported tracker mechanics. After three failed attempts on a slice, preserve evidence and report the blocker. Continue safe independent work without scope expansion.

## Verification and completion

Focused suite (use actual implemented test names):

```sh
go test -race ./internal/financialfixtures -run '^TestMappingEvidence' -count=1 -v
go test -race ./internal/financialfixtures -run '^TestHarness' -count=1 -v
go test -race ./internal/financialfixtures -run '^(TestFinancialTableV2MetricCompleteFixtures|TestFinancialTableV2RetainedCandidatesMatchPinnedStructure|TestStructureBindingNegativeControls)$' -count=1 -v
```

Required root/nested/publisher gates are in current `AGENTS.md`: gofmt, `go mod tidy -diff`, `go vet ./...`, `go test -race ./...`, traceability checker, cloud-bootstrap nested checks, shell syntax, adapter process-contract checks with `PDFUA_REAL_CLI_JAR=`, and publisher tests. Do not install tools or run Java/veraPDF, validator provisioners, Neuro download/auth, or historical repair experiments. Existing root tests may create temporary repair candidates and run text/render checks; do not call the entire suite read-only. If a governing restriction disallows that existing temporary-output test, stop before full suite rather than silently skipping it.

Record exact source/test/helper blob IDs, Go/platform versions, UTC time, commands and complete output/exits; oracle provenance; every case by family and outcome; each behavioral RED/GREEN; targeted mutant and restored result; fixture/output preservation; AC-to-evidence table; conditional control not run; exact-head reviews/checks. Compare tracked fixture/manifest blob IDs and hashes and inspect owned temporary destinations as well as `git diff`.

Plan completion means independent confirmation of facts, scope, acceptance mapping, and executable verification design. Evidence implementation requires accepted scoped observations, case ledger, mutation sensitivity, preservation, exact-head reviews/checks. Original-task acceptance requires a separate AC1–AC4 judgment. Publication, merge, and closure require their own authorized outcomes; none implies another.

## Closure remains separate

At the inspected implementation base, the issue was open. PR #143 subsequently changed the tracker record on main; this plan does not assess or endorse that transition. PR #133 is not an implementation baseline and must not be merged or re-enabled. Accepted evidence does not by itself authorize issue closure. Any closure assessment must separately establish original-criteria acceptance or an authorized product-reviewed split, current review disposition, exact-head checks and durable evidence, plus a supported/authorized tracker transition. Do not edit JSONL, initialize/write Dolt, normalize missing values broadly, invent close metadata, or extend frozen tooling. R5-02 remains undecided. The v3 source/resource choice is unrelated and unapproved here.

## Key sources

- `.beads/issues.jsonl`; `docs/decisions/0003-operator-confirmed-financial-table-mapping.md`; `docs/decisions/0004-approver-persona-review-protocol.md`; `docs/decisions/0011-round-five-owner-decisions.md`.
- `docs/research/financial-table-mapping-options-2026-10-16.md`; `docs/fixtures/remediation/financial-table-experiment.v1.json`; `financial-table-experiment.v2.json`; `internal/financialfixtures/fixtures_test.go`, `fixtures_v2_test.go`, `repair_attempt_test.go`.
- `docs/planning/imported-agent-skills.md`; `.agents/skills/dap-test-driven-development/SKILL.md`; `.agents/skills/dap-test-driven-development/writing-good-tests.md`; `.agents/skills/dap-requesting-code-review/reviewer-brief.md`.
- PR #133 finding and independent follow-up: https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/133#discussion_r4305118333 and https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/133#discussion_r4305221496. PR #122 findings: https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/122#discussion_r4303139773 and https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/122#discussion_r4303212365.
