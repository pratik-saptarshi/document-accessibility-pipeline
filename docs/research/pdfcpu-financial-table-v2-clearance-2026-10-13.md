# Financial-table v2 experiment clearance extension

**Status: cleared for the positive no-op control only.** This extends the exact test-only pdfcpu v0.16.2 boundary in the [candidate/resource record](pdfcpu-financial-table-clearance-2026-10-13.md) to the separately versioned, metric-complete v2 fixtures. It does not change the v1 no-go, approve a production writer, or claim repair fidelity.

## Reviewed input identity

The reviewed source is PR #48 head `55848cc3309f25fa271b1491009fbddfdab6a2b1`, merged by squash at `d635ffb6122ecce47cbf452fe6f58b8e55906820`. Its v2 manifest SHA-256 is `b577a10f74ddbf1729e246313348d6f7a6f9fc1da786c40f9f2e576b804a5db7`. The exact fixture identities are:

| Family | Positive SHA-256 | Damaged SHA-256 |
|---|---|---|
| Statement | `4bc3d2b0b6e9d46f3f22d742c7c54b7e8f3f889f60ff78943e3f5a49fa41f2ec` | `73e47728a0cc5a017a0a114d77ed877198704a22cae7679157c923ec0abb3836` |
| Invoice | `ed4bbc12847abebb8c169d2338684ab64a9aad92dc5fb66239b49ae71c26cb72` | `d007433bf725a42edcda2173339ec80ab380dd2ba01f59afaae3ba5705ef70ad` |

The exact manifest also binds the complete Helvetica StandardEncoding metrics, AFM source digest and mapping rule, exact text/order, independent observations, and the single damaged `/Headers` relation. Fixture review accepted the immutable input identities and provenance. Independent product/gate and Go/source reviews accepted PR #48's exact head, which bounds v2 reads and ties parsing to the same capped bytes that were hashed. The earlier independent candidate/resource review is evidence head `25c4ca6a4160f76aae3c62c61b1e9cae4cc41f0e`; it clears only the pinned v0.16.2 API, reachable build/resource boundary, and explicit stateless/offline configuration recorded in the linked candidate record.

## Permitted operation and limits

This extension permits only separate positive no-op full-write controls for the statement and invoice v2 positives, using the pinned Go API and existing reviewed stateless/offline resource limits. The input cap is 64 KiB, context deadline 30 seconds, private staging mode 0600, and full-write output cap 256 KiB. Each output must be freshly read and independently compared to the immutable input for table structure, exact strings/order/associations, IDs and marked-content bindings, page font resources and full font dictionaries, metadata, extracted text, and rendered image. The v2 manifest and sources remain immutable.

Any protected delta is a bounded no-go. Do not normalize a fixture, accept a preservation waiver, or open damaged inputs after a failed positive control. Even if both no-op controls pass, this record alone does not publish a candidate or claim that repair succeeded. The damaged-input step remains separately gated by the recorded feasibility task and requires changing only its declared `/Headers` relation. No candidate-bound PDF/UA-2 report is represented here; that validation remains isolated CI-only and must exercise new candidate bytes.

## Merge verification

After PR #48 merged, exact merge SHA `d635ffb6122ecce47cbf452fe6f58b8e55906820` passed the Go checks (including race and both nested modules), security scans, veraPDF PDF/UA-2 integration, and CodeQL Go/Actions analysis. The isolated `validator-container` and `pdfua2` jobs both succeeded. These are repository checks for PR #48; they are not a result for any future no-op output or repair candidate.
