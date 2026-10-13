# Pinned cloud-tools bootstrap source archive

This file preserves the exact official GitHub HTTPS archive downloaded for source commit `63126024666fd71b48c25df55d1e98924e5eb09a`:

- Original acquisition: downloaded 2026-10-13 from the official URL below
- Acquisition URL: <https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz>
- Downloaded size: 306,734 bytes
- SHA-256: `6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5`

The bytes were copied from the original downloaded archive without reconstruction or recompression. GitHub-generated archives can have different compressed bytes for the same retained source commit, so keeping this reviewed blob in Git preserves the exact byte sequence used by the checksum pin. Its contents are the public repository source at the revision above; that source includes the repository's [MIT `LICENSE`](../../LICENSE).

The `.gitattributes` rule marks this directory `export-ignore` for future Git archives. Other bundle producers must explicitly exclude `buildtools/source-archives/` and verify the generated member list contains no prior source archives; never include a stored bundle recursively in a later bundle.

From a repository checkout, verify the pinned bytes and bounded archive structure with:

```sh
go test ./internal/devsecops -run '^TestPinnedBootstrapSourceArchive$'
```
