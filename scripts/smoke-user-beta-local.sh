#!/usr/bin/env bash
set -euo pipefail

APPLY=0
KEEP=0
PROJECT_NAME="${PROJECT_NAME:-podcast_hub_user_beta_smoke}"
API_PORT="${API_PORT:-18080}"

usage() {
  cat <<'USAGE'
Usage: scripts/smoke-user-beta-local.sh [--dry-run|--apply] [--keep] [--api-port=18080] [--project-name=name]

Runs a local-only User Beta deployment smoke test with generated temporary data.
Default mode is dry-run. --apply may start Docker Compose and will remove its
temporary project volumes at the end unless --keep is provided.
USAGE
}

for arg in "$@"; do
  case "$arg" in
    --dry-run) APPLY=0 ;;
    --apply) APPLY=1 ;;
    --keep) KEEP=1 ;;
    --api-port=*) API_PORT="${arg#--api-port=}" ;;
    --project-name=*) PROJECT_NAME="${arg#--project-name=}" ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $arg" >&2; usage >&2; exit 2 ;;
  esac
done

log() {
  printf '[user-beta-smoke] %s\n' "$1"
}

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required tool: $1" >&2
    exit 2
  fi
}

if [[ "$APPLY" != "1" ]]; then
  log "dry-run only; no containers, volumes, users, tokens, or backup files will be created"
  log "would start Compose project '$PROJECT_NAME' with API bound to 127.0.0.1:$API_PORT"
  log "would create temporary admin/user/no-access accounts and a published fixture Program/Episode/media"
  log "would verify catalog visibility, private media GET/HEAD/Range, RSS create/feed/enclosure, token rotate, grant revoke, backup, and restore into a temporary database"
  log "re-run with --apply to execute the local smoke"
  exit 0
fi

require_tool docker
require_tool curl
require_tool openssl
require_tool shasum
require_tool awk
require_tool python3

if ! docker info >/dev/null 2>&1; then
  echo "ENVIRONMENT BLOCKED: Docker daemon is not available." >&2
  exit 2
fi

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/podcast-hub-user-beta-smoke.XXXXXX")"
ENV_FILE="$WORK_DIR/.env.user-beta.smoke"
OVERRIDE_FILE="$WORK_DIR/compose.override.yml"
BACKUP_DIR="$WORK_DIR/backups"
mkdir -p "$BACKUP_DIR"

COMPOSE=(docker compose -p "$PROJECT_NAME" -f deploy/compose.user-beta.yml -f "$OVERRIDE_FILE" --env-file "$ENV_FILE")

cleanup() {
  local exit_code=$?
  if [[ "$KEEP" == "1" ]]; then
    log "keeping temporary Compose project '$PROJECT_NAME' and work dir $WORK_DIR"
  else
    "${COMPOSE[@]}" down -v --remove-orphans >/dev/null 2>&1 || true
    rm -rf "$WORK_DIR"
  fi
  exit "$exit_code"
}
trap cleanup EXIT

random_hex() {
  openssl rand -hex "$1"
}

hash_with_pepper() {
  local raw="$1"
  local pepper="$2"
  printf '%s:%s' "$raw" "$pepper" | shasum -a 256 | awk '{print $1}'
}

json_get() {
  python3 - "$1" "$2" <<'PY'
import json
import sys

path, expr = sys.argv[1], sys.argv[2]
data = json.load(open(path, encoding="utf-8"))
for part in expr.split("."):
    data = data[part]
print(data)
PY
}

expect_status() {
  local expected="$1"
  local url="$2"
  shift 2
  local code
  code="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' "$@" "$url")"
  if [[ "$code" != "$expected" ]]; then
    echo "expected HTTP $expected but got $code" >&2
    exit 1
  fi
}

expect_contains() {
  local file="$1"
  local text="$2"
  if ! grep -Fq "$text" "$file"; then
    echo "expected response to contain required fixture text" >&2
    exit 1
  fi
}

expect_not_contains() {
  local file="$1"
  local text="$2"
  if grep -Fq "$text" "$file"; then
    echo "response leaked fixture text after authorization was removed" >&2
    exit 1
  fi
}

SESSION_PEPPER="smoke_session_$(random_hex 16)"
AUTH_CODE_PEPPER="smoke_auth_$(random_hex 16)"
SECRETS_MASTER_KEY="$(random_hex 32)"
POSTGRES_PASSWORD="smoke_db_$(random_hex 12)"
CSRF_TOKEN="smoke_csrf_$(random_hex 12)"
USER_SESSION_TOKEN="smoke_user_$(random_hex 24)"
NO_ACCESS_SESSION_TOKEN="smoke_noaccess_$(random_hex 24)"
ADMIN_SESSION_TOKEN="smoke_admin_$(random_hex 24)"
USER_SESSION_HASH="$(hash_with_pepper "$USER_SESSION_TOKEN" "$SESSION_PEPPER")"
NO_ACCESS_SESSION_HASH="$(hash_with_pepper "$NO_ACCESS_SESSION_TOKEN" "$SESSION_PEPPER")"
ADMIN_SESSION_HASH="$(hash_with_pepper "$ADMIN_SESSION_TOKEN" "$SESSION_PEPPER")"

cat > "$ENV_FILE" <<EOF
APP_ENV=development
HTTP_ADDR=:8080
PUBLIC_BASE_URL=http://127.0.0.1:$API_PORT
FRONTEND_ORIGIN=http://127.0.0.1:$API_PORT
POSTGRES_DB=podcast_hub
POSTGRES_USER=podcast_hub
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
DATABASE_URL=postgresql://podcast_hub:$POSTGRES_PASSWORD@postgres:5432/podcast_hub?sslmode=disable
REDIS_URL=redis://redis:6379/0
SESSION_COOKIE_NAME=podcast_hub_session
SESSION_COOKIE_SECURE=false
SESSION_TTL_SECONDS=3600
SESSION_PEPPER=$SESSION_PEPPER
AUTH_CODE_PEPPER=$AUTH_CODE_PEPPER
CSRF_HEADER_NAME=X-CSRF-Token
TURNSTILE_MODE=mock
VITE_TURNSTILE_MODE=mock
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_FROM=smoke@example.invalid
SECRETS_MASTER_KEY=$SECRETS_MASTER_KEY
CONNECTOR_PACKAGE_LOCAL_DIR=/data/connector-packages
IMPORT_ARTIFACT_STORE_DIR=/data/import-artifacts
STAGING_STORE_DIR=/data/staging
MEDIA_STORE_DIR=/data/private-media
RUNNER_MODE=disabled
RUNNER_TRUSTED_ADMIN_ENABLED=false
BACKUP_DIR=$BACKUP_DIR
EOF
chmod 0600 "$ENV_FILE"

cat > "$OVERRIDE_FILE" <<EOF
services:
  api:
    ports:
      - "127.0.0.1:$API_PORT:8080"
EOF

log "starting local Compose project '$PROJECT_NAME'"
if ! "${COMPOSE[@]}" up -d --build postgres redis api; then
  echo "ENVIRONMENT BLOCKED: local Compose stack could not start. If the failure is Docker Hub metadata or image pull timeout, rerun when base images are cached or reachable." >&2
  exit 2
fi

BASE_URL="http://127.0.0.1:$API_PORT"
for _ in $(seq 1 60); do
  if curl --silent --fail "$BASE_URL/healthz" >/dev/null 2>&1 && curl --silent --fail "$BASE_URL/readyz" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl --silent --fail "$BASE_URL/healthz" >/dev/null
curl --silent --fail "$BASE_URL/readyz" >/dev/null
log "healthz and readyz passed"

ADMIN_ID="11111111-1111-4111-8111-111111111111"
USER_ID="22222222-2222-4222-8222-222222222222"
NO_ACCESS_USER_ID="33333333-3333-4333-8333-333333333333"
CONNECTOR_ID="44444444-4444-4444-8444-444444444444"
CONNECTOR_VERSION_ID="55555555-5555-4555-8555-555555555555"
SOURCE_ID="66666666-6666-4666-8666-666666666666"
JOB_ID="77777777-7777-4777-8777-777777777777"
ARTIFACT_ID="88888888-8888-4888-8888-888888888888"
PROGRAM_ID="99999999-9999-4999-8999-999999999999"
EPISODE_ID="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
MEDIA_ID="bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
GRANT_ID="cccccccc-cccc-4ccc-8ccc-cccccccccccc"

MEDIA_TEXT="Podcast Hub user beta smoke media fixture."
MEDIA_SIZE="$(printf '%s' "$MEDIA_TEXT" | wc -c | awk '{print $1}')"
MEDIA_SHA="$(printf '%s' "$MEDIA_TEXT" | shasum -a 256 | awk '{print $1}')"

"${COMPOSE[@]}" exec -T -u 0 api sh -c 'mkdir -p /data/private-media/smoke && cat > /data/private-media/smoke/episode.txt && chown -R 10001:10001 /data/private-media' <<<"$MEDIA_TEXT"

log "loading temporary users and fixture content"
"${COMPOSE[@]}" exec -T postgres psql -q -v ON_ERROR_STOP=1 -U podcast_hub -d podcast_hub <<SQL
DELETE FROM user_sessions WHERE user_id IN ('$ADMIN_ID', '$USER_ID', '$NO_ACCESS_USER_ID');
DELETE FROM program_access_grants WHERE id='$GRANT_ID';
DELETE FROM media_assets WHERE id='$MEDIA_ID';
DELETE FROM episodes WHERE id='$EPISODE_ID';
DELETE FROM programs WHERE id='$PROGRAM_ID';
DELETE FROM import_job_artifacts WHERE id='$ARTIFACT_ID';
DELETE FROM import_jobs WHERE id='$JOB_ID';
DELETE FROM connector_sources WHERE id='$SOURCE_ID';
DELETE FROM connector_versions WHERE id='$CONNECTOR_VERSION_ID';
DELETE FROM connectors WHERE id='$CONNECTOR_ID';
DELETE FROM users WHERE id IN ('$ADMIN_ID', '$USER_ID', '$NO_ACCESS_USER_ID');

INSERT INTO users(id, email_normalized, display_name, role, status, verified_at)
VALUES
  ('$ADMIN_ID', 'smoke-admin@example.invalid', 'Smoke Admin', 'admin', 'active', NOW()),
  ('$USER_ID', 'smoke-user@example.invalid', 'Smoke User', 'user', 'active', NOW()),
  ('$NO_ACCESS_USER_ID', 'smoke-no-access@example.invalid', 'No Access User', 'user', 'active', NOW());

INSERT INTO user_sessions(id, user_id, session_hash, created_at, last_seen_at, expires_at, ip_summary, user_agent_summary)
VALUES
  (gen_random_uuid(), '$ADMIN_ID', '$ADMIN_SESSION_HASH', NOW(), NOW(), NOW() + INTERVAL '1 hour', 'local-smoke', 'local-smoke'),
  (gen_random_uuid(), '$USER_ID', '$USER_SESSION_HASH', NOW(), NOW(), NOW() + INTERVAL '1 hour', 'local-smoke', 'local-smoke'),
  (gen_random_uuid(), '$NO_ACCESS_USER_ID', '$NO_ACCESS_SESSION_HASH', NOW(), NOW(), NOW() + INTERVAL '1 hour', 'local-smoke', 'local-smoke');

INSERT INTO connectors(id, slug, name, description, status, created_by)
VALUES ('$CONNECTOR_ID', 'smoke-fixture', 'Smoke Fixture Connector', 'Local smoke fixture only', 'active', '$ADMIN_ID');

INSERT INTO connector_versions(id, connector_id, version, review_status, runtime_profile, entrypoint, manifest_json, package_sha256, package_size_bytes, package_storage_key, validation_summary_json, uploaded_by, reviewed_by, reviewed_at)
VALUES ('$CONNECTOR_VERSION_ID', '$CONNECTOR_ID', '0.0.1-smoke', 'approved', 'python-basic', 'connector.py', '{"name":"smoke"}', repeat('0', 64), 1, 'smoke/connector.zip', '{"status":"ok"}', '$ADMIN_ID', '$ADMIN_ID', NOW());

INSERT INTO connector_sources(id, connector_version_id, name, description, status, trigger_type, auth_mode, execution_mode, config_json, network_mode, created_by)
VALUES ('$SOURCE_ID', '$CONNECTOR_VERSION_ID', 'Smoke Source', 'Local smoke source only', 'active', 'manual', 'none', 'unattended', '{}', 'disabled', '$ADMIN_ID');

INSERT INTO import_jobs(id, connector_source_id, connector_version_id, status, requested_by, trigger_type, auth_mode, execution_mode, started_at, finished_at)
VALUES ('$JOB_ID', '$SOURCE_ID', '$CONNECTOR_VERSION_ID', 'completed', '$ADMIN_ID', 'manual', 'none', 'unattended', NOW(), NOW());

INSERT INTO import_job_artifacts(id, import_job_id, artifact_type, relative_path, size_bytes, sha256, storage_key)
VALUES ('$ARTIFACT_ID', '$JOB_ID', 'media', 'episode.txt', $MEDIA_SIZE, '$MEDIA_SHA', 'smoke/source/episode.txt');

INSERT INTO programs(id, canonical_key, title, description, author, language, status, created_from_source_id, created_from_job_id, published_at)
VALUES ('$PROGRAM_ID', 'smoke-program', 'Smoke Program', 'Local User Beta smoke program', 'Podcast Hub', 'en-US', 'published', '$SOURCE_ID', '$JOB_ID', NOW());

INSERT INTO program_sources(id, program_id, connector_source_id, external_program_id, first_import_job_id, last_import_job_id)
VALUES (gen_random_uuid(), '$PROGRAM_ID', '$SOURCE_ID', 'smoke-program', '$JOB_ID', '$JOB_ID');

INSERT INTO episodes(id, program_id, external_episode_id, title, description, published_at, duration_seconds, status, source_job_id, published_to_users_at)
VALUES ('$EPISODE_ID', '$PROGRAM_ID', 'smoke-episode', 'Smoke Episode', 'Local User Beta smoke episode', NOW(), 3, 'published', '$JOB_ID', NOW());

INSERT INTO media_assets(id, owner_type, owner_id, import_job_id, artifact_id, media_kind, staged_storage_key, content_type, size_bytes, sha256, status, delivery_status, published_storage_key, published_at)
VALUES ('$MEDIA_ID', 'episode', '$EPISODE_ID', '$JOB_ID', '$ARTIFACT_ID', 'audio', 'smoke/staged/episode.txt', 'text/plain', $MEDIA_SIZE, '$MEDIA_SHA', 'published', 'published', 'smoke/episode.txt', NOW());

INSERT INTO program_access_grants(id, user_id, program_id, status, granted_by, reason)
VALUES ('$GRANT_ID', '$USER_ID', '$PROGRAM_ID', 'active', '$ADMIN_ID', 'local smoke');
SQL

USER_COOKIE="podcast_hub_session=$USER_SESSION_TOKEN; podcast_hub_csrf=$CSRF_TOKEN"
NO_ACCESS_COOKIE="podcast_hub_session=$NO_ACCESS_SESSION_TOKEN; podcast_hub_csrf=$CSRF_TOKEN"
ORIGIN_HEADER="Origin: $BASE_URL"
CSRF_HEADER="X-CSRF-Token: $CSRF_TOKEN"

curl --silent --show-error --fail -H "Cookie: $USER_COOKIE" "$BASE_URL/programs" > "$WORK_DIR/programs.json"
expect_contains "$WORK_DIR/programs.json" "Smoke Program"
curl --silent --show-error --fail -H "Cookie: $NO_ACCESS_COOKIE" "$BASE_URL/programs" > "$WORK_DIR/no-access-programs.json"
expect_not_contains "$WORK_DIR/no-access-programs.json" "Smoke Program"
curl --silent --show-error --fail -H "Cookie: $USER_COOKIE" "$BASE_URL/programs/$PROGRAM_ID/episodes" > "$WORK_DIR/episodes.json"
expect_contains "$WORK_DIR/episodes.json" "Smoke Episode"
log "authorized catalog checks passed"

curl --silent --show-error --fail -H "Cookie: $USER_COOKIE" "$BASE_URL/media/episodes/$EPISODE_ID" > "$WORK_DIR/user-media.txt"
expect_contains "$WORK_DIR/user-media.txt" "$MEDIA_TEXT"
expect_status 200 "$BASE_URL/media/episodes/$EPISODE_ID" -I -H "Cookie: $USER_COOKIE"
range_code="$(curl --silent --show-error --output "$WORK_DIR/user-media-range.txt" --write-out '%{http_code}' -H "Range: bytes=0-6" -H "Cookie: $USER_COOKIE" "$BASE_URL/media/episodes/$EPISODE_ID")"
if [[ "$range_code" != "206" ]]; then
  echo "expected HTTP 206 for user media range request, got $range_code" >&2
  exit 1
fi
log "private user media checks passed"

curl --silent --show-error --fail \
  -X POST \
  -H "$ORIGIN_HEADER" \
  -H "$CSRF_HEADER" \
  -H "Cookie: $USER_COOKIE" \
  -H "Content-Type: application/json" \
  --data '{"name":"Smoke Feed"}' \
  "$BASE_URL/me/rss-feeds" > "$WORK_DIR/feed-create.json"
FEED_ID="$(json_get "$WORK_DIR/feed-create.json" "feed.id")"
FEED_URL="$(json_get "$WORK_DIR/feed-create.json" "feed_url")"
FEED_PATH="$(python3 - "$FEED_URL" <<'PY'
from urllib.parse import urlparse
import sys
print(urlparse(sys.argv[1]).path)
PY
)"
RSS_TOKEN="${FEED_PATH#/rss/private/}"
RSS_TOKEN="${RSS_TOKEN%.xml}"

curl --silent --show-error --fail "$BASE_URL$FEED_PATH" > "$WORK_DIR/feed.xml"
expect_contains "$WORK_DIR/feed.xml" "Smoke Episode"
curl --silent --show-error --fail "$BASE_URL/rss/private/$RSS_TOKEN/episodes/$EPISODE_ID/media" > "$WORK_DIR/rss-media.txt"
expect_contains "$WORK_DIR/rss-media.txt" "$MEDIA_TEXT"
expect_status 200 "$BASE_URL/rss/private/$RSS_TOKEN/episodes/$EPISODE_ID/media" -I
log "RSS feed and enclosure checks passed"

curl --silent --show-error --fail \
  -X POST \
  -H "$ORIGIN_HEADER" \
  -H "$CSRF_HEADER" \
  -H "Cookie: $USER_COOKIE" \
  -H "Content-Type: application/json" \
  --data '{}' \
  "$BASE_URL/me/rss-feeds/$FEED_ID/rotate" > "$WORK_DIR/feed-rotate.json"
ROTATED_FEED_URL="$(json_get "$WORK_DIR/feed-rotate.json" "feed_url")"
ROTATED_FEED_PATH="$(python3 - "$ROTATED_FEED_URL" <<'PY'
from urllib.parse import urlparse
import sys
print(urlparse(sys.argv[1]).path)
PY
)"
ROTATED_TOKEN="${ROTATED_FEED_PATH#/rss/private/}"
ROTATED_TOKEN="${ROTATED_TOKEN%.xml}"
expect_status 404 "$BASE_URL$FEED_PATH"
expect_status 404 "$BASE_URL/rss/private/$RSS_TOKEN/episodes/$EPISODE_ID/media"
curl --silent --show-error --fail "$BASE_URL$ROTATED_FEED_PATH" > "$WORK_DIR/feed-rotated.xml"
expect_contains "$WORK_DIR/feed-rotated.xml" "Smoke Episode"
log "RSS token rotation invalidation checks passed"

"${COMPOSE[@]}" exec -T postgres psql -q -v ON_ERROR_STOP=1 -U podcast_hub -d podcast_hub <<SQL
UPDATE program_access_grants
SET status='revoked', revoked_by='$ADMIN_ID', revoked_at=NOW(), updated_at=NOW(), reason='local smoke revoke'
WHERE id='$GRANT_ID';
SQL

curl --silent --show-error --fail -H "Cookie: $USER_COOKIE" "$BASE_URL/programs" > "$WORK_DIR/programs-after-revoke.json"
expect_not_contains "$WORK_DIR/programs-after-revoke.json" "Smoke Program"
expect_status 404 "$BASE_URL/media/episodes/$EPISODE_ID" -H "Cookie: $USER_COOKIE"
curl --silent --show-error --fail "$BASE_URL$ROTATED_FEED_PATH" > "$WORK_DIR/feed-after-revoke.xml"
expect_not_contains "$WORK_DIR/feed-after-revoke.xml" "Smoke Episode"
expect_status 404 "$BASE_URL/rss/private/$ROTATED_TOKEN/episodes/$EPISODE_ID/media"
log "grant revoke invalidation checks passed"

log "running backup and temporary restore verification"
COMPOSE_PROJECT_NAME="$PROJECT_NAME" ENV_FILE="$ENV_FILE" BACKUP_DIR="$BACKUP_DIR" scripts/backup-postgres.sh --env-file="$ENV_FILE" --backup-dir="$BACKUP_DIR" >/dev/null
backup_file="$(find "$BACKUP_DIR" -type f -name 'podcast_hub_*.sql' | head -1)"
if [[ -z "$backup_file" || ! -s "$backup_file" ]]; then
  echo "backup file was not created" >&2
  exit 1
fi
if grep -Fq "$POSTGRES_PASSWORD" "$backup_file"; then
  echo "backup file contains the generated database password" >&2
  exit 1
fi
"${COMPOSE[@]}" exec -T postgres createdb -U podcast_hub podcast_hub_restore
POSTGRES_DB=podcast_hub_restore COMPOSE_PROJECT_NAME="$PROJECT_NAME" ENV_FILE="$ENV_FILE" scripts/restore-postgres.sh --env-file="$ENV_FILE" --backup="$backup_file" --confirm-restore >/dev/null
restore_count="$("${COMPOSE[@]}" exec -T postgres psql -qAt -U podcast_hub -d podcast_hub_restore -c "SELECT COUNT(*) FROM programs WHERE id='$PROGRAM_ID';")"
if [[ "$restore_count" != "1" ]]; then
  echo "restore verification did not find the smoke Program" >&2
  exit 1
fi
"${COMPOSE[@]}" exec -T postgres dropdb -U podcast_hub podcast_hub_restore
log "backup and restore verification passed"

log "local User Beta smoke passed"
