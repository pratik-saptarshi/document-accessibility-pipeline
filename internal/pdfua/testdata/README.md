# veraPDF test data

`contract-*.xml` are hand-authored minimal XML samples for parser unit tests. Their `buildInformation/releaseDetails` entries model the expected direct `apps` component provenance; their component versions and report bodies are synthetic. They are not veraPDF output and do not demonstrate document conformance.

`known-pass/report.xml` and `known-finding/report.xml` are unmodified pinned-veraPDF 1.30.3 reports for the corpus controls below. They were extracted from the `pdfua2-reports` artifact (ID `12145709034`, run `39153381973`, source head `5ba4ace4b133a34c35f0242fa021d4f14ff514e8`, archive SHA-256 `a4ee6a184f72d5ed0284dd809c718cb862ae1215025be30066663b645f39cf15`).

| File | SHA-256 | Report result |
| --- | --- | --- |
| `known-pass/report.xml` | `8f43db6fca3e7cab730ebbf7d83ae3b3574eff135e81ae90cb18bdc99c7bdc17` | compliant |
| `known-finding/report.xml` | `60b0f2cfeef133656684c4eb079c2d720b1332893d90201546f68b8723a425c3` | noncompliant with one finding |

`oracle-ua2-pass.pdf` and `oracle-ua2-fail.pdf` are unchanged PDF/UA-2 corpus controls from the official [`veraPDF/veraPDF-corpus`](https://github.com/veraPDF/veraPDF-corpus) repository, commit `bb75f4f0073d9350dfd058c0162a367e6fadf25e`, licensed CC BY 4.0. The pair exercises ISO 14289-2 clause 8.2.4 (structure types). The 1.30.3 Greenfield CLI was run with `--flavour ua2 --format xml` and reported the pass control compliant and the finding control noncompliant.

| File | Upstream corpus path | SHA-256 |
| --- | --- | --- |
| `oracle-ua2-pass.pdf` | `PDF_UA-2/8.2 Logical structure/8.2.4 Structure types/8.2.4-t01-pass-a.pdf` | `c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b` |
| `oracle-ua2-fail.pdf` | `PDF_UA-2/8.2 Logical structure/8.2.4 Structure types/8.2.4-t01-fail-a.pdf` | `2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0` |

These are machine-validator oracle controls only. A machine pass does not establish semantic accessibility, reading order, or assistive-technology behavior; human review remains a separate release gate.
