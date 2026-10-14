#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
mkdir "$work/bin"
container_runner=${PDFUA_CONTAINER_RUNNER:-"$work/bin/verapdf-container-runner"}
if [[ ! -x "$container_runner" ]]; then
  (cd "$repo_root" && go build -o "$container_runner" ./cmd/verapdf-container-runner)
fi
export PDFUA_CONTAINER_RUNNER="$container_runner"
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
if [[ "${FAKE_DOCKER_MODE:-}" == exact-flood ]]; then
  head -c 8388608 /dev/zero
  exit 1
fi
if [[ "${FAKE_DOCKER_MODE:-}" == stderr-flood ]]; then
  cat "$FAKE_REPORT"
  head -c 9437184 /dev/zero >&2
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

export FAKE_DOCKER_EXIT=125
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$output" >"$work/stale-rerun.log" 2>&1; then
  printf 'accepted an operational rerun after a prior valid contract\n' >&2
  exit 1
fi
[[ ! -e "$output/known-pass/contract.json" ]] || {
  printf 'failed rerun left a stale known-pass contract\n' >&2
  exit 1
}
grep -Fq 'validator operational or control exit mismatch (125)' "$work/stale-rerun.log" || {
  printf 'operational container status was not reported distinctly\n' >&2
  exit 1
}

export FAKE_REPORT="$repo_root/internal/pdfua/testdata/contract-noncompliant.xml"
export FAKE_DOCKER_EXIT=1
bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-finding \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-fail.pdf" "$output"
grep -q '"exit_code": 1' "$output/known-finding/contract.json"

bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" financial-v2-statement \
  "$repo_root/docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf" "$output"
grep -q '"control": "financial-v2-statement"' "$output/financial-v2-statement/contract.json"
grep -q '"candidate_sha256": "8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52"' "$output/financial-v2-statement/contract.json"
grep -q '"exit_code": 1' "$output/financial-v2-statement/contract.json"

export FAKE_REPORT="$repo_root/internal/pdfua/testdata/contract-compliant.xml"
export FAKE_DOCKER_EXIT=0
bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" financial-v2-invoice \
  "$repo_root/docs/fixtures/remediation/financial-tables-v2/candidates/invoice-repaired-v2.pdf" "$output"
grep -q '"control": "financial-v2-invoice"' "$output/financial-v2-invoice/contract.json"
grep -q '"candidate_sha256": "31c0f7764200c55e0ccece318dca59a77437b14934c3d86fcb7eac3b02e54448"' "$output/financial-v2-invoice/contract.json"
grep -q '"exit_code": 0' "$output/financial-v2-invoice/contract.json"

cat >"$work/bin/sha256sum" <<'FAKE_SHA256SUM'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >>"$FAKE_SHA256SUM_LOG"
printf '8e944c31c8b1aa2bf34da542c950cab76356f6f608454f793ef889ecdf898e52  %s\n' "${@: -1}"
FAKE_SHA256SUM
chmod 0700 "$work/bin/sha256sum"
export FAKE_SHA256SUM_LOG="$work/sha256sum.log"
dd if=/dev/zero of="$work/oversized-candidate.pdf" bs=262145 count=1 status=none
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" financial-v2-statement \
  "$work/oversized-candidate.pdf" "$work/oversized-input" >"$work/oversized-input.log" 2>&1; then
  printf 'accepted an oversized v2 candidate input\n' >&2
  exit 1
fi
grep -Fq 'input exceeds the declared validator-input limit' "$work/oversized-input.log" || {
  printf 'oversized candidate did not fail at the declared input-size boundary\n' >&2
  cat "$work/oversized-input.log" >&2
  exit 1
}
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || {
  printf 'ran validator after oversized-input rejection\n' >&2
  exit 1
}
[[ ! -s "$FAKE_SHA256SUM_LOG" ]] || {
  printf 'computed a candidate digest before rejecting its oversized input\n' >&2
  cat "$FAKE_SHA256SUM_LOG" >&2
  exit 1
}
rm "$work/bin/sha256sum"
unset FAKE_SHA256SUM_LOG

for flag in --network=none --read-only --user=65532:65532 --cap-drop=ALL \
  --security-opt=no-new-privileges:true --memory=1536m --pids-limit=64 \
  --tmpfs; do
  grep -Fq -- "$flag" "$FAKE_DOCKER_LOG" || { printf 'missing isolation flag: %s\n' "$flag" >&2; exit 1; }
done
grep -Fq 'readonly' "$FAKE_DOCKER_LOG" || { printf 'input mount was not read-only\n' >&2; exit 1; }
grep -Fq -- '--pull=never' "$FAKE_DOCKER_LOG" || { printf 'runner did not disable image pulls\n' >&2; exit 1; }
grep -Fq -- '--log-driver=none' "$FAKE_DOCKER_LOG" || { printf 'runner did not disable daemon-side log retention\n' >&2; exit 1; }
grep -Fq -- 'sha256:'"$(printf '%064d' 0)" "$FAKE_DOCKER_LOG" || {
  printf 'runner did not bind execution to the inspected immutable image ID\n' >&2
  exit 1
}
if grep -Fq 'synthetic-validator:fixture' "$FAKE_DOCKER_LOG"; then
  printf 'runner executed a mutable image tag after inspecting it\n' >&2
  exit 1
fi

cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/tampered.pdf"
printf tamper >>"$work/tampered.pdf"
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass "$work/tampered.pdf" "$output"; then
  printf 'accepted a modified known control PDF\n' >&2
  exit 1
fi
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after a control digest mismatch\n' >&2; exit 1; }

cp "$repo_root/docs/fixtures/remediation/financial-tables-v2/candidates/statement-repaired-v2.pdf" "$work/tampered-candidate.pdf"
printf tamper >>"$work/tampered-candidate.pdf"
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" financial-v2-statement "$work/tampered-candidate.pdf" "$output"; then
  printf 'accepted modified repaired candidate bytes\n' >&2
  exit 1
fi
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after candidate digest mismatch\n' >&2; exit 1; }

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
[[ ! -e "$work/flood/known-pass/contract.json" ]] || {
  printf 'oversized output emitted a validator contract\n' >&2
  exit 1
}

export FAKE_DOCKER_MODE=exact-flood
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-finding \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-fail.pdf" "$work/exact-flood"; then
  printf 'accepted validator output exactly at the configured byte limit\n' >&2
  exit 1
fi
[[ $(stat -c %s "$work/exact-flood/known-finding/report.xml") -eq 8388608 ]] || {
  printf 'exact-limit control did not exercise the byte boundary\n' >&2
  exit 1
}
[[ ! -e "$work/exact-flood/known-finding/contract.json" ]] || {
  printf 'exact-limit output emitted a validator contract\n' >&2
  exit 1
}

export FAKE_DOCKER_MODE=stderr-flood
export FAKE_DOCKER_EXIT=1
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-finding \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-fail.pdf" "$work/stderr-flood"; then
  printf 'accepted validator stderr at the output limit\n' >&2
  exit 1
fi
[[ ! -e "$work/stderr-flood/known-finding/contract.json" ]] || {
  printf 'stderr overflow emitted a validator contract\n' >&2
  exit 1
}

# --- dap-jo9.1.6.6: bounded single-capture input snapshot ---
cat >"$work/bin/head" <<'FAKE_HEAD'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" != "-c" || "$3" != "--" ]]; then
  printf 'bounded capture did not use the expected head -c form\n' >&2
  exit 97
fi
if [[ "${FAKE_HEAD_MODE:-}" == grow ]]; then
  # This control is specifically for financial-v2's 262144-byte input cap.
  # Validate the requested limit, open the source, then grow that same inode
  # before reading through the already-open descriptor.
  [[ "$2" == 262145 ]] || {
    printf 'bounded growth capture requested %s bytes, want 262145\n' "$2" >&2
    exit 97
  }
  src="${@: -1}"
  exec 3<"$src"
  dd if=/dev/zero of="$src" oflag=append conv=notrunc bs=262145 count=1 status=none
  /usr/bin/head -c "$2" <&3
  exit 0
fi
if [[ "${FAKE_HEAD_MODE:-}" == replace-path ]]; then
  src="${@: -1}"
  mv -f -- "$FAKE_HEAD_REPLACEMENT" "$src"
fi
if [[ "${FAKE_HEAD_MODE:-}" == mutate-source ]]; then
  src="${@: -1}"
  printf x >>"$src"
fi
if [[ "${FAKE_HEAD_MODE:-}" == partial ]]; then
  /usr/bin/head -c 100 /dev/zero
  exit 1
fi
exec /usr/bin/head "$@"
FAKE_HEAD
chmod 0700 "$work/bin/head"

# growth during capture is rejected at the declared input limit
cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/growing-source.pdf"
export FAKE_HEAD_MODE=grow
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" financial-v2-statement \
  "$work/growing-source.pdf" "$work/growth" >"$work/growth.log" 2>&1; then
  printf 'accepted input that grew during bounded capture\n' >&2
  exit 1
fi
grep -Fq 'input exceeds the declared validator-input limit' "$work/growth.log" || {
  printf 'growth during capture did not fail at the declared input-size boundary\n' >&2
  cat "$work/growth.log" >&2
  exit 1
}
[[ ! -e "$work/growth/financial-v2-statement/contract.json" ]] || {
  printf 'growth during capture emitted a validator contract\n' >&2
  exit 1
}
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after growth rejection\n' >&2; exit 1; }
unset FAKE_HEAD_MODE

# replacing the input pathname after pre-checks is rejected by the captured
# snapshot's expected digest, without launching the validator
cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/replaced-during-capture.pdf"
printf 'replacement with unexpected bytes\n' >"$work/replacement.pdf"
export FAKE_HEAD_MODE=replace-path
export FAKE_HEAD_REPLACEMENT="$work/replacement.pdf"
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$work/replaced-during-capture.pdf" "$work/path-replacement" >"$work/path-replacement.log" 2>&1; then
  printf 'accepted a path-replaced source with the wrong digest\n' >&2
  exit 1
fi
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after path-replacement digest mismatch\n' >&2; exit 1; }
[[ ! -e "$work/path-replacement/known-pass/contract.json" ]] || {
  printf 'path-replacement digest mismatch emitted a validator contract\n' >&2
  exit 1
}
grep -Fq 'known control input digest mismatch' "$work/path-replacement.log" || {
  printf 'path replacement did not report a digest mismatch\n' >&2
  cat "$work/path-replacement.log" >&2
  exit 1
}
unset FAKE_HEAD_MODE FAKE_HEAD_REPLACEMENT

# mutation during bounded capture is caught by the snapshot digest
cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/mutated-during-capture.pdf"
export FAKE_HEAD_MODE=mutate-source
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$work/mutated-during-capture.pdf" "$work/capture-mutation" >"$work/capture-mutation.log" 2>&1; then
  printf 'accepted candidate bytes mutated during bounded capture\n' >&2
  exit 1
fi
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after capture digest mismatch\n' >&2; exit 1; }
[[ ! -e "$work/capture-mutation/known-pass/contract.json" ]] || {
  printf 'capture digest mismatch emitted a validator contract\n' >&2
  exit 1
}
grep -Fq 'known control input digest mismatch' "$work/capture-mutation.log" || {
  printf 'capture mutation did not report a digest mismatch\n' >&2
  cat "$work/capture-mutation.log" >&2
  exit 1
}
unset FAKE_HEAD_MODE

# partial capture fails closed with a distinct message
export FAKE_HEAD_MODE=partial
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/partial" >"$work/partial.log" 2>&1; then
  printf 'accepted a partial bounded capture\n' >&2
  exit 1
fi
grep -Fq 'failed to capture bounded validator input snapshot' "$work/partial.log" || {
  printf 'partial capture did not fail closed with a distinct message\n' >&2
  cat "$work/partial.log" >&2
  exit 1
}
[[ ! -e "$work/partial/known-pass/contract.json" ]] || {
  printf 'partial capture emitted a validator contract\n' >&2
  exit 1
}
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after partial-copy failure\n' >&2; exit 1; }
unset FAKE_HEAD_MODE
rm "$work/bin/head"
