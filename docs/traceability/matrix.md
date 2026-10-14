# Traceability matrix

The JSON catalog is the planning source of truth for stable IDs and explicit links. The checker validates these links against the tracked Beads JSONL export and generates this view. Task-case links describe contractual participation; all cases remain specification-only, not executed results.

```text
Capability → requirement → proposed persona/story → function contract → Beads task → acceptance case specification
```

## Stories and linked work

### US-REM-001 — dap-jo9.2.3

As a proposed document-remediation operator, I want a bounded baseline structure repair for supported existing PDFs, so that repair candidates can be assessed without losing characterized content.

- **Capability:** `CAP-REM-PRESERVE`
- **Requirements:** REQ-REM-001, REQ-REM-002, REQ-REM-003, REQ-REM-004
- **Persona:** proposed; role and authority unresolved

- **Functions:** FN-REM-INTAKE, FN-REM-CHARACTERIZE, FN-REM-REPAIR-BASELINE, FN-REM-COMPARE
- **Tasks:** `dap-jo9.2.3.1` (refinement-ready), `dap-jo9.2.3.2` (decision-blocked)
- **Cases:** `TC-REM-STATEMENT-001`, `TC-REM-INVOICE-001`, `TC-REM-CORRESPONDENCE-001`, `TC-REM-REPORT-001`, `TC-REM-MALFORMED-001`, `TC-REM-FIDELITY-001`

#### Task-case participation

- `dap-jo9.2.3.1` → `TC-REM-STATEMENT-001`: **specification owner**
- `dap-jo9.2.3.1` → `TC-REM-INVOICE-001`: **specification owner**
- `dap-jo9.2.3.1` → `TC-REM-CORRESPONDENCE-001`: **specification owner**
- `dap-jo9.2.3.1` → `TC-REM-REPORT-001`: **specification owner**
- `dap-jo9.2.3.1` → `TC-REM-MALFORMED-001`: **specification owner**
- `dap-jo9.2.3.1` → `TC-REM-FIDELITY-001`: **specification owner**
- `dap-jo9.2.3.2` → `TC-REM-STATEMENT-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — first increment: paired statement repair contract; not whole-task completion
- `dap-jo9.2.3.2` → `TC-REM-INVOICE-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — first increment: paired invoice repair contract; not whole-task completion
- `dap-jo9.2.3.2` → `TC-REM-CORRESPONDENCE-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — full-task obligation: separately reviewed later increment; outstanding after first pair
- `dap-jo9.2.3.2` → `TC-REM-REPORT-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — full-task obligation: separately reviewed later increment; outstanding after first pair
- `dap-jo9.2.3.2` → `TC-REM-MALFORMED-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — first increment includes pair-specific negative controls; remaining whole-task scope needs separate review
- `dap-jo9.2.3.2` → `TC-REM-FIDELITY-001`: **implements** (canonical owner `dap-jo9.2.3.1`) — first increment includes pair-specific preservation comparisons; remaining whole-task scope needs separate review

### US-REM-002 — dap-jo9.2.4

As a proposed accessibility reviewer, I want uncertain semantic repairs routed to review with evidence, so that ambiguous structure is not silently guessed.

- **Capability:** `CAP-REM-SEMANTIC`
- **Requirements:** REQ-REM-005, REQ-REM-006, REQ-REM-007
- **Persona:** proposed; role and authority unresolved

- **Functions:** FN-REM-SEMANTIC-REVIEW, FN-REM-VALIDATE-OUTPUT, FN-REM-MACHINE-VALIDATE, FN-REM-ASSISTIVE-TECH-REVIEW
- **Tasks:** `dap-jo9.2.4.1` (refinement-ready), `dap-jo9.2.4.2` (decision-blocked), `dap-jo9.2.5.1` (refinement-ready, group dap-jo9.2.5), `dap-jo9.2.5.2` (refinement-ready, group dap-jo9.2.5), `dap-jo9.2.5.3` (refinement-ready, group dap-jo9.2.5), `dap-jo9.2.5.4` (refinement-ready, group dap-jo9.2.5), `dap-jo9.2.5.9` (decision-blocked, group dap-jo9.2.5, standalone: Early acceptance follows direct evidence gates; the later .2.5 pipeline remains separately gated.)
- **Cases:** `TC-REM-AMBIGUOUS-001`, `TC-REM-REVIEW-ROUTE-001`, `TC-REM-MACHINE-001`, `TC-PDFUA2-REPORT-CONTRACT-001`, `TC-PDFUA2-REAL-ORACLE-001`, `TC-REM-HUMAN-001`, `TC-REM-ASSISTIVE-TECH-001`, `TC-GEN-VALIDATION-001`

#### Task-case participation

- `dap-jo9.2.4.1` → `TC-REM-AMBIGUOUS-001`: **specification owner**
- `dap-jo9.2.4.1` → `TC-REM-REVIEW-ROUTE-001`: **specification owner**
- `dap-jo9.2.4.1` → `TC-REM-HUMAN-001`: **consumes** (canonical owner `dap-jo9.2.5.2`) — human/AT review contract and eventual results; does not perform human or AT review
- `dap-jo9.2.4.1` → `TC-REM-ASSISTIVE-TECH-001`: **consumes** (canonical owner `dap-jo9.2.5.2`) — human/AT review contract and eventual results; does not perform human or AT review
- `dap-jo9.2.4.2` → `TC-REM-AMBIGUOUS-001`: **implements** (canonical owner `dap-jo9.2.4.1`) — ambiguity handling and review routing
- `dap-jo9.2.4.2` → `TC-REM-REVIEW-ROUTE-001`: **implements** (canonical owner `dap-jo9.2.4.1`) — ambiguity handling and review routing
- `dap-jo9.2.4.2` → `TC-REM-HUMAN-001`: **consumes** (canonical owner `dap-jo9.2.5.2`) — review contract and eventual results; does not perform human review
- `dap-jo9.2.4.2` → `TC-REM-ASSISTIVE-TECH-001`: **consumes** (canonical owner `dap-jo9.2.5.2`) — review contract and eventual results; does not perform AT review
- `dap-jo9.2.5.1` → `TC-REM-MACHINE-001`: **specification owner**
- `dap-jo9.2.5.1` → `TC-GEN-VALIDATION-001`: **specification owner**
- `dap-jo9.2.5.2` → `TC-REM-HUMAN-001`: **specification owner**
- `dap-jo9.2.5.2` → `TC-REM-ASSISTIVE-TECH-001`: **specification owner**
- `dap-jo9.2.5.2` → `TC-GEN-VALIDATION-001`: **contributes** (canonical owner `dap-jo9.2.5.1`) — human/AT separation and outcome requirements only; not generation validation or full composite acceptance
- `dap-jo9.2.5.3` → `TC-PDFUA2-REPORT-CONTRACT-001`: **specification owner**
- `dap-jo9.2.5.3` → `TC-REM-MACHINE-001`: **contributes** (canonical owner `dap-jo9.2.5.1`) — machine process/report subcontract only; no family candidate result or human/AT evidence
- `dap-jo9.2.5.3` → `TC-GEN-VALIDATION-001`: **contributes** (canonical owner `dap-jo9.2.5.1`) — machine process/report subcontract only; no family candidate result or human/AT evidence
- `dap-jo9.2.5.4` → `TC-PDFUA2-REAL-ORACLE-001`: **specification owner**
- `dap-jo9.2.5.4` → `TC-REM-MACHINE-001`: **contributes** (canonical owner `dap-jo9.2.5.1`) — authenticated tool/oracle contribution only; no family candidate result or human/AT evidence
- `dap-jo9.2.5.9` → `TC-REM-HUMAN-001`: **implements** (canonical owner `dap-jo9.2.5.2`) — candidate binding, evidence-state enforcement, and invalidation for the bounded review gate
- `dap-jo9.2.5.9` → `TC-REM-ASSISTIVE-TECH-001`: **implements** (canonical owner `dap-jo9.2.5.2`) — separate assistive-technology evidence-state enforcement for the bounded review gate

### US-GEN-001 — dap-jo9.3.3

As a proposed document-operations author, I want validated statement and invoice inputs rendered as semantically structured documents, so that repeatable outputs can be reviewed before release.

- **Capability:** `CAP-GEN-RENDER`
- **Requirements:** REQ-GEN-001, REQ-GEN-002, REQ-GEN-003, REQ-GEN-005, REQ-GEN-006
- **Persona:** proposed; role and authority unresolved

- **Functions:** FN-GEN-VALIDATE-DATA, FN-GEN-STATEMENT, FN-GEN-INVOICE
- **Tasks:** `dap-jo9.3.1.1` (refinement-ready, group dap-jo9.3.1), `dap-jo9.3.1.2` (decision-blocked, group dap-jo9.3.1), `dap-jo9.3.3.1` (refinement-ready), `dap-jo9.3.3.2` (decision-blocked), `dap-jo9.3.3.3` (decision-blocked)
- **Cases:** `TC-GEN-STATEMENT-001`, `TC-GEN-INVOICE-001`, `TC-GEN-INVALID-DATA-001`, `TC-GEN-VALIDATION-001`, `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001`

#### Task-case participation

- `dap-jo9.3.1.1` → `TC-GEN-INVALID-DATA-001`: **specification owner**
- `dap-jo9.3.1.1` → `TC-GEN-STATEMENT-001`: **contributes** (canonical owner `dap-jo9.3.3.1`) — input-contract/validation only; not rendering, machine validation, or human/AT completion
- `dap-jo9.3.1.1` → `TC-GEN-INVOICE-001`: **contributes** (canonical owner `dap-jo9.3.3.1`) — input-contract/validation only; not rendering, machine validation, or human/AT completion
- `dap-jo9.3.1.1` → `TC-GEN-CORRESPONDENCE-001`: **contributes** (canonical owner `dap-jo9.3.4.1`) — input-contract/validation only; not rendering, machine validation, or human/AT completion
- `dap-jo9.3.1.1` → `TC-GEN-REPORT-001`: **contributes** (canonical owner `dap-jo9.3.4.1`) — input-contract/validation only; not rendering, machine validation, or human/AT completion
- `dap-jo9.3.1.2` → `TC-GEN-INVALID-DATA-001`: **implements** (canonical owner `dap-jo9.3.1.1`) — pre-render input validation
- `dap-jo9.3.1.2` → `TC-GEN-STATEMENT-001`: **contributes** (canonical owner `dap-jo9.3.3.1`) — pre-render input validation only; not complete family acceptance
- `dap-jo9.3.1.2` → `TC-GEN-INVOICE-001`: **contributes** (canonical owner `dap-jo9.3.3.1`) — pre-render input validation only; not complete family acceptance
- `dap-jo9.3.1.2` → `TC-GEN-CORRESPONDENCE-001`: **contributes** (canonical owner `dap-jo9.3.4.1`) — pre-render input validation only; not complete family acceptance
- `dap-jo9.3.1.2` → `TC-GEN-REPORT-001`: **contributes** (canonical owner `dap-jo9.3.4.1`) — pre-render input validation only; not complete family acceptance
- `dap-jo9.3.3.1` → `TC-GEN-STATEMENT-001`: **specification owner**
- `dap-jo9.3.3.1` → `TC-GEN-INVOICE-001`: **specification owner**
- `dap-jo9.3.3.1` → `TC-GEN-INVALID-DATA-001`: **consumes** (canonical owner `dap-jo9.3.1.1`) — input-validation contract
- `dap-jo9.3.3.1` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — shared output-validation contract
- `dap-jo9.3.3.2` → `TC-GEN-STATEMENT-001`: **implements** (canonical owner `dap-jo9.3.3.1`) — statement output
- `dap-jo9.3.3.2` → `TC-GEN-INVALID-DATA-001`: **consumes** (canonical owner `dap-jo9.3.1.1`) — pre-render input-validation contract
- `dap-jo9.3.3.2` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — candidate-specific output-validation contract/result
- `dap-jo9.3.3.3` → `TC-GEN-INVOICE-001`: **implements** (canonical owner `dap-jo9.3.3.1`) — invoice output; no arithmetic policy implied
- `dap-jo9.3.3.3` → `TC-GEN-INVALID-DATA-001`: **consumes** (canonical owner `dap-jo9.3.1.1`) — pre-render input-validation contract
- `dap-jo9.3.3.3` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — candidate-specific output-validation contract/result

### US-GEN-002 — dap-jo9.3.4

As a proposed communications author, I want correspondence and reports rendered from validated inputs with headings and reading order represented, so that readers can navigate the generated content.

- **Capability:** `CAP-GEN-RENDER`
- **Requirements:** REQ-GEN-001, REQ-GEN-002, REQ-GEN-003, REQ-GEN-004, REQ-GEN-005, REQ-GEN-006
- **Persona:** proposed; role and authority unresolved

- **Functions:** FN-GEN-CORRESPONDENCE, FN-GEN-REPORT, FN-GEN-VALIDATE-OUTPUT
- **Tasks:** `dap-jo9.3.4.1` (refinement-ready), `dap-jo9.3.4.2` (decision-blocked), `dap-jo9.3.4.3` (decision-blocked)
- **Cases:** `TC-GEN-CORRESPONDENCE-001`, `TC-GEN-REPORT-001`, `TC-GEN-SOURCE-FEASIBILITY-001`, `TC-GEN-ENGINE-GATE-001`, `TC-GEN-VALIDATION-001`

#### Task-case participation

- `dap-jo9.3.4.1` → `TC-GEN-CORRESPONDENCE-001`: **specification owner**
- `dap-jo9.3.4.1` → `TC-GEN-REPORT-001`: **specification owner**
- `dap-jo9.3.4.1` → `TC-GEN-SOURCE-FEASIBILITY-001`: **specification owner**
- `dap-jo9.3.4.1` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — shared output-validation contract
- `dap-jo9.3.4.1` → `TC-GEN-ENGINE-GATE-001`: **specification owner**
- `dap-jo9.3.4.2` → `TC-GEN-CORRESPONDENCE-001`: **implements** (canonical owner `dap-jo9.3.4.1`) — correspondence output family
- `dap-jo9.3.4.2` → `TC-GEN-SOURCE-FEASIBILITY-001`: **consumes** (canonical owner `dap-jo9.3.4.1`) — source-feasibility obligation; no source dependency bypass
- `dap-jo9.3.4.2` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — candidate-specific output-validation contract/result
- `dap-jo9.3.4.3` → `TC-GEN-REPORT-001`: **implements** (canonical owner `dap-jo9.3.4.1`) — report output family
- `dap-jo9.3.4.3` → `TC-GEN-SOURCE-FEASIBILITY-001`: **consumes** (canonical owner `dap-jo9.3.4.1`) — source-feasibility obligation; no source dependency bypass
- `dap-jo9.3.4.3` → `TC-GEN-VALIDATION-001`: **consumes** (canonical owner `dap-jo9.2.5.1`) — candidate-specific output-validation contract/result

### US-X-001 — dap-jo9.5

As a proposed service operator, I want both lanes to classify failures and retain privacy-aware audit evidence, so that support can distinguish review, retry, and terminal outcomes.

- **Capability:** `CAP-CROSS-ORCHESTRATION`
- **Requirements:** REQ-X-001, REQ-X-002, REQ-X-003
- **Persona:** proposed; role and authority unresolved

- **Functions:** FN-X-JOB-AUDIT, FN-X-SAFE-FAILURE
- **Tasks:** `dap-jo9.5.1` (refinement-ready), `dap-jo9.5.2` (decision-blocked)
- **Cases:** `TC-CROSS-HOSTILE-001`, `TC-CROSS-TIMEOUT-001`, `TC-CROSS-PRIVACY-001`, `TC-CROSS-AUDIT-001`

#### Task-case participation

- `dap-jo9.5.1` → `TC-CROSS-HOSTILE-001`: **specification owner**
- `dap-jo9.5.1` → `TC-CROSS-TIMEOUT-001`: **specification owner**
- `dap-jo9.5.1` → `TC-CROSS-PRIVACY-001`: **specification owner**
- `dap-jo9.5.1` → `TC-CROSS-AUDIT-001`: **specification owner**
- `dap-jo9.5.2` → `TC-CROSS-HOSTILE-001`: **implements** (canonical owner `dap-jo9.5.1`) — shared hostile-input bounds and safe-failure behavior
- `dap-jo9.5.2` → `TC-CROSS-TIMEOUT-001`: **implements** (canonical owner `dap-jo9.5.1`) — shared timeout and cancellation behavior
- `dap-jo9.5.2` → `TC-CROSS-PRIVACY-001`: **implements** (canonical owner `dap-jo9.5.1`) — shared privacy-aware diagnostic behavior
- `dap-jo9.5.2` → `TC-CROSS-AUDIT-001`: **implements** (canonical owner `dap-jo9.5.1`) — shared bounded audit-evidence behavior

### US-AFP-001 — dap-jo9.4.1

As a proposed AFP print-operations owner, I want a bounded feasibility result for an agreed AFP input/output goal, so that a go/no-go decision can be made before migration is considered.

- **Capability:** `CAP-AFP-FEASIBILITY`
- **Requirements:** REQ-AFP-001
- **Persona:** proposed; role, target, and workload unresolved

- **Functions:** FN-AFP-ASSESS
- **Tasks:** `dap-jo9.4.1.1` (refinement-ready), `dap-jo9.4.1.2` (decision-blocked)
- **Cases:** `TC-AFP-FEASIBILITY-001`

#### Task-case participation

- `dap-jo9.4.1.1` → `TC-AFP-FEASIBILITY-001`: **specification owner**
- `dap-jo9.4.1.2` → `TC-AFP-FEASIBILITY-001`: **implements** (canonical owner `dap-jo9.4.1.1`) — bounded AFP feasibility assessment and recommendation only; no conversion implementation

## Readiness rules

- **Refinement-ready** means a planning/specification task can be detailed using current information; it does not mean implementation is authorized or product tests pass.
- **Decision-blocked** means one or more profile, feasibility, fidelity, tool, privacy, or authority decisions must be recorded before implementation.
- **Definition of Ready:** a specific story, proposed persona, linked requirement, function contract, actionable Beads task, synthetic case IDs with Given/When/Then, and explicit unresolved decisions.
- **Definition of Done for planning:** IDs and edges validate; each requirement traces through a story, function, task, and case; family coverage exists in both lanes; test states say specification-only; and open decisions are visible.
- **Definition of Done for future product work:** write failing tests first, implement the smallest slice, and record passing results; demonstrate preservation/failure behavior; verify licenses and versions; independently record machine, human semantic, and assistive-technology review. A planning case is not a passing test.
