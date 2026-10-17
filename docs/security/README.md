# Security scanning operations

Status: **Proposed repository checks**. These checks cover repository source and history; they do not establish the security of a future PDF-processing runtime.

The [2026-10-13 adversarial panel](adversarial-panel-2026-10-13-a6d6fc6.md) audits exact SHA `a6d6fc65e88351d710b3d8d4229fe9721c984884` and records follow-up state at then-current main `23b11d34d6654685f47e43853ef0d077350a5183` (PR30). Its five lenses cover isolation, PDF/UA-2 and semantics, Go reliability/resources, supply chain/CI, and test/roadmap gaps. All five tracked follow-up issues were closed in that snapshot; real validator controls remain CI-only. The earlier report [adversarial-panel-2026-10-12-a6d6fc6.md](adversarial-panel-2026-10-12-a6d6fc6.md) contains the original dated evidence and the later verified required-check status.

Neither report changes rulesets. The later verified required-check set contains nine checks, including `pdfua2` and `validator-container`, as recorded in the Oct 12 report. No actual local veraPDF container or full Neuro-assisted execution is claimed.

The [earlier adversarial panel at `595e7e9`](adversarial-panel-595e7e9.md) records the prior findings and regressions. The current report rechecks those findings against main after PR14.

## Workflow

`.github/workflows/security-scans.yml` runs on pull requests targeting `main`, pushes to `main`, and weekly (`17 05 * * 1` UTC). It sets `contents: read`, uses immutable action commit SHAs with release comments, applies job timeouts, and cancels superseded runs.

The separate `checks` job in `.github/workflows/go-checks.yml` tests and vets the root module. The isolated `go-security` job in this workflow runs actionlint and checks vulnerabilities with `govulncheck -test ./...` in the root, `cmd/cloud-tools-bootstrap`, and `ci/pdfua2/adapter` Go modules. Each scan uses ordinary text mode: JSON, SARIF, and OpenVEX output modes can return success even when they report findings. The `go-security` job also runs `go-licenses check --include_tests` against the explicit SPDX identifiers in `.security/allowed-go-licenses.txt` for the root and adapter modules. Unknown licenses and licenses absent from the allowlist must fail the job. The license smoke test uses a temporary local module imported only by a test: its production-only scan passes, while the test-inclusive scan identifies and rejects its GPL-3.0 license. The copied license text fixture is documented in [the fixture directory](../../.github/scripts/fixtures/README.md) and is not a product dependency.

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

## Local reproduction

Run from the repository root with the pinned tools installed:

```sh
go test -race ./...
go vet ./...
actionlint
govulncheck -test ./...
(cd cmd/cloud-tools-bootstrap && govulncheck -test ./...)
(cd ci/pdfua2/adapter && govulncheck -test ./...)
allowed_licenses="$(grep -Ev '^[[:space:]]*(#|$)' .security/allowed-go-licenses.txt | paste -sd, -)"
go-licenses check --include_tests --allowed_licenses="$allowed_licenses" ./...
bash .github/scripts/test-go-licenses.sh
bash .github/scripts/test-gitleaks.sh
if [ "$(git rev-parse --is-shallow-repository)" = true ]; then git fetch --unshallow; fi
gitleaks git --log-opts=--all --redact=100 --ignore-gitleaks-allow --config=.security/gitleaks.toml --gitleaks-ignore-path=/dev/null --no-banner --timeout=300 .
```

For a fresh clone, fetch all refs/history before the final Gitleaks command. The workflow does that with checkout `fetch-depth: 0`. Gitleaks findings are security-sensitive; retain redaction and do not paste unredacted tool output into an issue or chat.

## CodeQL distinction

The repository's CodeQL default setup is configured in GitHub repository settings; its runs are identified in GitHub under `dynamic/github-code-scanning/codeql`. The checked-in `.github/workflows/codeql.yml` is manually disabled in Actions settings and is not the active CodeQL execution path. The security-scans workflow does not enable or change CodeQL settings. Confirm runs and alerts in GitHub rather than treating the disabled YAML file as evidence of execution.
