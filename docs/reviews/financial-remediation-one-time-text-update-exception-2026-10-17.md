# One-time exception: financial-remediation text updates

**Status:** Applied once to the exact text batch and the narrow owner-approved extension below. This is an agent-transcribed record of scope; it is not owner-signed or cryptographically authenticated. No private conversation transcript or message wording is reproduced.

**Authorization provenance:** The initial one-time instruction to apply the exact reviewed text batch was recorded on 2026-10-17. A later affirmative owner approval extended that exact scope to correct the final sentences of three notes and set truthful updated_at values on the same nine issues. The approval is represented here only as agent-transcribed scope.

## Exact scope

- Reviewed base commit: e15b352e398581bef24aeeb679e45efdf442286d.
- Pre-change .beads/issues.jsonl SHA-256: 441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec.
- Prior text-only candidate SHA-256: 9d5438fa5b046dd44ca32da878d434cb548b8eff897934762abc7ac8b19566a8.
- Current candidate .beads/issues.jsonl SHA-256: 4e3fede04c2133e3bdba04e48fa13e703415e991d3e831cae850212bf52bd095.
- Current review JSON batch: [financial-remediation-updates.json](../planning/proposals/2026-10-17-financial-remediation-updates.json), SHA-256 bf68a9918cf9ca25c74962e74c1bd1342a181d671aaed6ff8cc35bb4ba532239.
- Current digest-only manifest: [financial-remediation-text-batch-delta-2026-10-17.json](evidence/financial-remediation-text-batch-delta-2026-10-17.json), SHA-256 71a08366db5fc7cc7f89348b49c81c04cf714ee47ab35708bf0a2e3eabe5b736.
- Original batch: ten text-field values on nine existing issue IDs: .1.6.3, .1.6.4, .2.1, .2.1.3, .2.2, .2.3, .2.3.2, .2.5.10, and .2.5.9.
- Approved extension: replace only the terminal sentence in notes for .1.6.3, .1.6.4, and .2.5.9, making explicit that note text is applied while status/future acceptance remains pending; set updated_at to 2026-10-17T23:44:58Z on those same nine issue records. A subsequent factual correction changed only the terminal sentence and updated_at of .1.6.3 to 2026-10-17T23:53:35Z; the other eight timestamps remain 2026-10-17T23:44:58Z.
- No other fields changed. Preserve issue IDs, statuses, types, priorities, owners, parents, all 168 dependency edges, and explicit-null .5.3.dependencies. Do not create or close issues, change status/relationships, alter guards/policy, or create the proposed .2.1 child task.

## Limited override and continuing controls

For the exact batch and extension above only, the owner-approved instruction permits these direct canonical JSONL edits because the guarded source-to-baseline check fails on .1.6.10 timestamp precision normalization and disposable import/export loses .5.3.dependencies: null. It does not modify Decision 0011 or create a standing direct-edit policy. R5-02 remains undecided; the general tracker/governance freeze and all other safeguards remain in force.

No platform permission denial was bypassed. The original ten text updates, three note-ending corrections, and nine updated_at values are captured in the current review JSON. The original note content preceding each corrected terminal sentence is unchanged. From the prior text-only candidate, the original extension changed updated_at on nine issue IDs and notes on those three IDs. The subsequent factual correction changed only .1.6.3 notes and updated_at; all other records and the other eight timestamps are unchanged. From the reviewed base, only the ten listed text fields and those nine updated_at fields differ. The manifest verifies 101 rows, 168 identical dependency edges, byte-identical non-target rows, unchanged explicit null, and field-level before/after hashes.

This is a direct one-time edit, not a Beads-generated bundle or promotion. The pre-change disposable guarded update probe produced no bundle and changed no canonical data; it failed before plan application because source-to-baseline validation detected .1.6.10 closed_at/updated_at precision normalization. Disposable import/export also confirmed loss of .5.3.dependencies: null. The current pinned Beads source/binary does not cover the current source hash in its null-loss allowance. The exception does not change those limitations or authorize another edit.

The revised note endings now state respectively: the planning text is applied while .1.6.3 remains in_progress, its historical status transition remains pending, and no v3 authoring or future implementation is authorized (the historical repair/candidate evidence exists and bounded Proceed was accepted); the planning text is applied while .1.6.4 remains open and historical disposition/future implementation or acceptance remain pending; and the planning text is applied while .2.5.9 remains open and genuine semantic/AT acceptance remains pending. No issue status was changed.
