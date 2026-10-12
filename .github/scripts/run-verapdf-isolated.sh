#!/usr/bin/env bash
set -euo pipefail

if (( $# != 3 )); then
  printf 'usage: run-verapdf-isolated.sh <known-pass|known-finding> <absolute-input.pdf> <absolute-output-directory>\n' >&2
  exit 2
fi
control=$1
input=$2
output_root=$3
image=${PDFUA_IMAGE:-dap-verapdf:1.30.3}
case "$control" in
  known-pass) expected_candidate=c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b ;;
  known-finding) expected_candidate=2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0 ;;
  *) printf 'unsupported control\n' >&2; exit 2 ;;
esac
if [[ "$input" != /* || -L "$input" || ! -f "$input" || "$output_root" != /* ]]; then
  printf 'input must be an absolute regular file; output directory must be absolute\n' >&2
  exit 2
fi
actual_candidate=$(sha256sum -- "$input" | awk '{print $1}')
if [[ "$actual_candidate" != "$expected_candidate" ]]; then
  printf 'known control input digest mismatch\n' >&2
  exit 1
fi
if (( $(stat -c %s -- "$input") > 67108864 )); then
  printf 'known control input exceeds the 64 MiB limit\n' >&2
  exit 1
fi
stage_dir=$(mktemp -d "${TMPDIR:-/tmp}/pdfua-input.XXXXXX")
container_name="dap-verapdf-${control}-$$"
cleanup() {
  timeout --signal=TERM --kill-after=2s 10s docker rm --force "$container_name" >/dev/null 2>&1 || true
  rm -rf -- "$stage_dir"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
staged_input="$stage_dir/document.pdf"
install -m 0444 "$input" "$staged_input"
image_id=$(docker image inspect --format '{{.Id}}' "$image")
if [[ ! "$image_id" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  printf 'built validator image identity is invalid\n' >&2
  exit 1
fi
control_dir="$output_root/$control"
mkdir -p -- "$control_dir"
chmod 700 "$control_dir"
report="$control_dir/report.xml"
stderr_file="$control_dir/.validator.stderr"
ulimit -f 8192
set +e
timeout --signal=TERM --kill-after=10s 120s docker run --pull=never --rm \
  --log-driver=none \
  --name "$container_name" \
  --network=none \
  --read-only \
  --user=65532:65532 \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  --memory=1536m --memory-swap=1536m --cpus=2 --pids-limit=64 \
  --ulimit nofile=1024:1024 \
  --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m \
  --mount "type=bind,src=$staged_input,dst=/input/document.pdf,readonly" \
  "$image_id" /input/document.pdf >"$report" 2>"$stderr_file"
exit_code=$?
set -e
rm -f -- "$stderr_file"
if [[ "$control" == known-pass && $exit_code -ne 0 ]] || [[ "$control" == known-finding && $exit_code -ne 1 ]]; then
  printf 'validator operational or control exit mismatch (%s)\n' "$exit_code" >&2
  exit 1
fi
if [[ ! -s "$report" || $(stat -c %s -- "$report") -gt 8388608 ]]; then
  printf 'validator report is missing or exceeds the 8 MiB limit\n' >&2
  exit 1
fi
report_hash=$(sha256sum -- "$report" | awk '{print $1}')
cat >"$control_dir/contract.json" <<EOF
{
  "schema_version": 1,
  "validator": "veraPDF",
  "validator_version": "1.30.3",
  "validator_sha256": "19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70",
  "base_image": "eclipse-temurin:21.0.12.1_1-jre-noble@sha256:000fd431958bc81a24abe1e8e5f0f0fd3ae365a594bd50aadb20696805f9408c",
  "container_image_id": "$image_id",
  "control": "$control",
  "profile": "PDF/UA-2 + Tagged PDF validation profile",
  "exit_code": $exit_code,
  "candidate_sha256": "$actual_candidate",
  "report_sha256": "$report_hash"
}
EOF
chmod 600 "$report" "$control_dir/contract.json"
printf 'Validated %s in isolated container %s (exit %s)\n' "$control" "$image_id" "$exit_code"
