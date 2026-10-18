# Financial-table repair candidate: pdfcpu v0.16.2

**Status: cleared for the finite, test-only synthetic experiment described below.** The independent product/gate and candidate/resource reviews approved evidence head `25c4ca6a4160f76aae3c62c61b1e9cae4cc41f0e`; that evidence was merged in PR #44 at accepted main `8dd62bac02b05d42cd2001d195613cb89b995276`, whose nine pre-merge and seven post-merge checks passed. Clearance covers only pdfcpu v0.16.2's pinned Go API, reachable package/resource/build boundary and explicit configuration below, for attempted synthetic statement and invoice controls. It does not establish repair capability or fidelity, approve a runtime dependency, select a production writer, or authorize broader inputs or operations.

## Candidate and operation boundary

- Candidate: `github.com/pdfcpu/pdfcpu` v0.16.2, source tag commit [`b41a41d5323f6119ff6a183ad3f287eb4a20a617`](https://github.com/pdfcpu/pdfcpu/tree/v0.16.2). The module requires Go 1.26.0 or later and the source is Apache-2.0. The repository pins v0.16.2 for characterization tests only.
- Cleared finite operation: first attempt the protocol-positive synthetic PDF 2.0 fixture as a separate **positive no-op full-write control** for both families. Only if that control preserves the protected structures may the experiment continue by reading each immutable **damaged** input and repairing only its declared `/Headers` association to the fixture-authored expected header ID, into separate private staging. The equal-priority statement and invoice inputs, expected associations, strings, topology, and controls are in the [financial-table protocol](../fixtures/remediation/financial-table-experiment.v1.md) and its [A fixture package](../fixtures/remediation/financial-table-experiment.v1.json). A positive-to-damaged mutation is not a repair, and a positive no-op alone cannot satisfy `dap-jo9.1.2`.
- The API surface is present: `api.LoadConfiguration`, `api.ReadContext`, `api.ValidateContext`, and `api.WriteContext`; generic PDF dictionaries are represented by `types.Dict`. A compile-only probe type-checked these API signatures and stateless configuration. **It did not open a PDF, mutate a PDF, run a repair, or establish fidelity.** The earlier preservation tests are separate characterization evidence, not evidence that this table operation works. Clearance authorizes the bounded experiment only; it is not production adoption.
- Do not use pdfcpu's CLI, `ReadContextFile` default configuration, incremental writer, or `WriteContextFile` as a publication decision. Keep Java and veraPDF isolated to the existing CI-only validator boundary. No other non-Go runtime exception is proposed.

The compile-only probe source was `/tmp/dap-pdfcpu-api-probe.go`:

```go
package main

import (
    "context"
    "io"

    "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
    conf, _ := api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})
    conf.Offline = true
    conf.Limits.MaxInputBytes = 64 << 10
    var read func(context.Context, io.ReadSeeker, *model.Configuration) (*model.Context, error) = api.ReadContext
    var validate func(context.Context, *model.Context) error = api.ValidateContext
    var write func(context.Context, *model.Context, io.Writer) error = api.WriteContext
    var root types.Dict
    _, _, _, _ = conf, read, validate, write
    _ = root
}
```

## Reproducible build target and reachable package inventory

Evidence was collected on 2026-10-13 from merged base `b462faa070a647176f9f0bf8264cd89cdd531bc1` using the checked-in modules and `/workspace/.codex-tools/cloud-build-tools/activate.sh`:

| Setting | Observed value |
|---|---|
| Go toolchain | `go1.27.2` |
| Target | `linux/amd64`, `GOAMD64=v1` |
| Default cgo setting | `CGO_ENABLED=1`; the compile-only probe also built with `CGO_ENABLED=0` |
| Build tags | none (`GOFLAGS` empty); `pdfcpu_eutl` is excluded |
| Dependency access for local evidence | `GOPROXY=off`, `GOSUMDB=off`; packages were already cached |

Command: `go list -deps -test -f '{{.ImportPath}} {{with .Module}}{{.Path}} {{.Version}}{{end}}' ./internal/pdfpreserve ./internal/financialfixtures`. It succeeded and returned **282 package records** across both test-inclusive project closures. The pdfcpu portion is these 28 package paths:

```text
github.com/pdfcpu/pdfcpu/internal/contextutil
github.com/pdfcpu/pdfcpu/internal/fileutil
github.com/pdfcpu/pdfcpu/internal/inputlimit
github.com/pdfcpu/pdfcpu/internal/filter/lzw
github.com/pdfcpu/pdfcpu/internal/corefont/metrics
github.com/pdfcpu/pdfcpu/internal/netutil
github.com/pdfcpu/pdfcpu/pkg/api
github.com/pdfcpu/pdfcpu/pkg/filter
github.com/pdfcpu/pdfcpu/pkg/font
github.com/pdfcpu/pdfcpu/pkg/log
github.com/pdfcpu/pdfcpu/pkg/pdfcpu
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/color
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/create
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/draw
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/fault
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/font
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/format
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/matrix
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/pkcs7
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/primitives
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/safemath
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/sanitize
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/scan
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/sign
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types
github.com/pdfcpu/pdfcpu/pkg/pdfcpu/validate
```

Command: `go-licenses report --include_tests ./internal/pdfpreserve ./internal/financialfixtures`. It exited 0 and reported the license IDs below. It warned that source URLs could not be discovered for `go.yaml.in/yaml/v3` and the `golang.org/x` modules with proxy access disabled; their license files were inspected in the pinned module cache. The project module itself is MIT and is omitted from this third-party list.

| Reachable module | Version | License evidence |
|---|---:|---|
| `github.com/pdfcpu/pdfcpu` | `v0.16.2` | Apache-2.0, upstream [`LICENSE.txt`](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/LICENSE.txt) |
| `github.com/clipperhouse/uax29/v2` | `v2.7.0` | MIT, `LICENSE` |
| `github.com/hhrutter/tiff` | `v1.0.7` | BSD-3-Clause, `LICENSE`; nested `internal/lzw/LICENSE` also BSD-3-Clause |
| `github.com/mattn/go-runewidth` | `v0.0.30` | MIT, `LICENSE` |
| `github.com/xdg-go/stringprep` | `v1.0.4` | Apache-2.0, `LICENSE` |
| `go.yaml.in/yaml/v3` | `v3.0.5` | MIT and Apache-2.0; module `LICENSE` and `NOTICE` identify the file split |
| `golang.org/x/crypto` | `v0.57.0` | BSD-3-Clause, module `LICENSE` |
| `golang.org/x/image` | `v0.46.0` | BSD-3-Clause, module `LICENSE` |
| `golang.org/x/text` | `v0.42.0` | BSD-3-Clause, module `LICENSE` |

The pdfcpu tree also has file-specific notices not represented by the module-level SPDX label: `pkg/pdfcpu/pkcs7/LICENSE` is MIT, and `internal/filter/lzw/LICENSE` is BSD-3-Clause. The ordinary Go package graph does not include pdfcpu CLI packages such as Cobra/pflag, and the probe build metadata contains the nine external modules above. This is an operation/test package inventory, not a complete packaged binary SBOM or approval of pdfcpu's full module graph.

## Pinned maintenance and security posture

The [`SECURITY.md` at the pinned source commit](https://github.com/pdfcpu/pdfcpu/blob/b41a41d5323f6119ff6a183ad3f287eb4a20a617/SECURITY.md) covers the Go library, CLI, official release binaries, and their included dependencies. It states that the project has one maintainer and no absence coverage, provides no maintenance releases or backports, and promises no fixed security-support period; security fixes are delivered in later releases. These are maintenance risks for a product dependency, not evidence that a vulnerability remains unpatched.

The [v0.16.2 release notes](https://github.com/pdfcpu/pdfcpu/releases/tag/v0.16.2) report fixes for plaintext strings in encrypted object-stream PDFs and process termination from recursive stream references, in addition to other parser, fidelity, and resource-limit changes. Those advisories affect earlier versions through v0.16.1. This evidence describes the fixes included in v0.16.2; it does not prove that no other vulnerabilities exist or replace the repository's authoritative GitHub CI security scan. No local vulnerability database was fetched or scanned.

## Embedded and generated resources

`go list -f '{{.EmbedFiles}}' github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model` reports, with no build tags, `resources/Roboto-Regular.ttf` and `resources/config.yml`. Their SHA-256 hashes in the pinned module cache are:

| Resource | SHA-256 | Notice/provenance evidence |
|---|---|---|
| `Roboto-Regular.ttf` | `017c0be9aaa6d0359737e1fa762ad304c0e0107927faff5a6c1f415c7f5244ed` | Git blob ID `2b6392ffe8712b9c5450733320cd220d6c0f4bce`. In the SFNT `name` table (offset 140424, length 658, format 0, count 13, string-storage-relative-offset 162), Windows Unicode platform 3 / encoding 1 / language `0x0409` records give: name ID 0, record offset 140430, string offset 140586, length 94: `Copyright 2011 Google Inc. All Rights Reserved.`; ID 5, record 140490, string 140706, length 38: `Version 2.137; 2017`; ID 13, record 140562, string 140906, length 92: `Licensed under the Apache License, Version 2.0`; ID 14, record 140574, string 140998, length 84: `http://www.apache.org/licenses/LICENSE-2.0`. Strings are UTF-16BE. The `head` table begins at 128216; its `fontRevision` uint32 at 128220 is `140050`. `fc-scan` 2.15.0 reports the same value; interpret it as 16.16 fixed-point: `140050/65536 = 2.136993408203125`, consistent with 2.137. It must not be read as 1.4.5. |
| `config.yml` | `a37e388e7241b0243ecb0f9dabe739fbdddb4e6ff7e65e860ba41e7d78724868` | Embedded pdfcpu default configuration; covered by the pinned pdfcpu source license. |

The pinned upstream [`pkg/testdata/fonts/LICENSE.txt`](https://github.com/pdfcpu/pdfcpu/blob/b41a41d5323f6119ff6a183ad3f287eb4a20a617/pkg/testdata/fonts/LICENSE.txt) lists Roboto with an Apache 2.0 license. Its sibling [`Roboto-Regular.ttf`](https://github.com/pdfcpu/pdfcpu/blob/b41a41d5323f6119ff6a183ad3f287eb4a20a617/pkg/testdata/fonts/Roboto-Regular.ttf) is byte-identical to the embedded resource: same 171272-byte length, SHA-256, and Git blob ID. That license file also names GPL-licensed GNU Unifont test fonts; those files are under testdata and are not reachable or embedded in this candidate's ordinary API build.

Reproduction used Go's standard library against the exact cached TTF path, without executing the font or accessing the network. The reader verified the 171272-byte input SHA-256 above; walked 16-byte SFNT directory entries beginning at byte 12 to find `name` and `head`; decoded the `name` table's 12-byte records; selected platform 3 / encoding 1 / language `0x0409`; computed each string position as `nameTableOffset + storageRelativeOffset + recordStringOffset`; and decoded UTF-16BE. It read `fontRevision` as a big-endian uint32 at `headTableOffset+4` and divided by 65536. The resulting table/record offsets and values are shown above. `git hash-object` on the same file returned `2b6392ffe8712b9c5450733320cd220d6c0f4bce`.

The extraction uses only byte offsets, big-endian integer reads, and UTF-16 decoding; the key calculations are:

```go
nameTableOffset := tableDirectoryOffset
storageOffset := nameTableOffset + int(binary.BigEndian.Uint16(font[nameTableOffset+4:]))
recordOffset := nameTableOffset + 6 + recordIndex*12
stringOffset := storageOffset + int(binary.BigEndian.Uint16(font[recordOffset+10:]))
fontRevision := binary.BigEndian.Uint32(font[headTableOffset+4:])
revision := float64(fontRevision) / 65536
```

The probe binary's `go tool nm` output contained `model.configFileBytes` and `model.robotoFontFileBytes`, confirming those resources are linked into the default API binary. It did not contain the tag-only certificate embedding symbols. `go list -tags=pdfcpu_eutl` adds the certificate `.p7c` files; this tag is excluded from the proposed build and requires a separate asset/notice review if enabled.

The ordinary package graph includes `internal/corefont/metrics` generated from Adobe Core 14 AFM measurements. The upstream tree includes `internal/corefont/Core14_AFMs/MustRead.html`: its permission allows use, copying, and distribution for any purpose without charge, with or without modification, subject to retaining copyright notices, not distributing AFM files without the accompanying notice, and prominently marking modifications. The Go build uses generated metric tables; the raw `.afm` files are not listed as embedded resources. Preserve the Adobe notice with any redistributed dependency evidence and do not strip it from the source package. This is a file-specific permissive grant with notice conditions, not an SPDX label.

The probe was built as `/tmp/dap-pdfcpu-api-probe` with the target above and also with `CGO_ENABLED=0`; `go version -m` confirmed pdfcpu v0.16.2 and the nine external modules listed above. The probe was not run. `pdfcpu_eutl` was not used. These are compile and linkage observations only.

## Configuration and bounded execution

Source inspection of pdfcpu v0.16.2 shows `api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})` returns built-in defaults without consulting filesystem configuration. However, the stateless defaults set `Offline=false`; they also leave `Limits.MaxInputBytes=0`, which means unlimited input. The independently reviewed, experiment-only operation settings are:

```go
conf.Offline = true
conf.Limits = model.ResourceLimits{
    MaxInputBytes: 64 << 10, MaxObjectBytes: 32 << 10,
    MaxStreamBytes: 32 << 10, MaxDecodeBytes: 64 << 10,
    MaxImagePixels: 1_000_000, MaxImageBytes: 4 << 20,
    MaxObjectCount: 256, MaxObjectStreamCount: 64,
    MaxObjectStreamFirst: 16 << 10, MaxXRefEntries: 256,
    MaxRecursionDepth: 32,
}
```

Also use a cancellable context with a 30-second experiment deadline and a separate candidate writer capped at 256 KiB. Inputs are only a few kilobytes and the declared table topology has no images; these are experiment limits to test against the exact synthetic fixtures, not product limits. If a fixture or writer needs more, stop and review the boundary rather than raising limits silently. The default object/stream/decode/image ceilings are much larger than these fixtures and do not by themselves make the experiment bounded.

Use a cancellable context and `api.ReadContext` over an immutable input, then `api.ValidateContext`. Direct the full writer to a separately staged, bounded output; a writer error can leave partial bytes. Freshly re-read and validate the candidate before comparing it. `WriteContextFile`'s staging/commit success does not establish comparison, and no candidate is published until all required evidence and disposition are accepted. Do not let a mutable pdfcpu context replace the immutable input or baseline.

## Capability and fidelity gaps that keep the gate open

1. Source inspection shows that `types.Dict` is a mutable `map[string]Object`, with an `Update` method for replacing existing entries ([pinned `dict.go`](https://github.com/pdfcpu/pdfcpu/blob/b41a41d5323f6119ff6a183ad3f287eb4a20a617/pkg/pdfcpu/types/dict.go#L28-L92)). The ordinary full writer dereferences and serializes the catalog root and writes its `/StructTreeRoot` reference; the reduced-feature path deletes that entry ([pinned `write.go`](https://github.com/pdfcpu/pdfcpu/blob/b41a41d5323f6119ff6a183ad3f287eb4a20a617/pkg/pdfcpu/write.go#L398-L463)). This is source-level editing/serialization capability evidence, alongside the compile-only API probe; it does not prove that the exact `/Headers` repair works or preserves fidelity. No table association was edited or repaired. The independent reviews and PR44 CI clear only the bounded experiment to begin with the positive no-op control; any actual preservation delta still governs whether damaged-input repair is attempted.
2. The protocol's source review predicted that PDF 2.0 full-write may add Helvetica `FirstChar`, `LastChar`, `Widths`, and a `FontDescriptor`; the positive no-op control has now observed all four additions on both financial fixtures (see the [bounded no-go evidence](pdfcpu-financial-table-noop-2026-10-13.md)). Earlier full-write tests also observed Producer/date metadata changes, but this experiment's Info metadata remained empty. The experiment stopped at that preservation gate and did not repair either damaged input. A positive-to-damaged mutation or no-op alone is not a repair result. The observed Helvetica expansion is a preservation stop unless a reviewed protocol decision explicitly accepts that exact expansion; never waive it by weakening the baseline.
3. `go-licenses` reports source license IDs and module paths but not every linked resource or a binary SBOM. The embedded Roboto TTF itself contains the copyright, version, Apache-2.0 statement, and canonical license URL recorded above; this byte-level finding resolved the earlier asset provenance gap and was included in the exact candidate/resource review. No local vulnerability-database fetch or local vulnerability scan was run; use the authoritative GitHub CI security result for this exact revision.
4. The table-specific semantic predicate and whole-document limits remain those in the protocol. A pdfcpu parse/validation result and the existing veraPDF structure-type control cannot establish that the target amount cells point to the intended header. Human semantic and assistive-technology review remain separate gates.

The independent product/gate and candidate/resource reviews accepted the exact evidence head `25c4ca6a4160f76aae3c62c61b1e9cae4cc41f0e`. PR #44 merged that evidence into accepted main `8dd62bac02b05d42cd2001d195613cb89b995276`; its nine pre-merge and seven post-merge checks passed. This closes `dap-jo9.1.3` only for use of pinned pdfcpu v0.16.2's reviewed Go API/build/dependency/resource boundary in the finite synthetic attempt. It does not clear any other version, package/build tag, asset set, input, operation, or product/runtime adoption. The next bounded action is the positive no-op full-write preservation control on both positive fixtures. A protected-structure delta requires recording the measured delta and a no-go/inconclusive stop before any damaged-input repair; no no-op or positive-to-damaged output is a repair result.

## Local verification on this increment

The first root race-test attempt used the environment's read-only default home and failed when existing pdfcpu fixture tests called `ReadContextFile` and tried to create the default user configuration directory. The checks were rerun with a writable temporary `HOME=/tmp/dap-clearance-home`, cached dependencies, and `GOPROXY=off`; this passed and confirms why the proposed candidate operation must explicitly use stateless configuration.

| Scope | Command | Result |
|---|---|---|
| Root | `test -z "$(gofmt -l .)"`; `go mod tidy -diff`; `go vet ./...`; `go test -race ./...`; `go run ./cmd/traceability-check` | Passed. Traceability: 17 requirements, 6 stories, 21 tasks, 26 specification-only cases. |
| `cmd/cloud-tools-bootstrap` | `go test -race ./...`; `go vet ./...`; `go mod tidy -diff`; `bash -n ../../scripts/bootstrap-cloud-build-tools.sh` | Passed. |
| `ci/pdfua2/adapter` | `PDFUA_REAL_CLI_JAR= go test -race ./... -count=1`; `go vet ./...`; `go mod tidy -diff` | Passed. No Java/verapdf invocation. |
| Source publisher | `bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh`; `bash .github/scripts/test-source-publisher.sh` | Passed. Publisher tests used their disposable local repositories. |
| API probe | `go build -o /tmp/dap-pdfcpu-api-probe /tmp/dap-pdfcpu-api-probe.go`; repeat with `CGO_ENABLED=0`; `go version -m /tmp/dap-pdfcpu-api-probe` | Both compile-only builds passed; module/build metadata matched the table above. Probe was not executed. |
| Embedded font metadata | Read-only Go `encoding/binary` SFNT reader over the exact cached TTF path; verify SHA-256 and Git blob ID | Passed. Name records and fixed-point `head.fontRevision` match the offsets and values above; no font execution or network access. |
| License inventory | `go-licenses report --include_tests ./internal/pdfpreserve ./internal/financialfixtures` | Exit 0; reachable license IDs recorded above. Source-URL lookup warnings for cached YAML and `golang.org/x` modules were resolved by inspecting their local pinned `LICENSE`/`NOTICE` files. |

No local vulnerability-database fetch or vulnerability scan was run. PR44's nine pre-merge and seven post-merge checks passed on the accepted evidence head as reported in the review record. This later experiment increment still requires its own exact-revision CI; only GitHub CI is authoritative for security scans.
