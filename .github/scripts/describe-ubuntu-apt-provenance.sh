#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" != 3 && ( "$#" != 4 || "$4" != --validate-source ) ]]; then
  printf 'usage: %s <version> <ubuntu.sources> <apt-mirrors.txt> [--validate-source]\n' "$0" >&2
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
# Validate its signing key here; the installer also isolates every APT command
# to this source file and disables sourceparts before acquiring package data.
# An allowlisted field set rejects Trusted and other insecure source options.
awk '
  function finish(    suite_count, suites, i) {
    if (!active) return
    if (field_count != 5 || count["types"] != 1 || count["uris"] != 1 ||
        count["suites"] != 1 || count["components"] != 1 || count["signed-by"] != 1 ||
        value["types"] != "deb" ||
        value["uris"] != "mirror+file:/etc/apt/apt-mirrors.txt" ||
        value["signed-by"] != "/usr/share/keyrings/ubuntu-archive-keyring.gpg" ||
        value["components"] == "") bad = 1
    suite_count = split(value["suites"], suites, /[[:space:]]+/)
    if (suite_count < 1) bad = 1
    for (i = 1; i <= suite_count; i++)
      if (suites[i] !~ /^noble(-updates|-security|-backports)?$/) bad = 1
    stanzas++
    delete count
    delete value
    field_count = 0
    active = 0
  }
  BEGIN {
    allowed["types"] = allowed["uris"] = allowed["suites"] = 1
    allowed["components"] = allowed["signed-by"] = 1
  }
  /^[[:space:]]*$/ { finish(); next }
  /^[[:space:]]*#/ { next }
  /^[[:space:]]/ { bad = 1; next }
  {
    if ($0 !~ /^[A-Za-z-]+:[[:space:]]*/) { bad = 1; next }
    key = tolower($0)
    sub(/:.*/, "", key)
    line = $0
    sub(/^[^:]*:[[:space:]]*/, "", line)
    if (!(key in allowed)) { bad = 1; next }
    count[key]++
    value[key] = line
    field_count++
    active = 1
  }
  END {
    finish()
    if (bad || stanzas == 0) exit 1
  }
' "$source_file" || {
  printf 'Ubuntu APT source configuration is not the authenticated pinned-runner configuration\n' >&2
  exit 1
}

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

print_source_configuration() {
  printf 'Configured Ubuntu APT source: mirror+file:/etc/apt/apt-mirrors.txt\n'
  printf 'APT source isolation: dedicated source file; sourceparts disabled\n'
  printf 'APT Release signing keyring: /usr/share/keyrings/ubuntu-archive-keyring.gpg\n'
  printf 'Validated configured mirror candidates (serving candidate not inferred):\n'
  printf -- '- %s\n' "${mirror_candidates[@]}"
}

if [[ "${4:-}" == --validate-source ]]; then
  print_source_configuration
  exit 0
fi

policy="$(cat)"
candidate_list="$(printf '%s\n' "${mirror_candidates[@]}")"
printf '%s\n' "$policy" | awk -v expected="$version" -v candidate_list="$candidate_list" '
  BEGIN {
    candidate_count = split(candidate_list, candidates, /\n/)
  }
  /^[[:space:]]/ {
    line = $0
    sub(/^[[:space:]]*\*+[[:space:]]*/, "", line)
    sub(/^[[:space:]]+/, "", line)
    fields = split(line, parts, /[[:space:]]+/)
    if (fields >= 2 && parts[2] ~ /^[0-9]+$/) {
      active = (parts[1] == expected)
      active_uri = 0
      active_release = 0
    }
    if (active && fields >= 5 && parts[1] ~ /^[0-9]+$/ &&
        parts[2] == "mirror+file:/etc/apt/apt-mirrors.txt") active_uri = 1
    if (active && fields >= 5 && parts[1] ~ /^[0-9]+$/ && parts[2] ~ /^https?:\/\//) {
      for (i = 1; i <= candidate_count; i++)
        if (parts[2] == candidates[i] || parts[2] == substr(candidates[i], 1, length(candidates[i]) - 1)) active_uri = 1
    }
    if (active && fields >= 2 && parts[2] !~ /^[0-9]+$/ &&
        parts[1] ~ /^release$/ && /(^|,)o=Ubuntu(,|$)/ && /(^|,)l=Ubuntu(,|$)/ &&
        /(^|,)n=noble(,|$)/ && /a=noble(-updates|-security|-backports)?(,|$)/) active_release = 1
  }
  END {
    if (active && active_uri && active_release) found = 1
    exit !found
  }
' || {
  printf 'pinned package version has no authenticated Ubuntu Noble release entry in APT policy\n' >&2
  exit 1
}

print_source_configuration
printf 'Authenticated APT release: origin=Ubuntu, label=Ubuntu, codename=noble; pinned version=%s\n' "$version"
