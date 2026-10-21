# Pinned cloud-tools bootstrap source archives

This directory preserves the original bootstrap archives at their historical
paths and keeps denylist-filtered replacements as distinct artifacts. The raw
archives remain unchanged for historical evidence and the benchmark fixture.
The active no-checkout pin is the filtered production-only bootstrap module
artifact described below, pinned after the artifact reached protected main.

## Current bootstrap pin

- Contained source revision: `9ee727854c9fb974a481a61a9e1657f4a05d93e0`
  (PR #173 source head with the Go 1.27.3 bootstrap pin)
- Filtered artifact (the active no-checkout pin):
  `cloud-tools-bootstrap-9ee727854c9fb974a481a61a9e1657f4a05d93e0-filtered-621636861060e435.tar.gz`
- Filtered size and SHA-256: 13,767 bytes;
  `621636861060e43558213acf64eef18d1a47300f09ff8eeea0586825dfd43949`
- Protected-main storage commit: `75d9f59615b89606e61e3e5f92435c6fcf0459d6`
  (PR #177)

The filtered replacement contains the LICENSE and production files from the
bootstrap module at the source revision above. The complete official GitHub
archive exceeds the existing 4 MiB expanded-size filter bound, so this
minimal source bundle was constructed from those files and passed through the
existing membership denylist (`internal/devsecops/source_archive_denylist.go`,
R5-09). It omits tests, `.beads/`, prior archive bundles, and unrelated
repository content. The bootstrap module manifest and `go.mod` both pin Go
1.27.3. Both manifests, the bootstrap shell entrypoint, and the pinned-source
tests pin this artifact by exact path, byte size, and SHA-256, with the raw URL
pinned to the protected-main storage commit.

## Previous bootstrap source

- Contained source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
  (PR #51 squash-merge, the dap-5hy.2 accepted bounded bootstrap source)
- Original official archive:
  `cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz`
- Original size and SHA-256: 469,726 bytes;
  `16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473`
- Superseded by the filtered artifact above; replaced as the active
  no-checkout pin in the synchronized 2026-10-18 repin, which pins the
  protected-main storage commit `4ce49c0b4face9388f036e08f3e05af7a3bf58a2`.

- The former filtered `ef4c356` artifact remains available at its immutable
  path and is retained for historical verification; it is no longer the
  active no-checkout pin.

- Contained source revision: `63126024666fd71b48c25df55d1e98924e5eb09a`
- Original official archive:
  `cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz`
- Original size and SHA-256: 306,734 bytes;
  `6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5`
- Distinct filtered replacement:
  `cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a-filtered-afeb27cdf5179516.tar.gz`
- Filtered size and SHA-256: 259,355 bytes;
  `afeb27cdf51795168e7e606cd736d20988aecfba85949f5d36af671d8b8eeaed`

## Archive scope

The `.gitattributes` rules mark this directory `export-ignore` for future Git
archives, so official GitHub source archives never embed stored source
archives, and mark `.beads/**` `export-ignore` so future GitHub-generated
archives never publish issue-tracker state. Other bundle producers must
explicitly exclude `buildtools/source-archives/` and `.beads/`, and verify the
generated member list contains no prior source archives; never include a
stored bundle recursively in a later bundle.

From a repository checkout, verify the active pin and historical byte
identities with:

```sh
go test ./internal/devsecops -run '^(TestPinnedBootstrapSourceArchive|TestPinnedBootstrapSourceArchivePairs|TestFilterSourceArchiveDropsDenylistedMembers|TestFilterSourceArchiveRejectsUnsafeMemberLayouts|TestFilteredBootstrapSourceArchiveHasNoDenylistedMembers)$'
```
