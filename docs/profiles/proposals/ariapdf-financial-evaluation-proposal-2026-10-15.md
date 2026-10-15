# Proposed ariapdf financial evaluation profile

**Status: proposed for owner review; not approved and not a product default.** This document defines a small synthetic evaluation contract for comparing the selected ariapdf source candidate against explicit text and structure expectations. It does not choose production language, currency, tax, rounding, date, confidentiality, retention, semantic-policy, reviewer, or layout defaults. It does not authorize a prototype, dependency addition, or production selection.

This is a generation-lane proposal. The existing financial-table remediation fixtures are used only as a source of synthetic example values. Their PDFs, remediation instructions, and remediation acceptance criteria are not generation inputs and are not reused as generation policy.

## Proposed evaluation boundary

- Families: statement and invoice are the two equally deep first cases; correspondence and report remain small smoke cases.
- Lane: structured synthetic input to PDF generation. No DOCX, Markdown, or general HTML source conversion.
- Strings: the concrete financial fixtures below use ASCII/English-compatible strings solely to make the first exact oracles bounded. This is not an accepted production language or language allowlist.
- PDF target: PDF 2.0 with explicit `ua2` mode in the isolated CI candidate run. A validator result is machine evidence only, not a conformance claim.
- Layout: one-page baseline for each fixture. Renderer settings (page size, margins, font, point size, and line/table spacing) are inputs only; they do not define expected layout or prove it is correct. The independent layout oracle below must specify expected regions/boxes and numeric tolerances before it is hashed and reviewed.
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

## Proposed independent layout oracle

The following are proposed spike invariants, not approved layout policy or product defaults. The complete oracle must be authored from the frozen synthetic source fixtures and an explicit page geometry, independently of any candidate PDF. Hash the oracle with the fixtures and obtain owner approval of the complete bundle before candidate execution. A page size, margin, font, or spacing setting alone is not an expected-layout oracle.

- Each one-page fixture has exactly one page with the owner-approved media box and content area. Every content glyph/shape box is inside the approved content area; no text or mark is clipped by the media box or another object.
- Expected content blocks occupy explicit oracle regions in the stated reading order. Their measured boxes stay within the oracle's numeric positional tolerance, and unrelated blocks do not overlap. The oracle must include the tolerance value and the overlap rule, not derive either from candidate output.
- Statement and invoice table cell boxes stay within the table region; columns align across header and body rows within the frozen numeric tolerance; cells do not collide; row and header associations match the independent table oracle.
- Correspondence link annotation bounds cover the exact linked words and remain within their paragraph region. The two list-item regions are distinct and appear in the expected order.
- Report figure bounds and its expected placement match the independent page oracle. Header and footer remain inside their reserved furniture regions and outside content reading order.
- Include a layout mutation oracle that detects clipping, a shifted/overlapping block, a misaligned table column, and misplaced page furniture. A text/tag-only match cannot pass when any layout invariant fails.

The owner-approved oracle must state page dimensions, content and furniture regions, expected block/cell/annotation boxes or normalized positions, numeric tolerances, and the measurement method. If those expectations change, regenerate the oracle hash and obtain renewed approval before execution.

## Proposed resource ceilings for spike controls

These are conservative **proposed spike-only ceilings**, not approved product limits. Before any candidate or validator control is run, the owner must approve these exact limits and the harness must enforce them as hard bounds. ariapdf's selected `Render()` API returns the complete PDF as `[]byte`; this proposal does not assume streaming or a bounded renderer allocation. Run the renderer in a process/container with the hard memory and wall-clock limits below, which bound allocation and runtime while `Render()` executes. After it returns, reject a result above 8 MiB before creating or replacing any output file. Test that output gate with an injected oversized result under the memory limit, without rendering the candidate. The post-return size check is an artifact-acceptance bound, not the memory-safety mechanism; a soft timeout or CI job timeout alone is insufficient.

| Resource | Proposed per-render ceiling | Required guard |
|---|---:|---|
| Serialized synthetic input | 256 KiB | Reject before parsing/rendering; count bytes from the captured immutable input. |
| Generated PDF | 8 MiB | After `Render()` returns, reject a byte slice above the limit before creating or replacing an output file. The separate hard resident-memory and wall-clock caps bound the complete in-memory render operation. Remove partial output on any failure. |
| Wall time | 30 seconds | Enforce an outer killable process/container deadline and reap the process on timeout/cancellation. |
| Resident memory | 512 MiB | Enforce a hard OS/container memory limit over the complete render process, including ariapdf's full-buffer `Render()` allocation. |
| Concurrent renders | 1 | Use a single-render worker; reject another request before constructing its `Doc` or calling `Render()`, and do not queue unbounded work. |
| Captured stdout and stderr combined | 1 MiB | Drain with a bounded capture and terminate on overflow. |

Each ceiling applies to the full render attempt, including negative controls; cancellation and cleanup are part of the bound. The isolated validator step must separately retain its existing stricter control-specific input limits and fixed allowlists unless a reviewed, additive ARIAPDF contract is implemented. No control may relax a pre-existing bound. Record the enforced values and prove input rejection before render, memory/deadline containment around full-buffer rendering, and oversized-result rejection before file write in tests before candidate execution. Any parallel-render characterization is out of scope here and requires a separate owner-approved profile and explicit hard limits before it can run.

## Test-first acceptance proposal

The owner has not accepted this profile, layout oracle, resource ceilings, or fixture values. First define, review, and hash the complete source fixtures and independent expected text/structure/layout oracles for all four families, including numeric layout tolerances and resource ceilings. Then submit that immutable bundle and the bounded profile together for owner approval. If any expectation was previously approved before being fully defined, seek renewed approval after the complete bundle is presented. Only after approval, author failing tests from the frozen oracles before any renderer adapter or behavior; first test that the harness rejects just-over-limit controls without invoking the candidate. Do not execute the candidate before approval and guard tests. No Beads ownership, status, or acceptance criterion is changed here.

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
6. The authoring record pins the source archive, dependency/module hashes, toolchain, explicit layout test parameters, independent expected-layout regions/boxes and numeric tolerances, fixture hashes, and independent oracle hashes.
7. Under `ARIAPDF-SPIKE-VALIDATION-001` only, exercise the candidate-hash-bound machine validator path with explicit `ua2` and no fallback. The current `.github/scripts/run-verapdf-isolated.sh` accepts a fixed set of control names and hashes and rejects ARIAPDF spike files. Before this step, implement and review a separate additive ARIAPDF allowlist contract in the isolated validator-container lane: exact spike fixture identity, frozen fixture/candidate hashes, explicit `ua2`, and a dedicated output location. Preserve the existing fixed oracle and remediation allowlists unchanged. Add negative tests for unknown IDs, changed hashes, profile mismatch/fallback, and path traversal, plus regression checks that existing controls retain their behavior. No wildcard or general allowlist broadening is permitted. This is planned work, not current support; a validator pass is partial machine evidence only, and human semantic and assistive-technology review remain separate while the canonical validation case remains open.

The statement and invoice cases remain the priority and must be reviewed as separate acceptance cases. The correspondence and report smoke cases stay in scope. A source/license blocker, semantic mismatch, or inaccessible output is a valid `reject` or `inconclusive` outcome.

## Explicitly not decided

This proposal does not set a production language, real-customer data policy, currency or locale, date interpretation, rounding or tax behavior, required business fields, layout system, font policy, confidentiality class, retention period, semantic policy, human reviewer authority, or assistive-technology test method. It does not establish source conversion, complex typography, fallback fonts, complex table support, PDF/UA-2 conformance, or legal clearance for embedding and distributing the Go font subset.

See the [ariapdf source clearance and bounded spike plan](../../research/ariapdf-generation-clearance-2026-10-15.md) for the pinned source evidence, unresolved font-notice interpretation, candidate limitations, alternatives, and pre-prototype gates.
