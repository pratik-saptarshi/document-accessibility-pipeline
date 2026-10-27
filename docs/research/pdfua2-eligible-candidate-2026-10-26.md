# Eligible-candidate evidence for the candidate-bound review gate

**Date:** 2026-10-26. **Issue:** `dap-jo9.2.5.10`. **Standard:**
[Decision 0026 — candidate eligibility standard: assessed
evidence](../decisions/0026-candidate-eligibility-standard.md) (accepted
2026-10-26 by the product owner).

This record states what exists for candidate eligibility under that
standard, what this change adds, and what remains open. It is not a
compliance declaration, a human semantic or assistive-technology
result, or a closure of `dap-jo9.2.5.10`.

## Candidates and declared operation

The candidates are the retained v2 repaired files, each produced from
immutable damaged bytes by the declared, bounded operation recorded for
the v2 feasibility work: changing only the row-two `/Headers`
association of one amount cell (see the
[v2 repair result](pdfcpu-financial-table-v2-repair-2026-10-13.md) and
the [reconciled evidence ledger](pdfua2-retained-bundle-ledger-2026-10-15.md)).

| Family | Damaged input SHA-256 | Candidate SHA-256 | Relation change |
|---|---|---|---|
| Statement | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | Row-two `$42.10`: `descriptionHeader` → `amountHeader` |
| Invoice | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | Row-two `$8.00`: `quantityHeader` → `unitAmountHeader` |

## Criterion-by-criterion status

1. **Produced from immutable damaged bytes by a declared, bounded
   operation** — recorded. The damaged fixtures and candidates above
   are the retained files; the operation boundary is the single
   `/Headers` association per family.
2. **Hash-bound preservation evidence complete** — recorded in the
   [v2 repair feasibility evidence](pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md):
   decoded-object comparison found the sole semantic delta is object
   20 `/A /Headers`; text extraction and rendering are byte-identical
   between each damaged input and its candidate, as exercised by
   `TestV2Row2AmountHeaderRepair`, `TestV2CandidateTextPreserved`, and
   `TestV2CandidateRenderIdentical` in `internal/financialfixtures`.
3. **Pinned `ua2` run on both the damaged input and the candidate,
   with full reports retained** — the gap this change closes. The
   isolated validator workflow previously ran only the candidates.
   The runner (`.github/scripts/run-verapdf-isolated.sh`) now also
   accepts `financial-v2-statement-damaged` and
   `financial-v2-invoice-damaged`, pinned to the damaged hashes above
   under the same input cap, and the `validator-container` CI job runs
   all four inputs with the pinned veraPDF 1.30.3 validator. Reports
   and contracts for all four are uploaded in the `pdfua2-reports`
   artifact of each workflow run. Before this change, the historical
   record contained no damaged-input validator reports, so inheritance
   was never established by a before/after validator comparison.
4. **Every finding classified inherited vs introduced, with zero
   introduced findings** — implemented in
   `internal/pdfua/eligibility.go`. `ParseFailedChecks` enumerates
   every failed check of a report (failing closed when a noncompliant
   report does not enumerate its checks completely), and
   `ClassifyFindings` performs a multiset comparison between the
   damaged-input and candidate findings over each finding's normative
   identity (specification, clause, test number, description, object,
   test expression, and error message): a candidate finding is
   inherited only when the damaged report contains the same failed
   check under that identity, and any excess candidate finding is
   introduced. A check's context is not part of its identity: context
   is validator location data (content-stream operator and glyph
   indices, object numbers, and the font-object identifier embedded in
   glyph contexts), and the first fresh paired run showed why it
   cannot identify a finding — its 64 (statement) and 45 (invoice)
   apparent introduced findings were exactly mirrored by 64 and 45
   resolved findings, and every pair differed only in the
   font-object identifier inside otherwise identical contexts (the
   bounded operation rewrites the font dictionary), while per-clause
   and per-identity counts were identical between each damaged input
   and its candidate (91 findings each for statement, 74 each for
   invoice). With identity-based classification the paired bundles
   classify every candidate finding as inherited, with zero
   introduced and zero resolved. In the `pdfua2` CI job,
   `TestEligibilityFromDownloadedCIPairedBundles` verifies both
   bundles of each family with `VerifyDamagedBundle` /
   `VerifyCandidateBundle` and asserts, via `AssessEligibility`, that
   the fresh paired run introduced zero findings. A candidate with
   even one introduced finding fails that CI test and is ineligible.
5. **Compliance itself is not required** — both candidates remain
   NONCOMPLIANT under the historical candidate reports (six failed
   rules each; 74 failed checks for invoice, 91 for statement). This
   record does not change that. Genuine human/AT review under
   `dap-jo9.2.5.9` remains the separate acceptance step; no machine
   evidence here satisfies it.

## What remains open

- `dap-jo9.2.5.10` stays open until a fresh paired CI run's bundles
  have been inspected and the zero-introduced result for the exact
  candidate hashes above is recorded against that run. The machinery
  in this change generates and checks that evidence on every
  `verapdf-integration` run; it does not by itself close the issue.
- Eligibility, acceptance, and compliance remain three distinct
  states (Decision 0026). Nothing here makes either candidate
  compliant, reviewed by a human, or released.
