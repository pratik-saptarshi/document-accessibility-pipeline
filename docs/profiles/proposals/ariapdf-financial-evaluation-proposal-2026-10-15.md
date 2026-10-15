# Proposed ariapdf financial evaluation profile

**Status: proposed for owner review; not approved and not a product default.** This document defines a small synthetic evaluation contract for comparing the selected ariapdf source candidate against explicit text and structure expectations. It does not choose production language, currency, tax, rounding, date, confidentiality, retention, semantic-policy, reviewer, or layout defaults. It does not authorize a prototype, dependency addition, or production selection.

This is a generation-lane proposal. The existing financial-table remediation fixtures are used only as a source of synthetic example values. Their PDFs, remediation instructions, and remediation acceptance criteria are not generation inputs and are not reused as generation policy.

## Proposed evaluation boundary

- Families: statement and invoice are the two equally deep first cases; correspondence and report remain small smoke cases.
- Lane: structured synthetic input to PDF generation. No DOCX, Markdown, or general HTML source conversion.
- Strings: the concrete financial fixtures below use ASCII/English-compatible strings solely to make the first exact oracles bounded. This is not an accepted production language or language allowlist.
- PDF target: PDF 2.0 with explicit `ua2` mode in the isolated CI candidate run. A validator result is machine evidence only, not a conformance claim.
- Layout: one-page baseline for each financial case. Before execution, the experiment record must state page size, margins, font, point size, and line/table spacing as test parameters. No renderer library defaults become product rules.
- Data: preserve each supplied field exactly, in supplied order. Do not infer or calculate balances, currency identity, quantities, subtotals, taxes, totals, or rounding.
- Handling: synthetic data only, temporary isolated workspace, and cleanup after the run. This proposal does not define production confidentiality, retention, or deletion policy.
- Font: candidate default Go Regular may be evaluated only after the embedded subset notice question is reviewed. No font is cleared for product output by this proposal.

## Synthetic financial fixtures

The following values are adapted from `docs/fixtures/remediation/financial-table-experiment.v2.json` as fixture values only. Every value must be asserted as supplied; it is not a financial calculation specification.

### Statement: `TC-GEN-STATEMENT-001`

Title: `Synthetic statement sample`

| Date | Description | Amount |
|---|---|---:|
| 2026-09-03 | North Harbor | -$18.25 |
| 2026-09-04 | City Transit | $42.10 |

Expect exactly the title, headers, row order, cell text, and one-column header-to-cell associations shown. Do not add opening/closing balances or derive a balance. The literal `$` is text; it does not declare a currency.

### Invoice: `TC-GEN-INVOICE-001`

Title: `Synthetic invoice sample`

| Item | Quantity | Unit amount |
|---|---:|---:|
| Paper | 2 | $12.50 |
| Envelopes | 1 | $8.00 |

The test may use the synthetic identifier `SYN-INV-001` if an identifier is required by the fixture harness; it is test-only and not part of the existing fixture values. Expect exact title, header labels, row order, and cell text. Do not compute line amounts, subtotal, tax, total, currency, or rounding. The canonical invoice acceptance case still requires declared currency and rounding rules and verification of totals against those rules; this rendering-only proposal cannot close or weaken that requirement.

## Remaining family smoke fixtures

- `TC-GEN-CORRESPONDENCE-001`: synthetic one-page letter with one heading, two short paragraphs, one link, and one short list. The expected reading order and link target are authored independently before candidate execution.
- `TC-GEN-REPORT-001`: synthetic one-page report with a heading, paragraph, simple table, one figure with authored alt text, and page furniture. Expected reading order, table associations, alt text, and furniture treatment are recorded independently.

These smoke cases keep all four generation families represented; they do not claim equal coverage. Their exact source text and expected manifests must be fixed and hashed before implementation. The proposal leaves profile schema and product profile examples unchanged.

## Test-first acceptance proposal

Owner acceptance of this exact bounded profile and its fixture expectations is the next prerequisite for authoring the failing tests/oracles. If accepted, write those tests before any renderer adapter or behavior. No Beads ownership, status, or acceptance criterion is changed here.

Tests should establish, independently from candidate output:

1. Exact text, order, signs, punctuation, decimal strings, table headers, and row/cell associations for both financial fixtures.
2. Mutations are detected: changed amount, removed text, reordered rows, swapped header association, or altered punctuation must fail.
3. Negative controls for missing glyphs, repeated spaces, tabs, NBSP, combining sequences, long unbroken identifiers, unsupported topology, and render-time errors fail closed. These are characterization inputs, not a claim that the candidate supports them.
4. No generated output is treated as successful when rendering fails or content has been silently dropped. Temporary artifacts are removed after each isolated run.
5. The authoring record pins the source archive, dependency/module hashes, toolchain, explicit layout test parameters, fixture hashes, and independent oracle hashes.

The statement and invoice cases remain the priority and must be reviewed as separate acceptance cases. The correspondence and report smoke cases stay in scope. A source/license blocker, semantic mismatch, or inaccessible output is a valid `reject` or `inconclusive` outcome.

## Explicitly not decided

This proposal does not set a production language, real-customer data policy, currency or locale, date interpretation, rounding or tax behavior, required business fields, layout system, font policy, confidentiality class, retention period, semantic policy, human reviewer authority, or assistive-technology test method. It does not establish source conversion, complex typography, fallback fonts, complex table support, PDF/UA-2 conformance, or legal clearance for embedding and distributing the Go font subset.

See the [ariapdf source clearance and bounded spike plan](../../research/ariapdf-generation-clearance-2026-10-15.md) for the pinned source evidence, unresolved font-notice interpretation, candidate limitations, alternatives, and pre-prototype gates.
