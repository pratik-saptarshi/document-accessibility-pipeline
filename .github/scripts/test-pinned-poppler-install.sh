#!/usr/bin/env bash
set -euo pipefail

readonly script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/install-pinned-poppler.sh"
readonly source_test="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/test-ubuntu-apt-provenance.sh"
readonly temp="$(mktemp -d)"
trap 'rm -rf -- "$temp"' EXIT
readonly mock_bin="$temp/bin"
mkdir -p "$mock_bin"
readonly package_file='poppler-utils_24.02.0-1ubuntu9.10_amd64.deb'
readonly metadata_path="pool/main/p/poppler/$package_file"
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
  download)
    cp "$MOCK_DEB" "$MOCK_DOWNLOAD_FILENAME"
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
case "$1" in
  show) cat "$MOCK_METADATA" ;;
  policy) cat "$MOCK_POLICY" ;;
  *) exit 2 ;;
esac
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
  Package) printf '%s\n' "${MOCK_DEB_PACKAGE:-poppler-utils}" ;;
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

write_runner_sources() {
  cat >"$temp/ubuntu.sources" <<'SOURCES'
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
  printf 'http://azure.archive.ubuntu.com/ubuntu/\tpriority:1\nhttps://archive.ubuntu.com/ubuntu/\tpriority:2\nhttps://security.ubuntu.com/ubuntu/\tpriority:3\n' >"$temp/apt-mirrors.txt"
  cat >"$temp/policy" <<'POLICY'
poppler-utils:
  Installed: (none)
  Candidate: 24.02.0-1ubuntu9.10
  Version table:
     24.02.0-1ubuntu9.10 500
        500 http://azure.archive.ubuntu.com/ubuntu noble-updates/universe amd64 Packages
           release v=24.04,o=Ubuntu,a=noble-updates,n=noble,l=Ubuntu,c=universe,b=amd64
           origin azure.archive.ubuntu.com
POLICY
}

run_installer() {
  local package_filename="$(basename "$metadata_path")"
  local summary="$temp/summary.md"
  PATH="$mock_bin:/usr/bin:/bin" \
    MOCK_DEB="${MOCK_DEB:-$deb_bytes}" MOCK_DOWNLOAD_FILENAME="$package_filename" \
    MOCK_METADATA="$temp/Packages" MOCK_POLICY="$temp/policy" \
    MOCK_DEB_VERSION="${MOCK_DEB_VERSION:-24.02.0-1ubuntu9.10}" \
    MOCK_DEB_ARCHITECTURE="${MOCK_DEB_ARCHITECTURE:-amd64}" \
    MOCK_APT_INSTALL_INVOCATION="$temp/apt-install-invocation" \
    POPPLER_APT_SOURCES_FILE="$temp/ubuntu.sources" POPPLER_APT_MIRRORS_FILE="$temp/apt-mirrors.txt" \
    GITHUB_STEP_SUMMARY="$summary" bash "$script"
}

expect_rejected() {
  local name="$1"
  if run_installer >"$temp/$name.out" 2>"$temp/$name.err"; then
    printf 'pinned Poppler installer unexpectedly accepted %s\n' "$name" >&2
    exit 1
  fi
}

expect_error() {
  local name="$1" message="$2"
  grep -Fq -- "$message" "$temp/$name.err" || {
    printf 'pinned Poppler installer did not report the expected %s failure\n' "$name" >&2
    cat "$temp/$name.err" >&2
    exit 1
  }
}

write_metadata
write_runner_sources
run_installer
grep -Fq -- '- Version: `24.02.0-1ubuntu9.10`' "$temp/summary.md"
grep -Fq -- '- Architecture: `amd64`' "$temp/summary.md"
grep -Fq -- 'mirror+file:/etc/apt/apt-mirrors.txt' "$temp/summary.md"
grep -Fq -- 'http://azure.archive.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'https://archive.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'https://security.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'serving candidate not inferred' "$temp/summary.md"
grep -Fq -- "- SHA-256 (matches signed APT package metadata): \`$deb_sha256\`" "$temp/summary.md"
grep -Fq 'install -y ' "$temp/apt-install-invocation"

write_metadata '24.02.0-1ubuntu9.9'
expect_rejected wrong-version-metadata
write_metadata '24.02.0-1ubuntu9.10' amd64 "$(printf wrong | sha256sum | awk '{print $1}')"
expect_rejected wrong-metadata-digest
write_metadata '24.02.0-1ubuntu9.10' arm64
expect_rejected wrong-metadata-architecture
write_metadata
write_metadata '24.02.0-1ubuntu9.10' amd64 "$deb_sha256" pool/main/p/poppler/unexpected.deb
expect_rejected wrong-metadata-filename
write_metadata
printf 'SHA256: %s\n' "$deb_sha256" >>"$temp/Packages"
expect_rejected duplicate-metadata-field
write_metadata
write_runner_sources
printf 'https://mirror.example.invalid/ubuntu/\tpriority:4\n' >>"$temp/apt-mirrors.txt"
expect_rejected untrusted-mirror-candidate
write_runner_sources
MOCK_DEB="$temp/tampered.deb"
printf 'tampered package bytes\n' >"$MOCK_DEB"
export MOCK_DEB
expect_rejected downloaded-package-digest-mismatch
unset MOCK_DEB
write_metadata
write_runner_sources
MOCK_DEB_VERSION='24.02.0-1ubuntu9.9'
export MOCK_DEB_VERSION
expect_rejected wrong-downloaded-version
unset MOCK_DEB_VERSION
MOCK_DEB_PACKAGE='untrusted-poppler-utils'
export MOCK_DEB_PACKAGE
expect_rejected wrong-downloaded-package-name
unset MOCK_DEB_PACKAGE
MOCK_DEB_ARCHITECTURE=arm64
export MOCK_DEB_ARCHITECTURE
expect_rejected wrong-downloaded-architecture
unset MOCK_DEB_ARCHITECTURE

bash "$source_test"

printf 'pinned Poppler acquisition provenance tests passed\n'
