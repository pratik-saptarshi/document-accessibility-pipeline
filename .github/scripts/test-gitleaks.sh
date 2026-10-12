#!/usr/bin/env bash
set -euo pipefail

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/clean" "$scratch/finding"
printf '%s\n' 'ordinary synthetic text' > "$scratch/clean/clean.txt"

if ! gitleaks dir --redact=100 --no-banner --timeout=30 "$scratch/clean" >/dev/null 2>&1; then
  printf '%s\n' 'Gitleaks rejected the clean synthetic fixture.' >&2
  exit 1
fi

# Split the documented synthetic AWS example so repository-history scanning never sees it as a token.
fake_key="$(printf '%s%s' 'AKIAIOSFODNN7' 'EXAMPLE')"
printf '%s\n' "$fake_key" > "$scratch/finding/synthetic.txt"
set +e
finding_output="$(gitleaks dir --redact=100 --no-banner --timeout=30 "$scratch/finding" 2>&1)"
finding_status=$?
set -e
if [[ "$finding_status" -eq 0 ]]; then
  printf '%s\n' 'Gitleaks did not fail on the synthetic finding.' >&2
  exit 1
fi
if [[ "$finding_output" != *"synthetic.txt"* || "$finding_output" != *"Finding:"* ]]; then
  printf '%s\n' 'Gitleaks failed the finding fixture without reporting the expected location.' >&2
  exit 1
fi
if [[ "$finding_output" == *"$fake_key"* ]]; then
  printf '%s\n' 'Gitleaks output was not fully redacted.' >&2
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
