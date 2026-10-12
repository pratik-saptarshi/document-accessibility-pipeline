# Security policy

## Reporting a vulnerability

Do not include private reports, customer documents, or sensitive data in a public issue. Use GitHub's private vulnerability reporting for this repository when it is available. If it is unavailable, contact the repository owner through an established private channel before sharing details publicly. No response-time commitment is currently defined.

Provide a concise description, affected commit or component, reproduction steps using synthetic data where possible, and potential impact. Please allow the maintainer to investigate and coordinate a fix before public disclosure.

## Scope and current status

This repository currently contains planning documentation, a Go standard-library traceability checker, a profile contract validator, and a test-only pdfcpu full-write characterization harness. It does not yet contain the document-processing product. Repository workflow scanning is not a security assessment of a future runtime system.

Dependabot configuration requests weekly Go module and GitHub Actions version-update pull requests. The Go module declares pdfcpu v0.16.2 for the bounded `internal/pdfpreserve` test harness, with indirect module dependencies. This is a test characterization candidate, not an approved product/runtime PDF dependency; the full packaged dependency and security review remains open. See the [remediation component inventory](docs/research/remediation-component-inventory.md). A Dependabot configuration file does not prove that repository-level alerts or security updates are enabled. Dependabot alerts do not cover SHA-pinned Actions; version updates can maintain the configured pins.

## Repository checks

The [security scanning workflow](.github/workflows/security-scans.yml) runs on pull requests to `main`, pushes to `main`, and weekly. It uses read-only repository permissions and immutable action commit pins. The Go job runs tests and `go vet`, lints workflow YAML, checks vulnerabilities with `govulncheck -test ./...` in text mode (so findings fail the job), and checks the test-inclusive Go dependency graph against the explicit [SPDX allowlist](.security/allowed-go-licenses.txt). Unknown and unlisted licenses fail. The allowlist is a current graph review, not a product-runtime dependency approval; pdfcpu remains test-only and its full packaged dependency/security review remains open.

The secret-scanning job fetches all Git history and runs Gitleaks with full redaction and native failure status. It pins the default rule set in `.security/gitleaks.toml`, rejects working-tree `.gitleaksignore` and `.gitleaks.toml` files before scanning, sets the ignore path to `/dev/null`, and disables inline allow comments. It does not write a full-scan report artifact. A synthetic smoke check exercises clean, finding, redaction, invalid-configuration, and suppression-bypass behavior before scanning. Do not add a real secret to a fixture or exception. If a finding may be a real credential, stop processing it and notify the repository owner privately without copying the value into logs, an issue, or a public report.

There are currently no approved secret exceptions. Future exceptions must cover exactly one synthetic Gitleaks test fixture and identify a known rule, reason, owner, Beads issue, and future expiry. Keep the registry in `.security/secret-exceptions.yaml`; broad paths and inline Gitleaks suppressions are rejected by policy. Review and remove an exception when its issue is resolved or its expiry is reached.

Go policy tests enforce workflow triggers and least privilege, SHA-pinned actions with version comments, fail-closed vulnerability/license scanner options, explicit Gitleaks config and suppression paths, the suppression-file guard, full-history checkout, redaction, and exception metadata. The license smoke test constructs a temporary local module reachable only from a test and proves the license check rejects its forbidden license. Reproduce locally with `go test -race ./...`, `go vet ./...`, `actionlint`, `govulncheck -test ./...`, `go-licenses check --include_tests --allowed_licenses=... ./...`, `.github/scripts/test-go-licenses.sh`, and the Gitleaks smoke script followed by the full-history scan. See [security scanning operations](docs/security/README.md) for tool pins and details.

GitHub's default CodeQL setup is separately configured in repository settings and its runs appear under `dynamic/github-code-scanning/codeql`. The checked-in `.github/workflows/codeql.yml` is manually disabled in GitHub Actions settings and does not execute. This change does not alter GitHub settings, branch rules, or repository permissions. Verify default-setup runs and findings in GitHub; do not infer them from the disabled workflow file.
