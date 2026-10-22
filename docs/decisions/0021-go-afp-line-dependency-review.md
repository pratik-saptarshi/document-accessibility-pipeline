# Decision 0021: go-afp-line dependency review (dap-jo9.4.4)

**Status: Proposed.** This record is due diligence only. It does **not** approve
dependency adoption, does **not** add `go-afp-line` to any `go.mod`, and does not
approve AFP→PDF conversion work. The adoption decision is `dap-jo9.4.2`, which
remains open; the owner makes the call. go-afp-line stays an unapproved candidate
library until reviewed **and** accepted per AGENTS.md ("Before adopting a
product/runtime dependency, record exact direct and transitive licenses,
maintenance/security posture, and requirement fit… Keep candidate libraries
unapproved until reviewed.").

**Date: 2026-10-22. Reviewer team: dap-jo9.4.4 (dependency-review lane).**

## Context

Decision 0020 (Accepted, owner-selected 2026-10-21) chose the afpworld
`Sample_1_health.zip` bundle (`01_Health_Coverage.afp`, 12,784 bytes, SHA-256
`553a336db9283ebbf85092320b68c060aacdbea4138fa720810f09431f86b73a`,
25 structured fields / 22 distinct types / 1 page) as the first AFP source
profile, with go-afp-line as the named conversion engine and PDF 1.7
visual-fidelity as the target. This review evaluates go-afp-line itself — its
revision, license graph, maintenance/security posture, versioning, subset
coverage against the 0020 profile file, and integration/operational fit —
against that evidence. Raw tables live in the evidence packs, cited below:

- Pack A (revision, licenses, maintenance/security, versioning):
  `evidence-pack-a-dap-jo9-4-4.md`
- Pack B (profile fit, integration, operational):
  `evidence-pack-b-dap-jo9-4-4.md`
(both packs live in the goal's hidden build-state directory for dap-jo9.4.4)

Two pack inaccuracies found by independent review are corrected here: the
external module count is **21** (the pack's own table lists all 21), and the
unregistered-SFID list in the profile file contains **10** SFIDs, not 9.
DAP-side go.mod facts are taken from this worktree (`go 1.27`,
`toolchain go1.27.3`, `pdfcpu v0.16.2`, `golang.org/x/sys v0.49.0`).

## Evidence

### (a) Exact source revision

- Repository: `github.com/pratik-saptarshi/go-afp-line` (Apache-2.0); evidence was
  gathered against a read-only local checkout pinned at the revision below.
- HEAD == annotated tag `v0.1.0-beta.1` == commit
  `fc2427b0ed90b79515185d190e0dcbd396bbfc32`; `git describe --tags --long`
  gives `v0.1.0-beta.1-0-gfc2427b` — **zero commits after the tag**;
  working tree clean.
- Published release: one, `v0.1.0-beta.1`, type **Pre-release**, published
  2026-10-22T04:01:58Z. No stable release exists. Tag unsigned.
- Full commit cadence in Pack A §1 (85 commits, 2026-10-13–2026-10-21).

### (b) Direct + transitive licenses

Effective build list has **21 external modules** (Pack A §2 table lists all
21; each license file read from the module cache). License set across the
whole graph — including the library itself: **MIT, Apache-2.0, BSD-3-Clause
only. Zero copyleft, zero non-permissive.**

License distribution: Apache-2.0 — go-afp-line (self), pdfcpu v0.16.2,
xdg-go/stringprep v1.0.4, inconshreveable/mousetrap v1.1.0,
spf13/cobra v1.10.2; MIT — clipperhouse/uax29/v2 v2.7.0, mattn/go-runewidth
v0.0.30, go.yaml.in/yaml/v3 v3.0.5 (dual MIT/Apache, inferred), google/jsonschema-go v0.4.3,
yuin/goldmark v1.4.13; BSD-3-Clause — hhrutter/tiff v1.0.7 and all
`golang.org/x/*` modules (crypto, image, text, sys, mod, net, sync, term,
tools, xerrors) plus spf13/pflag v1.0.10. Module→license table in Pack A §2.

**pdfcpu caveat (inferred):** pdfcpu v0.16.2 is the last release still under
Apache-2.0 per its own LICENSE.txt in the module cache; upstream pdfcpu has
since moved to a commercial license. The pin at **pdfcpu v0.16.2** is exactly
what preserves the permissive license — it must not be bumped without a new
license review.

### (c) Maintenance + security posture

- **Owner/maintainer:** pratik-saptarshi; release PR #18 (2026-10-21).
  85 commits (45 CloudBot Neo-green, 21 CloudBot Neo-blue, 17 Pratik
  Saptarshi, 2 dependabot[bot]); three adversarial review rounds; 0 open
  issues.
- **Tests:** `go test ./...` — all 5 packages pass; `go vet ./...` clean;
  85% coverage floor in CI (per CHANGELOG); 4 fuzz targets
  (`FuzzDecodeField`, `FuzzDecodeIntroducer`, `FuzzDecodePTOCA`,
  `FuzzReduce`) — long fuzzing not run.
- **Security:** `SECURITY.md` present; Dependabot (gomod + GitHub Actions,
  weekly, ≤5 open) with observed activity; **govulncheck binary is NOT
  available on this VM, so the local vulnerability scan result is unknown**
  (CI runs it per CHANGELOG; not verified locally).
- **Cadence:** dense burst 2026-10-13–2026-10-21; no commits since the beta
  tag — stable-but-young (inferred).

### (d) Release / versioning

Single pre-release tag `v0.1.0-beta.1` — valid semver and a valid Go module
pin (`go list -m …@v0.1.0-beta.1` resolves, exit 0). CHANGELOG.md follows
Keep a Changelog 1.1.0 (beta section 2026-10-21, Unreleased empty). Pre-release
caveat: **no stable release exists**; adoption pins a beta. See Pack A §4.

### (e) Supported vs unsupported vs the 0020 profile — the headline finding

**Supported:** contracted subset (contract §3): typed decoders BDT/EDT,
BPG/EPG; in-subset-as-`RawField` BNG/ENG, BAG/EAG, BPT/EPT, PTX, MCF, PTD;
PTOCA control sequences ×19 (AMB, AMI, RMB, RMI, SBI, SIM, SIA, SVI, STC,
BSU, ESU, SCFL, STO, TRN, RPS, DIR, DBR, BLN, NOP); triplets ×45; code page 037
(default) and code page 500 (via `ReduceWithCodePage`); naked MO:DCA and `5A`-wrapped framing; US Letter fixed geometry; PDF 1.7
output. See Pack B §1a–§1c.

**Unsupported:** closed exclusion list (contract §9) — IOCA images, GOCA
vector graphics, BCOCA barcodes, CMOCA/color, overlays/medium objects,
object containers, NOP/TLE/IPS/MPS, DBCS; FOCA font resources (pinned
synthetic metrics + `FIX-MONO-10`→Courier instead); FormDef/PageDef carriers
(BFM/EFM, BPM/EPM, BDM/EDM, BFG/EFG); output is untagged PDF only, no UA
tagging (matches the 0020 target only in the negative sense). See Pack B §1d.

**The 0020 profile file:** the owner-approved `01_Health_Coverage.afp`
(Hash-verified identical) field-census reproduces Decision 0020 exactly —
25 structured fields, 22 distinct types — but **10 of its 25 fields use SFIDs
absent from the v0.1.0-beta.1 registry**: BPF/EPF (`D3A8A5`/`D3A9A5`),
BRS/ERS (`D3A8CE`/`D3A9CE`), BOC/EOC (`D3A892`/`D3A992`), OCD (`D3EE92`),
PGD (`D3A6AF`), MDR (`D3ABC3`), IOB (`D3AFC3`). (Names for PGD/MDR/IOB rest
on openafp single-source research — unverified against IBM S550-1058; see
Material unknowns.) Unregistered field types **abort** rather than becoming
exclusion diagnostics, so the file **reproducibly aborts at field 0**:

`afp: decode error at offset 0 [validation] (field D3A8A5): unknown
structured field type: aborting (not in contracted subset)`

A secondary strictness finding: the file's bare 8-byte EDT (introducer
only) would still abort even with the 10 unknown types handled — the decoder
requires 10 data bytes (DocName[8]+reserved[2]).

**Net: the v0.1.0-beta.1 contracted subset cannot process the
owner-approved 0020 profile file as-is.** The in-subset control fixture
(`afptoline/fixtures/statement-001.afp`) reduces cleanly (2 pages,
0 exclusions) and renders `%PDF-1.7`, proving the pipeline works end-to-end
on in-subset input — the gap is coverage, not pipeline mechanics. Closing
the gap requires either (a) a registry/exclusion extension for the 10 SFIDs
plus relaxing the EDT length rule, or (b) a profile-scoped pre-processing
decision — both are product/contract decisions for the owner, not covered by
this review.

### (f) Integration behavior

- Module path `github.com/pratik-saptarshi/go-afp-line`; Go floor **go 1.27**
  both sides (DAP worktree: `go 1.27`, `toolchain go1.27.3`).
- **Pure Go: zero `import "C"`** in go-afp-line and in cached pdfcpu v0.16.2
  (non-test and test); no `net/*` in non-test code — offline-safe at runtime.
- Public API surface: `afp` (`Fields`/`RoundTrip`/`SplitFields`/`DecodeField`/`LookupCodePage`),
  `line` (Document/Page/TextRun/Exclusion), `afptoline`
  (`Reduce`/`ReduceWithBudget`/`ReduceWithCodePage`), `linetopdf` (`Render`
  only). Minimal, focused.
- **Zero version conflicts** with DAP's go.mod: both pin **pdfcpu v0.16.2**,
  all 8 shared indirect modules identical. Adoption would introduce no new
  transitive module beyond go-afp-line itself.

### (g) Operational fit

- Incremental dep weight ~**660 KB** (go-afp-line source tree); pdfcpu's
  28 MB module is already in DAP's graph (tests-only), so incremental
  module-cache cost is go-afp-line alone.
- No daemons, no goroutines/threads in the API path; CGO-free build.
- Pin recommendation (if adopted): **annotated tag `v0.1.0-beta.1`** over a
  pseudo-version — immutable, matches the reviewed state; re-check
  `go.sum` hashes at adoption time. Pre-release caveat applies.

### (h) Material unknowns

- **govulncheck result unknown** (binary absent from this VM; CI runs it —
  unverified locally).
- License of upstream pdfcpu beyond v0.16.2 if ever bumped (commercial now;
  do not bump without a new review).
- **n=1 profile sample** — the whole fit assessment rests on one 12,784-byte
  file; no second sample measured.
- No DAP-side integration test exists in this issue; the abort reproduction
  used a scratch driver outside both repos.
- Long fuzzing not run (4 targets present, brief seeds only).

### (i) Explicit limitations

- Single 12,784-byte sample: the abort finding is evidence about *this*
  revision × *this* file, not a characterization of all AFP input.
- **This review is evidence, not fidelity validation.** A zero-exclusion
  count alone is not a fidelity result — the control fixture had 0
  exclusions and the profile file aborted; neither establishes that
  converted output matches the source. Fidelity evidence would require the
  profile-file gap closed first, then byte/visual comparison against the
  0020 trusted references.
- Pack evidence was produced from a read-only checkout of go-afp-line at
  the pinned SHA; any later tag requires re-review.

### (j) Alternatives considered

- **Pseudo-version pin vs tag pin:** rejected — the annotated tag
  `v0.1.0-beta.1` is immutable and matches the reviewed revision.
- **Adopt now, fix the gap in go-afp-line later:** rejected — adopts against
  the only owner-approved profile file it cannot process.
- **Close the gap by registry extension vs profile-file pre-processing:**
  not decided here — this is an owner/product call (extend the contracted
  subset's registry, or re-scope the profile), and must land before
  `dap-jo9.4.2`.

### (k) Consequences

- If the owner resolves the subset gap and `dap-jo9.4.2` later approves
  adoption: pin the annotated tag, re-check `go.sum`, keep the pdfcpu
  v0.16.2 pin frozen, and require a DAP-side integration test plus fidelity
  comparison against the 0020 trusted references.
- Until then: go-afp-line remains an unapproved candidate; no go.mod
  changes; no conversion work authorized by this record.

### (l) Test evidence (what was actually run)

- `git rev-parse HEAD`, `git tag -l`, `git cat-file -t`, `git describe`,
  `gh release list`, `gh issue list` (all read-only, 2026-10-22) — §(a).
- `go mod graph` + per-module license-file header reads (Pack A §2).
- `go test ./...` (5 packages pass), `go vet ./...` clean,
  `go list -m …@v0.1.0-beta.1` resolves (Pack A §§3–4).
- Scratch driver (in `/tmp`, GOPROXY=off, nothing written to either repo):
  `SplitFields` census on the hash-verified profile file (25 fields /
  22 types — reproduced Decision 0020 exactly), `RoundTrip`/`Reduce` abort
  at field 0 with the quoted `unknown structured field type` error, bare-EDT
  length strictness confirmed, control fixture rendered `%PDF-1.7`
  (1046 bytes). Pack B §1e.
- `grep` for `import "C"` (zero hits) and `net/*` usage in non-test code
  (Pack B §2); version-conflict diff of both go.mods (Pack B §2).

## Recommendation

**Defer adoption of go-afp-line v0.1.0-beta.1 until the profile-file subset
gap is resolved by an owner call — either a go-afp-line registry/exclusion
extension covering the 10 unregistered SFIDs (plus the bare-EDT length
rule), or a profile-scoped pre-processing decision — and keep the adoption
decision itself with `dap-jo9.4.2`.** If adopted later, pin the annotated
tag `v0.1.0-beta.1` (not a pseudo-version), with `go.sum` re-checks at
adoption time.

Evidence: the single material fact against adopt-now is that the reviewed
revision reproducibly aborts the owner-approved 0020 profile file at field
0. Adopting it now would approve a dependency that cannot process the only
owner-approved input it exists to convert. The review otherwise found the
library well-formed for a beta — clean license graph (21 external modules,
all permissive, pdfcpu v0.16.2 pinned before the commercial move), no
version conflicts, pure-Go/offline-safe, tests green, Dependabot + CI
govulncheck — which is why the recommendation is deferral-with-a-clear-path
rather than rejection.

## Follow-up Beads issues

- `dap-jo9.4.2` (open): the adoption decision — owns the adopt/reject call,
  the owner/product resolution of the subset gap, and any later pin target.
- The profile-file subset gap (10 unregistered SFIDs + bare-EDT strictness)
  needs an owner/product call **before** adoption: either a go-afp-line
  registry extension (new issue) or a profile decision — **not** this issue
  (`dap-jo9.4.4`).
- `dap-jo9.4.4` closes on this review record; fidelity validation against
  the 0020 trusted references and a DAP-side integration test are
  follow-ups contingent on the gap being closed.
