# M1.3B Review Publication Acceptance

Status: accepted for Deployable Alpha.

## Scope

M1.3B adds the administrator review and publication state workflow for staging Program and Episode candidates.

Implemented:

- Admin review queue APIs.
- Review approve/reject with required reject reason.
- Safe Program and Episode metadata patch APIs.
- Program/Episode submit-review APIs.
- Program/Episode publish APIs with backend precondition checks.
- Program/Episode archive APIs.
- Publication event recording for approve, reject, publish, archive, and metadata updates.
- Real API-backed admin Program, Episode, Review list/detail pages.

Not implemented:

- User-visible catalog access.
- RSS generation or RSS management.
- Batch publish.
- Public media download.
- Real duoting.
- Scheduled, interactive, or QR jobs.

## State Rules

- `approved` is not `published`.
- Intake never publishes content.
- Review approve moves the target to `approved`.
- Review reject requires a reason and moves the target to `rejected`.
- Program publish requires:
  - Program status `approved`.
  - No pending Program review.
  - Valid title.
  - Existing Source reference.
- Episode publish requires:
  - Episode status `approved`.
  - Parent Program status `published`.
  - No pending Episode review.
  - Approved audio MediaAsset.
- Archive moves the target to `archived`.
- Published metadata updates write audit events and create or keep a pending metadata review.

## API

- `GET /admin/review`
- `GET /admin/review/{reviewId}`
- `POST /admin/review/{reviewId}/approve`
- `POST /admin/review/{reviewId}/reject`
- `GET /admin/programs`
- `GET /admin/programs/{programId}`
- `PATCH /admin/programs/{programId}`
- `POST /admin/programs/{programId}/submit-review`
- `POST /admin/programs/{programId}/publish`
- `POST /admin/programs/{programId}/archive`
- `GET /admin/episodes/{episodeId}`
- `PATCH /admin/episodes/{episodeId}`
- `POST /admin/episodes/{episodeId}/submit-review`
- `POST /admin/episodes/{episodeId}/publish`
- `POST /admin/episodes/{episodeId}/archive`

All write APIs require admin auth and CSRF validation.

## Security Boundary

Admin APIs do not return:

- Secret values.
- Connector package contents.
- Raw Job logs.
- Absolute filesystem paths.
- Staging storage keys.
- Artifact file contents.
- Public media download links.

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
