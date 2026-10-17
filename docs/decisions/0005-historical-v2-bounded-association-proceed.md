# Historical v2 bounded association-repair Proceed

**Status: Proposed repository decision record.** The narrow historical outcome below has independent technical and product-value acceptance, plus a final assessment in the source thread. The two acceptances are independent AI-assisted reviews, not human semantic/AT approval or product-owner release sign-off. This file is a reviewable draft only; `.1.6.3` and `.1.6.4` remain open/in progress pending a safe authoritative tracker update.

**Decision date:** 2026-10-15 UTC. **Criteria rechecked at the reviewed commit's actual base:** `bc4c136ded20a1b64c53075965d98f6ce892c0de` (current `main`, including PR #76). Relative to the earlier check at `e4ea5ad30eae19c9baa9a4df0439a323d4156e3a`, PR #76 adds two catalog strictness research plans and changes the traceability catalog implementation/tests; it changes neither `.beads/issues.jsonl` nor the v2 candidate/fixture files. This was verified by the `e4ea5ad..bc4c136` path diff and direct comparison of those tracker/fixture paths. The `.1.6.3` and `.1.6.4` acceptance criteria and dependencies therefore remain as recorded at the earlier check. PR #73 is merged at `b7b6e15eaf9c28eee2b9b0b79855198b3fe5e6c9`; PR #74 is included at the earlier snapshot.

## Decision

Accept **Proceed** for the exact paired historical engineering feasibility attempt: repair the single fixture-authored row-two `/Headers` association in each immutable, metric-complete v2 input, preserving every other declared string, relation, structure, font/resource and rendering observation. Both repaired outputs received candidate-bound explicit `ua2` assessments that completed normally and were **NONCOMPLIANT**. The inherited findings are fully enumerated and accepted as unchanged for this bounded operation only.

This finite engineering Proceed is not a conformance result, machine-compliant candidate, eligible-candidate determination, production writer/runtime approval, general table/document capability, human semantic or AT outcome, reviewer-authority proof, or release decision. Do not remediate inherited categories inside the old v2 preservation contract. Any UA-2-ready baseline must be separately versioned with its own allowed-change list and hashes.

## Exact operation and evidence

| Family | Immutable damaged input SHA-256 | Retained candidate SHA-256 | Only accepted semantic change |
|---|---|---|---|
| Statement | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | Row-two `$42.10`, `descriptionHeader` → `amountHeader` |
| Invoice | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | Row-two `$8.00`, `quantityHeader` → `unitAmountHeader` |

The positive no-op control passed on both families before damaged inputs were opened. The recorded repair used test-only pdfcpu v0.16.2. Independent comparisons covered the exact displayed strings, intended and non-target associations/order, structure, fonts/resources, text extraction, rendering, bounds, and negative/failure controls. A decoded comparison of all 23 original PDF objects found the sole semantic delta at object 20 `/A /Headers`. Source bytes, content streams, catalog/metadata, Helvetica metrics, and the other structures were preserved. These recorded historical checks were accepted by the reviewers; no check was rerun for this decision draft.

The complete retained reports and contracts reconcile four distinct archives:

| Artifact / run | ZIP SHA-256 | Invoice / statement report XML SHA-256 |
|---|---|---|
| `11848027193` / `38465226725` | `2ef4e9601b0cd5de5a12f8e97188fcef22a5e597e3bbd2aa22ee8d692152b42a` | `c601d25e6f66b63ddfea1019eff128473cfc26c583134c176c20da55fee918aa` / `940fb769d6f9bc0920d48a7d652804e588393fefc767d30d765317d2fb2dc78d` |
| `11848458063` / `38466062420` | `c172b77f34645299632a0258fd2273478a68bf25ef98e76f18677e53bd7f417b` | `a56c8b7c029c473eb5320401f414abceb4adb1b0c5d98643a9ca0728cc84344f` / `afea6666665c6183347dd20a34d1f4e764e7d09d2f38290d4122ab9dcc8d223a` |
| `11867387877` / `38505760173` | `6cbcc298210dfe207fde676b556f6b2c64dc7906302b59cce4e01593482df22e` | `7e13f3f2e8d8a1d8cc230f53e241c7a6e964b3987fb4f14e6e45dfb059e9b01d` / `5a1120ff5d06988d94409d46408c379b71832da2c7f0c6e1da1602bcbfe9fb9c` |
| `11867739001` / `38505929576` | `4fba9c4c14fb5264504059772047f170636eb8ca7f6cc821c877ed8408142ef0` | `1d9336833f0158e059342c0b691ecc94577e9365dc358a31d3a2aae6a7ce8e44` / `8fc8f3d86b2591fdcca487dc90542c904e0a754cd52bf93656fcd3f37f5a3f8f` |

All 16 contracts and reports identify veraPDF 1.30.3 and the explicit PDF/UA-2 + Tagged PDF validation profile. The eight financial-candidate reports (four per candidate) completed normally with exit 1 and `isCompliant=false`: invoice **6 failed rules / 74 failed checks** and statement **6 failed rules / 91 failed checks**. The four known-finding control reports completed normally with exit 1 for the expected finding. The four known-pass control reports completed with exit 0, `isCompliant=true`, and zero failed rules/checks. These are normal assessment outcomes, not process or validator failures.

The retained archives contain no damaged-input validator reports, and no before/after validator comparison was performed. Thus the reports alone do not demonstrate inherited findings or absence of regression. The inherited classification for this bounded association change rests on separate source/object comparisons and preservation evidence, including the no-op record, exact operation bounds, unchanged non-target structure/resources, and recorded text/render observations.

The six findings are accepted as inherited within this finite operation only: missing `DisplayDocTitle` (1/1); missing PDF 2.0 structure namespace (1/1); unembedded Helvetica program (1/1); missing Unicode mapping (61/78); absent XMP (1/1); and untagged/unartifacted content (9/9). Counts are not impact rankings. The v2 contract preserves font resources and XMP absence; no silent fixture or contract changes are allowed.

## Independent acceptance record

- Technical acceptance is an independent AI-assisted technical review, not human semantic or assistive-technology approval. It accepted the exact bounded historical Proceed evidence and reported no technical blocker. Source Library file `libfile_9e033eb55f98819196b53cf5dd19eac9`, v0; source SHA-256 `59e54cea88cf31ab6278b903cd648e1b8f54f1ed7d6940990ebd6d11ec929e3b`.
- Product acceptance is an independent AI-assisted product-value review, not a human review, personal product-owner signature, or release-signoff authority. It accepted the paired bounded Proceed and all six inherited findings within scope; no additional product criterion remained unmet. Source Library file `libfile_a25e1b6842d881918e65c287e42cdf90`, v0; source SHA-256 `cb013b7751ab50b64ce75546d334a156816cce48d50b43e7ccb284c5f7130c2f`.
- Evidence ledger: Library file `libfile_69bac0bb0ec48191bcc9c5c66a86c179`, v0, SHA-256 `3a1999a73e5c891e2ce04853d359e9565465514d904bc471e63171143c8aa011`.
- The source-thread **final assessment** accepts the same narrow paired Proceed under the current `.1.6.3/.1.6.4` criteria. It is a scoped assessment, not human semantic/AT evidence.
- Durable repository excerpts with source identities, exact hashes, review-type limits, and acceptance text are in [the acceptance register](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md).

The prior local materialization attempts failed in the Library backend/host-path handoff. That access failure is not missing evidence and does not override the complete Library ledger or two independent raw-byte readers. This record relies on their exact retained-evidence reviews; it does not claim this executor re-downloaded or re-inspected the ZIP bytes.

## Intended Beads updates (draft only)

Current main records `.1.6.3` as `in_progress` and `.1.6.4` as `open`, both owned by `waynemanorsaptarshi@gmail.com`. The intended status sequence is:

1. Close `dap-jo9.1.6.3` with finite outcome **Proceed**, citing the paired repair evidence, four retained bundles (eight financial-candidate reports), all findings assessment, and independent technical/product acceptance. Keep its acceptance criteria and dependencies unchanged.
2. Close `dap-jo9.1.6.4` as its **Proceed-only** gate after `.1.6.3` records Proceed, citing the same exact paired evidence and both acceptances. Preserve its parent and `.1.6.3` prerequisite edges and all seven downstream consumer edges listed below.
3. Correct the `.1.6.4` acceptance text's stale count “six existing downstream dependencies” to “seven existing downstream dependencies” without changing the seven actual graph edges or any consumer scope.

The seven `.1.6.4` consumers that must remain linked are:

- `dap-jo9.2.1.3`
- `dap-jo9.2.2`
- `dap-jo9.2.3`
- `dap-jo9.2.3.2`
- `dap-jo9.2.4.2`
- `dap-jo9.2.5.9`
- `dap-jo9.2.5.9.1`

Do not modify `dap-jo9.1.6.8`, `dap-jo9.1.6.9`, or `dap-jo9.5.4`; those remain under their separate owner. Closing `.1.6.4` does not close any dependent issue or establish eligibility; remaining blockers and independent closure criteria still apply.

### Draft close records

**`dap-jo9.1.6.3` — proposed close reason:**

> Finite outcome: **Proceed** for the exact paired historical v2 attempt. Both metric-complete positive no-op controls passed before the damaged-input attempt. From immutable inputs, the test-only pdfcpu v0.16.2 experiment produced only the fixture-authored row-two `/Headers` changes: statement `$42.10` `descriptionHeader`→`amountHeader` (input `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836`, candidate `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52`); invoice `$8.00` `quantityHeader`→`unitAmountHeader` (input `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad`, candidate `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448`). The preserved structures, exact strings/relations, text/render observations, bounds and failure controls were independently accepted. Four distinct retained bundles contain complete candidate-bound explicit-`ua2` reports/contracts; both assessments completed normally NONCOMPLIANT (six failed rules each; invoice 74, statement 91), with all six findings assessed as inherited within this operation only. Technical and product acceptance: Library files `libfile_9e033eb55f98819196b53cf5dd19eac9` and `libfile_a25e1b6842d881918e65c287e42cdf90`; evidence ledger `libfile_69bac0bb0ec48191bcc9c5c66a86c179`. This does not establish conformance, eligibility, production writer/runtime approval, human/AT acceptance, or release.

**`dap-jo9.1.6.4` — proposed close reason:**

> Accepted bounded association **Proceed** for both actual repaired candidates from immutable v2 inputs, changing only their declared row-two `/Headers` association. The statement candidate is `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52`; the invoice candidate is `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448`. Independent preservation/object, exact-string, order/relationship, text/render, bounds and failure-control evidence is recorded and accepted. Complete reports/contracts from four distinct retained bundles bind to those exact candidate hashes; explicit `ua2` assessments completed normally and remain NONCOMPLIANT (statement 91, invoice 74 failed checks, six inherited findings each). Independent technical and product reviewers accepted this exact evidence; see Library ledger `libfile_69bac0bb0ec48191bcc9c5c66a86c179`, technical acceptance `libfile_9e033eb55f98819196b53cf5dd19eac9`, and product acceptance `libfile_a25e1b6842d881918e65c287e42cdf90`. This closes only the finite historical engineering feasibility gate. It does not establish machine compliance, eligible-candidate evidence, production writer/runtime approval, human semantic/AT acceptance, or release. Preserve all seven downstream dependency edges unchanged.

## Consequences and limits

No new fixture or baseline implementation may begin until both gate records are accepted in the authoritative tracker and safely published. The next baseline proposal must version the complete needed changes independently; it must not alter historical v2 bytes, hashes, or findings. Keep statements and invoices paired and equal priority. Keep explicit `ua2`, CI-only isolated veraPDF, Go-first/permissive dependency boundaries, operator-confirmed mapping, and the open trusted role-source/temporal policy intact. Historical eligibility and release gates remain separate.

## Tracker export-fidelity blocker

The accepted disposition is ready as a reviewable draft; the status transition is not. A separate build task (`01a13cb9-7067-72e0-83d8-d2bcc1bb50c0`, completed turn `01a13d74-83e2-71d5-84ad-83d561410456`) verified bd 1.3.1 source and a fresh 99-issue/166-edge round-trip. The imported dependency graph was preserved, but export synthesized a missing `.5.5` `closed_at` as `2026-10-14T04:25:39` from `created_at` `2026-10-14T04:23:06` and `updated_at` `2026-10-14T04:25:38`. This derived timestamp is not factual closure evidence. That round-trip also normalized unrelated fields. The round-trip was not reproduced in this executor; its local observation that `bd` is unavailable is a separate environment limitation.

Do not hand-edit or publish that lossy export. Leave `.1.6.3`/`.1.6.4` transitions unapplied until an authoritative, fidelity-preserving tracker update path is available and the intended edits can be safely reviewed. Preserve the seven actual consumer edges and correct the stale “six” wording only within that safe path.

## Source records

- [Reconciled retained-bundle evidence ledger](../research/pdfua2-retained-bundle-ledger-2026-10-15.md).
- [V2 positive no-op record](../research/pdfcpu-financial-table-v2-noop-2026-10-13.md) and [v2 repair record](../research/pdfcpu-financial-table-v2-repair-2026-10-13.md).
- [Current acceptance and dependencies](https://github.com/pratik-saptarshi/document-accessibility-pipeline/blob/bc4c136ded20a1b64c53075965d98f6ce892c0de/.beads/issues.jsonl) (compared at the reviewed commit's actual base; unchanged from the earlier check).
- Historical CI runs: [38465226725](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38465226725), [38466062420](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38466062420), [38505760173](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38505760173), [38505929576](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38505929576).
