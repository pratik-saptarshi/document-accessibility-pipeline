#!/usr/bin/env bash
set -euo pipefail

repository_root="${1:-$(git rev-parse --show-toplevel)}"
repository_root="$(cd "$repository_root" && pwd -P)"
git_metadata="$repository_root/.git"

while IFS= read -r -d '' candidate; do
  printf 'Unmanaged Gitleaks config or suppression file is prohibited: %s\n' "${candidate#"$repository_root"/}" >&2
  exit 1
done < <(find "$repository_root" -path "$git_metadata" -prune -o \( -name .gitleaks.toml -o -name .gitleaksignore \) -print0)

printf '%s\n' 'No unmanaged Gitleaks config or suppression files found.'
