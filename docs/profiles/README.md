# Proposed data and review profiles

Profiles here are contract examples, not approved policy defaults. The shared family identifiers are accepted in [Decision 0017](../decisions/0017-approved-profile-family-set.md); the profile-policy questions are answered in [Decision 0014](../decisions/0014-profile-policy-decisions.md) (per-family required source fields, PDF/UA-2 target, machine/human semantic boundary) with the launch-scope deferrals recorded in [Decision 0016](../decisions/0016-deferred-scope-restricted-spanish.md) (English-only language allowlist; `restricted` confidentiality refused). The examples use synthetic names and values only.

The shared Go contract in `internal/profile` requires an explicit family, lane, source fields, language, confidentiality, retention, accessibility expectations, human-review expectations, and PDF/UA target. Its eight accepted family identifiers are recorded in [Decision 0017](../decisions/0017-approved-profile-family-set.md). Language is constrained to the English allowlist (`en`, `en-US`, `en-GB`) and `restricted` confidentiality is refused per Decision 0016; per-family required top-level source fields are enforced per Decision 0014 section 1. This does not validate nested form labels, alt-text presence, or link-text structure; those remain downstream machine checks. The accepted PDF/UA-2 target is represented independently; it does not approve the still-open semantic policy or human-review expectations. Remaining profile choices are labeled `open` or `proposed`. See the lane schemas and current examples:

Source-field names must be non-empty, have no leading or trailing whitespace, and be unique within a profile. Boundary whitespace follows Go's `strings.TrimSpace`/Unicode `IsSpace` set: U+0009–U+000D, U+0020, U+0085, U+00A0, U+1680, U+2000–U+200A, U+2028–U+2029, U+202F, U+205F, and U+3000. U+FEFF is not whitespace under this contract. Both lane schemas use an explicit boundary class and reject such names. The Go validator additionally checks uniqueness by field name; JSON Schema's `uniqueItems` only rejects identical field objects, so consumers must apply the Go contract or equivalent validation.

This specification supports [generation requirements REQ-GEN-001 and REQ-GEN-002](../requirements/catalog.md) and [remediation characterization requirements REQ-REM-001 through REQ-REM-004](../requirements/catalog.md). These are profile-contract tests, not the linked product acceptance cases.

- [Generation schema](../workflows/generation/profiles/profile.schema.json) and [examples](examples/generation/)
- [Proposed bounded ariapdf financial evaluation profile](proposals/ariapdf-financial-evaluation-proposal-2026-10-15.md) is a synthetic test proposal only; it is not an approved profile or product default.
- [Existing-PDF remediation schema](../workflows/remediation/profiles/profile.schema.json) and [examples](examples/remediation/)

Each lane has proposed examples for all eight families (statements, invoices, correspondence, reports, presentations, forms, marketing, hr_documents); the four families approved in Decision 0014 gained examples in this change. The examples are synthetic contract illustrations, not production configs. Per [Decision 0017](../decisions/0017-approved-profile-family-set.md), add per-family fixtures and acceptance criteria before claiming a family has been implemented or validated end to end. Generation profiles describe proposed source data fields. Remediation profiles describe observed or unknown fields in an existing PDF. This shared contract does not merge the workflow lanes.

The synthetic remediation statement example marks `transactions` and the invoice example marks `line_items` as `unknown` and `open`, matching the bounded statement/invoice table examples. These entries record only that table rows are present; they do not define required production fields, currency or arithmetic rules, or a supported-input profile.

## Decided

Recorded in Decisions 0014–0017; example values carry status `proposed`/`open` per the contract:

- Language: English-only launch allowlist (`en`, `en-US`, `en-GB`) — [Decision 0016](../decisions/0016-deferred-scope-restricted-spanish.md). Spanish deferred; no Spanish fixtures or `/Lang: es` claims while it stands.
- Confidentiality classes: `public`/`internal` stand; `restricted` refused at intake until the Q5 controls are implemented and independently reviewed — [Decision 0014](../decisions/0014-profile-policy-decisions.md), [Decision 0016](../decisions/0016-deferred-scope-restricted-spanish.md).
- Retention tiers: `transient` / `mandate-horizon` / `regulatory` — [Decision 0014](../decisions/0014-profile-policy-decisions.md).
- Review authority: three-role model (document owner / accessibility reviewer / approver), approver identity from the owner-approved evidence-source category — [Decision 0014](../decisions/0014-profile-policy-decisions.md), [Decision 0015](../decisions/0015-approver-role-evidence-source.md). SMB dual-hat reviews must be recorded as dual-hat, never hidden.
- APCA contrast checking: removed per owner direction; WCAG 2.2 contrast ratios are advisory measurements, not pass/fail gates — [Decision 0014](../decisions/0014-profile-policy-decisions.md).

## Still open

- Which non-financial family gets the first bounded fixture (the draft suggests `forms`)? — Decision 0014 follow-up; no bead yet.
- Named screen readers for the AT-review protocol (NVDA/JAWS/VoiceOver)? — Decision 0014 follow-up; no bead yet.
- The incident-notification window? — Decision 0014 follow-up; no bead yet.
- The `restricted`-document dual-hat question — open for the owner's judgment ([Decision 0015](../decisions/0015-approver-role-evidence-source.md)); `restricted` itself is refused at launch per Decision 0016.
- Retention-period assignment per case (the tier model is decided; the `mandate-horizon` horizon and per-case tier choice are open) — Decision 0014 follow-up; no bead yet.
- Semantic policy and conformance evidence — tracked in `dap-jo9.1.1`'s notes; semantic rules unapproved per [Decision 0017](../decisions/0017-approved-profile-family-set.md) Limits.
- Approver-evidence source category, temporal policy, and register form — accepted ([Decision 0015](../decisions/0015-approver-role-evidence-source.md), [Decision 0019](../decisions/0019-approver-register-form.md)). Still open: the adapter implementation (in progress), trust anchor and writer controls (out of scope per Decision 0019 — deployment concerns), and genuine release evidence (`dap-jo9.2.5.9.2`).
- Threat-data handling, access, deletion, and audit rules for real documents — proposed in Decision 0014 (Q5) but not implemented or independently reviewed; the `restricted` revisit waits on the independent security review ([Decision 0016](../decisions/0016-deferred-scope-restricted-spanish.md)).
- Production defaults, dependencies, privacy/retention policy detail, validation claims, release gates — no owner yet; unapproved per [Decision 0017](../decisions/0017-approved-profile-family-set.md) Limits.

The PDF/UA-2 output target is accepted in [Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md). Machine validation and tags do not replace semantic review or assistive-technology review.
