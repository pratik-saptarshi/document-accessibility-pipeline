# Decision 0015: Approver-role evidence source — owner-approved append-only role-grant register

**Status: Accepted as policy selection.** Product owner accepted 2026-10-18
with the instruction "provide and document details with this decision
outlining rationale." This record selects an evidence-source design and
temporal policy. It does not name an actual approver, create the register, or
implement a trusted verifier. No Beads status transition is recorded here;
implementation and release-evidence gates remain open.

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

The selected source/policy, the named person acting as approver, and the
trusted verifier are distinct. Naming and verifying the actual approver is
release-stage work; it is not a prerequisite to implementing or functionally
testing the approved profile, family-recognition, or WCAG 2.2 contrast
features. Synthetic identities may be used to test functional contracts but
never count as genuine release approval. A real release still requires
genuine authorized semantic and assistive-technology evidence under Decision
0004, bound to the exact candidate.

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
place; expiry and revocation are new appended entries.

### Rationale

1. **Append-only gives tamper-evident history.** The release gate must verify
   role authority at two points in time — the UTC `reviewed_at` timestamp and
   release re-verification. A register where every grant, expiry, and
   revocation is a new immutable entry lets the gate re-fetch the exact bytes,
   recompute the digest, and prove what was true at `reviewed_at` long after
   the fact. Retroactive edits are structurally impossible, which is the only
   audit-trail posture that survives scrutiny in the mandate-readiness story
   Neo is building toward Q2 2027.
2. **Owner-approval keeps it under Neo's control.** The register's authority
   is the product owner's approval itself — no third party, no provider
   integration, no credential the pipeline does not own. The register's
   integrity (who wrote what, when) is protected by controls the owner accepts
   alongside this record, not by an external trust root.
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

The following are decided, not deferred:

- **Source definition.** The register is the product owner's named register
  for approver-role grants. Each entry records: the person's canonical
  identity (exact match to the review record's `approver_identity`; any case or
  email normalization is owner-decided and pinned, never inferred),
  `not_before` / `not_after` UTC bounds (closed interval), entry sequence
  number, and the SHA-256 digest of the exact entry bytes.
- **Identity mapping.** `approver_identity` in the review record must match a
  register entry's canonical identity exactly. Self-asserted identity or role
  labels with no register entry at `reviewed_at` are rejected as missing
  authorization evidence.
- **Validity at `reviewed_at`.** A grant authorizes a review iff
  `not_before <= reviewed_at <= not_after`. A grant whose `not_after` precedes
  `reviewed_at` is **expired**: the review is rejected; expiry is automatic and
  irreversible except by a new grant entry.
- **Revocation.** A revocation entry applies **prospectively**: it stops future
  reviews under that grant. Reviews already recorded while the grant was valid
  remain valid evidence. Rationale: retroactive invalidation would make the
  register's own history unreliable as an audit trail; append-only semantics
  make retroactive deletion impossible by design. A revocation prompted by a
  security concern may be escalated to the product owner for a case-by-case
  determination, recorded as a new entry.
- **Retroactive grants.** Not permitted by default: `not_before` may not
  predate the entry's append time. Exception only with an owner-approved
  retroactive-grant justification entry naming the exact reviews it covers —
  expected to be rare and audited.
- **Unavailable source.** If the register cannot be read at verification or
  re-verification time, the gate returns blocked/review-needed. Unreachability
  is never treated as approval and never downgraded to "evidence absent."
- **Replaced evidence.** The immutable reference (register identifier, entry
  sequence, entry digest) plus a recomputed SHA-256 digest of the exact source
  bytes are retained in the review evidence record. A digest mismatch or
  replaced entry rejects the review, unless a newer entry also binds the
  identity to the approver role and covers `reviewed_at`, in which case the
  record is re-pointed with an audit trail.
- **Digest identifies bytes; it does not authenticate the issuer.** Issuer
  authenticity comes only from the owner-accepted register.

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

## Test evidence

None. This record is policy documentation, not implementation or verification
evidence. Synthetic identities/records may exercise functional test contracts
but are not real approvals. The negative-test specifications for future
trusted verification include:

reject operator-only, unknown, unauthorized, expired, not-yet-valid,
malformed/time-invalid, missing, unverifiable, wrong-source,
identity/role-mismatched, replaced, forged, and digest-mismatched evidence;
accept authority only when the trusted evidence binds the exact identity to
the approver role and its validity covers `reviewed_at`. Authority-independent
synthetic evidence-contract cases stay in `dap-jo9.2.5.9.1`.

## Follow-ups

- `dap-jo9.1.1.6` — policy selection is documented here; its tracker status is
  unchanged by this record.
- `dap-jo9.2.5.9.2` — remains open for trusted-role verification against the
  selected policy, as release-stage work; no named approver or verifier is
  presumed.
- DRAFT-0014 Q4 — source/policy design is documented; implementation and
  genuine release evidence remain distinct open work. The
  `restricted`-document dual-hat question remains open.
- Register maintenance controls (who may append grants/revocations, on what
  authority, with what change evidence) — for the implementation slice, not
  this record.
