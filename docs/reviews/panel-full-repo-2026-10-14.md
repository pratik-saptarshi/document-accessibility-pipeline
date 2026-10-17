# Adversarial panel review — full repository (2026-10-14)

**Revision:** `1bc7336` (same as `origin/main` at review time).
**Method:** four independent reviewers (architecture, code quality, test coverage, best practices); a judge verified, merged, and ranked the findings. The panel changed nothing in the repository.
**Result:** no critical defects. Two P1 problems need fixing first (below).
**Verification:** all Go tests passed including with `-race`, and `traceability-check` passed. The judge checked every High and Medium claim against the source; three key claims were independently re-checked. CI results for the reviewed revision were not checked.

## Prioritized findings

| P | Finding | Raised by | Evidence |
|---|---|---|---|
| **P1** | Validator reports are kept for 1 day, so evidence for the open release gates (`dap-jo9.1.6.3` and `.1.6.4`) expires before review. A research note already records this happening. → `dap-jo9.1.6.10` | Architecture | `.github/workflows/verapdf-integration.yml:80` |
| **P1** | The main Go CI workflow (`go-checks.yml`) never runs the adapter module's tests, `go vet` or `go mod tidy -diff`, or the bootstrap module's `tidy -diff`. The policy test doesn't require them either. The adapter tests only run behind the Java validator job. → `dap-jo9.5.6` | Architecture, test coverage, best practices | `go-checks.yml:26-42`, `internal/devsecops/policy_test.go:1030-1037` |
| P2 | The tests for how validator reports are parsed live only in the adapter module, which CI rarely runs. One branch (an `isCompliant` value that is neither true nor false) is tested nowhere. Folded into `dap-jo9.5.6`. | Test coverage | `internal/pdfua/report.go:139` |
| P2 | No fuzz tests exist anywhere, although `AGENTS.md` asks for them on hostile-input parsers. → `dap-jo9.2.1.4` | Code quality, test coverage | `ParseReport`, `VerifyBundle`, `corpus.Load`, the bootstrap manifest |
| P2 | Several docs contradict the code. → `dap-jo9.5.7` | Best practices, architecture | see list below |
| P2 | Fixture-generation code reads a font file from inside pdfcpu's module-cache directory. The file is hash-checked, but `AGENTS.md` says pdfcpu is used only in tests. Assessed as accepted risk: the path is Go's module cache (not pdfcpu-specific), the read is hash-checked, and fixture authoring falls under the test bucket. | Architecture | `internal/financialfixtures/author_v2.go:193` |
| P3 | `corpus.Load` accepts duplicate keys and keys that differ only in case. A reviewer showed `"PROVENANCE":"synthetic-only"` silently overriding `"provenance":"private-customer-reports"`. → `dap-jo9.2.1.4` | Code quality | `internal/corpus/manifest.go:70-77` |
| P3 | The cloud-tools bootstrap manifest has loose checks (see below). The manifest is checked in, so the risk is low. → `dap-jo9.2.1.4` | Code quality, test coverage | `manifest.go:83,103`, `bootstrap.go:426,570` |
| P3 | Process-group cleanup is written four separate ways. Two copies signal the process group after the leader has been reaped, which leaves a narrow PID-reuse window. → `dap-jo9.5.7` | Architecture, code quality | `ci/pdfua2/adapter/process_linux.go`, `cmd/verapdf-container-runner/process_linux.go` |
| P3 | Fixture publish and rollback have little coverage (rollback 0%), and the container runner's command-line entry point is untested. Not yet tracked; nominated for a follow-up item. | Test coverage | `author_v2.go:666-772`, `cmd/verapdf-container-runner/main.go:200` |
| P3 | Smaller issues (see below). Not yet tracked. | Several | |

Doc contradictions (→ `dap-jo9.5.7`):
- `SECURITY.md:17` says the security-scan job runs tests and `go vet`; it runs neither.
- `CONTRIBUTING.md:87` says pdfcpu is the only direct dependency, but `go.mod` lists three.
- `docs/validation/README.md:14` says a test keeps process execution out of the root module, but that test (`internal/pdfua/boundary_test.go`) only scans `internal/pdfua`.

Bootstrap manifest checks (→ `dap-jo9.2.1.4`):
- The `version` field isn't sanitized, so `../` in it can place downloads outside the cache directory.
- The version pin is a substring check, so `1.9` accepts `@v1.9.0`.
- Archive extraction silently truncates at 256 MiB instead of failing.

Smaller issues (not yet tracked):
- `LoadCatalog` silently drops misspelled fields.
- The dependency-cycle error can name a different node from run to run.
- One error message can print "trailing JSON data: <nil>" (`author_v2.go:625`).
- A few tests use fixed sleeps that could flake on a slow CI runner.

## Claims the judge downgraded or rejected
- **Hard-coded candidate hashes in `bundle.go`:** this pinning is intentional. Keeping the validator identity in one place would still be good.
- **Experiment tests that never run:** they are opt-in by design and documented as such.
- **"`version_source` can never verify":** the only manifest entry using it is a raw binary, where the check is correct, so this is a latent issue only.
- **Report-parsing coverage rated High:** downgraded to P2, because the adapter module's tests cover all but one branch. The real problem is that CI rarely runs them.

## Not reviewed by anyone
The source-publisher script, the line-by-line container flags in `run-verapdf-isolated.sh`, whether the docs separate the generation and remediation workflows, and whether these findings match existing Beads issues.

## Correction
The judge wrote that process-group descendant cleanup is "deferred under `dap-jo9.5.3`". That issue is actually closed, so that cleanup has no open issue tracking it; it is now tracked under `dap-jo9.5.7`.
