# Poppler CI acquisition repair

**Status: PR validation pending.** This repairs CI tooling only and does not
change product scope. Local shell regressions and independent review passed;
the full Go checks and hosted signed-acquisition gates remain pending.

## Evidence and scope

Run `39695006147`, job `119550135323`, stopped during Poppler acquisition with
`expected exactly one Package field in authenticated APT metadata`, skipping
15 later validations. At base commit
`4ca75e98f2df17f1f54f16c9d4900f492b27c35a`, the installer pinned
`poppler-utils=24.02.0-1ubuntu9.10`. The supplied investigation found only
`24.02.0-1ubuntu9.11` in current Noble updates/security indexes and reproduced
successful empty output for the unavailable exact pin with APT 3.0.3. This is
evidence for present-day pin unavailability, not proof of the historical
runner's cause: its APT metadata was not retained, and the supplied current
index evidence was not signature-validated. Hosted CI must establish signed
acquisition.

## Requirements and non-goals

- Detect successful `apt-cache show` calls that return empty or whitespace-only
  data and report that no metadata was returned for the exact package, version,
  and architecture. Keep command failures distinct and visible.
- Refresh the reviewed exact pin consistently to
  `poppler-utils=24.02.0-1ubuntu9.11`. Do not select a latest version, installed
  copy, or arbitrary first record. Pins may need future reviewed refreshes.
- Preserve source and Ubuntu keyring validation, isolated APT state, exact
  identity/path checks, authenticated-index SHA256 comparison, downloaded
  package identity, dependency source context, least-privilege state, and
  cleanup.
- Keep this a CI-only GPL tool repair. Do not change product behavior,
  install Neuro/models/credentials, edit Beads data, or publish.

## Acceptance criteria

1. Empty and whitespace-only successful metadata responses fail with a bounded
   diagnostic naming package, version, and architecture; do not query policy,
   download, install, or write a success summary.
2. An `apt-cache show` command failure remains distinguishable and is not
   mislabeled as an empty successful response.
3. The `.11` exact package pin is consistent across installer, tests, and CI
   policy text. No fallback behavior is introduced.
4. Existing provenance and security rejection cases continue to pass, with
   stage evidence reset per case and exact package/version/architecture
   operands asserted. Mocked normal and injected-failure cases verify that
   the cleanup command ran and the isolated APT directory was removed.
5. Deterministic acquisition tests run before live package installation in
   CI, and the ordering is protected by a workflow regression assertion.
6. Local required root and nested-module checks, formatting, traceability, and
   source-publisher tests pass where the environment supports them. Independent
   local diff review is complete. Hosted exact-head CI and post-merge main CI
   remain separate gates.

## Test plan

First add red cases for empty and whitespace-only successful `show` output,
query failure, exact query operands, and no later-stage calls or summary. Then
implement the smallest diagnostic and preserve native command failure. Expand
the harness to realistic full package stanzas and stage-scoped evidence; cover
duplicate fields/stanzas, missing fields, wrong identity/version/architecture/
path, malformed or mismatched digests, downloaded identity mismatch, foreign
source/keyring, update/download/install failures, and cleanup. An installed
status entry must never provide a fallback. Inject forbidden policy, download,
and install log rows to prove the stage-absence assertion itself rejects each
one even inside a conditional. Verify the intended call trace and expected
diagnostic for update, download, install, duplicate-stanza, and missing-field
failures; assert that the isolated APT directory no longer exists after normal
and injected-failure exits. This does not inject a cleanup-command failure.
Run shell syntax and the focused harness, then the repository's required root
and nested-module checks. These
mock tests do not establish real signed APT integration; hosted CI must retain
package/version/architecture/source/digest evidence and complete cleanup.
