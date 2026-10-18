# Decision 0014: Profile policy decisions

**Status: Accepted.** Supersedes and replaces DRAFT-0014 with the owner's
answers below. Where a DRAFT-0014 recommendation differs, the owner's answer
here governs.

**Date: 2026-10-18.**

## Context

Primary product goal (owner, 2026-10-18): help SMB clients **and** enterprises
prepare for WCAG document-accessibility requirements and for **WCAG 3.0**.
Secondary goal: tooling readiness for upcoming federal and state
digital/web/document accessibility mandates, anchored to our **internal
readiness target of April 1, 2027** — our own planning goal, not a statutory
deadline. That is distinct from the DOJ ADA Title II final rule's
**April 26, 2027** compliance date for larger public entities (populations
of 50,000 or more) per the DOJ's 2024 web rule guidance
(https://www.ada.gov/resources/2024-03-08-web-rule/). The two dates must not
be conflated: one is internal, the other is statutory applicability.

PDF/UA-2 is the accepted output target (Decision 0001). This record answers
the five open profile-policy questions in `docs/profiles/README.md` so the
`internal/profile` contract can carry real values instead of placeholders.

Regulatory backdrop as of 2026-10 (verify before relying on it): WCAG 2.2 AA
is the current baseline standard. WCAG 3.0 is a W3C draft with a different
conformance model (bronze/silver/gold outcomes, APCA-based contrast replacing
WCAG 2.x contrast ratios); it is not final — do not promise WCAG 3.0
conformance claims. US federal: Section 508 (refresh) references WCAG 2.0 AA.
The EU European Accessibility Act applies from June 2025.

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
   Approved per-family structural fields:
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

2. **DRAFT-0010 (approver-role evidence source): ACCEPTED.** Human review
   authority follows the three-role model with SMB collapse: the **document
   owner** declares the profile (family, language, confidentiality); the
   **accessibility reviewer** runs the human-judgment checklist; the
   **approver** signs off, with the approver's identity resolved from the
   owner-approved evidence source DRAFT-0010 selects. In SMB reality the
   reviewer is often the owner wearing a second hat — that must be *recorded*
   as a dual-hat review, not hidden. Full details are in the companion
   decision record
   [0015-approver-role-evidence-source.md](0015-approver-role-evidence-source.md).

3. **`restricted` confidentiality class: DEFERRED.** `restricted`-class
   documents (regulated or sensitive data) are out of scope until the Q5
   threat-data controls are implemented and independently reviewed; the
   pipeline must not accept them before then. The `public`/`internal`
   classes, the tiered retention model, and the Q5 handling rules stand.
   Deferred-scope details are in
   [0016-deferred-scope-restricted-spanish.md](0016-deferred-scope-restricted-spanish.md).

4. **APCA contrast checking: INCLUDED (advisory, draft-stage per WCAG 3.0).**
   The pipeline includes an APCA evaluation path, treated as advisory while
   WCAG 3.0 remains a draft. WCAG 2.x contrast ratios must not be hard-coded
   as pass/fail gates for generated output: WCAG 3.0's APCA model changes
   contrast evaluation and its conformance model is outcome-based
   (bronze/silver/gold), so a pipeline that conflates "tags present" with
   "document is accessible" would fail under it. Keeping the machine/human
   evidence boundary explicit is the WCAG-3.0-aligned posture.

5. **Internal readiness target: April 1, 2027 (year stated explicitly).**
   April 1, 2027 is the product's internal readiness target — our own
   planning goal, not a statutory deadline. It is separate from the DOJ ADA
   Title II final rule's **April 26, 2027** compliance date for larger public
   entities (populations of 50,000 or more), per the DOJ's 2024 web rule
   guidance. The two dates must not be conflated: one is internal, the
   other is statutory applicability.
   Retention tiers stand: `transient` (deleted after the pipeline run
   completes), `mandate-horizon` (through the horizon), `regulatory` (7
   years — the common US financial-record floor).

6. **Spanish language support: DEFERRED.** The pipeline launches English-only.
   Language is per-document, declared in the profile, and must match the
   PDF's `/Lang` entry. Deferred-scope details are in
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
- **Deferring APCA until WCAG 3.0 stabilizes:** superseded; APCA is included
  as advisory/draft-stage instead.
- **Hard-coding WCAG 2.x contrast ratios as pass/fail:** rejected; APCA and
  outcome-based conformance replace them.
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

- The four-family schema change to `internal/profile/profile.go` is
  implemented alongside this record. The accepted
  `Decision.Value` fields move from placeholders to the values above with
  status `proposed` (the contract's validation admits `proposed`/`open`
  values); `pdfua_target` stays `accepted` as PDF/UA-2.
- SMB vs enterprise is a deployment profile, not two products: the same
  contract serves both; only the selected values differ. Do not fork the
  schema.
- The machine/human semantic boundary (Q3) is now policy: machine-verifiable
  criteria (tagged-structure completeness, reading-order integrity, heading
  hierarchy, table header association, list structure, language declaration,
  form-field labeling, alt-text presence, link-text structure) are asserted
  by the pipeline; alt-text quality, reading-order sensibility, table
  sensibility, contrast intent, assistive-technology behavior, and
  chart/presentation sensibility remain human-judgment review items the
  pipeline must not claim.
- Open follow-ups: which non-financial family gets the first bounded fixture
  (the draft suggests `forms`); named screen readers for the AT-review
  protocol (NVDA/JAWS/VoiceOver); the incident-notification window. The
  `restricted` dual-hat question remains open for Dano's judgment (see 0015);
  Spanish and `restricted` revisit triggers live in the deferred scope
  (0016).
