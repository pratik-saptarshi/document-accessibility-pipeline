# Pinned cloud-tools bootstrap source archives

This directory preserves the original bootstrap archives at their historical
paths and keeps denylist-filtered replacements as distinct artifacts. The raw
archives remain unchanged for historical evidence and the benchmark fixture.
The active no-checkout pin is the filtered `ef4c356` artifact described
below, repinned by a synchronized update of both manifests, the bootstrap
shell entrypoint, and the pinned-source tests after the artifact reached
protected main.

## Current bootstrap pin

- Contained source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
  (PR #51 squash-merge, the dap-5hy.2 accepted bounded bootstrap source)
- Filtered artifact (the active no-checkout pin):
  `cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz`
- Filtered size and SHA-256: 388,215 bytes;
  `f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d`
- Protected-main storage commit: `4ce49c0b4face9388f036e08f3e05af7a3bf58a2`
  (PR #153)

The filtered replacement is the official GitHub archive for the source
revision above, repackaged once through the membership denylist
(`internal/devsecops/source_archive_denylist.go`, R5-09). It omits `.beads/`
and members matching the recorded personal contact-data patterns. Re-applying
the denylist removes nothing further. Both manifests, the bootstrap shell
entrypoint, and the pinned-source tests pin this artifact by its exact path,
byte size, and SHA-256, with the raw URL pinned to the protected-main
storage commit. The archived source includes the repository MIT `LICENSE`.

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
