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

The user selected PDF/UA-2 for both PDF output lanes and the Go standard-library subprocess approach for the validator adapter. Official veraPDF CLI documentation defines the `ua2` flavour and XML `validationReport` profile/compliance fields. The official Greenfield 1.30.3 installer and detached signature were downloaded from `software.verapdf.org`; the key obtained from its official key URL matches the fingerprint published in the official installation guide, and GPG verified the exact installer signature. The installer ZIP and installed CLI JAR hashes are pinned in [the acquisition record](../research/verapdf-cli-1.30.3.md). The separate Maven JAR candidate is not used and its signature is not inferred from the installer signature.

The adapter contract and CI provisioning are implemented in `internal/pdfua` and `.github`. The exact authenticated CLI was run with `--flavour ua2` against genuine official corpus pass/finding controls, and the real report profile label is pinned. Synthetic XML reports and fake process tests remain shim tests only. The real CLI integration does not establish semantic meaning, reading order, or assistive-technology behavior.

## Consequences and limits

- Profile data stores the accepted PDF/UA target separately from open confidentiality, retention, family fields, arithmetic, semantic policy, and reviewer authority.
- Generation and remediation need distinct UA-2 pass/fail oracle PDFs for their respective workflows. The remediation validator integration controls cannot stand in for production family fixtures or human review.
- Engineering issue completion and release eligibility are different. No output is eligible for a conformance claim without machine evidence tied to the candidate hash plus independent human semantic and assistive-technology evidence.
- The CI-only exception covers only the pinned official veraPDF CLI under the approved MPL-2.0 election after signer verification and hash pinning. No production runtime or other non-Go runtime exception is included.

## Follow-up

- `dap-jo9.2.5.3`: implement and test the bounded CI-only adapter; actual CLI and oracle integration is recorded in `dap-jo9.2.5.4`.
- `dap-jo9.1.2`, `dap-jo9.2.1`, and generation engine/source feasibility issues: test candidate feasibility against PDF 2.0/PDF/UA-2 without treating the existing PDF 1.7 characterization corpus as conformance evidence.
- `dap-jo9.2.5.2`: retain independent semantic and assistive-technology review as a separate release gate.
