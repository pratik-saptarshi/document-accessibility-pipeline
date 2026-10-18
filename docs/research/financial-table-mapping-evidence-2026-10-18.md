# Bounded financial-table mapping evidence

- Date: 2026-10-18 UTC
- Execution verification record finalized: `2026-10-18T06:21:54Z` (UTC), after the PR #142 coverage follow-up; the owner-accepted AC2f interpretation was added at 10:52:44 UTC.
- Status: **bounded AC2f interpretation accepted by owner** on 2026-10-18 at 10:52:44 UTC; this does not claim original-task acceptance, tracker closure, or broader product approval.
- Refreshed base: `6d42691d2c95ac0b369408f527d5994fb187d5d3` (`origin/main`); work branch: `test/financialfixtures-mapping-evidence`.
- Local toolchain: `/workspace/.codex-tools/cloud-build-tools/toolchains/go/1.27.2/bin/go`, Go `1.27.2 linux/amd64`.
- Follow-up implementation test file SHA-256: `b974c2e948fbe85c26ef0cb9b9484976a8a23a91d7bf7243de1cdd5b2990c2e8`.

This supplement adds actual fixed-fixture observations to the historical paper comparison. It preserves that paper's explicit limitation: its Options A–E associations were hypotheses, and its original harness read no PDFs. The existing harness remains separate from the new fixture checker.

## Authored oracle and immutable inputs

The original open issue requires a bounded options comparison with fixture-authored expected associations, assumptions, ambiguity behavior, and limits; synthetic accuracy, missing, multiple, conflicting, unsafe, and no-unconfirmed-application controls; a separate follow-up recommendation; and no automatic mapping or general-support claim. The current `.beads/issues.jsonl` still records the issue as open. No tracker change was made.

The independent normative source is the authored v1/v2 fixture manifest and its expected table associations. The fixed v2 PDFs were rehashed in the refreshed worktree; the positive/damaged PDF objects and structure observations were checked by test assertions and direct raw-object inspection. The target cell is second data row, third column. These position facts bind only the named fixtures.

| Family | Positive SHA-256 | Damaged SHA-256 | Authored target pair | Actual damaged relation |
|---|---|---|---|---|
| Statement | `4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec` | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | `$42.10` → `Amount` / `amountHeader` | `$42.10` → `Description` / `descriptionHeader` |
| Invoice | `ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72` | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | `$8.00` → `Unit amount` / `unitAmountHeader` | `$8.00` → `Quantity` / `quantityHeader` |

All four source files and both v1/v2 manifests are byte-for-byte unchanged from refreshed main. Each family has one page, one table, one header row, two data rows, three columns, and no spans/merges. The tests also assert every non-target `/Headers` relationship, reading order, target content, IDTree mapping, MCID/displayed-text binding, `/ParentTree`, `/StructParents`, `/Pg`, and `/P` facts. These are bounded fixture observations, not a general parser or inference algorithm.

## Executed case ledger

The focused mapping suite ran all 74 named leaf subcases below under `-race`; all passed. This follow-up adds assertions inside existing correct-proposal and classifier-failure cases, so the leaf count remains 74. Names are shown by exact test group and family so that each observation and failure remains traceable.

| Test and cases | Result / asserted reason | Preservation / scope |
|---|---|---|
| `TestMappingEvidenceV2FixtureObservations`: statement positive/damaged; invoice positive/damaged | All four pinned inputs parsed and matched the independent oracle for target and non-target relations, header IDs/titles, order, and content bindings. | Read-only immutable source bytes and fresh parsed contexts. |
| `TestMappingEvidenceRejectsCrossFamilySourceDigest`: statement-positive-as-invoice; statement-damaged-as-invoice; invoice-positive-as-statement; invoice-damaged-as-statement | All four reject at `SHA-256` identity check and return no bytes. | Wrong-family inputs fail before parsing/classification. |
| `TestMappingEvidenceSourceAcquisitionBoundaries`: changed-byte-keeps-original-digest; unknown-digest; oversized-input; malformed-pdf-is-not-missing-association | Changed/unknown digest → SHA mismatch; oversized input → configured byte-limit rejection; correctly pinned malformed bytes → parser failure, not an association status. | Only `t.TempDir` copies are changed; the changed-byte precondition starts from bounded, hash-verified bytes. |
| `TestMappingEvidenceProposalDiscrimination`, both families: positive/correct-proposal; damaged/correct-proposal; damaged/wrong-existing-header; damaged/wrong-target-cell; damaged/cross-family-proposal | Correct positive and damaged proposals → `valid` from both the proposal validator and `classifyMappingEvidence`; real but wrong header → `association_mismatch`; wrong cell → `wrong_target_cell`; wrong family → `family_mismatch`. | Proposals are explicit test inputs, not generated by an inference engine. Both family directions are exercised. The four correct-proposal cases assert classifier success, exact proposal preservation, target preservation, and fresh re-observation of all other fixed-table cells and marked-content bindings. |
| `TestMappingEvidenceProposalSourceAndHeaderReferences`, both families: correct-source-and-header-reference; wrong-source-digest; unknown-proposed-header; claimed-title-does-not-match-resolved-header | Baseline → `valid`; errors → `source_mismatch`, `proposal_header_unresolved`, `proposal_header_reference_mismatch`. | Distinct identity/reference failures reach their intended predicates. |
| `TestMappingEvidenceRejectsContentBindingMismatch`: statement-positive; statement-damaged; invoice-positive; invoice-damaged | Deliberate parsed-state MCID/content swaps → `content_binding_mismatch`; title/header observations remain present. | Each trial starts from a fresh verified parse; parsed mutation is test-only and not serialized. |
| `TestMappingEvidenceClassifiesParsedAssociationFailureStates`, each family: document/missing-field; document/empty-array; document/multiple-current-then-correct; document/multiple-correct-then-current; document/duplicate-id; document/unknown-header; proposal/missing; proposal/multiple-alternatives; proposal/conflicting-records-forward; proposal/conflicting-records-reversed | Missing relation → `rejected/missing_document_association`; multiple and duplicate relations → `held/multiple_document_associations`; unknown relation → `rejected/unknown_document_header`; absent proposal → `rejected/missing_proposal`; alternatives → `held/multiple_proposal_alternatives`; disagreeing records in either order → `held/conflicting_evidence`. | Each relation mutation is independently verified before classification. After each classifier call, tests reread the parsed target, header cells and all non-target cell observations, and rerun marked-content binding checks. All alternatives are surfaced; no winner is selected. |
| `TestMappingEvidenceClassifiesRequestedSnapshotDeltas`, each family: target-header-only; displayed-text; row-order; another-association; structure-object-content-same-reference; resource-object-content-same-reference; resource-reference-swap-same-type; struct-parents; decoded-content-bytes; decoded-mcid-text | Target-only requested snapshot delta → `allowed` with exactly `table.rows[1].cells[2]./Headers`; each other delta → `rejected/forbidden_requested_change` with its exact changed path. | This compares cloned fixed-fixture snapshots; the PDF and parsed context do not change. Resource facts are limited to the fixed fixture and one child-dictionary level. It is not a general object-delta checker or application path. |

For both families, the requested target-only delta changed the copied target `/Headers` field to the authored correct ID while the comparator returned exactly one allowed path. Displayed text, row order, other-cell association, structure/resource object facts, resource references, `/StructParents`, decoded page stream, and target MCID text were each independently changed in the requested snapshot and rejected. The source parse and both snapshots were unchanged by classification.

The acquisition and structure tests also characterize bounded file reads, regular-file/context checks, exact SHA-256, and PDF 2.0 prefix validation. The parser-error case is intentionally not relabeled as a missing association.

## Existing gate evidence and the AC2 interpretation

The historical `TestHarness*` suite passed all six controls on this refreshed worktree. `TestHarnessAppliesOnlySafeConfirmedSuggestion` and `TestHarnessUnconfirmedStubValuesNeverApply` call `spikeApplier.apply` and observe its accepted list. Those tests establish only test-local behavior for supplied literal values and booleans. They do not bind those gate inputs to a fixture's source digest, target cell, or header. New fixture proposal validation and snapshot classifications do not call the gate.

| Original clause | Evidence in this supplement | Acceptance boundary |
|---|---|---|
| AC1: bounded A–E comparison, assumptions, expected associations, ambiguity behavior, limits | The 2026-10-16 paper supplies the option-by-option comparison and explicitly labels its associations as hypotheses. This supplement adds common fixed-fixture facts and explains which option-specific questions remain untested. | Whether the paper and common fixture facts meet the unchanged criterion remains a separate overall-acceptance judgment. No option-specific runtime performance or implementation result. |
| AC2 accuracy | Correct versus wrong-but-valid fixture associations are distinguished on the named pair, including correct proposals on positive and damaged PDFs and explicit wrong-header/cell/family proposals. | Call this bounded fixture-association discrimination. It is not inferred mapping accuracy, a population rate, or user performance. |
| AC2 missing/multiple/conflicting/unsafe | Missing, multiple, duplicate, unknown, alternative, conflict, and forbidden-requested-delta cases have distinct outcomes and preservation assertions on both families. | Test-only fixed-shape observation/classification; no production repair or inference. |
| AC2 no unconfirmed suggestion applied | Existing literal-input gate tests show unconfirmed literals do not enter the gate's accepted list; fixture controls independently establish association facts. | **Owner accepted this bounded interpretation** on 2026-10-18 at 10:52:44 UTC: the two evidence strands suffice for this bounded AC2f criterion. They do not show a fixture-derived unconfirmed proposal traversing the gate, a combined application path, or causal document-state protection. The approval is recorded in this supplement and the owner's acceptance is preserved in the task conversation. |
| AC3 separate follow-up decision | The paper recommends no option and names any selection/build as a separate decision. | Preserved; this evidence does not authorize a choice. |
| AC4 no automatic mapping/general support | New work is limited to test-only fixed fixtures and explicit records. | No inference, runtime API, writer, general support, conformance, human semantic/AT authority, or tracker transition. |

### Option evidence: paper versus observations

| Option | Actual evidence now available | Still only a hypothesis / untested |
|---|---|---|
| A — operator-confirmed association | Bounded proposal discrimination against the fixed oracle; existing literal accepted-list gate characterization. | Human accuracy/usability and fixture-specific confirmation/application path. |
| B — fixture-authored lookup | Exact fixture keys and authored associations are now observable. | Runtime lookup, lookup selection, or automatic application. |
| C — header-text heuristic | Actual header IDs/titles and wrong-valid alternatives are observed. | Heuristic scoring, threshold calibration, confidence, or performance. |
| D — positional/structural inference | Exact fixed topology and unsupported boundaries are observed. | No inference engine or broader regularity claim. |
| E — external mapping file | Explicit proposal records are test inputs. | No file schema, importer, authoring protocol, or runtime integration. |

The research recommendation remains: authorize no option here; decide selection/build/use separately. Do not aggregate these unequal evidence levels into “five options tested” or an accuracy percentage.

## TDD and targeted mutation evidence

The implementation slices used focused behavioral tests, observed RED, the smallest test-only predicate, then GREEN. Existing behavior was characterized; no historical failure was invented. A compiling empty classifier produced behavioral failures before the missing/multiple/conflict classifier was added. One-at-a-time mutation checks were run and restored in the dedicated evidence worktree.

Freshly rerun mutation records on this refreshed base:

| Mutation | Command and observed failure | Restored result |
|---|---|---|
| Change the missing-document-association reason to `missing_proposal`. | `go test -race ./internal/financialfixtures -run '^TestMappingEvidenceClassifiesParsedAssociationFailureStates/statement/document/missing-field$' -count=1 -v` exited 1 at the exact status/reason assertion (`missing_proposal`, wanted `missing_document_association`). | Restore exact predicate; same command exited 0. |
| Remove the gate's confirmation predicate, leaving only the unsafe check. | `go test -race ./internal/financialfixtures -run '^(TestHarnessAppliesOnlySafeConfirmedSuggestion|TestHarnessUnconfirmedStubValuesNeverApply)$' -count=1 -v` exited 1: safe-but-unconfirmed reached the applied list, and the unconfirmed lookup stub was applied. | Restore `confirmedByOperator && !unsafe`; same command exited 0. At that run, file hashes returned to `mapping_evidence_test.go` `1b71860f…0c9237` and `mapping_options_spike_test.go` `963c794a…1d9069`. |
| Force the successful classifier result to `rejected`. | `go test -race ./internal/financialfixtures -run '^TestMappingEvidenceProposalDiscrimination/statement/positive/correct-proposal$' -count=1 -v` exited 1 at the exact classification assertion: got `rejected` with the correct proposal surfaced, wanted `valid` with that proposal. | Restore `result.status = "valid"`; same command exited 0. |
| After successful classification, mutate the first data row's first cell `/Headers` to `mutantHeader`. | Same focused command exited 1 at the fresh non-target observation assertion; before was `dateHeader`, after was `mutantHeader`. | Remove the temporary parsed-state mutation; same command exited 0. This proves the newly added post-call reader catches an unrelated association change. |

Earlier exploratory mutant runs were reported during implementation, but their exact per-mutant command output and tested-file identities were not retained. They are therefore **not counted as verified mutation evidence here**. Only the four fully logged, freshly rerun mutants above are claimed. No inference or product behavior was mutated. Confirmation-binding composition/application mutants were not run because that code is not part of the accepted slice.

## Verification record

The focused command on base `6d42691d2c95ac0b369408f527d5994fb187d5d3` was:

```sh
go test -race ./internal/financialfixtures -run '^TestMappingEvidence' -count=1 -v
```

Result: exit 0; all eight evidence test groups and 74 named leaf cases passed (`ok .../internal/financialfixtures 1.403s`). The six `TestHarness*` controls also passed under `go test -race ./internal/financialfixtures -run '^TestHarness' -count=1 -v` (exit 0). Fixture/structure characterization passed under `go test -race ./internal/financialfixtures -run '^(TestFinancialTableV2MetricCompleteFixtures|TestFinancialTableV2RetainedCandidatesMatchPinnedStructure|TestStructureBindingNegativeControls)$' -count=1 -v` (exit 0). Existing retained candidates were read and structurally compared; no validator was run or rerun.

The refreshed-main verification matrix was rerun after the coverage additions. After the final preservation assertion, `go test -race ./internal/financialfixtures -run '^TestMappingEvidence' -count=1` passed (all 74 leaf cases; package 1.349s) and `go test -race ./...` passed (financialfixtures package 11.400s). `test -z "$(gofmt -l .)"`, `go mod tidy -diff`, `go vet ./...`, and `go run ./cmd/traceability-check` passed; traceability reported 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases. The nested `cmd/cloud-tools-bootstrap` race tests, vet, and tidy-diff passed (cloudtools package 17.290s), and nested `ci/pdfua2/adapter` race tests with `PDFUA_REAL_CLI_JAR=` set, vet, and tidy-diff passed (adapter package 24.867s) after the main coverage additions and before the last assertion-only change. `bash -n` passed for the cloud bootstrap and source publisher scripts; `bash -n .github/scripts/test-source-publisher.sh` and `bash .github/scripts/test-source-publisher.sh` passed. The publisher test emits expected failure diagnostics for its negative cases before reporting `source publisher tests passed`. `git diff --check` passed. The verification record was finalized by `2026-10-18T06:21:54Z` with Go 1.27.2 linux/amd64. The prior base run's temporary pdfcpu config issue was isolated by setting temporary `XDG_CONFIG_HOME` and `XDG_CACHE_HOME`; no repository configuration was changed. No Java, veraPDF, validator provisioner, or Neuro tool was invoked.

## Owner-accepted AC2f interpretation and boundary

The owner accepted the bounded AC2f interpretation: fixture-association checks and literal-input gate characterization are separate evidence strands for this discovery criterion. This does not show a fixture-derived proposal traversing the gate or any combined application path. A source/cell/header-bound composition and causal in-memory repair remain stronger, separate controls and are not part of this accepted scope. AC1's option-by-option interpretation and the original issue's overall acceptance/closure remain separate judgments; this evidence does not close the issue or authorize automatic mapping.

Sources: the unchanged open issue in `.beads/issues.jsonl`; the [original paper comparison](financial-table-mapping-options-2026-10-16.md); v1/v2 manifests in `docs/fixtures/remediation/`; `internal/financialfixtures/mapping_options_spike_test.go`, `mapping_evidence_test.go`, `fixtures_v2_test.go`, and `repair_attempt_test.go`; the [PR #133 adversarial review](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/133#issuecomment-6246216077); and the [execution plan](../superpowers/plans/2026-10-18-pr133-mapping-acceptance-evidence.md).
