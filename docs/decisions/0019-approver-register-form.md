# Decision 0019: Approver-role register — standalone signed append-only file

**Status: Accepted.** The product owner selected the concrete register form on
2026-10-21: a standalone signed append-only JSONL file with hash chaining,
verified by a small adapter. This record does not name an actual approver,
create the register, select its signing key, or implement the adapter.

**Date: 2026-10-21.**

## Context

[Decision 0015](0015-approver-role-evidence-source.md) (accepted 2026-10-18)
accepted the source *category* — an owner-approved append-only role-grant
register — and the temporal policy, but left the concrete form open. The
2026-10-19 tracker correction on `dap-jo9.1.1.6` noted that completion still
required a concrete register identity and trust anchor, an authenticated
source adapter/interface, and authorization controls for grant/revocation
writers. This record resolves the concrete form; the writer-authorization
question is explicitly out of scope (see below).

## Decision

The register is a **standalone signed append-only JSONL file with hash
chaining**, verified by a **small adapter**:

1. **Standalone file** — the register lives outside the codebase (not in the
   repo's git history), giving clean separation between role evidence and
   product source.
2. **Append-only with hash chaining** — each entry commits to the previous
   entry's hash, so deletion or reordering is detectable. The file is never
   rewritten; grants and revocations are appended as new entries.
3. **Signed** — entries (or periodic checkpoints) carry signatures from the
   register's trust anchor, so the adapter can authenticate the issuer, not
   just the bytes.
4. **Small adapter** — the pipeline verifies through a narrow interface, not
   by reading the file directly. The adapter must prove, for a given UTC
   `reviewed_at` timestamp: the named identity held the approver role at that
   time, the grant had not been revoked as of that time, the snapshot is
   complete (high-watermark continuity — no gaps in the entry sequence), and
   the source append times are immutable.

## Explicitly out of scope

**Who is authorized to write grants and revocations is beyond the scope of
this repo and design** (product-owner ruling, 2026-10-21). The trust anchor's
key custody, the writer set, and the operational process for issuing grants
belong to the deploying organization, not to this pipeline's design. The
adapter contract above is specified independently of who holds the keys.

## Consequences

- The release gate can validate an approval against the register at
  `reviewed_at` and re-check the immutable reference at release time, without
  trusting a self-asserted role string.
- Implementing the adapter and provisioning the register (file location,
  signing key, writer process) is engineering and deployment work, tracked
  separately; it is not a product decision.
- Synthetic identities may test the adapter contract but never count as
  genuine release approval, per Decision 0015.

## Alternatives considered

- **Hash-chained log in the repo itself** (signed commits, git history as the
  timeline): zero new infrastructure, but couples role grants to the codebase.
  Rejected for separation.
- **Defer the concrete pick** (record only the adapter interface contract):
  rejected — the owner selected the standalone file as the concrete form now.
