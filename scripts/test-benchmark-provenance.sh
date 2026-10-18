#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/benchmark-provenance.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

good_go="${tmp}/go-good"
cat > "$good_go" <<'SH'
#!/bin/sh
case "$1" in
  version) printf 'go version go1.27.2 linux/amd64\n' ;;
  run) printf '7\n' ;;
  *) exit 2 ;;
esac
SH
chmod 755 "$good_go"
resolved="$(GO_BIN="$good_go" resolve_host_go 1.27.2)"
[[ "$resolved" == "$good_go" ]] || { printf 'GO_BIN override not selected: %s\n' "$resolved" >&2; exit 1; }
maxprocs="$(host_go_maxprocs "$good_go")"
[[ "$maxprocs" == 7 ]] || { printf 'runtime GOMAXPROCS query returned %s, expected 7\n' "$maxprocs" >&2; exit 1; }

relative_root="${tmp}/relative-root"
mkdir -p "$relative_root/bin"
cp "$good_go" "$relative_root/bin/go"
(
  cd "$relative_root"
  resolved="$(GO_BIN=bin/go resolve_host_go 1.27.2)"
  [[ "$resolved" == /* ]] || { printf 'relative GO_BIN did not resolve absolutely: %s\n' "$resolved" >&2; exit 1; }
  ( cd / && [[ -x "$resolved" ]] ) || { printf 'resolved GO_BIN failed after cwd change: %s\n' "$resolved" >&2; exit 1; }
)

path_bin="${tmp}/path-bin"
mkdir -p "$path_bin"
cp "$good_go" "$path_bin/go"
resolved="$(PATH="$path_bin:/usr/bin:/bin" GO_BIN= resolve_host_go 1.27.2)"
[[ "$resolved" == "$path_bin/go" ]] || { printf 'PATH go not discovered: %s\n' "$resolved" >&2; exit 1; }
(
  cd "$tmp"
  relative_path_bin="./path-bin"
  mkdir -p "$relative_path_bin"
  cp "$good_go" "$relative_path_bin/go"
  resolved="$(PATH="$relative_path_bin:/usr/bin:/bin" GO_BIN= resolve_host_go 1.27.2)"
  [[ "$resolved" == /* ]] || { printf 'relative PATH go did not resolve absolutely: %s\n' "$resolved" >&2; exit 1; }
  ( cd / && [[ -x "$resolved" ]] ) || { printf 'PATH go failed after cwd change: %s\n' "$resolved" >&2; exit 1; }
)

bad_go="${tmp}/go-bad"
printf '#!/bin/sh\nprintf "go version go1.26.9 linux/amd64\\n"\n' > "$bad_go"
chmod 755 "$bad_go"
if GO_BIN="$bad_go" resolve_host_go 1.27.2 >/dev/null 2>&1; then
  printf 'host Go toolchain accepted the wrong version\n' >&2
  exit 1
fi

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
