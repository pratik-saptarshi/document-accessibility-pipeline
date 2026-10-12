# Adversarial panel: recorded `main` snapshot

**Review date:** 2026-10-12 UTC  
**Reviewed commit:** `595e7e9224fc6f22abb75e9f706dc2558514f620`  
**Commit subject:** `feat(corpus): add synthetic remediation manifest validation (#4)`  
**Review status:** three independent reviewer roles and local Go probes supplied evidence, and the full Go test suite passed. This is **not a completed NeuroStrata/NeuroPlasticity-assisted panel**: NeuroStrata only ran its model-free status command, and NeuroPlasticity only printed an egress plan; neither tool reviewed the code. Their runtime prerequisites remained blocked as detailed below.

## Scope and exclusions

The reviewed tree is an isolated checkout of the fetched `origin/main` at the exact SHA above. PR #4's synthetic remediation manifest is included. PR #5 (`codex/go-generation-engine-evidence`, head `cb0dca5eb3ab8cac24548da30bd80b9c16cf53d3`) was open and excluded. Later draft PR #7 (`ci(pdfua): provision pinned veraPDF oracle integration`, head `c8e1a5b0a2884204453f11f0f558d8f73e0f1a9e`) completed signed veraPDF acquisition and UA-2 pass/fail controls after this pin; it remains unmerged and is outside this review. The PDF/UA adapter scaffold present at this pin was reviewed and tested with synthetic fake-process reports only.

The repository describes planning scaffolding plus CI/test-only validation components; no production PDF service was in scope. The panel covered hostile inputs and subprocess boundaries, PDF/UA-2 classification assumptions and accessibility claims, Go concurrency and resource controls, CI/supply-chain enforcement, and test adequacy and requirements traceability. No private/customer documents were used. No product code, fixtures, workflow policy, branch rules, or other repositories were changed.

## Panel charter and process

Three independent reviewer roles completed their review and sent findings for adjudication:

1. **Security and hostile-input reviewer:** adversarial XML, parser ambiguity, diagnostic disclosure, CI scanner bypasses, license policy variants, and process inputs.
2. **PDF accessibility and QA reviewer:** preservation-test coverage, catalog/accessibility metadata, semantic and assistive-technology evidence, and story/case traceability.
3. **Go runtime reviewer:** subprocess lifecycle, bounded I/O, cancellation, concurrency, resource controls, reproducibility, and false-positive assessment.

The chair then reproduced reported findings where feasible, challenged interpretations against the pinned source, recorded dissent, and assigned severity. Reviewer identities are intentionally not represented as evidence; this report records their roles and adjudication. NeuroStrata and NeuroPlasticity are build-machine tools, not shipped dependencies, and were not counted as independent reviewers.

### NeuroStrata and NeuroPlasticity

The approved provenance and isolation guidance is in [build-machine-tooling.md](../planning/build-machine-tooling.md). Official pinned binaries were installed outside the checkout and checksums matched the recorded official release metadata. Versions: NeuroStrata 1.9.0, NeuroPlasticity 2.1.0.

- NeuroStrata's local, model-free `status` command ran and reported no database file and no daemon. Pinned NeuroStrata v1.9.0 uses FastEmbed 7.0.1; Nomic Embed Text v1.5 resolves the ONNX file at `https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/onnx/model.onnx` (domain `huggingface.co`; about 523 MB). Source reuses existing FastEmbed cache directories but exposes no CLI prefetch or no-download mode; `TextEmbedding::try_new` initializes the model and downloads missing files. A namespace listing reached initialization and stopped at proxy HTTP CONNECT 403 for `huggingface.co`. No repository or issue content was ingested, no database/namespace was created, and no daemon was started. No model request was retried.
- NeuroPlasticity's `test … --print-egress-plan` ran on a disposable synthetic no-op manifest with `engine: podman`, `base_image: alpine:latest`, no external evaluator, and embedded optimizer. The output was `Data class: unclassified (no enforcement applies)` and `embedded → ALLOWED (local inference, nothing leaves the machine)`. This is an application-level provider plan, not a registry/image network check or firewall. Source supports Podman and falls back to Docker; this machine had Docker Engine 28.4.0 but no Podman executable and no cached images. Pulling `alpine:latest` was not approved and was not attempted. The run command stopped before container startup because the disposable manifest placed `target_rules_file` outside the project. A corrected manifest would still need a locally available base image. If an evaluator fails, the optimizer's default Qwen model would be needed: `https://huggingface.co/unsloth/Qwen3-4B-Instruct-2507-GGUF/resolve/main/Qwen3-4B-Instruct-2507-Q4_K_M.gguf` (domain `huggingface.co`, approximately 2.5 GB, documented SHA-256 `3605803b982cb64aead44f6c1b2ae36e3acdb41d8e46c8a94c6533bc4c67e597`; upstream path follows mutable `main`). No Qwen model was requested or downloaded. No model request was retried.
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

The pinned go-licenses v2.0.1 executable reported `GPL-3.0` for a synthetic forbidden dependency. Its check returned exit 0 when `GPL-3.0` was appended to the allowlist, and `go test ./internal/devsecops` still passed. The unit policy test lists `GPL-3.0-only` but not `GPL-3.0` among explicitly rejected allowlist entries (`internal/devsecops/policy_test.go:307-310`). The separate smoke script uses its own hard-coded permissive allowlist; it correctly tests scanner behavior but does not consume the checked-in project allowlist, so it does not detect that mutation. The current project dependency graph was not found to include the GPL fixture, and the checked-in allowlist was not changed.

**Acceptance test:** extend the workflow policy model/assertions to reject disabling conditions and job-level failure suppression; include the scanner's actual `GPL-3.0` identifier in the negative allowlist test. Ensure the license smoke test or a second test exercises the actual project allowlist rather than a duplicate hard-coded list.

#### F-04 — Round-trip comparison shares its writer context and omits the PDF catalog (P2)

`internal/pdfpreserve/roundtrip_test.go:27-62` passes the `before` context to the writer and later compares that same context with the output. A disposable probe changed the protected page object's `/StructParents` value in the writer context just before serialization: the same-context comparison passed while an independently reread input context detected the difference. This demonstrates an oracle-coupling false negative if a writer mutates its input context; it does **not** establish that the current writer does so. The test also checks object IDs 3–16 (with streams checked separately) but omits object 1, the catalog. The synthetic catalog contains `/Lang`, `/MarkInfo`, `/StructTreeRoot`, and `/AcroForm` links. A current-fixture read showed those values retained, but a future catalog regression would not fail this test.

**Acceptance test:** retain an immutable independently read input snapshot separate from the writer context; add a deliberate pre-write protected-value mutation control that the preservation comparator detects; compare the catalog and assert accessibility-relevant keys and form linkage survive; mutation tests that remove/change each key must fail. This is fidelity coverage, not PDF/UA conformance evidence.

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

- At the reviewed commit, no authenticated veraPDF 1.30.3 execution or genuine known-pass/known-fail PDF oracle had been run; synthetic reports only exercised the adapter parser. After this pinned review, open draft PR #7 at `c8e1a5b0a2884204453f11f0f558d8f73e0f1a9e` completed signed veraPDF 1.30.3 acquisition and UA-2 pass/fail controls. That later integration is outside this report pin and remains unmerged, so it does not change the baseline findings or constitute a claim of full product conformance.
- No human semantic or assistive-technology review was performed. The repository correctly keeps these as separate gates; no false PDF/UA-2 or WCAG conformance claim was found.
- No concurrency defect was confirmed in the bounded output budget/process tree implementation. The focused race suite passed.
- Resource limits for memory/CPU remain a design assumption until the adapter is used on hostile files.
- GitHub reported seven successful checks on the preceding published PR head, but branch-protection settings could not be read through the connected integration. Green checks are observed; their formal required status is unverified.
- The opened PR #5 and the separate veraPDF integration work were outside this immutable review SHA.

### Rejected interpretations and dissent

- **The current writer was shown to mutate its context:** not established. The original probe only checked catalog object 1 and saw no mutation. A broader deliberate-mutation control confirms that the current same-context test pattern can miss a changed protected object; this is an oracle-independence test gap, not evidence of current output corruption. Catalog omission is a separate confirmed assertion gap.
- **No corpus artifacts proves the corpus validator is broken:** rejected. The pinned manifest explicitly records planned synthetic cases, not product execution evidence.
- **The adapter or tagged fixture proves PDF/UA-2 conformance:** rejected. No real validator or human/AT evidence is present in this panel.
- **The Go job is unbounded:** rejected as stated; P3 reflects a missing explicit job bound while retaining GitHub's finite default.
- **No concurrency bug means resource bounds are complete:** rejected. Race safety and memory/CPU quotas are separate questions.
- No independent reviewer dissent was recorded after adjudication. At the pinned baseline, real-validator controls were missing; PR #7 later added them outside this review. Human/AT evidence and full NeuroStrata/NeuroPlasticity-assisted review remain incomplete.

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
| Temporary round-trip probe | Current fixture's catalog matched after serialization. A second probe changed page object 3 `/StructParents` in the writer context immediately before serialization: comparing that context to output passed, while an independent reread of input differed. This proves the same-context oracle can miss a mutation; it does not show the current writer mutates its context. Probe files removed; recipe is preserved in the reproduction appendix. |
| Four temporary policy mutations + `go test ./internal/devsecops` | Tests passed despite Gitleaks disabled, govulncheck disabled, job-level continue-on-error, and GPL-3.0 allowlist addition. Mutations restored. |
| `go-licenses check --include_tests --allowed_licenses=Apache-2.0,BSD-3-Clause,GPL-3.0,MIT ./...` on synthetic GPL-3.0 module using pinned go-licenses v2.0.1 | Scanner emitted `GPL-3.0`; allowing that emitted identifier returned exit 0. Synthetic module/cache outside repository. |
| Actual project allowlist command from `security-scans.yml`; `bash .github/scripts/test-go-licenses.sh` | Both passed with the current allowlist; smoke test rejected its synthetic test-only GPL dependency. |
| NeuroStrata 1.9.0 `status` | Passed model-free; no DB file and daemon down. `namespaces` attempted model initialization then proxy CONNECT 403; no retry. |
| NeuroPlasticity 2.1.0 `test /tmp/neuroplasticity-review.json --print-egress-plan` | Egress plan printed `Data class: unclassified (no enforcement applies)` and embedded/local allowed. The manifest requested Podman and `alpine:latest`; no container started. |
| NeuroPlasticity no-op `test` attempt | Stopped before container startup with the safe-relative-target-rules error. Podman is absent, Docker fallback is available, no image is cached, and no image pull/model download was attempted. |
| Beads v1.3.1 isolated import/create/export and JSONL integrity check | Imported 44 snapshot issues; created six linked follow-ups; exported 50. Verified all 44 original objects were unchanged, 6 new IDs are unique, and every dependency target resolves. Metrics off; no remote configured and no sync run. |
| `git diff --check` | Passed after report/export update. |

Versions: Git 2.52.0, Go 1.27.2, Beads 1.3.1, NeuroStrata 1.9.0, NeuroPlasticity 2.1.0, Docker Engine 28.4.0. No Podman runtime or container images were available. Exact executable release hashes and official provenance are recorded in [build-machine-tooling.md](../planning/build-machine-tooling.md).

### Published PR check evidence

PR #6 remains open and in draft state. At the 2026-10-12 verification checkpoint, report/export head `117449336699232073730d70edf99a3913706916` had seven successful checks: CodeQL, secrets, checks, go-security, Review dependency changes, Analyze (actions), and Analyze (go). GitHub CI associated the checks with synthetic PR merge commit `9b441f4f42b4045d7683f71ddbbffcb822072990` (parents `70728c6a` and `117449336699232073730d70edf99a3913706916`); they validate the PR merge result, not the original main pin in isolation. Local Go/probe evidence was run separately against `595e7e9224fc6f22abb75e9f706dc2558514f620`. The connected GitHub integration returned 403 when reading branch protection, so exact required-check settings remain unverified. The user authorized autonomous merge after the independent review verdict and required CI pass; keep this PR unmerged while the reviewer verdict is pending. The attempted `gh pr edit 6 --body-file /tmp/pr6-body.md` returned `Post "https://api.github.com/graphql": Forbidden` (HTTP 403/permission denial, not a transient response). No alternate endpoint was tried; the PR body remains stale.

## Review boundary and publication

This report adjudicates only `595e7e9224fc6f22abb75e9f706dc2558514f620`. PR #5 and the separate veraPDF integration changes were outside the pin. The results distinguish executed tests from static policy gaps, synthetic parser behavior, plausible resource risks, and untested genuine validator/human accessibility evidence. Keep PR #6 in draft and unmerged pending the independent reviewer verdict.

## Appendix: minimal reproduction recipes

These recipes use a disposable worktree at the reviewed SHA and synthetic fixtures only. Save temporary Go tests as `review_repro_test.go`, run them with the command shown, and remove the file. Product tests and CI configuration are not modified.

### XML pass classification and diagnostic leakage

In `internal/pdfua/review_repro_test.go`, add:

```go
package pdfua

import (
	"strings"
	"testing"
)

func TestReviewAmbiguousAndEmptyAttributes(t *testing.T) {
	duplicate := strings.Replace(report("true"), `isCompliant="true"`, `isCompliant="false" isCompliant="true"`, 1)
	if _, compliant, err := parseReport([]byte(duplicate)); err != nil || !compliant {
		t.Fatalf("duplicate attribute result: compliant=%t err=%v", compliant, err)
	}
	for _, attr := range []string{`failedToParse="0"`, `failedJobs="0"`} {
		empty := strings.Replace(attr, `="0"`, `=""`, 1)
		input := strings.Replace(report("true"), attr, empty, 1)
		if _, compliant, err := parseReport([]byte(input)); err != nil || !compliant {
			t.Fatalf("empty %s result: compliant=%t err=%v", attr, compliant, err)
		}
	}
}

func TestReviewParseErrorEchoesChildOutput(t *testing.T) {
	input := strings.Replace(report("true"), `totalJobs="1"`, `totalJobs="SENSITIVE_MARKER"`, 1)
	_, _, err := parseReport([]byte(input))
	if err == nil || !strings.Contains(err.Error(), "SENSITIVE_MARKER") {
		t.Fatalf("expected synthetic marker in current parse error; got %v", err)
	}
}
```

Run `go test ./internal/pdfua -run 'TestReview(AmbiguousAndEmptyAttributes|ParseErrorEchoesChildOutput)' -v`.

### Coupled round-trip baseline control

In `internal/pdfpreserve/review_repro_test.go`, add:

```go
package pdfpreserve

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestReviewCoupledBaselineMutation(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "in.pdf"), filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(input, syntheticPDF("statement"), 0o600); err != nil { t.Fatal(err) }
	independent, err := api.ReadContextFile(context.Background(), input)
	if err != nil { t.Fatal(err) }
	writerCtx, err := api.ReadContextFile(context.Background(), input)
	if err != nil { t.Fatal(err) }
	page := writerCtx.Table[3].Object.(types.Dict)
	page["StructParents"] = types.Integer(99)
	if err := api.WriteContextFile(context.Background(), writerCtx, output); err != nil { t.Fatal(err) }
	after, err := api.ReadContextFile(context.Background(), output)
	if err != nil { t.Fatal(err) }
	t.Logf("same-context comparison=%t", reflect.DeepEqual(writerCtx.Table[3].Object, after.Table[3].Object))
	t.Logf("independent-input comparison=%t", reflect.DeepEqual(independent.Table[3].Object, after.Table[3].Object))
	if !reflect.DeepEqual(writerCtx.Table[3].Object, after.Table[3].Object) || reflect.DeepEqual(independent.Table[3].Object, after.Table[3].Object) {
		t.Fatal("expected coupled baseline to miss synthetic mutation while independent baseline detects it")
	}
}
```

Run `go test ./internal/pdfpreserve -run TestReviewCoupledBaselineMutation -v`. This intentional synthetic mutation control does not assert that pdfcpu currently mutates its context.

### Relative TempDir invocation path

In `internal/pdfua/review_repro_test.go`, add the following test (the package's existing fake helpers are available):

```go
package pdfua

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewRelativeTempDirArgResolution(t *testing.T) {
	if err := os.Mkdir("review-relative-temp", 0o700); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("review-relative-temp")
	pdf := fakePDF(t)
	capture := filepath.Join(t.TempDir(), "args")
	t.Setenv("PDFUA_FAKE_REPORT", report("true"))
	t.Setenv("PDFUA_FAKE_EXIT", "0")
	t.Setenv("PDFUA_FAKE_CAPTURE", capture)
	options := fakeOptions(t, fakeArtifact(t))
	options.TempDir = "review-relative-temp"
	if _, err := Validate(context.Background(), options, pdf); err != nil {
		t.Fatal(err)
	}
	argsData, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Split(string(argsData), "\t")
	jarIndex := -1
	for i, arg := range args {
		if arg == "-jar" {
			jarIndex = i
			break
		}
	}
	if jarIndex < 0 || jarIndex+1 >= len(args) {
		t.Fatalf("missing -jar argument: %q", args)
	}
	jarArg := args[jarIndex+1]
	if !strings.HasPrefix(jarArg, "review-relative-temp/pdfua-validation-") {
		t.Fatalf("expected relative isolated JAR path, got %q", jarArg)
	}
	childCwd := filepath.Dir(jarArg)
	resolvedFromChildCwd := filepath.Join(childCwd, jarArg)
	expectedPrefix := filepath.Join(childCwd, childCwd) + string(os.PathSeparator)
	if !strings.HasPrefix(resolvedFromChildCwd, expectedPrefix) {
		t.Fatalf("expected child cwd to repeat the relative prefix, got %q", resolvedFromChildCwd)
	}
}
```

Run `go test ./internal/pdfua -run TestReviewRelativeTempDirArgResolution -v`. The fake process records the invocation but does not open a JAR, so this does not establish actual JVM behavior.

### Workflow-policy mutations

In a disposable copy, apply one change at a time: add `if: ${{ false }}` to the Gitleaks full-history step; add the same condition to govulncheck; add job-level `continue-on-error: true` to `go-security`; append `GPL-3.0` to `.security/allowed-go-licenses.txt`. After each mutation run `go test ./internal/devsecops`. All four mutations passed the current policy tests. Restore the file after each case. The synthetic `go-licenses` control used the scanner-emitted identifier `GPL-3.0` and returned exit 0 when that identifier was in `--allowed_licenses`.
