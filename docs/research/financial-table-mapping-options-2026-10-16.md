# dap-jo9.1.6.7 — Financial-table mapping options

- Date: 2026-10-16
- Branch: `feat/dap-jo9-1-6-7-mapping-options`
- Context: doc-first research spike exploring candidate financial-table mapping options (discovery only — no automatic mapping authorized). Accepted boundary: operator-confirmed mapping only (see `docs/decisions/0003-operator-confirmed-financial-table-mapping.md`); v2 repair evidence in `docs/research/pdfcpu-financial-table-v2-repair-evidence-2026-10-16.md`.

## Scope and authority

This spike compares mapping options for associating financial-table repair outputs with the v2 candidate data; it authorizes exploration only, and no automatic mapping will be applied without a follow-up decision.

### Option A — Operator-confirmed manual association (status quo)

#### Inputs

The operator supplies associations manually; the system records the confirmed associations.

#### Assumptions

Human confirmation remains the gate; nothing is applied without an explicit operator decision.

#### Fixture-authored expected associations

To be defined in a later task: expected operator confirmations recorded in fixtures.

#### Confidence and ambiguity behavior

To be defined in a later task: how ambiguity is surfaced and held until confirmation.

#### Unsupported boundaries

To be defined in a later task: what the option does not attempt (e.g. automatic association).

### Option B — Fixture-authored lookup

#### Inputs

A lookup authored in the fixture files maps repair outputs to expected associations.

#### Assumptions

Fixtures carry authoritative expected associations; the lookup is exact, not inferred.

#### Fixture-authored expected associations

To be defined in a later task: fixture entries with the authoritative expected mappings.

#### Confidence and ambiguity behavior

To be defined in a later task: handling when a repair output has no fixture entry.

#### Unsupported boundaries

To be defined in a later task: lookup coverage limits and exact-match requirements.

### Option C — Header-text heuristic with confidence threshold

#### Inputs

Column/row header text from the table plus a confidence threshold value.

#### Assumptions

Header text is machine-readable enough for a heuristic score against a threshold.

#### Fixture-authored expected associations

To be defined in a later task: fixture expectations for accepted, rejected, and borderline scores.

#### Confidence and ambiguity behavior

To be defined in a later task: behavior below the threshold and near-tie scoring.

#### Unsupported boundaries

To be defined in a later task: heuristics are advisory only and never applied unconfirmed.

### Option D — Positional/structural inference

#### Inputs

Structural signals from the table layout (row/column positions, spans, ordering).

#### Assumptions

Layout position carries enough regularity to infer associations structurally.

#### Fixture-authored expected associations

To be defined in a later task: fixture expectations for regular and irregular layouts.

#### Confidence and ambiguity behavior

To be defined in a later task: behavior on ambiguous or irregular structure.

#### Unsupported boundaries

To be defined in a later task: inferred associations require operator confirmation before use.

### Option E — Operator-authored external mapping file

#### Inputs

A mapping file authored by the operator outside the repair run, loaded as input.

#### Assumptions

The operator owns the file's content and correctness; the system validates format only.

#### Fixture-authored expected associations

To be defined in a later task: fixture copies of operator-authored mapping files.

#### Confidence and ambiguity behavior

To be defined in a later task: handling of malformed entries and partial coverage.

#### Unsupported boundaries

To be defined in a later task: the system does not generate or repair the file itself.

## Synthetic controls

| Control | Test | Result |
|---|---|---|
| accuracy | TBD | TBD |
| missing | TBD | TBD |
| multiple | TBD | TBD |
| conflicting | TBD | TBD |
| unsafe | TBD | TBD |
| no-unconfirmed-applied | TBD | TBD |

## Recommendation

To be written in a later task: comparative recommendation across Options A–E.

## Follow-up decision

To be written in a later task: the decision record following this spike's recommendation.

## Limits and non-goals

This spike explores options only; it does not implement mapping, apply any automatic mapping, or assert semantic or AT outcomes.
