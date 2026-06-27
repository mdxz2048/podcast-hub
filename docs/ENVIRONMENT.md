# Environment

## 1. Scope

This document defines M1.0 local environment and security configuration for the real account flow.

## 2. Git policy

Allowed in Git:

- `.env.example`

Forbidden in Git:

- `.env` and other real env files
- real passwords, tokens, cookies, keys
- private account data and media

## 3. Required variables

### Frontend

- `VITE_API_BASE_URL`
- `VITE_TURNSTILE_MODE` (`mock` or `cloudflare`, optional `off` for local debug)
- `VITE_TURNSTILE_SITE_KEY`

### Backend

- `APP_ENV`
- `HTTP_ADDR`
- `FRONTEND_ORIGIN`
- `DATABASE_URL`
- `REDIS_URL`
- `SESSION_PEPPER`
- `AUTH_CODE_PEPPER`
- `TURNSTILE_MODE`
- `TURNSTILE_SECRET_KEY` (required in `cloudflare` mode and production)
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM`
- `ADMIN_SEED_PASSWORD` (optional, local admin seed CLI only)

## 4. Production fail-closed rules

Production startup must fail when:

- `TURNSTILE_MODE != cloudflare`
- `TURNSTILE_SECRET_KEY` missing
- `REDIS_URL` missing
- `SESSION_COOKIE_SECURE != true`
- any required security pepper missing

## 5. Local development stack

M1.0 local dependencies:

- PostgreSQL
- Redis
- Mailpit

Start dependencies:

```bash
docker compose up -d postgres redis mailpit
```

Mailpit UI:

- http://127.0.0.1:8025

## 6. Run commands

```bash
set -a && source .env && set +a
go run ./cmd/api
```

```bash
corepack pnpm install
corepack pnpm dev
```

## 7. Local admin bootstrap

Use the shared login flow with an admin account seeded locally (no HTTP admin registration endpoint):

```bash
export ADMIN_SEED_PASSWORD='replace-with-strong-password'
go run ./cmd/admin seed --email admin@example.invalid
```

Development-only promotion of an existing user:

```bash
go run ./cmd/admin seed --email user@example.invalid --promote
```
## M1.1A environment additions

- `CONNECTOR_PACKAGE_LOCAL_DIR` (default: `.local/connector-packages`)
  - Local development package store root for Connector ZIP quarantine/approved files.
  - Must remain git-ignored.

## M1.1B environment additions

- `SECRETS_MASTER_KEY`
  - Required in production.
  - Must be 32 raw bytes or base64-encoded 32 bytes.
  - Used for Secret Reference encryption; never commit a real value.

Security reminder:

- Never place secrets, session dumps, cookies, tokens, or real media into connector package directories.

## M1.2D Alpha environment additions

- `RUNNER_MODE`
  - Default: `disabled`.
  - Allowed values: `disabled`, `docker_trusted_admin`.
  - API service may expose this as non-sensitive status metadata but does not execute Connector code.
- `RUNNER_PYTHON_BASIC_IMAGE`
  - Docker image used by the separate trusted-admin Runner for fixture execution.
- `RUNNER_PYTHON_TELEGRAM_IMAGE`
  - Reserved placeholder for a future runtime profile. M1.2D does not use Telegram.
- `RUNNER_WORKSPACE_ROOT`
  - Local Runner workspace root. Must remain git-ignored.
- `.env.alpha.example`
  - Placeholder-only local/internal Alpha environment file.
  - Copy to `.env.alpha` and replace every placeholder locally.
  - Do not commit `.env.alpha`.
- `POSTGRES_PASSWORD`
  - Placeholder for local/internal Alpha PostgreSQL.
  - Replace locally; do not use the example value outside disposable development.
- `STAGING_STORE_DIR` and `MEDIA_STORE_DIR`
  - Private Alpha volume mount points.
  - Not served by the frontend or API as static public files.

Generate a local Secret Master Key with one of:

```bash
openssl rand -base64 32
```

or a raw 32-byte value managed by the deployment secret store. Never commit the generated value.
