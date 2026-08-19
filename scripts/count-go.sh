#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

files="$(find . -type f -name '*.go' \
  | grep -v '/vendor/' \
  | grep -v '_test\.go$' \
  | sort)"

count="$(printf '%s\n' "$files" | grep -c . || true)"
lines="$(printf '%s\n' "$files" | xargs wc -l | tail -n 1 | awk '{print $1}')"

printf 'files=%s\n' "$count"
printf 'lines=%s\n' "$lines"
