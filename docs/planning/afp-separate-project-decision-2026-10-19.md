# AFP separate-project decision record (2026-10-19)

**Status:** Recorded.
**Beads story:** `dap-jo9.4.1` ("Decide whether IBM AFP migration merits a separate project").
**Requirement:** REQ-AFP-001.
**Date:** 2026-10-19.

## Decision

**Decision: YES — IBM AFP migration merits a separate project.**

The AFP effort is the separate project **go-afp-line**, published at
<https://github.com/pratik-saptarshi/go-afp-line> (Apache-2.0). AFP conversion
work is not a lane of this repository.

## Authority

Pratik Saptarshi's explicit **GO** decision on 2026-10-13 (recorded in
project memory; AFP track decision).

## Bounded, evidence-backed feasibility result

### Feasibility spike (2026-10-13)

The spike (`afp-feasibility-spike-2026-10-13.md` — a local planning note,
uncommitted in the main checkout as of 2026-10-19; findings summarized below)
found:

- No mature, permissively-licensed open-source AFP→(tagged)PDF converter exists.
  The strongest open-source converter found (Alpheus AFP Parser, maintained
  chatelao fork) is Java-only and GPL-3.0 — excluded by the permissive-license
  constraint; the Apache-2.0 parser found (yan74/afplib) is Java-only and not
  a converter; the only Apache-2.0-adjacent Go-relevant references (CC0
  Haskell/Perl implementations) are parsing/reference only.
- **No Go implementation of IBM AFP (MO:DCA) parsing exists at all**
  (GitHub/language searches for `modca`, `afpds`, and "Advanced Function
  Presentation" in Go returned zero relevant repositories).
- AFP is a final-form print stream (absolute positioning, no logical
  structure), so producing tagged PDF/UA-2 requires semantic reconstruction;
  Decision 0001 already records that machine validation does not establish
  meaning, reading order, or assistive-technology usability.

The spike recommended keeping AFP on a separate feasibility track rather than
a prerequisite for either PDF lane.

### Separate-project delivery evidence (go-afp-line)

The separate project was published 2026-10-14 at
<https://github.com/pratik-saptarshi/go-afp-line> under the **Apache-2.0**
license. Delivered scope (per its README):

- `afp/` — MO:DCA codec, structured fields, triplets, code pages 037/500.
- `line/` — line-data model.
- `afptoline/` — PTOCA reducer with conservation law.
- `linetopdf/` — untagged PDF rendering via pdfcpu.
- Fuzzers caught real bugs pre-ship (per its README).

Follow-up merges after publication:

- PR #16, `fix: persist PTOCA state across PTX fields; map NEL to space`,
  squash-merged as `50f7546`.
- PR #17, `docs(agents): add public-tree hygiene note`, squash-merged as
  `6af4cb2`.

Repository description (verified 2026-10-19 via `gh`): "Native Go library for
AFP/MO:DCA parsing, line-data extraction, and untagged PDF rendering".

## Closure

This record satisfies `dap-jo9.4.1`'s acceptance criterion: a bounded,
evidence-backed feasibility result for the agreed AFP input/output goal, with
the decision outcome recorded above.

This record does not authorize AFP conversion implementation in this
repository — the story's own acceptance criterion specifies decision outcome
only; implementation in this repo is out of scope.

## Open follow-ups

- `dap-jo9.4.2` (track AFP go/no-go) remains **Open** with Pratik.
