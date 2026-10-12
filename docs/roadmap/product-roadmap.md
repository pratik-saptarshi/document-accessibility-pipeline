# Product roadmap: test-driven accessible document pipeline

**Status: Proposed plan.** This roadmap is a sequenced planning artifact. It contains no delivery dates, budget, measured performance, approved production architecture, or compliance determination.

## Scope and organizing principles

The roadmap covers two distinct lanes: open-source remediation of existing PDFs and accessible document generation. Both lanes include statements, invoices, correspondence, and reports. Generation also covers data/templates and DOCX/Markdown sources; its PDF engine is deliberately unresolved. IBM AFP print migration is a separate feasibility track.

The current remediation design direction is a Go control plane for API/job orchestration, audit evidence, and semantic repair, with replaceable worker adapters. This is proposed, not approved. Remediation dependencies must remain open-source-only. Enterprise-only tagging or PDF/UA export dependencies are excluded. Candidate tool capabilities, versions, licenses, and integration behavior must be checked during the relevant spikes.

Every implementation vertical slice follows **red → green → refactor**: first write a failing test, make the smallest change that passes, then refactor while tests remain green. Use synthetic/deidentified gold fixtures and characterization/preservation tests. Layer unit, adapter contract, integration, and end-to-end coverage with failure injection and security tests. Release gates include explicit PDF/UA-2 machine validation, human semantic review, and assistive-technology checks. Tagged output alone is not a PDF/UA-2 or WCAG compliance result.

## Sequence and gates

## Refined implementation phases

The phase map below refines the existing Beads graph. IDs are the source of task state and dependencies; this roadmap records each milestone's intent and end gate. **Engineering completion does not mean a document is eligible for release or may be described as conformant.** Every implementation slice starts with a failing test, follows red/green/refactor, and records the test, fixture/tool versions and hashes, commit, and limitations. Do not mark an end gate passed while its human-review, evidence, or external-tool prerequisites are missing.

### Phase 0 — decisions and validation foundations

**Mapped work:** `US-REM-002`, `US-GEN-002`; `FN-REM-VALIDATE-OUTPUT`, `FN-GEN-VALIDATE-OUTPUT`; `dap-jo9.1.1`, `.1.3`, `.1.4`, `.1.5`, `.2.5.3` (adapter implementation successor to the closed `.2.5.1` contract specification), `.2.5.4` (real-tool authentication and oracle task; decision-blocked), and `.2.5.2` for the separately specified human/AT gate. Decision 0001 records the accepted PDF/UA-2 target, Go-first implementation constraint, permissive product-dependency rule, and CI-only veraPDF boundary. Other semantic, privacy, retention, reviewer-authority, arithmetic, and engine decisions remain open.

**Acceptance evidence:** profile schema v1.1 and tests demonstrate that the PDF/UA-2 target can be accepted separately while unrelated policies remain open; Linux fake-process adapter tests force `--flavour ua2`, bound preparation, process time, and pipe draining, terminate the isolated process group, reject symlink/special-file input, and redact source paths. The parser requires the report root and all used batch-summary attributes, rejects non-whitespace outside the single root, and checks profile/batch-summary/exit consistency. Evidence must label shim tests as adapter-only. Real veraPDF acquisition, signer identity/signature, immutable digest, real integration, and UA-2 PDF oracles are tracked in blocked task `.2.5.4`.

**End-of-phase validation:** run `go test -race ./...`, `go vet ./...`, profile examples/schema validation, traceability check, Beads lint, Markdown link check, Go license/security checks, and independent phase review. Keep the real-tool subgate blocked until artifact authentication succeeds; no semantic/AT release gate is satisfied by this phase.

### Phase 1 — input profiles, synthetic corpus, and engine feasibility

**Mapped work:** `REQ-REM-001`–`REQ-REM-004`, `REQ-GEN-001`–`REQ-GEN-004`; `US-REM-001`, `US-GEN-001`; `FN-REM-INTAKE`, `FN-REM-CHARACTERIZE`, `FN-REM-COMPARE`, `FN-GEN-VALIDATE-DATA`; `dap-jo9.1.1`–`.1.5`, `dap-jo9.2.1`, `dap-jo9.3.1.1` and `.3.2`. Keep generation and remediation profiles and fixtures distinct. Expand remediation characterization beyond the current small PDF 1.7 corpus to multiple pages, inherited resources, sparse/nonmonotonic MCIDs, ParentTree/MCR/OBJR links, fonts/Unicode, form appearances, and annotations. Explicitly classify signed, encrypted, embedded, scanned, malformed, and unsupported inputs. Generation needs separate PDF/UA-2 controls and one minimal semantic/content fixture per family.

**Acceptance evidence:** synthetic corpus manifests, provenance, baseline hashes, and explicit positive/negative/unsupported expectations; original-PDF SHA-256; fixed-version independent text and render comparison; source/engine/license/capability evidence; a no-go result is valid. Tiny PDF 1.7 preservation tests cannot establish writer selection or PDF/UA-2 conformance.

**End-of-phase validation:** run fixture manifest/schema checks, per-family contract tests, deterministic characterization twice, license/dependency review, and independent feasibility review. Any unexplained fidelity difference or missing baseline keeps dependent implementation issues blocked.

### Phase 2 — narrow existing-PDF remediation candidate pipeline

**Mapped work:** `US-REM-001`, `REQ-REM-001`–`REQ-REM-004`; `FN-REM-INTAKE`, `FN-REM-CHARACTERIZE`, `FN-REM-REPAIR-BASELINE`, `FN-REM-COMPARE`, `FN-REM-VALIDATE-OUTPUT`; `dap-jo9.2.2`, `.2.3`, `.2.3.2`, with Phase 0 and Phase 1 prerequisites.

**Acceptance evidence:** one explicitly bounded supported subset; intake, characterization, repair, comparison, explicit UA-2 machine validation, and review-needed state; original immutable; candidate output atomic; evidence tied to candidate hash. Each document family has positive and unsupported scenarios. Missing fidelity evidence, unexplained changes, crash, or partial artifact must not produce success.

**End-of-phase validation:** run race tests, unit/adapter/integration tests, fixed-version independent text/render/structure comparisons, malformed/unsupported/failure injection, dependency/license scans, and independent phase review. Keep semantic and assistive-technology release evidence separate.

### Phase 3 — statement then invoice generation

**Mapped work:** `US-GEN-001`, `REQ-GEN-001`–`REQ-GEN-006`; `FN-GEN-VALIDATE-DATA`, `FN-GEN-STATEMENT`, `FN-GEN-INVOICE`, `FN-GEN-VALIDATE-OUTPUT`; `dap-jo9.3.1`, `.3.1.1`, `.3.1.2`, `.3.2`, `.3.3`, `.3.3.1`–`.3.3.3`. Honor DOCX/Markdown and generation-engine feasibility dependencies; refine them explicitly if scope changes.

**Acceptance evidence:** validate data and template bindings before rendering; no invented tax, currency, or rounding policy; test missing/unknown/unsafe fields, arithmetic overflow, pagination, and table relationships. Keep one family slice at a time, statements before invoices. Each generated PDF has explicit UA-2 machine evidence and independent semantic/AT evidence; unavailable manual review remains blocked.

**End-of-phase validation:** run package and race tests, family golden tests, invalid-data and arithmetic boundaries, PDF text/render/structure comparisons, validator report checks, dependency/license/security scans, and independent review. Machine pass alone cannot close the human release gate.

### Phase 4 — semantic repair and complex families

**Mapped work:** `US-REM-002`, `US-GEN-002`, `REQ-REM-005`–`REQ-REM-007`, `REQ-GEN-003`–`REQ-GEN-006`; `FN-REM-SEMANTIC-REVIEW`, `FN-GEN-CORRESPONDENCE`, `FN-GEN-REPORT`, `FN-GEN-VALIDATE-OUTPUT`; `dap-jo9.2.4`, `.2.4.1`, `.2.4.2`, `.2.5`, `.2.5.2`, `.3.4`, `.3.4.1`–`.3.4.3`.

**Acceptance evidence:** bounded semantic operations and reviewer handoff; correspondence and reports; source-path feasibility; explicit findings for ambiguous reading order, tables, figures, labels, and language. Review records name authority, exact candidate hash, protocol, and unresolved findings. A changed output invalidates prior approval. Human review is never simulated as passed.

**End-of-phase validation:** run semantic contract tests for all four families, negative/ambiguous cases, machine validation, real human semantic and assistive-technology sessions, accessibility review of the evidence record, and independent phase review. Missing reviewers or protocol leave the release status blocked.

### Phase 5 — integrated release candidate and safety hardening

**Mapped work:** `US-X-001`, `REQ-X-001`–`REQ-X-003`; `FN-X-JOB-AUDIT`, `FN-X-SAFE-FAILURE`; `dap-jo9.5`, `.5.1`, `.5.2`, plus `.2.5` and all generation/remediation family tasks. Split the early machine adapter from this later integrated release validation. Split minimal safe-failure/audit primitives from late cross-lane hardening; avoid a generic workflow framework.

**Acceptance evidence:** eight lane/family end-to-end cases and negative controls; hostile inputs; byte/page/time/output/resource limits; isolation where required; cancellation, retries, partial artifact cleanup; privacy-safe logs and temp cleanup; audit failure behavior; access/retention decisions; repeatable benchmarks; full dependency/security review. Automation, rendering, semantic and AT outcomes remain separate. No certification claim.

**End-of-phase validation:** run all tests with race detector, end-to-end suite across eight family/lane combinations, hostile/failure injection, repeatability and benchmark protocol, license/vulnerability/action/secret scans, privacy/security review, human semantic and AT review, and independent release-candidate review. Any missing evidence blocks release eligibility even if engineering checks pass.

### Parallel track — IBM AFP feasibility

**Mapped work:** `REQ-AFP-001`; `dap-jo9.4`, `.4.1`, `.4.1.1`, `.4.1.2`. Prior official/GitHub research informs questions but is not fixture or fidelity evidence.

**Acceptance evidence:** missing-input checklist for resource-complete representative samples, fonts, overlays, barcodes, input/output target, fidelity criteria, licensing, and go/no-go authority; bounded evidence and explicit unknowns. No support claim or unapproved external converter implementation.

**End-of-phase validation:** review the fixture/resource checklist, verify cited official sources and license evidence, run only approved bounded feasibility checks, and record a go/no-go recommendation. Keep this track independent of the two PDF lanes.

### Unrefined horizon

SaaS, self-hosted, and desktop deployments have appeared in prior research but have no approved scope, deployment target, or acceptance requirements here. Keep them unrefined; do not create deployment gates or implementations from those mentions.

### 0. Profiles and feasibility gates

Resolve remaining data, semantic, privacy/security, retention, and human-review profiles; establish synthetic fixtures; run the existing-PDF writer/fidelity spike; verify the open-source remediation boundary; and run the generation-engine selection spike. The PDF/UA-2 target is already selected; the real veraPDF integration remains blocked pending artifact authentication. Record cross-lane decisions only from evidence. This is a prerequisite to committing to either production design.

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

The [traceability matrix](../traceability/matrix.md) is the review entry point for stable capabilities, requirements, proposed user stories, function contracts, actionable Beads tasks, and case specifications. The [catalog JSON](../traceability/catalog.json) is the source for generated Markdown views; the [Go checker](../../cmd/traceability-check) verifies catalog links against the tracked Beads export.

Beads stores the authoritative work state and prerequisite edges. Use `bd list`, `bd show <id>`, and `bd ready` for current status. Refresh `.beads/issues.jsonl` after issue mutations using the [contribution guide](../../CONTRIBUTING.md). The export is an issue interchange/recovery file, not a backup of embedded Dolt history.

The roadmap and issue descriptions are planning artifacts. **Refinement-ready** means a scoped planning/specification task can be detailed; it does not authorize implementation or mean a product case passed. **Decision-blocked** means one or more profile, feasibility, fidelity, tool, privacy, or review-authority decisions remain open. See the matrix for Definitions of Ready and Done and the catalog for open decisions.

The separate repository-security task is `dap-jo9.6`. Its local baseline is described in [SECURITY.md](../../SECURITY.md), `.github/dependabot.yml`, and the repository workflows. GitHub server-side security toggles and branch rules are outside this repository configuration and require separate repository-owner settings access.
