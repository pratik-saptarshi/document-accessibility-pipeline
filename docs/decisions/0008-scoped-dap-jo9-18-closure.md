# Decision 0008: Scoped canonical-export closure for `dap-jo9.18`

**Status: Accepted.** The repository owner authorized this exact operation in the task instruction dated 2026-10-16. This record intentionally contains no private conversation identifiers or transcript excerpts.

**Date: 2026-10-16.**

## Decision

Authorize a new, explicitly typed Beads close operation for exactly `dap-jo9.18` against the exact canonical source snapshot with SHA-256 `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`, using only Beads 1.3.1 with executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653` in isolated disposable stores. The operation may record `status=closed`, the actual Beads-generated `closed_at`, the approved evidence-bearing `close_reason`, and Beads-generated `updated_at` for this issue only.

After exact-source validation, real close-command evidence, candidate review bound to the immutable bundle and candidate digests, and the required protected checks, the owner authorizes the existing guarded `promote` operation to update the tracked canonical JSONL export. The close reason is one non-empty valid UTF-8 string no larger than 64 KiB and is passed as an argv value. A refused close must abort; force and automatic continuation are not authorized.

This decision authorizes no write to, initialization of, or migration of the configured live embedded Dolt database; no restoration or claim of Dolt history; no generic status/closure fields; no other issue closure; no creation/update mixed with closure; no dependency or other graph change; and no widening or alteration of Decision 0007's five explicit-null pairs, source digests, Beads version, or executable pin. The JSONL promotion does not itself change a live Beads store. No future updater run may treat the promoted candidate digest as authorized by Decision 0007 or this decision without further exact-source approval.

**Implementation update (2026-10-16; main `76f14f69a3283715f5d5278983363121168a2f09`):** PR #97 merged the authorized typed workflow and reviewed canonical-export promotion. The current `.beads/issues.jsonl` records `dap-jo9.18` as closed. The operation did not inspect, change, or verify the live embedded Dolt database; the authorization remains limited to the canonical export and exact source/binary scope above. See the [candidate review](../reviews/dap-jo9.18-closure-candidate-2026-10-16.md) for bundle, candidate, and verification evidence.

## Context and evidence

The accepted test-only work for `dap-jo9.18` is present on reviewed main. The existing updater accepts only planned string-field updates and one bounded creation path. A closure therefore requires a separate typed operation and must not weaken generic field validation. The existing lifecycle imports the canonical export into a disposable store, captures B0 and B1, derives a candidate from original source bytes, round-trips the candidate in a second disposable store, creates an immutable evidence bundle, and promotes only after bundle/source revalidation under a cooperating-writer lock.

The exact source digest above is the source approved by Decision 0007. This decision permits that exact source to receive one target-specific closure delta while retaining the same Beads 1.3.1 binary and the same five explicit-null pairs. It adds no source digest or null pair to the compatibility exception. Any other source, binary, version, null field/value, target, graph delta, or operation is outside this approval.

## Alternatives considered

- **Use generic status/closed fields in the existing update plan:** rejected because that would turn a narrow close into arbitrary lifecycle editing.
- **Close a live embedded Dolt issue directly:** rejected; this authorization is limited to isolated stores and the tracked JSONL candidate/promotion.
- **Promote without exact-source evidence and independent candidate review:** rejected; candidate derivation and promotion remain source-bound and guarded.
- **Broaden Decision 0007 to a future digest, binary, version, null pair, or graph delta:** rejected; each such change requires separate exact authorization.

## Consequences and limits

The implementation must enforce exact target, source digest/status, binary/version, close count and argv, reason bound, actual B0-to-B1 event, allowlisted candidate delta, unchanged dependency graph and unrelated bytes, existing five-null exception, second-store round-trip, immutable bundle verification, stale-source checks, cooperative lock behavior, and atomic-replacement failure semantics.

The resulting candidate and any promoted `.beads/issues.jsonl` record only an interchange-export closure. They do not prove or claim that the live embedded Dolt database was updated or that prior Dolt history was recovered. A failure after atomic rename but during directory synchronization must be reported as replacement already having occurred; rollback must not be claimed.

## Acceptance evidence

This decision records authorization, not completion. Completion requires the exact-source synthetic Beads close regression, focused red/green and negative/mutation tests, full repository-required local and protected CI gates on the final exact head, independent specification and code-quality review, exact bundle/candidate review, and verification that the promoted canonical export differs only by the authorized target fields plus Beads' `updated_at`. The decision remains distinct from the owner-authorized independent review's technical acceptance of `dap-jo9.18`'s existing criteria.

## Approval

Owner approval for Decision 0008 is recorded from the authenticated owner instruction in the task instruction dated 2026-10-16. The source instruction remains private. This decision record is an agent-authored transcription of that bounded authorization; it documents the approval but is not itself the approval. Implementation followed separately: PR #97 merged the authorized typed workflow after independent specification and code-quality review and the [candidate review](../reviews/dap-jo9.18-closure-candidate-2026-10-16.md). That review covers the candidate's technical acceptance, not a re-approval of this decision's scope.

## Follow-up

- Implement only under `docs/superpowers/plans/2026-10-16-beads-jo9-18-closure-updater.md` and the source-publisher/branch-protection process.
- Do not write or initialize the live Beads database.
- Any future closure, source continuation, null-loss change, binary/version change, extra field/issue, or graph delta requires a separately reviewed and accepted decision.
