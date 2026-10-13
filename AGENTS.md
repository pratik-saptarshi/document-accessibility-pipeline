# Repository guidance

## Project boundaries

- This repository is planning scaffolding. Read [the docs index](docs/README.md), [product scope](docs/requirements/product-scope.md), [roadmap](docs/roadmap/product-roadmap.md), and relevant decision records before changing behavior. A candidate tool or architecture is not approved unless an accepted decision says so; mark unresolved choices **Proposed** or **Open**.
- Keep existing-PDF remediation and document generation as separate workflow lanes. Use synthetic or deidentified fixtures only; never add private reports or source documents.
- Go is the primary implementation language. Prefer the standard library for small project-specific work. Reuse permissively licensed libraries where they fit; before adopting a product/runtime dependency, record exact direct and transitive licenses, maintenance/security posture, and requirement fit. Keep candidate libraries unapproved until reviewed. `pdfcpu` is currently used only by preservation characterization tests, not product runtime.
- Keep packages focused. Pass `context.Context` for cancellation or deadlines, return useful errors, and document concurrency guarantees and ownership where they matter. Do not assume a method is concurrency-safe. Bound work, memory, file sizes, subprocess output, and concurrency where inputs or child processes can grow. Fuzz parsers and other hostile-input boundaries where a useful invariant can be asserted. Prefer small interfaces owned by their consumers when actual use justifies them; do not add abstractions speculatively. See [Go code review guidance](https://go.dev/wiki/CodeReviewComments), [error wrapping guidance](https://go.dev/blog/go1.13-errors), and [race detector documentation](https://go.dev/doc/articles/race_detector); race tests cover only executed paths.

## Work and tests

- Track work in Beads. Run `bd prime`, `bd ready`, and `bd show <id>` before changing an issue. Link prerequisites, keep acceptance criteria testable, and close only with evidence that each criterion passed. After issue changes, export and review `.beads/issues.jsonl`; it is an interchange export, not a full Dolt backup. See [CONTRIBUTING.md](CONTRIBUTING.md) for recovery and command details.
- Use test-first vertical slices: add a failing test, make the smallest change that passes, then refactor with tests green. Use preservation/characterization tests for existing PDFs, and failure-injection/security tests for hostile input. Do not implement product features under a planning-only request.
- `go test -race` needs cgo and a C compiler in the test environment; this does not make cgo a product runtime requirement. The root `./...` commands do not traverse the two nested Go modules, so run their checks separately. These are the repository's relevant Go checks:

  ```sh
  # Root module
  test -z "$(gofmt -l .)"
  go mod tidy -diff
  go vet ./...
  go test -race ./...
  go run ./cmd/traceability-check

  # Standalone cloud-tools bootstrap module
  (cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./...)
  bash -n scripts/bootstrap-cloud-build-tools.sh

  # CI-only PDF/UA-2 adapter module: Go process-contract tests only
  (cd ci/pdfua2/adapter && go test -race ./... -count=1 && go vet ./...)

  # Source-publisher guard tests
  bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh
  bash .github/scripts/test-source-publisher.sh
  ```

- The PDF/UA-2 target is accepted in [Decision 0001](docs/decisions/0001-pdf-ua-2-validation-target.md). Select the explicit `ua2` profile; never silently fall back. Machine validation does not establish correct meaning, reading order, or assistive-technology usability. Keep human semantic and assistive-technology review as separate evidence and do not claim conformance from tags or one automated check.
- Java and veraPDF execution belong only in the isolated `validator-container` CI job; the Go adapter is CI/test-only, not product runtime. Do not run the validator provisioner or real CLI locally. Local Go tests may validate the adapter contract without invoking Java. Keep container network, user, resource, artifact, and pinned-validator boundaries intact.
- NeuroStrata and NeuroPlasticity are an accepted permanent cloud/environment limitation for this work. Do not retry model downloads, configure authentication, copy credentials, ingest or upload the repository, or claim Neuro-assisted execution unless a permitted operation genuinely ran. Follow the [build-machine tooling record](docs/planning/build-machine-tooling.md) for verified binary digests and bounded local probes; verify NeuroPlasticity by digest, never with its unsafe version flag, and do not broaden probe scope.
- CI security scans and repository protection checks are enforcement. Never bypass, disable, suppress, or weaken a failing check to get a change through. Do not claim a check passed unless its result is available for the exact revision.

## Git and review

- Preserve user work, hooks, and repository settings. Keep changes focused; do not rewrite shared history. Before publication inspect the staged diff and worktree.
- Publish only when explicitly authorized, using `scripts/source-publisher.sh` as described in [CONTRIBUTING.md](CONTRIBUTING.md). It enforces HTTPS destination and CloudBot Neo-green author/committer identity for authorized agent source commits. Use Conventional Commit subjects (`type(scope): imperative summary`). Do not use connector-created commits or ref updates, alter credentials, or bypass the source-publisher guard. Do not push or open a PR without authorization.
- Require independent review and current required CI checks before merge. Never merge on behalf of a reviewer or bypass branch protection.

## Documentation

- Keep docs concise, Markdown-only, and linked from the relevant index. Separate generation and remediation documentation. Do not invent schedules, budgets, measured performance, test results, or approved architecture.
- This guide uses the section organization of [Flow's AGENTS.md at a pinned revision](https://github.com/onflow/flow-go/blob/b85954c90eb9f8ab4b62a71641b2a53fb311a170/AGENTS.md) as a navigation reference; its project rules and wording are not adopted here.
