# Autonomous engineering platform fit assessment

**Status: proposed workflow guidance only. No upstream code, skills, scheduler, service, or memory backend is installed or approved by this note.** Reviewed 2026-10-17; source checked through `main` at `2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1`.

## Reference and provenance

The reference is [`pratik-saptarshi/autonomous-engineering-platform` at commit `2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1`](https://github.com/pratik-saptarshi/autonomous-engineering-platform/tree/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1). Its reviewed tree contains six files: `AGENTS.md`, `README.md`, `docs/runbook.md`, `docs/operations.md`, `docs/limitation-analysis.md`, and `scripts/watchdog-backup.sh`. No license file is present in this tree; the added `AGENTS.md` does not supply a license. This assessment contains original, project-specific findings and links; it does not copy the source runbook or script. Do not redistribute or port their text or code without license clearance.

The source describes a persistent Linux VM with 29 skills, a Markdown backlog, scheduled task loop, NeuroStrata memory, and Google Drive backup. Those are claims about the reference author's environment, not verified properties of this repository or this cloud executor. The reviewed history includes merged PR #3 (`b7f1db6`), which changes the watchdog to signal every matching daemon PID during stop, and subsequent PR #5 (`2c7b0db`), which changes its lifecycle to on-demand, idle shutdown, and log archiving. These upstream changes are assessed as documentation and source evidence only; none is installed or approved here. Source pointers are pinned to the reviewed commit:

- [Reference contributor instructions](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/AGENTS.md)
- [Reference README](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/README.md)
- [Reference runbook](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/docs/runbook.md)
- [Reference operations guide](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/docs/operations.md)
- [Reference limitation analysis](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/docs/limitation-analysis.md)
- [Reference watchdog script](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/2c7b0dbe1c854e06fbc5085f5ed9b57333adffd1/scripts/watchdog-backup.sh)
- [PR #3 watchdog PID fix](https://github.com/pratik-saptarshi/autonomous-engineering-platform/commit/b7f1db6)
- [PR #5 on-demand watchdog/log changes](https://github.com/pratik-saptarshi/autonomous-engineering-platform/commit/2c7b0db)

## Fit with this repository

This repository already has the parts of the workflow that fit: Beads is the canonical work tracker, [`AGENTS.md`](../../AGENTS.md) requires reading decisions and prior evidence, the existing `dap-` skills cover test-first changes, review, and verification, and the guarded publisher and protected CI govern publication. The compatible addition is a bounded-work rule: handle one authorized work item at a time; for an explicitly autonomous or repeated run, stop after six task iterations, or after three unsuccessful attempts on the same item, preserve the partial state, and report the blocker. These limits trigger escalation; they do not waive acceptance criteria or close work.

After verification, record a short result and reusable finding in the canonical Beads issue or relevant project document. Keep it sanitized and linked to evidence. Do not add a second Markdown backlog, send repository or Beads content to an external memory service, or install the reference's 29-skill library. This repository's existing skills have their own provenance and review record; use those rather than duplicate upstream packages.

The reference's runtime architecture does not fit the current approved environment. Its scheduler, NeuroStrata database, custom MCP client, connector gateway, Google Drive backup, and egress configuration would add persistent services, data movement, or network policy. None is authorized here. The upstream `AGENTS.md` also specifies the reference repository's commit identity and branch naming; those are not adopted as instructions or credentials for this project. The accepted permanent Neuro model and authentication blocker remains in force; this note does not reopen it. Keep Go as the primary implementation language, Beads as the work tracker, and Java/veraPDF in the isolated CI job.

## Reference reliability and security limits

The pinned material is not a ready-to-install package. It has no license file, automated tests, CI workflow, or general installer in the reviewed tree. The runbook downloads the Nomic model through a mutable `resolve/main` URL even though it separately names a model commit. Its Beads example saves the archive under `beads.tar.gz` but checks the checksum list using the release asset filenames, so the documented checksum command may not refer to the saved archive. Do not execute the runbook.

The source still has operational-documentation drift. The README and watchdog's top-of-file comments say it revives a dead daemon, and the runbook suggests restarting directly or waiting for the watchdog; the current executable logic instead treats a down daemon as the expected idle state and exits without restarting it. PR #5 changed the script to stop a healthy daemon after an idle interval, keep timestamped daemon logs, and archive logs with database backups. PR #3 fixed the earlier stop race by signalling every matching PID; the old first-PID stop behavior must not be attributed to the current stop logic. A failed 30-second stop is logged and causes the backup to be skipped, while the script's caller exits successfully. The watchdog remains tied to that VM's paths and Drive destination, and the reviewed source has no tested restore procedure or remote-retention proof. The runbook still uses a mutable model `resolve/main` URL despite separately naming a model commit, and its Beads checksum command checks the release filenames rather than the saved `beads.tar.gz` name. These are review findings about the pinned files, not claims that this repository has those services or failure modes.

## Current cloud snapshot

Observed in this executor on 2026-10-17, without starting services or loading models:

- The repository-local guidance and `dap-` workflow skills are present; the project's `AGENTS.md` is the controlling instruction file.
- The workspace-local cloud build-tools directory contains Go, `bd`, `actionlint`, `govulncheck`, `go-licenses`, `gitleaks`, and the NeuroStrata and NeuroPlasticity binaries.
- `codex mcp list` reported that no MCP servers are configured. No Neuro daemon process was running. The expected workspace model-cache and Neuro config/data roots recorded in the existing tooling proposal were absent. Podman and the reference's `hatch_gws_cli` were not on `PATH`; Docker availability does not satisfy the project's rootless-Podman approval gate.

This is a point-in-time inventory, not an attestation of saved-environment settings or future persistence. The tool binaries do not make either Neuro tool ready. See [build-machine tooling](../planning/build-machine-tooling.md) for the existing readiness gates. No runtime, model, authentication, MCP, scheduling, egress, or backup configuration was changed.
