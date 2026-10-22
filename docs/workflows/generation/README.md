# Generation workflow

The proposed input contract is documented in [generation profiles](profiles/profile.schema.json), with [synthetic family examples](../../profiles/examples/generation/). Its shared family enum follows the owner-approved set in [Decision 0017](../../decisions/0017-approved-profile-family-set.md): statements, invoices, correspondence, reports, presentations, forms, marketing, and HR documents (`hr_documents`). The enum does not claim end-to-end support; synthetic examples now exist for all eight families, and acceptance criteria per family remain open. PDF/UA-2 is the accepted output validation target (see [Decision 0001](../../decisions/0001-pdf-ua-2-validation-target.md)). Language, source fields, confidentiality, retention, semantic policy, and human-review expectations remain open.

**Status: Proposed workflow. PDF engine selection is open.**

1. Validate input data against an agreed profile for one of the eight owner-approved family identifiers.
2. Resolve a template and semantic structure, including language, headings, lists, tables, links, and meaningful image descriptions as applicable.
3. Accept DOCX or Markdown sources where their feasibility and conversion behavior have been verified.
4. Render through a selected generation engine. The engine must be chosen through the comparative spike; do not presume it is the remediation writer.
5. Validate data integrity, visible output, structure, metadata, and accessibility with independent machine checks and human semantic/assistive-technology review.
6. Record source-to-output traceability, warnings, and review evidence.

## Test-first slices

Start with synthetic golden cases and failing tests for each family. Cover data correctness, structure, visible fidelity, accessibility semantics, invalid inputs, and failure behavior. Make each narrow slice green before refactoring. Tagged output is only one artifact and does not establish PDF/UA or WCAG compliance.

## Open questions

- Which PDF engine, source converters, and template model pass the evidence gate?
- Are DOCX and Markdown inputs, authoring formats, or both?
- Which profile, language, output, and review requirements apply to each family?
- Which ambiguities require approval from a content owner?
