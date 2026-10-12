#!/usr/bin/env bash
set -euo pipefail

# Minimal first-stage bootstrap. The Go orchestrator performs all remaining
# installations from buildtools.manifest.json. Nothing is installed system-wide.
readonly go_version='1.27.2'
readonly go_archive='go1.27.2.linux-amd64.tar.gz'
readonly go_sha256='ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5'
readonly go_url="https://go.dev/dl/${go_archive}"
# The source archive is pinned in buildtools.manifest.json and compiled locally.
# This keeps no-checkout installs independent of Go module proxy/tag behavior.
readonly bootstrap_source_revision='01e5395ee7d596a7d75326256098a70bddaac133'
readonly bootstrap_source_url="https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/${bootstrap_source_revision}.tar.gz"
readonly bootstrap_source_sha256='e9da46eba76e7891a9d5c12fe3953234c764dcc09c2e250791327bbb2e62da16'

script_dir="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(CDPATH= cd -- "${script_dir}/.." && pwd)"
tools_root="${CODEX_BUILD_TOOLS_ROOT:-/workspace/.codex-tools/cloud-build-tools}"
cache_root="${CODEX_BUILD_CACHE_ROOT:-/workspace/.cache/cloud-build-tools}"
go_root="${tools_root}/toolchains/go/${go_version}"

if [[ "$(uname -s)" != Linux || "$(uname -m)" != x86_64 ]]; then
  printf 'Unsupported host %s/%s; this manifest currently supports linux/amd64 only.\n' "$(uname -s)" "$(uname -m)" >&2
  exit 2
fi
for utility in curl sha256sum tar timeout; do
  command -v "${utility}" >/dev/null || { printf 'Required bootstrap utility missing: %s\n' "${utility}" >&2; exit 2; }
done
mkdir -p "${tools_root}/toolchains/go" "${cache_root}/downloads"
chmod 700 "${tools_root}" "${cache_root}"

# mkdir is an atomic lock. A timed-out/stale lock is reported for inspection;
# this script never removes a lock owned by a possibly running bootstrap.
lock="${tools_root}/toolchains/go/.bootstrap.lock"
deadline=$((SECONDS + 120))
until mkdir "${lock}" 2>/dev/null; do
  if (( SECONDS >= deadline )); then
    printf 'Timed out waiting for %s; check for a live bootstrap before removing it.\n' "${lock}" >&2
    exit 1
  fi
  sleep 1
done
printf '%s\n' "$$" > "${lock}/pid"
cleanup() { rm -rf -- "${lock}"; }
part=''
tmp=''
cleanup_part() { [[ -z "${part}" ]] || rm -f -- "${part}"; }
cleanup_tmp() { [[ -z "${tmp}" ]] || rm -rf -- "${tmp}"; }
trap 'cleanup_tmp; cleanup_part; cleanup' EXIT INT TERM

go_bin="${go_root}/bin/go"
if [[ -x "${go_bin}" ]]; then
  if "${go_bin}" version | grep -Fq "go${go_version} "; then
    printf 'Verified existing Go toolchain %s\n' "${go_version}"
  else
    printf 'Existing Go path has an unexpected version: %s\n' "${go_bin}" >&2
    exit 1
  fi
else
  tmp="$(mktemp -d "${tools_root}/toolchains/go/.stage-${go_version}.XXXXXX")"
  trap 'cleanup_tmp; cleanup_part; cleanup' EXIT INT TERM
  archive="${cache_root}/downloads/${go_archive}"
  part="${archive}.part.$$"
  rm -f -- "${part}"
  curl --fail --location --silent --show-error "${go_url}" --output "${part}"
  printf '%s  %s\n' "${go_sha256}" "${part}" | sha256sum --check --status || {
    rm -f -- "${part}"
    printf 'Go archive checksum mismatch; no toolchain was installed.\n' >&2
    exit 1
  }
  mv -- "${part}" "${archive}"
  part=''
  mkdir "${tmp}/unpacked"
  tar -xzf "${archive}" -C "${tmp}/unpacked" --strip-components=1
  [[ -x "${tmp}/unpacked/bin/go" ]] || { printf 'Go archive did not contain bin/go.\n' >&2; exit 1; }
  "${tmp}/unpacked/bin/go" version | grep -Fq "go${go_version} " || {
    printf 'Go archive version did not match pinned version %s.\n' "${go_version}" >&2
    exit 1
  }
  mv -- "${tmp}/unpacked" "${go_root}"
  printf 'Installed Go toolchain %s in %s\n' "${go_version}" "${go_root}"
fi

mkdir -p "${cache_root}/go/mod" "${cache_root}/go/build" "${cache_root}/go/gopath" "${tools_root}/bin"
bootstrap_status=0
if [[ -f "${repo_root}/cmd/cloud-tools-bootstrap/go.mod" && -f "${repo_root}/cmd/cloud-tools-bootstrap/main.go" ]]; then
  # Local checkout path for review and maintenance. Saved environment Install
  # scripts run the immutable module branch below and need no repository clone.
  if (
    cd "${repo_root}/cmd/cloud-tools-bootstrap"
    GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
    GOMODCACHE="${cache_root}/go/mod" GOCACHE="${cache_root}/go/build" \
    "${go_bin}" run . \
      --manifest "${repo_root}/buildtools.manifest.json" \
      --root "${tools_root}" --cache "${cache_root}" --go "${go_bin}"
  ); then :; else bootstrap_status=$?; fi
else
  source_archive="${cache_root}/downloads/cloud-tools-bootstrap-${bootstrap_source_revision}.tar.gz"
  if [[ ! -f "${source_archive}" ]] || ! printf '%s  %s\n' "${bootstrap_source_sha256}" "${source_archive}" | sha256sum --check --status; then
    part="${source_archive}.part.$$"
    rm -f -- "${part}"
    if ! curl --fail --location --silent --show-error "${bootstrap_source_url}" --output "${part}"; then
      printf 'Could not retrieve the pinned bootstrap source archive.\n' >&2
      exit 1
    fi
    if ! printf '%s  %s\n' "${bootstrap_source_sha256}" "${part}" | sha256sum --check --status; then
      printf 'Bootstrap source checksum mismatch; no bootstrap command was built.\n' >&2
      exit 1
    fi
    mv -- "${part}" "${source_archive}"
    part=''
  fi
  tmp="$(mktemp -d "${cache_root}/.bootstrap-source.XXXXXX")"
  mkdir "${tmp}/source" "${tmp}/bin"
  tar -xzf "${source_archive}" -C "${tmp}/source" --strip-components=1
  if ! timeout 8m bash -c 'cd "$1/cmd/cloud-tools-bootstrap" && GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly GOMODCACHE="$2/go/mod" GOCACHE="$2/go/build" "$3" build -buildvcs=false -trimpath -o "$4/cloud-tools-bootstrap" .' _ "${tmp}/source" "${cache_root}" "${go_bin}" "${tmp}/bin"; then
    printf 'Could not build the checksum-verified bootstrap source.\n' >&2
    exit 1
  fi
  chmod 0755 "${tmp}/bin/cloud-tools-bootstrap"
  mv -f -- "${tmp}/bin/cloud-tools-bootstrap" "${tools_root}/bin/cloud-tools-bootstrap"
  if "${tools_root}/bin/cloud-tools-bootstrap" --root "${tools_root}" --cache "${cache_root}" --go "${go_bin}"; then :; else bootstrap_status=$?; fi
fi

printf '\nFor this task shell, source %s/activate.sh\n' "${tools_root}"
for tool in go bd actionlint govulncheck go-licenses gitleaks neurostrata-mcp neuroplasticity; do
  if [[ "${tool}" == go ]]; then tool_path="${go_bin}"; else tool_path="${tools_root}/bin/${tool}"; fi
  [[ -x "${tool_path}" ]] && printf 'READY: %s\n' "${tool}" || printf 'MISSING from isolated tool root: %s\n' "${tool}"
done
printf 'Fresh tasks must source %s/activate.sh and repeat the command-discovery check.\n' "${tools_root}"
exit "${bootstrap_status}"
