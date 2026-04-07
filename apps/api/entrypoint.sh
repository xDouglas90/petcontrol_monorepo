#!/bin/sh
set -eu

if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
  migrate -path /app/infra/migrations -database "${DATABASE_URL}" up
fi

exec /app/petcontrol-api