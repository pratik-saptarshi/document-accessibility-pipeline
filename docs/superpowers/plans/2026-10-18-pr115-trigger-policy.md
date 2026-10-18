# PR #115 Trigger-Policy Follow-Up

## Goal and scope

Address unresolved review finding [#4303867227](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/115#discussion_r4303867227): the module-gate validator accepts `pull_request_target` because it ignores the workflow `on` map, while checkout has no explicit `ref`. The follow-up is based on current `main` at `63ed6b0d1d4fa2b09b8b949bae7d3eaeb04bbe4c` and will use a separate branch. PR #115 remains untouched; its auto-merge and conflict state are not changed here.

This is a policy-test change only. Preserve PR #136's strict workflow parser and existing module-gate tests. Do not edit workflow files, Beads/tracker state, product code, issue status, or other open PRs. Do not interpret this source review as product acceptance.

## Acceptance criteria

1. `validateScannerIsolation` validates the Go checks workflow's trigger map.
2. The workflow must define `pull_request` with only the `main` branch filter. It must reject `pull_request_target` both when that event replaces `pull_request` and when it is added alongside it; explicit activity-type or path filters are also rejected so checks cannot be narrowed away from PR updates.
3. Regressions exercise the actual strict Go-checks parser and policy validator for a missing `pull_request`, a `pull_request_target` replacement, a coexisting `pull_request_target`, a `pull_request` branch filter changed to `release`, a closed-only activity filter, a path filter, and a catch-all `paths-ignore` filter.
4. The unmodified validator rejects none of those seven mutations (RED); the implementation rejects all seven while the unmodified `go-checks.yml` continues to pass (GREEN).
5. Existing parser-fidelity and module-gate behavior remains intact. Required local root and nested-module checks, traceability, shell syntax, and source-publisher regression checks pass on the final tree.
6. An independent plan review and an independent exact-head code review report no blockers. Hosted checks must pass on the exact published head before protected squash merge.

## Test-first sequence and evidence

1. Pin current `main` and inspect the complete PR #115 thread history, including resolved threads. The unresolved trigger finding is the bounded task. This follow-up does not claim the complete old PR diff is superseded; its current-main delta includes separate workflow, tracker-export, and policy-test changes, which require an independent disposition.
2. Add trigger mutation regressions to `internal/devsecops/policy_test.go`, using `parseGoChecksWorkflow` and `validateScannerIsolation`. The first four tests failed on unmodified base `63ed6b0`; the later three filter tests were also run against that same unmodified base and failed because the validator accepted them.
3. Add the narrow trigger validator: reject any `pull_request_target`, require a `pull_request` mapping with no filter keys beyond `branches`, and require its sole branch filter to be `main`. All seven tests then pass; the unmodified workflow and complete `internal/devsecops` package pass.
4. On code-only commit `ff80bcb6ff8622be2223581b3799aadc2f8e26d8`, root formatting, module tidy diff, vet, race tests, and traceability passed. Bootstrap and adapter race/vet/tidy checks, bootstrap shell syntax, publisher syntax/regression checks, and `git diff --check` passed. The root race run used task-scoped writable `HOME` and cache paths after the initial environment-default config path proved read-only. These results describe that code-only revision. The later activity/path-filter controls were confirmed RED on the unmodified base before the validator was tightened; rerun the full applicable checks on the eventual final candidate.
5. Obtain independent review of this plan and of the final exact candidate. Publish the separate branch only through the guarded HTTPS source publisher, open a follow-up PR referencing #115 and the finding, inspect its comments/threads, and wait for all required exact-head CI checks before protected squash merge.

## Publication and merge boundary

No publication, PR update, auto-merge change, ready-for-review action, merge, thread resolution, Beads operation, or issue closure has occurred in this follow-up checkpoint. Publication and protected merge are authorized only for the separately reviewed follow-up. PR #115's original branch and review thread remain as-is; after the follow-up merges, separately determine whether its original diff is fully superseded before proposing any disposition of #115.
