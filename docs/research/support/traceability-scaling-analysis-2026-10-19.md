# Traceability checker scaling analysis — 2026-10-19

Evidence: `traceability-scaling-benchmark-2026-10-19.txt` (raw `go test -v` output,
5 repetitions, `-cpu=1`, GOMAXPROCS=1, AMD EPYC 9D25).

Fixture note: the stored harness (`traceability_scaling_benchmark_test.go.txt`,
2026-10-13) predates the standalone task-policy validation rules added by PR #59
(2026-10-15, `standaloneTaskPolicies` for dap-jo9.2.5.2 / dap-jo9.2.5.9 in
`internal/traceability/catalog.go`), so the verbatim harness could not validate its
synthetic fixture and its synthetic benchmarks failed. A scratch-only adapter
merged the repository's real catalog and Beads issues into the synthetic fixture
(disjoint ID spaces; the repository base validates cleanly). The scratch test was
deleted before commit. Medians below therefore include the constant repository
base; growth exponents measure the synthetic scaling component.

## Measured medians (5 reps, `-cpu=1`)

`BenchmarkReviewValidate` (per-`Validate` call over the whole fixture):

| synthetic entities | median ns/op | median ms/op | median B/op | median allocs/op |
|---|---|---|---|---|
| entities_100 | 1,811,736 | 1.812 | 412,822 | 1,313 |
| entities_1000 | 29,938,924 | 29.939 | 2,260,611 | 8,559 |
| entities_5000 | 594,501,923 | 594.502 | 8,761,264 | 40,730 |

`BenchmarkReviewGeneratedMarkdown` (per-`GeneratedMarkdown` call):

| synthetic entities | median ns/op | median ms/op | median B/op | median allocs/op |
|---|---|---|---|---|
| entities_100 | 1,041,061 | 1.041 | 731,909 | 4,688 |
| entities_1000 | 35,196,636 | 35.197 | 5,623,752 | 38,023 |
| entities_5000 | 688,559,314 | 688.559 | 28,420,392 | 186,066 |

Repository-backed (real catalog + current Beads, same run):

| benchmark | median ms/op |
|---|---|
| `BenchmarkRepositoryTraceability/Validate` | 0.924 |
| `BenchmarkRepositoryTraceability/GeneratedMarkdown` | 0.169 |

## Fitted runtime growth exponents (log-log on median ns/op)

- `Validate` 100→1000: exponent **1.218**
- `Validate` 1000→5000: exponent **1.857**
- `GeneratedMarkdown` 100→1000: exponent **1.529**
- `GeneratedMarkdown` 1000→5000: exponent **1.848**

## Growth assumption (recorded)

Runtime growth is superlinear and steepens with size: near-linear at small sizes
(~1.2–1.5 on the 100→1000 interval) and approaching quadratic (~1.85 on the
1000→5000 interval). At 5,000 synthetic entities a single `Validate` already costs
~0.6 s at `-cpu=1`; extrapolated per-interval, a further 10x entity growth would
cost roughly 70x runtime. The absolute repository numbers (0.924 ms p50) are far
from any budget today, but the growth curve means a fixed (size, budget) threshold
pair would need re-fitting at each order of magnitude — which is why this run's
decision record adopts monitor mode rather than a hard threshold.

Variance caveat: `entities_5000` reps auto-tuned to `b.N=2` (~0.6 s/op), so the
5000-entity medians carry higher variance than the 100/1000-entity medians.
