# Reconciled retained PDF/UA-2 evidence ledger

**Record status:** reconciled historical evidence; reviewed for the bounded disposition below. This is not a new validator run, current-main test result, conformance declaration, candidate-eligibility decision, or tracker closure.

**Repository snapshot checked:** current main `e4ea5ad30eae19c9baa9a4df0439a323d4156e3a` (PR #75), rechecked from `ae2f5259361142575cb2a59811368195f9bbfcf9`. PR #75 changes six corpus/JSON/traceability/test files; it changes neither `.beads/issues.jsonl` nor v2 candidate/fixture files. The `.1.6.3/.1.6.4` acceptance criteria and dependencies therefore remain as recorded at the prior snapshot; states remain `in_progress` and `open`. PR #73 is merged at `b7b6e15eaf9c28eee2b9b0b79855198b3fe5e6c9`.

## Evidence sources and method

The underlying retained-bundle ledger is Library file `libfile_69bac0bb0ec48191bcc9c5c66a86c179`, version 0, `retained-bundle-ledger-2026-10-15.md`, SHA-256 `3a1999a73e5c891e2ce04853d359e9565465514d904bc471e63171143c8aa011`. It records a read-only inspection of all four ZIPs, raw contracts/XML, CI run logs, candidate/input identities, and pinned source reads. The historical inspector opened ZIPs in memory, validated the eight-entry archive contract and CRCs, hashed raw members, checked XML/report-to-contract bindings and control/profile/version identity, and compared report trees. No new PDF writer, veraPDF, Java, rendering, repository tests, or repair was run for this reconciliation.

Independent technical acceptance is Library file `libfile_9e033eb55f98819196b53cf5dd19eac9`, version 0, `historical-feasibility-technical-acceptance-2026-10-15.md`, SHA-256 `59e54cea88cf31ab6278b903cd648e1b8f54f1ed7d6940990ebd6d11ec929e3b`. This is an independent AI-assisted technical review, not human semantic or assistive-technology approval. It independently reopened the four ZIPs and compared every JSON contract and full failed-rule/check record to archive bytes. Independent product-value acceptance is Library file `libfile_a25e1b6842d881918e65c287e42cdf90`, version 0, `historical-feasibility-product-acceptance-2026-10-15.md`, SHA-256 `cb013b7751ab50b64ce75546d334a156816cce48d50b43e7ccb284c5f7130c2f`. This is an independent AI-assisted product review, not a human review, product-owner signature, or release-signoff authority. It reviewed the complete Markdown ledger, repository criteria, preservation/no-op records, and issue dependencies, relying on the separate raw-byte inspection. The source-thread final assessment accepts the same narrow paired historical Proceed under current criteria. The durable excerpts, source hashes, and review limits are captured in [the repository acceptance register](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md).

The earlier Library-to-executor materialization failures were backend/host-path transfer failures. They are not missing evidence: the retained Library ledger and both acceptances are readable, and two independent readers already checked the retained raw bytes. This reconciliation does not require another ZIP download.

## Bundle and run identities

Source revision and actual execution checkout are distinct fields. The original and PR runs were pull-request events; artifact `11867739001` was produced by a `push` to `main` at its historical source revision. None is represented as an execution of current main `e4ea5ad`.

| Artifact / run | Source head / actual checkout | ZIP bytes / SHA-256 | Validator jobs (container / PDF-UA) | Container image SHA-256 |
|---|---|---|---|---|
| `11848027193` / `38465226725` | `6cbbde390c01812dadc2c6bfe1896e320a85ce17` / `5eb9a1dae4cc2c6201b10ebdcf294746becc824f` | 9,385 / `2ef4e9601b0cd5de5a12f8e97188fcef22a5e597e3bbd2aa22ee8d692152b42a` | `115530269681` / `115530744934` | `8e3107066dc03dc058654ed56967bf975375dec4ec5a98685ca2558802e5928b` |
| `11848458063` / `38466062420` | `821a593eb7f4ed6ae3d44773f3dd5b13016937dd` / `bb1eefbc7c8d3df12f4e7b4db86bc82e03fe321c` | 9,372 / `c172b77f34645299632a0258fd2273478a68bf25ef98e76f18677e53bd7f417b` | `115533046337` / `115533422537` | `4140a434d8fd00d19daa7504d324ea8940a6f28792272d0a70368cf777bceae4` |
| `11867387877` / `38505760173` | `e7f9268b8f754e531a280645233f90f862c3803d` / `06548938b52a99e0f4736fe112431e4fe8c90a04` | 9,380 / `6cbcc298210dfe207fde676b556f6b2c64dc7906302b59cce4e01593482df22e` | `115662779970` / `115662970554` | `71d60aa86a60edd8b7a4b821519cef9c1877fad50089c5a2d9378b869a54718f` |
| `11867739001` / `38505929576` | `715a1d10f09da5e0c2a72789b3354cde5dc87295` / `715a1d10f09da5e0c2a72789b3354cde5dc87295` | 9,376 / `4fba9c4c14fb5264504059772047f170636eb8ca7f6cc821c877ed8408142ef0` | `115663313747` / `115663523532` | `222fa3dd47540c046c31e3648ff3ae5f6f9905773cef705fd8bb190aa32b4e9a` |

Retained Library ZIP identities, respectively: `libfile_cf20ce8674b08191af471bfce2ae6b6d`, `libfile_e89662ada5708191b9ed090f76d8b911`, `libfile_93279f7dbae88191b92a50d916d75a26`, and `libfile_a14015074f9881918de07625d81b2b17`. Each archive contains the same eight expected unique entries: invoice, statement, known-finding, and known-pass `contract.json`/`report.xml` pairs. The complete per-member CRC, contract digest, XML digest, controls, and provenance are preserved in the referenced Library ledger.

## Candidate outcomes and report identities

All 16 contracts identify veraPDF 1.30.3, validator binary SHA-256 `19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70`, profile `PDF/UA-2 + Tagged PDF validation profile`, and the same pinned Java 21 base image. Workflow jobs completed normally; the validator had no parse, encryption, memory, or exception failures. The report XML digest in each row is the `report_sha256` recorded by that candidate's contract.

| Artifact | Invoice candidate SHA-256 | Invoice report XML SHA-256 | Statement candidate SHA-256 | Statement report XML SHA-256 | Result |
|---|---|---|---|---|---|
| `11848027193` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | `c601d25e6f66b63ddfea1019eff128473cfc26c583134c176c20da55fee918aa` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | `940fb769d6f9bc0920d48a7d652804e588393fefc767d30d765317d2fb2dc78d` | Both completed normally, exit 1, `isCompliant=false` |
| `11848458063` | same invoice hash | `a56c8b7c029c473eb5320401f414abceb4adb1b0c5d98643a9ca0728cc84344f` | same statement hash | `afea6666665c6183347dd20a34d1f4e764e7d09d2f38290d4122ab9dcc8d223a` | Both completed normally, exit 1, `isCompliant=false` |
| `11867387877` | same invoice hash | `7e13f3f2e8d8a1d8cc230f53e241c7a6e964b3987fb4f14e6e45dfb059e9b01d` | same statement hash | `5a1120ff5d06988d94409d46408c379b71832da2c7f0c6e1da1602bcbfe9fb9c` | Both completed normally, exit 1, `isCompliant=false` |
| `11867739001` | same invoice hash | `1d9336833f0158e059342c0b691ecc94577e9365dc358a31d3a2aae6a7ce8e44` | same statement hash | `8fc8f3d86b2591fdcca487dc90542c904e0a754cd52bf93656fcd3f37f5a3f8f` | Both completed normally, exit 1, `isCompliant=false` |

Invoice candidate: 2,464 bytes, 6 failed rules / 74 failed checks. Statement candidate: 2,507 bytes, 6 failed rules / 91 failed checks. The same pinned candidate bytes were present at the four historical execution checkouts and at main `96962a4`; PRs #73, #74, and #75 changed-file lists contain no candidate or fixture paths. Thus the candidate identities remain unchanged through current main `e4ea5ad`, without representing that as a fresh validator run.

The controls behaved as intended across all four bundles:

- `known-pass`, SHA-256 `c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b`: exit 0, compliant, zero failed rules/checks.
- `known-finding`, SHA-256 `2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0`: normal exit 1, one expected namespace/role-mapping failure.
- Both financial candidates completed normal assessment with six failed rules, not process/validator execution failures. Green report-contract CI therefore means successful assessment of findings, not document compliance.

## Operation, preservation, and findings

The accepted operation is restricted to the fixture-authored row-two `/Headers` association repair on each immutable v2 financial table:

| Family | Damaged input SHA-256 | Retained candidate SHA-256 | Exact relation change |
|---|---|---|---|
| Statement | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` | `8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52` | Row-two `$42.10`: `descriptionHeader` → `amountHeader` |
| Invoice | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` | `31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448` | Row-two `$8.00`: `quantityHeader` → `unitAmountHeader` |

Recorded independent preservation evidence: all 23 original decoded PDF objects were compared; the sole semantic delta is object 20 `/A /Headers` for the named target relation. Content streams, catalog/metadata, Helvetica metrics, other structures, exact strings/ordering and non-target relationships matched. Recorded Poppler text and Ghostscript rendering hashes match the immutable damaged baselines. The positive no-op controls passed before the damaged-input experiment; private staging, bounds, mutation/failure controls, and separate candidate publication were recorded. These are historical recorded checks accepted by the two reviewers; they were not rerun in this reconciliation.

The retained archives contain candidate and control reports, but no damaged-input validator reports. No before/after validator comparison was performed. Therefore the reports alone do not demonstrate inherited findings or absence of regression. The accepted inherited classification rests on separate source/object comparison and preservation evidence, including exact operation bounds, unchanged non-target structure/resources, the no-op record, and recorded text/render observations. These six findings are inherited fixture conditions within this operation only:

| Finding | Invoice failed checks | Statement failed checks | Boundary |
|---|---:|---:|---|
| Missing title-display preference (`DisplayDocTitle`) | 1 | 1 | Conformance blocker remains |
| Missing PDF 2.0 structure namespace | 1 | 1 | Conformance blocker remains |
| Unembedded Helvetica font program | 1 | 1 | Embedding rights/capability remain unevaluated |
| Missing Unicode mapping | 61 | 78 | Check counts do not rank user impact |
| Missing XMP metadata | 1 | 1 | Absence is explicitly preserved by the v2 contract |
| Untagged/unartifacted content | 9 | 9 | Not nine broken header associations |
| **Total** | **74** | **91** | **Six failed rules per candidate; both NONCOMPLIANT** |

Inheritance is accepted from the combined exact-operation preservation and independent source/object comparison, not from the report-only archives alone. Resolving these six findings requires a separately versioned proposed baseline and allowed-change set; do not retrofit the immutable v2 input, candidate, or preservation claim.

## Review disposition and gate boundaries

The independent AI-assisted technical reviewer accepted a narrowly bounded historical Proceed under `.1.6.3`/`.1.6.4`, with no technical blocker to that disposition. The independent AI-assisted product-value reviewer accepted the paired operation under the same criteria. The source-thread final assessment accepts the same scope. These assessments are not human semantic/AT approval, a product-owner signature, or release sign-off. The source document digests and durable excerpts are in the repository acceptance register. The acceptance does not waive any finding for conformance or eligibility.

At current main `e4ea5ad30eae19c9baa9a4df0439a323d4156e3a`, `.1.6.3` remains `in_progress`, `.1.6.4` remains `open`, and the `.1.6.4` issue acceptance text still says “six existing downstream dependencies.” The actual graph has **seven** consumers. Any authorized tracker update must preserve all seven edges:

`dap-jo9.2.1.3`, `dap-jo9.2.2`, `dap-jo9.2.3`, `dap-jo9.2.3.2`, `dap-jo9.2.4.2`, `dap-jo9.2.5.9`, and `dap-jo9.2.5.9.1`.

This evidence establishes no PDF/UA-2 compliance, eligible candidate, production writer/runtime adoption, automatic mapping, general-document support, human semantic/AT review, reviewer authority, or release approval. `.1.6.8`, `.1.6.9`, and `.5.4` are outside this task and unchanged.

## Tracker export-fidelity blocker

A separate build task (`01a13cb9-7067-72e0-83d8-d2bcc1bb50c0`, completed turn `01a13d74-83e2-71d5-84ad-83d561410456`) verified bd 1.3.1 source and a fresh 99-issue/166-edge round-trip. The import preserved the dependency graph, but export synthesized a missing `.5.5` `closed_at` as `2026-10-14T04:25:39` from `created_at` `2026-10-14T04:23:06` and `updated_at` `2026-10-14T04:25:38`; this derived timestamp is not factual closure evidence. The export also normalized unrelated fields. That round-trip was not reproduced in this executor; local `bd` unavailability here is a separate environment limitation. Do not hand-edit or publish the lossy export. Leave the intended `.1.6.3/.1.6.4` tracker transitions unapplied until an authoritative, fidelity-preserving update path is available.

## Source records

- [V2 positive no-op record](pdfcpu-financial-table-v2-noop-2026-10-13.md) and [v2 repair record](pdfcpu-financial-table-v2-repair-2026-10-13.md).
- [Historical disposition draft](../decisions/0005-historical-v2-bounded-association-proceed.md) and [durable acceptance excerpts](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md).
- Current issue acceptance/status was read from `.beads/issues.jsonl` at `ae2f5259361142575cb2a59811368195f9bbfcf9` and rechecked unchanged at `e4ea5ad30eae19c9baa9a4df0439a323d4156e3a`; the `.1.6.3/.1.6.4` acceptance and dependency fields are unchanged from the PR #73 merge revision.
- Historical workflow runs: [38465226725](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38465226725), [38466062420](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38466062420), [38505760173](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38505760173), [38505929576](https://github.com/pratik-saptarshi/document-accessibility-pipeline/actions/runs/38505929576).
