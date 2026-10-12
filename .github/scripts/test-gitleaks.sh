#!/usr/bin/env bash
set -euo pipefail

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/clean" "$scratch/finding"
printf '%s\n' 'ordinary synthetic text' > "$scratch/clean/clean.txt"
cat > "$scratch/synthetic.toml" <<'TOML'
title = "Synthetic smoke test"

[[rules]]
id = "synthetic-smoke-token"
description = "Finds only the temporary synthetic smoke fixture."
regex = '''SYNTHETIC_TEST_SECRET_[A-Z0-9]{16}'''
secretGroup = 0
TOML

if ! gitleaks dir --config "$scratch/synthetic.toml" --redact=100 --no-banner --timeout=30 "$scratch/clean" >/dev/null 2>&1; then
  printf '%s\n' 'Gitleaks rejected the clean synthetic fixture.' >&2
  exit 1
fi

# Build a pattern-matched placeholder only in the temporary directory.
synthetic_token="$(printf '%s%s' 'SYNTHETIC_TEST_SECRET_' 'Q7M4V2C8N5R1T9X3')"
printf '%s\n' "$synthetic_token" > "$scratch/finding/synthetic.txt"
set +e
gitleaks dir --config "$scratch/synthetic.toml" --report-format=json --report-path="$scratch/finding-report.json" --redact=100 --no-banner --timeout=30 "$scratch/finding" >/dev/null 2>&1
finding_status=$?
set -e
if [[ "$finding_status" -eq 0 ]]; then
  printf '%s\n' 'Gitleaks did not fail on the synthetic finding.' >&2
  exit 1
fi
if ! grep -Fq 'synthetic.txt' "$scratch/finding-report.json" || ! grep -Fq 'synthetic-smoke-token' "$scratch/finding-report.json"; then
  printf '%s\n' 'Gitleaks did not report the expected synthetic finding and rule.' >&2
  exit 1
fi
if grep -Fq "$synthetic_token" "$scratch/finding-report.json"; then
  printf '%s\n' 'Gitleaks report was not fully redacted.' >&2
  exit 1
fi

printf 'version = [\n' > "$scratch/invalid.toml"
set +e
gitleaks dir --config "$scratch/invalid.toml" --redact=100 --no-banner --timeout=30 "$scratch/clean" >/dev/null 2>&1
config_status=$?
set -e
if [[ "$config_status" -eq 0 ]]; then
  printf '%s\n' 'Gitleaks accepted an invalid configuration.' >&2
  exit 1
fi

printf '%s\n' 'Gitleaks smoke checks passed: clean fixture, detected and redacted finding, invalid-config failure.'
