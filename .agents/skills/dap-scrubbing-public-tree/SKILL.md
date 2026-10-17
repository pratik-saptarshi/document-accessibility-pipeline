---
name: dap-scrubbing-public-tree
description: Use when preparing an agent-built repo or branch for public viewing — after agentic implementation work, before opening a PR or publishing. Triggers: docs or comments containing agent headers, absolute machine paths, or bot identity strings; process artifacts (plans, designs, reviews, integration logs) sitting in the public tree; nested scratch or spike modules; duplicate CHANGELOG sections.
---

# Scrubbing the Public Tree

## Overview

Agentic implementation can leave machine-specific paths, prompt headers, transient build output, and misplaced process artifacts in a public repo. **Core principle:** identify and remove confirmed accidental artifacts while preserving intentional project content, operational safeguards, and useful evidence.

Language-agnostic: the playbook detects the build system instead of assuming one — Go, Rust, TypeScript, Python, or anything else.

**Announce at start:** "I'm using the dap-scrubbing-public-tree skill to clean the public tree."

## When to Use

- After a subagent-driven (or any agentic) implementation run, before opening the PR
- Before publishing a repo or branch that agents have worked in
- When a review flags "agent artifacts" or machine-specific content in the tree

**When NOT to use:** private scratch repos nobody will read; repos where process artifacts are intentionally public at their current paths (check the repo's docs conventions first).

## The Playbook

Work through these in order. Each step ends with a checkable result.

| # | Step | What to do | Done when |
|---|------|-----------|-----------|
| 1 | Inventory | Find candidate process artifacts and machine/agent traces in tracked files; classify each match by file, purpose, and whether it is accidental or intentional | Every candidate has a documented disposition; a search hit alone is not a defect |
| 2 | Relocate | Relocate only confirmed misplaced material, following the repository's documentation and audit conventions; repair references as part of the move | The move is justified, links resolve, and the original record is preserved at its intended destination |
| 3 | Scrub | Remove or generalize confirmed accidental headers and machine-specific paths; retain required identities, examples, test fixtures, policy, and audit references | Each changed match has a verified reason and the edit preserves its meaning |
| 4 | Remove transient artifacts | Remove only confirmed disposable build output or scratch material that is in scope and not an intentional fixture, tracker export, evidence record, or source-of-truth document | Each removed item has a recorded disposition and none is a listed record type (evidence, plan, review, decision, fixture, tracker export, source-of-truth document) |
| 5 | Dedupe | Merge duplicate CHANGELOG sections only when they are truly duplicate entries; preserve distinct release evidence | Changelog structure is clear and each distinct entry remains |
| 6 | Repair | Sweep for intra-repository references that step 2's inline repair missed after a justified move or rename | Every affected reference resolves; do not rewrite unrelated history |
| 7 | Verify | Detect the build system and run the repository-prescribed checks, including separately declared workspaces/modules and applicable publication guards | Required checks pass for the changed scope; known CI-only steps remain in CI |

## Implementation

**Candidate search patterns (step 1)** (extend per repo; these are common examples). Search tracked files so local scratch checkouts do not pollute results. Exclude this skill file from both searches below — it documents the identity string and patterns and would otherwise flag itself. That exclusion is file-level (git pathspecs cannot exclude individual lines), so it also hides this file's own illustrative examples. Compensate with a targeted self-scan of just this file — every hit must be a documented pattern line or an intentional example; anything else is a genuine candidate:
```bash
git grep -nE '/home/|/root/|/tmp/|/Users/' -- .agents/skills/dap-scrubbing-public-tree/SKILL.md || true
git grep -niE 'cloudbot|neo-green' -- .agents/skills/dap-scrubbing-public-tree/SKILL.md || true
```
Maintenance trap: if the skill is renamed or moved, the exclusion's hardcoded pathspec must move with it. Treat results as review candidates, not a list that must become empty:
```bash
# '|| true': grep returns exit status 1 on zero matches — a clean tree is the expected success, not a failure.
git grep -nE '/home/|/root/|/tmp/|/Users/' -- ':(exclude).agents/skills/dap-scrubbing-public-tree/SKILL.md' || true
git grep -niE 'cloudbot|neo-green' -- ':(exclude).agents/skills/dap-scrubbing-public-tree/SKILL.md' || true
```
Review each result in context. Required publisher identities and their enforcement code, tests, policy, and documentation are intentional and must remain intact. Preserve intentional tracker and fixture data, audit evidence, and indexed plans, reviews, and decisions. Do not make a match disappear by weakening a guard, deleting useful records, or renaming the identity it enforces. Record disposition when a search is part of a cleanup review.

**Disposition record (step 1):** for every search run that is part of a cleanup review, record one disposition per candidate in the PR description (or cleanup note), with a minimal shape — file, match, verdict (accidental/intentional + action), reason. This is the audit trail that satisfies step 1's "every candidate has a documented disposition" and lets a reviewer audit the classifications, not just the cleaned tree.

**Classifying required matches (steps 1 and 3):** to decide whether a match is required (intentional) or accidental, trace the match to an enforcement mechanism — a publish script, a CI job, a policy doc, or a test asserting the identity. If it enforces, it is required and stays; if it merely appears, investigate; if neither, it is a removal candidate. When in doubt, preserve: contested classifications default to preserve, and the disposition record captures the doubt for the reviewer.

**Scrub, don't just delete (step 3):** when a machine path carries meaning ("build output lives at /home/hatch/build/out"), generalize it to a neutral description ("the configured build output directory") rather than deleting the sentence — the information matters, the absolute path doesn't.

**Remove (step 4) and Dedupe (step 5):** no additional implementation guidance beyond the playbook table — step 4 removals are covered by the disposition record above, and step 5 is a manual merge judged by its "Done when". Do not invent extra procedure here.

**Build detection** (step 7) — first read the repository's contributor/agent instructions and CI workflows. Then inspect root and nested workspace/module manifests; a root command may not include nested modules. Run the repository-prescribed checks for every affected workspace and any required publisher or policy tests. If a required tool or check is unavailable, report it instead of calling the suite complete. Keep integration tools that the repository reserves for CI in CI; for this repository, Java and veraPDF run only in the isolated `validator-container` CI job.

For `document-accessibility-pipeline`, follow [`AGENTS.md`](../../../AGENTS.md). The following run locally: the root Go checks, the standalone `cmd/cloud-tools-bootstrap` checks, and the `ci/pdfua2/adapter` Go contract checks with `PDFUA_REAL_CLI_JAR=` (empty, so the Java CLI is skipped), plus `bash .github/scripts/test-source-publisher.sh` and its syntax check. The following stay in CI: the full Java/veraPDF validation in the `validator-container` job — do not provision or run Java/veraPDF locally. The publisher implementation, its tests, and the policies that pin CloudBot Neo-green as the author/committer identity are intentional matches and are not cleanup targets. The exact commands for the checks above are the repository's Go checks in the "Work and tests" section of AGENTS.md (root module, standalone `cmd/cloud-tools-bootstrap`, CI-only `ci/pdfua2/adapter` with `PDFUA_REAL_CLI_JAR=`, source-publisher guard tests).

For repositories without more specific instructions, use manifests as a starting point and guard unavailable tools with `command -v`:

| Manifest | Run |
|---|---|
| `Cargo.toml` | `command -v cargo >/dev/null && cargo test` |
| `go.mod` | `command -v go >/dev/null && go test ./...` |
| `package.json` | `command -v npm >/dev/null && npm test` |
| `pyproject.toml` / `requirements.txt` | `command -v pytest >/dev/null && pytest` |

Manifest detection is not root-only. Check explicitly declared nested modules and workspace members, then follow the repository's prescribed commands rather than assuming root tests cover them. If a required toolchain is missing, note the incomplete check; do not describe it as a full pass.

**Relocation (steps 2 and 6):** when a move is justified, preserve the record and update affected links; do not move or delete indexed plans, reviews, decisions, or audit evidence merely because they look like process artifacts. Apply the step 6 sweep to each justified move because references can break silently, including in ADRs and decision records.

## Common Mistakes

| Mistake | Why it hurts | Fix |
|---------|--------------|-----|
| Treating every search hit as an artifact | It can remove required publisher checks, tracker records, fixtures, or audit evidence | Classify matches and preserve intentional content |
| Assuming the root suite covers every module | Nested modules can be omitted from root test commands | Inspect repository instructions and run each prescribed module check |
| Moving evidence without checking references | Decisions and reviews may point at the original path | Relocate only when justified, then verify affected links |
| Deleting plans or audit records as clutter | These may be indexed source-of-truth records | Preserve them unless the repository's process explicitly authorizes removal |
| Weakening a guard to silence a match | It can break publication or policy enforcement | Keep enforcement intact and classify its identity strings as intentional |
| Skipping step 7 | Cleanup edits can break scripts, fixtures, or build checks | Run the repository-prescribed checks after changes |

## Real-World Impact

Applied to `go-afp-line` PR #15: process artifacts relocated to `docs/process/`, agent headers/machine paths/identity scrubbed, a nested spike module deleted, CHANGELOG deduped, and references rebased — with one fix round for dangling refs caught by verification. The PR merged with no hygiene findings.
