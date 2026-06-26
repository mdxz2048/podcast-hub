# Podcast Hub — Project Rules

## Product Goal

Podcast Hub is a content ingestion, management, authorization, and RSS publishing platform.

The system allows administrators to manage podcast programs, import episodes through approved connectors or manual upload, review imported content, and publish authorized RSS feeds to permitted users.

The system is not a general web scraper and must not bypass paywalls, DRM, authentication controls, CAPTCHA, platform restrictions, or access controls.

Only content that the operator owns, is licensed to distribute, or is otherwise authorized to publish may be released through RSS.

---

## Core Domain Model

Program
  -> Source
  -> Import Job
  -> Imported Episode
  -> Review
  -> RSS Publication

Connector is attached through Source only when `ingestion_type` is `connector`.

Definitions:

- Program: A podcast, audio course, or audio content collection.
- Source: A specific platform or origin where the program content comes from.
- Connector: A versioned package that imports content from a source.
- Import Job: One execution attempt of a connector or manual import workflow.
- Imported Episode: An episode staged before review.
- Review: Approval, rejection, or revision of imported content.
- RSS Publication: A public, private, or authorized RSS feed.

Do not bind Program directly to a script file.
Do not let a connector write directly into production database tables.
Do not let a connector generate final RSS XML directly.

Frozen Source dimensions:

- ingestion_type: native_rss, connector, manual_upload
- trigger_type: manual, scheduled
- auth_mode: none, reusable_session, qr_each_run
- execution_mode: unattended, interactive

The frozen scheduled trigger value is `scheduled`.
Do not mix manual upload, scheduled trigger, and interactive auth into one trigger field.

---

## Connector Rules

Every connector must be a versioned package.

A connector package must include:

- manifest.yaml
- entrypoint source code
- dependency lock file
- README.md
- tests
- sample output fixture

Supported runtime in phase 1:
- Python only

Every connector execution must:

1. Receive a standard job JSON input.
2. Write files only into its assigned temporary output directory.
3. Emit machine-readable JSON Lines events to stdout.
4. Produce standardized episode metadata JSON files.
5. Exit after one import attempt.
6. Never run as a permanent daemon.
7. Never schedule itself internally.
8. Never access PostgreSQL, Redis, Docker socket, host root filesystem, or production secrets.

Connector execution rules:

- connector + auth_mode none may use manual or scheduled trigger_type and must be unattended.
- connector + auth_mode reusable_session may use manual or scheduled trigger_type only while the session is valid and must be unattended.
- connector + auth_mode qr_each_run may use manual trigger_type only and must be interactive.
- qr_each_run must include QR expiration, interaction timeout, cancellation, and failure states.
- manual_upload does not execute a Connector and cannot be scheduled.
- native_rss is a platform built-in Importer, not an uploaded Connector ZIP.
- native_rss may use manual or scheduled trigger_type, auth_mode none, and execution_mode unattended.

---

## Technology Stack Rules

Frozen M0 frontend stack:

- React
- TypeScript
- Vite
- pnpm
- React Router
- Tailwind CSS
- CSS Variables as the only source of design tokens
- lucide-react
- Playwright for E2E and screenshot acceptance

Storybook:

- Do not introduce Storybook in M0.
- Use an internal component showcase page or development route in M0.
- Re-evaluate Storybook only after the component system stabilizes.

Frozen long-term backend architecture:

- Go as the platform main backend.
- PostgreSQL as the main database.
- Redis for cache, rate limiting, and task queue.
- S3-compatible object storage for authorized media and task artifacts.
- Python only for Connector SDK and Connector execution environments.
- Do not use Python FastAPI as the platform main API.

M0.0 and M0.1 must not create backend services, databases, APIs, Docker, real auth, email, Turnstile, RSS, upload, Connector execution, or QR scanning.

---

## Security Rules

Never expose or commit:

- SSH private keys
- server root credentials
- database passwords
- Redis passwords
- object storage root keys
- platform account passwords
- cookies
- long-lived access tokens
- QR login session data

Never print secrets, cookies, tokens, passwords, or authorization headers in logs.

Uploaded connector ZIP files must never execute directly on the host machine.

Connector execution must eventually run in an isolated environment with:

- non-root user
- temporary workspace
- CPU limit
- memory limit
- disk limit
- execution timeout
- restricted network access
- no host filesystem mount
- no Docker socket
- no database access

---

## User Authentication Rules

Podcast Hub supports public registration for normal users only.

Supported user roles:

- user
- admin

Rules:

- New public registrations always create user accounts.
- Public administrator registration is not allowed.
- admin accounts can be created only during trusted initialization or granted by an existing authorized administrator.
- Account status must be one of pending_verification, active, suspended, or deleted.
- Do not use Pending invite or Disabled User as User statuses.
- Invitations are not implemented in M0 and must not introduce Invitation APIs or User statuses.
- System Owner, Operator, and Reviewer are admin responsibility labels or permission profiles, not account roles.
- Email verification is required before account activation.
- Registration must use email, password, Cloudflare Turnstile, and email verification code.
- Turnstile tokens must be verified by the backend through Cloudflare Siteverify.
- Turnstile is not the only security control; server-side rate limiting, audit logging, and risk controls are still required.
- Login failures must not reveal whether the email exists or whether the password was wrong.
- Verification codes and password reset proofs must be single-use, short-lived, attempt-limited, rate-limited, and never stored in plaintext.
- Passwords must be stored with Argon2id.
- Plaintext passwords, reversible encrypted passwords, and password logging are prohibited.
- Sessions must use HttpOnly, Secure, SameSite cookies.
- Login, logout, password change, and password reset must have explicit session state behavior.
- Authentication events must be audited with secrets redacted.

---

## UI Rules

The UI should feel like one coherent product.

Do not create unrelated visual styles across pages.

The visual direction is:

- restrained
- content-led
- editorial
- trustworthy audio content library

Avoid:

- generic enterprise admin look
- excessive gradients
- excessive glassmorphism
- different visual styles per page
- excessive card stacking
- meaningless icon decoration
- template-site or AI-generated collage feeling

User-facing pages and admin pages must share the same design tokens, typography scale, spacing system, semantic colors, button rules, and status expression.

Dark mode:

- Phase 1 explicitly does not support dark mode.
- Tokens and component structure must allow future `data-theme="dark"` extension.
- Phase 1 pages must not become unreadable because of browser or OS automatic dark mode behavior.

Use reusable components for:

- Button
- Input
- SearchBar
- Select
- Badge
- ProgramCard
- EpisodeRow
- ConnectorCard
- ImportJobCard
- EmptyState
- LoadingState
- ErrorState
- Drawer
- Dialog
- Toast

Every async page or action must include:

- loading state
- empty state
- error state
- success feedback
- mobile layout

All admin workflows must clearly show:

- current task status
- next required action
- authorization state
- execution logs
- review state
- publishing state

Every user-visible feature must provide visual acceptance evidence before completion:

- desktop screenshot
- mobile screenshot
- empty state screenshot
- loading state screenshot
- error state screenshot
- long text screenshot
- permission denied screenshot
- async success feedback screenshot
- explicit phase-1 dark-mode-unsupported note and readability check
- mock data verification
- boundary data verification

---

## Development Workflow

Before implementing a non-trivial feature:

1. Read relevant docs in /docs.
2. Provide a concise implementation plan.
3. List files that will change.
4. Identify risks and non-goals.
5. Implement only the agreed scope.
6. Add or update tests.
7. Run formatting, linting, and tests.
8. Verify desktop and mobile UI.
9. Provide screenshots for changed user-facing pages.
10. Update docs if public contracts or domain rules changed.

Do not make broad refactors during feature work unless explicitly requested.

Do not replace working architecture with a new stack without approval.

M0 must not be treated as engineering skeleton only.

M0 is split into:

- M0.0: documentation, Git, architecture, and decision freeze.
- M0.1: frontend skeleton, design tokens, base components, layouts, responsive rules, Mock data infrastructure, representative high-fidelity static pages, core states, and Playwright screenshot foundation.
- M0.2: complete core static pages, full Mock data, boundary states, and all core user/admin navigation paths without a real backend.
- M0.3: Playwright desktop/mobile screenshots, consistency checks, visual fixes, information hierarchy checks, accessibility basics, and visual acceptance report.

Do not enter M1 until M0.3 is accepted.

---

## Definition of Done

A task is done only when:

1. Feature scope is complete.
2. Tests pass.
3. Error and empty states are implemented.
4. Mobile layout is checked.
5. Relevant docs are updated.
6. No secrets are added to the repository.
7. No unauthorized content publishing path is introduced.
8. A concise change summary is provided.
