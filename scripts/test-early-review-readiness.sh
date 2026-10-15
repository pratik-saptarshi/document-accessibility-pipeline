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
specification=dap-jo9.2.5.2
gates=(dap-jo9.1.1.4 dap-jo9.1.1.5 dap-jo9.1.6.4 dap-jo9.2.5.10 dap-jo9.2.5.2 dap-jo9.2.5.9.2)
is_ready() {
  local issue="$1" output="$2"
  python3 -c 'import json,sys; target=sys.argv[1]; rows=json.load(sys.stdin); sys.exit(0 if any(row.get("id")==target for row in rows) else 1)' "$issue" <<<"$output"
}

is_blocked() {
  local issue="$1" output="$2"
  python3 -c 'import json,sys; target=sys.argv[1]; rows=json.load(sys.stdin); sys.exit(0 if any(row.get("id")==target for row in rows) else 1)' "$issue" <<<"$output"
}

has_status() {
  local issue="$1" expected="$2" output="$3"
  python3 -c 'import json,sys; target,status=sys.argv[1:3]; rows=json.load(sys.stdin); sys.exit(0 if any(row.get("id")==target and row.get("status")==status for row in rows) else 1)' "$issue" "$expected" <<<"$output"
}

readiness() {
  # `--limit 0` is explicitly unlimited in bd v1.3.1. `bd blocked` has no
  # limit/pagination option, so its structured JSON result covers the full set.
  ready_json="$(bd ready --json --brief --limit 0)"
  blocked_json="$(bd blocked --json)"
}

# Specification preparation is independently claimable; it is not review
# acceptance and does not close or approve the still-open consumer task.
bd update "$specification" --force --status open >/dev/null
readiness
consumer_json="$(bd show "$target" --json)"
if ! is_ready "$specification" "$ready_json" || is_blocked "$specification" "$blocked_json" || ! has_status "$target" open "$consumer_json"; then
  echo "FAIL: reopened specification is not independently ready with its consumer still open" >&2
  exit 1
fi
echo "PASS (simulation only): reopened specification is ready; consumer remains open"
bd update "$specification" --force --status closed >/dev/null

for gate in "${gates[@]}"; do
  bd update "$gate" --force --status closed >/dev/null
done
readiness
is_ready "$target" "$ready_json" || {
  echo "FAIL: $target not ready when all six direct gates are simulated closed" >&2
  exit 1
}
if is_blocked "$target" "$blocked_json"; then
  echo "FAIL: $target remains blocked when all six direct gates are simulated closed" >&2
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

# Make all other prerequisites of the broader gates synthetic-closed. The
# candidate-review task stays open, so it must hold .2.5 and .5 below ready.
for gate in dap-jo9.1.1.3 dap-jo9.2.4 dap-jo9.3.4 dap-jo9.2; do
  bd update "$gate" --force --status closed >/dev/null
done
readiness
if ! is_ready "$target" "$ready_json" || is_ready dap-jo9.2.5 "$ready_json" || ! is_blocked dap-jo9.2.5 "$blocked_json" || is_ready dap-jo9.5 "$ready_json" || ! is_blocked dap-jo9.5 "$blocked_json"; then
  echo "FAIL: open review did not hold the broader release gates after other blockers were simulated closed" >&2
  exit 1
fi
echo 'PASS (simulation only): open candidate review holds .2.5 and .5 after other blockers pass'

bd update "$target" --force --status closed >/dev/null
readiness
if ! is_ready dap-jo9.2.5 "$ready_json" || is_blocked dap-jo9.2.5 "$blocked_json" || is_ready dap-jo9.5 "$ready_json" || ! is_blocked dap-jo9.5 "$blocked_json"; then
  echo 'FAIL: simulated review closure did not unblock .2.5 while preserving .5 aggregate-gate blocking' >&2
  python3 - "$ready_json" "$blocked_json" <<'PY' >&2
import json,sys
ready={row['id'] for row in json.loads(sys.argv[1])}
blocked={row['id'] for row in json.loads(sys.argv[2])}
for issue in ('dap-jo9.2.5','dap-jo9.5'):
 print(issue, 'ready=',issue in ready,'blocked=',issue in blocked)
PY
  exit 1
fi
echo 'PASS (simulation only): simulated review closure makes .2.5 ready; .5 still awaits aggregate-gate completion'

bd update dap-jo9.2.5 --force --status closed >/dev/null
readiness
if ! is_ready dap-jo9.5 "$ready_json" || is_blocked dap-jo9.5 "$blocked_json"; then
  echo 'FAIL: simulated aggregate-gate completion did not unblock .5 after review closure' >&2
  exit 1
fi
echo 'PASS (simulation only): aggregate-gate completion unblocks .5 after review closure'

echo 'SIMULATION ONLY: no human/AT acceptance, eligible candidate, preservation result, or PDF/UA-2 result was created.'
