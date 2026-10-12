# Adversarial Go quality and performance panel

**Review date:** 2026-10-12 UTC

**Pinned source:** 197ba24596d105971b4d0d88b2dfab3edd724156 (fix(bootstrap): clean up child processes on signals and deadlines, after PR19)
**Scope:** Code quality, Go API/error design, maintainability, allocations, scaling, concurrency, lifecycle, resource use, all three Go modules, and CI tooling. This is a quality/performance audit; no product remediation was made.

## Executive summary

No high-priority product defect or concurrency failure was reproduced. The current tree is planning and CI scaffolding, not an end-to-end document service. The CI-only PDF/UA report verifier and traceability planner both have explicit input bounds and tests. The main actionable correctness gap is narrow: traceability.Validate does not include task-group IDs in its duplicate-ID check. A bounded fixture showed it accepts both duplicate group IDs and a task/group ID collision.

The synthetic performance runs expose repeated relationship scans in validation and Markdown generation. At the current catalog size (17 requirements, 6 stories, 17 functions, 20 tasks, 26 cases, 2 groups), those scans are not a practical bottleneck. The 8 MiB XML ceiling costs about 0.12–0.13 s and about 33.6 MB allocated per call on this VM. These are parser/bundle measurements, not PDF validation throughput.

Two low-priority resource/setup follow-ups are worth tracking: bootstrap subprocess output is collected with unbounded CombinedOutput, and the bootstrap shell runs the full eight-entry verifier once during bootstrap and again in a separate verify pass. Neither was measured as an incident. No remediation or optimization rewrite is recommended by this review.

## Panel method

The panel used five review lenses: idiomatic Go/API/error design; architecture, duplication, and test quality; allocation/throughput/scaling; concurrency/lifecycle/resource use; and an independent challenger separating measured evidence from hypotheses. Independent parent-supplied reviewer observations were incorporated where applicable. Reviewers distinguished the planning checker and CI/test adapters from product paths.

The exact SHA was fetched by object ID and reviewed in a detached worktree. The original work branch was not switched, moved, or published. Temporary benchmark and probe files were removed from that worktree after measurement; the reproducibility sources are retained under support/ as .txt files so they do not join the Go packages. For publication, the audit branch is based on freshly fetched origin/main at 9bb9f2a202632fa92359d288ab298853c89434e7 because the publisher guard requires new branches to descend from main. The diff from the review SHA to that parent is limited to the Beads export, a research note, and a test file; there are no product implementation, go.mod, or go.sum changes. This does not change the review cutoff or measured SHA. A separately discussed no-checkout bootstrap benchmark pinned to source archive 8f0072c885fad7811635523ced39df0f1a23ca33; those timings are not attributed to this 197ba245 review.

No .agents skills were present in the pinned tree. The installer-managed tool directory is /workspace/.codex-tools/cloud-build-tools and is not on the shell default PATH. Its Go 1.27.2, bd 1.3.1, actionlint 1.7.12, govulncheck 1.9.0, go-licenses 2.0.1, and Gitleaks 8.30.1 binaries passed the read-only cloud-tools-bootstrap --verify check. bd is installed, but bd prime/ready could not use this checkout: there is no local Dolt database (the repository has only the tracked issues.jsonl export), and bd prime also hit the read-only /home/agent/.dolt path. No Beads issue was changed, exported, or closed; the follow-ups below remain proposals to avoid altering the parallel Beads reconciliation. No Java or veraPDF executable was run, and no CLI sign-in or credential state was changed. Neuro remains the accepted permanent cloud limitation; no setup, model, or authentication work was attempted.

## Environment and verification

- Host: Linux x86_64 VM, AMD EPYC 9V74, 5 online vCPUs, approximately 18 GiB memory available at capture time.
- Toolchain: official golang:1.27.2 image (linux/amd64; image digest sha256:5bc7f572bbaa98885a3a1fd9c0aa76b59e3e14e8628bfc316bbfd0c701e4818c). GOTOOLCHAIN=local; CGO enabled. The shell-default /usr/bin/go is not a usable Go toolchain; the same Go 1.27.2 is available at /workspace/.codex-tools/cloud-build-tools/toolchains/go/1.27.2/bin/go, outside default PATH. Project tests and benchmarks ran in the pinned image. The environment CA was mounted there for fetching existing go.sum modules; no dependency versions or module files changed.
- Root module: go test -race ./..., go vet ./..., and go mod tidy -diff passed.
- cmd/cloud-tools-bootstrap: go test -race ./..., go vet ./..., and go mod tidy -diff passed.
- ci/pdfua2/adapter: go test -race ./..., go vet ./..., and go mod tidy -diff passed. The authenticated real-validator test remained skipped because PDFUA_REAL_CLI_JAR was not set.
- CI tooling checks: actionlint, bash -n scripts/*.sh .github/scripts/*.sh, test-source-publisher.sh, test-gitleaks.sh, test-go-licenses.sh, and the bounded synthetic test-verapdf-output-bound-shim.sh passed. The latter caps its synthetic output fixture at 16 MiB and verifies the 8 MiB capture boundary. The first go-licenses smoke-test attempt could not initialize the default read-only /home/agent/.cache/go-build; rerunning with GOCACHE and GOMODCACHE under /tmp passed. The scanner test used local synthetic modules and fixtures. On the publication parent, go run ./cmd/traceability-check passed with 17 requirements, 6 stories, 20 tasks, and 26 specification-only cases. Both saved benchmark harnesses compiled and completed a one-iteration smoke run; the group-ID probes passed with no diagnostics.
- gofmt was applied to the temporary benchmark/probe source before preserving it. No product source was reformatted.
- Static CI workflow review found root and bootstrap checks in go-checks.yml, adapter race tests in the separate veraPDF integration workflow, and pinned scanner/tool installs in security-scans.yml. govulncheck was not run because its configured vulnerability database is external; no network settings were changed. CI-only Java/veraPDF execution was left to CI.

## Measured benchmarks

Commands ran against this exact tree using the temporary harness in [support/pdfua_benchmark_test.go.txt](support/pdfua_benchmark_test.go.txt) and [support/traceability_benchmark_test.go.txt](support/traceability_benchmark_test.go.txt). Fixtures were generated and validated before timers started. Samples were serial, with five repeats per case, -benchmem; XML used -benchtime=500ms, traceability used -benchtime=300ms.

Reproduction after copying each support file to the indicated package as review_benchmark_test.go:

    go test -run '^$' -bench 'BenchmarkReview(ParseReport|VerifyBundle)' -benchmem -benchtime=500ms -count=5 ./internal/pdfua
    go test -run '^$' -bench 'BenchmarkReview' -benchmem -benchtime=300ms -count=5 ./internal/traceability

### Report parser and bundle verifier

The XML fixture is a valid single-job report with comment padding to the exact target size. It holds XML element count constant, so it exercises large text payloads rather than realistic validator issue density. VerifyBundle also computes the report SHA-256 before calling ParseReport.

| Input size | ParseReport median (range) | B/op; allocs/op | VerifyBundle median (range) | B/op; allocs/op |
|---:|---:|---:|---:|---:|
| 1 KiB | 28.5 µs (27.5–31.1) | 11,752; 172 | 39.5 µs (38.8–42.3) | 18,373; 252 |
| 64 KiB | 1.045 ms (1.026–1.065) | 269,801; 184 | 1.074 ms (1.070–1.088) | 276,844; 265 |
| 1 MiB | 16.43 ms (16.13–16.62) | 4,201,983; 192 | 16.93 ms (16.41–19.46) | 4,214,952; 287 |
| 8 MiB | 124.26 ms (122.44–130.70) | 33,562,107; 198 | 127.28 ms (125.84–131.53) | 33,575,076; 293 |

ParseReport does a token-validation pass and then a decoding pass (internal/pdfua/report.go:15–20, 130–160). VerifyBundle additionally hashes the report and caps it at 8 MiB (internal/pdfua/bundle.go:43–50, 77–84). The measured allocation is roughly four times the report size at the maximum fixture. The cap and duplicate-attribute defense are intentional. At this limit and rate, the panel found no immediate reason to change the parser. These measurements do not model veraPDF runtime, PDF complexity, container execution, or end-to-end throughput.

### Traceability validation and Markdown generation

Each valid synthetic catalog had N requirements, stories, functions, tasks, and acceptance cases, with one-to-one links and valid synthetic Beads issues. It included lane/family coverage. Validation ran once before timing; Markdown generation emitted and checked all four views.

| Entities in each linked collection | Validate median (range) | B/op; allocs/op | GeneratedMarkdown median (range) | B/op; allocs/op |
|---:|---:|---:|---:|---:|
| 100 | 0.283 ms (0.266–0.294) | 150,638; 230 | 0.351 ms (0.343–0.376) | 414,134; 3,560 |
| 1,000 | 15.45 ms (15.18–16.00) | 1,820,092; 2,080 | 15.19 ms (14.90–15.32) | 4,251,654; 35,105 |
| 5,000 | 376.79 ms (374.56–401.63) | 7,253,616; 10,252 | 326.85 ms (317.12–339.31) | 22,507,104; 175,151 |

Validate nests the requirement/story scan and then scans functions and tasks for each linked story (internal/traceability/catalog.go:310–343). renderMatrix scans all functions/tasks for each story and storyAcceptanceCases rescans tasks (internal/traceability/catalog.go:482–507, 513–529). This explains the measured near-quadratic growth in the synthetic one-to-one case. The real catalog is much smaller and the checker is a local planning tool; this is a scaling limit to track, not a current product-path defect.

## Findings and proposed Beads follow-ups

### P3 — Task-group IDs are omitted from duplicate-ID validation

**Evidence:** Validate adds capabilities, requirements, stories, functions, tasks, and tests to allIDs, but not catalog.Groups; groups are later inserted into a map by ID, where a duplicate silently overwrites the earlier entry (internal/traceability/catalog.go:152–177, 179–202). The README says the checker verifies unique IDs (docs/traceability/README.md:13). A bounded temporary test appended a duplicate group and separately appended a group whose ID equaled an existing task ID; both fixtures returned no validation diagnostics. Existing valid catalog results were otherwise unchanged.

**Impact/uncertainty:** A malformed planning catalog can carry ambiguous group ownership and map resolution while passing the advertised uniqueness check. This affects planning metadata integrity only; it does not affect PDF behavior or accessibility outcomes. Low priority because the current catalog has only two groups.

**Proposed follow-up:** Add group IDs to the same stable-ID namespace and reject collisions deterministically. Acceptance: tests cover an exact duplicate, a conflicting duplicate group with a different story, and a group/task ID collision; the current catalog still validates and reports retain stable output.

### P3 — Relationship scans scale quadratically in the planning checker

**Evidence:** The benchmark above grows from 0.283 ms at 100 linked entities to 376.79 ms at 5,000 for validation, and from 0.351 ms to 326.85 ms for Markdown generation. Source locations are listed above.

**Impact/uncertainty:** The growth is measurable but current catalog size is 17 requirements and 6 stories. The checker is repository tooling, not a document-processing hot path. No performance failure was measured at the current size.

**Proposed follow-up:** Track a planning-catalog size/runtime budget. If the catalog approaches hundreds or thousands of linked entities, precompute story-to-requirement/function/task/case indexes. Acceptance for any future indexing change: existing validation diagnostics and all four generated Markdown byte outputs remain identical, the same 100/1,000/5,000 fixture remains valid, and repeated benchmarks show materially flatter growth without an allocation regression. Do not treat these numbers as PDF throughput.

### P3 — Bootstrap subprocess output has no byte ceiling

**Evidence:** installGoModule uses CombinedOutput for a command allowed up to four minutes (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:304–336). installedOK also uses CombinedOutput for version commands under a 20-second context, including the fallback (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:443–479). This captures child output fully in memory. No output-flood experiment was run against bootstrap.

**Impact/uncertainty:** This is a resource-bound gap if a checksum-pinned tool or Go command emits unexpectedly large output. Normal installed tools and source modules are pinned and checksum-verified, so the review found no observed incident or ordinary-path pressure.

**Proposed follow-up:** Replace unbounded capture with a bounded writer that preserves exit status while discarding or truncating output; retain the current credential/path redaction behavior. Acceptance: small success/failure output remains classified as today; a safe synthetic writer exceeding the configured cap is stopped or truncated at that cap; diagnostics do not expose child output or environment secrets. No unbounded flood is needed.

### P3 — Bootstrap performs a second full verification pass

**Evidence:** In the checkout path, the shell runs the bootstrap command and then invokes it again with --verify (scripts/bootstrap-cloud-build-tools.sh:238–242). The no-checkout path likewise runs bootstrap and then verify (scripts/bootstrap-cloud-build-tools.sh:244–269). The manifest contains eight entries. Bootstrap already calls installedOK for installed tools and verifies each newly installed tool before returning (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:120–144).

**Impact/uncertainty:** This duplicates setup-time hashes/version subprocesses; it does not add per-document cost. No bootstrap duration was measured here, and the no-checkout path extracts/rebuilds an archived bootstrap source pinned separately at 8f0072c, so timings from that path must not be attributed to the current module baseline.

**Proposed follow-up:** Measure cold and warm checkout/no-checkout bootstrap with a synthetic manifest and local stub tools. Only then decide whether a separate verify pass is redundant. Acceptance: retain checksum/version validation for all eight entries and preserve failure exit behavior while reducing redundant work if a measurable setup cost is confirmed.

## Positive evidence and limits

- ParseReport and bundle verification reject malformed/noncanonical input and cap metadata/report size. Report outcome is derived from XML and cross-checked against process status; bundle metadata alone cannot assert a pass.
- The adapter copies candidate/artifact bytes into a private temporary directory, hashes both, bounds total output, propagates cancellation, and uses process-tree cleanup on Linux. The workflow runs adapter race tests separately from real Java/veraPDF execution.
- Bootstrap uses context deadlines, atomic staging/rename, checksums, a lock, and process-group cleanup; race tests passed. The successful-leader descendant observation from the previous panel was not reopened as a new finding.
- Error messages that withhold subprocess/path details are a deliberate privacy boundary, not an API-quality defect in this review.
- No end-to-end PDF throughput, live external service load, production data, hostile probe, product-source change, dependency addition, Beads change, issue closure, commit, push, or merge was performed.
