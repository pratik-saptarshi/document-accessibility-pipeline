# Proposed product breakdown: financial remediation pair

**Status:** Product refinement accepted for planning quality only; unapplied. This document does not authorize Beads changes, source selection, artifact authoring, implementation, or acceptance.

**Reviewed source:** `e15b352e398581bef24aeeb679e45efdf442286d` (2026-10-17). Current canonical export SHA-256: `441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec`.

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

Semantic review, assistive-technology review, eligibility, authority, and release remain distinct. `.2.5.2` owns review specifications; `.2.5.9.1` is labeled-simulation testing; `.2.5.9.2` verifies trusted authority after `.1.1.6`; `.2.5.10` is the candidate eligibility gate. No synthetic result or machine pass substitutes for a genuine review. The proposed rule that documenting a missing eligibility standard cannot satisfy `.2.5.10` closure is conditional on explicit owner approval; current acceptance permits that blocker-documentation alternative.

## Milestones and proposed new task

- **M0 historical audit:** reconcile `.1.6.3` and `.1.6.4` against Decision 0005, Decision 0011, the retained-bundle ledger, and exact candidate/run/report identities. Record bounded historical Proceed only if independent reviewers accept the evidence; both v2 candidates remain NONCOMPLIANT. No rerun, v3 authoring, eligibility claim, or tracker transition is implied.
- **M1 guarded recording:** the current path cannot close `.1.6.3`/`.1.6.4`; see the probe evidence below. Obtain owner direction and a supported exact-source path. Do not expand frozen tooling or edit canonical JSONL directly.
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

## Exact tracker proposal and limits

The companion [JSON proposal](2026-10-17-financial-remediation-updates.json) contains full replacement values for nine existing issues' named text fields, based on the pinned source. It changes no status, ID, type, priority, owner, parent, dependency, or edge. Proposed updates to `.1.6.3`/`.1.6.4` record historical evidence and gate boundaries; the other seven updates refine planning and acceptance wording for `.2.1`, `.2.1.3`, `.2.2`, `.2.3`, `.2.3.2`, `.2.5.10`, and `.2.5.9`. All are review-only and blocked from promotion by current source normalization. The proposed `.2.5.10` acceptance text is a **material product disposition**: it makes documenting a missing eligibility standard a blocker record that cannot satisfy candidate-production closure. The product panel and judge support this proposed clarification, but explicit owner acceptance remains required before any canonical update. It is not a new eligibility decision or authorization to update the tracker.

The proposal preserves the seven `.1.6.4` downstream consumers and all six `.2.3.2` obligations. It makes no graph change and does not create the proposed task.

### Current guarded-path probe (read-only/disposable)

- The embedded Beads database is incomplete; read-only status/ready/show fail with `database not found: dap`. No live database was initialized.
- Pinned Beads 1.3.1 binary SHA-256 is `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. A disposable import/export of the current source loses `.5.3.dependencies: null`.
- Current canonical source SHA-256 is `441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec`. A disposable guarded update probe failed before plan application: source-to-baseline validation reports `.1.6.10` `closed_at`/`updated_at` precision normalization. No bundle was produced or promoted.
- Existing updates support four string fields only. Create supports one complete numeric direct child of `dap-jo9`; it cannot create this multi-level hierarchy in one plan. Closure remains pinned to `.18`, Decision 0008, and an older exact source.

Therefore none of these changes has been applied. The workflow currently blocks even a proposed update due importer/source normalization and null-preservation constraints. Do not solve this by bypassing guards or expanding frozen tooling.

## Review questions

1. Is the feature grouping useful without creating a noncanonical hierarchy?
2. Are the first-pair proposed criteria precise while keeping all six whole-story obligations?
3. Is v3 source, resource, and prerequisite treatment still visibly unapproved?
4. Is the M0 historical assessment distinct from an unapplied tracker transition?
5. Is the `.2.5.10` tightening acceptable, or must the existing “missing standard” alternative remain closure-capable?
6. Can each proposed text update use the current guarded path without losing source fields/nulls?
7. Are eligibility, machine, semantic, AT, authority, and release outcomes kept separate?
