#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture_dir="$repo_root/docs/fixtures/remediation/financial-tables-v2"
output_dir="$fixture_dir/observations"
mkdir -p "$output_dir"

pdftotext_version="$(pdftotext -v 2>&1 | head -n 1)"
ghostscript_version="$(gs --version)"
records="$output_dir/.observations.records"
trap 'rm -f "$records"' EXIT
: >"$records"

for family in statement invoice; do
	for state in positive damaged; do
		name="$family-$state-v2"
		pdf="$fixture_dir/$name.pdf"
		text_output="$output_dir/$name.txt"
		render_output="$output_dir/$name.png"
		pdf_sha="$(sha256sum "$pdf" | cut -d ' ' -f 1)"
		pdftotext -layout "$pdf" "$text_output"
		gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r72 -sOutputFile="$render_output" "$pdf"
		text_sha="$(sha256sum "$text_output" | cut -d ' ' -f 1)"
		render_sha="$(sha256sum "$render_output" | cut -d ' ' -f 1)"
		printf '%s\t%s\t%s\t%s\n' "$name" "$pdf_sha" "$text_sha" "$render_sha" >>"$records"
	done
done

(
	cd "$repo_root"
	go run ./cmd/financial-fixtures-v2 observations \
		--records "$records" \
		--output "$output_dir/observations.v2.json" \
		--pdftotext-version "$pdftotext_version" \
		--ghostscript-version "$ghostscript_version"
)
