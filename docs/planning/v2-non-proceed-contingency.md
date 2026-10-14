# v2 non-proceed contingency

Decision record for the non-proceed branch of the v2 association-repair
feasibility attempt. Recorded by `dap-jo9.1.6.8`.

## Trigger

`dap-jo9.1.6.3` (the bounded feasibility attempt) closes with an assessed
**Reject** or **Inconclusive** disposition — its finite outcome, recorded
after independent acceptance of the actual attempt evidence. `dap-jo9.1.6.4`
is a Proceed-only gate: a negative outcome leaves it open, so the
contingency is keyed to the `.1.6.3` disposition, not to `.1.6.4` recording
anything.

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
| `dap-jo9.2.2` Specify smallest bounded remediation interface | Re-gate: re-scope to characterization/detection interface only, gated on `dap-jo9.2.1` | An interface for repair has no feasible basis, but a detection/reporting interface (what characterization observes and reports) can still be specified against the fixture plan and characterization evidence that `dap-jo9.2.1` produces. Repair operations are explicitly out of scope until a feasible repair path exists. |
| `dap-jo9.2.3` Deliver a bounded baseline remediation slice | Explicit closure | The slice is the repair. No feasible basis. |
| `dap-jo9.2.3.2` Implement the gated baseline structure-repair slice | Explicit closure | The implementation is the repair. No feasible basis. |
| `dap-jo9.2.4.2` Implement bounded semantic repair and reviewer handoff | Re-gate: re-scope to characterization-finding handoff only, gated on `dap-jo9.2.2` | The repair half closes; the routing-uncertain-findings-to-review half survives, because characterization still produces findings that need review routing. The handoff implements routing for the detection interface that re-scoped `dap-jo9.2.2` specifies, which itself rests on `dap-jo9.2.1`. |
| `dap-jo9.2.5.9` Implement candidate-bound semantic and AT review gate | Re-gate: keep open, gated on a future eligible candidate | Closure requires executable tests on an eligible candidate, and a no-go cannot satisfy closure. Closing it would falsely mark an unfinished gate complete and leave future v3 or narrower-repair alternatives without their tracked candidate-review gate. The gate stays open; its closure criterion is unchanged. |

## Consequential re-gate

`dap-jo9.2.5.10` (Produce an eligible candidate) does not carry a
`dap-jo9.1.6.4` edge, but it is substantively dependent on feasibility: no
eligible candidate can exist on a rejected basis. It stays open, re-gated
to a future candidate-producing basis. Closing it on trigger would satisfy
`.2.5.9`'s only candidate-producing dependency even though no candidate
exists, and would bypass its own acceptance criteria (an eligible candidate
with hash-bound evidence, or the eligibility standard recorded as the open
blocker with owner and date). Its closure criterion is unchanged.

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
2. Close `dap-jo9.2.1.3`, `dap-jo9.2.3`, and `dap-jo9.2.3.2` with a close
   reason citing this contingency and the `dap-jo9.1.6.3` disposition.
3. Re-scope `dap-jo9.2.2` to the characterization/detection interface and
   add a blocks edge to `dap-jo9.2.1` (its evidence prerequisite).
   Re-scope `dap-jo9.2.4.2` to characterization-finding handoff, remove
   its `dap-jo9.2.3` edge (the slice closes), and add a blocks edge to
   re-scoped `dap-jo9.2.2`.
4. Keep `dap-jo9.2.5.9` and `dap-jo9.2.5.10` open with a note that closure
   still requires a future eligible candidate from a new feasibility basis
   (`.2.5.10`), and executable tests on one (`.2.5.9`).
5. `dap-jo9.1.6.4` itself stays open as the Proceed-only gate record: it
   closes only on Proceed, and the `.1.6.3` Reject/Inconclusive record is
   the evidence that justifies this contingency.

After re-wiring, no open issue remains blocked on a disposition that cannot
arrive: the closed issues are closed, the re-scoped issues depend on live
evidence-producing work (`.2.1`, `.2.2`), the candidate gate stays open for
a future candidate, and the gate record stands on its own.
