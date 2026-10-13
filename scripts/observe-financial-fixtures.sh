#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture_dir="$repo_root/docs/fixtures/remediation/financial-tables"
output_dir="$fixture_dir/observations"
mkdir -p "$output_dir"

pdftotext_version="$(pdftotext -v 2>&1 | head -n 1)"
ghostscript_version="$(gs --version)"
records="$output_dir/.observations.records"
: >"$records"

for family in statement invoice; do
	for state in positive damaged; do
		name="$family-$state"
		pdf="$fixture_dir/$name.pdf"
		text_output="$output_dir/$name.txt"
		render_output="$output_dir/$name.png"
		pdf_sha="$(sha256sum "$pdf" | cut -d ' ' -f 1)"
		set +e
		pdftotext -layout "$pdf" "$text_output"
		text_status=$?
		set -e
		if (( text_status != 0 )); then
			echo "pdftotext failed for $name with exit $text_status" >&2
			exit "$text_status"
		fi
		set +e
		gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r72 -sOutputFile="$render_output" "$pdf"
		render_status=$?
		set -e
		if (( render_status != 0 )); then
			echo "Ghostscript failed for $name with exit $render_status" >&2
			exit "$render_status"
		fi
		text_sha="$(sha256sum "$text_output" | cut -d ' ' -f 1)"
		render_sha="$(sha256sum "$render_output" | cut -d ' ' -f 1)"
		printf '%s\t%s\t%s\t%s\t%d\t%s\t%d\n' "$name" "$pdf_sha" "$text_sha" "$render_sha" "$text_status" "$render_output" "$render_status" >>"$records"
	done
done

python3 - "$records" "$output_dir/observations.v1.json" "$pdftotext_version" "$ghostscript_version" <<'PY'
import json
import sys
from pathlib import Path

records_path = Path(sys.argv[1])
output_path = Path(sys.argv[2])
text_version, render_version = sys.argv[3:5]
observations = []
for line in records_path.read_text().splitlines():
    name, pdf_sha, text_sha, render_sha, text_status, render_path, render_status = line.split("\t")
    observations.append({
        "fixture": f"{name}.pdf",
        "pdf_sha256": pdf_sha,
        "text_command": f"pdftotext -layout docs/fixtures/remediation/financial-tables/{name}.pdf docs/fixtures/remediation/financial-tables/observations/{name}.txt",
        "text_tool": text_version,
        "text_exit_status": int(text_status),
        "text_output_path": f"docs/fixtures/remediation/financial-tables/observations/{name}.txt",
        "text_output_sha256": text_sha,
        "render_command": f"gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r72 -sOutputFile=docs/fixtures/remediation/financial-tables/observations/{name}.png docs/fixtures/remediation/financial-tables/{name}.pdf",
        "render_tool": f"Ghostscript {render_version}; png16m; 72 dpi",
        "render_exit_status": int(render_status),
        "render_output_path": f"docs/fixtures/remediation/financial-tables/observations/{name}.png",
        "render_output_sha256": render_sha,
    })
output_path.write_text(json.dumps({
    "schema_version": "1.0.0",
    "scope": "reproducible external text and rendering observations for synthetic fixtures",
    "observations": observations,
}, indent=2) + "\n")
PY
rm -f "$records"
