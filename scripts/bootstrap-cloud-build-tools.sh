#!/usr/bin/env bash
set -euo pipefail
set +m

if (( $# != 0 )); then
  printf 'This bootstrap entrypoint does not accept command-line arguments; set CODEX_BUILD_TOOLS_ROOT and CODEX_BUILD_CACHE_ROOT in the environment instead.\n' >&2
  exit 2
fi

# Minimal first-stage bootstrap. The Go orchestrator performs all remaining
# installations from buildtools.manifest.json. Nothing is installed system-wide.
readonly go_version='1.27.2'
readonly go_archive='go1.27.2.linux-amd64.tar.gz'
readonly go_sha256='ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5'
readonly go_url="https://go.dev/dl/${go_archive}"
# The source archive is pinned in buildtools.manifest.json and compiled locally.
# This keeps no-checkout installs independent of Go module proxy/tag behavior.
readonly bootstrap_source_revision='ff5c6dbb5cbcb7660c5a365755e98451de33088c'
readonly bootstrap_source_url="https://github.com/pratik-saptarshi/document-accessibility-pipeline/archive/${bootstrap_source_revision}.tar.gz"
readonly bootstrap_source_sha256='1dfa31cfd8b936cbdf8ca86c662fce80b7becedc4c63301327e1c5dbc769a5ac'

script_dir="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(CDPATH= cd -- "${script_dir}/.." && pwd)"
invocation_dir="$(pwd -P)"
tools_root="${CODEX_BUILD_TOOLS_ROOT:-/workspace/.codex-tools/cloud-build-tools}"
cache_root="${CODEX_BUILD_CACHE_ROOT:-/workspace/.cache/cloud-build-tools}"
[[ "${tools_root}" == /* ]] || tools_root="${invocation_dir}/${tools_root}"
[[ "${cache_root}" == /* ]] || cache_root="${invocation_dir}/${cache_root}"
go_root="${tools_root}/toolchains/go/${go_version}"

if [[ "$(uname -s)" != Linux || "$(uname -m)" != x86_64 ]]; then
  printf 'Unsupported host %s/%s; this manifest currently supports linux/amd64 only.\n' "$(uname -s)" "$(uname -m)" >&2
  exit 2
fi
for utility in curl sha256sum tar timeout setsid; do
  command -v "${utility}" >/dev/null || { printf 'Required bootstrap utility missing: %s\n' "${utility}" >&2; exit 2; }
done
mkdir -p "${tools_root}/toolchains/go" "${cache_root}/downloads"
chmod 700 "${tools_root}" "${cache_root}"

# mkdir is an atomic lock. A timed-out/stale lock is reported for inspection;
# this script never removes a lock owned by a possibly running bootstrap.
lock="${tools_root}/toolchains/go/.bootstrap.lock"
part=''
tmp=''
active_child=''
active_child_pgid=''
active_child_ready=''
child_starting=0
signal_status=0
lock_owned=0
terminate_active_child() {
  local child="${active_child}" group="${active_child_pgid}" attempt
  [[ -n "${child}" ]] || return 0
  # Before the child publishes its ready marker, signal only the launcher.
  # Afterwards, use the process group ID read from inside the new session.
  if [[ -n "${group}" ]]; then kill -TERM -- "-${group}" 2>/dev/null || true
  else kill -TERM "${child}" 2>/dev/null || true; fi
  for ((attempt = 0; attempt < 20; attempt++)); do
    if [[ -n "${group}" ]]; then
      kill -0 -- "-${group}" 2>/dev/null || break
    else
      kill -0 "${child}" 2>/dev/null || break
    fi
    sleep 0.1
  done
  if [[ -n "${group}" ]]; then kill -KILL -- "-${group}" 2>/dev/null || true
  else kill -KILL "${child}" 2>/dev/null || true; fi
  wait "${child}" 2>/dev/null || true
  active_child=''
  active_child_pgid=''
  [[ -z "${active_child_ready}" ]] || rm -f -- "${active_child_ready}"
  active_child_ready=''
}
run_group() {
  local status attempt child_finished=0
  active_child_ready="$(mktemp "${cache_root}/.bootstrap-child-ready.XXXXXX")"
  rm -f -- "${active_child_ready}"
  active_child_pgid=''
  child_starting=1
  setsid -- bash -c 'printf "%s\\n" "$BASHPID" > "$1"; shift; exec "$@"' _ "${active_child_ready}" "$@" &
  active_child=$!
  child_starting=0
  if (( signal_status != 0 )); then
    terminate_active_child
    exit "${signal_status}"
  fi

  # Wait only a bounded time for the new session to publish its actual PGID.
  # This closes the launch/assignment race and avoids treating $! as a PGID.
  for ((attempt = 0; attempt < 100; attempt++)); do
    if [[ -s "${active_child_ready}" ]]; then
      read -r active_child_pgid < "${active_child_ready}"
      break
    fi
    if (( signal_status != 0 )); then
      terminate_active_child
      exit "${signal_status}"
    fi
    if ! kill -0 "${active_child}" 2>/dev/null; then
      child_finished=1
      break
    fi
    sleep 0.05
  done
  if (( signal_status != 0 )); then
    terminate_active_child
    exit "${signal_status}"
  fi
  if [[ -z "${active_child_pgid}" ]]; then
    if (( child_finished )); then
      if wait "${active_child}"; then status=0; else status=$?; fi
      active_child=''
      rm -f -- "${active_child_ready}"
      active_child_ready=''
    else
      terminate_active_child
      status=1
    fi
    if (( signal_status != 0 )); then exit "${signal_status}"; fi
    return "${status}"
  fi

  if wait "${active_child}"; then status=0; else status=$?; fi
  if (( signal_status != 0 )); then
    terminate_active_child
    exit "${signal_status}"
  fi
  active_child=''
  active_child_pgid=''
  rm -f -- "${active_child_ready}"
  active_child_ready=''
  if (( signal_status != 0 )); then exit "${signal_status}"; fi
  return "${status}"
}
check_go_version() {
  run_group bash -c 'set -o pipefail; timeout 20s "$1" version | grep -Fq "go$2 "' _ "$1" "$2"
}
cleanup_part() { [[ -z "${part}" ]] || rm -f -- "${part}"; }
cleanup_tmp() { [[ -z "${tmp}" ]] || rm -rf -- "${tmp}"; }
cleanup() {
  terminate_active_child
  cleanup_tmp
  cleanup_part
  if (( lock_owned )); then rm -rf -- "${lock}"; fi
}
handle_signal() {
  if (( signal_status == 0 )); then signal_status="$1"; fi
  if (( child_starting == 0 )) && [[ -z "${active_child}" ]]; then
    exit "${signal_status}"
  fi
}
trap cleanup EXIT
trap 'handle_signal 130' INT
trap 'handle_signal 143' TERM

deadline=$((SECONDS + 120))
until mkdir "${lock}" 2>/dev/null; do
  if (( SECONDS >= deadline )); then
    printf 'Timed out waiting for %s; check for a live bootstrap before removing it.\n' "${lock}" >&2
    exit 1
  fi
  sleep 1
done
lock_owned=1
printf '%s\n' "$$" > "${lock}/pid"

go_bin="${go_root}/bin/go"
if [[ -x "${go_bin}" ]]; then
  if check_go_version "${go_bin}" "${go_version}"; then
    printf 'Verified existing Go toolchain %s\n' "${go_version}"
  else
    printf 'Existing Go path has an unexpected version: %s\n' "${go_bin}" >&2
    exit 1
  fi
else
  tmp="$(mktemp -d "${tools_root}/toolchains/go/.stage-${go_version}.XXXXXX")"
  archive="${cache_root}/downloads/${go_archive}"
  part="${archive}.part.$$"
  rm -f -- "${part}"
  run_group curl --connect-timeout 15 --max-time 300 --fail --location --silent --show-error "${go_url}" --output "${part}"
  printf '%s  %s\n' "${go_sha256}" "${part}" | sha256sum --check --status || {
    rm -f -- "${part}"
    printf 'Go archive checksum mismatch; no toolchain was installed.\n' >&2
    exit 1
  }
  mv -- "${part}" "${archive}"
  part=''
  mkdir "${tmp}/unpacked"
  run_group tar -xzf "${archive}" -C "${tmp}/unpacked" --strip-components=1
  [[ -x "${tmp}/unpacked/bin/go" ]] || { printf 'Go archive did not contain bin/go.\n' >&2; exit 1; }
  check_go_version "${tmp}/unpacked/bin/go" "${go_version}" || {
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
  if run_group bash -c 'cd "$1/cmd/cloud-tools-bootstrap" && GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly GOMODCACHE="$2/go/mod" GOCACHE="$2/go/build" "$3" run . --manifest "$1/buildtools.manifest.json" --root "$4" --cache "$2" --go "$3"' _ "${repo_root}" "${cache_root}" "${go_bin}" "${tools_root}"; then :; else bootstrap_status=$?; fi
  if run_group bash -c 'cd "$1/cmd/cloud-tools-bootstrap" && GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly GOMODCACHE="$2/go/mod" GOCACHE="$2/go/build" "$3" run . --manifest "$1/buildtools.manifest.json" --root "$4" --cache "$2" --go "$3" --verify' _ "${repo_root}" "${cache_root}" "${go_bin}" "${tools_root}"; then :; else bootstrap_status=1; fi
else
  source_archive="${cache_root}/downloads/cloud-tools-bootstrap-${bootstrap_source_revision}.tar.gz"
  if [[ ! -f "${source_archive}" ]] || ! printf '%s  %s\n' "${bootstrap_source_sha256}" "${source_archive}" | sha256sum --check --status; then
    part="${source_archive}.part.$$"
    rm -f -- "${part}"
    if ! run_group curl --connect-timeout 15 --max-time 180 --fail --location --silent --show-error "${bootstrap_source_url}" --output "${part}"; then
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
  run_group tar -xzf "${source_archive}" -C "${tmp}/source" --strip-components=1
  if ! run_group timeout 8m bash -c 'cd "$1/cmd/cloud-tools-bootstrap" && GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly GOMODCACHE="$2/go/mod" GOCACHE="$2/go/build" "$3" build -buildvcs=false -trimpath -o "$4/cloud-tools-bootstrap" .' _ "${tmp}/source" "${cache_root}" "${go_bin}" "${tmp}/bin"; then
    printf 'Could not build the checksum-verified bootstrap source.\n' >&2
    exit 1
  fi
  chmod 0755 "${tmp}/bin/cloud-tools-bootstrap"
  mv -f -- "${tmp}/bin/cloud-tools-bootstrap" "${tools_root}/bin/cloud-tools-bootstrap"
  if run_group "${tools_root}/bin/cloud-tools-bootstrap" --root "${tools_root}" --cache "${cache_root}" --go "${go_bin}"; then :; else bootstrap_status=$?; fi
  if run_group "${tools_root}/bin/cloud-tools-bootstrap" --root "${tools_root}" --cache "${cache_root}" --go "${go_bin}" --verify; then :; else bootstrap_status=1; fi
fi

printf '\nFor this task shell, source %s/activate.sh\n' "${tools_root}"
printf '\nOperational augmentation status: NOT VERIFIED by binary installation.\n'
printf 'NeuroStrata requires a writable persistent data home, approved local embedder model/runtime, and a successful MCP initialize plus synthetic namespace smoke check.\n'
printf 'NeuroPlasticity requires a reviewed plasticity.json, supported rootless Podman, approved optimizer model, safe agent invocation, and a disposable synthetic evaluation.\n'
printf 'Fresh tasks must source %s/activate.sh and repeat the command-discovery check.\n' "${tools_root}"
exit "${bootstrap_status}"
