# Traceability checker scaling threshold

**Status: Proposed review trigger.** This records planning-tool measurements and a point to remeasure. It is not a product performance target, CI service-level objective, or decision to add indexes.

## Current catalog measurement

At source revision `9eb160d2b58a258edb2076f1ab6be55142ec3098`, `go run ./cmd/traceability-check` validated 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases. A repository-backed benchmark loaded the tracked catalog and Beads export, then timed validation and generation of all four Markdown views. Five serial repetitions with `-benchtime=300ms -benchmem -cpu=1` had medians of 0.526 ms and 162,864 B/op for validation, and 0.107 ms and 215,802 B/op for Markdown generation. These figures are local CPU benchmark evidence, not end-to-end command or CI duration.

Benchmark environment: Go 1.27.2, `GOOS=linux`, `GOARCH=amd64`; the Codex execution container reported hostname `902195ed1ae9`, Linux 6.18.44 x86_64, and an AMD EPYC 9V74 80-Core Processor. The container saw CPUs 0–4 and a cgroup quota of four CPUs. Benchmarks explicitly used `-cpu=1`; the runtime probe reported `NumCPU=5` and `GOMAXPROCS=1`, so benchmark operation concurrency was one. These results are tied to this ephemeral runner and are not a cross-machine performance guarantee.

The synthetic one-to-one benchmark scales each of requirements, stories, function contracts, tasks, and cases together. Five serial repetitions produced these medians:

| Linked entities in each collection | Validate | Validate B/op | Generate four views | Generation B/op |
|---:|---:|---:|---:|---:|
| 100 | 0.557 ms | 174,776 | 0.464 ms | 492,168 |
| 1,000 | 19.896 ms | 2,050,224 | 21.196 ms | 5,486,916 |
| 5,000 | 441.795 ms | 8,452,576 | 495.460 ms | 28,356,104 |

This shows steep growth in this regular one-to-one synthetic fixture, while repository-backed validation and generation remain below 0.6 ms at the current catalog size. Each record links to one corresponding record; the links are sparse. The fixture is a stress signal, not a forecast of document throughput or future planning content. Five raw repetitions for each measurement and the runner probe are retained in [`support/traceability-scaling-benchmark-2026-10-13.txt`](support/traceability-scaling-benchmark-2026-10-13.txt).

## Proposed remeasurement trigger

The values **100 live tasks** and **10 ms median repository-backed runtime** are heuristic proposed review points: remeasure when either is reached. They are not measured growth forecasts, adopted policy, an SLO, or a performance promise. One hundred is about 4.8 times the current 21 tasks; no final catalog size is forecast. Ten milliseconds is below the roughly 20 ms measured at 1,000 synthetic linked entities and above this runner's sub-millisecond current-catalog medians. Product triage must decide whether to adopt, revise, or discard these review points.

Do not add relationship indexes solely because the synthetic curve is quadratic. If the trigger is reached, compare an indexed implementation against the current output: require identical validation diagnostics, byte-identical generated Markdown, repeated benchmarks showing materially flatter growth, and no allocation regression. Until then, keep the checker unchanged.

## Reproduction

The harness at [`support/traceability_scaling_benchmark_test.go.txt`](support/traceability_scaling_benchmark_test.go.txt) uses only the Go standard library and the current package. Copy it temporarily to `internal/traceability/review_benchmark_test.go`, then run:

```bash
go test -run '^TestReviewBenchmarkEnvironment$' -v -cpu=1 ./internal/traceability
go test -run '^$' -bench '^BenchmarkRepositoryTraceability$' -benchmem -benchtime=300ms -count=5 -cpu=1 ./internal/traceability
go test -run '^$' -bench '^BenchmarkReview(Validate|GeneratedMarkdown)$' -benchmem -benchtime=300ms -count=5 -cpu=1 ./internal/traceability
```

The repository-backed benchmark reads the actual tracked catalog and Beads export. The synthetic benchmark creates valid linked records at 100, 1,000, and 5,000 entities per collection. Its synthetic Beads tasks include the current parser-required `## Test cases` section. No production checker/parser code or dependencies were changed for these measurements.
