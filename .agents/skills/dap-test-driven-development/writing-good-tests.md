# Writing good tests

Use when adding or changing tests, mocks, or test helpers.

- Before writing a test, state the realistic production regression it should
  catch. Prefer user-visible or contract behavior over implementation details.
- Derive `want` values independently. A test that computes its expectation with
  the code under test can agree with the same bug.
- Prefer actual components. Mock only slow, external, or otherwise uncontrollable
  boundaries; understand the replaced method's side effects and assert behavior
  at the component boundary.
- Keep test-only cleanup/helpers out of production APIs. Use literal boundary,
  malformed, empty, and failure cases when they matter to the contract.
- Do not test documentation by grepping for required phrases. Check downstream
  behavior or review the instructions directly.

Repository test commands and limits are in `AGENTS.md`. Do not run commands that
the task or governing instructions do not authorize.
