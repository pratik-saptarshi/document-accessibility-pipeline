---
name: dap-verification-before-completion
description: Use before claiming work is complete or a check passed, and before proposing publication.
---

# Verification before completion

Match each claim to current evidence: exact revision, command or CI check,
complete output/status, and any scope limit. Run verification only when the
user's request and repository rules authorize it. A partial check proves only
that portion; a prior run or another agent's report is not fresh evidence for
this revision. Inspect the resulting diff and confirm relevant requirements
individually. Report unrun checks, failures, and unavailable evidence plainly.

For publication, follow `CONTRIBUTING.md` and `AGENTS.md`: protected checks,
independent review, the guarded HTTPS publisher, and user authorization apply.
Do not imply a pass, merge eligibility, or completion from intention or code
inspection alone.
