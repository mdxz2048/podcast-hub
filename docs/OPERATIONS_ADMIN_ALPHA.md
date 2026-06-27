# Operations Admin Alpha

## Scope

This runbook is for private administrator Alpha only. It is not a public production deployment guide.

## R0 Capability

R0 can:

- Admin login.
- Connector upload, static validation, review, enable, and disable.
- Source creation.
- Secret write and Secret Reference binding.
- Import Job creation and cancellation.
- Import Job metadata, redacted Events, and Artifact metadata viewing.

R0 cannot run Connector code.

## R1 Capability

R1 requires a separately started Runner service.

R1 can:

- Run the test fixture.
- Run an administrator-explicitly trusted Connector in `docker_trusted_admin` mode.

R1 is not suitable for untrusted third-party Connector execution.

## Not Supported

- duoting.
- Telegram.
- scheduled Jobs.
- interactive or QR Jobs.
- Program or Episode publication.
- RSS.
- user subscriptions.
- public or multi-tenant deployment.
- untrusted arbitrary code execution.

## Startup

Copy placeholders:

```bash
cp .env.alpha.example .env.alpha
```

Replace every placeholder locally. Do not commit `.env.alpha`.

Start local/internal Alpha services:

```bash
docker compose -f docker-compose.alpha.yml --env-file .env.alpha up -d postgres redis mailpit api
```

The API applies migrations at startup.

Health:

```bash
curl http://127.0.0.1:8080/healthz
```

Readiness:

```bash
curl http://127.0.0.1:8080/readyz
```

## Initial Admin

```bash
set -a && source .env.alpha && set +a
go run ./cmd/admin seed --email admin@example.invalid
```

## API And Runner Separation

The API service must not mount the Docker socket and must not execute Connector code. The API owns control-plane metadata, authentication, Source configuration, Job creation, and admin views.

The Runner is a separate trusted-admin service. Only the Runner may have Docker execution capability. Connector containers still must not receive the Docker socket.

Runner compose:

```bash
RUNNER_MODE=docker_trusted_admin docker compose -f deploy/docker-compose.runner-alpha.yml up runner
```

The Runner service exposes no public ports. It is trusted-admin Alpha only.

## Secret Boundary

Secrets are written through admin APIs as encrypted records and bound by reference. Admin APIs must not return plaintext Secret values. M1.2E allows only the separate Runner to decrypt required Source-bound Secrets after it claims a Job. Secret values are written only to temporary `/work/secrets` files and removed with workspace cleanup.

## Logs

Logs must not contain passwords, cookies, tokens, authorization headers, Secret payloads, session files, or real media paths.

## Backup

```bash
docker compose -f docker-compose.alpha.yml exec postgres pg_dump -U podcast_hub podcast_hub > podcast_hub_alpha.sql
```

Private package, staging, and media volumes must be backed up separately. Do not store real Secret values or media in Git.

## Restore

```bash
cat podcast_hub_alpha.sql | docker compose -f docker-compose.alpha.yml exec -T postgres psql -U podcast_hub podcast_hub
```

## Rollback

Stop services:

```bash
docker compose -f docker-compose.alpha.yml down
```

Restore a known-good database dump and restart the previous image. Do not use `git reset`, force push, or history rewrites as a deployment rollback mechanism.
