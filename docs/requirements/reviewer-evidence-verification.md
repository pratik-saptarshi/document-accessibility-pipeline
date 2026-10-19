# Reviewer authority and machine-evidence verification follow-up

**Status: Proposed; source category and temporal rules accepted, authenticated-source implementation open.** This follow-up records implementation acceptance requirements raised by [PR #71 review findings](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) and [machine-evidence binding finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180). [Decision 0015](../decisions/0015-approver-role-evidence-source.md) selects an owner-approved append-only register and defines entry, digest, append-time, validity, and revocation rules. This document does not select a trust anchor, register provider, or writer-control model, and it does not claim that a reviewer gate is implemented.

## P1 — Verify authority at the review time

The approver persona is the authorized role; an identity string or self-asserted role is not proof of membership. The selected source category is an owner-approved append-only role-grant register. Before accepting human/AT review evidence, a trusted source adapter must authenticate that register and return a complete, ordered snapshot with source-authenticated append times; the verifier applies Decision 0015's UTC validity and prospective-revocation rules. The review record retains an authenticated source reference, snapshot high-water mark, and constructible entry digest for release-time re-verification.

The trust anchor/authentication mechanism and grant/revocation writer controls remain open product/security prerequisites. The register and trusted source adapter do not yet exist; until they are implemented and verifiable, the gate returns blocked/review-needed. It must not infer authority from a supplied identity, role name, unverified reference, machine result, or simulation.

The executable contract must reject:

- operator-only, unknown, or otherwise unauthorized identities, including submitters who self-assert the approver role;
- missing, unverifiable, replaced, forged, wrong-source, identity-mismatched, role-mismatched, or digest-mismatched authorization evidence;
- grants expired or not yet valid at `reviewed_at`, malformed validity windows, and review timestamps outside the authoritative window;
- evidence that was valid at initial review but no longer satisfies the owner-decided revocation policy at release.

It may accept authority only when the authenticated register evidence binds the exact identity to the approver role, source completeness/authenticity is proven, and the grant covers `reviewed_at` under Decision 0015. These are future implementation acceptance criteria; no provider integration or real approver identity is selected here.

## P2 — Bind reviews to exact machine evidence

The review evidence must identify the exact machine evidence bundle shown to the approver. Retain the report digest, verified bundle-metadata digest, derived machine outcome, validator/profile/tool versions, and the candidate digest recorded inside the machine evidence. The machine candidate digest must equal the human review's `candidate_sha256`. Outcome is derived from the verified report and metadata, never asserted by the reviewer. A new evidence record is required if any bound machine evidence changes, even when tool versions and candidate bytes stay the same.

Executable contract cases must show that:

- a report, bundle, metadata, derived outcome, profile, validator identity, or candidate mismatch blocks the review record;
- replacing a report or bundle with materially different bytes at the same tool version changes its digest and cannot reuse the prior review;
- evidence tied to another candidate is rejected even if the report and tool version are otherwise valid;
- pass, finding, processing error, and missing evidence remain distinguishable; a machine result never satisfies semantic or AT review;
- simulations and dry runs remain labeled and never count as genuine human or AT approval.

The current Go verifier already derives candidate digest, report digest, profile, validator version, and outcome from the validated evidence. The narrow TDD increment adds the exact bundle-metadata digest to that result. It does not implement a human review record or release gate.

## Beads and implementation gate

`dap-jo9.1.1.6` is open after the prior closure was corrected: Decision 0015 accepts the register source category and temporal rules, while the original issue acceptance criteria remain unchanged. The issue remains incomplete until the concrete register identity and trust anchor, authenticated source interface (including snapshot completeness/high-watermark continuity and immutable append-time evidence), and grant/revocation writer controls are decided. These remain prerequisites to `dap-jo9.2.5.9.2`; no provider deployment or named release approver is selected. Authority-independent, explicitly simulated contract cases are isolated in `dap-jo9.2.5.9.1` and retain the existing authority/protocol, candidate, feasibility, and canonical-case prerequisites. The parent review gate requires both work slices and still cannot close without an actual eligible candidate and genuine authorized semantic/AT evidence. No provider or trust mechanism beyond the register category is selected here.
