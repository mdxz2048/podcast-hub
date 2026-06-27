# Routes

## 1. Scope

This document defines frontend routes after M1.1B Source and Secret Reference integration.

Account and admin identity routes are connected to real backend APIs. Connector and Source/Secret admin pages are also connected to real APIs. Program/collection/review/import workflow/RSS pages remain Mock-data driven.

## 2. Public Routes

| Route | Page | M0 Phase | Notes |
| --- | --- | --- | --- |
| `/` | Home | M0.1 | Static product home with content-led visual direction. |
| `/register` | Register | M1.0 | Real `POST /auth/register/request-code` + Turnstile token submit. |
| `/register/verify` | Email verification | M1.0 | Real `POST /auth/register/verify-code`, success creates cookie session. |
| `/login` | Login | M1.0C | Real `POST /auth/login`; admin and user share this route and role-based redirect. |
| `/forgot-password` | Password reset | M1.0 | Real `POST /auth/password-reset/request` and `POST /auth/password-reset/verify`. |
| `/reset-password` | Password reset verify | M1.0 | Real reset verification form (`code + new password`) calling `POST /auth/password-reset/verify`. |

## 3. User App Routes

| Route | Page | M0 Phase | Notes |
| --- | --- | --- | --- |
| `/programs` | Program browse | M1.3C | Real authorized Program API; no mock fallback. |
| `/programs/:programId` | Program detail | M1.3C | Real authorized Program and published Episode API. |
| `/collections` | My Collections | M1.3C | Real owner-scoped personal collections. |
| `/collections/:collectionId` | Collection detail/editor | M1.3C | Real collection edit, add, and remove Program API. |
| `/collections/:collectionId/subscribe` | Collection RSS subscription | M0.2A | Static simulated RSS URL using example.invalid; no RSS XML generation. |
| `/rss` | RSS feeds | M1.5C | Real private RSS Feed list/create/rotate/revoke/delete APIs. |
| `/account` | Account | M0.2 | Static account overview. |
| `/account/sessions` | Account sessions | M0.2 | Static future capability. |

## 4. Admin Routes

| Route | Page | M0 Phase | Notes |
| --- | --- | --- | --- |
| `/admin` | Admin overview | M0.1 | Static operational overview. |
| `/admin/programs` | Program list | M1.3B | Real admin Program metadata and lifecycle status. |
| `/admin/programs/:programId` | Program detail | M1.3B | Real metadata edit, submit review, publish, archive, and Episode links. |
| `/admin/episodes/:episodeId` | Episode detail | M1.3B | Real metadata edit, submit review, publish, and archive controls. |
| `/admin/connectors` | Connector list | M1.1A | Real admin Connector registry list API. |
| `/admin/connectors/new` | Connector upload | M1.1A | Real multipart ZIP upload + static validation result display; no execution. |
| `/admin/connectors/:connectorId` | Connector detail | M1.1A | Real version/review status + manifest summary + enable/disable actions. |
| `/admin/sources` | Connector Source list | M1.1B | Real admin Source list; no fake Source rows. |
| `/admin/sources/new` | Connector Source create | M1.1B | Creates draft Source from active approved ConnectorVersion only. |
| `/admin/sources/:sourceId` | Connector Source detail | M1.1B | Shows Secret binding state only, never Secret values. |
| `/admin/secrets` | Secret metadata | M1.1B | Encrypted Secret write and metadata list; no read-value API. |
| `/admin/import-jobs` | Import job list | M1.2B | Real Import Job metadata list; no fake jobs or Program/Episode output. |
| `/admin/import-jobs/:jobId` | Import job detail | M1.3A | Real Job metadata, redacted events, artifact metadata, cancel action, and guarded manual intake action. |
| `/admin/staging` | Staging intake list | M1.3A | Real review-pending Program/Episode candidates from completed ImportJob artifacts. |
| `/admin/staging/programs/:programId` | Staging Program detail | M1.3A | Admin-only candidate metadata; no publish/RSS/user link. |
| `/admin/staging/episodes/:episodeId` | Staging Episode detail | M1.3A | Admin-only candidate metadata; no path/download/RSS/user link. |
| `/admin/reviews` | Review queue | M1.3B | Real pending/approved/rejected review items with approve/reject actions. |
| `/admin/review/:reviewId` | Review detail | M1.3B | Real review detail and approve/reject actions. |
| `/admin/rss-feeds` | RSS feed metadata | M1.5C | Real safe RSS metadata list and admin revoke. |
| `/admin/publications` | Publications | M0.2 | Static RSS publication state. |
| `/admin/users` | Users and access | M0.2B | Static user/access view; no invitations or real user management in M0. |
| `/admin/audit` | Audit log | M0.2 | Static audit list. |
| `/admin/settings` | Settings | M0.2 | Static settings view. |

Related real admin APIs in M1.0C:

- `GET /healthz`
- `GET /readyz`
- `GET /admin/me`
- `GET /admin/system/status`

All API error responses may include `error.request_id`. This is not a route parameter or authentication credential; it is an opaque server-generated support correlation value and must not be logged or displayed as user identity.

Additional real admin Connector APIs in M1.1A:

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

Additional real admin Source and Secret APIs in M1.1B:

- `GET /admin/sources`
- `POST /admin/sources`
- `GET /admin/sources/{sourceId}`
- `PATCH /admin/sources/{sourceId}`
- `POST /admin/sources/{sourceId}/enable`
- `POST /admin/sources/{sourceId}/disable`
- `GET /admin/secrets`
- `POST /admin/secrets/text`
- `POST /admin/secrets/file`
- `POST /admin/secrets/{secretId}/revoke`
- `POST /admin/sources/{sourceId}/secret-bindings`
- `DELETE /admin/sources/{sourceId}/secret-bindings/{bindingId}`

Additional real admin Import Job APIs in M1.2A:

- `GET /admin/import-jobs`
- `POST /admin/sources/{sourceId}/import-jobs`
- `GET /admin/import-jobs/{jobId}`
- `GET /admin/import-jobs/{jobId}/events`
- `GET /admin/import-jobs/{jobId}/artifacts`
- `POST /admin/import-jobs/{jobId}/cancel`

M1.2B adds the independent `cmd/runner` process and does not add user-facing routes. Existing Import Job APIs continue to return metadata only: no Secret values, no package content, no absolute paths, no Artifact file contents, no media download, and no RSS publication actions.

M1.2D adds UI access to the Import Job workflow:

- Source detail can create a manual Import Job when the Source is active, manual, and unattended.
- Import Job list/detail pages show real API metadata and Runner disabled reason from `GET /admin/system/status`.
- Artifact display is metadata-only.
- No publish, RSS, subscription, or media download controls are exposed.
- `/healthz` is liveness only; `/readyz` is safe readiness only.

Additional real admin staging intake APIs in M1.3A:

- `POST /admin/import-jobs/{jobId}/intake`
- `GET /admin/import-jobs/{jobId}/intake-status`
- `GET /admin/staging/programs`
- `GET /admin/staging/programs/{programId}`
- `GET /admin/staging/episodes`
- `GET /admin/staging/episodes/{episodeId}`

M1.3A routes remain admin-only. They expose review-pending candidate metadata only: no Secret values, no storage keys, no absolute paths, no Artifact file contents, no published media URLs, no RSS, and no normal-user visibility.

Additional real admin review and publication APIs in M1.3B:

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

M1.3B still does not expose published content to normal users. User catalog and access grants are deferred to M1.3C.

M1.3C user catalog APIs:

- `GET /programs`
- `GET /programs/{programId}`
- `GET /programs/{programId}/episodes`
- `GET /episodes/{episodeId}`
- `GET /me/collections`
- `POST /me/collections`
- `PATCH /me/collections/{collectionId}`
- `DELETE /me/collections/{collectionId}`
- `POST /me/collections/{collectionId}/programs`
- `DELETE /me/collections/{collectionId}/programs/{programId}`

M1.3C admin access APIs:

- `GET /admin/programs/{programId}/access-grants`
- `POST /admin/programs/{programId}/access-grants`
- `POST /admin/program-access/{grantId}/revoke`

Normal-user routes expose only authorized, published catalog data. Unauthorized reads use generic not-found errors and never disclose Program titles, Source, Connector, ImportJob, Artifact, staging, storage keys, Secrets, or tokens.

M1.5C RSS APIs are connected to frontend routes:

- `GET /me/rss-feeds`
- `POST /me/rss-feeds`
- `POST /me/rss-feeds/{feedId}/rotate`
- `POST /me/rss-feeds/{feedId}/revoke`
- `DELETE /me/rss-feeds/{feedId}`
- `GET /admin/rss-feeds`
- `POST /admin/rss-feeds/{feedId}/revoke`

Only create and rotate responses expose plaintext token URLs. RSS list and admin list responses expose token prefixes only.

D2 deployment candidate adds no new application routes. It adds reverse proxy routing for frontend, API, and `/rss/private/*` token paths in `deploy/Caddyfile.user-beta.template`.

## 5. Route Guard States

Route guard behavior in M1.0C:

- `/admin/*` unauthenticated: redirect to `/login`.
- `/admin/*` authenticated `user`: render Permission Denied.
- `/admin/*` authenticated `admin`: allow access.
- Backend `RequireAdmin` enforces 401/403 regardless of frontend route state.
