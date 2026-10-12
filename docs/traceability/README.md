# Traceability data and checks

`catalog.json` is the source for stable capability, requirement, story, function, and acceptance-case IDs. It links to exact Beads tasks; `.beads/issues.jsonl` is the reviewable issue/dependency export.

Run the Go standard-library planning checker from the repository root:

```sh
go run ./cmd/traceability-check       # validate catalog, Beads links, and generated views
go run ./cmd/traceability-check --write  # regenerate the Markdown views
go test -race ./...
```

The Go checker verifies unique IDs, all references, Beads task/story parent-child edges and status labels, dependency endpoints and acyclic blocking edges, lane/family case coverage, Gherkin-style clauses, and generated-view consistency. It validates planning data only. It does not implement or test PDF behavior. The catalog's small, project-specific rules use Go's standard JSON and filesystem packages; no third-party Go module is needed for these checks.

Generated views are [requirements](../requirements/catalog.md), [function contracts](../requirements/functional-contracts.md), [acceptance cases](../validation/cases.md), and the [matrix](matrix.md). Edit the JSON source, regenerate, and review the diff rather than hand-editing generated views.
