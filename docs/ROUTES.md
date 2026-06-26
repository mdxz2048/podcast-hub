# Routes

## 1. Scope

This document defines frontend routes after M1.0 and M1.0C authentication/admin integration.

Account and admin identity routes are connected to real backend APIs. Program, collection, admin business workflow, RSS, and Connector pages remain Mock-data driven.

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
| `/admin/sources` | Source list | M0.2 | Static source status list. |
| `/admin/connectors` | Connector list | M0.2B | Static registry list. |
| `/admin/connectors/new` | Connector registration wizard | M0.2B | Static Mock wizard; no real ZIP upload, parsing, validation, or execution. |
| `/admin/connectors/:connectorId` | Connector detail | M0.2B | Static manifest/status view. |
| `/admin/import-jobs` | Import job list | M0.2B | Static job status list. |
| `/admin/import-jobs/:jobId` | Import job detail | M0.2B | Static sanitized log, timeline, QR placeholder, and output summary. |
| `/admin/reviews` | Review queue | M0.2B | Static review queue with Drawer details and confirmation Dialogs. |
| `/admin/publications` | Publications | M0.2 | Static RSS publication state. |
| `/admin/users` | Users and access | M0.2B | Static user/access view; no invitations or real user management in M0. |
| `/admin/audit` | Audit log | M0.2 | Static audit list. |
| `/admin/settings` | Settings | M0.2 | Static settings view. |

Related real admin APIs in M1.0C:

- `GET /admin/me`
- `GET /admin/system/status`

## 5. Route Guard States

Route guard behavior in M1.0C:

- `/admin/*` unauthenticated: redirect to `/login`.
- `/admin/*` authenticated `user`: render Permission Denied.
- `/admin/*` authenticated `admin`: allow access.
- Backend `RequireAdmin` enforces 401/403 regardless of frontend route state.
