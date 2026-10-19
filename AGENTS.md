# Repository guidance

## Project boundaries

- This repository is planning scaffolding. Read [the docs index](docs/README.md), [product scope](docs/requirements/product-scope.md), [roadmap](docs/roadmap/product-roadmap.md), and relevant decision records before changing behavior. A candidate tool or architecture is not approved unless an accepted decision says so; mark unresolved choices **Proposed** or **Open**.
- Keep existing-PDF remediation and document generation as separate workflow lanes. Use synthetic or deidentified fixtures only; never add private reports or source documents.
- Go is the primary implementation language. Product dependencies must be permissively licensed open-source software; remediation dependencies must remain open source, and enterprise-only tagging or PDF/UA export is out of scope. Prefer the standard library for small project-specific work. Before adopting a product/runtime dependency, record exact direct and transitive licenses, maintenance/security posture, and requirement fit. Keep candidate libraries unapproved until reviewed. `pdfcpu` is currently used by preservation characterization tests and financial-fixture experiments, not product runtime.
- Keep packages focused. Pass `context.Context` for cancellation or deadlines, return useful errors, and document concurrency guarantees and ownership where they matter. Do not assume a method is concurrency-safe. Bound work, memory, file sizes, subprocess output, and concurrency where inputs or child processes can grow. Fuzz parsers and other hostile-input boundaries where a useful invariant can be asserted. Prefer small interfaces owned by their consumers when actual use justifies them; do not add abstractions speculatively. See [Go code review guidance](https://go.dev/wiki/CodeReviewComments), [error wrapping guidance](https://go.dev/blog/go1.13-errors), and [race detector documentation](https://go.dev/doc/articles/race_detector); race tests cover only executed paths.

## Work and tests

- Track work in Beads. Run `bd prime`, `bd ready`, and `bd show <id>` before changing an issue. Link prerequisites, keep acceptance criteria testable, and close only with evidence that each criterion passed. After issue changes, export and review `.beads/issues.jsonl`; it is an interchange export, not a full Dolt backup. See [CONTRIBUTING.md](CONTRIBUTING.md) for recovery and command details.
- For an explicitly authorized autonomous or repeated task run, work one issue or PR slice at a time and stop after at most six task iterations. After three unsuccessful attempts on the same slice, preserve partial work and report the blocker and next evidence needed. This is an escalation boundary; it never marks incomplete work complete, waives acceptance criteria, or authorizes background scheduling. Before work, inspect the current issue, relevant decisions, and prior evidence; after verification, record a concise sanitized result in canonical Beads or project documentation. Do not create a parallel backlog or store repository/Beads content in an external memory service. See the [pinned autonomous-platform fit assessment](docs/research/autonomous-engineering-platform-fit-2026-10-17.md); it is an analysis record, not runtime approval.
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
  (cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./... && go mod tidy -diff)
  bash -n scripts/bootstrap-cloud-build-tools.sh

  # CI-only PDF/UA-2 adapter module: Go process-contract tests only
  (cd ci/pdfua2/adapter && PDFUA_REAL_CLI_JAR= go test -race ./... -count=1 && go vet ./... && go mod tidy -diff)

  # Source-publisher guard tests
  bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh
  bash .github/scripts/test-source-publisher.sh
  ```

- The PDF/UA-2 target is accepted in [Decision 0001](docs/decisions/0001-pdf-ua-2-validation-target.md). Select the explicit `ua2` profile; never silently fall back. Machine validation does not establish correct meaning, reading order, or assistive-technology usability. Keep human semantic and assistive-technology review as separate evidence and do not claim conformance from tags or one automated check.
- Java and veraPDF execution belong only in the isolated `validator-container` CI job; the Go adapter is CI/test-only, not product runtime. Do not run the validator provisioner or real CLI locally. Local Go tests may validate the adapter contract without invoking Java. Keep container network, user, resource, artifact, and pinned-validator boundaries intact.
- Neuro model download and authentication setup are an accepted permanent cloud/environment blocker; do not install models, retry downloads, configure authentication, or copy credentials. Existing binaries may be used only for safe, documented local probes with scoped state and synthetic/public inputs; do not ingest or upload the repository. Verify NeuroPlasticity by digest, never with its unsafe version flag. Follow the [build-machine tooling record](docs/planning/build-machine-tooling.md) for verified binary digests and probe limits. Report exactly which local tool functions ran, and never imply model-backed or full Neuro-assisted execution without evidence.
- CI security scans and repository protection checks are enforcement. Never bypass, disable, suppress, or weaken a failing check to get a change through. Do not claim a check passed unless its result is available for the exact revision.

## Imported agent skills

Skills are optional workflow guides under [`.agents/skills/`](.agents/skills/). Load the linked `SKILL.md` and any references it names when its trigger fits the task; there is no mandatory bootstrap. Imported guidance is adapted for this repository and has lower priority than system/developer instructions, the current user's request, and this file. A skill never grants authority to take an action the user has not authorized.

- [TDD](.agents/skills/dap-test-driven-development/SKILL.md): use for implementation changes; preserve pre-existing/user work and follow the test rules above.
- [Systematic debugging](.agents/skills/dap-systematic-debugging/SKILL.md): use for failures or unexpected behavior before proposing a fix.
- [Verification before completion](.agents/skills/dap-verification-before-completion/SKILL.md): use before reporting completion, test, or check results.
- [Requesting code review](.agents/skills/dap-requesting-code-review/SKILL.md): use to prepare a scoped independent review with exact base/head evidence.
- [Receiving code review](.agents/skills/dap-receiving-code-review/SKILL.md): use to assess review findings against the current revision and repository policy.
- [Parallel agent dispatch](.agents/skills/dap-dispatching-parallel-agents/SKILL.md): use only for explicitly delegated or clearly requested parallel work; bound and verify each task.
- [Feature to PR workflow](.agents/skills/dap-feature-pr-pipeline/SKILL.md): use for a multi-step change intended to end in an independently reviewed pull request.
- [Parallel work lanes](.agents/skills/dap-overnight-parallel-lanes/SKILL.md): use only for explicitly requested independent lanes with deterministic gates and a coordinating owner.
- [Adversarial review panel](.agents/skills/dap-overseer/SKILL.md): use when multiple independent review perspectives or a structured panel are requested.
- [Review-to-plan integration](.agents/skills/dap-plan-review-integrator/SKILL.md): use when asked to integrate findings into an existing authorized plan.

### Locally authored

- [Scrubbing the public tree](.agents/skills/dap-scrubbing-public-tree/SKILL.md): use after agentic implementation work, before opening a PR or publishing, to clean agent and machine artifacts from the public tree.

The [skill provenance record](docs/planning/imported-agent-skills.md) identifies pinned sources, local adaptations, licenses, and the update process. Skills are Markdown-only; do not run upstream installers or utilities.

### Paperthin refinement

When the owner explicitly selects Paperthin for backlog refinement or planning, use relevant model-invoked skills from the reviewed `autonomous-engineering-platform` snapshot [`7354d9eabfd60e2f6891d34b22745fc5b997b147`](https://github.com/pratik-saptarshi/autonomous-engineering-platform/tree/7354d9eabfd60e2f6891d34b22745fc5b997b147), which pins [`LilMGenius/paperthin` at `7d5dc6235990230e45a24a25c527d81701da5458`](https://github.com/LilMGenius/paperthin/tree/7d5dc6235990230e45a24a25c527d81701da5458). Check each skill's invocation status and use only the relevant workflow; do not invoke user-only skills implicitly or install upstream software. Paperthin guidance supplements this repository's rules and cannot waive product-owner decisions, acceptance evidence, security controls, review, or publication gates.

## Git and review

- Preserve user work, hooks, and repository settings. Keep changes focused; do not rewrite shared history. Before publication inspect the staged diff and worktree.
- Publish only when explicitly authorized, using `scripts/source-publisher.sh` as described in [CONTRIBUTING.md](CONTRIBUTING.md). It enforces HTTPS destination and CloudBot Neo-green author/committer identity for authorized agent source commits. Use Conventional Commit subjects (`type(scope): imperative summary`). Do not use connector-created commits or ref updates, alter credentials, or bypass the source-publisher guard. Do not push or open a PR without authorization.
- Name branches `type/scope-short-description` (e.g. `fix/scripts-watchdog-signal-all-daemon-pids`), matching the Conventional Commit type and scope of the change. PR titles use Conventional Commit subjects as well.
- Require independent review and current required CI checks before merge. Never merge on behalf of a reviewer or bypass branch protection.
- Review and analyze every review comment across revisions against current code and acceptance criteria; do not rely on resolved labels. Record an evidence-backed fix or disposition. Fix critical or acceptance-blocking findings immediately. Defer valid lower-priority findings only through product review into Beads with rationale, priority, dependencies, testable acceptance criteria, and a source-comment link. Document unsupported findings with evidence. This does not waive approval, security, or release gates.

## Documentation

- Keep docs concise, Markdown-only, and linked from the relevant index. Separate generation and remediation documentation. Do not invent schedules, budgets, measured performance, test results, or approved architecture.
- This guide uses the section organization of [Flow's AGENTS.md at a pinned revision](https://github.com/onflow/flow-go/blob/b85954c90eb9f8ab4b62a71641b2a53fb311a170/AGENTS.md) as a navigation reference; its project rules and wording are not adopted here.
