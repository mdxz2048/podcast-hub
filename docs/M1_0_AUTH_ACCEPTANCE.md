# M1.0 Authentication Acceptance & Security Audit

## 1. Scope Check (M1.0B)

### 1.1 `/reset-password` route/page

- **Before fix:** missing dedicated `/reset-password` route.
- **Now:** completed.
  - Route added: `/reset-password` (`src/main.tsx`)
  - Page added: `src/routes/ResetPasswordPage.tsx`
  - Navigation from `/forgot-password` to `/reset-password?email=...`

### 1.2 Reset flow completeness

- Input verification code: **Pass**
- Input new password + confirm password: **Pass**
- Call `POST /auth/password-reset/verify`: **Pass**
- Success jump to login: **Pass** (auto navigate to `/login?reset=done`)
- Show “other sessions logged out”: **Pass**
- Handle code wrong / expired / rate limit / network errors: **Pass**

Implementation references:

- `src/routes/ResetPasswordPage.tsx`
- `src/routes/ForgotPasswordPage.tsx`
- `src/routes/LoginPage.tsx`
- `src/api/auth.ts`
- `internal/http/server.go` (error mapping)
- `internal/auth/errors.go`
- `internal/store/postgres/auth_store.go`

---

## 2. Smoke Test Result (development)

## Required startup chain

- PostgreSQL
- Redis
- Mailpit
- Go API
- Vite frontend

## Execution status

- `docker version` / `docker compose version` → **Passed**
- `docker compose up -d postgres redis mailpit` / `docker compose ps` → **Passed**
- Go API startup with `.env.example` development placeholder config (`TURNSTILE_MODE=mock`, SMTP to Mailpit) → **Passed**
- migration check (`SELECT version FROM schema_migrations`) → **Passed** (`0001_auth_tables`)
- Real chain executed and validated:
  - register request
  - registration mail received in Mailpit
  - verify code and user status transition `pending_verification -> active`
  - `GET /auth/me` authenticated
  - logout then `GET /auth/me` unauthenticated
  - login with correct password restores session
  - login failures (known email vs unknown email) both return generic `invalid_credentials`
  - password-reset request mail received
  - password-reset verify succeeds
  - old sessions invalidated
  - old password login fails, new password login succeeds
- Frontend security prompt check (no screenshots):
  - `/login?reset=done` displays “其他登录设备已退出” → **Passed**

## Conclusion

- **Smoke test Passed**.
- No secrets, cookies, passwords, full codes, reset proofs, or tokens were printed in logs or report.

---

## 3. Security Audit Checklist (code-level)

## 3.1 Production mandatory security config

- TURNSTILE_MODE not mock in production: **Pass**
- TURNSTILE_SECRET_KEY required in production: **Pass**
- SESSION_COOKIE_SECURE must be true in production: **Pass**
- SESSION_PEPPER required: **Pass**
- AUTH_CODE_PEPPER required: **Pass**
- DATABASE_URL required: **Pass**
- REDIS_URL required in production: **Pass**
- FRONTEND_ORIGIN required: **Pass**

Evidence: `config/config.go`

## 3.2 Redis failure strategy

- Production cannot fall back to unlimited memory limiter: **Pass**
  - Production config requires REDIS_URL.
  - Redis ping failure in production exits process.
- Development/test memory limiter allowed: **Pass**

Evidence: `config/config.go`, `cmd/api/main.go`, `internal/ratelimit/limiter.go`

## 3.3 Frontend storage and API usage

- No `localStorage`/`sessionStorage` for auth secrets: **Pass**
- Auth API uses `credentials: include`: **Pass**
- Turnstile secret absent from frontend: **Pass**
- Password/code not written into URL: **Pass**

Evidence: `src/api/client.ts`, auth routes/components.

## 3.4 Cookie, CORS, CSRF

- CORS does not use credentials + wildcard: **Pass**
- Cookie uses HttpOnly + SameSite + configurable Secure/Domain: **Pass**
- State-changing CSRF protection is active (origin/referer + logout token): **Pass**
- Empty origin/referer cross-site POST is rejected: **Pass**

Evidence: `internal/http/middleware.go`, `internal/http/server.go`, `internal/http/auth_handlers.go`

## 3.5 Data handling and lifecycle

- Password/code/reset/session stored as hash/non-reusable representation: **Pass**
- Audit logs avoid sensitive values: **Pass**
- Email normalization consistent: **Pass**
- Resend verification invalidates old code: **Pass**
- Password reset revokes all user sessions: **Pass**

Evidence: `internal/security/password.go`, `internal/auth/service.go`, `internal/store/postgres/auth_store.go`, `migrations/0001_auth_tables.up.sql`

## 3.6 API behavior

- Register/login/password-reset request non-enumerating behavior: **Pass**
- Unified error response shape: **Pass**
- Internal details not exposed in API errors: **Pass**
- suspended/deleted users cannot establish effective session: **Pass**
  - Login blocked for non-active status.
  - Session resolve enforces active status.
  - Registration for suspended/deleted/admin blocked.

Evidence: `internal/auth/service.go`, `internal/http/server.go`

---

## 4. Fixes made during this acceptance pass

1. Added dedicated reset page/route:
   - `/reset-password`
   - separated from `/forgot-password`.
2. Added explicit reset result UX:
   - success message includes “其他登录设备已退出”
   - auto redirect to login.
3. Added reset-specific error codes:
   - `invalid_reset_proof`
   - `reset_proof_expired`
4. Hardened registration against status bypass:
   - suspended/deleted/admin accounts blocked from registration-code flow.
5. Added tests for suspended-account re-registration blocking.

---

## 5. Deferred items

1. Browser-level DevTools manual inspection for HttpOnly flag is not included in this CLI-only report; server-side cookie config remains verified in code and behavior tests.

---

## 6. M1.0 submission readiness

- `/reset-password` flow completeness: **Yes**
- Security audit critical issues: **No critical issue found in code audit**
- Build/tests: **Pass** (see command log)
- Full required smoke chain: **Passed**

## Final decision

- **Ready to submit/push for M1.0**.
