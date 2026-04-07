#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required"
  exit 1
fi

if [ -z "${MIGRATIONS_DIR:-}" ]; then
  echo "MIGRATIONS_DIR is required"
  exit 1
fi

command=${1:-}
shift || true

if [ -z "${command}" ]; then
  echo "usage: migrate.sh <up|down|force|goto|version|drop> [args...]"
  exit 1
fi

docker run --rm \
  --network host \
  -v "${MIGRATIONS_DIR}:/migrations" \
  migrate/migrate:v4.19.0 \
  -path=/migrations \
  -database "${DATABASE_URL}" \
  "${command}" "$@"