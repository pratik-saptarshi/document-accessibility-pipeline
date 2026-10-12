# veraPDF test data

`contract-*.xml` are hand-authored minimal XML samples for parser unit tests. They are not veraPDF output and do not demonstrate document conformance.

`oracle-ua2-pass.pdf` and `oracle-ua2-fail.pdf` are unchanged PDF/UA-2 corpus controls from the official [`veraPDF/veraPDF-corpus`](https://github.com/veraPDF/veraPDF-corpus) repository, commit `bb75f4f0073d9350dfd058c0162a367e6fadf25e`, licensed CC BY 4.0. The pair exercises ISO 14289-2 clause 8.2.4 (structure types). The 1.30.3 Greenfield CLI was run with `--flavour ua2 --format xml` and reported the pass control compliant and the finding control noncompliant.

| File | Upstream corpus path | SHA-256 |
| --- | --- | --- |
| `oracle-ua2-pass.pdf` | `PDF_UA-2/8.2 Logical structure/8.2.4 Structure types/8.2.4-t01-pass-a.pdf` | `c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b` |
| `oracle-ua2-fail.pdf` | `PDF_UA-2/8.2 Logical structure/8.2.4 Structure types/8.2.4-t01-fail-a.pdf` | `2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0` |

These are machine-validator oracle controls only. A machine pass does not establish semantic accessibility, reading order, or assistive-technology behavior; human review remains a separate release gate.
