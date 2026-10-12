# Security scanning operations

Status: **Proposed repository checks**. These checks cover repository source and history; they do not establish the security of a future PDF-processing runtime.

## Workflow

`.github/workflows/security-scans.yml` runs on pull requests targeting `main`, pushes to `main`, and weekly (`17 05 * * 1` UTC). It sets `contents: read`, uses immutable action commit SHAs with release comments, applies job timeouts, and cancels superseded runs.

The Go job tests and vets all Go packages, runs actionlint, and checks vulnerabilities using `govulncheck -test ./...`. Keep govulncheck in its ordinary text mode: its JSON, SARIF, and OpenVEX output modes can return success even when they report findings. It also runs `go-licenses check --include_tests` against the explicit SPDX identifiers in `.security/allowed-go-licenses.txt`. Unknown licenses and licenses absent from the allowlist must fail the job.

The current pinned tool versions are:

| Tool | Version | Reference |
| --- | --- | --- |
| govulncheck | v1.9.0 | [Go vulnerability tool](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |
| go-licenses | v2.0.1 | [go-licenses](https://github.com/google/go-licenses) |
| actionlint | v1.7.12 | [actionlint releases](https://github.com/rhysd/actionlint/releases) |
| Gitleaks | v8.30.1 Linux x64 release archive, SHA-256 pinned | [Gitleaks v8.30.1 release](https://github.com/gitleaks/gitleaks/releases/tag/v8.30.1) |

Action SHAs and displayed release versions are checked by `internal/devsecops` policy tests. Confirm a release's provenance and upstream repository before changing a pin. Gitleaks is downloaded from its official release and checked against the pinned SHA-256 before installation; the Go module source for v8.30.1 did not match the public Go checksum database in this environment, so CI deliberately uses the verified release archive instead of bypassing Go checksum verification.

## Secret scanning and exceptions

Gitleaks checks all Git refs/history from a full-depth checkout. Output uses `--redact=100`, no report is written, and the scanner's normal nonzero exit status is preserved for findings and errors. `--ignore-gitleaks-allow` prevents inline allow comments from suppressing findings. The smoke script uses only an ephemeral synthetic fixture and checks clean input, a synthetic finding, redaction, and invalid configuration behavior.

`.security/secret-exceptions.yaml` currently contains no exceptions. If a synthetic test fixture must be excepted, add one exact file under `internal/devsecops/testdata/gitleaks/` and provide a known rule identifier, short reason, maintainer owner, linked Beads issue, and future expiry. Policy tests reject wildcard or non-fixture paths, missing metadata, duplicate IDs, and expired dates. Do not create exceptions for live credentials, customer material, or private reports. Never print or commit a suspected real credential; stop and contact the repository owner privately.

No exception file is currently consumed by Gitleaks as a suppression list. It is a reviewed registry for tracking narrowly scoped synthetic fixture exceptions, and the scan intentionally stays fail-closed.

## Local reproduction

Run from the repository root with the pinned tools installed:

```sh
go test -race ./...
go vet ./...
actionlint
govulncheck -test ./...
allowed_licenses="$(grep -Ev '^[[:space:]]*(#|$)' .security/allowed-go-licenses.txt | paste -sd, -)"
go-licenses check --include_tests --allowed_licenses="$allowed_licenses" ./...
bash .github/scripts/test-gitleaks.sh
git fetch --unshallow 2>/dev/null || true
gitleaks git --log-opts=--all --redact=100 --ignore-gitleaks-allow --no-banner --timeout=300 .
```

For a fresh clone, fetch all refs/history before the final Gitleaks command. The workflow does that with checkout `fetch-depth: 0`. Gitleaks findings are security-sensitive; retain redaction and do not paste unredacted tool output into an issue or chat.

## CodeQL distinction

The repository's CodeQL default setup is configured in GitHub repository settings; its runs are identified in GitHub under `dynamic/github-code-scanning/codeql`. The checked-in `.github/workflows/codeql.yml` is manually disabled in Actions settings and is not the active CodeQL execution path. The security-scans workflow does not enable or change CodeQL settings. Confirm runs and alerts in GitHub rather than treating the disabled YAML file as evidence of execution.
