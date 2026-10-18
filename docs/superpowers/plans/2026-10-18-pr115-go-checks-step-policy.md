# PR #115 Go checks step inventory policy

**Status:** Authorized implementation slice; plan written against refreshed `origin/main` `6a80e22b683516deefbac7ad866bc84486f956b4`. No tracker state is changed by this work.

## Scope

Harden the existing `internal/devsecops` test-only workflow policy on current main. Preserve its strict workflow decoding, trigger/fail-open checks, scanner isolation, and all nine root/nested module gates. Do not merge PR #115's conflicting branch wholesale, copy its validator as a second implementation, change `.beads/issues.jsonl`, or alter workflow behavior.

The Go checks `checks` job currently has 16 ordered steps. The existing policy proves the nine module-gate commands remain present but does not bind the complete inventory/order or the configuration of non-gate setup steps. A future step could change Go's persistent configuration, repurpose dependency setup, or alter action inputs while the required module commands still appear.

## Acceptance criteria

1. One existing validator checks exactly the 16 current `checks` job steps in order. An added, removed, duplicated, or reordered step is rejected.
2. Every non-gate step is pinned by exact name, action reference, command, working directory, and complete `with` map. Existing action-SHA/comment validation remains in force.
3. Every non-gate step rejects step-level `env` and `shell`; existing module-gate `env`/`shell` rejection remains in force.
4. Tests demonstrate rejection of an added persistent-Go-configuration step, reordered steps, a repurposed `Install poppler-utils` command, and an altered checkout input. Tests also retain the nine gate-removal controls and exercise non-gate `env`/`shell` rejection.
5. Each new regression reaches the intended policy assertion, goes RED on the pre-fix main behavior, then GREEN after the smallest validator change. Existing current-main controls remain intact.
6. The issue record, workflow YAML, fixtures, and unrelated PR132/PR138 work remain untouched.

## Execution and verification

- Consolidate the nine current gate expectations and seven non-gate expectations into one ordered step-spec table; do not maintain parallel inventories or a second validator.
- Add compiling regression subtests against the existing `validateScannerIsolation` before implementing the missing policy.
- Run the focused `internal/devsecops` test first, then the repository checks in `AGENTS.md` (root, cloud bootstrap module, adapter contract module, traceability, shell syntax, and publisher guard). Java/veraPDF remains CI-only.
- Request independent specification and code-quality/security reviews against the exact merge base and head. Fix and re-review any findings.
- Publish with the guarded HTTPS source-publisher path and Conventional Commit subject, open a focused PR, and wait for every required check on the exact PR head. Merge only through protected squash after reviews/findings and checks are clear.
