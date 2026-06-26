# Testing and Visual QA

## 1. Scope

This document defines M0 testing and visual QA policy.

M0.1 uses Playwright for screenshot foundations and static route checks. It does not test real APIs, authentication, databases, upload, RSS generation, Connector execution, or Docker services.

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

