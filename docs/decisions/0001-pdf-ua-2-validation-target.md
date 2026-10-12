# Decision 0001: PDF/UA-2 validation target

**Status: Accepted for output verification.** This records the user's explicit target selection. It does not approve a writer, application architecture, input profile, release protocol, or claim of conformance.

## Decision

- Implementation is Go-first; product dependencies must have a permissive open-source license. This sets an implementation constraint, not an API/job architecture choice. Existing-PDF remediation remains the primary lane, with generation a separate lane.
- PDF/UA-2 is the target standard for generated PDF output and remediation candidate output.
- Validation must explicitly request veraPDF profile `ua2` with `--flavour ua2`. Automatic profile detection, PDF/UA-1, PDF/A, and other fallback profiles are prohibited.
- PDF/UA-2 output uses PDF 2.0 requirements. Existing PDF 1.7 preservation fixtures remain characterization inputs; they are not PDF/UA-2 positive controls.
- The initial veraPDF boundary is a CI-only Go subprocess adapter. It is not linked into product runtime or exposed as a service. No other runtime dependency exception is approved.
- A machine pass is one recorded validation outcome. It does not approve semantic meaning, reading order, or assistive-technology behavior. Independent human semantic and assistive-technology review remains required.
- Any writer or source-conversion feasibility gap against PDF 2.0/PDF/UA-2 is recorded as unsupported or blocked; the target is not weakened to make an implementation pass.

## Context and evidence

The user selected PDF/UA-2 for both PDF output lanes and the Go standard-library subprocess approach for the validator adapter. Official veraPDF CLI documentation defines the `ua2` flavour and XML `validationReport` profile/compliance fields. The Maven Central CLI JAR and its detached signature are available; signature packet metadata names fingerprint `13DD102B4DD69354D12DE5A83184863278B17FE7`, matching the fingerprint published in the official installation documentation. However, the public-key URL `https://software.verapdf.org/keys/KEY` is denied by the environment proxy, so the Maven signature is not yet cryptographically verified. The published SHA-256 sidecar is not yet a trusted pin; the JAR has not been downloaded or locally hashed. Exact URLs and proxy evidence are recorded in [the acquisition check](../research/verapdf-cli-1.30.3.md).

The adapter contract is prepared in `internal/pdfua`. Its synthetic XML reports and fake process are shim tests only; they are not veraPDF output. Real CLI execution, independently authenticated artifacts, and genuine PDF/UA-2 oracle PDFs remain blocked.

## Consequences and limits

- Profile data stores the accepted PDF/UA target separately from open confidentiality, retention, family fields, arithmetic, semantic policy, and reviewer authority.
- Generation and remediation need distinct UA-2 pass/fail oracle PDFs once the authentic tool is available. Existing preservation fixtures cannot stand in for those controls.
- Engineering issue completion and release eligibility are different. No output is eligible for a conformance claim without machine evidence tied to the candidate hash plus independent human semantic and assistive-technology evidence.
- The CI-only exception covers only a pinned official veraPDF CLI invocation after signer verification, hash pinning, and license election. No production runtime or other non-Go runtime exception is included.

## Follow-up

- `dap-jo9.2.5.3`: implement and test the bounded CI-only adapter; real-tool validation stays blocked until artifact authentication is resolved.
- `dap-jo9.1.2`, `dap-jo9.2.1`, and generation engine/source feasibility issues: test candidate feasibility against PDF 2.0/PDF/UA-2 without treating the existing PDF 1.7 characterization corpus as conformance evidence.
- `dap-jo9.2.5.2`: retain independent semantic and assistive-technology review as a separate release gate.
