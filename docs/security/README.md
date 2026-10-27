# Security scanning operations

Status: **Proposed repository checks**. These checks cover repository source and history; they do not establish the security of a future PDF-processing runtime.

The [2026-10-13 adversarial panel](adversarial-panel-2026-10-13-a6d6fc6.md) audits exact SHA `a6d6fc65e88351d710b3d8d4229fe9721c984884` and records follow-up state at then-current main `23b11d34d6654685f47e43853ef0d077350a5183` (PR30). Its five lenses cover isolation, PDF/UA-2 and semantics, Go reliability/resources, supply chain/CI, and test/roadmap gaps. All five tracked follow-up issues were closed in that snapshot; real validator controls remain CI-only. The earlier report [adversarial-panel-2026-10-12-a6d6fc6.md](adversarial-panel-2026-10-12-a6d6fc6.md) contains the original dated evidence and the later verified required-check status.

Neither report changes rulesets. The later verified required-check set contains nine checks, including `pdfua2` and `validator-container`, as recorded in the Oct 12 report. No actual local veraPDF container or full Neuro-assisted execution is claimed.

The [earlier adversarial panel at `595e7e9`](adversarial-panel-595e7e9.md) records the prior findings and regressions. The current report rechecks those findings against main after PR14.

## Workflow

`.github/workflows/security-scans.yml` runs on pull requests targeting `main`, pushes to `main`, and weekly (`17 05 * * 1` UTC). It sets `contents: read`, uses immutable action commit SHAs with release comments, applies job timeouts, and cancels superseded runs.

The separate `checks` job in `.github/workflows/go-checks.yml` tests and vets the root module. The isolated `go-security` job in this workflow runs actionlint and checks vulnerabilities with `govulncheck -test ./...` in the root, `cmd/cloud-tools-bootstrap`, and `ci/pdfua2/adapter` Go modules. Each scan uses ordinary text mode: JSON, SARIF, and OpenVEX output modes can return success even when they report findings. The `go-security` job also runs the license checks via `scripts/check-licenses.sh`: root and adapter `go-licenses check --include_tests` against the explicit SPDX identifiers in `.security/allowed-go-licenses.txt`, the bootstrap module's stdlib-only assertion, and the script behavior test (`.github/scripts/test-check-licenses.sh`). The same script is the local entry point for the license checks. Unknown licenses and licenses absent from the allowlist must fail the job. The license smoke test uses a temporary local module imported only by a test: its production-only scan passes, while the test-inclusive scan identifies and rejects its GPL-3.0 license. The copied license text fixture is documented in [the fixture directory](../../.github/scripts/fixtures/README.md) and is not a product dependency.

The current pinned tool versions are:

| Tool | Version | Reference |
| --- | --- | --- |
| govulncheck | v1.9.0 | [Go vulnerability tool](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |
| go-licenses | v2.0.1 | [go-licenses](https://github.com/google/go-licenses) |
| actionlint | v1.7.12 | [actionlint releases](https://github.com/rhysd/actionlint/releases) |
| Gitleaks | v8.30.1 Linux x64 release archive, SHA-256 pinned | [Gitleaks v8.30.1 release](https://github.com/gitleaks/gitleaks/releases/tag/v8.30.1) |

Action SHAs and displayed release versions are checked by `internal/devsecops` policy tests. Confirm a release's provenance and upstream repository before changing a pin. Gitleaks is downloaded from its official release and checked against the pinned SHA-256 before installation; the Go module source for v8.30.1 did not match the public Go checksum database in this environment, so CI deliberately uses the verified release archive instead of bypassing Go checksum verification.

## Secret scanning and exceptions

Gitleaks checks all Git refs/history from a full-depth checkout. The full-history scan uses `--redact=100`, writes no report, and preserves the scanner's normal nonzero exit status for findings and errors. It explicitly loads `.security/gitleaks.toml`, which only extends Gitleaks' built-in rules, and sets `--gitleaks-ignore-path=/dev/null`. Because Gitleaks can also auto-load `.gitleaksignore` files found under the scan source, CI first runs a guard that rejects any working-tree `.gitleaksignore` or `.gitleaks.toml` file. `--ignore-gitleaks-allow` also disables inline allow comments. The policy test requires the guard step and exact config/scan flags; adding custom rules or allowlists fails it. The structured exception registry remains empty and is not a scanner suppression mechanism.

The Gitleaks smoke script uses only ephemeral synthetic fixtures; its redacted JSON finding report stays in a temporary directory that is removed on exit. It checks clean input, a synthetic finding, report redaction, invalid configuration behavior, and both forms of repository suppression: a temporary `.gitleaks.toml` allowlist and `.gitleaksignore` fingerprint. The suppression controls use the same working directory and source path as the resulting scan, and the guard is verified to reject each unmanaged file before the hardened scan.

`.security/secret-exceptions.yaml` currently contains no exceptions. Repository suppressions remain disabled for this tranche; the registry does not create scanner bypasses. A future policy change for a synthetic fixture would need an exact file under `internal/devsecops/testdata/gitleaks/`, a verified Gitleaks rule ID, a reviewed maintainer owner, a linked Beads issue, reason, and future expiry. The current tests check metadata presence/shape, fixture-only paths, duplicate IDs, and future dates; they do not independently verify rule IDs or owner identities. Do not create exceptions for live credentials, customer material, or private reports. Never print or commit a suspected real credential; stop and contact the repository owner privately.

No exception file is currently consumed by Gitleaks as a suppression list. It is a reviewed registry for tracking narrowly scoped synthetic fixture exceptions, and the scan intentionally stays fail-closed.

## Scanner suppression census

A follow-up census (PR #196) enumerated the shell and YAML forms that can
suppress a pinned scanner invocation's findings or exit status, instead of
only fixing the one reported string. Each form was probed by mutating the
real `security-scans.yml` steps and running the `internal/devsecops` policy
validators; the "Flagged before" column is that probe's result.

Stream facts the classification rests on (govulncheck v1.9.0 source,
`internal/scan/run.go`): govulncheck's text handler writes findings to
**stdout** and errors to **stderr**, exiting 3 on findings. Gitleaks writes
its findings log to **stderr**, exiting 1 on leaks. The policy therefore
treats a redirection that sends **either** stream to a sink (a file,
`/dev/null`, or a closed descriptor) as suppression, simulated in shell
order with descriptor state — so `2>&1 1>/dev/null` is suppression (stderr
reaches the log, but the stdout findings are still discarded), while a bare
`2>&1` or `1>&2`, which keeps both streams on the step log, is accepted and
pinned as accepted in the tests. Detection is token-aware
(`shellTokens`/`scannerShellSuppression` in `policy_test.go`), not substring
matching. The nested govulncheck steps and other exactly pinned steps
already rejected every appended form before this census; the open class was
on the substring-pinned root govulncheck and Gitleaks steps probed here.

| Family | Form | Flagged before? | Disposition |
| --- | --- | --- | --- |
| Exit neutralization | `scanner \|\| true` | Yes (literal substring) | Kept; now token-aware on every step |
| Exit neutralization | `scanner \|\| exit 0`, `scanner \|\| :` | No | **Fixed** — generic fail-open list rule, every step |
| Exit neutralization | `scanner \|\| <any command>` (e.g. `\|\| echo ok`) | No | **Fixed** — any `\|\|` list on a scanner invocation is rejected |
| Exit neutralization | `scanner \|\| exit` (bare), `scanner \|\| false` | No | Flagged by the same any-`\|\|` rule as non-canonical condition context; neither form alone yields a silent success (bare `exit` carries the scanner status, `false` fails with the wrong status), and the census records that distinction rather than enumerating right-hand spellings |
| Scanner flags | `--exit-code <any>`, `--baseline-path <any>` | Yes (PR #196) | Kept; whole flag family, any spelling/value |
| Redirection | `2>/dev/null` | No | **Fixed** — stderr sink (Gitleaks findings log; govulncheck errors) |
| Redirection | `>/dev/null`, `1>/dev/null`, `1>/dev/null 2>&1`, `&>/dev/null` | No | **Fixed** — stdout/both-stream sink (govulncheck findings) |
| Redirection | `2>&1 1>/dev/null` | No | **Fixed** — descriptor simulation: stdout still discarded |
| Redirection | `&>file`, `>file 2>&1`, `>file`, `2>file`, spaced forms (`2> /dev/null`), `2>&-` | No | **Fixed** — sink on either stream |
| Redirection | `exec 2>/dev/null` on an earlier line of the step | No | **Fixed** — exec redirection persists into the scanner's segment |
| Redirection | bare `2>&1`, `1>&2` | No (not suppression) | **Accepted** — both streams stay on the step log; pinned as accepted tests |
| Pipes | `\| true`, `\| cat`, `\| tee file`, `\| tee /dev/null` | No | **Fixed** — any pipe on a scanner invocation; the pinned scanner steps set no explicit shell, so no pipefail is guaranteed and the step status would come from the downstream command |
| Shell control | `set +e` / `set +o errexit` prefix | No | **Fixed** — generic errexit rule, every step |
| Shell control | `if scanner; then …`, `! scanner` | No | **Fixed** — condition/negation context on a scanner invocation |
| Shell control | `out=$(scanner)`, backtick capture | No | **Fixed** — command substitution on a scanner invocation |
| Shell control | `( scanner \|\| true )` | Yes (inner `\|\| true`) | Kept; also covered by the scanner `\|\|` rule |
| Shell control | `( scanner )` bare subshell, `scanner; echo`, `scanner && echo` | No (not suppression) | **Accepted** — under the pinned default shell (`bash -e`) the step status still carries the scanner's nonzero status; pinned as accepted tests |
| Shell control | `scanner &` (background) | No | **Fixed** — step never waits for the scanner |
| Gitleaks flags | duplicate `--log-opts` (any value, including repeating `--all`), `--config`, `--redact`, `--gitleaks-ignore-path`, `--timeout` | No | **Fixed** — each pinned value flag must occur exactly once; flag parsing keeps the last occurrence, so a duplicate overrode the pinned value despite every required substring being present |
| YAML level | step `continue-on-error: true`, job `continue-on-error: true`, step/job `if:` | Yes | Kept |
| YAML level | `continue-on-error: ${{ true }}` or `"true"` (step or job) | Rejected at parse | Kept and pinned — the modeled fields are booleans, so expression/quoted forms fail closed in the strict workflow parser |

Reviewer items resolved with this census: the exit-code asymmetry (the
shell side knew only the literal `|| true` while the Gitleaks side matched
the whole `--exit-code` family) is resolved by unifying the shell side at
the operator level for scanner invocations, with the rationale recorded in
the `validateScannerSteps` comment; that comment now states the implemented
contract exactly. Two further reviewer references did not match this tree:
there is no `internal/devsecops/doc.go` or `internal/devsecops/policy.go`
(the validators live in `policy_test.go`), so the truncated-sentence and
comment-scope items were applied to the corresponding comment in
`policy_test.go` instead. A govulncheck scan-level narrowing form
(`-scan=module` appended to the pinned invocation) was probed unflagged; it
belongs to a separate flag-value census family and is filed as a follow-up,
not fixed here.

## Local reproduction

Run from the repository root with the pinned tools installed:

```sh
go test -race ./...
go vet ./...
actionlint
govulncheck -test ./...
(cd cmd/cloud-tools-bootstrap && govulncheck -test ./...)
(cd ci/pdfua2/adapter && govulncheck -test ./...)
bash scripts/check-licenses.sh
bash .github/scripts/test-go-licenses.sh
bash .github/scripts/test-gitleaks.sh
if [ "$(git rev-parse --is-shallow-repository)" = true ]; then git fetch --unshallow; fi
gitleaks git --log-opts=--all --redact=100 --ignore-gitleaks-allow --config=.security/gitleaks.toml --gitleaks-ignore-path=/dev/null --no-banner --timeout=300 .
```

For a fresh clone, fetch all refs/history before the final Gitleaks command. The workflow does that with checkout `fetch-depth: 0`. Gitleaks findings are security-sensitive; retain redaction and do not paste unredacted tool output into an issue or chat.

## License-check parity

The script exports `GOTOOLCHAIN=local` (plus `GOWORK=off`, `-mod=readonly`, and `GOROOT` taken from the runtime `go`): with toolchain auto-switching, go-licenses can be handed a mixed stdlib/module view and fails with `Package <stdlib> does not have module info` on a compliant graph (observed locally 2026-10-26, Go 1.27.2 against a `go.mod` pinning go1.27.3). CI avoids this because setup-go installs the pinned toolchain and installs go-licenses with it.

Under `GOTOOLCHAIN=local`, each module's `go` directive is a hard floor: a local Go older than a module's directive fails closed. Full local parity needs the pinned Go installed (root `go.mod` `toolchain` line); without it, `root` and `bootstrap` still check locally and `adapter` fails at its floor — run `bash scripts/check-licenses.sh` and read which target failed.

**Bump tripwire (advisory, extends Decision 0022 condition 5):** any change to a `go`/`toolchain` directive or to the go-licenses pin must run `bash scripts/check-licenses.sh` locally and pass before the change is pushed; CI's floor is setup-go's resolved toolchain, so a directive bump ahead of setup-go's resolution fails `go-security` until it catches up.

## CodeQL distinction

The repository's CodeQL default setup is configured in GitHub repository settings; its runs are identified in GitHub under `dynamic/github-code-scanning/codeql`. The checked-in `.github/workflows/codeql.yml` is manually disabled in Actions settings and is not the active CodeQL execution path. The security-scans workflow does not enable or change CodeQL settings. Confirm runs and alerts in GitHub rather than treating the disabled YAML file as evidence of execution.
