# Decision 0017: Approved shared profile-family set

**Status: Accepted.** The product owner approved the eight-family set on
2026-10-18. The family approval is recorded in [Decision 0014 at PR #151's
reviewed head](https://github.com/pratik-saptarshi/document-accessibility-pipeline/blob/5ba4ace4b133a34c35f0242fa021d4f14ff514e8/docs/decisions/0014-profile-policy-decisions.md).

## Decision

The shared profile-family identifiers for planning contracts in both the
generation and remediation lanes are:

- `statements`
- `invoices`
- `correspondence`
- `reports`
- `presentations`
- `forms`
- `marketing`
- `hr_documents`

Both lane schemas and the Go profile validator use this same set.

## Limits

This decision accepts the family identifiers for the profile contracts. It
does not establish that every family has complete fixtures, workflow behavior,
or production support. It does not approve production defaults, dependencies,
privacy or retention policy, semantic rules, language support, validation
claims, or release gates. Proposed and open profile values retain their
existing status. The initial remediation priority in Decision 0002 is
unchanged.

## Follow-up

Keep the family enum consistent in both lane schemas and the Go validator.
Add per-family fixtures and acceptance criteria before claiming that a family
has been implemented or validated end to end.
