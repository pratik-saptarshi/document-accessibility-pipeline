# Filesystem latency boundary for remediation inputs

Resolves `dap-jo9.1.6.5` (filesystem latency boundary for remediation
inputs), tracked from PR #50 review discussion_r4263777099. The v2
exact-fixture experiment is restricted to bounded regular local fixture
files. No observed stall or candidate-integrity defect motivated this
boundary; it is a truthful statement of what the input path guarantees
before broader storage/input support is added.

## Source and storage boundary

- Inputs are **regular local fixture files only**: absolute paths under the
  repository's fixture tree, opened and type-checked as regular files.
- Unsupported source types are rejected before any bytes are read:
  symlinks (rejected at open, not followed), FIFOs (rejected without
  blocking on open), directories, devices, and missing files.
- No network, FUSE, or removable storage is supported. A path on such a
  mount is not a supported input even if it happens to resolve to a regular
  file; the boundary is the fixture tree, not the file type alone.
- Memory and file-size bounds are preserved: reads are capped at the
  declared limit plus one byte (`64 KiB` for v2 fixture inputs, `256 KiB`
  for manifests and outputs), and oversized inputs are rejected.

## Truthful interruption and latency guarantee

- **Cancellation is checked before and after acquisition.** A context that
  is already cancelled fails before the file is opened; a cancellation that
  arrives during the read is observed after the read completes, and the
  read result is discarded.
- **No hard wall-clock deadline is promised for filesystem acquisition.**
  A blocked read on a slow or stalled source is not interrupted by context
  cancellation: Go's `context.Context` cannot abort a blocked `read`
  syscall. The context checks bound *when* acquisition may start and
  whether its result is used, not how long the read itself takes.
- **The latency bound is structural, not temporal.** The read cannot block
  on a FIFO (rejected without blocking), cannot follow a symlink into
  unexpected storage, and cannot allocate beyond the byte cap. What is not
  bounded is the wall-clock time of a read from a slow-but-regular local
  file. Any future hard filesystem deadline promise requires a separate
  mechanism (for example, a reader goroutine with a deadline, or a
  deadline-aware open) and its own tests; it must not be claimed from the
  context checks alone.

## What this does not change

- The current exact-fixture feasibility experiment is unaffected: it already
  used regular fixture files with context checks before and after
  acquisition.
- No broader storage or input support is added. Adding any is gated on
  this boundary: the new source type must be named, its interruption and
  latency behavior documented, and tests added for its unsupported cases
  before it is accepted.
