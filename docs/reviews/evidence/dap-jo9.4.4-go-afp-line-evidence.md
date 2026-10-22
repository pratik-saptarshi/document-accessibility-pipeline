# Evidence: go-afp-line dependency review (dap-jo9.4.4)

Dependency under review: `github.com/pratik-saptarshi/go-afp-line` —
the proposed adoption revision is the **v0.1.0-beta.1** tag.

Every claim below names the exact command or file it was read from.
Inspection was read-only against the go-afp-line checkout at the revision
below; nothing in that repository was modified. This file is raw evidence for
the Task 2 review document; it makes no adoption recommendation.

## Source revision

- Proposed revision: tag `v0.1.0-beta.1` (the repo's only tag;
  `git tag -l` in the go-afp-line checkout returns just `v0.1.0-beta.1`).
- `git rev-parse v0.1.0-beta.1` returns
  `516be13dfc7a4af03b200c40fae3523f6f520c0d`, which is an **annotated tag
  object** (`git cat-file -t v0.1.0-beta.1` → `tag`), not a commit.
- Peeled to the commit: `git rev-parse 'v0.1.0-beta.1^{}'` →
  `fc2427b0ed90b79515185d190e0dcbd396bbfc32`
  (`docs(changelog): cut v0.1.0-beta.1 release section (#18)`).
- The tagged commit is the checkout's HEAD:
  `git rev-parse HEAD` → `fc2427b0ed90b79515185d190e0dcbd396bbfc32`.
- `git status --porcelain` → empty (clean working tree).
- `git diff 'v0.1.0-beta.1^{}' HEAD --stat` → empty (no drift between the
  tagged revision and HEAD).

## Direct license evidence

- License file: `LICENSE` (root of the go-afp-line checkout). First lines
  read "Apache License / Version 2.0, January 2004" (read via
  `head -5 LICENSE`). License body is the standard Apache-2.0 text
  (also confirmed: `grep -m1 "Version 2.0" LICENSE` on the module-cache
  copies of sibling Apache-2.0 projects uses the same wording).
- No SPDX identifier string appears: `grep -ri "SPDX" LICENSE README.md`
  returned nothing.
- No per-file license headers: `grep -rl "Apache"` over the non-test `.go`
  files of `afp/`, `afptoline/`, `line/`, and `linetopdf/` returned nothing,
  and a case-insensitive `grep -ril "license"` over the same files returned
  nothing.
- No NOTICE file: `find . -iname "notice*"` (excluding `.git`) returned
  nothing.
- Apache-2.0 is permissive and satisfies the repository rule that product
  dependencies must be permissively licensed open-source software
  (per the worktree's `AGENTS.md`, "Project boundaries").

## Transitive license evidence

`go.mod` (read directly) declares module
`github.com/pratik-saptarshi/go-afp-line`, `go 1.27`, one direct require
(`github.com/pdfcpu/pdfcpu v0.16.2`) and eight indirect requires. The full
module graph (`go list -m all` in the checkout) adds pdfcpu's own build
graph. Licenses below were read from each module's extracted directory in
the Go module cache (`go env GOMODCACHE`), from that module's license file
— command pattern:
`grep -im1 "<marker>" "$GOMODCACHE/<escaped-module>@<version>/LICENSE*"`.

| Module | Version | License file | License found |
|---|---|---|---|
| github.com/pdfcpu/pdfcpu | v0.16.2 | LICENSE.txt | Apache License, Version 2.0 ("Licensed under the Apache License, Version 2.0") |
| github.com/hhrutter/tiff | v1.0.7 | LICENSE | BSD-style ("Copyright (c) 2009 The Go Authors… Redistribution and use in source and binary forms…") |
| github.com/mattn/go-runewidth | v0.0.30 | LICENSE | MIT License |
| github.com/xdg-go/stringprep | v1.0.4 | LICENSE | Apache License, Version 2.0, January 2004 |
| github.com/clipperhouse/uax29/v2 | v2.7.0 | LICENSE | MIT License |
| go.yaml.in/yaml/v3 | v3.0.5 | LICENSE | MIT License |
| golang.org/x/crypto | v0.57.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/image | v0.46.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/text | v0.42.0 | LICENSE | BSD-style (Go Authors) |
| github.com/spf13/cobra | v1.10.2 | LICENSE.txt | Apache License |
| github.com/spf13/pflag | v1.0.10 | LICENSE | BSD-style ("Copyright (c) 2012 Alex Ogier… Redistribution and use…") |
| github.com/inconshreveable/mousetrap | v1.1.0 | LICENSE | Apache License |
| github.com/yuin/goldmark | v1.4.13 | LICENSE | MIT License |
| github.com/google/jsonschema-go | v0.4.3 | LICENSE | MIT License ("Copyright (c) 2025 JSON Schema Go Project Authors") |
| golang.org/x/mod | v0.41.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/net | v0.58.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/sync | v0.23.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/sys | v0.48.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/term | v0.46.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/tools | v0.49.0 | LICENSE | BSD-style (Go Authors) |
| golang.org/x/xerrors | v0.0.0-20190717185122-a985d3407aa7 | LICENSE | BSD-style ("Copyright (c) 2019 The Go Authors…") |

Notes:

- The first nine rows (pdfcpu through x/text) are the nine modules in the
  go-afp-line `go.mod` require blocks; the remaining twelve reach the graph
  via pdfcpu's own dependencies and are not listed in go-afp-line's go.mod.
- Every license found is permissive (Apache-2.0, MIT, or BSD-style). No
  copyleft license appears anywhere in the graph. No module's license could
  not be determined — the "Material unknowns" section therefore lists none
  for licensing.
- pdfcpu is the only direct dependency, used by `linetopdf` for rendering.
  The repository's own `AGENTS.md` notes pdfcpu is currently used by
  preservation characterization tests, not product runtime — its use here
  would be a runtime dependency, which is the subject of this review, not a
  precedent.

## Maintenance and security posture

- History: `git log --format='%h %ad %an' --date=short | wc -l` → **85
  commits**; earliest `2026-10-13`, latest `2026-10-21` (9-day span).
  Cadence is bursty (concentrated review rounds: "Round 1", "Round 2 Phases
  A/B/C", beta cut), not a steady trickle.
- Authors (`git log --format='%an' | sort | uniq -c`): 45 commits
  `CloudBot Neo-green`, 21 `CloudBot Neo-blue`, 17 `Pratik Saptarshi`,
  2 `dependabot[bot]`. Four distinct identities; two are automation.
- `SECURITY.md` (read in full): only the latest revision of `main`
  receives security fixes; vulnerabilities are reported via GitHub's
  private vulnerability reporting (no public issues); the policy asks for
  affected revision, Go version, and reproduction steps, and aims to
  acknowledge within 5 business days. Notable implication: the
  v0.1.0-beta.1 tag is a non-`main` revision, so per the policy it would
  not receive security fixes unless the pin is refreshed.
- CI hygiene signals (from `CHANGELOG.md`): Dependabot coverage for Go
  modules and GitHub Actions (weekly); `govulncheck` vulnerability
  scanning in CI; two Dependabot-authored commits are in the history
  (bump actions/checkout 4.4.0→7.0.1, actions/setup-go 5.6.0→7.0.0).
- Quality signal measured in the checkout (read-only runs):
  `go build ./...` → success; `go vet ./...` → clean;
  `go test -count=1 ./...` → all 5 packages pass fresh
  (`afp` 2.124s, `afptoline` 0.018s, `afptoline/genfixture` 0.009s,
  `line` 0.004s, `linetopdf` 0.103s), zero failures.

## Release and versioning

- Tags: only `v0.1.0-beta.1` (`git tag -l`). No `v0.1.0` final, no
  earlier release tags, no release branches observed.
- `CHANGELOG.md` (Keep-a-Changelog format) has a `[0.1.0-beta.1] -
  2026-10-21` section: "First beta release." Added: Dependabot coverage,
  govulncheck in CI, runnable example tests (`ExampleFields` in `afp`,
  `ExampleReduce` in `afptoline`). Changed: API polish (12 internal-only
  identifiers unexported across `afp`, `line`, `afptoline`; `LookupFont`
  returns a defensive copy); PTOCA inline-position advancement past emitted
  text on the `Reduce` path. Fixed: 5A length-bound framing fallback,
  EBCDIC NEL→space mapping, PTOCA decoder state persisted across PTX fields.
  Earlier sections document the adversarial-panel review rounds
  (Round 1 P0 batch, Round 2 Phases A/B, Round 2 Phase C) with fail-fast
  budgets, typed error kinds, and fuzz targets.
- Go module versioning implications of the beta tag (standard Go module
  semantics, no repo-specific evidence needed): `v0.1.0-beta.1` is valid
  semver for `go get`; a consumer pins it with an explicit version
  (`go get github.com/pratik-saptarshi/go-afp-line@v0.1.0-beta.1`) —
  `@latest`/`@upgrade` will not select a pre-release. Being pre-1.0, the
  module path needs no `/vN` major-version suffix, and the author makes no
  compatibility promise. Pinning a beta means deliberately tracking
  pre-release API churn.

## Profile feature coverage

Reference: `docs/subset-contract.md` in the go-afp-line checkout
(§3–§6), contrasted with Decision 0020
(`docs/decisions/0020-first-afp-source-profile.md` in this worktree):
the sample `01_Health_Coverage.afp` has **25 structured fields, 22 distinct
field types, 1 page**, and is a COBRA correspondence letter.

Supported by go-afp-line at the reviewed revision
(from `docs/subset-contract.md` and `afp/field.go`, `afp/codepage.go`):

- **Structured fields (13):** BDT (D3A8A8), EDT (D3A9A8), BPG (D3A8AF),
  EPG (D3A9AF), BNG (D3A8AD), ENG (D3A9AD), BAG (D3A8C9), EAG (D3A9C9),
  BPT (D3A89B), EPT (D3A99B), PTX (D3EE9B) — the minimal document walk —
  plus text-environment companions MCF (D3AB8A, Map Coded Font) and PTD
  (D3B19B, Presentation Text Descriptor).
- **Triplets:** all 45 dual-corroborated triplet IDs registered
  (contract §4 lists all 45); v1 decode paths require only the Comment
  triplet (0x65). Triplet 0x21 is explicitly excluded pending a check
  against the IBM reference and surfaces a typed error naming the triplet.
- **PTOCA control sequences (19):** AMB (D2), AMI (C6), RMB (D4), RMI (C8),
  SBI (D0), SIM (C0), SIA (C2), SVI (C4), STC (74), BSU (F2), ESU (F4),
  SCFL (F0), STO (F6), TRN (DA), RPS (EE), DIR (E4), DBR (E6), BLN (D8),
  NOP (F8). Afplib-only codes (0x72, 0x76, 0x78, 0x80) are excluded.
- **Code pages:** 037 and 500 only (`afp/codepage.go`: registry
  `{37: …, 500: …}`; any other CPGID is a loud diagnostic:
  `"code page CPGID %d outside the v1 subset (supported: 37, 500)"`).
- **Framing:** naked MO:DCA streams and `5A`-wrapped variable-blocked
  records both accepted (contract §10).

Conformance read against Decision 0020's profile facts:

- The sample's 1 page and correspondence-letter (PTOCA text) nature fit
  the contracted single-page PTOCA-text walk (BDT … BPG … PTX … EPG …
  EDT).
- Decision 0020's measured 22 distinct field types vs. the contract's 13
  supported types is the open feasibility question from Decision 0020
  itself ("field types outside go-afp-line's contracted subset" are an
  open error boundary). Out-of-subset-but-well-formed fields become
  `line.Exclusion` diagnostics and reduction continues; malformed fields
  abort with a typed `DecodeError` (contract §9). IOCA, GOCA, BCOCA,
  FOCA, CMOCA/color, overlays/medium, FormDef/PageDef carriers, and object
  containers are all on the closed exclusion list — any occurrence in the
  sample produces a diagnostic, never a silent skip.
- Code pages 037/500 match Decision 0020's boundary ("code pages beyond
  037/500" unsupported).
- Multi-page documents are an open boundary in Decision 0020; the library
  models multi-page documents (`line.Document` has `Pages`), but the
  profile decision scopes to 1 page.

Material unknowns for coverage: the sample file `01_Health_Coverage.afp`
is **not present** in the go-afp-line checkout (`find . -iname "*health*"`
excluding `.git` → nothing), so which of the 22 field types it contains —
and whether each falls inside the 13-type contract or the exclusion list —
could not be measured here. That measurement belongs to the feasibility
spike, not this evidence task.

## Integration behavior

- Module path: `github.com/pratik-saptarshi/go-afp-line`; `go` directive
  `1.27` (from `go.mod`).
- Exported API surface (non-test `.go` files, via
  `grep -h "^func [A-Z]\|^type [A-Z]" <pkg>/*.go`):
  - `afp`: `type CodePage`, `LookupCodePage(cpgid uint16) (CodePage, error)`,
    `type ErrorKind`, `type DecodeError`, `type Field` interface,
    `type BDT/EDT/BPG/EPG/RawField`, `Acronym(typeID string) (string, bool)`,
    `DecodeField(b []byte) (Field, error)`, `type Introducer`,
    `type StreamField`, `Fields(r io.Reader) iter.Seq2[StreamField, error]`
    (streaming iterator), `SplitFields(r io.Reader) ([][]byte, error)`,
    `type Triplet`, `type Triplets`, `RoundTrip(r io.Reader) error`.
  - `afptoline`: `type Run`, `DecodePTOCA(f afp.Field, cp afp.CodePage)
    ([]Run, error)`, `Reduce(r io.Reader) (line.Document, []line.Exclusion,
    error)`, `ReduceWithBudget(r io.Reader, maxTextBytes int)
    (line.Document, []line.Exclusion, error)`,
    `ReduceWithCodePage(r io.Reader, cp afp.CodePage)
    (line.Document, []line.Exclusion, error)`.
  - `line`: `type Exclusion`, `ClassifyExclusion(acronym string)
    (Exclusion, bool)`, `type Document` (has `Pages`), `type Page`,
    `type TextRun`, `type ValidationError`, `type FontMetrics`,
    `LookupFont(id string) (FontMetrics, bool)` (returns a defensive
    copy per the beta changelog).
  - `linetopdf`: `Render(doc line.Document, w io.Writer) error`
    (single entry point; renders an untagged visual PDF — see below).
- Error-handling style: typed errors. `afp.DecodeError` carries an
  `ErrorKind` (structure, encoding, budget, validation — per
  `afp/errors.go` and `afp/errors_test.go`) plus stream offsets; errors
  from malformed input are returned, never silent. Out-of-subset content
  is returned as `line.Exclusion` diagnostics alongside the document, not
  dropped. Encode-side helpers panic on programmer errors
  (e.g. `EncodeUnmappableRunePanics`, `EncodeTooLongPanics` in tests), so
  callers must pre-validate or recover around the encode path.
- No `context.Context` on any public API (explicitly dropped per ADR 0001
  noted in the changelog: "no caller scenario justified the breaking
  change").
- `linetopdf.Render` emits an **untagged** visual PDF (there is an explicit
  `TestRender_Untagged`), uses WinAnsi text encoding, and resolves fonts
  through a pinned synthetic metrics table (`line.LookupFont`) — no FOCA
  font resources are read (FOCA is on the exclusion list). This matches
  Decision 0020's statement that go-afp-line "renders visual PDFs and
  performs no UA tagging".

## Operational fit

- Measured build: `go build ./...` in the go-afp-line checkout at the
  tagged revision → success, no warnings emitted.
- Measured conversion run (fixture from the go-afp-line repo, run from a
  throwaway module in scratch space with a `replace` directive pointing at
  the checkout — the go-afp-line repo itself was not modified):
  - Input: `afptoline/fixtures/statement-001.afp` (hand-authored fixture;
    other shipped fixtures: `exclusion-goca.afp`, `medium-ref.afp`).
  - Pipeline: `afptoline.Reduce` → `linetopdf.Render` to a new PDF file.
  - Result: 2 pages, 0 exclusions; output 1,046 bytes; output SHA-256
    `97f0189a802345aa8b8f1a3dd6ef820f0059b296873e10bfaf5817e505848b7a`.
  - Wall time: ~2.9s wall for the whole `go run` invocation (includes
    first-time compilation of the program and its dependency tree; not a
    pure conversion benchmark).
- Fixture availability: yes — the repo ships three `.afp` fixtures under
  `afptoline/fixtures/`, plus afplib research fixtures under
  `docs/_evidence/afp-research/` (third-party sample files used for
  research, not for the conversion pipeline). Decision 0020's
  `01_Health_Coverage.afp` is not among them (see "Material unknowns").
- Runtime footprint signals (not measured, observed in code): bounded
  decode budgets (`ReduceWithBudget`, architectural max field length
  32767 with fail-fast before allocation, coordinate/page-dimension
  bounds), and a streaming field iterator (`afp.Fields`) that bounds input
  retention — documented in the changelog.

## Material unknowns

1. Which of Decision 0020's 22 distinct field types in
   `01_Health_Coverage.afp` fall inside the 13-type contract vs. the
   exclusion list — the sample file is not in the go-afp-line checkout,
   so this could not be measured (belongs to the feasibility spike).
2. Whether the maintainers would backport a security fix to the
   `v0.1.0-beta.1` tag: `SECURITY.md` supports only the latest `main`.
3. Conversion throughput on production-shaped input: the measured run is a
   small hand-authored fixture, not a performance benchmark.
4. The "22 field types" figure comes from Decision 0020 (measured with
   go-afp-line per that record); it was not re-derived here.
