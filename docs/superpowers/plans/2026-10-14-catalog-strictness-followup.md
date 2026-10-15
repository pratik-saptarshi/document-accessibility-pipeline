# Catalog Strictness Follow-up Implementation Plan

> **Current status (2026-10-15; `main` at `31109862b2c232a7827ca5da30868e8e8a15108d`):** This plan preserves its historical scope and deferred-work record. The Unicode-folding refinement, manifest size bound, and path-scoped fuzz-oracle work listed as deferred below were merged in PR #75 (`e4ea5ad30eae19c9baa9a4df0439a323d4156e3a`); they are no longer pending items from this plan. PR #76 implemented the `LoadCatalog` exact-schema, collision, and EOF checks described here (`internal/traceability/catalog.go:152-175,179-225`). PR #82 (`31109862b2c232a7827ca5da30868e8e8a15108d`) later fixed and tested the corpus loader's standalone noncanonical-key gap. These catalog behaviors are landed; the broader Task 1 remains open only for distinct direct `jsonstrict`, `author_v2`, and `internal/traceability.LoadIssues` coverage and compatibility questions, subject to current-source review. See the [current execution update](2026-10-15-round-two-plan-current-main-delta.md). Commit and publication examples below follow the repository's guarded publisher policy.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address the two catalog-strictness Codex P2 review comments left unaddressed on merged document-accessibility-pipeline PR #74, completing `dap-jo9.2.1.4`'s acceptance criterion ("`corpus.Load` and `LoadCatalog` reject duplicate/case-variant keys and unknown fields").

**Architecture:** Mirror the strict-decoding pattern already established in `internal/corpus/manifest.go` (`Load`: `rejectAmbiguousKeys` token scan before decoding, trailing-value check after) inside `internal/traceability/catalog.go`'s `LoadCatalog`, with catalog-appropriate error messages. Two tasks, one per finding, each TDD with its own commit, on a new branch `fix/catalog-strictness-followup` cut from `origin/main` in a fresh worktree.

**Tech Stack:** Go, `encoding/json` token scanning (same technique as `manifest.go`).

**Spec:**
- Codex P2 on PR #74 (`internal/traceability/catalog.go:155`): "Reject duplicate and case-variant catalog keys — The new `DisallowUnknownFields` call still accepts duplicate keys and case-insensitive field matches, so a catalog containing both `"status":"..."` and a later `"STATUS":"..."` loads successfully with the latter value... Reject folded key collisions before decoding, as is being done for manifests."
- Codex P2 on PR #74 (`internal/traceability/catalog.go:155`): "Reject trailing catalog JSON — When `catalog.json` contains a valid catalog followed by another JSON value, `decoder.Decode` accepts the first value and `LoadCatalog` returns success without examining the remainder. The previous `json.Unmarshal` rejected this input... perform a second decode and require `io.EOF` before returning."

Out of scope (deferred, matching the panel's final review): the Unicode-folding refinement, manifest memory-use concern, and fuzz-oracle decoy comment — all previously deferred minors, not catalog strictness.

## Global Constraints

- New branch `fix/catalog-strictness-followup` from `origin/main` (which now contains the PR #74 merge); fresh worktree via the `using-git-worktrees` skill — do not reuse the merged `panel-findings` worktree's branch.
- This plan file should be committed on the new branch as `docs(plans): catalog strictness follow-up plan`.
- Go toolchain: `export PATH="$HOME/sdk/go/bin:$HOME/go/bin:$PATH"`.
- Conventional Commits subjects. TDD: failing test first per task.
- `bd` resolves to the main checkout's Dolt database even inside worktrees — do not use `bd` for issue edits here; no issue edits are needed (`dap-jo9.2.1.4` stays open until exact-head CI is green on the follow-up PR).

## Review Focus

- Nested objects (e.g. entries in `capabilities[]`) with case-variant keys must be rejected — the token scan must recurse into arrays and nested objects exactly like `manifest.go`'s `checkValueKeys`.
- A key that case-folds onto a *struct field name* (`"Status"` vs `"status"`) must be rejected — `encoding/json` binds case-insensitively and takes the last duplicate silently.
- Trailing whitespace-only input must still load fine; trailing garbage already fails at first decode (unchanged behavior).
- Error messages must identify the file: existing style is `fmt.Errorf("decode catalog %s: %w", path, err)`; new messages follow it (`"decode catalog %s: duplicate or case-variant key %q (also present as %q)"`, `"decode catalog %s: multiple JSON values"`).

---

### Task 1: LoadCatalog rejects duplicate/case-variant keys

**Files:**
- Modify: `internal/traceability/catalog.go` (`LoadCatalog`)
- Test: `internal/traceability/catalog_test.go` (next to `TestLoadCatalogRejectsUnknownFields` at line 55)

**Interfaces:**
- Consumes: existing `Catalog` struct and `LoadCatalog(path string) (Catalog, error)` signature — unchanged.
- Produces: strict `LoadCatalog`; no new exported API. (Deliberately a catalog-local `rejectAmbiguousKeys` mirroring `internal/corpus/manifest.go`'s unexported helpers, not a new cross-package dependency — same folding semantics, `strings.ToLower`, catalog-worded errors.)

- [ ] **Step 1: Write the failing test** `TestLoadCatalogRejectsDuplicateCaseVariantKeys` in `internal/traceability/catalog_test.go`, following the existing test's temp-file pattern:

```go
func TestLoadCatalogRejectsDuplicateCaseVariantKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	// Minimal valid catalog shape (copy the fields TestLoadCatalogRejectsUnknownFields uses)
	// with both "status" and "STATUS" present.
	data := `{"acceptance_criteria":"x","status":"first","STATUS":"second"}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(path); err == nil {
		t.Fatalf("LoadCatalog accepted duplicate case-variant keys")
	} else if !strings.Contains(err.Error(), "duplicate or case-variant key") {
		t.Fatalf("error %q does not name the duplicate key", err)
	}
}
```

(Check what the existing test uses for a minimal valid catalog and match its shape; `filepath`, `os`, `strings` are already imported in the test file — verify.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/traceability/ -run TestLoadCatalogRejectsDuplicateCaseVariantKeys -v`
Expected: FAIL (`LoadCatalog accepted duplicate case-variant keys`).

- [ ] **Step 3: Implement the key scan** in `internal/traceability/catalog.go`: add an unexported `rejectAmbiguousKeys(data []byte, path string) error` plus the recursive token-walking helper, mirroring `internal/corpus/manifest.go`'s `rejectAmbiguousKeys`/`checkValueKeys` line-for-line except error messages use `"decode catalog %s: ..."` with the path. Call it in `LoadCatalog` right after `os.ReadFile`, before creating the decoder.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test -count=1 ./internal/traceability/ -v -run 'TestLoadCatalog'`
Expected: PASS (including the pre-existing unknown-fields test).
Then: `go vet ./internal/traceability/`, `gofmt -l internal/traceability/` (empty), `go test -count=1 ./internal/traceability/`.

- [ ] **Step 5: Commit**

```bash
git add internal/traceability/catalog.go internal/traceability/catalog_test.go
scripts/source-publisher.sh commit "fix(traceability): reject duplicate/case-variant keys in LoadCatalog"
```

---

### Task 2: LoadCatalog rejects trailing JSON values

**Files:**
- Modify: `internal/traceability/catalog.go` (`LoadCatalog`)
- Test: `internal/traceability/catalog_test.go`

**Interfaces:**
- Consumes: Task 1's strict `LoadCatalog` (this task adds the trailing check after it; implement sequentially on the same branch).
- Produces: fully strict `LoadCatalog`.

- [ ] **Step 1: Write the failing test** `TestLoadCatalogRejectsTrailingData`:

```go
func TestLoadCatalogRejectsTrailingData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	data := `{"acceptance_criteria":"x","status":"ok"}` + "\n{}\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(path); err == nil {
		t.Fatalf("LoadCatalog accepted trailing JSON value")
	} else if !strings.Contains(err.Error(), "multiple JSON values") {
		t.Fatalf("error %q does not name the trailing data", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/traceability/ -run TestLoadCatalogRejectsTrailingData -v`
Expected: FAIL (`LoadCatalog accepted trailing JSON value`).

- [ ] **Step 3: Implement the trailing check** in `LoadCatalog`, mirroring `manifest.go`'s `Load` exactly:

```go
var trailing any
if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
    if err == nil {
        return catalog, fmt.Errorf("decode catalog %s: multiple JSON values", path)
    }
    return catalog, fmt.Errorf("decode catalog %s trailing data: %w", path, err)
}
```

(Add `errors`/`io` imports if not present.)

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test -count=1 ./internal/traceability/ -v`
Expected: PASS (all catalog tests, including Task 1's).
Then: `go vet ./internal/traceability/`, `gofmt -l internal/traceability/` (empty). Also run the module-wide checks the repo's CI runs for this module if quick (`go test -count=1 ./internal/...` scoped to affected trees).

- [ ] **Step 5: Commit and push**

```bash
git add internal/traceability/catalog.go internal/traceability/catalog_test.go
scripts/source-publisher.sh commit "fix(traceability): reject trailing JSON values in LoadCatalog"
scripts/source-publisher.sh publish
```

After both tasks, open a PR against `main` titled `fix: catalog strictness follow-up (duplicate keys, trailing JSON)`, body referencing PR #74's unaddressed Codex comments and `dap-jo9.2.1.4`. Keep the worktree for PR feedback. Do NOT close `dap-jo9.2.1.4` — it closes only on exact-head CI green per its acceptance criteria.
