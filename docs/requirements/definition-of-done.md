# Shared engineering definition of done

Canonical engineering completion criteria for Beads issues in this repository.
Per-issue acceptance criteria reference this record and add only checkable
per-issue criteria. This record does not approve a profile, writer, or
release, and does not replace per-issue evidence.

Extracted in `dap-jo9.5.4` from identical boilerplate pasted across nine
issues. Each clause below is stated in checkable form: the evidence that
satisfies it is named, and a pass in one outcome is never cited as a pass
in another.

## Test-first evidence

For each vertical slice, write failing tests before behavior. Record each
passing execution with six fields: artifact, input hash, candidate hash,
evidence hash, tool versions, and limits.

Check: the execution record exists with all six fields, or an explicit
not-applicable entry where a field cannot exist (see below).

For executions that correctly reject input before producing output
(negative/failure controls), record the rejection with the input hash and
the rejection reason; artifact, candidate hash, and evidence hash are
recorded as not applicable with that reason. A rejection recorded this way
satisfies this clause.

## Declared scope

State the supported subset and the explicit unsupported outcomes. Cover
negative and failure controls.

Check: the scope statement names both supported and unsupported outcomes;
negative and failure tests exist.

## Gates and policy defaults

Apply only gates that are both accepted and necessary for the slice.
Record each applied gate's accepted source and its relevance to the slice.
Every policy default applied is recorded with its source decision; no
default is applied without a recorded decision.

Check: the applied gates are listed with sources and relevance; no
accepted-but-irrelevant gate is applied. This replaces the old
unverifiable form ("Do not silently select policy defaults") with a
positive, auditable requirement: a default with no recorded source
decision fails this clause.

## Separate outcomes

Record engineering completion, machine validation, human semantic review,
assistive-technology review, and release eligibility as separate outcomes.
A pass in one is never cited as a pass in another.

Check: the evidence record carries five separate entries. In particular,
machine validation is not semantic approval, and a candidate pass is not
release eligibility.
