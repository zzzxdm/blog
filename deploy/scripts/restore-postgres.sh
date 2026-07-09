#!/usr/bin/env sh
set -eu

if [ "${1:-}" = "" ]; then
  echo "usage: deploy/scripts/restore-postgres.sh backups/blog-YYYYmmdd-HHMMSS.sql" >&2
  exit 1
fi

COMPOSE_FILE="${COMPOSE_FILE:-deploy/docker-compose.yml}"
POSTGRES_USER="${POSTGRES_USER:-blog}"
POSTGRES_DB="${POSTGRES_DB:-blog}"
SOURCE="$1"

if [ ! -f "$SOURCE" ]; then
  echo "backup file not found: $SOURCE" >&2
  exit 1
fi

docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  < "$SOURCE"
