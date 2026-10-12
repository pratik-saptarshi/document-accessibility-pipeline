# Workflows

Two product workflows are planned separately:

- [Accessible document generation](generation/README.md): validated data and templates or DOCX/Markdown sources flow through a PDF engine that remains to be selected.
- [Existing-PDF remediation](remediation/README.md): inspect and characterize an existing PDF, make bounded semantic repairs through open-source adapters, preserve content, then validate and review.

Both lanes cover statements, invoices, correspondence, and reports. Each implementation slice follows red/green/refactor and uses synthetic/deidentified fixtures. See the [validation strategy](../validation/README.md).
