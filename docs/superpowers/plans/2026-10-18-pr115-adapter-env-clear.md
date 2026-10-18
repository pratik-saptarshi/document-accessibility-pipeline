# PR #115 Adapter Test Environment Follow-Up

## Goal and scope

Port the one useful workflow correction remaining on old PR #115: prevent an inherited `PDFUA_REAL_CLI_JAR` value from activating the real veraPDF integration test in the Go checks job. The exact base is current `main` at `4736a75610ca4b3274c5686db91999c29917099c`. Work on a separate branch; preserve the isolated veraPDF CI job and do not edit PR #115 or `.beads/issues.jsonl`.

## Acceptance criteria

1. The adapter module gate in `.github/workflows/go-checks.yml` explicitly clears `PDFUA_REAL_CLI_JAR` before `go test -race ./... -count=1`.
2. `requiredGoChecksModuleGates` expects that exact command, and a regression proves the validator rejects the same step with the clearing prefix removed.
3. Confirm adapter skip semantics from `TestValidateWithAuthenticatedVeraPDFAndRealUA2Controls`: `os.Getenv` maps both unset and explicitly empty `PDFUA_REAL_CLI_JAR` to `""`, so both skip the authenticated integration test. Run the focused test with an inherited non-empty sentinel jar value and the workflow's `env PDFUA_REAL_CLI_JAR=` prefix; verify it skips before Java lookup/invocation using a test-only Java sentinel. Never run the real CLI locally.
4. Keep the existing strict policy coverage and the isolated Java/veraPDF CI boundary intact. Do not transplant the old PR #115 policy snapshot or stale Beads closure row.
5. Pass required root and nested-module checks, traceability, shell syntax, publisher tests, hygiene, independent exact-range review, and hosted required checks on the final PR head.

## Test-first sequence

1. Pin current `main`; inspect PR #115's remaining diff and adapter skip logic.
2. Add a policy regression that mutates the adapter gate back to the uncleared command. On the base, `go test ./internal/devsecops -run '^TestGoChecksAdapterGateClearsRealCLIEnvironment$' -count=1` failed because the validator accepted it.
3. Update the expected command and workflow together; confirm the regression and existing gate tests pass. Verify that unset and explicit-empty values skip and an inherited non-empty sentinel is cleared without invoking the Java sentinel.
4. Write this plan into the plan index, run prescribed gates, scrub the public tree, and request independent review of the exact base/head.
5. Publish a separate draft PR with the guarded HTTPS source publisher. Wait for independent review and exact-head CI; report the PR to the owner and wait for their review relay before any protected squash merge.

## Disposition boundary

Current `main` already closes `dap-jo9.5.6` with evidence for all nine module gates, and its policy tests supersede the old PR's weaker snapshot. Do not reapply that issue row. This follow-up addresses only the adapter environment-clear behavior; do not close PR #115 or resolve its review thread here.

## Verification evidence

- RED on the exact base: `TestGoChecksAdapterGateClearsRealCLIEnvironment` failed because the validator accepted the uncleared command.
- GREEN after the workflow and expected gate were changed: the focused policy regression passed.
- Adapter skip probe passed with the variable unset, explicitly empty, and inherited as a non-empty sentinel then cleared with the workflow prefix. A test-only `java` sentinel was not invoked.
- Root gofmt, tidy diff, vet, race suite, traceability; bootstrap race/vet/tidy and shell syntax; adapter race/vet/tidy; source-publisher syntax/regression; and `git diff --check` passed. Traceability reports 17 requirements, 6 stories, 21 tasks, and 26 specification-only cases.
- The generated, untracked Fontconfig cache created by local tests was removed after confirming the worktree was clean before test execution; no source, fixture, or Beads data was removed or changed.
