# Adversarial review: financial remediation planning

**Date:** 2026-10-17 UTC

**Reviewed base:** `e15b352e398581bef24aeeb679e45efdf442286d`
**Disposition:** Corrected planning proposal accepted for planning quality only. Tracker updates, issue creation, source choice, authoring, and implementation remain blocked or unapproved as described below.

## Scope and method

Reviewed the proposed plan, product hierarchy, and nine-issue text-update JSON under `docs/superpowers/plans/` and `docs/planning/proposals/`, against the pinned Beads export, traceability catalog, product roadmap, Decision 0005, Decision 0011, v3 experiment plan, and existing reconciliation constraints. Three independent read-only perspectives covered product/case ownership, technical preservation/validator evidence, and tracker/governance safety; a final judge disposition addressed the corrected acceptance proposal. Reviewers did not edit files, run code tests, or execute a validator.

Local review checks: JSON parses; its source SHA-256 matches the pinned export; all nine issue IDs exist; all ten field replacements use supported text fields; no status, owner, priority, parent, type, or dependency mutation is present; the seven `.1.6.4` consumer edges are explicit. `git diff --check` passes. No `.beads` or fixture path changed.

## Findings and dispositions

1. **Category-mutation controls — P2, fixed.** The initial criterion allowed a hash/manifest mismatch to fail the combined positive contract without showing validator sensitivity. The corrected Stage 1 and Stage 2 text now requires, for each of six approved categories, an isolated validator-visible PDF mutation, an owner-reviewed category-to-expected-finding mapping, and an actual-hash-bound completed explicit-`ua2` NONCOMPLIANT report with at least one category-linked failed rule/check. An unrelated failure is insufficient; no fixed failure count is imposed. Assertion tests must reject COMPLIANT mutation XML, hash mismatch, missing report, and process/validator error as the expected negative control. External provenance/license evidence is checked separately by the provenance oracle and is not described as a PDF/UA finding. The wrong-`/Headers` relation control remains separate because it may machine-pass; an independent relationship oracle must reject it.
2. **Feature label and hierarchy — P2, fixed.** The plan and product breakdown use the same planning-only feature grouping: “Operator-confirmed financial-table repair with preservation evidence.” No canonical Feature type/ID exists, so no issue/type is invented. Core `US-REM-001` functions remain separate from the `US-REM-002` validation contribution.
3. **Acceptance ownership and whole-story scope — no remaining finding.** `.2.3.1` retains case-specification ownership; `.2.3.2` retains implementation and all six statement, invoice, correspondence, report, malformed-input, and fidelity obligations. The paired statement/invoice work is only a first increment. The exact current blockers remain binding; a split is an owner decision.
4. **`.2.5.10` acceptance change — material product disposition, supported conditionally.** Product review and the final judge support the proposed rule that recording a missing eligibility standard is useful blocker evidence but cannot itself satisfy candidate-production closure. This is a substantive acceptance tightening, not a copy edit. The existing alternative is preserved as historical text in the proposal; explicit owner acceptance is still required before any tracker update. This grants no eligibility standard or candidate authorization.
5. **Consumer count and JSON scope — P2, fixed.** `.1.6.4` now says seven and lists all seven actual blocking consumers. Obsolete “six consumers” language was removed. The plan distinguishes the two historical-gate updates from the other seven planning/acceptance text refinements in the review-only JSON.
6. **Unresolved values and oracles — no finding.** Exact v3 values, resources, licenses, source choice, and prerequisite treatment remain stop conditions until approved. No unresolved value can qualify a baseline. The owner-reviewed v3 decision packet and independent oracle must precede authoring.

## Execution boundary and blockers

- The embedded Beads DB is incomplete; read-only status/ready/show fail with `database not found: dap`. No live database was initialized.
- The current canonical source SHA-256 is `441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec`. Pinned Beads 1.3.1 is `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`.
- The disposable guarded update probe failed before plan application on `.1.6.10` `closed_at`/`updated_at` timestamp precision normalization. Disposable import/export also drops `.5.3.dependencies: null`; this source is outside the approved null-loss source list. No bundle was produced or promoted.
- Existing closure support remains `.18`/Decision 0008/exact older source only. Existing creation allows one direct numeric `dap-jo9` child and cannot create the proposed child under `.2.1`. No source exception, guard change, direct JSONL edit, edge change, status transition, or issue creation is authorized here.
- The JSON is marked non-executable; it is not a validated reconciler input. If any owner-approved update is later attempted, rebase against exact source, convert to the actual guard schema, and repeat the supported-source probe first. Omit `.2.5.10` update if owner rejects the product disposition. Do not bypass the freeze or lose null/history.

No canonical Beads update was applied. The plan and proposal provide exact reviewable text and an honest blocked state; they do not claim hierarchy creation, historical closure, or implementation readiness.
