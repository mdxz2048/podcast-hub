# M1.2D Admin Alpha Acceptance

## Implemented

- `GET /healthz` public liveness probe with no dependency details.
- `GET /readyz` public readiness probe with safe `ready` / `not_ready`.
- Admin-only system status with safe dependency summary and non-sensitive Runner mode.
- Import Job admin list and detail pages backed by real APIs.
- Source detail manual Job creation.
- Runner disabled reason visible in admin UI.
- queued and running Job cancellation.
- Redacted Event display.
- Artifact metadata display only.
- `.env.alpha.example`.
- `Dockerfile.alpha` for non-root API image.
- `docker-compose.alpha.yml` for local/internal API, PostgreSQL, Redis, and Mailpit.
- `docs/OPERATIONS_ADMIN_ALPHA.md` private admin Alpha runbook.

## Security Checks

- `/healthz` does not return database, Redis, SMTP, Runner, path, version, or Secret details.
- `/readyz` does not return connection strings, internal errors, paths, or Secrets.
- API Compose service does not mount Docker socket.
- API Compose service does not use host networking.
- API image runs as non-root user.
- API image does not include Docker CLI or Runner binary.
- Package store, staging, and media paths are private volumes, not static frontend assets.
- Runner must be deployed separately.

## Explicit Non-Goals

- No duoting.
- No Telegram.
- No scheduled Jobs.
- No interactive or QR Jobs.
- No Program/Episode publishing.
- No RSS.
- No user subscriptions.
- No public deployment.
- No untrusted third-party sandbox.

## Required Verification

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
docker compose -f docker-compose.alpha.yml config
docker build -f Dockerfile.alpha -t podcast-hub-api:alpha-test .
```

If Docker is unavailable, Docker build must be reported as an environment blocker rather than marked as passed.
