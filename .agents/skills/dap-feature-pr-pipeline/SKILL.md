---
name: dap-feature-pr-pipeline
description: Use for a multi-step feature, fix, or documentation change that should end as a reviewable pull request.
---

# Feature to PR workflow

Use this guide for multi-step work that will be reviewed and delivered in a pull request. Skip it for a one-line change, an exploratory spike, or work the user asked to keep local. Give concise progress updates in commentary.

## Workflow

1. **Establish the scope.** Read `AGENTS.md`, relevant accepted decisions, and the applicable local skills. Check the current branch, worktree status, and intended base. Preserve user work. Use an isolated worktree when it helps separate the change.
2. **Plan the acceptance checks.** Map the requested outcome to observable acceptance criteria and required tests. Record a short plan in the repository's established location when the work warrants one. Do not turn unresolved product choices into implementation assumptions.
3. **Implement test first.** For behavior changes, add a committed regression in the owning language, confirm it fails on the unmodified behavior, then make the smallest fix and confirm it passes. Do not substitute uncommitted or temporary shell assertions for the required regression. For documentation-only changes, verify links, formatting, and any provenance or manifest constraints.
4. **Review independently.** Ask a fresh reviewer to assess the whole branch against the exact base and head, acceptance criteria, and project safeguards. A focused independent review is sufficient unless the user asks for a panel. Re-review any changed final head.
5. **Run hygiene and required gates.** Apply `dap-scrubbing-public-tree` before publication. Run the checks required by `AGENTS.md`, including nested modules and security or CI checks that apply. Keep CI-only tools within their designated CI job.
6. **Publish and integrate under existing authorization.** If the current request or session authorizes publication, use the repository's guarded source-publisher path, HTTPS remote, and required CloudBot author/committer identity. Open or update the authorized PR. Carry forward existing authorization for the same action and stated scope; ask again before any step that expands beyond it (a new PR, arming auto-merge, or merging). Do not bypass branch protection.
7. **Gate the final head.** Read every review comment and inline thread. Address findings or record an evidence-based disposition. Wait until every required check succeeds on the exact final head; any later commit or base synchronization requires fresh checks.
8. **Merge only through the protected repository path.** When the current authorization covers the exact PR and branch, use the repository's protected squash merge path after independent review and exact-head checks pass. Never force-push, rewrite shared history, bypass protection, or merge on behalf of a reviewer.

## Boundaries

- Existing user authorization persists across turns within its stated scope; this guide grants no new authority.
- Use the repository's Go-first and Beads rules. Do not mutate canonical Beads state without an authorized supported operation, or infer issue closure from code completion.
- Preserve partial work and label it accurately. A passing local check does not establish hosted CI, product acceptance, or release readiness.
- Record evidence-backed implementation dispositions when they fit accepted decisions and existing authorization. Route product dispositions through owner product review into Beads with rationale, priority, dependencies, and testable acceptance criteria; agents do not make product-level accept/defer calls on the owner's behalf. Resolve uncertainty only within an authorized, reversible implementation choice; escalate choices that change product scope, policy, or irreversible state.
