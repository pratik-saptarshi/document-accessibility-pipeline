# Pinned cloud-tools bootstrap source archive

This directory preserves exact official GitHub HTTPS archives. The currently
pinned archive is:

- Source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
  (PR #51 squash-merge, the dap-5hy.2 accepted bounded bootstrap source)
- Original acquisition: downloaded 2026-10-17 from the official URL below
- Acquisition URL: <https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz>
- Downloaded size: 469,726 bytes
- SHA-256: `16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473`

The bytes were copied from the original downloaded archive without reconstruction or recompression. GitHub-generated archives can have different compressed bytes for the same retained source commit, so keeping this reviewed blob in Git preserves the exact byte sequence used by the checksum pin. Its contents are the public repository source at the revision above; that source includes the repository's [MIT `LICENSE`](../../LICENSE). The archive contains the bounded bootstrap subprocess output capture accepted in dap-5hy.2 (PR #51).

The `.gitattributes` rule marks this directory `export-ignore` for future Git archives, so official GitHub source archives never embed a stored source archive. Other bundle producers must explicitly exclude `buildtools/source-archives/` and verify the generated member list contains no prior source archives; never include a stored bundle recursively in a later bundle.

The previous pinned archive (`cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz`, 306,734 bytes, SHA-256 `6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5`) is retained for history. It is superseded by the archive above; the manifest and entrypoint pins were updated to the archive above in the follow-up repin change.

From a repository checkout, verify the pinned bytes and bounded archive structure with:

```sh
go test ./internal/devsecops -run '^TestPinnedBootstrapSourceArchive$'
```
