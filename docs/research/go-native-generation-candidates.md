# Go-native generation candidates

**Source review: 2026-10-12. Preliminary; no engine is selected or approved.** Go has real PDF writers; this note records a read-only source comparison, not a build, integration test, fixture result, or PDF/UA conclusion.

## Candidate evidence

| Candidate | Source-level evidence | Remaining gate |
|---|---|---|
| [boxesandglue at `d2a6da76`](https://github.com/boxesandglue/boxesandglue/tree/d2a6da76abc262aacf7988e7297d6b66ef181fe6) | Its [document model and writer](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/backend/document/document.go) include PDF 2.0 structure namespaces, RoleMapNS, ParentTree, MCID/MCR/OBJR, Alt/ActualText, simple table Scope, associated files, and ordered content interleaving. | The typed model reviewed lacks explicit complex-table Headers/ID and RowSpan/ColSpan fields. Existing-page import is Form-XObject import, not evidence of safe remediation. The API is pre-production and upstream's veraPDF example claim was not reproduced. Test tables with spans, cross-page reading order, figures, links, multilingual text, and ActualText before considering selection. |
| [ariapdf v0.9.0 structure docs](https://pkg.go.dev/gitlab.com/andrehehn/ariapdf/structure) | Package docs describe namespace/ParentTree support, link OBJR, table Scope, Alt, and ToUnicode. | A complete pinned source, dependency, license, and maintenance review remains unverified. Missing glyphs may fall back to `.notdef`. Treat the top-level MIT label as candidate metadata, not a completed dependency clearance. |

## License and integration boundary

The boxesandglue repository's [top-level license](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/License.md) is Modified BSD (three-clause). That does not clear the full frontend: its generated [`hyphenationpatterns.go`](https://github.com/boxesandglue/boxesandglue/blob/d2a6da76abc262aacf7988e7297d6b66ef181fe6/frontend/hyphenationpatterns.go) embeds Hungarian TeX pattern data matching upstream [`hyph-hu.tex`](https://github.com/hyphenation/tex-hyphen/blob/5684c0f51c0b81133db2efbe60a408b4155a3ff5/hyph-utf8/tex/generic/hyph-utf8/patterns/tex/hyph-hu.tex), whose license options include MPL-1.1, GPL-2.0, and LGPL-2.1. Audit backend-only package reachability and packaged data before considering the permissive-only boundary. Do not import the full frontend under a permissive-only claim until that review is complete.

## Gate

boxesandglue is the strongest source-inspected candidate so far, not an adopted engine. Continue to prefer established libraries over a custom PDF engine, while keeping generation separate from existing-PDF remediation. Use synthetic cases only. No source-to-output behavior, real validator run, or conformance result is established here; the accepted PDF/UA-2 target still requires authenticated validation and human semantic/assistive-technology review.
