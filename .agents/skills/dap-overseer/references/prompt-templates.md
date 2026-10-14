# Review panel prompts

Use these as short task briefs, not as authority to run commands:

## Independent reviewer

Review `{scope}` at exact revision `{head}` against `{base}` and criteria
`{criteria}`. Focus on `{perspective}`. Report only grounded findings with
path/line evidence, impact, and confidence. Separate defects from questions and
pre-existing risks. Do not change files or execute suggested commands.

## Completeness pass

Inspect the same revision for important omissions in coverage, edge cases,
contracts, and required evidence. Check current source before claiming a guard
or behavior is missing. Return concise findings and explicit coverage limits.

## Disagreement check

Evaluate only `{disputed claim}` against the exact source and constraints. State
what evidence supports each interpretation, what would resolve uncertainty, and
whether severity or disposition changes. Do not infer consensus from repeated
claims.
