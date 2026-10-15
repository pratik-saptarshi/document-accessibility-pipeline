---
name: dap-requesting-code-review
description: Use to prepare an independent review of a meaningful change or before a protected merge.
---

# Requesting code review

Give the reviewer a bounded question, acceptance criteria, exact base and head
revisions, changed paths, and relevant constraints. For a task branch use the
merge base with the intended target (or another explicitly reviewed base), not
an assumed `HEAD~1`. Ask for findings with file/line evidence, impact, and a
reproduction or verification path where practical. Request an independent pass
without sharing other reviewers' conclusions first.

Check all findings against the current revision, repository policy, and
acceptance criteria. Follow the [reviewer brief](reviewer-brief.md). This skill
does not authorize sending a review request, publishing, or merging.
