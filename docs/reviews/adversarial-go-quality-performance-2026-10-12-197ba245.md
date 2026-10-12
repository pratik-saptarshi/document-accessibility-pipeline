# Adversarial Go quality and performance panel

**Review date:** 2026-10-12 UTC

**Pinned source:** 197ba24596d105971b4d0d88b2dfab3edd724156 (fix(bootstrap): clean up child processes on signals and deadlines, after PR19)
**Scope:** Code quality, Go API/error design, maintainability, allocations, scaling, concurrency, lifecycle, resource use, all three Go modules, and CI tooling. This is a quality/performance audit; no product remediation was made.

## Executive summary

No high-priority product defect or concurrency failure was reproduced. The current tree is planning and CI scaffolding, not an end-to-end document service. VerifyBundle enforces a 16 KiB metadata limit and an 8 MiB report limit; ParseReport has no input-length cap when called directly. The traceability loaders also lack an aggregate input-size limit: LoadCatalog reads the file with os.ReadFile, and LoadIssues limits each scanned line but not the total file. The main reproduced correctness gap is narrow: traceability.Validate omits task-group IDs from duplicate-ID validation. Bounded fixtures showed duplicate group IDs and a group/task ID collision pass with zero diagnostics.

The synthetic performance runs expose repeated relationship scans in validation and Markdown generation. At the current catalog size (17 requirements, 6 stories, 17 functions, 20 tasks, 26 cases, 2 groups), those scans are not a practical bottleneck. The corrected, exact-size 8 MiB comment-padded XML fixture takes median 127.44 ms for ParseReport and 135.74 ms for VerifyBundle. These are parser/bundle measurements, not PDF validation throughput.

Four follow-ups are being tracked: one confirmed P3 planning-integrity correction for duplicate task-group IDs, plus three optional follow-ups on synthetic scaling, subprocess output bounds, and repeated bootstrap verification. The optional items are not measured incidents or immediate optimization requests. No product remediation is included in this review.

## Panel method

The panel used five review lenses: idiomatic Go/API/error design; architecture, duplication, and test quality; allocation/throughput/scaling; concurrency/lifecycle/resource use; and an independent challenger separating measured evidence from hypotheses. Independent parent-supplied reviewer observations were incorporated where applicable. Reviewers distinguished the planning checker and CI/test adapters from product paths.

The review inputs and measurements remain pinned to the exact SHA stated above. Temporary benchmark and probe files were removed after measurement; reproducibility sources remain under support/ as .txt files so they do not join Go packages. The audit branch was based on current main because the publication guard requires new branches to descend from main. The source cutoff and measurements were not moved to that parent. A separately discussed no-checkout bootstrap benchmark pinned to source archive 8f0072c885fad7811635523ced39df0f1a23ca33; those timings are not attributed to this 197ba245 review.

No .agents skills were present in the pinned tree. Installer-managed Go 1.27.2 and pinned CI tools passed the read-only version/digest verification. The initial checkout's embedded Beads database was unavailable, so I restored the tracked 57-issue export into a temporary isolated workspace, disabled Beads metrics, and used it to create four follow-ups. The final export contains 61 issues. The convention warning applied exactly to the four new records; their issue descriptions now include the required Steps to Reproduce section for the bug and Goal and Findings sections for each spike. Acceptance criteria remain in Beads' acceptance field. After correction, all 37 open issues passed template lint, and artifact, stale, and orphan checks passed. No issue was closed. No Java or veraPDF executable was run, and no CLI sign-in or credential state was changed. Neuro remains the accepted permanent cloud limitation; no setup, model, or authentication work was attempted.

## Verification

- Go toolchain: 1.27.2. Machine and host details are intentionally omitted. No dependency versions or module files changed.
- Root module: go test -race ./..., go vet ./..., and go mod tidy -diff passed.
- cmd/cloud-tools-bootstrap: go test -race ./..., go vet ./..., and go mod tidy -diff passed.
- ci/pdfua2/adapter: go test -race ./..., go vet ./..., and go mod tidy -diff passed. The real-validator test was not invoked; Java/veraPDF execution remains CI-only.
- CI tooling checks: actionlint, bash -n scripts/*.sh .github/scripts/*.sh, test-source-publisher.sh, test-gitleaks.sh, test-go-licenses.sh, and the bounded synthetic test-verapdf-output-bound-shim.sh passed. The latter caps its synthetic output fixture at 16 MiB and verifies the 8 MiB capture boundary. The go-licenses smoke check passed using writable temporary caches and local synthetic modules. The scanner test used local synthetic modules and fixtures. On the publication parent, go run ./cmd/traceability-check passed with 17 requirements, 6 stories, 20 tasks, and 26 specification-only cases. Both saved benchmark harnesses compiled and completed a one-iteration smoke run; the group-ID probes passed with no diagnostics. Beads convention checks passed after adding the required sections to the four newly created follow-ups; no template warnings remained across 37 open issues.
- gofmt was applied to the temporary benchmark/probe source before preserving it. No product source was reformatted.
- Static CI workflow review found root and bootstrap checks in go-checks.yml, adapter race tests in the separate veraPDF integration workflow, and pinned scanner/tool installs in security-scans.yml. govulncheck was not run because its configured vulnerability database is external; no network settings were changed. CI-only Java/veraPDF execution was left to CI.

## Measured benchmarks

The harnesses are in [support/pdfua_benchmark_test.go.txt](support/pdfua_benchmark_test.go.txt) and [support/traceability_benchmark_test.go.txt](support/traceability_benchmark_test.go.txt). Fixtures were generated and validated before timing. XML fixture helpers assert the exact requested byte length. Samples were serial, with five repeats per case, -benchmem; XML used -benchtime=500ms, traceability used -benchtime=300ms. B/op is total allocation per operation, not peak live memory.

Reproduction after copying each support file to the indicated package as review_benchmark_test.go:

    go test -run '^$' -bench 'BenchmarkReview(ParseReport|VerifyBundle)' -benchmem -benchtime=500ms -count=5 ./internal/pdfua
    go test -run '^$' -bench 'BenchmarkReview' -benchmem -benchtime=300ms -count=5 ./internal/traceability

### Report parser and bundle verifier

The XML fixture is a valid single-job report with comment padding to the exact requested size. It holds XML element count constant, so it exercises large text payloads rather than realistic validator issue density. VerifyBundle also hashes the report before calling ParseReport.

| Input size | ParseReport median (range) | B/op; allocs/op | VerifyBundle median (range) | B/op; allocs/op |
|---:|---:|---:|---:|---:|
| 1 KiB | 27.920 µs (27.638–28.354) | 11,752; 172 | 39.610 µs (38.903–41.427) | 18,373; 252 |
| 64 KiB | 1.035 ms (1.026–1.048) | 269,801; 184 | 1.097 ms (1.089–1.117) | 276,904; 265 |
| 1 MiB | 16.11 ms (15.79–16.70) | 4,201,998; 192 | 16.82 ms (16.55–17.03) | 4,214,951; 287 |
| 8 MiB | 127.44 ms (126.02–136.48) | 33,562,126; 198 | 135.74 ms (130.34–139.25) | 33,575,104; 293 |

ParseReport does a token-validation pass and then a decoding pass (internal/pdfua/report.go:15–20, 130–160) and has no direct input-size cap. VerifyBundle enforces a 16 KiB metadata cap and 8 MiB report cap before hashing and parsing (internal/pdfua/bundle.go:43–50, 77–84). B/op is total allocation; it is not peak live memory. These comment-padded synthetic fixtures do not define worst-case parser cost or model veraPDF runtime, PDF complexity, container execution, or end-to-end throughput.

### Traceability validation and Markdown generation

Each valid synthetic catalog had N requirements, stories, functions, tasks, and acceptance cases, with one-to-one links and valid synthetic Beads issues. It included lane/family coverage. Validation ran once before timing; Markdown generation emitted and checked all four views.

| Entities in each linked collection | Validate median (range) | B/op; allocs/op | GeneratedMarkdown median (range) | B/op; allocs/op |
|---:|---:|---:|---:|---:|
| 100 | 0.283 ms (0.266–0.294) | 150,638; 230 | 0.351 ms (0.343–0.376) | 414,134; 3,560 |
| 1,000 | 15.45 ms (15.18–16.00) | 1,820,092; 2,080 | 15.19 ms (14.90–15.32) | 4,251,654; 35,105 |
| 5,000 | 376.79 ms (374.56–401.63) | 7,253,616; 10,252 | 326.85 ms (317.12–339.31) | 22,507,104; 175,151 |

Validate nests the requirement/story scan and then scans functions and tasks for each linked story (internal/traceability/catalog.go:310–343). renderMatrix scans all functions/tasks for each story and storyAcceptanceCases rescans tasks (internal/traceability/catalog.go:482–507, 513–529). This explains the measured near-quadratic growth in the synthetic one-to-one case. The real catalog is much smaller and the checker is a local planning tool; this is a scaling limit to track, not a current product-path defect. The loaders do not impose an aggregate file-size limit: LoadCatalog uses os.ReadFile (catalog.go:105–113), and LoadIssues enforces a 4 MiB per-line Scanner maximum while continuing across lines without a total-byte cap (catalog.go:116–145). These are local repository metadata readers, not upload endpoints; this report does not claim planner inputs are bounded.

## Findings and Beads follow-ups

### P3 — Task-group IDs are omitted from duplicate-ID validation

**Beads:** `dap-jo9.13` (confirmed defect)

**Evidence:** Validate adds capabilities, requirements, stories, functions, tasks, and tests to allIDs, but not catalog.Groups; groups are later inserted into a map by ID, where a duplicate silently overwrites the earlier entry (internal/traceability/catalog.go:152–177, 179–202). The README says the checker verifies unique IDs (docs/traceability/README.md:13). A bounded temporary test appended a duplicate group and separately appended a group whose ID equaled an existing task ID; both fixtures returned no validation diagnostics. Existing valid catalog results were otherwise unchanged.

**Impact/uncertainty:** A malformed planning catalog can carry ambiguous group ownership and map resolution while passing the advertised uniqueness check. This affects planning metadata integrity only; it does not affect PDF behavior or accessibility outcomes. Low priority because the current catalog has only two groups.

**Proposed follow-up:** Add group IDs to the same stable-ID namespace and reject collisions deterministically. Acceptance: tests cover an exact duplicate, a conflicting duplicate group with a different story, and a group/task ID collision; the current catalog still validates and reports retain stable output.

### P3 — Relationship scans scale quadratically in the planning checker

**Beads:** `dap-jo9.14` (optional threshold/indexing follow-up)

**Evidence:** The benchmark above grows from 0.283 ms at 100 linked entities to 376.79 ms at 5,000 for validation, and from 0.351 ms to 326.85 ms for Markdown generation. Source locations are listed above.

**Impact/uncertainty:** The growth is measurable but current catalog size is 17 requirements and 6 stories. The checker is repository tooling, not a document-processing hot path. No performance failure was measured at the current size.

**Proposed follow-up:** Track a planning-catalog size/runtime budget. If the catalog approaches hundreds or thousands of linked entities, precompute story-to-requirement/function/task/case indexes. Acceptance for any future indexing change: existing validation diagnostics and all four generated Markdown byte outputs remain identical, the same 100/1,000/5,000 fixture remains valid, and repeated benchmarks show materially flatter growth without an allocation regression. Do not treat these numbers as PDF throughput.

### P3 — Bootstrap subprocess output has no byte ceiling

**Beads:** `dap-5hy.2` (optional hardening hypothesis; no output-flood experiment)

**Evidence:** installGoModule uses CombinedOutput for a command allowed up to four minutes (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:304–336). installedOK also uses CombinedOutput for version commands under a 20-second context, including the fallback (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:443–479). This captures child output fully in memory. No output-flood experiment was run against bootstrap.

**Impact/uncertainty:** This is a resource-bound gap if a checksum-pinned tool or Go command emits unexpectedly large output. Normal installed tools and source modules are pinned and checksum-verified, so the review found no observed incident or ordinary-path pressure.

**Proposed follow-up:** Replace unbounded capture with a bounded writer that preserves exit status while discarding or truncating output; retain the current credential/path redaction behavior. Acceptance: small success/failure output remains classified as today; a safe synthetic writer exceeding the configured cap is stopped or truncated at that cap; diagnostics do not expose child output or environment secrets. No unbounded flood is needed.

### P3 — Bootstrap performs a second full verification pass

**Beads:** `dap-5hy.3` (optional measurement; no setup-cost measurement yet)

**Evidence:** In the checkout path, the shell runs the bootstrap command and then invokes it again with --verify (scripts/bootstrap-cloud-build-tools.sh:238–242). The no-checkout path likewise runs bootstrap and then verify (scripts/bootstrap-cloud-build-tools.sh:244–269). The manifest contains eight entries. Bootstrap already calls installedOK for installed tools and verifies each newly installed tool before returning (cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap.go:120–144).

**Impact/uncertainty:** This duplicates setup-time hashes/version subprocesses; it does not add per-document cost. No bootstrap duration was measured here, and the no-checkout path extracts/rebuilds an archived bootstrap source pinned separately at 8f0072c, so timings from that path must not be attributed to the current module baseline.

**Proposed follow-up:** Measure cold and warm checkout/no-checkout bootstrap with a synthetic manifest and local stub tools. Only then decide whether a separate verify pass is redundant. Acceptance: retain checksum/version validation for all eight entries and preserve failure exit behavior while reducing redundant work if a measurable setup cost is confirmed.

## Positive evidence and limits

- ParseReport rejects malformed/noncanonical XML but has no input-length cap. VerifyBundle enforces 16 KiB metadata and 8 MiB report limits; the outcome is derived from XML and cross-checked against process status. Bundle metadata alone cannot assert a pass.
- The adapter copies candidate/artifact bytes into a private temporary directory, hashes both, bounds total output, propagates cancellation, and uses process-tree cleanup on Linux. The workflow runs adapter race tests separately from real Java/veraPDF execution.
- Bootstrap uses context deadlines, atomic staging/rename, checksums, a lock, and process-group cleanup; race tests passed. The successful-leader descendant observation from the previous panel was not reopened as a new finding.
- Error messages that withhold subprocess/path details are a deliberate privacy boundary, not an API-quality defect in this review.
- No end-to-end PDF throughput, live external service load, production data, hostile probe, product-source change, or dependency addition was performed. The report and reproducibility harnesses were published for review; no merge was performed.
