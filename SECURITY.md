# Security policy

## Reporting a vulnerability

Do not include private reports, customer documents, or sensitive data in a public issue. Use GitHub's private vulnerability reporting for this repository when it is available. If it is unavailable, contact the repository owner through an established private channel before sharing details publicly. No response-time commitment is currently defined.

Provide a concise description, affected commit or component, reproduction steps using synthetic data where possible, and potential impact. Please allow the maintainer to investigate and coordinate a fix before public disclosure.

## Scope and current status

This repository currently contains planning documentation, a Go standard-library traceability checker, a profile contract validator, and a test-only pdfcpu full-write characterization harness. It does not yet contain the document-processing product. Repository workflow scanning is not a security assessment of a future runtime system.

Dependabot configuration requests weekly Go module and GitHub Actions version-update pull requests. The Go module declares pdfcpu v0.16.2 for the bounded `internal/pdfpreserve` test harness, with indirect module dependencies. This is a test characterization candidate, not an approved product/runtime PDF dependency; the full packaged dependency and security review remains open. See the [remediation component inventory](docs/research/remediation-component-inventory.md). A Dependabot configuration file does not prove that repository-level alerts or security updates are enabled. Dependabot alerts do not cover SHA-pinned Actions; version updates can maintain the configured pins. CodeQL workflows request repository-content read access and job-scoped `security-events: write` to publish Go and GitHub Actions analysis results. Workflow runs and findings must be verified in GitHub after publication.
