# Validation strategy

Validation is layered and evidence-based. No automated tagging or single machine report establishes semantic correctness or PDF/UA/WCAG conformance.

## Planned test layers

- **Characterization and preservation:** synthetic/deidentified input/output comparisons for visible text, page geometry/rendering, page count, links, forms, annotations, fonts, metadata, and structure where applicable.
- **Unit:** profile/data validation, semantic repair decisions, template behavior, error handling, and audit evidence.
- **Adapter contract:** inputs, outputs, deterministic evidence, unsupported capabilities, cancellation/timeouts, retries, and license/capability policy.
- **Integration:** real open-source candidate workers and writers against synthetic fixtures after versions/licenses are verified.
- **End to end:** complete generation and remediation flows for statements, invoices, correspondence, and reports.
- **Failure injection and security:** malformed PDFs, hostile metadata, oversized/resource-intensive input, worker crash/timeout, retry/idempotence, privacy leakage, and audit integrity.
- **Independent machine validation:** PDF/UA-2 is the accepted output target. A dedicated CI job authenticates veraPDF 1.30.3 and runs `--flavour ua2` in an ephemeral non-root container with no network, read-only input/root, and resource bounds. A separate Go-only job verifies the versioned report bundle. Installer signature, hashes, image provenance, and real pass/finding evidence are recorded in [the acquisition evidence](../research/verapdf-cli-1.30.3.md). Fake-process and synthetic XML tests verify contract handling, not semantic accessibility or PDF conformance.
- **Subprocess boundary:** the legacy Go adapter and its Linux process tests live in the CI-only nested module `ci/pdfua2/adapter`; they are not part of the root product Go module. A regression test covering the root module's `internal/pdfua` report package rejects Java or process execution markers in that package. Adapter tests reject symlinks and non-regular input/artifact files, bound preparation and process time, terminate the isolated process group after command completion (including cancellation), and cap pipe draining. The CI validator container separately enforces `--network=none`, a read-only root/input, non-root identity, dropped capabilities, no-new-privileges, memory/CPU/PID/tmp/output limits, and timeout. Synthetic reports require every batch-summary counter used for classification and reject non-whitespace outside the single report root. Real bundled reports are rechecked against the pinned profile, hashes, oracle input, and validator exit code in the Go-only job.
- **Human gates:** semantic correctness review and assistive-technology checks by people; reviewer authority, protocol, and sampling policy remain open. PDF/UA-2 is the accepted output-validation target.

For each vertical slice, tests follow red/green/refactor. Record actual test and validation results only after they are run; do not fabricate results, thresholds, or acceptance levels. Maintain synthetic fixture provenance and do not add confidential samples.

The [acceptance case catalog](cases.md) is a set of proposed Given/When/Then specifications, not executable product tests and not evidence that behavior passed. The [traceability matrix](../traceability/matrix.md) links each case to a story, function contract, task, and requirement.

The Go standard-library planning data checker is separate from product validation. Its structural tests run with `go test -race ./...`; those tests check catalog integrity and are not PDF accessibility tests. The CI-only adapter tests run separately with `(cd ci/pdfua2/adapter && go test -race ./... -count=1)`.
