# Adversarial panel: recorded `main` snapshot

**Review date:** 2026-10-12 UTC  
**Reviewed commit:** `595e7e9224fc6f22abb75e9f706dc2558514f620`  
**Commit subject:** `feat(corpus): add synthetic remediation manifest validation (#4)`  
**Review status:** source review, full Go test suite, selected adversarial mutations, and independent role reviews completed. External NeuroStrata memory and NeuroPlasticity evaluation paths remained constrained; details are below.

## Scope and exclusions

The reviewed tree is an isolated checkout of the fetched `origin/main` at the exact SHA above. PR #4's synthetic remediation manifest is included. PR #5 (`codex/go-generation-engine-evidence`, head `cb0dca5eb3ab8cac24548da30bd80b9c16cf53d3`) was open and excluded. The separate in-progress veraPDF integration work was also outside this pin. The PDF/UA adapter scaffold present at this pin was reviewed and tested with synthetic fake-process reports only.

The repository describes planning scaffolding plus CI/test-only validation components; no production PDF service was in scope. The panel covered hostile inputs and subprocess boundaries, PDF/UA-2 classification assumptions and accessibility claims, Go concurrency and resource controls, CI/supply-chain enforcement, and test adequacy and requirements traceability. No private/customer documents were used. No product code, fixtures, workflow policy, branch rules, or other repositories were changed.

## Panel charter and process

Three independent reviewer roles completed their review and sent findings for adjudication:

1. **Security and hostile-input reviewer:** adversarial XML, parser ambiguity, diagnostic disclosure, CI scanner bypasses, license policy variants, and process inputs.
2. **PDF accessibility and QA reviewer:** preservation-test coverage, catalog/accessibility metadata, semantic and assistive-technology evidence, and story/case traceability.
3. **Go runtime reviewer:** subprocess lifecycle, bounded I/O, cancellation, concurrency, resource controls, reproducibility, and false-positive assessment.

The chair then reproduced reported findings where feasible, challenged interpretations against the pinned source, recorded dissent, and assigned severity. Reviewer identities are intentionally not represented as evidence; this report records their roles and adjudication. NeuroStrata and NeuroPlasticity are build-machine tools, not shipped dependencies, and were not counted as independent reviewers.

### NeuroStrata and NeuroPlasticity

The approved provenance and isolation guidance is in [build-machine-tooling.md](../planning/build-machine-tooling.md). Official pinned binaries were installed outside the checkout and checksums matched the recorded official release metadata. Versions: NeuroStrata 1.9.0, NeuroPlasticity 2.1.0.

- NeuroStrata's local, model-free `status` command ran and reported no database file and no daemon. A namespace listing attempted to initialize its Nomic embedder and stopped at the configured proxy with HTTP CONNECT 403 for the model host. No repository or issue content was ingested, no database/namespace was created, and no daemon was started.
- NeuroPlasticity's `test … --print-egress-plan` ran on a disposable synthetic no-op manifest. It reported the embedded optimizer as local and allowed. This is an egress-policy inspection, not a code-review result. An actual test run could not proceed: no Podman runtime is installed, no container image is cached, and fetching an image was outside the approved network scope. Its normal failure path invokes an optimizer that could fetch the 2.5 GB Qwen model; no model was requested or downloaded. An initial no-op execution also stopped before container startup because its temporary manifest pointed `target_rules_file` outside the project; the manifest was in `/tmp` and was deleted after inspection.
- No hosted model/provider was used. No credentials were read, mounted, or expanded. Neither tool is claimed to have run a code review.

## Method

- Pin the fetched `origin/main` SHA and use a separate worktree.
- Read repository instructions, source, tests, workflows, security policy, PDF/UA decision records, corpus manifest, traceability catalog, and exported Beads records.
- Run the Go suite using the workspace-installed Go 1.27.2 toolchain, and use focused synthetic reproductions or temporary policy mutations. Remove all temporary tests/mutations afterward.
- Separate demonstrated defects from test-control gaps, deployment-dependent risks, untested assumptions, and rejected claims. Do not infer genuine veraPDF behavior or PDF/UA-2 conformance from fake-process tests.
- Do not repair overlapping code in this review. Record executable acceptance criteria and link follow-up tasks to the existing roadmap.

## Findings

### Confirmed defects and control gaps

#### F-01 — Ambiguous or empty validator XML can be classified as pass; parser errors expose output (P2)

`internal/pdfua/adapter.go:285-310, 342-394` decodes XML into Go structs and validates values after decoding. Go's XML decoder accepts duplicate attributes, taking a value for classification, and empty attributes can decode to zero-valued numeric pointers. A passing summary plus duplicate `isCompliant="false" isCompliant="true"` was classified `OutcomePass`; namespaced duplicate attributes also passed. Empty `failedToParse=""` and `failedJobs=""` counters also passed the synthetic report contract.

Separately, `internal/pdfua/adapter.go:230-233` wraps the underlying XML parse error. Synthetic malformed XML containing `SENSITIVE_MARKER` returned that marker in the public error; a malformed numeric attribute did the same. This contradicts the adapter comment that diagnostics omit child output and document content. Whether genuine veraPDF output can contain document-derived marker text is not established; the disclosure path is demonstrated with the child-output boundary treated as untrusted.

**Evidence and commands:** temporary fake-process tests under `internal/pdfua` altered only the synthetic report and were removed after execution. `go test ./internal/pdfua` passed the existing suite. The command is reproducible by adding the duplicate/empty attributes to `report("true")` in `internal/pdfua/adapter_test.go` and calling `Validate`; parser errors are returned through the normal adapter error path.

**Acceptance test:** reject duplicate attributes, including namespaced duplicates, and empty/invalid counters before classification; ensure errors are bounded and content-free. Add regression cases for both markers and retain valid pass/finding cases. Do not present this as a real-validator result.

#### F-02 — Relative temporary directories produce child paths interpreted from the wrong working directory (P2)

`internal/pdfua/adapter.go:130-137, 201-205` creates temporary paths from caller-provided `Options.TempDir`, passes those paths as JVM arguments, and also sets `cmd.Dir = tempDir`. With a relative temp directory, captured `-jar` argument was `review-relative-temp/pdfua-validation-…/validator.jar` while child cwd was that same relative subdirectory; resolving the argument from that cwd repeats the prefix. The absolute TempDir control succeeds under the fake process, but fake Java does not open the JAR, so actual JVM failure remains untested. The invocation is nonetheless constructed with a demonstrably wrong relative path.

**Acceptance test:** normalize the temp directory and all derived paths to absolute paths, or pass child-relative basenames; add a path-aware fake process that resolves and opens the passed JAR and input from `cmd.Dir`, for both relative and absolute TempDir.

#### F-03 — Go CI policy tests miss workflow bypass forms and a real license identifier (P2)

`internal/devsecops/policy_test.go` parses workflows into a policy model and enforces named controls, but the model does not include step `if` or job-level `continue-on-error`. On the pinned tree, each mutation below still passed `go test ./internal/devsecops`:

- Disable the full-history Gitleaks step with `if: ${{ false }}`.
- Disable the govulncheck step with the same condition.
- Set job-level `continue-on-error: true` on the Go security job.

The pinned go-licenses v2.0.1 executable reported `GPL-3.0` for a synthetic forbidden dependency. Its check returned exit 0 when `GPL-3.0` was appended to the allowlist, while the current policy regression coverage asserts `GPL-3.0-only` and the fixture's text. This is a regression-test blind spot; the current project dependency graph was not found to include that GPL fixture and no existing allowlist change was made.

**Acceptance test:** extend the workflow policy model/assertions to reject disabling conditions and job-level failure suppression; include the scanner's actual `GPL-3.0` identifier in the negative policy test. Keep the project allowlist review-based.

#### F-04 — Round-trip test omits the PDF catalog from its preservation assertions (P2)

`internal/pdfpreserve/roundtrip_test.go:27-62` checks page count/title, object IDs 3–16 (with streams checked separately), and sparse ParentTree behavior, but omits object 1, the catalog. The synthetic source catalog contains `/Lang`, `/MarkInfo`, `/StructTreeRoot`, and `/AcroForm` links. The current output retained those values in a read-only reproduction, but a future serialization regression changing the catalog would not fail this test.

**Acceptance test:** compare the catalog and assert the accessibility-relevant keys and form linkage survive; mutation tests that remove/change each key must fail. This is fidelity coverage, not PDF/UA conformance evidence.

#### F-05 — Generated traceability matrix drops acceptance cases linked only from tasks (P3)

The US-REM-002 story links task `dap-jo9.2.5.2`, whose closed, specification-only description names `TC-REM-HUMAN-001` and `TC-REM-ASSISTIVE-TECH-001`; the story's `Tests` list does not include them. `internal/traceability/catalog.go:503-507` renders the matrix only from `story.Tests`, so those task cases are absent from the generated story view. The validation path checks task cases separately but does not reconcile the two sets. This is a generated-view/traceability gap; it does not erase the human-review requirement.

**Acceptance test:** render the union of the story and linked-task cases, or validate explicit equivalence, and assert both human/AT cases appear under US-REM-002. Keep their status specification-only.

#### F-06 — Profile JSON schemas accept whitespace-only decision values rejected by Go (P3)

`internal/profile/profile.go:54-66` rejects decision values for which `strings.TrimSpace` is empty. Both lane profile schemas define decision `value` as a string with `minLength: 1` (`docs/workflows/generation/profiles/profile.schema.json:69-76`; same shape in remediation). A single space has length one and therefore meets that schema constraint while failing Go validation. Current schema parity coverage focuses on field-name whitespace, not decision values.

**Acceptance test:** make both schemas reject whitespace-only decision values across the same boundary set as Go, and add schema/Go parity cases.

#### F-07 — General Go CI job lacks an explicit timeout (P3 hardening)

`.github/workflows/go-checks.yml:12-31` has no job timeout. Security scan jobs do specify bounds. The suite currently passes quickly; the gap is a future hang/runner cost hardening issue, not an unbounded production service. The panel downgraded this from the initial P2 proposal: GitHub applies a finite platform maximum and Go's test command has its own default timeout.

**Acceptance test:** add an explicit reasonable job timeout and a policy assertion that rejects its removal or an invalid value. A hosted stalled-job exercise can verify the chosen bound.

### Plausible risks (not demonstrated as exploitable defects)

#### R-01 — Validator process lacks CPU and memory quotas (P2, deployment-dependent)

`internal/pdfua/adapter.go:105-112, 201-220` applies input/output limits, a wall-clock timeout, and process-tree cancellation. It does not impose CPU scheduling or memory limits on Java. A pathological PDF could still consume resources before termination. No OOM/CPU exhaustion was reproduced, and the adapter was not wired to production at the pinned commit.

**Acceptance test:** before accepting hostile PDFs in shared workers, run synthetic complexity/resource cases under enforceable per-process/container CPU and memory bounds and verify termination, cleanup, and host availability.

#### R-02 — Java runtime identity is caller-asserted (P3, deployment-dependent)

The adapter requires an absolute executable path and verifies the JAR hash, but `ValidatorVersion` is a caller-supplied string. No Java binary identity/version check is performed. No executable substitution was demonstrated and the adapter was not wired into CI at this pin.

**Acceptance test:** define the trusted Java distribution/source and verify observed runtime identity in the build image before enabling the adapter in a job.

### Untested assumptions and assurance limits

- No authenticated veraPDF 1.30.3 execution or genuine known-pass/known-fail PDF oracle was run. Synthetic reports only exercise the adapter parser.
- No human semantic or assistive-technology review was performed. The repository correctly keeps these as separate gates; no false PDF/UA-2 or WCAG conformance claim was found.
- No concurrency defect was confirmed in the bounded output budget/process tree implementation. The focused race suite passed.
- Resource limits for memory/CPU remain a design assumption until the adapter is used on hostile files.
- GitHub reported seven successful checks on the preceding published PR head, but branch-protection settings could not be read through the connected integration. Green checks are observed; their formal required status is unverified.
- The opened PR #5 and the separate veraPDF integration work were outside this immutable review SHA.

### Rejected interpretations and dissent

- **Round-trip baseline is mutated by the writer:** a reproduction snapshotted object-table entry 1 before `api.WriteContextFile`; it was unchanged afterward. This part of the accessibility reviewer concern is a false positive. The distinct omission of catalog assertions remains confirmed.
- **No corpus artifacts proves the corpus validator is broken:** rejected. The pinned manifest explicitly records planned synthetic cases, not product execution evidence.
- **The adapter or tagged fixture proves PDF/UA-2 conformance:** rejected. No real validator or human/AT evidence is present in this panel.
- **The Go job is unbounded:** rejected as stated; P3 reflects a missing explicit job bound while retaining GitHub's finite default.
- **No concurrency bug means resource bounds are complete:** rejected. Race safety and memory/CPU quotas are separate questions.
- No independent reviewer dissent was recorded after adjudication. The panel's limitations remain the missing real validator, human/AT evidence, and constrained model/runtime setup above.

## Prioritized remediation work (Beads)

Six tasks were created in an isolated Beads v1.3.1 workspace, imported from the pinned 44-issue JSONL snapshot, then exported as 50 issues. The updated export is included in this report branch. No Dolt remote was configured, no sync/push was performed, and Beads anonymous metrics were disabled.

| Priority | Beads | Work |
| --- | --- | --- |
| P2 | `dap-jo9.2.5.5` | Reject ambiguous/empty validator XML attributes and redact parse diagnostics; coordinate with active adapter task `dap-jo9.2.5.3`. |
| P2 | `dap-jo9.2.5.6` | Resolve validator temp/JAR/input paths before changing child cwd; add a path-aware test. |
| P2 | `dap-jo9.8` | Detect disabled security steps, job-level `continue-on-error`, and the scanner-emitted GPL-3.0 identifier in policy tests. |
| P2 | `dap-jo9.2.3.3` | Assert PDF catalog, language, MarkInfo, structure root, and AcroForm preservation. |
| P3 | `dap-jo9.9` | Reconcile linked task cases in matrix and align whitespace validation between schemas and Go. |
| P3 | `dap-jo9.10` | Add a bounded timeout to the general Go CI job and policy coverage. |

The authentic veraPDF/oracle work remains in the existing open `dap-jo9.2.5.4`. Human/AT review is tracked under open parent `dap-jo9.2.5`; `dap-jo9.2.5.2` is closed and specification-only, not an implementation task. The in-progress adapter implementation task is `dap-jo9.2.5.3`; findings that overlap it are follow-up proposals only.

## Commands, versions, and outcomes

All source commands below were run in `/workspace/review-probes-595e7e9`, a disposable checkout at the exact reviewed commit. Temporary tests/mutations were removed; the checkout was clean afterward. Tool downloads and Beads database were outside the reviewed checkout.

| Command/tool | Outcome |
| --- | --- |
| `git rev-parse HEAD` | `595e7e9224fc6f22abb75e9f706dc2558514f620`. |
| `/workspace/.toolchains/go1.27.2/bin/go version` | `go version go1.27.2 linux/amd64`. |
| `go test -race -timeout=120s ./...` | Passed all packages: corpus, devsecops, pdfpreserve, pdfua, profile, traceability; command package has no tests. |
| `go vet ./...` | Passed. |
| `go test -race -timeout=120s ./internal/pdfua ./internal/profile ./internal/devsecops ./internal/traceability ./internal/pdfpreserve` | Passed all five targeted packages. |
| Temporary fake-process XML probes; `go test ./internal/pdfua` | Confirmed duplicate/empty attributes classified pass and malformed output markers surfaced in errors; existing package suite passed. Probe file removed. |
| `go test ./internal/pdfua -run TestReviewRelativeTempDirArgResolution -v` (temporary test) | Captured relative `-jar` arg alongside relative `cmd.Dir`; derived effective path repeats the temp prefix. Absolute TempDir fake-process control passed. No JVM was run. |
| Temporary round-trip probe | Writer did not mutate pre-write catalog snapshot; output catalog matched source on current fixture. Confirms baseline-mutation allegation false positive and shows catalog's present-state values. Probe removed. |
| Four temporary policy mutations + `go test ./internal/devsecops` | Tests passed despite Gitleaks disabled, govulncheck disabled, job-level continue-on-error, and GPL-3.0 allowlist addition. Mutations restored. |
| `go-licenses check --include_tests --allowed_licenses=Apache-2.0,BSD-3-Clause,GPL-3.0,MIT ./...` on synthetic GPL-3.0 module using pinned go-licenses v2.0.1 | Scanner emitted `GPL-3.0`; allowing that emitted identifier returned exit 0. Synthetic module/cache outside repository. |
| NeuroStrata 1.9.0 `status` | Passed model-free; no DB file and daemon down. `namespaces` attempted model initialization then proxy CONNECT 403; no retry. |
| NeuroPlasticity 2.1.0 `test /tmp/neuroplasticity-review.json --print-egress-plan` | Egress plan printed: embedded optimizer is local/allowed. No-op run did not proceed to evaluation because manifest's target rule path was outside the project. |
| NeuroPlasticity no-op `test` attempt | Stopped before container startup with the safe-relative-target-rules error. No Podman binary or cached container images are available; no image pull/model download attempted. |
| Beads v1.3.1 isolated import/create/export | Imported 44 snapshot issues; created six linked follow-ups; exported 50. Metrics off; no remote configured and no sync run. |
| `git diff --check` | Passed after report/export update. |

Versions: Git 2.52.0, Go 1.27.2, Beads 1.3.1, NeuroStrata 1.9.0, NeuroPlasticity 2.1.0, Docker Engine 28.4.0. No Podman runtime or container images were available. Exact executable release hashes and official provenance are recorded in [build-machine-tooling.md](../planning/build-machine-tooling.md).

### Published PR check evidence

PR #6 is a draft report-only review branch. The preceding published head `585a483dfd834b46d11a890c2f4cc67bc639b1be` had seven successful check runs: CodeQL, secrets, checks, go-security, Review dependency changes, Analyze (actions), and Analyze (go). Branch-protection API access returned 403, so the report does not claim those checks are formally required. This updated report/export must itself be pushed and its checks rerun. No merge is authorized until independent human review and all seven required checks are confirmed green.

## Review boundary and publication

This report adjudicates only `595e7e9224fc6f22abb75e9f706dc2558514f620`. PR #5 and the separate veraPDF integration changes were outside the pin. The results distinguish executed tests from static policy gaps, synthetic parser behavior, plausible resource risks, and untested genuine validator/human accessibility evidence. Keep PR #6 in draft for independent review.
