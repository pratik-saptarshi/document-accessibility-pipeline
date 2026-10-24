# Decision 0022: go-afp-line dependency re-review (0021 §g)

**Status: Proposed.** This record is due diligence only — the re-review
required by Decision 0021 §g after the profile-file subset gap was
remediated. It does **not** approve dependency adoption, does **not** add
`go-afp-line` to any `go.mod`, and does not approve AFP→PDF conversion
work. The adoption decision is `dap-jo9.4.2`, which remains open; the owner
makes the call. go-afp-line stays an unapproved candidate library until
reviewed **and** accepted per AGENTS.md.

**Adoption status update (2026-10-23):** The status text above is the
evidence as written on 2026-10-22 and is not rewritten. The adoption call
it left open has since been made: the product owner gave the GO
("Adopt and proceed", 2026-10-22), `go-afp-line` is pinned as a direct
dependency in the root `go.mod` at
`v0.1.0-beta.1.0.20261022163835-492c4c7d35b9`, and the integration merged
via PR #187. The scope of the GO — test-only use or product runtime too —
is TBD pending owner decision #1; this record asserts no scope. See the
draft GO record, [Decision 0025](0025-go-afp-line-adoption-go.md).

**Date: 2026-10-22. Reviewer team: 0021 §g re-review lane (coordinator +
three evidence agents).**

## Context

Decision 0021 (Proposed; PR #184, unmerged at the time of this review)
evaluated go-afp-line `v0.1.0-beta.1` (`fc2427b0`) and recommended **DEFER**:
the revision reproducibly aborted the owner-approved 0020 profile file
(`01_Health_Coverage.afp`) at field 0 — 10 unregistered SFIDs — plus a bare
8-byte EDT strictness. §g required that, if the gap were closed by a
library fix, the re-review must run against the *newer fixed revision*
(never carry the beta forward) and must verify the *resolved adoption
graph* rather than the candidate's standalone build list. §k required a
DAP-side integration test and fidelity comparison against the 0020 trusted
references before adoption.

Since then, two owner rulings (2026-10-22, binding):
1. Remediation path (a) authorized: registry extension for the 10 SFIDs +
   EDT relaxation — implemented in go-afp-line as PR #19, merged to main.
2. Rejection of the afplib-only PTOCA code `0x80` is **acceptable expected
   behavior** — the diagnostic must be loud, not silent; it is not a defect
   and afplib-only codes are not to be admitted.

This record reviews **go-afp-line main @ `492c4c7`** ("feat: gap
remediation for DAP Decision 0021 (registry extension + lenient decode)",
squash of PR #19) against §g and §k.

## Evidence

### (a) Revision under review

- `github.com/pratik-saptarshi/go-afp-line` (Apache-2.0), commit
  `492c4c7` (verified present in the local read-only checkout; evidence
  agents extracted it via `git archive` to keep the checkout pristine).
- Changes vs `v0.1.0-beta.1`: 10 new SFID registry rows (BPF/EPF/BRS/ERS/
  PGD/MDR as contracted accepted-and-noted in Reduce; BOC/EOC/OCD/IOB as
  excluded passthrough), lenient decode mode (`DecodeFieldLenient`,
  opt-in), introducer-only EDT accepted with byte-identical re-encode
  (`BareForm` presence flag).
- go-afp-line CI green on the merged PR; Codex clean; independently
  reviewed (per the go-afp-line lane handoff).

### (b) §g — resolved adoption graph (the x/sys question, closed)

Full-graph MVS simulation (scratch module: DAP's requires + go-afp-line
require + `replace` to a pristine `492c4c7` extraction; `GOPROXY=off`,
module cache warm, zero downloads): **21 shared modules, 20
version-identical, exactly one conflict** — `golang.org/x/sys` resolves at
**v0.49.0** (DAP's direct require wins over the candidate's transitively
resolved v0.48.0, inherited via pdfcpu; go-afp-line's own go.mod does not
require x/sys).

License of the bumped module verified in the module cache:
`golang.org/x/sys@v0.49.0/LICENSE` is the Go Authors' **BSD-3-Clause** —
within the permitted MIT/Apache-2.0/BSD-3-Clause-only set. The remaining
20 modules are version-identical between the candidate's tested graph and
the adoption graph (MIT: uax29/v2, jsonschema-go, go-runewidth, goldmark;
Apache-2.0: mousetrap, pdfcpu v0.16.2, cobra, stringprep; BSD-3-Clause:
tiff, pflag, all other golang.org/x/*; dual MIT+Apache-2.0:
go.yaml.in/yaml/v3). **The resolved set is MIT/Apache-2.0/BSD-3-Clause-only;
zero copyleft.** Adoption introduces no new transitive module beyond
go-afp-line itself.

Caveat carried forward: the candidate's tests/fuzzers ran against x/sys
v0.48.0, not the resolved v0.49.0 (see Recommendation, condition 2).

### (c) Profile-file reproduction (§g)

Input verified: `01_Health_Coverage.afp`, 12,784 bytes, SHA-256
`553a336db9283ebbf85092320b68c060aacdbea4138fa720810f09431f86b73a`
(matches Decision 0020).

- **Field census: 25 structured fields, 22 distinct types, zero aborts**
  (strict `DecodeField`). All 10 remediated SFIDs decode: 6 as contracted
  passthrough kept (`Excluded=false`: BPF idx 0, BRS idx 2, ERS idx 6, MDR
  idx 11, PGD idx 12, EPF idx 24) and 4 as known-excluded passthrough
  (`Excluded=true`: BOC idx 3, OCD idx 4, EOC idx 5, IOB idx 18).
  Pre-existing exclusions BRG/ERG also present. Typed: BDT, BPG, EPG, EDT.
- **Byte-identical round-trip: true** — re-encoded stream is 12,784 bytes,
  SHA-256 `553a336d…b73a`, matching the original.
- **Bare EDT: accepted** — census idx 23, wire offset 12766, 8 bytes,
  decodes as `*afp.EDT` with `BareForm=true`, re-encodes byte-identically.
- **Control fixture** (`afptoline/fixtures/statement-001.afp`): Reduce
  clean — nil error, 2 pages, 0 exclusions, 4 text runs.

The 0021 headline finding (abort at field 0) is **resolved**.

### (d) The 0x80 PTOCA rejection — expected behavior, loud diagnostic

`afptoline.Reduce` on the profile file stops at field 16 of 25 (the first
PTX, wire offset 6039) with the verbatim diagnostic:

`afp: decode error at offset 6063 [validation] (field D3EE9B): unknown PTOCA control sequence 0x80`

Typed as `afp.DecodeError{FieldID:"D3EE9B", Offset:6063, Kind:"validation",
Reason:"unknown PTOCA control sequence 0x80"}` (supports `errors.As`).
**Loudness: LOUD** — names the code, the reason, the enclosing field, and
the kind. Per the binding 2026-10-22 owner ruling this rejection is
acceptable expected behavior, not a defect.

Notes: (1) 5 exclusions (BRG, BOC, OCD, EOC, ERG, each with its subset
reason) are returned *alongside* the error — partial diagnostics survive
the abort. (2) Cosmetic: the reported offset (6063) is off by one byte
from the true sequence offset (6064, verified by byte forensics) — the
PTX rebase misses the stripped 5A marker on wrapped streams. Cosmetic,
pre-existing, not a defect in the rejection itself; noted for upstream.
(3) IOB's exclusion is not recorded because the abort at field 16
pre-empts field 18 — consistent.

### (e) §k(a) — DAP-side integration probe

Exact DAP call path exercised (scratch driver, `/tmp`, `GOPROXY=off`):
`afp.Fields` (streaming, bounded memory, wire offsets) →
`afptoline.Reduce` → `linetopdf.Render` (pdfcpu, `%PDF-1.7`).

- Control fixture: full path clean — **1046-byte PDF, 2 pages**,
  pdfcpu-validated (`PageCountFile=2`, `ValidateFile` passed), delta 0 vs
  the repo's test expectation.
- Profile file: path stops at the 0x80 rejection (see §(d)) — `doc.pages=0`,
  no `line.Document`, nothing reaches Render. Expected per the ruling.

API ergonomics for a DAP integrator: errors are typed and actionable
(`afp.DecodeError{FieldID, Offset, Kind, Reason}`); offset semantics differ
by layer (Reduce: absolute wire offsets; direct `DecodePTOCA`: payload-
relative) — the integrator must know which layer produced the error;
`ReduceWithBudget` (default 1 MiB text cap) and `ReduceWithCodePage`
(CP 37 default, CP 500 exposed) are the knobs; no resume/skip past a
failing PTX — abort is total by design; `Render` has a closed font
registry (`FIX-MONO-10`→Courier only) and WinAnsi-only text. Fail-loud
throughout; nothing silently skipped.

### (f) §k(b) — fidelity comparison (blocked by design; partial measured)

End-to-end AFP→PDF fidelity against the 0020 trusted references is **not
measurable**: `DecodePTOCA` is atomic (nil runs on error), so zero text runs
decode before the accepted 0x80 stop — there is no `line.Document` to
render and no AFP→PDF comparison against the layout reference is possible.
This is the expected, ruling-sanctioned outcome.

Partial fidelity (measured, no thresholds invented — tolerances remain open
per Decision 0020): 0 text runs / 0 characters extracted before the block
vs the HTML reference's 2116 characters across 17 sentences — 0/2116
reference characters covered, all 17 reference sentences unmatched. Raw
measured deltas; not a pass/fail claim.

### (g) Licenses (re-check at resolved versions)

Per §(b): the MVS-resolved set is MIT/Apache-2.0/BSD-3-Clause-only,
including the single bumped module (x/sys v0.49.0, BSD-3-Clause, cache-
verified). pdfcpu remains v0.16.2 / Apache-2.0 (the 0021 commercial-
transition inference stays withdrawn; per-version re-read still required
at any bump).

### (h) Material unknowns (carried + new)

- **govulncheck result unknown** (binary absent from this VM; CI runs it —
  unverified locally). Carried from 0021.
- Fuzzing: 4 targets present; long fuzzing not run. Carried from 0021.
- **n=1 profile sample.** Carried from 0021.
- The candidate's test suite has not been run at the MVS-resolved x/sys
  v0.49.0 (only v0.48.0).
- No DAP production code consumes the library yet; the integration probe
  was a scratch driver, not a repo test.

### (i) Explicit limitations

- Fidelity is unmeasured end-to-end (blocked by the accepted 0x80
  behavior); a zero-exclusion count is not a fidelity result.
- Findings are evidence about this revision × this file, not a
  characterization of all AFP input.
- The 0x80 offset cosmetic (6063 vs 6064) is noted, not fixed here.

### (j) Alternatives considered

- **Admit afplib-only PTOCA codes (e.g. 0x80):** rejected by owner ruling
  2026-10-22 — rejection is acceptable expected behavior; admission would
  need IBM-source verification of each code's semantics.
- **Carry v0.1.0-beta.1 forward:** rejected per §g — the beta cannot
  process the profile file; the pin under review is `492c4c7`.
- **Defer again pending fidelity:** considered — fidelity is unmeasurable
  *because of* the accepted behavior, so deferral would wait on nothing
  actionable; hence CONDITIONAL rather than DEFER (see Recommendation).

### (k) Consequences

- If `dap-jo9.4.2` approves adoption: pin commit `492c4c7` + go.sum
  module checksums (hashes to be recorded at adoption time; never the
  beta tag); run the candidate's test suite at the resolved x/sys v0.49.0
  (condition 2); surface 0x80-class aborts as typed loud errors in the DAP
  conversion path (condition 3); keep per-version license review; keep the
  durable DAP-side integration test (`internal/afp/integration_test.go`,
  condition 6) green in CI.
- Fidelity validation against the 0020 trusted references stays open —
  satisfiable only when a convertible profile input exists or the PTOCA
  subset grows; tolerances still open per 0020.
- Until adoption: go-afp-line remains an unapproved candidate; no go.mod
  changes.

## Recommendation

**CONDITIONAL — the library is fit for adoption subject to all of:**

1. **Pin:** commit `492c4c7` + go.sum module checksums recorded at
   adoption time. Never the `v0.1.0-beta.1` tag/commit.
2. **x/sys:** run go-afp-line's test suite (`go test ./...`) at the
   MVS-resolved x/sys v0.49.0 before or at adoption (it was tested at
   v0.48.0), or have the owner explicitly accept the minor-version risk.
3. **0x80-class aborts:** DAP's conversion path must propagate
   `afp.DecodeError` (typed, `errors.As`-able) to the operator — loud,
   never swallowed. This is the owner-accepted expected behavior for
   afplib-only PTOCA content.
4. **Fidelity:** remains an open precondition — measurable only once a
   convertible profile input exists; do not claim fidelity from the
   zero-abort census or the byte-identical round-trip.
5. **Licenses:** ordinary per-version re-read at any bump (pdfcpu and
   x/sys included).
6. **DAP-side integration test (§k(a), discharged):**
   `internal/afp/integration_test.go` exercises the adopted revision
   through the intended DAP call path (`afp.Fields` →
   `afptoline.Reduce` → `linetopdf.Render`): a synthetic fixture
   (`testdata/statement-001.afp`) must reduce with zero exclusions and
   render a valid `%PDF-1.7`, and the 0020 profile file must abort with
   the loud typed `afp.DecodeError` naming code `0x80` (owner-accepted
   behavior). Runs in CI. This is the durable library-contract test;
   Phase 3's verification gate remains the separate DAP-lane wiring test
   (propagation of the typed error through DAP's own AFP lane code).

The 0021 DEFER condition (subset gap + owner call) is satisfied: the 10
SFIDs and bare EDT are fixed in the reviewed revision, and the owner has
ruled the 0x80 rejection acceptable. The remaining items above are
adoption-time conditions, not reasons to defer.

## Follow-up Beads issues

- `dap-jo9.4.2` (open): owns the adopt/reject call — this record is its
  evidence; the six conditions above are the proposed adoption terms.
- Decision 0021 (Proposed, PR #184 unmerged): the original review; this
  record supersedes its verdict for the remediated revision.
- Fidelity validation vs the 0020 trusted references: open, blocked by
  the accepted 0x80 behavior; tolerances open per 0020.

## Test evidence (what was actually run)

- `git archive 492c4c7` extraction (read-only) + `go build ./...`
  (`GOPROXY=off`) — candidate builds offline.
- Scratch MVS simulation: DAP requires + go-afp-line require + `replace`
  to the `492c4c7` extraction; `go list -m all` on candidate standalone,
  DAP root, and combined graphs (all `GOPROXY=off`, zero downloads).
- License-file header reads for all 21 shared modules + x/sys@v0.49.0 in
  the local module cache.
- Scratch drivers (`/tmp`, `GOPROXY=off`): field census (25/22, zero
  aborts), SHA-256 round-trip (`553a336d…b73a` match), bare-EDT check,
  `Reduce` on profile (verbatim 0x80 diagnostic captured) and control
  fixture (2 pages, 0 exclusions), full `Fields→Reduce→Render` path
  (1046-byte `%PDF-1.7`, pdfcpu `PageCountFile=2` + `ValidateFile` pass),
  HTML reference text extraction (2116 chars / 17 sentences) for the
  partial-fidelity delta.
- Byte forensics on the 0x80 chain (`od`, python3): failing sequence at
  true offset 6064, diagnostic reports 6063 (cosmetic rebase note).
- Durable DAP-side integration test: `internal/afp/integration_test.go`
  (`TestIntegrationControlFixtureEndToEnd`,
  `TestIntegrationProfileRejectsAfplibPTOCA`) with fixtures under
  `internal/afp/testdata/` — `go test -race ./internal/afp/ -count=1`
  green. This discharges 0021 §k(a); the scratch drivers above remain the
  exploratory evidence.

## Appendix: cited prerequisite excerpts from Decision 0021

Decision 0021 (Proposed, PR #184, unmerged at the time of this review).
The §g and §k sections this record relies on are quoted verbatim here so
the prerequisites are verifiable from this repository alone.

### (g) Operational fit

- Incremental dep weight ~**660 KB** (go-afp-line source tree); pdfcpu's
  28 MB module is already in DAP's graph (tests-only), so incremental
  module-cache cost is go-afp-line alone.
- No daemons, no goroutines/threads in the API path; CGO-free build.
- Pin recommendation (if adopted): **the exact reviewed commit
  `fc2427b0ed90b79515185d190e0dcbd396bbfc32`** via the module pin
  `github.com/pratik-saptarshi/go-afp-line v0.1.0-beta.1`, with `go.sum`
  re-checks at adoption time (go.mod hash
  `h1:fJU1EziLEOTHeKQPHPZ4n53sF5gBarhiMGDbFF2VpoY=`, zip hash
  `h1:4rNK2sGdG/HKonsJT6qRABHBysnwo0Nrjw/Uz3ejiEs=`, proxy VCS origin hash
  verified == the commit, 2026-10-22). An annotated tag is **not** immutable
  by itself — the repo owner can retag, and the release API reports
  `immutable:false`. The binding that matters is **commit + module
  checksum**, not the tag name. **Choose the pin by remediation path:**
  (a) registry/EDT fix → review and pin the *newer fixed revision*,
  rerunning profile/fidelity evidence against that actual candidate;
  (b) a separately approved and validated pre-processing path → pin the
  reviewed commit above. Never carry the old beta forward as the adoption
  pin for a library-fix path. Pre-release caveat applies.


### (k) Consequences

- If the owner resolves the subset gap and `dap-jo9.4.2` later approves
  adoption: pin the exact reviewed commit + module checksum (see §g), choose
  the pin revision by remediation path (registry fix → review and pin the
  newer fixed revision; validated pre-processing → the reviewed commit),
  apply ordinary per-version license review to every module version in the
  adopted graph, and require a DAP-side integration test plus fidelity
  comparison against the 0020 trusted references.
- Until then: go-afp-line remains an unapproved candidate; no go.mod
  changes; no conversion work authorized by this record.

