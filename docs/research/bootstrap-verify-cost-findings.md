# Bootstrap verification cost — findings (dap-5hy.3)

Spike: does the separate `cloud-tools-bootstrap --verify` pass run by
`scripts/bootstrap-cloud-build-tools.sh` after bootstrap create a meaningful
cold or warm setup cost, in checkout and no-checkout flows?

## Inputs

- Orchestrator: `cmd/cloud-tools-bootstrap` (stdlib-only Go module), built and
  run with host Go 1.27.2 on Linux x86_64 (5 CPUs; Go default GOMAXPROCS=5).
- Script flow under test: `scripts/bootstrap-cloud-build-tools.sh` invokes
  bootstrap once, then `--verify` once (both shell branches do this).
- Measurement harness: `scripts/measure-bootstrap-verify-cost.sh`.
- Raw data: `docs/research/support/bootstrap-verification-cost-2026-10-18.txt`
  (8 cells × 5 reps; provenance header with date, Go version, uname, CPUs,
  `harness_sha256` and `harness_support_sha256` identifying the executed
  harness and sourced provenance-helper bytes at run start).

## Assumptions

- Synthetic manifest + local stub tools only; no real network. Stub tools are
  served over a loopback TLS server (throwaway self-signed CA, trusted via
  `SSL_CERT_FILE`) because manifest validation requires absolute HTTPS URLs.
- Stub tools (4 `raw`, including a 4 MiB padded synthetic shell script, 1
  `archive`, plus the pinned `go` bootstrap-toolchain entry) carry
  `binary_sha256` plus `version_args`/`version_pattern`, so they exercise the
  same checksum + version-exec path (`installedOK`) the real `--verify` pass
  uses. The padded script probes larger hash inputs, but its shell parsing
  cost is not representative of native production binaries.
- "Checkout flow" = `go run` from the module checkout (mirrors the script's
  local-checkout branch); "no-checkout flow" = extract + rebuild the pinned
  source archive per invocation inside the timer, then run the built binary
  (mirrors the script's immutable source-archive branch, including timed
  extraction + build of the current pinned source revision
  `ef4c35629c467887b6749f9e31af111cbe8f8913`). The harness verifies the
  archive's SHA-256 and size against the current manifest and entrypoint pins
  once at fixture setup. That is setup-only provenance checking; it does not
  repeat those checks in each timed arm.
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
`--verify` pass repeats that round for installed tools; this duplicates
verification work, while bootstrap also downloads and installs tools when the
cache is cold.

## Results (median of 5 reps)

Regenerated 2026-10-18 with the per-arm extract+build harness (evidence
`docs/research/support/bootstrap-verification-cost-2026-10-18.txt`; see its
`harness_sha256` provenance header).
The earlier raw file is preserved as
`docs/research/support/bootstrap-verification-cost-2026-10-18-superseded-pre-harness-corrections.txt`.
It is superseded because its committed-HEAD hash did not identify dirty
executed harness bytes, its no-checkout flow used a superseded source archive,
and its archive provenance did not state the explicit size-check limit. The
current rerun uses the current pinned source archive and a 5-CPU host; the
earlier 2-CPU run remains available for historical comparison.
An intermediate 5-CPU result from 12:53 UTC used the wrong archive and
unpaired ordering; its raw samples were overwritten before publication and
could not be recovered. Its known metadata and supersession are recorded in
`docs/research/support/bootstrap-verification-cost-2026-10-18-superseded-intermediate-run.txt`;
those samples are excluded from findings and are not reconstructed.

| flow | cache | bootstrap-only | bootstrap + verify | delta |
|---|---|---|---|---|
| checkout | cold | 18.700s | 21.192s | +2.492s (+13.3%; share 11.8%) |
| checkout | warm | 0.157s | 0.258s | +0.101s (+64.4%; share 39.2%) |
| no-checkout | cold | 19.843s | 20.869s | +1.026s (+5.2%; share 4.9%) |
| no-checkout | warm | 0.151s | 0.190s | +0.039s (+26.1%; share 20.7%) |

For this table, relative increase is `(bootstrap+verify − bootstrap-only) /
bootstrap-only`; the denominator is the bootstrap-only baseline. The verify
pass's share of the combined measured flow is `(bootstrap+verify −
bootstrap-only) / (bootstrap+verify)`; its denominator is the combined
bootstrap-plus-verify median. These percentages answer different questions.

## Revised paired-median trigger

The original #132 declaration on 2026-10-17 registered a different criterion:
mean standalone `--verify` cost of at least 2 seconds **or** verification cost
at least 25% of combined total setup time. Do not treat the paired-median
calculation below as that historical rule or claim the old rule was met from
this dataset; the statistic and denominator differ.

For the corrected 2026-10-18 paired-median run, the salvage acceptance plan
recorded a revised exploratory reporting trigger at 2026-10-18T12:50Z, before
the corrected run generated at 13:10:51Z: mark “threshold met” when any cell
has a positive delta of at least 2 seconds **or** a positive relative increase
of at least 25% over the bootstrap-only median. By that revised measurement
rule, the trigger is **met** in checkout cold (+2.492 s), checkout warm
(+64.4%), and no-checkout warm (+26.1%); it is not met in no-checkout cold
(+1.026 s, +5.2%). This rule was not registered for the original 2026-10-17
measurements or the unrecoverable 12:53 intermediate run. It is an exploratory
result for the corrected run, not an owner-approved performance policy.

The two 25% denominators are not interchangeable: a 25% share of combined time
(`delta / combined`) equals a 33⅓% increase over the baseline (`delta /
bootstrap-only`); a 25% baseline-relative increase is only a 20% combined-time
share.

Reading the table honestly:

- Cold cells (~19–21 s) are dominated by compiling the orchestrator from an
  empty go build cache. All four current median deltas are positive, but cold
  rep ranges overlap substantially; five repetitions on one host do not
  establish a stable cold cost.
- Warm checkout: `--verify` adds +0.101 s, dominated by the second `go run`
  process spawn — not by verification work itself.
- Warm no-checkout: `--verify` adds ~39 ms, the second
  `installedOK` round (checksum + version execs) on the stub tools. This is
  measurement includes no-checkout extraction and rebuild, and is a minority
  of that combined flow.

Absolute deltas with stub tools range from ~39 ms (warm no-checkout) to
~2.5 s (cold checkout). Relative increases are +13.3% checkout cold,
+64.4% checkout warm, +5.2% no-checkout cold, and +26.1% no-checkout warm,
using bootstrap-only as the denominator. Shares of the combined measured flow
are +11.8%, +39.2%, +4.9%, and +20.7%, respectively, using
bootstrap-plus-verify as the denominator. The warm checkout delta is
dominated by a second `go run` process spawn; the warm no-checkout delta
(~39 ms) reflects the second `installedOK` round.

## Confidence and ambiguity behavior

- High confidence: the verify pass executes one additional `installedOK` round
  per tool (checksum + version execs); the measured warm deltas are consistent
  with this mechanism plus process overhead.
- The current cold deltas are positive but not conclusive: rep ranges overlap
  substantially and cold arms compile the orchestrator from an empty go build
  cache. The honest cold-flow claim is that this small study does not establish
  a stable delta, not that the delta is zero or negative.
- Ambiguity (absolute numbers): three stub tools are small shell scripts and
  one is a 4 MiB padded synthetic shell script. Real tools are multi-MB
  binaries, so digest hashing scales with size, but ELF loading and shell
  parsing differ. Real cold runs are dominated by network downloads; those
  and real tools were not measured here.
  The earlier "verification *is* the bootstrap, so `--verify` ~doubles it"
  transfer argument is withdrawn: with the real per-invocation extract+build
  in the baseline, the earlier 2-CPU warm no-checkout verify result was
  +7.9%; that superseded run is not the current measurement.

## Unsupported boundaries

- Real tool archives, real network downloads, and `go-install` tools were not
  measured (explicitly out of scope: synthetic fixture only).
- Single host, single load condition; no concurrent-bootstrap lock-contention
  measurement; Linux/x86_64 only.
- The harness TLS trust is a throwaway self-signed CA — valid for local
  measurement only, not for production trust claims.

## Accuracy checks

- The raw evidence header records SHA-256 digests for the executed
  `scripts/measure-bootstrap-verify-cost.sh` and sourced
  `scripts/lib/benchmark-provenance.sh` worktree bytes at run start, so dirty
  edits to either file are identified accurately. Earlier raw evidence is
  preserved in the explicitly superseded file and is not presented as current.
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

- Acceptance controls run before timing against a binary extracted and built
  from the current pinned archive: valid bootstrap and `--verify` return 0;
  checksum mismatch returns 1; wrong source and installed versions return 1
  with matching source/binary SHA-256 pins; flag misuse returns 2.
- Multiple tools per kind were exercised (4 raw, 1 archive) but not the
  `go-install` kind; stub execs are shell scripts, whose spawn and parse cost
  differ from native ELF binaries. Real-tool timings are outside this
  synthetic-only study.
- Direction: all four current comparisons have positive medians; the cold
  comparisons have substantial rep overlap. This single-host study does not
  establish a stable cold-flow effect.
- Unsafe-control note: none of the measured runs touched the real manifest,
  real network, or shared state; the `bd` CLI was never invoked in the
  worktree.

## Recommendation

The current run measures the separate `--verify` cost as measurable in warm
flows; cold-flow cost remains inconclusive under substantial rep overlap. The
exploratory study trigger is met in three of four cells, as reported above;
that result is separate from the safety recommendation and is not an
owner-approved performance policy.

Retain the separate verification pass as an explicit safety check. Any future
proposal to remove or narrow it requires a separately accepted decision that
defines the safety criterion, failure coverage, and a relevant performance
threshold. This study does not authorize production writers,
reviewer-authority changes, or any change to the orchestrator.
