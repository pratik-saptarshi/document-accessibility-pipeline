# Decision 0016: Defer `restricted` confidentiality class and Spanish language support

**Status: Accepted.**

**Date: 2026-10-18.**

## Context

Two launch-scope questions surfaced during the profile-policy Q&A round
(recorded by the parallel worker in `0014-profile-policy-decisions.md`; see
DRAFT-0014 Q2/Q5 for the `restricted`-class threat-data analysis). The owner
(Neo) answered both with explicit deferrals:

- `restricted` confidentiality class: "No, defer for now."
- Spanish language support: "for the time being, we are making a DECISION
  to defer Spanish documents."

This record preserves those deferrals, their rationale, the revisit triggers,
and what un-deferral requires.

## Decisions

1. **`restricted` confidentiality class: NOT in scope.**
   - Rationale: the Q5 threat-data controls (least-privilege access,
     NIST 800-88 deletion, hash-only audit log) are proposed but not
     implemented or independently reviewed. Restricted-class processing is
     out of product scope until the controls are evidenced. This is enforced
     at intake: the profile validator rejects `restricted` confidentiality
     values (ASCII whitespace is trimmed and ASCII case-folded before
     comparison, so ASCII whitespace or case variants cannot bypass the
     refusal; exotic Unicode variants are out of scope) with an error
     citing this decision. No production intake path for restricted-class
     documents is claimed here.
   - Revisit trigger: an independent security review of the Q5 controls
     passes.
   - Un-deferral requires: the Q5 controls implemented, an independent
     security review that passes, and an accepted decision record recording
     the review evidence before any restricted-class document is accepted.
2. **Spanish language support: deferred.**
   - Rationale: deferring Spanish keeps launch scope narrow; Spanish is not
     claimed as supported. The profile validator enforces an English-only
     language allowlist (`en`, `en-US`, `en-GB`); other tags such as `es`,
     `es-US`, or `fr` are rejected with an error citing this decision. No
     Spanish fixtures or `/Lang: es` support claims are added by this
     decision.
   - Revisit trigger: the first client request for Spanish documents, or a
     post-launch i18n milestone.
   - Un-deferral requires: an accepted decision record (or i18n milestone
     plan) that re-opens the profile contract, with Spanish fixtures and
     test coverage, before any `/Lang: es` support is claimed.

## Alternatives considered

- Accept `restricted`-class documents now under proposed-but-unevidenced
  controls: rejected; dishonest to claim protections not yet implemented or
  reviewed.
- Ship Spanish at launch as a US SMB differentiator: rejected; not a launch
  blocker and expands scope against the Q2 2027 target.

## Follow-up

- Reference `0014-profile-policy-decisions.md` (DRAFT-0014 Q2/Q5) for the
  threat-data analysis backing the `restricted` scope deferral.
- When the Q5 controls are implemented, schedule the independent security
  review; un-deferral waits on its outcome, not on implementation alone.
- Do not add Spanish fixtures, Spanish documentation claims, or `/Lang: es`
  support language to the profile contract while this decision stands.
