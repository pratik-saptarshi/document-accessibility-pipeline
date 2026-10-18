# Reviewer authority and machine-evidence verification follow-up

**Status: Proposed; trusted-verifier and release-evidence implementation remain open.** This follow-up records implementation acceptance requirements raised by [PR #71 review findings](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170) and [machine-evidence binding finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123180). [Decision 0015](../decisions/0015-approver-role-evidence-source.md) selects the owner-approved role-grant register and its temporal policy. This document does not name an approver, implement a trusted verifier, or claim that a reviewer/release gate is implemented.

## P1 — Verify authority at the review time

The approver persona is the authorized role; an identity string or self-asserted role is not proof of membership. Before release, a trusted verifier must check an immutable register entry for the actual acting identity and establish that the grant covered the UTC `reviewed_at` timestamp. The review evidence must retain a source reference and digest sufficient to re-verify the same evidence at release time. Naming and verifying the actual person is release-stage work; it does not block implementation or functional testing of the approved profile, family-recognition, and contrast features.

Decision 0015 selects exact identity matching, closed validity intervals, prospective revocation, and its stated retroactive-grant rules. Until the register and verifier exist and source evidence is verifiable, release remains blocked/review-needed. No authority may be inferred from a supplied identity, role name, unverified reference, machine result, or simulation. Synthetic identities can support functional tests but never count as real release approval.

Once the source is selected, the executable contract must reject:

- operator-only, unknown, or otherwise unauthorized identities, including submitters who self-assert the approver role;
- missing, unverifiable, replaced, forged, wrong-source, identity-mismatched, role-mismatched, or digest-mismatched authorization evidence;
- grants expired or not yet valid at `reviewed_at`, malformed validity windows, and review timestamps outside the authoritative window;
- evidence that was valid at initial review but no longer satisfies the owner-decided revocation policy at release.

It must accept authority only when the trusted evidence binds the exact identity to the approver role and its validity covers `reviewed_at`. These are future verifier requirements; their existence does not claim that role-source enforcement is implemented.

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

Decision 0015 records the selected source and temporal/revocation semantics. The canonical Beads status for `dap-jo9.1.1.6` and implementation tasks is not changed by this documentation. Authority-independent simulated contract cases are isolated in `dap-jo9.2.5.9.1` and retain their existing prerequisites. Trusted-role verification remains in `dap-jo9.2.5.9.2` as release-stage work. The parent review/release gate still requires an actual eligible candidate and genuine authorized semantic/AT evidence; synthetic records do not satisfy it.
