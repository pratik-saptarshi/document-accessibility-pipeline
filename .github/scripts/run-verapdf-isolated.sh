#!/usr/bin/env bash
set -euo pipefail

if (( $# != 3 )); then
  printf 'usage: run-verapdf-isolated.sh <known-pass|known-finding|financial-v2-statement|financial-v2-invoice> <absolute-input.pdf> <absolute-output-directory>\n' >&2
  exit 2
fi
control=$1
input=$2
output_root=$3
image=${PDFUA_IMAGE:-dap-verapdf:1.30.3}
container_runner=${PDFUA_CONTAINER_RUNNER:-}
max_output_bytes=8388608
max_input_bytes=67108864
case "$control" in
  known-pass) expected_candidate=c255fc1a6d4aeeef41d2b6ee7009dec43754e01dcd3d2c697e662f0040bef96b ;;
  known-finding) expected_candidate=2d0996fa1c49decc6efa1241461832d7724062b14244d24663684f58d28badd0 ;;
  financial-v2-statement) expected_candidate=8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52; max_input_bytes=262144 ;;
  financial-v2-invoice) expected_candidate=31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448; max_input_bytes=262144 ;;
  *) printf 'unsupported control\n' >&2; exit 2 ;;
esac
if [[ "$output_root" != /* ]]; then
  printf 'output directory must be absolute\n' >&2
  exit 2
fi
control_dir="$output_root/$control"
contract="$control_dir/contract.json"
contract_tmp="$control_dir/.contract.json.tmp.$$"
report="$control_dir/report.xml"
stderr_file="$control_dir/.validator.stderr"
stage_dir=
container_name="dap-verapdf-${control}-$$"
cleanup() {
  local status=$?
  trap - EXIT
  timeout --signal=TERM --kill-after=2s 10s docker rm --force "$container_name" >/dev/null 2>&1 || true
  if [[ -n "$stage_dir" ]]; then
    rm -rf -- "$stage_dir"
  fi
  rm -f -- "$stderr_file"
  if (( status != 0 )); then
    rm -f -- "$contract" "$contract_tmp"
  fi
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

mkdir -p -- "$control_dir"
chmod 700 "$control_dir"
rm -f -- "$contract" "$contract_tmp"
if [[ "$container_runner" != /* || ! -x "$container_runner" ]]; then
  printf 'bounded validator container runner must be an absolute executable\n' >&2
  exit 2
fi

if [[ "$input" != /* || -L "$input" || ! -f "$input" ]]; then
  printf 'input must be an absolute regular non-symlink file\n' >&2
  exit 2
fi
actual_candidate=$(sha256sum -- "$input" | awk '{print $1}')
if [[ "$actual_candidate" != "$expected_candidate" ]]; then
  printf 'known control input digest mismatch\n' >&2
  exit 1
fi
if (( $(stat -c %s -- "$input") > max_input_bytes )); then
  printf 'input exceeds the declared validator-input limit\n' >&2
  exit 1
fi
stage_dir=$(mktemp -d "${TMPDIR:-/tmp}/pdfua-input.XXXXXX")
staged_input="$stage_dir/document.pdf"
install -m 0444 "$input" "$staged_input"
staged_candidate=$(sha256sum -- "$staged_input" | awk '{print $1}')
if [[ "$staged_candidate" != "$expected_candidate" ]]; then
  printf 'staged control input digest mismatch\n' >&2
  exit 1
fi

image_id=$(docker image inspect --format '{{.Id}}' "$image")
if [[ ! "$image_id" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  printf 'built validator image identity is invalid\n' >&2
  exit 1
fi

set +e
"$container_runner" 120s "$max_output_bytes" "$report" "$stderr_file" -- \
  docker run --pull=never --rm \
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
    "$image_id" /input/document.pdf
run_status=$?
set -e

if [[ "$control" == known-pass && $run_status -ne 0 ]] || [[ "$control" == known-finding && $run_status -ne 1 ]] || { [[ "$control" == financial-v2-statement || "$control" == financial-v2-invoice ]] && (( run_status != 0 && run_status != 1 )); }; then
  printf 'validator operational or control exit mismatch (%s)\n' "$run_status" >&2
  exit 1
fi
report_size=$(stat -c %s -- "$report")
stderr_size=$(stat -c %s -- "$stderr_file")
if [[ ! -s "$report" ]] || (( report_size + stderr_size >= max_output_bytes )); then
  printf 'validator output is missing or reached the 8 MiB combined limit\n' >&2
  exit 1
fi
report_hash=$(sha256sum -- "$report" | awk '{print $1}')
cat >"$contract_tmp" <<EOF
{
  "schema_version": 1,
  "validator": "veraPDF",
  "validator_version": "1.30.3",
  "validator_sha256": "19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70",
  "base_image": "eclipse-temurin:21.0.12.1_1-jre-noble@sha256:000fd431958bc81a24abe1e8e5f0f0fd3ae365a594bd50aadb20696805f9408c",
  "container_image_id": "$image_id",
  "control": "$control",
  "profile": "PDF/UA-2 + Tagged PDF validation profile",
  "exit_code": $run_status,
  "candidate_sha256": "$staged_candidate",
  "report_sha256": "$report_hash"
}
EOF
chmod 600 "$report" "$contract_tmp"
mv -f -- "$contract_tmp" "$contract"
printf 'Validated %s in isolated container %s (exit %s)\n' "$control" "$image_id" "$run_status"
