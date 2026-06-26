# Podcast Hub Product Design

## 1. Product Positioning

Podcast Hub is a content ingestion, program management, authorization control, and RSS publishing platform.

It is designed for operators who own, license, or are otherwise authorized to distribute audio content through podcast feeds. It is not a general web scraper, account sharing tool, DRM bypass tool, paywall bypass tool, CAPTCHA bypass tool, or unauthorized redistribution system.

The platform provides:

- Controlled content ingestion from manual upload, platform-native RSS import, and approved Connectors.
- Program and episode management with review before publication.
- Connector package governance and execution tracking.
- User-level authorization and personal RSS delivery.
- Full traceability for imports, authentication states, review decisions, and feed publication.

## 2. Product Goals

### 2.1 Administrator Goals

Administrators can:

- Create and manage podcast programs.
- Configure one or more content sources for a program.
- Choose ingestion modes:
  - Manual upload.
  - Native RSS import.
  - Automatic Connector.
  - Scan-triggered Connector.
- Upload Connector ZIP packages that conform to the platform specification.
- Configure Connector versions, input parameters, trigger modes, authorization modes, and scheduling.
- View import jobs, logs, errors, authentication status, and manual action items.
- Review imported episodes before publication.
- Publish episodes to private, selected-user, or public RSS feeds.
- Audit who changed program, source, job, review, and publication settings.

### 2.2 User Goals

Users can:

- Register with email after Turnstile and email verification.
- Log in and maintain a secure session.
- Browse programs they are authorized to access.
- Create personal program collections.
- Obtain personal RSS feed URLs.
- Subscribe from external podcast clients.
- Regenerate personal RSS URLs if a URL is leaked or no longer trusted.
- Reset their password through a verified recovery flow.
- View and revoke login sessions in a later version.

## 3. Non-Goals

The first phase does not include:

- Business code implementation.
- A general-purpose scraping marketplace.
- Connector execution on the host machine.
- Any authentication bypass, DRM bypass, paywall bypass, CAPTCHA bypass, or platform restriction bypass.
- Connector-managed scheduling.
- Connector-managed RSS generation.
- Connector direct writes to database, Redis, object storage control plane, or production secrets.
- Multi-runtime Connector execution. Phase 1 supports Python Connectors only.
- Automated publication without review unless a future policy explicitly enables it.

Frozen M0.1 non-goals also include real authentication, email sending, Cloudflare Turnstile integration, password hashing, database, real APIs, audio upload, RSS XML generation, Connector ZIP upload, Python Connector execution, QR scanning, Docker, and production deployment.

## 4. Roles

### 4.0 Account Roles

System account roles:

- `user`: default role for public registration.
- `admin`: privileged role for platform administration.

Rules:

- Public administrator registration is not allowed.
- Newly registered accounts are always `user`.
- `admin` can be created only during trusted initialization or granted by an existing authorized administrator.
- Role changes must be audited.

### 4.1 System Owner

An `admin` responsibility profile that owns platform-level policies, runtime limits, connector approval rules, and security defaults.

### 4.2 Administrator

Manages programs, sources, connectors, import jobs, reviews, publishing, users, and access rules.

### 4.3 Reviewer

An `admin` responsibility profile that reviews staged imported episodes and decides whether they can be published, rejected, or sent back for revision.

### 4.4 Operator

An `admin` responsibility profile that monitors jobs, handles manual todos, follows authentication prompts, and retries failed imports.

### 4.5 End User

Consumes authorized programs through web browsing, personal collections, and RSS subscriptions.

## 4.6 User Account Status

Supported statuses:

- `pending_verification`: user has registered but has not completed email verification.
- `active`: user can log in and use authorized features.
- `suspended`: user cannot log in or use RSS access.
- `deleted`: user is logically deleted or anonymized according to retention policy.

Only `active` users can access personal RSS feeds.

## 5. Core Domain Objects

### 5.1 Program

A podcast, audio course, private collection, or long-running audio content product.

Key properties:

- Title, description, language, category, cover image.
- Ownership and rights notes.
- Visibility policy.
- Default review policy.
- RSS publication settings.

### 5.2 Source

A content origin attached to a Program.

Examples:

- Manual upload source.
- Native RSS source.
- Connector-backed external platform source.
- QR-authenticated source.

Key properties:

- Source type.
- Source display name.
- Connector binding, if applicable.
- Input parameters.
- `ingestion_type`: `native_rss`, `connector`, or `manual_upload`.
- `trigger_type`: `manual` or `scheduled`.
- `auth_mode`: `none`, `reusable_session`, or `qr_each_run`.
- `execution_mode`: `unattended` or `interactive`.
- Schedule policy.
- Rights notes.
- Active or disabled state.

### 5.3 Connector

A versioned ZIP package that imports content from a Source into standardized platform staging outputs.

Connectors are not programs. A Program can use multiple Sources, and a Source can reference one approved Connector version.

### 5.4 Import Job

One execution attempt of a Connector or manual import workflow.

Each Import Job has:

- Immutable job ID.
- Source and Program references.
- Connector package version, if applicable.
- Trigger reason.
- Input snapshot.
- Status timeline.
- JSON Lines events.
- Output artifact references.
- Error classification, if failed.
- Human action references, if blocked.

### 5.5 Imported Episode

An episode staged by a job before review and publication.

Imported Episodes are not automatically public. They require review unless a later phase adds explicit trusted automation rules.

### 5.6 Review

Approval, rejection, revision request, or hold decision for a staged episode.

### 5.7 RSS Publication

A feed output configuration for approved episodes.

Publication scopes:

- Private: only the program owner or selected internal users.
- Selected users: explicit user or group authorization.
- Public: available without user-specific authorization.

## 6. Content Ingestion Modes

### 6.1 Manual Upload

Used when an administrator uploads audio and metadata directly.

Rules:

- Uses `ingestion_type: manual_upload`.
- Cannot be scheduled.
- Does not use Connector execution.
- Creates content through a human import workflow.
- Requires the same review and publication pipeline as automated imports.

### 6.2 Native RSS

Used when content is available through an authorized RSS source.

Rules:

- Uses `ingestion_type: native_rss`.
- Is a platform built-in Importer, not an administrator-uploaded Connector ZIP.
- Platform fetches and normalizes RSS metadata in a controlled importer.
- May use `trigger_type: manual` or `trigger_type: scheduled`.
- Uses `auth_mode: none` and `execution_mode: unattended`.
- Source URL, ownership, and redistribution rights must be recorded.
- Imported episodes still enter review.

### 6.3 Automatic Connector

Used when a source requires custom but authorized integration logic.

Rules:

- Uses `ingestion_type: connector`.
- First phase supports Python Connector ZIP packages only.
- Connector receives one job input and exits after one import attempt.
- Connector does not schedule itself.
- Connector produces standardized episode JSON outputs and JSON Lines events.
- `auth_mode: none` can run manually or scheduled as `execution_mode: unattended`.
- `auth_mode: reusable_session` can run manually or scheduled only while session is valid as `execution_mode: unattended`.

### 6.4 Scan-Triggered Connector

Used when a source requires interactive QR authentication for each import.

Rules:

- Uses `ingestion_type: connector`, `auth_mode: qr_each_run`, and `execution_mode: interactive`.
- Can use only `trigger_type: manual`.
- Cannot have a periodic schedule.
- Each run creates a human todo for QR scan.
- Job must support QR expiration, interaction timeout, cancellation, and failure states.

## 7. Authorization and Publishing

Authorization depends on both account authentication and content entitlement.

Account authentication answers:

- Who is the user?
- Is the account active?
- What role does the user have?

Content authorization answers:

- Which Programs can this user access?
- Which RSS feeds can this user use?
- Which admin operations, if any, can this user perform?

### 7.1 Program Authorization

Program access can be granted by:

- Explicit user allowlist.
- Group membership.
- Organization-level entitlement.
- Public visibility.

### 7.2 RSS Authorization

RSS URLs can be:

- Public feed URLs.
- Program-private feed URLs.
- User-specific feed URLs with revocable tokens.
- Collection RSS URLs generated from a user's authorized programs.

User-specific feed URLs must never grant access beyond the user's current authorization. If authorization changes, the feed output must reflect the new permission state.

Every RSS request must validate RSS Token, User status, current Program or Collection access, and publication state. Token revocation or permission revocation must take effect immediately, and caches must not bypass authorization checks.

### 7.3 Publication Gate

An episode can be published only when:

- The Program is active.
- The Source is active.
- Rights policy allows publication.
- The episode review state is approved.
- The target RSS scope is configured and active.
- No active takedown, hold, or compliance block exists.

Published RSS must use reviewed and approved content only. External source URLs are provenance by default and are not automatically used as formal RSS enclosures.

## 8. Success Metrics

Product quality should be measured by:

- Import job success rate by source type and connector version.
- Time from import completion to review decision.
- Number of jobs blocked by authentication or manual todo.
- Number of episodes rejected for metadata or rights issues.
- RSS feed availability and latency.
- User subscription activation rate.
- Connector version rollback and failure frequency.
- Audit completeness for sensitive actions.

## 9. Development Milestones

### Milestone 0.0: Documentation and Decision Freeze

- Product, UX, architecture, connector protocol, rights, and decision documents.
- No business code.
- Authentication architecture and API contract.
- Visual direction and visual acceptance policy.
- Git baseline and frozen M0 decisions.

#### M0.1: Frontend Foundation and Representative Static Pages

- Frontend engineering skeleton.
- Design tokens.
- Base components.
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

#### M0.2: Complete Core Static Experience

- Complete all core static pages.
- Use complete Mock data.
- Cover long text, no permission, task failure, waiting QR, pending review, and empty-list boundary states.
- Cover all core navigation paths for user app and admin app.
- Do not connect a real backend.

#### M0.3: Visual Acceptance

- Playwright desktop and mobile screenshots.
- Page consistency checks.
- Visual defect fixes.
- Information hierarchy checks.
- Basic accessibility checks.
- Visual acceptance report.
- Do not enter M1 until M0.3 is accepted.

### Milestone 1: Core Domain and Admin Skeleton

- Program, Source, Connector Package, Import Job, Episode, Review, Publication domain model.
- User authentication domain model.
- Admin navigation and authenticated shell.
- Role and permission model draft.
- Audit event schema.

### Milestone 2: Manual Import and Review

- Manual audio and metadata upload flow.
- Staged episode review queue.
- Private RSS publication for approved episodes.
- Basic user-specific RSS URL generation and revocation.

### Milestone 3: Connector Registry and Validation

- Connector ZIP upload.
- Manifest validation.
- Package versioning and approval workflow.
- Static safety checks and sample output validation.

### Milestone 4: Isolated Job Execution

- Isolated Connector runtime.
- Standard job input.
- JSON Lines event collection.
- Episode output ingestion.
- Job timeout, resource limits, and retry policy.

### Milestone 5: Connector Auth and Scheduling

- `none`, `reusable_session`, and `qr_each_run` state handling.
- Scheduler controlled by platform.
- Human todo queue.
- Session expiration and reauthorization workflows.

### Milestone 6: Multi-Source Expansion

- Program-level multi-source merge.
- Conflict detection and duplicate handling.
- Connector version rollout and rollback.
- Advanced analytics and operational dashboards.
