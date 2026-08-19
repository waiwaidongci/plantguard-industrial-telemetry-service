#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PSQL="${PSQL:-psql}"
DATABASE_URL="${DATABASE_URL:-postgres://plantguard:plantguard@localhost:5432/plantguard?sslmode=disable}"

for file in "$ROOT"/migrations/postgres/*.sql; do
  echo "applying $(basename "$file")"
  "$PSQL" "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
done
