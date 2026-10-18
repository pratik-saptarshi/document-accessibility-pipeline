# Pinned cloud-tools bootstrap source archives

This directory preserves the original bootstrap archives at their historical
paths and keeps denylist-filtered replacements as distinct artifacts. The raw
archives remain unchanged for historical evidence and the benchmark fixture.
The filtered artifacts below are prepared for review; the active no-checkout
pin remains the original `ef4c356` archive until a protected-main storage
commit and separate synchronized pin update are available.

## Current bootstrap pin

- Contained source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
  (PR #51 squash-merge, the dap-5hy.2 accepted bounded bootstrap source)
- Original official archive: `cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz`
- Original size and SHA-256: 469,726 bytes;
  `16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473`
- Filtered replacement artifact:
  `cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz`
- Filtered size and SHA-256: 388,215 bytes;
  `f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d`

The filtered replacement is the official GitHub archive for the source
revision above, repackaged once through the membership denylist
(`internal/devsecops/source_archive_denylist.go`, R5-09). It omits `.beads/`
and members matching the recorded personal contact-data patterns. Re-applying
the denylist removes nothing further. The manifests and bootstrap shell
entrypoint still pin the original archive. A later update must use the
filtered artifact's protected-main storage commit and then synchronize its
exact path, byte size, and SHA-256 in both manifests, the shell entrypoint,
and tests. The archived source includes the repository MIT `LICENSE`.

## Previous bootstrap source

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
go test ./internal/devsecops -run '^(TestPinnedBootstrapSourceArchive|TestHistoricalBootstrapSourceArchiveBytesPreserved|TestFilteredBootstrapSourceArchiveReplacement|TestFilteredBootstrapSourceArchiveHasNoDenylistedMembers)$'
```
