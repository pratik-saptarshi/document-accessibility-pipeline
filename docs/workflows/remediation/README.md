# Existing-PDF remediation workflow

The proposed source characterization contract is documented in [remediation profiles](profiles/profile.schema.json), with [synthetic family examples](../../profiles/examples/remediation/). Its shared family enum follows the owner-approved set in [Decision 0017](../../decisions/0017-approved-profile-family-set.md): statements, invoices, correspondence, reports, presentations, forms, marketing, and HR documents (`hr_documents`). The enum does not claim end-to-end support; the current examples cover statements, invoices, correspondence, and reports. The profile describes observed or unknown source fields; it does not authorize edits or resolve semantic ambiguity.

**Status: Proposed; product dependencies require permissive open-source licenses.** The veraPDF validator remains isolated to CI as accepted in Decision 0001; no additional runtime exception is approved. The first technical gate is an existing-PDF writer/fidelity feasibility spike.

1. Validate the job and data/security profile; reject unsupported or unsafe input.
2. Inspect and characterize the PDF before changing it. Preserve a baseline of visible content and relevant links, forms, annotations, fonts, metadata, and geometry.
3. Route candidate work through replaceable worker adapters behind the Go-centered orchestration/audit/semantic-repair boundary.
4. Apply bounded structural or semantic repairs only where the source supports a defensible decision. Route ambiguous semantics to human review rather than guessing.
5. Write output using a verified open-source path, compare preservation assertions, and retain audit evidence.
6. Run independent machine validation, human semantic review, and assistive-technology checks before any release claim.

## Test-first slices

Begin with synthetic/deidentified characterization fixtures and failing preservation assertions. Add unit, adapter contract, integration, and end-to-end tests before implementation; use red/green/refactor for each vertical slice. Inject worker failures and hostile inputs. Keep scanned/OCR cases selective and subject to separate quality review.

## Dependency boundary

The [component inventory](../../research/remediation-component-inventory.md) records current release/license boundaries for OpenDataLoader OSS auto-tagging, PDFBox, veraPDF, qpdf, and selective OCRmyPDF/Tesseract. The full-write pdfcpu characterization is also recorded separately. Enterprise-only PDF/UA export is excluded; no runtime candidate is an accepted dependency yet.
