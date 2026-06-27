#!/usr/bin/env bash
set -euo pipefail

APPLY=0
ENV_FILE="${ENV_FILE:-.env.user-beta}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.user-beta.yml}"
SERVICE="${POSTGRES_SERVICE:-postgres}"
DB_NAME="${POSTGRES_DB:-podcast_hub}"
DB_USER="${POSTGRES_USER:-podcast_hub}"
OLDER_THAN_DAYS="${OLDER_THAN_DAYS:-30}"

for arg in "$@"; do
  case "$arg" in
    --dry-run) APPLY=0 ;;
    --apply) APPLY=1 ;;
    --env-file=*) ENV_FILE="${arg#--env-file=}" ;;
    --older-than-days=*) OLDER_THAN_DAYS="${arg#--older-than-days=}" ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

if ! [[ "$OLDER_THAN_DAYS" =~ ^[0-9]+$ ]] || [[ "$OLDER_THAN_DAYS" -lt 1 ]]; then
  echo "--older-than-days must be a positive integer" >&2
  exit 2
fi

echo "Staging cleanup (${APPLY:-0} = apply). Sensitive paths and storage keys are not printed."
[[ "$APPLY" == "1" ]] || echo "Dry run only. Re-run with --apply to delete eligible stale staging rows."

sql_count="
SELECT COUNT(*)
FROM programs
WHERE status IN ('staging','rejected')
  AND updated_at < NOW() - ($OLDER_THAN_DAYS || ' days')::interval
  AND NOT EXISTS (
    SELECT 1 FROM import_jobs j
    WHERE j.id = programs.created_from_job_id
      AND j.status IN ('queued','running')
  );
"

count="$(docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" psql -qAt -U "$DB_USER" -d "$DB_NAME" -c "$sql_count")"
echo "Eligible stale staging Programs: $count"

if [[ "$APPLY" != "1" || "$count" == "0" ]]; then
  exit 0
fi

docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" psql -q -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" <<SQL
WITH candidates AS (
  SELECT id
  FROM programs
  WHERE status IN ('staging','rejected')
    AND updated_at < NOW() - ('$OLDER_THAN_DAYS days')::interval
    AND NOT EXISTS (
      SELECT 1 FROM import_jobs j
      WHERE j.id = programs.created_from_job_id
        AND j.status IN ('queued','running')
    )
), audit AS (
  INSERT INTO publication_events(id, target_type, target_id, event_type, metadata_redacted)
  SELECT gen_random_uuid(), 'program', id, 'cleanup.staging_deleted', '{"retention":"stale_staging"}'::jsonb
  FROM candidates
)
DELETE FROM programs WHERE id IN (SELECT id FROM candidates);
SQL

echo "Staging cleanup applied."
