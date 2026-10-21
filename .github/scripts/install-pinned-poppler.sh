#!/usr/bin/env bash
set -euo pipefail

readonly package='poppler-utils'
readonly version='24.02.0-1ubuntu9.10'
readonly architecture="$(dpkg --print-architecture)"
readonly package_spec="${package}:${architecture}=${version}"

sudo apt-get update

# APT accepts repository indexes only after checking their signatures against
# the configured Ubuntu archive keyring. Compare the acquired .deb bytes with
# the SHA256 in that authenticated package index before installing them.
metadata="$(apt-cache show "$package_spec")"
metadata_field() {
  local field="$1" count=0 value
  value="$(awk -v key="${field}:" 'index($0, key) == 1 { print substr($0, length(key) + 2); count++ } END { if (count != 1) exit 1 }' <<<"$metadata")" || {
    printf 'expected exactly one %s field in authenticated APT metadata\n' "$field" >&2
    return 1
  }
  printf '%s' "$value"
}

metadata_package="$(metadata_field Package)"
metadata_version="$(metadata_field Version)"
metadata_architecture="$(metadata_field Architecture)"
metadata_filename="$(metadata_field Filename)"
metadata_sha256="$(metadata_field SHA256)"
[[ "$metadata_package" == "$package" && "$metadata_version" == "$version" && "$metadata_architecture" == "$architecture" ]] || {
  printf 'authenticated APT package identity differs from the pin\n' >&2
  exit 1
}
[[ "$metadata_filename" == pool/main/p/poppler/poppler-utils_*_"${architecture}".deb ]] || {
  printf 'authenticated APT package path is outside the official Ubuntu Poppler pool\n' >&2
  exit 1
}
[[ "$metadata_sha256" =~ ^[[:xdigit:]]{64}$ ]] || {
  printf 'authenticated APT metadata has no valid SHA256 for the pinned package\n' >&2
  exit 1
}

uri_output="$(apt-get --print-uris download "$package_spec")"
package_uri="$(sed -n "s/^[[:space:]]*'\\([^']*\\)'.*/\\1/p" <<<"$uri_output")"
[[ -n "$package_uri" && "$package_uri" != *$'\n'* ]] || {
  printf 'APT did not return exactly one pinned Poppler package URI\n' >&2
  exit 1
}
if [[ "$package_uri" =~ ^https?://([[:alnum:]-]+\.)?archive\.ubuntu\.com/ubuntu/(.+)$ ]]; then
  uri_filename="${BASH_REMATCH[2]}"
else
  printf 'pinned Poppler package resolved to an unsupported source URI: %s\n' "$package_uri" >&2
  exit 1
fi
[[ "$uri_filename" == "$metadata_filename" ]] || {
  printf 'resolved Ubuntu package URI differs from authenticated APT metadata\n' >&2
  exit 1
}

workdir="$(mktemp -d)"
trap 'rm -rf -- "$workdir"' EXIT
(
  cd "$workdir"
  apt-get download "$package_spec"
)
package_deb="$workdir/$(basename "$metadata_filename")"
[[ -f "$package_deb" ]] || {
  printf 'APT did not download the expected pinned package file\n' >&2
  exit 1
}
[[ "$(dpkg-deb --field "$package_deb" Version)" == "$version" && "$(dpkg-deb --field "$package_deb" Architecture)" == "$architecture" ]] || {
  printf 'downloaded Poppler package identity differs from the pin\n' >&2
  exit 1
}
actual_sha256="$(sha256sum "$package_deb" | awk '{print $1}')"
[[ "$actual_sha256" == "$metadata_sha256" ]] || {
  printf 'downloaded Poppler package SHA256 differs from authenticated APT metadata\n' >&2
  exit 1
}

sudo apt install -y "$package_deb"

summary="${GITHUB_STEP_SUMMARY:-/dev/stdout}"
{
  printf '### Poppler package acquisition provenance\n\n'
  printf -- '- Package: `%s`\n- Version: `%s`\n- Architecture: `%s`\n' "$metadata_package" "$metadata_version" "$metadata_architecture"
  printf -- '- Authenticated Ubuntu archive URI: `%s`\n- SHA-256 (matches signed APT package metadata): `%s`\n' "$package_uri" "$actual_sha256"
} >>"$summary"
