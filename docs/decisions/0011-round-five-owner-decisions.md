# Decision 0011: Bounded round-five owner decisions

**Status: Accepted within the limits below.** R5-02 remains undecided. This
record authorizes bounded issue closure under the existing safe workflow, but
does not approve generalized closure code, a generic tracker transition, or
null-loss policy expansion.

**Date: 2026-10-17.**

## Context

The round-five plan included a proposed reusable Beads transition policy. The
owner later superseded the earlier general “Proceed” direction with scoped
decisions and an explicit freeze on tracker/governance tooling until
`.1.6.3` and `.1.6.4` advance. This record preserves those decisions without
changing canonical tracker data, issue status, dependencies, or the live Beads
database.

## Decisions

1. **Tracker operations (R5-01/02).** Use `bd` for operations, then guarded
   validation and promotion. Never accept an unchecked export. Agents may
   authorize issue closure only when documented acceptance criteria and review
   checks pass. Preserve human, assistive-technology, and product-specific
   gates; do not invent evidence. Freeze tracker/governance tooling changes
   until `.1.6.3` and `.1.6.4` advance; use existing safe capabilities only
   meanwhile. R5-02 remains undecided and does not authorize generalized
   closure code or null-policy expansion. Preserve existing exact-source
   approvals and null behavior.
2. **Publisher identity (R5-03).** The documented historical exception is
   accepted. Going forward, use a synthetic identity or a verified GitHub
   noreply identity. Verify the actual noreply identity before changing
   publisher configuration. Preserve the CloudBot name and strong publisher
   guard. This decision does not authorize history rewriting or redaction of
   retained artifacts.
3. **Evidence record (R5-12/13).** Any one of a PR review, Beads note, or signed
   commit may record owner approval. The record must distinguish authenticated
   owner approval from an agent's assertion. Keep private chat private. A
   digest record plus a CI artifact is sufficient evidence; plan retention,
   retrievability, and hash verification before retiring the current bundle.
   Preserve history and unique evidence; do not delete it prematurely.

   Owner approval for this decision was independently verified and transcribed
   by an agent in the [PR #128 review](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/128#pullrequestreview-5559368515)
   on 2026-10-17.
4. **CI tool licensing (R5-15).** GPL tools are approved for CI-only use. This
   does not extend approval to product/runtime use. Pin the acquired package or
   source and record provenance within the CI scope.
5. **PR merges (R5-23).** Agents may merge PRs when repository conventions and
   required safeguards are met, including required review, hosted checks, and
   protected merge controls. This does not authorize bypassing protections.

## Limits and follow-up

- No direct or unchecked `.beads/issues.jsonl` edit, dependency change, or live
  Beads/Dolt write is authorized here. A bounded issue closure is authorized
  only under the acceptance/review conditions above and through `bd` followed
  by guarded validation and promotion.
- No generic reusable transition validator, null-loss exception, or generalized
  closure code is approved. R5-02 remains pending; closure under R5-01 is
  authorized only within the bounded criteria and existing safe workflow.
- Product evidence and prerequisites for `.1.6.3` and `.1.6.4` may be prepared
  during the tooling freeze. Their historical closures require verification of
  the exact v2 source/evidence pins, current written criteria, durable
  evidence, and review checks, followed by `bd` and guarded
  validation/promotion. The separate v3 source choice does not gate those
  historical closures; resolve it before v3 authoring. If an issue closure is
  proposed, preserve all human/AT/product-specific gates and do not invent
  evidence.
- Preserve existing CloudBot publisher controls until the actual verified
  noreply identity is known. Never print an identity value as part of
  verification.
- Historical archive bytes remain evidence. Any future replacement is a new
  artifact with provenance; do not silently rewrite old history.

## Alternatives considered

- Treat the earlier generic “Proceed” as permanent authorization for reusable
  transition/null-loss/closure tooling: rejected by the owner's later scoped
  direction. The later bounded issue-closure delegation remains in force.
- Treat a PR or commit as sufficient without authenticating the owner: rejected;
  the record must distinguish owner approval from agent assertion.
- Replace or delete historical artifacts to remove old identity/path content:
  not authorized; preserve unique evidence and plan retrieval before retirement.
- Extend CI-only GPL approval to runtime or product use: rejected.

## Follow-up

Prepare evidence for `.1.6.3` and `.1.6.4` from the pinned historical v2
operation; these closures do not require v3 qualification. For the distinct v3
baseline, the recommended source is byte-for-byte copies of the pinned repaired
v2 candidates, followed by an owner-approved object/value allow-list and fresh
output hashes. The [v3 proposal](../fixtures/remediation/financial-table-experiment.v3-plan.md)
lists that option and alternatives. Resolve the exact source choice before v3
authoring. Do not resume tracker/governance tooling changes until both issues
advance or a new owner direction explicitly changes the freeze. Existing safe
`bd` plus guarded validate/promote operations remain available for authorized
bounded closures.
