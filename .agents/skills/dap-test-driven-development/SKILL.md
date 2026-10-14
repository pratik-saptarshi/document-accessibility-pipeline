---
name: dap-test-driven-development
description: Use when planning an implementation or bug fix that needs behavior tests.
---

# Test-driven development

For a behavior change, identify a concrete observable regression first. Add a
focused test, run it to confirm the expected failure, implement the smallest
change, then run the relevant verification the task and repository policy call
for. Refactor only while the behavior remains covered. Keep expected values
independent of the code under test; prefer real behavior over assertions about
mock calls. See [writing good tests](writing-good-tests.md).

Preserve the user's and repository's existing work. Never delete or discard
pre-existing implementation to force a test-first sequence; characterize the
current behavior, then add a focused regression test. Do not add or run tests
unless the user asks or the active task expressly requires verification.

Follow Go and test requirements in `AGENTS.md`; repository guidance and the
current user request govern any exception or scope limit. This skill does not
authorize product implementation when the request is planning-only.
