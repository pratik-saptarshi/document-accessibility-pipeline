# Cloud build-tool bootstrap

**Status: proposed and reviewable.** This setup is for an isolated cloud build machine. It does not change host configuration, the product architecture, or shipped runtime images.

## Install and verify

The reviewed entrypoint is [`scripts/bootstrap-cloud-build-tools.sh`](../../scripts/bootstrap-cloud-build-tools.sh). It supports Linux x86_64 only. The first stage installs Go 1.27.2 from go.dev after checking the fixed SHA-256. Without a checkout, it downloads the bootstrap source archive at the immutable commit recorded in the [tool manifest](../../buildtools.manifest.json), verifies the pinned archive SHA-256, and builds the standard-library-only command locally. That command reads its embedded manifest, downloads only pinned official release assets, verifies their digests, and installs into workspace-local directories. Go-installed tools are fetched from their pinned source modules with the public Go checksum database enabled.

The bootstrap defaults are:

- Tools and toolchain: `/workspace/.codex-tools/cloud-build-tools/`
- Downloads and Go caches: `/workspace/.cache/cloud-build-tools/`
- Activation: `/workspace/.codex-tools/cloud-build-tools/activate.sh`

`CODEX_BUILD_TOOLS_ROOT` and `CODEX_BUILD_CACHE_ROOT` override these paths. The entrypoint acquires an atomic directory lock; a competing run waits up to two minutes. It does not remove a lock on timeout because another process may own it. Inspect the recorded PID and running processes before removing a stale lock. Each staged binary must pass its version check or exact release digest check before atomic publication. A rerun verifies already installed versions and reuses checksum-verified cached assets.

The shell entrypoint accepts no command-line arguments and rejects them before it inspects tools or writes files. Configure the tool and cache locations with `CODEX_BUILD_TOOLS_ROOT` and `CODEX_BUILD_CACHE_ROOT` environment variables; flags such as `--root` are not supported.

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

For the pinned binary verification in a saved Start skill, run the embedded verifier after sourcing `activate.sh`:

```bash
/workspace/.codex-tools/cloud-build-tools/bin/cloud-tools-bootstrap \
  --root /workspace/.codex-tools/cloud-build-tools \
  --cache /workspace/.cache/cloud-build-tools \
  --go /workspace/.codex-tools/cloud-build-tools/toolchains/go/1.27.2/bin/go \
  --verify
```

This read-only check executes each pinned version command or compares the configured artifact/executable digest. It reports `NOT VERIFIED` for missing, wrong-version, or wrong-digest binaries. It checks identity at run time; it is not a tamper-proof attestation of future file integrity. For veraPDF, it reports only installer-archive integrity; runtime readiness is not checked and PDF validation is not run by this bootstrap.

For command discovery only, use this loop after sourcing `activate.sh`:

```bash
for tool in go bd actionlint govulncheck go-licenses gitleaks neurostrata-mcp neuroplasticity; do
  command -v "$tool" || printf 'MISSING: %s\n' "$tool"
done
```

For an environment with no repository checkout, the saved Install script builds from the checksum-verified immutable source archive in the manifest; it does not depend on Go module version discovery. The publisher should place the same entrypoint in each saved environment's additive Install setup and put the activation plus command-discovery checks in its Start skill. Repo refresh alone does not rerun Install, and the Start instructions must verify command resolution for each fresh task. No saved environment configuration is modified by this repository change.

## Pins and boundaries

| Tool | Pin | Acquisition and version evidence | License / use boundary |
|---|---:|---|---|
| Go | 1.27.2 | Official Linux amd64 archive; SHA-256 in manifest; `go version` | BSD-3-Clause; build machine |
| Beads CLI (`bd`) | 1.3.1 | Official Linux amd64 release archive and official `checksums.txt`; `bd version` | MIT; issue tracking CLI |
| NeuroStrata | 1.9.0 | Official release archive SHA-256 `9232a9d56732413fa7612f16c96f61041d670d8370f730fa8ed87e7a8221a1cb`, extracted binary SHA-256 `9c029685bb881252b67de021d2ce5265e0acda0098532592d000a1586226166e`, and `--version` | MIT binary only; installing does not configure memory or models |
| NeuroPlasticity | 2.1.0 | Official release binary SHA-256; CLI has no version command, so the exact asset digest is the version proof | MIT binary only; installing does not configure evaluation, containers, or models |
| actionlint | 1.7.12 | Official Linux amd64 release archive and official checksums file; version output | MIT; workflow lint |
| govulncheck | 1.9.0 | Pinned Go module; Go checksum database; version output | BSD-3-Clause; vulnerability checks |
| go-licenses | 2.0.1 | Pinned Go module; Go checksum database and embedded Go build metadata | Apache-2.0; dependency-license checks |
| Gitleaks | 8.30.1 | Official Linux x64 release archive SHA-256; `version` output | MIT; secret scanning |
| veraPDF CLI | 1.30.3 | Official installer ZIP SHA-256 `b512bf24945503a33fc10738ecffc1f2dccaaf8593a448e94ee9ac86a9559a59`, CLI JAR SHA-256 `19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70`, signer fingerprint `13DD102B4DD69354D12DE5A83184863278B17FE7`; the persistent bootstrap caches and verifies only the signed installer ZIP; see the [authenticated CI provisioner](../../.github/scripts/provision-verapdf.sh) | GPL-3.0-or-later or MPL-2.0 per upstream notices; the existing CI workflow elects MPL-2.0; Java and the CLI runtime remain CI-only; never a product dependency or application runtime |

The manifest keeps the existing version decisions and checksums from [build-machine tooling research](build-machine-tooling.md) and [security policy](../security/README.md). Changing any version requires updating the manifest, embedded manifest, and verification tests together. Unknown or unreviewed licenses are not silently accepted.

## Operational readiness for augmentation

An installed executable is not a ready memory or evaluation service. The bootstrap reports NeuroStrata and NeuroPlasticity as **not verified** until their models, persistent data paths, agent integration, and runtime checks are completed. See [the build-machine readiness gates](build-machine-tooling.md#functional-readiness-and-required-approvals).

The minimum checks for a fresh environment are:

1. **NeuroStrata:** register the binary as a task-scoped MCP stdio server using its absolute path; give it a writable persistent workspace data/config/cache home outside the checkout; pin and make its local embedding model/runtime available; then verify MCP `initialize`, tool listing, and add/search against a synthetic `dap-build` namespace. Do not ingest the repository, PDFs, `.beads`, or private reports. Upstream describes a local stdio server, embedded database, and local fastembed fallback, but first use still needs its model assets. See the [pinned v1.9.0 README](https://github.com/Cognilogical/NeuroStrata/blob/v1.9.0/README.md).
2. **NeuroPlasticity:** create a reviewed `plasticity.json` that points to a fixed agent command, uses read-only project and isolated scratch mounts, and writes rules/logs only in a disposable copy. Verify a rootless Podman runtime, image availability, and model availability before running. The v2.1.0 release starts the agent in a container and may download its embedded Qwen model; failed runs can capture stdout/stderr and invoke the optimizer. Never mount existing Codex credentials. See the [pinned v2.1.0 README](https://github.com/Cognilogical/NeuroPlasticity/blob/v2.1.0/README.md).
3. Mark each tool ready only after its functional smoke test passes in a fresh task. `command -v`, `--version`, or a successful binary install alone does not pass this gate. The bootstrap prints this distinction every run.

## Operating boundaries

- This bootstrap installs executables only. It does not fetch NeuroStrata embedding weights, NeuroPlasticity model weights, or any other model/runtime data; configure hosts, credentials, MCP servers, hooks, daemons, databases, or network access; or run either Neuro tool.
- Do not treat installation as model/runtime readiness. Existing host and model gates remain those recorded in [build-machine tooling](build-machine-tooling.md).
- veraPDF remains CI/CD-only. This bootstrap reuses the existing authenticated provisioner in cache-only mode to retain the signed installer ZIP at `artifacts/verapdf/1.30.3/` and exports its path and pinned digests from `activate.sh`. The installer is not executed, Java is not installed, and no CLI runtime or PDF validation is run or added to `PATH`. The dedicated CI workflow uses the same provisioner in its existing install mode, checks the installed CLI JAR and version, and runs explicit PDF/UA-2 controls. This artifact cache does not extend the upstream GPL/MPL license election or add a production engine/runtime dependency.
- Tool binaries stay outside the checkout and product runtime. The activation file only prepends the isolated Go and tool directories when sourced, and avoids adding the same directory twice. Nothing writes a global profile or changes system PATH.
- If module source or checksum-database access fails for a tool, that tool remains uninstalled; the bootstrap reports the failure, continues with independent assets, and exits nonzero. It never turns off sum verification or falls back to an unpinned binary.

## Tests and diagnostics

The standalone bootstrap module's tests cover strict manifest parsing, checksummed download rejection and cleanup, version mismatch rejection, repeated-run idempotency, safe archive extraction, concurrent-lock timeout behavior, veraPDF artifact integrity separate from runtime/validation status, tamper detection and repair, cache-only provisioner invocation, and parity between the embedded and repository manifests. The real authenticated CLI and oracle-control tests remain in the CI-only PDF/UA-2 workflow. Useful checks are:

```bash
(cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./...)
bash -n scripts/bootstrap-cloud-build-tools.sh
```

The command prints per-tool `verified`, `installed`, or `cached artifact` status. It records tool-specific acquisition errors and continues with other tools, then exits nonzero if any installation failed. For diagnosis, check the selected root/cache, ensure required `curl`, `sha256sum`, `tar`, `timeout`, and `gpg` utilities exist, verify disk space and outbound access to the exact HTTPS endpoints in the manifest, and inspect a lock path before removing it. Module builds use `GOTOOLCHAIN=local`, `GOSUMDB=sum.golang.org`, isolated Go caches, and only `https://proxy.golang.org`; there is no direct-VCS fallback. The earlier proxy route redirected to `storage.googleapis.com` and returned HTTP 403 in this workspace; direct VCS access is disabled by the pinned proxy-only policy. Errors suppress subprocess output because proxy diagnostics may contain signed download URLs.
