# Decision 0024: Approver-register JSON key case policy — liberal parsing accepted

**Status: Accepted.** On 2026-10-23, the product owner accepted liberal key parsing for the approver-register JSONL wire format, with an evidence-based revisit trigger.

## Context

A P2 review thread on PR #183 ("Reject case-variant JSON field names") observed that `encoding/json` matches tagged struct fields case-insensitively (even with `DisallowUnknownFields`), while the duplicate-key scan compares names case-sensitively — so entries containing `"SEQ"` or both `"seq"` and `"SEQ"` are accepted and normalized into the same signed field. The thread asked for exact-spelling validation of every decoded key before struct decoding.

Note the boundary: duplicate keys (exact-match) are still rejected per the earlier Decision 0015 fix round. This decision concerns *case variants only*.

## Decision

Liberal parsing is accepted for now: case-variant JSON field names are normalized by `encoding/json` into the same signed field. No exact-spelling enforcement is added at this time. The behavior is documented in a code comment at the parse site.

## Rationale

Owner's call (2026-10-23): the register writer is a closed, owner-controlled producer — the realistic key source is the canonical writer, not adversarial hand-crafted JSON. Strictness here buys little while the duplicate-key rejection (the actual ambiguity hazard) is already enforced. Fail-closed behavior on signatures, hash chaining, and temporal policy is unaffected.

## Revisit trigger (evidence-based)

Revisit this decision if testing produces evidence that case-variant keys cause interop failures or hostile-input concerns — e.g., a fuzzer or integration test demonstrates two byte-distinct entries normalizing to the same signed field with divergent security consequences, or a second register producer emerges with different casing conventions. No numeric thresholds are invented here; the trigger is observed evidence, not speculation.

## Alternatives considered

- **Strict exact-spelling enforcement (deferred):** validate every decoded key against the exact permitted spelling before struct decoding, as the thread proposed. Deferred, not rejected — it remains the fallback if the revisit trigger fires.
