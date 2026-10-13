# Backlog migration — batch three — 2026-10-13

## Baseline and recovery boundary

This migration is based on fetched accepted `origin/main` commit `d398d320d6f01d330d1449e6bdccf3ec2d073497` (PR 41), on branch `docs/batch-three-backlog-refinement`. The Beads 1.3.1 export at the baseline contains 71 issues and has SHA-256 `144e7360da6d2cc9ac14e3cce08c48af1d71b7d68effc61b29b5fcf6ed89fc4b`. A supported isolated Dolt backup was synced before edits; its manifest SHA-256 is `057c58d0cc9df383ae9ae737925b89fb10f788e57ceee950883b183cc167e099`.

The checkout had no live `.beads/embeddeddolt/` database to use. No runtime database was deleted, replaced, or repaired. Beads was activated from the documented `/workspace/.codex-tools/cloud-build-tools/activate.sh` path and used with a writable temporary `HOME`. The isolated database was initialized from the exact accepted-main export; its backup preserves this imported snapshot, not unavailable earlier Dolt history. The JSONL export remains an issue interchange snapshot, not a full history backup.

The baseline `bd ready` command listed 10 records from the imported 71-issue snapshot. The baseline was not modified before its export and backup were captured.

## Scoped decisions and closures

- Refined `dap-jo9.1.1.1` to configuration-only. It defines the versioned protocol, equal-priority statement/invoice fixtures, fixture-authored target association, experiment-only parameters, expected outcomes, and evidence owners. Its acceptance states: “This issue completes when the versioned synthetic experiment protocol, declared engineering parameters, expected outcomes, and evidence responsibilities are reviewed. Actual fixture bytes, hashes, baselines, and normative-control verification belong to `dap-jo9.2.1.2`; attempted repair, candidate outputs, `ua2` reports, and feasibility disposition belong to `dap-jo9.1.2`. None is required to close this configuration issue.” No production defaults or arithmetic policy are selected.
- Refined `dap-jo9.1.2` to separate `proceed` evidence (actual repair and complete candidate evidence on both fixtures) from `reject`/`inconclusive` evidence (bounded attempted steps, exact failure or missing prerequisite, limits, and next decision; no invented output/report). Independent acceptance of any finite disposition completes the issue. No-go selects no production writer or closes unrelated product gates. No-op PDF 1.7 writes are insufficient; partially tagged PDF 2.0 proves neither PDF 1.7 conversion nor general untagged/complex-table support.
- Closed `dap-jo9.2.3.3` after independently confirming source tests for independent baseline/writer reads, `/StructParents` value `99` mutation rejection, removal of `/Lang`, `/MarkInfo`, `/StructTreeRoot`, and `/AcroForm`, and four-family positive writes. The focused three-test `go test -race` run passed. This is preservation characterization only, not writer or conformance evidence.
- At accepted PR 41 main, `dap-jo9.17` remained in progress pending its post-merge evidence. This batch closed it only after accepting that evidence: PR 41 merge `d398d320d6f01d330d1449e6bdccf3ec2d073497`, head `c801de2ebeb7e51dfe733c60a4939919027d2a11`, nine required pre-merge checks, seven post-merge checks, and accepted independent reviews. A read-only GitHub CLI query was denied by the environment; the PR/check/review result here is the accepted evidence supplied with this migration task.
- Replaced stale planning/decomposition language in all eight implementation successors (`dap-jo9.2.3.2`, `dap-jo9.2.4.2`, `dap-jo9.3.1.2`, `dap-jo9.3.3.2`, `dap-jo9.3.3.3`, `dap-jo9.3.4.2`, `dap-jo9.3.4.3`, and `dap-jo9.5.2`), keeping each linked case a specification rather than execution evidence. Their executable DoDs still require failing tests before behavior, recorded passing execution with artifact/hash/version evidence, explicit supported/unsupported scope, negative/failure controls, necessary accepted gates, no silent policy defaults, and separation of engineering from release evidence. In `dap-jo9.2.4.2`, the batch-two verbatim-acceptance note is now dated to that batch; batch three intentionally replaces the implementation DoD while preserving its four case mappings, roles, and canonical owners.
- Replaced the AFP implementation placeholder in `dap-jo9.4.1.2` with feasibility-only successor wording and bounded evidence requirements: representative resources/input, target criteria, licenses/versions where applicable, go/no-go/unknown, no converter implementation, and existing blockers preserved.
- Narrowed `dap-jo9.1.3` to clearance for only the exact candidate proposed for the finite repair operation, including its reachable code/dependencies/assets, pinned versions, applicable permissive licenses, capability evidence, and allowed execution boundary. Historical alternatives remain untested, are not universal prerequisites, and no engine or runtime exception is approved.
- Narrowed `dap-jo9.2.2` to the smallest consumed input/profile/candidate/evidence/error interface and permissive dependency boundary. Its valid/invalid, deterministic evidence, cancellation/timeout, applicable idempotence, unsupported, tool-failure, and license-policy cases remain; no six-interface worker framework was added.

## New children and exact dependency ledger

Created three real Beads children:

- `dap-jo9.2.1.2` (P1), under `dap-jo9.2.1`: author pre-spike financial-table fixtures and immutable independent baselines. It defines exact displayed values, provenance and hashes, target associations/order, support matrix/exclusions, independent text/render/structure observations, normative representation and validator limits, and pre-repair controls. It claims no repair.
- `dap-jo9.2.1.3` (P2), under `dap-jo9.2.1`: expand corpus only after bounded repair feasibility and only against evidenced capabilities and accepted profiles; retain all four remediation families. The existing closed `dap-jo9.2.1.1` remains a historical metadata-manifest task, not fixture-artifact evidence.
- `dap-jo9.2.5.9` (P1), under `dap-jo9.2.5`: implement a candidate-bound semantic and assistive-technology review gate. It uses existing `TC-REM-HUMAN-001` and `TC-REM-ASSISTIVE-TECH-001` specifications with bounded implementation roles. Their canonical owner remains `dap-jo9.2.5.2`; no case definitions or owner assignments changed. Simulated outcomes are not human passes.

`X depends on Y` describes a blocking edge. The complete ledger delta is:

| Operation | Edge |
|---|---|
| Removed | `dap-jo9.2.1` depends on `dap-jo9.1.2` |
| Added hierarchy | `dap-jo9.2.1.2` is child of `dap-jo9.2.1` |
| Added hierarchy | `dap-jo9.2.1.3` is child of `dap-jo9.2.1` |
| Added hierarchy | `dap-jo9.2.5.9` is child of `dap-jo9.2.5` |
| Added blocker | `dap-jo9.2.1.2` depends on `dap-jo9.1.1.1` |
| Added blocker | `dap-jo9.2.1.3` depends on `dap-jo9.1.1` |
| Added blocker | `dap-jo9.2.1.3` depends on `dap-jo9.1.2` |
| Added blocker | `dap-jo9.1.2` depends on `dap-jo9.2.1.2` |
| Added blocker | `dap-jo9.1.2` depends on `dap-jo9.1.3` |
| Added blocker | `dap-jo9.2.5.9` depends on `dap-jo9.1.1.4` |
| Added blocker | `dap-jo9.2.5.9` depends on `dap-jo9.1.2` |

No dependency makes invoice work wait on statement work. All other original edges remain unchanged, including DOCX/Markdown, generation-engine, safety, product, and release gates.

## Roadmap and traceability

The roadmap's duplicated phase lists are consolidated into one sequence: shared validation/review foundations; the A → finite writer/fidelity experiment → B remediation path; narrow remediation candidates; semantic/reviewer handoff; separate generation with unchanged DOCX/Markdown and engine gates; and separate AFP feasibility. Early bounds, cancellation, redaction, partial-output, and audit-failure controls remain early prerequisites; production and release gates remain intact. Historical Java/Python tool mentions are not selected and create no product-runtime exception.

Updated the product-scope and remediation-workflow wording to require permissive open-source product dependencies and to state that veraPDF remains isolated to CI; no other non-Go product-runtime exception is approved.

The canonical traceability catalog contains 21 tasks after adding only `dap-jo9.2.5.9`; A and B remain Beads children outside the catalog task list. All 20 existing catalog-task case bindings, all 26 case specifications, and their canonical owners are unchanged. C reuses two existing case specifications with the bounded implementation roles in the existing mapping fixtures. The generated matrix is current.

The current next step is the actual A fixture package and bounded feasibility work, or a precise no-go. `dap-jo9.2.5.9` can test labeled synthetic evidence-record behavior after its prerequisites, but closure additionally requires an actual eligible candidate with hash-bound preservation/`ua2` evidence from a successful bounded experiment or candidate slice. A no-go/inconclusive result does not close C. Phase 3 pipeline integration is not a prerequisite for the earlier contract tests or qualified candidate demonstration. This batch ran neither the repair experiment nor PDF repair and delivered no user-facing product flow. See [next steps](next-steps-batch-3-2026-10-13.md).

## Before/after health and readiness

| Measure | Before | After |
|---|---:|---:|
| Issue records | 71 | 74 |
| Open / closed / in progress | 35 / 30 / 6 | 38 / 32 / 4 |
| Parent-child / blocking edges | 69 / 52 | 72 / 58 |
| Total edges | 121 | 130 |
| Duplicate IDs / missing dependency targets / blocker cycles | 0 / 0 / 0 | 0 / 0 / 0 |
| Traceability tasks / case specifications | 20 / 26 | 21 / 26 |

All original 71 IDs, priorities, case-owner assignments, and non-listed statuses remain unchanged. The only status changes are `dap-jo9.17` and `dap-jo9.2.3.3`, both in progress to closed. The three new issues are open. The only removed edge and all ten added edges are listed above.

The export has 42 unfinished issues; 28 have an unresolved direct blocking edge (27 open and `dap-jo9.1.2` in progress). The isolated runtime reported 29 blocked (28 open plus `dap-jo9.1.2`); its additional classification is not equated with this direct-edge metric. The final `bd ready` command listed 10 records: `dap-jo9`, `dap-jo9.1`, `dap-jo9.1.1.1`, `dap-jo9.1.1.3`, `dap-jo9.1.1.4`, `dap-jo9.1.1.2`, `dap-jo9.10`, `dap-5hy.2`, `dap-jo9.14`, and `dap-5hy.3`. `bd stats` reported 9 ready-to-work and 29 dependency-blocked; `bd blocked` also listed 29. The difference between the `bd ready` listing count and stats readiness remains unestablished; no arithmetic explanation is inferred. In a disposable import of the final export, closing only `dap-jo9.1.1.1` caused `dap-jo9.2.1.2` to appear in `bd ready`; this confirms the hierarchy does not silently block A. That simulated close was not exported. In the final export, the configuration issue remains open and A waits on it. A blocks `dap-jo9.1.2`; B waits for `dap-jo9.1.1` and `dap-jo9.1.2`; C waits for `dap-jo9.1.1.4` and `dap-jo9.1.2`.

The corrected tracked export `.beads/issues.jsonl` has SHA-256 `8930efc66d52846702cffbbfb58b6786145ad99b8ef1ac2aa68893195d76e1f6`. A fresh isolated Beads database imported this exact export and re-exported all 74 issues; the IDs and status, priority, dependency, description, acceptance, label, close-reason, and close-time fields matched with zero differences. The supported after-state Dolt backup was synced from that isolated import of the corrected export; its manifest SHA-256 is `746885609d8dd3fc1de7bdb95873cfb4f4ee7d8de98947b7a6052a4fac3aca0f`.

## Validation and remaining gates

- Root Go module: formatting, `go mod tidy -diff`, `go vet ./...`, `go test -race ./...`, and traceability check passed. Traceability reports 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases.
- Cloud-tools bootstrap module: race tests, vet, and tidy-diff passed; bootstrap shell syntax passed.
- CI PDF/UA-2 adapter module: `PDFUA_REAL_CLI_JAR=` race tests, vet, and tidy-diff passed. No Java or real veraPDF process was run.
- Source-publisher shell syntax/guard tests and `actionlint` passed. Beads artifact checks, lint, conventions, `bd ready`, and dependency-cycle checks passed; no duplicates or missing targets were found.
- Post-correction review found 0 of 8 implementation successors with planning-only DoDs and 0 AFP implementation placeholders. This is planning refinement, not product delivery.
- No local vulnerability-database fetch, model download/authentication, candidate tool execution, writer implementation, or PDF repair experiment was performed.

Open gates include exact normative structure and validator limits until evidenced by A, writer/capability/license feasibility, supported topology, production/sensitive handling, accepted reviewer authority/protocol, generation engine and source conversion, and actual candidate-specific release evidence. Engineering may report unsupported or review-needed; no production defaults or human/AT pass were invented.
