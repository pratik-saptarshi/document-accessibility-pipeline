#!/usr/bin/env bash
set -euo pipefail

readonly script="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/install-pinned-poppler.sh"
readonly source_test="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/test-ubuntu-apt-provenance.sh"
readonly workflow="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/workflows/go-checks.yml"
readonly package_spec='poppler-utils:amd64=24.02.0-1ubuntu9.11'
readonly temp="$(mktemp -d)"
trap 'rm -rf -- "$temp"' EXIT
readonly mock_bin="$temp/bin"
mkdir -p "$mock_bin"
readonly package_file='poppler-utils_24.02.0-1ubuntu9.11_amd64.deb'
readonly metadata_path="pool/main/p/poppler/$package_file"
readonly deb_bytes="$temp/pinned-poppler.deb"
printf 'synthetic deb bytes for provenance tests\n' >"$deb_bytes"
readonly deb_sha256="$(sha256sum "$deb_bytes" | awk '{print $1}')"

cat >"$mock_bin/sudo" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
printf 'sudo %s\n' "$*" >>"$MOCK_SUDO_CALLS"
if [[ "${1:-}" == install ]]; then
  shift
  mode= path=''
  while (($#)); do
    case "$1" in
      -d) shift ;;
      -o|-g|-m) [[ "$1" == -m ]] && mode="$2"; shift 2 ;;
      *) path="$1"; shift ;;
    esac
  done
  mkdir -p "$path"
  chmod "$mode" "$path"
  exit 0
fi
exec "$@"
MOCK
cat >"$mock_bin/apt-get" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
printf 'apt-get %s\n' "$*" >>"$MOCK_APT_CALLS"
args="$*"
[[ "$args" == *"Dir::Etc::sourcelist=$MOCK_EXPECTED_SOURCE"* && "$args" == *'Dir::Etc::sourceparts=-'* && "$args" == *'Dir::State::lists='* ]] || {
  printf 'apt-get was not bound to the isolated Ubuntu source context\n' >&2
  exit 1
}
for option in "$@"; do
  [[ "$option" == Dir::State::lists=* ]] && lists_dir="${option#*=}"
done
[[ -n "${lists_dir:-}" && "$(stat -c '%a' "${lists_dir%/lists}")" == 711 && "$(stat -c '%a' "$lists_dir")" == 755 && "$(stat -c '%a' "$lists_dir/partial")" == 700 ]] || {
  printf 'isolated APT state does not provide least-privilege traversal for _apt\n' >&2
  exit 1
}
while [[ "${1:-}" == -o ]]; do shift 2; done
case "${1:-}" in
  update)
    if [[ "${MOCK_APT_GET_UPDATE_STATUS:-0}" != 0 ]]; then
      printf 'mock apt-get update failed\n' >&2
      exit "$MOCK_APT_GET_UPDATE_STATUS"
    fi
    exit 0
    ;;
  download)
    [[ "${2:-}" == "$MOCK_EXPECTED_PACKAGE_SPEC" ]] || {
      printf 'unexpected apt-get download operand: %s\n' "${2:-}" >&2
      exit 1
    }
    if [[ "${MOCK_APT_GET_DOWNLOAD_STATUS:-0}" != 0 ]]; then
      printf 'mock apt-get download failed\n' >&2
      exit "$MOCK_APT_GET_DOWNLOAD_STATUS"
    fi
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
printf 'apt-cache %s\n' "$*" >>"$MOCK_APT_CALLS"
args="$*"
[[ "$args" == *"Dir::Etc::sourcelist=$MOCK_EXPECTED_SOURCE"* && "$args" == *'Dir::Etc::sourceparts=-'* && "$args" == *'Dir::State::lists='* ]] || {
  printf 'apt-cache was not bound to the isolated Ubuntu source context\n' >&2
  exit 1
}
for option in "$@"; do
  [[ "$option" == Dir::State::lists=* ]] && lists_dir="${option#*=}"
done
[[ -n "${lists_dir:-}" && "$(stat -c '%a' "${lists_dir%/lists}")" == 711 && "$(stat -c '%a' "$lists_dir")" == 755 && "$(stat -c '%a' "$lists_dir/partial")" == 700 ]] || {
  printf 'isolated APT state does not provide least-privilege traversal for _apt\n' >&2
  exit 1
}
while [[ "${1:-}" == -o ]]; do shift 2; done
case "$1" in
  show)
    [[ "${2:-}" == "$MOCK_EXPECTED_PACKAGE_SPEC" ]] || {
      printf 'unexpected apt-cache show operand: %s\n' "${2:-}" >&2
      exit 1
    }
    if [[ -n "${MOCK_APT_CACHE_SHOW_OUTPUT+x}" ]]; then
      printf '%s' "$MOCK_APT_CACHE_SHOW_OUTPUT"
      exit "${MOCK_APT_CACHE_SHOW_STATUS:-0}"
    fi
    cat "$MOCK_METADATA"
    exit "${MOCK_APT_CACHE_SHOW_STATUS:-0}"
    ;;
  policy)
    [[ "${2:-}" == poppler-utils ]] || {
      printf 'unexpected apt-cache policy operand: %s\n' "${2:-}" >&2
      exit 1
    }
    cat "$MOCK_POLICY"
    exit "${MOCK_APT_CACHE_POLICY_STATUS:-0}"
    ;;
  *) exit 2 ;;
esac
MOCK
cat >"$mock_bin/apt" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
printf 'apt %s\n' "$*" >>"$MOCK_APT_CALLS"
args="$*"
[[ "$args" == *"Dir::Etc::sourcelist=$MOCK_EXPECTED_SOURCE"* && "$args" == *'Dir::Etc::sourceparts=-'* && "$args" == *'Dir::State::lists='* ]] || {
  printf 'apt install was not bound to the isolated Ubuntu source context\n' >&2
  exit 1
}
for option in "$@"; do
  [[ "$option" == Dir::State::lists=* ]] && lists_dir="${option#*=}"
done
[[ -n "${lists_dir:-}" && "$(stat -c '%a' "${lists_dir%/lists}")" == 711 && "$(stat -c '%a' "$lists_dir")" == 755 && "$(stat -c '%a' "$lists_dir/partial")" == 700 ]] || {
  printf 'apt install does not use the isolated APT state created for _apt\n' >&2
  exit 1
}
while [[ "${1:-}" == -o ]]; do shift 2; done
[[ "$1" == install && "$2" == -y && "$(basename "${@: -1}")" == "$MOCK_EXPECTED_DEB_FILENAME" && -f "${@: -1}" ]]
if [[ "${MOCK_APT_INSTALL_STATUS:-0}" != 0 ]]; then
  printf 'mock apt install failed\n' >&2
fi
printf '%s\n' "$*" >"$MOCK_APT_INSTALL_INVOCATION"
exit "${MOCK_APT_INSTALL_STATUS:-0}"
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
  Version) printf '%s\n' "${MOCK_DEB_VERSION:-24.02.0-1ubuntu9.11}" ;;
  Architecture) printf '%s\n' "${MOCK_DEB_ARCHITECTURE:-amd64}" ;;
  *) exit 2 ;;
esac
MOCK
chmod +x "$mock_bin"/*

write_metadata() {
  local version="${1:-24.02.0-1ubuntu9.11}" architecture="${2:-amd64}" digest="${3:-$deb_sha256}" filename="${4:-$metadata_path}"
  cat >"$temp/Packages" <<EOF
Package: poppler-utils
Version: $version
Architecture: $architecture
Filename: $filename
SHA256: $digest
Size: 123456
Description: PDF utilities
 Command line tools for extracting text and metadata from PDF files.
Homepage: https://poppler.freedesktop.org/
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
  Candidate: 24.02.0-1ubuntu9.11
  Version table:
     24.02.0-1ubuntu9.11 500
        500 http://azure.archive.ubuntu.com/ubuntu noble-updates/universe amd64 Packages
POLICY
}

run_installer() {
  local package_filename="$(basename "$metadata_path")"
  local summary="$temp/summary.md"
  rm -f "$temp/apt-calls" "$temp/sudo-calls" "$temp/apt-install-invocation" "$summary"
  PATH="$mock_bin:/usr/bin:/bin" \
    MOCK_DEB="${MOCK_DEB:-$deb_bytes}" MOCK_DOWNLOAD_FILENAME="$package_filename" \
    MOCK_METADATA="$temp/Packages" MOCK_POLICY="$temp/policy" \
    MOCK_DEB_VERSION="${MOCK_DEB_VERSION:-24.02.0-1ubuntu9.11}" \
    MOCK_DEB_ARCHITECTURE="${MOCK_DEB_ARCHITECTURE:-amd64}" \
    MOCK_EXPECTED_PACKAGE_SPEC="$package_spec" \
    MOCK_EXPECTED_DEB_FILENAME="$(basename "$metadata_path")" \
    MOCK_APT_INSTALL_INVOCATION="$temp/apt-install-invocation" \
    MOCK_APT_CALLS="$temp/apt-calls" \
    MOCK_SUDO_CALLS="$temp/sudo-calls" \
    MOCK_EXPECTED_SOURCE="$temp/ubuntu.sources" \
    POPPLER_APT_SOURCES_FILE="$temp/ubuntu.sources" POPPLER_APT_MIRRORS_FILE="$temp/apt-mirrors.txt" \
    GITHUB_STEP_SUMMARY="$summary" bash "$script"
}

expect_rejected() {
  local name="$1"
  if run_installer >"$temp/$name.out" 2>"$temp/$name.err"; then
    printf 'pinned Poppler installer unexpectedly accepted %s\n' "$name" >&2
    exit 1
  fi
  if [[ -s "$temp/summary.md" ]]; then
    printf 'pinned Poppler installer wrote a success summary after rejecting %s\n' "$name" >&2
    return 1
  fi
  assert_cleanup_completed
}

expect_error() {
  local name="$1" message="$2"
  grep -Fq -- "$message" "$temp/$name.err" || {
    printf 'pinned Poppler installer did not report the expected %s failure\n' "$name" >&2
    cat "$temp/$name.err" >&2
    exit 1
  }
}

assert_workflow_order() {
  local inventory expected
  inventory="$(awk '
    $0 == "      - name: Test pinned Poppler acquisition provenance" {
      if (getline <= 0 || $0 != "        run: bash .github/scripts/test-pinned-poppler-install.sh") exit 1
      print "test"
    }
    $0 == "      - name: Install poppler-utils" {
      if (getline <= 0 || $0 != "        run: bash .github/scripts/install-pinned-poppler.sh") exit 1
      print "install"
    }
  ' "$workflow")" || {
    printf 'Poppler workflow step names must use their expected commands\n' >&2
    exit 1
  }
  expected=$'test\ninstall'
  [[ "$inventory" == "$expected" ]] || {
    printf 'workflow must run deterministic Poppler provenance tests once before live installation\n' >&2
    exit 1
  }
}

assert_no_acquisition_after_metadata_query() {
  assert_call_absent '^apt-cache .* policy ' 'APT policy query' || return 1
  assert_call_absent '^apt-get .* download ' 'APT package download' || return 1
  assert_call_absent '^apt .* install ' 'APT package installation' || return 1
  if [[ -e "$temp/apt-install-invocation" ]]; then
    printf 'unexpected APT install invocation marker exists\n' >&2
    return 1
  fi
  if [[ -s "$temp/summary.md" ]]; then
    printf 'unexpected success summary exists\n' >&2
    return 1
  fi
}

assert_call_present() {
  local pattern="$1" label="$2" status
  if grep -Eq -- "$pattern" "$temp/apt-calls"; then
    return 0
  else
    status=$?
  fi
  printf 'expected call trace missing for %s (grep status %s)\n' "$label" "$status" >&2
  return 1
}

assert_call_absent() {
  local pattern="$1" label="$2" status
  if grep -Eq -- "$pattern" "$temp/apt-calls"; then
    printf 'unexpected call trace for %s\n' "$label" >&2
    return 1
  else
    status=$?
  fi
  if [[ "$status" != 1 ]]; then
    printf 'could not inspect call trace for %s (grep status %s)\n' "$label" "$status" >&2
    return 1
  fi
}

assert_cleanup_completed() {
  local cleanup_path status
  if cleanup_path="$(awk '
    /^sudo rm -rf -- / { count++; path = substr($0, length("sudo rm -rf -- ") + 1) }
    END { if (count != 1) exit 1; print path }
  ' "$temp/sudo-calls")"; then
    :
  else
    status=$?
    printf 'expected exactly one isolated APT cleanup invocation (status %s)\n' "$status" >&2
    return 1
  fi
  if [[ -z "$cleanup_path" || -e "$cleanup_path" ]]; then
    printf 'isolated APT directory was not removed: %s\n' "$cleanup_path" >&2
    return 1
  fi
}

# Mutation guard: each forbidden call with APT options must be caught even
# when the assertion helper is invoked in a conditional context.
while IFS= read -r forbidden_call; do
  printf '%s\n' "$forbidden_call" >"$temp/apt-calls"
  if assert_no_acquisition_after_metadata_query >"$temp/assertion.out" 2>"$temp/assertion.err"; then
    printf 'forbidden APT call mutation was accepted: %s\n' "$forbidden_call" >&2
    exit 1
  fi
done <<'FORBIDDEN_CALLS'
apt-cache -o Dir::State::lists=/tmp/lists policy poppler-utils
apt-get -o Dir::State::lists=/tmp/lists download poppler-utils:amd64=24.02.0-1ubuntu9.11
apt -o Dir::State::lists=/tmp/lists install -y /tmp/poppler.deb
FORBIDDEN_CALLS
: >"$temp/apt-calls"

write_metadata
write_runner_sources
assert_workflow_order
run_installer
grep -Fq -- '- Version: `24.02.0-1ubuntu9.11`' "$temp/summary.md"
grep -Fq -- '- Architecture: `amd64`' "$temp/summary.md"
grep -Fq -- 'mirror+file:/etc/apt/apt-mirrors.txt' "$temp/summary.md"
grep -Fq -- 'http://azure.archive.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'https://archive.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'https://security.ubuntu.com/ubuntu/' "$temp/summary.md"
grep -Fq -- 'serving candidate not inferred' "$temp/summary.md"
grep -Fq -- "- SHA-256 (matches signed APT package metadata): \`$deb_sha256\`" "$temp/summary.md"
grep -Fq 'install -y ' "$temp/apt-install-invocation"
grep -Fq 'apt -o ' "$temp/apt-calls"
grep -Fq 'Dir::Etc::sourcelist='"$temp/ubuntu.sources" "$temp/apt-calls"
grep -Fq 'Dir::Etc::sourceparts=-' "$temp/apt-calls"
grep -Fq 'Dir::State::lists=' "$temp/apt-calls"
grep -Fq 'apt-get -o ' "$temp/apt-calls"
grep -Fq 'apt-cache -o ' "$temp/apt-calls"
grep -Fq 'sudo install -d -o _apt -g root -m 0700 ' "$temp/sudo-calls"
assert_cleanup_completed
grep -Fq 'show poppler-utils:amd64=24.02.0-1ubuntu9.11' "$temp/apt-calls"
grep -Fq 'policy poppler-utils' "$temp/apt-calls"
grep -Fq 'download poppler-utils:amd64=24.02.0-1ubuntu9.11' "$temp/apt-calls"

# A successful exact-pin query with no records is different from an APT
# command failure and must stop before policy, download, install, or summary.
: >"$temp/Packages"
cat >"$temp/policy" <<'POLICY'
poppler-utils:
  Installed: 24.02.0-1ubuntu9.11
  Candidate: 24.02.0-1ubuntu9.11
POLICY
MOCK_APT_CACHE_SHOW_OUTPUT=''
export MOCK_APT_CACHE_SHOW_OUTPUT
expect_rejected empty-metadata
expect_error empty-metadata 'no authenticated APT metadata returned for exact package pin poppler-utils:amd64=24.02.0-1ubuntu9.11'
grep -Fq 'apt-cache ' "$temp/apt-calls"
grep -Fq 'show poppler-utils:amd64=24.02.0-1ubuntu9.11' "$temp/apt-calls"
assert_no_acquisition_after_metadata_query

MOCK_APT_CACHE_SHOW_OUTPUT=$' \n\t\n'
export MOCK_APT_CACHE_SHOW_OUTPUT
expect_rejected whitespace-metadata
expect_error whitespace-metadata 'no authenticated APT metadata returned for exact package pin poppler-utils:amd64=24.02.0-1ubuntu9.11'
assert_no_acquisition_after_metadata_query
unset MOCK_APT_CACHE_SHOW_OUTPUT
write_metadata
write_runner_sources
run_installer >/dev/null

MOCK_APT_CACHE_SHOW_OUTPUT=''
MOCK_APT_CACHE_SHOW_STATUS=100
export MOCK_APT_CACHE_SHOW_OUTPUT MOCK_APT_CACHE_SHOW_STATUS
expect_rejected apt-cache-query-failure
expect_error apt-cache-query-failure 'apt-cache show failed for exact package pin poppler-utils:amd64=24.02.0-1ubuntu9.11'
assert_no_acquisition_after_metadata_query
assert_cleanup_completed
unset MOCK_APT_CACHE_SHOW_OUTPUT MOCK_APT_CACHE_SHOW_STATUS

# Source update failure is surfaced before metadata query or download.
MOCK_APT_GET_UPDATE_STATUS=42
export MOCK_APT_GET_UPDATE_STATUS
expect_rejected update-failure
expect_error update-failure 'mock apt-get update failed'
assert_call_present '^apt-get .* update$' 'APT source update'
assert_call_absent '^apt-cache ' 'APT metadata query after failed update'
assert_call_absent '^apt-get .* download ' 'APT package download after failed update'
assert_cleanup_completed
unset MOCK_APT_GET_UPDATE_STATUS

# Metadata structure remains strict, including repeated records and missing
# required fields; descriptions in valid stanzas may contain indented lines.
write_metadata
cat "$temp/Packages" >>"$temp/Packages.duplicate"
cat "$temp/Packages.duplicate" >>"$temp/Packages"
expect_rejected duplicate-metadata-stanza
expect_error duplicate-metadata-stanza 'expected exactly one Package field in authenticated APT metadata'
assert_call_present '^apt-cache .* show ' 'APT metadata query for duplicate stanza'
assert_call_absent '^apt-cache .* policy ' 'APT policy query after duplicate stanza'
write_metadata
sed -i '/^SHA256:/d' "$temp/Packages"
expect_rejected missing-metadata-digest
expect_error missing-metadata-digest 'expected exactly one SHA256 field in authenticated APT metadata'
assert_call_present '^apt-cache .* show ' 'APT metadata query for missing digest'
assert_call_absent '^apt-cache .* policy ' 'APT policy query after missing digest'
write_metadata

# Download and installation failures stop at their stage and still clean up.
MOCK_APT_GET_DOWNLOAD_STATUS=43
export MOCK_APT_GET_DOWNLOAD_STATUS
expect_rejected download-failure
expect_error download-failure 'mock apt-get download failed'
assert_call_present '^apt-cache .* policy ' 'APT policy query before download failure'
assert_call_present '^apt-get .* download poppler-utils:amd64=24\.02\.0-1ubuntu9\.11$' 'exact pinned package download'
assert_call_absent '^apt .* install ' 'APT install after download failure'
assert_cleanup_completed
unset MOCK_APT_GET_DOWNLOAD_STATUS

MOCK_APT_INSTALL_STATUS=44
export MOCK_APT_INSTALL_STATUS
expect_rejected install-failure
expect_error install-failure 'mock apt install failed'
assert_call_present '^apt-get .* download poppler-utils:amd64=24\.02\.0-1ubuntu9\.11$' 'package download before install failure'
assert_call_present '^apt .* install -y ' 'APT install invocation after validated download'
[[ -s "$temp/apt-install-invocation" ]]
assert_cleanup_completed
[[ ! -s "$temp/summary.md" ]]
unset MOCK_APT_INSTALL_STATUS

# Invalid signing configuration is rejected before even apt-get update runs.
: >"$temp/apt-calls"
sed -i 's|Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg|Signed-By: /tmp/untrusted-keyring.gpg|' "$temp/ubuntu.sources"
expect_rejected wrong-keyring-before-update
[[ ! -s "$temp/apt-calls" ]]
sed -i 's|Signed-By: /tmp/untrusted-keyring.gpg|Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg|' "$temp/ubuntu.sources"

# The policy URI must be one of the independently validated Ubuntu mirror
# candidates; copied Ubuntu Release labels from another source are insufficient.
cp "$temp/policy" "$temp/policy.good"
sed -i 's|http://azure.archive.ubuntu.com/ubuntu|https://evil.example.invalid/ubuntu|g; s|origin azure.archive.ubuntu.com|origin evil.example.invalid|' "$temp/policy"
expect_rejected foreign-policy-source
mv "$temp/policy.good" "$temp/policy"

write_metadata '24.02.0-1ubuntu9.9'
expect_rejected wrong-version-metadata
write_metadata '24.02.0-1ubuntu9.11' amd64 "$deb_sha256" "$metadata_path"
sed -i 's/^Package: poppler-utils$/Package: other-package/' "$temp/Packages"
expect_rejected wrong-package-metadata
write_metadata
sed -i '/^Architecture:/d' "$temp/Packages"
expect_rejected missing-metadata-architecture
write_metadata
sed -i 's/^SHA256:.*/SHA256: not-a-digest/' "$temp/Packages"
expect_rejected malformed-metadata-digest
write_metadata '24.02.0-1ubuntu9.11' amd64 "$(printf wrong | sha256sum | awk '{print $1}')"
expect_rejected wrong-metadata-digest
write_metadata '24.02.0-1ubuntu9.11' arm64
expect_rejected wrong-metadata-architecture
write_metadata
write_metadata '24.02.0-1ubuntu9.11' amd64 "$deb_sha256" pool/main/p/poppler/unexpected.deb
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
