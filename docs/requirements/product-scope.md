# Product scope

**Status: Proposed scope for planning.** This document does not approve implementation, compliance targets, or schedules.

## Product lanes

### Existing-PDF accessibility remediation

The primary current design direction is open-source remediation of existing PDFs. A candidate Go control plane would orchestrate jobs, audit evidence, and a semantic repair model through replaceable worker adapters. Start implementation planning only after data/profile decisions and an existing-PDF writer/fidelity feasibility spike.

Remediation dependencies must be open-source-only. Enterprise-only tagging or PDF/UA export dependencies are excluded. Candidate tools still require current license, capability, version, and integration verification before acceptance.

### Accessible document generation

The broader product plan also covers generation from data and templates, DOCX and Markdown sources, and PDF output. The PDF generation engine has not been selected. Generation is a distinct lane from existing-PDF remediation and must pass its own engine and semantic-preservation gates.

### Separate AFP feasibility track

IBM AFP print migration belongs in a distinct feasibility track. Its input/output goals, preservation requirements, and go/no-go criteria remain open. It is not presumed to be part of either PDF lane's implementation.

## Document families

Both generation and remediation cover statements, invoices, correspondence, and reports. Each family needs synthetic or deidentified fixtures and explicit semantic, data-integrity, and visual-preservation acceptance criteria.

## Validation boundaries

Automated checks can identify machine-detectable structure and syntax problems; they cannot establish semantic correctness on their own. Tagged output alone is not proof of PDF/UA or WCAG conformance. Release decisions require independent machine validation, human semantic review, and assistive-technology checks, with applicable profiles/standards still to be decided.

## Out of scope for this planning setup

- Product implementation or production integrations.
- Commercial-only remediation dependencies.
- Prior private reports or sample-document migration.
- Approved performance targets, budgets, timelines, or compliance claims.
