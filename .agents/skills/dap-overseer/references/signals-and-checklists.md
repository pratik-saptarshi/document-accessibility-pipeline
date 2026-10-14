# Review signals and checklist

## Signals

- **Security:** input boundaries, authorization, secrets, injection, unsafe
  parsing, information exposure.
- **Reliability:** cancellation, deadlines, retries, idempotency, races,
  bounded resources, rollback.
- **Data:** schema and transformation invariants, malformed and missing values,
  preservation, provenance.
- **Build/operations:** CI boundaries, dependency provenance, deployment,
  credentials, external tools.
- **Accessibility/document behavior:** semantics, reading order, validation
  limits, human review evidence.
- **Plans/docs:** gaps, conflicting requirements, sequencing, ownership, and
  unverified assumptions.

## Checklist

- Trace critical inputs and outputs across component boundaries.
- Verify invariants, error paths, cancellation, and existing safeguards.
- Check current source and the complete task diff before calling behavior absent.
- Ask whether a realistic regression is covered; avoid source-text-only claims.
- Cite exact evidence and distinguish facts, inferences, and unknowns.
- Do not inflate severity for pre-existing debt unless the change worsens it.
- State what was not reviewed or could not be verified.
