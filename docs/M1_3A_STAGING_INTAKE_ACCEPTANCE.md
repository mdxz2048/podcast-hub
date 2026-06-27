# M1.3A Staging Intake Acceptance

Status: accepted for Deployable Alpha.

## Scope

M1.3A converts a completed Import Job's registered artifacts into review-pending staging content.

Implemented:

- `programs`, `program_sources`, `episodes`, `media_assets`, `review_items`, `publication_events`, and `intake_runs` tables.
- `import_job_artifacts.storage_key` for private promoted artifact retention.
- `internal/intake` service for strict metadata bundle parsing.
- Admin-only intake and staging APIs.
- Admin staging UI at `/admin/staging`.
- Import Job detail action for manual intake into the review-pending area.

Not implemented:

- Publishing.
- RSS generation.
- User-visible Program or Episode exposure.
- Real duoting.
- Scheduled, interactive, or QR jobs.
- Public media download URLs.
- Object storage production implementation.

## API

- `POST /admin/import-jobs/{jobId}/intake`
- `GET /admin/import-jobs/{jobId}/intake-status`
- `GET /admin/staging/programs`
- `GET /admin/staging/programs/{programId}`
- `GET /admin/staging/episodes`
- `GET /admin/staging/episodes/{episodeId}`

All endpoints require `admin`; write endpoints require CSRF protection.

## Intake Rules

- Only `completed` Import Jobs can be intaken.
- Metadata bundle must be an already registered `metadata_bundle` artifact.
- Bundle JSON rejects unknown fields.
- Artifact references must point to already registered ImportJob artifacts.
- Absolute paths, path traversal, URLs, and secret-like metadata keys are rejected.
- Intake never executes Connector code, never reads Docker Socket, never reads Secret values, and never reads arbitrary workspace files.
- Same Source + `external_program_id` updates the same Program candidate.
- Same Program + `external_episode_id` updates the same Episode candidate.
- Title equality alone never merges content across Sources.
- Intake is idempotent after success.

## Security Boundary

API responses do not return:

- Secret values.
- Connector package contents.
- Absolute filesystem paths.
- Internal storage keys.
- Artifact file contents.
- RSS links.
- Media download links.

Media assets stay private in staging metadata until later review and publish phases.

## Verification

Executed for this phase:

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```

Screenshot workflows were intentionally not run.
