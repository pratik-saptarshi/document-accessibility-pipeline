# Adversarial panel: main at a6d6fc6

**Review date:** 2026-10-13 UTC | **Reviewed commit:** `a6d6fc65e88351d710b3d8d4229fe9721c984884` | **Commit subject:** Isolate veraPDF PDF/UA-2 validation from Go (#14)
**Scope:** this exact post-PR14 snapshot. The fetched `origin/main` tip at final sync is `23b11d34d6654685f47e43853ef0d077350a5183` (PR30). Findings and reproductions below remain pinned to `a6d6fc6…`; the later-main status section reports follow-up evidence only and does not retroactively change that baseline. No product code was changed during this review.

## Scope and panel roles

The review covered security and isolation, PDF/UA-2 report classification and semantic limits, Go reliability and resource exhaustion, supply-chain/CI enforcement, and test/roadmap gaps. The five review roles were:

1. **Security and isolation:** hostile inputs, filesystem boundaries, process/container controls, and failure behavior.
2. **PDF/UA-2 and semantics:** validator profile and report consistency, preservation evidence, semantic/assistive-technology limits.
3. **Go reliability and resources:** cancellation, process trees, bounded I/O, and resource limits.
4. **Supply chain and CI:** pinned tools/actions, scanner-policy enforcement, and untrusted pull-request code execution.
5. **Test and roadmap gaps:** regression coverage, traceability, and planning-vs-product evidence.

The five roles were used to generate adversarial questions and targeted reproductions. The report author ran the local reproductions below in disposable fixtures and separated confirmed evidence from threat-model assumptions. Independent parent review is expected separately and is not represented as complete here.

Repository instructions and relevant material read: AGENTS.md, CONTRIBUTING.md, docs index, product scope, roadmap, PDF/UA-2 decision 0001, security operations, build-machine tooling proposal, current CI workflows, Beads export, and validator/preservation source and tests. The requested `.agents/skills` directories were absent from both the checkout and mounted workspace (`/workspace/.agents` was empty); no skill files were available. The repository is planning scaffolding: CI validator and preservation evidence do not establish production support or compliance. The five roles were applied as adversarial review lenses; independent parent review is expected separately and is not claimed as complete here.

## NeuroStrata and NeuroPlasticity evidence

Both executables were found in the workspace-local cloud build-tools bundle, outside the repository. Their digests matched the pinned repository manifest:

| Tool | Digest and safe local use | Limitation |
|---|---|---|
| NeuroStrata 1.9.0 | On 2026-10-13, binary SHA-256 `9c029685bb881252b67de021d2ce5265e0acda0098532592d000a1586226166e` matched the pinned manifest. `--help` and `status` ran with HOME/XDG paths scoped under `/tmp`; status reported no database file, daemon down, and free lock. | No memory add/search or model initialization was run. The task owner reports the exact `us.aws.cdn.hf.co` model-host addition was approved but the environment editor denied it twice. No download was retried and no network setting was changed. NeuroStrata did not review source code. |
| NeuroPlasticity 2.1.0 | On 2026-10-13, binary SHA-256 `85a476b1b13f0ddebeb9928946004ea2f55b04bee27038f71913456759db3634` matched the pinned manifest. `--help` and `test /tmp/dap-panel-neuroplasticity.json --print-egress-plan` ran with HOME/XDG paths under `/tmp` and an otherwise clean environment. The synthetic public-only manifest produced `Data class: unclassified (no enforcement applies)` and `embedded → ALLOWED (local inference, nothing leaves the machine)`. No version flag was invoked. | This was egress-plan inspection only, not an agent/evaluator run or model-assisted review. The plan is application output, not a network firewall guarantee. No container, optimizer, model, or credential path was used. |

No model was downloaded, no credentials were accessed or mounted, no repository content was uploaded to either tool, and no persistent MCP registration was created. NeuroStrata `status` reported no database file and no running daemon; no memory add/search or model initialization was attempted. The reported model-host allowlist approval and two environment-editor denials are supplied by the task owner; this panel did not retry the blocked download or change network settings. These tools are not counted as panel reviewers. This panel used source inspection and local Go/shell reproductions, not full Neuro-assisted execution.

## Confirmed findings

### F-01 — Round-trip comparator misses encoded stream mutations (P2)

`internal/pdfpreserve/roundtrip_test.go` compares stream dictionaries and decoded `Content` for objects 13 and 14, but not encoded `Raw`. The pinned pdfcpu v0.16.2 default configuration sets `DecodeAllStreams` to false, and its `saveDecodedStreamContent` returns before decoding. Both synthetic stream objects therefore had nil `Content`.

A temporary Go test mutated same-length `Raw` bytes in page content object 13 and form XObject 14, wrote and reread each PDF, and called the existing `compareRoundTrip`. In both cases the serialized `Raw` changed while `compareRoundTrip` returned nil. This is a confirmed preservation-oracle gap; it does not show that a production writer currently corrupts PDFs.

**Acceptance:** compare encoded bytes as well as decoded content where available; add same-length mutation controls for both stream types and require comparison against the independently read baseline to fail. Keep the result labeled characterization evidence, not PDF/UA-2 conformance.

### F-02 — Truncated validator output can satisfy the known-finding control (P2)

`.github/scripts/run-verapdf-isolated.sh` accepts exit code 1 for the known-finding control and rejects report size only when it is greater than 8 MiB. A temporary Go Docker shim wrote valid noncompliant XML, then whitespace until the file-size limit. The captured report was exactly 8 MiB; the shim returned 1, the runner emitted a contract, and `TestVerifyDownloadedCIControlBundles/known-finding` accepted it. `ParseReport` permits trailing XML whitespace, so the truncated output remained parseable. A second Go shim wrote a valid report to stdout and overflowed stderr; the runner deleted stderr and also emitted a known-finding contract after the shim returned 1.

This demonstrates that a Docker client/output-sink failure can be mistaken for veraPDF's expected noncompliant exit. It is a validator-control false positive under the reproduced output shape; no genuine veraPDF failure was induced, and this does not show that the fixed corpus PDF produces such output.

**Acceptance:** reject output at or above the byte limit, preserve and check output-channel errors, and distinguish Docker/client operational failures from the validator container's exit code. A synthetic over-limit report with a valid noncompliant prefix must not produce a bundle accepted by the Go verifier.

### F-03 — Candidate digest is not bound to the staged PDF bytes (P2)

The runner hashes the input before copying it to the staged file, then records that original digest in `contract.json`; it never verifies the staged copy. In a disposable reproduction, a `sha256sum` wrapper returned the correct known-control digest and then changed the copied input before the `install` operation. An `install` wrapper and Docker shim observed staged SHA-256 a01db41bb136f52de4b218ea0d1644e3b10fc99dbb2ec38b583a9ad65df4755a, while the contract retained the control's expected digest 2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0. The Go known-finding bundle test still accepted the contract/report.

The hash/copy race is confirmed. The current workflow uses fixed, trusted control files and the reproduction does not establish a PR-to-host exploit or concurrent writer in the hosted job. The evidence claim is nevertheless not bound to the bytes actually validated if mutation can occur in that interval.

**Acceptance:** hash the staged file after copying and bind that observed digest to the validator contract; fail if it differs from the permitted control digest. Include a deterministic mutation-between-check-and-copy test.

### F-04 — PR-controlled Go tests can replace later security scanners (P2)

The security workflow installs scanners with `go install`, runs `go test -race ./...`, then invokes actionlint, govulncheck, and go-licenses in later steps. A disposable Go module test overwrote a dummy govulncheck executable in writable GOBIN; the subsequent command resolved to the replacement. The actual workflow does not set GOBIN, but Go installs into the writable GOPATH/bin by default and the subsequent scanner commands must resolve those installed files. PR test code runs under the same job identity and can replace or shadow them before scan steps.

This is a confirmed CI gate-integrity exposure for untrusted pull requests. The job requests contents:read and this does not demonstrate token escalation. No actual workflow binary was replaced.

**Acceptance:** run PR-controlled tests separately from pinned scanner execution, or otherwise make scanner resolution immutable to the test job. A fixture test attempting to overwrite or shadow a scanner must not change the executable used by later scan steps.

### F-05 — Bootstrap continues successfully after SIGTERM (P3)

`scripts/bootstrap-cloud-build-tools.sh` traps EXIT, INT, and TERM with cleanup commands but does not exit from the INT/TERM trap. In a disposable tools/cache root, a fake pinned Go binary delayed the first run. Sending SIGTERM only to the bootstrap shell left the lock in place while the child remained active; after the child finished, the script ran the verification invocation, cleaned the lock, printed the success status, and exited 0.

No early lock removal or child termination occurred in this reproduction. The defect is that cancellation is reported as successful completion and subsequent work continues.

**Acceptance:** signal tests must show nonzero exit, no later verification after cancellation, lock retention until foreground work ends, and cleanup afterward.

### F-06 — Gitleaks suppression discovery fails open when find errors (P3)

`.github/scripts/check-gitleaks-suppressions.sh` reads `find` output via process substitution. Under `set -euo pipefail`, the process-substitution command's nonzero status is not propagated. A disposable `find` shim that returned 1 without output made the guard exit 0 and print “No unmanaged Gitleaks config or suppression files found.”

The CI runner normally has `find`; no present workflow failure was observed. The guard does not fail closed if discovery itself fails.

**Acceptance:** propagate `find` failures and add a shim test for nonzero/no-output behavior, while preserving clean-tree and unmanaged-file detection tests.

## Conditional risks and bounded observations

### R-01 — Failed rerun leaves a previous contract file (P3, caller-dependent)

A successful known-finding run followed by a failed run using the same output directory leaves the old `contract.json`. The current workflow uses a fresh `$RUNNER_TEMP` output root, and a failed shell step prevents later verification/upload under Actions' default success gating. Thus stale evidence was reproduced on disk, but no false pass was reproduced in the current workflow. Callers must not reuse an output root or consume artifacts after a failed runner invocation.

### R-02 — Push target can change during pre-push hook (P3, local-trust boundary)

In two temporary bare repositories, a local pre-push hook changed `remote.origin.pushurl` during `git push --dry-run`. Git ran the hook; the dry run used remote one and the subsequent push would resolve remote two. `scripts/source-publisher.sh` validates the HTTPS origin before the dry run and rechecks HEAD afterward, but it does not revalidate the destination between dry-run and push. No repository remote/hook was changed and no external publication occurred. This requires a local hook/config mutation; such a hook already executes with the publishing user's privileges, so the reproduction does not establish a privilege gain or attacker path. Treat destination revalidation as hardening, not a demonstrated remote exploit.

### Process-tree and resource checks

The independent descendant-survival candidate is already covered by `ci/pdfua2/adapter/process_linux_test.go`: tests exercise a child holding inherited pipes and a child with closed pipes after leader exit, then assert process-group termination. The adapter also bounds input/artifact/output and runtime; the validator container adds memory, CPU, PID, filesystem, network, and privilege bounds. Resource exhaustion outside the CI containerized boundary remains a deployment question because the adapter is explicitly CI/test-only.

## Prior-panel regression checks

The previous panel's findings were rechecked against this post-PR14 tree:

| Prior finding | Status at a6d6fc6 |
|---|---|
| Ambiguous/empty validator XML and leaked parser diagnostics | Fixed in parser token validation and generic errors; current PDF/UA tests pass. |
| Relative temporary directory paths resolved from the wrong child cwd | Fixed by making the temp directory absolute before deriving JAR/input paths; current adapter tests pass. |
| Security policy missed disabled steps, job continue-on-error, and GPL-3.0 | Fixed: policy tests model step conditions/job continue-on-error and reject GPL-3.0; mutation and license smoke tests pass. |
| Round-trip oracle used writer context and omitted catalog | Fixed: independent baseline, catalog and accessibility metadata comparison, and mutation tests exist. New F-01 is distinct: encoded stream Raw bytes are omitted. |
| Story traceability omitted task-linked human/AT cases | Fixed: catalog test requires TC-REM-HUMAN-001 and TC-REM-ASSISTIVE-TECH-001. |
| Schemas accepted whitespace-only decision values rejected by Go | Fixed: schema/Go parity tests cover whitespace boundaries. |
| General Go CI job lacked an explicit timeout | Fixed: go-checks job has timeout-minutes and policy tests reject missing/out-of-range values. |

## Follow-up status on current main

At fetched main tip `23b11d34d6654685f47e43853ef0d077350a5183`, follow-up implementation commits and issue reconciliations are present after the audited SHA. This table reports repository state, not a re-rating of historical findings.

| Finding | Current-main change | Beads status |
|---|---|---|
| F-01, F-02, F-03 | PR16 (`4760735`) added raw-stream preservation checks and validator staged-byte/output-bound handling. | `dap-jo9.2.3.4` and `dap-jo9.2.5.7` are closed in the current export; acceptance notes record review and checks for PR16. Real validator controls remain CI-only. |
| F-04 | PR17 (`edd0360`) isolates pinned scanner execution from project tests. | `dap-jo9.11` is closed in the current export. |
| F-06 | PR18 (`793c031`) propagates Gitleaks discovery errors. | `dap-jo9.12` is closed in the current export. |
| F-05 | PR19 (`197ba24`) cleans up bootstrap child processes on signals and deadlines. | `dap-5hy.1` is closed in the current export. |

Current main still does not turn CI validator evidence into human semantic/assistive-technology review. The separate ruleset investigation remains untouched: per task-owner status, the seven existing rules remain required, while GitHub returned `Unauthorized` for attempts to require `pdfua2` and `validator-container`.

## PDF/UA-2, semantics, and roadmap limits

The dedicated CI workflow uses the authenticated pinned veraPDF CLI, explicit `--flavour ua2`, a digest-pinned container base, network isolation, and genuine pass/finding corpus controls. Go verifies report metadata and report digests. These checks are intended to record machine-validation evidence for the named control and candidate hash, subject to the output-classification and staged-byte binding gaps in F-02 and F-03. They do not validate reading order or semantic intent and do not replace human semantic or assistive-technology review. No human/AT review was performed in this panel. The existing synthetic PDF 1.7 preservation cases remain characterization fixtures, not positive PDF/UA-2 controls. No production PDF service or release eligibility was assessed.

The branch-ruleset blocker is outside this panel and remains with the separate read-only investigation. Per task-owner status, the original seven required checks remain; `pdfua2` and `validator-container` run successfully, but GitHub rejected making them mandatory with `Unauthorized`. This panel did not query or change rulesets and does not claim those two checks are required.

## Reproduction and validation evidence

Commands ran on a6d6fc65e88351d710b3d8d4229fe9721c984884 using Go 1.27.2, with HOME scoped to /tmp because pdfcpu attempts to create a default config under the executor's read-only shared home.

| Check | Result |
|---|---|
| `go test -race -timeout=180s ./...`, `go vet ./...`, root `go mod tidy -diff` | Passed at the pinned SHA. |
| Cloud bootstrap and `ci/pdfua2/adapter` nested-module race tests, vet, and tidy checks | Passed. Windows adapter test-binary cross-compile passed. |
| `actionlint` and Bash syntax checks for the reviewed scripts | Passed. |
| Validator output-bound shim and source-publisher synthetic local-repository tests | Passed. No actual validator container or external publication ran. |
| Gitleaks smoke | Passed with synthetic/redacted fixtures. Separate `find` failure reproduced the fail-open gap F-06. |
| Go license smoke | The production-only fixture scan passed, but the test-inclusive scan emitted `Package testing does not have module info` under bundled Go 1.27.2 before reaching its expected GPL-3.0 assertion. This run does not establish the test-inclusive policy check. |
| Temporary Go-first stream-mutation test | On 2026-10-13 at the pinned SHA, a temporary `TestReviewRawStreamMutationMustBeDetected` failed for both objects 13 and 14 with `accepted same-length encoded stream mutation`; test file removed and worktree clean. The output confirms the expected red control. |
| Temporary validator shims | Reproduced exact-8-MiB/trailing-whitespace known-finding acceptance. A hash-then-mutate shim also showed contract digest `2d0996...badd0` while the staged PDF hash was `3868fec5...a2a287`; the Go bundle verifier accepted it. No actual Docker/veraPDF container ran. |
| Bootstrap signal test | SIGTERM sent only to the bootstrap shell left the lock while fake Go ran, then verification started and the script exited 0. Disposable tools/cache removed. |
| Scanner replacement fixture | A disposable Go test replaced a placeholder `govulncheck` in `GOBIN`; the later name-based invocation resolved to the replacement. No real scanner was changed. |
| NeuroStrata and NeuroPlasticity | On 2026-10-13, SHA-256 matched both manifest digests. Both `--help` commands ran with HOME/XDG paths under `/tmp`; NeuroStrata `status` reported no DB file, no daemon, and a free lock. NeuroPlasticity's synthetic egress-plan mode reported embedded/local inference only. No NeuroPlasticity version flag was used. No model, credentials, daemon, or container was started; detailed limits above. |

## Beads follow-ups

The original panel added five issues to the tracked export from an isolated import of the 52-issue snapshot; all original issue records were unchanged and that snapshot contained 57 issues. This revision reuses those Beads IDs already present on current main and does not rewrite the current export, create duplicates, or run a Dolt push/sync.

| Beads | Follow-up |
|---|---|
| `dap-jo9.2.3.4` | Compare encoded stream bytes in preservation checks. Closed after PR16 review and checks. |
| `dap-jo9.2.5.7` | Bind validator evidence to staged bytes and reject output errors. Closed after PR16 review and checks; real validator controls remain CI-only. |
| `dap-jo9.11` | Isolate pinned security scanners from PR-controlled tests. Closed after PR17. |
| `dap-5hy.1` | Propagate termination during cloud tool bootstrap. Closed after PR19. |
| `dap-jo9.12` | Fail closed when Gitleaks suppression discovery fails. Closed after PR18. |

## Review boundary

This report records findings at the exact SHA above. It does not implement remediation, claim PDF/UA-2 or WCAG conformance, claim a full Neuro-assisted review, or report a genuine local veraPDF container run. Findings distinguish demonstrated false-acceptance/oracle gaps from local-trust and caller-dependent risks.
