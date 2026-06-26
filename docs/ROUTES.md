# Routes

## 1. Scope

This document defines proposed frontend routes for M0.1 and M0.2 static experiences.

Routes are static and Mock-data driven until a later milestone. They do not imply real APIs, authentication, RSS generation, upload, Connector execution, or backend services.

## 2. Public Routes

| Route | Page | M0 Phase | Notes |
| --- | --- | --- | --- |
| `/` | Home | M0.1 | Static product home with content-led visual direction. |
| `/register` | Register | M0.1 | Static registration UI; no real Turnstile or email. |
| `/register/verify` | Email verification | M0.2A | Static six-digit code UI; no real verification service. |
| `/login` | Login | M0.1 | Static login UI; no real authentication. |
| `/forgot-password` | Password reset request | M0.2A | Static request flow with non-enumerating copy; no email service. |

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

## 5. Route Guard States

M0 static pages should include Mock route guard states:

- Public allowed.
- Authenticated user.
- Admin required.
- Permission denied.
- Suspended user.
- Deleted user.

No real route guard is implemented in M0.1.
