#!/usr/bin/env python3
"""Author only the four reviewed synthetic financial-table v2 PDFs.

This deliberately narrow fixture aid is not a general PDF parser or writer.
It accepts only the four exact v1 inputs listed in the v1 manifest, the
pinned pdfcpu Helvetica AFM, the expected 22-object classic-xref layout, and
the single existing minimal Helvetica dictionary.
"""

import hashlib
import json
import os
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
V1_MANIFEST = ROOT / "docs/fixtures/remediation/financial-table-experiment.v1.json"
V2_MANIFEST = ROOT / "docs/fixtures/remediation/financial-table-experiment.v2.json"
V1_DIR = ROOT / "docs/fixtures/remediation/financial-tables"
V2_DIR = ROOT / "docs/fixtures/remediation/financial-tables-v2"
PDFCPU_VERSION = "v0.16.2"
PDFCPU_COMMIT = "b41a41d5323f6119ff6a183ad3f287eb4a20a617"
AFM_RELATIVE = Path("github.com/pdfcpu/pdfcpu@v0.16.2/internal/corefont/Core14_AFMs/Helvetica.afm")


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def read_standard_helvetica_metrics(afm_path: Path) -> tuple[list[int], str]:
    afm = afm_path.read_bytes()
    glyph_by_code: dict[int, str] = {}
    width_by_glyph: dict[str, int] = {}
    row = re.compile(rb"C\s+(-?\d+)\s+;\s+WX\s+(\d+)\s+;\s+N\s+(\S+)\s+;")
    for line in afm.splitlines():
        match = row.match(line)
        if not match:
            continue
        code, width = int(match.group(1)), int(match.group(2))
        name = match.group(3).decode("ascii")
        width_by_glyph[name] = width
        if code >= 0:
            glyph_by_code[code] = name
    return [width_by_glyph.get(glyph_by_code.get(code, ""), 0) for code in range(256)], digest(afm)


def rewrite_one_fixture(source: Path, destination: Path, widths: list[int]) -> None:
    data = source.read_bytes()
    if len(data) > 64 * 1024 or not data.startswith(b"%PDF-2.0\n"):
        raise ValueError(f"unexpected bounded PDF input: {source}")
    objects: dict[int, bytes] = {}
    for match in re.finditer(rb"(?m)^(\d+) 0 obj\n(.*?)\nendobj\n", data, re.S):
        number = int(match.group(1))
        if number in objects:
            raise ValueError(f"duplicate object {number} in {source}")
        objects[number] = match.group(2)
    if set(objects) != set(range(1, 23)):
        raise ValueError(f"unexpected object inventory in {source}")
    if objects[5] != b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>":
        raise ValueError(f"unexpected source Helvetica dictionary in {source}")
    trailer = re.search(rb"trailer\n<< /Size 23 /Root (\d+) 0 R >>\nstartxref\n\d+\n%%EOF\n?$", data)
    if trailer is None:
        raise ValueError(f"unexpected source trailer in {source}")

    width_array = " ".join(str(value) for value in widths)
    objects[5] = (
        f"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /StandardEncoding "
        f"/FirstChar 0 /LastChar 255 /Widths [{width_array}] /FontDescriptor 23 0 R >>"
    ).encode("ascii")
    objects[23] = (
        b"<< /Type /FontDescriptor /FontName /Helvetica /Flags 32 "
        b"/FontBBox [-166 -225 1000 931] /ItalicAngle 0 /Ascent 718 "
        b"/Descent -207 /CapHeight 718 /StemV 88 >>"
    )

    output = bytearray(b"%PDF-2.0\n%\xe2\xe3\xcf\xd3\n")
    offsets = [0] * 24
    for number in range(1, 24):
        offsets[number] = len(output)
        output.extend(f"{number} 0 obj\n".encode("ascii"))
        output.extend(objects[number])
        output.extend(b"\nendobj\n")
    xref_offset = len(output)
    output.extend(b"xref\n0 24\n0000000000 65535 f \n")
    for offset in offsets[1:]:
        output.extend(f"{offset:010d} 00000 n \n".encode("ascii"))
    output.extend(
        f"trailer\n<< /Size 24 /Root {trailer.group(1).decode('ascii')} 0 R >>\n"
        f"startxref\n{xref_offset}\n%%EOF\n".encode("ascii")
    )
    destination.write_bytes(output)


def main() -> None:
    module_cache = Path(os.environ["GOMODCACHE"])
    afm_path = module_cache / AFM_RELATIVE
    if not afm_path.is_file():
        raise FileNotFoundError(f"pinned source not cached: {afm_path}")
    widths, afm_sha256 = read_standard_helvetica_metrics(afm_path)
    base = json.loads(V1_MANIFEST.read_text())
    v2 = json.loads(V2_MANIFEST.read_text())
    v2_by_family = {case["family"]: case for case in v2["cases"]}
    for case in base["cases"]:
        family = case["family"]
        stem = "statement" if family == "statements" else "invoice"
        expected = v2_by_family[family]
        for state in ("positive", "damaged"):
            source = ROOT / case[f"{state}_path"]
            source_bytes = source.read_bytes()
            if digest(source_bytes) != case[f"{state}_sha256"]:
                raise ValueError(f"v1 input hash changed: {source}")
            destination = V2_DIR / f"{stem}-{state}-v2.pdf"
            rewrite_one_fixture(source, destination, widths)
            if digest(destination.read_bytes()) != expected[f"{state}_sha256"]:
                raise ValueError(f"authored v2 hash differs from manifest: {destination}")
    metrics = v2["font_metrics"]
    if metrics["source_version"] != f"github.com/pdfcpu/pdfcpu {PDFCPU_VERSION}" or metrics["source_commit"] != PDFCPU_COMMIT:
        raise ValueError("v2 manifest pins a different pdfcpu source")
    if metrics["widths_1000_em"] != widths or metrics["afm_sha256"] != afm_sha256:
        raise ValueError("v2 manifest metrics differ from the pinned AFM")


if __name__ == "__main__":
    main()
