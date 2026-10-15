# Financial-table positive no-op control: bounded no-go

**Current disposition (2026-10-14): ACCEPTED REJECT for the exact v1 pdfcpu v0.16.2 full-write operation.** The positive no-op control changes the Helvetica font dictionary on both v1 fixtures. This is not a general production-writer rejection. No v1 damaged input was opened, no v1 `/Headers` repair was attempted, no v1 candidate was produced, and this v1 result has no PDF/UA-2 outcome.

**Historical publication state:** this report was first issued on 2026-10-13 as a proposed finite disposition pending independent acceptance. The product owner and judge later accepted the bounded REJECT for this exact operation. The accepted disposition does not authorize widening the v1 preservation contract or closing the separate v2 evidence gate.

## Scope and execution

The attempt used the exact experiment-only clearance in [the candidate record](pdfcpu-financial-table-clearance-2026-10-13.md): pdfcpu v0.16.2, `api.LoadConfiguration` stateless mode, explicit offline operation and reviewed resource limits, `api.ReadContext`, `api.ValidateContext`, and full `api.WriteContext`. Go was `go1.27.2 linux/amd64`, `GOAMD64=v1`, `CGO_ENABLED=1`, with no build tags. Each run used a 30-second context, a 64 KiB input limit, a 256 KiB output cap, immutable positive input bytes, a separate mode-0600 file under private temporary staging, and a fresh read/validation of the written bytes. The input was never replaced. This is test-only use of the already pinned dependency.

The TDD preservation assertion first ran red on both fixtures. It correctly observed that the full writer adds all four predicted Helvetica keys. The executable control now passes only by recording and asserting that exact no-go outcome for both families. It also compares pre-write snapshots to fresh-read output for the declared table structures and verifies the IDTree, marked-cell text bindings, ParentTree, `/Pg`, `/P`, and `/StructParents` relations.

## Measured output delta

| Fixture | Immutable input SHA-256 / bytes | Private no-op output SHA-256 / bytes | Font delta |
|---|---|---|---|
| Statement | `14c0456e3b8ec0146bc2a25fa9af51d0c5559f8a53e23c4e163a3c422e66cfed` / 3,764 | `6fea2741fda4a6afc062c5a0d8e55a13736b907e5091ece52951e13a61bca39f` / 2,508 | Font object `5 0 R` (`/F1`) gains `/FirstChar 0`, `/LastChar 255`, a 256-entry `/Widths` array, and `/FontDescriptor` reference `(23 0 R)`. |
| Invoice | `9600519a8b3c64e7c48796dabb6f09cbe0c432bfe2daa4d2f6764937d3f2695e` / 3,710 | `674d898fe618be9baa885c611cf4161c43452d98400be7ebfb079471596688d5` / 2,466 | Font object `5 0 R` (`/F1`) gains the same four keys and values; the descriptor is `(23 0 R)`. |

Both outputs' width arrays have SHA-256 `6cb5449864ad91580d3eb294e278eaca115f354a611fb37e3b92056349874e65`. Their descriptor dictionaries have keys `Ascent`, `CapHeight`, `Descent`, `Flags`, `FontBBox`, `FontName`, `ItalicAngle`, `StemV`, and `Type`, with SHA-256 `92a177916a70f32f53ffa12561ce3b05210feb36199ffc1b9eab1279c98cc108`. The source Helvetica dictionaries have none of these four metrics. pdfcpu rewrote and compressed the files; output bytes and whole-file hashes therefore differ. No Info metadata fields changed (all remained empty).

The PDF inputs remain the only fixture identities. The two temporary full-write files are private no-op control outputs, not candidates. Their hashes describe this run; do not expect deterministic full-file bytes across separate writes.

## Review finding disposition

The accepted finite REJECT means that the v1 full-write operation cannot proceed under its recorded preservation contract. It does not reject pdfcpu for all use, select an alternative writer, establish general feasibility, or satisfy any v2 candidate/release gate. The v2 positive no-op controls passed, actual repaired candidates were produced, and their candidate-bound machine assessments were NONCOMPLIANT; see the [v2 repair result](pdfcpu-financial-table-v2-repair-2026-10-13.md). At this note's historical evidence snapshot, the latest full raw XML/contracts were not durably available for review.

The [PR #45 P2 review comment](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/45#discussion_r4261643587) was valid: the first oracle version checked only for the presence of the four predicted keys, so changes or removals to existing font entries and added font resources could escape detection. The control now snapshots every page font resource, its indirect identity, every complete font dictionary entry, and the resolved values of indirect entries. It accepts only the exact four-key `/F1` expansion, then checks the reviewed values and representation (including the 256 widths, descriptor reference, complete descriptor key set, and value digests). Targeted mutations for both statement and invoice reject changed and removed existing entries, added font resources, changed direct/indirect width arrays, changed indirect descriptor values, changed metric values, and representation changes. The output staging uses `t.TempDir()` for automatic cleanup. These changes improve the characterization oracle only; the measured font-metric delta remains a no-go, and no damaged input was opened.

## Independent preservation observations

The Go fresh-read structural assertions passed for both outputs: header IDs/scopes, row and cell order, exact cell strings (including `$12.50`, `$8.00`, `-$18.25`, and `$42.10`), every `/Headers` association, IDTree resolution, MCID text, and ParentTree bindings were unchanged. Independent Poppler `pdftotext -layout` output byte-matched the retained positive-fixture text for each family. Independent Ghostscript 10.05.1 renders (`png16m`, 72 dpi) byte-matched the retained positive-fixture PNGs. Each output remained one letter-size page (612 × 792 pt).

| Family | Extracted text SHA-256 | Rendered PNG SHA-256 |
|---|---|---|
| Statement | `b20dd862dda3e51f2d65927f03a5df7ae01338acc76dda6f155e3f1a0b36b34b` | `7c9531c6b51d04423f5d46ce28c65584e35a00983636d49b3dad895275d1785e` |
| Invoice | `35fa4c4506c765eac234110017cc0592e0704bb3b51ae34589aac36d40485ea0` | `d1ccd35abd66f7f33af61c4f8198f08ba36bc4a9444a04e7a29dbbc927fd2718` |

## Disposition and next decision

The added font metrics are a protected structural delta under the reviewed protocol. Stop before opening either damaged fixture. Do not normalize the source, waive the font delta, mutate a positive input toward the damaged state, or infer that unchanged appearance/text makes the full write acceptable. The product owner and judge have accepted **REJECT for this exact v1 operation**; that disposition is final for the recorded metric expansion and preservation contract, with no further acceptance pending. At the historical evidence snapshot for this note, the separate v2 evidence gate was described as open because the latest full raw XML/contracts were not available for independent inspection. This accepted v1 result does not authorize production dependency adoption or imply general pdfcpu fidelity.

**V2 evidence reconciliation (2026-10-15):** the two preceding v2 report-availability statements describe the earlier evidence snapshot and are superseded: four retained bundles and their complete candidate-bound reports/contracts have since been reconciled and independently reviewed. The final assessment records a finite historical Proceed only for the paired row-two `/Headers` repairs; both candidates remain NONCOMPLIANT. See the [reconciled evidence ledger](pdfua2-retained-bundle-ledger-2026-10-15.md), [proposed Decision 0005](../decisions/0005-historical-v2-bounded-association-proceed.md), and [acceptance register](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md). This does not alter the accepted v1 **REJECT** or establish v2 candidate eligibility, PDF/UA-2 conformance, production adoption, human semantic/assistive-technology acceptance, or release. `.1.6.3` remains in progress and `.1.6.4` remains open; tracker transitions are unapplied.

No candidate-bound `ua2` job was run; Java and veraPDF remain CI-only. This no-op is not a repair result and does not satisfy successful-candidate evidence gates.
