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

### Statement renderer spike fixture: `ARIAPDF-SPIKE-STATEMENT-001`

This reduced fixture maps to canonical `TC-GEN-STATEMENT-001` as **partial coverage only**. It does not satisfy that case's proposed field and arithmetic profile, source-to-output manifest, or separately reviewable semantic structure obligations.

Title: `Synthetic statement sample`

| Date | Description | Amount |
|---|---|---:|
| 2026-09-03 | North Harbor | -$18.25 |
| 2026-09-04 | City Transit | $42.10 |

Expect exactly the title, headers, row order, cell text, and one-column header-to-cell associations shown. Do not add opening/closing balances or derive a balance. The literal `$` is text; it does not declare a currency.

### Invoice renderer spike fixture: `ARIAPDF-SPIKE-INVOICE-001`

This reduced fixture maps to canonical `TC-GEN-INVOICE-001` as **partial coverage only**. It does not exercise declared currency and rounding rules or verify totals, and cannot pass or close the canonical case.

Title: `Synthetic invoice sample`

| Item | Quantity | Unit amount |
|---|---:|---:|
| Paper | 2 | $12.50 |
| Envelopes | 1 | $8.00 |

The test may use the synthetic identifier `SYN-INV-001` if an identifier is required by the fixture harness; it is test-only and not part of the existing fixture values. Expect exact title, header labels, row order, and cell text. Do not compute line amounts, subtotal, tax, total, currency, or rounding. The canonical invoice acceptance case still requires declared currency and rounding rules and verification of totals against those rules; this rendering-only proposal cannot close or weaken that requirement.

## Proposed correspondence and report fixtures

These exact strings and expected structures are proposed synthetic test values, not approved business text or production defaults. They are included so the complete fixture and oracle bundle can be reviewed before owner acceptance. A future owner approval must cover these values and expectations together with the financial fixtures; if any are defined or changed after approval, obtain renewed owner approval before candidate execution.

### Correspondence: `ARIAPDF-SPIKE-CORRESPONDENCE-001`

Maps to canonical `TC-GEN-CORRESPONDENCE-001` as partial coverage only. The test-only fields are language `en-US`, recipient `Sam Example`, and title `Synthetic notice sample`.

Expected source text and order:

1. Heading: `Synthetic notice sample`
2. Recipient line: `To: Sam Example`
3. Paragraph: `Your sample statement is ready.`
4. Paragraph: `Open the Sample portal to review it.` The words `Sample portal` form a link to the exact target `https://example.invalid/statements/SYN-STATEMENT-001`.
5. List item: `Keep this notice for your records.`
6. List item: `Contact the sample desk with questions.`

The independent oracle expects that text in this order, the declared language and recipient bindings, one link with the exact target above, and a two-item list in the stated order. The values are synthetic and do not satisfy canonical correspondence acceptance, which also requires complete language/title/recipient binding behavior and explicit missing-binding outcomes.

### Report: `ARIAPDF-SPIKE-REPORT-001`

Maps to canonical `TC-GEN-REPORT-001` as partial coverage only. This one-page fixture is proposed synthetic content, not approved report policy.

Expected source content and semantic order:

1. Document title and level-one heading: `Synthetic activity report`
2. Level-two heading: `Overview`
3. Paragraph: `This test report contains two synthetic items.`
4. Table headers: `Item`, `Count`; body rows, in order: `Paper` / `2`, then `Envelopes` / `1`. The oracle expects each body cell associated with its respective column header.
5. Figure with authored alt text: `Bar chart: Paper 2; Envelopes 1.`

The page header text is `Synthetic activity report`; the footer text is `Page 1`. Both are expected as page furniture/artifacts outside the content reading order. The independent oracle checks the stated heading hierarchy, reading order, table associations, figure alt text, and furniture treatment. The fixture deliberately omits canonical provenance metadata and unmapped-value cases, so it cannot pass the canonical report case.

These smoke fixtures keep all four generation families represented but do not claim equal coverage. Their exact source strings, link target, order, relationships, and independent expected manifests are proposed and must be frozen and hashed before owner approval. The proposal leaves the profile schema and product profile examples unchanged.

## Test-first acceptance proposal

The owner has not accepted this profile or its fixture values. First define, review, and hash the complete source fixtures and independent expected text/structure oracles for all four families. Then submit that immutable bundle and the bounded profile together for owner approval. If any expectation was previously approved before being fully defined, seek renewed approval after the complete bundle is presented. Only after approval, author failing tests from the frozen oracles before any renderer adapter or behavior; do not execute the candidate before approval. No Beads ownership, status, or acceptance criterion is changed here.

### Spike IDs and canonical coverage boundary

Use these distinct IDs for every reduced renderer spike fixture and control. The mappings identify the canonical case each can inform; they do not transfer IDs or acceptance status:

| Spike-only ID | Canonical case informed | Coverage limit |
|---|---|---|
| `ARIAPDF-SPIKE-STATEMENT-001` | `TC-GEN-STATEMENT-001` | Partial: exact text/order/associations for the two listed rows only; no accepted statement field/arithmetic profile, complete source-to-output manifest, or separate semantic structure review. |
| `ARIAPDF-SPIKE-INVOICE-001` | `TC-GEN-INVOICE-001` | Partial: exact rendering of listed fields only; no declared currency/rounding rules or total verification. |
| `ARIAPDF-SPIKE-CORRESPONDENCE-001` | `TC-GEN-CORRESPONDENCE-001` | Partial: one positive example of declared language/title/recipient bindings; no missing-binding outcomes or complete required-field matrix. |
| `ARIAPDF-SPIKE-REPORT-001` | `TC-GEN-REPORT-001` | Partial: one-page structure only; no provenance metadata or unmapped-value handling. |
| `ARIAPDF-SPIKE-INVALID-DATA-001` | `TC-GEN-INVALID-DATA-001` | Partial: only the negative controls explicitly selected for this spike; no complete required-field/type validation or field-level findings contract. |
| `ARIAPDF-SPIKE-VALIDATION-001` | `TC-GEN-VALIDATION-001` | Partial: candidate-bound explicit-`ua2` machine-path controls only; not the complete machine, semantic, and assistive-technology gate. |

Record spike outcomes under the `ARIAPDF-SPIKE-*` IDs only. A spike pass may be described as partial evidence for its mapped canonical case; it must never be reported as a canonical `TC-GEN-*` pass or used to close that case. Every canonical acceptance obligation remains unchanged and must be satisfied by its own evidence before its canonical case can pass.

Tests should establish, independently from candidate output:

1. Under `ARIAPDF-SPIKE-STATEMENT-001` and `ARIAPDF-SPIKE-INVOICE-001` only, exact text, order, signs, punctuation, decimal strings, table headers, and row/cell associations for the listed reduced fixtures.
2. Mutations are detected: changed amount, removed text, reordered rows, swapped header association, or altered punctuation must fail.
3. Treat repeated spaces, tabs, NBSP, combining sequences, and long unbroken identifiers as valid input. Under the relevant `ARIAPDF-SPIKE-*` case, test for exact-preserving output; an exact-preserving result passes. Fail closed only when safe preservation is impossible within the owner-accepted bounds; never silently normalize, truncate, substitute, or drop content. These checks do not claim general text or pagination support.
4. Under `ARIAPDF-SPIKE-INVALID-DATA-001` only, negative controls for missing glyphs, unsupported topology, and render-time errors fail closed. These controls are not a claim of candidate support or complete canonical invalid-data coverage.
5. No generated output is treated as successful when rendering fails or content has been altered or dropped. Temporary artifacts are removed after each isolated run.
6. The authoring record pins the source archive, dependency/module hashes, toolchain, explicit layout test parameters, fixture hashes, and independent oracle hashes.
7. Under `ARIAPDF-SPIKE-VALIDATION-001` only, exercise the candidate-hash-bound machine validator path with explicit `ua2` and no fallback. This is partial evidence only; human semantic and assistive-technology review remain separate and the canonical validation case remains open.

The statement and invoice cases remain the priority and must be reviewed as separate acceptance cases. The correspondence and report smoke cases stay in scope. A source/license blocker, semantic mismatch, or inaccessible output is a valid `reject` or `inconclusive` outcome.

## Explicitly not decided

This proposal does not set a production language, real-customer data policy, currency or locale, date interpretation, rounding or tax behavior, required business fields, layout system, font policy, confidentiality class, retention period, semantic policy, human reviewer authority, or assistive-technology test method. It does not establish source conversion, complex typography, fallback fonts, complex table support, PDF/UA-2 conformance, or legal clearance for embedding and distributing the Go font subset.

See the [ariapdf source clearance and bounded spike plan](../../research/ariapdf-generation-clearance-2026-10-15.md) for the pinned source evidence, unresolved font-notice interpretation, candidate limitations, alternatives, and pre-prototype gates.
