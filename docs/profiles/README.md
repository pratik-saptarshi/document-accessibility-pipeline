# Proposed data and review profiles

Profiles here are contract examples, not approved policy. All choices remain **Open** until an owner accepts an evidence-backed decision. The examples use synthetic names and values only.

The shared Go contract in `internal/profile` requires an explicit family, lane, source fields, language, confidentiality, retention, accessibility expectations, human-review expectations, and PDF/UA target. The accepted PDF/UA-2 target is represented independently; it does not approve the still-open semantic policy or human-review expectations. Other profile choices are labeled `open` or `proposed`. See the lane schemas and eight examples:

Source-field names must be non-empty, have no leading or trailing whitespace, and be unique within a profile. Both lane schemas reject padded names. The Go validator additionally checks uniqueness by field name; JSON Schema's `uniqueItems` only rejects identical field objects, so consumers must apply the Go contract or equivalent validation.

This specification supports [generation requirements REQ-GEN-001 and REQ-GEN-002](../requirements/catalog.md) and [remediation characterization requirements REQ-REM-001 through REQ-REM-004](../requirements/catalog.md). These are profile-contract tests, not the linked product acceptance cases.

- [Generation schema](../workflows/generation/profiles/profile.schema.json) and [examples](examples/generation/)
- [Existing-PDF remediation schema](../workflows/remediation/profiles/profile.schema.json) and [examples](examples/remediation/)

Each lane has examples for statements, invoices, correspondence, and reports. Generation profiles describe proposed source data fields. Remediation profiles describe observed or unknown fields in an existing PDF. This shared contract does not merge the workflow lanes.

## Open decisions

- Which source fields and allowed values apply to each family and source system?
- Which languages, confidentiality classes, retention periods, semantic interpretation criteria, and human review authorities apply?
- What threat-data handling, access, deletion, and audit rules govern real documents?

The PDF/UA-2 output target is accepted in [Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md). Source fields, language, retention, production confidentiality labels, semantic policy, and reviewer authority remain open. Machine validation and tags do not replace semantic review or assistive-technology review.
