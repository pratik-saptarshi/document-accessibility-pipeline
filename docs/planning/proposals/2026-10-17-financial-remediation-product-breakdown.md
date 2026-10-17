# Proposed product breakdown: financial remediation pair

**Status:** Product refinement accepted for planning quality. The exact nine-issue/ten-text-field batch and its one-time approved note-tail/timestamp extension are recorded below. The extension changed only the three named note endings and updated_at on the same nine issues. It does not authorize source selection, artifact authoring, implementation, status changes, graph changes, issue creation, or closure.

**Reviewed source:** e15b352e398581bef24aeeb679e45efdf442286d (2026-10-17). Pre-change source JSONL SHA-256: 441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec. Prior text-only candidate JSONL SHA-256: 9d5438fa5b046dd44ca32da878d434cb548b8eff897934762abc7ac8b19566a8. Current candidate after approved note/timestamp extension JSONL SHA-256: 4e3fede04c2133e3bdba04e48fa13e703415e991d3e831cae850212bf52bd095.

## Product hierarchy

| Level | Existing mapping | Boundary |
|---|---|---|
| Capability | `CAP-REM-PRESERVE`, supported by `CAP-REM-INTAKE` and `CAP-REM-REVIEW` | Preserve characterized content while applying only an approved bounded structural change. |
| Feature grouping | **Operator-confirmed financial-table repair with preservation evidence** | Proposed display label only. There is no canonical Feature entity or stable ID; do not invent a Beads type or issue. |
| User story | `US-REM-001` / `dap-jo9.2.3` | Bounded repair of supported existing PDFs. The operator persona and authority are still subject to existing decisions. |
| Acceptance ownership | `dap-jo9.2.3.1` owns specifications; `dap-jo9.2.3.2` implements them | Keep all six cases: statements, invoices, correspondence, reports, malformed input, and fidelity/preservation. The first pair is one increment only. |
| Core functions | `FN-REM-INTAKE`, `FN-REM-CHARACTERIZE`, `FN-REM-REPAIR-BASELINE`, `FN-REM-COMPARE` in `US-REM-001` | Each is a requirement responsibility, not a mandate for a separate package, service, or API. `FN-REM-VALIDATE-OUTPUT` belongs to `US-REM-002`; it contributes output-validation evidence without silently changing the story/function mapping. |
| Existing tasks and gates | `.2.3.1`, `.2.3.2`, `.2.2`, `.1.6.3`, `.1.6.4`, `.2.5.9` and `.9.1`/`.9.2` | `.1.6.3` and `.1.6.4` are historical evidence gates, not implementation subtasks. `.2.2` specifies a consumed interface. `.2.5.9` is an actual review gate; `.9.1` tests labeled simulations and `.9.2` verifies trusted authority. |

### Proposed acceptance decomposition

These are proposed details for review, not claims that existing cases already specify or execute them.

1. **Intake:** accept only exact approved synthetic inputs under declared bounds. Preserve originals; malformed, encrypted, unsupported, or oversized inputs cannot become successful candidates. Functions: intake and characterization. Owner: existing malformed/family cases under `.2.3.1`; implementation remains `.2.3.2`.
2. **Positive baseline:** after exact owner approval of a v3 source, six-category object/value/resource allow-list, and license clearance, qualify both positive baselines and no-op outputs with exact financial/text/structure/order preservation and candidate-bound explicit-`ua2` COMPLIANT reports. Functions: characterization and comparison. This proposed qualification does not rewrite historical v2 evidence.
3. **Paired repair:** require the operator to confirm each exact row-two amount-header mapping. Produce separate candidates; change only the named relation. Independently reject wrong relations even if the machine validator passes. Functions: baseline repair and comparison. Cases: existing statement/invoice cases, whose canonical owner remains `.2.3.1`.
4. **Fidelity:** compare fresh reads, text extraction, structure, rendering, resources, and immutable inputs against observations independent of the writer. An unapproved delta or missing evidence fails closed. Case: existing fidelity case, with remaining whole-story scope still open.
5. **Machine evidence:** bind validator/version, explicit profile, exact candidate hash, report identity, process result, and assessment. Missing, mismatched, fallback, contradictory, malformed, or failed evidence cannot pass. `FN-REM-VALIDATE-OUTPUT` remains under `US-REM-002`; `TC-REM-MACHINE-001` remains owned by `.2.5.1`.
6. **Whole-story completion:** correspondence, reports, and any remaining malformed/fidelity criteria stay outstanding until separately reviewed increments satisfy all six obligations. Pair success, Reject, or Inconclusive does not close `.2.3.2`.

Semantic review, assistive-technology review, eligibility, authority, and release remain distinct. `.2.5.2` owns review specifications; `.2.5.9.1` is labeled-simulation testing; `.2.5.9.2` verifies trusted authority after `.1.1.6`; `.2.5.10` is the candidate eligibility gate. No synthetic result or machine pass substitutes for a genuine review. The applied text records that documenting a missing eligibility standard is a blocker only and cannot satisfy `.2.5.10` candidate-production closure. This is a text refinement, not a new eligibility standard or candidate decision.

## Milestones and proposed new task

- **M0 historical audit:** reconcile `.1.6.3` and `.1.6.4` against Decision 0005, Decision 0011, the retained-bundle ledger, and exact candidate/run/report identities. Independent product and technical reviewers have accepted the bounded historical repair Proceed; M0 still needs to reconcile its exact evidence and record identities. Both v2 candidates remain NONCOMPLIANT. No rerun, v3 authoring, eligibility claim, or tracker transition is implied.
- **M1 guarded recording:** the current path cannot close .1.6.3/.1.6.4; see the probe evidence below. The one-time text batch and narrow note/timestamp extension are documented below. No status transition or closure occurred; future changes still require a supported path or their own explicit scoped authorization. Do not expand frozen tooling.
- **M2 decision packet:** prepare options A/B/C, exact source pins, six change categories, resource/license evidence, and a narrow prerequisite proposal. A (copying pinned repaired v2 candidates as immutable v3 source material) is recommended but unapproved. No PDF is generated until owner approval.
- **M3 baseline qualification:** proposed task below. Preserve v1/v2 hashes; add independent financial/text/structure/render oracles, known-pass/finding controls, category-specific negative controls, and positive/no-op explicit-`ua2` results for both families.
- **M4 consumed Go contract:** refine `.2.2` only around fields and outcomes shown to be consumed. Cover exact input/output identity, operator-confirmed operation, preservation evidence, cancellation/deadline, invalid/unsupported/tool-failed/partial/audit-failed outcomes. No generic workflow framework.
- **M5 pair execution:** `.2.3.2` remains governed by its current direct blockers (`.1.3`, `.1.6.4`, `.2.1`, `.2.2`, `.1.1`) unless the owner approves and a supported guard applies a graph change. Require both families, independent relation and preservation checks, negative controls, and candidate-bound evidence. No broadening or whole-task closure.
- **M6 review/eligibility:** keep genuine semantic/AT acceptance, trusted authority, and candidate eligibility distinct and bound to exact candidate bytes and evidence.

### One possible task (no allocated Beads ID)

**Title:** Qualify paired v3 financial baselines. **Type:** task. **Priority:** proposed P1. **Intended parent:** `.2.1`, pending owner approval and a supported create/graph path. This is a proposal only; no issue ID or dependency edge is invented.

**Planning acceptance:** prepare exact options A/B/C; recommend A without treating it as accepted; list unresolved object/value/resource/license decisions, source pins required, independent financial oracle, validator/control plan, provenance requirements, and a proposed prerequisite split. No artifact/hash may be fabricated. Completion of this planning phase is not completion of baseline qualification.

**Qualification acceptance:** only after owner-approved exact inputs, resources/licenses, and supported prerequisites, author separately versioned baselines, record actual hashes/provenance, preserve v1/v2, pass independent financial/text/structure/render checks, and obtain complete candidate-bound positive and no-op explicit-`ua2` COMPLIANT evidence for both families. For each of the six approved categories, define a validator-visible PDF mutation and owner-reviewed category-to-expected-finding map before execution. Mutate one category at a time and require a completed explicit-`ua2` NONCOMPLIANT report bound to the exact mutated candidate hash, with at least one failed rule/check linked to that category. An unrelated finding alone is insufficient. Assertion tests must reject COMPLIANT mutation XML, hash mismatch, missing report, and process/validator error as satisfying the negative control. No fixed failure count is required. Test external license/provenance evidence separately with the provenance oracle; do not represent evidence-only changes as validator mutants. Separately prove the relationship oracle rejects the wrong association even when machine validation passes.

**Evidence:** reviewed decision packet, exact source/output hashes, resource licenses/notices, independent observations, full reports/contracts, controls, and mutation results. **Exclusions:** repair execution (`.2.3.2`), broad corpus (`.2.1.3`), new case ownership, production-engine selection, automatic inference, generation, human/AT acceptance, and release. Leave owner unassigned until expressly assigned. Defer issue creation until exact graph treatment and guarded creation are confirmed.

## Exact tracker batch and limits

The companion [JSON batch](2026-10-17-financial-remediation-updates.json) records ten approved text fields across nine existing issues, plus updated_at on those same nine issue records. The approved extension replaces only the terminal sentence in notes for .1.6.3, .1.6.4, and .2.5.9, clarifying that those text updates are applied while issue statuses remain unchanged; .1.6.3 also states that its historical repair/candidate evidence exists and bounded Proceed was accepted, while its status transition remains pending and no v3 authoring/future implementation is authorized. The [one-time exception record](../../reviews/financial-remediation-one-time-text-update-exception-2026-10-17.md) and [digest-only delta manifest](../../reviews/evidence/financial-remediation-text-batch-delta-2026-10-17.json) preserve authorization scope and verification evidence. No status, ID, type, priority, owner, parent, dependency, or edge changed; updated_at changed only on the nine target issues. The .1.6.3/.1.6.4 values record historical evidence and gate boundaries; the other seven issue updates refine planning and acceptance wording. The .2.5.10 text makes documenting a missing eligibility standard a blocker record that cannot satisfy candidate-production closure. Product review supported that clarification, and the user authorized this exact text batch and narrow extension; this is not a new eligibility standard or source/implementation decision. The direct edit is a one-time exception and creates no continuing authority. Historical closure and future implementation/acceptance remain unapplied or unapproved.

The proposal preserves the seven `.1.6.4` downstream consumers and all six `.2.3.2` obligations. It makes no graph change and does not create the proposed task.

### Current guarded-path probe (read-only/disposable)

- The embedded Beads database is incomplete; read-only status/ready/show fail with `database not found: dap`. No live database was initialized.
- Pinned Beads 1.3.1 binary SHA-256 is `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. A disposable import/export of the current source loses `.5.3.dependencies: null`.
- Pre-change canonical source SHA-256 was `441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec`. A disposable guarded update probe failed before plan application: source-to-baseline validation reports `.1.6.10` `closed_at`/`updated_at` precision normalization. No bundle was produced or promoted. A later exact one-time direct text edit is recorded above; it did not alter the guard or create a bundle.
- Existing updates support four string fields only. Create supports one complete numeric direct child of `dap-jo9`; it cannot create this multi-level hierarchy in one plan. Closure remains pinned to `.18`, Decision 0008, and an older exact source.

The text batch and its exact approved extension are applied. Current source normalization and null-preservation constraints still block this workflow from applying further changes. Do not treat this one-time exception as permission to bypass guards again or expand frozen tooling.

## Review questions

1. Is the feature grouping useful without creating a noncanonical hierarchy?
2. Are the first-pair proposed criteria precise while keeping all six whole-story obligations?
3. Is v3 source, resource, and prerequisite treatment still visibly unapproved?
4. Is the M0 historical assessment distinct from an unapplied tracker transition?
5. The exact `.2.5.10` text tightening was later authorized and applied as a text-only refinement. No eligibility standard or candidate is selected.
6. Can each proposed text update use the current guarded path without losing source fields/nulls?
7. Are eligibility, machine, semantic, AT, authority, and release outcomes kept separate?
