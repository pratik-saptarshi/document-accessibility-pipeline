#!/usr/bin/env bash
set -euo pipefail

# test-check-licenses.sh -- behavior test for scripts/check-licenses.sh.
#
# Builds one shared scratch tree that mimics the repository layout
# (scripts/, .security/, LICENSE, a root module, the nested adapter and
# bootstrap modules, and a GPL-3.0 stub module), copies the real script
# and allowlist into it, and runs the script target by target against
# clean and deliberately poisoned fixture graphs. The real repository
# files are copied, never edited.
#
# Exit 0 only if every case behaves as specified.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

log="$scratch/case.log"
err_log="$scratch/case.err"
status=0

mkdir -p "$scratch/scripts" "$scratch/.security" \
  "$scratch/ci/pdfua2/adapter" "$scratch/cmd/cloud-tools-bootstrap" \
  "$scratch/forbidden"

cp "$repo_root/scripts/check-licenses.sh" "$scratch/scripts/check-licenses.sh"
cp "$repo_root/.security/allowed-go-licenses.txt" "$scratch/.security/allowed-go-licenses.txt"
cp "$repo_root/LICENSE" "$scratch/LICENSE"
cp "$repo_root/LICENSE" "$scratch/ci/pdfua2/adapter/LICENSE"

# --- Fixture modules -------------------------------------------------------

write_clean_root_mod() {
  cat > "$scratch/go.mod" <<'MOD'
module example.test/check-licenses-fixture

go 1.27
MOD
}

write_clean_adapter_mod() {
  cat > "$scratch/ci/pdfua2/adapter/go.mod" <<'MOD'
module example.test/adapter-fixture

go 1.27
MOD
}

write_clean_bootstrap_mod() {
  cat > "$scratch/cmd/cloud-tools-bootstrap/go.mod" <<'MOD'
module github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap

go 1.27
MOD
}

cat > "$scratch/main.go" <<'GO'
package main

func main() {}
GO

cat > "$scratch/ci/pdfua2/adapter/adapter.go" <<'GO'
package adapter

func Name() string { return "adapter" }
GO

cat > "$scratch/cmd/cloud-tools-bootstrap/main.go" <<'GO'
package main

func main() {}
GO

cat > "$scratch/forbidden/go.mod" <<'MOD'
module example.test/forbidden

go 1.27
MOD
cat > "$scratch/forbidden/forbidden.go" <<'GO'
package forbidden

func Value() string { return "synthetic" }
GO
# This full GPL-3.0 license is a real classifier fixture, not a product dependency.
cp "$script_dir/fixtures/GPL-3.0.txt" "$scratch/forbidden/LICENSE"

# --- Helpers ---------------------------------------------------------------

reset_fixtures() {
  write_clean_root_mod
  write_clean_adapter_mod
  write_clean_bootstrap_mod
  rm -f "$scratch/main_test.go" "$scratch/ci/pdfua2/adapter/adapter_test.go" \
    "$scratch/go.work"
  cp "$repo_root/.security/allowed-go-licenses.txt" "$scratch/.security/allowed-go-licenses.txt"
}

run_check() {
  local target="$1"
  status=0
  : >"$log"
  : >"$err_log"
  bash "$scratch/scripts/check-licenses.sh" "$target" >"$log" 2>&1 || status=$?
}

fail_case() {
  printf '%s\n' "FAIL: $1" >&2
  if [[ -s "$log" ]]; then
    printf '%s\n' '--- check-licenses output ---' >&2
    cat "$log" >&2
  fi
  if [[ -s "$err_log" ]]; then
    printf '%s\n' '--- check-licenses stderr ---' >&2
    cat "$err_log" >&2
  fi
  exit 1
}

reset_fixtures

# --- Case 1: clean passes --------------------------------------------------

run_check all
if [[ "$status" -ne 0 ]]; then
  fail_case "case 1: clean tree: 'all' exited $status, want 0"
fi
printf '%s\n' 'case 1: clean passes (all exits 0)'

# --- Case 2: disallowed test-only license in root fails --------------------

cat > "$scratch/go.mod" <<'MOD'
module example.test/check-licenses-fixture

go 1.27

require example.test/forbidden v0.0.0

replace example.test/forbidden => ./forbidden
MOD
cat > "$scratch/main_test.go" <<'GO'
package main

import (
	"testing"

	"example.test/forbidden"
)

func TestForbiddenTestOnlyDependency(t *testing.T) {
	if forbidden.Value() != "synthetic" {
		t.Fatal("unexpected fixture value")
	}
}
GO
run_check root
if [[ "$status" -eq 0 ]]; then
  fail_case "case 2: root accepted a test-only GPL-3.0 dependency"
fi
if ! grep -Eq "Not allowed license 'GPL-3\.0' found for library 'example.test/forbidden'" "$log"; then
  fail_case "case 2: root failed without the expected GPL-3.0 message"
fi
printf '%s\n' 'case 2: disallowed test-only license in root fails with the GPL-3.0 message'
reset_fixtures

# --- Case 3: empty allowlist fails at the guard ----------------------------

printf '%s\n' '# comments only, no selectable licenses' > "$scratch/.security/allowed-go-licenses.txt"
status=0
bash "$scratch/scripts/check-licenses.sh" root >"$log" 2>"$err_log" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail_case "case 3: root accepted an empty allowlist"
fi
if ! grep -Fq 'check-licenses: allowed license list is empty' "$err_log"; then
  fail_case "case 3: empty allowlist did not produce the pinned guard message on stderr"
fi
printf '%s\n' 'case 3: empty allowlist fails at the guard with the pinned message'
reset_fixtures

# --- Case 4: bootstrap assertion fails -------------------------------------

cat > "$scratch/cmd/cloud-tools-bootstrap/go.mod" <<'MOD'
module github.com/pratik-saptarshi/document-accessibility-pipeline/cmd/cloud-tools-bootstrap

go 1.27

require example.test/forbidden v0.0.0

replace example.test/forbidden => ../../forbidden
MOD
run_check bootstrap
if [[ "$status" -eq 0 ]]; then
  fail_case "case 4: bootstrap accepted an extra module in its graph"
fi
if ! grep -Fq 'example.test/forbidden' "$log"; then
  fail_case "case 4: bootstrap failure output did not name example.test/forbidden"
fi
printf '%s\n' 'case 4: bootstrap assertion fails and names the extra module'
reset_fixtures

# --- Case 5: missing tool fails closed -------------------------------------

shim="$scratch/shim"
mkdir -p "$shim"
for tool in go grep paste dirname cat; do
  ln -s "$(command -v "$tool")" "$shim/$tool"
done
# Resolve bash before restricting PATH: with PATH set to the shim alone a
# bare 'bash' would not be found, so invoke the interpreter by the path
# resolved here. The script is still run via bash, never by direct exec,
# so its shebang/env resolution is not part of the case.
bash_bin="$(command -v bash)"
status=0
PATH="$shim" "$bash_bin" "$scratch/scripts/check-licenses.sh" root >"$log" 2>"$err_log" || status=$?
if [[ "$status" -eq 0 ]]; then
  fail_case "case 5: root passed without go-licenses on PATH"
fi
if ! grep -Fq 'check-licenses: the pinned license tool was not found on PATH' "$err_log"; then
  fail_case "case 5: missing go-licenses did not produce the not-found message on stderr"
fi
printf '%s\n' 'case 5: missing go-licenses fails closed with the not-found message'
reset_fixtures

# --- Case 6: stray go.work does not change verdicts -----------------------

cat > "$scratch/go.work" <<'WORK'
go 1.27

use .
WORK
run_check all
if [[ "$status" -ne 0 ]]; then
  fail_case "case 6: stray go.work changed the verdict: 'all' exited $status, want 0"
fi
printf '%s\n' 'case 6: stray go.work does not change verdicts (all exits 0)'
reset_fixtures

# --- Case 7: disallowed license in adapter only fails ----------------------

cat > "$scratch/ci/pdfua2/adapter/go.mod" <<'MOD'
module example.test/adapter-fixture

go 1.27

require example.test/forbidden v0.0.0

replace example.test/forbidden => ../../../forbidden
MOD
cat > "$scratch/ci/pdfua2/adapter/adapter_test.go" <<'GO'
package adapter

import (
	"testing"

	"example.test/forbidden"
)

func TestForbiddenTestOnlyDependency(t *testing.T) {
	if forbidden.Value() != "synthetic" {
		t.Fatal("unexpected fixture value")
	}
}
GO
run_check adapter
if [[ "$status" -eq 0 ]]; then
  fail_case "case 7: adapter accepted a test-only GPL-3.0 dependency"
fi
if ! grep -Eq "Not allowed license 'GPL-3\.0' found for library 'example.test/forbidden'" "$log"; then
  fail_case "case 7: adapter failed without the expected GPL-3.0 message"
fi
run_check root
if [[ "$status" -ne 0 ]]; then
  fail_case "case 7: root exited $status with a poisoned adapter graph, want 0"
fi
printf '%s\n' 'case 7: disallowed license in adapter only fails adapter; root still passes'
reset_fixtures

printf '%s\n' 'test-check-licenses: all 7 cases passed.'
