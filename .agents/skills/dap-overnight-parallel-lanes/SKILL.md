---
name: dap-overnight-parallel-lanes
description: Use when explicitly requested parallel work can be split into independent lanes with deterministic validation and review.
---

# Parallel work lanes

Use parallel lanes only when the user or governing instructions request delegation and the work can be separated by files, interfaces, and acceptance checks. Do not parallelize tightly coupled changes or work without a useful validation gate.

## Set up the run

1. **Split the work.** Give each lane one bounded deliverable, an owner, file boundaries, an acceptance test, and prerequisite issue IDs when tracked. Explain why lanes are independent. Keep shared interfaces and integration work with the coordinator.
2. **Isolate each lane.** Create separate worktrees from one verified base commit. Check the worktree path and repository instructions before dispatching. Workers must not share a working tree.
3. **Record limits before dispatch.** In the run ledger, record lane scope, owner, expected artifact, validation commands, a real elapsed-time limit, and a review/fix-round limit. Use a task-appropriate clock limit (one hour per lane by default) and at most three fix rounds unless the user sets another bound. Do not invent spending, token, or cost measurements; use only actual metering or a user-provided budget.
4. **Dispatch within actual runtime capabilities.** Assign work based on available agents, tools, and task needs. Do not promise model tiers or runtime limits the environment cannot enforce. Workers stay within their file boundaries and do not dispatch further agents unless authorized.

## Execute and integrate

- Run deterministic tests and checks before requesting review. Give each lane a fresh independent reviewer; the coordinator does not review its own implementation. A focused review is sufficient unless a full panel is requested.
- Keep lane commits local unless the user or coordinator explicitly authorizes publishing. The coordinator owns branch publication, PR creation, integration, and any protected squash merge covered by current authorization. Use the repository's HTTPS source-publisher guard, CloudBot identity, and branch protections.
- Preserve incomplete work when a time or review limit is reached. Mark it **partial** or **blocked**, list what passed and what did not, and do not describe it as validated. Stop or re-scope instead of extending a limit silently.
- Integrate sequentially enough to resolve overlap, then validate the assembled branch. Monitor every required check until it finishes on the exact final head. Do not cap CI waiting at one run; if the head changes after checks start, require fresh checks for that head.
- Read every review comment across revisions and resolve findings or record an evidence-based disposition. Never merge around a failing check or unresolved required review.

## Ledger and resume

The ledger is coordination evidence, not the source of tracker state. Link each lane to its canonical Beads issue when one exists, but do not use the ledger to replace Beads, infer status, or close issues. On resume, verify the current branch heads, worktree contents, issue states, review comments, and check runs. Re-run checks that are missing, stale, or tied to a different head; preserve partial work without treating previous progress notes as proof.

Use elapsed time and completed review rounds as stop conditions. If a lane reaches its limit, record the blocker and hand control to the coordinator. Only the coordinator continues, publishes, or merges when the existing user authorization covers that action.
