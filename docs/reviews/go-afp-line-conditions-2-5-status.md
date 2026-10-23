# go-afp-line adoption conditions 2–5: status and evidence (dap-jo9.4.2)

**Record date:** 2026-10-22.
**Scope:** documents-only evidence record for adoption conditions 2–5 from
Decision 0022 (`docs/decisions/0022-go-afp-line-re-review.md`). Conditions 1
(pin `492c4c7` + go.sum checksums) and 6 (DAP-side integration test) are
already discharged on `origin/main` (pin PR #187; `internal/afp/integration_test.go`
in CI). No production code changes in this record.

## Condition 2 — x/sys v0.49.0 suite run: DISCHARGED

Evidence run 2026-10-22 (all commands read-only against a `git archive`
extraction in `/tmp` scratch; nothing was written to the go-afp-line
checkout):

1. `git -C /home/hatch/workspace/repos/go-afp-line archive 492c4c7 | tar -x -C /tmp/gal-evidence`
2. `go mod edit -require=golang.org/x/sys@v0.49.0`
3. `go list -m golang.org/x/sys` → `golang.org/x/sys v0.49.0`
4. `go list -m all` (resolved graph): `github.com/pdfcpu/pdfcpu v0.16.2`,
   `go.yaml.in/yaml/v3 v3.0.5`, `golang.org/x/crypto v0.57.0`,
   `golang.org/x/image v0.46.0`, `golang.org/x/mod v0.41.0`,
   `golang.org/x/net v0.58.0`, `golang.org/x/sync v0.23.0`,
   **`golang.org/x/sys v0.49.0`**, `golang.org/x/term v0.46.0`,
   `golang.org/x/text v0.42.0`, `golang.org/x/tools v0.49.0`,
   `golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7`, plus the
   module itself and non-x/* modules (clipperhouse/uax29/v2, jsonschema-go,
   hhrutter/tiff, mousetrap, go-runewidth, cobra, pflag, xdg-go/stringprep,
   goldmark).
5. `GOPROXY=off GOFLAGS=-mod=mod go test -count=1 ./...` → exit 0:

   - `ok github.com/pratik-saptarshi/go-afp-line/afp 2.449s`
   - `ok github.com/pratik-saptarshi/go-afp-line/afptoline 0.018s`
   - `ok github.com/pratik-saptarshi/go-afp-line/afptoline/genfixture 0.003s`
   - `ok github.com/pratik-saptarshi/go-afp-line/line 0.004s`
   - `ok github.com/pratik-saptarshi/go-afp-line/linetopdf 0.063s`

Verdict: **DISCHARGED.** The suite is green at the MVS-resolved x/sys
v0.49.0 — no owner risk-acceptance needed for the minor version. This is the
exact condition Decision 0022 stated ("tested at v0.48.0" is now superseded
by this run).

## Condition 3 — 0x80-class abort propagation: covered, gate transfers

Covered by the already-discharged condition-6 test
`TestIntegrationProfileRejectsAfplibPTOCA`
(`internal/afp/integration_test.go`): it asserts the 0020 profile file aborts
through `afptoline.Reduce` with a loud, typed `afp.DecodeError`
(`errors.As`-able), `Kind == KindValidation`, the reason naming code `0x80`,
and a non-empty `FieldID`. No new code is needed here. The propagation gate —
the typed error flowing through DAP's own AFP lane code — transfers to the
future lane's own tests once a real lane exists. Boundary per dap-jo9.4.5's
acceptance criteria: "Do not begin implementation until the owner confirms
profile, audience, and delivery scope." Lane wiring verification stays parked
behind that owner call.

## Condition 4 — fidelity vs 0020 trusted references: standing watch

Still open by design. Decision 0022 §f records end-to-end AFP→PDF fidelity
as not measurable: `DecodePTOCA` is atomic (nil runs on error), so zero text
runs decode before the accepted 0x80 stop — no `line.Document` to render and
no comparison against the layout reference is possible. Measured partial
fidelity (no thresholds invented, tolerances open per Decision 0020): 0/2116
reference characters covered, all 17 HTML-reference sentences unmatched, all
before the ruling-sanctioned block. **Watch trigger:** when the 0020 profile
file — or another 0020-reference-backed input — converts end-to-end,
condition 4 becomes measurable and is then executed. Until then, do not claim
fidelity from the zero-abort census or the byte-identical round-trip.

## Condition 5 — license re-read rule: extended standing rule

Decision 0022's condition 5 (ordinary per-version re-read at any bump,
pdfcpu and x/sys included) is adopted and extended by a **new standing rule**
recorded in this change — its authority is this change, not Decision 0022
(0022 required only per-bump re-reads). Re-read licenses on every go-afp-line
version bump **and** on any `go.mod` graph change touching the AFP dependency
subtree, before the pin moves. The subtree is `{go-afp-line, the pdfcpu
subtree, DAP's direct `golang.org/x/sys` imports}`: x/sys does not transit
via go-afp-line at all — it enters DAP's graph through pdfcpu's transitive
subtree and through DAP's own direct require (`golang.org/x/sys v0.49.0` in
the direct require block of `go.mod`, imported by
`ci/pdfua2/adapter/process_linux.go`,
`cmd/beads-export-reconcile/process_linux.go`,
`cmd/verapdf-container-runner/process_linux.go`, and
`internal/financialfixtures/process_linux.go`). The rule is recorded in
`docs/planning/afp-refinement-2026-10-19.md` and `CONTRIBUTING.md`.
