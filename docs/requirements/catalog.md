# Requirements catalog

**Status: Proposed.** Requirement statements are planning contracts. They do not establish compliance, workload, or approved implementation.

## REQ-REM-001 — Existing PDF intake and characterization

Characterize input PDFs without silently changing source content; supported inputs and profiles remain open.

## REQ-REM-002 — Existing PDF intake and characterization

Cover statements, invoices, correspondence, and reports with synthetic or deidentified fixtures.

## REQ-REM-003 — Bounded structure repair with fidelity evidence

Record preservation evidence for visible text, pagination, geometry, links, forms, annotations, and metadata where applicable.

## REQ-REM-004 — Bounded structure repair with fidelity evidence

Limit automated repairs to explicitly supported structural cases; preserve an original and report unsupported inputs.

## REQ-REM-005 — Semantic repair and uncertainty routing

Ambiguous reading order, tables, figures, or meaning must produce a review-needed outcome rather than a guessed semantic repair.

## REQ-REM-006 — Independent remediation validation and review

Machine validation, human semantic review, and assistive-technology review have separately recorded outcomes.

## REQ-REM-007 — Independent remediation validation and review

Remediation candidate output is machine-checked against the explicit PDF/UA-2 profile without fallback; the result is separate from semantic and assistive-technology review.

## REQ-GEN-001 — Validated family data and template inputs

Validate required fields, types, arithmetic rules, and template bindings before rendering; family schemas and business rules remain open.

## REQ-GEN-002 — Validated family data and template inputs

Generation input examples cover statements, invoices, correspondence, and reports using synthetic or deidentified data.

## REQ-GEN-003 — Accessible document generation by family

Render semantic structure and family content from validated inputs while preserving declared source meaning.

## REQ-GEN-004 — Accessible document generation by family

Keep the PDF engine and DOCX/Markdown conversion support behind evidence-backed feasibility decisions.

## REQ-GEN-005 — Accessible document generation by family

Generated output is independently validated and not described as compliant from tags or a machine pass alone.

## REQ-GEN-006 — Accessible document generation by family

Generated PDF output is machine-checked with explicit PDF/UA-2 profile selection and no fallback; semantic and assistive-technology outcomes remain separate.

## REQ-X-001 — Cross-lane job outcomes and audit evidence

Each job has a traceable status, failure class, artifact identity, and review state without exposing document content in ordinary diagnostics.

## REQ-X-002 — Cross-lane safe failure and privacy controls

Malformed or hostile inputs, timeouts, resource exhaustion, and worker failures fail safely with bounded diagnostics and no silent success.

## REQ-X-003 — Cross-lane safe failure and privacy controls

Audit evidence records decisions and validation provenance while privacy, retention, and access policy remain open.

## REQ-AFP-001 — Separate AFP feasibility decision

An AFP feasibility recommendation requires an agreed input/output target, representative safe fixtures, fidelity criteria, and evidence; conversion is not authorized.

## Open decisions

- Actual persona titles, authority, and approval responsibilities.
- Family profile fields, source systems, workloads, volumes, and acceptance examples.
- Semantic interpretation criteria, human reviewer authority, assistive-technology protocol, and sign-off.
- Authenticated veraPDF artifact route, actual Maven signer identity/signature, immutable artifact digest, and reproducible acquisition.
- Generation engine, template model, DOCX/Markdown conversion support, and licensing evidence.
- Existing-PDF writer feasibility, supported structures, preservation thresholds, OCR policy, and open-source license verification.
- Business arithmetic, currency, tax, date, rounding, and exception rules.
- Audit integrity, data access, retention, diagnostic redaction, and deletion controls.
- Performance, availability, retry, resource-limit, and service-level goals.
- AFP input/output target, workload, fonts/barcodes, fidelity requirements, and go/no-go authority.
