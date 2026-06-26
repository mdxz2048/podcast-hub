# M1.1A Connector Registry Acceptance

## Implemented

- Connector registry schema:
  - `connectors`
  - `connector_versions`
  - `connector_events`
- Admin-only APIs:
  - `GET /admin/connectors`
  - `POST /admin/connectors/upload`
  - `GET /admin/connectors/{connectorId}`
  - `GET /admin/connectors/{connectorId}/versions`
  - `GET /admin/connector-versions/{versionId}`
  - `POST /admin/connector-versions/{versionId}/approve`
  - `POST /admin/connector-versions/{versionId}/reject`
  - `POST /admin/connector-versions/{versionId}/disable`
  - `POST /admin/connectors/{connectorId}/disable`
  - `POST /admin/connectors/{connectorId}/enable`
- Upload handling:
  - multipart/form-data only
  - quarantine storage
  - server-generated storage keys independent of uploaded filenames
  - SHA-256 + size capture
  - ZIP static security checks
  - strict manifest validation with unknown-field rejection
  - invalid non-ZIP uploads are rejected and cleaned up
  - parseable but policy-invalid ZIPs are recorded as rejected versions with redacted validation issues
- Frontend admin pages switched from mock to real API for connector registry pages.

## Security and policy boundaries

- Connector package is not executed in M1.1A.
- No Python process invocation.
- No shell execution.
- No Docker build/run.
- No Program/Source/ImportJob/Episode creation from connector upload.
- No secret value persistence from manifest; declaration-only secret metadata.
- No absolute local filesystem path returned to client.
- No package storage key or ZIP content returned to client.

## Explicit non-goals (Deferred)

- Production object storage package store
- Connector execution runner/sandbox
- Source binding workflow
- Import job orchestration
- Episode staging/review publishing path
- RSS publishing integration
- Runtime QR flow integration

## Domain clarification

- Program != Connector
- Source != Connector
- Duoting is treated as future external uploaded package material, not main repository first-party source code.
