# PR #125 Non-Gate Step Controls

## Goal

Address the active PR #125 review finding that a non-gate step can be skipped
with `if:` or allowed to fail with `continue-on-error:`. Work from
current `main` (`26091508650f7658894957df1c4f78c6f72daa5b`) on a separate
branch. Preserve #125, its review threads, and tracker state.

## Acceptance checks

1. Regression tests show `validateScannerIsolation` rejects a conditional
   non-gate step and a non-gate step with `continue-on-error: true`, naming the
   affected step.
2. The unmodified current workflow remains accepted; existing gate-control and
   workflow policy tests remain green.
3. The validator applies the two execution-control checks to every step in the
   `checks` job, without adding body pins for unrelated steps or changing
   workflow syntax.
4. Required root and nested-module gates, traceability, publisher checks,
   hygiene, independent review, and exact-head hosted checks pass before merge.

## Test-first sequence

1. Add conditional and fail-open mutations for existing non-gate steps; confirm
   the current validator accepts them (RED).
2. Add the smallest per-step rejection in `validateScannerIsolation`; confirm
   both regressions and the unchanged workflow pass (GREEN).
3. Review the exact base/head, run required checks, scrub, and publish a
   separate draft PR. Wait for owner review before protected squash merge.

## Scope boundary

Do not amend or resolve #125's review comments, update `.beads/issues.jsonl`,
or transplant policy tests already superseded on current `main` by PRs #134–139.
