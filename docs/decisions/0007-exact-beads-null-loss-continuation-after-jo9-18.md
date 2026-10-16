# Decision 0007: Exact Beads null-loss continuation after `dap-jo9.18`

**Status: Accepted.** The repository owner approved this exact, bounded source continuation on 2026-10-16.

**Date: 2026-10-16.**

## Decision

Add exactly one canonical source digest to the Beads explicit-null compatibility exception: `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`. This is the candidate export that adds `dap-jo9.18` to the previously approved continuation. The authorization retains the same Beads 1.3.1 executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653` and exactly the five field pairs below.

This supplements Decision 0006. The complete approved source set is the historic source `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`, its approved continuation `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`, and the exact continuation in this decision. No other source digest, Beads version or binary, null pair, null value, field change, or dependency-graph change is approved.

| Issue | Field whose original explicit `null` is preserved |
|---|---|
| `dap-jo9.5.3` | `dependencies` |
| `dap-jo9.1.6.10` | `closed_at` |
| `dap-jo9.5.6` | `closed_at` |
| `dap-jo9.2.1.4` | `closed_at` |
| `dap-jo9.5.7` | `closed_at` |

## Context and evidence

PR #87 added one reviewed P2 issue row and preserved the prior 99 records. Its canonical candidate digest is the exact source authorized here. The guarded reconciler captured Beads 1.3.1 imports in disposable stores, retained the five original null tokens in the candidate, and verified the candidate through a separate-store round trip. This decision authorizes only that source lineage for the existing compatibility behavior; it does not itself claim that a live Beads store was updated or that Dolt history was restored.

## Alternatives considered

- Keep only the two previously approved source digests and reject this continuation: rejected because the repository owner explicitly approved this exact digest while retaining the same binary and five field pairs.
- Accept a digest prefix, wildcard, configuration-provided digest, or future source: rejected because it would exceed the bounded approval and weaken source-lineage checks.
- Broaden the exception to additional null pairs, changed values, binaries, versions, or dependency changes: rejected because none is included in this decision.

## Consequences and limits

The reconciler may apply the existing five-pair null-preservation path to this exact source and pinned Beads executable. Any change to source bytes creates a different digest and remains rejected unless separately reviewed and explicitly approved. The reconciler must continue rejecting altered baseline values, unexpected semantic or graph changes, unlisted fields, and wrong Beads versions or binary digests.

This decision authorizes no live-store write, canonical tracker mutation, issue update, or issue closure. It records only the bounded source-lineage approval. Any tracker operation remains a separately reviewed guarded action.

## Test evidence

The implementation adds an exact-byte source fixture and `TestApprovedIssueCreationSourceContinuationPreservesNulls`. The test checks the approved digest, unchanged candidate bytes, exactly the five approved losses, and second-store candidate re-derivation; it also rejects a changed source, a wrong binary or version, a non-absent baseline value, and an unexpected dependency change. Existing historic/current-source, wrong-field, wrong-binary, and tamper regressions remain in place.

## Follow-up

- Keep all future source continuations blocked until each exact digest receives separate owner approval, an indexed accepted decision, and a reviewed test change.
- Do not modify `.beads/issues.jsonl` or a live Beads store as part of this decision implementation.
