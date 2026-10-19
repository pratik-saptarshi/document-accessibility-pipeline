# Decision 0018: Keep AFP within the pipeline; retain its feasibility gates

**Status: Accepted.** On 2026-10-19, the product owner decided that AFP migration does not merit a separate project. `go-afp-line` remains a separately maintained library intended for use by `document-accessibility-pipeline`; the organizational decision does not adopt the library, select an AFP source profile, or authorize conversion delivery.

## Decision

Keep AFP planning and any later delivery within this repository's product and governance scope. Treat `go-afp-line` as a candidate library dependency with its own repository, releases, and maintenance. Maintain an AFP feasibility workstream in this pipeline so its source profile, fidelity evidence, dependency fit, and go/no-go are decided before delivery work is authorized.

## Evidence and limits

The pinned source findings and their limits are maintained in the [AFP refinement record](../planning/afp-refinement-2026-10-19.md). They apply only to the reviewed `go-afp-line` revision and do not establish behavior of later revisions or suitability for an unselected profile. A zero-exclusion result by itself is not proof of faithful output.

This decision resolves only the project-organization question. It does not accept an input profile, customer or reviewer persona, workload, reference corpus, fidelity threshold, dependency, adoption, output accessibility, or delivery scope. The internal document-operations review outcome is a working hypothesis pending owner confirmation. `dap-jo9.4.1`, `dap-jo9.4.1.2`, and `dap-jo9.4.2` must not be closed or unblocked by this decision while their feasibility and go/no-go acceptance obligations remain unmet.

## Follow-up

- Refine the AFP backlog as a gated sequence: owner profile/reference decisions; feasibility of the first useful profile; separate library dependency review and adoption decision; then delivery refinement only if the owner approves the outcome and gates.
- Preserve the existing feasibility criteria in `REQ-AFP-001` and `TC-AFP-FEASIBILITY-001` until their coverage is explicitly reconciled into the feasibility work and traceability artifacts.
- The proposed issue decomposition, priorities, dependencies, delivery acceptance, and unresolved owner decisions are recorded in [AFP refinement plan](../planning/afp-refinement-2026-10-19.md). Tracker changes remain pending a supported and reviewed Beads reconciliation.
