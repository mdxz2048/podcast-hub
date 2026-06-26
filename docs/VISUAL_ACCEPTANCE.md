# Visual Acceptance

## 1. Scope

Every user-visible feature must pass visual and interaction acceptance before it is considered complete.

This applies to:

- Public pages.
- Authentication pages.
- User app pages.
- Admin pages.
- Error and permission states.
- Async workflows.
- Review and publication workflows.

## 2. Required Evidence

For every user-visible feature, provide:

- Desktop screenshot.
- Mobile screenshot.
- Empty state screenshot.
- Loading state screenshot.
- Error state screenshot.
- Long text screenshot.
- Permission denied screenshot.
- Async operation success feedback screenshot.
- Explicit note that dark mode is not supported in phase 1, plus a check that browser or OS automatic dark mode does not make pages unreadable.
- Mock data verification.
- Real boundary data verification.

Screenshots must represent the actual implemented UI in later phases, not design mockups only.

## 3. Data Scenarios

### 3.1 Mock Data

Use Mock data to verify:

- Normal content density.
- Common user journey.
- Common admin workflow.
- Expected successful states.

### 3.2 Real Boundary Data

Use boundary data to verify:

- Very long Program titles.
- Missing cover images.
- Many Sources on one Program.
- Many Connector versions.
- Failed jobs with long error messages.
- Large review queues.
- Users with no access.
- RSS token revoked state.
- Suspended user state.

## 4. State Coverage

Each page must define and verify:

- Initial state.
- Loading state.
- Empty state.
- Error state.
- Permission denied state.
- Success state.
- Long text state.
- Mobile state.

Async actions must verify:

- Button disabled or progress state during submission.
- Success feedback.
- Error feedback.
- Retry path when safe.

## 5. Authentication Acceptance

Registration must include screenshots for:

- Initial form.
- Turnstile area.
- Field validation.
- Code sent.
- Code expired.
- Too many attempts.
- Success state.

Login must include screenshots for:

- Initial form.
- Generic invalid credentials error.
- Turnstile required.
- Loading.
- Success redirect or authenticated shell.

Password reset must include screenshots for:

- Request form.
- Generic instruction sent state.
- Reset proof verification.
- New password validation.
- Success and session invalidation note.

## 6. User App Acceptance

Program browse must verify:

- Program grid or list on desktop.
- Program list on mobile.
- No authorized Programs.
- Long Program titles and descriptions.
- Missing cover fallback.
- Permission changed state.

My Collections must verify:

- No collections.
- Collection list.
- RSS URL active.
- RSS URL revoked.
- Program access removed from collection.

Collection editor must verify:

- Empty collection.
- Many authorized Programs.
- Unsaved changes.
- Save success.
- Save error.

## 7. Admin Acceptance

Admin overview must verify:

- Normal operating state.
- Failed imports.
- QR tasks waiting.
- Review backlog.
- Rights hold.

Program management must verify:

- Program list.
- Long titles.
- No Programs.
- Program disabled.
- Rights hold.
- Publication paused.

Connector management must verify:

- Upload pending.
- Validation failed.
- Pending approval.
- Approved.
- Deprecated.
- Revoked.

Import jobs must verify:

- Queued.
- Running.
- Waiting for authentication.
- Completed.
- Completed with warnings.
- Failed.
- Timed out.
- Sanitized log viewer.

Review queue must verify:

- Empty queue.
- Pending review.
- Duplicate warning.
- Rights hold block.
- Approval success.
- Rejection success.

## 8. Design Consistency Checklist

Before acceptance:

- Uses shared design tokens.
- Uses shared typography scale.
- Uses shared spacing system.
- Uses shared color semantics.
- Uses shared button hierarchy.
- Uses shared status badges.
- Uses shared empty/loading/error/success patterns.
- Does not introduce one-off visual language.
- Does not overuse cards.
- Does not rely on meaningless icons.
- Does not use excessive gradients or glass effects.

## 9. Responsive Acceptance

Desktop screenshots should cover:

- Primary laptop width.
- Wide desktop if layout meaningfully changes.

Mobile screenshots should cover:

- Narrow mobile.
- Long content scroll.
- Sticky or bottom actions where relevant.

Mobile requirements:

- No incoherent overlap.
- No horizontal page overflow.
- Text fits inside controls.
- Main actions remain discoverable.
- Drawers become full-screen sheets where needed.

## 10. Accessibility Acceptance

Before visual acceptance:

- Keyboard focus is visible.
- Forms have labels.
- Errors are associated with fields.
- Color is not the only state indicator.
- Contrast meets target accessibility levels.
- Icon-only buttons have accessible names.
- Loading and async state changes are announced where appropriate.

## 11. Dark Mode Policy

Phase 1 explicitly does not support dark mode.

Because dark mode is not supported in phase 1:

- Document the decision on each visual acceptance report.
- Use semantic color tokens instead of hard-coded palette values.
- Avoid assets that only work on one background.
- Avoid shadows and borders that cannot be tokenized later.
- Ensure future dark mode can be added without redesigning component APIs.
- Keep future `data-theme="dark"` extension possible.
- Ensure browser automatic dark mode does not create unreadable UI.

## 12. M0 Visual Gate

M0 must not be treated as engineering skeleton only.

M0.1 must include:

- Frontend engineering skeleton.
- Design tokens.
- Component library.
- Public layout, user layout, and admin layout.
- Responsive rules.
- Mock data infrastructure.
- Representative high-fidelity static pages:
  - Home page.
  - Registration page.
  - Login page.
  - Program browse page.
  - Admin overview page.
  - Admin Program list page.
- Loading, Empty, Error, Permission Denied, and Success Feedback states.
- Playwright screenshot foundation.

M0.2 must include:

- All core static pages.
- Complete Mock data.
- Long text, no permission, task failure, waiting QR, pending review, and empty-list boundary states.
- User app and admin app core navigation paths.
- No real backend.

M0.2A user-side visual evidence must include:

- Program detail normal and access-restricted states.
- Add-to-collection Drawer.
- My Collections normal and empty states.
- Collection editor normal, empty collection, no-preview rule, unsaved, and permission-denied states.
- Collection subscribe page normal, copied Toast, reset confirmation Dialog, revoked, and permission-denied states.
- Email verification page.
- Forgot password page.
- Desktop and mobile screenshots for the core user path.

M0.2B admin-side visual evidence must include:

- Admin Program detail normal, draft, authorization-pending, paused, no-source, no-episode, long-text, and permission-denied states.
- Connector list, detail, and static registration wizard.
- Import job list and detail states for queued, running, waiting_auth, waiting_manual_upload, review_pending, completed, failed, and cancelled.
- Non-scannable QR placeholder viewport screenshot for waiting_auth interactive jobs.
- Review queue normal, empty, error-like risk, permission-denied, Drawer, rejection Dialog, and success Toast states.
- Users and access page showing user/admin roles, user statuses, responsibility labels, RSS state, and permission summary.
- Overlay states should include viewport screenshots with names that contain `viewport-overlay`.

M0.3 must include:

- Playwright desktop and mobile screenshots.
- Page consistency check.
- Visual defect fixes.
- Information hierarchy check.
- Basic accessibility check.
- Visual acceptance report.

M1 must not begin until M0.3 is accepted.
