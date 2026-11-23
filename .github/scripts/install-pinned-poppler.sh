#!/usr/bin/env bash
set -euo pipefail

readonly package='poppler-utils'
readonly version='24.02.0-1ubuntu9.11'
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

# Stage banners go to stderr so a CI log identifies the last stage reached
# before a failure; they never alter the step summary or any check result.
stage() {
  printf 'poppler-acquisition: stage %s\n' "$1" >&2
}

# Reject untrusted or insecure runner APT configuration before it is used to
# fetch indexes. The same file is selected explicitly in every APT invocation.
stage validate-source
bash "$provenance_helper" "$version" "$ubuntu_sources_file" "$ubuntu_mirrors_file" --validate-source >/dev/null

# A single mirror or index fetch can fail transiently, or one configured
# source can be hard-broken, while the remaining sources still deliver the
# pin's signature-verified indexes into this run's fresh lists directory.
# Retry the update a bounded number of times; if every attempt fails, let
# the authenticated metadata stage below act as the gate instead of the
# update exit status. That stage rejects when the pin's metadata is absent
# or does not verify, so no package is ever downloaded or installed from
# unverified metadata. The retry bounds can be redirected only by the
# isolated shell harness; the workflow does not set either variable.

# Parse a non-negative integer setting explicitly as base 10. Bash
# arithmetic treats leading-zero constants as octal, so fed to the retry
# loop directly 08 would raise an arithmetic error, 00 would silently
# evaluate to zero (skipping the update while it is treated as having
# succeeded), and 010 would silently mean eight attempts. Routing every
# spelling through this parser makes each input either its decimal value
# or a loud rejection before any APT call. Leading zeros are stripped
# before parsing and the stripped form is capped at nine digits, so the
# value handed to 10# always parses exactly and cannot overflow into a
# wrapped, wrong value.
parse_decimal_setting() {
  local name="$1" raw="$2" digits
  case "$raw" in
    '' | *[!0-9]*)
      printf '%s must be a non-negative integer, got %q\n' "$name" "$raw" >&2
      return 1
      ;;
  esac
  digits="${raw#"${raw%%[!0]*}"}"
  if [[ -z "$digits" ]]; then
    printf '0'
    return 0
  fi
  if ((${#digits} > 9)); then
    printf '%s must be a non-negative integer of at most nine digits, got %q\n' "$name" "$raw" >&2
    return 1
  fi
  printf '%d' "$((10#$digits))"
}

update_attempts_raw="${POPPLER_APT_UPDATE_ATTEMPTS:-3}"
update_delay_seconds_raw="${POPPLER_APT_UPDATE_DELAY_SECONDS:-5}"
update_attempts="$(parse_decimal_setting POPPLER_APT_UPDATE_ATTEMPTS "$update_attempts_raw")" || exit 1
update_delay_seconds="$(parse_decimal_setting POPPLER_APT_UPDATE_DELAY_SECONDS "$update_delay_seconds_raw")" || exit 1
if ((update_attempts == 0)); then
  printf 'POPPLER_APT_UPDATE_ATTEMPTS must be a positive integer, got %q\n' "$update_attempts_raw" >&2
  exit 1
fi
readonly update_attempts update_delay_seconds
update_status=0
update_attempt=1
while ((update_attempt <= update_attempts)); do
  stage "update (attempt $update_attempt of $update_attempts)"
  if sudo apt-get "${apt_options[@]}" update; then
    update_status=0
    break
  else
    update_status=$?
  fi
  printf 'poppler-acquisition: apt-get update attempt %d of %d failed with exit status %d\n' "$update_attempt" "$update_attempts" "$update_status" >&2
  if ((update_attempt < update_attempts)); then
    sleep "$update_delay_seconds"
  fi
  update_attempt=$((update_attempt + 1))
done
update_failed=0
if ((update_status != 0)); then
  update_failed=1
  printf 'poppler-acquisition: apt-get update failed after %d attempts (last exit status %d); continuing to authenticated metadata verification - the pinned package is installed only if its signed APT metadata verifies\n' "$update_attempts" "$update_status" >&2
fi

# APT accepts repository indexes only after checking their signatures against
# the configured Ubuntu archive keyring. Compare the acquired .deb bytes with
# the SHA256 in that authenticated package index before installing them.
stage metadata
if ! metadata="$(apt-cache "${apt_options[@]}" show "$package_spec")"; then
  if ((update_failed)); then
    printf 'apt-cache show failed for exact package pin %s after apt-get update failed with exit status %d; no authenticated metadata is available for the pin\n' "$package_spec" "$update_status" >&2
  else
    printf 'apt-cache show failed for exact package pin %s\n' "$package_spec" >&2
  fi
  exit 1
fi
[[ "$metadata" =~ [^[:space:]] ]] || {
  if ((update_failed)); then
    printf 'no authenticated APT metadata returned for exact package pin %s after apt-get update failed with exit status %d\n' "$package_spec" "$update_status" >&2
  else
    printf 'no authenticated APT metadata returned for exact package pin %s\n' "$package_spec" >&2
  fi
  exit 1
}
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
stage provenance
source_provenance="$(printf '%s\n' "$apt_policy" | bash "$provenance_helper" "$version" "$ubuntu_sources_file" "$ubuntu_mirrors_file")"

stage download
(
  cd "$download_dir"
  apt-get "${apt_options[@]}" download "$package_spec"
) || {
  printf 'poppler-acquisition: stage download failed for exact package pin %s\n' "$package_spec" >&2
  exit 1
}
package_deb="$download_dir/$(basename "$metadata_filename")"
[[ -f "$package_deb" ]] || {
  printf 'APT did not download the expected pinned package file\n' >&2
  exit 1
}
[[ "$(dpkg-deb --field "$package_deb" Package)" == "$package" && "$(dpkg-deb --field "$package_deb" Version)" == "$version" && "$(dpkg-deb --field "$package_deb" Architecture)" == "$architecture" ]] || {
  printf 'downloaded Poppler package identity differs from the pin\n' >&2
  exit 1
}
stage verify
actual_sha256="$(sha256sum "$package_deb" | awk '{print $1}')"
[[ "$actual_sha256" == "$metadata_sha256" ]] || {
  printf 'downloaded Poppler package SHA256 differs from authenticated APT metadata\n' >&2
  exit 1
}

stage install
sudo apt "${apt_options[@]}" install -y "$package_deb" || {
  printf 'poppler-acquisition: stage install failed for %s\n' "$package_deb" >&2
  exit 1
}

summary="${GITHUB_STEP_SUMMARY:-/dev/stdout}"
{
  printf '### Poppler package acquisition provenance\n\n'
  printf -- '- Package: `%s`\n- Version: `%s`\n- Architecture: `%s`\n' "$metadata_package" "$metadata_version" "$metadata_architecture"
  printf '\n```text\n%s\n```\n\n' "$source_provenance"
  printf -- '- SHA-256 (matches signed APT package metadata): `%s`\n' "$actual_sha256"
} >>"$summary"
