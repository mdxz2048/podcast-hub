# Design System

## 1. Product Feel

Podcast Hub should feel like a coherent operational console for content, jobs, authorization, and publishing.

The design should be:

- Clear.
- Calm.
- Dense enough for repeated admin work.
- Explicit about state and next action.
- Consistent across admin and user surfaces.

Avoid:

- Decorative layouts that hide operational data.
- One-off component styles.
- Ambiguous status colors.
- Hidden destructive actions.
- Text-heavy explanations where state and controls can be clearer.

Brand temperament:

- Restrained.
- Content-led.
- Editorial.
- Trustworthy.
- Organized.
- Warm without being decorative.

The product should feel like a credible audio content library with operational depth, not a generic enterprise backend.

## 1.1 Design Tokens

User-facing pages and admin pages must share one token system.

CSS Variables are the only source of design tokens. Tailwind CSS must consume or mirror those variables rather than becoming a second source of truth.

### 1.1.1 Color Tokens

Use semantic tokens rather than hard-coded page colors.

Core tokens:

- `color.background.canvas`: page background.
- `color.background.surface`: panels, tables, drawers.
- `color.background.subtle`: soft section background.
- `color.text.primary`: main text.
- `color.text.secondary`: supporting text.
- `color.text.muted`: low-emphasis metadata.
- `color.border.default`: normal border.
- `color.border.strong`: emphasized border.
- `color.action.primary`: primary action.
- `color.action.primaryHover`: primary action hover.
- `color.action.secondary`: secondary action.
- `color.status.success`: completed, approved, active.
- `color.status.warning`: pending, expiring, needs attention.
- `color.status.danger`: failed, rejected, destructive.
- `color.status.info`: running, processing, informational.
- `color.status.neutral`: draft, disabled, archived.
- `color.focus.ring`: keyboard focus.

Palette guidance:

- Prefer quiet neutrals and restrained accent colors.
- Use accent color to clarify action and status, not decoration.
- Avoid purple-blue gradient dominance, excessive beige, dark slate-only UI, or one-hue themes.

### 1.1.2 Typography Tokens

Typography should support editorial content and operational scanning.

Recommended tokens:

- `font.family.sans`: primary UI font.
- `font.family.mono`: logs, IDs, protocol fields.
- `font.size.xs`: metadata and compact labels.
- `font.size.sm`: secondary UI text.
- `font.size.md`: default body and controls.
- `font.size.lg`: page section heading.
- `font.size.xl`: page title.
- `font.size.2xl`: marketing or home headline.
- `font.weight.regular`
- `font.weight.medium`
- `font.weight.semibold`
- `lineHeight.tight`
- `lineHeight.normal`
- `lineHeight.relaxed`

Rules:

- Do not scale font size directly with viewport width.
- Letter spacing should stay at `0` unless a specific component has a documented reason.
- Hero-scale typography is reserved for home and major editorial moments.

### 1.1.3 Spacing Tokens

Use a predictable spacing scale:

- `space.1`: 4px
- `space.2`: 8px
- `space.3`: 12px
- `space.4`: 16px
- `space.5`: 20px
- `space.6`: 24px
- `space.8`: 32px
- `space.10`: 40px
- `space.12`: 48px
- `space.16`: 64px

Rules:

- Dense admin areas can use `space.2` to `space.4`.
- User content pages should use more breathing room without becoming sparse.
- Repeated lists should keep stable row heights and predictable gaps.

### 1.1.4 Shape, Shadow, and Border Tokens

Recommended tokens:

- `radius.sm`: 4px
- `radius.md`: 6px
- `radius.lg`: 8px
- `border.width.default`: 1px
- `shadow.none`
- `shadow.subtle`
- `shadow.overlay`

Rules:

- Cards should use 8px radius or less unless a future brand revision changes it.
- Avoid heavy shadows.
- Use borders and spacing before shadows for structure.
- Avoid glassmorphism as a default surface treatment.

### 1.1.5 Responsive Breakpoints

Recommended breakpoints:

- `breakpoint.mobile`: 0-639px.
- `breakpoint.tablet`: 640-1023px.
- `breakpoint.desktop`: 1024-1439px.
- `breakpoint.wide`: 1440px and above.

Rules:

- Mobile is not a squeezed desktop table.
- Drawers become full-screen sheets on mobile.
- Tables become stacked list rows on mobile.
- Main actions remain discoverable.

### 1.1.6 Dark Mode Strategy

Phase 1 explicitly does not support dark mode.

Even if unsupported:

- Use semantic color tokens.
- Avoid hard-coded foreground/background assumptions.
- Avoid images or shadows that only work on light backgrounds.
- Keep component APIs compatible with future theme switching.
- Keep future `data-theme="dark"` extension possible.
- Prevent browser or OS automatic dark mode from producing unreadable pages.
- Visual acceptance must state "dark mode not supported in phase 1" when no dark screenshot is provided.

## 2. Layout Principles

### 2.1 Admin Layout

Recommended structure:

- Persistent left navigation on desktop.
- Top bar with environment, user menu, and global search.
- Page header with title, status, and primary action.
- Main content with tabs or split panels.
- Right-side drawer for secondary detail when useful.

Admin pages should favor:

- Tables for repeatable operational records.
- Compact cards for status summaries.
- Detail panels for one selected object.
- Drawers for edit forms and previews.
- Dialogs for confirmations and destructive actions.

Admin pages should not become traditional table stacks. Use tables where scanning is the job, but pair them with summary strips, next-action panels, timelines, and drawers that explain what needs attention.

### 2.2 User Layout

User-facing pages should be simpler:

- Authorized Programs.
- Collections.
- RSS URLs.
- Account.

The user should never need to understand Connector, job, or review internals to subscribe to a feed.

### 2.3 Information Hierarchy

Page hierarchy:

- Page title and primary state.
- One primary action.
- Critical blocker or next action.
- Main content.
- Secondary metadata.
- Audit or historical detail.

Rules:

- Rights holds, auth blockers, failed jobs, and permission problems must appear near the top.
- Do not bury next action inside logs or secondary tabs.
- Use status badges consistently and sparingly.

## 3. Component Inventory

Required shared components:

- `Button`
- `Input`
- `SearchBar`
- `Select`
- `Badge`
- `ProgramCard`
- `EpisodeRow`
- `ConnectorCard`
- `ImportJobCard`
- `EmptyState`
- `LoadingState`
- `ErrorState`
- `Drawer`
- `Dialog`
- `Toast`

Recommended additions:

- `StatusTimeline`
- `LogViewer`
- `MetadataPanel`
- `PermissionPicker`
- `ScheduleEditor`
- `ManifestPreview`
- `AuditEventRow`
- `ReviewDecisionBar`
- `RssUrlField`

## 4. Buttons

Button hierarchy:

- Primary: one main page action.
- Secondary: common non-destructive actions.
- Tertiary or ghost: low-emphasis navigation or utility.
- Danger: destructive or access-changing actions.

Button rules:

- Use clear verbs: Save, Run Import, Approve, Reject, Pause Schedule.
- Disable unavailable actions with a short reason in tooltip or inline helper.
- Dangerous actions require confirmation.
- Long-running actions show progress and success or error feedback.

## 5. Status Badges

Badges should use stable labels and colors.

### 5.1 Program Status

| State | Label | Meaning |
| --- | --- | --- |
| `draft` | Draft | Program is not ready for active operation. |
| `active` | Active | Program can ingest and publish according to policy. |
| `disabled` | Disabled | Jobs and publication changes are blocked. |
| `rights_hold` | Rights Hold | Publication is blocked by policy. |
| `archived` | Archived | Program is retained for history. |

### 5.2 Source Status

| State | Label | Meaning |
| --- | --- | --- |
| `ready` | Ready | Source can run. |
| `auth_required` | Auth Required | Human authentication is needed. |
| `schedule_paused` | Schedule Paused | Automatic runs are paused. |
| `connector_unavailable` | Connector Unavailable | Bound Connector version cannot run. |
| `disabled` | Disabled | Source cannot run. |

### 5.3 Connector Version Status

| State | Label | Meaning |
| --- | --- | --- |
| `uploaded` | Uploaded | ZIP received but not validated. |
| `validating` | Validating | Checks are running. |
| `validation_failed` | Validation Failed | Package cannot be approved. |
| `pending_approval` | Pending Approval | Awaiting human approval. |
| `approved` | Approved | May be used by Sources. |
| `deprecated` | Deprecated | Avoid new usage. |
| `revoked` | Revoked | Cannot run. |

### 5.4 Job Status

| State | Label | Meaning |
| --- | --- | --- |
| `queued` | Queued | Job is waiting to start. |
| `preparing` | Preparing | Platform is preparing execution. |
| `waiting_for_auth` | Waiting for Auth | Human authentication is required. |
| `running` | Running | Connector is executing. |
| `processing_output` | Processing Output | Platform is validating artifacts. |
| `completed` | Completed | Job succeeded. |
| `completed_with_warnings` | Completed with Warnings | Job succeeded with warnings. |
| `failed` | Failed | Job failed. |
| `cancelled` | Cancelled | Job was cancelled. |
| `timed_out` | Timed Out | Job exceeded timeout. |

### 5.5 Review Status

| State | Label | Meaning |
| --- | --- | --- |
| `pending_review` | Pending Review | Waiting for decision. |
| `approved` | Approved | Eligible for publication. |
| `rejected` | Rejected | Not eligible. |
| `needs_revision` | Needs Revision | Must be corrected. |
| `on_hold` | On Hold | Blocked by policy or investigation. |
| `published` | Published | Included in a feed. |

## 6. Empty, Loading, Error, and Success States

Every page and async action must include these states.

### 6.1 Loading State

Use skeletons for tables and details. Use spinners only for short inline actions.

### 6.2 Empty State

Empty states should say:

- What is empty.
- Why it matters.
- The next safe action.

Examples:

- No Programs yet: Create Program.
- No Sources: Add Source.
- No Review Items: All imported episodes have decisions.

### 6.3 Error State

Error states should include:

- Human-readable summary.
- Retry action if safe.
- Link to logs or audit details when relevant.
- No secrets.

### 6.4 Success Feedback

Use toast or inline success feedback for:

- Saved settings.
- Job queued.
- Review decision recorded.
- RSS URL regenerated.
- Connector uploaded.

## 7. Tables and Lists

Tables are preferred for:

- Programs.
- Sources.
- Connector versions.
- Import jobs.
- Review queue.
- Users.
- Audit log.

Table rules:

- First column should identify the object.
- Status column should be near the object name.
- Time columns should show relative and exact time on hover.
- Actions should be right-aligned.
- Filters should preserve URL state.
- Mobile layout should convert rows into stacked cards.

Lists should use richer rows when pure tables would obscure content. Program, Connector, Import Job, and Episode rows may include status summaries, next action, and compact metadata.

## 7.1 Content Card Rules

Cards are for individual repeated content or operational objects, not for wrapping whole page sections.

Program cards:

- Show cover, title, short description, episode count, access state, and primary content action.
- Handle missing cover with a quiet, branded fallback.
- Clamp long descriptions while preserving full text in detail.

Connector cards or rows:

- Show runtime, version, validation state, usage count, and last job health.
- Avoid decorative icons unless they clarify status or action.

Import Job cards or rows:

- Show job status, source, trigger, duration, last event, and next action.

## 8. Detail Pages

Detail pages should include:

- Header with object name, status, and primary action.
- Tabs for major subareas.
- Activity timeline or audit summary.
- Clear disabled states for blocked actions.

Common tabs:

- Overview.
- Configuration.
- Jobs.
- Review.
- Publications.
- Audit.

## 9. Drawers and Dialogs

Use drawers for:

- Editing metadata.
- Previewing manifest details.
- Viewing job logs from a list page.
- Reviewing source configuration.

Use dialogs for:

- Confirmation.
- Destructive actions.
- Access changes with broad impact.
- RSS token regeneration.

Dialog rules:

- Show the affected object.
- Show consequences.
- Require explicit confirmation for destructive actions.

## 10. Log Viewer

The Log Viewer must:

- Display JSON Lines events in chronological order.
- Show level, event type, time, and message.
- Allow filtering by level and event type.
- Preserve structured `data` display.
- Redact secrets.
- Support copy or download only for sanitized logs.

Error events should link to:

- Job status.
- Connector version.
- Source configuration.
- Output artifact validation error.

## 11. Review Interface

Review screens should prioritize:

- Episode title.
- Source provenance.
- Audio preview.
- Required metadata completeness.
- Rights notes.
- Duplicate candidates.
- Decision history.

Review actions:

- Approve.
- Reject.
- Request Revision.
- Put on Hold.

Review actions should remain visible on desktop and sticky on mobile detail pages.

## 12. RSS URL Interface

RSS URL display rules:

- Treat personal RSS URLs as sensitive.
- Do not show full token in broad table views.
- Show copy action only on the user's own feed or authorized admin views.
- Regeneration requires confirmation.
- Revocation requires confirmation.

States:

- Active.
- Revoked.
- Regenerating.
- Permission blocked.

## 13. Accessibility

Minimum requirements:

- Keyboard-accessible controls.
- Visible focus states.
- Color is not the only state indicator.
- Form fields have labels.
- Error messages are attached to fields.
- Tables have meaningful headers.
- Log viewer supports text selection and screen-reader labels.

## 13.1 Forms

Form rules:

- Labels are always visible or programmatically available.
- Required fields are clear.
- Validation appears near the field.
- Password fields support show/hide controls only when implemented accessibly.
- Authentication forms do not reveal whether an email exists.
- Submit buttons show loading and prevent duplicate submission.
- Resend actions show cooldown state.

## 13.2 Permission and Dangerous States

Permission states:

- Use clear text and neutral structure.
- Explain what access is missing.
- Provide a safe next action where possible.

Danger states:

- Use danger color only for destructive, rejected, failed, or policy-blocking states.
- Destructive actions require confirmation.
- Confirmation dialogs must show affected object and consequence.
- Admin role changes, user suspension, RSS token revocation, Connector revocation, and publication takedown must be auditable.

## 14. Mobile Rules

Mobile layout must:

- Avoid horizontal page overflow.
- Convert tables into stacked rows.
- Keep primary action visible.
- Keep review decision actions accessible.
- Use drawers as full-screen sheets.
- Keep log text readable with contained horizontal scrolling.

## 15. Copy Style

Use direct operational language.

Good:

- "Session expired. Reauthorize this Source before the next scheduled run."
- "This Connector version is revoked and cannot start new jobs."
- "Public publication is blocked until rights notes are complete."

Avoid:

- Vague errors.
- Blaming users.
- Revealing internal stack traces.
- Mentioning secrets or raw credential data.

## 16. Visual Acceptance Link

Every user-visible page must satisfy `VISUAL_ACCEPTANCE.md` before it is considered complete.

Required evidence includes:

- Desktop screenshot.
- Mobile screenshot.
- Empty state screenshot.
- Loading state screenshot.
- Error state screenshot.
- Long text screenshot.
- Permission denied screenshot.
- Async success feedback screenshot.
- Dark mode screenshot or explicit phase-1 unsupported note.
- Mock data and boundary data verification.
