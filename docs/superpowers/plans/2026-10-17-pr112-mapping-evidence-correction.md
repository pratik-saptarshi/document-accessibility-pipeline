# PR #112 Financial-Table Mapping Evidence Correction

**Status (2026-10-17):** Prepared on the correction branch after synchronizing
with `origin/main` at `09209554b94d4252387eda290cc1cfa1d8a45e81`. The repository
owner approved this bounded correction. Publication and final merge remain
subject to the independent review and exact-head CI gates below.

## Goal and scope

Correct the two unresolved PR #112 review findings without broadening the
mapping spike or changing unrelated tracker state:

- Correct documentation and tests that described inline stub examples as
  statement/invoice fixture evidence. State clearly that no PDF fixture was
  opened or parsed and that all expected associations are hypotheses.
- Restore only the malformed `notes` array on `dap-jo9.1.6.7` to its prior
  absence. Do not change any other field, status, dependency edge, or issue
  row from current main.

Owner authorization is recorded in [Decision 0010](../../decisions/0010-bounded-pr112-correction-and-source-continuation.md).
That decision is bound to the exact preimage and candidate export digests.

## Guidance applied

No standalone upstream `superpowers:*` skill was present in the installed cloud
or executor skill catalogs. Planning follows the repository's
[`docs/superpowers/plans` index](README.md) and its reviewed-plan conventions,
alongside the installed local adaptations: TDD,
[`feature-to-PR`](../../../.agents/skills/dap-feature-pr-pipeline/SKILL.md),
[`requesting-code-review`](../../../.agents/skills/dap-requesting-code-review/SKILL.md),
and [`verification-before-completion`](../../../.agents/skills/dap-verification-before-completion/SKILL.md).
These are repository-local adapted guides; this plan does not claim that the
full upstream Superpowers implementation or subagent-driven planning skill is
installed.

## Current-main preservation invariants

- Review base: `09209554b94d4252387eda290cc1cfa1d8a45e81`; it includes PR #114
  (`205903b2cc5c230fa7cd61f8c1605baea618b07f`), which closed
  `dap-jo9.1.6.10`, and PR #120 branch guidance.
- Exact `.beads/issues.jsonl` preimage SHA-256:
  `5bb63fe2b6678b391360966af394adbab1c5b4e8bf7b9f99ad01e535a723087a`.
- Exact correction candidate SHA-256:
  `6d2993b1248396fb1b180423551b67346e4f6a576060d487dd7d4cb768448459`.
- The candidate is derived from current main by deleting only the `notes`
  property on `dap-jo9.1.6.7`. The property was absent immediately before PR
  #112. Every other field in that row and every other row is preserved, with
  status and dependency edges unchanged. The PR #114 closure remains intact.
- Do not run Beads import/export normalization or alter the live embedded
  store. The canonical JSONL change is the specifically authorized operation.

## Test-first sequence and acceptance

1. Pin the observed mismatch from PR #112 review comments #4301459486 and
   #4301459487: the `notes` array violates the scalar notes contract, while
   the report attributes inline test stubs to fixtures and overstates invoice
   coverage (including the missing `Unit amount` example).
2. Keep regressions that exercise only constructed strings, local stubs, and
   the confirmation/safety gate. Confirm the tests do not read or parse either
   PDF fixture and do not assert mapping accuracy or fixture-authored data.
   The original focused TDD run was red when the unconfirmed/unsafe guards
   were removed and green with those guards restored.
3. Correct the report to call this a paper comparison and to label proposed
   associations as hypotheses. Name the actual tests and state the limited
   behavior each observes.
4. Remove only the target `notes` property. Compare parsed JSONL against the
   exact current-main preimage and verify 101 rows remain, no other field or
   dependency changes, no status changes, all other rows unchanged, and the
   PR #114 closure remains present.
5. Add the exact resulting JSONL digest to the source-bound explicit-null
   guard only, retaining the same Beads binary/version and original field
   boundary. Because PR #114 has populated one previously-null `closed_at`,
   the new snapshot authorizes only the other four actual nulls. Add a
   regression that exercises that exact source through reconciliation and
   second-store round-trip, verifies those four nulls and PR #114 closure, and
   rejects a one-field source mutation. First run it red against the current
   allowlist; then add the exact digest and run it green.
6. Run the focused financialfixtures and beadsreconcile tests and the repository-required root
   and nested Go, traceability, shell, and source-publisher checks. Review
   `git diff --check`, links, tree hygiene, publisher identity, and the exact
   staged diff.
7. Obtain an independent exact-base/exact-head review of all changed paths,
   including both original review findings and the JSONL preservation proof.
   Publish through the guarded HTTPS source-publisher only after that review;
   open a correction PR and require fresh exact-head CI and review before
   protected squash merge.

## Out of scope

No PDF fixture analysis, mapping implementation, accuracy claim, new mapping
decision, automatic mapping, Beads status/closure/dependency change, live
database operation, PR #112 comment mutation, or AEP work. The correction
does not imply semantic or accessibility acceptance.
