# PR #161: proposed one-time Beads closure exception

**Status: Applied once under the exact scoped exception below. The proposal snapshot committed at `c58b3ff` reports a pre-promotion audit of temporary source/B0/B1/B2/candidate artifacts, while still marking independent exception review and authorization as pending. Separately, task history reports that an independent reviewer checked the candidate at 03:57Z using the public Markdown proposal at `c58b3ff` plus immutable main, accepted the one-operation exception, and did not inspect temporary B0/B1/B2 exports. That review report is not a timestamped repository artifact, so the repository cannot establish that the reported event or acceptance occurred. The published reconstruction projection was created after promotion and does not substantiate either earlier check.**

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
JSONL timestamps have second precision. The exact shell-history transcript is
not present in retained repository evidence; whether another copy exists
outside the repository is not established. Durable provenance here is the
pinned binary identity, plan digest, command sequence, source/B0/B1/B2 digests,
and read-only Beads history record.

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

The proposal snapshot committed in `c58b3ff` at `2026-10-19T03:56:07Z` reports
an independent read-only audit of temporary source/B0/B1/B2 and candidate
artifacts before promotion. That audit reported checking source and snapshot
digests, 101 rows, 168 edges, exact B1/B2 round-trip, raw-byte preservation of
all 100 non-target lines, the four-field target delta, edge equality, and the
unrelated explicit null. At the c58 snapshot, independent exception review and
authorization were still described as pending.

Separately, the independent reviewer reports a pre-promotion check at 03:57Z
using the public Markdown proposal at exact commit
`c58b3ff9757e82bd7a24592b12dcc7674aed6014` and immutable main source
`d2779c1979149807de8c207acdd718db2a1eb8c5`. The reviewer reports reconstructing
the candidate from those inputs and accepting the one-operation exception;
they did not inspect the temporary B0/B1/B2 exports. This check therefore
supports the proposed candidate projection and bounded target delta, not the
temporary Beads snapshot digests or round-trip. The report of this check is in
task history, not a timestamped repository artifact; the exact transcript is
not retained. The c58 Markdown proposal and immutable main are identified by
commit, but the repository cannot independently establish the reported 03:57
review event or its acceptance.

A sanitized, durable evidence projection is retained at
[`evidence/pr161-beads-roundtrip-2026-10-19.json`](evidence/pr161-beads-roundtrip-2026-10-19.json). It includes the source/B0/B1/B2 original digests and counts, exact target-field projections, a sanitized Beads history event, and the full dependency-edge inventory. Its SHA-256 is
`16ed63a2264206b74ac2a3d5eaceeb858d1ae19eef316da0981d74c610caf82b`. It retains every issue ID, status, updated/closed timestamp when present, dependency edge, and the target close reason; descriptions, acceptance text, notes, labels, assignees, creator data, and other issue fields are omitted. It is a non-importable audit projection, not a second canonical export. It includes exact reconstruction commands and a minimal one-row B0-to-B1 field delta. I reran the pinned Beads import/export from the exact source: B0 reproduced byte-for-byte at its recorded digest; applying the target delta reproduced B1 byte-for-byte at its recorded digest; importing that reconstructed B1 into a fresh store and exporting B2 reproduced the same digest byte-for-byte. The sanitized event history retains only the event hash, committer, time, issue ID, status, timestamps, and close reason; the artifact lists omitted fields. The exact candidate remains reconstructible from the identified source and B1 values above.

The projection's
`created_utc` is `2026-10-19T04:02:00Z`, after the recorded working-tree
promotion at `2026-10-19T03:58:47Z`. The projection was first committed later,
in `13926c63b46c16183c3447b250c221d9382e674a` at `04:03:08Z`. It was not the
artifact used for the pre-promotion review. Its independent immutable-source
replay is post-promotion reproducibility evidence only; it cannot establish
the timing or inputs of the earlier review.

This candidate was computed after the guarded reconciler rejected its
preflight. It is not a reconciler evidence bundle and must not
be described as a guarded promotion. The exact reviewed candidate is now the
canonical row in this PR branch; its one-time promotion is recorded below.
The correction was authorized under the owner’s standing instruction for
factually invalid content after independent verification, scoped to this exact
source digest, target, candidate digest, Beads binary/version, four changed
fields, and close-event provenance. The reported 03:57 pre-promotion review
used the c58 Markdown proposal and immutable main to reconstruct the candidate;
it did not inspect temporary B0/B1/B2 exports or the later published
projection. The separate c58 audit report covers temporary
source/B0/B1/B2/candidate artifacts but still marks exception review and
authorization as pending. Their scopes and limits are stated above. A separate
independent replay from the published projection and
exact main source later confirmed the candidate digest and row/field/graph/null
invariants, but that post-promotion result is not evidence of pre-promotion
review. The exact source digest was rechecked immediately before replacing the
canonical export.
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
