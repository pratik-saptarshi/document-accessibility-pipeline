# Bootstrap verification cost — findings (dap-5hy.3)

Spike: does the separate `cloud-tools-bootstrap --verify` pass run by
`scripts/bootstrap-cloud-build-tools.sh` after bootstrap create a meaningful
cold or warm setup cost, in checkout and no-checkout flows?

## Inputs

- Orchestrator: `cmd/cloud-tools-bootstrap` (stdlib-only Go module), built and
  run with host Go 1.27.2 on Linux x86_64 (2 CPUs).
- Script flow under test: `scripts/bootstrap-cloud-build-tools.sh` invokes
  bootstrap once, then `--verify` once (both shell branches do this).
- Measurement harness: `scripts/measure-bootstrap-verify-cost.sh`.
- Raw data: `docs/research/support/bootstrap-verification-cost-2026-10-17.txt`
  (8 cells × 5 reps; provenance header with date, Go version, uname, CPUs).

## Assumptions

- Synthetic manifest + local stub tools only; no real network. Stub tools are
  served over a loopback TLS server (throwaway self-signed CA, trusted via
  `SSL_CERT_FILE`) because manifest validation requires absolute HTTPS URLs.
- Stub tools (3 `raw`, 1 `archive`, plus the pinned `go` bootstrap-toolchain
  entry) carry `binary_sha256` plus `version_args`/`version_pattern`, so they
  exercise the same checksum + version-exec path (`installedOK`) the real
  `--verify` pass uses.
- "Checkout flow" = `go run` from the module checkout (mirrors the script's
  local-checkout branch); "no-checkout flow" = prebuilt binary (mirrors the
  immutable source-archive branch).
- "Cold" = empty cache + empty tools root; "warm" = tools installed and cache
  populated by one prior bootstrap (repeat-run scenario).
- Measured quantity is wall time of (bootstrap) vs (bootstrap then `--verify`),
  per the script's two-invocation sequence.

## Fixture-authored expectations

Bootstrap already runs `installedOK` (version exec + digest) for every tool —
once for already-installed tools, and again after each install. The separate
`--verify` pass repeats exactly that round. Expectation before measuring: the
verify pass adds roughly one `installedOK` round per tool — i.e. it doubles the
*verification* component, while bootstrap dominates total cost when installs
happen (cold). Results confirmed this.

## Results (median of 5 reps)

| flow | cache | bootstrap-only | bootstrap + verify | delta |
|---|---|---|---|---|
| checkout | cold | 0.213s | 0.341s | +0.128s (+60%) |
| checkout | warm | 0.159s | 0.309s | +0.151s (+95%) |
| no-checkout | cold | 0.051s | 0.067s | +0.016s (+31%) |
| no-checkout | warm | 0.017s | 0.032s | +0.016s (+92%) |

Absolute deltas are 15–150 ms with stub tools; relative deltas are +31% to
+95%. The checkout-flow delta is dominated by a second `go run` process spawn;
the no-checkout delta (~16 ms) is the pure second `installedOK` round.

## Confidence and ambiguity behavior

- High confidence: the verify pass costs ≈ one additional `installedOK` round
  per tool, in every cell. The mechanism is visible in code and the measured
  ratios match.
- Ambiguity (absolute numbers): stub tools are ~100-byte shell scripts.
  Real tools are multi-MB binaries, so the digest-hashing component scales
  with size, and real cold runs are dominated by network downloads — neither
  was measured here (network fetches are out of scope for this spike).
  Relative cost, however, transfers: in warm/no-checkout flows verification
  *is* the bootstrap, so `--verify` ~doubles it regardless of tool size.

## Unsupported boundaries

- Real tool archives, real network downloads, and `go-install` tools were not
  measured (explicitly out of scope: synthetic fixture only).
- Single host, single load condition; no concurrent-bootstrap lock-contention
  measurement; Linux/x86_64 only.
- The harness TLS trust is a throwaway self-signed CA — valid for local
  measurement only, not for production trust claims.

## Accuracy checks

- The harness aborts a cell on any nonzero orchestrator exit (no half-measured
  data); every recorded rep is `rc=0`.
- A `--verify` exit of 0 proves the checksum + version checks passed on the
  stubs, so measured time includes real digest hashing and real version execs.
- Guard: the harness refuses to run unless the manifest is the synthetic
  fixture and every fetched tool source is `https://127.0.0.1/...`; `go-install`
  kinds are rejected.
- All measurement state under a `chmod 700` temp dir; nothing installed
  system-wide; main checkout untouched.

## Missing, multiple, conflicting, and unsafe controls

- Missing control: real-tool timings (needs network; not permitted here).
- Multiple tools per kind were exercised (3 raw, 1 archive) but not the
  `go-install` kind; stub execs are shell scripts, whose spawn cost differs
  from native ELF binaries.
- No conflicting results: all 8 cells show the same directional effect.
- Unsafe-control note: none of the measured runs touched the real manifest,
  real network, or shared state; the `bd` CLI was never invoked in the
  worktree.

## Recommendation

This spike confirms the repeated work is real but small in absolute terms:
the separate `--verify` pass adds one `installedOK` round (~16 ms pure
orchestrator cost; ~0.13–0.15 s via `go run` in the checkout flow). Whether
that is "meaningful" depends on the setup context: in cold real-world runs it
is noise against multi-MB downloads; in warm/no-checkout repeats it roughly
doubles the (already small) bootstrap time.

No automatic behavior change is proposed. If the cost is judged meaningful
for a specific context, the candidate decision for product review is to skip
the separate `--verify` invocation where bootstrap itself just verified
(warm/no-checkout), while keeping it for cold checkouts — to be recorded as a
decision (Proposed/Open) per the repo's decision process, not implemented
unilaterally. This spike does not authorize production writers,
reviewer-authority changes, or any change to the orchestrator.
