# Decision 0014: Profile policy decisions

**Status: Accepted.** Supersedes and replaces DRAFT-0014 with the owner's
answers below. Where a DRAFT-0014 recommendation differs, the owner's answer
here governs.

**Date: 2026-10-18.**

## Context

Primary product goal (owner, 2026-10-18): help SMB clients **and** enterprises
prepare for WCAG document-accessibility requirements and for **WCAG 3.0**.
Secondary goal: tooling readiness for upcoming federal and state
digital/web/document accessibility requirements, with an **April 1, 2027
internal readiness target** selected for planning. This is not a legal
compliance date. DOJ's Title II web and mobile-app rule applies to state and
local public entities, including special districts; the DOJ fact sheet states
April 26, 2027 for public entities with a population of 50,000 or more and
April 26, 2028 for public entities below 50,000 and special district
governments. It is not a blanket deadline for SMBs or private enterprises.
The rule covers applicable web content and mobile apps and specifies WCAG 2.1
Level AA.

PDF/UA-2 is the accepted output target (Decision 0001). This record answers
the five open profile-policy questions in `docs/profiles/README.md` so the
`internal/profile` contract can carry real values instead of placeholders.

WCAG 2.2 SC 1.4.3 specifies contrast ratios of at least 4.5:1 for normal text
and 3:1 for large text, subject to its stated exceptions. WCAG 2.2 SC 1.4.6
sets separate AAA thresholds of 7:1 for normal text and 4.5:1 for large text.
SC 1.4.11 is a separate non-text contrast criterion. WCAG 3.0 remains a W3C
draft; its text-contrast requirement says the contrast algorithm has yet to
be determined. See the [DOJ Title II rule fact sheet](https://www.ada.gov/resources/2024-03-08-web-rule/),
[WCAG 2.2 SC 1.4.3](https://www.w3.org/TR/WCAG22/#contrast-minimum),
[WCAG 2.2 SC 1.4.6](https://www.w3.org/TR/WCAG22/#contrast-enhanced),
[WCAG 2.2 SC 1.4.11](https://www.w3.org/TR/WCAG22/#non-text-contrast),
[WCAG 2.2 contrast-ratio and relative-luminance definitions](https://www.w3.org/TR/WCAG22/#dfn-contrast-ratio), and
[WCAG 3.0 text contrast](https://www.w3.org/TR/wcag-3.0/#text-contrast-sufficient-minimum).

## Decisions

1. **Four new document families: APPROVED — `presentations`, `forms`,
   `marketing`, `hr_documents`.** The pipeline handles a variety of PDF and
   Word documents — financial and non-financial, with tables, graphics, and
   other content — and the existing `statements`, `invoices`,
   `correspondence`, `reports` family set cannot describe an SMB Word
   newsletter or an enterprise slide deck. The schema change to
   `internal/profile/profile.go` is implemented alongside this record
   (branch `feat/profile-policy-decisions-0014-0016`); this record
   approves the families themselves. The structural principle is retained:
   source fields are structural, not domain-semantic — the pipeline tags
   structure (headings, tables, figures, form fields, reading order), it does
   not interpret content. A field is required only if the tagging pipeline
   consumes it. Domain meaning (what a dollar amount *means*, what a
   marketing claim *says*) is human-review territory, never a source field.
   The following are descriptive proposed profile fields, not new required
   `source_fields` or proof of production handling:
   - `presentations`: `slide_count`, `slide_titles`, `speaker_notes_included`,
     `embedded_media` (transcript/caption required), `charts_per_slide`.
     Accessibility hinges on per-slide reading order, alt-text for visuals,
     and not stranding content in the slide master.
   - `forms`: `form_fields` (each with required `label`, `field_type`,
     `required`), `instructions`, `sections`. The critical invariant: every
     field has an explicit label — unlabeled fields are the most common form
     accessibility failure.
   - `marketing`: `headings`, `images` (alt-text required; decorative images
     explicitly marked), `links` (descriptive text — not "click here" — plus
     target), `language`. Covers brochures, flyers, and the SMB Word
     newsletter. Link-text quality and the decorative-vs-informative image
     distinction are the two things the pipeline must capture.
   - `hr_documents`: `title`, `effective_date`, `sections`, `tables`
     (benefits comparisons are table-heavy), `signature_fields`.
   - Existing families keep their draft-proposed fields: `statements`
     (`period_start`, `period_end`, `account_holder`, `transactions` table),
     `invoices` (`invoice_number`, `invoice_date`, `bill_to`, `line_items`
     table), `correspondence` (`title`, `date`, `sender`, `recipient`,
     `body`, `signature_block`, `language`), `reports` (`title`,
     `report_date`, `author`, `sections`, `tables`, `figures` with required
     alt-text, `footnotes`).

2. **DRAFT-0010 (approver-role evidence source and temporal policy):
   ACCEPTED as a release-stage policy decision.** Human review authority
   follows the three-role model with SMB collapse: the **document
   owner** declares the profile (family, language, confidentiality); the
   **accessibility reviewer** runs the human-judgment checklist; the
   **approver** signs off, with the approver's identity resolved from the
   owner-approved evidence source DRAFT-0010 selects. In SMB reality the
   reviewer is often the owner wearing a second hat — that must be *recorded*
   as a dual-hat review, not hidden. Full details are in the companion
   decision record
   [0015-approver-role-evidence-source.md](0015-approver-role-evidence-source.md).
   This selection does not name a person, implement a trusted verifier, or
   block functional profile, schema, or contrast work. Genuine authorized
   semantic/AT evidence remains a release requirement; synthetic identities
   are test fixtures only and cannot count as release approval.

3. **`restricted` confidentiality class: DEFERRED.** `restricted`-class
   documents (regulated or sensitive data) are out of scope until the Q5
   threat-data controls are implemented and independently reviewed. This is
   a scope decision, not evidence that profile validation currently rejects
   the class. Retention choices remain independently open.
   Deferred-scope details are in
   [0016-deferred-scope-restricted-spanish.md](0016-deferred-scope-restricted-spanish.md).

4. **Contrast calculation: WCAG 2.2 relative-luminance ratio.** The
   calculation reports a numeric ratio for two rendered sRGB colors. Its
   caller supplies the applicable criterion and context; the calculation
   does not infer text size, classify content, apply criterion exceptions,
   or establish WCAG conformance. SC 1.4.3 uses 4.5:1 for normal text and
   3:1 for large text. SC 1.4.6 has distinct AAA thresholds of 7:1 and
   4.5:1. Non-text contrast is separate under SC 1.4.11. A ratio result
   alone is not a whole-document conformance result.

   **Supersession record:** The initial 2026-10-18 text selected APCA as an
   advisory path and incorrectly described WCAG 3.0 as selecting it. The
   owner's subsequent 2026-10-18 direction replaces that choice with the
   WCAG 2.2 contrast-ratio calculation above. APCA is not part of WCAG 2.2,
   is not selected by the WCAG 3.0 draft, and has been removed from current
   implementation and policy. This correction makes no APCA
   license/provenance clearance or waiver; that question remains unresolved
   and is not a current feature requirement.

5. **Internal readiness target: April 1, 2027.** This is a planning target,
   not a regulatory compliance deadline. Retention policy is separate and
   remains open. No universal seven-year legal retention floor is selected;
   any period must be based on applicable record type and jurisdiction. Do
   not extend, delete, or retain evidence solely because of this target.

6. **Spanish language support: DEFERRED.** Spanish is not claimed as
   supported. The generic profile contract does not enforce an English-only
   language allowlist. Language is per-document and should match the PDF's
   `/Lang` entry when that behavior is implemented. Deferred-scope details are in
   [0016-deferred-scope-restricted-spanish.md](0016-deferred-scope-restricted-spanish.md).

## Rejected options

- **Q1 Option B — Rich domain model per family:** rejected; couples the
  accessibility contract to domain parsing the pipeline will never do.
- **Q1 Option C — Fully open (every field unknown/open):** rejected; gives
  clients nothing to implement against before the mandate horizon.
- **Q2 Option B — Minimal (en + internal + transient only):** rejected;
  left Spanish-speaking SMB clients and any evidence-retention story
  unaddressed.
- **Q2 Option C — Full enterprise taxonomy:** rejected; premature for the
  current client base and impossible to evidence by the mandate horizon.
- **Q4 Option B — Single signer:** rejected; provides no segregation signal
  for enterprise clients and weakens the mandate-readiness evidence story.
- **Q4 Option C — External auditor required:** rejected; commercially
  infeasible for SMB.
- **Using APCA as a current contrast feature:** superseded by the owner's
  subsequent direction to use WCAG 2.2 contrast ratios. APCA is not a WCAG
  2.2 criterion and the WCAG 3.0 draft has not selected an algorithm.
- **Accepting `restricted` documents before an independent security review of
  the Q5 controls:** rejected; deferred until the controls are evidenced.

## Q5 threat-data handling, access, deletion, audit

Boundary (unchanged): the repo mandates synthetic/deidentified fixtures only.
Real client documents must not enter the pipeline until these rules are
implemented and independently reviewed: least-privilege access for named
individuals only (owner, reviewer, approver), with logged grants;
`restricted`-class documents processed only where the synthetic-fixture
corpus is provably absent from the data path; cryptographic erasure
(key destruction or NIST 800-88 Clear/Purge) within 24 hours of run
completion for `transient` retention, with a deletion record; an append-only
audit log of intake hashes, access grants, transformation runs, review
events, approvals, and deletions — hashes and metadata only, no document
content; and P0 incident response (quarantine, owner notification, log
entry) for any over-retention or unauthorized access. Storage backend,
encryption scheme, and key management are architecture decisions for a later
record.

## Consequences and follow-up

- Go profile validation and both lane schemas recognize the same eight
  approved family identifiers. This family recognition, and the descriptive
  proposed profile fields above, do not imply production processing or
  family-specific validation/handling. The accepted `Decision.Value` fields
  use proposed/open statuses; `pdfua_target` remains independently accepted
  as PDF/UA-2.
- SMB vs enterprise is a deployment profile, not two products: the same
  contract serves both; only the selected values differ. Do not fork the
  schema.
- The machine/human semantic boundary (Q3) is policy guidance, not evidence
  that checks are implemented: proposed machine-verifiable criteria include
  tagged-structure completeness, reading-order integrity, heading hierarchy,
  table-header association, list structure, language declaration,
  form-field labeling, alt-text presence, and link-text structure. Alt-text
  quality, reading-order sensibility, table sensibility, contrast intent,
  assistive-technology behavior, and chart/presentation sensibility remain
  human-judgment review items. The numeric contrast calculation does not
  classify raw colors or establish conformance.
- Open follow-ups: which non-financial family gets the first bounded fixture
  (the draft suggests `forms`); named screen readers for the AT-review
  protocol (NVDA/JAWS/VoiceOver); the incident-notification window. The
  `restricted` dual-hat question remains open for Dano's judgment (see 0015);
  Spanish and `restricted` revisit triggers live in the deferred scope
  (0016).
