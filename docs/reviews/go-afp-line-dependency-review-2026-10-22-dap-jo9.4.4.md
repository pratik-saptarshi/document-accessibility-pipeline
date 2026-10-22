# go-afp-line dependency review (dap-jo9.4.4)

**Review date:** 2026-10-22
**Decision under review:** Adoption of `github.com/pratik-saptarshi/go-afp-line` as the AFP source-decoding dependency for the agreed AFP profile in Decision 0020 (`docs/decisions/0020-first-afp-source-profile.md`).
**Proposed revision:** tag `v0.1.0-beta.1`.
**Evidence:** all claims below are grounded in the Task 1 evidence file, `docs/reviews/evidence/dap-jo9.4.4-go-afp-line-evidence.md`. Citations name the evidence-file section. No numbers were re-derived and none are invented here.

This review is due diligence only. It records whether the library, at the proposed revision, is fit for the agreed AFP profile. It does **not** approve adoption — adoption is dap-jo9.4.2, a product decision reserved for Pratik — and it does not authorize any integration work.

## Exact source revision

Evidence: §Source revision.

The proposed revision is the tag `v0.1.0-beta.1`, the repository's only tag (`git tag -l` returns just that tag). `git rev-parse v0.1.0-beta.1` returns `516be13dfc7a4af03b200c40fae3523f6f520c0d`, an annotated tag object (`git cat-file -t v0.1.0-beta.1` → `tag`), not a commit. Peeled to the commit: `git rev-parse 'v0.1.0-beta.1^{}'` → `fc2427b0ed90b79515185d190e0dcbd396bbfc32` (`docs(changelog): cut v0.1.0-beta.1 release section (#18)`). The tagged commit is the checkout's HEAD, the working tree is clean (`git status --porcelain` empty), and `git diff 'v0.1.0-beta.1^{}' HEAD --stat` is empty, so there is no drift between the tagged revision and HEAD.

## License evidence

Evidence: §Direct license evidence, §Transitive license evidence.

- The library's `LICENSE` reads "Apache License / Version 2.0, January 2004"; the body is the standard Apache-2.0 text. No SPDX identifier string appears, there are no per-file license headers in the non-test `.go` files of `afp/`, `afptoline/`, `line/`, and `linetopdf/`, and no NOTICE file exists. Apache-2.0 is permissive and satisfies the repository rule that product dependencies must be permissively licensed open-source software.
- The module graph was read in full. `go.mod` declares module `github.com/pratik-saptarshi/go-afp-line`, `go 1.27`, one direct require (`github.com/pdfcpu/pdfcpu v0.16.2`) and nine indirect requires. Licenses were read from each module's extracted directory in the Go module cache from that module's own license file. Every module in the graph carries a permissive license (Apache-2.0, MIT, or BSD-style); no copyleft license appears anywhere; no module's license could not be determined.
- pdfcpu is the only direct dependency, used by `linetopdf` for rendering. The repository's own guidance notes pdfcpu is currently used by preservation characterization tests, not product runtime — its use here would be a runtime dependency, which is the subject of this review, not a precedent.

## Maintenance and security posture

Evidence: §Maintenance and security posture.

- History: 85 commits (`git log --format='%h %ad %an' --date=short | wc -l`), earliest 2026-10-13, latest 2026-10-21 — a 9-day span. Cadence is bursty (concentrated review rounds: "Round 1", "Round 2 Phases A/B/C", the beta cut), not a steady trickle.
- Authors: 45 commits `CloudBot Neo-green`, 21 `CloudBot Neo-blue`, 17 `Pratik Saptarshi`, 2 `dependabot[bot]` — four distinct identities, two of them automation.
- `SECURITY.md` was read in full: only the latest revision of `main` receives security fixes; vulnerabilities are reported via GitHub's private vulnerability reporting (no public issues); the policy asks for affected revision, Go version, and reproduction steps, and aims to acknowledge within 5 business days. Implication: the `v0.1.0-beta.1` tag is a non-`main` revision, so per the policy it would not receive security fixes unless the pin is refreshed to a covered revision.
- CI hygiene signals (from `CHANGELOG.md`): Dependabot coverage for Go modules and GitHub Actions (weekly); `govulncheck` vulnerability scanning in CI; two Dependabot-authored commits are in the history (bump actions/checkout 4.4.0→7.0.1, actions/setup-go 5.6.0→7.0.0).
- Measured in the checkout (read-only runs): `go build ./...` succeeds; `go vet ./...` is clean; `go test -count=1 ./...` passes fresh across all 5 packages (`afp` 2.124s, `afptoline` 0.018s, `afptoline/genfixture` 0.009s, `line` 0.004s, `linetopdf` 0.103s), zero failures.

## Release and versioning

Evidence: §Release and versioning.

- Tags: only `v0.1.0-beta.1`. No `v0.1.0` final, no earlier release tags, no release branches observed.
- `CHANGELOG.md` (Keep-a-Changelog) has a `[0.1.0-beta.1] - 2026-10-21` section: "First beta release." Added: Dependabot coverage, govulncheck in CI, runnable example tests (`ExampleFields` in `afp`, `ExampleReduce` in `afptoline`). Changed: API polish (12 internal-only identifiers unexported across `afp`, `line`, `afptoline`; `LookupFont` returns a defensive copy); PTOCA inline-position advancement past emitted text on the `Reduce` path. Fixed: 5A length-bound framing fallback, EBCDIC NEL→space mapping, PTOCA decoder state persisted across PTX fields. Earlier sections document the adversarial-panel review rounds (Round 1 P0 batch, Round 2 Phases A/B, Round 2 Phase C) with fail-fast budgets, typed error kinds, and fuzz targets.
- Go module versioning implications of the beta tag (standard Go module semantics): `v0.1.0-beta.1` is valid semver for `go get`; a consumer pins it with an explicit version (`go get github.com/pratik-saptarshi/go-afp-line@v0.1.0-beta.1`) — `@latest`/`@upgrade` will not select a pre-release. Being pre-1.0, the module path needs no `/vN` major-version suffix, and the author makes no compatibility promise. Pinning a beta means deliberately tracking pre-release API churn.

## Supported and unsupported profile features

Evidence: §Profile feature coverage (reference: `docs/subset-contract.md` in the go-afp-line checkout, contrasted with Decision 0020).

Supported at the reviewed revision:

- **Structured fields (13):** BDT (D3A8A8), EDT (D3A9A8), BPG (D3A8AF), EPG (D3A9AF), BNG (D3A8AD), ENG (D3A9AD), BAG (D3A8C9), EAG (D3A9C9), BPT (D3A89B), EPT (D3A99B), PTX (D3EE9B) — the minimal document walk — plus text-environment companions MCF (D3AB8A, Map Coded Font) and PTD (D3B19B, Presentation Text Descriptor).
- **Triplets:** all 45 dual-corroborated triplet IDs registered; v1 decode paths require only the Comment triplet (0x65). Triplet 0x21 is explicitly excluded pending a check against the IBM reference and surfaces a typed error naming the triplet.
- **PTOCA control sequences (19):** AMB (D2), AMI (C6), RMB (D4), RMI (C8), SBI (D0), SIM (C0), SIA (C2), SVI (C4), STC (74), BSU (F2), ESU (F4), SCFL (F0), STO (F6), TRN (DA), RPS (EE), DIR (E4), DBR (E6), BLN (D8), NOP (F8). Afplib-only codes (0x72, 0x76, 0x78, 0x80) are excluded.
- **Code pages:** 037 and 500 only; any other CPGID is a loud diagnostic (`"code page CPGID %d outside the v1 subset (supported: 37, 500)"`). Code pages 037/500 match Decision 0020's boundary ("code pages beyond 037/500" unsupported).
- **Framing:** naked MO:DCA streams and `5A`-wrapped variable-blocked records both accepted.

Conformance read against Decision 0020's profile facts:

- The sample's 1 page and correspondence-letter (PTOCA text) nature fit the contracted single-page PTOCA-text walk (BDT … BPG … PTX … EPG … EDT). Multi-page documents are an open boundary in Decision 0020; the library models multi-page documents (`line.Document` has `Pages`), but the profile decision scopes to 1 page.
- Decision 0020's measured 22 distinct field types vs. the contract's 13 supported types is the open feasibility question from Decision 0020 itself ("field types outside go-afp-line's contracted subset" are an open error boundary). Out-of-subset-but-well-formed fields become `line.Exclusion` diagnostics and reduction continues; malformed fields abort with a typed `DecodeError` (contract §9). IOCA, GOCA, BCOCA, FOCA, CMOCA/color, overlays/medium, FormDef/PageDef carriers, and object containers are all on the closed exclusion list — any occurrence in the sample produces a diagnostic, never a silent skip.

## Integration behavior

Evidence: §Integration behavior.

- Module path `github.com/pratik-saptarshi/go-afp-line`; `go` directive `1.27` (from `go.mod`).
- Exported API surface (non-test `.go` files): `afp` provides `CodePage`, `LookupCodePage`, `ErrorKind`, `DecodeError`, the `Field` interface, the BDT/EDT/BPG/EPG/RawField types, `Acronym`, `DecodeField`, `Introducer`, `StreamField`, the streaming iterator `Fields(r io.Reader) iter.Seq2[StreamField, error]`, `SplitFields`, `Triplet`, `Triplets`, and `RoundTrip`. `afptoline` provides `Run`, `DecodePTOCA`, `Reduce`, `ReduceWithBudget`, and `ReduceWithCodePage`. `line` provides `Exclusion`, `ClassifyExclusion`, `Document` (with `Pages`), `Page`, `TextRun`, `ValidationError`, `FontMetrics`, and `LookupFont` (returns a defensive copy per the beta changelog). `linetopdf` exposes a single entry point: `Render(doc line.Document, w io.Writer) error`.
- Error-handling style: typed errors. `afp.DecodeError` carries an `ErrorKind` (structure, encoding, budget, validation) plus stream offsets; errors from malformed input are returned, never silent. Out-of-subset content is returned as `line.Exclusion` diagnostics alongside the document, not dropped. Encode-side helpers panic on programmer errors (e.g. `EncodeUnmappableRunePanics`, `EncodeTooLongPanics` in tests), so callers must pre-validate or recover around the encode path.
- No `context.Context` on any public API (explicitly dropped per ADR 0001 noted in the changelog: "no caller scenario justified the breaking change"). Any adopting integration must satisfy the repository's "pass `context.Context` for cancellation or deadlines" boundary for the product-facing surface it wraps — this is integration work, not a blocker of the review, and remains outside its scope.
- `linetopdf.Render` emits an **untagged** visual PDF (explicit `TestRender_Untagged`), uses WinAnsi text encoding, and resolves fonts through a pinned synthetic metrics table (`line.LookupFont`) — no FOCA font resources are read (FOCA is on the exclusion list). This matches Decision 0020's statement that go-afp-line "renders visual PDFs and performs no UA tagging".

## Operational fit

Evidence: §Operational fit.

- Measured build: `go build ./...` in the go-afp-line checkout at the tagged revision succeeds, no warnings emitted.
- Measured conversion run (fixture from the go-afp-line repo, run from a throwaway module with a `replace` directive pointing at the checkout — the go-afp-line repo itself was not modified): input `afptoline/fixtures/statement-001.afp` (hand-authored fixture; other shipped fixtures: `exclusion-goca.afp`, `medium-ref.afp`). Pipeline `afptoline.Reduce` → `linetopdf.Render` produced 2 pages, 0 exclusions, output 1,046 bytes, output SHA-256 `97f0189a802345aa8b8f1a3dd6ef820f0059b296873e10bfaf5817e505848b7a`. Wall time ~2.9s for the whole `go run` invocation (includes first-time compilation of the program and its dependency tree; not a pure conversion benchmark).
- That "0 exclusions" count is a **measurement of one hand-authored fixture run**, not a fidelity result and not evidence for or against adoption. The feasibility question is which of the 22 field types in the Decision 0020 sample fall inside the 13-type contract, and that measurement could not be made here.
- Runtime footprint signals (not measured, observed in code): bounded decode budgets (`ReduceWithBudget`, architectural max field length 32767 with fail-fast before allocation, coordinate/page-dimension bounds), and a streaming field iterator (`afp.Fields`) that bounds input retention — documented in the changelog.
- Fixture availability: the repo ships three `.afp` fixtures under `afptoline/fixtures/`, plus afplib research fixtures under `docs/_evidence/afp-research/` (third-party sample files used for research, not for the conversion pipeline). Decision 0020's `01_Health_Coverage.afp` is not among them.

## Material unknowns

Evidence: §Material unknowns.

1. Which of Decision 0020's 22 distinct field types in `01_Health_Coverage.afp` fall inside the 13-type contract vs. the exclusion list — the sample file is not in the go-afp-line checkout, so this could not be measured (belongs to the feasibility spike).
2. Whether the maintainers would backport a security fix to the `v0.1.0-beta.1` tag: `SECURITY.md` supports only the latest `main`.
3. Conversion throughput on production-shaped input: the measured run is a small hand-authored fixture, not a performance benchmark.
4. The "22 field types" figure comes from Decision 0020 (measured with go-afp-line per that record); it was not re-derived here.

## Recommendation

Defer: `v0.1.0-beta.1` is a clean, permissively licensed, typed-error, budget-bounded library, but adoption should wait until the feasibility spike resolves the open coverage question and the pre-release policy gaps are addressed. The contract's 13 supported structured-field types against Decision 0020's 22 distinct field types in `01_Health_Coverage.afp` is an unresolved material unknown — the sample is not in the library checkout, so no coverage measurement exists, and "0 exclusions" on a hand-authored fixture says nothing about the real sample. A pre-release pin carries no compatibility promise, and `SECURITY.md` supports only latest `main`, so the tag would sit outside the supported revision for security fixes unless the pin is refreshed. Adopting would also move pdfcpu from test-only characterization use to product runtime — a shift this repository has not yet accepted. Defer until: (1) the feasibility spike measures the Decision 0020 sample against the 13-type contract and records the resulting exclusion profile; (2) a stable release exists or the security-policy gap is addressed; (3) Pratik decides pdfcpu's runtime use. Limitations: this review is due diligence only and does **not** approve adoption — adoption is dap-jo9.4.2, reserved for Pratik — and it does not authorize or scope any integration work.
