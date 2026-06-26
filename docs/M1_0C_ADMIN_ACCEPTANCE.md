# M1.0C Admin Acceptance

## Scope

M1.0C only delivers:

- local first-admin bootstrap CLI
- backend `RequireAuth` + `RequireAdmin` enforcement
- minimal admin identity/status APIs
- frontend admin role redirect and `/admin/*` guard

Out of scope (still Mock / not implemented here):

- real program/source/import job CRUD
- connector execution
- RSS publishing
- review workflow backend

## Admin bootstrap

- Command: `go run ./cmd/admin seed --email admin@example.invalid`
- Password source:
  - `--password`
  - `--password-env` (default key: `ADMIN_SEED_PASSWORD`)
  - development interactive prompt fallback
- No HTTP endpoint for admin creation.
- Existing admin email: idempotent success.
- Existing non-admin email:
  - default reject
  - `--promote` only allowed in development.

## Backend permission chain

- `RequireAuth`
  - no session: `401 not_authenticated`
- `RequireAdmin`
  - authenticated non-admin: `403 forbidden`
- suspended/deleted admin cannot pass auth resolution (`403 account_unavailable`)

APIs:

- `GET /admin/me`
- `GET /admin/system/status`

## Frontend behavior

- `/login` shared by user/admin.
- login success:
  - `role=admin` -> `/admin`
  - `role=user` -> `/programs`
- `/admin/*` guard:
  - unauthenticated -> `/login`
  - non-admin -> PermissionDenied
  - admin -> existing Mock admin pages
- admin layout shows current admin email + logout action.

## Validation status

- Go unit tests: includes admin seed service + admin middleware/API access control.
- Frontend tests: includes admin/user login redirect and `/admin` guard behavior.
- Build/type checks: pass in this phase.

## Deferred

- dedicated admin login UX (`/admin/login`) intentionally not introduced.
- admin business modules remain Mock by plan.
