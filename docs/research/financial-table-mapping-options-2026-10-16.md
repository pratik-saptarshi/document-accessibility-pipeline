# dap-jo9.1.6.7 — Financial-table mapping options

- Date: 2026-10-16
- Branch: `feat/dap-jo9-1-6-7-mapping-options`
- Context: doc-first research spike exploring candidate financial-table mapping options (discovery only — no automatic mapping authorized). Accepted boundary: operator-confirmed mapping only (see `docs/decisions/0003-operator-confirmed-financial-table-mapping.md`); v2 repair evidence in `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

## Scope and authority

This spike compares mapping options for associating financial-table repair outputs with the v2 candidate data; it authorizes exploration only, and no automatic mapping will be applied without a follow-up decision.

### Option A — Operator-confirmed manual association (status quo)

#### Inputs

- The repair output with the damaged row-2 amount-cell `/Headers` association exposed as an unapplied proposal.
- An explicit operator confirmation record: operator identity, timestamp, the exact association confirmed, and the byte-identical candidate it applies to.
- The accepted workflow boundary in `docs/decisions/0003-operator-confirmed-financial-table-mapping.md`; no confirmation is valid outside that boundary.

#### Assumptions

- The operator acts within the bounded first workflow only; confirmation does not grant semantic-review, assistive-technology, or release-signoff authority (per 0003).
- A confirmation applies only to the named row-2 amount-cell pair and only while the candidate bytes are unchanged; any byte change invalidates it.
- The system never pre-applies a proposal: the proposal stays inert until the operator confirms.

#### Fixture-authored expected associations

- Each synthetic family records the expected operator confirmation in its fixture: the statement family's row-2 amount cell is confirmed associated with the statement amount-column header; the invoice family's row-2 amount cell is confirmed associated with the invoice amount-column header.
- The fixture records exactly one confirmation per family, targeting only the deliberately damaged `/Headers` association; no other association is confirmed.

#### Confidence and ambiguity behavior

- Missing confirmation: nothing is applied; the proposal is held and surfaced as pending.
- Multiple or conflicting proposals for one cell: the system does not choose; the operator must select exactly one explicitly, and the record notes the choice.
- Unsafe proposal (one that would alter displayed values, extend beyond the bounded row-2 pair, or change more than the damaged `/Headers`): rejected outright; applying it would require a separate decision, not a confirmation.

#### Unsupported boundaries

- No automatic inference or application of any association, with or without a heuristic score.
- No general documents and no non-synthetic inputs; scope stays with the synthetic statement/invoice families.
- Not a production writer: the option describes a confirmation gate, not processing.
- Operator confirmation does not establish conformance, candidate eligibility, semantic acceptability, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option A is not authorized now: Option A is the current accepted boundary, not an authorized expansion. Decision 0003 accepts operator confirmation only for the first bounded workflow; this spike is discovery-only, and any comparative selection belongs to Task 6's recommendation and a follow-up decision. Scaling A beyond the bounded workflow (general documents, non-synthetic inputs) is not decided, and confirmation alone establishes no conformance, semantic, or assistive-technology outcome.

### Option B — Fixture-authored lookup

#### Inputs

- The repair output's identity as an exact fixture key (structure snapshot hash pinned to the damaged fixture revision).
- A fixture-authored table mapping each exact output key to its expected association; matching is exact, never approximate.

#### Assumptions

- Fixtures enumerate every case the option is asked about; there is no fallback outside the table.
- Fixture authorship is trusted for expected associations only; fixtures record expectations, not production authorization.
- Entries are pinned to the exact damaged byte state of the fixtures; a fixture revision change invalidates its entries.

#### Fixture-authored expected associations

- Statement family entry: the statement row-2 amount cell maps to the statement amount-column header.
- Invoice family entry: the invoice row-2 amount cell maps to the invoice amount-column header.
- Each entry covers only the deliberately damaged `/Headers` association; no entry maps any other cell.

#### Confidence and ambiguity behavior

- Missing entry for a repair output: no association is applied; the output is flagged as uncovered.
- Multiple entries for one output: held; the fixture must be deduplicated before any decision can consider it.
- Conflicting entries (same output key, different expected associations): held; nothing is applied.
- Unsafe entry (one whose association would change displayed text or render output): never applied, even on exact match.

#### Unsupported boundaries

- No non-synthetic inputs; the table only covers the synthetic statement/invoice families.
- No approximate, scored, or inferred matching; exact keys only.
- No unconfirmed application: even a matched entry is a proposal until a follow-up decision says otherwise.
- Fixtures do not establish conformance, candidate eligibility, semantic acceptability, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option B is not authorized now: no fixture-authored lookup entries exist yet - the Synthetic controls results have not been recorded, so there is nothing to look up. Even once fixtures exist, automatic application of a lookup contradicts decision 0003's no-automatic-mapping boundary and would require a follow-up decision; fixture maintenance cadence and authorship authority are also undecided. Selection among options belongs to Task 6's recommendation.

### Option C — Header-text heuristic with confidence threshold

#### Inputs

- Header text strings extracted from the table's tagged structure for the candidate header cells.
- The row-2 amount cell's candidate headers with a deterministic heuristic score per candidate.
- A confidence threshold value recorded with the run; scores and the threshold are advisory only.

#### Assumptions

- Header text is extractable deterministically for the synthetic families; the scoring function is deterministic for given inputs.
- The threshold is a recorded parameter, not a calibrated constant; it carries no validated meaning until measured.
- Scores propose; they never apply an association on their own.

#### Fixture-authored expected associations

- Statement family fixture: the heuristic is expected to score the statement row-2 amount cell's association with the statement amount-column header at or above the threshold, and below the threshold for the other headers.
- Invoice family fixture: the same expectation holds for the invoice row-2 amount cell and the invoice amount-column header.
- Fixtures also record expected rejections (header text missing or unreadable yields no score) and borderline cases (scores within epsilon of the threshold are held).

#### Confidence and ambiguity behavior

- Score at or above threshold: proposal only; still requires operator confirmation before any use.
- Below threshold: held; the proposal is surfaced as unconfirmed with its score.
- Near-tie (two candidates within epsilon): flagged ambiguous; no preference applied.
- Missing header text: no score computed; held.
- Conflicting signals (header text suggests one candidate, structure order suggests another): held; the conflict is surfaced.
- Unsafe score (driven by duplicated or manipulated header text): rejected; the proposal is discarded.

#### Unsupported boundaries

- Heuristics are advisory only and are never applied unconfirmed; unconfirmed application would cross the 0003 boundary.
- No general documents and no non-synthetic inputs.
- Not a production writer; the option describes scoring, not processing.
- A high score does not establish semantic correctness, conformance, candidate eligibility, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option C is not authorized now: the confidence threshold has no calibrated basis - no measured accuracy, borderline, or rejection behavior exists on the synthetic controls, so any threshold value would be arbitrary. Any score is advisory-only and still needs the operator gate under decision 0003. A follow-up decision must define the threshold methodology, tie-breaking, and failure controls before any use; comparative selection belongs to Task 6's recommendation.

### Option D — Positional/structural inference

#### Inputs

- Structural signals from the tagged table layout: row and column indices, cell spans, and reading order from the structure tree.
- The row-2 amount cell's position and the amount-column header's position within each synthetic family.

#### Assumptions

- The synthetic statement and invoice families have regular layouts where position deterministically identifies the correct header for the bounded pair.
- The structure tree parses without error; inference inputs are machine-derived, not human-authored.
- Inferred associations are proposals; inference is not confirmation.

#### Fixture-authored expected associations

- Statement family fixture: the statement row-2 amount cell (row 2, amount column) is inferred to associate with the statement amount-column header (header row, amount column).
- Invoice family fixture: the invoice row-2 amount cell (row 2, amount column) is inferred to associate with the invoice amount-column header (header row, amount column).
- Irregular-layout fixtures (spanning cells, reordered structure) record the expected outcome: held with no inference.

#### Confidence and ambiguity behavior

- Regular structure matching the synthetic pair: inferred proposal only; still requires operator confirmation before any use.
- Missing structure signals: no inference; held.
- Multiple plausible headers (spanning or overlapping header cells): held; the system does not prefer one.
- Conflicting order signals: held; the conflict is surfaced.
- Unsafe inference (one that would change displayed values or extend beyond the bounded pair): never applied.

#### Unsupported boundaries

- No irregular or general documents and no non-synthetic inputs; regularity is demonstrated only on the synthetic pair.
- No unconfirmed application: inference counts as automatic mapping in decision 0003's terms and is not accepted.
- Not a production writer; the option describes inference, not processing.
- A structurally regular inference does not establish semantic correctness, conformance, candidate eligibility, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option D is not authorized now: structural regularity is demonstrated only on the synthetic statement/invoice pair; there is no evidence on irregular or general layouts, and inference is automatic mapping in 0003 terms, which is not accepted. A follow-up decision must define the regularity evidence and the irregular-input controls before any use; comparative selection belongs to Task 6's recommendation.

### Option E — Operator-authored external mapping file

#### Inputs

- A mapping file authored by the operator outside the repair run, loaded as input: entries of fixture-key, cell reference, and header reference, with a schema version.
- The operator's identity attached to the file; the file is versioned against the run it targets.

#### Assumptions

- The operator owns the file's content and correctness; the system validates format and schema only, never content.
- The file applies only to the exact damaged fixture revision it names; a revision change invalidates it.
- Loading a file is not confirmation; entries are proposals until a confirmation gate says otherwise.

#### Fixture-authored expected associations

- Fixtures bundle copies of operator-authored mapping files: the statement file carries an entry associating the statement row-2 amount cell with the statement amount-column header; the invoice file carries an entry associating the invoice row-2 amount cell with the invoice amount-column header.
- Each fixture file entry targets only the deliberately damaged `/Headers` association.

#### Confidence and ambiguity behavior

- Malformed entry (bad format, unknown schema version, dangling cell or header reference): rejected with an error; no partial application of the file.
- Missing entry for a repair output: held; the output is flagged as uncovered.
- Multiple or conflicting entries for one cell: held; the file must be corrected by its author.
- Unsafe entry (outside the bounded pair, or implying changed displayed values): rejected.

#### Unsupported boundaries

- The system does not generate, complete, or repair the mapping file itself.
- No non-synthetic inputs and no general documents.
- No unconfirmed application of file entries; operator authorship does not substitute for the 0003 confirmation gate.
- A mapping file does not establish conformance, candidate eligibility, semantic acceptability, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option E is not authorized now: no mapping-file schema, validation rules, or operator-authoring protocol exist. Operator authorship moves the confirmation burden rather than resolving it - the file's contents still need a confirmation gate, and whether an external file satisfies the 0003 boundary is undecided. A follow-up decision must define the schema, validation, and authoring protocol before any use; comparative selection belongs to Task 6's recommendation.

## Synthetic controls

All controls run locally against the synthetic statement/invoice fixtures
only; they exercise the spike harness behaviors (hold, flag, or reject
proposals), not production mapping.

| Control | Test | Result |
|---|---|---|
| accuracy | `TestMappingOptionsAccuracy` | PASS (local) |
| missing | `TestMappingOptionsMissing` | PASS (local) |
| multiple | `TestMappingOptionsMultiple` | PASS (local) |
| conflicting | `TestMappingOptionsConflicting` | PASS (local) |
| unsafe | `TestMappingOptionsUnsafe` | PASS (local) |
| no-unconfirmed-applied | `TestMappingOptionsNoUnconfirmedApplied` | PASS (local) |

## Recommendation

No mapping option is authorized by this spike. This is a recommendation only,
not a selection: per the issue (dap-jo9.1.6.7) this spike recommends whether
a separately authorized decision is warranted, without selecting or
implementing any approach.

Of the five options, **Option B (fixture-authored lookup)** and **Option E
(operator-authored external mapping file)** are the lowest-risk candidates for
that future decision: both keep associations exact and fixture-pinned, hold
on missing/multiple/conflicting entries, reject unsafe entries, and apply no
association unconfirmed, so they stay inside Decision 0003's
no-automatic-mapping boundary. Options C (heuristic scoring) and D
(structural inference) remain advisory proposals that would still require the
operator gate plus further threshold/regularity calibration. Option A is the
accepted bounded boundary as-is, not an expansion.

Guard (per Decision 0003): operator confirmation grants no semantic-review,
assistive-technology, or release-signoff authority. A confirmation — for the
bounded row-2 association or for any future mapping approach — establishes no
conformance, candidate eligibility, semantic acceptability, or
assistive-technology acceptability; both v2 retained candidates remain
NONCOMPLIANT.

## Follow-up decision

A separate decision is required before any mapping approach is selected,
built, or used: a new decision-scope beads issue for the financial-table
mapping approach, referencing this spike's option evidence, the synthetic
control results, and Decision 0003's boundary. This spike does not open that
issue; it records the need and the recommended low-risk candidates.

Related items, per the local beads tracker:
- dap-jo9.1.1.6 (open) — select the authoritative approver-role evidence
  source and temporal policy; its outcome constrains any confirmation-gated
  mapping workflow.
- dap-jo9.1.1.4 (closed as accepted Decision 0004, approver-user persona) —
  human semantic and assistive-technology reviewer authority is decided in
  principle; reviewer identity, authority, protocol, outcome handling, and
  approval invalidation remain separate evidence gates and are not
  established by any mapping confirmation.

## Limits and non-goals

This spike explores options only; it does not implement mapping, apply any
automatic mapping, or assert semantic or AT outcomes. Specifically:

- No automatic mapping is applied or authorized; every option is evaluated as
  a confirmation-gated proposal under Decision 0003.
- Synthetic fixtures only: results cover the statement/invoice families and
  generalize to no other documents, layouts, or inputs.
- No production writer selection, profile defaults, or release decisions are
  made here.
- No conformance, candidate eligibility, semantic, or assistive-technology
  claims are made; operator confirmation is not semantic/AT authority.
- This spike opens no beads issues and closes none; the follow-up mapping
  decision and the dap-jo9.1.1.6 decision remain separate.
