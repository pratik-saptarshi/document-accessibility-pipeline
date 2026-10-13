# Traceability checker scaling threshold

**Status: Proposed review trigger.** This records planning-tool measurements and a point to remeasure. It is not a product performance target, CI service-level objective, or decision to add indexes.

## Current catalog measurement

At source revision `9eb160d2b58a258edb2076f1ab6be55142ec3098`, `go run ./cmd/traceability-check` validated 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases. A repository-backed benchmark loaded the tracked catalog and Beads export, then timed validation and generation of all four Markdown views. Five serial repetitions with `-benchtime=300ms -benchmem` had medians of 0.464 ms and 164,108 B/op for validation, and 0.112 ms and 215,958 B/op for Markdown generation. These figures are local CPU benchmark evidence, not end-to-end command or CI duration.

The synthetic one-to-one benchmark scales each of requirements, stories, function contracts, tasks, and cases together. Five serial repetitions produced these medians:

| Linked entities in each collection | Validate | Validate B/op | Generate four views | Generation B/op |
|---:|---:|---:|---:|---:|
| 100 | 0.524 ms | 176,091 | 0.473 ms | 492,369 |
| 1,000 | 19.389 ms | 2,073,541 | 20.426 ms | 5,490,021 |
| 5,000 | 423.537 ms | 8,490,288 | 482.676 ms | 28,361,168 |

This confirms steep synthetic growth, while the repository-backed validation and generation remain below 0.5 ms at the current catalog size. The synthetic fixture is deliberately dense and regular; it is a stress signal, not a forecast of document throughput or an exact model of future planning content.

## Proposed remeasurement trigger

Rerun the repository-backed and synthetic benchmarks when the live catalog reaches **100 tasks** (about 4.8 times the current 21, and room for roughly a dozen tasks across each of the eight lane/family combinations), or when either repository-backed operation has a median above **10 ms**, whichever occurs first. This is a review trigger, not a performance promise. The size count is grounded in current catalog size and the roadmap's eight lane/family combinations; the runtime trigger is below the measured 19–20 ms at 1,000 synthetic linked entities and well above the current sub-millisecond baseline.

Do not add relationship indexes solely because the synthetic curve is quadratic. If the trigger is reached, compare an indexed implementation against the current output: require identical validation diagnostics, byte-identical generated Markdown, repeated benchmarks showing materially flatter growth, and no allocation regression. Until then, keep the checker unchanged.

## Reproduction

The harness at [`support/traceability_scaling_benchmark_test.go.txt`](support/traceability_scaling_benchmark_test.go.txt) uses only the Go standard library and the current package. Copy it temporarily to `internal/traceability/review_benchmark_test.go`, then run:

```bash
go test -run '^$' -bench '^BenchmarkRepositoryTraceability$' -benchmem -benchtime=300ms -count=5 ./internal/traceability
go test -run '^$' -bench '^BenchmarkReview(Validate|GeneratedMarkdown)$' -benchmem -benchtime=300ms -count=5 ./internal/traceability
```

The repository-backed benchmark reads the actual tracked catalog and Beads export. The synthetic benchmark creates valid linked records at 100, 1,000, and 5,000 entities per collection. Its synthetic Beads tasks include the current parser-required `## Test cases` section. No production checker/parser code or dependencies were changed for these measurements.
