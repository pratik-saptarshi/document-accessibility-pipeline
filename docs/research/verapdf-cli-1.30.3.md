# veraPDF CLI 1.30.3 authenticated CI integration

**Checked:** 2026-10-12. **State:** official artifact authenticated; CI-only pin and real CLI integration are implemented. This is machine-validation evidence only, not certification, semantic review, or an assistive-technology result.

## Official artifact and signature

The selected artifact is the official Greenfield installer [`verapdf-greenfield-1.30.3-installer.zip`](https://software.verapdf.org/releases/1.30/verapdf-greenfield-1.30.3-installer.zip) and its detached `.asc` signature. The official [installation guide](https://docs.verapdf.org/install/) directs users to [`https://software.verapdf.org/keys/KEY`](https://software.verapdf.org/keys/KEY) and publishes primary-key fingerprint `13DD 102B 4DD6 9354 D12D E5A8 3184 8632 78B1 7FE7`. On 2026-10-12, the installer, signature, and public-key routes each returned HTTP 200. The downloaded key's primary fingerprint exactly matched the fingerprint published in the official documentation.

GPG verified the downloaded `.asc` against the exact installer ZIP and reported a good RSA signature by `13DD102B4DD69354D12DE5A83184863278B17FE7`, created 2026-10-07 13:53:13 UTC. GPG marked the key's trust as unknown because this checkout does not establish a separate web-of-trust certification; signer identity is anchored by comparing the key from veraPDF's official key endpoint with the fingerprint in its official documentation. The primary signing key is valid through 2027-07-20; its expired encryption-only subkey was not used to verify the signature.

| Artifact | SHA-256 | Use |
| --- | --- | --- |
| Official installer ZIP | `b512bf24945503a33fc10738ecffc1f2dccaaf8593a448e94ee9ac86a9559a59` | Downloaded, locally hashed, and GPG-verified before extraction/execution |
| Installed `bin/cli-1.30.3.jar` | `19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70` | Inner CLI JAR extracted from that authenticated installer; pinned and rechecked in CI and by the Go adapter |

The Maven Central CLI JAR is not used. Its packaging differs from the installer-distributed JAR, so its detached signature and checksum are not inferred from the installer signature. No artifact is committed to the repository or included in product/runtime dependencies.

The reproducible CI acquisition is implemented in [the provisioning script](../../.github/scripts/provision-verapdf.sh) and [the integration workflow](../../.github/workflows/verapdf-integration.yml). The script checks the official key fingerprint, installer SHA-256, and actual detached signature before extracting or running the installer; it then checks the installed CLI JAR hash and reported version. The workflow keeps Java and veraPDF inside CI and runs the Go adapter test against the exact pinned JAR. The separate cloud build-tool bootstrap calls the same script in `--cache-only` mode: it verifies and retains only the signed installer ZIP under its workspace-local tool root. That path does not extract or execute the installer, add Java or a veraPDF executable to `PATH`, or run document validation.

## License review and boundary

The official [veraPDF-apps project](https://github.com/veraPDF/veraPDF-apps/tree/v1.30.3) offers its software under GPL-3.0-or-later or MPL-2.0-or-later. For this explicitly CI/CD-only validator use, the approved choice is MPL-2.0. No Java code or executable is a product/runtime dependency, and no downloaded artifact is stored in this repository.

The installer ZIP contains the IzPack installer JAR and launcher scripts, but no separate top-level license/notice file. The installed shaded CLI JAR includes bundled notice/license resources: Rhino MPL license and notices; Apache License 2.0 and Jackson notice; W3C license text; MIT and Boost notices for FastDoubleParser; Schubfach MIT terms; Oracle BSD-style `LICENSE.md`; and Eclipse JAXB notice. The CLI `--version` output also states GPLv3 and MPLv2-or-later. The notices remain within the downloaded JAR during CI execution. This is a bounded package-notice inspection, not a generated SBOM or a legal conclusion for redistribution; redistribution and production use remain outside this exception.

## Real PDF/UA-2 integration evidence

The CLI is invoked with `--flavour ua2 --format xml`; no automatic detection or fallback profile is allowed. In the exact 1.30.3 output, the selected profile is named `PDF/UA-2 + Tagged PDF validation profile`; the adapter pins that exact report value and fails closed for any other profile. The actual CLI uses exit code 0 for a compliant report and 1 for a noncompliant finding, consistent with the adapter's existing exit/report check.

Two unchanged CC BY 4.0 oracle controls come from the official [`veraPDF/veraPDF-corpus`](https://github.com/veraPDF/veraPDF-corpus) repository at commit `bb75f4f0073d9350dfd058c0162a367e6fadf25e`. Their provenance and SHA-256 values are recorded in [`internal/pdfua/testdata/README.md`](../../internal/pdfua/testdata/README.md): `8.2.4-t01-pass-a.pdf` (pass) and `8.2.4-t01-fail-a.pdf` (finding), both testing ISO 14289-2 clause 8.2.4. Running the authenticated CLI produced:

- Pass control: exit 0, profile `PDF/UA-2 + Tagged PDF validation profile`, `isCompliant="true"`, 0 failed rules/checks.
- Finding control: exit 1, the same profile, `isCompliant="false"`, 1 failed rule/check for a nonstandard structure element.
- The Go integration test exercises those real controls through the adapter and verifies that a modified copy of the authenticated JAR fails its pinned digest check. The package also retains synthetic-process operational tests for missing executables, cancellation/timeouts, output limits, malformed and mismatched reports, cleanup, and failure exit codes.

These controls are machine-validator oracles. A machine pass cannot establish that document meaning, reading order, or assistive-technology behavior is correct. Human semantic and assistive-technology review remain separate release gates.

## Route observations

The exact installer, `.asc`, and key URLs returned HTTP 200 from this environment. The previous Go module download route was also checked using the exact `pdfcpu v0.16.2` proxy URL: it redirected to `storage.googleapis.com` and completed with HTTP 200. The root storage URL alone returned HTTP 400, which is not used as the download test. No alternate artifact/key host or network/credential configuration was used.
