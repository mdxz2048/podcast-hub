# Podcast Hub UX Flows

## 1. UX Principles

Podcast Hub is an operational product. The interface should be dense enough for repeated admin work, but calm enough that job states, authorization states, and review requirements are easy to understand.

Every async workflow must include:

- Loading state.
- Empty state.
- Error state.
- Success feedback.
- Mobile layout.

M0.3 interaction baseline:

- Shared controls keep visible keyboard focus.
- Dialog and Drawer support Escape close and focus return.
- Dangerous operations use explicit confirmation dialogs.
- Long text and long email boundaries are verified on desktop and mobile.

Every admin workflow must show:

- Current status.
- Next required action.
- Authorization state.
- Import job logs.
- Review state.
- Publishing state.

## 2. Information Architecture

### 2.1 Admin Navigation

Primary admin sections:

- Overview
- Programs
- Sources
- Connector Registry
- Import Jobs
- Review Queue
- Publications
- Users and Access
- Audit Log
- Settings

### 2.2 User Navigation

Primary user sections:

- Home
- Register
- Login
- Authorized Programs
- My Collections
- RSS Feeds
- Account

## 3. Admin Pages

### 3.1 Overview Dashboard

Goal:

- Help administrators understand system health and identify the next operational action.

Main areas:

- Job status summary.
- Authentication alerts.
- Review queue summary.
- Recently failed imports.
- Recently published episodes.
- Connector version health.
- Manual todos.

States:

- Loading dashboard metrics.
- Empty state when no programs exist.
- Partial error when one metric source fails.
- Degraded state when job runner or scheduler is unavailable.

Key buttons:

- Create Program.
- Upload Connector.
- View Failed Jobs.
- Open Review Queue.
- View Manual Todos.

Exceptions:

- If metrics are stale, show last successful refresh time.
- If scheduler is disabled, show a prominent operational warning.
- If there are blocked QR jobs, show the number and oldest waiting time.

### 3.2 Program List

Goal:

- Let admins find, compare, create, and manage Programs.

Main areas:

- Search and filters.
- Program table or card list.
- Status badges.
- Source count, episode count, review pending count, publication scope.
- Bulk actions for archive or disable, subject to permissions.

States:

- Loading programs.
- Empty state with Create Program action.
- No search results.
- Error state with retry.

Key buttons:

- Create Program.
- Import Program Metadata.
- Open Program.
- Disable Program.

Exceptions:

- A Program with active publications cannot be deleted; it can be disabled or archived.
- A Program with pending reviews should warn before archival.

### 3.3 Program Detail

Goal:

- Provide one place to understand a Program's identity, ingestion, review, and publication state.

Main areas:

- Program header: title, cover, status, visibility.
- Metadata panel.
- Sources tab.
- Episodes tab.
- Review tab.
- Publications tab.
- Rights notes.
- Activity timeline.

States:

- Draft.
- Active.
- Disabled.
- Archived.
- Rights hold.
- Publication paused.

Key buttons:

- Edit Metadata.
- Add Source.
- Run Import.
- Open Review Queue.
- Configure RSS.
- Disable Program.

Exceptions:

- Missing rights notes should block public publication.
- Disabled Program should prevent scheduled jobs from starting.
- Archived Program should be read-only except for audit and export.

M0.2B static behavior:

- Program detail summarizes Source count, latest Import Job state, pending Review count, publication state, and the next recommended action.
- Tabs/sections cover Overview, Sources, Episodes, Access and Publication, and Activity.
- Mock states cover normal, draft, authorization pending, paused, no permission, no Sources, no Episodes, long title, and long description.

### 3.4 Program Create and Edit

Goal:

- Capture enough metadata and rights context to safely ingest and publish a Program.

Main areas:

- Basic metadata.
- Cover artwork.
- Category and language.
- Ownership and rights notes.
- Default publication scope.
- Default review policy.

States:

- Draft form.
- Validation errors.
- Unsaved changes.
- Save success.

Key buttons:

- Save Draft.
- Save and Add Source.
- Cancel.

Exceptions:

- Public publication cannot be enabled without rights notes.
- Invalid cover dimensions should show a recoverable validation error.

### 3.5 Source List

Goal:

- Show all content sources across Programs and their operational status.

Main areas:

- Filters by Program, source type, auth mode, schedule state, connector version.
- Source table.
- Last job status.
- Next scheduled run.
- Authentication status.

States:

- Active.
- Suspended.
- Deleted.
- Auth required.
- Schedule paused.
- Connector deprecated.
- Error.

Key buttons:

- Add Source.
- Open Source.
- Run Now.
- Pause Schedule.
- Reauthorize.

Exceptions:

- `qr_each_run` sources do not show schedule controls.
- `manual_upload` sources show Upload action instead of Run Now.

### 3.6 Source Detail

Goal:

- Configure and operate a single Program Source.

Main areas:

- Source identity.
- Ingestion mode.
- Connector binding and version, when applicable.
- Input parameter snapshot.
- Authentication panel.
- Scheduling panel.
- Recent jobs.
- Source-level rights notes.

States:

- Ready.
- Disabled.
- Auth missing.
- Auth expired.
- Waiting for QR scan.
- Schedule active.
- Schedule paused.
- Connector version unavailable.

Key buttons:

- Edit Source.
- Change Connector Version.
- Run Import.
- Upload Manual Episode.
- Start Authentication.
- Pause Schedule.
- Resume Schedule.
- Disable Source.

Exceptions:

- Running a job with missing required input parameters is blocked.
- Running a `reusable_session` source with expired session creates a reauthorization todo.
- Running a `qr_each_run` source creates a QR scan todo and cannot be scheduled.
- `native_rss` is a built-in Importer, not a Connector ZIP.

### 3.7 Connector Registry

Goal:

- Govern approved Connector packages and versions.

Main areas:

- Connector list.
- Runtime and version filters.
- Approval state.
- Usage count.
- Latest job health.
- Deprecation status.

States:

- Uploaded.
- Validating.
- Validation failed.
- Pending approval.
- Approved.
- Deprecated.
- Revoked.

Key buttons:

- Upload Connector ZIP.
- Open Connector.
- Approve Version.
- Deprecate Version.
- Revoke Version.

Exceptions:

- A Connector version used by active Sources cannot be hard deleted.
- Revoked versions cannot start new jobs.
- Deprecated versions can keep existing Sources running only if policy allows.

### 3.8 Connector Upload

Goal:

- Safely upload and validate a Connector ZIP package.

Main areas:

- ZIP upload field.
- Validation result.
- Manifest preview.
- Required file checklist.
- Sample output validation.
- Security and policy warnings.

States:

- Waiting for file.
- Uploading.
- Validating.
- Validation passed.
- Validation failed.
- Pending approval.

Key buttons:

- Select ZIP.
- Validate.
- Submit for Approval.
- Cancel.

Exceptions:

- Missing `manifest.yaml`, dependency lock file, README, tests, or sample output blocks approval.
- Unsupported runtime blocks upload.
- Manifest schedule/auth mismatch blocks approval.

### 3.9 Connector Detail

Goal:

- Understand a Connector version's contract, status, usage, and operational history.

Main areas:

- Manifest summary.
- Version history.
- Required inputs.
- Supported trigger modes.
- Supported auth modes.
- Package validation log.
- Usage by Source.
- Recent jobs.

States:

- Approved.
- Deprecated.
- Revoked.
- Validation failed.
- In use.

Key buttons:

- Approve.
- Deprecate.
- Revoke.
- Download Package for Audit.
- View Jobs.

Exceptions:

- Revocation requires confirmation and impact preview.
- Approval requires passing validation and required documentation.

M0.2B static behavior:

- Connector list and detail distinguish platform Native RSS Importer, approved Python Connector, and manual import workflow.
- Connector registration wizard is static only and must not read, upload, unzip, validate, or execute real ZIP files.
- Manifest, network policy, resource limits, version history, bound Sources, and recent job result are shown from Mock data.

### 3.10 Import Jobs

Goal:

- Track all import attempts, failures, retries, outputs, and manual actions.

Main areas:

- Job filters.
- Job table.
- Status badge.
- Trigger reason.
- Source, Program, Connector version.
- Duration and output counts.
- Error classification.

States:

- Queued.
- Preparing.
- Waiting for authentication.
- Running.
- Processing output.
- Completed.
- Completed with warnings.
- Failed.
- Cancelled.
- Timed out.

Key buttons:

- Open Job.
- Retry.
- Cancel.
- View Logs.
- Open Review Items.

Exceptions:

- Failed jobs without retryable error do not show Retry.
- Jobs waiting for QR scan show Open Todo.
- Cancelled or timed out jobs remain auditable.

### 3.11 Import Job Detail

Goal:

- Give operators enough context to diagnose one job without exposing secrets.

Main areas:

- Job header and status timeline.
- Input snapshot with secrets redacted.
- JSON Lines event viewer.
- Output artifact list.
- Imported episode list.
- Error and retry panel.
- Manual todo panel.
- Audit events.

States:

- Live streaming logs.
- Completed log archive.
- Redacted sensitive fields.
- Output processing failed.

Key buttons:

- Retry Job.
- Cancel Job.
- Download Sanitized Logs.
- Open Source.
- Open Program.
- Open Review.

Exceptions:

- Logs must never display cookies, tokens, passwords, authorization headers, or QR session data.
- If output files are malformed, show a validation error linked to the specific artifact.

M0.2B static behavior:

- Import job detail shows Program, Source, Connector version, trigger type, auth mode, status timeline, sanitized logs, progress, output summary, failure reason, and next action.
- `waiting_auth` with `qr_each_run` displays a non-scannable QR placeholder and copy stating there is no real QR, Cookie, Token, or authentication data.
- Retry and cancel are Mock actions only; cancel uses confirmation Dialog.

### 3.12 Authentication and Manual Todos

Goal:

- Centralize human intervention required for blocked imports.

Main areas:

- Todo queue.
- Auth mode filter.
- QR scan panel.
- Reusable session expiration list.
- Assignment and due time.
- Related job/source links.

States:

- Open.
- Waiting for operator.
- In progress.
- Expired.
- Completed.
- Cancelled.

Key buttons:

- Start QR Scan.
- Mark Cannot Complete.
- Reassign.
- Cancel Job.
- Open Source.

Exceptions:

- Expired QR tasks cannot be reused; create a new job or retry.
- Manual todo completion must be auditable.

### 3.13 Review Queue

Goal:

- Let reviewers approve, reject, hold, or request revision for staged episodes.

Main areas:

- Review filters.
- Episode rows with metadata preview.
- Audio preview if available.
- Rights and source context.
- Duplicate warnings.
- Batch review controls.

States:

- Pending review.
- Approved.
- Rejected.
- Needs revision.
- On hold.
- Superseded.

Key buttons:

- Approve.
- Reject.
- Request Revision.
- Put on Hold.
- Open Episode.
- Compare Duplicates.

Exceptions:

- Episodes from rights-hold Programs cannot be approved.
- Missing required metadata blocks approval.
- Duplicate conflict requires explicit reviewer decision.

M0.2B static behavior:

- Review queue uses Drawer details to avoid unnecessary page jumps.
- Mock actions include approve, reject, request more information, mark duplicate, and pause publication.
- Rejection uses confirmation Dialog; success uses Toast.

### 3.14 Episode Review Detail

Goal:

- Review a single staged episode with full context and safe publication controls.

Main areas:

- Metadata fields.
- Source provenance.
- Audio artifact preview.
- Transcript or notes, if available.
- Validation issues.
- Rights notes.
- Publication targets.
- Decision history.

States:

- Pending.
- Approved.
- Rejected.
- Revision requested.
- Hold.

Key buttons:

- Save Metadata Edits.
- Approve.
- Reject.
- Request Revision.
- Publish to Configured Feeds.

Exceptions:

- Publication button is disabled until approval.
- Metadata edits are versioned and auditable.
- Reviewers cannot approve content outside their permission scope.

### 3.15 Publications

Goal:

- Configure and monitor RSS outputs.

Main areas:

- Publication scopes.
- Feed status.
- Episode inclusion rules.
- User or group access rules.
- RSS URL management.
- Last generation status.

States:

- Draft.
- Active.
- Paused.
- Error.
- Rights hold.

Key buttons:

- Create Publication.
- Preview Feed.
- Regenerate Feed.
- Pause Publication.
- Resume Publication.
- Revoke User Feed URL.

Exceptions:

- Public feeds require rights confirmation.
- Private and selected-user feeds require active authorization policy.
- A feed generation error should not delete the previous valid feed.

### 3.16 Users and Access

Goal:

- Manage which users can access Programs and RSS feeds.

Main areas:

- User list.
- Groups.
- Program entitlements.
- RSS token state.
- Recent access changes.

States:

- Active.
- Suspended.
- Deleted.
- Pending verification.
- RSS token active.
- RSS token revoked.

Key buttons:

- Grant Access.
- Revoke Access.
- Regenerate RSS URL.
- Suspend User.

Exceptions:

- Revoking access must immediately affect personal RSS output.
- Suspending a user revokes or suspends personal feed access by policy.
- Invitations are not implemented in M0. If represented in static UI, they are future capability placeholders only and must not become User statuses.

M0.2B static behavior:

- Users page displays `user` and `admin` account roles separately from System Owner, Operator, and Reviewer responsibility labels.
- Mock actions include suspend, restore, revoke RSS Token, inspect access, and view activity.
- No real invitation, user management API, session control, or RSS revocation is implemented.

### 3.17 Audit Log

Goal:

- Provide traceability for sensitive operations.

Main areas:

- Search and filters.
- Actor, action, target, timestamp.
- Before and after snapshots when appropriate.
- Related job/source/program links.

States:

- Loading.
- Empty.
- Filtered no results.
- Exporting.

Key buttons:

- Export Audit Log.
- Open Target.
- Clear Filters.

Exceptions:

- Secrets must be redacted in audit snapshots.
- Audit records should be append-only.

### 3.18 Settings

Goal:

- Manage platform-level defaults and safety controls.

Main areas:

- Runtime limits.
- Connector approval policy.
- Review policy.
- RSS defaults.
- Retention policy.
- Notification settings.

States:

- Saved.
- Unsaved changes.
- Validation error.

Key buttons:

- Save Settings.
- Reset to Defaults.
- Test Notification.

Exceptions:

- Lowering retention must warn about operational impact.
- Disabling review gates should be unavailable in phase 1.

## 4. User Flows

### 4.0 Register with Email

Goal:

- Let a new person create a normal `user` account safely.

Flow:

1. User opens Register.
2. User enters email, password, and password confirmation.
3. User completes Cloudflare Turnstile.
4. Platform verifies Turnstile token on the server.
5. Platform sends email verification code.
6. User enters verification code.
7. Platform validates code, activates account, and creates a secure session.

Main areas:

- Registration form.
- Turnstile area.
- Email verification step.
- Resend code action.
- Login link.

States:

- Initial form.
- Field validation error.
- Turnstile required.
- Turnstile failed, expired, or reused.
- Verification code sent.
- Resend cooling down.
- Code expired.
- Too many attempts.
- Account activated.

Key buttons:

- Create Account.
- Verify Email.
- Resend Code.
- Go to Login.

Exceptions:

- New registrations always create `user`, never `admin`.
- Errors must not reveal sensitive account existence details.
- Full verification codes must never appear in logs or UI diagnostics.

### 4.0.1 Login

Goal:

- Let users and administrators enter the product without leaking credential details.

Flow:

1. User opens Login.
2. User enters email and password.
3. Platform validates credentials and risk state.
4. High-risk login may require Turnstile.
5. Platform creates secure session and routes by role.

Main areas:

- Email field.
- Password field.
- Optional Turnstile challenge.
- Forgot password link.
- Register link.

States:

- Initial form.
- Loading.
- Generic invalid credentials.
- Turnstile required.
- Rate limited.
- Account unavailable.
- Success.

Key buttons:

- Log In.
- Continue after Turnstile.
- Forgot Password.

Exceptions:

- Failure messages must not distinguish nonexistent email from wrong password.
- Suspended, deleted, or pending verification accounts cannot receive active sessions.

### 4.0.2 Password Reset

Goal:

- Let a user regain access without exposing whether an email exists.

Flow:

1. User requests password reset with email.
2. Platform applies rate limits and Turnstile when required.
3. Platform sends a code or one-time link if the account can be reset.
4. User verifies proof and enters a new password.
5. Platform updates password, invalidates old sessions, and sends security notification.

Main areas:

- Reset request form.
- Proof verification form.
- New password form.
- Success notice.

States:

- Initial request.
- Instructions sent if account exists.
- Proof expired.
- Too many attempts.
- Password validation error.
- Reset success.

Key buttons:

- Send Reset Instructions.
- Set New Password.
- Back to Login.

Exceptions:

- Response must not reveal whether the email exists.
- Old sessions must be invalidated after successful reset.

### 4.0.3 Account Sessions

Goal:

- Let users inspect and revoke login devices in a later version.

Main areas:

- Current session.
- Other sessions.
- Device label.
- Last seen time.
- Revoke action.

States:

- No other sessions.
- Session revoked.
- Revoke failed.

Key buttons:

- Revoke Session.
- Revoke Other Sessions, if later supported.

Exceptions:

- Users can revoke only their own sessions unless a future admin endpoint is introduced.

### 4.1 Browse Authorized Programs

Flow:

1. User opens Authorized Programs.
2. Platform lists Programs currently allowed by user, group, or public access.
3. User opens Program detail.
4. User can subscribe to a Program feed or add it to a collection.

M0.2A static behavior:

- Search and scope filters run only against local Mock data.
- Program detail shows cover, author, category, language, update frequency, access state, rights state, and recent Mock episodes.
- "Add to Collection" opens a Drawer in the Program context.
- The Drawer can select an existing collection or create a new local Mock collection.
- Success feedback is shown with Toast; no API is called.

States:

- No authorized Programs.
- Program temporarily unavailable.
- Program has no published episodes.
- Access restricted.
- Long title or long description.

### 4.2 Create Personal Collection

Flow:

1. User opens My Collections.
2. User creates a collection name.
3. User adds authorized Programs.
4. Platform creates a collection feed containing only authorized content.

M0.2A static behavior:

- Collections are held in front-end memory only.
- Collection editor uses a local draft for title, description, selected Programs, order, and rules.
- RSS preview updates immediately from Mock episodes.
- Save only writes to local Mock state and shows success feedback.

Collection editor states:

- Normal collection.
- Empty collection.
- Rule set that produces no preview episodes.
- Unsaved changes.
- Save success.
- Permission denied.
- Long Program names.

Exceptions:

- If access to a Program is later revoked, its episodes disappear from the collection feed.

### 4.3 Get Personal RSS URL

Flow:

1. User opens a Collection subscribe page.
2. Platform displays a simulated personal RSS URL.
3. User copies URL into an external podcast client.
4. User can reset the simulated RSS address through a confirmation Dialog.

M0.2A static behavior:

- RSS URL uses `example.invalid`.
- Copy action shows success or failure Toast.
- Reset action shows confirmation Dialog and Mock success feedback.
- No RSS XML is generated.
- No real token is created.

States:

- Feed active.
- Feed revoked.
- Feed regenerating.
- Feed unavailable due to permissions.
- Copy success.
- Copy failure.

Key buttons:

- Copy RSS URL.
- Regenerate URL.
- Revoke URL.

Exceptions:

- Regenerating a URL invalidates the previous URL.
- A disabled account cannot use personal RSS URLs.

## 5. End-to-End Admin Flow

### 5.1 Create a Program and Add Manual Import

1. Create Program.
2. Add Source with `ingestion_type: manual_upload`.
3. Upload audio and metadata.
4. Platform creates Import Job.
5. Imported Episode enters Review Queue.
6. Reviewer approves.
7. Administrator publishes to configured RSS scope.

### 5.2 Add Connector Source with Reusable Session

1. Upload and approve Connector ZIP.
2. Add Source to Program.
3. Select Connector version.
4. Configure required input parameters.
5. Start authentication.
6. Session becomes valid.
7. Enable schedule.
8. Platform creates jobs according to schedule while session remains valid.

### 5.3 Run QR Each Time Connector

1. Add Source with `ingestion_type: connector`, `auth_mode: qr_each_run`, and `execution_mode: interactive`.
2. Operator clicks Run Import.
3. Platform creates Import Job and QR todo.
4. Operator scans QR.
5. Job runs once.
6. Connector exits.
7. Output episodes enter review.

## 6. Mobile Layout Expectations

Mobile admin pages should prioritize action clarity:

- Tables collapse into stacked rows.
- Primary status and next action appear at the top.
- Logs use horizontal scroll within a bounded panel.
- Dangerous actions remain behind confirmation dialogs.
- Review actions remain sticky at the bottom of Episode Review Detail.

## 7. Visual Experience Governance

All user-facing and admin flows must follow `VISUAL_DIRECTION.md`, `DESIGN_SYSTEM.md`, and `VISUAL_ACCEPTANCE.md`.

Shared requirements:

- User app and admin app share tokens, typography, spacing, semantic colors, button rules, and status language.
- User pages prioritize content, cover artwork, collections, and subscription clarity.
- Admin pages prioritize status clarity, next action, traceability, and moderate information density.
- Avoid generic enterprise dashboard appearance, excessive gradients, excessive glassmorphism, unrelated page styles, unnecessary icon decoration, and card-heavy stacking.

Every user-visible page must define:

- Desktop layout.
- Mobile layout.
- Loading state.
- Empty state.
- Error state.
- Permission denied state.
- Long text handling.
- Async success feedback.

Pages that must exist in M0.1 as representative high-fidelity static pages:

- Home.
- Register.
- Login.
- Program browse.
- Admin overview.
- Admin Program list.

Pages completed in M0.2:

- My Collections.
- RSS feeds.
- Account.
- Collection editor.
- Admin Source list.
- Admin Connector list.
- Admin Import Job list.
- Admin Review queue.
- Admin Publications.
- Admin Users and Access.
- Admin Audit.
- Admin Settings.
