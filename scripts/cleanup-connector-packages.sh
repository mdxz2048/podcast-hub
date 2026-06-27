#!/usr/bin/env bash
set -euo pipefail

APPLY=0
ENV_FILE="${ENV_FILE:-.env.user-beta}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.user-beta.yml}"
SERVICE="${POSTGRES_SERVICE:-postgres}"
DB_NAME="${POSTGRES_DB:-podcast_hub}"
DB_USER="${POSTGRES_USER:-podcast_hub}"
CONNECTOR_PACKAGE_LOCAL_DIR="${CONNECTOR_PACKAGE_LOCAL_DIR:-.local/connector-packages}"
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

root="$(cd "$CONNECTOR_PACKAGE_LOCAL_DIR" 2>/dev/null && pwd -P || true)"
if [[ -z "$root" ]]; then
  echo "Connector package directory does not exist. Nothing to clean."
  exit 0
fi
case "$root" in
  /|/Users|/Users/*/Code|/tmp) echo "Refusing unsafe connector package root." >&2; exit 2 ;;
esac

active_keys_file="$(mktemp "${TMPDIR:-/tmp}/podcast-hub-active-connectors.XXXXXX")"
trap 'rm -f "$active_keys_file"' EXIT
docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" psql -qAt -U "$DB_USER" -d "$DB_NAME" -c "
SELECT package_storage_key
FROM connector_versions cv
JOIN connectors c ON c.id=cv.connector_id
WHERE c.status='active' AND cv.review_status IN ('pending_review','approved');
" > "$active_keys_file"

echo "Connector package cleanup (${APPLY:-0} = apply). Storage keys and absolute paths are not printed."
[[ "$APPLY" == "1" ]] || echo "Dry run only. Re-run with --apply to delete eligible unreferenced packages."

count=0
while IFS= read -r -d '' path; do
  rel="${path#"$root"/}"
  if grep -Fxq "$rel" "$active_keys_file"; then
    continue
  fi
  count=$((count + 1))
  printf 'Eligible package artifact: %s\n' "$(basename "$path")"
  if [[ "$APPLY" == "1" ]]; then
    rm -f -- "$path"
  fi
done < <(find "$root" -type f -mtime +"$OLDER_THAN_DAYS" -print0)

echo "Eligible package artifacts: $count"
[[ "$APPLY" == "1" ]] && echo "Connector package cleanup applied."
