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

### Accepted initial remediation priority

For the first remediation increment, **statements and invoices with tables matter first and most**, as accepted for prioritization only in [Decision 0002](../decisions/0002-initial-remediation-family-priority.md). This does not select an engine, approve production processing, or resolve language, topology, arithmetic, privacy/retention, or reviewer-authority policy. The bounded synthetic experiment proposed for the next batch is recorded in the [next-step decision register](../planning/next-step-decision-register-2026-10-13.md). Its fixture parameters are experimental and are not production defaults.

## Validation boundaries

PDF/UA-2 is the accepted target for generated and remediated PDF output, with explicit `ua2` validation and no profile fallback ([Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md)). This target decision does not approve semantic policies, human-review authority, assistive-technology protocol, a writer, or release eligibility. Automated checks cannot establish semantic correctness on their own. Tagged output or a machine pass alone is not proof of PDF/UA-2 or WCAG conformance. Human semantic and assistive-technology checks remain separate release gates.

## Out of scope for this planning setup

- Product implementation or production integrations.
- Commercial-only remediation dependencies.
- Prior private reports or sample-document migration.
- Approved performance targets, budgets, timelines, or compliance claims.
