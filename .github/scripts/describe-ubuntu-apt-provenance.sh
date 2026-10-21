#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" != 3 ]]; then
  printf 'usage: %s <version> <ubuntu.sources> <apt-mirrors.txt>\n' "$0" >&2
  exit 2
fi

readonly version="$1"
readonly source_file="$2"
readonly mirrors_file="$3"
[[ -r "$source_file" && -r "$mirrors_file" ]] || {
  printf 'Ubuntu APT source configuration or mirror list is unreadable\n' >&2
  exit 1
}

# The wrapper URI is the source configured by the pinned ubuntu-24.04 runner.
# It does not identify which candidate mirror served a particular request.
source_uris="$(awk '
  /^[[:space:]]*URIs:[[:space:]]*/ {
    sub(/^[[:space:]]*URIs:[[:space:]]*/, "")
    count = split($0, values, /[[:space:]]+/)
    for (i = 1; i <= count; i++) if (values[i] != "") print values[i]
  }
' "$source_file")"
[[ -n "$source_uris" ]] || {
  printf 'Ubuntu APT source file has no URI entries\n' >&2
  exit 1
}
while IFS= read -r source_uri; do
  [[ "$source_uri" == 'mirror+file:/etc/apt/apt-mirrors.txt' ]] || {
    printf 'Ubuntu APT source URI differs from the pinned runner mirror configuration\n' >&2
    exit 1
  }
done <<<"$source_uris"

mirror_candidates=()
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ "$line" =~ ^[[:space:]]*(#|$) ]] && continue
  read -r mirror_uri _metadata <<<"$line"
  if [[ ! "$mirror_uri" =~ ^https?://([^/:?#@]+)(/[^?#]*)$ ]]; then
    printf 'Ubuntu mirror list contains a non-HTTP(S) or malformed candidate\n' >&2
    exit 1
  fi
  host="${BASH_REMATCH[1]}"
  path="${BASH_REMATCH[2]}"
  case "$host" in
    azure.archive.ubuntu.com|archive.ubuntu.com|security.ubuntu.com) ;;
    *)
      printf 'Ubuntu mirror list contains an unapproved host: %s\n' "$host" >&2
      exit 1
      ;;
  esac
  [[ "$path" == /ubuntu || "$path" == /ubuntu/ ]] || {
    printf 'Ubuntu mirror list candidate has an unexpected archive path\n' >&2
    exit 1
  }
  mirror_candidates+=("${mirror_uri%/}/")
done <"$mirrors_file"

[[ "${#mirror_candidates[@]}" -gt 0 ]] || {
  printf 'Ubuntu APT mirror list has no validated candidates\n' >&2
  exit 1
}

policy="$(cat)"
printf '%s\n' "$policy" | awk -v expected="$version" '
  /^[[:space:]]/ {
    line = $0
    sub(/^[[:space:]]*\*+[[:space:]]*/, "", line)
    sub(/^[[:space:]]+/, "", line)
    fields = split(line, parts, /[[:space:]]+/)
    if (fields >= 2 && parts[2] ~ /^[0-9]+$/) active = (parts[1] == expected)
  }
  active && /release / && /(^|,)o=Ubuntu(,|$)/ && /(^|,)l=Ubuntu(,|$)/ && /(^|,)n=noble(,|$)/ && /a=noble(-updates|-security|-backports)?(,|$)/ {
    found = 1
  }
  END { exit !found }
' || {
  printf 'pinned package version has no authenticated Ubuntu Noble release entry in APT policy\n' >&2
  exit 1
}

printf 'Configured Ubuntu APT source: mirror+file:/etc/apt/apt-mirrors.txt\n'
printf 'Validated configured mirror candidates (serving candidate not inferred):\n'
printf -- '- %s\n' "${mirror_candidates[@]}"
printf 'Authenticated APT release: origin=Ubuntu, label=Ubuntu, codename=noble; pinned version=%s\n' "$version"
