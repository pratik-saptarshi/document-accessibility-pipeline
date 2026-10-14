# Backlog Refinement Execution Plan — 2026-10-14

**Status: Proposed execution plan.** Records panel findings, judge dispositions,
and the execution sequence. Does not adopt product decisions; deferred items
remain explicitly open.

## Sources

- Panel report (this run, 2026-10-14): five role agents (refinement, elicitation,
  product owner, roadmapping, prioritization) + judge synthesis, read-only over
  the 41 active issues, validated against `docs/requirements/product-scope.md`,
  `docs/roadmap/product-roadmap.md`, Decisions 0001/0002.
- Prior panel: `docs/planning/backlog-refinement-panel-2026-10-13.md`
  (six-role review; complementary, not superseded).

## Judge dispositions

- **FIX-NOW (22):** 5 reprioritizations, 7 stale-dependency removals
  (`dap-jo9.1.2`, closed after a no-go, no longer a gate), 5 dependency
  additions, 9 type/label corrections, 4 retitles/rewords, 1 rollup-AC rewrite
  (`dap-jo9.2.1`), tooling lane/phase labels, 1 test-case annotation, 1 wording
  fix, 1 goal clarification.
- **DEFER-TO-BACKLOG (12):** 5 new issues + 7 decision/notes recorded on
  existing issues. All need a product-owner decision or larger rewrite before
  work starts.
- **REJECTED (8):** recorded with rationale; no action.

## Backlog health (panel, against the product-value brief)

| Expectation | Verdict |
|---|---|
| Names lane and phase | FAIL — 0/41 name a phase; ~20 name no lane → DEFER batch issue `dap-jo9.5.5` |
| Checkable acceptance criteria | PARTIAL — boilerplate unverifiable clauses in 9 issues → DEFER `dap-jo9.5.4`; bare catalog IDs in 6 stories → DEFER notes |
| No gate-forbidden promises | PASS — product owner found none |
| No duplicate scope | PASS — one overlap already disambiguated in text |
| Live dependency chains | PARTIAL — no cycles/dangling refs; 7 edges point at closed-REJECT `.1.2` → FIX-NOW |
| Priority consistent with Decision 0002 / phase order | PARTIAL — 5 demotions → FIX-NOW; P0s and tip-of-spear P1s correct |

## Execution sequence

1. **FIX-NOW** — apply the 22 edits on branch `docs/backlog-refinement-2026-10-14`
   (JSONL-level edits, verified by diff; no `bd` DB writes).
2. **DEFER** — file 5 new issues (`dap-jo9.1.6.5`, `dap-jo9.2.5.10`, `dap-jo9.4.2`,
   `dap-jo9.5.4`, `dap-jo9.5.5`) and append decision-notes to 7 existing issues.
3. **Verify** — `bd ready` re-run, health re-check, repo checks
   (`gofmt`, `go vet`, `go run ./cmd/traceability-check`).
4. **Autonomous TDD** — proceed down the judge's ranked top-12, starting with the
   first ready, unowned, ungated item; red → green → refactor per issue AC.

## New deferred issues (to file)

| ID | Title | Type | Disposition source |
|---|---|---|---|
| `dap-jo9.1.6.5` | Record disposition contingency for a v2 non-proceed outcome | task | Judge #7 |
| `dap-jo9.2.5.10` | Produce an eligible candidate for the candidate-bound review gate | task | Judge #8 |
| `dap-jo9.4.2` | Decide AFP track go/no-go/unknown with owner and input criteria | decision | Judge #10 |
| `dap-jo9.5.4` | Extract shared engineering definition-of-done to one canonical record | task | Judge #20 |
| `dap-jo9.5.5` | Batch-label lane/phase on all open issues (or record exemption) | task | Judge #27 |

Decision-notes to append: `.1.1.4` (authorized owner, #6), `.2.1.3` (evidence
policy, #9), `.1.1.3` ("becomes material" trigger, #11), `.5` (domain
authority, #12), `.2.5.9` (byte-change carve-out, #14), `.2.3` (story
re-scope, #19), 6 stories (catalog-row linking, #21).
