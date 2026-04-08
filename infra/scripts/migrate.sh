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

db_url="${DATABASE_URL}"
network_arg=""

if [ -n "${DOCKER_NETWORK:-}" ]; then
  network_arg="--network ${DOCKER_NETWORK}"
else
  case "$(uname -s)" in
    Linux*)
      network_arg="--network host"
      ;;
    *)
      db_url=$(printf '%s' "${db_url}" | sed 's/@localhost:/@host.docker.internal:/g; s/@127\.0\.0\.1:/@host.docker.internal:/g')
      ;;
  esac
fi

docker run --rm \
  ${network_arg} \
  -v "${MIGRATIONS_DIR}:/migrations" \
  migrate/migrate:v4.19.0 \
  -path=/migrations \
  -database "${db_url}" \
  "${command}" "$@"
