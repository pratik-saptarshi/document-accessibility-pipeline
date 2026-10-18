# Pinned cloud-tools bootstrap source archive

This directory preserves the pinned bootstrap source archives. The currently
pinned archive is:

- Source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
  (PR #51 squash-merge, the dap-5hy.2 accepted bounded bootstrap source)
- Original acquisition: downloaded 2026-10-17 from the official URL below
- Acquisition URL: <https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz>
- Stored size: 388,215 bytes
- SHA-256: `f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d`

The stored bytes are the official GitHub archive for the source revision
above, repackaged once through the membership denylist
(`internal/devsecops/source_archive_denylist.go`, R5-09): the `.beads/`
issue-tracker state and the three members whose contents matched personal
contact-data patterns were removed, and the remaining members were rewritten
with deterministic gzip settings. Re-applying the denylist to the stored
bytes removes nothing further. Its contents are the public repository source
at the revision above; that source includes the repository's
[MIT `LICENSE`](../../LICENSE). The archive contains the bounded bootstrap
subprocess output capture accepted in dap-5hy.2 (PR #51).

The `.gitattributes` rules mark this directory `export-ignore` for future Git
archives, so official GitHub source archives never embed a stored source
archive, and mark `.beads/**` `export-ignore` so future GitHub-generated
archives never publish issue-tracker state. Other bundle producers must
explicitly exclude `buildtools/source-archives/` and `.beads/`, and verify the
generated member list contains no prior source archives; never include a
stored bundle recursively in a later bundle.

The previous pinned archive (`cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz`, 259,355 bytes, SHA-256 `afeb27cdf51795168e7e606cd736d20988aecfba85949f5d36af671d8b8eeaed`) is retained for history. It received the same denylist repackaging. It is superseded by the archive above; the manifest and entrypoint pins were updated to the archive above in the follow-up repin change.

From a repository checkout, verify the pinned bytes and bounded archive structure with:

```sh
go test ./internal/devsecops -run '^TestPinnedBootstrapSourceArchive'
```
