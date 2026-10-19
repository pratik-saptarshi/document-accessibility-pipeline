# AFP refinement and prioritization — 2026-10-19

**Status: Proposed planning refinement.** This record preserves the product owner's decision that AFP does not merit a separate project while keeping the feasibility, dependency, adoption, and delivery gates open. It does not authorize converter implementation, library adoption, or delivery.

## Product outcome under consideration

A document-operations reviewer receives a reviewable PDF and explicit diagnostics when the pipeline converts AFP from an owner-agreed source profile. Internal review is a working hypothesis, not a confirmed customer-delivery scope or accepted persona. The owner must confirm the source profile, target output, reviewer role, and intended use before feasibility criteria can be finalized.

The separately maintained [`go-afp-line`](https://github.com/pratik-saptarshi/go-afp-line) library is a candidate dependency used by this pipeline. Keeping AFP in this product does not merge the library's maintenance or release governance into this repository, and does not approve dependency adoption.

## Proposed sequence and priorities

Priority values below are proposals for the AFP items only (`P1` blocks the next decision gate; `P2` follows after its prerequisites). Existing unrelated product priorities remain unchanged. Priority edits and new items are pending supported Beads reconciliation.

| Order | Beads item | Proposed priority | Work and gate | Dependencies / rationale |
| --- | --- | --- | --- | --- |
| 1 | `dap-jo9.4.1` (existing, P2) | P2 | Narrow to the organizational decision: AFP remains in this pipeline and is not a separate project; `go-afp-line` remains separately maintained. Keep open until its acceptance criteria and traceability are separated from feasibility. | Existing child `.4.1.2` and feasibility links currently share this story. Closing it against unchanged `REQ-AFP-001` / `TC-AFP-FEASIBILITY-001` would falsely satisfy feasibility. |
| 2 | Owner profile/reference decision (no ID allocated) | P1 prerequisite | Agree the first useful AFP source profile and PDF target; identify authorized synthetic or deidentified fixtures and trusted reference observations/acceptors. | Required before the existing blocked feasibility assessment can produce a recommendation. Do not infer a profile or persona from repository evidence. |
| 3 | `dap-jo9.4.1.2` (existing, P2) | Proposed P1 | Assess `go-afp-line` against that first useful input profile and the accepted fidelity/error criteria. Return evidence-backed proceed, reject, or unknown; no converter implementation. | Preserve existing `dap-jo9.1.1` blocker and add the owner profile/reference prerequisite. This is the first product feasibility blocker. |
| 4 | Library dependency review (new item; no ID allocated) | Proposed P1 | Separately review the exact library revision, direct/transitive licenses, maintenance/security posture, versions, supported features, integration behavior, and operational fit. Return an adoption recommendation; a zero-exclusion count alone is insufficient. | Starts after the profile is agreed and can proceed in parallel with feasibility. Adoption requires both assessments; keep it distinct from project organization and delivery. |
| 5 | `dap-jo9.4.2` (existing, P2) | Proposed P1 | Owner records go/no-go/unknown or explicit deferral, with input criteria and date. | Retain existing `.4.1` relationship and add blocking dependencies on `.4.1.2` and the dependency-review item. A go/no-go cannot be inferred from the organizational NO. |
| 6 | AFP delivery story (new item; no ID allocated) | Proposed P2 | Refine implementation only after the owner authorizes the outcome and the feasibility, dependency, and go/no-go gates allow it. The acceptance contract below is proposed for that story. | Child of the existing AFP epic; blocked by `.4.2` and the owner's explicit adoption/scope decisions. Do not make AFP a prerequisite for unrelated PDF lanes. |

### Current canonical tracker state

At main `82c903bcb53bde803418064b80933b37f578a52d`, the tracker remains: `dap-jo9.4` P2 open (child of `.9`, blocks `.1`); `.4.1` P2 open (child of `.4`); `.4.1.1` P2 closed (child of `.4.1`); `.4.1.2` P2 open (child of `.4.1`, blocked by `.1.1`); and `.4.2` P2 open (blocked by `.4.1`). The proposed priority changes and dependency additions above are not applied. Unrelated issues are untouched.

PR [#170](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170) proposes closing `.4.1` while leaving its feasibility requirement and case links unchanged. The existing review findings identify that this can falsely satisfy feasibility and unblock `.4.2` ([acceptance-criteria finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170#discussion_r4314287716), [decision-scope finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170#discussion_r4314287730)). This refinement keeps `.4.1` open on canonical main until its scope and traceability are separated; it does not treat the unmerged close as evidence.

The exact dependency edges, priority changes, new Beads IDs, and traceability ownership remain proposals. The issue export is not edited in this record. The canonical export SHA-256 is `dd048534e6e007841e0750d5370a25398178f2b0c70750078eabe1f4b1508409`, outside the reconciler's four exact approved snapshots; its supported operations do not authorize arbitrary dependency-graph changes. Use the guarded reconciliation path after source-specific authorization and independent review are available; do not hand-edit the export.

## Proposed delivery acceptance contract

For an owner-agreed source profile, acceptance should require all of the following without choosing numeric thresholds before the owner/reference decisions:

1. Independently compare extracted text, page order, and page geometry with trusted reference observations. Record the comparison method, reference identity, result, and unresolved differences; do not invent tolerances or claim fidelity from parser round-trip alone.
2. Preflight semantic features against the agreed profile and report supported, unsupported, and not-evaluated features explicitly. An unsupported required feature, malformed input, resource exhaustion, cancellation, encoding error, or unavailable/unsupported required font produces a classified failure and no successful partial result. The owner must decide whether a declared exclusion is permissible for the profile; do not silently treat one as success.
3. Bound input, memory, time, output, and concurrency according to limits approved during engineering/product review. Cancellation and all failure paths leave no output presented as successful; publish complete output atomically only after required checks pass.
4. Preserve the original input. Record input/output identities and hashes, exact library and pipeline versions, profile identity, and processing outcome as provenance.
5. Keep ordinary diagnostics privacy-safe: do not include source text or other document content. Retention, access, and deletion policy remain owner/security decisions.
6. State that the candidate PDF is untagged and is not an accessible or PDF/UA-conformant deliverable. No accessibility claim follows from conversion or visual/text fidelity.
7. Maintain CI positive and negative fixtures for the agreed profile and explicit exclusions. A human reviewer authorized by the owner accepts the comparison against trusted references; fixture tests and tool output do not substitute for that acceptance.

This is a proposed product contract, not proof that the current library meets it. Current acceptance cannot claim fidelity beyond the agreed profile and independently accepted evidence.

## Pinned feasibility evidence and limits

Reviewed source: [`go-afp-line` commit `6af4cb2a416bb9f1a2b966992e8e98dd855c7370`](https://github.com/pratik-saptarshi/go-afp-line/tree/6af4cb2a416bb9f1a2b966992e8e98dd855c7370). At that revision:

- `afptoline/ptoca.go` documents that SVI increments are tracked but not applied and that rules/color are parsed but not emitted ([lines 190–192](https://github.com/pratik-saptarshi/go-afp-line/blob/6af4cb2a416bb9f1a2b966992e8e98dd855c7370/afptoline/ptoca.go#L190-L192)).
- `afptoline/medium.go` resolves every page to fixed US Letter dimensions ([lines 5–23](https://github.com/pratik-saptarshi/go-afp-line/blob/6af4cb2a416bb9f1a2b966992e8e98dd855c7370/afptoline/medium.go#L5-L23)).
- `linetopdf/render.go` labels the output untagged/not accessible and maps the synthetic `FIX-MONO-10` font to Courier ([lines 1–15](https://github.com/pratik-saptarshi/go-afp-line/blob/6af4cb2a416bb9f1a2b966992e8e98dd855c7370/linetopdf/render.go#L1-L15)).

These findings are pinned to that reviewed revision; they do not establish behavior of future revisions, suitability for an unselected profile, or license/security approval. A declared exclusion policy is useful diagnostics, not a sufficient fidelity result.

## Owner decisions still needed

- Which AFP source profile and PDF target define the first useful feasibility input?
- Which fixtures and independently trusted text/order/geometry references are authorized, and who accepts them?
- Which semantic and layout features must be represented, explicitly unsupported, or cause rejection?
- Which declared exclusions, if any, are acceptable within the first profile, and which must block output?
- Who owns profile agreement, dependency adoption, go/no-go, and human reference acceptance?
- Which resource limits, cancellation behavior, diagnostic retention/access rules, and operational boundary are acceptable?
- Is internal document-operations review the intended first delivery use, or is another use/customer scope intended?
- After feasibility and dependency review, should the library be adopted, rejected, or held pending further evidence?

## Paperthin workflow record

This refinement follows the owner's selected Paperthin workflow. The reviewed `autonomous-engineering-platform` commit `7354d9eabfd60e2f6891d34b22745fc5b997b147` pins `LilMGenius/paperthin` at `7d5dc6235990230e45a24a25c527d81701da5458`. Applied `readchk`, `autobahn`, `factchk`, `re0`, and `sip`; the `sip` pass used a blind `shower` read and read-only `ssotize` audit. The user-invoked `re0-plan` was not run. Paperthin supplements repository policy and does not waive product-owner, security, review, or publication gates. See [repository guidance](../../AGENTS.md#paperthin-refinement).
