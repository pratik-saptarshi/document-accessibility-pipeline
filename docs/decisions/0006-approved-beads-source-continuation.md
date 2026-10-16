# Decision 0006: Exact Beads source-continuation approval

**Status: Accepted.** The product owner approved the exact continuation in the owner conversation at 02:25 on 2026-10-16 (timezone not stated). The independent post-merge review confirmed that the approval was valid. The PR #85 bot review comment [`r4290631458`](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/85#discussion_r4290631458) identified that the approval lacked an indexed durable decision record and prompted this record; it is not the source of the approval. This record makes the owner's bounded approval durable without expanding it.

**Date: 2026-10-16.**

## Decision

Accept one exact continuation of the pre-existing Beads import/export null-preservation exception: source snapshot `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890` and its explicitly approved continuation `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de` may use the existing Beads 1.3.1 compatibility path for the same five explicit-null field pairs listed below, using only the Beads executable with SHA-256 `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`.

The continuation preserves the original authorization's exact scope. It does not authorize any other source digest, binary, version, field pair, changed null value, or dependency-graph delta. Any mismatch continues to fail closed and requires separate approval before the allowlist or behavior changes.

## Context and evidence

Beads 1.3.1 import/export omits explicit JSON `null` values for the listed fields. The reconciliation exception preserves those original `null` bytes in the candidate only when the source snapshot, importer version and binary, and exact field pair all match the approval. The current implementation encodes the two full source digests in an exact allowlist and pins version and binary digest; it does not accept prefixes, wildcards, or caller-supplied digests as authorization.

The owner approval is identified as Sentinel `ed038ff9e220819189f25a67aad33ae0` in the owner conversation at 02:25 on 2026-10-16 (timezone not stated). It approves continuation digest `768d13dfa4ab4b1fc337a6fde9bf071e6c3267c2013311e9871db6ab54de96de`, the existing five pairs, and the pinned executable above. The source snapshot before that continuation is `7ce7e649acd2d9574b785f655802f3afecbf6a001fc4ccb04f5ac56bf1125890`.

For each listed pair, the authorized canonical source contains the explicit JSON token `null`; the two imported/exported baselines omit the field; and the candidate retains the original `null` bytes. No field or dependency-graph change is covered by this exception.

| Issue | Field whose original explicit `null` is preserved |
|---|---|
| `dap-jo9.5.3` | `dependencies` |
| `dap-jo9.1.6.10` | `closed_at` |
| `dap-jo9.5.6` | `closed_at` |
| `dap-jo9.2.1.4` | `closed_at` |
| `dap-jo9.5.7` | `closed_at` |

The merged implementation and regression tests provide the technical enforcement evidence. `TestApprovedCurrentSourceContinuationPreservesNulls` exercises source-specific reconciliation with the importer-response fixture and verifies null-byte retention plus second-store re-derivation; it does not invoke the Beads executable. `TestApprovedNullLossRegressionUsesHistoricFixtureAndRejectsLaterCanonicalEdit` retains the historic-source case and rejects altered source bytes. `TestReconcileScopedNullLossRejectsWrongBinaryAndChangedNullField` rejects a wrong binary digest or changed null field. Separately, `TestCreateCommandReconcilesApprovedCurrentSourceInDisposableStores` invokes the pinned Beads executable in disposable stores and checks the real import/export and creation path. No live tracker promotion or canonical JSONL mutation is evidence for this decision.

## Alternatives considered

- **Keep only the historic digest and reject the continuation:** rejected because the owner explicitly approved this one exact continuation while retaining the same pairs and pinned importer.
- **Allow a digest prefix, wildcard, configuration-provided digest, or any later source:** rejected; that would exceed the approval and weaken fail-closed lineage checks.
- **Broaden the exception to other issue fields, values, graph changes, versions, or binaries:** rejected; none is included in the approval.
- **Replace Beads or migrate the export format as part of this decision:** not selected; that would be a separate technical and product decision.

## Consequences and limits

The exact `768d13…96de` continuation can use the already implemented null-preservation path under the same Beads 1.3.1 executable. A source change produces a different digest and remains unauthorized unless a separately reviewed and explicitly approved decision adds that exact digest. The old snapshot and this single continuation are the complete approved source set.

This decision records source-lineage approval only. It does not add or modify a Beads issue, promote a live tracker state, authorize arbitrary issue creation, or permit hand-editing `.beads/issues.jsonl`. Any actual tracker operation remains a separate next step after this record is accepted in protected main; it must use the guarded reconciler and independently reviewed evidence. No new tracker ID is assigned by this record.

## Test evidence

The exact-source continuation, historical exception, altered-source rejection, version/binary pinning, and disposable creation round trip are covered by the merged PR #85 tests in `internal/beadsreconcile/reconcile_test.go` and `cmd/beads-export-reconcile/creation_contract_test.go`. The PR's validation and review record is in [PR #85](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/85). This documentation change does not claim to rerun those behavioral tests.

## Follow-up

- PR #85 post-merge bot review comment [`r4290631458`](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/85#discussion_r4290631458) prompted this durable record; the owner conversation above is the approval source. After this decision is merged, handle the separately requested actual tracker issue creation/promotion as its own reviewed, guarded action. It has no new Beads ID yet; do not invent or pre-assign one here.
- Any future continuation, importer change, additional pair, or graph delta requires separate explicit owner approval, tests, independent review, and an indexed decision record before promotion.
