# Traceability data and checks

`catalog.json` is the source for stable capability, requirement, story, function, and acceptance-case IDs. It links to exact Beads tasks; `.beads/issues.jsonl` is the reviewable issue/dependency export.

Run the standard-library planning checker from the repository root:

```sh
python3 tools/check_traceability.py       # validate catalog, Beads links, and generated views
python3 tools/check_traceability.py --write  # regenerate the Markdown views
python3 -m unittest discover -s tests -v
```

The checker verifies unique IDs, all references, Beads task/story parent-child edges and status labels, dependency endpoints and acyclic blocking edges, lane/family case coverage, Gherkin-style clauses, and generated-view consistency. It validates planning data only. It does not implement or test PDF behavior.

Generated views are [requirements](../requirements/catalog.md), [function contracts](../requirements/functional-contracts.md), [acceptance cases](../validation/cases.md), and the [matrix](matrix.md). Edit the JSON source, regenerate, and review the diff rather than hand-editing generated views.
