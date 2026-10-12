# Beads issue tracking

This repository uses Beads with the `dap` issue prefix and embedded Dolt storage.

- The live embedded database is under `.beads/embeddeddolt/` and is ignored by Git.
- `.beads/issues.jsonl` is the tracked issue and dependency export. Refresh it after issue changes with `bd export -o .beads/issues.jsonl` and review the diff before committing.
- The JSONL file can restore issue records and dependency edges in a fresh checkout, but it does not contain Dolt commit history or working-set state. See [contributor guidance](../CONTRIBUTING.md) for restore steps.
- Use `bd ready`, `bd show <id>`, `bd update <id> --claim`, `bd create`, `bd dep add`, and `bd close` to manage work. See the [roadmap](../docs/roadmap/product-roadmap.md) for the current issue graph.
- Beads Dolt synchronization and GitHub publishing are separate operations. Do not push either without a request to publish.
