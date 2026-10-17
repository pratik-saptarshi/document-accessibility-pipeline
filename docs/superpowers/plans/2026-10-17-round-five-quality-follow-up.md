# Round-five quality follow-up

**Status (2026-10-17):** Plan prepared on fresh `origin/main` at
`4585a23457ac2f7a16650de5fc0358b2f900327f`. The owner authorized the
six-phase review plan, then superseded the earlier generic tracker-policy
direction with the bounded decisions recorded in [Decision 0011](../../decisions/0011-round-five-owner-decisions.md).
The R5-02 policy question remains undecided. Freeze tracker/governance tooling
changes until `.1.6.3` and `.1.6.4` advance; do product-evidence and
prerequisite work in the meantime. Closures are authorized only under the
bounded R5-01 criteria and existing `bd` plus guarded validate/promote path;
unchecked or direct canonical export changes are not authorized. Agents may
merge PRs when repository conventions and safeguards are satisfied.

## Goal and boundaries

Resolve the round-five findings that remain valid on the stated base, preserve
audit history and active work, and avoid treating repeated panel observations
as independent defects. Use test-first changes, exact-base/exact-head review,
and hosted checks. Do not edit Musebot branches or source/branches for PRs
#115–#118 or #122. Coordinate with #125 before touching its DevSecOps policy
tests. Do not rewrite history, alter credentials or broader security policy, or
run a live Beads database repair. Publisher identity changes are limited to
the R5-03 path and require identity verification first.

Use `bd` for tracker operations followed by guarded validation and promotion;
never accept an unchecked export. Agents may authorize issue closure when
documented acceptance criteria and review checks pass, while preserving all
human, assistive-technology, and product-specific gates and without inventing
evidence. The freeze is on tracker/governance tooling changes until `.1.6.3`
and `.1.6.4` advance; use existing safe capabilities only in the meantime.
R5-02 remains undecided and does not authorize generalized closure code or
null-policy expansion. Preserve existing source-specific approvals and null
distinctions.

The owner accepts the documented historical publisher-identity exception.
Going forward, use a synthetic identity or a verified GitHub noreply identity;
verify the actual noreply identity before changing publisher configuration.
Preserve the CloudBot name and strong publisher guard. The owner approved GPL
tools for CI-only use; pin and record provenance under that scope. A digest
record plus a CI artifact is sufficient evidence, provided retention,
retrievability, and hash verification are planned before retiring the current
bundle. Do not rewrite history or delete unique evidence prematurely.

Owner approval may be recorded in any one of a PR review, a Beads note, or a
signed commit. Records must distinguish authenticated owner approval from an
agent's assertion; private chat remains private. Agents may merge PRs when
repository conventions and safeguards are met, including protected merge
requirements. The bounded decisions and limits are in Decision 0011.

## Finding validation and disposition

This matrix is anchored to the exact base above. “Verified” means supported by
the current source or a reproduced failure; “partial” means a narrower gap
survives while part of the panel claim is stale, covered, or policy-only.

| ID | Disposition | Evidence and planned response |
|---|---|---|
| R5-01 | Verified gap; bounded operation model decided | A direct `.beads/issues.jsonl` edit can bypass guarded reconciliation. PR #114 landed the closure, but that fact does not establish that it was unauthorized; do not characterize it that way. Use `bd`, then guarded validation/promotion; never accept unchecked exports. Agents may authorize closure when documented acceptance criteria and review checks pass, preserving human/AT/product-specific gates and using no invented evidence. Tooling changes remain frozen until `.1.6.3` and `.1.6.4` advance. |
| R5-02 | Verified policy gap; owner decision pending | Current reconciliation is pinned to source hashes, null pairs, binary/version, and exact operations. Do not replace these with a generic null-loss or closure policy while R5-02 is undecided. Keep tracker/governance tooling frozen until `.1.6.3` and `.1.6.4` advance. |
| R5-03 | Verified propagation; exact totals unresolved | Sensitive email-like data appears in multiple tracked text files and retained archives. Independent scans disagree on target/scope totals. Reconcile with a bounded in-memory scanner that emits only file/member counts and digests, never the matched value. The documented historical exception is accepted. Going forward use a synthetic or verified GitHub noreply identity; verify the actual noreply value before changing publisher settings. Preserve CloudBot name and strong guard. Preserve old ZIPs/archives; no retroactive redaction. |
| R5-04 | Partly valid; thresholds partly covered | Repository guidance already caps review iterations and stops after repeated unsuccessful attempts. Apply accepted limits consistently; do not invent numerical performance thresholds or override owner decisions. |
| R5-05 | Partial | Closure timestamps are UTC and update fields are written, but monotonicity within a command window is not explicitly asserted; environment/version checks also have narrow fail-closed weaknesses. Add only tests for demonstrated boundary failures; do not claim timestamp forgery. |
| R5-06 | Mostly covered; test-strength opportunity | Production creation validates fields, parent, priority, count, and content. Independently exercise isolated predicates and import content-hash behavior; retain existing final B1 equivalence. Do not claim a production bypass without a failing test/mutant. |
| R5-07 | Narrow test gap | XML duplicate attributes/namespaces are rejected; add explicit directive/DOCTYPE rejection coverage if the parser contract requires it. Keep tests synthetic and independent of real financial fixtures. |
| R5-08 | Partial | Version source is digest-bound and version presence is checked, but module-version matching uses a substring test. Replace with exact module/version parsing and boundary tests; retain valid source-pinning behavior. |
| R5-09 | Verified stale executable fallback | Existing archive integrity controls (bounds, digest, safe paths, recursive protection) do not establish source freshness. `bootstrap-cloud-build-tools.sh` extracts and executes the pinned archive's `cmd/cloud-tools-bootstrap` when the current download path falls back; that snapshot predates PR #109 and lacks its manifest key validation (118 added lines). Add a regression that proves fallback source has the required validation or fails closed. Preserve all historical archive bytes; any refreshed archive is a new version with provenance. |
| R5-10 | Verified CI gap; fuzz absence is already dispositioned | Traceability and real-Beads coverage are not enforced in current CI; real-Beads tests may skip without the binary. Add deterministic traceability and a pinned real-Beads CI check if feasible. Bounded fuzz remains decision-only/P3 and is not a new blocker. |
| R5-11 | Verified fail-open bug; coordinate first | `go list` piped to `grep` can hide producer failure without `pipefail`; a synthetic failure reproduced fail-open behavior. Fix the check to fail closed and test both valid/missing-dependency and producer-error cases. PR #125 covers adjacent parser/policy tests but does not change this command; coordinate and avoid editing its branch. |
| R5-12 | Approval record format decided | Owner approval may be recorded in any one of a PR review, Beads note, or signed commit. Clearly distinguish authenticated owner approval from agent self-assertion; keep private chat private. Do not change canonical records while governance tooling is frozen. |
| R5-13 | Claim correction, preserve evidence | A retained ZIP contains `/tmp` paths; the existing scrub claim applies to new prose/tests, not every immutable archive byte. A digest record plus CI artifact is sufficient. Plan retention, retrievability, and hash verification before retiring the current bundle; preserve history and unique evidence. |
| R5-14 | Verified stale count | Current tracker has 101 nonblank issue rows; the export guide still describes 100 as current. Correct current counts and clearly label historical counts/snapshots. |
| R5-15 | Verified reproducibility gap; CI-only GPL use approved | CI installs Poppler without a version pin; metadata indicates 25.03.0, while an older comment says 24.02.0 and does not establish the current CI version. GPL tools are approved for CI-only use. Pin the reproducible package/source and record provenance within that scope. |
| R5-16 | Partial; preserve schema-specific behavior | Shared `jsonstrict` exists, while Beads and reconciliation code have separate schema-aware duplicate-key scanners and intentionally different unknown-field rules. Characterize schemas first; consolidate only where tests prove identical semantics. Do not blanket-unify loaders. |
| R5-17 | Duplicate work, no new issue | Process-group cleanup is duplicated across financial fixture, veraPDF, and adapter code. Issue `.5.7` says not to consolidate helpers solely to reduce duplication; check current ownership before related changes. Do not open duplicate work or consolidate for deduplication alone. |
| R5-18 | Valid operational concern; manual recovery remains | Bootstrap may leave stale locks for inspection; promotion rejects an existing lock, and HTML rendering is not used by current Go commands. Keep recovery manual; no auto-delete or stale-lock takeover without a separately accepted contract. Separately test whether `runCommandWithTimeout` must honor caller cancellation; it currently derives timeout from `context.Background()`. |
| R5-19 | Partial | `LoadCatalog` lacks an aggregate read cap and traversal order can be nondeterministic; `LoadIssues` already applies a line cap and projection-key scanner. Dependency cycle detection exists, while parent/child graph validation is a separate contract question. Add bounded reads/deterministic traversal and tests; do not repeat the stale plain-unmarshal claim. |
| R5-20 | Partial harness gap | Publisher tests isolate global Git config and synthetic traceability tests exist, but config-injection coverage and a JSONL newline policy are absent. Add injection tests only if the guarded wrapper promises to defend against those variables; define newline handling without modifying identities or live fixtures. |
| R5-21 | Verified workflow hardening gaps | Dependabot covers root Go/Actions but not nested modules; curl follows redirects without an HTTPS-only policy and CodeQL timeout is not explicit. Add nested-module coverage, safe redirect/protocol restrictions, and bounded timeouts. The managed CodeQL default workflow is active; YAML-only edits do not change that external setting. |
| R5-22 | Partial; several claims stale | Generic JSON errors, `%w` wrapping in `LoadCatalog`, and output-parent symlink rejection already exist. Verify races before expanding claims; test parent-directory confinement and whether `readRegularBounded`/`rejectSymlinkPath` diagnostics expose absolute paths. Document seven packages lacking API docs if still true. `author_v2` diagnostics are not a security bypass. |
| R5-23 | Mostly stale; merge authority clarified | `AGENTS.md` already defines guidance precedence and records skill provenance. Agents may merge PRs when repository conventions and required safeguards are met. No blanket guidance rewrite; update only a specific contradiction if one remains. |

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

### Phase 1 — bounded tracker decisions and prerequisite evidence (R5-01/02/12)

1. Record the bounded owner decisions in Decision 0011. R5-02 remains undecided;
   this record does not activate a generic transition or closure policy.
2. Keep tracker/governance tooling changes frozen until `.1.6.3` and `.1.6.4`
   advance. Existing safe capabilities may be used: make intended operations
   with `bd`, then validate and promote through the guarded path. Never accept
   an unchecked export. Do not change dependency edges or write directly to
   the live Beads/Dolt store.
3. Prepare the product evidence and prerequisites for `.1.6.3` and `.1.6.4`.
   Before advancing either issue, resolve the exact baseline/source choice and
   use only already authorized tooling. Do not synthesize product acceptance.
4. Preserve existing exact-source guards and decisions. No generic null-loss
   behavior or generalized closure code may replace them absent a new explicit
   owner decision and a separate reviewed implementation plan. Closure may be
   authorized only when documented acceptance criteria and review checks pass;
   preserve human/AT/product-specific gates and do not invent evidence.
5. Approval may be represented by one PR review, Beads note, or signed commit;
   verify owner authentication and do not reproduce private chat.

### Phase 2 — privacy and immutable provenance (R5-03/09/13/15)

1. Finish no-print scope/count reconciliation; do not include matched values in
   logs, plans, PR bodies, CI output, or comments.
2. Use a synthetic or verified GitHub noreply identity going forward. Verify
   the actual noreply identity before changing publisher configuration;
   preserve CloudBot name and the strong guard. The documented historical
   exception is accepted; do not rewrite history.
3. Test the bootstrap fallback against the current `cmd/cloud-tools-bootstrap`
   manifest-validation contract. It must fail closed if the selected archived
   source is stale. Before publishing a refreshed archive, verify the selected
   publisher identity and record the new artifact's source/version provenance.
4. If the executable fallback is refreshed, publish it as a new version with
   source/version provenance and a reviewed minimal allowlist. Keep all older
   ZIP/archive bytes and digests unchanged. Use synthetic values in tests;
   treat source freshness and identity minimization as separate requirements.
5. Pin Poppler acquisition/version and record source/version/digest. GPL use is
   approved for CI-only; do not extend that approval beyond CI.
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
readiness decision to the owner. Agents may mark PRs ready and merge them when
repository conventions and required safeguards are met. After a protected
merge, verify the resulting main head and only then begin the next increment.

## Explicitly preserved facts

- Prior archive/ZIP bytes, canonical tracker history, issue statuses, and
  dependency edges remain intact. Existing exact-source decisions apply only
  within their written scope; no generic transition guard is approved.
- `R5-10` bounded fuzz absence remains a prior decision-only/P3 disposition.
- `R5-13` does not authorize deleting historical evidence.
- `R5-17` does not justify helper consolidation solely for deduplication or
  duplicate `.5.7` work.
- `R5-18` does not authorize automatic stale-lock cleanup.
- `R5-23` does not require a broad AGENTS/skill rewrite.
- No email-like value, credential, or personal identity string is reproduced
  in this plan.
