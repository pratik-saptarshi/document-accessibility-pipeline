# Decision 0020: First AFP source profile and trusted references

**Status: Accepted.** The product owner selected the first AFP input profile on
2026-10-21: the `Sample_1_health.zip` bundle from afpworld.com
(`01_Health_Coverage.afp`), converted AFP→PDF with the go-afp-line
repo/library. This record does not approve dependency adoption (that is
`dap-jo9.4.2`), invent a customer or delivery scope, or set tolerance
thresholds.

**Date: 2026-10-21. Owner: Pratik Saptarshi.**

## AFP source profile

- **File:** `01_Health_Coverage.afp` — 12,784 bytes,
  SHA-256 `553a336db9283ebbf85092320b68c060aacdbea4138fa720810f09431f86b73a`.
- **Source:** public sample bundle `Sample_1_health.zip`
  (https://www.afpworld.com/wp-content/uploads/Sample_1_health.zip).
- **Structure (measured with go-afp-line):** 25 structured fields, 22 distinct
  field types, **1 page**. Document is a COBRA health-coverage continuation
  letter — the **correspondence** family.
- **Fixture data:** the sample carries placeholder PII-shaped values
  ("John Doe", "4711 Any Street", "Any City, TX 12345"). These are synthetic
  sample placeholders, not real personal data. Handling constraint: fixtures
  are test inputs only; never present them as real customer documents or
  redistribute them as real PII.

## Engine

AFP→PDF conversion uses the **go-afp-line repo and library**
(https://github.com/pratik-saptarshi/go-afp-line): `afp/` codec,
`afptoline/` PTOCA reducer, `linetopdf/` pdfcpu rendering. No engine
alternative is selected; no other AFP library is approved for this track.

## PDF target

**PDF 1.7 visual-fidelity output** for this first profile. PDF/UA-2 is the
pipeline's accepted target per Decision 0001, but UA tagging of AFP-sourced
PDFs is out of scope for the first profile — go-afp-line renders visual PDFs
and performs no UA tagging. Tagging remains an open follow-up.

## Trusted references

| Reference | Identity (SHA-256) | Checks |
|-----------|--------------------|----|
| `01_Health_Coverage.html` | `95dd15248ddb72e4f994073ddccd1c6991219cb28f9c05957f8a28610d9aefdd` | **Text:** extract text from the generated PDF, normalize whitespace, compare against the HTML text content |
| `01_Health_Coverage.pdf` | `51b8b17a229a690fff1ac252a5756e0b0f05958210920cfca1d4dd4431d2e7ed` | **Page order:** page count must be 1, in sequence (trivially single-page). **Geometry:** visual/bounding-box comparison of the generated PDF against the reference PDF |

## Semantic features and boundaries

- **In scope:** single-page PTOCA text presentation; the correspondence
  letter's text content, reading order, and page geometry as observable in the
  references.
- **Unsupported / error boundaries (open, for the feasibility spike):** field
  types outside go-afp-line's contracted subset; multi-page documents;
  non-PTOCA content (images, barcodes, vector graphics); code pages beyond
  037/500. The spike must enumerate what the sample actually exercises.

## Authorized reference acceptor

Pratik Saptarshi (product owner). *Proposed — to be confirmed.*

## Explicitly unresolved

- Text/geometry comparison tolerance thresholds (exact vs fuzzy, and how much).
- Whether the HTML or the PDF is authoritative when they disagree.
- go-afp-line revision pin (belongs to `dap-jo9.4.2`, not this decision).
- PDF/UA-2 tagging follow-up scope and owner.
