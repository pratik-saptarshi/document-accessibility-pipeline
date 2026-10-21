#!/usr/bin/env bash
set -euo pipefail

readonly package='poppler-utils'
readonly version='24.02.0-1ubuntu9.10'
readonly architecture="$(dpkg --print-architecture)"
readonly package_spec="${package}:${architecture}=${version}"
readonly provenance_helper="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/describe-ubuntu-apt-provenance.sh"
# These runner-owned paths can be redirected only by the isolated shell
# harness; the workflow does not set either variable.
readonly ubuntu_sources_file="${POPPLER_APT_SOURCES_FILE:-/etc/apt/sources.list.d/ubuntu.sources}"
readonly ubuntu_mirrors_file="${POPPLER_APT_MIRRORS_FILE:-/etc/apt/apt-mirrors.txt}"
readonly workdir="$(mktemp -d)"
trap 'sudo rm -rf -- "$workdir"' EXIT
readonly lists_dir="$workdir/lists"
readonly cache_dir="$workdir/cache"
readonly download_dir="$workdir/download"
mkdir -p "$lists_dir" "$cache_dir" "$download_dir"
chmod 0711 "$workdir"
chmod 0755 "$lists_dir"
chmod 0700 "$cache_dir" "$download_dir"
# APT's unprivileged _apt downloader needs to traverse the temp parent and
# write only to its dedicated partial-download directory.
sudo install -d -o _apt -g root -m 0700 "$lists_dir/partial"
readonly apt_options=(
  -o "Dir::Etc::sourcelist=$ubuntu_sources_file"
  -o 'Dir::Etc::sourceparts=-'
  -o "Dir::State::lists=$lists_dir"
  -o "Dir::Cache::pkgcache=$cache_dir/pkgcache.bin"
  -o "Dir::Cache::srcpkgcache=$cache_dir/srcpkgcache.bin"
)

# Reject untrusted or insecure runner APT configuration before it is used to
# fetch indexes. The same file is selected explicitly in every APT invocation.
bash "$provenance_helper" "$version" "$ubuntu_sources_file" "$ubuntu_mirrors_file" --validate-source >/dev/null
sudo apt-get "${apt_options[@]}" update

# APT accepts repository indexes only after checking their signatures against
# the configured Ubuntu archive keyring. Compare the acquired .deb bytes with
# the SHA256 in that authenticated package index before installing them.
metadata="$(apt-cache "${apt_options[@]}" show "$package_spec")"
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
apt_policy="$(apt-cache "${apt_options[@]}" policy "$package")"
[[ "$metadata_package" == "$package" && "$metadata_version" == "$version" && "$metadata_architecture" == "$architecture" ]] || {
  printf 'authenticated APT package identity differs from the pin\n' >&2
  exit 1
}
[[ "$metadata_filename" == "pool/main/p/poppler/poppler-utils_${version}_${architecture}.deb" ]] || {
  printf 'authenticated APT package path is outside the official Ubuntu Poppler pool\n' >&2
  exit 1
}
[[ "$metadata_sha256" =~ ^[[:xdigit:]]{64}$ ]] || {
  printf 'authenticated APT metadata has no valid SHA256 for the pinned package\n' >&2
  exit 1
}
source_provenance="$(printf '%s\n' "$apt_policy" | bash "$provenance_helper" "$version" "$ubuntu_sources_file" "$ubuntu_mirrors_file")"

(
  cd "$download_dir"
  apt-get "${apt_options[@]}" download "$package_spec"
)
package_deb="$download_dir/$(basename "$metadata_filename")"
[[ -f "$package_deb" ]] || {
  printf 'APT did not download the expected pinned package file\n' >&2
  exit 1
}
[[ "$(dpkg-deb --field "$package_deb" Package)" == "$package" && "$(dpkg-deb --field "$package_deb" Version)" == "$version" && "$(dpkg-deb --field "$package_deb" Architecture)" == "$architecture" ]] || {
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
  printf '\n```text\n%s\n```\n\n' "$source_provenance"
  printf -- '- SHA-256 (matches signed APT package metadata): `%s`\n' "$actual_sha256"
} >>"$summary"
