# Proposed function contracts

These are behavior contracts for planning only. They do not select a language library, engine, service boundary, or production design.

## FN-REM-INTAKE (US-REM-001)

- **Inputs:** PDF bytes or a privacy-safe fixture plus a proposed profile
- **Outputs:** input classification, artifact identity, supported/unsupported result
- **Rules:** No source mutation; supported input profile is a decision gate
- **Error states:** malformed, encrypted, unsupported, or oversized input is explicit

## FN-REM-CHARACTERIZE (US-REM-001)

- **Inputs:** accepted input and characterization profile
- **Outputs:** baseline text, page, geometry, and feature observations
- **Rules:** Capture only applicable, declared preservation fields
- **Error states:** unavailable feature checks are marked unknown, not pass

## FN-REM-REPAIR-BASELINE (US-REM-001)

- **Inputs:** accepted artifact and approved bounded repair rules
- **Outputs:** candidate repaired PDF and change manifest
- **Rules:** Proposed contract; never overwrite original; unsupported feature stops or routes to review
- **Error states:** repair failure retains original and emits classified error

## FN-REM-COMPARE (US-REM-001)

- **Inputs:** original, candidate, characterization observations
- **Outputs:** preservation comparison evidence
- **Rules:** Thresholds and acceptable changes remain open
- **Error states:** missing evidence cannot be represented as pass

## FN-REM-SEMANTIC-REVIEW (US-REM-002)

- **Inputs:** candidate structure, uncertainty markers, evidence
- **Outputs:** review-needed or reviewer decision record
- **Rules:** Ambiguity is not auto-resolved; human authority remains open
- **Error states:** missing reviewer decision remains blocked

## FN-GEN-VALIDATE-DATA (US-GEN-001)

- **Inputs:** family data, template binding, proposed profile
- **Outputs:** validated model or field-level findings
- **Rules:** Required fields, types, and arithmetic policies are profile-defined
- **Error states:** invalid/ambiguous values stop render

## FN-GEN-STATEMENT (US-GEN-001)

- **Inputs:** validated statement data and template
- **Outputs:** candidate accessible PDF and render manifest
- **Rules:** Semantic structure and arithmetic acceptance remain profile-gated
- **Error states:** missing binding or failed rendering is explicit

## FN-GEN-INVOICE (US-GEN-001)

- **Inputs:** validated invoice data and template
- **Outputs:** candidate accessible PDF and render manifest
- **Rules:** Currency, tax, rounding, and totals policy remain open
- **Error states:** invalid totals or missing binding stop render

## FN-GEN-CORRESPONDENCE (US-GEN-002)

- **Inputs:** validated correspondence data and template
- **Outputs:** candidate document with declared heading and reading order
- **Rules:** Language, heading, and content rules are profile-defined
- **Error states:** missing or ambiguous content is explicit

## FN-GEN-REPORT (US-GEN-002)

- **Inputs:** validated report data and template
- **Outputs:** candidate report with declared sections and tables
- **Rules:** Table semantics and data provenance are profile-defined
- **Error states:** unmapped values or unsupported layout stops render

## FN-GEN-VALIDATE-OUTPUT (US-GEN-002)

- **Inputs:** generated candidate and validation profile
- **Outputs:** machine findings and separate human-review status
- **Rules:** No compliance conclusion from tags or one validator
- **Error states:** unsupported checks remain unknown

## FN-X-JOB-AUDIT (US-X-001)

- **Inputs:** job events, artifact identities, decision events
- **Outputs:** bounded status and audit record
- **Rules:** Content logging, access, retention, and integrity mechanisms are open
- **Error states:** audit failure cannot be silently ignored

## FN-X-SAFE-FAILURE (US-X-001)

- **Inputs:** hostile input, timeout, worker/resource failure
- **Outputs:** classified terminal/retry/review state
- **Rules:** Resource limits and retry policy are open
- **Error states:** no success state without complete required evidence

## FN-AFP-ASSESS (US-AFP-001)

- **Inputs:** agreed AFP target, safe fixtures, fidelity criteria
- **Outputs:** evidence-backed go/no-go recommendation
- **Rules:** No conversion implementation; target agreement is prerequisite
- **Error states:** missing target/fixture/criteria returns decision-blocked

## FN-REM-MACHINE-VALIDATE (US-REM-002)

- **Inputs:** candidate output and proposed validator profile
- **Outputs:** machine pass, finding, unsupported, or error with provenance
- **Rules:** Machine result is distinct from human semantic and assistive-technology review
- **Error states:** missing profile or unavailable validator is unknown, not pass

## FN-REM-ASSISTIVE-TECH-REVIEW (US-REM-002)

- **Inputs:** candidate and proposed assistive-technology review protocol
- **Outputs:** separate observation record and review status
- **Rules:** Tool, version, steps, and reviewer evidence are captured after protocol approval
- **Error states:** unavailable test environment is blocked/unknown
