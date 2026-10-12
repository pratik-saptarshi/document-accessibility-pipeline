#!/usr/bin/env bash
set -euo pipefail

if (( $# != 1 )) || [[ ! "$1" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  printf 'usage: test-verapdf-output-bound.sh <immutable-sha256-image-id>\n' >&2
  exit 2
fi
image_id=$1
work=$(mktemp -d "${TMPDIR:-/tmp}/pdfua-output-bound.XXXXXX")
trap 'rm -rf -- "$work"' EXIT
container_name="dap-verapdf-output-flood-$$"
ulimit -f 8192
set +e
timeout --signal=TERM --kill-after=3s 30s docker run --pull=never --rm \
  --name "$container_name" \
  --log-driver=none \
  --network=none \
  --read-only \
  --user=65532:65532 \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  --memory=1536m --memory-swap=1536m --cpus=2 --pids-limit=64 \
  --ulimit nofile=1024:1024 \
  --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m \
  --entrypoint /bin/sh "$image_id" -c 'printf "FLOOD_STARTED\\n" >&2; head -c 16777216 /dev/zero' \
  >"$work/stdout" 2>"$work/stderr"
status=$?
set -e
size=$(stat -c %s "$work/stdout")
docker rm --force "$container_name" >/dev/null 2>&1 || true
container_absent=false
for attempt in {1..20}; do
  if docker container inspect --format '{{.Id}}' "$container_name" >"$work/inspect.out" 2>&1; then
    sleep 0.25
    continue
  fi
  if grep -Eq 'No such (object|container)' "$work/inspect.out"; then
    container_absent=true
    break
  fi
  break
done
if [[ "$container_absent" != true ]]; then
  printf 'could not confirm output-flood container cleanup\n' >&2
  cat "$work/inspect.out" >&2
  exit 1
fi
if ! grep -Fxq 'FLOOD_STARTED' "$work/stderr"; then
  printf 'container did not prove successful startup before the output flood\n' >&2
  exit 1
fi
if (( size != 8388608 )); then
  printf 'host report stopped at %s bytes; expected the exact 8 MiB output limit\n' "$size" >&2
  exit 1
fi
if (( status != 1 )); then
  printf 'output flood ended with status %s; expected the bounded Docker output error (1)\n' "$status" >&2
  exit 1
fi
printf 'Container stdout flood was stopped at %s bytes (status %s).\n' "$size" "$status"
