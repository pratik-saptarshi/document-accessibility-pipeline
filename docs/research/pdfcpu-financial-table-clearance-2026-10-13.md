# Financial-table repair candidate: pdfcpu v0.16.2

**Status: proposed evidence; `dap-jo9.1.3` remains in progress.** This record narrows evidence to the exact Go API candidate proposed for the finite synthetic statement/invoice table experiment. It does not clear the candidate, establish repair capability or fidelity, approve a runtime dependency, or select a production writer. Independent review and exact-branch CI are pending.

## Candidate and operation boundary

- Candidate: `github.com/pdfcpu/pdfcpu` v0.16.2, source tag commit [`b41a41d5323f6119ff6a183ad3f287eb4a20a617`](https://github.com/pdfcpu/pdfcpu/tree/v0.16.2). The module requires Go 1.26.0 or later and the source is Apache-2.0. The repository pins v0.16.2 for characterization tests only.
- Proposed finite operation: use the Go API to read and validate one protocol-positive synthetic PDF 2.0 fixture, make one fixture-authored `/Headers` association change in an isolated in-memory context, and serialize a separate candidate. The equal-priority statement and invoice inputs, expected associations, strings, topology, and controls are in the [financial-table protocol](../fixtures/remediation/financial-table-experiment.v1.md) and its [A fixture package](../fixtures/remediation/financial-table-experiment.v1.json).
- The API surface is present: `api.LoadConfiguration`, `api.ReadContext`, `api.ValidateContext`, and `api.WriteContext`; generic PDF dictionaries are represented by `types.Dict`. A compile-only probe type-checked these API signatures and stateless configuration. **It did not open a PDF, mutate a PDF, run a repair, or establish fidelity.** The earlier preservation tests are separate characterization evidence, not evidence that this table operation works.
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

## Embedded and generated resources

`go list -f '{{.EmbedFiles}}' github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model` reports, with no build tags, `resources/Roboto-Regular.ttf` and `resources/config.yml`. Their SHA-256 hashes in the pinned module cache are:

| Resource | SHA-256 | Notice/provenance evidence |
|---|---|---|
| `Roboto-Regular.ttf` | `017c0be9aaa6d0359737e1fa762ad304c0e0107927faff5a6c1f415c7f5244ed` | Git blob ID `2b6392ffe8712b9c5450733320cd220d6c0f4bce`; `fc-scan` 2.15.0 identifies family Roboto, PostScript name `Roboto-Regular`, font version `140050` (1.4.5), but reports no copyright/license field. Prior research mapped this exact blob to Roboto v2.137 / Apache-2.0, but that archived upstream notice could not be independently retrieved in this pass: current Google Fonts paths have moved and the GitHub API returned 403. Treat the version/license linkage as **unresolved evidence**, despite the matching blob ID. |
| `config.yml` | `a37e388e7241b0243ecb0f9dabe739fbdddb4e6ff7e65e860ba41e7d78724868` | Embedded pdfcpu default configuration; covered by the pinned pdfcpu source license. |

The probe binary's `go tool nm` output contained `model.configFileBytes` and `model.robotoFontFileBytes`, confirming those resources are linked into the default API binary. It did not contain the tag-only certificate embedding symbols. `go list -tags=pdfcpu_eutl` adds the certificate `.p7c` files; this tag is excluded from the proposed build and requires a separate asset/notice review if enabled.

The ordinary package graph includes `internal/corefont/metrics` generated from Adobe Core 14 AFM measurements. The upstream tree includes `internal/corefont/Core14_AFMs/MustRead.html`: its permission allows use, copying, and distribution for any purpose without charge, with or without modification, subject to retaining copyright notices, not distributing AFM files without the accompanying notice, and prominently marking modifications. The Go build uses generated metric tables; the raw `.afm` files are not listed as embedded resources. Preserve the Adobe notice with any redistributed dependency evidence and do not strip it from the source package. This is a file-specific permissive grant with notice conditions, not an SPDX label.

The probe was built as `/tmp/dap-pdfcpu-api-probe` with the target above and also with `CGO_ENABLED=0`; `go version -m` confirmed pdfcpu v0.16.2 and the nine external modules listed above. The probe was not run. `pdfcpu_eutl` was not used. These are compile and linkage observations only.

## Configuration and bounded execution

Source inspection of pdfcpu v0.16.2 shows `api.LoadConfiguration(api.ConfigurationOptions{Mode: api.ConfigurationModeStateless})` returns built-in defaults without consulting filesystem configuration. However, the stateless defaults set `Offline=false`; they also leave `Limits.MaxInputBytes=0`, which means unlimited input. The proposed operation-specific settings, pending independent review, are:

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

1. The compile-only probe confirms API availability only. No table association was edited or repaired. The first repair attempt must remain blocked until this document/evidence increment receives independent review and exact-branch CI passes.
2. The protocol's source review predicts that PDF 2.0 full-write may add Helvetica `FirstChar`, `LastChar`, `Widths`, and a `FontDescriptor`; earlier full-write tests also observed Producer and date metadata changes. No output delta has been measured on the two financial fixtures. The repair spike must compare the exact positive fixture and record all structural, text, and rendered changes; a successful write is not a fidelity result.
3. `go-licenses` reports source license IDs and module paths but not every linked resource or a binary SBOM. The Roboto archived notice/version mapping remains unresolved. No local vulnerability-database fetch or local vulnerability scan was run; use the authoritative GitHub CI security result for this exact revision.
4. The table-specific semantic predicate and whole-document limits remain those in the protocol. A pdfcpu parse/validation result and the existing veraPDF structure-type control cannot establish that the target amount cells point to the intended header. Human semantic and assistive-technology review remain separate gates.

The bounded next action is independent review of this evidence and the A package, followed by exact-branch CI. Keep `dap-jo9.1.3` in progress until those gates accept the evidence and the Roboto notice question is resolved or the candidate is explicitly dispositioned as no-go/inconclusive. Do not run the repair experiment before then.

## Local verification on this increment

The first root race-test attempt used the environment's read-only default home and failed when existing pdfcpu fixture tests called `ReadContextFile` and tried to create `/home/agent/.config/pdfcpu`. The checks were rerun with a writable temporary `HOME=/tmp/dap-clearance-home`, cached dependencies, and `GOPROXY=off`; this passed and confirms why the proposed candidate operation must explicitly use stateless configuration.

| Scope | Command | Result |
|---|---|---|
| Root | `test -z "$(gofmt -l .)"`; `go mod tidy -diff`; `go vet ./...`; `go test -race ./...`; `go run ./cmd/traceability-check` | Passed. Traceability: 17 requirements, 6 stories, 21 tasks, 26 specification-only cases. |
| `cmd/cloud-tools-bootstrap` | `go test -race ./...`; `go vet ./...`; `go mod tidy -diff`; `bash -n ../../scripts/bootstrap-cloud-build-tools.sh` | Passed. |
| `ci/pdfua2/adapter` | `PDFUA_REAL_CLI_JAR= go test -race ./... -count=1`; `go vet ./...`; `go mod tidy -diff` | Passed. No Java/verapdf invocation. |
| Source publisher | `bash -n scripts/source-publisher.sh .github/scripts/test-source-publisher.sh`; `bash .github/scripts/test-source-publisher.sh` | Passed. Publisher tests used their disposable local repositories. |
| API probe | `go build -o /tmp/dap-pdfcpu-api-probe /tmp/dap-pdfcpu-api-probe.go`; repeat with `CGO_ENABLED=0`; `go version -m /tmp/dap-pdfcpu-api-probe` | Both compile-only builds passed; module/build metadata matched the table above. Probe was not executed. |
| License inventory | `go-licenses report --include_tests ./internal/pdfpreserve ./internal/financialfixtures` | Exit 0; reachable license IDs recorded above. Source-URL lookup warnings for cached YAML and `golang.org/x` modules were resolved by inspecting their local pinned `LICENSE`/`NOTICE` files. |

No local vulnerability-database fetch or vulnerability scan was run. The user reports that the merged PR's required and post-merge checks passed; exact-branch CI for this evidence increment is still pending.
