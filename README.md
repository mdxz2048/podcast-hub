# Podcast Hub

Podcast Hub is a content ingestion and publishing platform.  
Current phase includes real account/admin flows, Connector registry, Source/Secret metadata, Import Job lifecycle, fixture-only Runner protocol, trusted-admin Docker Runner boundaries, local/internal Alpha deployment preparation, admin-only staging intake, administrator review/publish state controls, private media/RSS backend flows, and privacy-safe request correlation IDs.

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
- `GET /admin/sources` (admin only)
- `POST /admin/sources` (admin only)
- `GET /admin/sources/{sourceId}` (admin only)
- `PATCH /admin/sources/{sourceId}` (admin only)
- `POST /admin/sources/{sourceId}/enable` (admin only)
- `POST /admin/sources/{sourceId}/disable` (admin only)
- `GET /admin/secrets` (admin only)
- `POST /admin/secrets/text` (admin only)
- `POST /admin/secrets/file` (admin only)
- `POST /admin/secrets/{secretId}/revoke` (admin only)
- `POST /admin/sources/{sourceId}/secret-bindings` (admin only)
- `DELETE /admin/sources/{sourceId}/secret-bindings/{bindingId}` (admin only)
- `GET /admin/import-jobs` (admin only)
- `POST /admin/sources/{sourceId}/import-jobs` (admin only)
- `GET /admin/import-jobs/{jobId}` (admin only)
- `GET /admin/import-jobs/{jobId}/events` (admin only)
- `GET /admin/import-jobs/{jobId}/artifacts` (admin only)
- `POST /admin/import-jobs/{jobId}/cancel` (admin only)
- `POST /admin/import-jobs/{jobId}/intake` (admin only)
- `GET /admin/import-jobs/{jobId}/intake-status` (admin only)
- `GET /admin/staging/programs` (admin only)
- `GET /admin/staging/programs/{programId}` (admin only)
- `GET /admin/staging/episodes` (admin only)
- `GET /admin/staging/episodes/{episodeId}` (admin only)
- `GET /admin/review` (admin only)
- `GET /admin/review/{reviewId}` (admin only)
- `POST /admin/review/{reviewId}/approve` (admin only)
- `POST /admin/review/{reviewId}/reject` (admin only)
- `GET /admin/programs` (admin only)
- `GET /admin/programs/{programId}` (admin only)
- `PATCH /admin/programs/{programId}` (admin only)
- `POST /admin/programs/{programId}/submit-review` (admin only)
- `POST /admin/programs/{programId}/publish` (admin only)
- `POST /admin/programs/{programId}/archive` (admin only)
- `GET /admin/episodes/{episodeId}` (admin only)
- `PATCH /admin/episodes/{episodeId}` (admin only)
- `POST /admin/episodes/{episodeId}/submit-review` (admin only)
- `POST /admin/episodes/{episodeId}/publish` (admin only)
- `POST /admin/episodes/{episodeId}/archive` (admin only)

## M1.1A scope note

- Connector is an uploaded versioned package, not a Program and not a Source.
- This phase performs only upload, static validation, registry review, enable/disable states.
- This phase does **not** execute connector code, does not create Source/ImportJob, and does not download media.
- Do not upload secrets/cookies/session files/media; manifest secrets are declaration-only.
- Duoting remains an external future package upload target, not repository-internal platform source code.

## M1.1B scope note

- Source is a Connector configuration instance, not a Program.
- Secrets are stored as encrypted records and bound by reference; APIs never return values.
- Alpha Source creation supports only manual + none/reusable_session + unattended.
- This phase still does **not** execute Connector code and does not create Program, Episode, RSS, or ImportJob data.

## Admin bootstrap (local only)

```bash
export ADMIN_SEED_PASSWORD='replace-with-strong-password'
go run ./cmd/admin seed --email admin@example.invalid
```

Admin and normal users share `/login`; admin role is enforced by backend middleware (`RequireAdmin`).

## Verification commands

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```

Error responses include a server-generated `request_id` for log correlation. The ID is an opaque random value and must not contain hostnames, local paths, caller-provided IDs, tokens, cookies, or other environment-derived information.

## Deployable Alpha notes

Local/internal Alpha files:

- `.env.alpha.example`
- `docker-compose.alpha.yml`
- `Dockerfile.alpha`
- `docs/ALPHA_DEPLOYMENT.md`
- `docs/OPERATIONS_ADMIN_ALPHA.md`

The API service starts with `RUNNER_MODE=disabled` by default and does not mount Docker socket access. Start the Runner separately only for trusted-admin fixture execution. Alpha still does not support public deployment, RSS, real duoting, scheduled jobs, interactive/QR jobs, user subscriptions, or real media download. M1.3B can mark content published in admin state, but normal users cannot see it until M1.3C access grants and catalog APIs are enabled.

Run the optional real Docker fixture smoke test explicitly:

```bash
RUNNER_INTEGRATION_TEST=1 CGO_ENABLED=0 GOTOOLCHAIN=local go test ./... -run Integration
```
