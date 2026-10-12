# Contributing

## Before starting

Read the [docs index](docs/README.md), [product scope](docs/requirements/product-scope.md), and [roadmap](docs/roadmap/product-roadmap.md). Confirm the relevant Beads issue and its dependencies before taking work.

## Git workflow

The local `work` branch tracks `origin/main`. Create a focused branch for a contribution, keep changes reviewable, and do not rewrite shared history. Inspect `git status` before and after work. Do not commit private documents, credentials, or machine-specific Beads runtime files. The maintained `.beads/issues.jsonl` issue export is intentionally reviewable and versionable. Do not push or publish unless explicitly requested.

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for every commit: `type(scope): imperative summary`. Keep the scope specific and the summary concise. Examples:

```text
docs(roadmap): define test-driven delivery plan
chore(beads): initialize issue tracking
```

## Beads workflow

Install Beads using the [official installation guide](https://beads.gascity.com/). This repository was initialized with Beads 1.3.1, embedded Dolt storage, issue prefix `dap`, and optional hooks/agent setup skipped. The live runtime database under `.beads/embeddeddolt/` is ignored by Git. `.beads/issues.jsonl` is the tracked recovery/interchange export for issue records and dependency edges; it is not a full backup of Dolt history or working-set state. Cross-clone synchronization through a Dolt remote is a separate operation. Do not run `bd dolt push` unless asked.

Common commands:

```sh
bd prime
bd ready
bd show dap-jo9
bd create "Describe the work" -t task -p 2 --description "..." --acceptance "..."
bd update <id> --claim
bd close <id> "Evidence and outcome"
bd doctor --check=artifacts
bd doctor --check=conventions
```

The current embedded Beads backend does not support `bd doctor --deep`; the artifact and convention checks above are supported in embedded mode.

Link prerequisites with `bd dep add <dependent-id> <prerequisite-id>`. Keep descriptions and acceptance criteria concrete, testable, and scoped. Use issues for work tracking; use `docs/roadmap/product-roadmap.md` for the durable, indexed roadmap.

After changing Beads issues, refresh and review the tracked export before committing:

```sh
bd export -o .beads/issues.jsonl
git diff -- .beads/issues.jsonl
```

To restore issue records and dependencies in a fresh clone, use a writable home directory and the committed export:

```sh
bd init --non-interactive --skip-agents --skip-hooks --prefix dap
bd import .beads/issues.jsonl
bd list
bd doctor --check=artifacts
bd doctor --check=conventions
```

Import is an upsert, so use the exact exported file and review the resulting IDs/count before continuing. A new local embedded database does not restore prior Dolt commit history. In this saved executor only, the default home is read-only; set `HOME` to a writable temporary directory for Beads commands if the same environment error occurs (for example, `mkdir -p /tmp/bd-home && HOME=/tmp/bd-home bd ...`). A normal checkout with writable home does not need that workaround.

## Test-first implementation convention

For every vertical slice, write a failing test that captures expected behavior, make it pass with the smallest implementation, and refactor with tests still green. Add the appropriate unit, adapter contract, integration, and end-to-end coverage. Use characterization and preservation checks for existing PDFs, failure injection and security tests for hostile inputs, and independent machine plus human semantic and assistive-technology validation. Do not claim PDF/UA or WCAG compliance from tags or a single automated check.

## Go and dependency policy

Go is the primary implementation language. Follow the checked-in Go module's toolchain version, keep packages focused, return errors with useful context, and use `context.Context` for cancellable or time-bounded operations. Prefer the Go standard library for small, project-specific tasks. The current direct third-party Go module, pdfcpu v0.16.2, is used only by `internal/pdfpreserve` characterization tests; its component and module-license evidence is in the [remediation component inventory](docs/research/remediation-component-inventory.md). It is not an approved product/runtime dependency, and full packaged dependency/security review remains open. Before accepting an external library for product/runtime use, record its exact version, direct and transitive license footprint, maintenance and security posture, and fit to the requirement. Keep candidates proposed until those checks and any needed feasibility decision are recorded; do not infer approval from a candidate list. Use `gofmt`, `go vet`, `go test -race ./...`, and `go mod tidy -diff` before publishing changes.

## Documentation conventions

- Keep documents concise, Markdown-only, and linked from the relevant index.
- Mark unresolved decisions **Proposed** or **Open**; include evidence and owner questions when known.
- Do not invent schedules, budgets, measured performance, test results, or approved architecture.
- Keep generation and remediation documentation in separate subfolders.
- Fixtures and examples must be synthetic or deidentified; never commit confidential source material.
