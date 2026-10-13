# Codex review remediation plan

**Plan date:** 2026-10-13 UTC
**Audit cutoff:** `9dde9846074c727b50e3d128ab59d2f94137d072`
**Planning branch starting point:** `95718bd0df0d4ed4357910c1988b7931617aef60` (PR #26 merge). Synchronized with current `origin/main` at `b85e7a48811b62b10e1f759c04b98cbb0af97286` using `scripts/source-publisher.sh merge-main`; local sync commit `1756fb9f1e29580a6cef6fed3c72a6d3a6b8eac8`.

## Audit boundary and disposition

The source audit covers merged PRs #1–14 and #16–25 at the exact cutoff above. It recorded 16 actionable Codex review comments: eight addressed, seven unaddressed, and one partial. This is a historical classification, not the current task count. PR #26 and PR #27 merged after that audit cutoff and are recorded as subsequent evidence without changing the baseline.

The [post-PR14 adversarial panel](../security/adversarial-panel-2026-10-12-a6d6fc6.md) records the SHA-pinned five-role review and tool evidence. Tracked remediation requirements are in [review-audit-remediation.md](../requirements/review-audit-remediation.md); detailed audit links and acceptance criteria live there and in the linked Beads issues.

## Execution order

1. **Close verified stale records and repair report facts.** `.5.5` meets its recorded criteria and is closed with PR #7 plus focused local test evidence. `.5.6` remains open: `TestValidateNormalizesRelativeTempDir` confirms absolute `-jar`/input arguments and cleanup, but the acceptance criterion explicitly calls for a path-aware fake Java process to resolve both arguments. The PR #6 report’s dated correction now gives PR #7’s actual merge state, the historical issue count, and PR #6’s later merge state while preserving the `595e7e9` source review pin and findings.
2. **Restore security and activation invariants.** Extend `dap-jo9.8` to cover all three Go modules in test-inclusive `govulncheck` policy and mutation negatives. Add `dap-5hy.4` for activation path selection, with alternate `--go` and default-path tests.
3. **Make the pinned bootstrap source durable.** `dap-5hy.5` must preserve source availability without rewriting history and prove fresh no-checkout installation plus exact digest, missing-source, and tamper behavior.
4. **Bind validator version evidence.** `dap-jo9.2.5.8` defines an explicit artifact/report version contract and negative tests for missing, mismatched, and malformed evidence. Do not presume a common version field across report schemas.
5. **Record PR #27 as addressed after the cutoff.** Its CloudBot source commit `b4595e45cc82111bc2833aec6b0061bf282aaf94` added nested tidy checks; synchronized head `7af937c5659b4a4e441320d1acb042d71077e4a5` passed all nine checks. Merge commit `b85e7a48811b62b10e1f759c04b98cbb0af97286` corresponds to tested tree `3e894e876b155bdbb735530f41ebd55e1dad8493`; all seven post-merge checks passed (`pdfua2`, `Analyze (actions)`, `Analyze (go)`, `validator-container`, `go-security`, `checks`, `secrets`). The integration/sync commit identity is Pratik/GitHub, not CloudBot. This partial audit item is now addressed; no duplicate task is needed.

## Risk and assurance limits

The findings describe security-policy coverage, tool path selection, durable source availability, XML version provenance, and report/tracker accuracy. They do not establish an active exploit or current download outage. The blocked NeuroStrata model download remains blocked; do not retry it or alter network settings. NeuroPlasticity requires digest verification; never use its unsafe version flag. Do not claim model-assisted code review or full Neuro execution without evidence. No private documents, credentials, repository uploads, model downloads, persistent MCP registration, container-isolation weakening, or unrelated fixes are part of this plan.

The separate ruleset-permission investigation remains outside this audit plan: retain the existing seven required main checks; `pdfua2` and `validator-container` have passed CI but GitHub rejected making them mandatory with `Unauthorized`. Do not change rulesets or retry permission changes here.

Do not merge without independent review and required checks. Keep PDF/UA-2 machine validation separate from semantic correctness and assistive-technology review.
