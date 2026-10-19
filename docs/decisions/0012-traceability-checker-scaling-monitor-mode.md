# Decision 0012: Traceability checker scaling — monitor mode, no threshold adopted

**Status: Accepted within the limits below.**

**Date: 2026-10-19.**

## Context

Beads issue `dap-jo9.14` required a scaling decision for the traceability
checker: either adopt an explicit (catalog size, runtime budget) threshold pair
(AC1 alternative 1) or explicitly record that no threshold is adopted and the
checker runs in monitor mode (AC1 alternative 2). A (size, runtime-budget) pair
would be a product performance budget, and the issue forbids imposing one, so
this record takes AC1 alternative 2.

## Decision

**No threshold is adopted — the traceability checker runs in monitor mode.**
No (entity count, runtime budget) pair is recorded as a gate, budget, or
acceptance bar for the checker.

## Evidence

Fresh scaling measurement from the 2026-10-19 benchmark run (5 repetitions,
`-cpu=1`, GOMAXPROCS=1; raw output retained in
`docs/research/support/traceability-scaling-benchmark-2026-10-19.txt`):

- Repository catalog + current Beads (101 issues): `Validate` p50 **0.924 ms**,
  `GeneratedMarkdown` p50 **0.169 ms** — far below any budget that would be
  proposed today.
- Synthetic scaling (`Validate`, median ns/op): 100 entities → 1.812 ms;
  1,000 entities → 29.939 ms; 5,000 entities → 594.502 ms.
- Synthetic scaling (`GeneratedMarkdown`, median ns/op): 100 → 1.041 ms;
  1,000 → 35.197 ms; 5,000 → 688.559 ms.

Fitted runtime growth exponents (log-log on medians), from
`docs/research/support/traceability-scaling-analysis-2026-10-19.md`:

- `Validate` 100→1000: **1.218**; 1000→5000: **1.857**.
- `GeneratedMarkdown` 100→1000: **1.529**; 1000→5000: **1.848**.

Growth assumption recorded: runtime grows superlinearly and steepens with
size — near-linear at small sizes (~1.2–1.5) and approaching quadratic (~1.85)
on the 1000→5000 interval. A fixed threshold pair would need re-fitting at
each order of magnitude, which is why no threshold is adopted.

Fixture note: the 2026-10-13 benchmark harness predates the standalone
task-policy validation rules (PR #59, 2026-10-15); a scratch-only adapter
merged the repository catalog/issues as a constant base so the synthetic
fixture validates. The scratch test was deleted before commit.

## Owner

Owner: dap-nightly-build automation. It re-runs
the scaling benchmark whenever the trigger below fires and re-decides (adopts
a threshold pair or re-affirms monitor mode) with fresh numbers.

## Re-check trigger

Re-run the benchmark and re-decide when either condition holds:

- Catalog total entities exceeds **1,000** (87 on 2026-10-19), OR
- repository-catalog `Validate` p50 (`-cpu=1`) exceeds **5 ms** (0.924 ms on
  2026-10-19).

**Review-by: 2027-04-19** — revisit this decision no later than this date even
if neither trigger fires.

## Limits

- This record is **not a product performance budget**. It authorizes no
  runtime or size gate and may not be cited as one in acceptance criteria.
- Any future decision to impose a (size, runtime-budget) threshold pair
  re-opens `dap-jo9.14` AC3 (a product decision, not an automation decision).
- The fitted exponents are a measurement of the current implementation on the
  benchmark host; they are not a guarantee of behavior on other hosts or for
  future implementations.

## Follow-up

- `dap-jo9.14` is closed by this record.
