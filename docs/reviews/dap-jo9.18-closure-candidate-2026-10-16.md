# `dap-jo9.18` scoped closure candidate review

**Status:** The candidate passed independent read-only review. Promotion remains gated on exact implementation-head protected checks. No live Beads database was read or written.

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

Root Go formatting, tidy, vet, race tests, traceability, both nested Go module checks, bootstrap/source-publisher shell syntax, and source-publisher regression tests passed on the implementation tree. The full reconciler race suite also passed with the pinned Beads binary and disposable-store integrations enabled. Hosted exact-head checks have not yet run; canonical candidate promotion is pending those checks and will produce a new exact head requiring fresh checks.
