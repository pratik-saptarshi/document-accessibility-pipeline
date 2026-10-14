# Review Panel Report — dap-jo9.14 "Track traceability checker scaling threshold"

- **Date:** 2026-10-13
- **Mode:** Exhaustive (documentation review)
- **Reviewers:** Clarity Editor, Technical Accuracy Reviewer, Completeness Checker, Devil's Advocate
- **Process:** independent review → one debate round → judge synthesis
- **Subject:** beads issue dap-jo9.14 [P3, OPEN, type: spike]

## Verdict

**Keep open, but reframe.** The panel converged 3–1 against closing the issue outright
(Devil's Advocate conceded "close" in favor of "keep open but reframe the goal").
The benchmark evidence is honest and verified — every median reconciles with
`docs/research/support/traceability-scaling-benchmark-2026-10-13.txt`, the provenance
block matches the harness output, and the cited commit resolves. The defects are in
the issue *text*, not the data: the acceptance criteria promise a deliverable
(a recorded threshold) that has been explicitly punted to product triage twice,
with no owner, trigger, or review date. An issue whose AC can be neither satisfied
nor falsified stays open forever by construction.

Panel score: **5/10** (Technical Accuracy scored the evidence 7/10; the issue text
scores 4–5 across the other three reviewers).

## P1 — Merged: the acceptance criteria are the load-bearing defect

All four reviewers independently hit the AC from different angles; the debate merged
them into one finding:

1. The goal says "set a realistic catalog-size/runtime threshold" and the AC demands
   "record a size/runtime threshold based on realistic repository growth" — but the
   notes disclaim the only candidates: "100 tasks and 10 ms are heuristic proposed
   review points only, not measured growth forecasts or adopted policy."
2. No artifact is named (a number pair? a policy doc? an alert?), no owner, no trigger
   condition, no review-by date. "Pending separate product triage and independent
   review" names no venue, assignee, or SLA.
3. The AC's second clause is purely conditional ("If indexing is proposed…") with no
   actor or timing.

**Recommendation:** rewrite the goal from "set a threshold" to "retain verified scaling
evidence pending product triage," and rewrite the AC as a decision record:
threshold value (or explicit "none adopted"), the growth assumption behind it,
the owner, the trigger condition, and a review/expiry date.

## P2 findings

- **"Near-quadratic" needs its interval qualifier.** Fitted exponents from the
  reported medians: ~1.55/1.66 (Validate/Markdown) for 100→1,000 entities,
  ~1.93/1.96 for 1,000→5,000, ~1.71/1.78 overall. The claim is solid above ~1k
  entities and loose below it; state the per-interval fits instead of the label.
- **The description is stale.** The Description's Findings still present the
  superseded 2026-10-12 exploratory numbers as primary; the controlled 2026-10-13
  rerun lives in Notes. Invert: controlled results in Description, 10-12 data as history.
- **Terminology drift.** Three names for the subject artifact ("traceability checker",
  "repository planning checker", "parser/checker") and two for its outputs ("outputs"
  vs "diagnostics" / "generated views"). Pick one term per concept.
- **Unit mismatch.** Benchmarks scale in "entities"; the heuristic review point is in
  "tasks" and "10 ms" of an unspecified operation. No tasks↔entities conversion exists,
  so the heuristics can't be plotted against the benchmark axis. Worse, the two
  heuristics are mutually inconsistent: at 100 entities Validate is ~0.557 ms, an order
  of magnitude below the 10 ms trigger, which fires around ~500+ entities — the "100
  tasks" prong is nearly irrelevant as a pair.
- **Cold-pickup gaps.** No harness path or rerun command is cited; the AC demands "no
  allocation regression" but no allocation baseline is recorded (the raw file contains
  B/op and allocs/op — e.g. 8,452,576 B/op, 40,247 allocs/op at 5,000 entities — but the
  issue never cites them); "byte-identical generated views" names no reference views.
- **Inverted evidence pyramid.** The most decision-relevant point — the 5,000-entity
  median (~440 ms) that would motivate any indexing work — is the least replicated
  (single iteration per repetition); the caveat is undisclosed in the issue.
- **Synthetic linkage ungrounded.** "One-to-one" synthetic fixtures are never defined
  against real-catalog linkage density (depends-on, blocks, children, hubs); a threshold
  extrapolated from them may misfire if real growth is sparse or hub-driven. The
  measured exponent itself varies 1.55→1.96 across intervals — any extrapolated
  threshold is regime-dependent, and the issue never says so.
- **AC over-constraint.** "Byte-identical generated views" conflates no-regression with
  no-change — it would veto legitimate future output improvements; scope to "no
  *unintended* output change." "Materially flatter growth" is undefined (2×? exponent
  < 1.5?). "Unchanged diagnostics" is unscoped against future checker evolution.

## P3 findings

- Slash-delimited number triples force re-derivation; lead with results, then provenance.
- Opaque jargon: "concurrency suffix", "provenance edit", "the refreshed harness",
  "one-to-one catalogs", "allocation regression" — define on first use.
- Cited commit hash has no repository qualifier; current-catalog snapshot
  (17 requirements, 6 stories, 21 tasks, 26 cases) has no date/version pin.
- No sunset or review cadence if the catalog never grows.

## Verified sound (no finding)

- All reported medians reconcile with the raw benchmark file; provenance block matches;
  cited source commit exists; superseded-data handling in the notes is correct
  (restore + explicit disqualification, not silent overwrite).
- "Scaling signal, not a current performance incident" is supported: live catalog runs
  in ~0.5 ms; the superlinear regime only appears 20–300× above current catalog size.
- The "-cpu=1 vs NumCPU=5" note is internally consistent (`-cpu=1` sets GOMAXPROCS=1;
  NumCPU=5 is the machine's CPU count) — the "discrepancy" raised in review was a misread.

## Dropped in debate

- "Evidence overwritten once" as a live defect (transparently corrected and verified).
- Devil's Advocate's "close the issue" verdict (refined to "keep open but reframe").
