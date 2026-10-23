# AFP adoption — progressive enhancement phases

Product-owner GO on `dap-jo9.4.2` (2026-10-22): adopt go-afp-line @ `492c4c7`
(pseudo-version `v0.1.0-beta.1.0.20261022163835-492c4c7d35b9`), enhance
progressively. Basis: Decision 0021 (DEFER on the beta) + Decision 0022
(CONDITIONAL on 492c4c7, six conditions).

**2026-10-23: this plan was run through the planning-wringer** (adversarial
panel vs ground truth, paperthin refine, independent review). The wringer
reshaped it: Phase 2 became an evidence run, Phase 3 was killed as a build
phase, Phase 4 became a standing watch, Phase 5 became a small docs edit.
Current conditions status lives in
`docs/reviews/go-afp-line-conditions-2-5-status.md` (PR #188) — this plan is
the historical record of intent, not the status source.

## Phase 1 — Pin (done)

- Entry: owner GO.
- Work: require the exact pseudo-version in `go.mod`, `go.sum` checksums,
  `internal/afp` anchor + `TestPinnedRevision` (0022 condition 1 gate).
- Verification gate: `go mod tidy -diff` clean, `go build ./...`,
  `go test -race ./...`, `go mod verify`, traceability-check valid.
- Done: PR #187 (in flight at time of writing).

## Phase 2 — Candidate suite at the resolved x/sys (0022 condition 2; DISCHARGED)

- Wringer verdict: reshaped to an evidence run, not a phase PR — go-afp-line
  carries zero x/sys references (transitive-only delta), so a standalone
  phase PR was process overhead.
- Evidence: suite re-run at x/sys v0.49.0 in scratch (`git archive 492c4c7`,
  `go mod edit -require`, `GOPROXY=off`): 5/5 packages ok, `go list -m`
  confirms v0.49.0. Recorded in `docs/reviews/go-afp-line-conditions-2-5-status.md`.

## Phase 3 — Typed loud-error propagation (0022 condition 3; COVERED, build phase killed)

- Wringer verdict: killed as a build phase. No DAP lane production code
  exists; the planned gate duplicated the condition-6 integration test
  assertion-for-assertion; new plumbing would violate dap-jo9.4.5's AC
  ("Do not begin implementation until the owner confirms profile, audience,
  and delivery scope").
- Condition 3 is covered by the condition-6 test (`internal/afp/integration_test.go`:
  profile aborts Reduce with a loud typed `afp.DecodeError` naming 0x80 —
  the owner-accepted expected behavior, ruling 2026-10-22). The gate
  transfers to the future lane's tests when lane code is built.

## Phase 4 — Fidelity measurement (0022 condition 4; standing watch)

- Wringer verdict: not a phase — entry may never fire. Block (0/2116
  reference characters decode before the accepted 0x80 stop) and deferral
  already recorded in 0022 §f/§k.
- Watch trigger: a *reference-backed* input converts end-to-end. Then:
  AFP→PDF fidelity comparison vs the 0020 trusted references; report
  measured deltas, never invent pass/fail thresholds (tolerances open per
  0020); thresholds decided by the product owner if/when needed.

## Phase 5 — Per-version license re-read hook (0022 condition 5; DONE)

- Wringer verdict: reshaped to a small docs edit (no "lane PR checklist"
  exists). Rule recorded in the AFP refinement record and CONTRIBUTING.md
  (PR #188): every go-afp-line version bump — and any go.mod change
  touching the AFP subtree — gets a license re-read of the adopted graph
  before the pin moves. Recorded as a new rule with owner sign-off, not
  0022-derived.
