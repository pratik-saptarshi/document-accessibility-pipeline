# Imported agent skills

This repository carries a small, adapted set of Markdown workflow guides in
[`../../.agents/skills/`](../../.agents/skills/). The repository-local index in
[`AGENTS.md`](../../AGENTS.md) names each skill, its trigger, and its direct
`SKILL.md` link. Load that file and the references it links only when the task
matches its description; no installer, bootstrap, or global configuration is
required.

## Pinned sources and license

| Source | Pin | License | Imported guidance |
|---|---|---|---|
| [obra/superpowers](https://github.com/obra/superpowers) | `bb92a77741419a4ab5f06e711a283343f1ada0c3` (v7.0.0) | MIT, Jesse Vincent, 2025; full notice at [local copy](../../.agents/skills/licenses/superpowers-MIT.txt) | TDD, systematic debugging, verification, requesting/receiving review, bounded parallel work |
| [pratik-saptarshi/opencode-rountable-review](https://github.com/pratik-saptarshi/opencode-rountable-review) | `dee1d66ea252fa3b31d5073b84dd4a947d881916` | MIT, Huiyan Wan, 2026; full notice at [local copy](../../.agents/skills/licenses/roundtable-MIT.txt) | Adversarial review panel and review-to-plan integration |

The files are concise local adaptations, not a verbatim mirror. They use unique
`dap-` names, repository paths and Go/Beads/review rules, and explicitly state
the instruction hierarchy and lack of action authority. The source file blob
IDs and SHA-256 values for every imported local file are recorded in
[`skill-provenance.json`](../../.agents/skills/skill-provenance.json).

## Scope and exclusions

Only the skills listed in `AGENTS.md` are included. No upstream install scripts,
executables, commands, plugin manifests, hooks, packages, remote execution,
telemetry, or saved/global configuration are included. In particular, the
roundtable `excalibur-quest` skill and its helper are excluded. Instructions
that require Claude-specific models/tools, private reasoning logs, full
transcripts, CDN HTML, global memory edits, unconditional bootstrap, automatic
GitHub actions, or unbounded waits were removed or adapted.

## Updating

To update, inspect the upstream repository at a new immutable commit over HTTPS.
Review each source file and license, retain only the approved Markdown allowlist,
adapt it to current `AGENTS.md` and user/system boundaries, then update the
source commit, original path/blob ID, local SHA-256, and adaptation notes in the
manifest. Review the complete diff, verify frontmatter, links, allowlist and
absence of executable files/hooks, then publish only through the authorized
repository workflow for independent review and required CI. Do not merge until
those gates pass. Do not execute upstream installers or utilities.
