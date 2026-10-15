# Decision 0002: Initial remediation family priority

**Status: Accepted for prioritization only.**

**Date: 2026-10-13.**

This records the product owner's explicit instruction. It does not approve product implementation or production scope.

**Evidence reconciliation (2026-10-15):** the statements below about the v2 reports being unavailable and no finite Proceed being accepted describe the evidence state when this decision was recorded. The four retained bundles and complete candidate-bound reports/contracts have since been reconciled and independently reviewed; the source-thread final assessment records a finite historical engineering Proceed only for the paired row-two `/Headers` repairs. See [proposed Decision 0005](0005-historical-v2-bounded-association-proceed.md), the [reconciled evidence ledger](../research/pdfua2-retained-bundle-ledger-2026-10-15.md), and the [acceptance register](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md). Both candidates remain NONCOMPLIANT. Decision 0005 remains a proposed repository record; this evidence reconciliation does not change this prioritization decision or authorize implementation, candidate eligibility, conformance, production use, human semantic/assistive-technology acceptance, or release. `.1.6.3` remains in progress and `.1.6.4` remains open; their tracker transitions have not been applied pending a fidelity-preserving update path. The dated notice supersedes only the later v2-report availability, review, and no-Proceed statements below; the original decision scope and history remain.

## Authority

Product owner, by explicit instruction: “Statements / invoices with tables matter first and most”.

## Decision

Statements and invoices with tables are the first remediation priority. Neither family is ranked ahead of the other. This decision does not select a writer or generation engine; approve production processing; or resolve language, table topology, semantic mapping, arithmetic, fidelity exceptions, privacy, retention, or reviewer authority.

## Context and evidence

Existing-PDF accessibility remediation is the primary lane under [Decision 0001](0001-pdf-ua-2-validation-target.md); generation remains separate. The [backlog refinement panel](../planning/backlog-refinement-panel-2026-10-13.md) proposed selecting an initial remediation input class. The product owner's instruction selects both statements and invoices with tables as the first priority. Technical tests cannot establish that preference, and this decision claims no product capability.

## Alternatives

Simple correspondence first was not selected. Correspondence and reports remain in the broader remediation scope. Statement-only versus invoice-only precedence is undecided.

## Consequences and limits

Prioritize bounded financial-table requirements and feasibility. Preserve displayed values without calculating or changing them. Generation remains a separate lane. Decision 0003 records the separately authorized operator-confirmed mapping boundary for the first workflow. The v1 pdfcpu full-write no-op is a **REJECT for that exact operation only**; it does not reject a production writer generally. V2 positive no-op controls passed and a bounded repair attempt produced separate statement/invoice candidates, but their completed candidate-bound `ua2` assessments were NONCOMPLIANT. The latest full raw XML/contracts are not durably available for review, so there is no accepted v2 finite Proceed disposition, candidate eligibility, production support, or release approval.

Keep the accepted Go-first approach, permissive product-dependency boundary, explicit PDF/UA-2 target, isolated CI-only validator, and separate human semantic and assistive-technology review boundaries from Decision 0001. None of those constraints are evidence that the proposed repair works.

## Test evidence

This is an owner-authorized prioritization, not a repair, fidelity, machine-validation, semantic-review, or assistive-technology test outcome.

## Open follow-ups

- `dap-jo9.1.1.1` is complete: the reviewed configuration-only protocol is versioned in [financial-table experiment protocol v1](../fixtures/remediation/financial-table-experiment.v1.md). Its fixed parameters remain experiment-only.
- `dap-jo9.2.1.2` is complete at PR #43 merge `b462faa070a647176f9f0bf8264cd89cdd531bc1`; its immutable fixtures, independent observations, structural predicates, and negative controls are documented under [financial-table fixtures](../fixtures/README.md). This closes only the pre-spike baseline task; it reports no repair outcome.
- The accepted v1 finite REJECT is documented in the [v1 no-op result](../research/pdfcpu-financial-table-noop-2026-10-13.md). It is limited to pdfcpu v0.16.2's measured font-metric expansion on the exact v1 positive fixtures.
- The [v2 repair result](../research/pdfcpu-financial-table-v2-repair-2026-10-13.md) records the historical candidate bytes and completed NONCOMPLIANT assessments. The report retention, assessment, and independent acceptance have since been completed for the exact paired operation and are recorded in the dated reconciliation above. `.1.6.3` remains in progress and `.1.6.4` remains open because their tracker transitions remain unapplied pending a fidelity-preserving update path; do not infer conformance or candidate eligibility.
- `dap-jo9.1.6.6` separately tracks a P2 mutable-input snapshot boundary. `dap-jo9.1.6.5` remains the distinct filesystem-latency follow-up; neither adds a blocker to the current fixed-fixture attempt.
- `dap-jo9.1.6.7` tracks future financial-table mapping-options discovery. It does not change the operator-confirmed first-workflow boundary or authorize automatic mapping.
- `dap-jo9.1.2` and `dap-jo9.1.3` record earlier writer/fidelity and exact candidate/license boundaries; their closure does not resolve the separate v2 `.1.6.4` gate.
- `dap-jo9.1.1.3` and `dap-jo9.1.1.4` retain open production-handling and reviewer-authority questions.
- `dap-jo9.17` reconciles mapping integrity without removing case obligations.
