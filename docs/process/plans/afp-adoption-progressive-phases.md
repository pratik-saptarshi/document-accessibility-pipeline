# AFP adoption — progressive enhancement phases

Product-owner GO on `dap-jo9.4.2` (2026-10-22): adopt go-afp-line @ `492c4c7`
(pseudo-version `v0.1.0-beta.1.0.20261022163835-492c4c7d35b9`), enhance
progressively. Basis: Decision 0021 (DEFER on the beta) + Decision 0022
(CONDITIONAL on 492c4c7). The 5 adoption conditions in 0022 become the phase
gates. Each phase lands as its own PR; no phase starts until the previous
phase's done criteria are verified.

## Phase 1 — Pin (done)

- Entry: owner GO.
- Work: require the exact pseudo-version in `go.mod`, `go.sum` checksums,
  `internal/afp` anchor + `TestPinnedRevision` (0022 condition 1 gate).
- Verification gate: `go mod tidy -diff` clean, `go build ./...`,
  `go test -race ./...`, `go mod verify`, traceability-check valid.
- Done: merged to main; bead closure for `dap-jo9.4.2` follows once its AC
  chain is satisfied (blocked: `dap-jo9.4.1.2` still open; Decision 0022
  pending merge via PR #186).

## Phase 2 — Candidate suite at the resolved x/sys (0022 condition 2)

- Entry: Phase 1 merged.
- Work: run the go-afp-line test suite at the MVS-resolved x/sys v0.49.0
  (candidate was tested at v0.48.0). Record results; if green, no code
  change needed in DAP.
- Verification gate: candidate suite green at v0.49.0, or a written risk
  acceptance if the owner waives the run.
- Done: evidence recorded in the phase PR; condition 2 discharged.

## Phase 3 — Typed loud-error propagation (0022 condition 3)

- Entry: Phase 2 done.
- Work: wire `Fields→Reduce→Render` through DAP's AFP lane so 0x80-class
  aborts surface as typed `afp.DecodeError` (errors.As-able), never
  swallowed or downgraded to silent skips. Per owner ruling 2026-10-22,
  rejection of afplib-only PTOCA codes is expected behavior — the
  diagnostic must stay loud.
- Verification gate: failing test first — feed the 0020 profile file,
  assert the typed diagnostic names code, reason, field, and kind;
  `go test -race`, vet, gofmt clean.
- Done: merged; condition 3 discharged.

## Phase 4 — Fidelity measurement (0022 condition 4, stays open)

- Entry: a convertible input exists (today: 0/2116 reference characters
  decode before the accepted 0x80 stop — not measurable by design).
- Work: AFP→PDF fidelity comparison vs the 0020 trusted references once an
  input converts end-to-end. Do not invent pass/fail thresholds (tolerances
  open per 0020) — report measured deltas.
- Verification gate: measured delta recorded; thresholds decided by the
  product owner if/when needed.
- Done: condition 4 discharged or explicitly re-deferred by owner decision.

## Phase 5 — Per-version license re-read hook (0022 condition 5)

- Entry: Phase 1 merged.
- Work: document the standing rule — every go-afp-line version bump gets a
  license re-read of the adopted graph (same 21-module discipline as 0022
  §g) before the pin moves. Add the check to the lane's PR checklist.
- Verification gate: rule recorded in the AFP lane docs; next bump
  demonstrates it.
- Done: condition 5 operational.
