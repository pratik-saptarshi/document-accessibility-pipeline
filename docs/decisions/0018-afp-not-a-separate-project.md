# Decision 0018: AFP migration is not a separate project

**Status: Accepted.** The product owner (Pratik Saptarshi) decided NO on
2026-10-19: IBM AFP migration does not merit a separate project. This
overrules the erroneous YES recorded by the autonomous pipeline in unmerged
PR #169 (closed without merging the same day). That PR treated the published
go-afp-line library as evidence *for* a separate project; the owner
corrected the reading: the library's existence is evidence *against* one.

## Decision

**Decision: NO.**

go-afp-line (https://github.com/pratik-saptarshi/go-afp-line, Apache-2.0,
published 2026-10-14) is a library project and repository meant to be used by
document-accessibility-pipeline. AFP migration work therefore proceeds as a
library dependency of the pipeline, not as a separate project or track.

## Limits

This decision answers only dap-jo9.4.1 ("Decide whether IBM AFP migration
merits a separate project"). It does not decide the AFP track go/no-go
(dap-jo9.4.2 stays open with the product owner), and it does not change
go-afp-line's own repository governance, licensing, or release process.

## Follow-up

- dap-jo9.4.1 is closed with this record as evidence.
- dap-jo9.4.2 (AFP track go/no-go/unknown with owner and input criteria)
  remains open and still needs the owner's call.
