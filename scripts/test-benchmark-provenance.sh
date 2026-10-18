#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/benchmark-provenance.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

archive="${tmp}/archive"
printf 'fixture archive bytes\n' > "$archive"
size="$(wc -c < "$archive" | tr -d '[:space:]')"
sha="$(sha256sum "$archive" | cut -d' ' -f1)"

verify_archive_fixture "$archive" "$sha" "$size"
if verify_archive_fixture "$archive" "$sha" "$((size + 1))" >/dev/null 2>&1; then
  printf 'archive fixture accepted a mismatched size\n' >&2
  exit 1
fi
if verify_archive_fixture "$archive" "$(printf '%064d' 0)" "$size" >/dev/null 2>&1; then
  printf 'archive fixture accepted a mismatched checksum\n' >&2
  exit 1
fi

committed="${tmp}/committed.sh"
executed="${tmp}/executed.sh"
printf 'committed version\n' > "$committed"
cp "$committed" "$executed"
before="$(harness_file_sha256 "$executed")"
printf 'working-tree edit\n' >> "$executed"
after="$(harness_file_sha256 "$executed")"
if [[ "$before" == "$after" ]]; then
  printf 'harness digest did not change when executed bytes changed\n' >&2
  exit 1
fi

pcts="$(benchmark_delta_percentages 0.146909 0.361391)"
if [[ "$pcts" != "146.0 59.3" ]]; then
  printf 'percentage denominators collapsed or changed: got %s, want relative increase 146.0 and combined share 59.3\n' "$pcts" >&2
  exit 1
fi

printf 'benchmark provenance regression checks passed\n'
