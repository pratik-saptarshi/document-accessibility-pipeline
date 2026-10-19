# Cloud build-tool bootstrap

**Status: proposed and reviewable.** This setup is for an isolated cloud build machine. It does not change host configuration, the product architecture, or shipped runtime images.

## Install and verify

The reviewed entrypoint is [`scripts/bootstrap-cloud-build-tools.sh`](../../scripts/bootstrap-cloud-build-tools.sh). It supports Linux x86_64 only. The first stage installs Go 1.27.2 from go.dev after checking the fixed SHA-256. Without a checkout, it fetches the tracked source archive from a raw HTTPS URL pinned to an immutable storage commit and exact repository path. The manifest separately records the source revision contained in that archive. The shell verifies exact byte size and SHA-256 before extraction or building the standard-library-only command; it has no generated-archive fallback. That command reads its embedded manifest, downloads only pinned official release assets, verifies their digests, and installs into workspace-local directories. Go-installed tools are fetched from their pinned source modules with the public Go checksum database enabled.

The archive contains source revision `ef4c35629c467887b6749f9e31af111cbe8f8913` — exactly the dap-5hy.2-accepted PR #51 bounded bootstrap source (not the current main tip; #109's manifest hardening is outside dap-5hy.2 acceptance and is not part of this pin). The active no-checkout pin now uses the filtered 388,215-byte artifact at `buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz`, SHA-256 `f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d`, stored in protected-main commit `4ce49c0b4face9388f036e08f3e05af7a3bf58a2` (PR #153). The original 469,726-byte archive at its historical filename, SHA-256 `16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473` (protected-main commit `05c9a4a710eef0bb6b525e7f853da7e874430a00`), remains in the tree for historical evidence. Its membership filtering is reviewable in `internal/devsecops/source_archive_denylist.go`; it removes `.beads/` and members matching the documented personal-contact patterns. Both artifacts derive from the official GitHub archive; neither is reconstructed from a local Git archive. Generated archive URLs can change compression bytes for the same source revision. Missing, wrong-size, or wrong-hash bytes fail before extraction or Go build.

The repository bootstrap discards output from `go install` and captures at most 64 KiB total stdout and stderr when checking a tool's version or fallback Go build metadata. Output beyond that limit fails verification closed; diagnostics report only that evidence exceeded the limit and never include child output. The no-checkout pin was replaced with the filtered artifact in a synchronized update of both manifests, the bootstrap shell entrypoint, and the pinned-source tests, using the protected-main storage commit `4ce49c0b4face9388f036e08f3e05af7a3bf58a2` as the raw-URL anchor. New no-checkout execution was verified 2026-10-19T00:20:09Z in an isolated script-only run (no repo checkout): the filtered archive downloaded from the pinned raw URL, size and SHA-256 verified before extraction, `cloud-tools-bootstrap` built from the archived source, and install plus `--verify` completed with exit 0 (see the dap-5hy.6 closure note for the evidence summary). Only exact-head hosted CI on the landing PR remains pending. Saved environment configuration remains unchanged.

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

This read-only check executes each pinned version command or compares the configured artifact/executable digest. It reports `NOT VERIFIED` for missing, wrong-version, or wrong-digest binaries. It checks identity at run time; it is not a tamper-proof attestation of future file integrity. The separate veraPDF CI job is outside this bootstrap.

Activation uses the exact validated `--go` executable. When its basename is not `go`, bootstrap atomically writes a regular launcher at `<tools-root>/.activation/bin/go`; the launcher executes that absolute path and forwards arguments and exit status. The activation prepends the launcher directory, so `command -v go` reports the launcher for a renamed selection. This binds command invocation to the selected executable but does not provide tamper-proof future integrity; rerun `--verify` to recheck the selected tool.

For command discovery only, use this loop after sourcing `activate.sh`:

```bash
for tool in go bd actionlint govulncheck go-licenses gitleaks neurostrata-mcp neuroplasticity; do
  command -v "$tool" || printf 'MISSING: %s\n' "$tool"
done
```

For an environment with no repository checkout, the saved Install script builds from the checksum-verified immutable source archive in the manifest; it does not depend on Go module version discovery. The publisher should place the same entrypoint in each saved environment's additive Install setup and put the activation plus command-discovery checks in its Start skill. Repo refresh alone does not rerun Install, and the Start instructions must verify command resolution for each fresh task. No saved environment configuration is modified by this repository change.

### Updating a previously saved environment

After this change is merged, update the saved Install entrypoint to the current `scripts/bootstrap-cloud-build-tools.sh` and rerun it in a fresh task shell. The no-checkout bootstrap archive is pinned to the decoupled source revision in the manifest; it no longer has a veraPDF provisioner option. Rerunning bootstrap refreshes the generated activation script, which has no `VERAPDF_*` variables.

The earlier setup may have a separately verified installer ZIP at `<tools-root>/artifacts/verapdf/1.30.3/verapdf-greenfield-1.30.3-installer.zip`. The new CI flow does not read it. Leave the file in place by default. If an operator later chooses to reclaim that one file, first verify its exact path and SHA-256 against the recorded official artifact digest, confirm the new bootstrap and CI flow are working, and remove only that exact file; do not recursively remove an artifacts, cache, or tools directory. Preserve the Go toolchain, module/build caches, other installed binaries, `.beads/embeddeddolt`, Beads exports, and all other persistent data. Rerunning bootstrap replaces activation with the current version; shells already open may retain old exported variables, so inspect and unset only the obsolete `VERAPDF_*` variables in that shell if needed. This repository change does not access or alter a saved environment.

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
- The shared cloud bootstrap contains exactly eight build tools. It does not acquire or expose a veraPDF ZIP, Java runtime, executable, or environment variable. The historical workspace cache remains outside this repository change and is not removed or modified. CI independently authenticates the official installer, builds a temporary digest-pinned Temurin/veraPDF image, runs the real controls in an offline non-root container, and passes only bounded XML reports plus versioned JSON metadata to a separate Go-only job. The image and JAR are neither published nor uploaded.
- Tool binaries stay outside the checkout and product runtime. The activation file only prepends the isolated Go and tool directories when sourced, and avoids adding the same directory twice. Nothing writes a global profile or changes system PATH.
- If module source or checksum-database access fails for a tool, that tool remains uninstalled; the bootstrap reports the failure, continues with independent assets, and exits nonzero. It never turns off sum verification or falls back to an unpinned binary.

## Tests and diagnostics

The standalone bootstrap module's tests cover strict manifest parsing, checksummed download rejection and cleanup, version mismatch rejection, repeated-run idempotency, safe archive extraction, concurrent-lock timeout behavior, eight-tool-only manifest/activation boundaries, and parity between the embedded and repository manifests. Synthetic Docker-shim tests exercise container flags, control hash rejection, output bounds, operational exits, and forced cleanup; the CI-only PDF/UA-2 workflow runs the real validator controls and the Go bundle verifier. Useful checks are:

```bash
(cd cmd/cloud-tools-bootstrap && go test -race ./... && go vet ./...)
bash -n scripts/bootstrap-cloud-build-tools.sh
```

The command prints per-tool `verified` or `installed` status. It records tool-specific acquisition errors and continues with other tools, then exits nonzero if any installation failed. For diagnosis, check the selected root/cache, ensure required `curl`, `sha256sum`, `stat`, `tar`, and `timeout` utilities exist, verify disk space and outbound access to the exact HTTPS endpoints in the manifest, and inspect a lock path before removing it. Module builds use `GOTOOLCHAIN=local`, `GOSUMDB=sum.golang.org`, isolated Go caches, and only `https://proxy.golang.org`; there is no direct-VCS fallback. The exact `pdfcpu v0.16.2` proxy URL redirected to `storage.googleapis.com` and returned HTTP 200 in this environment. Errors suppress subprocess output because proxy diagnostics may contain signed download URLs.
