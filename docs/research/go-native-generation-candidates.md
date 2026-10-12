# Go-native generation candidates

**Source review: 2026-10-12. Preliminary; no engine is selected or approved.** Go has real PDF writers; this note records a read-only source comparison, not a build, integration test, fixture result, or PDF/UA conclusion.

## Candidate evidence

| Candidate | Source-level evidence | Remaining gate |
|---|---|---|
| [boxesandglue at `d2a6da76`](https://github.com/boxesandglue/boxesandglue/tree/d2a6da76abc262aacf7988e7297d6b66ef181fe6) | Its [document model and writer](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/backend/document/document.go) include PDF 2.0 structure namespaces, RoleMapNS, ParentTree, MCID/MCR/OBJR, Alt/ActualText, simple table Scope, associated files, and ordered content interleaving. | `StructureElement.ID` is an integer and is not serialized as PDF `/ID`; no `Headers`, `RowSpan`, `ColSpan`, or IDTree support was found. Existing-page import is Form-XObject import, not evidence of safe remediation. The API is pre-production and upstream's veraPDF example claim was not reproduced. A bounded experiment might cover H1/P, single-header tables with Scope, basic links/figures, and isolated single-script runs; complex tables, bidi, and fallback need more library/application work. |
| [ariapdf v0.9.0 structure docs](https://pkg.go.dev/gitlab.com/andrehehn/ariapdf/structure@v0.9.0?tab=doc) | Package docs describe namespace/ParentTree support, link OBJR, table Scope, Alt, and ToUnicode. | The [v0.9.0 font docs](https://pkg.go.dev/gitlab.com/andrehehn/ariapdf/font@v0.9.0?tab=doc) describe ToUnicode mapping and `.notdef` fallback behavior. A complete pinned source, dependency, license, and maintenance review remains unverified. Treat the top-level MIT label as candidate metadata, not a completed dependency clearance. |

## License and integration boundary

The boxesandglue repository's [top-level license](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/License.md) is Modified BSD (three-clause). The inspected `backend/document` + font + node import closure reaches eight in-module packages, including the separate `frontend/pdfdraw` leaf, but not the `frontend` root; it therefore avoids the root package's generated [`hyphenationpatterns.go`](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/frontend/hyphenationpatterns.go) and its Hungarian TeX data matching upstream [`hyph-hu.tex`](https://github.com/hyphenation/tex-hyphen/blob/5684c0f51c0b81133db2efbe60a408b4155a3ff5/hyph-utf8/tex/generic/hyph-utf8/patterns/tex/hyph-hu.tex). This narrows the frontend concern but does **not** permissively clear the backend closure.

[`backend/document/colorprofile.go`](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/backend/document/colorprofile.go) unconditionally embeds `ISOcoated_v2_eci.icc` and `sRGB2014.icc`. The sRGB profile's embedded copyright and official ICC terms were reviewed as permissive. The exact 1.8 MB ECI binary and license tag were not inspected; official ECI distribution/publisher materials raise modification and redistribution questions, so its terms remain **unverified**, not conclusively nonpermissive. A PDF/UA-2 output path not needing ICC does not prove the compiled or distributed package excludes these embedded resources.

The reachable `textshape` v0.0.18 source at commit [`03f5f92`](https://github.com/boxesandglue/textshape/tree/03f5f92cf09df07f3ea9374de6f6ca93d9cfa7ce) has a root MIT label, but its bundled HarfBuzz/Unicode-derived notice and provenance review is incomplete. [`ot/indic_machine.go`](https://github.com/boxesandglue/textshape/blob/03f5f92cf09df07f3ea9374de6f6ca93d9cfa7ce/ot/indic_machine.go) retains copyright without the full upstream permission/disclaimer; generated Unicode tables still need exact version and license mapping against the [Unicode terms](https://www.unicode.org/license.txt). These are notice/provenance gaps, not evidence of GPL or AGPL code.

Complete pinned font/image licenses and determine source/build resource exclusion or permission before adoption. Do not assume a no-ICC output path excludes compiled resources. No new dependency or nonpermissive exception is approved.

## Capability limits

Source inspection found `font.Shape` guesses one script/direction per buffer. A bounded experiment could use H1/P, single-header simple tables with Scope, basic links/figures, and isolated single-script runs; complex tables, bidirectional text, and fallback behavior require more library/application work. These limits are not a general existing-PDF remediation result.

## Gate

boxesandglue is the strongest source-inspected candidate so far, not an adopted engine. Continue to prefer established libraries over a custom PDF engine, while keeping generation separate from existing-PDF remediation. Use synthetic cases only. No build or runtime validation was executed. No source-to-output behavior, real validator run, or conformance result is established here; the accepted PDF/UA-2 target still requires authenticated validation and human semantic/assistive-technology review.
