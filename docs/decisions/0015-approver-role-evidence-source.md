# Decision 0015: Approver-role evidence source — owner-approved append-only role-grant register

**Status: Accepted source category and temporal policy; authenticated-source
contract partly specified, implementation prerequisite remains open.** The
product owner accepted the register source category on 2026-10-18 with the
instruction "provide and document details with this decision outlining
rationale." This record does not name an actual approver, create the
register, select its trust anchor or maintenance controls, or implement a
trusted verifier. The issue `dap-jo9.1.1.6` records the source-category
decision; it does not close authenticated-source implementation or release
evidence requirements.

**Date: 2026-10-18.**

## Context

DRAFT-0014 Q4 recommended the three-role human review authority model
(document owner / accessibility reviewer / approver) but could not close Q4
without an owner decision on where the approver's identity comes from — the
open question carried as DRAFT-0010 (`dap-jo9.1.1.6`, "Select authoritative
approver-role evidence source and temporal policy"). Neo directed that the
decision be provided with documented details and rationale; this record is
that decision.

[Decision 0004](0004-approver-persona-review-protocol.md) (accepted 2026-10-14)
designates the approver as a product **role** — the authority to review,
approve, or reject remediation candidates on semantic and assistive-technology
grounds — and defines the review-evidence record format (`approver_identity`,
`reviewed_at`, outcomes). It deliberately does **not** name an authoritative,
time-aware source proving the acting person held the approver role at
`reviewed_at`. The product-owner designation of the role itself
(`dap-jo9.1.1.4`, closed 2026-10-14) is kept intact; this record changes
neither it nor Decision 0004.

The selected source category/policy, its authenticated-source contract, the
named person acting as approver, and the trusted verifier are distinct.
Naming the actual approver is release-stage work; it is not a prerequisite to
implementing or functionally testing the approved profile,
family-recognition, or WCAG 2.2 contrast features. Synthetic identities may
be used to test functional contracts but never count as genuine release
approval. A real release still requires genuine authorized semantic and
assistive-technology evidence under Decision 0004, bound to the exact
candidate.

The PR #71 P1 review finding that raised the evidence-source question is
<https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/71#discussion_r4277123170>.

## Decision 1: Three-role model adopted (with recorded, never hidden, dual-hatting)

The pipeline adopts the three-role model from [DRAFT-0014 Q4](0014-profile-policy-decisions.md):

1. **Document owner** — accountable for the document's content (SMB: the
   business owner; enterprise: the publishing team lead). Declares the profile
   (family, language, confidentiality).
2. **Accessibility reviewer** — runs the human-judgment checklist. Must be
   someone other than the document owner where feasible; where it is not
   (the common SMB case), the dual-hat review is **recorded as dual-hat in the
   review evidence**, never silently collapsed.
3. **Approver** — holds release-signoff authority per Decision 0004.

**Dual-hat rule.** One person may hold two roles on a given document only if
the dual-hat flag is recorded in the review evidence. A review or release
record that omits the flag when roles were shared is treated as missing
evidence and blocks release. Recording is the control; separation is the
preference.

## Decision 2: Evidence source — owner-approved append-only role-grant register

The authoritative evidence source for the approver identity is the
**owner-approved append-only role-grant register**: a register owned and
explicitly approved by the product owner, recording each grant of the approver
role to a named person with a UTC validity interval. Grants are never edited in
place; expiry and revocation are new appended entries. The register read must
authenticate the register identity and prove a complete, ordered snapshot;
an identity string, sequence number, or entry digest alone is not source
authentication.

### Rationale

1. **Append-only gives tamper-evident history.** The release gate must verify
   role authority at two points in time — the UTC `reviewed_at` timestamp and
   release re-verification. A register where every grant, expiry, and
   revocation is a new immutable entry lets the gate re-fetch the exact bytes,
   recompute the digest, and prove what was true at `reviewed_at` long after
   the fact. Retroactive edits are structurally impossible, which is the only
   audit-trail posture that survives scrutiny in the mandate-readiness story
   Neo is building toward Q2 2027.
2. **Owner-approval keeps policy under the product owner's control.** The
   owner approves the source category and grant policy; the authenticated
   read mechanism and writer controls still need an explicit trust contract.
   This record selects no third party, provider integration, credential, or
   trust root. A digest does not establish who appended an entry or when.
3. **No dependency on an external identity system the pipeline doesn't own.**
   A named organizational IAM system was the considered alternative; it was
   rejected because it introduces a provider integration and an availability
   dependency at the release gate, both outside this repository's accepted
   scope (see product scope and AGENTS.md). The fail-closed rule would still
   require an availability contract the owner does not currently hold with any
   provider. The register fits the operation Neo actually runs today: named
   individuals, small teams, explicit owner sign-off.

### Alternatives rejected

- **Named organizational identity and access-control system.** Reasonable for a
  larger organization with an existing system of record, but it adds an
  external runtime dependency and availability contract at the release gate,
  outside current accepted scope. Reconsidered only if the client base reaches
  enterprise scale with an IAM the owner trusts.
- **No source selected (fail-closed current posture).** Preserves safety but
  blocks every genuine human/AT approval from ever passing the release gate,
  so the bounded workflow could never complete a release. Rejected as the
  steady state; it remains the interim posture only in the form below.

## Decision 3: Register semantics (temporal policy)

The following source and temporal rules are accepted. They define a
constructible record contract, but do not decide a specific trust mechanism
or register-maintenance authority:

- **Entry shape and digest.** Each immutable grant records `schema_version`,
  `entry_type`, canonical `identity`, `role`, UTC `not_before` and
  `not_after`, strictly increasing `sequence`, and source-authenticated
  `appended_at`. A revocation records `schema_version`, `entry_type`,
  `revokes_sequence`, its own increasing `sequence`, source-authenticated
  `appended_at`, and UTC `effective_at`. All times use RFC 3339 UTC `Z` form
  with whole-second precision. Grant, revocation, and exception records may
  not contain unrecognized fields. The stored `entry_sha256` is excluded from
  its own digest input. Compute SHA-256 over the RFC 8785 JSON Canonicalization
  Scheme serialization of the entry with `entry_sha256` omitted; reject
  duplicate keys, unknown fields, noncanonical times, and any digest
  mismatch. This makes the digest deterministic and constructible without
  changing its own input bytes. A digest identifies bytes; it does not
  authenticate an issuer or append time.
- **Narrow exception entry shapes.** Each exception is a JSON object with
  `schema_version: 1`, its exact `entry_type`, a strictly increasing unique
  `sequence`, source-authenticated UTC whole-second `appended_at`, and the
  type-specific fields below. `retroactive_grant_exception` has
  `entry_type: "retroactive_grant_exception"`, one positive integer
  `grant_sequence`, a non-empty lexicographically sorted array of unique
  non-empty string `affected_review_ids`, and a non-empty string
  `decision_evidence_ref`. It permits that grant's `not_before` to predate its
  append time only for those listed reviews. `revocation_exception` has
  `entry_type: "revocation_exception"`, one positive integer
  `revocation_sequence`, a non-empty lexicographically sorted array of unique
  non-empty string `affected_review_ids`, and a non-empty string
  `decision_evidence_ref`; it may invalidate only those named past reviews.
  Neither shape permits additional fields except its `entry_sha256`. Both use
  the same canonical digest rule and require an authenticated record of the
  product owner's case-specific decision. The reference identifies that
  decision evidence; it does not authenticate it.
- **Append time and ordering.** `appended_at` is returned by the authenticated
  source and is immutable; it is never supplied or backdated by a grant
  author. Entries have unique, contiguous, strictly increasing sequence
  numbers through the authenticated snapshot high-water mark, and
  `appended_at` values do not decrease with sequence. A missing, duplicate,
  reordered, malformed, or unauthenticated entry makes the snapshot
  incomplete and fails closed.
- **Golden digest vectors.** The grant vector's canonical digest input (the
  `entry_sha256` property omitted) is
  `{"appended_at":"2026-10-19T00:00:00Z","entry_type":"grant","identity":"reviewer@example.test","not_after":"2027-12-31T23:59:59Z","not_before":"2026-10-19T00:00:00Z","role":"approver","schema_version":1,"sequence":1}`.
  Its SHA-256 is
  `9db7010b43620e12be1a0d907bc826fee6083ddb3de44ab848f0ad248cfdab05`. The
  revocation vector's canonical digest input is
  `{"appended_at":"2026-10-19T12:00:00Z","effective_at":"2026-10-19T12:00:00Z","entry_type":"revocation","revokes_sequence":1,"schema_version":1,"sequence":2}`.
  Its SHA-256 is
  `35ead8167dce1f74a4090bfdaf3927a732814f9fd195848e036381546a2b28ad`.
  These vectors test serialization and digest construction only; they are
  synthetic and confer no authority.
- **Identity mapping.** `approver_identity` in the review record must match a
  register entry's canonical identity exactly. Self-asserted identity or role
  labels with no register entry at `reviewed_at` are rejected as missing
  authorization evidence.
- **Validity at `reviewed_at`.** A grant authorizes a review iff
  `not_before <= reviewed_at <= not_after`. A grant whose `not_after` precedes
  `reviewed_at` is **expired**: the review is rejected; expiry is automatic and
  irreversible except by a new grant entry. `reviewed_at` must be a valid UTC
  timestamp. The closed interval applies at both exact boundaries.
- **Revocation.** A revocation names exactly one grant by `revokes_sequence`
  and applies prospectively from its `effective_at`, inclusive: reviews with
  `reviewed_at < effective_at` remain valid under that grant; reviews at or
  after that instant are rejected under it. `effective_at` may not precede
  the revocation's authenticated `appended_at`. A new grant is a distinct
  successor entry and does not erase the revocation. A revocation prompted by
  a security concern may be escalated to the product owner for a case-by-case
  determination. It changes past-review validity only through an authenticated
  `revocation_exception` naming the revocation and affected review IDs; no
  such exception is inferred.
- **Retroactive grants.** Not permitted by default: `not_before` may not
  predate the grant's authenticated append time. Exception only with an
  authenticated `retroactive_grant_exception` naming the exact grant and
  reviews it covers — expected to be rare and audited.
- **Unavailable or incomplete source.** If the register cannot be read, its
  authenticity cannot be established, or the snapshot has a sequence gap,
  malformed entry, unverifiable digest, or untrusted append time at
  verification or re-verification time, the gate returns blocked/review-needed.
  Unreachability is never treated as approval and never downgraded to
  "evidence absent."
- **Replaced evidence.** The immutable reference (authenticated register
  identifier, entry sequence, entry digest, and snapshot high-water mark) and
  recomputed digest are retained in the review evidence. Replaced bytes,
  wrong-source references, or a mismatch reject the evidence. Re-verification
  may use an authenticated successor snapshot only when it preserves the
  referenced history and proves the required entry was valid at
  `reviewed_at`; any re-pointing is explicit and audited.
- **Digest identifies bytes; it does not authenticate the source.** The
  verifier must authenticate the configured register identity and snapshot
  completeness through a separately accepted source adapter. No provider,
  trust root, signing mechanism, or administrative control is selected here.

### Abstract authenticated-source/verifier contract

The verifier consumes an `AuthenticatedRegisterSnapshot` containing the
stable register identifier, source-authentication evidence, snapshot
high-water mark, and the complete ordered entry set through that mark. Its
source adapter must authenticate the configured register and bind the
high-water mark and append times to that authenticated snapshot before the
verifier uses any entry. The verifier then validates entry shape, sequence,
digest, identity/role, time boundaries, revocations, and source continuity.
The adapter's trust anchor and the controls authorizing grant/revocation
writers remain an explicit owner/security prerequisite; this abstract
contract does not select or deploy an implementation.

## Decision 4: Interim posture until the register is implemented

The register does not exist yet, and no implementation is authorized from this
record beyond what `dap-jo9.2.5.9.2` covers. Until the register is implemented
and populated:

- The approver's actual identity, UTC timestamp, and dual-hat flag are
  recorded for any genuine review, as Decision 0004 requires. An identity
  string alone is **unverifiable-authority** and cannot satisfy release
  authorization. Until the register and verifier exist and produce
  verifiable evidence, release remains blocked/review-needed.
- No authority may be inferred from simulations or machine results; Decision
  0004's "genuine review only" rule applies unchanged.

## Consequences and limits

- This record documents the selected source and policy. It is not evidence
  that the source exists, that a verifier is implemented, or that any Beads
  issue was closed. Trusted-role verification and genuine candidate-bound
  human/AT release evidence remain open follow-up work.
- It does **not** implement the register, the verification interface, or any
  authentication mechanism; it authorizes no provider integration, no
  human-review workflow changes, and no new dependencies.
- It does not decide whether dual-hat reviews are acceptable for `restricted`
  documents — that remains Neo's open judgment from DRAFT-0014 Q4.
- Decision 0004's protocol, invalidation rules, and "machine results are
  context, not evidence" rule remain intact and in force.

## Synthetic contract cases

This record is policy documentation, not implementation or real verification
evidence. Synthetic identities/records may exercise functional test contracts
but are not real approvals. The future contract tests must include:

reject operator-only, unknown, unauthorized, expired, not-yet-valid,
malformed/time-invalid, missing, unverifiable, wrong-source,
identity/role-mismatched, replaced, forged, and digest-mismatched evidence;
wrong-source data with a correct digest; tampered bytes; a backdated grant;
revocation immediately before, at, and after its effective time; and a
successor snapshot that authenticates and preserves prior history. Accept
only when authenticated evidence binds the exact identity to the approver
role and its validity covers `reviewed_at`. Authority-independent synthetic
evidence-contract cases stay in `dap-jo9.2.5.9.1`; the provider-authentication
and writer-control prerequisite remains open and no synthetic result is a
human/AT pass.

## Follow-ups

- `dap-jo9.1.1.6` — records the accepted source-category selection. It does
  not satisfy authenticated-source implementation or release evidence; those
  remain separately gated below. The tracker record must not imply those
  criteria are complete.
- `dap-jo9.2.5.9.2` — remains open for trusted-role verification against the
  selected policy, as release-stage work; no named approver or verifier is
  presumed.
- DRAFT-0014 Q4 — source/policy design is documented; implementation and
  genuine release evidence remain distinct open work. The
  `restricted`-document dual-hat question remains open.
- Register authentication/trust anchor and maintenance controls (who may
  append grants/revocations, on what authority, with what change evidence) —
  unresolved owner/security prerequisites for trusted verification.
