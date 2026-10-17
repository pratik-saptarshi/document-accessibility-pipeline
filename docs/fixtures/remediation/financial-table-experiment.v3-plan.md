# Financial-table v3 UA-2-ready baseline proposal

**Status: proposed plan only.** The user authorized planning, not baseline-file generation or candidate execution. No v3 artifacts or hashes exist. This proposal does not approve a writer, production profile, sensitive-data flow, human/AT outcome, candidate eligibility, or conformance claim.

## Purpose and boundary

The first remediation increment remains the equal-priority synthetic statement and invoice pair from [Decision 0002](../../decisions/0002-initial-remediation-family-priority.md), using the operator-confirmed mapping boundary in [Decision 0003](../../decisions/0003-operator-confirmed-financial-table-mapping.md). Its target remains two candidate-specific outputs that pass the explicit `ua2` machine gate, as scoped in `dap-jo9.2.3.2`.

The historical v1 and v2 files, hashes, observations, and candidate evidence remain immutable. The v2 candidates remain NONCOMPLIANT for six inherited findings that the bounded v2 operation leaves unchanged; the v2 contract explicitly preserves the font resources and XMP absence. The repository's [proposed Decision 0005 record](../../decisions/0005-historical-v2-bounded-association-proceed.md) documents Proceed only for the historical `/Headers` repairs and requires any UA-2-ready baseline to be separately versioned with its own allowed-change set and hashes. Its `.1.6.3` and `.1.6.4` tracker transitions remain pending. This proposal describes that next versioned pair; it does not authorize implementation or alter v2.

The [ariaPDF profile](../../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md) and [ariaPDF source-clearance plan](../../research/ariapdf-generation-clearance-2026-10-15.md) belong to document generation. They explicitly do not approve remediation inputs, policies, or a remediation engine. No ariaPDF code, values, layout choices, or resource limits are imported into this plan.

## Proposed v3 baseline structure

Create a new manifest and four new files under a separately versioned directory, for example:

- statement positive and damaged inputs;
- invoice positive and damaged inputs.

Do not create these files until the owner-approved allow-list below and exact source/resource provenance are complete. The positive and damaged member of each pair must contain the same approved baseline-conformance changes. The damaged member must differ from its positive member only at the one fixture-authored row-two amount-cell `/Headers` association already specified for that family. The historical v2 manifest, files, and findings are not normalized or replaced.

When artifacts exist, record their actual SHA-256 values, exact source v2 identities, authoring method/version, tool and resource identities, and independent observations in the new manifest. Do not prefill or infer hashes. Preserve exact displayed strings, cell values, row/column topology, table-header meanings, and reading order. Do not calculate, translate, normalize, or infer a balance, currency, tax, date, quantity, subtotal, total, or rounding rule.

## Proposed baseline change categories

The following categories map one-to-one to the six shared inherited findings in the [retained-bundle ledger](../../research/pdfua2-retained-bundle-ledger-2026-10-15.md). They are proposed baseline changes only. Engineering must produce an exact object/value-level allow-list and evidence for owner review before authoring; no category permits edits to historical v1/v2.

| v2 inherited finding | Proposed v3 baseline change category | Required invariant/evidence | Open owner approval before authoring |
|---|---|---|---|
| Missing `DisplayDocTitle` | Add the required viewer preference to both family baselines. | Explicit `ua2` report identifies the finding as resolved; document title remains the approved synthetic title. | Approve the exact preference value and title binding. |
| Missing PDF 2.0 structure namespace | Add the normative namespace declaration and bindings required by the selected PDF 2.0 structure elements. | Normative structure inspection and validator report; no change to table roles, headers, cell relations, or reading order. | Approve the exact structure vocabulary/namespace map after technical review. |
| Missing Unicode mapping | Add a valid mapping for every glyph used by the unchanged synthetic strings. | Extracted Unicode text matches every source string byte-for-byte at the text level; no normalization, substitution, or dropped glyph. | Approve the exact permitted mapping and any unsupported-glyph behavior. |
| Unembedded Helvetica program | Use an embedding-permitted font/resource for the glyphs used by both baselines. | Record exact font source, version/hash, embedding rights, notice, metrics, and rendered/text observations; check the embedding did not alter visible financial content. | Approve the font/resource and its output notice only after license and capability review. If no acceptable resource is available, stop. |
| Missing XMP metadata | Add the minimal metadata required for the selected PDF 2.0/UA-2 baseline. | XMP and document information values agree; metadata contains no personal or real-document data; exact field set is recorded. | Approve the exact metadata property/value set. |
| Untagged or unartifacted content | Tag meaningful content and mark only genuine page furniture/artifacts as artifacts. | Independent structure/content oracle verifies all content, heading semantics, reading order, table associations, and artifact treatment. | Approve the exact semantic mapping and any artifact classification. |

The baseline proposal does not authorize waiving any of these six findings. A proposed category that cannot be implemented with reviewed normative and license evidence is a stop condition; it must not become a silent exception.

## Baseline and repair controls

Before the damaged inputs are opened, complete and review the following controls for both families:

1. **Independent positive baseline:** verify the exact financial strings, table topology, IDs, header/cell associations, reading order, structure tree, Unicode extraction, font resources, metadata, and rendered observations against an oracle authored independently of the resulting PDFs.
2. **UA-2 baseline and validator controls:** run the positive v3 inputs through the isolated CI validator with explicit `ua2`. Each positive must complete with `isCompliant=true` and zero failed rules/checks. Run the repository's known-pass and known-finding controls. Do not infer acceptance from generic CI or from a `ua2` label alone.
3. **Baseline mutation controls:** for each of the six categories, omit or invalidate that category in an isolated control and verify the report no longer meets the positive baseline contract. Do not require a particular failure count if validator rules cascade; require that no mutated control is misreported as the accepted positive baseline.
4. **Positive no-op writer controls:** before opening damaged inputs, test the exact proposed writer operation on both positive v3 inputs. Stop if protected strings, topology, associations, structure, font/resource, rendering, or `ua2` status changes outside the approved output allow-list.
5. **Damaged-input identity control:** derive each damaged v3 input from its corresponding positive baseline and verify that the sole intended semantic delta is the named row-two `/Headers` association. The external validator may not detect a semantically wrong association; the independent relationship oracle must reject it.
6. **Repair and evidence controls:** require explicit operator confirmation of each proposed mapping, create separate candidates from immutable damaged bytes, repair only the confirmed association, and recheck strings, topology, structure, font/resource, text, rendering, and bounds. Run candidate-bound explicit-`ua2` validation on the exact candidate hashes; both repaired candidates must be COMPLIANT. Retain full raw XML and report contracts with candidate hash, validator/profile/version, report identity/digest, process result, and assessment.
7. **Failure controls:** preserve the first increment's amount-change, wrong-association, malformed, unsupported, ambiguous-map, cancellation, timeout, partial-write, audit-failure, and cleanup controls. No failure, missing report, validator error, or partial artifact may be surfaced as success.

Any candidate or control report applies only to the exact bytes it names. Machine compliance does not establish semantic/AT acceptance, eligible-candidate status, or release.

## Prerequisite split for the synthetic financial slice

This is a proposed planning split, not a tracker update. Preserve the full acceptance criteria and dependencies of `dap-jo9.2.3.2`; it retains six obligations and is not closed by success, rejection, or inconclusive results for the first pair.

| Stage | Exact prerequisites for this work | Separate work that remains open |
|---|---|---|
| Prepare the v3 baseline proposal | Accepted experiment scope `dap-jo9.1.1.1`; immutable source/baselines from `dap-jo9.2.1.2`; technical closure of historical `dap-jo9.1.6.3` and `.1.6.4` through an authorized, fidelity-preserving tracker path as set out in proposed repository Decision 0005; owner acceptance of the exact v3 change list and unresolved field/resource choices above before artifact authoring. | No production profile defaults or real/sensitive-document handling. General profile examples remain open. |
| Author and verify baseline artifacts | Approved object/value-level allow-list; exact candidate/tool/resource capability and license clearance (recheck `dap-jo9.1.3` if v3 changes its reachable materials); independently validated positive/negative validator controls; immutable v2 source and new v3 output hashes recorded only after creation. | No engine is approved for product use. No human/AT or release decision is implied. |
| Implement the paired first increment | Accepted v3 positive/damaged baselines; `dap-jo9.2.2`'s smallest consumer interface; exact candidate/license boundary; operator confirmation under Decision 0003; the pair-specific negative and failure controls in `dap-jo9.2.3.2`. | `dap-jo9.2.1.3`'s broader evidence-limited corpus expansion, generic profile completion, correspondence/report implementation, trusted reviewer-role source, candidate eligibility, human/AT review, and release remain separate. |

The current issue graph directly blocks `dap-jo9.2.3.2` on `dap-jo9.1.6.4`, `dap-jo9.2.1`, `dap-jo9.2.2`, and the broad profile parent `dap-jo9.1.1` (as well as the already-closed `.1.3` candidate boundary). `dap-jo9.2.1` includes later corpus expansion; `.1.1.3` already says production handling does not block synthetic engineering. To run only this first pair while leaving general profile and corpus work open, the owner/issue maintainers must approve a narrow prerequisite split in the tracker before implementation. This plan neither changes edges nor treats current open blockers as satisfied. Keep all seven downstream edges of `.1.6.4` intact.

The trusted role-source issue `dap-jo9.1.1.6`, eligibility issue `dap-jo9.2.5.10`, and human/AT gate `dap-jo9.2.5.9` are downstream release/review gates, not prerequisites for authoring synthetic baselines. This does not remove their existing criteria or dependencies.

## Decisions still required before artifact generation

The owner-approved baseline bundle must specify:

- the exact object/value-level allow-list implementing all six proposed categories;
- the approved synthetic title and exact metadata properties/values;
- the normative structure namespace and content-to-structure mapping, including what is meaningful content versus artifact;
- the exact font/resource, its embedding rights and notice, and evidence that the selected method can produce the required mapping without changing visible values;
- the validator positive/finding controls and baseline acceptance condition (`ua2`, completed, COMPLIANT, zero failed rules/checks); and
- the final prerequisite split that lets the paired synthetic slice proceed while generic profile/corpus work remains open.

If any owner choice or exact license/capability evidence is missing, keep the v3 artifacts ungenerated. Once artifacts exist, present their actual hashes and complete evidence for approval; never approve a hash that has not been produced.

## Review-to-plan integration

| Existing record/finding | Plan disposition |
|---|---|
| Decision 0002 and Decision 0003: paired priority and operator-confirmed mapping | Preserved; no automatic mapping or family precedence introduced. |
| Decision 0005: historical v2 Proceed only; v2 findings may not be repaired in place | Preserved; separate v3 pair and pre-authoring tracker gate are explicit. |
| Retained-bundle ledger: six baseline findings and both candidates NONCOMPLIANT | Mapped individually to proposed v3 categories; no finding is waived or claimed fixed. |
| `dap-jo9.2.3.2`: paired COMPLIANT first increment, six full-task obligations | Preserved; pair evidence does not close the full task or downstream review gates. |
| ariaPDF proposal and clearance: generation lane only | Excluded from remediation values, baseline authoring, and engine selection. |
| Current issue graph: broad profile/corpus prerequisites on the pair task | Reported as an unresolved prerequisite split; no graph/status mutation is proposed here. |

## Source records

- [Financial-table experiment protocol v1](financial-table-experiment.v1.md) and [v2](financial-table-experiment.v2.md).
- [Decision 0005: historical v2 bounded association Proceed](../../decisions/0005-historical-v2-bounded-association-proceed.md).
- [Retained-bundle ledger](../../research/pdfua2-retained-bundle-ledger-2026-10-15.md).
- [Product roadmap, Phase 2](../../roadmap/product-roadmap.md).
- [ariaPDF generation profile proposal](../../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md) and [source-clearance plan](../../research/ariapdf-generation-clearance-2026-10-15.md), explicitly excluded from this remediation plan.
