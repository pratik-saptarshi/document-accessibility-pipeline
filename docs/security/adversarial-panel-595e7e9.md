# Adversarial panel: recorded `main` snapshot

**Review date:** 2026-10-12 UTC  
**Reviewed commit:** `595e7e9224fc6f22abb75e9f706dc2558514f620`  
**Commit subject:** `feat(corpus): add synthetic remediation manifest validation (#4)`  
**Scope status:** source and workflow review completed; executable Go and external-tool review blocked as detailed below.

## Scope and exclusions

The reviewed tree was created in a separate worktree from the fetched `origin/main` ref. Its commit includes PR #4's synthetic remediation manifest work. PR #5 (`codex/go-generation-engine-evidence`, head `cb0dca5eb3ab8cac24548da30bd80b9c16cf53d3`) was open and not part of this snapshot. The separate in-progress veraPDF integration work was also outside the pinned commit. The `internal/pdfua` scaffold present at this pin is included in the static review.

The repository describes itself as planning scaffolding plus a CI/test-only validator adapter; there is no production PDF processing service to test. The review covered repository code and configs for hostile inputs, subprocess handling, PDF/UA-2 claims, Go concurrency/resource bounds, GitHub Actions and dependency controls, and test/requirements traceability. No private or customer documents were used. No production code, fixtures, dependency pins, workflow policy, branch rules, or other repositories were changed.

## Panel charter and execution

The review was organized around these adversarial roles:

1. **Security and hostile-input reviewer:** attacker-controlled PDF bytes, XML output, paths, symlinks/FIFOs, timeouts, output floods, subprocess descendants, and failure disclosure.
2. **PDF/UA-2 and accessibility reviewer:** profile selection, parsing/classification assumptions, real-validator evidence, semantic correctness, and assistive-technology release gates.
3. **Go runtime reviewer:** cancellation, goroutine/process lifecycle, I/O bounds, memory/CPU bounds, and failure cleanup.
4. **CI and supply-chain reviewer:** workflow permissions, action/tool provenance, secret-scan suppression paths, dependency/license checks, and job limits.
5. **QA and traceability reviewer:** requirements-to-cases-to-Beads links, executable versus specification-only tests, adversarial coverage, and reproducibility.

These are reviewer lenses, not five independent people or agents. This environment had no NeuroStrata or NeuroPlasticity binaries or integration, so neither tool ran and they supplied no reviewer output, dissent, or adjudication. The panel outcome below is one manual source review organized by role, followed by a moderator rebuttal/adjudication. It must not be represented as independent-tool review.

### NeuroStrata and NeuroPlasticity availability

The repository's approved setup/provenance record is [`build-machine-tooling.md`](../planning/build-machine-tooling.md). It places the previously verified NeuroStrata 1.9.0 and NeuroPlasticity 2.1.0 binaries outside the checkout, keeps their runtime data local, forbids ingesting the repository/Beads/private data, and records model and authentication gates. In this fresh environment, `/workspace/.codex-tools` and `/workspace/.neurotools-home` did not exist; `neurostrata-mcp`, `neurostrata`, and `neuroplasticity` were not found on `PATH`; no matching tool integration was available. No help/version invocation was treated as review. No model or executable was downloaded, no credentials were accessed, and neither tool is claimed to have run.

## Method

- Pin the fetched `origin/main` commit and review in an isolated worktree.
- Inspect repository instructions, scoped source, tests, workflows, security policy, PDF/UA decision record, fixture manifest, and traceability/Beads export.
- Trace hostile-input checks to implementation and tests, while distinguishing source assertions from tests actually executed here.
- Record direct commands and outcomes; use line-level evidence for findings. Do not infer real veraPDF behavior or PDF/UA-2 conformance from fake-process/XML tests.
- Separate confirmed control gaps, plausible risks, untested assumptions, and disproved/false-positive interpretations. No source fixes were made because this is a review task and any code remediation should first be presented as a scoped task.

## Findings

### Confirmed control gap

#### F-01 — General Go CI job has no explicit time limit (P2 / medium)

`.github/workflows/go-checks.yml:12-31` defines a job that runs `gofmt`, `go mod tidy`, `go vet`, and the race-enabled test suite without `timeout-minutes`. Unlike `.github/workflows/security-scans.yml`, which sets job timeouts at lines 19 and 60, the general Go checks can occupy a hosted runner until the platform maximum if a test or tool hangs. The checks are read-only and checkout credentials are disabled, so this is a CI availability/cost exposure, not evidence of repository-secret exposure.

**Reproducer/evidence:** inspect `.github/workflows/go-checks.yml:12-31`; no timeout key is present at job or step scope. This is a static policy finding, not a timed exploit demonstration.

**Acceptance test:** add a workflow-policy test that fails if the Go checks job lacks an explicit, bounded timeout. Confirm the timeout is applied by GitHub Actions on a deliberately stalled test in a disposable branch or equivalent workflow harness. Choose the actual limit with maintainers; no measured runtime threshold is asserted here.

### Plausible risks (not demonstrated as exploitable defects)

#### R-01 — Validator subprocess has no CPU or memory quota (P2 / medium, unverified impact)

`internal/pdfua/adapter.go:105-106` requires caller-provided timeout, input, and combined output limits; lines 111-112 apply a wall-clock deadline, and line 201 starts Java with `exec.CommandContext`. These controls bound duration and bytes copied/emitted, but do not cap process memory or CPU scheduling. A hostile or pathological PDF could cause a Java/veraPDF allocation spike before the timeout, with potential runner OOM impact. The repository is not yet a production PDF service and the real validator was not run, so impact is plausible rather than demonstrated.

**Acceptance test:** after the execution boundary and runner policy are selected, exercise synthetic high-complexity PDFs under enforceable per-process/container memory and CPU limits; demonstrate that the worker is terminated, the host remains available, temporary files are removed, and the result is classified as a bounded failure. Keep the limit outside caller-controlled PDF metadata.

#### R-02 — Java executable identity is not pinned by the adapter contract (P3 / low-to-medium, deployment-dependent)

The adapter validates that `Options.Executable` is an absolute path (`internal/pdfua/adapter.go:92-94`) and validates a SHA-256 for the copied veraPDF JAR (`:98-103`, `:157-178`), but it neither authenticates nor hashes the Java executable. `ValidatorVersion` is a non-empty caller-supplied string (`:95-96`), not an observed runtime version. If the configured path or CI image is outside the trusted boundary, execution may use an unexpected Java runtime. No such path substitution was reproduced, and the adapter is not wired into the workflow at this pin.

**Acceptance test:** define the trusted Java distribution/version and acquisition source; capture the observed executable version and verify its installed package/image identity or digest in CI. Add tests for a wrong runtime identity and for executable-path replacement/symlink policy before enabling the adapter in a job.

### Untested assumptions and assurance gaps

#### U-01 — Real veraPDF XML and PDF/UA-2 classification have not been exercised

`internal/pdfua/adapter.go:201` explicitly passes `--flavour ua2 --format xml`; `parseReport` at `:285-381` requires a single job, profile name, compliance flag, and batch counters. Existing tests use a fake executable and synthetic XML (`internal/pdfua/adapter_test.go`), and fake PDF bytes are explicitly labeled non-oracle. No authentic veraPDF artifact or genuine positive/negative PDF/UA-2 controls were run at this review. Thus compatibility between the expected XML shape/counters and veraPDF 1.30.3 remains an untested assumption. The record in `docs/decisions/0001-pdf-ua-2-validation-target.md:17-26` already describes the signature-key and real-oracle gates.

**Acceptance test:** complete the existing `dap-jo9.2.5.4` gate using an authenticated official artifact, explicitly run `ua2`, hash and retain genuine pass/fail controls, and test the parser against that tool's real reports. Keep human semantic and assistive-technology review separate.

#### U-02 — No PDF/UA-2 conformance or end-user accessibility result exists

The current files include profile contracts and planned corpus metadata, not generated/remediated PDF artifacts. The remediation manifest marks all four cases `planned-not-created` and `not-run` (`docs/fixtures/remediation/manifest.v1.json:18-95`). The 26 entries in `docs/traceability/catalog.json` are all `specification-only, not executed` (confirmed by JSON inspection). This is an evidence gap, not a defect in a planning-only repository. No machine, semantic, reading-order, screen-reader, or assistive-technology result can be inferred from tagged-output scaffolding or these specifications.

**Acceptance test:** follow the existing `dap-jo9.2.1`, `dap-jo9.2.5`, and `dap-jo9.2.5.2` gates with synthetic documents for all four families; record the candidate hashes, independent machine results, human semantic findings, and assistive-technology outcomes. Do not claim PDF/UA-2 based on the adapter alone.

#### U-03 — Go behavior and test suite did not execute in this environment

The repository requires Go 1.27.2 (`go.mod:1-4`), but `/usr/bin/go` in this environment is a board-game application. `go version` returned `Go: Unknown option: version`; direct `go test` and `go vet` invocations also returned `Unknown option`. The Docker daemon is present but has no cached images; no image was pulled. Consequently, the source-level review of process cleanup, bounded buffers, XML parsing, profile validation, corpus validation, and traceability is not corroborated by an executed Go suite here.

**Acceptance test:** run the prescribed checks in a Go 1.27.2 environment: `gofmt -l .`, `go mod tidy -diff`, `go test -race ./...`, `go vet ./...`, plus the workflow security tools. Record exact output and status before treating the review as fully reproducible.

#### U-04 — Tool policy and live branch protection/check requirements are not established locally

The source pins the tools and scanner steps in `.github/workflows/security-scans.yml`, but `actionlint`, `govulncheck`, `go-licenses`, and `gitleaks` were unavailable here. The GitHub connector returned an empty combined-status list for the pinned main commit and no PR-triggered runs for that SHA. That does not establish whether repository branch protection currently requires seven checks or whether CodeQL default setup is active; the checked-in `codeql.yml` and repository prose are not live settings evidence.

**Acceptance test:** execute the complete CI workflow or inspect the PR's seven required checks and current GitHub rules using authorized repository settings. Confirm CodeQL default-setup runs in GitHub. Do not infer active enforcement from YAML alone.

### False positives / rebuttals

- **“The repository claims PDF/UA-2 conformance because it has a validator adapter.” Rejected.** Package comments, decision records, and validation documentation explicitly distinguish machine-only adapter outcomes from semantic and assistive-technology approval.
- **“No PDF fixtures means the corpus validator is broken.” Rejected.** The pinned manifest expressly defines planned synthetic cases and says no artifacts/results exist; this is the accepted planning state.
- **“The security scan is unbounded.” Rejected for that workflow.** `security-scans.yml` sets job timeouts. F-01 is limited to the separate general Go checks job.
- **“The disabled CodeQL workflow proves CodeQL is absent.” Rejected.** Repository docs distinguish the disabled YAML workflow from the claimed GitHub default setup. Live configuration/runs remain unverified here.
- **“PR #5 is part of this reviewed main snapshot.” Rejected.** GitHub reported it open and based on the exact reviewed main SHA; it was excluded from the pin.

### Dissent

No independent reviewer dissents were collected: the specialized tools and separate reviewer agents were unavailable. The absence of dissent is not consensus evidence.

## Rebuttal and adjudication

The strongest counterargument to F-01 is that the general suite is small and likely to finish quickly. No runtime measurement could be made here, and that does not protect against a future deadlock or pathological test. The missing bound is concrete and the security workflow already demonstrates a bounded-job convention. Adjudication: retain F-01 at P2 as a narrowly scoped CI availability gap.

For R-01 and R-02, current code is explicitly CI/test-only, limits input/output/time, and is not wired into production or even the current workflow. There is no evidence of a present exploit. Adjudication: retain them as deployment-dependent design risks, not confirmed vulnerabilities; require resolution before untrusted PDFs reach this adapter in a shared/production worker.

For U-01 and U-02, the existing documentation already states the restrictions accurately. Adjudication: no PDF/UA-2 false claim was found; the review itself cannot provide missing external-tool or human-test evidence. Keep release/conformance gates open.

## Prioritized remediation work

These are concrete task proposals. The Beads CLI and embedded database were absent in this worktree, so no Beads issue was created or modified and no ID is invented. Existing related Beads issues are noted for reuse.

1. **P2 — Bound general Go CI execution.** Add an explicit `timeout-minutes` to the `checks` job and a policy regression test. Acceptance: workflow lint/policy tests reject a missing or unbounded timeout; a disposable stalled-job exercise shows cancellation at the configured limit. *New Beads issue not created: `bd` unavailable.*
2. **P2 — Define and enforce subprocess resource limits before accepting hostile PDFs.** Add a design/implementation slice for CPU and memory limits, process count, filesystem write scope, and kill/cleanup behavior around Java/veraPDF. Acceptance: synthetic resource-exhaustion controls fail closed without destabilizing the runner. *New Beads issue not created: `bd` unavailable.*
3. **P2 — Close the genuine validator and accessibility evidence gate.** Continue existing `dap-jo9.2.5.4` for artifact authentication/oracle controls and `dap-jo9.2.5.2` for human semantic/assistive-technology evidence; do not duplicate those issues. Acceptance is as specified under U-01/U-02.
4. **P3 — Pin the Java runtime identity in CI.** Before enabling the adapter, record the official runtime source/version and verify the actual executable identity, not only the JAR. *New Beads issue not created: `bd` unavailable.*

## Commands, versions, and outcomes

Commands were run from the isolated worktree at the pinned commit on 2026-10-12. The available host versions were Git 2.52.0, Python 3.12.14, GNU Bash 5.2.37, and Docker Engine 28.4.0. `/usr/bin/go` was not the Go toolchain.

| Command | Outcome |
| --- | --- |
| `git fetch --no-tags origin refs/heads/main:refs/remotes/origin/main` | Passed; fetched main at `595e7e9224fc6f22abb75e9f706dc2558514f620`. |
| `git worktree add -b review/adversarial-panel-595e7e9 /workspace/review-adversarial-595e7e9 origin/main` | Passed; isolated review branch/worktree created. |
| `git status --short --branch` | Passed; clean at pin before report edits. |
| `git --version` | `git version 2.52.0`. |
| `python3 --version` | `Python 3.12.14`. |
| `bash --version \| head -1` | `GNU bash, version 5.2.37(1)-release`. |
| `docker info --format '{{.ServerVersion}}'`; `docker images --format ...` | Engine reported `28.4.0`; no cached images were listed. No pull/run attempted. |
| `go version` | Failed: `Go: Unknown option: version` (the `go` command is a board game). |
| `go test -race -timeout=120s ./...` | Failed before Go execution: `Go: Unknown option: test`. Suite not run. |
| `go vet ./...` | Failed before Go execution: `Go: Unknown option: vet`. Vet not run. |
| `go mod tidy -diff` | Failed before Go execution: `Go: Unknown option: mod`. |
| `bash .github/scripts/check-gitleaks-suppressions.sh` | Passed: `No unmanaged Gitleaks config or suppression files found.` |
| `bash .github/scripts/test-gitleaks.sh` | Failed: `Gitleaks rejected the clean synthetic fixture.` Gitleaks is absent, so this is not a scan result. |
| `bash .github/scripts/test-go-licenses.sh` | Failed: `go-licenses: command not found`. |
| `actionlint` | Not run: command not found. |
| `govulncheck -test ./...` | Not run: command not found. |
| `bd prime`; `bd ready` | Not run: `bd: command not found`. |
| `git diff --check` | Passed at review-document creation checkpoint. |
| Python JSON read of `docs/traceability/catalog.json` and `.beads/issues.jsonl` | Passed: 17 requirements, 6 stories, 17 functions, 2 task groups, 20 catalog tasks, 26 specification-only acceptance cases, 44 exported Beads issues. |

`neurostrata-mcp`, `neurostrata`, `neuroplasticity`, `verapdf`, `actionlint`, `govulncheck`, `go-licenses`, and `gitleaks` were not found on `PATH`. No model, JAR, container image, or alternate CLI was fetched. No credentials or persistent access were expanded.

### Post-publication PR check evidence

For the first published PR head `46bd7173fbde18496432ef61bd34ec47e9ce56ca`, GitHub reported seven completed check runs, all with conclusion `success`: `CodeQL`, `secrets`, `go-security`, `checks`, `Review dependency changes`, `Analyze (actions)`, and `Analyze (go)`. The three action workflows also completed their constituent steps successfully. The branch-protection endpoint returned HTTP 403 to the connected integration, so this records the seven observed check runs and does not claim independent verification of the branch-protection configuration. The PR remains a draft pending independent human review.

## Review boundary and publication

This report concerns exactly `595e7e9224fc6f22abb75e9f706dc2558514f620`. It excludes PR #5 and the separately pending veraPDF integration changes. The report is static/manual evidence plus the commands above; it is not a completed independent NeuroStrata/NeuroPlasticity panel and not an executable Go test result. Do not merge any report PR until an independent reviewer has reviewed it and all seven required checks are confirmed green.
