# ariapdf generation source clearance and bounded spike plan

**Date:** 2026-10-15 UTC

**Status:** Source pin and source-level package license inventory are verified. Interpretation and notice handling for the bundled font in generated PDF outputs remain open. Generation-profile and test-fixture acceptance also remain open. No prototype, dependency change, PDF, or conformance result is authorized or represented here. The owner-selected lead is ariapdf for evaluation only; production engine selection remains open.

## Decision boundary

The owner selected ariapdf as the lead candidate to evaluate. This replaces the earlier suggestion to begin with boxesandglue; it does not select a production engine, approve a runtime dependency, close `dap-jo9.1.5`, or establish PDF/UA-2 conformance. Generation remains separate from existing-PDF remediation. No remediation priority or profile is copied into this generation lane.

This pass retrieved and statically inspected the official source archive and dependency metadata. It did not run candidate code, install dependencies, build, execute upstream CI, make a PDF, or invoke Java/veraPDF. Treat source statements and upstream tests as source evidence, not observed output behavior.

## Source and version evidence

The official GitLab repository API identifies `main` at `499ef3f23e8f1a4b8b6c368a3c497d8b94c1dbef`; the repository's tag list contains `v0.9.0` at the same commit and no higher version tag. The commit was created 2026-06-29. The immutable GitLab archive for that commit was retrieved read-only; SHA-256: `816776bb1559e934a4684f4d634d453d8c9e33ba0fcc97d1d2c4f0696278836b` (122,826 bytes). The raw `go.mod` and `LICENSE` at the tag agree with the archive.

This establishes `v0.9.0` as the latest tag observed in the official repository at review time, and establishes that `main` currently resolves to the same source. GitLab's Releases API returned HTTP 403 in this environment, so a separate formal Release record/date was not verified. Do not describe package-index metadata as proof that v0.9.0 is the newest release.

| Component | Version / pin | Role and license evidence | Status |
|---|---|---|---|
| `gitlab.com/andrehehn/ariapdf` | `v0.9.0`, commit `499ef3f23e8f1a4b8b6c368a3c497d8b94c1dbef`; archive SHA-256 above | Root `LICENSE` is MIT, copyright André Hehn, 2026. | Permissive for evaluation, with notices retained. No product adoption. |
| `golang.org/x/image` | `v0.43.0`; go.sum `h1:FLxcP4ec2350nTfOC8ysKtqYSIFbk/QGjw1ZHNP4tsY=` | Required by ariapdf; BSD-3-Clause root license. Its `font/sfnt` imports `golang.org/x/text/encoding/charmap`. | Permissive; carry copyright/license notices with redistributed software. |
| `golang.org/x/text` | `v0.38.0`; go.sum `h1:sXmwo9DwP3OK9EZ7PqAdaooSGozfl/3a6/xJcbzPRhE=` | Indirect in ariapdf; BSD-3-Clause root license. `encoding/charmap` is in the inspected production import closure. | Permissive; carry notices. |
| `golang.org/x/tools`, `golang.org/x/mod`, `golang.org/x/sync` | `v0.45.0`, `v0.36.0`, `v0.21.0` | Declared by x/text's `go.mod` for its development/tooling graph; official root licenses are BSD-3-Clause. They were not found imported by ariapdf's selected `x/image/font`, `font/sfnt`, `font/gofont/goregular`, or `math/fixed` closure. | Record in the full module graph; do not misstate them as demonstrated ariapdf runtime imports. Verify the consumer module graph in any later isolated spike. |

The ariapdf `go.mod` declares Go `1.25.0`, x/image `v0.43.0`, and x/text `v0.38.0`; its `go.sum` records both module and `go.mod` checksums. The repository declares toolchain `go1.27.2`, which is above ariapdf's declared Go floor; no compatibility build was run.

### Fonts, resources, and notices

The source archive contains 57 Go files, one root license, documentation/configuration, and no bundled standalone font, image, ICC profile, or other non-code runtime resource. No `go:embed`, cgo source, or product-source import of `net`, `net/http`, `os/exec`, or `syscall` was found. Test/example code does use `os/exec`; it is not part of the library's product runtime.

`font.Default()` imports Go Regular from `golang.org/x/image/font/gofont/goregular`; that font's bytes are compiled into the dependency as Go source. ariapdf's `font.Embed` source writes a subset of the font program into the PDF output. The upstream Go-font notice identifies Bigelow & Holmes Inc. (2016) and permits source and binary redistribution subject to retaining notices; for binary redistribution it requires reproducing the copyright notice, conditions, and disclaimer in documentation or other accompanying materials. That text supports a notice requirement when the font itself is redistributed in binary form. It does **not**, by itself, settle whether a generated PDF containing a subset font constitutes binary redistribution of the font, nor where a notice must be presented for such a PDF. Treat output-distribution clearance and notice placement as **unresolved** pending a reviewed interpretation; do not claim that generated PDFs either do or do not trigger the clause. Software distributions retain the x/image and x/text notices. Any caller-supplied TTF/OTF/WOFF needs its own provenance, embedding permission, and notice review; it is outside this candidate-source inventory. WOFF2 is rejected by the candidate source.

The reachable source import set is Go standard library plus x/image and x/text. ariapdf's `.gitlab-ci.yml` runs Go formatting, vet, lint and race tests, and describes a separate veraPDF job and manual release job. It uses a mutable `golangci-lint:latest` image and fetches its validator installer in that upstream CI configuration. This was not executed or adopted. The source archive contains no `SECURITY.md`; security advisory history, maintainer response expectations, and reproducible release attestations were not established in this pass. The project was created 2026-06-26, its latest observed activity was 2026-07-04, and its tags are concentrated between 2026-06-27 and 2026-06-29. This is limited maintenance history, not evidence of a vulnerability or a security audit.

**License disposition:** The ariapdf source and inspected software module licenses are permissive on the evidence reviewed. That source-level result is sufficient to continue clearance review, but not to add a dependency or clear generated-output distribution. The embedded-font redistribution question and notice mechanism remain open; caller-supplied fonts remain outside scope. No legal conclusion is made. The formal release-object lookup and fuller security/maintenance review also remain open.

## Source-level behavior to challenge

Documented APIs include `document.New`, block-level headings/paragraphs/tables/lists/figures, `Render() ([]byte, error)`, reusable templates, and UA-1/UA-2 flags. Source shows a simple facade table creates `TH` cells with `Scope=Column`; that is not proof that serialized associations work for the proposed family outputs or that complex tables are supported.

The following source behaviors make explicit regression tests necessary:

- `font.Encode` maps an unmapped rune to glyph 0 (`.notdef`); the library does not make that a preflight rejection. Missing glyphs in financial strings must be caught by the application before render.
- `font.Wrap` uses `strings.Fields`, collapsing runs of whitespace, and hard-breaks long words between runes. Preserve exact requested financial text separately from any line-wrapped drawing representation; test spaces, tabs, nonbreaking spaces, combining marks, and long identifiers.
- `document.HTML` is lenient and has no error return for malformed or unknown markup; unknown tags can be transparent and whitespace is collapsed. Do not treat this as a verified DOCX/Markdown/source converter or allow silent degradation.
- Block-time failures can be sticky and surface only at `Render`. The caller must propagate the render error and produce no apparent success/partial artifact.
- `Doc` and `font.Font` are documented as mutable and not safe for concurrent use; templates create per-document font state, but the actual isolation must be tested. Use one document per render and do not share mutable font instances across concurrent renders.
- Complex-script shaping, bidi, fallback fonts, emoji, WOFF2, complex/spanning/multi-level tables, and malformed-data behavior remain unsupported or unverified. Do not imply support from generic structure APIs.

## Bounded test-first generation spike proposal

This is a proposed experiment outline, not an approved profile. The companion [bounded ariapdf evaluation-profile proposal](../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md) defines the same reduced fixture scope and makes its test values and boundaries reviewable; neither document supersedes the canonical acceptance cases. Do not select production language, currency, tax, rounding, date, confidentiality, retention, semantic-policy, or reviewer defaults in this note.

### Fixtures and order

Keep all four required families. Build failing fixture/oracle tests before a renderer adapter or behavior is written. The statement and invoice cases are the two equally deep first cases; correspondence and report are small smoke cases, not dropped families.

1. **Statement renderer spike fixture (`ARIAPDF-SPIKE-STATEMENT-001`):** use the proposed profile's exact two-row synthetic statement fixture as a one-page baseline. It contains the title, two transaction rows, and three column headers only: no opening or closing balance, long description, or multi-page content is part of this fixture. It maps to canonical `TC-GEN-STATEMENT-001` as partial coverage only; it does not exercise that case's proposed field/arithmetic profile, complete source-to-output manifest, or separate semantic structure review.
2. **Invoice renderer spike fixture (`ARIAPDF-SPIKE-INVOICE-001`):** use the proposed profile's exact two-row synthetic invoice fixture as a one-page baseline. A test-only identifier is optional; no supplied invoice total, currency identity, tax, subtotal, or rounding input is part of this fixture. It maps to canonical `TC-GEN-INVOICE-001` as partial coverage only; it does not exercise declared currency/rounding rules or verify totals.
3. **Correspondence smoke fixture (`ARIAPDF-SPIKE-CORRESPONDENCE-001`):** one proposed positive example of declared language, recipient, and title fields, with exact source text, one link/target, a short list, and a fixed reading-order oracle. It maps to canonical `TC-GEN-CORRESPONDENCE-001` as partial coverage only; it does not exercise missing-binding outcomes or a complete required-field matrix.
4. **Report smoke fixture (`ARIAPDF-SPIKE-REPORT-001`):** headings, paragraph, simple table, figure with authored alt text, and header/footer/page furniture. It maps to canonical `TC-GEN-REPORT-001` as partial coverage only; it does not exercise canonical provenance metadata or unmapped-value handling.

Use independent expected text, order, link-target, and association oracles defined in the [proposed evaluation profile](../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md), never reconstructed from candidate output. That profile now spells out proposed correspondence and report text and structure as well as the two financial fixtures; all values remain unapproved synthetic proposals. Before owner approval, freeze and hash the complete source fixture and oracle bundle for all four families. Candidate execution requires owner approval of that exact bundle; if any details were approved before they were defined, obtain renewed approval after definition. The two financial baselines are one page each and use the listed values verbatim. Any overflow or pagination characterization is separate and cannot add a financial field or imply pagination support. Profiles remain `proposed`; no fixture values become production rules.

### Spike IDs and canonical coverage boundary

All reduced renderer fixtures and the control suites mapped to canonical cases use spike-only IDs. The complete mapping and coverage limits are stated in the [proposed evaluation profile](../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md); in particular, `ARIAPDF-SPIKE-INVALID-DATA-001` maps only partially to canonical `TC-GEN-INVALID-DATA-001`, and `ARIAPDF-SPIKE-VALIDATION-001` maps only partially to `TC-GEN-VALIDATION-001`. The latter does not replace the separate semantic or assistive-technology outcomes. HTML-behavior and mutable-state characterization controls are outside those canonical mappings and must not be reported under a `TC-GEN-*` ID.

Record outcomes under spike-only IDs. A reduced-spike pass is partial evidence for its mapped canonical case; it must never be reported as a canonical `TC-GEN-*` pass or used to close that case. The canonical cases remain specification-only with their complete acceptance obligations unchanged. This plan neither changes canonical Beads ownership nor closes any case.

### Pre-prototype gates

1. Complete all four synthetic source fixtures and independent oracles, including exact correspondence text/link target, report text/table associations, and expected reading order. The linked profile contains proposed values only; do not treat them as approved.
2. Obtain owner approval of the immutable, hashed profile/fixture/oracle bundle before candidate execution. If any fixture detail or oracle was approved only at a high level before it was fully defined, seek renewed approval after completion. Any subsequent change requires renewed approval.
3. Verify the exact tagged source archive and module graph again at the isolated spike's start; pin source and dependency hashes. Retain MIT/BSD software/font notices and approve how notices accompany generated PDFs.
4. Define the bounded test subset: the two one-page financial fixtures with simple one-row headers and `Scope=Column`, plus the proposed one-page correspondence and report fixtures. Explicitly defer pagination and repeated page-header behavior, complex associations/spans, fallback/shaping, general HTML conversion, arbitrary documents, and caller fonts unless independently cleared.
5. Author failing tests from the owner-approved, frozen oracles before any renderer adapter or behavior. Candidate execution must not precede approval and test readiness. Verify mutation oracles detect changed text, link target, amount, table association, and order.

### Candidate tests and bounded execution

After all gates above pass, run a throwaway, pinned Go-only spike isolated from product runtime and repository dependency files. The smallest candidate path should use ariapdf's documented block API and a fresh document/font instance per render. Do not use `HTML` as a source conversion shortcut. Record toolchain, source/module hashes, fixture/output hashes, limits, and all failures.

- **Data/text:** compare exact fixture text and test repeated spaces, tabs, NBSP, combining sequences, and long unbroken identifiers as valid inputs. Exact-preserving output passes these checks. Fail closed only when safe preservation is impossible within the accepted bounded profile; never normalize, truncate, substitute, or drop content. Missing glyph coverage or `.notdef` must also fail closed. These checks do not add a multi-page financial acceptance case or claim general text support.
- **Structure:** independently inspect serialized PDF version/metadata, reading order, headings, table `TH`/`Scope` relationships and each body-cell association, one-page figure alt text, links, page furniture, and ParentTree/MCID bindings. Repeated page headers remain outside this subset. Include mutation tests for amount changes, missing content, association swaps, and order changes.
- **Failure safety (`ARIAPDF-SPIKE-INVALID-DATA-001`):** invalid profile values, invalid font data, unsupported topology, missing alt text/font, delayed render errors, oversized input/output, cancellation at the wrapper boundary, and cleanup/atomic-write failure must fail closed without retaining a partial artifact. This is partial evidence only for canonical `TC-GEN-INVALID-DATA-001`; it does not implement the full missing-required-field/wrong-type validation and field-level findings contract.
- **HTML behavior:** characterize malformed/unknown markup and the no-error-return behavior as a negative control. If it can silently change requested content, mark the HTML path unsupported; do not weaken the test or silently accept output.
- **State:** test that each concurrent render owns a separate `Doc` and mutable font state; verify template-spawned instances do not leak glyph subsets. Race tests cover only the tested paths.
- **Validation (`ARIAPDF-SPIKE-VALIDATION-001`):** use the project's existing isolated CI validator path with explicit `ua2`, candidate-hash-bound reports, and valid positive/negative controls. This is partial evidence only for canonical `TC-GEN-VALIDATION-001`; do not report it as a canonical pass or close semantic/assistive-technology gates. Do not run Java/veraPDF locally, do not use profile fallback, and do not convert a validator pass into a semantic or conformance claim.

End the experiment with `proceed`, `reject`, or `inconclusive` for this exact subset. A source/license blocker or unsupported test result is a valid bounded outcome. Keep human semantic review and genuine assistive-technology review as separate evidence; neither is replaced by code inspection, extracted text, race tests, or veraPDF.

## Alternatives and non-adoption

| Candidate | Preserved finding | Current disposition |
|---|---|---|
| [boxesandglue](https://github.com/boxesandglue/boxesandglue) | Existing source review supports only a simple `Scope`-based header scope. Complex `/Headers`, spans, IDTree associations and complete resource/font/ICC terms remain unresolved; do not treat the simple case as broad capability. | Keep as an alternative with unresolved resource and complex-association work. It is not selected. |
| [GoPdfSuit](https://github.com/chinmay-sawant/gopdfsuit) | Header semantics and serialized associations remain unproven in this project's acceptance model. | No replacement recommendation; exact source/license/capability clearance would be separate work. |
| [Typst](https://typst.app/) | Official material identifies PDF/UA-1 support, which does not meet the accepted PDF/UA-2 target. | Does not satisfy the unchanged target on current evidence. |

The comparison is not a rank ordering. Preserve these limitations in the next relevant PR; do not silently switch the user-selected ariapdf evaluation lead.

## Blockers and next decision

- GitLab's formal Releases endpoint returned 403; the latest tag and branch identity are known, but a formal Release record was not independently read.
- Security advisory posture and maintainer support commitments remain unknown; the short observed project history and upstream mutable CI tool reference merit review before adoption.
- Product-accepted generation profile/fixture fields and independent expected semantics are still open. This blocks prototype execution even though the candidate source is now retrievable and the inspected software licenses are permissive.
- DOCX/Markdown source conversion remains a separate `dap-jo9.3.2` gate; this plan tests structured generation inputs only. Existing case specifications remain in force and are not closed by this source note.
- Legal interpretation and notice handling for the Go font subset embedded in generated PDFs remain open; caller-supplied fonts require separate provenance and embedding-license clearance.
- Complex typography/tables, source conversion, machine conformance, human semantic review, and genuine AT usability remain distinct gates.

The canonical `.beads/issues.jsonl` row at this base records `dap-jo9.1.5` as open. A read-only `bd show` could not confirm live database state: the embedded Dolt store reports an incomplete database directory, and the default Dolt home is read-only. No reset or manual export edit was attempted; live tracker state remains unverified beyond the canonical export. This note contributes source-pin/license evidence and a proposed bounded plan only. It does not change acceptance criteria, the generation profile, the Beads graph/export, or production selection. Do not hand-edit or publish an export from an isolated Beads database.

## Evidence references

- [Official ariapdf project](https://gitlab.com/andrehehn/ariapdf)
- [Official repository tag list](https://gitlab.com/api/v4/projects/83785085/repository/tags?per_page=100)
- [Official `main` branch metadata](https://gitlab.com/api/v4/projects/83785085/repository/branches/main)
- [Immutable source archive at the selected commit](https://gitlab.com/andrehehn/ariapdf/-/archive/499ef3f23e8f1a4b8b6c368a3c497d8b94c1dbef/ariapdf-499ef3f23e8f1a4b8b6c368a3c497d8b94c1dbef.tar.gz)
- [v0.9.0 `go.mod`](https://gitlab.com/andrehehn/ariapdf/-/raw/v0.9.0/go.mod) and [root LICENSE](https://gitlab.com/andrehehn/ariapdf/-/raw/v0.9.0/LICENSE)
- [x/image v0.43.0 license](https://github.com/golang/image/blob/v0.43.0/LICENSE) and [raw Go font notice](https://raw.githubusercontent.com/golang/image/v0.43.0/font/gofont/ttfs/README)
- [x/text v0.38.0 `go.mod`](https://github.com/golang/text/blob/v0.38.0/go.mod) and [license](https://github.com/golang/text/blob/v0.38.0/LICENSE)
- [x/tools v0.45.0](https://github.com/golang/tools/tree/v0.45.0), [x/mod v0.36.0](https://github.com/golang/mod/tree/v0.36.0), and [x/sync v0.21.0](https://github.com/golang/sync/tree/v0.21.0) license files
- [Existing Go-native candidate comparison](go-native-generation-candidates.md), [generation profile schema](../workflows/generation/profiles/profile.schema.json), [generation family examples](../profiles/examples/generation/), and [accepted PDF/UA-2 target](../decisions/0001-pdf-ua-2-validation-target.md)
- [Existing acceptance case specifications](../validation/cases.md) and [bounded evaluation-profile proposal](../profiles/proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md)
