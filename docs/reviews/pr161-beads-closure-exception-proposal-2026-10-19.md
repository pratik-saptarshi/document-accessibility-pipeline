# PR #161: proposed one-time Beads closure exception

**Status: Proposed for independent review. Not authorized for promotion.**

This record separates the bounded acceptance judgment from the tracker
transition. The written discovery criteria are supported; AC2f retains only
the owner's accepted fixture-association and literal-input gate evidence
strands. This is not an owner decision, a new acceptance interpretation, or
evidence of production mapping behavior.

## Exact guarded-workflow result

The canonical source at main commit
`d2779c1979149807de8c207acdd718db2a1eb8c5` has SHA-256
`d74a67530a728c991bfcb96943cb210b9acde94bddc564a5bf3f2486508ad563`.
The existing reconciler was invoked with a one-target closure plan for
`dap-jo9.1.6.7`, Beads version `1.3.1`, the pinned executable, and the current
source digest. It stopped before running Beads with:

```text
create: invalid update plan: closure source bytes do not match the exact authorized snapshot
```

The typed closure validator accepts only `dap-jo9.18`, its exact source
digest, and Decision 0008. Decision 0011 freezes tracker/governance tooling
changes and authorizes only existing safe capabilities. No code, source
allowlist, or null-preservation policy was changed.

## Disposable Beads evidence

The failure above was checked separately from a disposable Beads 1.3.1 run.
The pinned binary SHA-256 was
`a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`.
`bd metrics off` was run and confirmed `OFF` in the temporary Beads home.
The source was imported into a temporary embedded store; `bd prime`,
`bd ready`, and `bd show dap-jo9.1.6.7` confirmed the source issue was open.
The issue was closed there with a bounded AC1–AC4 evidence reason, then
exported. No configured workspace store or live tracker was written.

| Snapshot | SHA-256 | Result |
|---|---|---|
| Source `.beads/issues.jsonl` | `d74a67530a728c991bfcb96943cb210b9acde94bddc564a5bf3f2486508ad563` | 101 rows; 168 dependency edges |
| Beads B0 export | `6f20b6fde4b06a929834e975731bcf94b33d76062baaca3214c7ff49e04810c0` | Imported baseline |
| Beads B1 export after close | `7ba50e418135e3b1b1b0e795f7ac4e3501f41206449d633ed573052bb17a5638` | Only target row changed; fields `status`, `closed_at`, `close_reason`, `updated_at` |
| Fresh-store B2 export of B1 | `7ba50e418135e3b1b1b0e795f7ac4e3501f41206449d633ed573052bb17a5638` | Exact round-trip digest match |

All four snapshots contain 101 rows and 168 dependency edges; normalized edge
sets are identical. Beads import/export omits two explicit source nulls:
`dap-jo9.1.6.7.closed_at` and `dap-jo9.5.3.dependencies`. The proposed
canonical candidate preserves the unrelated `.5.3.dependencies: null`; the
target's `closed_at: null` is replaced by the actual B1 close timestamp
`2026-10-19T03:47:34Z`. No synthetic timestamp was used.

## Review candidate and proposed scope

Candidate SHA-256:
`e5d5317f2ea1de989e19d1e667790b7513601522e6c6633c75dc1c753f42f9c5`.
It is formed from the exact main source bytes by taking the actual B1 values
for the target's four closure fields. A field-level comparison verified:

- only `dap-jo9.1.6.7` changes;
- only `status`, `closed_at`, `close_reason`, and `updated_at` change;
- the existing parent-child edge and all other 167 dependency edges remain
  unchanged;
- the unrelated explicit null at `dap-jo9.5.3.dependencies` is preserved;
- the target close reason cites the bounded paper, synthetic controls, and
  owner-accepted AC2f, and explicitly withholds implementation or production
  authorization.

This candidate was computed for review outside the guarded reconciler after
its preflight rejection. It is not a reconciler evidence bundle and must not
be described as a guarded promotion. The PR branch retains the canonical open row; this candidate is not
checked in, promoted, or applied to any tracker. The proposed correction is
one operation-specific process exception, if independently accepted:
authorize this exact source digest, target, candidate digest, Beads
binary/version, four changed fields, and close-event provenance for a
protected canonical export update. Preserve the unrelated null and dependency graph; do not write
the live Beads database; add no generalized closure code, source wildcard,
future source digest, null pair, or policy change. Any promotion remains
blocked pending independent review and recorded authorization of this exact
exception.

## Public-tree hygiene disposition

The tracked-tree scrub searched for absolute machine paths and publisher
identity strings, excluding this skill's pattern examples. The path search found 18
lines in 9 files: `.github/workflows/security-scans.yml`, `CONTRIBUTING.md`,
`docs/planning/backlog-migration-batch-2-2026-10-13.md`,
`docs/research/pdfcpu-financial-table-clearance-2026-10-13.md`, the two
`docs/security/adversarial-panel-*.md` records,
`docs/superpowers/plans/2026-10-15-round-two-review-validation-and-plan.md`,
`docs/superpowers/plans/2026-10-17-dap-jo9-5-7.md`, and
`docs/validation/README.md`. These are intentional command examples, temporary
paths in retained historical evidence, and isolated workflow/test boundaries;
none is a newly introduced machine-specific path, so none was changed.

The identity search found 455 lines in 22 files: repository guidance and
imported skills, `.beads/issues.jsonl` historical tracker records, publisher
implementation and tests, required policy/decision text, and retained Beads
reconciler test fixtures. Those matches preserve the guarded CloudBot author
and committer identity or historical provenance; no accidental identity trace
was introduced, and enforcement was not altered.
