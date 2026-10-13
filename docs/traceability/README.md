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

The [proposed profile contract](../profiles/README.md) and its generation/remediation schemas support existing data requirements `REQ-GEN-001` / `REQ-GEN-002` and remediation characterization requirements `REQ-REM-001` through `REQ-REM-004` in the generated requirements catalog. Profile tests are implementation-level contract tests; they do not change product acceptance-case status or approve the open profiles.

The [legacy requirement crosswalk](legacy-requirement-crosswalk-2026-10-13.md) preserves eleven historical story references as proposed continuity mappings to current catalog requirements. It does not claim proven semantic equivalence.
