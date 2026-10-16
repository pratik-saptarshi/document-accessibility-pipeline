# Beads export reconciliation

`beads-export-reconcile` creates a reviewable, immutable evidence bundle for narrowly planned existing-issue updates, one bounded new issue, or the exact scoped closure described below. It does not edit the canonical JSONL until the separate `promote` command passes all checks.

**Current tracker-export status (2026-10-16; main `76f14f69a3283715f5d5278983363121168a2f09`):** the canonical `.beads/issues.jsonl` contains 100 issue rows and records `dap-jo9.18` as closed. PR #97 merged the scoped closure workflow and its reviewed export update. This is an interchange-export status only; the live embedded Dolt database was not changed or verified. The closure evidence and explicit limits are recorded in [Decision 0008](../decisions/0008-scoped-dap-jo9-18-closure.md) and the [candidate review](../reviews/dap-jo9.18-closure-candidate-2026-10-16.md).

## Plan

Write a JSON plan with unique issue IDs and exact desired string values. Supported fields are `title`, `description`, `acceptance_criteria`, and `notes`:

```json
{
  "updates": [
    {
      "id": "dap-jo9.1.1.6",
      "fields": {
        "title": "Select the authoritative role-evidence source"
      }
    }
  ]
}
```

The command pins Beads 1.3.1, hashes the executable, and records each command's exact arguments, exit status, and stdout/stderr digests. It runs `bd init`, imports the canonical export into a disposable store, and captures B0. For existing issues, it applies only planned field updates via `bd update`; a creation plan may import only its one complete planned row into that disposable store; a closure plan may perform only the exact separately authorized close described below. It then captures B1. A separate disposable store imports the candidate and produces a round-trip export. The candidate is built from canonical source bytes; only planned fields and Beads' changed `updated_at` fields can be spliced into existing records, and a new row is appended only if it exactly matches the single planned creation. Untouched records and fields retain their original bytes. The live Beads store is never changed by `create`.

### Scoped closure currently supported

The typed `closes` plan operation supports at most one closure, for `dap-jo9.18` only, and only against source SHA-256 `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465` with the Beads 1.3.1 executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. It requires the expected source status `open`, Decision 0008, the exact accepted reason and evidence references, no mixed update/create operations, and an actual successful `bd close` transition in the disposable store. Beads supplies `closed_at` and `updated_at`; the caller cannot supply timestamps or generic lifecycle fields. The operation is not a general status-edit or bulk-close facility. Do not treat the old source digest as authorization to rerun a close against current bytes.

The bundle ZIP contains `source.jsonl` (C), `baseline.jsonl` (B0), `updated.jsonl` (B1), `plan.json`, `candidate.jsonl`, `round-trip.jsonl`, and `manifest.json`. The manifest binds the immutable bundle identity to the source, machine exports, plan, candidate, round-trip, and every command result. Copy the printed ZIP digest and candidate digest into the review record. Promotion requires the separately recorded ZIP digest, so replacing the bundle even with the same Beads version fails closed.

## Fail-closed checks

Reconciliation rejects duplicate JSON keys, noncanonical case-insensitive schema-key aliases or collisions, null/non-string values for string fields, issue IDs, and dependency edges; missing or extra issues; incomplete or changed dependency records; source/import deltas outside the pinned Beads 1.3.1 normalization rules; any post-update change outside the plan and Beads' `updated_at`; stale source bytes; and malformed, missing, mismatched, replaced, or same-version-but-different evidence. Explicit empty strings remain distinct from JSON null; clearing a description requires an explicit empty string and Beads' `--allow-empty-description` flag. A single deterministic importer timestamp exception is allowed only for `dap-jo9.5.5.closed_at` when source is absent and Beads 1.3.1 produces exactly `max(created_at, updated_at) + 1s`; the derived value is verified but never copied to C.

### Pinned explicit-null compatibility exception

An approved, temporary compatibility exception covers exactly five Beads 1.3.1 import losses for the three exact canonical source snapshots listed below and Beads executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`:

| Issue | Field |
| --- | --- |
| `dap-jo9.5.3` | `dependencies` |
| `dap-jo9.1.6.10` | `closed_at` |
| `dap-jo9.5.6` | `closed_at` |
| `dap-jo9.2.1.4` | `closed_at` |
| `dap-jo9.5.7` | `closed_at` |

The exact authorized source digests are `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`, `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`, and the one further approved continuation `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`. For every pair, C must contain the explicit JSON token `null`; B0 and B1 must both omit the field; and the candidate must retain the original `null` bytes, including when another field on that issue is edited. No field or dependency graph change is covered by this exception. A different source digest, executable digest, Beads version, null pair, changed B0/B1 value, or any other semantic/graph delta fails closed. The second-store check re-derives the candidate from C, B0, B1, and the plan, then records the exact candidate line spans containing `null`; a caller-supplied candidate digest or span cannot authorize the round-trip omission.

The evidence bundle records the source and executable digests, all snapshot digests, each exact issue/field pair and its C/B0/B1/candidate/round-trip presence, plus candidate line and byte spans. Audit by comparing those records to the retained bundle snapshots and confirming that the candidate bytes preserve C's null tokens. Keep the ZIP digest and candidate digest with the reviewed change. Remove this exception only after a separately verified Beads version/binary preserves all five nulls in fresh import/export and the negative tests still reject altered values, graph changes, replacements, and unlisted losses. Any source digest outside the three exact approved snapshots, or any different executable digest or Beads version, fails closed automatically.

**Approved continuations:** source-specific authorization covers the historic snapshot `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`, the continuation `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`, and the exact continuation `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465` approved by repository-owner decision [0007](../decisions/0007-exact-beads-null-loss-continuation-after-jo9-18.md). All three use Beads 1.3.1 executable SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653` and the same five pairs above. The continuation is encoded as an exact source allowlist in the reconciler; no prefix, wildcard, configuration-supplied digest, other binary, or additional pair is accepted. Fresh-source regressions import and export disposable stores and check the candidate and second-store round trip. A bundle from another source remains stale against the current source. **Future updates and promotions** still require their own authorized exact-source plan, behavioral checks, and independent review; the completed `.18` promotion does not authorize another closure or source digest.

The earlier pre-creation source snapshot `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de` contained 99 rows. The historical run record reports a dependency warning for its disposable import/export and says B0 retained the edge and graph comparison passed. The raw warning output and B0 bytes are not retained in the current tree; this records the warning for that pre-creation snapshot only, not an authorized graph exception. The later closure source C, `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`, contained 100 rows, including `.18` as open. The reviewed candidate closed that existing row without changing the count; the current canonical export contains 100 rows.

## Promotion

`promote` reads the bundle as an untrusted input, rejects symlink and hard-link aliases, checks the bundle and evidence digests, reruns both semantic validations, verifies that the canonical source still matches C, then takes an exclusive lock and rechecks the source immediately before atomic replacement. The lock serializes only writers that honor the same sibling `.beads-reconcile.lock`; it is not a filesystem compare-and-swap. Every process or person that can change the canonical export must acquire and hold that lock, or remain quiescent for the full promotion window. Direct `bd export` and other uncoordinated writes are outside this guarantee; if exclusive writer coordination cannot be ensured, do not promote. Platforms where a hard-link count cannot be obtained fail closed. An existing lock, stale source, or verification failure leaves the canonical export unchanged. The lock-contention test verifies that a second cooperating publisher fails without changing the source. Review the resulting `.beads/issues.jsonl` diff and keep the evidence ZIP and its printed digest with the change record.
