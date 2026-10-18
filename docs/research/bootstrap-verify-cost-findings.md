# Bootstrap verification cost — findings (dap-5hy.3)

This bounded synthetic study measures the direct orchestrator invocation cost of
adding a separate `cloud-tools-bootstrap --verify` pass in checkout and
no-checkout modes. It does not measure the production entrypoint.

## Method and scope

The harness directly invokes `go run` or the built orchestrator binary. It
excludes the production `scripts/bootstrap-cloud-build-tools.sh` `run_group` wrapper,
including its `setsid`, launcher, readiness/ack polling, wait, and
cleanup overhead. These are direct orchestrator microbenchmarks, not production
entrypoint timings; wrapper overhead may affect the comparison.

The fixture uses synthetic manifests and loopback-only stub tools, including a
4 MiB padded shell script. Every measured tool has checksum and version checks.
Checkout uses `go run` from the module checkout. No-checkout extracts and
rebuilds the pinned source archive per timed arm. The archive SHA and size are
validated once during fixture setup; those checks are not repeated inside each
timed arm. Each cold arm starts with a fresh tools root and scoped Go build and
module caches. Warm arms use separate state seeded by one successful untimed
bootstrap. Five repetitions per cell are paired and interleaved, alternating
which mode runs first.

The current raw data is
[`bootstrap-verification-cost-2026-10-18.txt`](support/bootstrap-verification-cost-2026-10-18.txt):
four cells × two modes × five repetitions. Its provenance header identifies
the exact executed harness and helper bytes, Go 1.27.2, Linux x86_64, five
CPUs, and the run time (2026-10-18T14:20:20Z). The recorded harness/helper
hashes match the source files at commit `4f7009c`, and the run explicitly used
an absolute Go 1.27.2 executable. A subsequent path-only correction normalizes
relative `GO_BIN` and `PATH` selections before the harness changes directories;
it does not change the absolute executable used for these measurements.
The original `gomaxprocs=5` header value was copied from `nproc`, not queried
from the Go runtime. A post-run query with the pinned Go 1.27.2 runtime on this
same five-CPU runner returned 5; the raw transcript marks this correction and
does not imply that the old harness captured it. The current harness now records
`runtime.GOMAXPROCS(0)` directly using an isolated Go build cache, before
acceptance controls and timing.

Controls pass before timing: valid bootstrap and verify return 0; source
checksum mismatch, wrong source version with matching hashes, and installed
wrong version with matching hash return 1; CLI misuse returns 2. A direct
three-entry manifest regression also proves that verification returns a result
for every entry and rejects both a matching-digest wrong-version entry and a
checksum-mismatch entry.

## Results (median of five repetitions)

| flow | cache | bootstrap only | bootstrap + verify | delta |
|---|---|---:|---:|---:|
| checkout | cold | 14.454s | 14.382s | -0.073s (-0.5%; share -0.5%) |
| checkout | warm | 0.119s | 0.232s | +0.113s (+95.6%; share 48.9%) |
| no-checkout | cold | 14.881s | 14.797s | -0.084s (-0.6%; share -0.6%) |
| no-checkout | warm | 0.147s | 0.183s | +0.036s (+24.2%; share 19.5%) |

Delta is `median(bootstrap + verify) − median(bootstrap only)`. Relative
increase is `delta / median(bootstrap only)`, using the baseline as denominator.
The verify pass's share is `delta / median(bootstrap + verify)`, using the
combined result as denominator. These percentages answer different questions:
a 25% share of combined time equals a 33⅓% increase over baseline, while a 25%
baseline-relative increase is a 20% combined-time share. Negative cold deltas
are retained as observed; they do not mean verification makes bootstrap faster.

### Paired-delta sensitivity

The prespecified result above is the difference between marginal medians. As a
sensitivity check, the raw transcript also reports the median of the five
within-repetition verify-minus-bootstrap differences. It is +0.226240s for
checkout cold, +0.113493s for checkout warm, -0.201527s for no-checkout cold,
and +0.036316s for no-checkout warm. The sensitivity calculation does not
replace the prespecified marginal-median estimator.

## Exploratory reporting trigger

The original #132 declaration on 2026-10-17 registered a different criterion:
mean standalone `--verify` cost of at least 2 seconds **or** verification cost
at least 25% of combined total setup time. That is not the criterion applied to
this dataset; its statistic and denominator differ.

The salvage plan recorded a revised exploratory marginal-median reporting
trigger at 2026-10-18T12:50Z, before the earlier corrected run and this 14:20
rerun. It marks “threshold met” if any cell has a positive marginal delta of
at least 2 seconds or a positive relative increase of at least 25% over the
bootstrap-only median. For this rerun, the trigger is **met only in checkout
warm** (+95.6%). It is not met in checkout cold (-0.5%), no-checkout cold
(-0.6%), or no-checkout warm (+24.2%). This exploratory result is not an
owner-approved performance policy and does not recommend changing verification
behavior.

## Limits and interpretation

- Warm deltas are approximately 113 ms in checkout and 36 ms in no-checkout.
  The checkout increment includes a second `go run` spawn; no-checkout includes
  extraction and rebuild per arm plus the repeated checksum and version checks.
- Cold deltas are small and negative in this run. Cold repetition ranges vary
  substantially; five repetitions on one host do not establish a stable cold
  effect.
- Real tool archives, network downloads, `go-install` tools, production
  wrapper timings, and real validator performance were not measured. Stub shell
  execution differs from native binaries, and real downloads can dominate cold
  flows.
- This is one Linux x86_64 host and one load condition. It does not measure
  concurrent bootstrap contention.
- The self-signed loopback TLS trust is for local measurement only.

## Evidence history and safety

The 13:10:51Z raw run remains preserved as
[`superseded-before-host-portability-and-manifest-controls`](support/bootstrap-verification-cost-2026-10-18-superseded-before-host-portability-and-manifest-controls.txt).
The 14:12:20Z run is preserved as
[`superseded-before-paired-summary`](support/bootstrap-verification-cost-2026-10-18-superseded-before-paired-summary.txt): the measurements were retained, but that harness revision did not emit the paired-delta block required by its fact-check. Earlier superseded evidence is preserved in [`superseded-pre-review-corrections`](support/bootstrap-verification-cost-2026-10-18-superseded-pre-review-corrections.txt) and the other `superseded-pre-*` files. The 12:53:48Z intermediate run's raw samples were overwritten and unrecoverable; its known metadata and supersession are recorded in
[`superseded-intermediate-run`](support/bootstrap-verification-cost-2026-10-18-superseded-intermediate-run.txt), with no samples reconstructed.

The harness refuses non-synthetic manifests and non-loopback tool sources. It
aborts on any nonzero measured invocation. No real manifest, real network,
shared installation state, or `bd` command was used. Retain the separate
verification pass as an explicit safety check. Any proposal to remove or
narrow it requires a separately accepted decision defining safety criteria,
failure coverage, and a relevant performance threshold. This study does not
authorize production writers, reviewer-authority changes, or orchestrator
changes.
