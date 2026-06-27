#!/usr/bin/env bash
set -euo pipefail

APPLY=0
ENV_FILE="${ENV_FILE:-.env.user-beta}"
RUNNER_WORKSPACE_ROOT="${RUNNER_WORKSPACE_ROOT:-.local/runner-workspaces}"
OLDER_THAN_DAYS="${OLDER_THAN_DAYS:-7}"

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

root="$(cd "$RUNNER_WORKSPACE_ROOT" 2>/dev/null && pwd -P || true)"
if [[ -z "$root" ]]; then
  echo "Runner workspace root does not exist. Nothing to clean."
  exit 0
fi
case "$root" in
  /|/Users|/Users/*/Code|/tmp) echo "Refusing unsafe workspace root." >&2; exit 2 ;;
esac

echo "Job workspace cleanup (${APPLY:-0} = apply). Absolute paths are not printed."
[[ "$APPLY" == "1" ]] || echo "Dry run only. Re-run with --apply to delete eligible old workspaces."

count=0
while IFS= read -r -d '' path; do
  count=$((count + 1))
  printf 'Eligible workspace: %s\n' "$(basename "$path")"
  if [[ "$APPLY" == "1" ]]; then
    rm -rf -- "$path"
  fi
done < <(find "$root" -mindepth 1 -maxdepth 1 -type d -mtime +"$OLDER_THAN_DAYS" -print0)

echo "Eligible workspaces: $count"
[[ "$APPLY" == "1" ]] && echo "Job workspace cleanup applied."
