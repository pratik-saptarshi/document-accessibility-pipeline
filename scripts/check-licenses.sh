#!/usr/bin/env bash
# check-licenses.sh -- single entry point for the test-inclusive Go license
# checks, used by both CI (.github/workflows/security-scans.yml) and local
# runs so the two cannot drift apart.
#
# Usage: bash scripts/check-licenses.sh [root|adapter|bootstrap|all]
#   root      license-check the root module
#   adapter   license-check the nested ci/pdfua2/adapter module
#   bootstrap assert the nested cmd/cloud-tools-bootstrap module graph
#             contains exactly its own module and nothing else
#   all       root, adapter, and bootstrap, in that order (default)
#
# Exit 0 means every requested check passed; any failure aborts nonzero
# (fail-closed).
#
# The parity exports below pin the toolchain to the locally installed Go
# (GOTOOLCHAIN=local) and pin GOROOT to that same installation. Without
# them, a cached newer toolchain can be selected whose GOROOT does not
# match the license tool's build environment, which made local and CI
# results diverge.
#
# Do not source this script: it sets shell options and exports that would mutate the caller.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  printf '%s\n' 'usage: bash scripts/check-licenses.sh [root|adapter|bootstrap|all]' >&2
}

# Parse the target argument before any tool preconditions, so the
# bootstrap target never requires the license tool or the allowlist.
target="${1:-all}"
if [[ $# -gt 1 ]]; then
  usage
  exit 2
fi
case "$target" in
  root|adapter|bootstrap|all) ;;
  *)
    usage
    exit 2
    ;;
esac

# Parity environment: exported before any go invocation.
export GOTOOLCHAIN=local
export GOWORK=off
export GOFLAGS=-mod=readonly

if ! command -v go; then
  printf '%s\n' 'check-licenses: go not found on PATH' >&2
  exit 1
fi
export GOROOT="$(go env GOROOT)"
printf '%s\n' "check-licenses: $(go version) (GOTOOLCHAIN=local GOROOT=$GOROOT)"

allowed_licenses=""
if [[ "$target" == "root" || "$target" == "adapter" || "$target" == "all" ]]; then
  if ! command -v go-licenses; then
    {
      printf '%s\n' 'check-licenses: the pinned license tool was not found on PATH.'
      printf '%s\n' 'Install pinned v2.0.1 the same way as the Install pinned security tools step in .github/workflows/security-scans.yml:'
      printf '%s\n' '  go install github.com/google/go-licenses/v2@v2.0.1'
    } >&2
    exit 1
  fi
  # The status guard matters: with no selectable lines grep exits 1, and an
  # unguarded assignment under set -e/pipefail would exit silently before
  # the empty-list message below could print.
  if ! allowed_licenses="$(grep -Ev '^[[:space:]]*(#|$)' "$repo_root/.security/allowed-go-licenses.txt" | paste -sd, -)"; then
    allowed_licenses=""
  fi
  if [[ -z "$allowed_licenses" ]]; then
    printf '%s\n' 'check-licenses: allowed license list is empty (check .security/allowed-go-licenses.txt)' >&2
    exit 1
  fi
fi

check_root() {
  (cd "$repo_root" && go-licenses check --include_tests --allowed_licenses="$allowed_licenses" ./...)
}

check_adapter() {
  (cd "$repo_root/ci/pdfua2/adapter" && go-licenses check --include_tests --allowed_licenses="$allowed_licenses" ./...)
}

check_bootstrap() {
  local deps
  local line
  if ! deps="$(cd "$repo_root/cmd/cloud-tools-bootstrap" && go list -m all)"; then
    printf '%s\n' 'check-licenses: go list -m all failed in cmd/cloud-tools-bootstrap (see the error above)' >&2
    return 1
  fi
  if [[ "$deps" != 'github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap' ]]; then
    printf '%s\n' 'check-licenses: bootstrap module graph must contain only its own module; full module list:' >&2
    printf '%s\n' "$deps" >&2
    printf '%s\n' 'check-licenses: unexpected modules:' >&2
    while IFS= read -r line; do
      if [[ "$line" != 'github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap' ]]; then
        printf '%s\n' "$line" >&2
      fi
    done <<< "$deps"
    return 1
  fi
}

case "$target" in
  root)
    printf '%s\n' 'check-licenses: checking root module'
    check_root
    printf '%s\n' 'check-licenses: root: pass'
    ;;
  adapter)
    printf '%s\n' 'check-licenses: checking adapter module'
    check_adapter
    printf '%s\n' 'check-licenses: adapter: pass'
    ;;
  bootstrap)
    printf '%s\n' 'check-licenses: checking bootstrap module graph'
    check_bootstrap
    printf '%s\n' 'check-licenses: bootstrap: pass'
    ;;
  all)
    printf '%s\n' 'check-licenses: checking root module'
    check_root
    printf '%s\n' 'check-licenses: root: pass'
    printf '%s\n' 'check-licenses: checking adapter module'
    check_adapter
    printf '%s\n' 'check-licenses: adapter: pass'
    printf '%s\n' 'check-licenses: checking bootstrap module graph'
    check_bootstrap
    printf '%s\n' 'check-licenses: bootstrap: pass'
    ;;
esac
