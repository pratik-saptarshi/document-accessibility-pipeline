# One-time exception: financial-remediation text updates

**Status:** Applied once to the exact batch below. This is an agent-transcribed record of a user instruction, not a document manually authored, signed, or cryptographically authenticated by the owner. It creates no continuing authority or policy.

**Instruction received:** 2026-10-17 23:20:50 UTC. In response to the precise guarded-path blocker, the owner instructed: “Bypass safeguard and proceed.” This is interpreted narrowly as a one-time exception for the exact reviewed text-only update batch below.

## Exact scope

- Reviewed base commit: `e15b352e398581bef24aeeb679e45efdf442286d`.
- Pre-change source `.beads/issues.jsonl` SHA-256: `441e193c7b5ed4dc189102f378f164dcce34cda7d212c1bae19d6267f230daec`.
- Resulting canonical candidate `.beads/issues.jsonl` SHA-256: `9d5438fa5b046dd44ca32da878d434cb548b8eff897934762abc7ac8b19566a`.
- Exact reviewed JSON value batch: [financial-remediation-updates.json](../planning/proposals/2026-10-17-financial-remediation-updates.json), SHA-256 `4d5359b8d3896428bc59d2e88ccdef6496a70fd57548048ad93cf58148e29aee`.
- Scope: nine existing issues, ten text-field values, each preserving its current value as an exact prefix: `.1.6.3` notes; `.1.6.4` description and notes; `.2.1` description; `.2.1.3` acceptance criteria; `.2.2` acceptance criteria; `.2.3` description; `.2.3.2` acceptance criteria; `.2.5.10` acceptance criteria; `.2.5.9` notes.
- Preserve all other bytes and every other field: all issue IDs, statuses, types, priorities, owners, parents, the 168 dependency edges, and the explicit-null `.5.3.dependencies` value. Do not use a full `bd export` to rewrite the source.
- No issue creation, status transition, closure, dependency change, source/null exception, guard-code change, or policy-tool change is authorized. In particular, the proposed `.2.1` child task is not created; its source/resource/prerequisite decisions remain open.

## Limited override and continuing controls

For this exact batch only, the instruction overrides Decision 0011's prohibition on direct canonical JSONL text edits because the existing guarded source-to-baseline check fails on `.1.6.10` timestamp precision normalization and disposable import/export loses `.5.3.dependencies: null`. It does not modify Decision 0011 or create a standing direct-edit policy. R5-02 remains undecided; the general tracker/governance freeze and all other safeguards remain in force.

No platform permission denial was bypassed. The exact base source and each target field were rechecked; the original serialized bytes for each existing string were preserved as a prefix. Only the ten approved string values changed. Independent comparison confirmed all non-target rows byte-identical, all non-target issue fields and all 168 edges unchanged, and explicit-null `.5.3.dependencies` retained. The canonical candidate SHA-256 is `9d5438fa5b046dd44ca32da878d434cb548b8eff897934762abc7ac8b19566a8`. The machine-readable digest-only comparison is [financial-remediation-text-batch-delta-2026-10-17.json](evidence/financial-remediation-text-batch-delta-2026-10-17.json). This is a direct one-time edit, not a Beads-generated bundle or promotion. Repository schema/traceability and diff checks and independent review remain required before publication; protected review and merge rules remain unchanged.

## Evidence

The pre-change disposable guarded update probe produced no bundle and changed no canonical data. It failed before plan application because source-to-baseline validation detected `.1.6.10` `closed_at`/`updated_at` timestamp-precision normalization. A separate disposable import/export confirmed loss of `.5.3.dependencies: null`. The current pinned Beads 1.3.1 source and binary do not cover the current source hash in their null-loss allowance.

The manifest contains hashes and field names only, not full tracker snapshots or private chat. No issue was created or closed, no status or relationship changed, and no guard or policy changed. Only the ten listed text fields on nine existing issues changed.

A later review instruction requested clarification of the historical terminal wording in three applied note values. No canonical text value was changed in response: the proposal and plan now state that the phrase is retained from the pre-authorization draft, while the containing text fields are applied and historical closure/future implementation remain unapplied. This documentation clarification does not broaden the original field scope.

Three applied note values (`.1.6.3`, `.1.6.4`, and `.2.5.9`) contain a terminal phrase saying a paragraph/text is “proposed and unapplied.” That phrase is retained from the original pre-authorization draft text; the exact containing note values are applied in this batch. The phrase refers to the status of the historical paragraph when written, not the present application status of the text update. Historical issue closure/status transitions and future implementation or acceptance remain unapplied/unapproved.
