# Remediation component candidates

**Reviewed 2026-10-12; none are selected or approved.** This inventory records release-level license/capability boundaries to support the open-source remediation spike. “Latest” means the official release/tag visible on the review date, not a permanent recommendation.

| Candidate | Release and license | Verified fit and boundary |
|---|---|---|
| pdfcpu | [v0.16.2](https://github.com/pdfcpu/pdfcpu/tree/v0.16.2), Apache-2.0; Go-first | Full-write preservation characterization passed on the four synthetic fixtures; see [the evidence and test limits](pdfcpu-v0.16.2.md). No automatic semantic remediation selected. |
| OpenDataLoader PDF | [v2.5.12](https://github.com/opendataloader-project/opendataloader-pdf/releases/tag/v2.5.12), Apache-2.0 | OSS can read tags and auto-tag to Tagged PDF. PDF/UA export and visual editor are enterprise add-ons and excluded. Java/Python/Node packaging is not a Go library; integration and full transitive license/runtime footprint remain open. |
| Apache PDFBox | [3.0.8](https://pdfbox.apache.org/download.html), Apache-2.0 | Java PDF read/write APIs and preflight tooling; useful writer comparison, but this phase did not test its tag-tree preservation or confirm the complete transitive license footprint. |
| qpdf | [12.4.2](https://github.com/qpdf/qpdf/releases/tag/v12.4.2), Apache-2.0 | C++ PDF transformation and inspection tool. It can preserve structural objects during ordinary transformations, but is not a semantic tag authoring system; its `--remove-structure` feature would be inappropriate for preservation. Runtime dependency licenses and adapter behavior remain open. |
| veraPDF | [1.30.3](https://github.com/veraPDF/veraPDF-library/releases/tag/v1.30.3), dual GPL-3.0-or-later / MPL-2.0-or-later | Machine validation for PDF/A and PDF/UA; it does not establish correct meaning or replace human semantic and assistive-technology review. Do not embed until the chosen license option and complete package/dependency footprint are reviewed. |
| OCRmyPDF | [17.13.0](https://github.com/ocrmypdf/OCRmyPDF/releases/tag/v17.13.0), MPL-2.0 | Adds OCR text layers to scanned PDFs; not a semantic tagging or repair engine. It invokes external OCR/PDF components (including Tesseract; Ghostscript in some conversion paths); no OCR runtime is selected. |
| Tesseract | [5.5.3](https://github.com/tesseract-ocr/tesseract/tree/5.5.3), Apache-2.0 | OCR engine only. It provides recognized text, not PDF structure repair or reading-order decisions. Language data and Leptonica/runtime footprint still need verification before use. |

The Go characterization test declares `github.com/pdfcpu/pdfcpu v0.16.2` as the only direct third-party module in `go.mod`; Go also records the required indirect modules there. The resolved `go list -m all` graph at review time was checked against release/source license files and source headers; all identified modules use permissive licenses:

- Apache-2.0: pdfcpu v0.16.2; `xdg-go/stringprep` v1.0.4; Cobra v1.10.2; `inconshreveable/mousetrap` v1.1.0.
- MIT: `clipperhouse/uax29/v2` v2.7.0; `google/jsonschema-go` v0.4.3; `mattn/go-runewidth` v0.30; Goldmark v1.4.13.
- BSD-3-Clause: `hhrutter/tiff` v1.0.7; pflag v1.0.10; Go `x/crypto` v0.57.0, `x/image` v0.46.0, `x/text` v0.42.0, `x/mod` v0.41.0, `x/net` v0.58.0, `x/sync` v0.23.0, `x/sys` v0.48.0, `x/term` v0.46.0, `x/tools` v0.49.0, and `x/xerrors` 2019-07-17 revision.
- go-yaml v3.0.5 is dual MIT/Apache-2.0.

This is a source/module license inventory for the test graph, not a legal approval or a binary software bill of materials. Re-check the graph and packaged files if dependency versions or build targets change. No OCR, validator, enterprise feature, model, or other candidate was installed or selected for product use.
