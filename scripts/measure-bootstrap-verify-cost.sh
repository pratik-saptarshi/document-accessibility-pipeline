#!/usr/bin/env bash
# Measure repeated bootstrap verification cost (dap-5hy.3).
#
# Synthetic fixture ONLY: a generated buildtools.manifest.json referencing
# local stub tools served over a loopback TLS server with a throwaway
# self-signed CA. No real network fetches, no system-wide installs.
# All measurement state lives under a fresh temp dir (chmod 700).
#
# Matrix: (bootstrap-only | bootstrap + separate --verify pass)
#       x (cold cache | warm cache)
#       x (checkout flow | no-checkout flow), >=5 repetitions per cell.
#
# Usage:
#   MEASURE_REPS=5 ./scripts/measure-bootstrap-verify-cost.sh --out <results.txt>
set -euo pipefail

REPS="${MEASURE_REPS:-5}"
OUT=""
CONTROLS_ONLY=false
while (( $# > 0 )); do
  case "$1" in
    --out) OUT="${2:?--out needs a path}"; shift 2 ;;
    --controls-only) CONTROLS_ONLY=true; shift ;;
    *) printf 'unknown argument: %s\n' "$1" >&2; exit 2 ;;
  esac
done
if [[ -z "$OUT" && "$CONTROLS_ONLY" == false ]]; then
  printf 'usage: %s --out <results.txt>\n' "${BASH_SOURCE[0]}" >&2
  exit 2
fi
if ! [[ "$REPS" =~ ^[0-9]+$ ]] || (( REPS < 5 )); then
  printf 'MEASURE_REPS must be an integer >= 5 (got %s)\n' "$REPS" >&2
  exit 2
fi

command -v curl >/dev/null || { printf 'curl is required\n' >&2; exit 2; }
command -v openssl >/dev/null || { printf 'openssl is required\n' >&2; exit 2; }

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/lib/benchmark-provenance.sh"
# Freeze exact executable bytes before any controls or timing work begins.
HARNESS_SHA="$(harness_file_sha256 "${BASH_SOURCE[0]}")"
HARNESS_SUPPORT_SHA="$(harness_file_sha256 "${SCRIPT_DIR}/lib/benchmark-provenance.sh")"
MODULE_DIR="${REPO_ROOT}/cmd/cloud-tools-bootstrap"
GO_VERSION_PIN="$(awk '$1 == "toolchain" { sub(/^go/, "", $2); print $2; exit }' "${REPO_ROOT}/go.mod")"
[[ -n "$GO_VERSION_PIN" ]] \
  || { printf 'go.mod must pin the benchmark toolchain version\n' >&2; exit 2; }
GO_BIN="$(resolve_host_go "$GO_VERSION_PIN")" \
  || { printf 'host Go must match go%s; set GO_BIN or put the pinned version on PATH\n' "$GO_VERSION_PIN" >&2; exit 2; }
GO_MAXPROCS="$(host_go_maxprocs "$GO_BIN")" \
  || { printf 'could not determine the host Go runtime GOMAXPROCS\n' >&2; exit 2; }

FIX="$(mktemp -d "${TMPDIR:-/tmp}/dap5hy3-measure.XXXXXX")"
chmod 700 "$FIX"
SRV_PID=""
cleanup() {
  if [[ -n "${SRV_PID}" ]]; then kill "${SRV_PID}" 2>/dev/null || true; fi
  rm -rf "$FIX"
}
trap cleanup EXIT

MANIFEST="${FIX}/manifest/buildtools.manifest.json"

# ---------------------------------------------------------------- fixtures
printf 'building synthetic stub fixtures in %s\n' "$FIX"
mkdir -p "${FIX}/manifest" "${FIX}/assets" "${FIX}/server" "${FIX}/work"

make_stub() { # name version
  local name="$1" version="$2" f="${FIX}/assets/${1}.sh"
  printf '#!/bin/sh\n# Synthetic stub tool for dap-5hy.3 measurements. Not a real tool.\nprintf "%%s version %s (synthetic stub fixture)\\n" "%s"\n' "$version" "$name" > "$f"
  chmod 755 "$f"
}
make_padded_stub() { # name version bytes
  python3 - "$FIX/assets/${1}.sh" "$1" "$2" "$3" <<'PY'
import sys
path, name, version, size = sys.argv[1], sys.argv[2], sys.argv[3], int(sys.argv[4])
prefix = f'#!/bin/sh\nprintf "%s version {version} (synthetic stub fixture)\\n" "{name}"\n# synthetic padding follows\n'.encode()
padding = b"#" * max(0, size - len(prefix) - 1) + b"\n"
with open(path, "wb") as out:
    out.write(prefix + padding)
PY
  chmod 755 "$FIX/assets/${1}.sh"
}
make_stub stub-raw-1 1.2.3
make_stub stub-raw-2 2.0.0
make_stub stub-raw-3 0.9.1
make_padded_stub stub-raw-large-4m 4.0.0 4194304

ARCH_STAGE="${FIX}/assets/archive-stage"
mkdir -p "$ARCH_STAGE"
printf '#!/bin/sh\n# Synthetic stub tool for dap-5hy.3 measurements. Not a real tool.\nprintf "stub-archive-1 version 3.4.5 (synthetic stub fixture)\\n"\n' > "${ARCH_STAGE}/stub-archive-1"
chmod 755 "${ARCH_STAGE}/stub-archive-1"
tar -czf "${FIX}/assets/stub-archive-1.tar.gz" -C "$ARCH_STAGE" stub-archive-1

# Synthetic manifest: real bootstrap_source pins (validated but never fetched
# in these flows), the real go bootstrap-toolchain entry, and four synthetic
# stub tools. Stub tools carry binary_sha256 (installed-binary digest check)
# plus version_args/version_pattern (exec version check), so they exercise
# the orchestrator's full checksum + version verification path.
PORT="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')"
python3 - "$FIX" "$PORT" <<'EOF'
import hashlib, json, os, sys
fix, port = sys.argv[1], sys.argv[2]
def sha(p):
    h = hashlib.sha256()
    with open(p, "rb") as f:
        h.update(f.read())
    return h.hexdigest()
asset = lambda n: os.path.join(fix, "assets", n)
stubs = [
    dict(name="stub-raw-1", version="1.2.3", kind="raw", file="stub-raw-1.sh"),
    dict(name="stub-raw-2", version="2.0.0", kind="raw", file="stub-raw-2.sh"),
    dict(name="stub-raw-3", version="0.9.1", kind="raw", file="stub-raw-3.sh"),
    dict(name="stub-raw-large-4m", version="4.0.0", kind="raw", file="stub-raw-large-4m.sh"),
]
tools = []
for s in stubs:
    p = asset(s["file"])
    d = sha(p)
    tools.append({
        "name": s["name"], "version": s["version"], "license": "Apache-2.0",
        "source": f"https://127.0.0.1:{port}/assets/{s['file']}",
        "sha256": d, "binary_sha256": d,
        "checksum_source": "synthetic local stub fixture (dap-5hy.3)",
        "kind": "raw", "binary": s["name"],
        "version_args": ["--version"],
        "version_pattern": r"(?i)\b" + s["version"].replace(".", r"\.") + r"\b",
    })
member = os.path.join(fix, "assets", "archive-stage", "stub-archive-1")
tools.append({
    "name": "stub-archive-1", "version": "3.4.5", "license": "Apache-2.0",
    "source": f"https://127.0.0.1:{port}/assets/stub-archive-1.tar.gz",
    "sha256": sha(asset("stub-archive-1.tar.gz")),
    "binary_sha256": sha(member),
    "checksum_source": "synthetic local stub fixture (dap-5hy.3)",
    "kind": "archive", "archive_member": "stub-archive-1", "binary": "stub-archive-1",
    "version_args": ["--version"],
    "version_pattern": r"(?i)\b3\.4\.5\b",
})
go_tool = {
    "name": "go", "version": "1.27.3", "license": "BSD-3-Clause",
    "source": "https://go.dev/dl/go1.27.3.linux-amd64.tar.gz",
    "sha256": "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40",
    "checksum_source": "https://go.dev/dl/",
    "kind": "bootstrap-toolchain", "binary": "go",
    "version_args": ["version"],
    "version_pattern": r"go1\.27\.3(?:[^0-9]|$)",
}
manifest = {
    "schema_version": 1,
    "platform": "linux/amd64",
    "bootstrap_source": {
        "revision": "ef4c35629c467887b6749f9e31af111cbe8f8913",
        "storage_revision": "05c9a4a710eef0bb6b525e7f853da7e874430a00",
        "path": "buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz",
        "url": "https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/05c9a4a710eef0bb6b525e7f853da7e874430a00/buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz",
        "sha256": "16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473",
        "size_bytes": 469726,
    },
    "tools": [go_tool] + tools,
}
mpath = os.path.join(fix, "manifest", "buildtools.manifest.json")
with open(mpath, "w") as f:
    json.dump(manifest, f, indent=2)
    f.write("\n")
print("wrote", mpath)
# The no-checkout flow builds the orchestrator from the current pinned source
# archive. Emit its matching current-schema manifest; the archive pin is the
# same as the checkout manifest and current bootstrap entrypoint.
manifest_nc = dict(manifest)
mpath_nc = os.path.join(fix, "manifest", "buildtools.manifest.nocheckout.json")
with open(mpath_nc, "w") as f:
    json.dump(manifest_nc, f, indent=2)
    f.write("\n")
print("wrote", mpath_nc)
# Wrong-version negative control: the alternate body has a correct downloadable
# and installed-binary SHA-256, but reports 9.9.9 instead of the pinned 1.2.3.
source = asset("stub-raw-1.sh")
wrong = asset("stub-wrong-version.sh")
with open(source, "rb") as f:
    body = f.read().replace(b"1.2.3", b"9.9.9")
with open(wrong, "wb") as f:
    f.write(body)
os.chmod(wrong, 0o755)
print("wrote wrong-version control with matching SHA-256", sha(wrong))
EOF

MANIFEST_NC="${FIX}/manifest/buildtools.manifest.nocheckout.json"

# ------------------------------------------------------- guard: synthetic only
# Refuse to measure against the real manifest or any real network host.
# Both fixture manifests (current schema for the checkout flow, archived
# schema for the no-checkout flow) are checked.
for GUARD_MANIFEST in "$MANIFEST" "$MANIFEST_NC"; do
python3 - "$GUARD_MANIFEST" "$FIX" <<'EOF'
import json, re, sys
mpath, fix = sys.argv[1], sys.argv[2]
if not mpath.startswith(fix + "/"):
    sys.exit("GUARD: manifest is not the synthetic fixture manifest: " + mpath)
m = json.load(open(mpath))
marker = "synthetic local stub fixture"
bad = []
for t in m["tools"]:
    if t.get("kind") == "bootstrap-toolchain":
        continue  # pinned host go entry; never fetched, version-checked only
    src = t.get("source", "")
    if not re.match(r"^https://127\.0\.0\.1(:\d+)?/", src):
        bad.append((t["name"], src))
    if marker not in t.get("checksum_source", ""):
        bad.append((t["name"], "missing synthetic marker"))
    if "stub" not in t["name"] or "stub" not in t.get("binary", ""):
        bad.append((t["name"], "not a stub tool"))
    if t.get("kind") == "go-install":
        bad.append((t["name"], "go-install would hit the real module proxy"))
    for host in ("github.com", "go.dev", "proxy.golang.org", "sum.golang.org",
                 "raw.githubusercontent.com"):
        if host in src:
            bad.append((t["name"], "real network host " + host))
    # bootstrap_source URLs are schema placeholders and are never fetched;
    # the guard only constrains fetched tool sources above.
if bad:
    for name, why in bad:
        print(f"GUARD: tool {name}: {why}", file=sys.stderr)
    sys.exit("GUARD: refusing to run against non-synthetic tool sources")
print("guard OK: synthetic manifest, loopback-only stub tool sources:", mpath)
EOF
done

# ------------------------------------------------- loopback TLS stub server
openssl req -x509 -newkey rsa:2048 -keyout "${FIX}/server/ca.key" \
  -out "${FIX}/server/ca.pem" -days 1 -nodes -subj "/CN=dap-5hy.3 stub CA" >/dev/null 2>&1
openssl req -newkey rsa:2048 -keyout "${FIX}/server/key.pem" \
  -out "${FIX}/server/csr.pem" -nodes -subj "/CN=127.0.0.1" >/dev/null 2>&1
cat > "${FIX}/server/ext.cnf" <<'EOF'
subjectAltName=IP:127.0.0.1
EOF
openssl x509 -req -in "${FIX}/server/csr.pem" -CA "${FIX}/server/ca.pem" \
  -CAkey "${FIX}/server/ca.key" -CAcreateserial -out "${FIX}/server/cert.pem" \
  -days 1 -extfile "${FIX}/server/ext.cnf" >/dev/null 2>&1
cat > "${FIX}/server/server.py" <<'EOF'
import functools, http.server, ssl, sys
port, root, cert, key = int(sys.argv[1]), sys.argv[2], sys.argv[3], sys.argv[4]
h = functools.partial(http.server.SimpleHTTPRequestHandler, directory=root)
srv = http.server.ThreadingHTTPServer(("127.0.0.1", port), h)
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain(cert, key)
srv.socket = ctx.wrap_socket(srv.socket, server_side=True)
srv.serve_forever()
EOF
# Serve the fixture root so /assets/<file> resolves; use -f so the readiness
# probe fails on HTTP errors (a 404 is not "up").
python3 "${FIX}/server/server.py" "$PORT" "${FIX}" \
  "${FIX}/server/cert.pem" "${FIX}/server/key.pem" \
  >/dev/null 2>&1 &
SRV_PID=$!
for _ in $(seq 1 100); do
  if curl -kf --max-time 1 -s -o /dev/null "https://127.0.0.1:${PORT}/assets/stub-raw-1.sh"; then
    break
  fi
  sleep 0.1
done
curl -kf --max-time 2 -s -o /dev/null "https://127.0.0.1:${PORT}/assets/stub-raw-1.sh" \
  || { printf 'stub server did not come up\n' >&2; exit 1; }
# Go's x509 honors SSL_CERT_FILE: trust only our throwaway stub CA.
export SSL_CERT_FILE="${FIX}/server/ca.pem"
printf 'stub server up on 127.0.0.1:%s (self-signed, trusted via SSL_CERT_FILE)\n' "$PORT"

# --------------------------------- source-archive fixture (no-checkout flow)
# The real scripts/bootstrap-cloud-build-tools.sh no-checkout branch, on EVERY
# invocation: verifies the pinned source archive, extracts it
# (tar -xzf --strip-components=1) into a fresh tmp dir, rebuilds the
# orchestrator (go build -buildvcs=false -trimpath), installs it into the
# tools root, then runs it for bootstrap and again with --verify. The harness
# mirrors that per timed arm: the extract+build below runs INSIDE the timer,
# once per arm, never once globally. The archive is the repo's real pinned
# source archive; its pin is verified against the manifest pin at fixture
# setup and the run fails closed on mismatch. No network fetch is performed:
# the real script would fetch from bootstrap_source_url only when the local
# archive fails the pin check.
SOURCE_ARCHIVE="${REPO_ROOT}/buildtools/source-archives/cloud-tools-bootstrap-ef4c35629c467887b6749f9e31af111cbe8f8913.tar.gz"
SOURCE_ARCHIVE_SHA256="16a72c55c81b96da07555a441330a408c42ed640f66fa3a60e8a84ee0f406473"
SOURCE_ARCHIVE_SIZE=469726
manifest_pin="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["bootstrap_source"]["sha256"])' "$MANIFEST")"
manifest_nc_pin="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["bootstrap_source"]["sha256"])' "$MANIFEST_NC")"
[[ "$manifest_pin" == "$SOURCE_ARCHIVE_SHA256" ]] \
  || { printf 'fixture pin drift: manifest bootstrap_source.sha256=%s\n' "$manifest_pin" >&2; exit 1; }
[[ "$manifest_nc_pin" == "$SOURCE_ARCHIVE_SHA256" ]] \
  || { printf 'fixture pin drift: no-checkout manifest bootstrap_source.sha256=%s\n' "$manifest_nc_pin" >&2; exit 1; }
verify_archive_fixture "$SOURCE_ARCHIVE" "$SOURCE_ARCHIVE_SHA256" "$SOURCE_ARCHIVE_SIZE" \
  || { printf 'source-archive fixture pin mismatch; refusing to measure\n' >&2; exit 1; }
printf 'source-archive fixture SHA-256 and size verified at setup\n'

# prepare_no_checkout_binary: per-arm extract+rebuild of the pinned source
# archive, mirroring the real script's no-checkout branch. Installs the built
# orchestrator into the arm's tools-root bin/ (as the script installs into
# tools_root/bin on every invocation). Called INSIDE the timed region, once
# per timed arm. GOCACHE/GOMODCACHE follow the real script's cache_root
# semantics (cache/go/build, cache/go/mod under the orchestrator --cache
# dir), so cold arms get genuinely cold builds and warm arms reuse the seeded
# cache.
prepare_no_checkout_binary() { # tools_root cache
  local tools_root="$1" cache="$2"
  local stage bindir srcdir
  stage="$(mktemp -d "${FIX}/stage-ncbin.XXXXXX")" || return 1
  bindir="${tools_root}/bin"
  mkdir -p "$stage/source" "$bindir" || { rm -rf "$stage"; return 1; }
  tar -xzf "$SOURCE_ARCHIVE" -C "$stage/source" --strip-components=1 \
    || { rm -rf "$stage"; return 1; }
  srcdir="$stage/source/cmd/cloud-tools-bootstrap"
  [[ -f "$srcdir/main.go" ]] \
    || { printf 'extracted source archive missing cmd/cloud-tools-bootstrap/main.go\n' >&2; rm -rf "$stage"; return 1; }
  ( cd "$srcdir" && env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
      GOMODCACHE="$cache/go/mod" GOCACHE="$cache/go/build" \
      "$GO_BIN" build -buildvcs=false -trimpath -o "$bindir/cloud-tools-bootstrap" . ) \
    || { rm -rf "$stage"; return 1; }
  chmod 0755 "$bindir/cloud-tools-bootstrap" || { rm -rf "$stage"; return 1; }
  rm -rf "$stage"
  return 0
}

invoke_bootstrap() { # flow root cache [extra args...]
  local flow="$1" root="$2" cache="$3"; shift 3
# The no-checkout archive contains the current orchestrator source and supports
# the same current-schema synthetic manifest as the checkout flow.
  local man="$MANIFEST"
  if [[ "$flow" == no-checkout ]]; then man="$MANIFEST_NC"; fi
  if [[ "$flow" == checkout ]]; then
    # Mirrors the real script's local-checkout branch (`go run`), with
    # GOCACHE/GOMODCACHE under the arm's cache dir (the script's cache_root
    # semantics): cold arms compile from scratch, warm arms reuse the seed.
    ( cd "$MODULE_DIR" && env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
        GOMODCACHE="$cache/go/mod" GOCACHE="$cache/go/build" \
        "$GO_BIN" run . --manifest "$man" --root "$root" --cache "$cache" \
        --go "$GO_BIN" "$@" )
  else
    # Mirrors the real script's no-checkout branch: the binary installed by
    # the per-arm prepare step above is executed directly.
    "$root/bin/cloud-tools-bootstrap" --manifest "$man" --root "$root" \
      --cache "$cache" --go "$GO_BIN" "$@"
  fi
}

# Acceptance controls run before any timing. A fresh binary built from the
# current pinned no-checkout archive must pass valid bootstrap/verify, reject
# bad source checksums, reject a wrong version whose bytes match both digest
# pins, reject an installed wrong version whose digest is explicitly pinned,
# and preserve CLI misuse exit 2.
CONTROL_ROOT="${FIX}/controls/tools"
CONTROL_CACHE="${FIX}/controls/cache"
CONTROL_MANIFEST="${FIX}/controls/manifest"
mkdir -p "$CONTROL_ROOT" "$CONTROL_CACHE" "$CONTROL_MANIFEST"
prepare_no_checkout_binary "$CONTROL_ROOT" "$CONTROL_CACHE" >/dev/null
invoke_bootstrap no-checkout "$CONTROL_ROOT" "$CONTROL_CACHE" >/dev/null
invoke_bootstrap no-checkout "$CONTROL_ROOT" "$CONTROL_CACHE" --verify >/dev/null

python3 - "$MANIFEST_NC" "$CONTROL_MANIFEST" "$FIX" <<'PY'
import copy, hashlib, json, os, sys
base_path, out_dir, fix = sys.argv[1:]
with open(base_path) as f:
    base = json.load(f)
raw = next(t for t in base["tools"] if t["name"] == "stub-raw-1")
bad_checksum = copy.deepcopy(base)
next(t for t in bad_checksum["tools"] if t["name"] == "stub-raw-1")["sha256"] = "0" * 64
with open(os.path.join(out_dir, "bad-checksum.json"), "w") as f:
    json.dump(bad_checksum, f)
    f.write("\n")

wrong_path = os.path.join(fix, "assets", "stub-wrong-version.sh")
h = hashlib.sha256()
with open(wrong_path, "rb") as f:
    h.update(f.read())
wrong_sha = h.hexdigest()
wrong_version = copy.deepcopy(base)
wrong_tool = next(t for t in wrong_version["tools"] if t["name"] == "stub-raw-1")
wrong_tool["source"] = wrong_tool["source"].replace("stub-raw-1.sh", "stub-wrong-version.sh")
wrong_tool["sha256"] = wrong_sha
wrong_tool["binary_sha256"] = wrong_sha
with open(os.path.join(out_dir, "wrong-source-version.json"), "w") as f:
    json.dump(wrong_version, f)
    f.write("\n")

verify_wrong = copy.deepcopy(base)
next(t for t in verify_wrong["tools"] if t["name"] == "stub-raw-1")["binary_sha256"] = wrong_sha
with open(os.path.join(out_dir, "wrong-installed-version.json"), "w") as f:
    json.dump(verify_wrong, f)
    f.write("\n")
PY

expect_control_failure() { # expected_exit expected_output_pattern label args...
  local expected_exit="$1" expected_pattern="$2" label="$3" output rc
  shift 3
  if output="$("$@" 2>&1)"; then rc=0; else rc=$?; fi
  [[ "$rc" == "$expected_exit" ]] \
    || { printf 'acceptance control %s exit=%s; expected %s\n%s\n' "$label" "$rc" "$expected_exit" "$output" >&2; exit 1; }
  [[ "$output" == *"$expected_pattern"* ]] \
    || { printf 'acceptance control %s output lacks %s\n%s\n' "$label" "$expected_pattern" "$output" >&2; exit 1; }
}

expect_control_failure 1 'SHA-256 mismatch' 'source checksum mismatch' \
  "$CONTROL_ROOT/bin/cloud-tools-bootstrap" --manifest "$CONTROL_MANIFEST/bad-checksum.json" \
  --root "${FIX}/controls/checksum-tools" --cache "${FIX}/controls/checksum-cache" --go "$GO_BIN"
expect_control_failure 1 'version validation' 'wrong source version with matching hashes' \
  "$CONTROL_ROOT/bin/cloud-tools-bootstrap" --manifest "$CONTROL_MANIFEST/wrong-source-version.json" \
  --root "${FIX}/controls/wrong-source-tools" --cache "${FIX}/controls/wrong-source-cache" --go "$GO_BIN"
wrong_version_body="${FIX}/assets/stub-wrong-version.sh"
cp "$wrong_version_body" "$CONTROL_ROOT/bin/stub-raw-1"
expect_control_failure 1 'NOT VERIFIED' 'installed wrong version with matching hash' \
  "$CONTROL_ROOT/bin/cloud-tools-bootstrap" --manifest "$CONTROL_MANIFEST/wrong-installed-version.json" \
  --root "$CONTROL_ROOT" --cache "$CONTROL_CACHE" --go "$GO_BIN" --verify
expect_control_failure 2 '--go must point' 'CLI flag misuse' \
  "$CONTROL_ROOT/bin/cloud-tools-bootstrap" --manifest "$MANIFEST_NC" \
  --root "${FIX}/controls/misuse-tools" --cache "${FIX}/controls/misuse-cache"
if ! ( cd "$MODULE_DIR" && env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
    GOMODCACHE="${FIX}/controls/go/mod" GOCACHE="${FIX}/controls/go/build" \
    "$GO_BIN" test ./internal/cloudtools \
      -run '^TestVerifyInstalledValidatesEveryManifestToolEntry$' -count=1 >/dev/null ); then
  printf 'multi-entry checksum/version validation regression failed\n' >&2
  exit 1
fi
CONTROL_SUMMARY='valid bootstrap=0 valid verify=0 source-checksum mismatch=1 wrong source version with matching hashes=1 installed wrong version with matching hash=1 multi-entry checksum/version validation=pass flag misuse=2'
printf 'acceptance controls passed: %s\n' "$CONTROL_SUMMARY"
if [[ "$CONTROLS_ONLY" == true ]]; then
  exit 0
fi

RESULTS_TMP="${OUT}.tmp"
{
  printf '# bootstrap verification cost — raw measurements (dap-5hy.3)\n'
  printf '# provenance: generated %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '# provenance: %s\n' "$("$GO_BIN" version)"
  printf '# provenance: %s\n' "$(uname -a)"
  printf '# provenance: cpus=%s gomaxprocs=%s (runtime.GOMAXPROCS(0))\n' "$(nproc)" "$GO_MAXPROCS"
  printf '# provenance: reps_per_cell=%s\n' "$REPS"
  printf '# provenance: harness_sha256=%s (executed scripts/measure-bootstrap-verify-cost.sh bytes at run start)\n' "$HARNESS_SHA"
  printf '# provenance: harness_support_sha256=%s (sourced scripts/lib/benchmark-provenance.sh bytes at run start)\n' "$HARNESS_SUPPORT_SHA"
  printf '# fixture: synthetic manifest + local stub tools over https://127.0.0.1:%s (throwaway self-signed CA); no real network\n' "$PORT"
  printf '# stub tools: three small raw stubs, stub-raw-large-4m (4 MiB synthetic shell script), stub-archive-1, and go bootstrap-toolchain entry\n'
  printf '# each stub exercises the checksum (binary_sha256) + version exec checks in the orchestrator installedOK path\n'
  printf '# acceptance controls passed before timing: %s\n' "$CONTROL_SUMMARY"
  printf '# modes: bootstrap = single bootstrap invocation; verify = bootstrap followed by a separate --verify pass\n'
  printf '# order: paired and interleaved by repetition; odd pairs run bootstrap then verify, even pairs verify then bootstrap\n'
  printf '# flows: checkout = `go run` from the module checkout (real script local-checkout branch); no-checkout = extract+rebuild the pinned source archive per arm inside the timer, then run the built binary (real script no-checkout branch; archive SHA-256 and size checked once at setup)\n'
  printf '# caches: cold = empty cache + empty tools root (genuinely cold per-arm go builds); warm = seeded by one prior bootstrap (tools installed, cache populated)\n'
  printf '\n'

  declare -A MEDIAN=() SAMPLES=() RECORDS=() REP_SAMPLE=()
  for flow in checkout no-checkout; do
    for cache in cold warm; do
      seeddir="${FIX}/work/${flow}/warm/seed"
      if [[ "$cache" == warm && ! -f "${seeddir}/.seeded" ]]; then
        mkdir -p "${seeddir}/tools" "${seeddir}/cache"
        if [[ "$flow" == no-checkout ]]; then
          prepare_no_checkout_binary "${seeddir}/tools" "${seeddir}/cache" >/dev/null \
            || { printf 'warm seed prepare failed\n' >&2; exit 1; }
        fi
        invoke_bootstrap "$flow" "${seeddir}/tools" "${seeddir}/cache" >/dev/null
        touch "${seeddir}/.seeded"
      fi
      for ((i=1; i<=REPS; i++)); do
        if (( i % 2 == 1 )); then
          pair_modes=(bootstrap verify)
        else
          pair_modes=(verify bootstrap)
        fi
        position=0
        for mode in "${pair_modes[@]}"; do
          position=$((position + 1))
          cell_dir="${FIX}/work/${flow}/${cache}/${mode}"
          mkdir -p "$cell_dir"
          if [[ "$cache" == warm ]]; then
            root="${seeddir}/tools"; cdir="${seeddir}/cache"
          else
            root="${cell_dir}/rep${i}/tools"; cdir="${cell_dir}/rep${i}/cache"
            rm -rf "$root" "$cdir"; mkdir -p "$root" "$cdir"
          fi
          start="$(date +%s.%N)"
          rc=0
          if [[ "$flow" == no-checkout ]]; then
            # One pinned archive extract+build per timed arm, matching the
            # current no-checkout script which builds one CLI for two calls.
            prepare_no_checkout_binary "$root" "$cdir" || rc=$?
          fi
          if (( rc == 0 )); then
            invoke_bootstrap "$flow" "$root" "$cdir" >/dev/null || rc=$?
            if [[ "$mode" == verify ]] && (( rc == 0 )); then
              invoke_bootstrap "$flow" "$root" "$cdir" --verify >/dev/null || rc=$?
            fi
          fi
          end="$(date +%s.%N)"
          if (( rc != 0 )); then
            printf 'rep %d mode=%s FAILED rc=%d — aborting study (no half-measured data)\n' "$i" "$mode" "$rc" >&2
            exit 1
          fi
          elapsed="$(awk -v s="$start" -v e="$end" 'BEGIN{printf "%.6f", e-s}')"
          key="${flow}/${cache}/${mode}"
          SAMPLES[$key]="${SAMPLES[$key]-}${elapsed} "
          REP_SAMPLE["${key}/${i}"]="$elapsed"
          RECORDS[$key]+="rep ${i} ${elapsed}s rc=0 pair_position=${position}"$'\n'
          if [[ "$cache" == cold ]]; then
            # Keep /tmp bounded: no later cold rep may reuse its build cache.
            rm -rf "$cdir/go"
          fi
        done
      done

      for mode in bootstrap verify; do
        key="${flow}/${cache}/${mode}"
        printf '## cell flow=%s mode=%s cache=%s\n' "$flow" "$mode" "$cache"
        printf '%s' "${RECORDS[$key]}"
        read -r -a times <<< "${SAMPLES[$key]}"
        stats="$(python3 - "${times[@]}" <<'EOF'
import statistics, sys
ts = [float(x) for x in sys.argv[1:]]
print(f"n={len(ts)} min={min(ts):.6f} median={statistics.median(ts):.6f} "
      f"mean={statistics.fmean(ts):.6f} max={max(ts):.6f}")
median = statistics.median(ts)
print(f"MEDIAN={median:.6f}")
EOF
)"
        printf '## stats %s\n' "$stats"
        median="$(printf '%s\n' "$stats" | grep '^MEDIAN=' | cut -d= -f2)"
        MEDIAN["${flow}/${cache}/${mode}"]="$median"
        printf '\n'
      done
    done
  done

  printf '## verify-pass deltas (median of bootstrap+verify minus median of bootstrap-only)\n'
  printf '# relative increase %% = delta / bootstrap-only median; share of combined %% = delta / bootstrap+verify median\n'
  for flow in checkout no-checkout; do
    for cache in cold warm; do
      b="${MEDIAN[${flow}/${cache}/bootstrap]}"
      v="${MEDIAN[${flow}/${cache}/verify]}"
      delta="$(awk -v b="$b" -v v="$v" 'BEGIN{printf "%.6f", v-b}')"
      read -r pct share <<< "$(benchmark_delta_percentages "$b" "$v")"
      printf 'delta %s %s: bootstrap=%.6fs verify=%.6fs delta=%ss relative_increase=%s%% share_of_combined=%s%%\n' \
        "$flow" "$cache" "$b" "$v" "$delta" "$pct" "$share"
    done
  done

  printf '\n## Paired-delta sensitivity analysis\n'
  printf '# median of per-repetition verify-minus-bootstrap deltas; sensitivities use marginal bootstrap/verify medians as denominators\n'
  for flow in checkout no-checkout; do
    for cache in cold warm; do
      paired_samples=()
      for ((i=1; i<=REPS; i++)); do
        b="${REP_SAMPLE[${flow}/${cache}/bootstrap/${i}]}"
        v="${REP_SAMPLE[${flow}/${cache}/verify/${i}]}"
        paired_samples+=("$(awk -v b="$b" -v v="$v" 'BEGIN{printf "%.6f", v-b}')")
      done
      pair_median="$(python3 - "${paired_samples[@]}" <<'EOF'
import statistics, sys
print(f"{statistics.median(map(float, sys.argv[1:])):.6f}")
EOF
)"
      b_median="${MEDIAN[${flow}/${cache}/bootstrap]}"
      v_median="${MEDIAN[${flow}/${cache}/verify]}"
      read -r relative share <<< "$(awk -v d="$pair_median" -v b="$b_median" -v v="$v_median" 'BEGIN {
        if (b == 0 || v == 0) exit 1
        printf "%.1f %.1f\n", d/b*100, d/v*100
      }')"
      printf 'paired-delta sensitivity %s %s: samples=%s median=%ss relative_to_bootstrap_median=%s%% share_of_combined_median=%s%%\n' \
        "$flow" "$cache" "$(IFS=,; printf '%s' "${paired_samples[*]}")" \
        "$pair_median" "$relative" "$share"
    done
  done
} > "$RESULTS_TMP"
mv "$RESULTS_TMP" "$OUT"
printf '\nwrote raw results to %s\n' "$OUT"
