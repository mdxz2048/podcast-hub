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
| `/programs` | Program browse | M0.1 | Static Mock authorized Programs. |
| `/programs/:programId` | Program detail | M0.2 | Static Program and episode preview. |
| `/collections` | My Collections | M0.2 | Static collection list. |
| `/collections/:collectionId` | Collection detail/editor | M0.2 | Static editor and RSS state. |
| `/collections/:collectionId/subscribe` | Collection RSS subscription | M0.2A | Static simulated RSS URL using example.invalid; no RSS XML generation. |
| `/rss` | RSS feeds | M0.2 | Static token state only. |
| `/account` | Account | M0.2 | Static account overview. |
| `/account/sessions` | Account sessions | M0.2 | Static future capability. |

## 4. Admin Routes

| Route | Page | M0 Phase | Notes |
| --- | --- | --- | --- |
| `/admin` | Admin overview | M0.1 | Static operational overview. |
| `/admin/programs` | Program list | M0.1 | Static Program management list. |
| `/admin/programs/:programId` | Program detail | M0.2B | Static Program operation view with Sources, Episodes, publication, and activity sections. |
| `/admin/connectors` | Connector list | M1.1A | Real admin Connector registry list API. |
| `/admin/connectors/new` | Connector upload | M1.1A | Real multipart ZIP upload + static validation result display; no execution. |
| `/admin/connectors/:connectorId` | Connector detail | M1.1A | Real version/review status + manifest summary + enable/disable actions. |
| `/admin/sources` | Connector Source list | M1.1B | Real admin Source list; no fake Source rows. |
| `/admin/sources/new` | Connector Source create | M1.1B | Creates draft Source from active approved ConnectorVersion only. |
| `/admin/sources/:sourceId` | Connector Source detail | M1.1B | Shows Secret binding state only, never Secret values. |
| `/admin/secrets` | Secret metadata | M1.1B | Encrypted Secret write and metadata list; no read-value API. |
| `/admin/import-jobs` | Import job list | M1.2B | Real Import Job metadata list; no fake jobs or Program/Episode output. |
| `/admin/import-jobs/:jobId` | Import job detail | M1.2B | Real Job metadata, redacted events, artifact metadata, and cancel action. |
| `/admin/reviews` | Review queue | M0.2B | Static review queue with Drawer details and confirmation Dialogs. |
| `/admin/publications` | Publications | M0.2 | Static RSS publication state. |
| `/admin/users` | Users and access | M0.2B | Static user/access view; no invitations or real user management in M0. |
| `/admin/audit` | Audit log | M0.2 | Static audit list. |
| `/admin/settings` | Settings | M0.2 | Static settings view. |

Related real admin APIs in M1.0C:

- `GET /healthz`
- `GET /admin/me`
- `GET /admin/system/status`

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

## 5. Route Guard States

Route guard behavior in M1.0C:

- `/admin/*` unauthenticated: redirect to `/login`.
- `/admin/*` authenticated `user`: render Permission Denied.
- `/admin/*` authenticated `admin`: allow access.
- Backend `RequireAdmin` enforces 401/403 regardless of frontend route state.
