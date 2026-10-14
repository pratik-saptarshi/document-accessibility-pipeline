# Decision 0004: Approver-persona review protocol and evidence handling

**Status: Accepted.** Product owner approved 2026-10-14; `dap-jo9.2.5.9` may now consume this protocol.

**Date: 2026-10-14.**

## Authority

Product owner designated the reviewer authority as an approver-user persona in `dap-jo9.1.1.4` (closed 2026-10-14). This record proposes the protocol, evidence format, outcome handling, and invalidation rules that were left Open there. It decides no product values beyond the review process itself.

## The persona

The **approver** is a product role, not a named individual. Any person acting as approver exercises the authority to review, approve, or reject remediation candidates on semantic and assistive-technology grounds. The acting person's identity is recorded per review (see evidence format), but the authority belongs to the role.

The approver is distinct from the **operator** defined in [Decision 0003](0003-operator-confirmed-financial-table-mapping.md): the operator confirms a proposed table-header mapping before a repair is applied; the operator holds no semantic-review, assistive-technology, or release-signoff authority. The approver reviews the repaired candidate after the fact and holds the release-gate authority described below.

## Review protocol

1. **Trigger.** A review is triggered when a remediation candidate is presented with a complete evidence bundle: the candidate's SHA-256 hash, the exact change applied, before/after observations, and the candidate-bound machine validation result (for context only).
2. **Separate outcomes.** The approver records two independent outcomes per candidate:
   - **Semantic review:** does the repaired structure say what it should say (e.g., does the row-2 amount cell associate with the correct header)?
   - **Assistive-technology review:** does the structure expose the correct reading order and associations to assistive technology?
   
   Each may pass or fail independently. A semantic pass never implies an AT pass, and neither implies the other.
3. **Machine results are context, not evidence.** A COMPLIANT machine assessment does not satisfy either review; a NONCOMPLIANT one does not decide them either. The approver judges meaning and AT behavior, not validator output.
4. **Genuine review only.** A recorded approval requires a real human review event. Simulations, dry runs, and automated checks must be labeled as such and can never be recorded as a human/AT pass. No auto-approval path exists.
5. **Quorum.** One acting approver per candidate is sufficient for the bounded first workflow. Escalation (below) routes to the product owner, not to a second approver.

## Evidence record format

Each review produces one record bound to the exact candidate bytes:

| Field | Required | Content |
|---|---|---|
| `candidate_sha256` | yes | SHA-256 of the exact reviewed bytes |
| `profile` | yes | Document profile identifier (e.g., financial-tables v2) |
| `tool_versions` | yes | Validator/reader versions used for the machine context (e.g., veraPDF explicit `ua2`, pdfcpu version) |
| `protocol_version` | yes | This decision's identifier and status (0004, Accepted). Records citing a non-Accepted protocol version are treated as missing evidence and block release. |
| `approver_identity` | yes | Who acted as approver for this review |
| `reviewed_at` | yes | UTC timestamp of the review event |
| `semantic_outcome` | yes | `pass` / `fail` / `escalated`, with rationale |
| `at_outcome` | yes | `pass` / `fail` / `escalated`, with rationale |
| `at_tool` | yes | Assistive-technology name and version used for the AT review |
| `at_steps` | yes | Test steps performed during the AT review |
| `at_observations` | yes | Observed issues and behavior during the AT review |
| `rationale` | yes | What was checked and why the outcome was reached |
| `supersedes` | no | Hash of any prior record this review replaces |

The record is immutable once written. Corrections are new records that supersede the old one; the old record is retained.

## Outcome handling

- **Approve** (`pass` on an outcome): the candidate may proceed through the release gate for that outcome. Both outcomes must pass for release; a single pass does not release the candidate.
- **Reject** (`fail` on either outcome): the candidate is blocked from release. The rationale must state what failed. A rejected candidate may be re-presented only after the underlying issue is addressed, as a new candidate with a new hash and a new review.
- **Escalate** (`escalated`): the approver cannot reach a verdict — ambiguous evidence, conflicting observations, or a question outside the approver's scope. The record states the blocking questions and routes to the product owner. An escalated candidate is blocked from release until the product owner resolves it and a fresh review is recorded.
- **Missing evidence**: no review record, or a record missing any required field, blocks release. Absence of evidence is never treated as approval.

## Approval-invalidation rules

An approval is valid only for the exact bytes reviewed. It is invalidated by:

1. **Any byte change to the candidate.** The approval binds to `candidate_sha256`; a different hash is a different candidate and requires a new review.
2. **Profile change.** Reviewing under a different document profile invalidates prior approvals.
3. **Tool/protocol version change.** If the machine-validation context or this protocol changes materially between review and release, the approval is invalidated and the candidate is re-presented.
4. **Supersession.** A newer record naming this record in `supersedes` replaces it.

Invalidation is automatic on hash mismatch; the release gate must re-verify the binding before every release decision.

## Alternatives

- **Named-individual authority:** not adopted. Authority belongs to the role so reviews survive personnel change; identity is still recorded per review.
- **Combined semantic/AT verdict:** not adopted. The two outcomes test different things and fail independently.
- **Auto-approval on machine COMPLIANT:** not adopted. Machine validation is a separate outcome by definition of done and never implies a human/AT pass.

## Consequences and limits

This protocol enables `dap-jo9.2.5.9` to implement the candidate-bound semantic and AT review gate: accepted, missing, rejected, and invalidated paths must be demonstrated with executable tests on an eligible candidate. It does not approve any candidate, does not claim any human/AT pass has occurred, and does not replace the eligibility, machine-validation, or release-approval gates. Simulations remain labeled and never count as review evidence.

## Open follow-ups

- `dap-jo9.1.1.5` — remains open until accepted.
- `dap-jo9.2.5.9` — consumes the accepted protocol; blocked until then.
