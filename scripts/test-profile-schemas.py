#!/usr/bin/env python3
"""Validate both published profile schemas and their behavioral contract.

Requires jsonschema 4.26.0. This exercises the complete Draft 2020-12 schemas,
not only parsed JSON fields or extracted regular expressions.
"""

from __future__ import annotations

import json
import hashlib
import re
import unittest
from pathlib import Path

from jsonschema import Draft202012Validator


ROOT = Path(__file__).resolve().parents[1]
LANES = ("generation", "remediation")
REQUIRED_FIELDS = {
    "statements": ("period_start", "period_end", "account_holder", "transactions"),
    "invoices": ("invoice_number", "invoice_date", "bill_to", "line_items"),
    "correspondence": ("title", "date", "sender", "recipient", "body", "signature_block", "language"),
    "reports": ("title", "report_date", "author", "sections", "tables", "figures", "footnotes"),
    "presentations": ("slide_count", "slide_titles", "speaker_notes_included", "embedded_media", "charts_per_slide"),
    "forms": ("form_fields", "instructions", "sections"),
    "marketing": ("headings", "images", "links", "language"),
    "hr_documents": ("title", "effective_date", "sections", "tables", "signature_fields"),
}


def valid_profile(lane: str, family: str, confidentiality: str = "internal", language: str = "en") -> dict:
    return {
        "schema_version": "1.1",
        "status": "proposed",
        "lane": lane,
        "family": family,
        "language": {"value": language, "status": "proposed"},
        "confidentiality": {"value": confidentiality, "status": "proposed"},
        "retention": {"value": "transient", "status": "proposed"},
        "accessibility": {"value": "semantic expectations open", "status": "open"},
        "pdfua_target": {"value": "PDF/UA-2", "status": "accepted"},
        "human_review": {"value": "review required; protocol open", "status": "open"},
        "source_fields": [
            {"name": name, "type": "string", "status": "proposed"}
            for name in REQUIRED_FIELDS[family]
        ],
    }


class ProfileSchemaBehaviorTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.validators = {}
        for lane in LANES:
            path = ROOT / "docs" / "workflows" / lane / "profiles" / "profile.schema.json"
            schema = json.loads(path.read_text(encoding="utf-8"))
            Draft202012Validator.check_schema(schema)
            cls.validators[lane] = Draft202012Validator(schema)

    def test_public_and_internal_profiles_cover_all_eight_families(self) -> None:
        self.assertEqual(
            set(REQUIRED_FIELDS),
            {"statements", "invoices", "correspondence", "reports", "presentations", "forms", "marketing", "hr_documents"},
        )
        for lane, validator in self.validators.items():
            for family in REQUIRED_FIELDS:
                for confidentiality in ("public", "internal"):
                    with self.subTest(lane=lane, family=family, confidentiality=confidentiality):
                        validator.validate(valid_profile(lane, family, confidentiality))

    def test_restricted_case_and_supported_boundary_whitespace_is_refused(self) -> None:
        values = (
            "restricted",
            "RESTRICTED",
            "ReStRiCtEd",
            " restricted ",
            "\trestricted\n",
            "\u00a0restricted\u00a0",
            "\u0085restricted\u0085",
            "\ufeffrestricted\ufeff",
        )
        for lane, validator in self.validators.items():
            for value in values:
                with self.subTest(lane=lane, value=value):
                    self.assertFalse(validator.is_valid(valid_profile(lane, "forms", value)))

    def test_canonical_english_tags_pass_and_padded_or_other_tags_fail(self) -> None:
        for lane, validator in self.validators.items():
            for language in ("en", "en-US", "en-GB"):
                with self.subTest(lane=lane, accepted=language):
                    validator.validate(valid_profile(lane, "forms", language=language))
            for language in (" en ", " en", "en ", "es", "es-US", "fr"):
                with self.subTest(lane=lane, rejected=language):
                    self.assertFalse(validator.is_valid(valid_profile(lane, "forms", language=language)))

    def test_each_family_required_field_is_enforced(self) -> None:
        for lane, validator in self.validators.items():
            for family, required in REQUIRED_FIELDS.items():
                for missing in required:
                    profile = valid_profile(lane, family)
                    profile["source_fields"] = [field for field in profile["source_fields"] if field["name"] != missing]
                    with self.subTest(lane=lane, family=family, missing=missing):
                        self.assertFalse(validator.is_valid(profile))

    def test_decision_0015_constructible_digest_golden_vectors(self) -> None:
        # These entries use only ASCII keys/values and integer sequence fields,
        # so sorted compact JSON is the RFC 8785 serialization for the vector.
        canonical_inputs = (
            '{"appended_at":"2026-10-19T00:00:00Z","entry_type":"grant","identity":"reviewer@example.test","not_after":"2027-12-31T23:59:59Z","not_before":"2026-10-19T00:00:00Z","role":"approver","schema_version":1,"sequence":1}',
            '{"appended_at":"2026-10-19T12:00:00Z","effective_at":"2026-10-19T12:00:00Z","entry_type":"revocation","revokes_sequence":1,"schema_version":1,"sequence":2}',
        )
        doc = (ROOT / "docs" / "decisions" / "0015-approver-role-evidence-source.md").read_text(encoding="utf-8")
        documented_hashes = re.findall(r"SHA-256 is\s+`([0-9a-f]{64})`", doc)
        expected_hashes = [hashlib.sha256(value.encode("utf-8")).hexdigest() for value in canonical_inputs]
        self.assertEqual(documented_hashes, expected_hashes)
        for value in canonical_inputs:
            self.assertIn(f"`{value}`", doc)


if __name__ == "__main__":
    unittest.main(verbosity=2)
