# M1.2A Import Job Lifecycle Acceptance

## Implemented

- Import Job schema:
  - `import_jobs`
  - `import_job_events`
  - `import_job_artifacts`
- Admin-only Import Job APIs:
  - `GET /admin/import-jobs`
  - `POST /admin/sources/{sourceId}/import-jobs`
  - `GET /admin/import-jobs/{jobId}`
  - `GET /admin/import-jobs/{jobId}/events`
  - `GET /admin/import-jobs/{jobId}/artifacts`
  - `POST /admin/import-jobs/{jobId}/cancel`
- Job lifecycle states:
  - `queued`
  - `running`
  - `completed`
  - `failed`
  - `cancelled`
- Legal state transitions:
  - `queued -> running`
  - `queued -> cancelled`
  - `running -> completed`
  - `running -> failed`
  - `running -> cancelled`
- Job creation preconditions:
  - admin only
  - active Source
  - active Connector
  - approved ConnectorVersion
  - required Secrets bound and not revoked
  - manual + unattended only
  - one queued/running Job per Source

## Explicit Boundaries

- M1.2A creates queued Jobs only.
- M1.2A does not run Connectors.
- M1.2A does not implement Runner, Docker, workspace execution, JSON Lines parsing, or Artifact production.
- Running Job cancellation only records `cancellation_requested_at`; it does not claim to kill a process.
- Completed, failed, and cancelled Job cancellation is idempotent.
- APIs return metadata only and do not return Secret values, Connector package content, internal paths, media, Program, Episode, RSS, or subscription data.

## Deferred

- Runner protocol and fixture execution: M1.2B.
- Docker trusted-admin runner: M1.2C.
- Full admin task console and Alpha deployment preparation: M1.2D.
