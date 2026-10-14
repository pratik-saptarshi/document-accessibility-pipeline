# Traceability Checker Scaling Threshold — Issue Reframe Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reframe beads issue dap-jo9.14 per the adversarial panel's verdict so its goal, acceptance criteria, and evidence are satisfiable, verifiable, and pick-up-able cold.

**Architecture:** Docs-only change set: rewrite the issue's Description/Goal/AC/Notes via `bd edit` in three bounded edits (evidence promotion, goal+AC reframe, cold-pickup notes), then verify with `bd show` and the repo's checks. No product code changes; no indexing work is authorized.

**Tech Stack:** Beads CLI (`bd`), Markdown.

**Spec:** beads issue dap-jo9.14 (run `bd show dap-jo9.14`); panel findings in `docs/reviews/panel-dap-jo9.14-2026-10-13.md` — the plan argues from the panel report, so both travel with it; executors read both.

## Global Constraints

- Work stays in the document-accessibility-pipeline repo; track changes there.
- Run `bd prime`, `bd ready`, and `bd show dap-jo9.14` before editing the issue (repo AGENTS.md).
- Do not invent schedules, budgets, measured performance, or approved architecture (repo AGENTS.md).
- Mark unresolved choices **Proposed** or **Open**.
- Do not implement product features (no checker/indexing code changes) under this planning request.
- Use Conventional Commit subjects; do not push without authorization.

## Review Focus

- A future reader must be able to state the threshold (or its explicit absence), the growth assumption, the owner, and the review date from the issue alone — if any of those is missing after the edits, the reframe failed.
- The Description must lead with the controlled 2026-10-13 rerun, never the superseded 2026-10-12 numbers.
- "Near-quadratic" must appear only with its interval-qualified exponents, or not at all.

---

### Task 1: Promote controlled evidence; unify terminology

**Files:**
- Modify: beads issue dap-jo9.14 (Description/Findings via `bd edit`)

**Interfaces:**
- Consumes: panel report §P2 "description is stale", §P2 "terminology drift", §P2 "interval qualifier"
- Produces: canonical issue text that Tasks 2–3 edit further

- [ ] **Step 1: Read the current issue and panel report**

Run: `bd prime && bd ready && bd show dap-jo9.14`
Expected: issue text matches the panel's quoted evidence; note the exact Description/Notes/AC section boundaries for editing.

- [ ] **Step 2: Rewrite the Description Findings section**

Replace the 2026-10-12 exploratory paragraph with the controlled 2026-10-13 rerun as primary:
current catalog (17 requirements, 6 stories, 21 tasks, 26 cases) medians 0.526 ms
Validate / 0.107 ms GeneratedMarkdown; synthetic one-to-one medians at
100/1,000/5,000 entities 0.557/19.896/441.795 ms Validate and
0.464/21.196/495.460 ms GeneratedMarkdown. Demote the 2026-10-12 numbers to a
"Superseded" subsection with the existing disqualification note. State the fitted
exponents per interval (Validate ~1.55/1.93, Markdown ~1.66/1.96 for 100→1,000 /
1,000→5,000) instead of the bare "near-quadratic" label.

- [ ] **Step 3: Unify terminology across the issue**

Use exactly one term per concept everywhere in the issue: "traceability checker"
for the subject artifact; "diagnostics" and "generated Markdown views" for its
outputs. Replace "repository planning checker" and "parser/checker".

- [ ] **Step 4: Verify the edit**

Run: `bd show dap-jo9.14`
Expected: Description leads with 2026-10-13 controlled numbers; "near-quadratic"
appears only with interval exponents; no occurrence of "repository planning checker"
or "parser/checker" remains (check with `bd show dap-jo9.14 | grep -i "planning checker\|parser/checker"` returning nothing).

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "docs(beads): promote controlled benchmark evidence in dap-jo9.14"
```

### Task 2: Reframe the goal and rewrite the acceptance criteria

**Files:**
- Modify: beads issue dap-jo9.14 (Goal + Acceptance Criteria via `bd edit`)

**Interfaces:**
- Consumes: Task 1's canonical text; panel report §P1 (merged AC defect), §P2 (AC over-constraint)
- Produces: a satisfiable, falsifiable goal + AC

- [ ] **Step 1: Rewrite the Goal**

New goal: "Retain verified scaling evidence for the traceability checker pending
product triage; record an explicit threshold decision (value or 'none adopted'),
the growth assumption behind it, the owner, and a review date." Mark the choice
between adopting a threshold vs. monitor-mode as **Proposed** for triage.

- [ ] **Step 2: Rewrite the Acceptance Criteria as a decision record**

Replace the current AC with:
1. A recorded threshold decision: either a (size, runtime-budget) pair with the
   growth assumption and review date that justify it, or an explicit "no threshold
   adopted — monitor mode" statement.
2. A named owner and a trigger condition for re-checking (e.g. catalog exceeds N
   entities or Validate exceeds X ms), plus a review-by date.
3. If indexing is ever proposed: unchanged diagnostics and no *unintended* change
   to generated views (byte-identical against the pre-indexing baseline, which must
   be named), plus repeated benchmarks showing flatter growth — define "materially"
   as exponent < 1.5 on the 1k→5k interval — with no allocation regression against
   the cited baseline (5,000-entity: 8,452,576 B/op, 40,247 allocs/op per
   `docs/research/support/traceability-scaling-benchmark-2026-10-13.txt`).

- [ ] **Step 3: Resolve the "100 tasks and 10 ms" heuristics**

Either define them (100 tasks of what entity mix; 10 ms of Validate, Markdown, or
combined; median or p99; the tasks↔entities conversion used) or record them as
**Proposed** review points explicitly not adopted, noting the panel's finding that
as a pair they are mutually inconsistent (10 ms fires around ~500+ entities,
making "100 tasks" nearly irrelevant).

- [ ] **Step 4: Verify the edit**

Run: `bd show dap-jo9.14`
Expected: AC lists the decision-record items; each item is checkable (a future
reader can answer yes/no); "materially" and the allocation baseline have numbers;
the heuristics are either defined or explicitly marked Proposed/not-adopted.

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "docs(beads): reframe dap-jo9.14 goal and acceptance criteria per panel"
```

### Task 3: Add cold-pickup notes and sunset

**Files:**
- Modify: beads issue dap-jo9.14 (Notes via `bd edit`)

**Interfaces:**
- Consumes: Task 2's text; panel report §P2 (cold-pickup gaps), §P3 (sunset)
- Produces: an issue a stranger can act on

- [ ] **Step 1: Add the rerun path**

Record the harness location and exact rerun command for the benchmark
(`docs/research/support/traceability-scaling-benchmark-2026-10-13.txt` is the raw
output; cite the benchmark target/command that produced it), the Go toolchain and
runner provenance already in the notes, and the N=1 caveat: the 5,000-entity
points are single-iteration measurements, so the headline worst-case number has
no variance estimate.

- [ ] **Step 2: Add the sunset**

Record a review-by date: if the catalog has not approached the trigger by that
date, the issue is re-confirmed or closed. Name the triage venue that owns the
"100 tasks / 10 ms" decision.

- [ ] **Step 3: Verify the edit**

Run: `bd show dap-jo9.14`
Expected: a reader can locate the harness, the rerun command, the raw data file,
and the review-by date; the N=1 caveat appears next to the 5,000-entity numbers.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "docs(beads): add cold-pickup notes and review cadence to dap-jo9.14"
```

- [ ] **Step 5: Run repo checks and export beads**

Run: `test -z "$(gofmt -l .)" && go vet ./... && go run ./cmd/traceability-check` and
export/review `.beads/issues.jsonl` per AGENTS.md.
Expected: checks pass; the export reflects the three edits.

## Self-Review

1. **Spec coverage:** panel P1 (goal/AC reframe) → Task 2; P2 evidence promotion → Task 1;
   P2 terminology → Task 1; P2 exponents → Task 1; P2 cold-pickup → Task 3; P2
   over-constraint → Task 2; P3 sunset → Task 3; P3 jargon/hash-pinning → Task 1/3
   notes edits. The panel's "verified sound" items need no task.
2. **Step scan:** each step names one action with a checkable result; `bd show`
   verifications are concrete.
3. **Type consistency:** N/A (docs-only; no code interfaces).
4. **Review Focus:** the three focus lines map to Task 2 (decision record
   completeness), Task 1 (description currency), Task 1 (exponent qualification).
5. **Proportion:** plan is shorter than the panel report it implements; no code
   bodies transcribed.
