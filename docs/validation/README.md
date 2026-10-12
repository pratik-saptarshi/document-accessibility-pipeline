# Validation strategy

Validation is layered and evidence-based. No automated tagging or single machine report establishes semantic correctness or PDF/UA/WCAG conformance.

## Planned test layers

- **Characterization and preservation:** synthetic/deidentified input/output comparisons for visible text, page geometry/rendering, page count, links, forms, annotations, fonts, metadata, and structure where applicable.
- **Unit:** profile/data validation, semantic repair decisions, template behavior, error handling, and audit evidence.
- **Adapter contract:** inputs, outputs, deterministic evidence, unsupported capabilities, cancellation/timeouts, retries, and license/capability policy.
- **Integration:** real open-source candidate workers and writers against synthetic fixtures after versions/licenses are verified.
- **End to end:** complete generation and remediation flows for statements, invoices, correspondence, and reports.
- **Failure injection and security:** malformed PDFs, hostile metadata, oversized/resource-intensive input, worker crash/timeout, retry/idempotence, privacy leakage, and audit integrity.
- **Independent machine validation:** PDF/UA-2 is the accepted output target. The CI-only veraPDF adapter explicitly uses `--flavour ua2`; its official installer signature, artifact hashes, and real pass/finding controls are recorded in [the acquisition evidence](../research/verapdf-cli-1.30.3.md). Fake-process and synthetic XML tests verify only the Go adapter contract, not validator behavior or PDF conformance.
- **Subprocess boundary:** the current adapter is supported only on Linux CI. It rejects symlinks and non-regular input/artifact files before reading, bounds preparation and process time, terminates the isolated validator process group after every command completion (including cancellation), and caps pipe draining. Other platforms fail closed. Synthetic XML tests require every batch-summary counter used for classification and reject non-whitespace outside the single report root; the dedicated integration test additionally checks the real 1.30.3 CLI report against genuine controls.
- **Human gates:** semantic correctness review and assistive-technology checks by people; reviewer authority, protocol, and sampling policy remain open. PDF/UA-2 is the accepted output-validation target.

For each vertical slice, tests follow red/green/refactor. Record actual test and validation results only after they are run; do not fabricate results, thresholds, or acceptance levels. Maintain synthetic fixture provenance and do not add confidential samples.

The [acceptance case catalog](cases.md) is a set of proposed Given/When/Then specifications, not executable product tests and not evidence that behavior passed. The [traceability matrix](../traceability/matrix.md) links each case to a story, function contract, task, and requirement.

The Go standard-library planning data checker is separate from product validation. Its structural tests run with `go test -race ./...`; those tests check catalog integrity and are not PDF accessibility tests.
