# Architecture

## Current status

**Proposed; not approved.** No production implementation exists. The Go-centered remediation design is a candidate direction only, and the generation lane remains less specified.

## Candidate direction

- A Go control plane may own the API, job orchestration, audit evidence, and semantic repair model.
- Existing-PDF remediation may use replaceable worker adapters so open-source tools can be evaluated or replaced independently.
- Generation should remain a distinct lane with an explicit PDF engine selection gate. Do not assume its engine is the remediation writer.
- Shared data profiles, security controls, validation evidence, and human review may be reusable; lane-specific transformations and acceptance remain separate.

## Open architecture questions

- Which data, accessibility, privacy, retention, and document-family profiles must be supported?
- Can an open-source PDF writer repair existing documents without unacceptable visible or structural regressions?
- Which generation PDF engine best preserves semantics from templates and source formats?
- Which tasks can be automated safely, and which require human semantic decisions?
- What are the deployment, audit, threat-model, and operational boundaries?

See the [roadmap](../roadmap/product-roadmap.md), [decision index](../decisions/README.md), and [research index](../research/README.md) for the evidence gates.
