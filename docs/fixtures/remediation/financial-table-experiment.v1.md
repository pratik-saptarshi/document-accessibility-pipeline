# Financial-table experiment protocol v1

**Status:** reviewed engineering configuration for a finite synthetic experiment. This protocol does not approve a production profile, writer, or support claim.

## Priority and operation

Statements and invoices with tables have equal priority under [Decision 0002](../../decisions/0002-initial-remediation-family-priority.md). Exercise exactly one one-page statement transaction table and one one-page invoice line-item table. Each has one header row, two data rows, three columns, and one explicitly mapped amount header. The experiment parameter is PDF 2.0, English text, Helvetica, and a simple unmerged table.

The semantic map is fixture-authored. The normative positive representation marks each column header as `/TH` with a unique `/ID`. Each data cell has a table attribute dictionary with `/O /Table` and `/Headers` containing the associated header ID. The damaged input is derived from that positive representation by changing only the `Headers` association on the second data-row amount cell: it points to the second-column header ID instead of the amount header ID. Its visual content, text, structure order, header scopes, and every other cell association remain unchanged. This directly tests the PDF table header association defined by the standard; it does not imply PDF/UA-2 conformance. See the logical structure and table standard-attribute definitions in [ISO 32000-2:2020](https://pdfa.org/resource/iso-32000-2/) and Adobe's [accessibility attribute reference](https://opensource.adobe.com/dc-acrobat-sdk-docs/library/accessibility/Access_DOM.html).

The explicit expected association is the target amount header to the amount cells in both data rows. The independent predicate checks each TD's `/Headers` IDs against the corresponding TH `/ID`, along with table/row/cell ancestry, target text, and order. A structure-type pass alone does not establish this predicate.

## Evidence and controls

For each family, retain the normative positive PDF and its derived damaged input as separate immutable files. The damaged input hash is the input identity for any later repair attempt. Independently record extracted text, rendered-page digest, and parsed structure/association observations. Keep exact displayed strings, including signs, decimals, and currency symbols; do not calculate, normalize, or infer totals.

Controls are defined before repair code:

- **Positive representation:** each TD's `/Headers` references its expected column TH `/ID`; the independent association predicate passes.
- **Association defect:** the derived damaged input's second-row amount TD references the second-column TH `/ID`; the independent predicate identifies exactly that misassociation while text and rendering remain identical to the positive.
- **Amount change:** changing any displayed financial string fails exact-string comparison, regardless of whether a machine validator accepts the PDF.
- **Malformed:** truncated or structurally unparsable PDF input is rejected before a repair candidate is emitted.
- **Unsupported:** merged, nested, multi-page, or otherwise out-of-scope table topology returns unsupported; no inferred repair is attempted.
- **Ambiguous map:** absent, multiply matched, or contradictory fixture-authored target associations require review-needed; no guessed association is accepted.

The existing veraPDF positive/finding oracle pair only tests ISO 14289-2:2024 clause 8.2.4 structure types ([control provenance](../../../internal/pdfua/testdata/README.md)). It authenticates the isolated validator/report path, not table-header semantics. No table-specific veraPDF outcome is claimed here. Any future table-specific validator result requires explicit report-only CI wiring for the exact fixture/candidate hash and an actual CI report. Validator acceptance and semantic correctness remain separate observations.

## Boundaries and ownership

Fixed language, font, values, and topology are experiment parameters, not production defaults, restrictions, or a general support promise. No arithmetic, tax, currency, date, rounding, privacy, retention, or reviewer-authority policy is selected. No writer or runtime dependency is selected. `dap-jo9.2.1.2` owns fixture bytes, hashes, baselines, and the pre-repair controls; `dap-jo9.1.2` owns attempted repair, candidate evidence, and finite disposition after exact candidate clearance in `dap-jo9.1.3`.
