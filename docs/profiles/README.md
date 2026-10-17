# Proposed data and review profiles

Profiles here are contract examples, not approved policy. All choices remain **Open** until an owner accepts an evidence-backed decision. The examples use synthetic names and values only.

The shared Go contract in `internal/profile` requires an explicit family, lane, source fields, language, confidentiality, retention, accessibility expectations, human-review expectations, and PDF/UA target. The accepted PDF/UA-2 target is represented independently; it does not approve the still-open semantic policy or human-review expectations. Other profile choices are labeled `open` or `proposed`. See the lane schemas and eight examples:

Source-field names must be non-empty, have no leading or trailing whitespace, and be unique within a profile. Boundary whitespace follows Go's `strings.TrimSpace`/Unicode `IsSpace` set: U+0009–U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000–U+200A, U+2028–U+2029, U+202F, U+205F, and U+3000. U+FEFF is not whitespace under this contract. Both lane schemas use an explicit boundary class and reject such names. The Go validator additionally checks uniqueness by field name; JSON Schema's `uniqueItems` only rejects identical field objects, so consumers must apply the Go contract or equivalent validation.

This specification supports [generation requirements REQ-GEN-001 and REQ-GEN-002](../requirements/catalog.md) and [remediation characterization requirements REQ-REM-001 through REQ-REM-004](../requirements/catalog.md). These are profile-contract tests, not the linked product acceptance cases.

- [Generation schema](../workflows/generation/profiles/profile.schema.json) and [examples](examples/generation/)
- [Proposed bounded ariapdf financial evaluation profile](proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md) is a synthetic test proposal only; it is not an approved profile or product default.
- [Existing-PDF remediation schema](../workflows/remediation/profiles/profile.schema.json) and [examples](examples/remediation/)

Each lane has examples for statements, invoices, correspondence, and reports. Generation profiles describe proposed source data fields. Remediation profiles describe observed or unknown fields in an existing PDF. This shared contract does not merge the workflow lanes.

The synthetic remediation statement example marks `transactions` and the invoice example marks `line_items` as `unknown` and `open`, matching the bounded statement/invoice table examples. These entries record only that table rows are present; they do not define required production fields, currency or arithmetic rules, or a supported-input profile.

## Fixture and data-handling rules (Proposed)

The following rules are proposed, not approved policy. They govern all profile examples, fixtures, and test inputs in this repository.

- **Synthetic or deidentified only.** No real customer documents, private reports, or production source data in the repository, issues, or fixtures. All eight examples use invented names, values, and identifiers; a scan of the example corpus confirms no email addresses or card-like number sequences.
- **Deidentification bar.** A deidentified fixture must not contain real account numbers, tax IDs, personal names, addresses, or any identifier that maps to a real person or account. When in doubt, synthesize rather than scrub.
- **Threat-data handling.** If real or hostile data is encountered — for example, a production document submitted as a test case — do not ingest it into fixtures: quarantine the copy, delete it, and record the incident as an owner question. Hostile-input testing uses purpose-built synthetic fixtures only, per the failure-injection guidance in `AGENTS.md`.
- **Fixture retention.** Test fixtures are versioned with the repository. No separate fixture retention schedule is defined yet; see the owner questions below.

## Open decisions and owner questions

Each item needs an explicit owner decision before it can become approved policy. Until then, related profile choices stay labeled `open` or `proposed`.

- **Source fields and allowed values** — which fields and value sets apply per family and source system? (Gates `dap-jo9.2.1`, `dap-jo9.1.5`, `dap-jo9.3.1`.)
- **Languages** — which languages must generated and remediated documents support?
- **Confidentiality classes** — which production confidentiality labels apply, and what handling does each require?
- **Retention periods** — how long are source documents, fixtures, and audit evidence retained?
- **Semantic interpretation criteria** — what counts as correct meaning and reading order for each family?
- **Human review authority** — who reviews, with what authority, and what evidence do they produce?
- **Threat-data governance** — what access, deletion, and audit rules govern real documents if they ever enter scope?

The PDF/UA-2 output target is accepted in [Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md). Source fields, language, retention, production confidentiality labels, semantic policy, and reviewer authority remain open. Machine validation and tags do not replace semantic review or assistive-technology review.
