# Legacy requirement crosswalk — 2026-10-13

The six affected Beads stories cited eleven unique `REQ-*` identifiers that are not current catalog IDs. Their original standalone definitions were not found in the tracked requirements catalog. This crosswalk records **proposed continuity** from the current story prose to the canonical requirements that now appear in the catalog and Beads story acceptance text. These mappings are **not proven semantic equivalence** and do not add requirements beyond the current catalog.

| Legacy story reference | Beads story | Proposed current requirement targets | Basis and limit |
|---|---|---|---|
| `REQ-REM-INGEST-001` | `dap-jo9.2.3` | `REQ-REM-001`, `REQ-REM-004` | Existing-PDF intake/provenance and the story's supported-input boundary; does not decide production authority or eligible topology. |
| `REQ-REM-FID-001` | `dap-jo9.2.3` | `REQ-REM-003` | Preservation/fidelity language; no quantitative threshold is inferred. |
| `REQ-SCOPE-001` | `dap-jo9.2.3`; `dap-jo9.3.3` | `REQ-REM-002`; `REQ-GEN-002` | The stories describe distinct remediation scope and generation data scope. This does not unify their profiles or approve arithmetic policy. |
| `REQ-REM-SEM-001` | `dap-jo9.2.4` | `REQ-REM-005` | Ambiguous semantic repair and review routing in the story prose. |
| `REQ-VAL-HUMAN-001` | `dap-jo9.2.4`; `dap-jo9.3.4` | `REQ-REM-006`; `REQ-GEN-005`, `REQ-GEN-006` | Human/AT review boundaries in the two lanes. Reviewer identity, authority, protocol, and sign-off remain unresolved. |
| `REQ-GEN-DATA-001` | `dap-jo9.3.3` | `REQ-GEN-001` | Statement/invoice generation from profile-validated data. |
| `REQ-GEN-OUTPUT-001` | `dap-jo9.3.3`; `dap-jo9.3.4` | `REQ-GEN-003`; `REQ-GEN-005`, `REQ-GEN-006` | Generation output and independent review outcomes, preserving lane-specific scope. |
| `REQ-GEN-SOURCE-001` | `dap-jo9.3.4` | `REQ-GEN-004` | Source-format feasibility remains a separate gate. |
| `REQ-CROSS-SEC-001` | `dap-jo9.5` | `REQ-X-002`; bounded/private behavior in `REQ-X-001` | Security and bounded execution context. |
| `REQ-CROSS-AUDIT-001` | `dap-jo9.5` | `REQ-X-003`; identity/status aspects of `REQ-X-001` | Audit evidence and bounded job identity/status. |
| `REQ-AFP-FEAS-001` | `dap-jo9.4.1` | `REQ-AFP-001` | The story remains a feasibility decision only, not conversion authorization. |

The active story records now cite only their canonical catalog requirement IDs. The table above preserves the older identifiers as historical context; it is not a synonym registry, an implementation authorization, or evidence that any acceptance case has passed.
