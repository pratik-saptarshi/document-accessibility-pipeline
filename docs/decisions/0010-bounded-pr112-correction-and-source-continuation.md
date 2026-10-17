# Decision 0010: Bounded PR #112 correction and Beads source continuation

**Status: Accepted.** The repository owner approved this exact correction in
the 2026-10-17 task instruction. This record contains no private conversation
identifiers or transcript excerpts.

**Date: 2026-10-17.**

## Decision

Authorize exactly one canonical-export correction: from the source export with
SHA-256
`5bb63fe2b6678b391360966af394adbab1c5b4e8bf7b9f99ad01e535a723087a`, remove
the `notes` property from issue `dap-jo9.1.6.7`, restoring the field's absence
in the pre-PR #112 export. The resulting exact export SHA-256 is
`6d2993b1248396fb1b180423551b67346e4f6a576060d487dd7d4cb768448459`.

The operation must preserve every other issue row and field, including the
target's status, timestamps, labels, counts, and dependency edge. It must
preserve PR #114's closure of `dap-jo9.1.6.10`. It authorizes no Beads
normalization, live embedded-store mutation, other field or issue change,
status/closure change, or dependency-graph change.

Because the exact candidate export becomes the next canonical source, add
only this candidate digest to the source allowlist for Decision 0007's
existing explicit-null preservation behavior. This snapshot contains four of
the existing five null pairs: PR #114 has populated the factual
`dap-jo9.1.6.10.closed_at`, so that field must be preserved as-is and must not
be treated as a lost null. The remaining pairs are
`dap-jo9.5.3.dependencies`, `dap-jo9.5.6.closed_at`,
`dap-jo9.2.1.4.closed_at`, and `dap-jo9.5.7.closed_at`. The Beads 1.3.1
executable SHA-256 remains
`a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. No new
null pair, value, binary, version, or operation is approved.

## Context and evidence

PR #112 review finding [4301459486](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/112#discussion_r4301459486)
identified that its `notes` array on `dap-jo9.1.6.7` violates the repository's
scalar Beads notes contract. Finding
[4301459487](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/112#discussion_r4301459487)
identified that the spike report misstated test-local stub examples as
statement/invoice fixture controls. The corrected report and tests distinguish
paper hypotheses and gate-only tests from fixture evidence; no PDF fixture is
opened or parsed by this correction.

The source is the current main export after PR #114 and PR #120. The bounded
execution plan is
[`2026-10-17-pr112-mapping-evidence-correction.md`](../superpowers/plans/2026-10-17-pr112-mapping-evidence-correction.md).

## Alternatives considered

- Restore a whole historical export: rejected because it would discard PR
  #114 and other current-main changes.
- Reconcile through Beads import/export: rejected because this repair requires
  restoring one malformed source field and must not normalize unrelated
  values.
- Leave the malformed array and correct only documentation: rejected because
  the array violates the established scalar notes contract.
- Broaden the correction to other fields, issues, statuses, or graph edges:
  rejected because the owner approved only this exact field removal.

## Consequences and limits

The candidate digest is the only newly approved source digest for the
source-specific set of four explicit-null losses described above. The older
source snapshots retain their existing five-pair set. Any later source edit
creates a new digest and remains blocked until separately reviewed and
explicitly approved. This correction does not change a live Beads store,
restore Dolt history, or establish fixture results or mapping accuracy.

## Acceptance evidence

Completion requires exact preimage/candidate digest verification; confirmation
that only the target `notes` property was removed; regression and prescribed
root/nested checks; independent review on the exact diff; publication through
the guarded HTTPS CloudBot Neo-green Conventional Commit path; and fresh
required checks on the exact correction PR head. The decision itself records
authorization, not completion.

## Follow-up

- Keep any later canonical-export source digest, field change, issue change,
  status/closure update, or graph change outside this approval.
- Do not write to or initialize the configured live Beads database.
