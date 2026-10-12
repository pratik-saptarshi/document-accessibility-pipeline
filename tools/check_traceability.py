#!/usr/bin/env python3
"""Validate and render the planning traceability catalog (standard library only)."""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CATALOG = ROOT / "docs/traceability/catalog.json"
ISSUES = ROOT / ".beads/issues.jsonl"
OUTPUTS = {
    ROOT / "docs/requirements/catalog.md": "requirements",
    ROOT / "docs/requirements/functional-contracts.md": "functions",
    ROOT / "docs/validation/cases.md": "tests",
    ROOT / "docs/traceability/matrix.md": "matrix",
}


def load_issues(path: Path) -> dict[str, dict]:
    result = {}
    for number, line in enumerate(path.read_text().splitlines(), 1):
        if not line.strip():
            continue
        item = json.loads(line)
        if item["id"] in result:
            raise ValueError(f"duplicate Beads issue ID {item['id']} at line {number}")
        result[item["id"]] = item
    return result


def validate(data: dict, issues: dict[str, dict]) -> None:
    errors: list[str] = []
    collections = ["capabilities", "requirements", "stories", "functions", "tasks", "tests"]
    by_id: dict[str, dict] = {}
    for collection in collections:
        for item in data.get(collection, []):
            key = item["id"]
            if key in by_id:
                errors.append(f"duplicate catalog ID: {key}")
            by_id[key] = item

    def ref(owner: str, key: str, target: str, allowed: set[str] | None = None) -> None:
        if target not in by_id or (allowed is not None and target not in allowed):
            errors.append(f"{owner}: unresolved {key} reference {target}")

    caps = {x["id"] for x in data.get("capabilities", [])}
    reqs = {x["id"] for x in data.get("requirements", [])}
    stories = {x["id"] for x in data.get("stories", [])}
    funcs = {x["id"] for x in data.get("functions", [])}
    tasks = {x["id"] for x in data.get("tasks", [])}
    tests = {x["id"] for x in data.get("tests", [])}
    groups = {x["id"]: x for x in data.get("groups", [])}
    for group_id, group in groups.items():
        ref(group_id, "group story", group["story"], stories)
        if group_id not in issues:
            errors.append(f"missing Beads task group {group_id}")
    for req in data.get("requirements", []):
        ref(req["id"], "capability", req["capability"], caps)
    for story in data.get("stories", []):
        ref(story["id"], "capability", story["capability"], caps)
        if not story.get("requirements"):
            errors.append(f"{story['id']}: no requirement links")
        for item in story.get("requirements", []): ref(story["id"], "requirement", item, reqs)
        for item in story.get("tests", []): ref(story["id"], "test", item, tests)
        if story.get("beads") not in issues:
            errors.append(f"{story['id']}: missing Beads story {story.get('beads')}")
    for fn in data.get("functions", []):
        ref(fn["id"], "story", fn["story"], stories)
        for field in ("inputs", "outputs", "rules", "errors"):
            if not fn.get(field, "").strip(): errors.append(f"{fn['id']}: missing {field} contract")
    for task in data.get("tasks", []):
        ref(task["id"], "story", task["story"], stories)
        issue = issues.get(task["id"])
        group_id = task.get("group")
        if group_id:
            if group_id not in groups:
                errors.append(f"{task['id']}: unresolved task group {group_id}")
            elif groups[group_id]["story"] != task["story"]:
                errors.append(f"{task['id']}: task group {group_id} is mapped to a different story")
        expected_parent = group_id or by_id[task["story"]].get("beads")
        if issue is None:
            errors.append(f"missing Beads task {task['id']}")
        elif not any(e.get("type") == "parent-child" and e.get("depends_on_id") == expected_parent for e in issue.get("dependencies", [])):
            errors.append(f"{task['id']}: Beads parent does not match expected story/group {expected_parent}")
        elif task["status"] not in issue.get("labels", []):
            errors.append(f"{task['id']}: status {task['status']} missing from Beads labels")
        if not task.get("tests"):
            errors.append(f"{task['id']}: no test case links")
        for test_id in task.get("tests", []): ref(task["id"], "test", test_id, tests)
    for case in data.get("tests", []):
        for field in ("given", "when", "then"):
            if not case.get(field, "").strip(): errors.append(f"{case['id']}: missing {field.upper()} clause")
        if case.get("status") != "specification-only, not executed":
            errors.append(f"{case['id']}: test status must say specification-only, not executed")
        if case.get("beads") not in issues:
            errors.append(f"{case['id']}: missing Beads test-spec task {case.get('beads')}")
        elif case["beads"] not in tasks:
            errors.append(f"{case['id']}: Beads test-spec issue is not a catalog refinement task")

    # Existing Beads acceptance references must resolve to the same stable test IDs.
    for issue in issues.values():
        referenced_cases = set(re.findall(r"TC-[A-Z0-9-]+", issue.get("description", "") + " " + issue.get("acceptance_criteria", "")))
        for case_id in sorted(referenced_cases - tests):
            errors.append(f"{issue['id']}: Beads references unknown acceptance case {case_id}")

    # Every requirement must reach a user story, function contract, actionable task, and case.
    for req in data.get("requirements", []):
        linked_stories = [s for s in data["stories"] if req["id"] in s["requirements"]]
        if not linked_stories: errors.append(f"{req['id']}: no user story")
        for story in linked_stories:
            linked_funcs = [f for f in data["functions"] if f["story"] == story["id"]]
            linked_tasks = [t for t in data["tasks"] if t["story"] == story["id"]]
            if not linked_funcs: errors.append(f"{req['id']}: story {story['id']} has no function contract")
            if not linked_tasks: errors.append(f"{req['id']}: story {story['id']} has no actionable task")
            if not any(tests.intersection(t["tests"]) for t in linked_tasks):
                errors.append(f"{req['id']}: story {story['id']} has no task-linked test case")

    # The supported family set must appear in each product lane.
    expected = {("remediation", family) for family in ("statements", "invoices", "correspondence", "reports")}
    expected |= {("generation", family) for family in ("statements", "invoices", "correspondence", "reports")}
    actual = {(case.get("lane"), case.get("family")) for case in data.get("tests", [])}
    for pair in sorted(expected - actual): errors.append(f"missing lane/family case coverage: {pair[0]} / {pair[1]}")

    # Validate the Beads prerequisite graph (parent-child edges excluded from cycle detection).
    graph: dict[str, list[str]] = defaultdict(list)
    for issue in issues.values():
        parents = [e for e in issue.get("dependencies", []) if e.get("type") == "parent-child"]
        if len(parents) > 1:
            errors.append(f"{issue['id']}: multiple Beads parent-child edges")
        for edge in issue.get("dependencies", []):
            target = edge["depends_on_id"]
            if target not in issues:
                errors.append(f"{issue['id']}: dependency endpoint missing {target}")
            elif edge.get("type") != "parent-child":
                graph[issue["id"]].append(target)
    visiting: set[str] = set()
    visited: set[str] = set()
    def visit(node: str) -> None:
        if node in visiting:
            errors.append(f"dependency cycle through {node}")
            return
        if node in visited: return
        visiting.add(node)
        for nxt in graph[node]: visit(nxt)
        visiting.remove(node)
        visited.add(node)
    for node in list(graph): visit(node)

    if errors:
        raise ValueError("\n".join(errors))


def render(data: dict, kind: str) -> str:
    if kind == "requirements":
        lines = ["# Requirements catalog", "", "**Status: Proposed.** Requirement statements are planning contracts. They do not establish compliance, workload, or approved implementation.", ""]
        caps = {c["id"]: c for c in data["capabilities"]}
        for req in data["requirements"]:
            lines += [f"## {req['id']} — {caps[req['capability']]['title']}", "", req["text"], ""]
        lines += ["## Open decisions", ""] + [f"- {x}" for x in data["open_decisions"]] + [""]
    elif kind == "functions":
        lines = ["# Proposed function contracts", "", "These are behavior contracts for planning only. They do not select a language library, engine, service boundary, or production design.", ""]
        for fn in data["functions"]:
            lines += [f"## {fn['id']} ({fn['story']})", "", f"- **Inputs:** {fn['inputs']}", f"- **Outputs:** {fn['outputs']}", f"- **Rules:** {fn['rules']}", f"- **Error states:** {fn['errors']}", ""]
    elif kind == "tests":
        lines = ["# Acceptance case specifications", "", "**Status: specifications only; none of these product cases have been executed.** Cases use synthetic/deidentified fixtures. Open thresholds, profiles, and reviewer authority must be resolved before implementation readiness.", ""]
        for case in data["tests"]:
            lines += [f"## {case['id']} — {case['lane']} / {case['family']}", "", f"- **Beads:** `{case['beads']}`", f"- **Status:** {case['status']}", f"- **Given:** {case['given']}", f"- **When:** {case['when']}", f"- **Then:** {case['then']}", ""]
    else:
        lines = ["# Traceability matrix", "", "The JSON catalog is the planning source of truth for stable IDs and explicit links. The checker validates these links against the tracked Beads JSONL export and generates this view.", "", "```text", "Capability → requirement → proposed persona/story → function contract → Beads task → acceptance case specification", "```", "", "## Stories and linked work", ""]
        for story in data["stories"]:
            reqs = ", ".join(story["requirements"])
            function_ids = ", ".join(f["id"] for f in data["functions"] if f["story"] == story["id"])
            task_ids = ", ".join("`%s` (%s%s)" % (t["id"], t["status"], ", group " + t["group"] if t.get("group") else "") for t in data["tasks"] if t["story"] == story["id"])
            case_ids = ", ".join("`%s`" % x for x in story["tests"])
            lines += [f"### {story['id']} — {story['beads']}", "", story["text"], "", f"- **Capability:** `{story['capability']}`", f"- **Requirements:** {reqs}", f"- **Persona:** {story['persona_status']}", f"- **Functions:** {function_ids}", f"- **Tasks:** {task_ids}", f"- **Cases:** {case_ids}", ""]
        lines += ["## Readiness rules", "", "- **Refinement-ready** means a planning/specification task can be detailed using current information; it does not mean implementation is authorized or product tests pass.", "- **Decision-blocked** means one or more profile, feasibility, fidelity, tool, privacy, or authority decisions must be recorded before implementation.", "- **Definition of Ready:** a specific story, proposed persona, linked requirement, function contract, actionable Beads task, synthetic case IDs with Given/When/Then, and explicit unresolved decisions.", "- **Definition of Done for planning:** IDs and edges validate; each requirement traces through a story, function, task, and case; family coverage exists in both lanes; test states say specification-only; and open decisions are visible.", "- **Definition of Done for future product work:** write failing tests first, implement the smallest slice, and record passing results; demonstrate preservation/failure behavior; verify licenses and versions; independently record machine, human semantic, and assistive-technology review. A planning case is not a passing test.", ""]
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--write", action="store_true", help="write generated Markdown")
    parser.add_argument("--catalog", type=Path, default=CATALOG)
    parser.add_argument("--issues", type=Path, default=ISSUES)
    parser.add_argument("--output-dir", type=Path, default=ROOT)
    args = parser.parse_args()
    data = json.loads(args.catalog.read_text())
    issues = load_issues(args.issues)
    validate(data, issues)
    errors = []
    for path, kind in OUTPUTS.items():
        dest = args.output_dir / path.relative_to(ROOT)
        expected = render(data, kind)
        if args.write:
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_text(expected)
        elif not dest.exists() or dest.read_text() != expected:
            errors.append(f"generated file differs: {dest.relative_to(args.output_dir)} (run with --write)")
    if errors:
        print("\n".join(errors), file=sys.stderr)
        return 1
    print(f"traceability valid: {len(data['requirements'])} requirements, {len(data['stories'])} stories, {len(data['tasks'])} tasks, {len(data['tests'])} specification-only cases")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
