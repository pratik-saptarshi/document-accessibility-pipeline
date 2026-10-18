# PR #134 DevSecOps Parser-Fidelity Remediation Plan

**Status:** Independently reviewed and cleared for implementation on 2026-10-18. The branch is based on protected `main` at `7ba21bcf49edde780eccb0e98cdfe38fba0d8c42`.

## Goal and source

Address the parser-fidelity and factual-test-comment findings in the owner-posted [PR #134 overseer review](https://github.com/pratik-saptarshi/document-accessibility-pipeline/pull/134#issuecomment-6244773889). Preserve the #135 permission-resource spelling correction. This is test-policy tooling only: it does not change workflows, product behavior, issue state, or product acceptance.

## Acceptance criteria

1. `parseGoChecksWorkflow` preserves YAML key spelling. It rejects structural-key variants such as `ruN` and `PERMISSIONS` before policy validation; a valid lowercase workflow still parses and passes policy validation.
2. The strict parser rejects unmodeled fields at the workflow, job, and step levels. These regressions call the actual strict parser, not direct non-strict `yaml.Unmarshal`.
3. Permission-resource spelling remains case-sensitive: `CONTENTS` still reaches the policy validator and is rejected, preserving #135.
4. Same-input comparisons with pinned actionlint 1.7.12 cover exactly these four cases: lowercase baseline (repository parser and validator accept; actionlint exits 0), `run` → `ruN` (strict parser rejects; actionlint exits 1 because the step has neither `run` nor `uses`), `permissions` → `PERMISSIONS` (strict parser rejects; actionlint exits 1 for an unexpected workflow key), and `contents` → `CONTENTS` (parser accepts the case-sensitive permission value, policy validator rejects it; actionlint exits 1 for an unknown permission scope). No parser dependency is added.
5. Correct the historical and semantic comments identified in F2: current top-level `env` modeling, `continue-on-error` failure propagation versus job skipping, unsupported step-level `container`, `GOFLAGS=-run=^$` versus `go vet`, and the obsolete actionlint case-insensitivity attribution.
6. Existing lowercase security and execution controls remain enforced. Root and both nested-module checks required by `AGENTS.md`, traceability, and publisher regression checks pass on the final code tree.
7. An independent spec/quality review finds no blockers; publication uses the guarded HTTPS publisher and exact-head required CI is reported separately.

## Finding-to-test traceability

| Source finding | Observable check |
| --- | --- |
| F1: case-folding accepts invalid structural keys | New parser-level regression for `ruN` and `PERMISSIONS`; compare the same four fixtures (baseline, `ruN`, `PERMISSIONS`, `CONTENTS`) with actionlint 1.7.12 and record each exit status. |
| F1/#135 boundary | `CONTENTS` remains parseable as a permission value and is rejected by `validateScannerIsolation`; actionlint rejects it as an unknown permission scope; lowercase baseline is accepted by all three. |
| C1: strict unknown-field coverage absent | New table-driven regression passes unknown workflow-, job-, and step-level keys through `parseGoChecksWorkflow` and requires errors naming the key. |
| F2: factual comments misstate behavior | Review each named comment against the parser structs, validator branches, Go command behavior, and actionlint evidence; no claim of runtime execution for unsupported syntax. |

## Bounded task sequence

### 1. Add parser contract regressions first

- Add tests for structural variants and strict unknown fields at workflow, job, and step levels. Update the existing case-variant test so structural variants must fail in parsing, while `CONTENTS` continues to fail in policy validation.
- Run the focused tests before changing parser code. Expected RED: `ruN` and `PERMISSIONS` are accepted by the current parser.

### 2. Preserve key spelling in strict decoding

- Remove structural-key lowercasing from the strict Go-checks path. Keep `KnownFields(true)` and preserve the existing non-strict security-workflow decoding behavior.
- Run the new tests and affected existing policy tests. Confirm valid workflow keys still pass and permission values are not normalized.

### 3. Correct explanatory comments

- Update only comments/test names necessary to distinguish current guards from historical bypasses, execution from skipped jobs, unsupported step syntax from runtime behavior, and test filtering from vet behavior.
- Keep all existing rejection controls and workflow files unchanged.

### 4. Verify, review, and publish

- Compare the exact test inputs (lowercase baseline, `ruN`, `PERMISSIONS`, and `CONTENTS`) against pinned actionlint 1.7.12; record actual outputs and exit statuses.
- Run focused tests, then all applicable `AGENTS.md` checks. Run only actual mutations and report only observed mutation outcomes.
- Review the exact diff, scrub the public tree, request independent review pinned to exact base/head, and use `scripts/source-publisher.sh` for CloudBot Conventional commits and HTTPS publication. The owner explicitly authorized committing and pushing this validated branch in the task request. Do not create a PR or merge until coordinated with the owner.

## Verification evidence

Using the pinned actionlint 1.7.12 binary, the exact four inputs were generated from the same `go-checks.yml` baseline by changing only the indicated key. Results:

| Input | Repository parser/policy | actionlint 1.7.12 |
| --- | --- | --- |
| Lowercase baseline | Accepted | Exit 0 |
| `run` → `ruN` | Strict parser rejects the unknown field | Exit 1: step lacks `run` or `uses` |
| Workflow `permissions` → `PERMISSIONS` | Strict parser rejects the unknown workflow key | Exit 1: unexpected workflow key |
| Permission resource `contents` → `CONTENTS` | Strict parser accepts the modeled permission map; policy rejects the unsupported resource spelling | Exit 1: unknown permission scope |

The focused parser and devsecops package tests pass after the parser change. Full root, both nested-module, traceability, and publisher results are recorded in the implementation review/PR once the final exact branch head is established.

## Constraints and exclusions

- Base: `main` at the freshly verified SHA above; branch: `fix/devsecops-validator-parser-fidelity`; use an isolated worktree.
- Go 1.27.2 and actionlint 1.7.12 are the verified local tools; activate Go using `/workspace/.codex-tools/cloud-build-tools/activate.sh`.
- No workflow edits, new dependencies, tracker/Beads edits, generic governance implementation, product acceptance, issue closure, branch-protection changes, or changes to PR #132/#133.
- Do not claim an executed mutation, runtime behavior, or exact-head CI result unless its output is available for the exact revision.
