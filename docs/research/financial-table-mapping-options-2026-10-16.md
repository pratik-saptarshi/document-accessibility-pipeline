# dap-jo9.1.6.7 — Financial-table mapping options

- Date: 2026-10-16
- Branch: `feat/dap-jo9-1-6-7-mapping-options`
- Context: doc-first research spike exploring candidate financial-table mapping options (discovery only — no automatic mapping authorized). Accepted boundary: operator-confirmed mapping only (see `docs/decisions/0003-operator-confirmed-financial-table-mapping.md`); v2 repair evidence in `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

## Correction (2026-10-17)

PR #112 review finding [4301459487](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/112#discussion_r4301459487) identified that the earlier report described test-local stub behavior as statement/invoice fixture controls. The claims below are corrected to distinguish the paper comparison and gate-only tests from any future fixture experiment. No real mapping accuracy or statement/invoice fixture result is claimed.

## Correction update (2026-10-18)

After PR #112, fixture-authored control tests were added in
`internal/financialfixtures/mapping_options_fixture_controls_test.go`: a
hand-authored manifest holds the expected associations for the bounded
synthetic statement and invoice fixtures (accuracy match applies;
confirmed-but-contradicting-authored not applied; missing cell yields no
suggestion; ambiguous equal-confidence candidates held; contradictory
authored entries unapplied; authored-unsafe rejected; no unconfirmed
suggestion applied). Each manifest entry is test data typed by hand from the
human-readable observation records in
`docs/fixtures/remediation/financial-tables-v2/observations/`; the tests do
not parse PDFs, implement a mapping algorithm, or perform inference. The
controls exercise the operator-confirmation gate semantics against these
authored associations only.

## Evidence status and limits

This document is a paper comparison of proposed approaches, not a fixture
experiment. The paper comparison of the five options is UNCHANGED: the
expected associations listed under each option remain hypotheses for a
possible future, separately authorized experiment — no option has been tested
against fixture data as a mapping implementation.

What is NEW since the 2026-10-17 correction: the operator-confirmation gate
semantics are now exercised with fixture-authored expected associations as
test inputs (`mapping_options_fixture_controls_test.go`). This is still a
gate-behavior test, not a mapping implementation: it contains no mapping
algorithm, performs no inference, parses no PDF fixtures, applies no
automatic mapping, and establishes no statement/invoice accuracy, ambiguity
detection from document data, or completed mapping comparison. The original
stub harness tests (`mapping_options_spike_test.go`) remain literal stub-only
checks; the new tests reuse the same gate semantics with hand-authored
associations (typed from the human-readable observation records) instead of
fully constructed values.

Unchanged disclaimers: both v2 retained candidates remain NONCOMPLIANT per
`docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`;
operator confirmation grants no semantic or assistive-technology authority;
the follow-up selection decision is still separate and not made here.

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

#### Proposed expected associations (hypotheses only)

- A future statement fixture could propose confirmation of its row-2 amount cell with its amount-column header; a future invoice fixture could propose the corresponding `Unit amount` header association.
- No such confirmation record was read or verified in this spike. The single arbitrary cell/header pair in the test harness is not fixture metadata.

#### Confidence and ambiguity behavior

- Missing confirmation: nothing is applied; the proposal is held and surfaced as pending.
- Multiple or conflicting proposals for one cell: the system does not choose; the operator must select exactly one explicitly, and the record notes the choice.
- Unsafe proposal (one that would alter displayed values, extend beyond the bounded row-2 pair, or change more than the damaged `/Headers`): rejected outright; applying it would require a separate decision, not a confirmation.

#### Unsupported boundaries

- No automatic inference or application of any association, with or without a heuristic score.
- The proposed scope is limited to synthetic statement/invoice cases; no inputs from either family were exercised by the mapping-spike tests.
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

#### Proposed expected associations (hypotheses only)

- A future statement entry could associate the statement row-2 amount cell with its amount-column header; a future invoice entry could associate the invoice row-2 amount cell with its `Unit amount` header.
- These are example expectations for a future fixture-authored lookup, not entries present in or validated against the current fixtures. No lookup data is implemented here.

#### Confidence and ambiguity behavior

- Missing entry for a repair output: no association is applied; the output is flagged as uncovered.
- Multiple entries for one output: held; the fixture must be deduplicated before any decision can consider it.
- Conflicting entries (same output key, different expected associations): held; nothing is applied.
- Unsafe entry (one whose association would change displayed text or render output): never applied, even on exact match.

#### Unsupported boundaries

- No non-synthetic inputs are in the proposal scope; no lookup table exists or was tested.
- No approximate, scored, or inferred matching; exact keys only.
- No unconfirmed application: even a matched entry is a proposal until a follow-up decision says otherwise.
- Any future fixture table would not establish conformance, candidate eligibility, semantic acceptability, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option B is not authorized now: no fixture-authored lookup entries were inspected or validated, and the tests do not exercise a lookup against either PDF fixture. The tests formerly described as synthetic controls exercise only local stub/gate behavior. Even once real fixture entries exist, automatic application of a lookup contradicts Decision 0003's no-automatic-mapping boundary and would require a follow-up decision; fixture maintenance cadence and authorship authority are also undecided. Selection among options belongs to Task 6's recommendation.

### Option C — Header-text heuristic with confidence threshold

#### Inputs

- Header text strings extracted from the table's tagged structure for the candidate header cells.
- The row-2 amount cell's candidate headers with a deterministic heuristic score per candidate.
- A confidence threshold value recorded with the run; scores and the threshold are advisory only.

#### Assumptions

- Hypothesis: header text for a future synthetic case can be extracted deterministically and a scoring function can be deterministic for given inputs.
- The threshold is a recorded parameter, not a calibrated constant; it carries no validated meaning until measured.
- Scores propose; they never apply an association on their own.

#### Proposed expected associations (hypotheses only)

- Hypothesis for a future statement experiment: score its row-2 amount cell against its amount-column header and compare with alternatives.
- Hypothesis for a future invoice experiment: evaluate the row-2 amount cell against the `Unit amount` header and alternatives.
- Missing-text and near-threshold cases are proposed controls only. No heuristic, threshold, PDF header extraction, or expected result is implemented or tested here.

#### Confidence and ambiguity behavior

- Score at or above threshold: proposal only; still requires operator confirmation before any use.
- Below threshold: held; the proposal is surfaced as unconfirmed with its score.
- Near-tie (two candidates within epsilon): flagged ambiguous; no preference applied.
- Missing header text: no score computed; held.
- Conflicting signals (header text suggests one candidate, structure order suggests another): held; the conflict is surfaced.
- Unsafe score (driven by duplicated or manipulated header text): rejected; the proposal is discarded.

#### Unsupported boundaries

- Heuristics are advisory only and are never applied unconfirmed; unconfirmed application would cross the 0003 boundary.
- The proposed scope excludes general documents and non-synthetic inputs; no heuristic or input family was tested.
- Not a production writer; the option describes scoring, not processing.
- A high score does not establish semantic correctness, conformance, candidate eligibility, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option C is not authorized now: no heuristic or threshold implementation exists, and no accuracy, borderline, or rejection behavior has been observed on fixture data. Any threshold would be arbitrary. Any future score must remain advisory and still require the operator gate under Decision 0003. A follow-up decision must define the threshold methodology, tie-breaking, and failure controls before any use; comparative selection belongs to Task 6's recommendation.

### Option D — Positional/structural inference

#### Inputs

- Structural signals from the tagged table layout: row and column indices, cell spans, and reading order from the structure tree.
- The row-2 amount cell's position and the amount-column header's position within each synthetic family.

#### Assumptions

- Hypothesis: a future statement or invoice case may have a regular layout where position identifies the expected header for the bounded pair.
- Hypothesis: a future structure-tree parser could supply machine-derived row, column, span, and reading-order inputs without parse error.
- Inferred associations are proposals; inference is not confirmation.

#### Proposed expected associations (hypotheses only)

- Hypotheses for a future statement/invoice experiment would compare the row-2 amount cell with the amount-column header using actual structure-tree positions.
- Irregular layouts such as spanning cells or reordered structure are possible future negative controls. No structure tree was inspected and no positional inference was tested in this spike.

#### Confidence and ambiguity behavior

- A future layout classified as regular: inferred proposal only; still requires operator confirmation before any use.
- Missing structure signals: no inference; held.
- Multiple plausible headers (spanning or overlapping header cells): held; the system does not prefer one.
- Conflicting order signals: held; the conflict is surfaced.
- Unsafe inference (one that would change displayed values or extend beyond the bounded pair): never applied.

#### Unsupported boundaries

- No layout regularity is demonstrated: neither the synthetic pair nor irregular or general documents were evaluated by this spike.
- No unconfirmed application: inference counts as automatic mapping in decision 0003's terms and is not accepted.
- Not a production writer; the option describes inference, not processing.
- A structurally regular inference does not establish semantic correctness, conformance, candidate eligibility, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option D is not authorized now: structural regularity has not been evaluated on statement, invoice, irregular, or general layouts. Inference would be automatic mapping in Decision 0003 terms and is not accepted. A follow-up decision must define regularity evidence and irregular-input controls before any use; comparative selection belongs to Task 6's recommendation.

### Option E — Operator-authored external mapping file

#### Inputs

- A mapping file authored by the operator outside the repair run, loaded as input: entries of fixture-key, cell reference, and header reference, with a schema version.
- The operator's identity attached to the file; the file is versioned against the run it targets.

#### Assumptions

- The operator owns the file's content and correctness; the system validates format and schema only, never content.
- The file applies only to the exact damaged fixture revision it names; a revision change invalidates it.
- Loading a file is not confirmation; entries are proposals until a confirmation gate says otherwise.

#### Proposed expected associations (hypotheses only)

- A future operator-authored mapping file could propose the statement row-2 amount-to-amount-column pair and the invoice row-2 amount-to-`Unit amount` pair.
- No mapping file, schema, or fixture bundle was loaded or validated here; these are hypothetical examples only.

#### Confidence and ambiguity behavior

- Malformed entry (bad format, unknown schema version, dangling cell or header reference): rejected with an error; no partial application of the file.
- Missing entry for a repair output: held; the output is flagged as uncovered.
- Multiple or conflicting entries for one cell: held; the file must be corrected by its author.
- Unsafe entry (outside the bounded pair, or implying changed displayed values): rejected.

#### Unsupported boundaries

- The system does not generate, complete, or repair the mapping file itself.
- The proposed scope excludes non-synthetic inputs and general documents; no mapping file was loaded or tested.
- No unconfirmed application of file entries; operator authorship does not substitute for the 0003 confirmation gate.
- A mapping file does not establish conformance, candidate eligibility, semantic acceptability, or assistive-technology acceptability; both v2 retained candidates remain NONCOMPLIANT per `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

Why Option E is not authorized now: no mapping-file schema, validation rules, or operator-authoring protocol exist. Operator authorship moves the confirmation burden rather than resolving it - the file's contents still need a confirmation gate, and whether an external file satisfies the 0003 boundary is undecided. A follow-up decision must define the schema, validation, and authoring protocol before any use; comparative selection belongs to Task 6's recommendation.

## Test-local harness checks

Two generations of test-only checks exist. Neither reads either PDF fixture,
executes a lookup, heuristic, structural inference, or mapping-file
implementation, and neither establishes mapping accuracy or fixture-control
results beyond the gate behavior shown. The "Evidence scope" column marks
which generation each check belongs to: **stub-only** (constructed literals,
no fixture input) vs **fixture-authored-association** (gate inputs typed by
hand from the human-readable observation records, still no PDF parsing and no
mapping algorithm).

### Original stub harness tests (`mapping_options_spike_test.go`) — stub-only scope

| Harness behavior | Test | Evidence scope |
|---|---|---|
| A safe, confirmed value passed directly to the gate is accepted; an unconfirmed or unsafe value is not. | `TestHarnessAppliesOnlySafeConfirmedSuggestion` | Passes locally; only gate behavior on literal inputs. |
| A stub closure represents a missing value. | `TestHarnessMissingStubReturnsNoSuggestion` | Passes locally; no fixture lookup. |
| Two constructed equal-confidence values are left unconfirmed and marked ambiguous by the test. | `TestHarnessAmbiguousStubSuggestionsRemainUnconfirmed` | Passes locally; no heuristic or ambiguity detector. |
| Two constructed different values are left unapplied and marked conflicting by the test. | `TestHarnessConflictingStubSuggestionsRemainUnapplied` | Passes locally; no resolver. |
| An explicitly unsafe constructed value is rejected by the gate. | `TestHarnessUnsafeConfirmedSuggestionIsRejected` | Passes locally; no document-safety analysis. |
| Constructed values without confirmation are not applied. | `TestHarnessUnconfirmedStubValuesNeverApply` | Passes locally; no mapping source. |

### Fixture-authored control tests (`mapping_options_fixture_controls_test.go`) — fixture-authored-association scope

The hand-authored manifest cites
`docs/fixtures/remediation/financial-tables-v2/observations/statement-positive-v2.txt`
(statement row-2 amount cell → `Amount` header, confidence 0.95) and
`invoice-positive-v2.txt` (invoice row-2 unit-amount cell → `Unit amount`
header, confidence 0.95). The driver (`fixtureDriver`) feeds these authored
entries through the existing gate semantics; it never confirms anything
itself — confirmation always comes from the test as a simulated operator.
Ambiguity/conflict/unsafe negative controls in this file are constructed test
controls, marked as such at each site, not fixture evidence.

| Gate behavior exercised | Test | Evidence scope |
|---|---|---|
| Authored entry plus matching confirmed suggestion is applied. | `TestFixtureAuthoredAccuracyMatchApplies` | Passes locally; fixture-authored association as gate input, not an inference result. |
| A confirmed suggestion contradicting the authored entry is recorded as a contradiction and never applied. | `TestFixtureAuthoredContradictionNotApplied` | Passes locally; gate behavior on authored input. |
| A cell with no authored entry produces no suggestion; the gate is never touched for it. | `TestFixtureAuthoredMissingCellNoSuggestion` | Passes locally; no fixture lookup or fallback. |
| Two constructed equal-confidence candidates are left unconfirmed and flagged ambiguous. | `TestFixtureAuthoredAmbiguousCandidatesHeld` | Passes locally; constructed control, not fixture evidence. |
| Two constructed contradictory authored entries for one cell are left unapplied and flagged conflicting. | `TestFixtureAuthoredConflictingEntriesUnapplied` | Passes locally; constructed control, not fixture evidence. |
| An authored entry flagged unsafe causes rejection even of a matching confirmed suggestion. | `TestFixtureAuthoredUnsafeEntryRejected` | Passes locally; constructed control, not document-safety analysis. |
| Authored entries with no confirmation never apply (statement and invoice cells). | `TestFixtureAuthoredUnconfirmedNeverApplies` | Passes locally; fixture-authored associations as gate inputs. |

## Recommendation

No mapping option is authorized by this spike. This is a recommendation only,
not a selection: per the issue (dap-jo9.1.6.7) this spike recommends whether
a separately authorized decision is warranted, without selecting or
implementing any approach.

Of the five options, **Option B (fixture-authored lookup)** and **Option E
(operator-authored external mapping file)** are the lowest-risk design
candidates for a future decision based on their proposed exact-match and
hold-on-ambiguity properties. This is a paper assessment only: no option was
tested against fixture data, and the local gate stubs do not prove these
properties for any implementation. Options C (heuristic scoring) and D
(structural inference) remain unimplemented hypotheses requiring the operator
gate plus future threshold/regularity calibration. Option A is the accepted
bounded boundary as-is, not an expansion.

Guard (per Decision 0003): operator confirmation grants no semantic-review,
assistive-technology, or release-signoff authority. A confirmation — for the
bounded row-2 association or for any future mapping approach — establishes no
conformance, candidate eligibility, semantic acceptability, or
assistive-technology acceptability; both v2 retained candidates remain
NONCOMPLIANT.

## Follow-up decision

A separate decision is required before any mapping approach is selected,
built, or used: a new decision-scope beads issue for the financial-table
mapping approach, referencing this paper comparison, the test-local gate
evidence, and Decision 0003's boundary. This spike does not open that
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
- No statement/invoice fixture data was read by the test-local harness: no
  test opens or parses a PDF fixture. The fixture-authored control tests use
  associations typed by hand from the human-readable observation records
  (test data citing the evidence, not fixture metadata and not parsed
  output), so these tests establish no family-specific result or
  generalization; the paper comparison remains hypotheses only.
- No production writer selection, profile defaults, or release decisions are
  made here.
- No conformance, candidate eligibility, semantic, or assistive-technology
  claims are made; operator confirmation is not semantic/AT authority.
- This spike opens no beads issues and closes none; the follow-up mapping
  decision and the dap-jo9.1.1.6 decision remain separate.
