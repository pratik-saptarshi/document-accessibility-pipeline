# Validation strategy

Validation is layered and evidence-based. No automated tagging or single machine report establishes semantic correctness or PDF/UA/WCAG conformance.

## Planned test layers

- **Characterization and preservation:** synthetic/deidentified input/output comparisons for visible text, page geometry/rendering, page count, links, forms, annotations, fonts, metadata, and structure where applicable.
- **Unit:** profile/data validation, semantic repair decisions, template behavior, error handling, and audit evidence.
- **Adapter contract:** inputs, outputs, deterministic evidence, unsupported capabilities, cancellation/timeouts, retries, and license/capability policy.
- **Integration:** real open-source candidate workers and writers against synthetic fixtures after versions/licenses are verified.
- **End to end:** complete generation and remediation flows for statements, invoices, correspondence, and reports.
- **Failure injection and security:** malformed PDFs, hostile metadata, oversized/resource-intensive input, worker crash/timeout, retry/idempotence, privacy leakage, and audit integrity.
- **Independent machine validation:** use a separately selected validator; veraPDF is a candidate machine-only check, not a semantic or conformance guarantee.
- **Human gates:** semantic correctness review and assistive-technology checks by people, with applicable standards/profile and sampling policy still open.

For each vertical slice, tests follow red/green/refactor. Record actual test and validation results only after they are run; do not fabricate results, thresholds, or acceptance levels. Maintain synthetic fixture provenance and do not add confidential samples.
