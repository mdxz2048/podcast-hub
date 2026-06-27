# Alpha Deployment

## Scope

This is a local/internal Alpha runbook. It is not a public production deployment guide.

Deployable Alpha supports:

- Go API
- PostgreSQL
- Redis
- Mailpit
- Connector package metadata storage
- Import Job metadata APIs
- Separate trusted-admin Runner startup for fixture execution
- Admin-only staging intake for completed fixture artifacts

Deployable Alpha does not support:

- public HTTPS
- public internet deployment
- RSS
- real duoting
- real external Connector execution
- scheduled Jobs
- interactive or QR Jobs
- review approval, Program/Episode/user subscription publishing
- real media download

## Environment

Copy the placeholder file locally:

```bash
cp .env.alpha.example .env.alpha
```

Replace all placeholder values before running. Do not commit `.env.alpha`.

Generate a local Secret Master Key:

```bash
openssl rand -base64 32
```

Use the generated value as `SECRETS_MASTER_KEY` in the local environment only.

`IMPORT_ARTIFACT_STORE_DIR` must be a private directory shared by the API and Runner when using M1.3A intake. The default Alpha compose files bind this to `.local/import-artifacts`; keep it out of Git and do not serve it as static files.

## Start Dependencies and API

```bash
docker compose -f docker-compose.alpha.yml --env-file .env.alpha up -d postgres redis mailpit api
```

The API container applies migrations on startup through `cmd/api`.

Health check:

```bash
curl http://127.0.0.1:8080/healthz
```

Readiness check:

```bash
curl http://127.0.0.1:8080/readyz
```

Mailpit UI:

```text
http://127.0.0.1:8025
```

## Admin Bootstrap

Run locally with the same `.env.alpha` values:

```bash
set -a && source .env.alpha && set +a
go run ./cmd/admin seed --email admin@example.invalid
```

## Runner Startup

Default Alpha API mode is:

```text
RUNNER_MODE=disabled
```

Queued Import Jobs remain queued until a separate Runner is explicitly started. Runner disabled does not make API readiness fail.

For trusted-admin fixture execution only:

```bash
set -a && source .env.alpha && set +a
export RUNNER_MODE=docker_trusted_admin
export DATABASE_URL='postgresql://podcast_hub:change_me_alpha_password@127.0.0.1:5432/podcast_hub?sslmode=disable'
export RUNNER_PYTHON_BASIC_IMAGE='python:3.12-alpine'
go run ./cmd/runner
```

Do not use this for real duoting or untrusted third-party Connectors. M1.2D does not provide a strong sandbox for arbitrary Connector code.

When using the compose-based Runner, `deploy/docker-compose.runner-alpha.yml` mounts `../.local/import-artifacts` into the Runner as `/artifacts`. Use the same host path for the API so intake can read only promoted, registered artifacts after the Runner workspace has been cleaned.

## Logs

API and Runner logs must not include:

- passwords
- cookies
- tokens
- authorization headers
- session files
- Secret payloads
- real media paths

Runner Event persistence performs redaction, but operators must still avoid passing secrets through Connector stdout.

## Backup and Restore

Minimal local backup:

```bash
docker compose -f docker-compose.alpha.yml exec postgres pg_dump -U podcast_hub podcast_hub > podcast_hub_alpha.sql
```

Minimal local restore:

```bash
cat podcast_hub_alpha.sql | docker compose -f docker-compose.alpha.yml exec -T postgres psql -U podcast_hub podcast_hub
```

Connector package metadata, private ImportJob artifacts, and future staging storage must be backed up separately from PostgreSQL. Do not store real secrets or media in Git.
