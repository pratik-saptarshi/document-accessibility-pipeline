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

cat >"$work/bin/install" <<'FAKE_INSTALL'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${FAKE_INSTALL_MODE:-}" == mutate-source ]]; then
  source_file=${@: -2:1}
  printf x >>"$source_file"
fi
exec /usr/bin/install "$@"
FAKE_INSTALL
chmod 0700 "$work/bin/install"
cp "$repo_root/internal/pdfua/testdata/oracle-ua2-pass.pdf" "$work/mutated-during-stage.pdf"
export FAKE_INSTALL_MODE=mutate-source
before=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
if bash "$repo_root/.github/scripts/run-verapdf-isolated.sh" known-pass \
  "$work/mutated-during-stage.pdf" "$work/staged-mutation"; then
  printf 'accepted candidate bytes mutated during staging\n' >&2
  exit 1
fi
after=$(grep -c '^run ' "$FAKE_DOCKER_LOG" || true)
[[ "$before" == "$after" ]] || { printf 'ran validator after staged digest mismatch\n' >&2; exit 1; }
[[ ! -e "$work/staged-mutation/known-pass/contract.json" ]] || {
  printf 'staged digest mismatch emitted a validator contract\n' >&2
  exit 1
}
unset FAKE_INSTALL_MODE

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
