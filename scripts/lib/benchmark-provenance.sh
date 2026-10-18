#!/usr/bin/env bash

harness_file_sha256() { # path
  [[ -f "$1" ]] || { printf 'harness file not found: %s\n' "$1" >&2; return 1; }
  sha256sum -- "$1" | cut -d' ' -f1
}

host_go_maxprocs() { # go_executable
  local go_bin="$1" probe_dir result
  probe_dir="$(mktemp -d "${TMPDIR:-/tmp}/dap-go-procs.XXXXXX")" || return 1
  cat > "${probe_dir}/main.go" <<'GO'
package main

import (
	"fmt"
	"runtime"
)

func main() { fmt.Println(runtime.GOMAXPROCS(0)) }
GO
  if ! result="$(GOCACHE="${probe_dir}/cache" GOTOOLCHAIN=local GOWORK=off GO111MODULE=off \
      "$go_bin" run "${probe_dir}/main.go" 2>/dev/null)"; then
    rm -rf "$probe_dir"
    printf 'could not query Go runtime GOMAXPROCS using %s\n' "$go_bin" >&2
    return 1
  fi
  rm -rf "$probe_dir"
  [[ "$result" =~ ^[1-9][0-9]*$ ]] \
    || { printf 'Go runtime returned invalid GOMAXPROCS: %s\n' "$result" >&2; return 1; }
  printf '%s\n' "$result"
}

verify_archive_fixture() { # path expected_sha256 expected_size_bytes
  local archive="$1" expected_sha="$2" expected_size="$3" actual_size
  [[ -f "$archive" ]] || { printf 'source-archive fixture missing: %s\n' "$archive" >&2; return 1; }
  [[ "$expected_sha" =~ ^[[:xdigit:]]{64}$ ]] \
    || { printf 'invalid source-archive SHA-256 pin\n' >&2; return 1; }
  [[ "$expected_size" =~ ^[0-9]+$ ]] \
    || { printf 'invalid source-archive size pin\n' >&2; return 1; }
  actual_size="$(wc -c < "$archive" | tr -d '[:space:]')"
  [[ "$actual_size" == "$expected_size" ]] \
    || { printf 'source-archive fixture size mismatch: expected %s, got %s\n' "$expected_size" "$actual_size" >&2; return 1; }
  printf '%s  %s\n' "$expected_sha" "$archive" | sha256sum --check --status \
    || { printf 'source-archive fixture checksum mismatch\n' >&2; return 1; }
}

benchmark_delta_percentages() { # bootstrap_only_seconds bootstrap_plus_verify_seconds
  awk -v b="$1" -v v="$2" 'BEGIN {
    delta = v - b
    relative = (b > 0) ? delta / b * 100 : 0
    combined = (v > 0) ? delta / v * 100 : 0
    printf "%.1f %.1f\n", relative, combined
  }'
}

resolve_host_go() { # expected_version
  local expected="$1" candidate output candidate_dir candidate_name
  candidate="${GO_BIN:-}"
  if [[ -z "$candidate" ]]; then
    candidate="$(command -v go || true)"
  elif [[ "$candidate" != */* ]]; then
    candidate="$(command -v -- "$candidate" || true)"
  fi
  if [[ "$candidate" == */* ]]; then
    candidate_dir="${candidate%/*}"
    candidate_name="${candidate##*/}"
  else
    candidate_dir=.
    candidate_name="$candidate"
  fi
  if [[ -n "$candidate" ]]; then
    candidate="$(cd -- "$candidate_dir" && pwd -P)/$candidate_name"
  fi
  [[ -n "$candidate" && -x "$candidate" ]] \
    || { printf 'host Go executable not found; set GO_BIN or put go on PATH\n' >&2; return 1; }
  output="$("$candidate" version 2>&1)" \
    || { printf 'could not run host Go executable at %s\n' "$candidate" >&2; return 1; }
  [[ "$output" =~ go${expected}([[:space:]]|$) ]] \
    || { printf 'host Go version mismatch: expected go%s, got %s\n' "$expected" "$output" >&2; return 1; }
  printf '%s\n' "$candidate"
}
