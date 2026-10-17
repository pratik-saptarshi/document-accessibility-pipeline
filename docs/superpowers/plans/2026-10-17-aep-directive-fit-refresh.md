# AEP directive fit refresh

**Status:** plan for a documentation-only update; no runtime installation or configuration.

## Scope

Refresh the existing fit assessment against AEP `main` at `2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1`, including merged PR #3 (`b7f1db6`) and later watchdog changes. Preserve the project-specific adaptation, license caution, current-environment boundaries, and existing `AGENTS.md`/index links. Do not port upstream text/code or install services, skills, memory, schedules, credentials, or network settings.

## Sequence and acceptance checks

1. **Inspect:** compare the current upstream tree and docs/script behavior with the previously pinned source; verify the license-file state and PR #3/#5 effects.
2. **Update:** revise only the fit assessment and its existing index references; keep unresolved runtime choices explicitly unapproved. Preserve the existing `codex/aep-directive-fit` branch and synchronize `main` with the guarded publisher, without rewriting history.
3. **Review:** inspect the exact diff, verify pinned links and claims, run the repository-required root and nested checks plus publisher/scrub checks, and request an independent review on the exact base/head.
4. **Publish:** use the guarded HTTPS CloudBot publisher and open a draft PR under existing user authorization. Wait for exact-head hosted checks and report the review URL/head; do not merge.

## Branch-name disposition

`codex/aep-directive-fit` predates PR #120's `type/scope-short-description` branch convention. Keep this already-published branch to preserve its review history; this is a bounded exception for this branch only. Use the current convention for new branches.

## Local workflow guidance applied

No standalone upstream Superpowers skill is installed in this environment. This plan follows the repository's locally adapted `dap-feature-pr-pipeline`, `dap-requesting-code-review`, `dap-verification-before-completion`, and `dap-scrubbing-public-tree` guidance, subject to `AGENTS.md` and the user's scope.
