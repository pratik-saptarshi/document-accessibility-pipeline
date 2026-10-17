# DRAFT Decision 0011: Eligibility standard for the candidate-bound review gate — open product-owner decision

**Status: DRAFT — NOT accepted. NO ELIGIBILITY STANDARD IS RECORDED.**

**Date: 2026-10-17.**

> **Decision pending owner acceptance.** This record is a draft for product-owner
> review only. It records an open blocker; it selects nothing. Until the
> product owner explicitly records the eligibility standard for the
> candidate-bound review gate (`dap-jo9.2.5.9`), no eligible candidate can be
> produced, `dap-jo9.2.5.10` stays open and decision-blocked, and the aggregate
> review gate fails closed. Nothing here invents a candidate, infers
> eligibility from feasibility evidence, or presents any selection as approved.

## The missing decision

`dap-jo9.2.5.9` closure requires an eligible candidate with hash-bound
preservation/ua2 evidence. `dap-jo9.1.6.4` is explicitly "not eligible-candidate
evidence", and no open issue produces an eligible candidate. The 2026-10-14
panel finding (validity high) recorded in `dap-jo9.2.5.10` states: **the
eligibility standard itself is an open product-owner decision** — what makes a
candidate "eligible", what evidence the standard demands, and what acceptance
semantics the gate honors. A 2026-10-17 verification searched
`docs/decisions/` (accepted records 0001–0009), `docs/requirements/`, the
review-gate docs, and `dap-jo9.2.5.9`/`dap-jo9.2.5.9.1` issue text: no accepted
or proposed record defines the eligibility standard. The existing mentions
(e.g. `docs/planning/backlog-panel-2026-10-14.md`,
`docs/planning/v2-non-proceed-contingency.md`, `docs/roadmap/product-roadmap.md`)
only restate this acceptance criterion; they do not record the standard.

## What unblocks this

A product-owner **accepted** decision record that names, testably:

1. the eligibility standard — the definition of an "eligible candidate" for
   the candidate-bound review gate;
2. its evidence requirements — what hash-bound preservation and explicit
   `ua2` evidence an eligible candidate must carry;
3. its acceptance semantics — what makes the candidate acceptable for gate
   closure and what invalidates it.

Only after that record is accepted can `dap-jo9.2.5.10` produce an eligible
candidate (the "Either" branch of its acceptance criteria).

## Owner and follow-up

- **Owner:** product owner (unnamed in repo records).
- **Date recorded:** 2026-10-17.
- **Follow-up Beads issues:** `dap-jo9.2.5.10` (blocked, decision-blocked
  label kept); blockers `dap-jo9.2.5.9` and `dap-jo9.2.5.9.1` remain gated on
  it.
- Until the owner decides, producing any "candidate" against a self-invented
  standard would bypass `dap-jo9.2.5.10`'s acceptance criteria and is
  prohibited by the "do not invent" house rule (no fabricated verification
  results, no self-invented approval).
