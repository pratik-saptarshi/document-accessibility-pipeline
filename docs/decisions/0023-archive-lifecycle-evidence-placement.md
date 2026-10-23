# Decision 0023: Archive-lifecycle validation evidence belongs in `docs/research/`

**Status: Accepted.** On 2026-10-22, the product owner accepted the paperthin adversarial panel's process recommendation arising from PR #179's review.

## Context

PR #179 ("close dap-5hy — bootstrap acceptance validated") drew a P1 review thread ("Reopen dap-5hy until the active archive is validated") claiming no record existed of the required lifecycle run against the active Go 1.27.3 archive (fresh install, repeat install, separate `--verify`, integrity/root negative controls). A paperthin adversarial panel refuted the thread: the complete run was recorded in PR #173's merged body (2026-10-21), against the active digest, in the venue the planning doc designated — so the closure stood.

But the dispute exposed a real process weakness: the evidence lived only in a PR body, which is invisible to tree-only audits. The thread's "no record" claim was factually wrong yet *audit-plausible*, and adjudicating it required a panel. Durable evidence must live where audits look.

## Decision

Archive and toolchain lifecycle validation runs — fresh install, repeat install, separate `--verify`, integrity and root negative controls, with exact artifact digests — must be recorded as durable files under `docs/research/`, following the precedent of the 2026-10-19 record. PR bodies may summarize and link to the record, but the record of record is the tree.

## Alternatives considered

- **PR-body-only (status quo ante): rejected.** Invisible to `git grep` and tree audits; forces future reviewers to re-litigate settled validation.
- **Beads notes as the evidence venue: rejected.** Notes are issue metadata, not an evidence venue; they don't carry the run's full detail and aren't the planning doc's designated location.

## Consequences

- Future archive/toolchain lifecycle PRs add (or update) a `docs/research/` record before merge; reviewers check the tree, not just the PR body.
- Existing PR-body-only records (e.g. PR #173's Validation section) stand as-is; backfilling is not required unless an audit disputes them.

## Follow-up

- None blocking. If a future lifecycle run's tree record is challenged, the challenger cites the record's path and section.
