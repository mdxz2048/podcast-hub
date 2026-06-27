#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-.env.user-beta}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.user-beta.yml}"
SERVICE="${POSTGRES_SERVICE:-postgres}"
DB_NAME="${POSTGRES_DB:-podcast_hub}"
DB_USER="${POSTGRES_USER:-podcast_hub}"
BACKUP_FILE=""
CONFIRM=0

for arg in "$@"; do
  case "$arg" in
    --env-file=*) ENV_FILE="${arg#--env-file=}" ;;
    --backup=*) BACKUP_FILE="${arg#--backup=}" ;;
    --confirm-restore) CONFIRM=1 ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

if [[ -z "$BACKUP_FILE" || ! -f "$BACKUP_FILE" ]]; then
  echo "Provide --backup=/path/to/backup.sql" >&2
  exit 2
fi

if [[ "$CONFIRM" != "1" ]]; then
  echo "Restore is destructive. Re-run with --confirm-restore after verifying the target database."
  exit 2
fi

echo "Restoring PostgreSQL backup from $BACKUP_FILE"
docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" psql -U "$DB_USER" "$DB_NAME" < "$BACKUP_FILE"
echo "Restore complete. Passwords were not printed."
