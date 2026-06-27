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

if command -v docker >/dev/null 2>&1; then
  check "compose user beta config" "$(docker compose -f deploy/compose.user-beta.yml config >/dev/null && echo 1 || echo 0)"
  check "runner compose config" "$(docker compose -f deploy/runner-compose.user-beta.yml config >/dev/null && echo 1 || echo 0)"
else
  echo "WARN docker not available; compose static checks skipped"
fi

if [[ "$failures" -gt 0 ]]; then
  echo "Preflight failed with $failures issue(s)."
  exit 1
fi

echo "Preflight passed."
