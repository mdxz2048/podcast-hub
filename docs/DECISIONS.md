# Decisions and Open Questions

This document records product and architecture decisions for the current design phase.

## 1. Current Phase

Current phase:

- Product, interaction, and engineering architecture design only.

Explicitly out of scope:

- Business code.
- Connector implementation code.
- Download, crawling, authentication bypass, DRM bypass, CAPTCHA bypass, paywall bypass, or paid-content bypass code.
- Real server accounts, root passwords, SSH private keys, cookies, platform passwords, or production tokens.

## 2. Key Decisions

### ADR-001: Program Does Not Bind Directly to Script Files

Decision:

- Program binds to Sources.
- Source may bind to an approved Connector version.

Reason:

- Keeps product model stable as multiple Sources are added.
- Allows one Program to combine manual upload, native RSS, and Connector-based ingestion.
- Avoids coupling content identity to execution details.

Consequence:

- Source configuration becomes the operational unit for imports, auth, and scheduling.

### ADR-002: Connector Is a Versioned ZIP Package

Decision:

- Every Connector is uploaded as a ZIP package with manifest, entrypoint, dependency lock, README, tests, and sample output.

Reason:

- Enables validation, approval, audit, rollback, and reproducible execution.

Consequence:

- Loose script uploads are not supported.

### ADR-003: Phase 1 Supports Python Connectors Only

Decision:

- `runtime: python` is the only supported Connector runtime in phase 1.

Reason:

- Reduces execution sandbox complexity.
- Keeps packaging, validation, and dependency behavior focused.

Consequence:

- Manifest keeps runtime fields so future runtimes can be added without redesigning the job protocol.

### ADR-004: Connector Executes One Job and Exits

Decision:

- A Connector execution performs exactly one Import Job attempt and exits.

Reason:

- Makes jobs traceable and bounded.
- Prevents untracked background work.
- Allows platform-level retries, cancellation, and scheduling.

Consequence:

- Long-running daemons and internal Connector schedulers are prohibited.

### ADR-005: Platform Owns Scheduling

Decision:

- The platform scheduler creates jobs.
- Connectors cannot create timers or future runs.

Reason:

- Centralizes policy enforcement for auth mode, source state, program state, and connector version state.

Consequence:

- `qr_each_run` cannot have scheduled Connector jobs.
- `manual_upload` cannot have scheduled jobs.

### ADR-006: Platform Owns Authentication State

Decision:

- The platform tracks auth mode, session validity, QR todo state, and auth-related scheduling blocks.

Reason:

- Keeps credentials and session lifecycle out of Connector package ownership.
- Makes human intervention visible and auditable.

Consequence:

- Connectors must not store or print credentials, cookies, tokens, or raw session material.

### ADR-007: JSON Input, JSON Lines Events, Episode JSON Output

Decision:

- Platform and Connector communicate through:
  - One job input JSON.
  - JSON Lines events to stdout.
  - Standardized episode JSON output files.

Reason:

- Works across runtimes.
- Easy to validate and archive.
- Supports streaming logs and structured diagnostics.

Consequence:

- Connectors cannot write directly to the database or generate final RSS XML.

### ADR-008: Imported Episodes Require Review Before Publication

Decision:

- Imported Episodes enter review before publication.

Reason:

- Reduces risk from metadata errors, duplicate imports, and rights ambiguity.

Consequence:

- Publication depends on review approval and rights policy.

### ADR-009: RSS Access Is Authorization-Aware

Decision:

- RSS generation and serving must enforce current Program and user authorization.

Reason:

- Personal RSS URLs can leak, be revoked, or become stale after access changes.

Consequence:

- RSS tokens are revocable and should never grant broader access than current permissions.

### ADR-010: Last Valid Feed Should Survive Generation Errors

Decision:

- Feed generation failure should not delete the previous valid feed.

Reason:

- External podcast clients rely on RSS stability.

Consequence:

- Publication service must distinguish active feed content from failed generation attempts.

### ADR-011: Connector Versions Are Immutable After Approval

Decision:

- Approved Connector package artifacts and manifest contracts are immutable.

Reason:

- Enables repeatable job auditing and reliable rollback.

Consequence:

- Fixes require a new version.

### ADR-012: Audit Records Are Required for Sensitive Actions

Decision:

- Sensitive actions must create append-only audit records.

Reason:

- Content publishing, authorization, and Connector execution need traceability.

Consequence:

- Audit redaction is mandatory for secrets and session material.

### ADR-013: Public Registration Creates User Accounts Only

Decision:

- Public registration is allowed for normal users.
- Newly registered accounts always receive the `user` role.

Reason:

- Users need self-service access to authorized Programs, collections, and RSS URLs.
- Administrator access must remain controlled.

Consequence:

- Registration UI must not imply admin registration.
- Admin APIs must reject public role escalation.

### ADR-014: Admin Accounts Are Granted Only Through Trusted Paths

Decision:

- `admin` accounts can be created only during trusted initialization or granted by an existing authorized administrator.

Reason:

- Admin users can affect content ingestion, publication, rights, Connectors, and user access.

Consequence:

- Role changes require audit events.
- There is no public admin signup flow.

### ADR-015: Email Verification Gates Account Activation

Decision:

- New accounts start as `pending_verification`.
- Successful email verification changes status to `active`.

Reason:

- Email verification reduces account abuse and supports account recovery.

Consequence:

- Pending users cannot log in, access RSS, or use authorized content features.

### ADR-016: Turnstile Is Required for High-Risk Entry Points But Is Not Sufficient Alone

Decision:

- Cloudflare Turnstile is used for registration, resending verification codes, password reset, and high-risk login.
- Backend must verify Turnstile tokens with Cloudflare Siteverify.
- Server-side rate limiting, audit, and risk controls are still required.

Reason:

- Bot resistance is useful but cannot replace server-side controls.

Consequence:

- Expired, duplicated, or failed Turnstile tokens have explicit error states.

### ADR-017: Passwords Use Argon2id

Decision:

- Passwords must be stored using Argon2id.
- Plaintext passwords, reversible encrypted passwords, and password logging are prohibited.

Reason:

- Password storage must be resilient against database disclosure.

Consequence:

- Authentication implementation must choose safe Argon2id parameters before coding begins.

### ADR-018: Sessions Use Secure Cookies and Explicit Lifecycle Rules

Decision:

- Login sessions use `HttpOnly`, `Secure`, `SameSite` cookies.
- Login, logout, password change, password reset, suspension, and deletion must define session behavior.

Reason:

- Session security and revocation are core to personal RSS access and admin safety.

Consequence:

- Later implementation must support session invalidation and future user-visible session revocation.

### ADR-019: Verification and Reset Proofs Are One-Time, Short-Lived, and Hashed

Decision:

- Email verification codes and password reset proofs are single-use, short-lived, attempt-limited, rate-limited, and not stored in plaintext.
- Resending a verification code invalidates the previous code for that purpose.

Reason:

- Registration and password reset are high-risk abuse paths.

Consequence:

- Logs and audit records must never contain full codes or raw reset proofs.

### ADR-020: One Shared Visual System for User App and Admin App

Decision:

- User-facing pages and admin pages share design tokens, typography, spacing, semantic colors, button rules, and state expression.

Reason:

- Podcast Hub should feel like one product, not a set of unrelated screens.

Consequence:

- New pages must reuse the shared component and token system.

### ADR-021: Visual Direction Is Restrained, Content-Led, Editorial, and Trustworthy

Decision:

- The visual direction is "restrained, content-led, editorial, trustworthy audio content library."

Reason:

- The product manages rights-sensitive audio content and should build confidence while keeping content central.

Consequence:

- Avoid generic enterprise admin look, excessive gradients, excessive glassmorphism, unrelated page styles, excessive card stacking, meaningless icons, and template-site collage.

### ADR-022: M0 Includes Visual Foundation, Static Prototype, and Visual Acceptance

Decision:

- M0 is split into M0.1, M0.2, and M0.3.
- M1 cannot begin until M0.3 is accepted.

Reason:

- Engineering skeleton without visual and interaction governance would create expensive rework and inconsistent UX.

Consequence:

- M0 must produce tokens, components, layouts, high-fidelity static prototypes, screenshots, component inventory, and design consistency checks.

### ADR-023: Frozen Frontend Stack for M0

Decision:

- Frontend uses React, TypeScript, Vite, pnpm, React Router, Tailwind CSS, CSS Variables, lucide-react, and Playwright.
- CSS Variables are the only source of design tokens.
- Storybook is not introduced in M0.

Reason:

- M0 needs a focused, lightweight static UI foundation with screenshot acceptance and no extra component tooling overhead.

Consequence:

- M0 uses an internal component showcase page or development route.
- Storybook may be reconsidered after the component system stabilizes.

### ADR-024: Frozen Long-Term Backend Stack

Decision:

- Go is the platform main backend.
- PostgreSQL is the main database.
- Redis is used for cache, rate limiting, and task queue.
- S3-compatible object storage is used for authorized media and task artifacts.
- Python is used only for Connector SDK and Connector execution environments.
- Python FastAPI is not used as the platform main API.

Reason:

- The platform backend should be operationally predictable and separate from Python Connector execution.

Consequence:

- M0.1 does not create backend services.
- Future backend planning should assume Go service boundaries.

### ADR-025: Source Execution Uses Four Dimensions

Decision:

- Source execution is modeled with `ingestion_type`, `trigger_type`, `auth_mode`, and `execution_mode`.
- The frozen scheduled trigger value is `scheduled`.
- `manual_import` and `interactive_auth` are not trigger values.

Reason:

- Mixing ingestion type, trigger, auth, and execution behavior caused ambiguity across Connector, manual upload, and native RSS workflows.

Consequence:

- Connector, Native RSS, Manual Upload, Job Protocol, UX, and scheduling docs use the same four-dimensional model.

### ADR-026: Native RSS Is a Built-In Importer

Decision:

- Native RSS is a platform built-in Importer, not an administrator-uploaded Connector ZIP.
- It uses `ingestion_type: native_rss`, `auth_mode: none`, `execution_mode: unattended`, and `trigger_type: manual` or `scheduled`.

Reason:

- Native RSS should follow the same Source/Job/Review/Publication pipeline without requiring uploaded Connector governance.

Consequence:

- Native RSS has its own specification and does not execute Connector packages.

### ADR-027: RSS Authorization Is Real-Time

Decision:

- Every RSS request validates RSS Token, User status, current Program or Collection access, and Program publication state.
- Cache must not bypass authorization checks.

Reason:

- RSS URLs can leak and entitlements can change after subscription.

Consequence:

- Token revocation or permission revocation must take effect immediately.

### ADR-028: Storage Baseline Is Controlled Object Storage and Isolated Staging

Decision:

- Approved and published media will use controlled S3-compatible object storage.
- Connector outputs first enter isolated staging.
- Failed job raw media artifacts are deleted immediately by default.
- Sanitized failed job logs and metadata are retained for 30 days by default.

Reason:

- Unreviewed and failed artifacts should not become publication paths or long-lived sensitive storage.

Consequence:

- External source URLs are provenance by default and are not formal RSS enclosures without explicit policy.

## 3. Architecture Risks

### 3.1 Rights Verification

Risk:

- The platform can record rights notes, but cannot automatically prove the operator's legal authorization.

Mitigation:

- Require Program and Source rights notes before publication.
- Add rights hold and takedown workflows.
- Audit publication scope changes.

### 3.2 Connector Safety

Risk:

- Uploaded packages may contain unsafe or policy-violating behavior.

Mitigation:

- Validate ZIP structure and manifest.
- Require human approval.
- Run only in isolated environments.
- Restrict filesystem, network, secrets, resources, and timeout.

### 3.3 Session Handling

Risk:

- Reusable sessions may contain sensitive material.

Mitigation:

- Use opaque session references.
- Redact logs and audit snapshots.
- Keep session lifecycle platform-owned.
- Define secure storage in a later engineering phase.

### 3.4 QR Workflow Reliability

Risk:

- QR authentication may expire or require operator presence.

Mitigation:

- Represent QR as a manual todo with expiration.
- Block scheduling for `qr_each_run`.
- Make retry create a new job and QR state.

### 3.5 Duplicate Episodes Across Sources

Risk:

- Multi-source Programs may stage duplicate episodes.

Mitigation:

- Store source episode ID, GUID, canonical URL, title, publication time, and connector provenance.
- Flag uncertain duplicates for review.

### 3.6 RSS URL Leakage

Risk:

- Personal RSS URLs may be copied outside the intended user.

Mitigation:

- Allow token regeneration and revocation.
- Enforce current authorization on feed request or generation.
- Avoid showing full tokens in admin tables.

### 3.7 Feed Client Caching

Risk:

- External podcast clients may cache removed items.

Mitigation:

- Treat takedown as best-effort for external clients.
- Remove affected items from future feeds.
- Keep audit trail and operator notices.

## 4. Open Questions

### 4.1 Storage Policy

Status:

- Baseline storage policy is frozen in `STORAGE_POLICY.md`.

Remaining non-blocking questions:

- Exact object key naming.
- Exact lifecycle rules beyond the default 30-day sanitized log retention.

### 4.2 Runner Isolation

Questions:

- Which isolation technology will be used for Connector execution?
- How will network allowlists be enforced?
- What default CPU, memory, disk, and timeout limits should phase 1 use?

### 4.3 Secret Management

Questions:

- Where will reusable session material be stored?
- How will session references be scoped to Source and runner?
- What rotation policy applies to encryption keys?

### 4.4 Review Automation

Questions:

- Should trusted Sources later allow auto-approval for low-risk metadata-only updates?
- If yes, what evidence and safeguards are required?

### 4.5 Native RSS Import

Status:

- Native RSS is frozen as a platform built-in Importer in `NATIVE_RSS_SPEC.md`.

Remaining non-blocking questions:

- Exact feed normalization edge cases.
- Feed polling interval defaults.

### 4.6 User Entitlements

Questions:

- Are user permissions direct, group-based, organization-based, or integrated with an external billing system?
- How quickly must revocations propagate to RSS clients?

## 5. Milestones

### M0.0: Documentation and Decision Freeze

Deliverables:

- Product design.
- UX flows.
- Architecture design.
- Connector specification.
- Job protocol.
- User authentication contract.
- Authentication and scheduling policy.
- Rights policy.
- Design system.
- Visual direction.
- Visual acceptance policy.
- Decision record.
- Git baseline.
- Frozen technology stack.
- Frozen Connector execution model.
- Frozen RSS and storage baseline.

### M0.1: Frontend Foundation and Representative Static Pages

Deliverables:

- Frontend engineering skeleton.
- Design tokens.
- Base components.
- Public layout, user layout, and admin layout.
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

### M0.2: Complete Core Static Experience

Deliverables:

- All core static pages.
- Complete Mock data.
- Long text, no permission, task failure, waiting QR, pending review, and empty-list boundary states.
- All user app and admin app core navigation paths.
- No real backend.

### M0.3: Visual Acceptance

Deliverables:

- Playwright desktop and mobile screenshots.
- Page consistency checks.
- Visual defect fixes.
- Information hierarchy checks.
- Basic accessibility checks.
- Visual acceptance report.

Gate:

- Do not enter M1 until M0.3 is accepted.

### M1: Domain Foundation

Deliverables:

- Domain model for Programs, Sources, Connectors, Jobs, Episodes, Reviews, Publications, Users, and Audit.
- Authentication domain model for User, UserCredential, EmailVerification, PasswordReset, UserSession, AuthAuditLog, UserRole, and UserStatus.
- Permission model.
- Authenticated app shell.
- Admin navigation skeleton.
- Read-only status pages.

### M2: Manual Import Path

Deliverables:

- Manual import Source.
- File and metadata staging.
- Review queue.
- Private RSS publication.

### M3: Connector Registry

Deliverables:

- Connector ZIP upload.
- Manifest validation.
- Version approval.
- Deprecation and revocation.
- Usage impact preview.

### M4: Isolated Job Runner

Deliverables:

- Job input creation.
- Runner isolation.
- JSON Lines event collection.
- Episode output validation.
- Retry and timeout behavior.

### M5: Auth and Scheduling

Deliverables:

- Reusable session state tracking.
- QR each run todo flow.
- Scheduler policy enforcement.
- Skipped run auditing.

### M6: RSS and Access Expansion

Deliverables:

- Selected-user RSS.
- User collections.
- RSS token regeneration and revocation.
- Program access management.

### M7: Multi-Source Operations

Deliverables:

- Duplicate detection.
- Source merge review.
- Connector health dashboard.
- Version rollout and rollback workflow.

## 6. Immediate Non-Goals

The next implementation phase should not start with:

- Full marketplace mechanics.
- Multi-runtime Connectors.
- Automated rights verification.
- Public discovery site.
- Payment and billing.
- Advanced media transcoding pipeline.
- Connector self-updating.

## 7. Required Design Invariants

These should remain true unless a future ADR changes them:

- Program does not bind directly to scripts.
- Source is the operational unit for ingestion.
- Connector versions are immutable after approval.
- Connector execution is isolated and one-shot.
- Scheduling belongs to the platform.
- Authentication state belongs to the platform.
- Source execution uses `ingestion_type`, `trigger_type`, `auth_mode`, and `execution_mode`.
- Native RSS is a built-in Importer, not a Connector ZIP.
- Review gates publication.
- RSS access reflects current authorization on every request.
- Cache must not bypass RSS authorization.
- Secrets are never printed, committed, or exposed in logs.
- Public registration creates only `user` accounts.
- `admin` is granted only through trusted initialization or existing admin action.
- Email verification gates account activation.
- Passwords use Argon2id.
- User app and admin app share one visual system.
- Phase 1 does not support dark mode but must preserve future `data-theme="dark"` extensibility.
- M1 cannot start before M0.3 visual acceptance.
