# v2 non-proceed contingency

Decision record for the non-proceed branch of `dap-jo9.1.6.4` (bounded v2
association-repair feasibility). Recorded by `dap-jo9.1.6.8`.

## Trigger

`dap-jo9.1.6.4` records an assessed **Reject** or **Inconclusive** disposition
for the v2 association-repair feasibility attempt, with evidence. v1 already
ended Reject; this record covers the case where v2 does not proceed.

This record does not itself close, re-gate, or re-scope any issue. When the
trigger fires, the product owner applies the table below and signs off the
re-wiring. Until then, all six stranded issues remain blocked on
`dap-jo9.1.6.4` as designed.

## Stranded issues and proposed paths

Six open issues carry `dap-jo9.1.6.4` as a blocks dependency. Each gets a
re-gated path or an explicit closure. Paths are proposals requiring
product-owner sign-off at trigger time.

| Issue | Proposal | Rationale |
| --- | --- | --- |
| `dap-jo9.2.1.3` Expand remediation corpus after feasibility | Explicit closure | Its premise is "after feasibility". With feasibility rejected, there is no repair capability to expand a corpus for. The characterization fixtures from closed `dap-jo9.2.1.2` remain as evidence. |
| `dap-jo9.2.2` Specify smallest bounded remediation interface | Re-gate: re-scope to characterization/detection interface only | An interface for repair has no feasible basis, but a detection/reporting interface (what characterization observes and reports) can still be specified. Repair operations are explicitly out of scope until a feasible repair path exists. |
| `dap-jo9.2.3` Deliver a bounded baseline remediation slice | Explicit closure | The slice is the repair. No feasible basis. |
| `dap-jo9.2.3.2` Implement the gated baseline structure-repair slice | Explicit closure | The implementation is the repair. No feasible basis. |
| `dap-jo9.2.4.2` Implement bounded semantic repair and reviewer handoff | Re-gate: re-scope to characterization-finding handoff only | The repair half closes; the routing-uncertain-findings-to-review half survives, because characterization still produces findings that need review routing. |
| `dap-jo9.2.5.9` Implement candidate-bound semantic and AT review gate | Explicit closure | The gate needs an eligible candidate. None can exist on a rejected feasibility basis. |

## Consequential closure

`dap-jo9.2.5.10` (Produce an eligible candidate) does not carry a
`dap-jo9.1.6.4` edge, but it is substantively dependent on feasibility: no
eligible candidate can exist on a rejected basis. On trigger, close it with
the same disposition reference.

## Alternative candidate paths

No alternative repair path opens automatically. The alternatives that would
require new issues (not re-gating) are:

- A narrower repair scope (a different association type, or a single family)
  would need a new feasibility spike with its own evidence bar.
- A future v3 attempt with different evidence would need a new feasibility
  spike; it does not revive the closed issues.
- Detection-only / characterization tooling is the surviving track, covered
  by the `.2.2` and `.2.4.2` re-scopes above.

## Re-wiring at trigger

When the trigger fires and the owner signs off:

1. Remove the `dap-jo9.1.6.4` blocks edges from all six stranded issues.
2. Close `dap-jo9.2.1.3`, `dap-jo9.2.3`, `dap-jo9.2.3.2`, `dap-jo9.2.5.9`,
   and `dap-jo9.2.5.10` with a close reason citing this contingency and the
   `dap-jo9.1.6.4` disposition.
3. Re-scope `dap-jo9.2.2` and `dap-jo9.2.4.2` to the narrowed scopes above
   (update descriptions and acceptance criteria; remove repair criteria).
4. `dap-jo9.1.6.4` itself stays open as the disposition record: the gate
   closes only on Proceed, and a Reject/Inconclusive record is the evidence
   that justifies this contingency.

After re-wiring, no open issue remains blocked on a disposition that cannot
arrive: the closed issues are closed, the re-scoped issues no longer depend
on `dap-jo9.1.6.4`, and the gate record stands on its own.
