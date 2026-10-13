# Backlog migration — batch two — 2026-10-13

## Scope and preserved baseline

This batch implements the mapping and legacy-requirement reconciliation assigned to P1 issue `dap-jo9.17`. It records the accepted first remediation priority—statements and invoices with tables—without selecting unresolved profile values or authorizing a writer, financial arithmetic, production handling, reviewer authority, engine, or runtime exception. The next product experiment remains actual bounded table-repair feasibility toward PDF 2.0/PDF/UA-2, not another PDF 1.7 round trip.

The batch starts from `origin/main` `fd06b9b2b0a69f2cac6cea8dc6a1760caad4bb95`, where Beads had 71 records and 121 edges. The checkout Beads runtime directory remains incomplete and untouched. The isolated Dolt store `/tmp/dap-beads-batch2` was initialized from that tracked export; the before snapshot is preserved at `/tmp/dap-issues-batch2-before.jsonl`, and the supported full Dolt backup is at `/tmp/dap-beads-batch2-baseline-backup`. The JSONL export is an upsert interchange snapshot, not a full Dolt backup.

## Accepted case mapping before → after

The 12 measured task-set mismatches and two additional ancillary task mappings are shown below. Each row records the catalog case set before this batch and its reconciled set after. The task records also have participation roles where they share a case without owning it; see the generated [traceability matrix](../traceability/matrix.md) and [mapping fixture](../../internal/traceability/testdata/mapping-expectations.json).

| Beads task | Catalog cases before | Catalog cases after |
|---|---|---|
| `dap-jo9.2.3.2` | `TC-REM-STATEMENT-001`, `TC-REM-INVOICE-001`, `TC-REM-CORRESPONDENCE-001`, `TC-REM-REPORT-001` | `TC-REM-STATEMENT-001`, `TC-REM-INVOICE-001`, `TC-REM-CORRESPONDENCE-001`, `TC-REM-REPORT-001`, `TC-REM-MALFORMED-001`, `TC-REM-FIDELITY-001` |
| `dap-jo9.2.4.1` | `TC-REM-AMBIGUOUS-001`, `TC-REM-REVIEW-ROUTE-001` | `TC-REM-AMBIGUOUS-001`, `TC-REM-REVIEW-ROUTE-001`, `TC-REM-HUMAN-001`, `TC-REM-ASSISTIVE-TECH-001` |
| `dap-jo9.2.4.2` | `TC-REM-AMBIGUOUS-001`, `TC-REM-REVIEW-ROUTE-001` | `TC-REM-AMBIGUOUS-001`, `TC-REM-REVIEW-ROUTE-001`, `TC-REM-HUMAN-001`, `TC-REM-ASSISTIVE-TECH-001` |
| `dap-jo9.2.5.2` | `TC-REM-HUMAN-001`, `TC-REM-ASSISTIVE-TECH-001` | `TC-REM-HUMAN-001`, `TC-REM-ASSISTIVE-TECH-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.2.5.3` | `TC-PDFUA2-REPORT-CONTRACT-001` | `TC-PDFUA2-REPORT-CONTRACT-001`, `TC-REM-MACHINE-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.2.5.4` | `TC-PDFUA2-REAL-ORACLE-001` | `TC-PDFUA2-REAL-ORACLE-001`, `TC-REM-MACHINE-001` |
| `dap-jo9.3.1.1` | `TC-GEN-INVALID-DATA-001` | `TC-GEN-INVALID-DATA-001`, `TC-GEN-STATEMENT-001`, `TC-GEN-INVOICE-001`, `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001` |
| `dap-jo9.3.1.2` | `TC-GEN-INVALID-DATA-001` | `TC-GEN-INVALID-DATA-001`, `TC-GEN-STATEMENT-001`, `TC-GEN-INVOICE-001`, `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001` |
| `dap-jo9.3.3.1` | `TC-GEN-STATEMENT-001`, `TC-GEN-INVOICE-001`, `TC-GEN-INVALID-DATA-001` | `TC-GEN-STATEMENT-001`, `TC-GEN-INVOICE-001`, `TC-GEN-INVALID-DATA-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.3.3.2` | `TC-GEN-STATEMENT-001` | `TC-GEN-STATEMENT-001`, `TC-GEN-INVALID-DATA-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.3.3.3` | `TC-GEN-INVOICE-001` | `TC-GEN-INVOICE-001`, `TC-GEN-INVALID-DATA-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.3.4.1` | `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001`, `TC-GEN-SOURCE-FEASIBILITY-001` | `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001`, `TC-GEN-SOURCE-FEASIBILITY-001`, `TC-GEN-VALIDATION-001`, `TC-GEN-ENGINE-GATE-001` |
| `dap-jo9.3.4.2` | `TC-GEN-CORRESPONDENCE-001` | `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-SOURCE-FEASIBILITY-001`, `TC-GEN-VALIDATION-001` |
| `dap-jo9.3.4.3` | `TC-GEN-REPORT-001` | `TC-GEN-REPORT-001`, `TC-GEN-SOURCE-FEASIBILITY-001`, `TC-GEN-VALIDATION-001` |

Two ancillary contract/oracle tasks also gained scoped participation links: `dap-jo9.2.5.3` adds REM machine and GEN validation process/report contributions; `dap-jo9.2.5.4` adds the REM machine authenticated-tool/oracle contribution. These do not claim a family candidate result or human/AT completion. The canonical case owner set remains explicit for all 26 cases. The missing inverse owner link for `TC-GEN-ENGINE-GATE-001` → `dap-jo9.3.4.1` is now present and checked.

Two existing non-owner task links are also now explicitly role-scoped without changing their case sets: `dap-jo9.5.2` participates in four cases owned by `dap-jo9.5.1`, and `dap-jo9.4.1.2` participates in the AFP feasibility case owned by `dap-jo9.4.1.1`. The matrix derives canonical ownership from each case's `Beads` field and does not infer ownership from a missing role.

Task `Test cases` links in Beads now equal catalog `tests` for all 14 mapped records. Existing task `acceptance_criteria`, implementation evidence, closure status, and CI boundaries are preserved from the reviewed baseline, including the substantive adapter and authenticated-oracle evidence for `dap-jo9.2.5.3` and `.5.4`. Only the catalog case states remain specification-only. Participation roles name the canonical owner and limit the contributor's scope. The generated matrix states that linked cases are specifications, not executed results. No business values are recalculated, no tax/rounding rule is introduced, and no unsupported input is accepted by implication.

## Legacy requirement reconciliation

The six affected story records cited **eleven unique** historical requirement IDs. Their active Beads acceptance text now uses the canonical current catalog IDs. [The dated crosswalk](../traceability/legacy-requirement-crosswalk-2026-10-13.md) preserves each historical identifier and its contextual target as proposed continuity; it does not claim proven semantic equivalence. These are the before → after canonical sets:

| Story | Canonical catalog requirements after |
|---|---|
| `dap-jo9.2.3` | `REQ-REM-001`–`REQ-REM-004` |
| `dap-jo9.2.4` | `REQ-REM-005`–`REQ-REM-007` |
| `dap-jo9.3.3` | `REQ-GEN-001`, `002`, `003`, `005`, `006` |
| `dap-jo9.3.4` | `REQ-GEN-001`–`REQ-GEN-006` |
| `dap-jo9.5` | `REQ-X-001`–`REQ-X-003` |
| `dap-jo9.4.1` | `REQ-AFP-001` |

The prior story records' legacy references and source prose remain recoverable from the baseline Beads export; the crosswalk is linked from the traceability index.

## Dependency mapping and health

| Metric | Before | After |
|---|---:|---:|
| Beads issue records / unique IDs | 71 / 71 | 71 / 71 |
| Open / closed / in progress | 36 / 30 / 5 | 35 / 30 / 6 |
| Priority counts P0 / P1 / P2 / P3 | 6 / 42 / 15 / 8 | 6 / 42 / 15 / 8 |
| Parent-child / blocking edges | 69 / 52 | 69 / 52 |
| Total edges | 121 | 121 |
| `bd ready` records | 11 | 10 |
| Duplicate IDs / missing targets / blocking cycles | 0 / 0 / 0 | 0 / 0 / 0 |

All 71 original issue IDs, priorities, and all 121 dependency edges are preserved. `dap-jo9.17` alone moved from open to in progress under the captured claim; no other issue status changed. It remains in progress until independent review and required GitHub CI are accepted. A premature close was corrected by reopening the issue, clearing `closed_at` and `close_reason`, restoring `in_progress`, and recording an honest correction note. The Beads reopen event and history remain preserved. There were no added, removed, or retargeted edges. The broad parent remains open. Safety and release-gate dependencies are unchanged; wider graph rewiring remains deferred.

Before `bd ready` listed `dap-jo9.17` plus 10 existing items. After the claim and correction, `bd ready` lists `dap-jo9`, `dap-jo9.1`, `dap-jo9.1.1.1`, `.1.1.2`, `.1.1.3`, `.1.1.4`, `dap-jo9.10`, `dap-5hy.2`, `dap-jo9.14`, and `dap-5hy.3`; `.17` remains in progress. The available product-path planning work is the statement/invoice table remediation profile and related generation/production/human-AT decisions; none is implementation-ready. The first bounded table-repair experiment remains proposed for a later batch.

The isolated live Dolt export was written to `.beads/issues.jsonl`; its final SHA-256 is `65fd89501ee0a7cfeece4526a4887639273608fc2e511b5e698689faf8910c77`. The before export SHA-256 was `c48b9b0112ece3c6ed60eeb4b3c0b41ed4216d4ad8ae0a93a698c9a8a236ddb0`. The after-state Dolt backup was synced after restoring `.17` to in progress; its manifest SHA-256 is `0fbe6bfab08abff3b2e418d3913bed522dc21438d3777d9573e7056e231c178a`. The final tracker export was reimported into a fresh isolated store and re-exported; all 71 IDs and the status, priority, dependency, description, acceptance, and label fields matched exactly. The broken checkout runtime was not repaired, deleted, or replaced.

## Validation and remaining work

- Root, cloud-tools bootstrap, and CI PDF/UA-2 adapter Go modules passed their prescribed `gofmt`, tidy-diff, vet, and race-test checks. `PDFUA_REAL_CLI_JAR` was empty for local adapter tests.
- Traceability validation passes with 17 requirements, 6 stories, 20 tasks, and 26 specification-only cases. Tests compare task-case obligations with a static reviewed fixture, canonical owners and roles, story requirement sets, Beads task-link parity, and deliberate same-case omission mutations. A synthetic case outside the reviewed fixture proves that every non-owner link requires a valid role and the matrix will not label it as an owner; current cross-family and AFP task links also render their canonical owners. Crosswalk tests validate story/target links within each row and reject swapped targets or deleted story links. The checker enforces catalog/Beads task-case equality and inverse owner links; generated views are current.
- The isolated Beads run passed `bd doctor --check=artifacts`, `bd doctor --check=conventions`, and export/readiness checks. Duplicate issue IDs, missing dependency targets, and blocker cycles are zero.
- `actionlint`, shell syntax, and source-publisher guard tests passed. GitHub CI remains authoritative for vulnerability scans. No local vulnerability database fetch, Java/veraPDF execution, model download/auth work, dependency addition, runtime change, PDF writer, new form fixture, or broad graph rewrite occurred.

Remaining decisions: exact statement/invoice remediation profile values and table topology; source/candidate and preservation boundary; normative structure and validator-control validity; financial arithmetic/tax/rounding; sensitive-document access/privacy/retention; human/AT reviewer identity, authority, protocol, and sign-off; separate generation profile and engine; production workflow and release eligibility. Next batch proposal: implement the actual bounded synthetic table repair value experiment toward PDF 2.0/PDF/UA-2, preserving source values, financial content, table header/data relationships, and reading order, with explicit unsupported cases.
