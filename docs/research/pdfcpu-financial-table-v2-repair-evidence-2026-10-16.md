# dap-jo9.1.6.3 — v2 repair feasibility evidence package

Date: 2026-10-16. Branch: `feat/dap-jo9.1.6.3-repair-attempt`. Context: the
bounded v2 repair feasibility attempt (candidate bytes produced 2026-10-14,
local repair test green at 065c717).

Prior work: [v2 bounded repair result](pdfcpu-financial-table-v2-repair-2026-10-13.md)
(locally verified candidate bytes + candidate-bound CI with noncompliant findings);
[v2 clearance extension](pdfcpu-financial-table-v2-clearance-2026-10-13.md)
(metric-complete fixture identities, test-only boundary).

## Candidate identities (computed fresh 2026-10-16 with sha256sum)

| Candidate | SHA-256 | Size |
|---|---|---|
| `docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | 2,507 B |
| `docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | 2,464 B |

Both hashes and sizes matched the expected values committed with the repair attempt.
Damaged sources (unchanged, not repaired in place):
- statement: `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836`
- invoice: `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad`

## Validation matrix

| Dimension | Check | Result |
|---|---|---|
| Structure | `TestV2Row2AmountHeaderRepair` (065c717) asserts the repaired row-2 amount-cell carries the corrected `/Headers` value while every OTHER structure element, IDTree, ParentTree, and page resources are dict-identical (pdfcpu `validate` clean). The within-object single-key scope follows from the repair performing exactly one key mutation on the hash-pinned candidate bytes, whose text and render outputs are independently byte-identical. | PASS (local, 2026-10-14) |
| Text | New `TestV2CandidateTextPreserved` (f193e38): `pdftotext -layout` output byte-identical damaged→repaired. Statement text SHA `b20dd862dda3e51f2d65927f03a5df7ae01338acc76dda6f155e3f1a0b36b34b` (174 B); invoice `35fa4c4506c765eac234110017cc0592e0704bb3b51ae34589aac36d40485ea0` (169 B) | PASS (local, 2026-10-16) |
| Render | New `TestV2CandidateRenderIdentical` (a36bc5e): `pdftoppm -png -r 72` output byte-identical damaged→repaired. Statement PNG SHA `b6aed3881cd102b79a7abce7ea95c175766ecc059d16092f8927049afea3a358` (13,463 B); invoice `4cc23ec2c4a0b4ef5216c5cd0134e686bfd518d89e2ff839fc68a10fc9352b3c` (11,086 B) | PASS (local, 2026-10-16) |
| veraPDF CI wiring | `.github/workflows/verapdf-integration.yml` invokes `run-verapdf-isolated.sh` for both candidates and uploads XML/reports via `actions/upload-artifact` as `pdfua2-reports` | VERIFIED (workflow content reviewed 2026-10-16; execution requires CI) |

## Pending gate — read this before any Proceed claim

- veraPDF full XML/contracts from the latest exact head are **NOT durably retained**.
  The 2026-10-13 candidate-bound CI runs returned **exit code 1 /
  `isCompliant=false` for both candidates (6 failed rules each)** — details in the
  [2026-10-13 repair doc](pdfcpu-financial-table-v2-repair-2026-10-13.md).
- **A fresh CI run after push is required before the evidence gate can close.**
  Local preservation checks (text/render/struct) do not substitute for it.
- No Proceed disposition is claimed by this package.
- Machine validation alone establishes nothing about semantic or AT acceptability
  (repo AGENTS.md).
