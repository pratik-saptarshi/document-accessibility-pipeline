#!/usr/bin/env python3
"""Fact-check the committed synthetic benchmark evidence and its interpretation."""

from pathlib import Path
import json
import hashlib
import re
import statistics
import subprocess


def store_unique_rep(target, rep_no, value):
    if rep_no in target:
        raise ValueError(f"duplicate raw repetition {rep_no}")
    target[rep_no] = value


def expected_rep_numbers(repetitions):
    if repetitions < 5:
        raise ValueError("benchmark requires at least five repetitions")
    return set(range(1, repetitions + 1))


def parse_repetition_rows(text):
    reps = {}
    counts = {}
    times = {}
    current = None
    for line in text.splitlines():
        cell_header = re.fullmatch(
            r"## cell flow=(checkout|no-checkout) mode=(bootstrap|verify) cache=(cold|warm)", line
        )
        if cell_header:
            current = cell_header.groups()
            if current in reps:
                raise ValueError(f"duplicate raw measurement cell {current}")
            reps[current] = {}
            counts[current] = 0
            continue
        rep = re.fullmatch(r"rep ([1-9][0-9]*) ([0-9.]+)s rc=0 pair_position=([12])", line)
        if rep and current:
            rep_no, elapsed, order_pos = rep.groups()
            rep_no = int(rep_no)
            store_unique_rep(reps[current], rep_no, int(order_pos))
            counts[current] += 1
            times[current, rep_no] = float(elapsed)
    return reps, counts, times


duplicate_probe = {1: "original"}
try:
    store_unique_rep(duplicate_probe, 1, "duplicate")
except ValueError:
    pass
else:
    raise AssertionError("duplicate repetition guard accepts an overwritten row")
assert expected_rep_numbers(5) == {1, 2, 3, 4, 5}
assert expected_rep_numbers(6) == {1, 2, 3, 4, 5, 6}, "valid six-repetition evidence is truncated"
synthetic_six_reps = "\n".join(
    f"## cell flow={flow} mode={mode} cache={cache}\n"
    + "\n".join(
        f"rep {rep_no} {rep_no + 0.1:.1f}s rc=0 pair_position={1 if rep_no % 2 else 2}"
        for rep_no in range(1, 7)
    )
    for flow in ("checkout", "no-checkout")
    for cache in ("cold", "warm")
    for mode in ("bootstrap", "verify")
)
six_reps, six_counts, six_times = parse_repetition_rows(synthetic_six_reps)
for cell, cell_values in six_reps.items():
    assert set(cell_values) == expected_rep_numbers(6)
    assert six_counts[cell] == 6
    assert round(statistics.median(six_times[(cell, rep)] for rep in expected_rep_numbers(6)), 1) == 3.6

root = Path(__file__).resolve().parents[1]
raw_path = root / "docs/research/support/bootstrap-verification-cost-2026-10-18.txt"
raw = raw_path.read_text()
findings = (root / "docs/research/bootstrap-verify-cost-findings.md").read_text()
measure_harness = (root / "scripts/measure-bootstrap-verify-cost.sh").read_text()
entrypoint = (root / "scripts/bootstrap-cloud-build-tools.sh").read_text()
bootstrap_tests = (root / "cmd/cloud-tools-bootstrap/internal/cloudtools/bootstrap_test.go").read_text()
manifest = json.loads((root / "cmd/cloud-tools-bootstrap/buildtools.manifest.json").read_text())
superseded = (
    root
    / "docs/research/support/bootstrap-verification-cost-2026-10-18-"
    "superseded-pre-harness-corrections.txt"
)

assert "executed scripts/measure-bootstrap-verify-cost.sh bytes at run start" in raw
assert "sourced scripts/lib/benchmark-provenance.sh bytes at run start" in raw
assert re.search(r"# provenance: .* gomaxprocs=[1-9][0-9]*\b", raw), "raw data must record the Go runtime's actual GOMAXPROCS"
assert "runtime.GOMAXPROCS(0); original header copied nproc" in raw, "raw data must disclose the post-run GOMAXPROCS correction"
assert "runtime.GOMAXPROCS(0)" in (root / "scripts/lib/benchmark-provenance.sh").read_text()
assert 'GO_MAXPROCS="$(host_go_maxprocs "$GO_BIN")"' in measure_harness
assert "archive SHA-256 and size checked once at setup" in raw
assert "harness_sha256=24d629ac992909770ce55fb169b6196e9907b47234e7b2a57d431bd143c0cd4c" in superseded.read_text()
control_start = measure_harness.index("# Acceptance controls run before any timing.")
assert measure_harness.index('HARNESS_SHA="$(harness_file_sha256') < control_start
assert measure_harness.index('HARNESS_SUPPORT_SHA="$(harness_file_sha256') < control_start
assert "stub-raw-large-4m (4 MiB synthetic shell script)" in raw
assert "acceptance controls passed before timing" in raw
assert "TestVerifyInstalledValidatesEveryManifestToolEntry" in bootstrap_tests, "all-entry validation regression is missing"
assert "multi-entry checksum/version validation=pass" in raw, "raw evidence omits the multi-entry validation control"

measured_harness_commit = "4f7009ce53ea29d1d33dc0edbb2d3518ebab5762"
for label, path in (
    ("harness_sha256", "scripts/measure-bootstrap-verify-cost.sh"),
    ("harness_support_sha256", "scripts/lib/benchmark-provenance.sh"),
):
    measured_bytes = subprocess.check_output(
        ["git", "show", f"{measured_harness_commit}:{path}"], cwd=root
    )
    digest = hashlib.sha256(measured_bytes).hexdigest()
    assert f"{label}={digest} " in raw, f"raw data does not identify measured {label} bytes"
assert measured_harness_commit[:7] in findings, "findings omit the measured harness revision"

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
paired_delta_rows = []
medians = {}
cell_reps = {}
cell_reps, cell_raw_row_counts, cell_times = parse_repetition_rows(raw)
current_cell = None
rep_header = re.search(r"^# provenance: reps_per_cell=([0-9]+)$", raw, re.MULTILINE)
assert rep_header, "raw transcript must declare its repetition count"
reps_per_cell = int(rep_header.group(1))
assert reps_per_cell >= 5
for line in raw.splitlines():
    cell_header = re.fullmatch(r"## cell flow=(checkout|no-checkout) mode=(bootstrap|verify) cache=(cold|warm)", line)
    if cell_header:
        current_cell = cell_header.groups()
        continue
    median_line = re.fullmatch(r"MEDIAN=([0-9.]+)", line)
    if median_line and current_cell:
        medians[current_cell] = float(median_line.group(1))
        continue
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
        paired_delta_rows.append((flow, cache))
        samples = [float(value) for value in sample_text.split(",")]
        computed = [
            cell_times[(flow, "verify", cache), rep_no]
            - cell_times[(flow, "bootstrap", cache), rep_no]
            for rep_no in expected_rep_numbers(reps_per_cell)
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
assert deltas[("checkout", "cold")] < 0
assert deltas[("checkout", "warm")] > 0
assert deltas[("no-checkout", "cold")] < 0
assert deltas[("no-checkout", "warm")] > 0
assert len(paired_deltas) == 4, f"expected four paired sensitivity rows, got {len(paired_deltas)}"
assert len(paired_delta_rows) == 4, f"duplicate or missing paired sensitivity rows: {paired_delta_rows}"
assert set(paired_delta_rows) == {
    (flow, cache)
    for flow in ("checkout", "no-checkout")
    for cache in ("cold", "warm")
}
assert raw.count("## Paired-delta sensitivity analysis") == 1
triggered = {
    key for key, delta in deltas.items()
    if delta >= 2.0 or delta / medians[(key[0], "bootstrap", key[1])] * 100 >= 25.0
}
assert triggered == {
    ("checkout", "warm"),
}, triggered
assert len(medians) == 8, f"expected 8 raw medians, got {len(medians)}"
for flow in ("checkout", "no-checkout"):
    for mode in ("bootstrap", "verify"):
        for cache in ("cold", "warm"):
            computed = statistics.median(
                cell_times[(flow, mode, cache), rep_no]
                for rep_no in expected_rep_numbers(reps_per_cell)
            )
            assert round(computed, 6) == medians[(flow, mode, cache)], (
                f"raw median mismatch for {flow}/{cache}/{mode}: "
                f"reported={medians[(flow, mode, cache)]}, computed={computed}"
            )
table_rows = {}
for line in findings.splitlines():
    row = re.fullmatch(
        r"\| (checkout|no-checkout) \| (cold|warm) \| ([0-9.]+)s \| "
        r"([0-9.]+)s \| ([+-][0-9.]+)s \(([+-][0-9.]+)%; share ([+-]?[0-9.]+)%\) \|",
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
assert len(cell_raw_row_counts) == 8
for cell, count in cell_raw_row_counts.items():
    assert count == reps_per_cell, f"{cell} has {count} raw rows; expected {reps_per_cell}"
    assert set(cell_reps[cell]) == expected_rep_numbers(reps_per_cell), f"{cell} has missing or duplicate reps"
for flow in ("checkout", "no-checkout"):
    for cache in ("cold", "warm"):
        for rep_no in expected_rep_numbers(reps_per_cell):
            bootstrap_pos = cell_reps[(flow, "bootstrap", cache)][rep_no]
            verify_pos = cell_reps[(flow, "verify", cache)][rep_no]
            expected_bootstrap = 1 if rep_no % 2 == 1 else 2
            assert bootstrap_pos == expected_bootstrap
            assert verify_pos == 3 - expected_bootstrap
assert "all 8 cells show the same directional effect" not in findings
assert "Negative cold deltas" in findings
assert "original #132 declaration on 2026-10-17 registered a different criterion" in findings
assert "mean standalone `--verify` cost of at least 2 seconds" in findings
assert "verification cost\nat least 25% of combined total setup time" in findings
assert "Exploratory reporting trigger" in findings
assert "2026-10-18T12:50Z" in findings and "14:20:20Z" in findings
assert "That is not the criterion applied to\nthis dataset" in findings
assert "25% share of combined time" in findings and "33⅓% increase over baseline" in findings
assert "Registered study trigger" not in findings
assert "the trigger is **met only in checkout\nwarm** (+95.6%)" in findings
assert "Paired-delta sensitivity" in findings
assert "It is +0.226240s for\ncheckout cold" in findings
assert "sensitivity calculation does not\nreplace the prespecified" in findings
assert "direct orchestrator microbenchmarks" in findings
assert "excludes the production" in findings and "`run_group` wrapper" in findings
assert "superseded-pre-review-corrections.txt" in findings
assert "flows: checkout = `go run`" in raw
assert "excludes the production `scripts/bootstrap-cloud-build-tools.sh` `run_group` wrapper" in findings
assert "not an\nowner-approved performance policy" in findings
assert "Relative\nincrease is `delta / median(bootstrap only)`" in findings
assert "verify)`" in findings
assert "validated once during fixture setup" in findings
assert "superseded-before-paired-summary" in findings
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
