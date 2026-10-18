# Reviewer authority and machine-evidence verification follow-up

**Status: Proposed; authority-source decision Open.** This follow-up records implementation acceptance requirements raised by [PR #71 review findings](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) and [machine-evidence binding finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180). It does not amend or replace accepted [Decision 0004](../decisions/0004-approver-persona-review-protocol.md), select an authority provider, or claim that a reviewer gate is implemented.

## P1 — Verify authority at the review time

The approver persona is the authorized role; an identity string or self-asserted role is not proof of membership. Before accepting human/AT review evidence, the product owner must select a trusted source and interface that can verify an immutable authorization record for the acting identity and establish that the grant covered the UTC `reviewed_at` timestamp. The record must retain a source reference and digest sufficient to re-verify the same evidence at release time.

The choice of source, identity mapping, validity interval rules, and how retroactive grants or later revocation affect past reviews are open product/security decisions. Until those choices are recorded and the source is verifiable, the gate returns blocked/review-needed. It must not infer authority from a supplied identity, role name, unverified reference, machine result, or simulation.

Once the source is selected, the executable contract must reject:

- operator-only, unknown, or otherwise unauthorized identities, including submitters who self-assert the approver role;
- missing, unverifiable, replaced, forged, wrong-source, identity-mismatched, role-mismatched, or digest-mismatched authorization evidence;
- grants expired or not yet valid at `reviewed_at`, malformed validity windows, and review timestamps outside the authoritative window;
- evidence that was valid at initial review but no longer satisfies the owner-decided revocation policy at release.

It must accept authority only when the trusted evidence binds the exact identity to the approver role and its validity covers `reviewed_at`. These are negative-test specifications; no role-source Go enforcement is authorized until the source and temporal/revocation policy are decided and recorded in Beads.

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

`dap-jo9.1.1.6` is the satisfied-and-closed owner-decision issue for the trusted source and temporal/revocation semantics, decided in [Decision 0015](../decisions/0015-approver-role-evidence-source.md). Authority-independent, explicitly simulated contract cases are isolated in `dap-jo9.2.5.9.1` and retain the existing authority/protocol, candidate, feasibility, and canonical-case prerequisites. Trusted role verification is isolated in `dap-jo9.2.5.9.2`, which depended on that owner decision; the parent review gate requires both work slices and still cannot close without an actual eligible candidate and genuine authorized semantic/AT evidence. No provider or policy beyond the Decision 0015 register is selected or accepted here.
