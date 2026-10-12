# Acceptance case specifications

**Status: specifications only; none of these product cases have been executed.** Cases use synthetic/deidentified fixtures. Open thresholds, profiles, and reviewer authority must be resolved before implementation readiness.

## TC-REM-STATEMENT-001 — remediation / statements

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic statement PDF matches the agreed supported-input profile and baseline fixture
- **When:** The bounded baseline characterization and repair contract is applied
- **Then:** A candidate and change manifest are produced; original remains intact and declared preservation observations are available.

## TC-REM-INVOICE-001 — remediation / invoices

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic invoice PDF contains declared text, totals, and page geometry
- **When:** It is processed by the proposed baseline repair contract
- **Then:** The output is compared against the invoice fixture and any unsupported arithmetic or layout change is reported.

## TC-REM-CORRESPONDENCE-001 — remediation / correspondence

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic letter PDF has known language, paragraphs, and reading order
- **When:** It is characterized and repaired within the supported baseline profile
- **Then:** Text, page count, links, metadata, and declared structure are compared and unverified semantics remain unverified.

## TC-REM-REPORT-001 — remediation / reports

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic report PDF contains headings, a table, and a chart caption
- **When:** Baseline repair is attempted
- **Then:** Known table and figure semantics are preserved or routed as unsupported; no guessed reading order is reported as correct.

## TC-REM-MALFORMED-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic malformed or encrypted PDF is outside the declared supported profile
- **When:** Intake classifies the file
- **Then:** The result is a bounded unsupported/error outcome and no candidate is emitted as successful.

## TC-REM-FIDELITY-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.3.1`
- **Status:** specification-only, not executed
- **Given:** An agreed synthetic fixture has baseline text, pages, geometry, links, forms, and annotations
- **When:** A candidate is compared with the baseline
- **Then:** Each applicable feature has pass, difference, or unknown evidence; missing checks do not become pass.

## TC-REM-AMBIGUOUS-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.4.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic PDF contains ambiguous reading order or table semantics
- **When:** Semantic analysis cannot meet an agreed confidence/decision rule
- **Then:** The output is review-needed with uncertainty evidence, not an automatic semantic guess.

## TC-REM-REVIEW-ROUTE-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.4.1`
- **Status:** specification-only, not executed
- **Given:** A repair candidate has unresolved semantic findings
- **When:** The human review gate is evaluated
- **Then:** Release remains pending until an authorized reviewer records a decision; reviewer authority is still an open decision.

## TC-REM-MACHINE-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.5.1`
- **Status:** specification-only, not executed
- **Given:** A candidate and a machine-check profile are supplied
- **When:** The independent validator returns pass, finding, unsupported, or error
- **Then:** The result records the validator and version with its status; it is not presented as a semantic or compliance verdict.

## TC-REM-HUMAN-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.5.2`
- **Status:** specification-only, not executed
- **Given:** A candidate has machine findings and a human semantic checklist
- **When:** A reviewer assesses reading order, labels, and meaning
- **Then:** The semantic decision and unresolved findings are separately recorded from machine results.

## TC-REM-ASSISTIVE-TECH-001 — remediation / cross-family

- **Beads:** `dap-jo9.2.5.2`
- **Status:** specification-only, not executed
- **Given:** A candidate is eligible for assistive-technology review under an agreed protocol
- **When:** The review is performed
- **Then:** Tool/version, test steps, observed issues, and outcome are recorded separately; no protocol is presumed approved.

## TC-GEN-STATEMENT-001 — generation / statements

- **Beads:** `dap-jo9.3.3.1`
- **Status:** specification-only, not executed
- **Given:** Synthetic statement data satisfies its proposed field and arithmetic profile
- **When:** The proposed statement renderer is invoked
- **Then:** A candidate output, source-to-output manifest, and separately reviewable semantic structure are produced.

## TC-GEN-INVOICE-001 — generation / invoices

- **Beads:** `dap-jo9.3.3.1`
- **Status:** specification-only, not executed
- **Given:** Synthetic invoice data has line items and declared currency/rounding rules
- **When:** Validation and rendering are requested
- **Then:** Totals are checked against the declared rules before a candidate is produced; unresolved policy blocks acceptance.

## TC-GEN-CORRESPONDENCE-001 — generation / correspondence

- **Beads:** `dap-jo9.3.4.1`
- **Status:** specification-only, not executed
- **Given:** Synthetic correspondence data provides language, title, paragraphs, and recipient fields
- **When:** The proposed renderer processes the template
- **Then:** Declared title, headings, language, and reading order are represented and missing bindings are explicit.

## TC-GEN-REPORT-001 — generation / reports

- **Beads:** `dap-jo9.3.4.1`
- **Status:** specification-only, not executed
- **Given:** Synthetic report data includes sections, a table, and provenance metadata
- **When:** The proposed report renderer is invoked
- **Then:** Sections and tabular relationships have explicit structure and unmapped values stop or flag rendering.

## TC-GEN-INVALID-DATA-001 — generation / cross-family

- **Beads:** `dap-jo9.3.1.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic input omits a required field or supplies a value with the wrong type
- **When:** Data/template validation runs
- **Then:** Field-level findings are returned and rendering does not proceed.

## TC-GEN-VALIDATION-001 — generation / cross-family

- **Beads:** `dap-jo9.2.5.1`
- **Status:** specification-only, not executed
- **Given:** A generated candidate is checked with a proposed independent machine profile
- **When:** Machine validation and human review gates are evaluated
- **Then:** Machine, semantic, and assistive-technology outcomes remain separate; no single result implies compliance.

## TC-GEN-SOURCE-FEASIBILITY-001 — generation / cross-family

- **Beads:** `dap-jo9.3.4.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic DOCX or Markdown source fixture is supplied
- **When:** Source conversion feasibility is assessed
- **Then:** Structure, fidelity, licensing, and unsupported features are documented before support is claimed.

## TC-GEN-ENGINE-GATE-001 — generation / cross-family

- **Beads:** `dap-jo9.3.4.1`
- **Status:** specification-only, not executed
- **Given:** No generation engine has an approved evidence record
- **When:** A generation implementation task is considered
- **Then:** Engine-dependent work remains decision-blocked until a comparative feasibility decision is recorded.

## TC-CROSS-HOSTILE-001 — cross-lane / cross-family

- **Beads:** `dap-jo9.5.1`
- **Status:** specification-only, not executed
- **Given:** Either lane receives a synthetic hostile metadata or resource-exhaustion fixture
- **When:** The input is processed under a proposed resource policy
- **Then:** Processing is bounded and classified with no silent success or content-bearing diagnostic leakage.

## TC-CROSS-TIMEOUT-001 — cross-lane / cross-family

- **Beads:** `dap-jo9.5.1`
- **Status:** specification-only, not executed
- **Given:** A synthetic worker exceeds the proposed timeout or crashes
- **When:** The job controller observes the failure
- **Then:** The job enters an explicit retry, terminal, or review state and records bounded evidence.

## TC-CROSS-PRIVACY-001 — cross-lane / cross-family

- **Beads:** `dap-jo9.5.1`
- **Status:** specification-only, not executed
- **Given:** A fixture contains a synthetic marker representing sensitive content
- **When:** Errors and audit events are emitted
- **Then:** Ordinary diagnostics omit the marker and retention/access policy is clearly left for decision.

## TC-CROSS-AUDIT-001 — cross-lane / cross-family

- **Beads:** `dap-jo9.5.1`
- **Status:** specification-only, not executed
- **Given:** A job records intake, transformation, validation, and review events
- **When:** An audit record is inspected
- **Then:** Event ordering, artifact identity, and decision provenance are traceable without claiming an unselected integrity mechanism.

## TC-AFP-FEASIBILITY-001 — AFP feasibility / AFP separate track

- **Beads:** `dap-jo9.4.1.1`
- **Status:** specification-only, not executed
- **Given:** AFP input/output goals, safe representative fixtures, and fidelity criteria are not yet agreed
- **When:** The feasibility issue is refined
- **Then:** Missing prerequisites are listed and the go/no-go task stays blocked; no converter is selected or implemented.
