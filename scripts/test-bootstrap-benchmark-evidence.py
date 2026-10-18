#!/usr/bin/env python3
"""Fact-check the committed synthetic benchmark evidence and its interpretation."""

from pathlib import Path
import json
import hashlib
import re
import statistics

root = Path(__file__).resolve().parents[1]
raw_path = root / "docs/research/support/bootstrap-verification-cost-2026-10-18.txt"
raw = raw_path.read_text()
findings = (root / "docs/research/bootstrap-verify-cost-findings.md").read_text()
measure_harness = (root / "scripts/measure-bootstrap-verify-cost.sh").read_text()
entrypoint = (root / "scripts/bootstrap-cloud-build-tools.sh").read_text()
manifest = json.loads((root / "cmd/cloud-tools-bootstrap/buildtools.manifest.json").read_text())
superseded = (
    root
    / "docs/research/support/bootstrap-verification-cost-2026-10-18-"
    "superseded-pre-harness-corrections.txt"
)

assert "executed scripts/measure-bootstrap-verify-cost.sh bytes at run start" in raw
assert "sourced scripts/lib/benchmark-provenance.sh bytes at run start" in raw
assert "archive SHA-256 and size checked once at setup" in raw
assert "harness_sha256=24d629ac992909770ce55fb169b6196e9907b47234e7b2a57d431bd143c0cd4c" in superseded.read_text()
assert "stub-raw-large-4m (4 MiB synthetic shell script)" in raw
assert "acceptance controls passed before timing" in raw

for label, path in (
    ("harness_sha256", root / "scripts/measure-bootstrap-verify-cost.sh"),
    ("harness_support_sha256", root / "scripts/lib/benchmark-provenance.sh"),
):
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    assert f"{label}={digest} " in raw, f"raw data does not identify current {label} bytes"

source_pin = manifest["bootstrap_source"]
for field, shell_pin in (
    ("revision", "bootstrap_source_revision"),
    ("storage_revision", "bootstrap_source_storage_revision"),
    ("sha256", "bootstrap_source_sha256"),
    ("size_bytes", "bootstrap_source_size_bytes"),
):
    value = source_pin[field]
    pattern = rf"readonly {shell_pin}='?{re.escape(str(value))}'?"
    assert re.search(pattern, entrypoint), f"entrypoint pin drift for {field}"

assert source_pin["revision"] in measure_harness
assert source_pin["sha256"] in measure_harness
assert str(source_pin["size_bytes"]) in measure_harness
for control in (
    "source checksum mismatch",
    "wrong source version with matching hashes",
    "installed wrong version with matching hash",
    "CLI flag misuse",
):
    assert control in measure_harness, f"harness lacks required negative control: {control}"

deltas = {}
paired_deltas = {}
medians = {}
cell_reps = {}
cell_times = {}
current_cell = None
for line in raw.splitlines():
    cell_header = re.fullmatch(r"## cell flow=(checkout|no-checkout) mode=(bootstrap|verify) cache=(cold|warm)", line)
    if cell_header:
        current_cell = cell_header.groups()
        cell_reps[current_cell] = {}
        continue
    median_line = re.fullmatch(r"MEDIAN=([0-9.]+)", line)
    if median_line and current_cell:
        medians[current_cell] = float(median_line.group(1))
        continue
    rep = re.fullmatch(r"rep ([1-5]) ([0-9.]+)s rc=0 pair_position=([12])", line)
    if rep and current_cell:
        rep_no, elapsed, order_pos = rep.groups()
        cell_reps[current_cell][int(rep_no)] = int(order_pos)
        cell_times[current_cell, int(rep_no)] = float(elapsed)

    match = re.fullmatch(
        r"delta (checkout|no-checkout) (cold|warm): bootstrap=([0-9.]+)s "
        r"verify=([0-9.]+)s delta=([-0-9.]+)s "
        r"relative_increase=([-0-9.]+)% share_of_combined=([-0-9.]+)%",
        line,
    )
    if match:
        flow, cache, bootstrap, combined, delta, relative, share = match.groups()
        b, v, d = map(float, (bootstrap, combined, delta))
        assert abs((v - b) - d) < 0.000002, line
        assert round(d / b * 100, 1) == float(relative), line
        assert round(d / v * 100, 1) == float(share), line
        deltas[(flow, cache)] = d

    pair_match = re.fullmatch(
        r"paired-delta sensitivity (checkout|no-checkout) (cold|warm): "
        r"samples=([-0-9.,]+) median=([-0-9.]+)s "
        r"relative_to_bootstrap_median=([-0-9.]+)% "
        r"share_of_combined_median=([-0-9.]+)%",
        line,
    )
    if pair_match:
        flow, cache, sample_text, delta_text, relative_text, share_text = pair_match.groups()
        samples = [float(value) for value in sample_text.split(",")]
        computed = [
            cell_times[(flow, "verify", cache), rep_no]
            - cell_times[(flow, "bootstrap", cache), rep_no]
            for rep_no in range(1, 6)
        ]
        assert samples == [round(value, 6) for value in computed]
        delta = statistics.median(computed)
        assert round(delta, 6) == float(delta_text)
        b = medians[(flow, "bootstrap", cache)]
        v = medians[(flow, "verify", cache)]
        assert round(delta / b * 100, 1) == float(relative_text)
        assert round(delta / v * 100, 1) == float(share_text)
        paired_deltas[(flow, cache)] = delta

assert len(deltas) == 4, f"expected four flow/cache comparisons, got {len(deltas)}"
assert all(delta > 0 for delta in deltas.values()), deltas
assert len(paired_deltas) == 4, f"expected four paired sensitivity rows, got {len(paired_deltas)}"
triggered = {
    key for key, delta in deltas.items()
    if delta >= 2.0 or delta / medians[(key[0], "bootstrap", key[1])] * 100 >= 25.0
}
assert triggered == {
    ("checkout", "cold"),
    ("checkout", "warm"),
    ("no-checkout", "warm"),
}, triggered
assert len(medians) == 8, f"expected 8 raw medians, got {len(medians)}"
table_rows = {}
for line in findings.splitlines():
    row = re.fullmatch(
        r"\| (checkout|no-checkout) \| (cold|warm) \| ([0-9.]+)s \| "
        r"([0-9.]+)s \| \+([0-9.]+)s \(\+([0-9.]+)%; share ([0-9.]+)%\) \|",
        line,
    )
    if row:
        flow, cache, bootstrap, combined, delta, relative, share = row.groups()
        key = (flow, cache)
        table_rows[key] = tuple(map(float, (bootstrap, combined, delta, relative, share)))
assert set(table_rows) == set(deltas), "findings table rows do not cover exactly the raw comparisons"
for flow, cache in deltas:
    row = table_rows[(flow, cache)]
    expected = (
        round(medians[(flow, "bootstrap", cache)], 3),
        round(medians[(flow, "verify", cache)], 3),
        round(deltas[(flow, cache)], 3),
    )
    assert row[:3] == expected, f"findings table {flow}/{cache} does not match raw medians/delta"
    baseline = medians[(flow, "bootstrap", cache)]
    combined = medians[(flow, "verify", cache)]
    assert row[3] == round(deltas[(flow, cache)] / baseline * 100, 1)
    assert row[4] == round(deltas[(flow, cache)] / combined * 100, 1)
assert len(cell_reps) == 8, f"expected 8 measured cells, got {len(cell_reps)}"
for flow in ("checkout", "no-checkout"):
    for cache in ("cold", "warm"):
        for rep_no in range(1, 6):
            bootstrap_pos = cell_reps[(flow, "bootstrap", cache)][rep_no]
            verify_pos = cell_reps[(flow, "verify", cache)][rep_no]
            expected_bootstrap = 1 if rep_no % 2 == 1 else 2
            assert bootstrap_pos == expected_bootstrap
            assert verify_pos == 3 - expected_bootstrap
assert "all 8 cells show the same directional effect" not in findings
assert "All four marginal-median deltas are positive" in findings
assert "original #132 declaration on 2026-10-17 registered a different criterion" in findings
assert "mean standalone `--verify` cost of at least 2 seconds" in findings
assert "verification cost\nat least 25% of combined total setup time" in findings
assert "Revised marginal-median trigger" in findings
assert "2026-10-18T12:50Z" in findings and "13:10:51Z" in findings
assert "Do not treat either current\ncalculation below as that historical rule" in findings
assert "25% share of combined time" in findings and "33⅓% increase over the baseline" in findings
assert "Registered study trigger" not in findings
assert "the trigger is **met**" in findings
assert "checkout cold (+2.492 s), checkout warm" in findings
assert "+64.4%), and no-checkout warm (+26.1%)" in findings
assert "Paired-delta sensitivity" in findings
assert "The cold checkout paired median is\n+1.830191s" in findings
assert "this estimator is not the median of" in findings
assert "direct orchestrator microbenchmarks" in findings
assert "excludes the production" in findings and "`run_group` wrapper" in findings
assert "superseded-pre-review-corrections.txt" in findings
assert "# scope correction: these measurements exclude the production run_group wrapper" in raw
assert "not the median of within-repetition differences" in raw
assert "not an\nowner-approved performance policy" in findings
assert "Relative increase is `delta / median(bootstrap-only)`" in findings
assert "pass's share is `delta /\nmedian(bootstrap+verify)`" in findings
assert "setup-only provenance" in findings
assert "superseded-pre-harness-corrections.txt" in findings
assert "superseded-intermediate-run.txt" in findings
intermediate = (
    root / "docs/research/support/bootstrap-verification-cost-2026-10-18-"
    "superseded-intermediate-run.txt"
).read_text()
assert "12:53:48 UTC" in intermediate and "8640a4" in intermediate
assert "warm no-checkout verify pass is +7.9%" not in findings
for line in raw.splitlines():
    match = re.fullmatch(
        r"delta (checkout|no-checkout) (cold|warm): bootstrap=([0-9.]+)s "
        r"verify=([0-9.]+)s delta=([-0-9.]+)s "
        r"relative_increase=([-0-9.]+)% share_of_combined=([-0-9.]+)%",
        line,
    )
    if match:
        relative, share = match.group(6), match.group(7)
        assert f"{relative}%" in findings, f"findings omit raw relative increase {relative}%"
        assert f"{share}%" in findings, f"findings omit raw combined share {share}%"

print("benchmark evidence estimators, denominators, provenance, and supersession checks passed")
