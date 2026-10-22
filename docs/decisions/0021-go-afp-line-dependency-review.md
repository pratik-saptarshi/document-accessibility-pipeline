# Decision 0021: go-afp-line dependency review (dap-jo9.4.4)

**Status: Proposed.** This record is due diligence only. It does **not** approve
dependency adoption, does **not** add `go-afp-line` to any `go.mod`, and does not
approve AFP→PDF conversion work. The adoption decision is `dap-jo9.4.2`, which
remains open; the owner makes the call. go-afp-line stays an unapproved candidate
library until reviewed **and** accepted per AGENTS.md ("Before adopting a
product/runtime dependency, record exact direct and transitive licenses,
maintenance/security posture, and requirement fit… Keep candidate libraries
unapproved until reviewed.").

**Correction round 2026-10-22** (review findings on PR #184): the pdfcpu
commercial-transition claim is withdrawn and corrected (§b); published
pdfcpu advisories are recorded with call-path applicability (§c); the
tag-immutability claim is corrected and the pin is bound to commit +
module checksum with the pin chosen by remediation path (§g); the
evidence annex (§A1–A4) commits the sanitized load-bearing evidence to the
tree. The DEFER recommendation and the owner adoption gate are unchanged.

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
against that evidence. The load-bearing tables are committed in the
**Evidence annex** (§A1–A4) below; the raw worktree evidence packs are
superseded by the annex.

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
- Full commit cadence: 85 commits, 2026-10-13–2026-10-21.

### (b) Direct + transitive licenses

Effective build list has **21 external modules** (Annex A1 table lists all 21; each license file read from the module cache). License set across the
whole graph — including the library itself: **MIT, Apache-2.0, BSD-3-Clause
only. Zero copyleft, zero non-permissive.**

License distribution: Apache-2.0 — go-afp-line (self), pdfcpu v0.16.2,
xdg-go/stringprep v1.0.4, inconshreveable/mousetrap v1.1.0,
spf13/cobra v1.10.2; MIT — clipperhouse/uax29/v2 v2.7.0, mattn/go-runewidth
v0.0.30, go.yaml.in/yaml/v3 v3.0.5 (dual MIT/Apache, inferred), google/jsonschema-go v0.4.3,
yuin/goldmark v1.4.13; BSD-3-Clause — hhrutter/tiff v1.0.7 and all
`golang.org/x/*` modules (crypto, image, text, sys, mod, net, sync, term,
tools, xerrors) plus spf13/pflag v1.0.10. Module→license table in Annex A1.

**pdfcpu license correction (2026-10-22):** the earlier inference that
upstream pdfcpu had moved to a commercial license is **withdrawn** — it rested
on reading only v0.16.2's cached license file and could not establish a later
license change. Verified 2026-10-22 against upstream: `v0.16.3`'s `LICENSE.txt`
and the current upstream `LICENSE.txt` are both **Apache License 2.0**.
**Retain ordinary per-version license review:** any bump off v0.16.2 requires
re-reading that version's license file; no version is exempt, and no pin-freeze
rationale may rest on an unverified transition claim.

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
  (CI runs it per CHANGELOG; not verified locally). The absent local binary
  is an execution limitation — it is not evidence that published advisories
  are unknown or inapplicable.
- **Published advisories at/after the pinned pdfcpu (recorded 2026-10-22):**
  pdfcpu v0.16.3 (still Apache-2.0) is a security release:
  **GHSA-jp42-qm45-2x6g** — decoded-byte-limit bypasses in stream decoding
  and cross-reference repair ("applications handling untrusted PDFs ...
  should upgrade"); **GHSA-44h9-p34h-2w4p** — local image access in
  `create`/`form` input confined to an allowed directory. **Applicability to
  go-afp-line's actual call paths (assessed 2026-10-22, not assumed):**
  `linetopdf.Render` builds new PDFs via `create.UpdatePageTree` +
  `api.WriteContext` — a *create* path, never a PDF-decode path, and it adds
  no images (no image/form/fill APIs in non-test code; IOCA images are
  outside the contracted subset). GHSA-jp42-qm45-2x6g (untrusted-PDF
  *decoding*) has no reachable call path today; GHSA-44h9-p34h-2w4p (local
  images in create/form input) is not reachable today because no image
  paths are passed — but the `create` package *is* in the call graph, so
  both advisories must be re-assessed against the actual adoption revision's
  call paths at adoption time (`dap-jo9.4.2`), including whether a newer
  pdfcpu carrying the fixes is the pin target (with its own license
  re-read). No exploitability claim is made for DAP or go-afp-line, and no
  upgrade or adoption is authorized by this record.
- **Cadence:** dense burst 2026-10-13–2026-10-21; no commits since the beta
  tag — stable-but-young (inferred).

### (d) Release / versioning

Single pre-release tag `v0.1.0-beta.1` — valid semver and a valid Go module
pin (`go list -m …@v0.1.0-beta.1` resolves, exit 0). CHANGELOG.md follows
Keep a Changelog 1.1.0 (beta section 2026-10-21, Unreleased empty). Pre-release
caveat: **no stable release exists**; adoption pins a beta. See Annex A3.

### (e) Supported vs unsupported vs the 0020 profile — the headline finding

**Supported:** contracted subset (contract §3): typed decoders BDT/EDT,
BPG/EPG; in-subset-as-`RawField` BNG/ENG, BAG/EAG, BPT/EPT, PTX, MCF, PTD;
PTOCA control sequences ×19 (AMB, AMI, RMB, RMI, SBI, SIM, SIA, SVI, STC,
BSU, ESU, SCFL, STO, TRN, RPS, DIR, DBR, BLN, NOP); triplets ×45; code page 037
(default) and code page 500 (via `ReduceWithCodePage`); naked MO:DCA and `5A`-wrapped framing; US Letter fixed geometry; PDF 1.7
output. See §(e) and Annex A2.

**Unsupported:** closed exclusion list (contract §9) — IOCA images, GOCA
vector graphics, BCOCA barcodes, CMOCA/color, overlays/medium objects,
object containers, NOP/TLE/IPS/MPS, DBCS; FOCA font resources (pinned
synthetic metrics + `FIX-MONO-10`→Courier instead); FormDef/PageDef carriers
(BFM/EFM, BPM/EPM, BDM/EDM, BFG/EFG); output is untagged PDF only, no UA
tagging (matches the 0020 target only in the negative sense). See §(e) and Annex A2.

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

### (h) Material unknowns

- **govulncheck result unknown** (binary absent from this VM; CI runs it —
  unverified locally).
- License of upstream pdfcpu beyond v0.16.2 if ever bumped — ordinary
  per-version license review required at any bump (the earlier
  commercial-transition inference is withdrawn; see §b).
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

- **Pseudo-version pin vs commit-bound tag pin:** rejected for the
  pseudo-version — bind the **commit + module checksum** (see §g); the tag
  name alone is not an immutability guarantee (`immutable:false` on the
  release API).
- **Adopt now, fix the gap in go-afp-line later:** rejected — adopts against
  the only owner-approved profile file it cannot process.
- **Close the gap by registry extension vs profile-file pre-processing:**
  not decided here — this is an owner/product call (extend the contracted
  subset's registry, or re-scope the profile), and must land before
  `dap-jo9.4.2`.

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

### (l) Test evidence (what was actually run)

- `git rev-parse HEAD`, `git tag -l`, `git cat-file -t`, `git describe`,
  `gh release list`, `gh issue list` (all read-only, 2026-10-22) — §(a).
- `go mod graph` + per-module license-file header reads (Annex A1).
- `go test ./...` (5 packages pass), `go vet ./...` clean,
  `go list -m …@v0.1.0-beta.1` resolves (Annex A3).
- Scratch driver (in `/tmp`, GOPROXY=off, nothing written to either repo):
  `SplitFields` census on the hash-verified profile file (25 fields /
  22 types — reproduced Decision 0020 exactly), `RoundTrip`/`Reduce` abort
  at field 0 with the quoted `unknown structured field type` error, bare-EDT
  length strictness confirmed, control fixture rendered `%PDF-1.7`
  (1046 bytes). Annex A2.
- `grep` for `import "C"` (zero hits) and `net/*` usage in non-test code
  (see §(f)); version-conflict diff of both go.mods (see §(f)).

## Recommendation

**Defer adoption of go-afp-line v0.1.0-beta.1 until the profile-file subset
gap is resolved by an owner call — either a go-afp-line registry/exclusion
extension covering the 10 unregistered SFIDs (plus the bare-EDT length
rule), or a profile-scoped pre-processing decision — and keep the adoption
decision itself with `dap-jo9.4.2`.** If adopted later, bind the exact
reviewed commit + module checksum (see §g), choosing the pin revision by
remediation path, with ordinary per-version license review.

Evidence: the single material fact against adopt-now is that the reviewed
revision reproducibly aborts the owner-approved 0020 profile file at field
0. Adopting it now would approve a dependency that cannot process the only
owner-approved input it exists to convert. The review otherwise found the
library well-formed for a beta — clean license graph (21 external modules,
all permissive), no version conflicts, pure-Go/offline-safe, tests green,
Dependabot + CI govulncheck — which is why the recommendation is
deferral-with-a-clear-path
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

## Evidence annex (committed 2026-10-22)

Sanitized load-bearing evidence, so the record stands without the worktree
evidence packs. Machine paths and the tagger's email address from the raw
packs are omitted here; all facts below were re-verified 2026-10-22.

### A1. Exact module / version / license / source mappings (21 external modules)

Effective build list at the reviewed revision (`go list -m all`, excluding
the module itself). License files read from the local module cache
2026-10-22; "source" = the license file inspected in the cached module.

| Module | Version | License | Source file |
|---|---|---|---|
| github.com/clipperhouse/uax29/v2 | v2.7.0 | MIT | LICENSE |
| github.com/google/jsonschema-go | v0.4.3 | MIT | LICENSE |
| github.com/hhrutter/tiff | v1.0.7 | BSD-3-Clause | LICENSE |
| github.com/inconshreveable/mousetrap | v1.1.0 | Apache-2.0 | LICENSE |
| github.com/mattn/go-runewidth | v0.0.30 | MIT | LICENSE |
| github.com/pdfcpu/pdfcpu | v0.16.2 | Apache-2.0 | LICENSE.txt |
| github.com/spf13/cobra | v1.10.2 | Apache-2.0 | LICENSE.txt |
| github.com/spf13/pflag | v1.0.10 | BSD-3-Clause | LICENSE |
| github.com/xdg-go/stringprep | v1.0.4 | Apache-2.0 | LICENSE |
| github.com/yuin/goldmark | v1.4.13 | MIT | LICENSE |
| go.yaml.in/yaml/v3 | v3.0.5 | MIT OR Apache-2.0 (inferred: "covered by two different licenses: MIT and Apache") | LICENSE |
| golang.org/x/crypto | v0.57.0 | BSD-3-Clause | LICENSE |
| golang.org/x/image | v0.46.0 | BSD-3-Clause | LICENSE |
| golang.org/x/mod | v0.41.0 | BSD-3-Clause | LICENSE |
| golang.org/x/net | v0.58.0 | BSD-3-Clause | LICENSE |
| golang.org/x/sync | v0.23.0 | BSD-3-Clause | LICENSE |
| golang.org/x/sys | v0.48.0 | BSD-3-Clause | LICENSE |
| golang.org/x/term | v0.46.0 | BSD-3-Clause | LICENSE |
| golang.org/x/text | v0.42.0 | BSD-3-Clause | LICENSE |
| golang.org/x/tools | v0.49.0 | BSD-3-Clause | LICENSE |
| golang.org/x/xerrors | v0.0.0-20190717185122-a985d3407aa7 | BSD-3-Clause | LICENSE |
| github.com/pratik-saptarshi/go-afp-line (self) | v0.1.0-beta.1 | Apache-2.0 | LICENSE (repo root) |

License set across the whole graph: **MIT, Apache-2.0, BSD-3-Clause only —
zero copyleft, zero non-permissive.** The go-afp-line author's own DAP
dependency uses the same pdfcpu pin, so adoption introduces no new
transitive module beyond go-afp-line itself.

### A2. Profile-file reproduction (headline finding)

- Input: `01_Health_Coverage.afp` — 12,784 bytes, SHA-256
  `553a336db9283ebbf85092320b68c060aacdbea4138fa720810f09431f86b73a`
  (matches Decision 0020).
- Scratch driver (`/tmp`, `GOPROXY=off`, `replace` directive to a read-only
  checkout at the pinned commit; nothing written to either repo):
  - `afp.SplitFields` → **25 structured fields** (reproduces the 0020
    census exactly).
  - `afp.DecodeField(fields[0])` → `afp: decode error at offset 0
    [validation] (field D3A8A5): unknown structured field type: aborting
    (not in contracted subset)`.
- Unregistered SFIDs in the profile file (10): `D3A8A5` (BPF),
  `D3A9A5` (EPF), `D3A8CE` (BRS), `D3A9CE` (ERS), `D3A892` (BOC),
  `D3A992` (EOC), `D3EE92` (OCD), `D3A6AF` (PGD), `D3ABC3` (MDR),
  `D3AFC3` (IOB). PGD/MDR/IOB names rest on openafp single-source research
  (unverified against IBM S550-1058).
- Secondary strictness: the file's bare 8-byte EDT (introducer only) would
  still abort even with the 10 types handled — the decoder requires 10 data
  bytes (`DocName[8]`+reserved[2]).
- Control: the library's own fixture `afptoline/fixtures/statement-001.afp`
  reduces cleanly (2 pages, 0 exclusions) and renders a 1046-byte PDF with
  header `%PDF-1.7`.

### A3. Revision binding

- git: HEAD == annotated tag `v0.1.0-beta.1` == commit
  `fc2427b0ed90b79515185d190e0dcbd396bbfc32`;
  `git describe --tags --long` → `v0.1.0-beta.1-0-gfc2427b` (zero commits
  after the tag); tag **unsigned** (`git verify-tag`: no signature found).
- Release: one, `v0.1.0-beta.1`, type **Pre-release**, published
  2026-10-22T04:01:58Z; no stable release.
- Module pin: `github.com/pratik-saptarshi/go-afp-line v0.1.0-beta.1`;
  go.mod hash `h1:fJU1EziLEOTHeKQPHPZ4n53sF5gBarhiMGDbFF2VpoY=`; zip hash
  `h1:4rNK2sGdG/HKonsJT6qRABHBysnwo0Nrjw/Uz3ejiEs=`; Go module proxy origin
  VCS hash == `fc2427b0ed90b79515185d190e0dcbd396bbfc32` (verified
  2026-10-22). Tags are mutable by the repo owner; bind commit + checksum.

### A4. Unexecuted limits (preserved)

- `govulncheck`: binary absent from the build VM; local vulnerability scan
  **unknown** (CI runs it per CHANGELOG — unverified locally).
- Fuzzing: 4 targets present (`FuzzDecodeField`, `FuzzDecodeIntroducer`,
  `FuzzDecodePTOCA`, `FuzzReduce`); long fuzzing **not run**.
- **n=1 profile sample** — the fit assessment rests on one 12,784-byte file.
- No DAP-side integration test exists in this issue.
