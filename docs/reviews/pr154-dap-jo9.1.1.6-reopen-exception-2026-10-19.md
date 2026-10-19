# PR #154 one-time tracker correction: reopen `dap-jo9.1.1.6`

## Authorization and scope

This is a one-time, issue-specific process exception authorized by the product owner in the PR #154 remediation delegation on 2026-10-19. The owner standing instruction in the authorized task record permits correction of factually invalid content after independent review, overriding repository process restrictions only as needed and recording the exception. The exact correction is limited to the tracked `.beads/issues.jsonl` interchange export and the directly affected roadmap and reviewer-evidence requirement statements. It does not authorize a live Dolt write, a general tracker-edit path, unrelated backlog work, or a product/security choice.

The owner-authorized correction changes only `dap-jo9.1.1.6` in the canonical export. Its status is reopened; the stale `closed_at` and `close_reason` properties are removed; `updated_at` records the actual edit time; and a factual note records what Decision 0015 accepted and what remains incomplete. The original description and acceptance criteria are unchanged. All issue IDs, dependencies, dependency counts, priorities, owner/creator, creation fields, labels, and every non-target row remain unchanged byte-for-byte. In particular, the explicit `closed_at: null` on `dap-jo9.1.6.7` is retained byte-for-byte. No new issue, dependency, or provider identity is introduced.

## Pinned source and candidate

- Source PR head: `90e6f9f6b8dad11b496e50f0086e332b2b3fcde1` (`fix/profile-policy-codex-findings`), with current `origin/main` `5f3f76dbbc9e96443c4788821efceb6cb6b26eef` as an ancestor.
- Source `.beads/issues.jsonl` SHA-256: `d53250d543c0f982294ae9c5eb340f714ead50c4005aa18534ab978e39c56201`.
- Candidate `.beads/issues.jsonl` SHA-256: `879e7e391aa7a2128a17f82ac53be491e6ba0a256b231eafad4b1fa28772c6ac`.
- The edit script aborted unless the source digest matched exactly, required exactly one target row, and compared all other raw rows before writing. The file was checked clean before this edit; no concurrent canonical-export edit was present.
- Allowed target fields only: `status`, removal of `closed_at`, removal of `close_reason`, `updated_at`, and `notes`. The note preserves the accepted append-only register source category and temporal policy, states that the original acceptance criteria are unchanged, and identifies the still-open concrete register identity/trust anchor, authenticated source interface and snapshot/high-watermark continuity, immutable source append times, and grant/revocation writer authorization.

## Guarded-workflow attempts and limits

Before this explicit owner authorization, the guarded disposable-store workflow failed closed. A status-only plan was rejected before Beads commands because `status` is not an allowed generic update field. A description-only plan advanced through isolated import, then stopped before applying updates because source-to-baseline validation found an unplanned `closed_at` delta on `dap-jo9.1.6.7`. No importer-loss exception was added and no timestamp or null value was normalized. The pinned `create` path used a disposable store/home; `promote` was not run. The canonical source SHA remained unchanged through those attempts.

The present correction therefore uses the narrowly authorized, directly reviewed export exception above. It still does not change, initialize, migrate, or claim to verify the ignored embedded Dolt database. The live store and its history remain divergent/unverified relative to this interchange export.

## Decision and documentation alignment

Decision 0015 accepts the owner-approved append-only register as the source category and defines temporal rules. It does not authenticate a specific register or establish an implemented adapter. Reopening this issue preserves its original acceptance criteria and records unfinished requirements: an approved concrete register identity and trust anchor; a trusted adapter/interface proving source authenticity, immutable append times, complete snapshot/high-watermark continuity, and re-verifiable evidence; and controls authorizing grant/revocation writers. `.9.2` remains dependent on the reopened issue. No provider deployment or named release-approver identity is selected or required here.

The roadmap and reviewer-evidence gate statements were updated to stop calling `.1.1.6` closed and to describe the unchanged dependency and open controls. Historical dated snapshots were left intact as historical records.

## Independent review and validation

- An independent reviewer reviewed the exact candidate export SHA-256 above and found no issues: only the target row changed; its original description, acceptance criteria, dependencies, counts, and all other preserved fields remain unchanged; all other raw rows (including `.1.6.7` explicit null) are byte-identical; the two current docs agree with the reopened tracker status.
- Canonical/graph validation passed: all 101 JSONL rows parse, IDs are unique, exactly the authorized target field delta is present, all dependency edges/counts are unchanged, and every non-target raw row is byte-identical. `git diff --check` passed.
- `bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh` and `bash .github/scripts/test-source-publisher.sh` passed.
- Local Go checks were attempted (`go test -race ./...` and `go run ./cmd/traceability-check`) but this environment's `/usr/bin/go` is not the Go CLI and returned `Go: Unknown option: test` / `Go: Unknown option: run`; `go version` likewise returned `Go: Unknown option: version`. Root and nested Go suites/traceability therefore remain for exact-head hosted CI.
- No Beads live-store mutation or promotion was performed. At the time this exception record was first added, hosted exact-head CI and fresh review were pending; use the PR’s latest-head checks for current status.
