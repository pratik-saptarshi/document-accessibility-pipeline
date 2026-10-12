#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/pdfua-output-bound-shim.XXXXXX")
trap 'rm -rf -- "$work"' EXIT
mkdir "$work/bin"
cat >"$work/bin/docker" <<'FAKE_DOCKER'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == container && "$2" == inspect ]]; then
  printf 'Error: No such object: %s\n' "${4:-unknown}" >&2
  exit 1
fi
if [[ "$1" == rm ]]; then
  exit 0
fi
if [[ "$1" != run ]]; then
  printf 'unexpected Docker shim command\n' >&2
  exit 2
fi
if [[ "$FAKE_DOCKER_MODE" == startup-failure ]]; then
  printf 'docker: Error response from daemon: image not found\n' >&2
  exit 125
fi
if [[ "$FAKE_DOCKER_MODE" == flood ]]; then
  printf 'FLOOD_STARTED\n' >&2
  set +e
  head -c 16777216 /dev/zero
  exit 1
fi
printf 'unsupported Docker shim mode\n' >&2
exit 2
FAKE_DOCKER
chmod 0700 "$work/bin/docker"
export PATH="$work/bin:$PATH"

export FAKE_DOCKER_MODE=flood
bash "$repo_root/.github/scripts/test-verapdf-output-bound.sh" \
  "sha256:$(printf '%064d' 0)"

export FAKE_DOCKER_MODE=startup-failure
if bash "$repo_root/.github/scripts/test-verapdf-output-bound.sh" \
    "sha256:$(printf '%064d' 0)" >"$work/startup-failure.log" 2>&1; then
  printf 'output-bound test accepted a failed container startup\n' >&2
  exit 1
fi
grep -Fq 'container did not prove successful startup before the output flood' "$work/startup-failure.log" || {
  printf 'output-bound test failed for an unexpected startup-failure reason\n' >&2
  cat "$work/startup-failure.log" >&2
  exit 1
}
printf 'Startup-failure control was rejected without a false positive.\n'
