# M1.2D Alpha Deployment Acceptance

## Implemented

- Admin Import Job console backed by real APIs.
- Source detail manual Import Job creation for runnable active Sources.
- Import Job list shows status, Source, ConnectorVersion, creation time, start time, and finish time.
- Import Job detail shows metadata, redacted Events, Artifact metadata, cancellation, and Runner disabled reason.
- Artifact UI never displays file contents, absolute paths, storage keys, RSS, media download, or publication actions.
- Public `/healthz` endpoint returns redacted service/dependency health.
- `GET /admin/system/status` returns non-sensitive Runner status.
- `.env.alpha.example` with placeholder-only values.
- `docker-compose.alpha.yml` for local/internal PostgreSQL, Redis, Mailpit, and API.
- `Dockerfile.alpha` builds API and Runner binaries.
- `docs/ALPHA_DEPLOYMENT.md` documents local/internal startup, migrations, Runner startup, Secret Master Key generation, backup/restore, and logs.

## Tests

- Job list empty state.
- Unauthenticated admin access redirects to login.
- Normal user sees permission denied.
- Manual Job creation from Source detail.
- Runner disabled reason.
- queued Job cancel.
- Job detail redacted Event display.
- Artifact metadata-only display.
- No external content run, publish, RSS, or media download controls.
- Frontend tests use mock APIs and do not depend on a real Go API or manual Vite server.

## Boundaries

- No public deployment.
- No HTTPS termination.
- No real duoting.
- No real external Connector execution.
- No scheduled, interactive, or QR Job.
- No Program, Episode, RSS, user subscription, or media download capability.
- API service does not require Docker socket access.
- Connector containers must never receive Docker socket access.
