# Podcast Hub

Podcast Hub is a content ingestion and publishing platform.  
Current phase implements M1.0 account/security infrastructure only.

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

## Verification commands

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm build
corepack pnpm screenshots
git diff --check
```
