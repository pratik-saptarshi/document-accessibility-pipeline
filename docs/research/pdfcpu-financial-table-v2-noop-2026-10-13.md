# Financial-table v2 positive no-op control

**Result: both positive controls passed the declared preservation checks.** This is only a positive full-write characterization. It does not establish repair, a publishable candidate, PDF/UA-2 conformance, production writer selection, or runtime adoption. The v1 font-metric no-go remains unchanged.

## Exact scope and execution

The controls used the reviewed v2 input and operation boundary in the [clearance extension](pdfcpu-financial-table-v2-clearance-2026-10-13.md): pdfcpu v0.16.2, Go API ReadContext, ValidateContext, and WriteContext, a stateless configuration, offline operation, 30-second context, 64 KiB input cap, 256 KiB output cap, and private mode-0600 temporary staging. Input bytes were SHA-256 checked and not modified. Output bytes were freshly read and validated.

The exact targeted invocation was DAP_RUN_V2_NOOP=1 go test ./internal/financialfixtures -run '^TestFinancialTableV2PositiveNoopControl$' -count=1 -v, with GOPROXY=off, GOSUMDB=off, and PDFUA_REAL_CLI_JAR=. It passed. The package test is opt-in; the normal test suite does not invoke a PDF writer. Poppler and Ghostscript commands ran with 30-second contexts, 250 ms wait delays, 256 KiB captured-text cap, 16 KiB diagnostic caps, and a 4 MiB rendered-image cap. These direct-command bounds do not claim descendant cleanup, hard RAM/CPU isolation, or adversarial executable containment; fixed-tool descendant cleanup remains deferred under dap-jo9.5.3.

## Results

| Family | Input SHA-256 / bytes | Private no-op output SHA-256 / bytes | Extracted text SHA-256 | Rendered PNG SHA-256 |
|---|---|---|---|---|
| Statement | 4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec / 4,860 | 2b5fa11f688b5c4ad161da69f003c57862ae7e63415402fa20be87aa28e1b344 / 2,507 | b20dd862dda3e51f2d65927f03a5df7ae01338acc76dda6f155e3f1a0b36b34b | 7c9531c6b51d04423f5d46ce28c65584e35a00983636d49b3dad895275d1785e |
| Invoice | ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72 / 4,806 | afd20e3673394449d37a1d3c921030c2b0343bf785aa17a4bfc0386184194f6b / 2,464 | 35fa4c4506c765eac234110017cc0592e0704bb3b51ae34589aac36d40485ea0 | d1ccd35abd66f7f33af61c4f8198f08ba36bc4a9444a04e7a29dbbc927fd2718 |

For each family, the fresh read preserved the table headers and rows, exact financial strings and reading order, every /Headers relation, /IDTree, and marked-cell content/ParentTree bindings. Complete page font resource names, references, dictionaries, and resolved indirect entries matched. PDF Info metadata matched. pdftotext -layout output and Ghostscript 10.05.1 72 dpi PNG bytes matched the retained independent positive baselines. The full-write PDF byte hashes differ from the immutable inputs because the writer rewrites and compresses the files; these temporary outputs were deleted with the test staging directory and are not candidates.

## Disposition

Both positive no-op preservation gates passed, so the bounded experiment may proceed to a separately reviewed damaged-input step. That step must open each immutable damaged fixture and change only its declared row-2 amount-cell /Headers relation to the expected positive association. It must preserve the exact input, all non-target structures and financial strings, write to separate private staging, and freshly verify outputs independently. No damaged input was opened in this no-op run; no repair, candidate, or validator report exists. Any candidate-bound PDF/UA-2 validation remains CI-only and must explicitly exercise new candidate bytes.
