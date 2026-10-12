# Synthetic fixture planning

Fixture manifests describe planned characterization work. They do not contain PDF artifacts or assert observed preservation results. The current remediation manifest is versioned at [`manifest.v1.json`](remediation/manifest.v1.json).

The v1 manifest maps one planned synthetic case to each family and references the corresponding synthetic remediation profile example. All cases are marked `planned-not-created` and `not-run`. Listed dimensions are items to assess for applicability and evidence; they are not expected values or promises that every fixture contains each feature.

Run `go test ./internal/corpus` to parse and validate the tracked manifest and exercise rejection cases.

The Go validator rejects unknown fields, non-synthetic provenance, unplanned artifacts, incomplete family or requirement coverage, duplicate IDs, and unsafe or mismatched profile paths. It validates planning metadata only. It does not generate, open, compare, or validate PDFs. Future malformed, scanned, or mixed-content cases must also be synthetic and remain proposed until safely authored and characterized.

This slice supports [REQ-REM-001 through REQ-REM-003](../requirements/catalog.md) and Beads issues `dap-jo9.1.1`, `dap-jo9.1.2`, and `dap-jo9.2.1`. Their profile and writer feasibility gates remain open.
