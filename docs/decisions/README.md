# Decision records

The accepted output-validation target is recorded in [Decision 0001](0001-pdf-ua-2-validation-target.md). [Decision 0002](0002-initial-remediation-family-priority.md) gives statements and invoices with tables equal initial remediation priority. [Decision 0003](0003-operator-confirmed-financial-table-mapping.md) records operator-confirmed mapping for the first bounded workflow and future options discovery. These decisions do not approve a production writer. Add one Markdown record per other material decision, including status (**Proposed**, **Accepted**, or **Superseded**), context, evidence, alternatives, consequences, test evidence, and follow-up Beads issues.

## Accepted decisions

- [Decision 0001 — PDF/UA-2 validation target](0001-pdf-ua-2-validation-target.md).
- [Decision 0002 — Initial remediation family priority](0002-initial-remediation-family-priority.md).
- [Decision 0003 — Operator-confirmed financial-table mapping](0003-operator-confirmed-financial-table-mapping.md).
- [Decision 0004 — Approver-persona review protocol and evidence handling](0004-approver-persona-review-protocol.md).
- [Decision 0006 — Exact Beads source-continuation approval](0006-approved-beads-source-continuation.md).
- [Decision 0007 — Exact Beads null-loss continuation after `dap-jo9.18`](0007-exact-beads-null-loss-continuation-after-jo9-18.md).
- [Decision 0008 — Scoped canonical-export closure for `dap-jo9.18`](0008-scoped-dap-jo9-18-closure.md).
- [Decision 0009 — Exact Beads source continuation after `dap-jo9.18` closure](0009-exact-beads-source-continuation-after-jo9-18-closure.md).
- [Decision 0010 — Bounded PR #112 correction and source continuation](0010-bounded-pr112-correction-and-source-continuation.md).
- [Decision 0011 — Bounded round-five owner decisions](0011-round-five-owner-decisions.md). Bounded R5-01 closure is delegated through `bd` plus guarded validation/promotion; R5-02 remains open, with no generic transition, closure-code, or null-loss policy authorized.
- [Decision 0015 — Approver-role evidence source](0015-approver-role-evidence-source.md). The authoritative source for the approver identity is the owner-approved append-only role-grant register (with full temporal policy: validity intervals, expiry, prospective revocation, retroactive-grant bar, unavailable-source and replaced-evidence handling). Adopts the three-role model (document owner / accessibility reviewer / approver) with dual-hatting recorded, never hidden. Interim posture until the register is implemented: approver recorded by name + timestamp + dual-hat flag in the release record.
- [Decision 0014 — Profile policy decisions](0014-profile-policy-decisions.md). Approves four new document families (`presentations`, `forms`, `marketing`, `hr_documents`); accepts the DRAFT-0010 approver-role evidence source (see Decision 0015); defers `restricted` confidentiality and Spanish language support (see Decision 0016); includes APCA contrast checking as advisory draft-stage; anchors the mandate horizon to April 1, 2027.
- [Decision 0016 — Deferred scope: `restricted` confidentiality class and Spanish](0016-deferred-scope-restricted-spanish.md). The pipeline refuses `restricted`-class documents until the Q5 threat-data controls are implemented and independently reviewed; Spanish support is deferred while the profile contract stays English-only.

## Drafts pending repository adoption

- [Historical v2 bounded association-repair Proceed](0005-historical-v2-bounded-association-proceed.md) records the accepted narrow historical disposition and exact proposed `.1.6.3`/`.1.6.4` updates. The linked [acceptance register](../reviews/pdfua2-bounded-proceed-acceptance-register-2026-10-15.md) preserves review excerpts and source hashes. Tracker states remain unchanged because the bd 1.3.1 export round-trip synthesized a non-factual `.5.5` `closed_at` and normalized unrelated fields; see Decision 0005 for provenance and limits.

## Proposed decision topics

- Data, accessibility, privacy, retention, and security profiles.
- Existing-PDF writer/fidelity and remediation worker boundaries.
- Open-source tool and license verification; enterprise-only remediation exports remain excluded.
- Generation PDF engine and DOCX/Markdown conversion approach.
- Semantic interpretation policy, human semantic review authority, and assistive-technology protocol.
- IBM AFP migration feasibility and go/no-go criteria.

Until a record is reviewed and accepted, roadmap candidates remain proposed. The accepted PDF/UA-2 target does not imply that any implementation or output has passed it.
