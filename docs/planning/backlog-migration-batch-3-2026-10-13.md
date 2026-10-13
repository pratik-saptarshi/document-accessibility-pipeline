# Backlog migration — batch three — 2026-10-13

## Baseline and recovery boundary

This migration is based on fetched accepted `origin/main` commit `d398d320d6f01d330d1449e6bdccf3ec2d073497` (PR 41), on branch `docs/batch-three-backlog-refinement`. The Beads 1.3.1 export at the baseline contains 71 issues and has SHA-256 `144e7360da6d2cc9ac14e3cce08c48af1d71b7d68effc61b29b5fcf6ed89fc4b`. A supported isolated Dolt backup was synced before edits; its manifest SHA-256 is `057c58d0cc9df383ae9ae737925b89fb10f788e57ceee950883b183cc167e099`.

The checkout had no live `.beads/embeddeddolt/` database to use. No runtime database was deleted, replaced, or repaired. Beads was activated from the documented `/workspace/.codex-tools/cloud-build-tools/activate.sh` path and used with a writable temporary `HOME`. The isolated database was initialized from the exact accepted-main export; its backup preserves this imported snapshot, not unavailable earlier Dolt history. The JSONL export remains an issue interchange snapshot, not a full history backup.

The baseline `bd ready` command listed 10 records from the imported 71-issue snapshot. The baseline was not modified before its export and backup were captured.

## Scoped decisions and closures

- Refined `dap-jo9.1.1.1` to one synthetic statement transaction table and one synthetic invoice line-item table, equal in priority. It specifies a fixture-authored target association, preservation of all other declared relationships/order and exact displayed values, experiment-only language/font/topology parameters, and explicit validator, mutation, malformed, unsupported, and ambiguous controls. It does not select a writer, production default, arithmetic policy, or generation profile.
- Replaced the broad readiness of `dap-jo9.1.2` with an actual, bounded structural-change experiment on A's two fixtures. No-op PDF 1.7 writes are insufficient. Partially tagged PDF 2.0 can isolate editing only; it does not prove PDF 1.7 conversion or general untagged/complex-table support. Exact candidate/license evidence, independent comparisons, explicit candidate-hash-bound `ua2` evidence, early failure controls, and proceed/reject/inconclusive outcomes are required.
- Closed `dap-jo9.2.3.3` after independently confirming source tests for independent baseline/writer reads, `/StructParents` value `99` mutation rejection, removal of `/Lang`, `/MarkInfo`, `/StructTreeRoot`, and `/AcroForm`, and four-family positive writes. The focused three-test `go test -race` run passed. This is preservation characterization only, not writer or conformance evidence.
- Closed `dap-jo9.17` for its accepted mapping scope: PR 41 merge `d398d320d6f01d330d1449e6bdccf3ec2d073497`, head `c801de2ebeb7e51dfe733c60a4939919027d2a11`, nine required pre-merge checks, seven post-merge checks, and accepted independent reviews. A read-only GitHub CLI query was denied by the environment; the PR/check/review result here is the accepted evidence supplied with this migration task.
- Replaced all eight planning-only implementation definitions of done in `dap-jo9.2.3.2`, `dap-jo9.2.4.2`, `dap-jo9.3.1.2`, `dap-jo9.3.3.2`, `dap-jo9.3.3.3`, `dap-jo9.3.4.2`, `dap-jo9.3.4.3`, and `dap-jo9.5.2`. Each now requires failing tests before behavior, recorded passing execution with artifact/hash/version evidence, explicit supported/unsupported scope, negative/failure controls, necessary accepted gates, no silent policy defaults, and separation of engineering from release evidence.
- Replaced the AFP implementation placeholder in `dap-jo9.4.1.2` with bounded feasibility evidence, resource/input and target criteria, licenses/versions where applicable, a go/no-go/unknown result, no converter implementation, and explicit preservation of existing blockers.

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

The canonical traceability catalog contains 21 tasks after adding only `dap-jo9.2.5.9`; A and B remain Beads children outside the catalog task list. All 26 case specifications and their canonical owners are unchanged. The human and assistive-technology role conventions are recorded in the existing mapping fixtures. The generated matrix is current.

The current next step is the actual A fixture package and bounded feasibility work, or a precise no-go. This batch ran neither the experiment nor PDF repair and delivered no user-facing product flow. See [next steps](next-steps-batch-3-2026-10-13.md).

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

The actual post-change `bd ready` output listed 10 records: `dap-jo9`, `dap-jo9.1`, `dap-jo9.1.1.1`, `dap-jo9.1.1.3`, `dap-jo9.1.1.4`, `dap-jo9.1.1.2`, `dap-jo9.10`, `dap-5hy.2`, `dap-jo9.14`, and `dap-5hy.3`. It includes epics, so it is not a count of implementation-ready leaf tasks. Child A is absent because its explicit prerequisite `dap-jo9.1.1.1` remains open; if that decision is resolved, A can proceed. A blocks `dap-jo9.1.2`; child B waits for `dap-jo9.1.1` and `dap-jo9.1.2`; C waits for `dap-jo9.1.1.4` and `dap-jo9.1.2`. `bd stats` reports 9 ready-to-work leaf issues; that summary excludes the two listed epics, while `bd ready` lists them. No hierarchy edge silently adds a blocker to A.

The final export SHA-256 is `47d11433227b61130c34b9dccbef4f49967c182cb0342c1808a740c1c7f7bcd5`. A fresh isolated Beads import/export reproduced all 74 IDs and the status, priority, dependency, description, acceptance, label, close-reason, and close-time fields with zero differences. The supported after-state Dolt backup manifest SHA-256 is `fea42a417622b3e8d4cb1045495aafbafcb5689e51240b96a8dc9c304fcffaba`.

## Validation and remaining gates

- Root Go module: formatting, `go mod tidy -diff`, `go vet ./...`, `go test -race ./...`, and traceability check passed. Traceability reports 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases.
- Cloud-tools bootstrap module: race tests, vet, and tidy-diff passed; bootstrap shell syntax passed.
- CI PDF/UA-2 adapter module: `PDFUA_REAL_CLI_JAR=` race tests, vet, and tidy-diff passed. No Java or real veraPDF process was run.
- Source-publisher shell syntax/guard tests and `actionlint` passed. Beads artifact checks, lint, conventions, `bd ready`, and dependency-cycle checks passed; no duplicates or missing targets were found.
- No local vulnerability-database fetch, model download/authentication, candidate tool execution, writer implementation, or PDF repair experiment was performed.

Open gates include exact normative structure and validator limits until evidenced by A, writer/capability/license feasibility, supported topology, production/sensitive handling, accepted reviewer authority/protocol, generation engine and source conversion, and actual candidate-specific release evidence. Engineering may report unsupported or review-needed; no production defaults or human/AT pass were invented.
