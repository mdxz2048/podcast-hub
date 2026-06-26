# Testing and Visual QA

## 1. Scope

This document defines M0+M1 testing and visual QA policy.

M1.0 keeps Playwright screenshot acceptance and adds Go unit tests for authentication/security logic. Only account flows are real in M1.0.

## 2. Frozen Tooling

- Playwright for E2E-style static route checks and screenshot acceptance.
- No Storybook in M0.
- Internal component showcase route may be used in M0.

## 3. M0.1 Screenshot Targets

Representative pages:

- Home.
- Register.
- Login.
- Program browse.
- Admin overview.
- Admin Program list.

Required viewport classes:

- Desktop.
- Mobile.

Required state variants:

- Loading.
- Empty.
- Error.
- Permission Denied.
- Success Feedback.
- Long text.

## 4. M0.2 Screenshot Targets

M0.2 expands screenshots to all core static pages:

- My Collections.
- Collection editor.
- RSS feeds.
- Account.
- Admin Source list.
- Admin Connector list and detail.
- Admin Import Job list and detail.
- Admin Review queue.
- Admin Publications.
- Admin Users and Access.
- Admin Audit.
- Admin Settings.

## 5. Visual QA Checks

Every visual acceptance pass should check:

- Design Token usage.
- Shared component usage.
- Consistent spacing.
- Consistent typography.
- Consistent status colors.
- No unreadable text.
- No overlapping UI.
- No horizontal overflow on mobile.
- No browser auto-dark-mode readability regressions.

## 6. Dark Mode QA

Phase 1 does not support dark mode.

QA requirement:

- Explicitly document dark mode unsupported in screenshot reports.
- Ensure pages remain readable if browser or OS has dark mode enabled.
- Components must remain ready for future `data-theme="dark"` token overrides.

## 7. Non-Goals

Do not add in M0.1:

- API integration tests.
- Database tests.
- Connector runtime tests.
- Upload tests.
- RSS XML tests.
- Docker smoke tests.

## 8. M1.0 Authentication baseline

Backend:

- `go test ./...` for auth service, security helpers, turnstile verifier, and production config fail-closed checks.

Frontend:

- Keep screenshot coverage for auth key states:
  - register turnstile mock
  - verification code error
  - verification code expired
  - login error
  - forgot-password generic success hint
  - mobile register verify page
## M1.1A test focus

Backend:

- Admin authz on `/admin/connectors*` APIs.
- ZIP static validation rules (zip-slip, symlink, forbidden files/extensions, manifest checks).
- Connector version immutability and review state transitions.

Frontend:

- `/admin/connectors`, `/admin/connectors/new`, `/admin/connectors/:connectorId` real API integration behavior.
- Empty/loading/error states and admin-only access behavior.
- No execution/run/download-media controls shown on connector management pages.

This phase does not run screenshot workflows.
