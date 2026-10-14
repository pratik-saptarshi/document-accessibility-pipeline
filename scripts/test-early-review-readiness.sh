#!/usr/bin/env bash
set -euo pipefail

# This is a synthetic Beads graph-reachability test only. Closing prerequisite
# statuses below simulates readiness; it is not evidence or acceptance for any
# product, candidate, preservation, ua2, semantic, or AT outcome.

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_root="$(mktemp -d "${TMPDIR:-/tmp}/dap-early-review-readiness.XXXXXX")"
trap 'rm -rf "$tmp_root"' EXIT

export HOME="$tmp_root/home"
export XDG_CONFIG_HOME="$HOME/.config"
mkdir -p "$HOME"

cd "$tmp_root"
bd init --non-interactive --skip-agents --skip-hooks --prefix dap >/dev/null
bd import "$repo_root/.beads/issues.jsonl" >/dev/null

target=dap-jo9.2.5.9
gates=(dap-jo9.1.1.4 dap-jo9.1.6.4 dap-jo9.2.5.2)
is_ready() {
  local issue="$1" output="$2"
  python3 -c 'import json,sys; target=sys.argv[1]; rows=json.load(sys.stdin); sys.exit(0 if any(row.get("id")==target for row in rows) else 1)' "$issue" <<<"$output"
}

is_blocked() {
  local issue="$1" output="$2"
  python3 -c 'import json,sys; target=sys.argv[1]; rows=json.load(sys.stdin); sys.exit(0 if any(row.get("id")==target for row in rows) else 1)' "$issue" <<<"$output"
}

readiness() {
  # `--limit 0` is explicitly unlimited in bd v1.3.1. `bd blocked` has no
  # limit/pagination option, so its structured JSON result covers the full set.
  ready_json="$(bd ready --json --brief --limit 0)"
  blocked_json="$(bd blocked --json)"
}

for gate in "${gates[@]}"; do
  bd update "$gate" --force --status closed >/dev/null
done
readiness
is_ready "$target" "$ready_json" || {
  echo "FAIL: $target not ready when all three direct gates are simulated closed" >&2
  exit 1
}
if is_blocked "$target" "$blocked_json"; then
  echo "FAIL: $target remains blocked when all three direct gates are simulated closed" >&2
  exit 1
fi
for later in dap-jo9.2.5 dap-jo9.2.4 dap-jo9.2.4.2; do
  if is_ready "$later" "$ready_json" || ! is_blocked "$later" "$blocked_json"; then
    echo "FAIL: later pipeline issue $later is not blocked in the early-review simulation" >&2
    exit 1
  fi
done
echo "PASS (simulation only): $target graph-ready; later pipeline remains blocked"

for unmet in "${gates[@]}"; do
  bd update "$unmet" --force --status open >/dev/null
  readiness
  if is_ready "$target" "$ready_json" || ! is_blocked "$target" "$blocked_json"; then
    echo "FAIL: $target became ready with unmet direct gate $unmet" >&2
    exit 1
  fi
  echo "PASS (simulation only): unmet gate $unmet blocks $target"
  bd update "$unmet" --force --status closed >/dev/null
done

echo 'SIMULATION ONLY: no human/AT acceptance, eligible candidate, preservation result, or PDF/UA-2 result was created.'
