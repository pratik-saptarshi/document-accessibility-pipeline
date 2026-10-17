# Autonomous engineering platform fit assessment

**Status: proposed workflow guidance only. No upstream code, skills, scheduler, service, or memory backend is installed or approved by this note.** Reviewed 2026-10-17.

## Reference and provenance

The reference is [`pratik-saptarshi/autonomous-engineering-platform` at commit `db4dd93780aa98916d0002c89bed16a616fc0490`](https://github.com/pratik-saptarshi/autonomous-engineering-platform/tree/db4dd93780aa98916d0002c89bed16a616fc0490). Its pinned tree contains five files: `README.md`, `docs/runbook.md`, `docs/operations.md`, `docs/limitation-analysis.md`, and `scripts/watchdog-backup.sh`. No license file is present in that tree. This assessment contains original, project-specific findings and links; it does not copy the source runbook or script. Do not redistribute or port their text or code without license clearance.

The source describes a persistent Linux VM with 29 skills, a Markdown backlog, scheduled task loop, NeuroStrata memory, and Google Drive backup. Those are claims about the reference author's environment, not verified properties of this repository or this cloud executor. Source pointers are pinned to the reviewed commit:

- [Reference README](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/db4dd93780aa98916d0002c89bed16a616fc0490/README.md)
- [Reference runbook](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/db4dd93780aa98916d0002c89bed16a616fc0490/docs/runbook.md)
- [Reference operations guide](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/db4dd93780aa98916d0002c89bed16a616fc0490/docs/operations.md)
- [Reference limitation analysis](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/db4dd93780aa98916d0002c89bed16a616fc0490/docs/limitation-analysis.md)
- [Reference watchdog script](https://github.com/pratik-saptarshi/autonomous-engineering-platform/blob/db4dd93780aa98916d0002c89bed16a616fc0490/scripts/watchdog-backup.sh)

## Fit with this repository

This repository already has the parts of the workflow that fit: Beads is the canonical work tracker, [`AGENTS.md`](../../AGENTS.md) requires reading decisions and prior evidence, the existing `dap-` skills cover test-first changes, review, and verification, and the guarded publisher and protected CI govern publication. The compatible addition is a bounded-work rule: handle one authorized work item at a time; for an explicitly autonomous or repeated run, stop after six task iterations, or after three unsuccessful attempts on the same item, preserve the partial state, and report the blocker. These limits trigger escalation; they do not waive acceptance criteria or close work.

After verification, record a short result and reusable finding in the canonical Beads issue or relevant project document. Keep it sanitized and linked to evidence. Do not add a second Markdown backlog, send repository or Beads content to an external memory service, or install the reference's 29-skill library. This repository's existing skills have their own provenance and review record; use those rather than duplicate upstream packages.

The reference's runtime architecture does not fit the current approved environment. Its scheduler, NeuroStrata database, custom MCP client, connector gateway, Google Drive backup, and egress configuration would add persistent services, data movement, or network policy. None is authorized here. The accepted permanent Neuro model and authentication blocker remains in force; this note does not reopen it. Keep Go as the primary implementation language, Beads as the work tracker, and Java/veraPDF in the isolated CI job.

## Reference reliability and security limits

The pinned material is not a ready-to-install package. It has no license file, automated tests, CI workflow, or general installer in the reviewed tree. The runbook downloads the Nomic model through a mutable `resolve/main` URL even though it separately names a model commit. Its Beads example saves the archive under `beads.tar.gz` but checks the checksum list using the release asset filenames, so the documented checksum command may not refer to the saved archive. Do not execute the runbook.

The source documents also disagree about daemon auto-start: the README describes MCP-spawn recovery, while the runbook and limitation analysis say it was unreliable and require direct restart or a watchdog. The limitation analysis characterizes the egress variable as a DNS/SSRF exception, not a network firewall. The watchdog is specific to that VM's paths and Drive destination; its PID selection uses the first process matching a pattern, and its status checks can log a failed restart without making the whole run fail. It has no tested restore procedure or remote-retention proof. These are review findings about the pinned files, not claims that this repository has those services or failure modes.

## Current cloud snapshot

Observed in this executor on 2026-10-17, without starting services or loading models:

- The repository-local guidance and `dap-` workflow skills are present; the project's `AGENTS.md` is the controlling instruction file.
- The workspace-local cloud build-tools directory contains Go, `bd`, `actionlint`, `govulncheck`, `go-licenses`, `gitleaks`, and the NeuroStrata and NeuroPlasticity binaries.
- `codex mcp list` reported that no MCP servers are configured. No Neuro daemon process was running. The expected workspace model-cache and Neuro config/data roots recorded in the existing tooling proposal were absent. Podman and the reference's `hatch_gws_cli` were not on `PATH`; Docker availability does not satisfy the project's rootless-Podman approval gate.

This is a point-in-time inventory, not an attestation of saved-environment settings or future persistence. The tool binaries do not make either Neuro tool ready. See [build-machine tooling](../planning/build-machine-tooling.md) for the existing readiness gates. No runtime, model, authentication, MCP, scheduling, egress, or backup configuration was changed.
