# Cloud build-tool bootstrap

**Status: proposed and reviewable.** This setup is for an isolated cloud build machine. It does not change host configuration, the product architecture, or shipped runtime images.

## Install and verify

The reviewed entrypoint is [`scripts/bootstrap-cloud-build-tools.sh`](../../scripts/bootstrap-cloud-build-tools.sh). It supports Linux x86_64 only. The first stage installs Go 1.27.2 from go.dev after checking the fixed SHA-256. A standard-library-only Go command then reads the self-contained [tool manifest](../../buildtools.manifest.json), downloads only pinned official release assets, verifies their digests, and installs into workspace-local directories. Go-installed tools are fetched from their pinned source modules with the public Go checksum database enabled. The Go manifest is embedded in the command so a saved environment can install it from an immutable module revision without a repository checkout.

The bootstrap defaults are:

- Tools and toolchain: `/workspace/.codex-tools/cloud-build-tools/`
- Downloads and Go caches: `/workspace/.cache/cloud-build-tools/`
- Activation: `/workspace/.codex-tools/cloud-build-tools/activate.sh`

`CODEX_BUILD_TOOLS_ROOT` and `CODEX_BUILD_CACHE_ROOT` override these paths. The entrypoint acquires an atomic directory lock; a competing run waits up to two minutes. It does not remove a lock on timeout because another process may own it. Inspect the recorded PID and running processes before removing a stale lock. Each staged binary must pass its version check or exact release digest check before atomic publication. A rerun verifies already installed versions and reuses checksum-verified cached assets.

To run from a repository checkout:

```bash
scripts/bootstrap-cloud-build-tools.sh
source /workspace/.codex-tools/cloud-build-tools/activate.sh
go version
bd version
actionlint -version
govulncheck -version
gitleaks version
```

For a full missing-tool report in a saved Start skill, use this loop after sourcing `activate.sh`:

```bash
for tool in go bd actionlint govulncheck go-licenses gitleaks neurostrata-mcp neuroplasticity; do
  command -v "$tool" || printf 'MISSING: %s\n' "$tool"
done
```

For an environment with no repository checkout, the saved Install script should perform the small Go archive stage and then `go install` the published `cmd/cloud-tools-bootstrap` module at the immutable reviewed commit. The exact commit value is supplied with the draft pull request handoff; do not substitute a branch or `latest`. The publisher should place the same entrypoint in each saved environment's additive Install setup and put the activation plus command-discovery checks in its Start skill. Repo refresh alone does not rerun Install, and the Start instructions must verify command resolution for each fresh task. No saved environment configuration is modified by this repository change.

## Pins and boundaries

| Tool | Pin | Acquisition and version evidence | License / use boundary |
|---|---:|---|---|
| Go | 1.27.2 | Official Linux amd64 archive; SHA-256 in manifest; `go version` | BSD-3-Clause; build machine |
| Beads CLI (`bd`) | 1.3.1 | Official Linux amd64 release archive and official `checksums.txt`; `bd version` | MIT; issue tracking CLI |
| NeuroStrata | 1.9.0 | Official release archive SHA-256; `--version` | MIT binary only; installing does not configure memory or models |
| NeuroPlasticity | 2.1.0 | Official release binary SHA-256; CLI has no version command, so the exact asset digest is the version proof | MIT binary only; installing does not configure evaluation, containers, or models |
| actionlint | 1.7.12 | Official Linux amd64 release archive and official checksums file; version output | MIT; workflow lint |
| govulncheck | 1.9.0 | Pinned Go module; Go checksum database; version output | BSD-3-Clause; vulnerability checks |
| go-licenses | 2.0.1 | Pinned Go module; Go checksum database and embedded Go build metadata | Apache-2.0; dependency-license checks |
| Gitleaks | 8.30.1 | Official Linux x64 release archive SHA-256; `version` output | MIT; secret scanning |
| veraPDF CLI | 1.30.3 | Independently authenticated installer, installer SHA-256 `b512bf24945503a33fc10738ecffc1f2dccaaf8593a448e94ee9ac86a9559a59`, CLI JAR SHA-256 `19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70`, signer fingerprint `13DD102B4DD69354D12DE5A83184863278B17FE7`; see [the existing CI provisioner in PR #7](https://github.com/pratik-saptarshi/document-accessibility-pipeline/blob/c8e1a5b0a2884204453f11f0f558d8f73e0f1a9e/.github/scripts/provision-verapdf.sh) | GPL-3.0-or-later or MPL-2.0 per upstream POM; only the existing CI workflow elects MPL-2.0; never install in the persistent build-tool root or app runtime |

The manifest keeps the existing version decisions and checksums from [build-machine tooling research](build-machine-tooling.md) and [security policy](../security/README.md). Changing any version requires updating the manifest, embedded manifest, and verification tests together. Unknown or unreviewed licenses are not silently accepted.

## Operating boundaries

- This bootstrap installs executables only. It does not fetch NeuroStrata embedding weights, NeuroPlasticity model weights, or any other model/runtime data; configure hosts, credentials, MCP servers, hooks, daemons, databases, or network access; or run either Neuro tool.
- Do not treat installation as model/runtime readiness. Existing host and model gates remain those recorded in [build-machine tooling](build-machine-tooling.md).
- veraPDF is not installed by this persistent build-tool bootstrap. The existing CI-only provisioner performs authenticated acquisition and version/hash checks in a temporary runner directory. Reuse that script; do not create a second veraPDF installer or keep its JAR in the persistent build-tool root.
- Tool binaries stay outside the checkout and product runtime. The activation file only prepends the isolated Go and tool directories when sourced, and avoids adding the same directory twice. Nothing writes a global profile or changes system PATH.
- If module source or checksum-database access fails for a tool, that tool remains uninstalled; the bootstrap reports the failure, continues with independent assets, and exits nonzero. It never turns off sum verification or falls back to an unpinned binary.

## Tests and diagnostics

The standalone bootstrap module's tests cover strict manifest parsing, checksummed download rejection and cleanup, version mismatch rejection, repeated-run idempotency, safe archive extraction, concurrent-lock timeout behavior, CI-only veraPDF exclusion, and parity between the embedded and repository manifests. Useful checks are:

```bash
(cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./...)
bash -n scripts/bootstrap-cloud-build-tools.sh
```

The command prints per-tool `verified`, `installed`, or `external` status. It records tool-specific acquisition errors and continues with other tools, then exits nonzero if any installation failed. For diagnosis, check the selected root/cache, ensure required `curl`, `sha256sum`, `tar`, and `timeout` utilities exist, verify disk space and outbound access to the exact HTTPS endpoints in the manifest, and inspect a lock path before removing it. Module builds use `GOTOOLCHAIN=local`, `GOSUMDB=sum.golang.org`, isolated Go caches, and direct VCS retrieval to avoid the blocked module-proxy redirect observed in this workspace. Direct source resolution for the current `govulncheck` and `go-licenses` pins was also denied with HTTP 403 for `golang.org`, `go.opencensus.io`, and `k8s.io`. Errors suppress subprocess output because proxy diagnostics may contain signed download URLs.
