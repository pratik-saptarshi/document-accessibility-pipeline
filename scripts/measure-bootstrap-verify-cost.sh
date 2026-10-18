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
while (( $# > 0 )); do
  case "$1" in
    --out) OUT="${2:?--out needs a path}"; shift 2 ;;
    *) printf 'unknown argument: %s\n' "$1" >&2; exit 2 ;;
  esac
done
if [[ -z "$OUT" ]]; then
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
MODULE_DIR="${REPO_ROOT}/cmd/cloud-tools-bootstrap"
GO_BIN="${HOME}/sdk/go/bin/go"
[[ -x "$GO_BIN" ]] || { printf 'host go not found at %s\n' "$GO_BIN" >&2; exit 2; }

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
mkdir -p "${FIX}/manifest" "${FIX}/assets" "${FIX}/server" "${FIX}/work" \
         "${FIX}/gocache" "${FIX}/gomodcache" "${FIX}/bin"

make_stub() { # name version
  local name="$1" version="$2" f="${FIX}/assets/${1}.sh"
  printf '#!/bin/sh\n# Synthetic stub tool for dap-5hy.3 measurements. Not a real tool.\nprintf "%%s version %s (synthetic stub fixture)\\n" "%s"\n' "$version" "$name" > "$f"
  chmod 755 "$f"
}
make_stub stub-raw-1 1.2.3
make_stub stub-raw-2 2.0.0
make_stub stub-raw-3 0.9.1

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
    "name": "go", "version": "1.27.2", "license": "BSD-3-Clause",
    "source": "https://go.dev/dl/go1.27.2.linux-amd64.tar.gz",
    "sha256": "ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5",
    "checksum_source": "https://go.dev/dl/",
    "kind": "bootstrap-toolchain", "binary": "go",
    "version_args": ["version"],
    "version_pattern": r"go1\.27\.2(?:[^0-9]|$)",
}
manifest = {
    "schema_version": 1,
    "platform": "linux/amd64",
    "bootstrap_source": {
        "revision": "63126024666fd71b48c25df55d1e98924e5eb09a",
        "storage_revision": "02f2b449a50c1e24af0f11495080ee19e1077fa2",
        "path": "buildtools/source-archives/cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz",
        "url": "https://raw.githubusercontent.com/pratik-saptarshi/document-accessibility-pipeline/02f2b449a50c1e24af0f11495080ee19e1077fa2/buildtools/source-archives/cloud-tools-bootstrap-63126024666fd71b48c25df55d1e98924e5eb09a.tar.gz",
        "sha256": "6fcd40bb10e90dc43efdf01094a38260b5bdf36710ec7a26677477c254675ed5",
        "size_bytes": 306734,
    },
    "tools": [go_tool] + tools,
}
mpath = os.path.join(fix, "manifest", "buildtools.manifest.json")
with open(mpath, "w") as f:
    json.dump(manifest, f, indent=2)
    f.write("\n")
print("wrote", mpath)
EOF

# ------------------------------------------------------- guard: synthetic only
# Refuse to measure against the real manifest or any real network host.
python3 - "$MANIFEST" "$FIX" <<'EOF'
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
if bad:
    for name, why in bad:
        print(f"GUARD: tool {name}: {why}", file=sys.stderr)
    sys.exit("GUARD: refusing to run against non-synthetic tool sources")
print("guard OK: synthetic manifest, loopback-only stub tool sources")
EOF

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

# ------------------------------------------------- build orchestrator binary
# no-checkout flow runs the prebuilt binary (mirrors the script's immutable
# source-archive branch); checkout flow uses `go run` from the checkout
# (mirrors the script's local-checkout branch).
env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
    GOMODCACHE="${FIX}/gomodcache" GOCACHE="${FIX}/gocache" \
    "$GO_BIN" build -o /dev/null . >/dev/null 2>&1 || true
(
  cd "$MODULE_DIR"
  env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
      GOMODCACHE="${FIX}/gomodcache" GOCACHE="${FIX}/gocache" \
      "$GO_BIN" build -o /dev/null .
)
printf 'go run compile cache warmed\n'
(
  cd "$MODULE_DIR"
  env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
      GOMODCACHE="${FIX}/gomodcache" GOCACHE="${FIX}/gocache" \
      "$GO_BIN" build -buildvcs=false -trimpath -o "${FIX}/bin/cloud-tools-bootstrap" .
)
printf 'no-checkout binary built\n'

invoke_bootstrap() { # flow root cache [extra args...]
  local flow="$1" root="$2" cache="$3"; shift 3
  if [[ "$flow" == checkout ]]; then
    ( cd "$MODULE_DIR" && env GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=readonly \
        GOMODCACHE="${FIX}/gomodcache" GOCACHE="${FIX}/gocache" \
        "$GO_BIN" run . --manifest "$MANIFEST" --root "$root" --cache "$cache" \
        --go "$GO_BIN" "$@" )
  else
    "${FIX}/bin/cloud-tools-bootstrap" --manifest "$MANIFEST" --root "$root" \
      --cache "$cache" --go "$GO_BIN" "$@"
  fi
}

RESULTS_TMP="${OUT}.tmp"
{
  printf '# bootstrap verification cost — raw measurements (dap-5hy.3)\n'
  printf '# provenance: generated %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '# provenance: %s\n' "$("$GO_BIN" version)"
  printf '# provenance: %s\n' "$(uname -a)"
  # `go env GOMAXPROCS` is empty on this toolchain; Go defaults GOMAXPROCS to NumCPU.
  printf '# provenance: cpus=%s gomaxprocs=%s (Go default = NumCPU)\n' "$(nproc)" "${GOMAXPROCS:-$(nproc)}"
  printf '# provenance: reps_per_cell=%s\n' "$REPS"
  printf '# fixture: synthetic manifest + local stub tools over https://127.0.0.1:%s (throwaway self-signed CA); no real network\n' "$PORT"
  printf '# stub tools: stub-raw-1, stub-raw-2, stub-raw-3 (kind=raw), stub-archive-1 (kind=archive), plus go bootstrap-toolchain entry\n'
  printf '# each stub exercises the checksum (binary_sha256) + version exec checks in the orchestrator installedOK path\n'
  printf '# modes: bootstrap = single bootstrap invocation; verify = bootstrap followed by a separate --verify pass\n'
  printf '# caches: cold = empty cache + empty tools root; warm = seeded by one prior bootstrap (tools installed, cache populated)\n'
  printf '\n'

  declare -A MEDIAN=()
  for flow in checkout no-checkout; do
    for cache in cold warm; do
      for mode in bootstrap verify; do
        cell_label="flow=${flow} mode=${mode} cache=${cache}"
        printf '## cell %s\n' "$cell_label"
        celldir="${FIX}/work/${flow}/${cache}/${mode}"
        mkdir -p "$celldir"
        seeddir="${FIX}/work/${flow}/warm/seed"
        if [[ "$cache" == warm && ! -f "${seeddir}/.seeded" ]]; then
          mkdir -p "${seeddir}/tools" "${seeddir}/cache"
          invoke_bootstrap "$flow" "${seeddir}/tools" "${seeddir}/cache" >/dev/null
          touch "${seeddir}/.seeded"
          printf '# warm seed installed (untimed)\n'
        fi
        times=()
        for ((i=1; i<=REPS; i++)); do
          if [[ "$cache" == warm ]]; then
            root="${seeddir}/tools"; cdir="${seeddir}/cache"
          else
            root="${celldir}/rep${i}/tools"; cdir="${celldir}/rep${i}/cache"
            rm -rf "$root" "$cdir"; mkdir -p "$root" "$cdir"
          fi
          start="$(date +%s.%N)"
          rc=0
          if [[ "$mode" == bootstrap ]]; then
            invoke_bootstrap "$flow" "$root" "$cdir" >/dev/null || rc=$?
          else
            invoke_bootstrap "$flow" "$root" "$cdir" >/dev/null || rc=$?
            if (( rc == 0 )); then
              invoke_bootstrap "$flow" "$root" "$cdir" --verify >/dev/null || rc=$?
            fi
          fi
          end="$(date +%s.%N)"
          if (( rc != 0 )); then
            printf 'rep %d FAILED rc=%d — aborting cell (no half-measured data)\n' "$i" "$rc" >&2
            exit 1
          fi
          elapsed="$(awk -v s="$start" -v e="$end" 'BEGIN{printf "%.6f", e-s}')"
          times+=("$elapsed")
          printf 'rep %d %.6fs rc=0\n' "$i" "$elapsed"
        done
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
  for flow in checkout no-checkout; do
    for cache in cold warm; do
      b="${MEDIAN[${flow}/${cache}/bootstrap]}"
      v="${MEDIAN[${flow}/${cache}/verify]}"
      delta="$(awk -v b="$b" -v v="$v" 'BEGIN{printf "%.6f", v-b}')"
      pct="$(awk -v b="$b" -v v="$v" 'BEGIN{printf "%.1f", (b>0)?(v-b)/b*100:0}')"
      printf 'delta %s %s: bootstrap=%.6fs verify=%.6fs delta=+%ss (+%s%%)\n' \
        "$flow" "$cache" "$b" "$v" "$delta" "$pct"
    done
  done
} > "$RESULTS_TMP"
mv "$RESULTS_TMP" "$OUT"
printf '\nwrote raw results to %s\n' "$OUT"
