# DRAFT Decision 0010: Authoritative approver-role evidence source and temporal policy

**Status: DRAFT — NOT accepted. NO SOURCE IS SELECTED.**

**Date: 2026-10-17.**

> **Decision pending owner acceptance.** This record is a draft for product-owner
> review only. It selects nothing. Until the product owner explicitly accepts a
> source, the authoritative approver-role evidence source remains **unselected**,
> `dap-jo9.1.1.6` stays open, and the release gate returns **blocked/review-needed**
> for any review whose authority cannot be verified. No implementation of trusted
> role verification is authorized from this draft; `dap-jo9.2.5.9.2` remains
> blocked. Nothing here infers authority from simulations or machine results.

## Authority and relationship to prior records

[Decision 0004](0004-approver-persona-review-protocol.md) (accepted 2026-10-14)
designates the approver as a product **role** — the authority to review, approve,
or reject remediation candidates on semantic and assistive-technology grounds —
and defines the review-evidence record format (`approver_identity`,
`reviewed_at`, outcomes). It does **not** identify an authoritative, time-aware
source that proves the acting person held the approver role at `reviewed_at`.

The product-owner designation of the approver role itself was made in
`dap-jo9.1.1.4` (closed 2026-10-14) and is kept intact; this draft changes
neither. The owner-decision issue for the evidence source is `dap-jo9.1.1.6`
(P1, decision, open); the PR #71 P1 review finding that raised it is
<https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170>.
Implementation acceptance requirements are further recorded in
[docs/requirements/reviewer-evidence-verification.md](../requirements/reviewer-evidence-verification.md)
(status: Proposed; authority-source decision open).

## Problem

The release gate must distinguish the *accepted approver role* from the
*evidence proving* that a particular person held that role when a review
occurred. An identity string (`approver_identity`) or a self-asserted role
label is insufficient. A SHA-256 digest can identify evidence bytes but cannot
authenticate the issuer or establish role authority.

The product owner must therefore select, in one accepted decision:

1. the **trusted source and verification interface** for approver-role evidence;
2. the **identity mapping** binding a recorded `approver_identity` to the
   source's records;
3. the **validity policy** for grants at the UTC `reviewed_at` timestamp:
   validity-interval semantics, expiry, revocation, retroactive authorization,
   and the effects of an unavailable source or replaced evidence.

## Candidate trusted sources

The issue lists two considered alternatives; neither is selected. A third
candidate — preserving the current "no source selected" posture — is included
because the gate must fail closed until the owner decides.

### Option A — Owner-approved append-only role-grant register

- **Source.** A register (file, ledger, or database table) owned and explicitly
  approved by the product owner, recording each grant of the approver role to a
  named person with a validity interval.
- **Verification interface.** At review-verification time, the gate looks up the
  record matching the acting identity; at release time it re-verifies the same
  record by its immutable reference and recomputed digest. Append-only: grants
  are never edited in place; expiry and revocation are recorded as new entries.
- **Identity mapping.** The register's person identifier is the single canonical
  identity; the review record's `approver_identity` must match it exactly
  (mapping rules, e.g. case/email normalization, are owner-decided, not
  inferred).
- **Validity/revocation semantics.** Each grant carries `[grant_start, grant_end]`
  in UTC. A grant is usable for a review iff
  `grant_start <= reviewed_at <= grant_end`. Expiry is the interval end; a later
  revocation entry for the same grant marks it revoked. The accepted policy must
  state whether revocation retroactively invalidates reviews recorded inside
  the grant interval (owner decision; this draft does not decide it).
- **Tradeoffs.**
  - (+) Fully owner-controlled; no external provider, matching this repository's
    scope limit against provider integration and enterprise-only dependencies.
  - (+) Naturally produces immutable references and recomputable digests from
    ordinary append-only records.
  - (−) The owner (or their delegate) must maintain grant entries and revocation
    entries by hand; operational burden and discipline risk.
  - (−) No independent issuer to authenticate against — the register's authority
    is the owner's approval itself, so register integrity (who wrote what, when)
    must be protected by separate controls the owner also accepts.

### Option B — Named organizational identity and access-control system

- **Source.** An existing organizational IAM/IdP (the owner names it explicitly
  at acceptance; the name is blank in this draft) that is already the system of
  record for role grants.
- **Verification interface.** A defined read interface (API or exported signed
  record) against that system, usable at both review-verification time and
  release re-verification time, returning the grant record for the acting
  identity with its validity window and a source reference.
- **Identity mapping.** The organizational user identifier is canonical; the
  review record's `approver_identity` must map to it under mapping rules the
  owner accepts (no self-asserted labels ever pass).
- **Validity/revocation semantics.** As issued by the system: grant validity
  intervals, expiry, and revocation entries as recorded there. The owner must
  still decide retroactive-authorization policy, unavailable-source behavior
  (fail closed), and how a grant revoked between review and release affects the
  already-recorded review.
- **Tradeoffs.**
  - (+) Independent issuer and existing operational practice for grant/revoke.
  - (+) Time-aware queries ("was this person in the approver role at time T")
    are typically a first-class feature.
  - (−) Introduces a provider integration and an external runtime dependency at
    the release gate — out of this repository's current accepted scope and an
    enterprise-only boundary the product scope excludes.
  - (−) Unavailable-source behavior at release time becomes a hard dependency
    with real availability consequences; the fail-closed rule must still hold.
  - (−) The system's semantics (interval inclusivity, retroactive changes,
    audit retention) must be pinned in the accepted record, not assumed.

### Option C — No source selected; gate remains fail-closed (current posture)

- **Source.** None. The release gate returns blocked/review-needed for every
  review whose authority cannot be verified against an owner-accepted source.
- **Tradeoffs.**
  - (+) Fail-closed by construction; cannot grant release authority to an
    untrusted claim.
  - (−) No genuine human/AT approval can ever pass the release gate, so the
    bounded workflow cannot complete a release until the owner decides.

## Required semantics for the accepted record (source-independent)

Whatever source the owner selects, the accepted record must define, testably:

1. **Validity interval at UTC `reviewed_at`.** Each grant has a closed UTC
   validity interval `[not_before, not_after]`; a review's authority is
   established only if `not_before <= reviewed_at <= not_after` under the
   source's evidence.
2. **Expiry.** A grant whose `not_after` is before `reviewed_at` is
   **expired**: the review is rejected; expiry is automatic and irreversible
   except by a new grant entry.
3. **Revocation.** The record defines whether a revocation entry invalidates
   (a) only future reviews, or (b) also reviews already recorded inside the
   grant interval. Until the owner decides, verification implementations must
   assume the stricter reading (revoked grant ⇒ no release) and flag the
   ambiguity back to the owner.
4. **Retroactive grants.** The record defines whether a grant with
   `not_before` earlier than its own issuance time can validate a review
   recorded before the grant existed. The default is **not permitted** unless
   the accepted policy explicitly allows it with owner-level authorization
   evidence.
5. **Unavailable source.** If the source cannot be reached at verification or
   re-verification time, the gate returns blocked/review-needed. Unreachability
   is never treated as approval and never downgraded to "evidence absent."
6. **Replaced evidence.** Evidence replaced at the source (a newer record
   superseding the referenced one) invalidates the stored reference; the review
   is rejected unless the new record also binds the identity to the approver
   role and covers `reviewed_at`, in which case the review evidence record is
   re-pointed with an audit trail.

## Immutable evidence reference and digest

The review evidence record (Decision 0004's format, extended when the source is
accepted) must retain:

- an **immutable reference** to the exact grant record in the trusted source
  (source identifier, record id, and version/sequence, sufficient to retrieve
  the identical bytes again at release);
- a **SHA-256 digest** of those exact source bytes.

At release, the gate re-fetches the referenced bytes, recomputes the digest,
and compares; a mismatch (replaced, forged, or digest-mismatched evidence)
rejects the review. The digest identifies bytes; it does not authenticate the
issuer — issuer authenticity comes only from the owner-accepted source.

**Self-asserted identity or role labels are rejected.** A review record whose
authority rests on an identity string or role name not corroborated by the
trusted source at `reviewed_at` is treated as missing authorization evidence.

## Negative-test specifications (testable contract)

Carried verbatim from `dap-jo9.1.1.6` and
[reviewer-evidence-verification.md](../requirements/reviewer-evidence-verification.md);
the implementing contract (`dap-jo9.2.5.9.2`, authorized only after this
decision is accepted) must require rejection of:

- operator-only, unknown, or otherwise unauthorized identities, including
  submitters who self-assert the approver role;
- missing, unverifiable, replaced, forged, wrong-source, identity-mismatched,
  role-mismatched, or digest-mismatched authorization evidence;
- grants expired or not yet valid at `reviewed_at`, malformed validity windows,
  and review timestamps outside the authoritative window;
- evidence that was valid at initial review but no longer satisfies the
  owner-decided revocation policy at release;
- in the issue's terms: operator-only, unknown, unauthorized, expired,
  not-yet-valid, malformed/time-invalid, missing, unverifiable, wrong-source,
  identity/role-mismatched, replaced, forged, and digest-mismatched evidence.

The contract must **accept authority only** when the trusted evidence binds the
exact identity to the approver role and its validity covers `reviewed_at`.
Authority-independent synthetic evidence-contract cases are tracked separately
in `dap-jo9.2.5.9.1`; no source or authority is inferred from simulations or
machine results.

## Non-binding recommendation (pending owner acceptance)

**Draft recommendation: Option A** — an owner-approved append-only role-grant
register.

Rationale: it keeps the trusted source entirely under the product owner's
control, produces immutable references and recomputable digests from ordinary
append-only records, requires no provider integration (consistent with the
issue's scope limits), and matches the repository's existing pattern of
owner-accepted decision records as the authority of last resort. Option B is a
reasonable choice for a larger organization with an existing system of record,
at the cost of an external dependency at the release gate and an availability
contract the owner must also accept. Option C preserves the current fail-closed
posture but blocks all releases.

This recommendation is **non-binding**: it proposes, it does not select. The
product owner may accept it, accept another option, or define a new one.

## Owner acceptance checklist

Acceptance is a product-owner action on `dap-jo9.1.1.6`. The accepted record
must name:

- [ ] the trusted source and its verification interface (Option A, B, or a
      defined alternative — no blanks);
- [ ] the identity mapping from `approver_identity` to the source's records;
- [ ] the validity-interval, expiry, revocation, and retroactive-grant policy;
- [ ] unavailable-source behavior (fail closed) and replaced-evidence effects;
- [ ] the immutable reference format and digest algorithm.

## Test evidence

None. No implementation is authorized from this draft; the negative-test
specifications above are the executable contract that `dap-jo9.2.5.9.2` must
satisfy **after** the owner accepts a source.

## Follow-ups

- `dap-jo9.1.1.6` — open; blocked on product-owner acceptance of a source.
- `dap-jo9.2.5.9.2` — stays blocked until this decision is accepted and the
  implementation contract can verify the selected source.
- `dap-jo9.2.5.9.1` — authority-independent simulated contract cases; separate
  and unaffected by the source decision.
- Scope limits (from the issue): this decision records an owner decision and
  testable contract, not a provider integration, authentication mechanism, or
  human-review implementation. The accepted history of `dap-jo9.1.1.4` and
  Decision 0004 remains intact. No claim of human or assistive-technology
  acceptance is made here.
