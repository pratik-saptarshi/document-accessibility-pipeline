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
- Raw data: `docs/research/support/bootstrap-verification-cost-2026-10-18.txt`
  (8 cells × 5 reps; provenance header with date, Go version, uname, CPUs,
  and `harness_sha256` identifying the exact harness revision that produced
  it).

## Assumptions

- Synthetic manifest + local stub tools only; no real network. Stub tools are
  served over a loopback TLS server (throwaway self-signed CA, trusted via
  `SSL_CERT_FILE`) because manifest validation requires absolute HTTPS URLs.
- Stub tools (3 `raw`, 1 `archive`, plus the pinned `go` bootstrap-toolchain
  entry) carry `binary_sha256` plus `version_args`/`version_pattern`, so they
  exercise the same checksum + version-exec path (`installedOK`) the real
  `--verify` pass uses.
- "Checkout flow" = `go run` from the module checkout (mirrors the script's
  local-checkout branch); "no-checkout flow" = extract + rebuild the pinned
  source archive per invocation inside the timer, then run the built binary
  (mirrors the script's immutable source-archive branch, including its
  sha256/size pin verification — the harness verifies the archive against
  the manifest pin at fixture setup and fails closed on mismatch).
- "Cold" = empty cache + empty tools root, with GOCACHE/GOMODCACHE scoped to
  the arm's cache dir (the script's `cache_root` semantics), so cold arms
  compile from scratch; "warm" = tools installed and cache populated by one
  prior bootstrap (repeat-run scenario), reusing the seeded go caches.
- Measured quantity is wall time of (bootstrap) vs (bootstrap then `--verify`)
  per the script's two-invocation sequence; in the no-checkout flow the
  extract+rebuild happens once per arm inside the timer in both modes, so
  the delta isolates the `--verify` pass cost.
- Methodology revision 2026-10-18: the first harness revision measured the
  no-checkout cells with a single prebuilt binary (an orchestrator
  microbenchmark that omitted the per-invocation archive extract+rebuild the
  real script performs). The +31%/+92% no-checkout figures and the
  "roughly doubles" claim below were computed against those microbenchmark
  baselines and are superseded. New measurements below use the per-arm
  extract+build flow; see the revision note under Results.

## Fixture-authored expectations

Bootstrap already runs `installedOK` (version exec + digest) for every tool —
once for already-installed tools, and again after each install. The separate
`--verify` pass repeats exactly that round. Expectation before measuring: the
verify pass adds roughly one `installedOK` round per tool — i.e. it doubles the
*verification* component, while bootstrap dominates total cost when installs
happen (cold). Results confirmed this.

## Results (median of 5 reps)

Regenerated 2026-10-18 with the per-arm extract+build harness (evidence
`docs/research/support/bootstrap-verification-cost-2026-10-18.txt`,
`harness_sha256=24d629ac992909770ce55fb169b6196e9907b47234e7b2a57d431bd143c0cd4c`).
The earlier table (prebuilt-binary microbenchmark baselines, +31%/+92%
no-checkout) is superseded and removed.

| flow | cache | bootstrap-only | bootstrap + verify | delta |
|---|---|---|---|---|
| checkout | cold | 38.879s | 36.191s | −2.688s (−6.9%, within run variance) |
| checkout | warm | 0.147s | 0.361s | +0.214s (+146%) |
| no-checkout | cold | 38.336s | 37.604s | −0.732s (−1.9%, within run variance) |
| no-checkout | warm | 0.185s | 0.200s | +0.015s (+7.9%) |

Reading the table honestly:

- Cold cells (~38 s) are dominated by compiling the orchestrator from an
  empty go build cache — the real script's cold behavior, now measured
  instead of masked. Against that baseline the `--verify` delta is noise
  (negative medians, smaller than the rep spread of ~35–42 s): the verify
  pass cost is undetectable in cold flows.
- Warm checkout: `--verify` adds +0.21 s, dominated by the second `go run`
  process spawn — not by verification work itself.
- Warm no-checkout: `--verify` adds ~15 ms (+7.9%), the pure second
  `installedOK` round (checksum + version execs) on the stub tools. This is
  the same ~16 ms the microbenchmark measured; what changed is the baseline
  it is compared against (0.185 s with per-invocation extract+build, not
  0.017 s). The "roughly doubles" claim is withdrawn: with the real
  per-invocation build in the baseline, the separate verify pass is under
  a tenth of the no-checkout flow, not double it.

Absolute deltas are 15–150 ms with stub tools; relative deltas are +31% to
+95%. The checkout-flow delta is dominated by a second `go run` process spawn;
the no-checkout delta (~16 ms) is the pure second `installedOK` round.

## Confidence and ambiguity behavior

- High confidence: the verify pass costs ≈ one additional `installedOK` round
  per tool (checksum + version execs). The mechanism is visible in code, and
  the warm no-checkout delta (~15 ms) matches the microbenchmark's ~16 ms —
  the round itself did not change; the baseline it is compared against did.
- The cold-flow deltas are noise, not signal: both cold deltas are negative
  and smaller than the rep spread (~35–42 s), because cold arms compile the
  orchestrator from an empty go build cache (~38 s). The honest cold-flow
  claim is "undetectable against the compile-dominated baseline", not a
  signed delta.
- Ambiguity (absolute numbers): stub tools are ~100-byte shell scripts.
  Real tools are multi-MB binaries, so the digest-hashing component scales
  with size, and real cold runs are dominated by network downloads — neither
  was measured here (network fetches are out of scope for this spike).
  The earlier "verification *is* the bootstrap, so `--verify` ~doubles it"
  transfer argument is withdrawn: with the real per-invocation extract+build
  in the baseline, the warm no-checkout verify pass is +7.9%, not ~double.

## Unsupported boundaries

- Real tool archives, real network downloads, and `go-install` tools were not
  measured (explicitly out of scope: synthetic fixture only).
- Single host, single load condition; no concurrent-bootstrap lock-contention
  measurement; Linux/x86_64 only.
- The harness TLS trust is a throwaway self-signed CA — valid for local
  measurement only, not for production trust claims.

## Accuracy checks

- The raw evidence header records `harness_sha256` (the sha256 of
  `scripts/measure-bootstrap-verify-cost.sh` as committed at HEAD when the
  evidence was generated), so future harness drift is mechanically
  detectable; the 2026-10-17 evidence file (different harness revision) was
  deleted.
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

Recomputed 2026-10-18 against the per-arm extract+build baselines: the
separate `--verify` pass is undetectable in cold flows (noise against the
~38 s compile-dominated baseline), adds +0.21 s in warm checkout (second
`go run` spawn), and adds ~15 ms (+7.9%) in warm no-checkout (the pure
`installedOK` round). The earlier "roughly doubles" framing was an artifact
of microbenchmark baselines and is withdrawn.

No automatic behavior change is proposed. Whether the cost is "meaningful"
depends on context: in cold real-world runs it is noise against compiles and
multi-MB downloads; in warm repeats it is milliseconds against a
sub-second flow. If the cost is judged meaningful for a specific context, the
candidate decision for product review is to skip the separate `--verify`
invocation where bootstrap itself just verified (warm/no-checkout), while
keeping it for cold checkouts — to be recorded as a decision (Proposed/Open)
per the repo's decision process, not implemented unilaterally. This spike
does not authorize production writers, reviewer-authority changes, or any
change to the orchestrator.

No automatic behavior change is proposed. If the cost is judged meaningful
for a specific context, the candidate decision for product review is to skip
the separate `--verify` invocation where bootstrap itself just verified
(warm/no-checkout), while keeping it for cold checkouts — to be recorded as a
decision (Proposed/Open) per the repo's decision process, not implemented
unilaterally. This spike does not authorize production writers,
reviewer-authority changes, or any change to the orchestrator.
