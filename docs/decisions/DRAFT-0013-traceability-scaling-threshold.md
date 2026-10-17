# DRAFT Decision 0013: Traceability checker scaling — monitor mode, no threshold adopted — open owner decision

**Status: DRAFT — NOT accepted. NO THRESHOLD IS RECORDED. NO PERFORMANCE BUDGET EXISTS.**

**Date: 2026-10-17.**

> **Decision pending owner acceptance.** This record is a draft for owner
> review only. It recommends a posture, not a policy. The recommendation is
> explicitly non-binding: until the owner explicitly accepts it (or an
> alternative), no (size, runtime-budget) pair may be cited as project
> policy, and no threshold may be enforced in CI or review. Any threshold
> pair adopted later is tooling-triage housekeeping — **it is NOT a product
> performance budget.** `dap-jo9.14` stays open and decision-blocked until
> then.

## The question

Beads spike `dap-jo9.14`: the traceability checker shows near-quadratic
Validate/GeneratedMarkdown growth above ~1,000 synthetic entities. Record
an explicit threshold decision — either a (size, runtime-budget) pair with
the growth assumption and review date that justify it, or an explicit "no
threshold adopted — monitor mode" statement — plus the owner and the
trigger condition for re-checking.

## Non-binding recommendation

**No threshold adopted — monitor mode**, with a concrete proposed re-check
trigger and a proposed review-by date (both below). This is the honest
answer the evidence supports: the scaling signal is real but rests on a
thin sample with a regime-dependent exponent, and the current catalog runs
~0.79 ms Validate — two orders of magnitude below any practical concern.
Adopting a (size, runtime-budget) pair today would be an invented forecast,
not an evidence-backed decision; the 2026-10-13 panel already rejected the
proposed "100 tasks and 10 ms" heuristics as mutually inconsistent and
undefined, and re-adopting a pair now would repeat that invention.

## Evidence

- **Controlled rerun (2026-10-13)** (`dap-jo9.14` description; raw
  repetitions in
  `docs/research/support/traceability-scaling-benchmark-2026-10-13.txt`):
  Go 1.27.2, linux/amd64, `-cpu=1` (GOMAXPROCS=1). Current catalog
  (17 requirements, 6 stories, 21 tasks, 26 cases): **0.526 ms Validate,
  0.107 ms GeneratedMarkdown** (medians). Synthetic one-to-one catalogs:
  100 entities → 0.557/0.464 ms; 1,000 → 19.896/21.196 ms; 5,000 →
  441.795/495.460 ms. Fitted exponents: Validate ~1.55 (100→1,000),
  ~1.93 (1,000→5,000); GeneratedMarkdown ~1.66, ~1.96. Each 5,000-entity
  repetition is a single benchmark iteration (N=1); five repetitions only.
  This is a scaling signal, not a current performance incident. No
  optimization was performed.
- **Re-verification (2026-10-17, this draft):** repository-backed
  benchmark re-run with the documented cold-pickup harness
  (`docs/research/support/traceability_scaling_benchmark_test.go.txt`,
  copied to `internal/traceability/` temporarily and removed after),
  `-cpu=1`, five repetitions, auto-scaled `b.N` (1,400–1,531 iterations).
  Environment: Go 1.27.2, linux/amd64, AMD EPYC 9D25, GOMAXPROCS=1.
  - Validate: 804151 / 789846 / 788855 / 787479 / 796082 ns/op →
    **median 0.790 ms** (baseline 0.526 ms).
  - GeneratedMarkdown: ~0.149 ms (baseline 0.107 ms).
  - **No checker regression.** `GeneratedMarkdown` is catalog-only
    (`docs/traceability/catalog.json` unchanged — CLI still reports
    17 requirements, 6 stories, 21 tasks, 26 specification-only cases),
    so its 1.40x uplift is environmental (different CPU: EPYC 9D25 now
    vs EPYC 9V74 on 2026-10-13). Validate's 1.50x is the same
    environmental factor plus tracker growth: `.beads/issues.jsonl`
    90 lines (pre-2026-10-14 commit) → 101 lines now; B/op 162864 →
    197890. Absolute scale is still sub-millisecond.
  - `go run ./cmd/traceability-check` (Go 1.27.2 from
    `~/workspace/toolchains/go/bin`): prints
    "traceability valid: 17 requirements, 6 stories, 21 tasks,
    26 specification-only cases", exit 0 — catalog identical to the
    2026-10-13 baseline counts; generated Markdown views compare
    byte-equal to the checked-in files (the command fails on any
    difference). Wall time ~12–18 ms including process startup and I/O.
- **Superseded:** the 2026-10-12 exploratory run is superseded (no
  contemporaneous toolchain/runner provenance) and is not the comparison
  baseline.

## Why monitor mode, not a threshold pair

1. **Thin sample.** The 5,000-entity medians (~442 ms Validate, ~495 ms
   GeneratedMarkdown) rest on five repetitions of a single iteration
   (N=1). A threshold built on them would pretend to a precision the
   data does not support.
2. **Regime-dependent exponent.** The fitted exponent changes with n
   (1.55→1.93 for Validate). Any extrapolated threshold is
   regime-dependent; picking one regime's exponent as "the" growth
   assumption would be a forecast, not a measurement.
3. **No current incident.** Current Validate is 0.79 ms — ~500x below
   even a conservative 500-entity re-check trigger's expected few-ms
   range. There is no performance problem to budget against.
4. **Housekeeping, not budget.** Per the issue's own goal statement, any
   adopted (size, runtime-budget) pair is tooling-triage housekeeping,
   not a product performance budget. Adopting a pair now risks it being
   misread as the latter.
5. **Prior panel resolution.** The 2026-10-13 panel recorded the proposed
   "100 tasks and 10 ms" review points as Proposed, not adopted —
   mutually inconsistent (at 100 synthetic entities Validate is
   ~0.557 ms, an order of magnitude below the 10 ms trigger) and
   undefined (entity mix, which measurement, median vs p99). They are
   not re-adopted here.

## Tradeoffs

| Option | Upside | Downside |
|---|---|---|
| Adopt a (size, runtime-budget) pair now | Concrete number to plan against; visible in CI | Invented forecast on a thin, regime-dependent sample; the 2026-10-13 panel already rejected the candidate heuristics as inconsistent/undefined; risks being misread as a product performance budget |
| **Monitor mode with concrete trigger (recommended)** | Zero invented authority; keeps the scaling signal on the record; re-check is deterministic and cheap | Threshold question stays open; requires owner to confirm the trigger and review-by date |
| Adopt indexing now to flatten growth | Would address the near-quadratic signal directly | No current incident justifies it; `dap-jo9.14` AC 3's byte-identical-views and exponent-<1.5 gates cannot be claimed without running them; premature optimization on a sub-millisecond tool |

## Proposed re-check trigger and review-by date (owner to confirm)

Re-measure when **any** of the following first occurs; the owner may
revise or discard each:

- **Catalog size:** total linked catalog entities
  (requirements + stories + tasks + cases) exceeds **500**. Order-of-magnitude
  justification, explicitly heuristic: at 1,000 synthetic entities Validate
  is ~19.9 ms with a 100→1,000 exponent of ~1.55, so 500 entities would be
  on the order of a few milliseconds — a safe early-warning point well
  before the ~1,000-entity knee. Current: 70.
- **Runtime:** repository-backed Validate benchmark (same cold-pickup
  harness, `-cpu=1`) exceeds **5 ms median**, or
  `go run ./cmd/traceability-check` wall time exceeds **1 s**.
- **Sunset:** proposed review-by **2027-04-13** (carried from the
  2026-10-13 cold-pickup note) — if the catalog has not approached the
  trigger by then, re-confirm or close.

All three are **Proposed**, not adopted.

## Owner and follow-up

- **Owner:** product triage — **no individual is named in repo records.**
  The generic "product owner" role appears in
  `docs/roadmap/product-roadmap.md` (v1 no-op REJECT acceptance) but is
  not tied to this decision; `dap-jo9.14` records the triage venue as
  open/unassigned. No owner is invented here.
- **Date recorded:** 2026-10-17.
- **Acceptance gate:** this DRAFT is accepted only when the owner
  explicitly records the threshold call (adopt a pair / monitor mode /
  alternative). Until then `dap-jo9.14` stays open, in-review,
  needs-input — never done.
- **Follow-up Beads issues:** `dap-jo9.14` (this decision); the proposed
  trigger points are recorded in the issue notes.
- **Indexing:** N/A — no indexing is proposed and no traceability-checker
  code changed. `go run ./cmd/traceability-check` exits 0 with the
  catalog valid, so diagnostics are unchanged and generated Markdown
  views are byte-identical to the checked-in baseline by construction.

## Unresolved — marked Proposed/Open

- Owner individual: unnamed; triage venue open/unassigned.
- Re-check trigger values (500 entities / 5 ms / 1 s): proposed, owner to confirm.
- Review-by date 2027-04-13: proposed, owner to confirm.
- Whether a (size, runtime-budget) pair should ever be adopted, and with
  what growth assumption: open.
