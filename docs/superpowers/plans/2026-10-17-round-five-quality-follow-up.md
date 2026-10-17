# Round-five quality follow-up

**Status (2026-10-17):** Plan prepared on fresh `origin/main` at
`4585a23457ac2f7a16650de5fc0358b2f900327f`. The owner has authorized the
six-phase plan and adoption of a reusable, approval/evidence-bounded tracker
policy. This plan records that authorization; the policy must be made durable
and independently reviewed before it is used for tracker operations. This is
not authorization to perform unresolved privacy, archive-retention, or license
decisions listed below. Protected merges remain with the parent.

## Goal and boundaries

Resolve the round-five findings that remain valid on the stated base, preserve
audit history and active work, and avoid treating repeated panel observations
as independent defects. Use test-first changes, exact-base/exact-head review,
and hosted checks. Do not edit Musebot branches or source/branches for PRs
#115–#118 or #122. Coordinate with #125 before touching its DevSecOps policy
tests. Do not rewrite history, alter credentials/identity/security policy, or
run a live Beads database repair.

The owner approved a generic guarded tracker update/closure policy after the
six-phase proposal. Its durable record must require a compact, validated
transition record with exact source and candidate digests, schema/policy
version, operation and changed fields, pinned Beads version and binary digest,
and evidence/decision reference for changes requiring acceptance. The guarded
path and CI must fail closed on ambiguous base/head selection, malformed or
extra fields, unapproved fields/statuses/edges, source drift, and invalid
evidence. Null preservation remains field- and version-specific; no general
`null`/empty equivalence is allowed. A record is an audit trail, not proof of
approval by itself. Existing ADRs remain immutable historical records. The
new policy may replace per-operation ADR constants only after its durable
record, tests, independent review, and CI enforcement land; it cannot silently
retroactively authorize a past change.

Open owner choices remain open: whether and how to stop copying the publisher
email-like identity into future artifacts; whether any historical archive
content may be redacted or republished; and whether the CI-only Poppler package
has approved license/provenance. Do not infer an answer from “Proceed.” Retain
existing durable evidence and current publisher identity controls until those
choices are explicitly resolved.

## Finding validation and disposition

This matrix is anchored to the exact base above. “Verified” means supported by
the current source or a reproduced failure; “partial” means a narrower gap
survives while part of the panel claim is stale, covered, or policy-only.

| ID | Disposition | Evidence and planned response |
|---|---|---|
| R5-01 | Verified gap; policy authorized | A direct `.beads/issues.jsonl` edit can bypass the guarded reconciliation path. PR #114 landed the closure, but that fact does not establish that it was unauthorized; do not characterize it that way. Add reusable guarded transition validation and a CI check that rejects unguarded tracker deltas. |
| R5-02 | Verified policy gap; policy authorized | Current reconciliation is pinned to source hashes, null pairs, binary/version, and a single closure. Replace the repeated per-source constants with the approved durable policy, exact transition records, and fail-closed validation. Preserve ADRs as audit history and require evidence for accepted closures. |
| R5-03 | Verified propagation; exact totals unresolved | Sensitive email-like data appears in multiple tracked text files and retained archives. Independent scans disagree on target/scope totals. Reconcile with a bounded in-memory scanner that emits only file/member counts and digests, never the matched value. Stop future propagation only after owner chooses the publisher identity treatment. Preserve old ZIPs/archives; no retroactive redaction. |
| R5-04 | Partly valid; thresholds partly covered | Repository guidance already caps review iterations and stops after repeated unsuccessful attempts. Apply accepted limits consistently; do not invent numerical performance thresholds or override owner decisions. |
| R5-05 | Partial | Closure timestamps are UTC and update fields are written, but monotonicity within a command window is not explicitly asserted; environment/version checks also have narrow fail-closed weaknesses. Add only tests for demonstrated boundary failures; do not claim timestamp forgery. |
| R5-06 | Mostly covered; test-strength opportunity | Production creation validates fields, parent, priority, count, and content. Independently exercise isolated predicates and import content-hash behavior; retain existing final B1 equivalence. Do not claim a production bypass without a failing test/mutant. |
| R5-07 | Narrow test gap | XML duplicate attributes/namespaces are rejected; add explicit directive/DOCTYPE rejection coverage if the parser contract requires it. Keep tests synthetic and independent of real financial fixtures. |
| R5-08 | Partial | Version source is digest-bound and version presence is checked, but module-version matching uses a substring test. Replace with exact module/version parsing and boundary tests; retain valid source-pinning behavior. |
| R5-09 | Verified stale executable fallback | Existing archive integrity controls (bounds, digest, safe paths, recursive protection) do not establish source freshness. `bootstrap-cloud-build-tools.sh` extracts and executes the pinned archive's `cmd/cloud-tools-bootstrap` when the current download path falls back; that snapshot predates PR #109 and lacks its manifest key validation (118 added lines). Add a regression that proves fallback source has the required validation or fails closed. Preserve all historical archive bytes; any refreshed archive is a new version with provenance. |
| R5-10 | Verified CI gap; fuzz absence is already dispositioned | Traceability and real-Beads coverage are not enforced in current CI; real-Beads tests may skip without the binary. Add deterministic traceability and a pinned real-Beads CI check if feasible. Bounded fuzz remains decision-only/P3 and is not a new blocker. |
| R5-11 | Verified fail-open bug; coordinate first | `go list` piped to `grep` can hide producer failure without `pipefail`; a synthetic failure reproduced fail-open behavior. Fix the check to fail closed and test both valid/missing-dependency and producer-error cases. PR #125 covers adjacent parser/policy tests but does not change this command; coordinate and avoid editing its branch. |
| R5-12 | Documentation opportunity | ADRs 0006–0010 are indexed and code/tools reference decisions. Add privacy-safe ADR links/metadata to canonical tracker records only where the accepted policy schema supports it; do not invent missing ADRs or claim absent authorization. |
| R5-13 | Claim correction, preserve evidence | A retained ZIP contains `/tmp` paths; the existing scrub claim applies to new prose/tests, not every immutable archive byte. Clarify this boundary and any retention guarantees. Do not delete or expire historical evidence. |
| R5-14 | Verified stale count | Current tracker has 101 nonblank issue rows; the export guide still describes 100 as current. Correct current counts and clearly label historical counts/snapshots. |
| R5-15 | Verified reproducibility gap; license choice pending | CI installs Poppler without a version pin; metadata indicates 25.03.0, while an older comment says 24.02.0 and does not establish the current CI version. Pin a reproducible package/source and record provenance. Do not assert license approval or add a license allowlist until owner resolves it. |
| R5-16 | Partial; preserve schema-specific behavior | Shared `jsonstrict` exists, while Beads and reconciliation code have separate schema-aware duplicate-key scanners and intentionally different unknown-field rules. Characterize schemas first; consolidate only where tests prove identical semantics. Do not blanket-unify loaders. |
| R5-17 | Duplicate work, no new issue | Process-group cleanup is duplicated across financial fixture, veraPDF, and adapter code. Issue `.5.7` says not to consolidate helpers solely to reduce duplication; check current ownership before related changes. Do not open duplicate work or consolidate for deduplication alone. |
| R5-18 | Valid operational concern; manual recovery remains | Bootstrap may leave stale locks for inspection; promotion rejects an existing lock, and HTML rendering is not used by current Go commands. Keep recovery manual; no auto-delete or stale-lock takeover without a separately accepted contract. Separately test whether `runCommandWithTimeout` must honor caller cancellation; it currently derives timeout from `context.Background()`. |
| R5-19 | Partial | `LoadCatalog` lacks an aggregate read cap and traversal order can be nondeterministic; `LoadIssues` already applies a line cap and projection-key scanner. Dependency cycle detection exists, while parent/child graph validation is a separate contract question. Add bounded reads/deterministic traversal and tests; do not repeat the stale plain-unmarshal claim. |
| R5-20 | Partial harness gap | Publisher tests isolate global Git config and synthetic traceability tests exist, but config-injection coverage and a JSONL newline policy are absent. Add injection tests only if the guarded wrapper promises to defend against those variables; define newline handling without modifying identities or live fixtures. |
| R5-21 | Verified workflow hardening gaps | Dependabot covers root Go/Actions but not nested modules; curl follows redirects without an HTTPS-only policy and CodeQL timeout is not explicit. Add nested-module coverage, safe redirect/protocol restrictions, and bounded timeouts. The managed CodeQL default workflow is active; YAML-only edits do not change that external setting. |
| R5-22 | Partial; several claims stale | Generic JSON errors, `%w` wrapping in `LoadCatalog`, and output-parent symlink rejection already exist. Verify races before expanding claims; test parent-directory confinement and whether `readRegularBounded`/`rejectSymlinkPath` diagnostics expose absolute paths. Document seven packages lacking API docs if still true. `author_v2` diagnostics are not a security bypass. |
| R5-23 | Mostly stale | `AGENTS.md` already defines precedence for system/user/repository guidance over imported optional skills and records skill provenance. No blanket rewrite; update only a specific contradictory instruction if one remains on the selected base. |

Finding counts are not used as an acceptance metric. Work is prioritized by
security/correctness impact, demonstrable failure, decision status, and
ownership—not by the panel's remaining-item estimate.

## Execution phases

### Phase 0 — preserve and baseline

1. Use fresh `origin/main` and capture exact base/head for each increment.
2. Record current worktree, fixture, canonical JSONL, archive, and provenance
   digests without printing sensitive values. Do not normalize/import/export
   the live embedded store.
3. Reconcile privacy-scan totals with a no-print scanner: list only path/member,
   count, and digest; do not print candidate strings. Retain current artifacts.
4. Re-check PR #115–#118, #122, and #125 ownership before each overlapping
   change. Never alter those branches.

### Phase 1 — tracker transition policy (R5-01/02/12)

1. Draft a durable policy decision/record on this branch that captures the
   owner's approved replacement model and preserves earlier ADRs unchanged.
2. Define a small versioned transition-record schema: exact base and candidate
   JSONL digests; operation; issue IDs and fields; old/new values; Beads
   version/binary digest; evidence and decision reference; provenance and
   validation result. Avoid storing unnecessary personal or raw fixture data.
3. Add unit tests first for allowed bounded updates, evidence-backed closure,
   null preservation, and rejection of source drift, unapproved status/edge/
   field changes, ambiguous source selection, duplicate IDs, malformed records,
   and unguarded direct tracker edits. Demonstrate each rejection by mutation
   or a red test before implementing its guard.
4. Enforce the same checker in CI on the exact base-to-head tracker diff. Ensure
   it cannot be bypassed by moving the tracker file, changing the claimed base,
   omitting the record, or selecting a different candidate. Do not change
   canonical tracker contents as part of the policy implementation.
5. Independently review the policy and threat model before using it for any
   future canonical update. Parent retains protected merge and tracker
   promotion responsibility.

### Phase 2 — privacy and immutable provenance (R5-03/09/13/15)

1. Finish no-print scope/count reconciliation; do not include matched values in
   logs, plans, PR bodies, CI output, or comments.
2. Ask the owner to choose future publisher identity handling and any license
   allowlist. Until resolved, preserve current publisher identity and do not
   make a license approval claim.
3. Test the bootstrap fallback against the current `cmd/cloud-tools-bootstrap`
   manifest-validation contract. It must fail closed if the selected archived
   source is stale. Do not republish an archive while the identity-treatment
   choice is pending, since repackaging could propagate the unresolved value.
4. After the owner resolves identity treatment, refresh the executable fallback
   as a new version through the guarded publisher, with source/version
   provenance and a reviewed minimal allowlist. Keep all older ZIP/archive
   bytes and digests unchanged. Use synthetic values in tests; treat source
   freshness and identity minimization as separately tested requirements.
5. Pin Poppler acquisition/version and record source/version/digest and license
   evidence only after its CI-only license/provenance choice is resolved.
6. Keep the current scrub statement explicitly scoped to newly authored prose
   and tests, with a separate statement about retained archives and paths.

### Phase 3 — test-first reliability and parsing (R5-05/06/07/08/16/18/19/20/22)

1. Add narrow regressions for observed defects only: timestamp boundary,
   version exactness, directive/DOCTYPE policy, isolated creation predicates,
   bounded aggregate catalog reads, deterministic traversal, Git config
   injection where promised, parent-directory confinement, caller cancellation
   for command execution, and absolute-path disclosure in diagnostics.
2. Keep schema policies explicit. Reuse `jsonstrict` only where exact schema
   behavior matches; retain future-field tolerance where intentional.
3. Preserve manual stale-lock recovery and do not broaden HTML or filesystem
   semantics without an accepted contract.
4. For each fix, use TDD red/green, then mutation/reversal to prove the test
   detects the regression. Avoid tests against real finance PDFs or canonical
   tracker data.

### Phase 4 — CI and documentation (R5-04/10/11/14/17/21/23)

1. Coordinate with PR #125, then add fail-closed `go list` handling and a
   synthetic producer-failure test without editing its branch.
2. Add traceability and pinned real-Beads CI coverage where deterministic and
   supported; do not turn decision-only fuzz work into a gate.
3. Add nested Go module dependency/update coverage; harden curl protocol,
   redirects, and workflow timeouts. Confirm managed CodeQL settings through
   their actual control plane rather than assuming workflow YAML changes it.
4. Correct current issue and source counts while labeling historical snapshots.
   Check `.5.7` ownership before any related changes; do not consolidate
   helpers solely to reduce duplication.
5. Make only evidenced guidance corrections; avoid blanket skill/AGENTS
   rewrites.

### Phase 5 — financial baseline and finalization

Keep financial fixture work synthetic and fixture hashes immutable. Existing
decisions prioritize statement/invoice field cases and operator confirmation;
do not add new numeric thresholds or treat fixture parsing as product
acceptance. Preserve known-good outputs, assert no mixed publication on
failure, and distinguish best-effort rollback from filesystem-wide atomicity.

For every implementation increment: run focused red/green tests, relevant
mutation/reversal checks, root and nested Go race/vet/tidy gates, traceability,
shell and publisher checks, hygiene and diff checks; obtain independent
exact-base/exact-head review; publish only through the guarded HTTPS
source-publisher; wait for all exact-head hosted checks; then return the
readiness decision to the parent. Do not mark PRs ready or merge them. After a
parent-controlled protected merge, verify the resulting main head and only
then begin the next increment.

## Explicitly preserved facts

- Prior archive/ZIP bytes, canonical tracker history, issue statuses, and
  dependency edges remain intact unless a future transition passes the newly
  approved guard and separate evidence requirements.
- `R5-10` bounded fuzz absence remains a prior decision-only/P3 disposition.
- `R5-13` does not authorize deleting historical evidence.
- `R5-17` does not justify helper consolidation solely for deduplication or
  duplicate `.5.7` work.
- `R5-18` does not authorize automatic stale-lock cleanup.
- `R5-23` does not require a broad AGENTS/skill rewrite.
- No email-like value, credential, or personal identity string is reproduced
  in this plan.
