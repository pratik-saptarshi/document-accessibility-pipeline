# Decision 0002: Initial remediation family priority

**Status: Accepted for prioritization only.**

**Date: 2026-10-13.**

This records the product owner's explicit instruction. It does not approve product implementation or production scope.

## Authority

Product owner, by explicit instruction: “Statements / invoices with tables matter first and most”.

## Decision

Statements and invoices with tables are the first remediation priority. Neither family is ranked ahead of the other. This decision does not select a writer or generation engine; approve production processing; or resolve language, table topology, semantic mapping, arithmetic, fidelity exceptions, privacy, retention, or reviewer authority.

## Context and evidence

Existing-PDF accessibility remediation is the primary lane under [Decision 0001](0001-pdf-ua-2-validation-target.md); generation remains separate. The [backlog refinement panel](../planning/backlog-refinement-panel-2026-10-13.md) proposed selecting an initial remediation input class. The product owner's instruction selects both statements and invoices with tables as the first priority. Technical tests cannot establish that preference, and this decision claims no product capability.

## Alternatives

Simple correspondence first was not selected. Correspondence and reports remain in the broader remediation scope. Statement-only versus invoice-only precedence is undecided.

## Consequences and limits

Prioritize bounded financial-table requirements and feasibility. Preserve displayed values without calculating or changing them. Generation remains a separate lane. The two-fixture table-header repair experiment remains proposed; this decision approves no detailed experimental parameters, engine, production support, or release eligibility.

Keep the accepted Go-first approach, permissive product-dependency boundary, explicit PDF/UA-2 target, isolated CI-only validator, and separate human semantic and assistive-technology review boundaries from Decision 0001. None of those constraints are evidence that the proposed repair works.

## Test evidence

This is an owner-authorized prioritization, not a repair, fidelity, machine-validation, semantic-review, or assistive-technology test outcome.

## Open follow-ups

- `dap-jo9.1.1.1` is complete: the reviewed configuration-only protocol is versioned in [financial-table experiment protocol v1](../fixtures/remediation/financial-table-experiment.v1.md). Its fixed parameters remain experiment-only.
- `dap-jo9.2.1.2` is complete at PR #43 merge `b462faa070a647176f9f0bf8264cd89cdd531bc1`; its immutable fixtures, independent observations, structural predicates, and negative controls are documented under [financial-table fixtures](../fixtures/README.md). This closes only the pre-spike baseline task; it reports no repair outcome.
- [pdfcpu v0.16.2 financial-table candidate evidence](../research/pdfcpu-financial-table-clearance-2026-10-13.md) is prepared for review. `dap-jo9.1.3` remains in progress; no candidate clearance or repair attempt is implied.
- `dap-jo9.1.2` and `dap-jo9.1.3` cover writer/fidelity feasibility and remediation capability/license clearance.
- `dap-jo9.1.1.3` and `dap-jo9.1.1.4` retain open production-handling and reviewer-authority questions.
- `dap-jo9.17` reconciles mapping integrity without removing case obligations.
