# PR #161: proposed one-time Beads closure exception

**Status: Applied once under the exact scoped exception below; independently verified before promotion.**

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
source digest. Plan SHA-256: `060872d84fcd895a69784341f999aaf79c817859056e17b5313153ce7c6bcecb`.
The sanitized invocation was `go run ./cmd/beads-export-reconcile create
-canonical $SOURCE -plan $PLAN -output $BUNDLE -bd $BD`. It stopped before
running Beads with:

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
The isolated workflow used `bd init --non-interactive --prefix dap`,
`bd metrics off` followed by `bd metrics` (output: `Anonymous usage metrics: OFF`),
`bd import $SOURCE`, `bd prime`, `bd ready`,
`bd show --json dap-jo9.1.6.7` (open), and `bd export -o $B0`. The one-issue
transition was `bd close dap-jo9.1.6.7 --reason "$REASON"`; `$REASON` is the
exact `close_reason` value printed below. Then `bd export -o $B1` captured the
close. A second fresh store used `bd init --non-interactive --prefix
dap-pr161-b2`, `bd import $B1`, and `bd export -o $B2`. All commands ran
against temporary stores; no configured workspace store or live tracker was
written. `bd history dap-jo9.1.6.7 --json` records the close event at
`2026-10-19T03:47:34.017Z` (Dolt commit `ujvsurh8eetmpt5ud1mu93emmu6qnn18`).
The sanitized history output SHA-256 is
`24d326b4718e115f5501ea7c17552a14aba71f8d28b5a0e2632172372a3f2eaa`; the B1
JSONL timestamps have second precision. The exact shell-history transcript was
not retained, so the durable provenance here is the pinned binary identity,
plan digest, command sequence, source/B0/B1/B2 digests, and read-only Beads
history record.

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

The exact source target row had `status: open`, `closed_at: null`,
`close_reason: ""`, and `updated_at: 2026-10-18T15:55:00Z`. The projection
replaces only those values with the exact B1 values below.

Candidate SHA-256:
`77b72b7b4efb994fd7f4d1b15a989043e0ddf9215c5d646f4842e6b68368fb40`.
The projection is reproducible from the source digest above and these exact B1
values (shown in a fixed field order):

```json
{
  "status": "closed",
  "closed_at": "2026-10-19T03:47:34Z",
  "close_reason": "Bounded acceptance evidence: AC1 A-E comparison is documented in docs/research/financial-table-mapping-options-2026-10-16.md; AC2 synthetic controls and owner-approved bounded AC2f are documented in docs/research/financial-table-mapping-evidence-2026-10-18.md; AC3 leaves option selection to a separate decision; AC4 preserves the no-inference, no-automatic-mapping, and no-general-support boundaries. No mapping implementation or production authorization.",
  "updated_at": "2026-10-19T03:47:34Z"
}
```

To reconstruct the candidate, preserve the exact source bytes for the 100
non-target rows. Parse only the `dap-jo9.1.6.7` row as JSON, set exactly the
four values above, serialize that one row as compact UTF-8 JSON with
`ensure_ascii=False` and separators `(',', ':')`, and retain its final newline.
This yields one changed raw line and the candidate digest above. A field-level
comparison verified:

- only `dap-jo9.1.6.7` changes;
- only `status`, `closed_at`, `close_reason`, and `updated_at` change;
- the existing parent-child edge and all other 167 dependency edges remain
  unchanged;
- the unrelated explicit null at `dap-jo9.5.3.dependencies` is preserved;
- the target close reason cites the bounded paper, synthetic controls, and
  owner-accepted AC2f, and explicitly withholds implementation or production
  authorization.

An independent read-only audit of the temporary source/B0/B1/B2 and candidate
artifacts verified the source, B0, B1/B2 digests, 101 rows, 168 edges, exact
round-trip, raw-byte preservation of all 100 non-target lines, the four-field
target delta, edge equality, and preservation of the unrelated explicit null.
A sanitized, durable evidence projection is retained at
[`evidence/pr161-beads-roundtrip-2026-10-19.json`](evidence/pr161-beads-roundtrip-2026-10-19.json). It includes the source/B0/B1/B2 original digests and counts, exact target-field projections, a sanitized Beads history event, and the full dependency-edge inventory. Its SHA-256 is
`b69d5c717dfc7c983ddd8d0fc4f31ddc0f31bd6e5acfda400070808e977c744c`. Descriptions, acceptance text, notes, labels, assignees, creator data, and other issue fields are omitted; this is not an importable tracker or a second canonical export. The original raw B0/B1/B2 exports remain temporary and their whole-file digests cannot be recomputed from the sanitized projection. The projection does allow inspection of the target values, B1/B2 target and graph equality, event timestamp/reason, and the source/candidate null distinction. The exact candidate remains reconstructible from the identified source and B1 values above.

This candidate was computed after the guarded reconciler rejected its
preflight. It is not a reconciler evidence bundle and must not
be described as a guarded promotion. The exact reviewed candidate is now the
canonical row in this PR branch; its one-time promotion is recorded below.
The correction was authorized under the owner’s standing instruction for
factually invalid content after independent verification, scoped to this exact
source digest, target, candidate digest, Beads binary/version, four changed
fields, and close-event provenance. The parent-side independent reviewer
reconstructed this candidate from the published projection and exact main
source, confirmed the candidate digest and all row/field/graph/null invariants,
and accepted it for this one-operation exception before promotion. The exact
source digest was rechecked immediately before replacing the canonical export.
The candidate was promoted to `.beads/issues.jsonl` at
`2026-10-19T03:58:47Z` in the working tree and committed as
`a56340f7c3c88222dbb6791ffa44464887a5d22e`; the resulting file digest is the
independently verified candidate digest above. The exception preserves the unrelated null and
dependency graph, does not write the live Beads database, and adds no
generalized closure code, source wildcard, future source digest, null pair, or
policy change. It does not claim historical guarded provenance or change the
existing guarded workflow. The close event is the isolated Beads event
recorded above; canonical export promotion is this later, exact-source
correction.

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
