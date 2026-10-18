#!/usr/bin/env bash

harness_file_sha256() { # path
  [[ -f "$1" ]] || { printf 'harness file not found: %s\n' "$1" >&2; return 1; }
  sha256sum -- "$1" | cut -d' ' -f1
}

verify_archive_fixture() { # path expected_sha256 expected_size_bytes
  local archive="$1" expected_sha="$2" expected_size="$3" actual_size
  [[ -f "$archive" ]] || { printf 'source-archive fixture missing: %s\n' "$archive" >&2; return 1; }
  [[ "$expected_sha" =~ ^[[:xdigit:]]{64}$ ]] \
    || { printf 'invalid source-archive SHA-256 pin\n' >&2; return 1; }
  [[ "$expected_size" =~ ^[0-9]+$ ]] \
    || { printf 'invalid source-archive size pin\n' >&2; return 1; }
  actual_size="$(wc -c < "$archive" | tr -d '[:space:]')"
  [[ "$actual_size" == "$expected_size" ]] \
    || { printf 'source-archive fixture size mismatch: expected %s, got %s\n' "$expected_size" "$actual_size" >&2; return 1; }
  printf '%s  %s\n' "$expected_sha" "$archive" | sha256sum --check --status \
    || { printf 'source-archive fixture checksum mismatch\n' >&2; return 1; }
}

benchmark_delta_percentages() { # bootstrap_only_seconds bootstrap_plus_verify_seconds
  awk -v b="$1" -v v="$2" 'BEGIN {
    delta = v - b
    relative = (b > 0) ? delta / b * 100 : 0
    combined = (v > 0) ? delta / v * 100 : 0
    printf "%.1f %.1f\n", relative, combined
  }'
}
