#!/usr/bin/env bash
# migrate.sh — thin wrapper around the golang-migrate CLI.
#
# Usage:
#   ./scripts/migrate.sh up      # apply all pending migrations
#   ./scripts/migrate.sh down    # roll back the most recent migration
#   ./scripts/migrate.sh down N  # roll back N migrations
#   ./scripts/migrate.sh version # print current migration version
#
# Required environment variable:
#   DATABASE_URL  — PostgreSQL connection string, e.g.
#                   postgres://user:pass@localhost:5432/agribid?sslmode=disable

set -euo pipefail

DIRECTION="${1:-up}"
STEPS="${2:-}"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "ERROR: DATABASE_URL environment variable is not set." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="${SCRIPT_DIR}/../migrations"

if [[ "${DIRECTION}" == "up" ]]; then
  migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" up ${STEPS}
elif [[ "${DIRECTION}" == "down" ]]; then
  migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" down ${STEPS:-1}
elif [[ "${DIRECTION}" == "version" ]]; then
  migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" version
else
  echo "Usage: $0 [up|down|version] [steps]" >&2
  exit 1
fi
