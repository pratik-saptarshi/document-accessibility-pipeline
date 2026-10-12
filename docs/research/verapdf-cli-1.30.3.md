# veraPDF CLI 1.30.3 acquisition check

**Checked:** 2026-10-12. **State: candidate identified; not authenticated or installed.** This document records route observations only. It does not pin the artifact or approve a validator run.

## Maven Central candidate

- Artifact: [`org.verapdf.apps:cli:1.30.3`](https://repo.maven.apache.org/maven2/org/verapdf/apps/cli/1.30.3/cli-1.30.3.jar)
- POM: [`cli-1.30.3.pom`](https://repo.maven.apache.org/maven2/org/verapdf/apps/cli/1.30.3/cli-1.30.3.pom)
- Detached signature: [`cli-1.30.3.jar.asc`](https://repo.maven.apache.org/maven2/org/verapdf/apps/cli/1.30.3/cli-1.30.3.jar.asc)
- Published SHA-256 sidecar: [`cli-1.30.3.jar.sha256`](https://repo.maven.apache.org/maven2/org/verapdf/apps/cli/1.30.3/cli-1.30.3.jar.sha256)

The artifact URL returned HTTP 200 with `application/java-archive` and `content-length: 16057132`. The detached signature, POM, POM signature, and SHA-256 sidecar URLs also returned HTTP 200. The sidecar text is `4c8acbaeb16405456c8d6d76deac8cf40d94618cb4e33aedf44e51ee1b68be91`. This digest is **not a trusted pin**: the JAR has not been downloaded and hashed locally, and neither signature has been verified cryptographically.

Inspection of the detached signature packet (without a key) reports issuer fingerprint `13DD102B4DD69354D12DE5A83184863278B17FE7`, key ID `3184863278B17FE7`, and signature creation date 2026-10-07. The published veraPDF installer documentation identifies the same fingerprint. This is signer metadata, not signature verification. The POM describes the CLI as GPLv3+ or MPLv2+ and sets `org.verapdf.apps.GreenfieldCliWrapper` as the shaded JAR main class. Any allowed use remains CI-only with an explicit MPL-2.0 election after license review; no product-runtime exception is approved.

The official [CLI validation guide](https://docs.verapdf.org/cli/validation/) lists `ua2`, documents explicit profile selection with `--flavour`, and shows the XML report structure as `report/jobs/job/validationReport`; the inner element supplies `profileName` and `isCompliant`. The [v1.30.3 CLI source](https://github.com/veraPDF/veraPDF-apps/blob/v1.30.3/cli/src/main/java/org/verapdf/cli/CliConstants.java) defines distinct compliant, noncompliant, and processing-error exit codes. The adapter accepts only one job and one inner validation report, requires the exact UA-2 profile name, and checks that pass/finding status agrees with exit codes 0/1. Its hand-authored report fixtures mirror only the documented XML envelope and are clearly marked as synthetic parser samples.

## Authentication blocker and exact network response

The public key needed to verify the detached Maven signature is linked by the official installation documentation at [`https://software.verapdf.org/keys/KEY`](https://software.verapdf.org/keys/KEY). This URL, the official installer `https://software.verapdf.org/releases/verapdf-installer.zip`, and its signature `https://software.verapdf.org/releases/verapdf-installer.zip.asc` each returned the following on 2026-10-12 04:33 UTC:

```text
curl: (56) CONNECT tunnel failed, response 403
HTTP/1.1 403 Forbidden
content-length: 16
content-type: text/plain
server: envoy
connection: close
```

The CONNECT tunnel was denied before an upstream HTTP response completed; the 16-byte proxy body was not retrieved. This establishes an environment proxy denial, not an upstream veraPDF rejection. No alternate key host or mirror was used. A narrow environment/setup change is needed to retrieve the official public key through the allowed route.

## Gate before use

1. Retrieve the authoritative key at the official key URL through an allowed route and verify its fingerprint against the documented fingerprint.
2. Cryptographically verify the Maven JAR signature with that key. Do not infer that the installer signature authenticates Maven.
3. Download the exact JAR only after the signature-verification route is available, compute SHA-256 locally, compare with the publisher sidecar, and record both verified values and the acquisition process.
4. Record the MPL-2.0 election and CI-only use. Pin the locally computed digest in any future CI download step.
5. Run the actual CLI with explicit `--flavour ua2`; do not use automatic profile detection or fallback. Create separate genuine PDF/UA-2 pass and fail controls only after the authenticated CLI is available. Keep fake-process/XML report tests labeled as shim tests.

No executable was installed, no artifact was added to the repository, and no checksum pin was created during this check.
