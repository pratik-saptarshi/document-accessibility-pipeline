import copy
import json
import unittest
from pathlib import Path

from tools.check_traceability import load_issues, render, validate

ROOT = Path(__file__).resolve().parents[1]


class TraceabilityTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.catalog = json.loads((ROOT / "docs/traceability/catalog.json").read_text())
        cls.issues = load_issues(ROOT / ".beads/issues.jsonl")

    def test_catalog_validates_against_beads_export(self):
        validate(self.catalog, self.issues)

    def test_generated_views_include_specification_only_status(self):
        cases = render(self.catalog, "tests")
        matrix = render(self.catalog, "matrix")
        self.assertIn("specifications only; none of these product cases have been executed", cases)
        self.assertIn("Definition of Ready", matrix)
        self.assertIn("Definition of Done", matrix)

    def test_duplicate_id_is_rejected(self):
        data = copy.deepcopy(self.catalog)
        data["tests"].append(copy.deepcopy(data["tests"][0]))
        with self.assertRaisesRegex(ValueError, "duplicate catalog ID"):
            validate(data, self.issues)

    def test_broken_reference_is_rejected(self):
        data = copy.deepcopy(self.catalog)
        data["stories"][0]["requirements"].append("REQ-MISSING")
        with self.assertRaisesRegex(ValueError, "unresolved requirement reference"):
            validate(data, self.issues)

    def test_beads_dependency_cycle_is_rejected(self):
        issues = copy.deepcopy(self.issues)
        issues["dap-jo9"]["dependencies"] = [{"type": "blocks", "depends_on_id": "dap-jo9.2"}]
        issues["dap-jo9.2"]["dependencies"].append({"type": "blocks", "depends_on_id": "dap-jo9"})
        with self.assertRaisesRegex(ValueError, "dependency cycle"):
            validate(self.catalog, issues)

    def test_stale_beads_case_id_is_rejected(self):
        issues = copy.deepcopy(self.issues)
        issues["dap-jo9.2.3"]["acceptance_criteria"] += " TC-NOT-IN-CATALOG-001"
        with self.assertRaisesRegex(ValueError, "unknown acceptance case"):
            validate(self.catalog, issues)

    def test_missing_lane_family_case_is_rejected(self):
        data = copy.deepcopy(self.catalog)
        data["tests"] = [c for c in data["tests"] if not (c["lane"] == "generation" and c["family"] == "reports")]
        with self.assertRaisesRegex(ValueError, "missing lane/family case coverage"):
            validate(data, self.issues)


if __name__ == "__main__":
    unittest.main()
