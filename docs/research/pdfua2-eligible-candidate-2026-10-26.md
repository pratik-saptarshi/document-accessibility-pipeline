# Eligible-candidate evidence for the candidate-bound review gate

**Date:** 2026-10-26. **Issue:** `dap-jo9.2.5.10`. **Standard:**
[Decision 0026 — candidate eligibility standard: assessed
evidence](../decisions/0026-candidate-eligibility-standard.md) (accepted
2026-10-26 by the product owner).

This record states what exists for candidate eligibility under that
standard, what this change adds, and what remains open. It is not a
compliance declaration or a human semantic or assistive-technology
result. As amended on 2026-10-26 to record the fresh paired CI
inspection below, it supports closure of `dap-jo9.2.5.10`.

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
   The fresh paired runs inspected on 2026-10-26 are recorded in the
   inspection section below.
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
   The inspected fresh paired runs recorded below classify every
   candidate finding as inherited, with zero introduced findings.
5. **Compliance itself is not required** — both candidates remain
   NONCOMPLIANT under the historical candidate reports (six failed
   rules each; 74 failed checks for invoice, 91 for statement). This
   record does not change that. Genuine human/AT review under
   `dap-jo9.2.5.9` remains the separate acceptance step; no machine
   evidence here satisfies it.

## Fresh paired CI inspection (2026-10-26)

On 2026-10-26, the `pdfua2-reports` artifact (90-day retention) was
downloaded and inspected from two fresh `main` runs of the
"veraPDF PDF/UA-2 integration" workflow. Both runs concluded
success, and both the `validator-container` and `pdfua2` jobs
succeeded in each run. The validator was pinned at veraPDF 1.30.3.

| Run | Head SHA | Event / time (UTC) |
|---|---|---|
| `40429773708` | `ea5984e19bc88e257f2eacefe1aa0e8689dc5aff` | push, 2026-10-27T00:58:46Z |
| `40433697710` | `18c1b523272c20ed4374e1f9da2ec80c9018c14d` | push, 2026-10-27T01:37:46Z |

All contract `report_sha256` values match the downloaded
`report.xml` files in both runs. The report SHA-256 values below
are from run `40429773708`; they are identical in run
`40433697710`.

| Family | Damaged input SHA-256 | Candidate SHA-256 | Damaged report SHA-256 | Candidate report SHA-256 | Classification | Eligible |
|---|---|---|---|---|---|---|
| Statement | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | `cb49b04dc621e487c41ec6296e84ca806acc5cc5e2b2488146fe78fe36241cc0` | `65ce88a950cf62fb135f96ea9d9ac531f7487fc8e705d84cd83a05d8a6651ce0` | inherited=91 introduced=0 resolved=0 | true |
| Invoice | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | `45ec04da6866b11ac152385b2375051acb06c07b0c1b44390c862789d35f5a3a` | `53ea0629fb2a472d7ca4831085a4a2b35e61bded4fe7eecbab3f9843ec919fdf` | inherited=74 introduced=0 resolved=0 | true |

Inspection method: the repository's own
`TestEligibilityFromDownloadedCIPairedBundles` test was re-run
locally with `PDFUA_BUNDLE_DIR` set to each downloaded bundle
directory. It passed for both runs, with the classification counts
above: statement inherited=91 introduced=0 resolved=0,
eligible=true; invoice inherited=74 introduced=0 resolved=0,
eligible=true.

## What remains open

- `dap-jo9.2.5.10` is resolved by this inspection recording: the
  fresh paired CI runs above were inspected and the zero-introduced
  result for the exact candidate hashes above is recorded against
  those runs. The bead is closed on this basis. The machinery in
  this change generates and checks that evidence on every
  `verapdf-integration` run; the machinery alone did not close the
  issue — this recorded inspection does.
- Eligibility, acceptance, and compliance remain three distinct
  states (Decision 0026). Nothing here makes either candidate
  compliant, reviewed by a human, or released. Genuine human/AT
  review under `dap-jo9.2.5.9` remains separate and open.
