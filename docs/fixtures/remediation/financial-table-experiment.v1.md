# Financial-table experiment protocol v1

**Status:** reviewed engineering configuration for a finite synthetic experiment. This protocol does not approve a production profile, writer, or support claim.

## Priority and operation

Statements and invoices with tables have equal priority under [Decision 0002](../../decisions/0002-initial-remediation-family-priority.md). Exercise exactly one one-page statement transaction table and one one-page invoice line-item table. Each has one header row, two data rows, three columns, and one explicitly mapped amount header. The experiment parameter is PDF 2.0, English text, Helvetica, and a simple unmerged table.

The semantic map is fixture-authored. The normative positive representation marks each column header as `/TH` with a unique `/ID`, and `/StructTreeRoot /IDTree` is a sorted name tree mapping each identifier to that exact header element. Each data cell has a table attribute dictionary with `/O /Table` and `/Headers` containing the associated header ID. The damaged input is derived from that positive representation by changing only the `Headers` association on the second data-row amount cell: it points to the second-column header ID instead of the amount header ID. Its visual content, text, structure order, header scopes, identifier tree, and every other cell association remain unchanged. This directly tests the PDF table header association defined by the standard; it does not imply PDF/UA-2 conformance. The `/IDTree` requirement and mappings follow the logical-structure rules in [ISO 32000-2:2020](https://pdfa.org/resource/iso-32000-2/) (see also the [PDF Association logical-structure cheat sheet](https://pdfa.org/download-area/cheat-sheets/LogicalStructureObjects.pdf)) and the table attribute definitions in Adobe's [accessibility attribute reference](https://opensource.adobe.com/dc-acrobat-sdk-docs/library/accessibility/Access_DOM.html).

The explicit expected association is the target amount header to the amount cells in both data rows. The Go characterization check resolves the name tree to the actual `/TH` elements, follows each cell's `/K` MCID to the page's marked-content text, and cross-checks `/Pg`, `/P`, `/StructParents`, and the page `/ParentTree` reverse binding. It requires exactly one supported literal `Tj` string for each marked cell. Negative controls swap MCIDs while keeping structure `/T` titles and `/Headers` intact, add a second `Tj`, and remove or mismatch `/StructParents`; each must fail the intended check. This is a bounded check of the fixture's simple, single-stream syntax using pdfcpu's stream decoder, not a general PDF content parser. A structure-type pass alone does not establish this predicate.

## Evidence and controls

For each family, retain the normative positive PDF and its derived damaged input as separate immutable files. The damaged input hash is the input identity for any later repair attempt. Rerun the exact Poppler and Ghostscript commands with `scripts/observe-financial-fixtures.sh`; it records input/output SHA-256, tool version, command, and exit status in `financial-tables/observations/observations.v1.json` and retains all extracted text and rendered PNG outputs next to the manifest. The Go unit test checks the retained outputs and recorded hashes against the fixture PDF identities; it does not rerun those external tools. Parsed structure/association observations are independently checked with pdfcpu in the Go test. Keep exact displayed strings, including signs, decimals, and currency symbols; do not calculate, normalize, or infer totals.

These samples are intentionally not whole-document conformance fixtures: the sample heading is untagged, and the Helvetica base font is not embedded. Later PDF/UA-2 validation may therefore report unrelated whole-document failures. Do not attribute those failures to table association or broaden this work into whole-document certification.

Controls are defined before repair code:

- **Positive representation:** each TD's `/Headers` references its expected column TH `/ID`; the independent association predicate passes.
- **Association defect:** the derived damaged input's second-row amount TD references the second-column TH `/ID`; the independent predicate identifies exactly that misassociation while text and rendering remain identical to the positive.
- **Amount change:** changing any displayed financial string fails exact-string comparison, regardless of whether a machine validator accepts the PDF.
- **Malformed:** truncated or structurally unparsable PDF input is rejected before a repair candidate is emitted.
- **Unsupported:** merged, nested, multi-page, or otherwise out-of-scope table topology returns unsupported; no inferred repair is attempted.
- **Ambiguous map:** absent, multiply matched, or contradictory fixture-authored target associations require review-needed; no guessed association is accepted.

The existing veraPDF positive/finding oracle pair only tests ISO 14289-2:2024 clause 8.2.4 structure types ([control provenance](../../../internal/pdfua/testdata/README.md)). It authenticates the isolated validator/report path, not table-header semantics. No table-specific veraPDF outcome is claimed here. Any future table-specific validator result requires explicit report-only CI wiring for the exact fixture/candidate hash and an actual CI report. Validator acceptance and semantic correctness remain separate observations.

The fixture test also checks the one uncompressed classic xref table used by each authored PDF: exactly one table, correct object offsets, matching trailer size, `startxref` pointing to the table, and a final EOF marker. This checks the syntax of these intentionally simple fixtures; it is not general PDF syntax validation.

## Boundaries and ownership

Fixed language, font, values, and topology are experiment parameters, not production defaults, restrictions, or a general support promise. No arithmetic, tax, currency, date, rounding, privacy, retention, or reviewer-authority policy is selected. No writer or runtime dependency is selected. `dap-jo9.2.1.2` owns fixture bytes, hashes, baselines, and the pre-repair controls; `dap-jo9.1.2` owns attempted repair, candidate evidence, and finite disposition after exact candidate clearance in `dap-jo9.1.3`.

## Later writer feasibility risk

Source inspection of pinned pdfcpu v0.16.2 predicts that its ordinary PDF 2.0 full-write path may expand this fixture's minimal Helvetica dictionary with `FirstChar`, `LastChar`, `Widths`, and a `FontDescriptor` through `EnsureCoreFontMetrics` ([pinned source](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/pkg/pdfcpu/font/fontDict.go#L560-L590), [write-path caller](https://github.com/pdfcpu/pdfcpu/blob/v0.16.2/pkg/pdfcpu/writeFont.go#L30-L42)). This is a source-derived prediction, not an executed experiment or an observed output delta. The later bounded spike must test the exact positive fixture, record the full object/render delta, and make a reviewed no-go or explicit protocol decision if extra changes conflict with preservation. Do not weaken the current baseline to hide them. A later candidate must be staged separately, validated against immutable input and its intended disposition, then explicitly published; a successful `WriteContextFile` call alone does not publish a validated candidate atomically.
