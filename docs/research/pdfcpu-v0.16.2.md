# Existing-PDF write feasibility: pdfcpu v0.16.2

**Status: candidate evidence only. Not a production selection or remediation implementation.** Reviewed and exercised 2026-10-12. Source tag `v0.16.2` resolves to commit `b41a41d5323f6119ff6a183ad3f287eb4a20a617`; the release `go.mod` requires Go 1.26.0 or later. The release license is Apache-2.0. Sources: [pinned source tag](https://github.com/pdfcpu/pdfcpu/tree/v0.16.2), [release module file](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/go.mod), [release license](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/LICENSE.txt).

## Bounded full-write characterization

The repository harness is `internal/pdfpreserve/roundtrip_test.go`. It creates synthetic, one-page fixtures for statements, invoices, correspondence, and reports; each contains marked content, a deliberately non-monotonic MCID sequence, a sparse ParentTree with a null slot, `/P` parent references, an MCR with `/Pg`, an MCR with `/Pg` and `/Stm`, a link annotation/OBJR, a text widget/OBJR with a field value, a base font, page geometry, and title metadata. It performs `api.ReadContextFile` → `api.WriteContextFile` with the ordinary full-write path → fresh `ReadContextFile`.

Observed on all four fixtures: page count, media box/page dictionary, page content and form-XObject decoded streams, font dictionary, link and widget dictionaries, structure dictionaries, ParentTree entries and null hole, document title, and Poppler's rendered pixels, text, tag-tree view, and link target are unchanged. The form field value remains `42.00`. In this tiny fixture, rendered PNG hashes match exactly at 72 DPI. The writer changes the Producer and adds creation/modification timestamps; the title is retained. These are characterization results for the fixture only, not a general fidelity threshold.

The fixture covers sparse/non-monotonic MCIDs, `/P`, `/Pg`, `/Stm`, annotation/widget OBJR, and a null ParentTree slot. It does not establish correctness for multiple pages, arbitrary third-party structure trees, signatures, embedded files, OCR, complex fonts, tagged forms/appearance streams, or all annotations. The small form widget has no appearance stream. Visual comparison was run with the workspace Poppler `pdftoppm` at 72 DPI; the committed Go test does not require Poppler, so image/text/tag-view comparisons are recorded here as a workspace run, not as CI assertions.

## Source-reviewed behavior and limits

At the pinned release, the normal full writer serializes the catalog `StructTreeRoot`, `AcroForm`, page annotations, and `StructParents`. Reduced-feature modes remove `StructTreeRoot`; therefore, the static-review risk is mode-specific. `WriteIncrement` is an incremental-writing primitive, not a complete semantic remediation API. A candidate adapter must choose the ordinary full-write path explicitly and test its mode settings. Source evidence: [`write.go` at v0.16.2](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/pkg/pdfcpu/write.go).

## Dependency evidence and remaining gate

The library itself is Apache-2.0 and is Go-first. Its release `go.mod` lists direct and indirect modules, including Cobra, YAML, x/crypto, x/image, x/text, TIFF, and Unicode helpers. This Phase 0 harness pins pdfcpu exactly at v0.16.2; it does not make the library an approved runtime dependency. Before any production adoption, record and review the full resolved transitive license inventory, security posture, supported versions, and compatibility with the approved input profile. The initial Go proxy fetch returned 403 in this environment; the module was subsequently downloaded from the exact official VCS tag, and `go.sum` records its checksums. The checked-in module dependency uses the tagged version and standard Go checksum verification.

## Proposed next test boundary

Keep remediation and generation adapters replaceable. Before choosing this writer, extend characterization to multi-page tagged examples, page-tree inheritance, form appearances, annotation variants, font embedding/subsets, XMP/Info metadata policy, signatures, and failure/unsupported behavior. A successful parse or round-trip is not a semantic repair result, PDF/UA validation, WCAG conformance, or accessibility release approval.
