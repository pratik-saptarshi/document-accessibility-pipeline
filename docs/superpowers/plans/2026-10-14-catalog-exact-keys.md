# LoadCatalog Exact-Spelling Key Validation Implementation Plan

> **Current status (2026-10-15; `main` at `31109862b2c232a7827ca5da30868e8e8a15108d`):** This is a historical implementation plan. Its described `LoadCatalog` exact-spelling behavior, together with collision and EOF checks, landed in PR #76 (`internal/traceability/catalog.go:152-175,179-225`); the earlier text below records the plan as written before implementation. PR #82 (`31109862b2c232a7827ca5da30868e8e8a15108d`) subsequently closed the separate corpus loader's standalone noncanonical-key gap with regression tests. Remaining Task 1 work is limited to distinct direct `jsonstrict`, `author_v2`, and `internal/traceability.LoadIssues` coverage and compatibility questions, subject to current-source review. See the [current execution update](2026-10-15-round-two-plan-current-main-delta.md); do not infer that broader Beads acceptance is complete.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the Codex P2 on dap PR #76 ("Reject standalone case-variant schema keys") by making `LoadCatalog` reject any JSON object key that does not exactly match a schema field spelling at its struct level.

**Architecture:** Replace the folding-collision scan (`rejectAmbiguousKeys`/`checkValueKeys`) in `internal/traceability/catalog.go` with a single schema-aware token walk. `encoding/json` binds keys case-insensitively, so the current scan misses a standalone `"STATUS"` (no collision to find) which then populates the `status` field despite `DisallowUnknownFields`. The new walk resolves each JSON object against its expected struct type via reflection, and rejects (a) exact duplicate keys and (b) keys with no exact tag spelling. This subsumes the old scan: a folding collision like `"status"`+`"STATUS"` is caught by rule (b). Keep `DisallowUnknownFields` on the decoder as a backstop.

**Tech Stack:** Go, `encoding/json` token scanning + `reflect` for tag sets (same module, no new dependencies).

**Spec:** Codex P2 on PR #76 (`internal/traceability/catalog.go:206`): 'When a catalog contains only a case-variant key such as `"STATUS":"ok"`, `seen` has no prior folded key, so this scan succeeds and `encoding/json` then binds `STATUS` to the `status` field case-insensitively despite `DisallowUnknownFields`. Consequently, `LoadCatalog` still accepts noncanonical schema keys; validate encountered keys against their exact schema spellings at every nested struct level and cover a standalone variant as a hostile-input regression.' Plus `dap-jo9.2.1.4`'s acceptance criterion: "`corpus.Load` and `LoadCatalog` reject duplicate/case-variant keys and unknown fields".

Out of scope (deferred, as before): the identical standalone-variant gap in `internal/corpus/manifest.go`'s `Load` — note it in the commit message as a known residual.

## Global Constraints

- Work in the existing `catalog-strictness` feature worktree on branch `fix/catalog-strictness-followup` (PR #76's branch). Do not create a new branch.
- Commit this plan file first on the branch: `docs(plans): exact-spelling key validation plan`.
- Go toolchain: `export PATH="$HOME/sdk/go/bin:$HOME/go/bin:$PATH"`.
- Conventional Commits subjects. TDD: failing tests first.
- Do not touch Beads items; do not close `dap-jo9.2.1.4` (closes only on exact-head CI green).

## Review Focus

- Exact duplicate keys (`"status"` twice) must still be rejected — the old scan caught these; the new walk must too (rule (a)).
- Nested standalone variants (`capabilities[0]."ID"`) must be rejected at their struct level (rule (b) recursing through `[]Capability` etc.).
- `omitempty` tags (`group`, `standalone_reason`, `case_roles`) must parse to their base spellings.
- Arrays of scalars (`requirements[]string` inside `Story`, `open_decisions[]string`) must be consumed without misinterpreting them as struct levels.
- Unknown keys must still be rejected (now by the walk; `DisallowUnknownFields` stays as backstop).
- Valid catalogs must still load — existing tests plus `go run ./cmd/traceability-check` must stay green.
- The manifest.go residual: flag in commit message, do not fix here.

---

### Task 1: Schema-aware exact-spelling key validation in LoadCatalog

**Files:**
- Modify: `internal/traceability/catalog.go` (replace `rejectAmbiguousKeys`/`checkValueKeys` with the schema walk; update the `LoadCatalog` call site)
- Test: `internal/traceability/catalog_test.go` (new tests next to the Task-1/Task-2 tests from PR #76)

**Interfaces:**
- Consumes: `Catalog` and nested struct types with their `json` tags (all simple `json:"name"` / `json:"name,omitempty"`; no embedded structs); existing `LoadCatalog(path string) (Catalog, error)` signature — unchanged.
- Produces: strict `LoadCatalog`; no new exported API.

- [ ] **Step 1: Write the failing tests** in `internal/traceability/catalog_test.go`:

```go
func TestLoadCatalogRejectsStandaloneCaseVariantKey(t *testing.T) {
	// Minimal valid catalog shape (copy from TestLoadCatalogRejectsUnknownFields),
	// with "status" replaced by the noncanonical "STATUS".
	...
	if _, err := LoadCatalog(path); err == nil {
		t.Fatalf("LoadCatalog accepted standalone case-variant key")
	} else if !strings.Contains(err.Error(), `"STATUS"`) {
		t.Fatalf("error %q does not name the offending key", err)
	}
}

func TestLoadCatalogRejectsNestedCaseVariantKey(t *testing.T) {
	// Same shape, but capabilities[0] uses "ID" instead of "id".
	...
	if _, err := LoadCatalog(path); err == nil {
		t.Fatalf("LoadCatalog accepted nested case-variant key")
	} else if !strings.Contains(err.Error(), `"ID"`) {
		t.Fatalf("error %q does not name the offending key", err)
	}
}

func TestLoadCatalogStillRejectsExactDuplicateKeys(t *testing.T) {
	// "status" twice — regression guard for the replaced folding scan.
	...
	if _, err := LoadCatalog(path); err == nil {
		t.Fatalf("LoadCatalog accepted exact duplicate keys")
	}
}
```

(Check the existing tests' minimal-valid-catalog shape and temp-file pattern; reuse it.)

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/traceability/ -run 'TestLoadCatalogRejectsStandaloneCaseVariantKey|TestLoadCatalogRejectsNestedCaseVariantKey' -v`
Expected: FAIL (both accepted). The exact-duplicate test should PASS already (old scan catches it) — it is a regression guard.

- [ ] **Step 3: Implement the schema walk** in `internal/traceability/catalog.go`, replacing `rejectAmbiguousKeys`/`checkValueKeys`:

```go
// rejectNoncanonicalKeys rejects JSON object keys that are exact duplicates
// or that do not exactly match a schema field spelling at their struct level.
// encoding/json binds keys case-insensitively, so without this check a
// standalone "STATUS" would populate the "status" field despite
// DisallowUnknownFields.
func rejectNoncanonicalKeys(data []byte, path string) error {
	return checkSchemaKeys(json.NewDecoder(bytes.NewReader(data)), reflect.TypeOf(Catalog{}), path)
}
```

`checkSchemaKeys(dec *json.Decoder, typ reflect.Type, path string) error`:
- Read one token. If it is not a `{` delimiter, return nil (the subsequent `Decode` produces the real error).
- Deref `typ` through pointers. If it is a struct: build the exact tag set once per call — for each field, take the `json` tag, split on `,`, keep the name part; skip fields tagged `json:"-"`. (No embedded structs exist in the catalog schema; if one appears, its fields are not promoted — acceptable, document in a comment.)
- For each key in the object (`dec.More()` loop): if the key is already in `seen`, return `fmt.Errorf("decode catalog %s: duplicate key %q", path, key)`; if it is not in the exact tag set, return `fmt.Errorf("decode catalog %s: key %q does not exactly match a schema field", path, key)`. Record in `seen`.
- For the value: resolve the field by exact tag, deref its type, and recurse with `checkSchemaValue`.
- Consume the closing `}` token.

`checkSchemaValue(dec *json.Decoder, typ reflect.Type, path string) error`:
- Peek the next token via `dec.Token()`. If `{`: if derefed `typ` is a struct, recurse into `checkSchemaKeys` (re-feed the already-read token — simplest: unread it by handling the `{` inside `checkSchemaKeys`... to avoid token pushback complexity, structure as: `checkValue` reads the token; on `{` with struct typ → validate keys inline; on `{` with non-struct typ → generic skip). If `[`: deref `typ` to slice/array elem type (if not a slice/array, use `any`), walk each element with `checkValue`. Else scalar: nothing to do.
- Generic skip for schema-less values: recursive token consumption until the value is complete (mirror the old `checkValueKeys` recursion shape).

Keep it simple and total: every token the decoder produces is consumed exactly once; unknown shapes are skipped, never misread. Add `reflect` to imports; `strings` stays (tag splitting).

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test -count=1 ./internal/traceability/ -v -run 'TestLoadCatalog'`
Expected: PASS — all catalog tests, old and new.
Then: `go vet ./internal/traceability/`, `gofmt -l internal/traceability/` (empty), `go test -count=1 ./internal/...`, and `go run ./cmd/traceability-check` (must stay green on the real catalog).

- [ ] **Step 5: Commit and publish** (branch `fix/catalog-strictness-followup`; repo rule: publish via `scripts/source-publisher.sh`, never raw `git push`)

```bash
git add internal/traceability/catalog.go internal/traceability/catalog_test.go
scripts/source-publisher.sh commit "fix(traceability): reject noncanonical key spellings in LoadCatalog

The folding-collision scan missed standalone case-variant keys: a lone
\"STATUS\" has no collision yet encoding/json binds it to the status field
case-insensitively. LoadCatalog now validates every key against the exact
schema spelling at its struct level (subsuming the old scan, which also
caught exact duplicates — still covered).

Known residual (deferred): internal/corpus Load has the same standalone-
variant gap; unchanged by this commit."
scripts/source-publisher.sh publish
```
