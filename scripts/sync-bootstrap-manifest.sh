#!/usr/bin/env bash
# sync-bootstrap-manifest.sh -- keep the embedded copy of the bootstrap tool
# manifest in sync with its single source of truth.
#
# The root buildtools.manifest.json is the only hand-edited statement of the
# bootstrap pin set (tool versions, digests, sources, and the bootstrap
# source-archive pin). cmd/cloud-tools-bootstrap embeds a copy so cloud
# environments can install without a repository checkout; that copy is a
# generated artifact and must never be edited by hand.
#
# Canonical generated form: the root manifest decoded and re-encoded as
# indented JSON (two spaces) in which each element of a top-level array of
# objects (the tools list) is written compactly on its own line, with a
# trailing newline. This reproduces the embedded copy's long-standing byte
# form exactly, so generation is a pure transform with no format churn.
#
# Usage:
#   scripts/sync-bootstrap-manifest.sh           regenerate the embedded copy
#   scripts/sync-bootstrap-manifest.sh --check   write nothing; exit non-zero
#                                                if the embedded copy is not
#                                                exactly what generation
#                                                would produce
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
source_manifest="${repo_root}/buildtools.manifest.json"
embedded_manifest="${repo_root}/cmd/cloud-tools-bootstrap/buildtools.manifest.json"

generate() {
  # $1 = output path
  python3 - "$source_manifest" "$1" <<'PYEOF'
import copy
import json
import sys

with open(sys.argv[1], encoding="utf-8") as fh:
    manifest = json.load(fh)

# Render top-level arrays of objects with one compact element per line;
# everything else uses the standard two-space indented form.
rendered = copy.deepcopy(manifest)
placeholders = {}
for key, value in manifest.items():
    if isinstance(value, list) and value and all(isinstance(v, dict) for v in value):
        marker = "__ARRAY_OF_OBJECTS_%s__" % key
        placeholders[marker] = (key, value)
        rendered[key] = marker

text = json.dumps(rendered, indent=2, ensure_ascii=False)
for marker, (key, value) in placeholders.items():
    lines = [json.dumps(item, ensure_ascii=False, separators=(",", ":")) for item in value]
    block = '"%s": [\n%s\n  ]' % (key, ",\n".join("    " + line for line in lines))
    text = text.replace('"%s": "%s"' % (key, marker), block)

with open(sys.argv[2], "w", encoding="utf-8") as fh:
    fh.write(text)
    fh.write("\n")
PYEOF
}

check_only=0
case "${1:-}" in
  "") ;;
  --check) check_only=1 ;;
  *)
    echo "usage: $0 [--check]" >&2
    exit 2
    ;;
esac

generated="$(mktemp)"
trap 'rm -f "$generated"' EXIT
generate "$generated"

if cmp -s "$generated" "$embedded_manifest"; then
  echo "embedded bootstrap manifest is in sync with buildtools.manifest.json"
  exit 0
fi

if [ "$check_only" -eq 1 ]; then
  echo "embedded bootstrap manifest is OUT OF SYNC with buildtools.manifest.json:" >&2
  diff -u "$embedded_manifest" "$generated" | head -n 40 >&2 || true
  echo "run scripts/sync-bootstrap-manifest.sh to regenerate it" >&2
  exit 1
fi

cat "$generated" > "$embedded_manifest"
echo "regenerated cmd/cloud-tools-bootstrap/buildtools.manifest.json from buildtools.manifest.json"
