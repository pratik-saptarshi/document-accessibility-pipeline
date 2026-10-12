# Existing-PDF remediation workflow

**Status: Proposed; open-source-only dependencies.** The first technical gate is an existing-PDF writer/fidelity feasibility spike.

1. Validate the job and data/security profile; reject unsupported or unsafe input.
2. Inspect and characterize the PDF before changing it. Preserve a baseline of visible content and relevant links, forms, annotations, fonts, metadata, and geometry.
3. Route candidate work through replaceable worker adapters behind the Go-centered orchestration/audit/semantic-repair boundary.
4. Apply bounded structural or semantic repairs only where the source supports a defensible decision. Route ambiguous semantics to human review rather than guessing.
5. Write output using a verified open-source path, compare preservation assertions, and retain audit evidence.
6. Run independent machine validation, human semantic review, and assistive-technology checks before any release claim.

## Test-first slices

Begin with synthetic/deidentified characterization fixtures and failing preservation assertions. Add unit, adapter contract, integration, and end-to-end tests before implementation; use red/green/refactor for each vertical slice. Inject worker failures and hostile inputs. Keep scanned/OCR cases selective and subject to separate quality review.

## Dependency boundary

The current notes list OpenDataLoader OSS auto-tagging, PDFBox custom writing, veraPDF machine validation, qpdf structural operations, and selective OCRmyPDF/Tesseract as candidates. Verify current version, license, capability, and integration behavior during spikes. Enterprise-only tagging/PDF/UA export is excluded; no candidate is an accepted dependency yet.
