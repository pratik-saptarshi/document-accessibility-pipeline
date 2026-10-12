#!/usr/bin/env bash
set -euo pipefail

if (( $# != 1 )) || [[ ! "$1" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  printf 'usage: test-verapdf-output-bound.sh <immutable-sha256-image-id>\n' >&2
  exit 2
fi
image_id=$1
work=$(mktemp -d "${TMPDIR:-/tmp}/pdfua-output-bound.XXXXXX")
trap 'rm -rf -- "$work"' EXIT
ulimit -f 8192
set +e
timeout --signal=TERM --kill-after=3s 30s docker run --pull=never --rm \
  --log-driver=none \
  --network=none \
  --read-only \
  --user=65532:65532 \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  --memory=1536m --memory-swap=1536m --cpus=2 --pids-limit=64 \
  --ulimit nofile=1024:1024 \
  --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m \
  --entrypoint /bin/sh "$image_id" -c 'head -c 16777216 /dev/zero' \
  >"$work/stdout" 2>"$work/stderr"
status=$?
set -e
size=$(stat -c %s "$work/stdout")
if (( status == 0 )); then
  printf 'container output flood unexpectedly completed successfully\n' >&2
  exit 1
fi
if (( size > 8388608 )); then
  printf 'host report exceeded 8 MiB output limit (%s bytes)\n' "$size" >&2
  exit 1
fi
printf 'Container stdout flood was stopped at %s bytes (status %s).\n' "$size" "$status"
