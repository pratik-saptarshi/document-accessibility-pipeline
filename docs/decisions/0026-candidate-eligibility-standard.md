# Decision 0026: Candidate eligibility standard — assessed evidence

**Status: Accepted.**

**Date: 2026-10-26.** Decided by the product owner (Pratik Saptarshi),
who selected the assessed-evidence standard (Option C) from the
2026-10-25 decision brief for `dap-jo9.2.5.10`.

## Context

`dap-jo9.2.5.9` (the candidate-bound semantic + AT review gate) cannot
close without an actual eligible candidate with hash-bound
preservation/`ua2` evidence, and `dap-jo9.2.5.10` exists to produce
one. The bead was decision-blocked: the eligibility standard itself
was an open product-owner decision, so the autonomous build loop
skipped the item from 2026-10-18 onward.

What the record already settled, and this decision does not disturb:

- The bounded v2 association-repair Proceed under `dap-jo9.1.6.4` /
  Decision 0005 is feasibility evidence only. It does not make either
  historical candidate eligible or compliant; both remain
  NONCOMPLIANT (invoice 74, statement 91 failed `ua2` checks), with
  findings accepted as inherited for the feasibility decision only.
- The historical record contains no damaged-input validator reports,
  so inheritance was never established by a before/after validator
  comparison.
- Machine (`ua2`) evidence is separate from, and can never satisfy,
  the genuine human semantic / AT outcomes required by
  `dap-jo9.2.5.9`. Eligibility for entering review is distinct from
  the human/AT acceptance that closes `.2.5.9`.

## Decision

A candidate is **eligible** for the candidate-bound review gate when
all five hold:

1. The candidate is produced from immutable damaged bytes by a
   declared, bounded operation.
2. Hash-bound preservation evidence is complete.
3. A pinned `ua2` run is executed on **both** the damaged input and
   the candidate, with full reports retained — closing the historical
   record's inheritance gap.
4. Every finding is classified inherited vs introduced, with **zero
   introduced findings**. A candidate with even one introduced
   finding is ineligible.
5. Compliance itself is **not** required. The genuine human/AT review
   under `dap-jo9.2.5.9` remains the acceptance step, and machine
   evidence still cannot satisfy it.

Eligibility, acceptance, and compliance remain three distinct states.

## Alternatives considered

- **Option A — Full conformance** (eligible = zero failed `ua2`
  checks). Rejected by the owner: neither historical candidate
  qualifies (74 / 91 failures), so producing an eligible candidate
  becomes a full-remediation effort of unknown size and `.2.5.9`'s
  demonstration stays blocked until one exists.
- **Option B — Bounded-Proceed parity** (eligible = the `.1.6.4`
  Proceed standard). Rejected by the owner: it collapses the
  feasibility/eligibility distinction that Decision 0005, the
  acceptance register, and `.1.6.4`'s own text were written to
  preserve, and adopting it would mean formally superseding those
  recorded boundaries.

## Consequences

- `dap-jo9.2.5.10` is unblocked: the standard is recorded here with
  owner and date, satisfying the recording requirement of the bead's
  acceptance-criteria alternative 2. The bead's `decision-blocked`
  label is removed and a Beads note records this decision. The bead
  itself stays open — recording the standard does not satisfy
  candidate-production acceptance and does not authorize closing it.
- Candidate production may now proceed under this standard. The only
  new work the standard requires is the paired damaged-input /
  candidate `ua2` runs; the evidence must be generated fresh, not
  inherited from the October archives.
- Nothing in this decision changes Decision 0005, the acceptance
  register, or `.1.6.4`; the feasibility/eligibility boundary stands.
