# Financial-table mapping options comparison (bounded spike)

**Date:** 2026-10-17. **Beads:** `dap-jo9.1.6.7` (spike, P2, CLOSED). **Ralph task:** RALPH-022.

**Authority boundary:** this spike is exploration only and authorizes nothing.
No automatic mapping is authorized. No inference is implemented. No
unconfirmed suggestion is applied. No general document support, production
writer selection, or human semantic/AT reviewer authority is defined here.

**Baseline:** [Decision 0003](decisions/0003-operator-confirmed-financial-table-mapping.md)
established the operator-confirmed workflow for the first bounded
statement/invoice slice: an operator confirms the intended table-header
association before any repair changes it; no automatic or unconfirmed
association is applied. The candidate options below are *future alternatives
considered after* that baseline — not replacements, and none is selected.

**Fixture ground truth:** the synthetic v2 families carry one deliberately
damaged row-two `/Headers` association each, per
[Decision 0005](decisions/0005-historical-v2-bounded-association-proceed.md):
statement cell `$42.10` currently `descriptionHeader`, expected `amountHeader`;
invoice cell `$8.00` currently `quantityHeader`, expected `unitAmountHeader`.
"Expected associations" below always mean fixture-authored expectations on
these synthetic tables, never inferred semantic truth.

## Candidate options

### Option A — Operator-confirmed mapping (accepted baseline, future variants)

- **Inputs:** current `/Headers` associations, candidate header set,
  operator's explicit confirmation naming one exact association.
- **Assumptions:** the operator knows the intended association; confirmation
  is per-change and names the exact cell→header pair; the operator's
  authority excludes semantic/AT signoff and release (`dap-jo9.1.1.4` open).
- **Fixture-authored expected associations:** statement `$42.10` → `amountHeader`;
  invoice `$8.00` → `unitAmountHeader` — applied only after the operator names
  them exactly.
- **Ambiguity behavior:** any ambiguity (multiple candidates, conflicting
  signals) is exposed to the operator; no default pick, no auto-apply.
- **Limits:** human throughput; no confidence scoring; confirmation wording
  and reviewer-authority protocols still open.
- **Unsupported boundaries:** unattended/production processing, general
  documents, any writer selection, semantic correctness judgment.

### Option B — Exact header-string matching, suggestion-only

- **Inputs:** header cell strings, the amount cell's row context.
- **Assumptions:** header strings are stable across synthetic families; exactly
  one header matches the target string per family.
- **Fixture-authored expected associations:** statement `$42.10` → `amountHeader`;
  invoice `$8.00` → `unitAmountHeader` — surfaced as suggestions, never applied.
- **Ambiguity behavior:** zero matches → missing (no suggestion); multiple
  exact matches → ambiguous, surfaced with all candidates, no silent pick.
- **Limits:** fails on synonyms/paraphrase, localized strings, merged or
  split headers; case/whitespace normalization choices undecided.
- **Unsupported boundaries:** auto-application, semantic meaning of a header,
  documents beyond the synthetic statement/invoice families.

### Option C — Column-geometry alignment, suggestion-only

- **Inputs:** x-bounds of amount cells vs. header-column bounds (Helvetica
  metric-complete fixtures, per the v2 fixture record).
- **Assumptions:** columns do not shift between fixture generations; metrics
  are complete and stable.
- **Fixture-authored expected associations:** statement `$42.10` column aligns
  with `amountHeader`; invoice `$8.00` column aligns with `unitAmountHeader` —
  surfaced as suggestions, never applied.
- **Ambiguity behavior:** overlapping/tied column bounds → ambiguous; no
  geometry overlap → missing.
- **Limits:** the v1 REJECT finding (pdfcpu v0.16.2 full-write font-metric
  expansion) means geometry is not writer-stable; ragged layouts, rotated
  text, and multi-line headers break the alignment assumption.
- **Unsupported boundaries:** auto-application, real-world layout variance,
  any writer that expands metrics.

### Option D — Structure-tree association reuse, suggestion-only

- **Inputs:** existing `/Headers` and `/ColSpan` entries in the structure tree.
- **Assumptions:** a structure tree exists and its entries are trustworthy
  inputs for suggestion generation.
- **Fixture-authored expected associations:** none trusted — the v2 fixtures
  carry deliberately damaged row-two `/Headers`; this option must treat the
  current association as a conflicting signal, not as truth.
- **Ambiguity behavior:** inherited (damaged) associations surface as
  conflict, not confirmation; absent structure → missing.
- **Limits:** structure cannot attest semantic correctness; machine-only,
  never a meaning claim; damaged fixtures actively mislead naive reuse.
- **Unsupported boundaries:** auto-application, trusting structure as
  semantic evidence, general document structure repair.

## Synthetic controls and results

Controls are implemented as an exploration harness in
`internal/mappingoptions/` (package `mappingoptions`): it models candidate
suggestions and the operator-confirmation guardrail policy only. It performs
no inference and contains no writer. Guardrail under test: **no unconfirmed
suggestion is applied; unsafe associations are never applied.**

| Control | Input | Expected | Result |
|---|---|---|---|
| Accuracy | fixture-authored suggestion (statement `$42.10`→`amountHeader`, invoice `$8.00`→`unitAmountHeader`) | recognized as suggestion, not applied; `Apply` without confirmation refused | PASS |
| Missing | no table/header data | no suggestion; `Apply` refused | PASS |
| Multiple | two matching candidates (`amountHeader`, `totalHeader`) | ambiguous; no silent pick; operator confirmation of one exact candidate resolves per baseline | PASS |
| Conflicting | header-string vs structure-tree signals disagree | no suggestion surfaced; `Apply` refused even with a confirmation | PASS |
| Unsafe | association flagged unsafe (would alter displayed text / reorder rows) | blocked; `Apply` refused even with operator confirmation (Decision 0003 preserves exact displayed values) | PASS |
| Guardrail | suggestion with no/mismatched confirmation | never applied | PASS |

Test evidence: `go test -race ./internal/mappingoptions/ -v -count=1` — 6/6
controls green. Tests were written first and observed failing (build failed:
undefined symbols) before the minimal harness was added.

## Recommendation

No implementation decision is warranted by this spike. The operator-confirmed
workflow (Decision 0003) remains the only accepted boundary; Options B–D are
recorded as suggestion-only candidates with the limits and ambiguity behavior
above, and none may auto-apply. Any future decision to implement one of these
options — including exact matching, geometry, or structure-reuse signals, or
any change to confirmation protocols — requires a **separately authorized
decision record** with its own accuracy/ambiguity evidence and authority
scoping. In particular, defining operator/semantic/AT reviewer authority
remains open under `dap-jo9.1.6.7`'s exclusions and `dap-jo9.1.1.4`.
