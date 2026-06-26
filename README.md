# Podcast Hub

Podcast Hub is a content ingestion and publishing platform.  
Current phase implements M1.0 account/security infrastructure only.
M1.0C adds admin bootstrap and backend/admin permission chain.

## Local development

1. Copy placeholders:

```bash
cp .env.example .env
```

2. Start local dependencies:

```bash
docker compose up -d postgres redis mailpit
```

3. Start Go API:

```bash
set -a && source .env && set +a
go run ./cmd/api
```

4. Start frontend:

```bash
corepack pnpm install
corepack pnpm dev
```

## M1.0 real APIs

- `POST /auth/register/request-code`
- `POST /auth/register/verify-code`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/password-reset/request`
- `POST /auth/password-reset/verify`
- `GET /auth/me`
- `GET /admin/me` (admin only)
- `GET /admin/system/status` (admin only)

## Admin bootstrap (local only)

```bash
export ADMIN_SEED_PASSWORD='replace-with-strong-password'
go run ./cmd/admin seed --email admin@example.invalid
```

Admin and normal users share `/login`; admin role is enforced by backend middleware (`RequireAdmin`).

## Verification commands

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```
