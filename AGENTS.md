# Repository guidance

- This repository is planning scaffolding. Do not infer that a proposed tool, PDF engine, architecture, compliance target, or schedule has been approved.
- Use synthetic or deidentified fixtures only. Do not add prior private reports or sample documents.
- Keep existing-PDF remediation and document generation as separate workflow lanes. Both lanes include statements, invoices, correspondence, and reports.
- Remediation dependencies must remain open-source-only. Enterprise-only tagging or PDF/UA export is out of scope.
- Accessibility claims require independent machine validation plus human semantic and assistive-technology review. Tagged output alone is not a compliance result.
- Track work in Beads. Run `bd prime` for current workflow context, `bd ready` to find unblocked work, and `bd show <id>` before changing an issue. Do not replace Beads with Markdown TODO lists.
- After issue changes, refresh `.beads/issues.jsonl` with `bd export -o .beads/issues.jsonl`; this tracked interchange export preserves issue records and dependencies for fresh-clone recovery, but is not a full Dolt backup. See `CONTRIBUTING.md` for restore steps.
- For each implementation slice, write failing tests first, make the smallest change that passes, then refactor while tests stay green. Do not implement product features under a planning-only request.
- Preserve existing Git hooks and local user work. Never push, publish, or open a PR unless asked.
- For authorized agent-authored source publication, use `scripts/source-publisher.sh commit` and `scripts/source-publisher.sh publish` as documented in `CONTRIBUTING.md`. This sets identity only in the current repository, verifies the new source commit and HTTPS fast-forward push before publication, then checks the fetched remote commit. Do not rewrite or reclassify existing approved history; GitHub-generated merge commits are reported separately and are not attested as CloudBot commits.
