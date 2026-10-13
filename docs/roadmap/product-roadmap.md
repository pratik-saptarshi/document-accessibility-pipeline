# Product roadmap: test-driven accessible document pipeline

**Status: Proposed plan.** This roadmap is a sequenced planning artifact. It contains no delivery dates, budget, measured performance, approved production architecture, or compliance determination.

## Scope and organizing principles

The roadmap covers two distinct lanes: open-source remediation of existing PDFs and accessible document generation. Both lanes include statements, invoices, correspondence, and reports. Generation also covers data/templates and DOCX/Markdown sources; its PDF engine is deliberately unresolved. IBM AFP print migration is a separate feasibility track.

The current remediation design direction is a Go control plane for API/job orchestration, audit evidence, and semantic repair, with replaceable worker adapters. This is proposed, not approved. Product dependencies require permissive open-source licenses. Enterprise-only tagging or PDF/UA export dependencies are excluded. The accepted veraPDF exception is isolated to CI-only validation; no other non-Go runtime exception is approved. Candidate tool capabilities, versions, licenses, and integration behavior must be checked during the relevant spikes.

Every implementation vertical slice follows **red → green → refactor**: first write a failing test, make the smallest change that passes, then refactor while tests remain green. Use synthetic/deidentified gold fixtures and characterization/preservation tests. Layer unit, adapter contract, integration, and end-to-end coverage with failure injection and security tests. Release gates include explicit PDF/UA-2 machine validation, human semantic review, and assistive-technology checks. Tagged output alone is not a PDF/UA-2 or WCAG compliance result.

## Sequence and gates

### Phase 0 — shared validation and review foundations

**Entry:** the accepted constraints in [Decision 0001](../decisions/0001-pdf-ua-2-validation-target.md): Go-first implementation; permissive open-source product dependencies; explicit PDF/UA-2 (`ua2`) output validation; and veraPDF only in the isolated CI adapter boundary. Remediation and generation profiles remain separate.

**Work:** resolve lane-specific profile and evidence contracts; prepare the exact candidate-bound machine, human semantic, and assistive-technology evidence interfaces; and implement only focused early bounds, cancellation, redaction, partial-output, and audit-failure primitives needed by the first vertical slices. After its existing authority and candidate prerequisites, `dap-jo9.2.5.9` may define and exercise labeled synthetic evidence-record contract tests here or in Phase 1. Those tests establish engineering behavior only. Its closure requires an actual eligible candidate with hash-bound preservation/`ua2` evidence from a successful bounded experiment or candidate slice; a no-go does not satisfy that requirement. The reviewer authority/protocol remains open until an authorized owner decides it. Engineering can return review-needed and cannot simulate a reviewer result.

**Exit evidence:** schema and profile tests keep unrelated policies independently open; the isolated adapter reports explicit `ua2`; evidence binds to exact candidate bytes; human semantic and assistive-technology outcomes are separate; and applicable safety failures cannot be reported as success. A machine pass is not semantic or release approval.

Contract-test readiness is distinct from issue completion. A qualified experimental candidate may exercise the bounded review gate after its existing prerequisites; Phase 3 pipeline integration is not required for those contract tests or that candidate-specific demonstration.

### Phase 1 — bounded existing-PDF financial-table feasibility

**Priority:** [Decision 0002](../decisions/0002-initial-remediation-family-priority.md) gives statements and invoices with tables equal first priority. Neither family precedes the other. This is remediation prioritization only; it does not choose a writer, generation family, production workflow, or arithmetic policy.

**Phase entry:** the versioned, equal-priority synthetic configuration in `dap-jo9.1.1.1` is reviewed. Actual fixture artifacts and candidate clearance are work within this phase, not phase-entry requirements. Fixed synthetic language, font, values, and topology are experiment parameters, not product defaults or production restrictions. Synthetic engineering may proceed while production privacy/access/retention choices remain open; resolve those choices before sensitive or production documents are used.

**Order:** child A, `dap-jo9.2.1.2`, authors one small synthetic statement transaction-table fixture and one small synthetic invoice line-item fixture with immutable source hashes, exact displayed strings, fixture-authored target associations and order, supported boundaries, exclusions, and independent text/render/structure observations. A also verifies the normative relationship representation and validator limits and defines amount, association, malformed, unsupported, and ambiguous controls before repair code. For the repair-spike step, require A's fixtures and baselines plus `dap-jo9.1.3` clearance for the exact candidate and its reachable dependencies/assets. Then `dap-jo9.1.2` attempts an actual declared relationship change on each fixture, preserving all non-target declared relations/order and displayed strings, comparing independently, writing separate atomic candidates, and binding explicit `ua2` CI evidence to each candidate hash. A no-op PDF 1.7 round trip is insufficient. Partially tagged PDF 2.0 can isolate editing only; it proves neither PDF 1.7 conversion nor arbitrary untagged or complex-table support. Child B, `dap-jo9.2.1.3`, expands the corpus only after the bounded disposition and only for evidenced capabilities and accepted profiles; if the result is no-go, B records no unsupported expansion.

**Finite disposition:** `proceed` requires actual repair and complete candidate evidence for both fixtures. `reject` or `inconclusive` requires bounded attempted-step evidence, the exact failure or missing prerequisite, limits, and the next decision; do not invent a candidate, output, or validator report. Independent review and acceptance of any finite disposition completes `dap-jo9.1.2`. A no-go selects no production writer and closes no unrelated profile, product, or release gate. These outcomes apply only to the declared operation and do not imply all-four-family repair support. After the existing authority and candidate prerequisites, Phase 0/1 may exercise `dap-jo9.2.5.9` on a qualified experimental candidate; a closed no-go or inconclusive result cannot satisfy that issue's candidate-evidence closure requirement. Phase 3 pipeline integration is not a prerequisite for those contract tests or that candidate-specific demonstration.

### Phase 2 — narrow remediation candidate pipeline

**Entry:** a `proceed` result for the declared operation, accepted input/profile and candidate boundaries, verified permissive candidate and complete relevant license evidence, required adapter contract, and applicable safety controls. A `reject` or `inconclusive` result can close the finite feasibility issue after independent acceptance, but returns implementation work to alternatives or profile decisions; it does not silently select a writer. Existing PDF 1.7 characterization remains preservation evidence only.

**Work:** implement one test-first bounded candidate path: reproducible intake/profile, immutable baseline, declared repair, separate atomic output, independent fidelity comparison, explicit candidate-bound `ua2`, and distinct reviewer evidence or review-needed/unsupported/classified failure. Expand only against demonstrated capability. Preserve all four remediation families in corpus planning; the first financial fixtures do not close correspondence/report obligations.

**Exit evidence:** passing executions identify fixture, input/candidate/evidence hashes and tool versions; negative/failure controls show missing evidence, partial writes, audit errors, and unexplained changes cannot become success; unsupported cases are explicit; and any human/AT result is genuine, authorized, and bound to the reviewed candidate. Engineering completion is not release eligibility.

### Phase 3 — semantic remediation and reviewer handoff

**Entry:** Phase 2's bounded candidate/evidence interfaces for user-facing pipeline integration, approved semantic operations, and an accepted reviewer authority and protocol. Keep `dap-jo9.2.5.2` as canonical owner of the existing human and assistive-technology case specifications. The candidate-bound implementation task consumes those specifications without changing their ownership. Phase 3 integration is not a prerequisite for `dap-jo9.2.5.9`'s earlier contract tests or a qualified experimental-candidate demonstration; its closure still requires successful candidate evidence, not a synthetic test pass.

**Work:** implement only approved semantic changes; route ambiguity with context and rationale instead of guessing; bind distinct human semantic and AT outcomes to candidate hash, profile, protocol/tool versions, and reviewer authority; invalidate both after candidate bytes change. Missing or rejected required evidence blocks release.

**Exit evidence:** accepted, missing, rejected, and invalidated evidence paths are executable; simulated unit outcomes are labeled as simulations; eligible-candidate evidence is actual; and no synthetic test result is presented as human/AT sign-off. Correspondence and report remediation remain family-specific work, not implied by the financial-table slice.

### Separate generation lane — validated data and family outputs

Generation retains its independent profile, source conversion, and engine dependencies. Keep DOCX and Markdown source feasibility (`dap-jo9.3.2`) and generation engine selection (`dap-jo9.1.5`) as existing gates; this remediation prioritization does not retarget or satisfy them. After those gates, validate fields and template bindings before rendering, then deliver bounded family slices: statements, invoices, correspondence, and reports. Do not invent arithmetic, tax, currency, date, or rounding rules. Validate each candidate against accepted vectors and explicit `ua2` evidence; human semantic and AT review remain separate release gates.

**Exit evidence:** field-level invalid-data controls run before rendering; outputs preserve accepted source content and profile/template identity; each family has positive, negative, unsupported, pagination, and candidate-specific validation outcomes; source and engine licenses/capabilities are reviewed; and missing human/AT evidence blocks release. Generation's engine is not presumed to be the remediation writer.

### Separate parallel track — IBM AFP feasibility

Use the bounded resource/input checklist, explicit target and fidelity criteria, verified licensing and operational constraints, limitations, and a go/no-go/unknown recommendation. Do not implement a converter or make AFP a prerequisite for either PDF lane. Existing profile and owner decisions remain gates; no go/no-go evidence is presumed by this roadmap.

### Cross-lane quality and release gates

Exercise both lanes independently with malformed and hostile inputs, resource bounds, timeouts, cancellation, retries, partial artifacts, privacy-safe diagnostics, and audit failures. Retain the later eight lane/family end-to-end, security, and release gates; the first financial-table slice closes none of them. Reuse focused early controls; do not create a general workflow framework before a consumer requires it. Keep lane/family candidate machine evidence, human semantic review, and assistive-technology evidence distinguishable. Production access/retention decisions apply before sensitive or production processing; security review, candidate-specific validation, and required authorized reviewers remain release gates. No tag-only or machine-only conformance claim is allowed.

### Unrefined horizon

SaaS, self-hosted, and desktop deployment mentions have no approved scope, target, or acceptance criteria. Keep them unrefined and create no deployment implementation gates from those mentions.

## Proposed architecture and candidate tools

- **Remediation control plane:** Go API, job orchestration, audit trail, and semantic repair model; worker adapters are replaceable. Proposed pending profile, writer/fidelity, and architecture gates.
- **Remediation candidates:** OpenDataLoader Apache-2.0 OSS auto-tagging, PDFBox Apache-2.0 custom writer, veraPDF machine-only validation, qpdf structural operations, and selective OCRmyPDF/Tesseract appear in historical research. That list is not a universal trial plan or an approved dependency set. `dap-jo9.1.3` clears only the exact candidate proposed for the finite operation, including its reachable dependencies and packaged assets; historical alternatives remain untested. Enterprise-only PDF/UA export/tagging is excluded from remediation.
- Historical Java/Python tool mentions are not selected. No product-runtime exception is approved; the only accepted non-Go execution boundary is the isolated CI-only veraPDF validator recorded in Decision 0001.
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
