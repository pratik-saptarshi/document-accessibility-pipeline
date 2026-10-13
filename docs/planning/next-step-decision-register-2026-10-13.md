# Next-step decision register — 2026-10-13

## Accepted priority

For the first remediation increment, **statements and invoices with tables matter first and most**, as accepted for prioritization only in [Decision 0002](../decisions/0002-initial-remediation-family-priority.md). Both families are in scope for prioritization; no individual family is chosen ahead of the other. This does not approve a production workflow, a PDF engine/writer, or unresolved profile values.

## Proposed finite next experiment — not implemented or production-approved

Test whether an eligible Go-first candidate can make one explicitly identified table-header relationship repair from a fixture-authored semantic map, preserve the document, and produce candidate-bound PDF/UA-2 evidence through the existing isolated CI path. Use two small synthetic fixtures: one statement transaction table and one invoice line-item table. Fixed synthetic language, fonts, and values are experiment parameters only, not accepted product restrictions or production defaults.

Before the experiment, engineering must verify the exact normative representation and that the validator's positive and negative controls are valid for the claimed check. The experiment must use an independent immutable input baseline and produce a separate atomic candidate. It must demonstrate an actual structural change and independently check cell/header/row association; a validator pass alone is not proof that the association is correct.

Targeted future acceptance:

- Preserve displayed values exactly, including dates, signs, decimals, currency, quantities, and totals. Do not recalculate, normalize, or invent financial policy.
- Bind all machine evidence to the exact candidate hash; select explicit `ua2` in the existing isolated CI validator path. Disclose metadata changes.
- Include deliberate amount-change and cell-association mutation controls, plus valid positive and negative validator controls. A control result does not establish that the validator detects all semantic mistakes.
- Return review-needed for ambiguous mapping. Fail safely for unsupported topology, missing evidence, cancellation, or partial writes; do not report success or retain a partial candidate.
- End with `proceed`, `reject`, or `inconclusive` for this precise operation. A no-go is a valid result.

Production privacy/access/retention decisions do not block this synthetic feasibility work; they must be resolved before sensitive or production documents are used. Human semantic and AT outcomes remain distinct evidence, with authority/sign-off details unresolved until needed for the candidate-review flow.

Partially tagged PDF 2.0 fixtures may be used to isolate editing capability if their limitations are explicit. They do not imply PDF 1.7-to-2.0 conversion, arbitrary untagged input, multi-page or merged-table support, or general financial-document support. The next experiment is not authorized by this record; it is a proposed target for the next planning/fixture/feasibility batch. No PDF writer or additional form fixture is part of migration batch one.

## Decisions that remain open

- Exact profile fields and semantic map requirements; supported table topology; input eligibility and unsupported-case boundaries.
- Candidate capability, exact normative structure, license/capability clearance, and preservation threshold.
- Real-document language, complex table topology, and whether operator-supplied mapping is acceptable.
- Business arithmetic, tax, currency, dates, and rounding rules. The experiment preserves values and computes none.
- Privacy, access, retention, deletion, and diagnostic redaction before sensitive or production documents are used.
- Human semantic and assistive-technology reviewer identity, authority, protocol, outcome handling, and sign-off when required.
- Production workflow and release eligibility.

Generation remains a separate lane with an independent profile and unresolved engine and family details. The accepted remediation priority does not select its scope.

## Next tracker work

`dap-jo9.17` is the P1 acceptance-case mapping and legacy-requirement reconciliation successor for batch two. It must resolve the 12 audited task mappings and six legacy requirement references without dropping obligations. Broad dependency rewiring remains deferred until those mappings are accepted; preserve current safety and release gates meanwhile.
