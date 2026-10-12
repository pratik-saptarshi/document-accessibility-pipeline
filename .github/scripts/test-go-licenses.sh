#!/usr/bin/env bash
set -euo pipefail

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/project" "$scratch/forbidden"
cat > "$scratch/project/go.mod" <<'MOD'
module example.test/license-policy-smoke

go 1.27

require example.test/forbidden v0.0.0

replace example.test/forbidden => ../forbidden
MOD
cat > "$scratch/project/main.go" <<'GO'
package main

func main() {}
GO
cp LICENSE "$scratch/project/LICENSE"
cat > "$scratch/project/main_test.go" <<'GO'
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
cat > "$scratch/forbidden/go.mod" <<'MOD'
module example.test/forbidden

go 1.27
MOD
cat > "$scratch/forbidden/forbidden.go" <<'GO'
package forbidden

func Value() string { return "synthetic" }
GO
# This full GPL-3.0 license is a real classifier fixture, not a product dependency.
cp "$(dirname "${BASH_SOURCE[0]}")/fixtures/GPL-3.0.txt" "$scratch/forbidden/LICENSE"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
allowed_licenses="$(grep -Ev '^[[:space:]]*(#|$)' "$repo_root/.security/allowed-go-licenses.txt" | paste -sd, -)"
if [[ -z "$allowed_licenses" ]]; then
  printf '%s\n' 'The checked-in Go license allowlist is empty.' >&2
  exit 1
fi
(
  cd "$scratch/project"
  go-licenses check --allowed_licenses="$allowed_licenses" ./...
)

set +e
(
  cd "$scratch/project"
  go-licenses check --include_tests --allowed_licenses="$allowed_licenses" ./...
) > "$scratch/include-tests.log" 2>&1
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  printf '%s\n' 'go-licenses accepted a forbidden license reachable only from a test.' >&2
  exit 1
fi
if ! grep -Eiq "Not allowed license 'GPL-3\.0' found for library 'example.test/forbidden'" "$scratch/include-tests.log"; then
  printf '%s\n' 'go-licenses failed the test-only fixture for an unexpected reason.' >&2
  exit 1
fi

printf '%s\n' 'go-licenses smoke check passed: clean production graph accepted; forbidden test-only dependency rejected.'
