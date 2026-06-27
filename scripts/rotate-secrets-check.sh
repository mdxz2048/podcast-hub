#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-.env.user-beta}"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

failures=0
check_present() {
  local name="$1"
  local value="${2:-}"
  if [[ -n "$value" && "$value" != *replace_with_* && "$value" != *change_me* ]]; then
    printf 'PASS %s configured\n' "$name"
  else
    printf 'FAIL %s missing or placeholder\n' "$name"
    failures=$((failures + 1))
  fi
}

echo "Checking secret rotation readiness. Sensitive values are not printed."
check_present "SESSION_PEPPER" "${SESSION_PEPPER:-}"
check_present "AUTH_CODE_PEPPER" "${AUTH_CODE_PEPPER:-}"
check_present "SECRETS_MASTER_KEY" "${SECRETS_MASTER_KEY:-}"
check_present "TURNSTILE_SECRET_KEY" "${TURNSTILE_SECRET_KEY:-}"
check_present "SMTP_PASSWORD" "${SMTP_PASSWORD:-}"

if [[ "$failures" -gt 0 ]]; then
  exit 1
fi

echo "Rotation readiness check passed."
