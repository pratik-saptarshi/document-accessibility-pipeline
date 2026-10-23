# AFP refinement and prioritization — 2026-10-19

**Status: Planning refinement promoted to the PR #171 branch; product and delivery gates remain open.** This record preserves the product owner's decision that AFP does not merit a separate project while keeping the feasibility, dependency, adoption, and delivery gates open. The reviewed bounded Beads candidate was promoted to this branch's canonical export through the guarded promotion command. This does not authorize converter implementation, library adoption, or delivery.

## Product outcome under consideration

A document-operations reviewer receives a reviewable PDF and explicit diagnostics when the pipeline converts AFP from an owner-agreed source profile. Internal review is a working hypothesis, not a confirmed customer-delivery scope or accepted persona. The owner must confirm the source profile, target output, reviewer role, and intended use before feasibility criteria can be finalized.

The separately maintained [`go-afp-line`](https://github.com/pratik-saptarshi/go-afp-line) library is a candidate dependency intended for use by this pipeline. Keeping AFP in this product does not merge the library's maintenance or release governance into this repository, and does not approve dependency adoption.

## Proposed sequence and priorities

Priority values below are proposals for the AFP items only (`P1` blocks the next decision gate; `P2` follows after its prerequisites). Existing unrelated product priorities remain unchanged. Priority edits and new items are pending supported Beads reconciliation.

| Order | Beads item | Proposed priority | Work and gate | Dependencies / rationale |
| --- | --- | --- | --- | --- |
| 0 | `dap-jo9.4` (existing, P2) | P2 unchanged | Keep this AFP workstream within the pipeline and the library separately maintained. Update only its description and acceptance criteria to reflect profile agreement, feasibility, dependency review, and an explicit owner decision; preserve its identity, status, priority, parent, and current `.1` prerequisite. | Its current edge is `dap-jo9.4` depends on `dap-jo9.1`. AFP does not block unrelated PDF lanes. |
| 1 | `dap-jo9.4.1` (existing, P2) | P2 unchanged | Record the organizational decision that AFP remains in this pipeline, is not a separate project, and `go-afp-line` remains separately maintained. Keep open. Update title, description, and notes; preserve the current feasibility acceptance criteria unchanged in this batch. | It remains parent of `.4.1.1` and `.4.1.2`, and currently blocks `.4.2`. Do not close it or reassign/remove feasibility traceability until that reassignment is separately reviewed. |
| 2 | Owner profile/reference decision (new child; ID allocated by Beads) | P1 proposed | Agree the first useful AFP source profile and PDF target; identify authorized fixtures, trusted text/order/geometry references, feature/error boundaries, and who accepts the reference comparison. | This is an explicit prerequisite for both feasibility and dependency review. Do not infer a profile, sample count, thresholds, persona, or acceptor from repository evidence. |
| 3 | `dap-jo9.4.1.2` (existing, P2) | P1 proposed | Assess `go-afp-line` against the agreed profile and accepted fidelity/error criteria. Return evidence-backed proceed, reject, or unknown; no converter implementation. | Preserve its parent `.4.1` and existing dependency on `dap-jo9.1.1`; add dependency on the profile/reference decision. First product feasibility blocker. |
| 4 | Library dependency review (new child; ID allocated by Beads) | P1 proposed | Separately review the exact library revision, direct/transitive licenses, maintenance/security posture, supported features, integration behavior, and operational fit. Return an adoption recommendation; a zero-exclusion count alone is insufficient. | Depends on the profile/reference decision and can proceed in parallel with feasibility after that gate. Adoption requires both assessments. |
| 5 | `dap-jo9.4.2` (existing, P2) | P1 proposed | Owner records go/no-go/unknown or explicit deferral after reviewing feasibility and dependency evidence; a go must include an explicit adoption decision, input criteria, owner, and date. | Preserve its current dependency on `.4.1`; add dependencies on `.4.1.2` and the dependency-review item. Do not infer a go/no-go from the organizational NO. |
| 6 | AFP delivery implementation story (new child; ID allocated by Beads) | P2 proposed | Implement and validate delivery only after the owner authorizes the outcome, adoption, audience/scope, and feasibility evidence. Its completion requires executed evidence under the contract below. This refinement task does not authorize implementation. | Child of `.4`; depends on `.4.2` and requires the explicit go/adoption and scope decision. Do not make AFP a prerequisite for unrelated PDF lanes. |

### Current canonical tracker state

At main `82c903bcb53bde803418064b80933b37f578a52d`, the tracker remains: `dap-jo9.4` P2 open (child of `dap-jo9`; depends on `dap-jo9.1`); `.4.1` P2 open (child of `.4`); `.4.1.1` P2 closed (child of `.4.1`); `.4.1.2` P2 open (child of `.4.1`, depends on `dap-jo9.1.1`); and `.4.2` P2 open (depends on `.4.1`). The proposed priority changes and dependency additions above are not applied. Unrelated issues are untouched.

PR [#170](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170) proposes closing `.4.1` while leaving its feasibility requirement and case links unchanged. The existing review findings identify that this can falsely satisfy feasibility and unblock `.4.2` ([acceptance-criteria finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170#discussion_r4314287716), [decision-scope finding](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/170#discussion_r4314287730)). This refinement keeps `.4.1` open on canonical main until its scope and traceability are separated; it does not treat the unmerged close as evidence.

PR [#171](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/171) also adds `docs/decisions/0018-afp-not-a-separate-project.md`, with content that differs from PR #170's version at the same path. Reconcile or supersede the competing file versions; do not merge both PRs independently or rely on conflict resolution to choose the intended decision. PR #170 currently has auto-merge enabled, and this refinement leaves that setting untouched. If it merges first, PR #171 must be updated against current main to replace the decision file with the reviewed final version and must not merge until that update is independently reviewed. If PR #170 closes `.4.1`, that remains a disputed closure against unchanged feasibility acceptance and does not provide feasibility evidence. Because current `.4.2` depends only on `.4.1`, Beads may report `.4.2` unblocked after that close; that tracker status is not accepted as feasibility readiness. Do not treat a Decision 0018 document update as repairing this tracker state. Reassess the new canonical source digest and `.4.1` status, then use a separately authorized, null-preserving guarded Beads correction before claiming the feasibility gate is available. If PR #170 remains open, reconcile both proposed decision-file versions before either PR is treated as the final decision record.

The exact dependency edges, priority changes, and generated Beads IDs are now in the PR branch's canonical export. Traceability ownership remains an open decision. The initial candidate was derived from source snapshot SHA-256 `dd048534e6e007841e0750d5370a25398178f2b0c70750078eabe1f4b1508409`; its original promoted export SHA-256 was `9e7efe1632810cc1d89e533a9656c6c178b277834a222fbd81c04b6f1903f033`, superseded by the exact-source correction documented below. A narrowly source-bound AFP operation supports this exact plan and pinned Beads 1.3.1 binary; its reviewed evidence is recorded below. It rejects unplanned field, priority, edge, lifecycle, and counter changes. The export update remains in the unmerged PR branch.

## Bounded Beads change proposal

This is the exact batch reviewed and applied through guarded promotion. Preserve `.4`'s identity, title, open status, P2 priority, parent edge, and current dependency on `.1`; change only its description and acceptance criteria below. Preserve `.4.1.1` as closed and every unrelated issue. Keep `.4.1` open and preserve its feasibility acceptance criteria and existing edges while the organizational decision is recorded in its title, description, and notes. No issue is closed or reopened.

### Exact existing-issue field changes

Fields not listed remain unchanged. Existing `status`, parent edges, and current prerequisites remain unchanged unless an edge is explicitly added below. Exact field values:

```text
dap-jo9.4.description = "Evaluate the first owner-agreed AFP source profile as an in-product feasibility workstream in this pipeline, using the separately maintained go-afp-line candidate library intended for this pipeline. This work does not approve dependency adoption or converter implementation."
dap-jo9.4.acceptance_criteria = "The workstream records (1) an owner-agreed AFP source profile, PDF target, authorized fixtures, trusted reference observations, and acceptor; (2) a bounded feasibility assessment of the pinned candidate library against that profile, including independently checked text, page order, and geometry; (3) a separate dependency review; and (4) an explicit owner go/no-go/unknown or deferral before delivery refinement. Do not infer fidelity from parser round-trip, treat a zero-exclusion count as sufficient, authorize implementation, or make AFP a prerequisite for unrelated PDF lanes."

dap-jo9.4.1.title = "Record AFP organizational decision and preserve feasibility traceability"
dap-jo9.4.1.description = "The owner decided AFP remains within this pipeline and is not a separate project. go-afp-line remains separately maintained and is only a candidate dependency intended for this pipeline; adoption is unapproved. This issue records the organizational boundary. Its existing feasibility acceptance criteria and case links remain in place until a separately reviewed traceability reassignment; this issue stays open and does not authorize implementation."
dap-jo9.4.1.acceptance_criteria = "CAP-AFP-FEASIBILITY. Canonical catalog requirements: REQ-AFP-001. Acceptance test specification: TC-AFP-FEASIBILITY-001. This story is a decision outcome only; it does not authorize AFP conversion implementation."
dap-jo9.4.1.notes = "\n\n2026-10-14 panel (deferred): link exact catalog rows (docs/requirements/catalog.md) instead of bare CAP-/REQ-/TC- IDs in story AC.\n\nDecision 0018 records that AFP remains in this pipeline, is not a separate project, and go-afp-line remains separately maintained. This organizational decision does not satisfy or close the existing feasibility acceptance criteria. Preserve feasibility traceability until its separately reviewed reassignment."

dap-jo9.4.1.2.title = "Assess go-afp-line against the owner-agreed AFP profile"
dap-jo9.4.1.2.description = "## Goal\n\nAssess the exact go-afp-line revision against the source profile, PDF target, fixtures, trusted references, and feature/error boundaries recorded by the owner profile/reference decision. Return evidence-backed proceed, reject, or unknown. Do not implement a converter or adopt the library.\n\n## Evidence\n\nIndependently compare extracted text, page order, and page geometry with the authorized trusted references. Preflight semantic features and classify unsupported, malformed, resource, cancellation, encoding, and font failures. Record bounded-resource and cancellation behavior, no-successful-partial-output behavior, original preservation, input/output/version provenance, privacy-safe diagnostics, and the untagged/non-accessibility boundary. Use CI positive and negative fixtures and human acceptance against references. Do not invent numeric tolerances, budgets, fixture counts, or accepted personas.\n\n## Test cases\n\nTC-AFP-FEASIBILITY-001\n\n## Traceability\n\nTC-AFP-FEASIBILITY-001 remains a specification, not executed evidence. Preserve the existing parent and dap-jo9.1.1 prerequisite; this assessment does not close the profile decision or authorize either PDF lane.\n\n## Case participation and evidence boundary\n\n- `TC-AFP-FEASIBILITY-001` — implements; canonical owner `dap-jo9.4.1.1`; scope: bounded AFP feasibility assessment and recommendation only; no conversion implementation.\n\nThese are task participation links only; the canonical case owner remains authoritative and no case execution is implied."
dap-jo9.4.1.2.acceptance_criteria = "The owner profile/reference decision and dap-jo9.1.1 prerequisite are satisfied. Record the exact candidate revision and evidence for the agreed profile: independent text/order/geometry comparison to trusted references; semantic preflight; classified unsupported, malformed, resource, cancellation, encoding, and font failures; approved resource bounds; no successful partial output; original preservation; input/output/version provenance; privacy-safe diagnostics; and explicit untagged/non-accessibility limits. Include positive and negative CI fixture results and human reference acceptance. Return proceed, reject, or unknown with unresolved gaps. TC-AFP-FEASIBILITY-001 remains a specification link, not executed evidence. No converter implementation, inferred thresholds, or closure of the profile/organizational issues."

dap-jo9.4.2.description = "## Goal\n\nRecord the owner's AFP go/no-go/unknown decision or explicit deferral only after the agreed profile, feasibility assessment, and separately maintained library dependency review have produced their evidence. A go requires an explicit adoption decision and authorized delivery scope.\n\n## Context\n\nAFP remains in this pipeline and is not a separate project; this organizational decision does not determine feasibility, dependency adoption, or delivery."
dap-jo9.4.2.acceptance_criteria = "1. Record go/no-go/unknown, the named decision owner, input profile and criteria, evidence reviewed, and date. A GO explicitly approves adoption of the exact separately maintained go-afp-line revision and names the authorized next scope. A NO GO records non-adoption and rationale. UNKNOWN or deferral records unresolved evidence and the trigger that re-opens the decision. Require completed dap-jo9.4.1.2 and dependency-review evidence; this decision does not itself implement conversion."
```

Proposed priority changes are `.4.1.2: P2→P1` and `.4.2: P2→P1`. Existing `.4` and `.4.1` remain P2. The P1 assessments block the owner decision; delivery is P2 because it follows a GO/adoption and scope decision.

### Complete templates for the three new issues

Beads allocates the three IDs; these names are role labels for the proposed edges, not guessed IDs. Values below are the complete caller-supplied issue fields. Beads 1.3.1 generates the ID and created/updated metadata, dependency timestamps/creator metadata, and derived counters during the disposable-store run. No caller-supplied field is `null`.

**Owner profile/reference decision (new direct child of `dap-jo9.4`):**

```yaml
title: Agree first AFP source profile and trusted references
description: |
  Record the owner-approved first useful AFP input profile and PDF target
  before feasibility or dependency review starts. Identify the profile's
  source characteristics, authorized synthetic or deidentified fixtures,
  trusted expected text/page-order/geometry observations, relevant semantic
  features and failure boundaries, and the people authorized to accept the
  comparison. Do not invent a customer, persona, sample count, tolerance,
  resource budget, or delivery scope.
acceptance_criteria: |
  Record the named owner and date; AFP source profile and PDF target; authorized
  fixture identities and handling constraints; trusted reference identities
  and how text, page order, and geometry will be checked; semantic feature and
  unsupported/error boundaries; and the authorized reference acceptor. Mark
  every unresolved item explicitly. A feasibility recommendation cannot pass
  this gate while required profile/reference inputs remain unresolved. This
  decision does not approve dependency adoption or implementation.
notes: ""
status: open
priority: P1
issue_type: decision
labels: [proposed, lane-afp, phase-na]
parent: dap-jo9.4
dependencies: []
```

**Library dependency review (new direct child of `dap-jo9.4`):**

```yaml
title: Review go-afp-line dependency for the agreed AFP profile
description: |
  Review the separately maintained go-afp-line revision proposed for use by
  this pipeline, against the owner-approved AFP profile. This is dependency
  due diligence and an adoption recommendation, not an approval or integration.
acceptance_criteria: |
  The owner profile/reference decision is complete. Record the exact source
  revision; direct and transitive license evidence; maintenance and security
  posture; release/versioning evidence; supported and unsupported profile
  features; integration behavior; operational fit; and material unknowns.
  Return adopt, reject, or defer with evidence and explicit limitations. A
  zero-exclusion count alone is not a fidelity or adoption result. Do not add
  the dependency to the product or implement conversion in this issue.
notes: ""
status: open
priority: P1
issue_type: task
labels: [proposed, lane-afp, phase-na]
parent: dap-jo9.4
dependencies:
  - depends_on: <owner profile/reference decision ID allocated by Beads>
```

**Future delivery implementation story (new direct child of `dap-jo9.4`):**

```yaml
title: Convert supported AFP documents into reviewable PDFs
description: |
  Implement an owner-authorized workflow that converts AFP from the agreed
  source profile into a reviewable PDF with explicit diagnostics. Internal
  document-operations review remains a hypothesis until the owner confirms
  audience and scope. This is a future implementation story, currently blocked
  pending feasibility, dependency review, and explicit owner GO/adoption/scope
  authorization. This refinement task does not authorize implementation.
acceptance_criteria: |
  This is implementation acceptance, not completion of the current planning or
  refinement task. Do not begin implementation until the owner confirms profile,
  audience, and delivery scope; the profile/reference decision is complete;
  feasibility recommends proceed; dependency review recommends adopting the
  exact library revision; and dap-jo9.4.2 records GO and adoption approval.
  Completion requires the implemented behavior and executed evidence below:
  1. Independently compare output text, page order, and page geometry with the
     authorized trusted references; record method, reference identity, result,
     and differences without inventing tolerances.
  2. Implement semantic-feature preflight that reports supported, unsupported,
     and not-evaluated features. Required unsupported features, malformed input,
     resource exhaustion, cancellation, encoding errors, or unavailable
     required fonts produce classified failures and no successful partial
     result. The owner decides whether any declared exclusion is permissible.
  3. Enforce approved bounds for input, memory, time, output, and concurrency.
     Executed cancellation/failure checks show no output is presented as
     successful; complete output is published only after required checks pass.
  4. Preserve the original. Record input/output identities and hashes, exact
     library and pipeline versions, profile identity, and processing outcome.
  5. Keep ordinary diagnostics privacy-safe; do not include source text or
     document content. Retention, access, and deletion follow owner/security
     decisions.
  6. State that the output PDF is untagged and is not an accessible or
     PDF/UA-conformant deliverable. Make no accessibility claim.
  7. Run CI positive and negative fixtures for the agreed profile and explicit
     exclusions, record their results, and obtain human acceptance from the
     authorized reviewer against trusted references. Tests do not replace
     human acceptance.
  Do not assign requirement or case IDs until traceability ownership and scope
  are approved. A plan, template, or unexecuted test specification does not
  satisfy this story.
notes: ""
status: open
priority: P2
issue_type: story
labels: [proposed, user-story, lane-afp, phase-na, decision-blocked]
parent: dap-jo9.4
dependencies:
  - depends_on: dap-jo9.4.2
```

### Exact dependency-edge delta

The canonical JSONL uses one dependency row for each edge; the dependent issue is `issue_id` and the prerequisite is `depends_on_id`. Preserve current edges and add only these:

| `issue_id` (dependent) | `depends_on_id` (prerequisite) | Edge type | Status |
| --- | --- | --- | --- |
| `<profile-ID>` | `dap-jo9.4` | `parent-child` | New profile decision is a direct child of `.4`. |
| `dap-jo9.4.1.2` | `<profile-ID>` | `blocks` | New prerequisite for feasibility; preserve `.4.1` parent and `.1.1` prerequisite. |
| `<dependency-review-ID>` | `dap-jo9.4` | `parent-child` | New dependency review is a direct child of `.4`. |
| `<dependency-review-ID>` | `<profile-ID>` | `blocks` | New prerequisite for dependency review. |
| `dap-jo9.4.2` | `dap-jo9.4.1.2` | `blocks` | New feasibility evidence prerequisite. |
| `dap-jo9.4.2` | `<dependency-review-ID>` | `blocks` | New dependency evidence prerequisite. |
| `<delivery-ID>` | `dap-jo9.4` | `parent-child` | New delivery story is a direct child of `.4`. |
| `<delivery-ID>` | `dap-jo9.4.2` | `blocks` | New delivery gate. |

Preserve the current edges exactly: `dap-jo9.4` is a child of `dap-jo9` and depends on `dap-jo9.1`; `.4.1` is a child of `.4` and blocks `.4.2`; `.4.1.1` is a child of `.4.1`; `.4.1.2` is a child of `.4.1` and depends on `.1.1`; `.4.2` depends on `.4.1`. The new symbolic IDs must be resolved by the pinned disposable-store run and shown in its candidate bundle before any promotion approval; do not invent IDs in advance.

### Derived counter effects and candidate assertions

The pinned Beads 1.3.1 export semantics were checked in a fresh embedded-Dolt disposable store imported from this exact source; the executable SHA-256 was `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`. Three temporary children were auto-allocated, the proposed parent and blocking edges were added, and before/after exports were compared. The probe changed only the following existing issue counters:

| Existing issue | Counter | Source value | Expected candidate value | Cause |
| --- | --- | ---: | ---: | --- |
| `dap-jo9.4.1.2` | `dependency_count` | 1 | 2 | Adds the profile/reference prerequisite. |
| `dap-jo9.4.1.2` | `dependent_count` | 0 | 1 | `.4.2` adds a dependency on this feasibility issue. |
| `dap-jo9.4.2` | `dependency_count` | 1 | 3 | Adds dependencies on `.4.1.2` and the dependency-review issue. |
| `dap-jo9.4.2` | `dependent_count` | 0 | 1 | The new delivery story depends on `.4.2`. |

No other existing issue's `dependency_count` or `dependent_count` may change. Beads-allocated new rows should export these derived counters from the verified graph: profile/reference decision `(dependency_count=0, dependent_count=2)`; dependency review `(1,1)`; delivery story `(1,0)`. These are assertions about derived output, not caller-controlled template fields. In the disposable-run bundle and candidate review, assert each listed edge occurs exactly once, all four existing counter deltas match this table, every other existing counter equals its source value, the three new-row counters match the values above, and the candidate and second-store round-trip have identical graph/counters. The current Beads 1.3.1 import/export probe omits the source's explicit null from baseline, updated, and second-store round-trip exports; therefore the candidate must be derived from canonical source bytes and retain that exact null token, while evidence records this representation loss and the candidate byte span. The null-loss exception is narrowly authorized in the source-bound candidate path for this exact source and Beads binary; it does not authorize canonical mutation. Do not assert that Beads' round-trip export itself retains the null. Reject any additional counter change or null loss, even if the added issue fields or edges otherwise look valid.

### Source, null, and publication boundary

The reviewed source snapshot `.beads/issues.jsonl` had SHA-256 `dd048534e6e007841e0750d5370a25398178f2b0c70750078eabe1f4b1508409`. It contains one explicit JSON null: `dap-jo9.5.3.dependencies`. The pinned importer/exporter omits this field after import, so the exact-source reconciler exception uses source-byte splicing and records the representation loss; that exception alone does not authorize canonical mutation. The owner-authorized candidate promotion preserves the value byte-for-byte; it is not normalized to `[]`, omitted, or otherwise altered. The promoted export preserves every other unplanned record and field byte-for-byte, except the documented Beads-generated `updated_at` for updated issues, the exact existing-record counter changes in the table above, and generated metadata/counters on the three new rows and edges. No null-valued fields were added to new issues; the templates specify empty `notes` as `""`. A candidate that loses or changes the explicit null or changes any unlisted existing counter is rejected.

The code now implements a narrow source-bound candidate path. It accepts only the exact source digest, approved AFP field values and priority changes, three issue templates, eight added edges, the Beads 1.3.1 executable digest, the one explicit-null loss, and two timestamp normalization pairs observed in this source. A pinned disposable-store run generated `dap-jo9.4.3` (profile decision, P1), `dap-jo9.4.4` (dependency review, P1), and `dap-jo9.4.5` (delivery story, P2). Initial candidate and promoted canonical SHA-256: `9e7efe1632810cc1d89e533a9656c6c178b277834a222fbd81c04b6f1903f033`; initial evidence ZIP SHA-256: `de9b14f911947320e82a08e6d003eab21c704d8b0c526c469e19c7fc4b354d2a`. This initial candidate was superseded by the exact-source correction below. Evidence confirms the specified graph and counter deltas, preserves the source null token, and records Beads timestamp normalization for `.1.6.10` and `.19`. Each new edge and created row timestamp must use whole-second precision and fall within the nearest-second-rounded start/finish window recorded for the exact pinned Beads command that emitted it; existing dependency records retain their full metadata. The guarded promotion rederived this candidate, validated its second-store export, and replaced the canonical file only after checking the source digest and file identity. The promoted change is on the unmerged PR #171 branch; no converter implementation, library adoption, or delivery is authorized.

2026-10-20 exact-source correction: after the root race suite exposed that `.4.1.2` retained only the case ID and had lost its reviewed participation role/owner/scope link, the parent authorized regeneration from the same original source snapshot and pinned Beads binary. The correction restores exactly `TC-AFP-FEASIBILITY-001` as an implements link owned by `dap-jo9.4.1.1`, scoped to bounded feasibility assessment and recommendation only, with no conversion implementation. Candidate SHA-256: `968a1776acb735447f18b013538843d77b30b2db7661d51f323539ef4dbdfa6a`; evidence ZIP SHA-256: `26a888ca0997c1ee8091840944bfea5c485722240900ed9a8a7f75382eb7303c`; source remains `dd048534e6e007841e0750d5370a25398178f2b0c70750078eabe1f4b1508409`. The guarded promoter validated this exact bundle and replaced the local canonical file. Full root race tests, vet, tidy-diff, and traceability passed against this canonical candidate. No alert was suppressed or bypassed, and the PR remains unmerged pending hosted verification and final review.

## Proposed delivery acceptance contract

This proposed contract is for the future delivery implementation story, not the current planning/refinement task. Completion requires implemented behavior and executed evidence after the owner/profile, feasibility, dependency, adoption, and scope gates pass. It must require the following without choosing numeric thresholds before the owner/reference decisions:

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
- `linetopdf/render.go` identifies the output as untagged and maps the synthetic `FIX-MONO-10` font to Courier ([lines 1–15](https://github.com/pratik-saptarshi/go-afp-line/blob/6af4cb2a416bb9f1a2b966992e8e98dd855c7370/linetopdf/render.go#L1-L15)). The non-accessibility boundary here is a product requirement, not a claim about literal wording in those source lines.

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

## Standing rule: AFP dependency license re-read (added 2026-10-22)

This is a new standing rule established by the
[conditions 2–5 status record](../reviews/go-afp-line-conditions-2-5-status.md),
extending Decision 0022's condition 5 (0022 required only per-bump re-reads).
Re-read licenses on every go-afp-line version bump **and** on any `go.mod`
graph change touching the AFP dependency subtree, before the pin moves. The
subtree is `{go-afp-line, the pdfcpu subtree, DAP's direct golang.org/x/sys
imports}`: x/sys does not transit via go-afp-line at all — it enters DAP's
graph through pdfcpu's transitive subtree and through DAP's own direct
require (`golang.org/x/sys v0.49.0` in the direct require block of `go.mod`,
imported by `ci/pdfua2/adapter/process_linux.go`,
`cmd/beads-export-reconcile/process_linux.go`,
`cmd/verapdf-container-runner/process_linux.go`, and
`internal/financialfixtures/process_linux.go`).

## Paperthin workflow record

This refinement follows the owner's selected Paperthin workflow. The reviewed `autonomous-engineering-platform` commit `7354d9eabfd60e2f6891d34b22745fc5b997b147` pins `LilMGenius/paperthin` at `7d5dc6235990230e45a24a25c527d81701da5458`. Applied `readchk`, `autobahn`, `factchk`, `re0`, and `sip`; the `sip` pass used a blind `shower` read and read-only `ssotize` audit. The user-invoked `re0-plan` was not run. Paperthin supplements repository policy and does not waive product-owner, security, review, or publication gates. See [repository guidance](../../AGENTS.md#paperthin-refinement).
