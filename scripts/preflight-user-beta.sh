#!/usr/bin/env bash
set -euo pipefail

APPLY=0
ENV_FILE="${ENV_FILE:-.env.user-beta}"

for arg in "$@"; do
  case "$arg" in
    --dry-run) APPLY=0 ;;
    --apply) APPLY=1 ;;
    --env-file=*) ENV_FILE="${arg#--env-file=}" ;;
    *) echo "unknown argument: $arg" >&2; exit 2 ;;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

failures=0
warnings=0

check() {
  local name="$1"
  local ok="$2"
  if [[ "$ok" == "1" ]]; then
    printf 'PASS %s\n' "$name"
  else
    printf 'FAIL %s\n' "$name"
    failures=$((failures + 1))
  fi
}

warn() {
  printf 'WARN %s\n' "$1"
  warnings=$((warnings + 1))
}

not_placeholder() {
  local value="${1:-}"
  [[ -n "$value" && "$value" != *replace_with_* && "$value" != *change_me* ]]
}

if [[ "$APPLY" == "1" ]]; then
  mode="apply"
else
  mode="dry-run"
fi

echo "Podcast Hub User Beta preflight (mode=$mode). Sensitive values are not printed."
[[ "$APPLY" == "1" ]] || echo "Dry run only. Re-run with --apply to mark this as an operator preflight."

check "APP_ENV production" "$([[ "${APP_ENV:-}" == "production" ]] && echo 1 || echo 0)"
check "DATABASE_URL configured" "$(not_placeholder "${DATABASE_URL:-}" && echo 1 || echo 0)"
check "REDIS_URL configured" "$(not_placeholder "${REDIS_URL:-}" && echo 1 || echo 0)"
check "SMTP host is not Mailpit" "$([[ -n "${SMTP_HOST:-}" && "${SMTP_HOST:-}" != "mailpit" && "${SMTP_HOST:-}" != "127.0.0.1" && "${SMTP_PORT:-}" != "1025" ]] && echo 1 || echo 0)"
check "SESSION_COOKIE_SECURE true" "$([[ "${SESSION_COOKIE_SECURE:-}" == "true" ]] && echo 1 || echo 0)"
check "SECRETS_MASTER_KEY configured" "$(not_placeholder "${SECRETS_MASTER_KEY:-}" && echo 1 || echo 0)"
check "Turnstile cloudflare mode" "$([[ "${TURNSTILE_MODE:-}" == "cloudflare" && "${VITE_TURNSTILE_MODE:-}" == "cloudflare" ]] && echo 1 || echo 0)"
check "Turnstile secret configured" "$(not_placeholder "${TURNSTILE_SECRET_KEY:-}" && echo 1 || echo 0)"
check "Frontend origin is not wildcard" "$([[ -n "${FRONTEND_ORIGIN:-}" && "${FRONTEND_ORIGIN:-}" != "*" ]] && echo 1 || echo 0)"
check "Runner mode explicitly disabled or trusted" "$([[ "${RUNNER_MODE:-}" == "disabled" || "${RUNNER_MODE:-}" == "docker_trusted_admin" ]] && echo 1 || echo 0)"
check "Runner trusted admin explicitly configured" "$([[ -n "${RUNNER_TRUSTED_ADMIN_ENABLED:-}" ]] && echo 1 || echo 0)"
check "API private stores isolated" "$([[ "${STAGING_STORE_DIR:-}" != "${MEDIA_STORE_DIR:-}" && "${CONNECTOR_PACKAGE_LOCAL_DIR:-}" != "${MEDIA_STORE_DIR:-}" ]] && echo 1 || echo 0)"
check "No default admin password" "$([[ -z "${DEFAULT_ADMIN_PASSWORD:-}" ]] && echo 1 || echo 0)"
check "Backup directory configured" "$([[ -n "${BACKUP_DIR:-}" ]] && echo 1 || echo 0)"
if [[ -d "${BACKUP_DIR:-}" ]]; then
  if [[ -w "${BACKUP_DIR:-}" && ! -L "${BACKUP_DIR:-}" ]]; then
    check "Backup directory is writable and not symlink" "1"
  else
    check "Backup directory is writable and not symlink" "0"
  fi
fi
check "RSS token redaction configured" "$(grep -Fq '/rss/private/[redacted]' deploy/Caddyfile.user-beta.template && echo 1 || echo 0)"
check ".env.user-beta remains git-ignored" "$(git check-ignore -q .env.user-beta && echo 1 || echo 0)"

if command -v docker >/dev/null 2>&1; then
  compose_tmp="$(mktemp "${TMPDIR:-/tmp}/podcast-hub-compose.XXXXXX")"
  runner_tmp="$(mktemp "${TMPDIR:-/tmp}/podcast-hub-runner-compose.XXXXXX")"
  if docker compose -f deploy/compose.user-beta.yml config >"$compose_tmp"; then
    check "compose user beta config" "1"
    check "api-volume-init rootfs writable for one-shot init" "$(! awk '/^  api-volume-init:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && /read_only: true/{found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "api-volume-init network disabled" "$(awk '/^  api-volume-init:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && /network_mode: none/{found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "api-volume-init exposes no ports" "$(! awk '/^  api-volume-init:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && /^[[:space:]]+ports:/{found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "api-volume-init excludes Docker socket" "$(! awk '/^  api-volume-init:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && index($0, "/var/run/docker.sock"){found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "api-volume-init uses controlled volume mounts" "$(awk '/^  api-volume-init:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && index($0, "target: /data/connector-packages"){targets["connector-packages"]=1} in_service && index($0, "target: /data/import-artifacts"){targets["import-artifacts"]=1} in_service && index($0, "target: /data/staging"){targets["staging"]=1} in_service && index($0, "target: /data/private-media"){targets["private-media"]=1} in_service && index($0, "target: /data/tmp"){targets["tmp"]=1} in_service && /type: bind/{bind=1} END{exit (length(targets)==5 && !bind) ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "API container non-root" "$(grep -q 'user: 10001:10001' "$compose_tmp" && echo 1 || echo 0)"
    check "API container remains read-only" "$(awk '/^  api:/{in_service=1; next} /^  [a-zA-Z0-9_-]+:/{in_service=0} in_service && /read_only: true/{found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
    check "API compose excludes Docker socket" "$(! grep -q '/var/run/docker.sock' "$compose_tmp" && echo 1 || echo 0)"
    check "Private media not mounted into frontend" "$(! awk '/^  frontend:/{in_frontend=1; next} /^  [a-zA-Z0-9_-]+:/{in_frontend=0} in_frontend && /private-media|staging|connector-packages|import-artifacts/{found=1} END{exit found ? 0 : 1}' "$compose_tmp" && echo 1 || echo 0)"
  else
    check "compose user beta config" "0"
  fi
  if docker compose -f deploy/runner-compose.user-beta.yml config >"$runner_tmp"; then
    check "runner compose config" "1"
    check "Runner compose owns Docker socket boundary" "$(grep -q '/var/run/docker.sock' "$runner_tmp" && echo 1 || echo 0)"
  else
    check "runner compose config" "0"
  fi
  rm -f "$compose_tmp" "$runner_tmp"
else
  warn "docker not available; compose static checks skipped"
fi

if command -v psql >/dev/null 2>&1 && not_placeholder "${DATABASE_URL:-}"; then
  if psql "$DATABASE_URL" -qAt -c "SELECT COUNT(*) FROM schema_migrations" >/dev/null 2>&1; then
    check "Migration table reachable" "1"
  else
    warn "migration table not reachable from local psql; verify on target host"
  fi
else
  warn "psql unavailable or DATABASE_URL placeholder; migration status check skipped"
fi

if [[ "$failures" -gt 0 ]]; then
  echo "Preflight failed with $failures issue(s)."
  exit 1
fi

echo "Preflight passed with $warnings warning(s)."
