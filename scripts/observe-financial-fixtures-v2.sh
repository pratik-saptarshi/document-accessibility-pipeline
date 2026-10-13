#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"
export GOTOOLCHAIN=local
go run ./cmd/financial-fixtures-v2 observe --repo-root "$repo_root"
