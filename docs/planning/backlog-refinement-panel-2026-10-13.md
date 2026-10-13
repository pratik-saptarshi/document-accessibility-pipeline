# Backlog Refinement Panel — 2026-10-13

**Status: Proposed recommendations; no tracker or roadmap decisions adopted.** This report records a six-role planning review and judge synthesis. It does not change Beads, the traceability catalog or generated matrix, the canonical roadmap, product scope, or issue status. Implementation remains subject to the relevant user decisions and evidence gates.

## Executive synthesis

The repository has a credible planning foundation: remediation and generation are distinct lanes, AFP is separate, the four document families are named, and automated validation is not treated as a substitute for human semantic and assistive-technology review. The backlog graph is structurally intact. The largest planning risks are an unselected first user workflow, work whose acceptance criteria are still too descriptive to test, stale requirement links in story records, and open tasks that may already have implementation evidence but lack a verified closeout.

The panel recommends a short, decision-led sequence: establish the first user workflow and lane profile; finish a finite, evidence-based remediation feasibility gate; move only a narrow, real structure-changing remediation slice into a candidate/review flow if that gate passes; keep generation feasibility parallel and its production lane separate; and defer broad family expansion and release hardening until evidence justifies them. PR #38's checkbox appearance-state work is useful bounded spike evidence, not writer selection or product approval.

**First-use-case question for the product owner:** Which first end-to-end workflow should the team prioritize: remediating an existing PDF or generating a new one? If remediation, which document family and must-have features or exclusions should define the first slice? If generation, what source and data path should it cover? The answer should select a workflow and constraints; it need not select a library.

No production workflow, product engine, writer, first family, release protocol, or user outcome has been demonstrated. Accordingly, statements such as “reduce review time” remain hypotheses until a baseline and measured outcome are agreed.

## Evidence baseline and limits

| Evidence | Observed baseline | Interpretation |
|---|---|---|
| Main/review cutoff | `163919bca4247c3f767af23e3d6aa669f10dae1b` (2026-10-13) | Includes PR #38, `test(pdfpreserve): cover checkbox appearance states`. |
| PR #38 | Owner-reported accepted result: 9 pre-merge and 7 post-merge checks passed | Test-only pdfcpu spike evidence for checkbox `/AP /N` state dictionaries and `/V`/`/AS` bindings. It does not select a production writer or demonstrate PDF/UA conformance. This panel did not rerun those checks. [PR #38](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/38) |
| Issue export | 66 issue records: 32 open, 28 closed, 6 in progress; 116 dependency edges (64 parent-child, 52 blocks) | Export-derived snapshot; it is not a full Beads database backup or complete event history. |
| Graph integrity | No duplicate IDs, missing dependency targets, or blocker cycles in the exported snapshot | Structural integrity is sound at this cutoff. |
| Unfinished work | 38 issues are not closed; 24 have an unresolved direct blocker and 14 have none | “No unresolved direct blocker” is an export-derived observation, not a `bd ready` result or proof that an issue is actionable. Some are parents or active work. |
| Active feasibility work | `.jo9.1.1`, `.jo9.1.2`, `.jo9.1.3` are in progress and have expired leases at the cutoff | The export shows assignments and lease timestamps, not whether work is abandoned. Reconcile ownership/heartbeat before reassigning. |
| Traceability cases | 20 task rows, 6 story rows, 26 case specifications | These are specifications, not 26 executed product tests. The rendered matrix includes linked task cases in its story view. |
| Fixture manifest | Four planned fixture rows, none created or run | These manifest entries are distinct from inline test fixtures. |
| CI and validation | PR #38 is reported green; veraPDF integration is isolated CI work | Do not infer local access to CI-only Java/veraPDF, product outcomes, or conformance. The permanently accepted Neuro model blocker is not a work item to revisit. |

The graph audit and repository inspection for this report were read-only. This report contains planning analysis; it is not a test report. No issue, catalog, matrix, or canonical roadmap was edited.

### Suggested health checks for the next refinement review

These are proposed targets, not claims about the current backlog. Retain zero duplicate or unknown IDs, zero dangling dependency references or cycles, and zero unresolved mapping-owner conflicts. Preserve all 66 issue IDs and 116 edges unless a reviewed decision justifies a specific change. After decisions are recorded, aim for at least one genuinely actionable product leaf to appear in live `bd ready`; confirm with Beads rather than inferring readiness from this export. Report blockers separately from incomplete work and from work awaiting CI or human review.

## Connection to the original goal

The initial repository description was “Enterprise document generation and remediation accessibility pipeline.” Current product scope turns that into two distinct lanes—existing-PDF remediation and new accessible-document generation—across statements, invoices, correspondence, and reports, with AFP as a separate feasibility track. The accepted PDF/UA-2 target and Go-first/permissive-dependency constraints are useful boundaries, while writer, generation engine, workflow, profiles, reviewer authority, and release eligibility remain unresolved.

The goal can be expressed as four user outcomes to validate: produce a useful candidate; preserve content and meaning; make unsupported, uncertain, and review-required cases visible; and reduce an agreed, measured risk or effort. Current evidence supports planning and synthetic feasibility only. It does not yet show a user can complete either lane, a production transformation is safe, review effort falls, or a conformance claim is justified.

## Panel recommendations and judge decisions

| Role | Recommendation | Judge disposition and rationale |
|---|---|---|
| **Backlog refinement** | Make profile decisions finite and lane-scoped; define executable acceptance tests before implementation; keep the writer experiment bounded; specify early safe-failure behavior and a separate human/AT review handoff. Extend `.jo9.9` only if an audit finds missing behavior; its stated matrix and whitespace behaviors already have tests. Create a separate mapping-integrity successor only after auditing old story links. | **Accept** the scoped profile, finite experiment, executable DoD, and separated review gates. **Modify** the `.jo9.9` proposal to audit existing evidence first. **Defer** any new issue until the audit demonstrates a gap. |
| **Elicitation** | Ask the first-use-case question above. Resolve only decisions needed for the next gate; avoid assuming correspondence, no forms, no signatures, or no scans. Engineers should verify candidate capabilities and licensing rather than asking the user to choose a library. | **Accept** first-use-case elicitation and just-in-time decisions. **Defer** family and feature defaults until the product owner answers. |
| **Product owner** | Define the first user journey from operator intake through candidate generation, exception/review, and disposition. Require an actual repair feasibility result rather than more no-op round-trips. Define reviewer authority and outcome measures early; retain generation as a separate lane. | **Accept** the journey framing and measured outcomes. **Modify** the proposed correspondence-first choice: it is not approved and should follow the first-use-case answer. |
| **Roadmapping** | Order work by evidence and decision gates, with fixture expectations before experiments, explicit phase entry/exit criteria, and early safety separated from later operational hardening. | **Accept** outcome-based phase gates. **Modify** the proposed arbitrary “one corrective iteration” cap into an effort boundary with explicit stop/go rules. **Defer** a fixed cycle count or duration until estimates exist. |
| **Prioritization** | Finish profile/corpus/candidate clearance and real repair evidence before broad implementation; keep generation feasibility secondary and parallel where independent; preserve PR #38; defer broad forms/family expansion. Use `3V + 2R + 2D - 2E - U` only as an illustrative ordinal discussion aid, never as ROI or measured value. | **Accept** the ordering. **Modify** the score's status to illustrative only. **Defer** expansion until a first slice demonstrates value and constraints. Audit `.jo9.9` and `.jo9.10` before proposing follow-up tasks. |
| **Judge** | Synthesize the recommendations as reversible, evidence-led refinement: choose the first workflow, gate real remediation on feasibility, correct traceability only after audit, distinguish blocked from incomplete, and keep CI authority and review evidence explicit. | **Accept** the staged approach and existing accepted boundaries. **Modify** only where evidence calls for a narrow audit or where a proposed family/cycle count was premature. **Defer** user-dependent scope and unsupported defaults. |

### Issue-level refinement proposal (not applied)

1. **Reconcile active feasibility work.** Review status and ownership for `dap-jo9.1.1`, `.jo9.1.2`, and `.jo9.1.3` before extending leases or changing assignees. Keep `.jo9.1.1` as a rollup; consider scoped children for remediation profile, generation profile, real-document handling, and semantic/AT authority only after the first-use-case decision.
2. **Close the evidence loop on `.jo9.1.2`.** Under `.jo9.2.1`, define the initial synthetic remediation fixture and baseline before the next experiment. The checkbox-state work in PR #38 is one preservation slice. A go/no-go experiment should make a real structure change and compare source and candidate independently for content, rendering, structure, metadata, and relevant interactive elements; include negative controls and immutable-input evidence. Bind CI `ua2` output to the exact candidate hash, report supported/unsupported/reject/inconclusive outcomes, and keep human semantic/AT review separate. A no-op PDF 1.7 round-trip is insufficient. Keep `.jo9.1.3` focused on exact candidate versions, licenses, capabilities, and build/runtime/assets clearance rather than an open-ended tool survey.
3. **Audit traceability ownership.** Six story records appear to cite requirement IDs absent from the current catalog: `.jo9.2.3` (`REQ-REM-INGEST-001`, `REQ-REM-FID-001`, `REQ-SCOPE-001`); `.jo9.2.4` (`REQ-REM-SEM-001`, `REQ-VAL-HUMAN-001`); `.jo9.3.3` (`REQ-GEN-DATA-001`, `REQ-GEN-OUTPUT-001`, `REQ-SCOPE-001`); `.jo9.3.4` (`REQ-GEN-OUTPUT-001`, `REQ-GEN-SOURCE-001`, `REQ-VAL-HUMAN-001`); `.jo9.5` (`REQ-CROSS-SEC-001`, `REQ-CROSS-AUDIT-001`); and `.jo9.4.1` (`REQ-AFP-FEAS-001`). Verify whether these are intentional legacy references, map to current requirement IDs with owner review, or need a clearly labeled successor. Do not invent replacement IDs or rewrite history. If a gap is confirmed, propose one P1 mapping-integrity task with case-level evidence and preserved issue IDs/edges.
4. **Audit candidate-case ownership, not just matrix output.** Review the cases and owner/shared-evidence intent for `.jo9.2.3.2`, `.jo9.2.4.1`, `.jo9.2.4.2`, `.jo9.3.3.1`, `.jo9.3.3.2`, `.jo9.3.3.3`, `.jo9.3.4.1`, `.jo9.3.4.2`, `.jo9.3.4.3`, `.jo9.3.1.1`, `.jo9.3.1.2`, and `.jo9.2.5.2`. Do not copy all cases into every story: distinguish the task that owns executable evidence from stories that consume or summarize shared evidence. The current generator already unions linked task cases into the matrix, and human/AT cases are represented; story-local omissions alone are not a rendered-matrix defect.
5. **Rewrite thin implementation DoD where appropriate.** The panel flags `.jo9.2.3.2`, `.jo9.2.4.2`, `.jo9.3.1.2`, `.jo9.3.3.2`, `.jo9.3.3.3`, `.jo9.3.4.2`, `.jo9.4.1.2` (conditional AFP work), `.jo9.3.4.3`, and `.jo9.5.2` for refinement into executable behavior, pinned fixtures/results, boundaries and negative controls, evidence/review expectations, and applicable CI. Keep engineering completion distinct from release approval. Leave closed specification issue `.jo9.2.5.2` closed; do not reopen it to duplicate human/AT gate definition.
6. **Audit possible completed work before adding tasks.** `.jo9.9` acceptance asks for test failures if task cases are omitted from story views and for ASCII/Unicode whitespace-only profile values to be rejected. `TestGeneratedStoryMatrixIncludesLinkedTaskCases` and `TestLaneSchemasRejectWhitespaceOnlyDecisionValues` cover those behaviors. `.jo9.10` asks for a bounded CI timeout and policy rejection of missing/invalid values; `go-checks.yml` has `timeout-minutes: 20`, and policy tests exercise missing/out-of-range values. Verify required workflow-run evidence (including a disposable stalled-job cancellation if required by `.jo9.10`) and review issue scope before closing or creating follow-up work. Do not treat code presence alone as all acceptance evidence.
7. **Preserve useful priorities and conditionals.** Statements before invoices may be a sequencing preference, not a dependency edge. Consider decoupling DOCX/Markdown source feasibility from data/template work only if evidence and ownership permit. Keep `.jo9.2.2` as a minimal consumer-driven adapter contract, not a generic framework; specify early review-needed handoff in `.jo9.2.4.2` separately from later automatic semantic repair; ensure output approval is bound to a candidate hash. Keep `.jo9.5.2` focused on early bounded safe failures and create no duplicate of `.jo9.5.1`'s cross-lane policy unless its acceptance gap is confirmed. Keep AFP's input/output, resources, fidelity, and owner decision independent.

## Proposed delivery sequence

These phases recommend ordering; they do not replace the canonical roadmap or authorize implementation.

| Phase | Focus | Exit evidence |
|---|---|---|
| **0 — Backlog integrity and first scope** | Answer the first-use-case question; reconcile active leases; audit story-to-requirement references, case ownership, and apparently implemented `.jo9.9`/`.jo9.10` acceptance. Agree on the initial profile, privacy/arithmetic boundaries, reviewer role, and fixture-handling rules. | Owner-approved first workflow and constraints; verified mapping/evidence audit; scoped issues with executable acceptance and an accurate dependency graph. |
| **1 — Remediation feasibility gate** | Create a small, synthetic baseline corpus; verify exact permissive candidate versions/licenses; run the bounded, structure-changing writer experiment. In parallel, gather generation source/engine evidence where independent. | Reproducible preservation, mutation, negative-control and CI `ua2` evidence; explicit limits and go/no-go; exact candidate hash; review needs and open decisions recorded. Failure returns to alternatives/scope instead of automatic tool selection. |
| **2 — First remediation candidate flow** | If Phase 1 passes, implement only the chosen input-to-candidate slice, with immutable source, bounded operations, audit evidence, explicit unsupported/review state, and reviewer handoff. | User can complete the agreed synthetic first journey; preservation and failure behavior are evidenced; human/AT review is recorded separately; candidate changes invalidate prior approval. |
| **3 — First generation slice** | After its profile/source/engine decisions, deliver the selected family and data/template path as a separate lane. A statement-first sequence remains a proposed default pending user scope. | Validated inputs and semantic structure, deterministic output, family-specific tests, PDF/UA-2 machine evidence and separate human review. |
| **4 — Family/source expansion** | Add the remaining statement, invoice, correspondence, and report needs; broaden source handling and semantic cases only from demonstrated demand. Keep AFP on its own decision track. | Each expanded case has an agreed user, fixture, acceptance evidence, and explicit support boundary. |
| **5 — Release operations and broader safety** | Complete cross-lane resource, isolation, retention, audit, privacy, security, performance, accessibility review, and operational support gates. | Reviewed release protocol and measured outcomes; no machine-only or automated claim substitutes for human semantic/AT acceptance. |

Optional scaling work `.jo9.14`, `.5hy.3`, and `.5hy.2` should wait for a demonstrated trigger or measured need. Do not weaken safety, skip the feasibility gate, or introduce a custom-engine fallback to accelerate a phase.

## Decisions and evidence still needed

| Decision | Current status | Needed next |
|---|---|---|
| PDF/UA-2 output target, explicit `ua2` profile, Go-first approach, permissive product dependencies, veraPDF CI-only boundary | Accepted in Decision 0001 | Preserve as constraints; do not overstate what acceptance proves. |
| First user workflow and first family/features | Open | Product-owner answer to the first-use-case question. |
| Remediation input/profile and supported structure changes | Open | User constraints plus synthetic corpus and feasibility evidence. |
| Writer/candidate and generation engine | Open | Exact capability, version, license, fidelity and integration evidence; no user library selection required. |
| Reviewer authority, protocol, and approval invalidation | Open | Define who may approve which findings and bind approval to immutable candidate identity. |
| Arithmetic, privacy, retention, and real-document handling | Open | Resolve just in time for the selected workflow, before sensitive inputs or retention are enabled. |
| Human semantic and assistive-technology protocol | Required separate gate; details open | Agree reviewer, protocol, evidence, and outcome handling; never simulate a pass. |
| Neuro-assisted testing | Permanently blocked/accepted | Do not revisit model downloads or claim Neuro-assisted testing. |

## Migration and rollback guidance for any future tracker change

This report authorizes no tracker mutation. If the user later approves a refinement batch:

1. Record the reviewed repository head, issue-export hash, and a supported full Beads/Dolt backup. The JSONL export is useful for comparison but is not a full backup or history.
2. Present the exact affected issue IDs, proposed text, status changes, and edge changes for review. Keep existing IDs; avoid blanket reparenting and duplicate work. Add a successor only when an audited gap is established.
3. Apply a small, reversible batch, then compare before/after issue counts, all IDs, statuses, notes, edge counts and endpoints, blocker DAG, catalog references, and generated views. Check live `bd show`/`bd ready` as well as the export.
4. If incorrect, revert the repository change and apply reviewed inverse tracker updates. An export upsert cannot remove a newly created issue or restore Dolt history; preserve concurrent work and use the supported database backup for recovery. Do not delete or replace a runtime database to resolve an incomplete environment.

## Recommendation

Keep this panel as a proposed refinement plan. First obtain the first-workflow answer, then validate the narrow audits and feasibility evidence before changing issue state or scope. Keep all open issues open until their own acceptance evidence and any required workflow/review evidence have been checked.
