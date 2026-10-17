# Repeated bootstrap `--verify` cost (spike dap-5hy.3)

- Date: 2026-10-17
- Spike: beads `dap-5hy.3` — "Measure repeated bootstrap verification cost"
- Tool: `cmd/cloud-tools-bootstrap` (standalone Go module)
- Verdict: **no change justified** — the separate `--verify` pass is not a
  meaningful setup cost. The tool is unchanged by this spike.

## Question

`scripts/bootstrap-cloud-build-tools.sh` runs bootstrap and then a separate
`--verify` pass in both the checkout flow (`go run .` from a repo checkout)
and the no-checkout flow (prebuilt binary from the pinned source archive).
Bootstrap already verifies every tool internally (`installedOK` per tool,
before install, after staging, and again on warm runs). Does the extra
`--verify` pass add a meaningful cold or warm setup cost?

## Method (synthetic only)

- Harness: `cmd/cloud-tools-bootstrap/spikemeasure/spike_test.go` (isolated
  spike package; timing test runs only with `SPIKE_MEASURE_TIMINGS=1`).
- Synthetic manifest: 1 bootstrap-toolchain entry pinned to a stub Go
  (`go1.99.0`, local shell script) + 8 `raw`-kind stub tools totaling
  ~15.8 MB (1 KB–8 MB each, one 8 MB binary to approximate hashing cost of a
  large real binary). All served from a local TLS test server; CLI
  subprocesses trust it via `SSL_CERT_FILE`. **No real tool downloads, no
  production manifest.**
- Checksum (`SHA256` + `BinarySHA256`) and version (`--version` +
  `VersionPattern`) validation retained for every manifest entry.
- Flows measured at CLI wall-clock:
  - checkout: `go run . --manifest … --root … --cache … --go <stub>`
    (fresh `GOCACHE`/`GOMODCACHE` on cold), then `--verify`
  - no-checkout: prebuilt binary (built once, not timed), then `--verify`
- Cold = fresh root/cache dirs per iteration (n=3); warm = reused installed
  state (n=5). Environment: linux/amd64, go1.27.2, 2 vCPU, 7 GB RAM.

Pre-registered decision rule: the `--verify` pass justifies deduplication work
only if its mean cost is **>= 2 s or >= 25% of the flow's total setup time**
in any measured cell.

## Results (wall-clock ms)

| flow       | state | n | bootstrap mean/max | verify mean/max | verify share of total |
|------------|-------|---|--------------------|-----------------|-----------------------|
| checkout   | cold  | 3 | 36601 / 36694      | 380 / 424       | 1.0%                  |
| checkout   | warm  | 5 | 7529 / 36428       | 324 / 402       | 4.1%                  |
| nocheckout | cold  | 3 | 558 / 651          | 182 / 204       | 24.6%                 |
| nocheckout | warm  | 5 | 245 / 528          | 184 / 222       | 42.8%                 |

Raw per-iteration data: `/tmp/spike-dap-5hy.3-results.json` (transient;
re-run the harness to regenerate).

Notes:

- Checkout cold/warm bootstrap includes `go run` compiling the module
  (~36 s on a cold Go build cache; the warm max 36428 ms is the first warm
  iteration populating its `GOCACHE`).
- Warm bootstrap is itself a verification pass: with everything installed,
  `Bootstrap` only re-runs `installedOK` per tool, so the separate `--verify`
  literally duplicates that work — yet the duplicate costs only ~0.2 s.

## Analysis

- The absolute `--verify` cost is **~0.2–0.4 s** in all cells: subprocess
  version probes plus SHA-256 hashing of ~15.8 MB of installed binaries.
- The pre-registered relative prong (>= 25%) fired only in
  nocheckout/warm (42.8%) — an artifact of the denominator being tiny
  (245 ms warm bootstrap), not of the numerator being large. The absolute
  prong (>= 2 s) fired nowhere.
- Scaling to the real pinned set: the largest real binary (`bd`, ~154 MB)
  would add roughly ~0.3 s of hashing; a realistic real-world `--verify`
  pass is on the order of 1–2 s — still negligible against a cold cloud
  setup measured in tens of seconds to minutes (toolchain download,
  `go install` builds).
- Judgment: removing the pass would save a fraction of a second in exchange
  for losing an independent read-only attestation step that the shell paths
  rely on for their exit status. **Not meaningful; no change justified.**

## Validation preserved (test-first controls)

All in `spikemeasure`, green under `go test -race`:

- Corrupted `SHA256` on one entry: `Bootstrap` rejects with "SHA-256
  mismatch"; CLI exits **1**; nothing installed for that entry; temp
  downloads cleaned.
- Wrong version bytes (checksum valid): `Bootstrap` rejects with "version
  validation"; after a good install, corrupting one installed binary makes
  `--verify` print `NOT VERIFIED` and exit **1**.
- Exit-code contract locked in: clean bootstrap **0**, clean `--verify`
  **0**, bootstrap failure **1**, `--verify` failure **1**, `--verify` on
  empty root **1**, flag misuse (missing `--go`) **2**.

## Limitations

- Synthetic stub sizes/heuristic scaling, not the real pinned binaries.
- Localhost TLS downloads, not real CDN/proxy network behavior.
- Single machine, single run (n=3 cold / n=5 warm); re-run the harness to
  reproduce or extend.

## Reproduce

```sh
export PATH="$HOME/workspace/toolchains/go/bin:$PATH"
cd cmd/cloud-tools-bootstrap
SPIKE_MEASURE_TIMINGS=1 SPIKE_RESULTS_PATH=/tmp/spike-results.json \
  go test -run 'TestSpikeVerifyCost_MeasureColdWarm' -v ./spikemeasure/
go test -race ./spikemeasure/   # controls only
```
