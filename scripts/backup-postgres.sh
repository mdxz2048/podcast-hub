#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-.env.user-beta}"
BACKUP_DIR="${BACKUP_DIR:-./backups/user-beta}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.user-beta.yml}"
SERVICE="${POSTGRES_SERVICE:-postgres}"
DB_NAME="${POSTGRES_DB:-podcast_hub}"
DB_USER="${POSTGRES_USER:-podcast_hub}"

for arg in "$@"; do
  case "$arg" in
    --env-file=*) ENV_FILE="${arg#--env-file=}" ;;
    --backup-dir=*) BACKUP_DIR="${arg#--backup-dir=}" ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

mkdir -p "$BACKUP_DIR"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="$BACKUP_DIR/podcast_hub_${stamp}.sql"

echo "Writing PostgreSQL backup to $target"
docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" pg_dump -U "$DB_USER" "$DB_NAME" > "$target"
chmod 0600 "$target"
echo "Backup complete. Passwords were not printed."
