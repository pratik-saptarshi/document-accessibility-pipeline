# Recovered PDF/UA-2 evidence: planning and validation sequence

**Status: Proposed execution plan.** This document integrates the PDF/UA-2 evidence assessment dated 2026-10-14 with the current product scope and accepted decisions. It does not implement features, approve a new fixture baseline, close a tracker issue, select a writer, or establish conformance or release readiness.

**Status refresh:** checked the repository's `main` source at merge commit `6a41979bf1154b4739d6d0aef90f6ea0eb9470cc` (PR #71). Decision 0004 is Accepted and `dap-jo9.1.1.5` is closed. The late, unresolved PR review comments [r4277123170](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) and [r4277123180](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180) identify two additional review-gate proof requirements; they do not change Decision 0004's accepted status by themselves. The local checkout used for planning remains at the panel's original base revision; the PR and decision status statements below are verified against the exact remote merge commit.

## Evidence baseline and decision boundary

The panel inspected four historical bundles, their contracts/reports, controls, and provenance. The primary bundle is artifact `11848027193`, run `38465226725`, ZIP SHA-256 `2ef4e9601b0cd5de5a12f8e97188fcef22a5e597e3bbd2aa22ee8d692152b42a`. The three recovered supplements are:

| Artifact | Run | ZIP SHA-256 | Retained Library identity | Panel coverage |
|---|---:|---|---|---|
| `11848027193` | `38465226725` | `2ef4e9601b0cd5de5a12f8e97188fcef22a5e597e3bbd2aa22ee8d692152b42a` | `libfile_cf20ce8674b08191af471bfce2ae6b6d` — `pdfua2-reports-11848027193.zip` | Primary panel |
| `11848458063` | `38466062420` | `c172b77f34645299632a0258fd2273478a68bf25ef98e76f18677e53bd7f417b` | `libfile_e89662ada5708191b9ed090f76d8b911` — `pdfua2-reports-11848458063.zip` | Required follow-up; independently inspected |
| `11867387877` | `38505760173` | `6cbcc298210dfe207fde676b556f6b2c64dc7906302b59cce4e01593482df22e` | `libfile_93279f7dbae88191b92a50d916d75a26` — `pdfua2-reports-11867387877.zip` | Independent evidence supplement |
| `11867739001` | `38505929576` | `4fba9c4c14fb5264504059772047f170636eb8ca7f6cc821c877ed8408142ef0` | `libfile_a14015074f9881918de07625d81b2b17` — `pdfua2-reports-11867739001.zip` | Independent evidence supplement |

The panel reports that all four bundles were retained separately; contracts, report hashes, profiles, versions, control outcomes, input identities, and provenance were checked. The newer XML trees match the reports in artifact `11848458063` after removing duration fields; contract differences are report hashes and per-build image IDs. None is a current-main execution. The original panel inspected source revision `b51c01e5e55e8616d40d873c7b4e3d17c6cbd167`; the primary historical CI checkout was PR merge `5eb9a1dae4cc2c6201b10ebdcf294746becc824f`, with source head `6cbbde390c01812dadc2c6bfe1896e320a85ce17`.

This recovery supersedes earlier planning notes that described the required/later reports as unavailable. They are now retained and independently inspected. It does not supersede the still-open formal `dap-jo9.1.6.3`/`dap-jo9.1.6.4` acceptance requirements or turn the recovered historical runs into a current-main execution.

All four bundles consistently identify veraPDF 1.30.3 and explicit PDF/UA-2 plus Tagged PDF validation. Normal execution and pass/finding controls behaved as expected. The two candidate assessments completed **NONCOMPLIANT**: invoice had 74 failed checks, statement had 91, with the six inherited categories below. Successful assessment is not a compliant result; the historical test accepts completed compliant or noncompliant candidate assessment.

The panel's recommendation is limited to a **historical, bounded engineering Proceed recommendation** for the declared row-two `/Headers` association edits, subject to formal evidence reconciliation and acceptance. It is not formal `dap-jo9.1.6.3` disposition or `dap-jo9.1.6.4` closure, candidate eligibility, PDF/UA-2 conformance, general writer feasibility, human semantic approval, AT acceptance, or release authorization. The panel ran no current tests, repair implementation, Java, veraPDF, rendering, or human/AT session, and closed no Beads issue.

PR #71 merged as `6a41979bf1154b4739d6d0aef90f6ea0eb9470cc` (source head `cc2665b2cae0cc0a1835437afd789544811fa5b9`). The merged [Decision 0004](https://github.com/pratik-saptarshi/document-accessibility-pipeline/blob/6a41979bf1154b4739d6d0aef90f6ea0eb9470cc/docs/decisions/0004-approver-persona-review-protocol.md) is Accepted, and the exported issue record marks `dap-jo9.1.1.5` closed on product-owner acceptance. This settles the protocol-design prerequisite and permits `dap-jo9.2.5.9` to consume it. It does not mean a named person has been appointed or has performed a review; Decision 0004 defines a role and requires identity to be recorded per review.

Two later PR #71 inline findings remain open and non-outdated in the review-thread source as checked against the merged PR:

| Review comment | Severity and evidence | Planning disposition and ownership |
|---|---|---|
| [r4277123170](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) | P1: a submitted `approver_identity` value does not prove the submitter held the approver role at `reviewed_at`; an operator or other unauthorized submitter could self-assert. The review asks for role-membership verification at review time plus an immutable authorization reference or equivalent evidence. | Treat as an unresolved candidate-gate acceptance requirement: a future release gate must verify authority at the event time and retain evidence of that verification. The current `dap-jo9.2.5.9` issue has no recorded owner in the exported issue record, and the comment does not assign one. Route policy/source-of-authority selection to the Decision 0004 product owner and implementation to the `dap-jo9.2.5.9` owner once assigned; do not infer an owner or resolve it by changing Decision 0004 in this plan. |
| [r4277123180](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180) | P2: `tool_versions` alone does not identify the exact machine report, its outcome, or the evidence bundle. A materially different report from the same tool version could be substituted while every recorded field remains unchanged. The review asks for immutable report/evidence-bundle binding. | Treat as an unresolved candidate-gate acceptance requirement: retain a stable report identity and digest (or complete evidence-bundle digest), report outcome, and binding to the exact candidate. This records which machine context was presented; it does not make machine results semantic or AT evidence. The comment assigns no owner; resolve record-schema policy with the Decision 0004 product owner and gate implementation with `dap-jo9.2.5.9` after ownership is recorded. |

These comments were made against PR head `144731016a4f802c9f8478cac8fb5c3dfd92f90d`. The review-thread listing marks both current threads unresolved and not outdated. The comments do not establish whether the accepted product owner intends a Decision 0004 amendment or considers them implementation-level acceptance details; that distinction and ownership remain open. Keep actual role proof and immutable report binding as explicit release-gate evidence requirements until an authorized owner records a disposition.

## Finding-to-disposition traceability

The panel classifies each category as inherited from the fixtures under the narrow v2 preservation contract. The plan below records the category as a separate remediation/validation obligation for any future UA-2-ready baseline; it does not retrofit the historic inputs or silently fold these changes into the association experiment.

| Panel finding | Invoice | Statement | Panel assessment | Plan disposition |
|---|---:|---:|---|---|
| Missing title-display preferences | 1 failed check | 1 | Inherited conformance blocker | Add an explicit title-display expectation to a separately versioned proposed baseline; test its presence and candidate-specific validator outcome. |
| Document structure namespace | 1 | 1 | Inherited conformance blocker | Specify the intended structure namespace in the proposed baseline; test correct namespace and a negative/malformed control. |
| Unembedded Helvetica program | 1 | 1 | Inherited conformance blocker; rights/capability need separate evaluation | Treat font choice, embedding permission, and exact candidate/resource behavior as an unapproved feasibility proposal. Require provenance/licensing and embedding-capability evidence before choosing or embedding a font. |
| Unicode mapping | 61 | 78 | Inherited conformance blocker; count is not an impact ranking | Define per-glyph/text expectations for a new baseline. Test exact displayed strings and Unicode extraction/mapping; do not infer that a lower check count means lower user impact. |
| XMP metadata | 1 | 1 | Inherited conformance blocker | Define required metadata for a new baseline and assert metadata both before/after repair. The v2 preservation contract requires XMP absence, so do not edit it retroactively. |
| Untagged/unartifacted content | 9 | 9 | Inherited conformance blocker; source is consistent with grid strokes and a sample heading, not nine broken cell associations | Specify meaningful content versus decorative artifacts in the new baseline. Add structure/artifact controls; do not relabel these findings as association defects. |
| **Total** | **74** | **91** | **Both candidates NONCOMPLIANT** | A future candidate must pass the explicit `ua2` assessment for its exact hash and independently preserve the declared financial semantics. |

The association repair changes the declared row-two `/Headers` relation. The panel's source inspection and recorded object comparison support treating these six categories as inherited. Since v2 explicitly preserves both XMP absence and the font resources, resolving those categories requires a **separately versioned proposed fixture/baseline** with its own allowed-change list and hashes. Do not alter old fixture bytes, baselines, preservation claims, or historical results to make the new work appear to have passed.

## Phases, dependencies, and validation gates

Phases are ordered by evidence dependency, not calendar date. No dates, effort estimates, or performance targets are implied. Work labelled parallel may proceed as planning/reconciliation without changing tracker state or running validators.

### Phase A — reconcile retained evidence and record a finite disposition

**Depends on:** the four retained bundles above; current accepted `dap-jo9.1.6.3`/`dap-jo9.1.6.4` criteria; independent acceptance authority. This phase records/reconciles evidence; it does not reproduce a run. At the verified main commit, `dap-jo9.1.6.3` remains `in_progress` and `dap-jo9.1.6.4` remains `open`.

**Work:** map each bundle to its exact artifact/run/hash, input and candidate hashes, report/contract, profile/version, controls, and provenance. Compare those facts with repository evidence references and the current issue acceptance. Explicitly classify all six findings per candidate as inherited or unexplained, preserving report identities and any disagreement. Then record only a precise disposition for the original bounded attempt.

**TDD/validation intent:** no product behavior change is planned here. If an evidence-index or bundle-reference behavior is later implemented, first add failing tests for wrong/missing hash, report mismatch, profile mismatch, and a completed NONCOMPLIANT report being mistaken for PASS; then make the smallest correction and keep those tests green. Do not add tests merely to close this planning task.

**Exit evidence:** an independently reviewable evidence ledger covers all four bundles, required contracts and findings; exact candidate/input identities agree; all six categories have a reasoned inherited/regression classification; and the original attempt is recorded as Proceed, Reject, or Inconclusive under the actual issue criteria. A bounded Proceed may coexist with assessed inherited NONCOMPLIANT findings, but does not make candidates eligible. `dap-jo9.1.6.4` remains a Proceed-only gate and can close only when its paired operation criteria, complete retained candidate-bound evidence, findings assessment, and independent reviews are met.

**Hard implementation gate:** do not implement or create a new remediation candidate until `dap-jo9.1.6.3` has an independently accepted finite disposition and `dap-jo9.1.6.4` has an accepted Proceed decision under its current criteria. Baseline design and planning may continue while those decisions are open. A Reject or Inconclusive outcome does not unlock the existing candidate slice; follow the accepted contingency through its owner and any coordinated policy/test updates.

**Stop/hold:** stop disposition if an artifact cannot be opened, identities conflict, a required report/contract is absent, provenance is insufficient, a control is invalid, or any unexplained regression remains. Missing evidence is not an invented Reject or Inconclusive experiment outcome. Do not claim current-main execution from these historical bundles. If the actual outcome is Reject/Inconclusive, consult the accepted non-proceed contingency through its current coordinated owner; do not independently rewire issues or reopen a duplicate candidate task in this plan.

### Phase B — propose a separate UA-2-ready synthetic baseline

**Depends on:** Phase A evidence review and the accepted equal-priority statement/invoice scope in Decisions 0002 and 0003. Statements and invoices remain paired and equal priority. Operator confirmation is required for table-header mapping; it does not establish human/AT reviewer authority.

**Work — specification:** define a new versioned synthetic baseline specification for both families. State exact allowed changes for title preferences, structure namespace, font/resource and embedding, Unicode mappings, XMP metadata, and meaningful-content tagging/artifacting. Record exact text and financial values, intended table associations/order, expected render/structure observations, boundaries, unsupported cases, and expected control outcomes. Keep arithmetic, tax, currency, dates, rounding, and production defaults unspecified unless separately accepted. Historical fixture bytes and the v2 contract remain immutable.

**Gate B1 — specification approval:** product-owner review approves the versioned specification, allowed-change list, and test/control definitions. This approves what artifacts should contain; it does not qualify fixture bytes, demonstrate preservation, or establish any validator result.

**TDD/validation intent:** before a baseline generator or checker is implemented, write failing checks that prove legacy inputs and hashes remain unchanged, the new version is distinct, required strings/associations are explicit, and all six baseline conditions are inspectable. A green specification check only proves the plan is represented correctly; it is not evidence that generated PDFs pass.

**Gate B2 — artifact qualification:** after B1, produce the actual versioned synthetic statement and invoice baseline PDFs and record fixture provenance, immutable input hashes, validator/profile/tool versions, limits, and complete exact-byte-hash-bound reports/contracts. Run explicit `ua2` in the isolated CI validator path on each exact baseline PDF; both must return **COMPLIANT**. Then perform a positive no-op on each qualified baseline, compare it independently against its baseline for exact strings, associations/order, structure, font/resources, metadata, extracted text, and rendering, and run explicit `ua2` on each no-op output; both no-op candidates must also be **COMPLIANT**. Preserve the actual report identities and evidence hashes. These are executed qualification results, distinct from B1's approved expectations. Complete B2 before any single-defect repair.

Keep the two COMPLIANT family baselines and their successful no-op outputs immutable. For each later single-defect test, create a separate defect derivative from the exact hash-pinned baseline: record the baseline hash, one declared mutation, allowed object/content diff, derivative hash, and intended control result. Do not mutate or relabel the COMPLIANT baseline to create a defect. A defect derivative is a controlled test input, not a baseline, repair candidate, or evidence of candidate eligibility. For the header-association slice, the derivative changes only the declared `/Headers` relation.

**Exit evidence:** B1 approval is recorded, and B2 has actual retained reports showing COMPLIANT for both baseline PDFs and both no-op candidates, plus independent preservation comparisons and exact hashes. The baselines/no-op outputs remain immutable; every defect derivative has its own source-baseline hash, narrowly declared change, and derivative hash. All allowed changes, preserved values, positive/negative/unsupported controls, and intended relationships are documented. If an implementation uses a font or other asset, its provenance, license, and required capability must be cleared before that use; research may remain unresolved at this phase if the asset is not used. No-op qualification is preservation evidence, not repair evidence.

**Stop/hold:** do not begin single-defect repair if either baseline or no-op output lacks an actual COMPLIANT `ua2` result, if any preservation comparison fails, or if reports/hashes do not bind to the exact bytes. Do not author/repair against a baseline with ambiguous financial content, unconfirmed mapping, unsupported topology, or undeclared changes. Font substitution or embedding is not presumed feasible or approved. If resolving an issue changes the old preservation boundary, create a new baseline version and new hashes.

### Phase C — investigate the inherited blocker families with test-first slices

**Depends on:** B1 specification approval and B2 qualification of actual baseline and no-op artifacts; accepted Proceed decisions for `dap-jo9.1.6.3` and `dap-jo9.1.6.4`; and candidate-specific permissive-license/capability clearance before implementation uses each candidate or asset. Keep Go-first product constraints, the test-only pdfcpu role, and the isolated CI-only Java/veraPDF boundary in Decision 0001. The formal historical decisions must precede implementation, not merely be inferred from recovery of the reports.

**Work:** plan narrow vertical slices using separate hash-pinned derivatives from the immutable qualified baselines: (1) title preferences and structure namespace, (2) meaningful structure versus decorative artifacts, (3) Unicode mapping and exact text extraction, and (4) font/resource embedding. Each derivative introduces one declared defect only; retain its independent baseline/input hash and exact diff. Keep metadata requirements explicit and test them independently. Pre-implementation font/resource/license/capability research may begin before clearance, including in parallel with B1/B2 once exact resource needs are known. Research establishes whether a proposed asset is usable; it is not itself clearance. Do not implement with or embed an asset until candidate-specific permissive-license, provenance, and capability clearance is recorded. Preserve independent semantic/object comparison and rendering evidence.

**TDD/validation intent:** only after B2's actual COMPLIANT baseline and no-op qualification, take one hash-pinned defect derivative at a time. For each proposed behavior, first create a failing test against that derivative that proves the declared defect; implement only the smallest scoped change into a separate candidate; then keep the regression, positive, malformed/negative, preservation, and failure-injection checks green. Verify that each output restores the intended baseline property without unrelated drift. Include a wrong, ambiguous, and unconfirmed mapping case; resource-limit, cancellation, partial-write, and audit-failure controls. Verify exact financial strings, signs, dates, quantities, labels, totals, intended and non-target relationships/order, structure, text, and rendering. No inferred arithmetic or broadened support.

**Exit evidence:** each blocker has an explicit supported treatment and candidate-bound evidence; the immutable baseline and no-op qualification still show COMPLIANT; every derivative has a hash-pinned single-defect diff; single-defect repairs restore the baseline property without unrelated changes; independent preservation comparisons pass; failure/cancellation/partial-write/audit failures cannot emit a successful or retained partial candidate; and licenses/capability are cleared before implementation uses each dependency or resource. Only the isolated CI validator may provide explicit `ua2` results. Every report and metadata disclosure is retained against the exact candidate hash and assessed, including NONCOMPLIANT findings.

**Stop/hold:** stop a slice for any unexplained visual/text/association change, failed control, missing report, resource/license uncertainty, unsupported topology, partial output, or mismatch between the candidate hash and evidence. Record a bounded no-go or unresolved question with evidence; do not expand to untagged/general PDFs or another document family.

### Phase D — bounded paired repair and machine-qualified review-ready candidate decision

**Depends on:** B1-approved baseline specification; B2-qualified statement and invoice baselines plus their COMPLIANT no-op candidates; prior assessed evidence; verified candidate clearance; applicable bounded safety behavior; and reconciled gate criteria. This is the future implementation path associated with the existing paired statement/invoice slice (`dap-jo9.2.3.2`), not new authorization to implement it now.

Following the incremental decomposition proposed in open PR #59, this statement/invoice pair is only the first increment of `dap-jo9.2.3.2`, a six-obligation task. The full task retains `TC-REM-STATEMENT-001`, `TC-REM-INVOICE-001`, `TC-REM-CORRESPONDENCE-001`, `TC-REM-REPORT-001`, `TC-REM-MALFORMED-001`, and `TC-REM-FIDELITY-001`, with canonical ownership retained by `dap-jo9.2.3.1`. Pair evidence covers only the pair increment and its bounded malformed/fidelity controls; correspondence/report and remaining malformed/fidelity obligations remain outstanding. A paired outcome, positive or negative, does not close the full task or transfer case ownership. This planning statement does not imply PR #59 has merged or that the tracker was changed.

**Work:** attempt only the declared operator-confirmed header association repair on its separately hash-pinned defect derivative. The repair changes only the `/Headers` relation, restoring the association declared by the immutable COMPLIANT baseline; it does not make title, namespace, font, Unicode, XMP, or tagging changes in this slice. Keep the derivative immutable and write a separate atomic candidate. Compare each candidate independently against the baseline and derivative, run explicit `ua2` on exact bytes in the isolated CI path, retain full raw reports/contracts, and assess each finding. Phase D produces machine-qualified, review-ready statement and invoice candidates when preservation checks pass and both candidates receive completed COMPLIANT results. This is not full candidate eligibility: genuine semantic and AT outcomes remain for Phase E, and full eligibility follows only after those outcomes and the applicable accepted eligibility standard are satisfied. An engineering Proceed with inherited findings cannot satisfy candidate eligibility.

**TDD/validation intent:** tests first cover the paired positive case, exact preservation, candidate/hash binding, one amount-change mutation, one association mutation, validator pass/finding controls, and malformed, ambiguous, unsupported, cancelled, partial-write, and audit-failure behavior. Green requires two independently inspected candidate results; neither a no-op round trip nor a validator-only pass establishes semantic correctness. Keep the historical repair and no-op characterization tests as characterization evidence, not as proof for the new baseline.

**Exit evidence:** both candidates have separate hashes, verified structural/semantic/render comparisons against the clean baseline and defect derivative, retained explicit-`ua2` reports and contracts, all findings assessed, both family results accepted independently, and no unexplained failure. If both meet the candidate-specific COMPLIANT contract, hand the exact machine-qualified review-ready candidates to Phase E. Do not describe them as fully eligible before genuine semantic and AT review. A one-family success is not a paired success.

**Stop/hold:** any NONCOMPLIANT result, unassessed finding, missing report, mismatch, or family asymmetry blocks candidate eligibility and scope expansion. Do not relabel NONCOMPLIANT as PASS because the assessment job completed. Do not run production documents or claim general writer capability from this bounded pair.

For every future execution, preserve the shared definition-of-done evidence fields: artifact, input hash, candidate hash, evidence hash, tool versions, and limits (or an explicit not-applicable entry with reason for a rejection before output). Record five distinct outcomes: engineering completion, machine validation, human semantic review, AT review, and release eligibility. A result in one field never fills another.

### Phase E — authorized semantic and assistive-technology review

**Depends on:** machine-qualified, review-ready candidates from Phase D and the accepted Decision 0004 protocol. `dap-jo9.1.1.5` is closed at the checked main commit, so protocol design is no longer an open prerequisite. Phase E establishes genuine semantic and AT outcomes before the candidates can meet full eligibility; it remains blocked on closing the unresolved role-proof and immutable machine-report-binding requirements described above. The approver-user persona is a product role; it is not evidence that a named human was appointed or performed a review.

**Work:** obtain genuine authorized semantic review and AT review as separate outcomes on the exact candidate bytes. Follow Decision 0004's immutable review record, accepted protocol revision, candidate hash, profile, and outcome requirements. Add/verify evidence that the acting reviewer held the approver role at `reviewed_at`, and bind the record to the immutable machine report or complete evidence bundle actually presented. Keep machine evidence as context only; it cannot substitute for either review. Define accepted, rejected, missing, escalated, and invalidated outcomes. Candidate-byte change invalidates existing review evidence.

**TDD/validation intent:** test accepted/missing/rejected/escalated/invalidated enforcement with clearly labelled simulations, and verify simulations can never be surfaced as actual human or AT approval. Add failing tests for an unauthorized operator self-asserting the approver role, missing/stale role authorization evidence at `reviewed_at`, and swapping in a materially different report from the same tool version while preserving the record. The green gate must reject each case and verify the retained authorization proof and exact report/evidence-bundle identity against the candidate. Then capture real authorized review evidence separately; synthetic tests do not substitute for it.

**Exit evidence:** Decision 0004's accepted protocol revision is recorded; the actual reviewer identity and immutable evidence that their authority was valid at `reviewed_at` are verified; the exact machine report/evidence bundle is identified and bound to the reviewed candidate; both genuine review outcomes are bound to exact candidate hashes and accepted under the protocol. Together with Phase D's machine/preservation evidence, these results may satisfy the applicable full eligibility standard once that standard is resolved; missing/rejected evidence blocks eligibility and release. This is not implied by machine compliance alone.

**Stop/hold:** if role membership cannot be proved for the review time, the exact machine report/evidence bundle cannot be identified, the accepted protocol revision cannot be verified, or the AT session is unavailable, retain a review-needed/blocked outcome. Do not invent a reviewer, convert operator mapping confirmation into reviewer approval, or claim release readiness.

### Phase F — evidence-led scope review

Expansion is a separate product decision after Phases A–E. Consider only capabilities demonstrated by accepted evidence and explicit profiles. Correspondence/report remediation, broad untagged-input support, generation, deployment, and IBM AFP remain separate scopes/tracks. Generation retains its own source-conversion and engine gates. Do not extrapolate from two synthetic financial tables to general documents.

**Exit evidence for any proposed expansion:** a named capability, accepted input/profile boundary, synthetic/deidentified cases, test-first plan, exact evidence contract, negative/failure controls, dependency/license review, and human/AT gate. Without these, the expansion remains proposed or open.

## Cross-cutting integration prerequisites and parallel lanes

These coordination issues affect later execution but are not edited or resolved by this document:

- Reconcile PR #59 at exact head `406b430` and its three-blocker policy before relying on its gate wiring. PR #70 introduced `dap-jo9.1.1.5` as a prerequisite; that issue is now closed through PR #71's accepted Decision 0004. Carry the unresolved role-proof and machine-report-binding findings into `dap-jo9.2.5.9` acceptance work without treating them as a reopened `.1.1.5` issue. The reviewer persona remains a role, not a named-human appointment.
- Coordinate any future `dap-jo9.1.6.4` re-gating with the applicable policy and tests. Do not alter gate meaning or tracker dependencies as part of this evidence plan.
- Keep explicit distinctions among cancellation, timeout, successful completion, and failure; the proposed work must not interpret cancellation as completed assessment or success.
- The separate `dap-jo9.2.4` parent/semantics ownership correction requires its own owner and acceptance treatment; do not fold it into this evidence disposition.
- [PR #59](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/59) is open and unmerged at head `406b430d69035050b5029903451d5b0f0d7a2d2f`. It proposes clarifying `dap-jo9.2.3.2` as six obligations delivered in reviewed increments, with the statement/invoice pair only the first increment; this plan follows that proposed scope but does not claim it is merged or alter the issue. PR #59 also proposes administrative retirement of `dap-jo9.2.5.10` as a duplicate/superseded issue, with no implementation or acceptance claim. On the checked main source `dap-jo9.2.5.10` remains open and its eligibility-standard question remains unresolved; do not call it retired or treat full eligibility as defined unless an authorized source decision is merged. The existing contingency and issue ownership require coordinated reconciliation; this plan does not change tracker state.
- `dap-jo9.1.6.9`, `dap-jo9.5.4`, and `dap-jo9.1.6.8` are outside this document's edit scope. Consult their owners/current accepted records where relevant; do not duplicate their work or change their status here.

Evidence reconciliation, baseline specification, and the above coordination review can proceed as distinct planning lanes. Actual baseline qualification requires produced artifacts and executed candidate-bound validation; it is not satisfied by an approved spec or planned hashes. Candidate changes and candidate-specific validation remain serial behind evidence review, baseline qualification, and candidate clearance. Font feasibility research may run before clearance once exact resource needs are known; use of an asset in implementation waits for its clearance. Phase D outputs machine-qualified review-ready candidates only; Phase E records genuine semantic and AT outcomes before any full-eligibility determination under the resolved standard and accepted protocol.

## Planned repository validation if implementation is later authorized

These are execution plans for future code changes, not results of this planning task. Run the applicable checks on the exact implementation revision and record their outputs with candidate/artifact hashes. Do not install dependencies or execute Java/veraPDF locally. Follow the repository boundary that Java and veraPDF run only in the isolated validator-container CI job; the Go adapter can be tested without a real CLI.

For an implementation revision, the repository checks are:

```sh
# Root module
test -z "$(gofmt -l .)"
go mod tidy -diff
go vet ./...
go test -race ./...
go run ./cmd/traceability-check

# Standalone cloud-tools bootstrap module
(cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./... && go mod tidy -diff)
bash -n scripts/bootstrap-cloud-build-tools.sh

# CI-only PDF/UA-2 adapter module; process-contract tests only
(cd ci/pdfua2/adapter && PDFUA_REAL_CLI_JAR= go test -race ./... -count=1 && go vet ./... && go mod tidy -diff)

# Source-publisher guard tests
bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh
bash .github/scripts/test-source-publisher.sh
```

These checks do not replace the future candidate-bound isolated `ua2` run, preservation comparison, actual human semantic review, or actual AT review. This planning change itself did not run the commands above or execute any validator.

## Validation and completion record for this planning change

This is planning-only documentation. Completion checks are: (1) all six findings and both per-family totals appear in the disposition map; (2) the four retained bundle identities and the distinction between historical engineering Proceed, formal gate closure, compliance, eligibility, and release are explicit; (3) the immutable COMPLIANT baseline, B1/B2 gates, and hash-pinned defect derivatives are distinct; (4) `.2.3.2` retains all six case obligations and the pair is only its first increment; (5) Phase D stops at machine-qualified review readiness and Phase E records genuine semantic/AT outcomes before full eligibility; (6) PR #71 status and its two unresolved proof findings, plus PR #59's open/unmerged proposals, are represented without inventing ownership; (7) historical `.1.6.3`/`.1.6.4` decisions are hard prerequisites to new candidate implementation; and (8) the document is linked from the planning index. No Go tests, PDF validators, Java, rendering, Beads commands/mutations, or implementation checks are authorized or represented as run for this plan.

**Evidence access note:** the complete 73-line Library panel report was read using its exact identity, version `1`, file ID `file_000000008e2481f596a979dca148e62a`, Library file ID `libfile_8585efc3c1448191ba326f0aa9daf1fe`, and exact name `pdfua2-evidence-panel-2026-10-14.md`. The supported local materialization helper failed twice with `download failed`; no local report file was created. The planning content was cross-checked against the complete Library read and repository sources, but local-file materialization remains unverified.

## Source records

- [Product scope](../requirements/product-scope.md), [product roadmap](../roadmap/product-roadmap.md), and [Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md), [Decision 0002](../decisions/0002-initial-remediation-family-priority.md), and [Decision 0003](../decisions/0003-operator-confirmed-financial-table-mapping.md).
- [Next-step decision register](next-step-decision-register-2026-10-13.md), [v2 repair evidence](../research/pdfcpu-financial-table-v2-repair-2026-10-13.md), [backlog panel](backlog-panel-2026-10-14.md), and [v2 non-proceed contingency](v2-non-proceed-contingency.md).
- [Decision 0004](https://github.com/pratik-saptarshi/document-accessibility-pipeline/blob/6a41979bf1154b4739d6d0aef90f6ea0eb9470cc/docs/decisions/0004-approver-persona-review-protocol.md), [PR #71](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71), and unresolved [review comment r4277123170](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) and [review comment r4277123180](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180); status checked at merge commit `6a41979bf1154b4739d6d0aef90f6ea0eb9470cc`.
- [PR #59](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/59) is referenced as an open proposal at exact head `406b430d69035050b5029903451d5b0f0d7a2d2f`; its status is not represented as merged.
- Panel evidence: Library artifact/run table above; source report `pdfua2-evidence-panel-2026-10-14.md`, version `1`.
