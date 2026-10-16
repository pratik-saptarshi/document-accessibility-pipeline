# `dap-jo9.18` scoped closure candidate review

**Status:** The candidate passed independent read-only review and was promoted to the canonical JSONL by the guarded promoter after the exact implementation head passed hosted checks. Promotion commit `49f61d1f3c160cc1f6a767e13c5d1a43e34f6567` passed exact-head hosted workflows and independent review. Shared API executable-pin correction `095ca84ac4139f59192870fbd12481d805e9da3c` also passed independent review and all four PR-triggered hosted workflows. No live Beads database was read or written.

## Pinned inputs and artifact digests

- Canonical source C: `5ef3435af8d2e75fa1ec67c423a32b3793a72952b2a9b763529ff04c2f90d465`
- Beads version / executable SHA-256: `1.3.1` / `a8f48d771b9e11eccfced4aed72923b59eda746f7a9b425e7ca1af7a51251653`
- Evidence ZIP: [dap-jo9.18-closure-83e05aeb.zip](evidence/dap-jo9.18-closure-83e05aeb.zip), SHA-256 `83e05aebafeda799b72d3435acd6551d52cf04769545fa998f93016e2b048b5a`
- Derived candidate: SHA-256 `69d6079754284f5b81fea74cda717594851f498848cda69c5809044711a754cb`
- Beads close event timestamp: `2026-10-16T21:13:44Z`

## Observed transition

The ZIP contains the exact source, baseline B0, updated B1, typed plan, candidate, second-store round-trip, manifest, command history, and null spans. `bd metrics off` and `bd metrics` completed successfully before `bd --version` and all database operations. The status output confirmed OFF. The single close argv was `close dap-jo9.18 --reason=<the exact reason in plan.json>` and exited 0 in a disposable `/tmp` store. The verification store imported the candidate and exported the round-trip successfully.

Source and candidate contain the same 100 issue IDs. Exactly one source row changed: `dap-jo9.18`, whose fields changed only by adding `closed_at`, `close_reason`, `status=closed`, and the Beads-generated `updated_at`. All 167 dependency edges are unchanged. The five Decision 0007 explicit-null values remain literal `null` tokens in the candidate; the manifest pins each source/candidate span and records their expected absence from the importer exports. B1 and the second-store export match; the candidate preserves the approved source nulls.

## Independent review and gates

An independent bundle reviewer verified the ZIP and candidate digests, command order, executable pin, B0/B1 transition, exact reason, only-target delta, graph preservation, null spans, and round-trip. No blocker was found. An independent code-quality reviewer separately reviewed the typed close implementation and tests against base `48d2380722bab9398e4e979875b8d1c35dd094c6`; after requiring and verifying actual-source digest preflight, command provenance tests, and narrower reason wording, no actionable blocker remained.

The public-tree scrub found no machine paths or agent headers in the new plan, review prose, or source tests. Existing absolute-path research notes, issue fixture text, and CloudBot publisher policy are intentional repository evidence. The reviewed ZIP retains its `/tmp` store paths as immutable command provenance; the archived paths contain no credentials and are not copied into the review prose.

Root Go formatting, tidy, vet, race tests, traceability, both nested Go module checks, bootstrap/source-publisher shell syntax, and source-publisher regression tests passed on the implementation tree. The full reconciler race suite also passed with the pinned Beads binary and disposable-store integrations enabled. All exact-head hosted workflows passed on the implementation commit; the guarded canonical promotion and follow-up validation are recorded below.

## Guarded canonical promotion

The implementation commit `f956d0f8cef05a0300e3e32ffa6f0af73141f25b` passed the exact-head hosted Go checks, security scans, dependency review, and veraPDF integration workflows. The PR had no issue comments, inline review threads, or submitted reviews at the time of promotion. The existing guarded `promote` command accepted the reviewed ZIP digest and replaced only the canonical JSONL in the isolated feature worktree. It printed the candidate digest above; `sha256sum .beads/issues.jsonl` matched that digest afterward. A direct diff showed only the `dap-jo9.18` row changing, with `status=closed`, Beads-generated `closed_at` and `updated_at`, and the approved `close_reason`; all other issue rows and dependency data remained byte-identical. This operation did not access or change the live embedded Dolt database. Promotion commit `49f61d1f3c160cc1f6a767e13c5d1a43e34f6567` passed independent review and all four PR-triggered hosted workflows. The separate shared API executable-pin correction described below must pass its own exact-head review and checks before merge.

## Shared API executable-pin correction

An independent merge review identified a public API gap: `ReconcileWithBinary` checked the authorized Beads executable digest, but the exported convenience API `Reconcile` reached the shared reconciliation function without one. A focused regression reproduced acceptance through `Reconcile` using the exact source, a baseline retaining all five historical null values and normalizing the historical `dap-jo9.5.5.closed_at`, and a valid closure export. The test failed before the correction. The shared reconciliation function now requires the exact pinned Beads digest whenever a close is present; the valid pinned wrapper remains accepted, and the internal second-store validation with no close is unchanged. This closes a library API validation gap; the guarded CLI path already verified the executable and used `ReconcileWithBinary`. The retained candidate remains valid: its archived command history records the pinned executable and successful close, its candidate digest is unchanged, and it was produced by the pinned CLI path. No new close event or timestamp was generated. Independent review found no blocker on `095ca84ac4139f59192870fbd12481d805e9da3c`; its Go checks, security scans, dependency review, and veraPDF integration workflows all succeeded.
