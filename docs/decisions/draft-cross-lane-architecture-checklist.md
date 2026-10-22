# DRAFT — Cross-lane architecture decision checklist

**Status: DRAFT (not accepted).** This checklist structures `dap-jo9.1.4`
("Record cross-lane architecture and acceptance decisions") so the open
architecture gates are explicit, each with its required evidence, fixtures,
owner, alternatives, and status. **Nothing below is approved.** Accepted vs
proposed is explicit per gate; no unverified engine, compliance level, or
production design is presented as approved. The bead stays open until the
linked profile and feasibility work provides reviewable evidence.

**Date: 2026-10-21. Owner: Pratik Saptarshi.**

## How to read this checklist

Each gate lists: **Decision** (what must be recorded), **Status**,
**Evidence required**, **Test fixtures**, **Alternatives**, **Acceptance
owner**. A gate flips from PROPOSED to ACCEPTED only when the evidence is
reviewable and the owner signs. PROPOSED values below are starting
positions, not decisions.

## Gate 1 — Control plane: Go API, job orchestration, audit, semantic repair

- **Status:** PROPOSED.
- **Decision:** the control-plane shape — Go API surface, job orchestration
  model, audit-trail format, and where semantic repair lives (control plane
  vs worker).
- **Evidence required:** profile feasibility evidence from `dap-jo9.1.1` (open:
  semantic policy not approved per Decision 0017) and audit requirements from
  `dap-jo9.1.3`.
- **Test fixtures:** none yet.
- **Alternatives:** not yet enumerated.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 2 — Replaceable remediation worker adapters

- **Status:** PROPOSED.
- **Decision:** the adapter interface boundary that lets remediation workers
  be swapped (e.g., different repair engines per document family).
- **Evidence required:** at least one implemented worker behind the proposed
  interface, proving the boundary is real and not speculative.
- **Test fixtures:** none yet.
- **Alternatives:** not yet enumerated.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 3 — Independent generation lane

- **Status:** PROPOSED (compartmentalization accepted, engine open).
- **Decision:** generation stays a separate lane (Decision .1.1.2 — accepted);
  **which generation engine** is undecided.
- **Evidence required:** generation-engine spike `dap-jo9.1.5` (not run).
- **Test fixtures:** none yet.
- **Alternatives:** not yet enumerated.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 4 — Data profiles

- **Status:** PARTIALLY ACCEPTED.
- **Decision:** 8 document families with per-family required fields — accepted
  (Decision 0014). **Semantic rules for those families — not approved**
  (Decision 0017 explicitly approves no semantic rules).
- **Evidence required:** profile schema + validator (`internal/profile`,
  done) and the semantic-policy decision (open).
- **Test fixtures:** 8-family example sets (PR #181, in flight).
- **Alternatives:** recorded in Decisions 0014/0017.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 5 — PDF writer / fidelity

- **Status:** PROPOSED.
- **Decision:** which PDF writer produces pipeline output and what fidelity
  bar it must clear.
- **Evidence required:** writer evaluation against the 8-family examples;
  for AFP-sourced PDFs, the go-afp-line output vs Decision 0020's trusted
  references.
- **Test fixtures:** 8-family examples (PR #181); AFP fixtures
  (Decision 0020).
- **Alternatives:** not yet enumerated.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 6 — Generation engine

- **Status:** PROPOSED.
- **Decision:** the engine behind the generation lane.
- **Evidence required:** generation-engine spike `dap-jo9.1.5` (not run).
- **Test fixtures:** none yet.
- **Alternatives:** not yet enumerated.
- **Acceptance owner:** Pratik Saptarshi.

## Gate 7 — Compliance target

- **Status:** PARTIALLY ACCEPTED.
- **Decision:** WCAG 2.2 only, APCA removed — accepted (Decision 0014);
  PDF/UA-2 is the pipeline's accepted target (Decision 0001). **AFP-sourced
  PDF tagging — open** (Decision 0020 scopes the first AFP profile to PDF 1.7
  visual fidelity; UA tagging is a follow-up).
- **Evidence required:** validator evidence (`dap-jo9.1.6.3` — blocked on
  durable raw validator XML/contracts).
- **Test fixtures:** none yet.
- **Alternatives:** recorded in Decisions 0001/0014.
- **Acceptance owner:** Pratik Saptarshi.

## Explicitly out of scope for this checklist

Writer authorization and trust-anchor key custody (out of scope per
Decision 0019); go-afp-line dependency adoption (belongs to `dap-jo9.4.2`);
AFP go/no-go (belongs to `dap-jo9.4.2`).
