# DRAFT Decision 0012: Whether IBM AFP migration merits a separate project — open owner decision

**Status: DRAFT — NOT accepted. NO PROJECT-STRUCTURE DECISION IS RECORDED.**

**Date: 2026-10-17.**

> **Decision pending owner acceptance.** This record is a draft for owner
> review only. It recommends a posture, not a project structure. The
> recommendation is explicitly non-binding: until the owner explicitly
> accepts it (or an alternative), AFP work remains a feasibility track
> inside this repository's planning scaffolding and no project separation
> may be claimed, funded, or organized. Nothing here authorizes AFP
> conversion implementation. `dap-jo9.4.1` stays open and decision-blocked
> until then.

## The question

Beads story `dap-jo9.4.1`: as a *proposed* AFP print-operations owner, a
bounded, evidence-backed feasibility result for an *agreed* AFP
input/output goal is wanted, so a go/no-go decision can be made before
migration implementation is considered. The story asks whether IBM AFP
migration merits a separate project.

## Non-binding recommendation

**Defer the separate-project decision: keep AFP as a separate parallel
feasibility track within this repository (not a separate project), and
revisit project separation only after `dap-jo9.4.1.2` produces bounded
go/no-go evidence AND the owner agrees the AFP input/output target,
fidelity criteria, persona, and go/no-go authority.**

This is the honest answer the evidence supports: none of the four
prerequisites for a structural decision currently exist, so a
go/no-go project-separation call today would be invented, not
evidence-backed.

## Evidence

- **Capability** `CAP-AFP-FEASIBILITY` ("Separate AFP feasibility decision",
  `docs/traceability/catalog.json:45`, lane "AFP feasibility"): the
  capability exists only as a planning contract, not an approved
  implementation scope.
- **Requirement** `REQ-AFP-001` ("Separate AFP feasibility decision",
  `docs/requirements/catalog.md:69-71`): "An AFP feasibility recommendation
  requires an agreed input/output target, representative safe fixtures,
  fidelity criteria, and evidence; conversion is not authorized." Its open
  decision (`docs/requirements/catalog.md:83`) leaves "AFP input/output
  target, workload, fonts/barcodes, fidelity requirements, and go/no-go
  authority" unresolved — unchanged since the story was opened 2026-10-12.
- **Acceptance test specification** `TC-AFP-FEASIBILITY-001`
  (`docs/validation/cases.md:205`; `docs/traceability/catalog.json:1141-1148`,
  status "specification-only, not executed"): **Given** "AFP input/output
  goals, safe representative fixtures, and fidelity criteria are not yet
  agreed" — **When** the feasibility issue is refined — **Then** "Missing
  prerequisites are listed and the go/no-go task stays blocked; no converter
  is selected or implemented." The story's own test spec *prescribes* the
  blocked outcome while prerequisites are missing.
- **Fixture/fidelity protocol** (`dap-jo9.4.1.1`, closed 2026-10-12): the
  specification exists — case IDs in the traceability catalog with
  Given/When/Then outcomes, refinement-ready. But the definition of done
  for that task explicitly recorded that "unresolved profiles/feasibility"
  remain blockers and "no implementation result is claimed". A protocol for
  measuring fidelity is not evidence of feasibility.
- **Options assessment** (`dap-jo9.4.1.2`, OPEN, P2): the bounded
  assessment that must produce the representative input/resource checklist,
  target and fidelity criteria, exact candidate versions and license
  evidence, operational boundaries, gaps, and an evidence-backed
  go/no-go/unknown recommendation. It has not run; its evidence does not
  exist yet.
- **Repository planning position** (roadmap, scope): "IBM AFP print migration
  is a separate feasibility track" (`docs/roadmap/product-roadmap.md:7`);
  "Keep AFP separate and stop at feasibility until goals are agreed"
  (`docs/roadmap/product-roadmap.md:104`); "It is not presumed to be part of
  either PDF lane's implementation" (`docs/requirements/product-scope.md:17-19`).
  No research note, accepted decision (0001–0009), or plan sets a project
  boundary, budget, team, or migration target for AFP.

## Tradeoffs

| Option | Upside | Downside |
|---|---|---|
| Separate project now | Dedicated staffing/budget; clear accountability | Would invent scope, target, and fidelity criteria that REQ-AFP-001 requires to be *agreed* first; contradicts TC-AFP-FEASIBILITY-001's Then-clause |
| Separate project after `dap-jo9.4.1.2` evidence | Decision becomes evidence-backed; satisfies the story's acceptance | Costs waiting for the assessment and owner agreement; may still come out no-go/unknown |
| Absorb AFP into a PDF lane | Fewer tracks to coordinate | Roadmap and scope explicitly forbid it: "Do not implement a converter or make AFP a prerequisite for either PDF lane" |
| Stay feasibility-track-in-repo (recommended) | Zero invented authority; matches all accepted planning records; keeps the go/no-go gate explicit via `dap-jo9.4.2` | Project-separation question stays open; requires owner engagement to advance |

## Unresolved — marked Proposed/Open

- **Persona:** "proposed AFP print-operations owner" — actual title, authority, and approval responsibilities are open (REQ-AFP-001 open decisions; story description).
- **Workload:** input profile, source systems, volumes, and representative AFP data characteristics are unresolved; only synthetic fixtures are permitted. Candidate sample source (reference only — do not ingest files without deidentification review): https://www.afpworld.com/educate/sample-afp-files.html
- **Output target:** agreed input/output goal is unresolved — the single prerequisite everything else hangs on.
- **Fidelity criteria:** fonts, barcodes, page fidelity, variable data constraints unresolved; `dap-jo9.4.1.1` defined the protocol, not the criteria.
- **Go/no-go authority:** who decides, and what unknown means for project status, unresolved.

## Owner and follow-up

- **Owner:** proposed AFP print-operations owner (unnamed in repo records).
- **Date recorded:** 2026-10-17.
- **Acceptance gate:** this DRAFT is accepted only when the owner explicitly
  records the project-structure call (separate project / not / revisit after
  evidence). Until then `dap-jo9.4.1` stays open, in-review, needs-input —
  never done.
- **Follow-up Beads issues:** `dap-jo9.4.1.2` (produce the missing
  go/no-go evidence; RALPH-034); `dap-jo9.4.2` (blocked on this decision;
  records the go/no-go/unknown outcome with the owner once prerequisites
  resolve). Verification scope for the decision-blocked follow-ups: verify
  the recorded prerequisites and the evidence chain, not the recommendation
  itself.
- **No implementation authorized:** this record does not approve an AFP
  converter, any candidate tool, or a separate repository/organization.
