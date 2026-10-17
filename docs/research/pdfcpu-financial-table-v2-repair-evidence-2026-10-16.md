# dap-jo9.1.6.3 — v2 repair feasibility evidence package

Date: 2026-10-16. Branch: `feat/dap-jo9-1-6-3-evidence`. Context: the
bounded v2 repair feasibility attempt (candidate bytes produced 2026-10-14,
local repair test green at 065c717).

Prior work: [v2 bounded repair result](pdfcpu-financial-table-v2-repair-2026-10-13.md)
(locally verified candidate bytes + candidate-bound CI with noncompliant findings);
[v2 clearance extension](pdfcpu-financial-table-v2-clearance-2026-10-13.md)
(metric-complete fixture identities, test-only boundary).

## Retained historical candidate identities

These are immutable files already retained in the repository. Their hashes
identify those files; they are not expected hashes for new pdfcpu
serialization runs. Repeated runs have produced different candidate hashes,
as documented in the [earlier feasibility report](pdfcpu-financial-table-v2-repair-2026-10-13.md).

| Candidate | SHA-256 | Size |
|---|---|---|
| `docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | 2,507 B |
| `docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | 2,464 B |

Both hashes and sizes match the expected values committed with the repair
attempt. The historical text and render checks below apply to these retained
files only.
Damaged sources (unchanged, not repaired in place):
- statement: `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836`
- invoice: `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad`

## Validation matrix

| Dimension | Check | Result |
|---|---|---|
| Structure | `TestV2Row2AmountHeaderRepair` asserts the repaired row-2 amount-cell has the corrected `/Headers` value while all traversed other structure elements, IDTree, ParentTree, and page resources are dict-identical (pdfcpu `validate` clean). The walker handles singular and array `/K` children and rejects malformed, cyclic, or over-limit trees. | PASS (local; see exact PR-head test run) |
| Text | `TestV2CandidateTextPreserved` runs `pdftotext -layout` on the damaged fixture and retained historical candidate and checks byte identity. `TestV2Row2AmountHeaderRepair` also extracts text from the just-written fresh candidate in the same run that generated and parsed its bytes; output must match the damaged fixture. Fresh candidate hashes are logged for diagnosis, not pinned. | PASS (local; fresh-candidate check is candidate-bound) |
| Render | `TestV2CandidateRenderIdentical` renders the damaged fixture and retained historical candidate at 72 dpi and checks byte identity. `TestV2Row2AmountHeaderRepair` also renders the just-written fresh candidate in the same run that generated and parsed its bytes and checks identical output against the damaged fixture. Fresh pdfcpu output is not hash-pinned. | PASS (local; fresh-candidate check is candidate-bound) |
| veraPDF CI wiring | `.github/workflows/verapdf-integration.yml` invokes `run-verapdf-isolated.sh` for both candidates and uploads XML/reports via `actions/upload-artifact` as `pdfua2-reports` | VERIFIED (workflow content reviewed 2026-10-16; execution requires CI) |

## Evidence scope and pending gate — read this before any Proceed claim

- The 2026-10-13 candidate-bound CI reports for the two named retained
  historical candidates were reconciled and recorded: both returned **exit
  code 1 / `isCompliant=false` with 6 failed rules each** — details in the
  [2026-10-13 repair doc](pdfcpu-financial-table-v2-repair-2026-10-13.md).
- New local repair attempts generate candidates in temporary directories.
  Their exact bytes are checked by the test that writes them, including text,
  render, structure, and pdfcpu validation. Those temporary bytes and any
  corresponding veraPDF report are not retained as durable artifacts by this
  test. The workflow's veraPDF jobs validate the checked-in historical
  candidate paths; reports apply to those exact file hashes.
- This document does not report fresh-candidate veraPDF compliance or claim
  that the fresh temporary output has a retained, candidate-bound report.
- No Proceed disposition is claimed by this package.
- Machine validation alone establishes nothing about semantic or AT acceptability
  (repo AGENTS.md).
