# Proposed data and review profiles

Profiles here are contract examples, not approved policy. All choices remain **Open** until an owner accepts an evidence-backed decision. The examples use synthetic names and values only.

The shared Go contract in `internal/profile` requires an explicit family, lane, source fields, language, confidentiality, retention, accessibility expectations, and human-review expectations. Each policy is labeled `open` or `proposed`; the validator does not decide the policy. See the lane schemas and eight examples:

This specification supports [generation requirements REQ-GEN-001 and REQ-GEN-002](../requirements/catalog.md) and [remediation characterization requirements REQ-REM-001 through REQ-REM-004](../requirements/catalog.md). These are profile-contract tests, not the linked product acceptance cases.

- [Generation schema](../workflows/generation/profiles/profile.schema.json) and [examples](examples/generation/)
- [Existing-PDF remediation schema](../workflows/remediation/profiles/profile.schema.json) and [examples](examples/remediation/)

Each lane has examples for statements, invoices, correspondence, and reports. Generation profiles describe proposed source data fields. Remediation profiles describe observed or unknown fields in an existing PDF. This shared contract does not merge the workflow lanes.

## Open decisions

- Which source fields and allowed values apply to each family and source system?
- Which languages, confidentiality classes, retention periods, accessibility targets, and human review authorities apply?
- What threat-data handling, access, deletion, and audit rules govern real documents?

No standards target, retention period, or production confidentiality label is selected. Machine validation and tags do not replace semantic review or assistive-technology review.
