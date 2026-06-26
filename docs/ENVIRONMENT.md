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
