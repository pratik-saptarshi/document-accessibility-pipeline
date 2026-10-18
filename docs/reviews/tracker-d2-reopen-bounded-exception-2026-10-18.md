# Bounded tracker correction record: D2 reopen

This record documents the narrow process exception used to correct the
tracked interchange export. It is not a new owner decision and does not claim
that the live Dolt database was changed.

## Scope

- Trusted source: `.beads/issues.jsonl` at protected-main commit
  `040fbdc5f6e84a50c65881adbbc40e18664b4b85`, SHA-256
  `341b2e2d96fe4bce8ad4ab2efbca8ad301d13e382f792b2dd8b28015fe5daaf1`.
- Candidate: SHA-256
  `1b57d8a08002789b08dc4fb1b02b0cebf48253e4106c14e8bc590765874bfc15`.
- Sole record changed: `dap-jo9.1.6.7`. Its previously recorded
  `2026-10-18T09:00:32Z` close preceded acceptance resolution; its existing
  D2 reopen note records the correction. The candidate sets `status` to
  `open`, `closed_at` to explicit `null`, `close_reason` to the empty string,
  `updated_at` to the recorded `2026-10-18T15:55:00Z`, and appends only that
  D2 note.
- Every other issue row and field is preserved from the trusted source,
  including `dap-5hy.3`'s PR #152 closure and the still-open/pending tracker
  states for `dap-jo9.2.2`, `.1.6.3`, and `.1.6.4`.

## Guard evidence and limits

`ValidateDirectEditTransition` pins the exact source row bytes for the D2
target, compares the complete source and candidate exports field by field,
and rejects ID changes, added or removed fields, explicit-null loss, timestamp
restoration, and all other lifecycle changes. The live-file test resolves its
trusted source from the PR base SHA or push-before SHA. The focused test passed
with `DAP_BEADS_SOURCE_REF=origin/main` after this candidate was written.

The bounded correction was applied as an explicit process exception because
the configured `bd`/embedded-Dolt path was not usable in this environment. No
claim is made that `bd`, the historical guarded `promote`, or a live database
transition ran. The candidate remains an interchange export only.
