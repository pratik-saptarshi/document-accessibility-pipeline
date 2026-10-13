#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture_dir="$repo_root/docs/fixtures/remediation/financial-tables-v2"
output_dir="$fixture_dir/observations"
mkdir -p "$output_dir"

pdftotext_version="$(pdftotext -v 2>&1 | head -n 1)"
ghostscript_version="$(gs --version)"
records="$output_dir/.observations.records"
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

python3 - "$records" "$output_dir/observations.v2.json" "$pdftotext_version" "$ghostscript_version" <<'PY'
import json
import sys
from pathlib import Path

records_path = Path(sys.argv[1])
output_path = Path(sys.argv[2])
text_version, render_version = sys.argv[3:5]
observations = []
for line in records_path.read_text().splitlines():
    name, pdf_sha, text_sha, render_sha = line.split("\t")
    observations.append({
        "fixture": f"{name}.pdf",
        "pdf_sha256": pdf_sha,
        "text_command": f"pdftotext -layout docs/fixtures/remediation/financial-tables-v2/{name}.pdf docs/fixtures/remediation/financial-tables-v2/observations/{name}.txt",
        "text_tool": text_version,
        "text_exit_status": 0,
        "text_output_path": f"docs/fixtures/remediation/financial-tables-v2/observations/{name}.txt",
        "text_output_sha256": text_sha,
        "render_command": f"gs -q -dSAFER -dBATCH -dNOPAUSE -sDEVICE=png16m -r72 -sOutputFile=docs/fixtures/remediation/financial-tables-v2/observations/{name}.png docs/fixtures/remediation/financial-tables-v2/{name}.pdf",
        "render_tool": f"Ghostscript {render_version}; png16m; 72 dpi",
        "render_exit_status": 0,
        "render_output_path": f"docs/fixtures/remediation/financial-tables-v2/observations/{name}.png",
        "render_output_sha256": render_sha,
    })
output_path.write_text(json.dumps({
    "schema_version": "2.0.0",
    "scope": "independent text extraction and rendering observations for synthetic metric-complete v2 fixtures",
    "observations": observations,
}, indent=2) + "\n")
PY
rm -f "$records"
