#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
mkdir "$work/bin"
cat >"$work/bin/docker" <<'FAKE_DOCKER'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == image && "$2" == inspect ]]; then
  printf 'sha256:%064d\n' 0
  exit 0
fi
if [[ "$1" == rm ]]; then
  printf '%s\n' "$*" >>"$FAKE_DOCKER_LOG"
  exit 0
fi
printf '%s\n' "$*" >>"$FAKE_DOCKER_LOG"
if [[ "${FAKE_DOCKER_MODE:-}" == flood ]]; then
  head -c 9437184 /dev/zero
  exit 1
fi
cat "$FAKE_REPORT"
exit "${FAKE_DOCKER_EXIT:-0}"
FAKE_DOCKER
chmod 0700 "$work/bin/docker"
export PATH="$work/bin:$PATH"
export PDFUA_IMAGE=synthetic-validator:fixture
export FAKE_DOCKER_LOG="$work/docker-args.log"
export FAKE_REPORT="$repo_root/internal/pdfua/testdata/contract-compliant.xml"
output="$work/out"

bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$output"
grep -q '"exit_code": 0' "$output/known-pass/contract.json"

export FAKE_REPORT="$repo_root/internal/pdfua/testdata/contract-noncompliant.xml"
export FAKE_DOCKER_EXIT=1
bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-finding \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-fail.pdf" "$output"
grep -q '"exit_code": 1' "$output/known-finding/contract.json"

for flag in --network=none --read-only --user=65532:65532 --cap-drop=ALL \
  --security-opt=no-new-privileges:true --memory=1536m --pids-limit=64 \
  --tmpfs; do
  grep -Fq -- "$flag" "$FAKE_DOCKER_LOG" || { printf 'missing isolation flag: %s\n' "$flag" >&2; exit 1; }
done
grep -Fq 'readonly' "$FAKE_DOCKER_LOG" || { printf 'input mount was not read-only\n' >&2; exit 1; }

cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/tampered.pdf"
printf tamper >>"$work/tampered.pdf"
before=$(wc -l <"$FAKE_DOCKER_LOG")
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass "$work/tampered.pdf" "$output"; then
  printf 'accepted a modified known control PDF\n' >&2
  exit 1
fi
after=$(wc -l <"$FAKE_DOCKER_LOG")
[[ "$before" == "$after" ]] || { printf 'ran validator after a control digest mismatch\n' >&2; exit 1; }

unset FAKE_DOCKER_MODE
export FAKE_REPORT="$repo_root/internal/pdfua/testdata/contract-compliant.xml"
export FAKE_DOCKER_EXIT=137
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/operational"; then
  printf 'accepted an operational validator failure\n' >&2
  exit 1
fi
[[ ! -e "$work/operational/known-pass/contract.json" ]] || {
  printf 'operational failure emitted an accepted contract\n' >&2
  exit 1
}
grep -Fq 'rm --force dap-verapdf-known-pass-' "$FAKE_DOCKER_LOG" || {
  printf 'container cleanup was not requested after an operational failure\n' >&2
  exit 1
}

export FAKE_DOCKER_MODE=flood
unset FAKE_DOCKER_EXIT
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/flood"; then
  printf 'accepted oversized validator output\n' >&2
  exit 1
fi
[[ $(stat -c %s "$work/flood/known-pass/report.xml") -le 8388608 ]] || {
  printf 'oversized output exceeded the report file limit\n' >&2
  exit 1
}
