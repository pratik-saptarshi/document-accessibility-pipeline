# Product roadmap: test-driven accessible document pipeline

**Status: Proposed plan.** This roadmap is a sequenced planning artifact. It contains no delivery dates, budget, measured performance, approved production architecture, or compliance determination.

## Scope and organizing principles

The roadmap covers two distinct lanes: open-source remediation of existing PDFs and accessible document generation. Both lanes include statements, invoices, correspondence, and reports. Generation also covers data/templates and DOCX/Markdown sources; its PDF engine is deliberately unresolved. IBM AFP print migration is a separate feasibility track.

The current remediation design direction is a Go control plane for API/job orchestration, audit evidence, and semantic repair, with replaceable worker adapters. This is proposed, not approved. Remediation dependencies must remain open-source-only. Enterprise-only tagging or PDF/UA export dependencies are excluded. Candidate tool capabilities, versions, licenses, and integration behavior must be checked during the relevant spikes.

Every implementation vertical slice follows **red → green → refactor**: first write a failing test, make the smallest change that passes, then refactor while tests remain green. Use synthetic/deidentified gold fixtures and characterization/preservation tests. Layer unit, adapter contract, integration, and end-to-end coverage with failure injection and security tests. Release gates include independent machine validation, human semantic review, and assistive-technology checks. Tagged output alone is not a PDF/UA or WCAG compliance result.

## Sequence and gates

### 0. Profiles and feasibility gates

Resolve data, accessibility, privacy/security, retention, and human-review profiles; establish synthetic fixtures; run the existing-PDF writer/fidelity spike; verify the open-source remediation boundary; and run the generation-engine selection spike. Record cross-lane decisions only from evidence. This is a prerequisite to committing to either production design.

**Deliverables:** profile and fixture definitions; reproducible writer/fidelity comparison; license/capability inventory; generation-engine evidence; decision records separating accepted facts from proposed choices.

**Acceptance gate:** test cases and fixture expectations exist before prototypes; preservation and semantic limits are reported; candidate versions/licenses are verified; open decisions and human review needs are explicit. A failed feasibility gate returns to alternatives and scope decisions rather than silently selecting a tool.

### 1. Existing-PDF remediation

Build the characterization and gold-fixture plan, specify replaceable Go adapter contracts, then proceed to a narrow baseline structure-repair slice and a later semantic/complex-content slice only after the writer/fidelity and open-source checks pass. Finish with independent validation and human release gates.

**Deliverables:** synthetic corpus plan; Go adapter contracts; test-first remediation slices; preservation evidence; machine validation plus semantic/assistive-technology review process.

**Acceptance gate:** preservation assertions pass for supported fixture cases; unsupported and ambiguous cases are explicit or routed to review; all remediation dependencies are verified open source; machine checks and human gates are recorded separately. No tags-only compliance claim.

### 2. Accessible generation

After profile and PDF-engine evidence, define the validated data/template model, test DOCX/Markdown source handling, then implement separate family slices: statements/invoices, followed by correspondence/reports. Keep generation engine selection explicit until the spike is reviewed.

**Deliverables:** family profiles and synthetic gold cases; tested data/template boundaries; DOCX/Markdown feasibility evidence; test-first generation slices and independent review evidence.

**Acceptance gate:** data integrity, visible output, semantics, and failure behavior are tested for each family; the selected engine and conversions have verified license/capability evidence; machine validation, human semantic review, and assistive-technology gates are satisfied before any compliance claim.

### 3. IBM AFP feasibility (separate track)

Assess AFP input/output goals, representative synthetic or safely sourced non-confidential fixtures, visual/pagination/font/barcode behavior, conversion options, licensing, and operational constraints. This track does not presume AFP conversion belongs in either PDF lane.

**Deliverables:** test-first feasibility protocol, evidence-backed options, open questions, and a go/no-go recommendation.

**Acceptance gate:** preservation expectations and limitations are documented; no conversion implementation begins without an explicit follow-up decision.

### Cross-lane quality and security

Define shared failure-injection, privacy/security, audit-integrity, and end-to-end gates that exercise both lanes independently. Include malformed PDFs, hostile metadata, resource exhaustion, worker failure, timeout, retries, and data-leakage checks. Keep family-specific outcomes and human review evidence distinguishable.

## Proposed architecture and candidate tools

- **Remediation control plane:** Go API, job orchestration, audit trail, and semantic repair model; worker adapters are replaceable. Proposed pending profile, writer/fidelity, and architecture gates.
- **Remediation candidates:** OpenDataLoader Apache-2.0 OSS auto-tagging, PDFBox Apache-2.0 custom writer, veraPDF machine-only validation, qpdf structural operations, and selective OCRmyPDF/Tesseract. These candidates come from prior design notes and remain unverified here. Verify current versions, license texts, capabilities, and adapter feasibility. Enterprise-only PDF/UA export/tagging is excluded from remediation.
- **Generation engine:** unresolved. Compare explicit candidates against synthetic family fixtures, accessible structure support, fidelity, source conversion, licensing, and operations. Do not assume the remediation writer is the generation engine.
- **IBM AFP:** distinct feasibility question; no accepted converter or target workflow.

## Test strategy and definition of done

Each story begins with failing tests and explicit fixtures. Red/green/refactor applies at each vertical slice, not just at project level.

1. Characterize and preserve existing content before changing PDFs: visible text, page geometry/rendering, page count, links, forms, annotations, fonts, metadata, and structure where applicable.
2. Add unit tests for profile validation, template/data rules, repair decisions, and audit evidence.
3. Add adapter contract tests for deterministic outputs, unsupported capabilities, timeouts, cancellation, retries, failure reporting, and license policy.
4. Add integration tests against verified open-source tools and end-to-end tests across all four families in both lanes.
5. Inject malformed/hostile inputs, resource limits, worker crashes/timeouts, and retry/idempotence scenarios; verify privacy and audit integrity.
6. Independently validate output with machine tools, then require human semantic and assistive-technology review. Record actual checks and limitations. A machine pass or tags alone is insufficient for a compliance claim.

**Definition of done for any future implementation issue:** acceptance tests were written before implementation; tests pass and are recorded; preservation and failure behavior are evidenced; applicable unit/contract/integration/end-to-end layers are covered; privacy-safe synthetic fixtures are used; licenses and versions are verified; machine and human validation are distinguished; unresolved decisions remain labeled; documentation and audit evidence are updated.

## Risks and decision gates

- **PDF writer fidelity:** structural edits can damage rendering or embedded document features. Gate the first remediation implementation on round-trip characterization and explicit supported-feature boundaries.
- **Automated semantics:** tagging tools can misinterpret reading order, tables, figures, or meaning. Route ambiguity to human review and retain audit evidence.
- **License/capability drift:** tool versions and licensing can change. Verify current source and exact versions during spikes; reject enterprise-only remediation dependencies.
- **Generation engine ambiguity:** one engine may not satisfy templates, DOCX/Markdown conversion, accessible structure, and rendering. Keep selection open until comparative fixtures pass.
- **False compliance confidence:** automated checks cannot establish semantic correctness alone. Require independent machine checks plus human semantic and assistive-technology gates.
- **Sensitive data exposure:** fixture or job data may contain personal information. Keep examples synthetic/deidentified and decide retention/security profiles before implementation.
- **AFP scope uncertainty:** conversion targets and fidelity expectations are unknown. Keep AFP separate and stop at feasibility until goals are agreed.

## Traceability and Beads issue graph

The [traceability matrix](../traceability/matrix.md) is the review entry point for stable capabilities, requirements, proposed user stories, function contracts, actionable Beads tasks, and case specifications. The [catalog JSON](../traceability/catalog.json) is the source for generated Markdown views; the [checker](../../tools/check_traceability.py) verifies catalog links against the tracked Beads export.

Beads stores the authoritative work state and prerequisite edges. Use `bd list`, `bd show <id>`, and `bd ready` for current status. Refresh `.beads/issues.jsonl` after issue mutations using the [contribution guide](../../CONTRIBUTING.md). The export is an issue interchange/recovery file, not a backup of embedded Dolt history.

The roadmap and issue descriptions are planning artifacts. **Refinement-ready** means a scoped planning/specification task can be detailed; it does not authorize implementation or mean a product case passed. **Decision-blocked** means one or more profile, feasibility, fidelity, tool, privacy, or review-authority decisions remain open. See the matrix for Definitions of Ready and Done and the catalog for open decisions.

The separate repository-security task is `dap-jo9.6`. Its local baseline is described in [SECURITY.md](../../SECURITY.md), `.github/dependabot.yml`, and the repository workflows. GitHub server-side security toggles and branch rules are outside this repository configuration and require separate repository-owner settings access.
