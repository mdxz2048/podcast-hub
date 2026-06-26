# M1.1B Source And Secret Reference Acceptance

## Implemented

- Admin-only Connector Source APIs for list, create, detail, update, enable, and disable.
- Admin-only Secret metadata/write APIs for encrypted text/file Secret creation and revoke.
- Source Secret binding APIs.
- `connector_sources`, `secret_records`, `source_secret_bindings`, and `source_events` schema.
- AES-GCM Secret encryption with `SECRETS_MASTER_KEY`.
- Production startup fails when `SECRETS_MASTER_KEY` is missing.
- Frontend admin pages: `/admin/sources`, `/admin/sources/new`, `/admin/sources/:sourceId`, and `/admin/secrets`.

## Security Boundaries

- Source stores Secret references only, never Secret values.
- Secret APIs never return plaintext, encrypted payloads, internal paths, file names, cookies, sessions, or tokens.
- Source can only use an approved ConnectorVersion under an active Connector.
- Alpha Source creation supports only manual + none/reusable_session + unattended.
- `qr_each_run`, scheduled, and interactive Source creation are deferred to M3 interactive / QR Connector work.
- No Connector execution occurs in M1.1B.

## Deferred

- Program and Episode creation.
- ImportJob runner.
- Staging review.
- Private RSS.
- Scheduled jobs.
- Interactive / QR jobs.
- Real duoting.
- Untrusted third-party Connector isolation.
- Production deployment.
