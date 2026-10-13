# Backlog migration — batch one — 2026-10-13

## Baseline and recovery boundary

- Started from the fetched, judge-approved `origin/main` commit `3ec8096aac63c97152597a2456014e2bc9e30b20` on focused branch `docs/statement-invoice-first-migration`.
- Before export: 66 records; SHA-256 `327306f333cdea01c4c32f19c2e4f088de9e408075a4e071582a2c5cd203c9f6`.
- The checkout's `.beads/embeddeddolt/dap` contains only incomplete interrupted-create markers. `bd prime`/`bd ready` could not open it. That directory was preserved untouched; no database was deleted or replaced.
- Beads 1.3.1 imported the tracked export into an isolated temporary Dolt repository. `bd ready` reported 7 at that imported baseline. A supported `bd backup init`/`bd backup sync` full backup was captured for that imported snapshot (manifest SHA-256 `5b4b555c1e6d0942a547d9011e7acfd7aabeffddd846ea67bea868cce3175183`). This backup preserves the isolated import's Dolt state/history; it cannot recover the checkout's unavailable pre-existing Dolt history.

## Applied changes

- Recorded the accepted initial remediation priority: statements and invoices with tables matter first and most. Both families are prioritized; no finer ranking or production scope is inferred.
- Created four real Beads decision children under `dap-jo9.1.1`:
  - `dap-jo9.1.1.1` — initial statement/invoice table remediation experiment profile and accepted priority.
  - `dap-jo9.1.1.2` — separate generation profile; no generation family or engine selected.
  - `dap-jo9.1.1.3` — production document handling; sensitive/production handling remains gated while synthetic engineering is not blocked by unresolved production defaults.
  - `dap-jo9.1.1.4` — human semantic/AT authority; reviewer, protocol, and approval details remain Open.
- Created `dap-jo9.17`, a distinct P1 batch-two task for the 12 measured acceptance-case mapping conflicts and six legacy requirement references. Its executable acceptance preserves all case obligations, requires owner/evidence mapping, and links the affected issue IDs and catalog/matrix. No broad dependency retargeting is assigned before mapping acceptance.
- Closed `dap-jo9.9` after the existing union/matrix and schema whitespace tests passed; its closure explicitly leaves the 12 owner/mapping conflicts to `dap-jo9.17`.
- Closed `dap-jo9.1.5.1` after reviewing its source-pinned generation research evidence. This closes only that bounded research-note task; engine selection remains open.
- Reviewed `dap-jo9.2.3.3`: acceptance tests pass, but status remains in progress per its existing independent-review/current-CI note. Reviewed `dap-jo9.10`: the 20-minute timeout and policy mutation tests pass, but it remains open because no disposable stalled-workflow cancellation run is evidenced.
- Updated [product scope](../requirements/product-scope.md) and added the [next-step decision register](next-step-decision-register-2026-10-13.md). The six-role [panel report](backlog-refinement-panel-2026-10-13.md) remains historical evidence of its original proposed state.

## Proposed next experiment (not implemented)

The decision register records the proposed finite next experiment: two small synthetic fixtures (a statement transaction table and an invoice line-item table) for one fixture-mapped table-header relationship repair. Targeted acceptance preserves displayed financial strings exactly, performs no arithmetic, requires an actual structural change, an independent immutable baseline, a separate atomic candidate, an independent cell/header/row association predicate, candidate-hash-bound explicit `ua2` CI evidence, metadata disclosure, mutation controls, and safe review/unsupported/failure outcomes. It must end proceed/reject/inconclusive for this precise operation. This is not production approval or a PDF-writer implementation; partial PDF 2.0 evidence does not imply conversion or general table support.

## Dependency mapping and health

| Metric | Before | After |
|---|---:|---:|
| Issue records | 66 | 71 |
| Open / closed / in progress | 32 / 28 / 6 | 36 / 30 / 5 |
| Parent-child edges / blocking edges | 64 / 52 | 69 / 52 |
| Total edges | 116 | 121 |
| `bd ready` records | 7 | 11 |
| Duplicate IDs / missing dependency targets / blocker cycles | 0 / 0 / 0 | 0 / 0 / 0 |

All original 66 IDs and all 116 original dependency edges are preserved. The only additions are five allocated IDs listed above and these five parent-child edges: each of `dap-jo9.1.1.1` through `.1.4` → parent `dap-jo9.1.1`; `dap-jo9.17` → parent `dap-jo9`. There are no removed or retargeted edges and no new blocking edges. Full dependency-cycle check also returned zero. The parent `dap-jo9.1.1` remains open. Safety and release-gate dependencies are unchanged; wider graph rewiring is deferred to the acceptance/mapping batch. The existing `dap-jo9.2.1` corpus task remains cross-family and retains its broad profile prerequisite; selectively retargeting it before a narrowly scoped fixture child/acceptance exists could create a partial unsafe graph. Production handling choices must not become blockers for the proposed synthetic experiment; resolve them before sensitive or production inputs instead.

After export, a fresh isolated Beads import/export reproduced the same 71 IDs and dependency set; the tracked export matches the isolated live export. Final export SHA-256: `599f3bf45a14b092e96887410ae65746e97f7894ee7622cf3f6a6ec4cbac7e08`. `bd doctor --check=artifacts` found no classic artifacts; `bd doctor --check=conventions` passed all checks. A supported after-state Dolt backup was synced in the isolated repository (manifest SHA-256 `d869c8f88f1fb541aaa814ce0227d65c4b868e7ea70168ebf28df0b4e1b3760a`). `bd ready` is verified against the imported issue snapshot, not the unavailable checkout database; its 11 entries include parent/decision/mapping work and are not evidence of a product implementation being ready.

## Validation and remaining gates

- Passed root formatting, `go mod tidy -diff`, `go vet ./...`, and `go test -race ./...`.
- Passed `go run ./cmd/traceability-check` (17 requirements, 6 stories, 20 tasks, 26 specification-only cases).
- Passed standalone cloud-bootstrap and CI PDF/UA-2 adapter module race tests, vet, and tidy; `PDFUA_REAL_CLI_JAR=` was empty for adapter tests.
- Passed bootstrap/publisher shell syntax, source-publisher guard tests, and `actionlint`.
- No local vulnerability database fetch, Java/veraPDF run, model download/auth work, PDF writer, or additional form fixture was performed.
- Unresolved: exact experiment profile values and candidate/normative structure; supported topology; financial arithmetic policy; sensitive-document privacy/access/retention; reviewer authority/protocol; generation engine; production workflow and release eligibility. The synthetic repair experiment is proposed for the next batch only.
