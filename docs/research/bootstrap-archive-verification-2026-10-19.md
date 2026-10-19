# Bootstrap archive verification evidence

Date: 2026-10-19 UTC
PR head reviewed: `dc228cd0946ad60f00bb107d70f1824911fe7cd8`
Platform: Linux x86_64
Go: `go1.27.2 linux/amd64`, obtained from the pinned official distribution; archive SHA-256 `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`

## Artifact identity

- Source revision: `ef4c35629c467887b6749f9e31af111cbe8f8913`
- Storage revision: `4ce49c0b4face9388f036e08f3e05af7a3bf58a2`
- Raw path: `buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz`
- Immutable URL: `https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/4ce49c0b4face9388f036e08f3e05af7a3bf58a2/buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913-filtered-f58e49c961e21ab1.tar.gz`
- Expected and observed size: `388215` bytes
- Expected and observed SHA-256: `f58e49c961e21ab179f3f82637a1234eafb58f048dc6f93bd0ae8cec62aa314d`
- Archive root: `document-accessibility-pipeline-ef4c35629c467887b6749f9e31af111cbe8f8913/`

The exact size and digest were checked before extraction. The extracted source came from this archive; no checkout source was substituted for the module under test or built for the no-checkout install. The entrypoint was staged alone in a disposable directory with no `cmd/cloud-tools-bootstrap/go.mod` or repository manifest present, forcing its immutable archive path.

## Archived tests

Command, from the extracted archive's `cmd/cloud-tools-bootstrap` directory:

```sh
GOTOOLCHAIN=local GOWORK=off go test -race -count=1 -v ./internal/cloudtools
```

`GOCACHE` and `GOMODCACHE` were directed to writable disposable storage. Run: `2026-10-19T02:43:05Z`–`2026-10-19T02:43:22Z`; exit status `0`; package result `ok`, 16.702s. `-count=1` forced test execution. The required individual controls passed:

- `TestInstalledOKRejectsOversizedVersionEvidence`
- `TestInstalledOKRejectsOversizedFallbackVersionEvidence`
- `TestBoundedOutputCapsFiniteWritesAndMarksTruncation`
- `TestInstallGoModuleDiscardsOutputAndPreservesExitClassification` (success and nonzero-exit cases)
- `TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification` (success and nonzero-exit cases)
- `TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification` (success and nonzero-metadata cases)
- `TestVerifyInstalledCancellationDuringVersionCommandRemainsCancellation`
- `TestCommandCancellationKillsDescendantHoldingCombinedOutputPipe`

The full verbose test output follows in the transcript section. The tests' assertions verify fail-closed truncation, preserved classification, cancellation, and absence of synthetic child-output/secret markers in diagnostics.

## No-checkout install and repeat

The staged entrypoint ran on the same Linux x86_64 platform, using its pinned Go 1.27.2 distribution and immutable archive. All tool versions and binary digests remained at the existing manifest pins. Fresh install ran `2026-10-19T02:43:30Z`–`02:44:30Z` with exit status `0`; the separate repeat invocation ran `02:44:30Z`–`02:44:33Z` with exit status `0`; the separate `cloud-tools-bootstrap --verify` ran `02:44:33Z`–`02:44:34Z` with exit status `0`. Command discovery found all eight tools. The installer reported NeuroStrata and NeuroPlasticity as not functionally verified; no models were downloaded and neither tool was executed. Java and veraPDF were not run.

## Negative controls

- `TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/wrong_size` and `/wrong_checksum_at_pinned_size` passed. For separate timestamped disposable runs at `2026-10-19T02:46:03Z`, both the extraction and Go-build sentinels remained absent. Wrong size reported `Bootstrap source size mismatch; no bootstrap command was built.`; wrong hash at the pinned size reported `Bootstrap source checksum mismatch; no bootstrap command was built.`. Both exited `1` before extraction or build.
- `TestNoCheckoutRejectsWrongRevisionArchiveBytesBeforeExecution` passed at `2026-10-19T02:58:01Z`–`02:58:03Z`. Its reproducible harness is [in the bootstrap signal tests](../../cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap_signals_linux_test.go). It creates a valid synthetic tar with a different 40-character source-revision root, pads its compressed bytes to the unchanged pinned size, and serves it through an isolated `curl` stub. The unmodified entrypoint's trusted size and SHA pins remain intact; the equal-size bytes fail the expected SHA check. The extraction, Go-build, and bootstrap-command execution sentinels all remain absent.
- A disposable manifest with source revision `1111111111111111111111111111111111111111` but the pinned source path was rejected by the archive-built command at `2026-10-19T02:45:49Z` with exit status `1` and `bootstrap source must pin the expected raw repository file, distinct full source/storage revisions, SHA-256, and positive size`. The supplied Go-command sentinel and bootstrap-install sentinel both remained absent. This validates rejection before any pinned Go/tool action.
- `TestManifestRejectsMutableAndUnpinnedSources` passed, including wrong repository path and storage revision/path mismatch cases.

The earlier no-checkout transcript reported for `2026-10-19T00:20:09Z` was not present in this workspace. This fresh, timestamped run replaces it for the acceptance evidence; no claim is made about inspecting that earlier transcript.

## Criterion map and limits

1. Protected-main immutable source publication and synchronized pins: existing PR #153 artifact; identity rechecked above. No pin or tool version changed.
2. Download, pre-extraction size/digest verification, and accepted source root: passed above.
3. Archived output, classification, and cancellation controls: all named tests passed uncached.
4. Fresh no-checkout install, separate repeat, and separate verification: passed above with no checkout fallback.
5. Wrong source revision, size, and hash controls: passed as described. The same-size wrong-revision archive bytes were rejected at the entrypoint acquisition boundary with extraction, build, and execution sentinels absent; wrong-size and wrong-hash controls also rejected before extraction/build.
6. Saved-environment configuration and CI-only validator boundary: documentation continues to state saved configurations are unchanged and Java/veraPDF remains CI-only.

The canonical tracker export still records `dap-5hy.6` as closed. The documented export reconciler permits only `title`, `description`, `acceptance_criteria`, and `notes` updates for existing issues; it does not support lifecycle status, `closed_at`, or `close_reason` corrections. A guarded update plan limited to this one issue and those three lifecycle fields was rejected before bundle creation with `create: invalid update plan: plan field "status" is not allowed` (exit status `2` from the reconciler; canonical export unchanged). `bd show dap-5hy.6` against this workspace's live Beads store returned “Issue dap-5hy.6 not found”; that live store was not initialized or mutated. Therefore the status discrepancy remains explicit and unresolved in this evidence record. Fresh independent review and final-head CI are also pending.

## Sanitized command transcript

Private machine paths are replaced with `<isolated-task-path>`. No credentials or real child secrets are present.


### Archived race-test transcript

```text
start_utc=2026-10-19T02:43:05Z
=== RUN   TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock
=== RUN   TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock/interrupt
=== RUN   TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock/terminate
--- PASS: TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock (4.50s)
    --- PASS: TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock/interrupt (2.22s)
    --- PASS: TestBootstrapEntrypointSignalsStopChildBeforeReleasingLock/terminate (2.28s)
=== RUN   TestBootstrapEntrypointStopsSignalDuringChildStartup
--- PASS: TestBootstrapEntrypointStopsSignalDuringChildStartup (2.15s)
=== RUN   TestBootstrapEntrypointSignalAfterReadinessBeforeAcknowledgement
--- PASS: TestBootstrapEntrypointSignalAfterReadinessBeforeAcknowledgement (2.17s)
=== RUN   TestBootstrapEntrypointSignalDuringVersionCheckKillsForegroundTool
--- PASS: TestBootstrapEntrypointSignalDuringVersionCheckKillsForegroundTool (2.16s)
=== RUN   TestBootstrapEntrypointSourceBuildDeadlineCleansDescendantBeforeUnlock
--- PASS: TestBootstrapEntrypointSourceBuildDeadlineCleansDescendantBeforeUnlock (3.29s)
=== RUN   TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput
=== RUN   TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput/nonzero-version
=== RUN   TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput/timeout-after-output
--- PASS: TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput (0.19s)
    --- PASS: TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput/nonzero-version (0.09s)
    --- PASS: TestBootstrapEntrypointRejectsVersionPipelineFailureAfterExpectedOutput/timeout-after-output (0.09s)
=== RUN   TestBootstrapEntrypointRejectsUnsuccessfulDownloadedToolchainVersion
--- PASS: TestBootstrapEntrypointRejectsUnsuccessfulDownloadedToolchainVersion (0.25s)
=== RUN   TestBootstrapEntrypointPreservesSuccessfulInstallAndVerifyFlow
--- PASS: TestBootstrapEntrypointPreservesSuccessfulInstallAndVerifyFlow (0.23s)
=== RUN   TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction
=== RUN   TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/wrong_size
=== RUN   TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/wrong_checksum_at_pinned_size
=== RUN   TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/unreachable_after_invalid_cached_archive
--- PASS: TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction (0.51s)
    --- PASS: TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/wrong_size (0.17s)
    --- PASS: TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/wrong_checksum_at_pinned_size (0.17s)
    --- PASS: TestNoCheckoutBootstrapRejectsInvalidSourceArchiveBeforeExtraction/unreachable_after_invalid_cached_archive (0.17s)
=== RUN   TestBootstrapChecksChecksumBeforeInstallAndCleansTemporaryFiles
--- PASS: TestBootstrapChecksChecksumBeforeInstallAndCleansTemporaryFiles (0.02s)
=== RUN   TestBootstrapIsIdempotentAndVerifiesVersion
--- PASS: TestBootstrapIsIdempotentAndVerifiesVersion (0.02s)
=== RUN   TestBootstrapRejectsWrongCLIVersionEvenWithCorrectChecksum
--- PASS: TestBootstrapRejectsWrongCLIVersionEvenWithCorrectChecksum (0.01s)
=== RUN   TestBlockedRepairDoesNotReportPreexistingWrongVersionAsVerified
--- PASS: TestBlockedRepairDoesNotReportPreexistingWrongVersionAsVerified (0.02s)
=== RUN   TestTarExtractionRequiresSafeRegularMember
--- PASS: TestTarExtractionRequiresSafeRegularMember (0.02s)
=== RUN   TestConcurrentBootstrapTimesOutOnOwnedLock
--- PASS: TestConcurrentBootstrapTimesOutOnOwnedLock (0.05s)
=== RUN   TestCommandCancellationKillsDescendantHoldingCombinedOutputPipe
--- PASS: TestCommandCancellationKillsDescendantHoldingCombinedOutputPipe (0.01s)
=== RUN   TestActivationCanBeSourcedRepeatedlyWithoutChangingPATHAgain
--- PASS: TestActivationCanBeSourcedRepeatedlyWithoutChangingPATHAgain (0.02s)
=== RUN   TestWriteGoLauncherDoesNotOverwriteUnmanagedFile
--- PASS: TestWriteGoLauncherDoesNotOverwriteUnmanagedFile (0.00s)
=== RUN   TestWriteGoLauncherRejectsActivationDirectorySymlink
--- PASS: TestWriteGoLauncherRejectsActivationDirectorySymlink (0.00s)
=== RUN   TestWriteGoLauncherRejectsSymlinkSelectedBinary
--- PASS: TestWriteGoLauncherRejectsSymlinkSelectedBinary (0.00s)
=== RUN   TestBootstrapCancellationDoesNotPublishGoLauncher
--- PASS: TestBootstrapCancellationDoesNotPublishGoLauncher (0.00s)
=== RUN   TestCloudBootstrapEntryPointRejectsUnexpectedArgsBeforeWrites
--- PASS: TestCloudBootstrapEntryPointRejectsUnexpectedArgsBeforeWrites (0.00s)
=== RUN   TestManifestErrorTextDoesNotIncludeURLQuery
--- PASS: TestManifestErrorTextDoesNotIncludeURLQuery (0.01s)
=== RUN   TestEveryConfiguredVersionPatternRejectsLongerNumericPrefix
--- PASS: TestEveryConfiguredVersionPatternRejectsLongerNumericPrefix (0.00s)
=== RUN   TestGoBuildInfoVersionFallbackRequiresExactModuleAndVersion
--- PASS: TestGoBuildInfoVersionFallbackRequiresExactModuleAndVersion (0.00s)
=== RUN   TestInstalledOKRejectsOversizedVersionEvidence
--- PASS: TestInstalledOKRejectsOversizedVersionEvidence (0.00s)
=== RUN   TestInstalledOKRejectsOversizedFallbackVersionEvidence
--- PASS: TestInstalledOKRejectsOversizedFallbackVersionEvidence (0.01s)
=== RUN   TestBoundedOutputCapsFiniteWritesAndMarksTruncation
--- PASS: TestBoundedOutputCapsFiniteWritesAndMarksTruncation (0.00s)
=== RUN   TestInstallGoModuleDiscardsOutputAndPreservesExitClassification
=== RUN   TestInstallGoModuleDiscardsOutputAndPreservesExitClassification/success
=== RUN   TestInstallGoModuleDiscardsOutputAndPreservesExitClassification/nonzero_exit
--- PASS: TestInstallGoModuleDiscardsOutputAndPreservesExitClassification (0.00s)
    --- PASS: TestInstallGoModuleDiscardsOutputAndPreservesExitClassification/success (0.00s)
    --- PASS: TestInstallGoModuleDiscardsOutputAndPreservesExitClassification/nonzero_exit (0.00s)
=== RUN   TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification
=== RUN   TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification/successful_version_output
=== RUN   TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification/nonzero_exit_despite_matching_output
--- PASS: TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification (0.00s)
    --- PASS: TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification/successful_version_output (0.00s)
    --- PASS: TestInstalledOKKeepsSmallVersionSuccessAndFailureClassification/nonzero_exit_despite_matching_output (0.00s)
=== RUN   TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification
=== RUN   TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification/successful_module_metadata
=== RUN   TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification/nonzero_metadata_command
--- PASS: TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification (0.01s)
    --- PASS: TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification/successful_module_metadata (0.00s)
    --- PASS: TestInstalledOKKeepsSmallFallbackSuccessAndFailureClassification/nonzero_metadata_command (0.00s)
=== RUN   TestVerifyInstalledCancellationDuringVersionCommandRemainsCancellation
--- PASS: TestVerifyInstalledCancellationDuringVersionCommandRemainsCancellation (0.01s)
=== RUN   TestInstallGoModuleUsesOfficialProxyAndChecksumDatabaseWithoutFallback
--- PASS: TestInstallGoModuleUsesOfficialProxyAndChecksumDatabaseWithoutFallback (0.00s)
=== RUN   TestInstallGoModuleIgnoresPersistedGoEnvProxyBypass
--- PASS: TestInstallGoModuleIgnoresPersistedGoEnvProxyBypass (0.01s)
=== RUN   TestRepositoryManifestKeepsEightToolsAndNoValidatorArtifacts
--- PASS: TestRepositoryManifestKeepsEightToolsAndNoValidatorArtifacts (0.00s)
=== RUN   TestManifestRejectsMutableAndUnpinnedSources
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/missing_size
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/zero_size
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/negative_size
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/missing_storage_revision
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/abbreviated_storage_revision
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/same_content_and_storage_revision
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/mutable_branch_URL
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/mutable_tag_URL
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/generated_archive_URL
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/wrong_repository
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/wrong_manifest_path
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/wrong_URL_path
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/storage_revision_mismatch
=== RUN   TestManifestRejectsMutableAndUnpinnedSources/query_string
--- PASS: TestManifestRejectsMutableAndUnpinnedSources (0.01s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/missing_size (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/zero_size (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/negative_size (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/missing_storage_revision (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/abbreviated_storage_revision (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/same_content_and_storage_revision (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/mutable_branch_URL (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/mutable_tag_URL (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/generated_archive_URL (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/wrong_repository (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/wrong_manifest_path (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/wrong_URL_path (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/storage_revision_mismatch (0.00s)
    --- PASS: TestManifestRejectsMutableAndUnpinnedSources/query_string (0.00s)
PASS
ok  	github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap/internal/cloudtools	16.702s
exit_status=0
end_utc=2026-10-19T02:43:22Z
```

### Fresh no-checkout install transcript

```text
fresh_install_start_utc=2026-10-19T02:43:30Z
Installed Go toolchain 1.27.2 in <isolated-task-path>
verified bootstrap toolchain go 1.27.2
installed bd 1.3.1
installed neurostrata 1.9.0
installed neuroplasticity 2.1.0
installed actionlint 1.7.12
installed govulncheck 1.9.0
installed go-licenses 2.0.1
installed gitleaks 8.30.1
Tool root: <isolated-task-path>
Activate for this shell with: source <isolated-task-path>
VERIFIED go 1.27.2 at <isolated-task-path> (verified version/digest)
VERIFIED bd 1.3.1 at <isolated-task-path> (verified version/digest)
VERIFIED neurostrata 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED neuroplasticity 2.1.0 at <isolated-task-path> (verified version/digest)
VERIFIED actionlint 1.7.12 at <isolated-task-path> (verified version/digest)
VERIFIED govulncheck 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED go-licenses 2.0.1 at <isolated-task-path> (verified version/digest)
VERIFIED gitleaks 8.30.1 at <isolated-task-path> (verified version/digest)

For this task shell, source <isolated-task-path>

Operational augmentation status: NOT VERIFIED by binary installation.
NeuroStrata requires a writable persistent data home, approved local embedder model/runtime, and a successful MCP initialize plus synthetic namespace smoke check.
NeuroPlasticity requires a reviewed plasticity.json, supported rootless Podman, approved optimizer model, safe agent invocation, and a disposable synthetic evaluation.
Fresh tasks must source <isolated-task-path> and repeat the command-discovery check.
fresh_install_exit_status=0
fresh_install_end_utc=2026-10-19T02:44:30Z
```

### Separate repeat invocation transcript

```text
repeat_start_utc=2026-10-19T02:44:30Z
Verified existing Go toolchain 1.27.2
verified bootstrap toolchain go 1.27.2
verified bd 1.3.1
verified neurostrata 1.9.0
verified neuroplasticity 2.1.0
verified actionlint 1.7.12
verified govulncheck 1.9.0
verified go-licenses 2.0.1
verified gitleaks 8.30.1
Tool root: <isolated-task-path>
Activate for this shell with: source <isolated-task-path>
VERIFIED go 1.27.2 at <isolated-task-path> (verified version/digest)
VERIFIED bd 1.3.1 at <isolated-task-path> (verified version/digest)
VERIFIED neurostrata 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED neuroplasticity 2.1.0 at <isolated-task-path> (verified version/digest)
VERIFIED actionlint 1.7.12 at <isolated-task-path> (verified version/digest)
VERIFIED govulncheck 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED go-licenses 2.0.1 at <isolated-task-path> (verified version/digest)
VERIFIED gitleaks 8.30.1 at <isolated-task-path> (verified version/digest)

For this task shell, source <isolated-task-path>

Operational augmentation status: NOT VERIFIED by binary installation.
NeuroStrata requires a writable persistent data home, approved local embedder model/runtime, and a successful MCP initialize plus synthetic namespace smoke check.
NeuroPlasticity requires a reviewed plasticity.json, supported rootless Podman, approved optimizer model, safe agent invocation, and a disposable synthetic evaluation.
Fresh tasks must source <isolated-task-path> and repeat the command-discovery check.
repeat_exit_status=0
repeat_end_utc=2026-10-19T02:44:33Z
```

### Separate verification and command-discovery transcript

```text
separate_verify_start_utc=2026-10-19T02:44:33Z
VERIFIED go 1.27.2 at <isolated-task-path> (verified version/digest)
VERIFIED bd 1.3.1 at <isolated-task-path> (verified version/digest)
VERIFIED neurostrata 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED neuroplasticity 2.1.0 at <isolated-task-path> (verified version/digest)
VERIFIED actionlint 1.7.12 at <isolated-task-path> (verified version/digest)
VERIFIED govulncheck 1.9.0 at <isolated-task-path> (verified version/digest)
VERIFIED go-licenses 2.0.1 at <isolated-task-path> (verified version/digest)
VERIFIED gitleaks 8.30.1 at <isolated-task-path> (verified version/digest)
separate_verify_exit_status=0
separate_verify_end_utc=2026-10-19T02:44:34Z
command_discovery_exit_status=0 tools=8
```


### Source revision control transcript

```text
start_utc=2026-10-19T02:45:49Z
bootstrap source must pin the expected raw repository file, distinct full source/storage revisions, SHA-256, and positive size
exit_status=1
go_execution_sentinel=absent
bootstrap_install_sentinel=absent
end_utc=2026-10-19T02:45:49Z
```

### Wrong-size and wrong-hash extraction/build sentinel transcripts

```text
start_utc=2026-10-19T02:46:03Z
Verified existing Go toolchain 1.27.2
Bootstrap source size mismatch; no bootstrap command was built.
exit_status=1
extraction_sentinel=absent
build_sentinel=absent
end_utc=2026-10-19T02:46:03Z
```

### Wrong-revision archive acquisition control

Command from `cmd/cloud-tools-bootstrap`:

```sh
GOTOOLCHAIN=local GOWORK=off go test -race -count=1 -v ./internal/cloudtools -run '^TestNoCheckoutRejectsWrongRevisionArchiveBytesBeforeExecution$'
```

The unmodified entrypoint pin values are read from `scripts/bootstrap-cloud-build-tools.sh`. The synthetic wrong-root archive has the pinned size but a different digest. Extraction, Go-build, and bootstrap execution are observable in the committed harness and remain absent in this run.

```text
start_utc=2026-10-19T02:58:01Z
=== RUN   TestNoCheckoutRejectsWrongRevisionArchiveBytesBeforeExecution
--- PASS: TestNoCheckoutRejectsWrongRevisionArchiveBytesBeforeExecution (0.21s)
PASS
ok   github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap/internal/cloudtools 1.234s
exit_status=0
end_utc=2026-10-19T02:58:03Z
```

### Guarded tracker lifecycle correction transcript

```text
plan fields: status=in_progress, closed_at="", close_reason=""
create: invalid update plan: plan field "status" is not allowed
exit status 2
result: no evidence bundle created; canonical export unchanged
```

```text
start_utc=2026-10-19T02:46:03Z
Verified existing Go toolchain 1.27.2
Bootstrap source checksum mismatch; no bootstrap command was built.
exit_status=1
extraction_sentinel=absent
build_sentinel=absent
end_utc=2026-10-19T02:46:03Z
```
