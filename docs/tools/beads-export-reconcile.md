# Beads export reconciliation

`beads-export-reconcile` creates a reviewable, immutable evidence bundle for a narrow set of planned issue-field updates. It does not edit the canonical JSONL until the separate `promote` command passes all checks.

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

The command pins Beads 1.3.1, hashes the executable, and records each command's exact arguments, exit status, and stdout/stderr digests. It runs `bd init`, imports the canonical export into a disposable store, captures B0, applies only the plan via `bd update`, then captures B1. A separate disposable store imports the candidate and produces a round-trip export. The candidate is built from canonical source bytes; only planned fields and Beads' changed `updated_at` fields can be spliced into records. Untouched records and fields retain their original bytes.

The bundle ZIP contains `source.jsonl` (C), `baseline.jsonl` (B0), `updated.jsonl` (B1), `plan.json`, `candidate.jsonl`, `round-trip.jsonl`, and `manifest.json`. The manifest binds the immutable bundle identity to the source, machine exports, plan, candidate, round-trip, and every command result. Copy the printed ZIP digest and candidate digest into the review record. Promotion requires the separately recorded ZIP digest, so replacing the bundle even with the same Beads version fails closed.

## Fail-closed checks

Reconciliation rejects duplicate JSON keys, issue IDs, and dependency edges; missing or extra issues; incomplete or changed dependency records; source/import deltas outside the pinned Beads 1.3.1 normalization rules; any post-update change outside the plan and Beads' `updated_at`; stale source bytes; and malformed, missing, mismatched, replaced, or same-version-but-different evidence. It distinguishes explicit `null` from an absent field and preserves the source representation in the candidate. A single deterministic importer timestamp exception is allowed only for `dap-jo9.5.5.closed_at` when source is absent and Beads 1.3.1 produces exactly `max(created_at, updated_at) + 1s`; the derived value is verified but never copied to C.

The current canonical export contains five explicit null values (four `closed_at` fields and one `dependencies` field), which Beads 1.3.1 removes on import. The reconciler therefore rejects the current export before candidate publication. Do not null-normalize or hand-edit the canonical file to bypass this check. The representation policy needs a reviewed resolution that preserves the accepted absent-versus-null distinction before this workflow can promote updates against that export.

## Promotion

`promote` reads the bundle as an untrusted input, rejects symlink and hard-link aliases, checks the bundle and evidence digests, reruns both semantic validations, verifies that the canonical source still matches C, then takes an exclusive lock and rechecks the source immediately before atomic replacement. The lock serializes only writers that honor the same sibling `.beads-reconcile.lock`; it is not a filesystem compare-and-swap. Every process or person that can change the canonical export must acquire and hold that lock, or remain quiescent for the full promotion window. Direct `bd export` and other uncoordinated writes are outside this guarantee; if exclusive writer coordination cannot be ensured, do not promote. An existing lock, stale source, or verification failure leaves the canonical export unchanged. The lock-contention test verifies that a second cooperating publisher fails without changing the source. Review the resulting `.beads/issues.jsonl` diff and keep the evidence ZIP and its printed digest with the change record.
