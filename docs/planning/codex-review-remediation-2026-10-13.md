# Codex review remediation plan

**Plan date:** 2026-10-13 UTC
**Audit cutoff:** `9dde9846074c727b50e3d128ab59d2f94137d072`
**Planning branch starting point:** `95718bd0df0d4ed4357910c1988b7931617aef60` (PR #26 merge). Synchronized with current `origin/main` at `b85e7a48811b62b10e1f759c04b98cbb0af97286` using `scripts/source-publisher.sh merge-main`; local sync commit `1756fb9f1e29580a6cef6fed3c72a6d3a6b8eac8`.

## Audit boundary and disposition

The source audit covers merged PRs #1–14 and #16–25 at the exact cutoff above. It recorded 16 actionable Codex review comments: eight addressed, seven unaddressed, and one partial. CRR-006 (PR6 publication wording) was partial; CRR-001–005, CRR-007, and CRR-008 were unaddressed at that cutoff. This is a historical classification, not the current task count. PR #26 and PR #27 merged after that audit cutoff and are recorded as subsequent evidence without changing the baseline.

The [post-PR14 adversarial panel](../security/adversarial-panel-2026-10-12-a6d6fc6.md) records the SHA-pinned five-role review and tool evidence. Tracked remediation requirements are in [review-audit-remediation.md](../requirements/review-audit-remediation.md); detailed audit links and acceptance criteria live there and in the linked Beads issues.

## Execution order at the audit cutoff

1. **Close verified stale records and repair report facts.** `.5.5` meets its recorded criteria and is closed with PR #7 plus focused local test evidence. `.5.6` remains open: `TestValidateNormalizesRelativeTempDir` confirms absolute `-jar`/input arguments and cleanup, but the acceptance criterion explicitly calls for a path-aware fake Java process to resolve both arguments. The PR #6 report’s dated correction now gives PR #7’s actual merge state, the historical issue count, and PR #6’s later merge state while preserving the `595e7e9` source review pin and findings.
2. **Restore security and activation invariants.** Extend `dap-jo9.8` to cover all three Go modules in test-inclusive `govulncheck` policy and mutation negatives. Add `dap-5hy.4` for activation path selection, with alternate `--go` and default-path tests.
3. **Make the pinned bootstrap source durable.** `dap-5hy.5` must preserve source availability without rewriting history and prove fresh no-checkout installation plus exact digest, missing-source, and tamper behavior.
4. **Bind validator version evidence.** `dap-jo9.2.5.8` is accepted in PR #33. For the pinned veraPDF 1.30.3 XML contract, the parser binds the `<buildInformation><releaseDetails id="apps" version="…"/></buildInformation>` value to the verified artifact and recorded result. Synthetic versionless reports are rejected; no alternate report-format support is claimed.
5. **Record PR #27 as addressed after the cutoff.** Its CloudBot source commit `b4595e45cc82111bc2833aec6b0061bf282aaf94` added nested tidy checks; synchronized head `7af937c5659b4a4e441320d1acb042d71077e4a5` passed all nine checks. Merge commit `b85e7a48811b62b10e1f759c04b98cbb0af97286` corresponds to tested tree `3e894e876b155bdbb735530f41ebd55e1dad8493`; all seven post-merge checks passed (`pdfua2`, `Analyze (actions)`, `Analyze (go)`, `validator-container`, `go-security`, `checks`, `secrets`). The integration/sync commit identity is Pratik/GitHub, not CloudBot. This finding was unaddressed at the audit cutoff and has since been addressed; no duplicate task is needed.

## Current disposition after PR #33 (2026-10-13)

- **Accepted and closed:** the scanner policy (`dap-jo9.8`) merged in PR #29 at `97feb71dfa931400a25ba2fe465614c7a08fa96c`; the child-path resolver (`dap-jo9.2.5.6`) in PR #30 at `23b11d34d6654685f47e43853ef0d077350a5183`; selected-Go activation (`dap-5hy.4`) in PR #31 at `63126024666fd71b48c25df55d1e98924e5eb09a`; and XML version provenance (`dap-jo9.2.5.8`) in PR #33 at `c3d2f678d77621afcc415323ebf2f82d101054cf`. Each merged implementation passed nine required pre-merge checks and seven post-merge checks. PR #33 also passed final Codex review and ran the real pinned veraPDF controls in CI.
- **Still pending:** `dap-5hy.5` remains in progress while PR #34 is open. Its source archive pin is the immutable protected-main commit `63126024666fd71b48c25df55d1e98924e5eb09a`, with downloaded-byte SHA-256 `6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5` and exact size 306,734 bytes. PR #34 is synchronized with main `c3d2f678d77621afcc415323ebf2f82d101054cf` at head `5090a772f690225a93bd1b2bdc54f329ba9d6f22`; acceptance remains pending until PR #34 actually merges. The prior `8f0072c` archive pin is historical context only; it remained readable and no outage was asserted.

## Risk and assurance limits

The findings describe security-policy coverage, tool path selection, durable source availability, XML version provenance, and report/tracker accuracy. They do not establish an active exploit or current download outage. The NeuroStrata model-download limitation is accepted and not reopened by this plan; do not retry it or alter network settings. NeuroPlasticity requires digest verification; never use its unsafe version flag. Do not claim model-assisted code review or full Neuro execution without evidence. No private documents, credentials, repository uploads, model downloads, persistent MCP registration, container-isolation weakening, or unrelated fixes are part of this plan.

Current CI enforcement requires all nine checks and enforces no-bypass rules; the original seven remain and `pdfua2` plus `validator-container` are included. The branch-ruleset investigation is resolved, with no current permission blocker. PR #26 merge `95718bd0df0d4ed4357910c1988b7931617aef60` and PR #27 merge `b85e7a48811b62b10e1f759c04b98cbb0af97286` each have seven completed successful post-merge checks; PR #27’s synchronized head also passed all nine checks. The ruleset verification is recorded in the [post-PR14 adversarial panel](../security/adversarial-panel-2026-10-12-a6d6fc6.md).

Do not merge without independent review and required checks. Keep PDF/UA-2 machine validation separate from semantic correctness and assistive-technology review.
