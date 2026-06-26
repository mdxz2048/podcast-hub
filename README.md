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
- `GET /admin/connectors` (admin only)
- `POST /admin/connectors/upload` (admin only, multipart ZIP upload + static validation)
- `GET /admin/connectors/{connectorId}` (admin only)
- `GET /admin/connectors/{connectorId}/versions` (admin only)
- `GET /admin/connector-versions/{versionId}` (admin only)
- `POST /admin/connector-versions/{versionId}/approve` (admin only)
- `POST /admin/connector-versions/{versionId}/reject` (admin only)
- `POST /admin/connector-versions/{versionId}/disable` (admin only)
- `POST /admin/connectors/{connectorId}/disable` (admin only)
- `POST /admin/connectors/{connectorId}/enable` (admin only)

## M1.1A scope note

- Connector is an uploaded versioned package, not a Program and not a Source.
- This phase performs only upload, static validation, registry review, enable/disable states.
- This phase does **not** execute connector code, does not create Source/ImportJob, and does not download media.
- Do not upload secrets/cookies/session files/media; manifest secrets are declaration-only.
- Duoting remains an external future package upload target, not repository-internal platform source code.

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
