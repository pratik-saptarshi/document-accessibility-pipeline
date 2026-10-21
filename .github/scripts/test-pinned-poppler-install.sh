#!/usr/bin/env bash
set -euo pipefail

readonly script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/install-pinned-poppler.sh"
readonly temp="$(mktemp -d)"
trap 'rm -rf -- "$temp"' EXIT
readonly mock_bin="$temp/bin"
mkdir -p "$mock_bin"
readonly package_file='poppler-utils_24.02.0-1ubuntu9.10_amd64.deb'
readonly metadata_path="pool/main/p/poppler/$package_file"
readonly official_uri="http://archive.ubuntu.com/ubuntu/$metadata_path"
readonly deb_bytes="$temp/pinned-poppler.deb"
printf 'synthetic deb bytes for provenance tests\n' >"$deb_bytes"
readonly deb_sha256="$(sha256sum "$deb_bytes" | awk '{print $1}')"

cat >"$mock_bin/sudo" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
exec "$@"
MOCK
cat >"$mock_bin/apt-get" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  update) exit 0 ;;
  --print-uris)
    printf "  '%s' %s 1 SHA256:%s\n" "$MOCK_URI" "$(basename "$MOCK_URI")" "$MOCK_SHA256"
    ;;
  download)
    cp "$MOCK_DEB" "$(basename "$MOCK_URI")"
    ;;
  install)
    [[ -f "${@: -1}" ]]
    ;;
  *)
    printf 'unexpected apt-get invocation: %s\n' "$*" >&2
    exit 1
    ;;
esac
MOCK
cat >"$mock_bin/apt-cache" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == show ]]
cat "$MOCK_METADATA"
MOCK
cat >"$mock_bin/apt" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == install && "$2" == -y && -f "${@: -1}" ]]
printf '%s\n' "$*" >"$MOCK_APT_INSTALL_INVOCATION"
MOCK
cat >"$mock_bin/dpkg" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == --print-architecture ]]
printf '%s\n' amd64
MOCK
cat >"$mock_bin/dpkg-deb" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == --field ]]
case "$3" in
  Version) printf '%s\n' "${MOCK_DEB_VERSION:-24.02.0-1ubuntu9.10}" ;;
  Architecture) printf '%s\n' "${MOCK_DEB_ARCHITECTURE:-amd64}" ;;
  *) exit 2 ;;
esac
MOCK
chmod +x "$mock_bin"/*

write_metadata() {
  local version="${1:-24.02.0-1ubuntu9.10}" architecture="${2:-amd64}" digest="${3:-$deb_sha256}" filename="${4:-$metadata_path}"
  cat >"$temp/Packages" <<EOF
Package: poppler-utils
Version: $version
Architecture: $architecture
Filename: $filename
SHA256: $digest
EOF
}

run_installer() {
  local summary="$temp/summary.md"
  PATH="$mock_bin:/usr/bin:/bin" \
    MOCK_DEB="$deb_bytes" MOCK_URI="${MOCK_URI:-$official_uri}" \
    MOCK_SHA256="${MOCK_SHA256:-$deb_sha256}" MOCK_METADATA="$temp/Packages" \
    MOCK_DEB_VERSION="${MOCK_DEB_VERSION:-24.02.0-1ubuntu9.10}" \
    MOCK_DEB_ARCHITECTURE="${MOCK_DEB_ARCHITECTURE:-amd64}" \
    MOCK_APT_INSTALL_INVOCATION="$temp/apt-install-invocation" \
    GITHUB_STEP_SUMMARY="$summary" bash "$script"
}

expect_rejected() {
  local name="$1"
  if run_installer >"$temp/$name.out" 2>"$temp/$name.err"; then
    printf 'pinned Poppler installer unexpectedly accepted %s\n' "$name" >&2
    exit 1
  fi
}

write_metadata
run_installer
grep -Fq -- '- Version: `24.02.0-1ubuntu9.10`' "$temp/summary.md"
grep -Fq -- '- Architecture: `amd64`' "$temp/summary.md"
grep -Fq -- "- Authenticated Ubuntu archive URI: \`$official_uri\`" "$temp/summary.md"
grep -Fq -- "- SHA-256 (matches signed APT package metadata): \`$deb_sha256\`" "$temp/summary.md"
grep -Fq 'install -y ' "$temp/apt-install-invocation"

MOCK_URI="http://azure.archive.ubuntu.com/ubuntu/$metadata_path"
export MOCK_URI
write_metadata
run_installer
grep -Fq -- "- Authenticated Ubuntu archive URI: \`$MOCK_URI\`" "$temp/summary.md"
unset MOCK_URI

write_metadata '24.02.0-1ubuntu9.9'
expect_rejected wrong-version-metadata
write_metadata '24.02.0-1ubuntu9.10' amd64 "$(printf wrong | sha256sum | awk '{print $1}')"
expect_rejected wrong-metadata-digest
write_metadata '24.02.0-1ubuntu9.10' arm64
expect_rejected wrong-metadata-architecture
write_metadata
MOCK_URI='http://mirror.example.invalid/ubuntu/pool/main/p/poppler/poppler-utils_24.02.0-1ubuntu9.10_amd64.deb'
export MOCK_URI
expect_rejected non-ubuntu-package-source
unset MOCK_URI
write_metadata
MOCK_DEB_ARCHITECTURE=arm64
export MOCK_DEB_ARCHITECTURE
expect_rejected wrong-downloaded-architecture
unset MOCK_DEB_ARCHITECTURE

printf 'pinned Poppler acquisition provenance tests passed\n'
