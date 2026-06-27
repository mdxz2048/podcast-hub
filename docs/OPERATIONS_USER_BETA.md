# Operations User Beta

This runbook prepares a private User Beta deployment candidate. It does not claim that a public deployment, backup restore drill, or real external Connector pilot has been performed.

## Scope

User Beta candidate includes:

- API, frontend, PostgreSQL, Redis, and Caddy reverse proxy.
- Isolated volumes for Connector packages, Import artifacts, staging files, private media, and PostgreSQL data.
- Separate Runner compose file.
- Default `RUNNER_MODE=disabled`.
- Preflight, backup, restore, and secret-rotation readiness scripts.

User Beta candidate does not include:

- real duoting access
- real Telegram
- real public deployment
- real Connector ZIP upload or execution
- real media download
- scheduled Jobs
- interactive or QR Jobs

## Deployment Files

- `deploy/compose.user-beta.yml`
- `deploy/runner-compose.user-beta.yml`
- `deploy/Caddyfile.user-beta.template`
- `.env.user-beta.example`
- `scripts/preflight-user-beta.sh`
- `scripts/backup-postgres.sh`
- `scripts/restore-postgres.sh`
- `scripts/rotate-secrets-check.sh`

## First-Time Preparation

```bash
cp .env.user-beta.example .env.user-beta
```

Replace every placeholder locally. Do not commit `.env.user-beta`.

Generate independent secrets for:

- `POSTGRES_PASSWORD`
- `SESSION_PEPPER`
- `AUTH_CODE_PEPPER`
- `SECRETS_MASTER_KEY`
- `TURNSTILE_SECRET_KEY`
- `SMTP_PASSWORD`

Do not reuse development values.

## Preflight

Dry run:

```bash
scripts/preflight-user-beta.sh --env-file=.env.user-beta
```

Operator-marked run:

```bash
scripts/preflight-user-beta.sh --env-file=.env.user-beta --apply
```

Preflight is fail-closed for production shape checks, including database, Redis, SMTP, secure cookies, Cloudflare Turnstile mode, Secret master key, broad CORS, Mailpit-like SMTP, default admin password, private storage isolation, Runner mode, and explicit trusted-admin Runner setting.

## Start API Stack

```bash
docker compose -f deploy/compose.user-beta.yml --env-file .env.user-beta up -d postgres redis api frontend caddy
```

The API applies migrations on startup. PostgreSQL and Redis are not published to public host ports by this compose file.
The compose file creates the stable private Docker network `podcast_hub_user_beta_private` and stable named volumes used by the separate Runner compose file.

## Runner

Runner is separate and defaults to disabled:

```bash
docker compose -f deploy/runner-compose.user-beta.yml --env-file .env.user-beta up runner
```

Only the Runner compose may mount `/var/run/docker.sock`. Connector containers must not receive the Docker socket. Do not enable `docker_trusted_admin` for untrusted Connector packages.
Start the API stack first so the external `podcast_hub_user_beta_private` network and `podcast_hub_user_beta_import_artifacts` volume exist for the Runner compose file.

## Backup

```bash
scripts/backup-postgres.sh --env-file=.env.user-beta --backup-dir=./backups/user-beta
```

Backup output is chmod `0600`. The script does not print database passwords. Private package, staging, artifact, secret, and media volumes need separate storage-level backup according to the operator's private infrastructure.

## Restore

Restore requires explicit confirmation:

```bash
scripts/restore-postgres.sh --env-file=.env.user-beta --backup=./backups/user-beta/podcast_hub_YYYYMMDDTHHMMSSZ.sql --confirm-restore
```

Do not run restore against an existing private database until the operator has confirmed the target.

## Logging

Reverse proxy logs must redact `/rss/private/{token}` paths. The Caddy template includes a filter for RSS token path segments. API audit metadata records token prefixes or redacted token values only.

Logs must not contain:

- cookies
- Authorization headers
- RSS tokens
- session values
- passwords
- Secret values
- private storage keys
- absolute filesystem paths

## Health And Readiness

- `/healthz` is liveness.
- `/readyz` is dependency readiness.
- Runner disabled must not fail API readiness.

Restrict operational endpoints at the reverse proxy and infrastructure layer.
