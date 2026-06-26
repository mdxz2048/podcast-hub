# Frontend Architecture

## 1. Frozen M0 Technology Stack

Frontend stack:

- React.
- TypeScript.
- Vite.
- pnpm.
- React Router.
- Tailwind CSS.
- CSS Variables as the only source of design tokens.
- lucide-react.
- Playwright for E2E and screenshot acceptance.

Storybook:

- Not introduced in M0.
- M0 uses an internal component showcase page or development route.
- Storybook may be reconsidered after the component system stabilizes.

## 2. M0.1 Scope

M0.1 may include:

- Frontend engineering skeleton.
- Design Token setup.
- Base component library.
- Public, user, and admin layouts.
- Responsive rules.
- Mock data infrastructure.
- Representative high-fidelity static pages:
  - Home.
  - Register.
  - Login.
  - Program browse.
  - Admin overview.
  - Admin Program list.
- Loading, Empty, Error, Permission Denied, and Success Feedback states.
- Playwright screenshot foundation.

M0.1 must not include:

- Real authentication.
- Email sending.
- Cloudflare Turnstile integration.
- Password hashing.
- Database.
- Real APIs.
- Audio upload.
- RSS XML.
- Connector ZIP upload.
- Python Connector execution.
- QR scanning.
- Docker.
- Production deployment.

## 3. Token Source

CSS Variables are the only source of design tokens.

Rules:

- Tailwind must consume or mirror CSS variable tokens.
- Components must not hard-code design values that bypass tokens.
- Future `data-theme="dark"` support must be structurally possible.
- Phase 1 does not support dark mode.

## 4. Suggested Directory Shape

This is a planning shape only, not created in M0.0:

```text
src/
  app/
  routes/
  layouts/
  components/
  components/showcase/
  styles/
  mock/
  types/
  test/
```

## 5. Static Data Strategy

M0.1 uses local Mock modules only.

Rules:

- No network calls.
- No API client.
- No database.
- No auth provider.
- No upload runtime.

## 6. Screenshot Strategy

Playwright should later cover:

- Desktop viewport.
- Mobile viewport.
- Loading state.
- Empty state.
- Error state.
- Permission denied state.
- Long text state.
- Success feedback state.

