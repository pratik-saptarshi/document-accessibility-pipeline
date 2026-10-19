---
name: dap-verification-before-completion
description: Use before claiming work is complete or a check passed, and before proposing publication.
---

# Verification before completion

Match each claim to current evidence: exact revision, gate/check identity,
invocation or CI workflow, status, and any scope limit. For a gate result,
record whether it is mandatory, the exact command/profile, exit or timeout
status (or execution error), and the retained evidence location. Distinguish
required evidence from optional diagnostics: if acceptance requires an
artifact, failure to create or retain it fails that criterion even when the
command exits successfully.

For a new or changed gate, verify that the required workflow actually invokes
it and that a meaningful negative control is rejected. A standalone self-test
or a green run alone does not prove the gate is wired or non-vacuous. When a
committed regression is practical, establish its expected failure before the
fix and its pass after the fix. Restore any temporary mutation and confirm the
working tree is clean.

For review findings, map each finding to its changed diff hunk and acceptance
criterion. After the final change, reread the latest review comments/threads
and rerun affected checks at the exact final head. Run verification only when
the user's request and repository rules authorize it. A partial check proves
only that portion; a prior run or another agent's report is not fresh evidence
for this revision. Inspect the resulting diff and confirm relevant requirements
individually. Report unrun checks, failures, and unavailable evidence plainly.

For publication, follow `CONTRIBUTING.md` and `AGENTS.md`: protected checks,
independent review, the guarded HTTPS publisher, and user authorization apply.
Do not imply a pass, merge eligibility, or completion from intention or code
inspection alone.
