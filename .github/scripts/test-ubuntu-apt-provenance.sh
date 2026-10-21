#!/usr/bin/env bash
set -euo pipefail

readonly temp="$(mktemp -d)"
trap 'rm -rf -- "$temp"' EXIT
readonly helper="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/describe-ubuntu-apt-provenance.sh"
readonly source_file="$temp/ubuntu.sources"
readonly mirrors_file="$temp/apt-mirrors.txt"
readonly package_version='24.02.0-1ubuntu9.10'

cat >"$source_file" <<'SOURCES'
Types: deb
URIs: mirror+file:/etc/apt/apt-mirrors.txt
Suites: noble noble-updates noble-backports
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: mirror+file:/etc/apt/apt-mirrors.txt
Suites: noble-security
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
SOURCES
cat >"$mirrors_file" <<'MIRRORS'
# GitHub ubuntu-24.04 image mirror candidates
MIRRORS
printf 'http://azure.archive.ubuntu.com/ubuntu/\tpriority:1\nhttps://archive.ubuntu.com/ubuntu/\tpriority:2\nhttps://security.ubuntu.com/ubuntu/\tpriority:3\n' >>"$mirrors_file"
readonly policy="$(cat <<EOF
poppler-utils:
  Installed: (none)
  Candidate: $package_version
  Version table:
     $package_version 500
        500 http://azure.archive.ubuntu.com/ubuntu noble-updates/universe amd64 Packages
EOF
)"

describe() {
  printf '%s\n' "$policy" | bash "$helper" "$package_version" "$source_file" "$mirrors_file"
}

describe_policy() {
  local supplied_policy="$1"
  printf '%s\n' "$supplied_policy" | bash "$helper" "$package_version" "$source_file" "$mirrors_file"
}

expect_rejected() {
  local name="$1"
  if describe >"$temp/$name.out" 2>"$temp/$name.err"; then
    printf 'Ubuntu APT provenance helper unexpectedly accepted %s\n' "$name" >&2
    exit 1
  fi
}

description="$(describe)"
grep -Fq 'Ubuntu archive keyring' <<<"$description"
grep -Fq 'APT source isolation: dedicated source file; sourceparts disabled' <<<"$description"
grep -Fq 'mirror+file:/etc/apt/apt-mirrors.txt' <<<"$description"
grep -Fq 'http://azure.archive.ubuntu.com/ubuntu/' <<<"$description"
grep -Fq 'https://archive.ubuntu.com/ubuntu/' <<<"$description"
grep -Fq 'https://security.ubuntu.com/ubuntu/' <<<"$description"

# Package-specific `apt-cache policy <package>` does not include Release rows.
# The authenticated source configuration binds this URI to Ubuntu's keyring.
# Copied labels cannot rescue a foreign package-source URI.
foreign_policy="${policy/http:\/\/azure.archive.ubuntu.com\/ubuntu/https:\/\/evil.example.invalid\/ubuntu}"
foreign_policy="${foreign_policy/origin azure.archive.ubuntu.com/origin evil.example.invalid}"
if describe_policy "$foreign_policy" >"$temp/foreign-release.out" 2>"$temp/foreign-release.err"; then
  printf 'Ubuntu APT provenance helper unexpectedly accepted foreign-source metadata with copied Ubuntu labels\n' >&2
  exit 1
fi
mixed_policy="$policy
        500 https://evil.example.invalid/ubuntu noble-updates/universe amd64 Packages
           release v=24.04,o=Ubuntu,a=noble-updates,n=noble,l=Ubuntu,c=universe,b=amd64
           origin evil.example.invalid"
if describe_policy "$mixed_policy" >"$temp/mixed-release.out" 2>"$temp/mixed-release.err"; then
  printf 'Ubuntu APT provenance helper unexpectedly combined an approved URI with foreign Release labels\n' >&2
  exit 1
fi
local_policy="${policy/http:\/\/azure.archive.ubuntu.com\/ubuntu/file:\/tmp\/untrusted-mirror}"
if describe_policy "$local_policy" >"$temp/local-policy.out" 2>"$temp/local-policy.err"; then
  printf 'Ubuntu APT provenance helper unexpectedly accepted a local-file package source\n' >&2
  exit 1
fi

# The configured runner source must bind each APT Release signature to the
# Ubuntu archive keyring, not merely an origin/label string.
sed -i 's|Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg|Signed-By: /tmp/untrusted-archive-keyring.gpg|' "$source_file"
if describe >"$temp/foreign-key.out" 2>"$temp/foreign-key.err"; then
  printf 'Ubuntu APT provenance helper unexpectedly accepted a non-Ubuntu signing keyring\n' >&2
  exit 1
fi
sed -i 's|Signed-By: /tmp/untrusted-archive-keyring.gpg|Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg|' "$source_file"
printf 'Trusted: yes\n' >>"$source_file"
if describe >"$temp/insecure-source-option.out" 2>"$temp/insecure-source-option.err"; then
  printf 'Ubuntu APT provenance helper unexpectedly accepted an insecure Trusted source option\n' >&2
  exit 1
fi
sed -i '$d' "$source_file"

printf 'https://mirror.example.invalid/ubuntu/\tpriority:4\n' >>"$mirrors_file"
expect_rejected foreign-mirror
sed -i '$d' "$mirrors_file"
printf 'file:/tmp/untrusted-ubuntu-mirror/\tpriority:4\n' >>"$mirrors_file"
expect_rejected local-mirror
sed -i '$d' "$mirrors_file"
sed -i 's|mirror+file:/etc/apt/apt-mirrors.txt|file:/tmp/untrusted-mirror|' "$source_file"
expect_rejected unexpected-source-uri
sed -i 's|file:/tmp/untrusted-mirror|mirror+file:/etc/apt/apt-mirrors.txt|' "$source_file"
printf '%s\n' "${policy/noble-updates/noble-proposed}" | bash "$helper" "$package_version" "$source_file" "$mirrors_file" >"$temp/unconfigured-suite.out" 2>"$temp/unconfigured-suite.err" && {
  printf 'Ubuntu APT provenance helper unexpectedly accepted a suite absent from the pinned source configuration\n' >&2
  exit 1
}

printf 'Ubuntu APT provenance tests passed\n'
