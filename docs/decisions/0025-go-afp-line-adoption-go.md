# Decision 0025: go-afp-line adoption GO — owner record

**Status: DRAFT.** This is a draft record for the product owner to ratify.
It is not an adopted decision until the owner ratifies it. It records the
owner's GO as given; it does not extend it.

**Date drafted: 2026-10-23.**

**SCOPE: TBD — owner decision #1.** The GO recorded below does not state
whether it covers test-only use or product runtime use too. No scope is
asserted anywhere in this record. The scope question is owner decision #1
and remains open until the owner answers it.

## Context

[Decision 0022](0022-go-afp-line-re-review.md) re-reviewed `go-afp-line`
at commit `492c4c7` per Decision 0021 §g + §k and returned a verdict of
CONDITIONAL adoption, subject to 6 recorded conditions. Decision 0022 as
written did not approve adoption: the call stayed with the product owner
under `dap-jo9.4.2`.

## Recorded GO

On 2026-10-22, the product owner gave the adoption GO in the words
"Adopt and proceed" for the exact reviewed revision. Following that GO,
`go-afp-line` was added as a direct dependency in the root `go.mod` and
the DAP-side integration (pin test and integration test under
`internal/afp/`, Decision 0022 condition 6) merged via PR #187
("feat(deps): pin go-afp-line 492c4c7 for AFP profile (dap-jo9.4.2)").

## Exact pin

- Module: `github.com/pratik-saptarshi/go-afp-line`
- Version: `v0.1.0-beta.1.0.20261022163835-492c4c7d35b9`
- Commit: `492c4c7d35b9` (the revision reviewed in Decision 0022)
- Placement: direct dependency in the root `go.mod`; replacement by any
  other path or version is rejected by the pin test in
  `internal/afp/pin_test.go`.

Any future version change is a new adoption review, not a silent upgrade
of this record.

## Scope

SCOPE: TBD — owner decision #1.

- If the owner decides test-only scope: `go-afp-line` must stay out of
  the non-test dependency graph, and the CI guard
  `.github/workflows/afp-import-guard.yml` remains in force.
- If the owner decides product-runtime scope: that decision, recorded by
  ratifying or superseding this draft, is the point at which the guard is
  removed or transferred, per the guard's stated removal condition.

Until the owner decides, the current state is descriptive only: the
library is exercised by the `internal/afp/` tests, and no non-test
package imports it.

## Consequences

- The stale "adoption not approved" statements in the roadmap, product
  scope, README, and decisions index were reconciled to this record's
  state on 2026-10-23; Decision 0022's body is unchanged evidence and
  carries a status pointer to this draft.
- Fidelity validation against the Decision 0020 trusted references
  remains open; the GO does not close it.
- `dap-jo9.4.2` remains open and owner-blocked until the owner ratifies
  the durable form of this GO and answers the scope question
  (decision #1).

## Ratification

To ratify: the owner confirms this record (with the scope question
answered), and this file's status changes from DRAFT to Accepted with
the ratification date. Alternatives for the durable form — a Beads note
or a signed commit — were considered; this ADR-form draft is the form
prepared for the owner's decision #1.
