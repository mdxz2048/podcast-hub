# Visual Direction

## 1. Direction

Podcast Hub should feel like a restrained, content-led, editorial, trustworthy audio library.

Visual keywords:

- Restrained.
- Editorial.
- Content-first.
- Trustworthy.
- Organized.
- Warm but not decorative.
- Operational without feeling like a generic enterprise dashboard.

Avoid:

- Generic enterprise admin look.
- Excessive gradients.
- Excessive glassmorphism.
- Different visual styles per page.
- Excessive card stacking.
- Large numbers of meaningless icons.
- Template-site or AI-generated page collage feeling.

User-facing pages and admin pages must share the same design tokens, typography scale, spacing system, semantic colors, button rules, and state language.

## 2. Product Home Page

Goal:

- Help a new user quickly understand that Podcast Hub is for authorized audio programs, personal collections, and RSS subscription.

Visual structure:

- Editorial hero with product name and concise value proposition.
- One primary action for sign up or enter app.
- One secondary action for login.
- A visible hint of program-library content below the first viewport.
- Selected program covers or tasteful content previews, not abstract dashboard decoration.
- Trust and authorization note presented as product policy, not legal clutter.

Layout:

- Desktop: strong left-aligned editorial headline, content preview region, clear actions.
- Mobile: headline first, actions immediately visible, content preview below.

Avoid:

- Overlarge SaaS marketing gradients.
- Fake analytics charts as hero decoration.
- Feature cards that bury the actual product.

## 3. Registration Page

Goal:

- Make account creation understandable and safe.

Visual structure:

- Compact centered or split editorial layout.
- Email, password, confirm password fields.
- Turnstile area.
- Clear primary action.
- Link to login.
- Brief note that new accounts are normal user accounts, not administrator accounts.

States:

- Initial form.
- Field validation.
- Turnstile required or failed.
- Code sent.
- Rate-limited resend.
- Email verification step.

Avoid:

- Asking for unnecessary profile fields.
- Making admin registration appear available.

## 4. Login Page

Goal:

- Provide a low-friction return path for users and admins.

Visual structure:

- Email and password fields.
- Optional Turnstile challenge area when required.
- Forgot password link.
- Register link.
- Generic failed-login error copy.

States:

- Initial.
- Loading.
- Invalid credentials.
- Turnstile required.
- Account unavailable.
- Success redirect.

Avoid:

- Revealing whether email exists.
- Multiple competing calls to action.

## 5. User Program Browse Page

Goal:

- Put program content, cover artwork, and subscription intent first.

Visual structure:

- Search and simple filters.
- Program cover grid or editorial list.
- Each Program shows cover, title, short description, episode count, authorization state, and subscribe/add-to-collection action.
- Empty state explains that no authorized programs are available.

Layout:

- Desktop: balanced grid with content preview and compact metadata.
- Mobile: single-column list with cover and primary action.

Avoid:

- Dense admin-style tables.
- Too many badges per Program.

## 6. My Collections Page

Goal:

- Help users manage personal RSS collections.

Visual structure:

- Collection list.
- Each collection shows title, included Program count, RSS state, and last updated time.
- Prominent create collection action.
- RSS URL access in a controlled field.

States:

- No collections.
- Collection active.
- RSS URL revoked.
- Permission changed.

Avoid:

- Treating collections like backend records.
- Showing full RSS tokens in broad list views.

## 7. Collection Editor

Goal:

- Make building a personal RSS collection clear and reversible.

Visual structure:

- Collection title field.
- Two-region layout:
  - Available authorized Programs.
  - Programs included in the collection.
- Search and filters.
- Save state and RSS preview.

States:

- Unsaved changes.
- Program access revoked.
- Empty collection.
- Save success.
- Save failure.

Avoid:

- Drag-only interactions without accessible alternatives.
- Hidden permission consequences.

## 8. Admin Overview Page

Goal:

- Show operational health, next actions, and risk without becoming a generic KPI wall.

Visual structure:

- A concise status strip for jobs, review, auth, and publication.
- Priority action queue.
- Recent import failures.
- Review queue summary.
- Connector health summary.
- Rights or publication holds.

Layout:

- Desktop: two-column operational layout with status and action queue.
- Mobile: stacked sections ordered by urgency.

Avoid:

- Decorative metric cards with no action.
- Large charts that do not support decisions.

## 9. Program Detail Management Page

Goal:

- Make one Program's content, sources, review, rights, and publishing state understandable at a glance.

Visual structure:

- Program masthead with cover, title, status, rights state, and publication state.
- Tabs for Sources, Episodes, Review, Publications, Activity.
- Right-side contextual panel for next actions and blockers.
- Source cards or rows that show ingestion mode, auth state, schedule, last job, and next action.

Avoid:

- Hiding rights state below the fold.
- Showing only raw tables without narrative status.

## 10. Connector Management Page

Goal:

- Govern Connector packages and versions safely.

Visual structure:

- Registry list with Connector name, runtime, latest approved version, validation state, usage count, failure rate, and deprecation state.
- Upload action.
- Manifest preview drawer.
- Version timeline.

States:

- Uploaded.
- Validating.
- Failed validation.
- Pending approval.
- Approved.
- Deprecated.
- Revoked.

Avoid:

- Treating Connector upload like a simple file attachment.
- Hiding validation failures behind generic errors.

## 11. Import Jobs and QR Tasks Page

Goal:

- Let operators understand job progress, blocked auth, and required human action.

Visual structure:

- Status filters.
- Job list with timeline-style status cues.
- QR/manual todo panel.
- Log preview drawer.
- Clear next action per blocked job.

States:

- Queued.
- Running.
- Waiting for QR scan.
- Failed.
- Completed with warnings.
- Timed out.

Avoid:

- Making QR tasks look like normal errors.
- Showing raw logs as the primary page content.

## 12. Review Queue Page

Goal:

- Help reviewers make fast, informed, auditable decisions.

Visual structure:

- Queue grouped by Program or risk level.
- Episode row with title, source, duration, publication date, validation warnings, rights hints, and audio preview.
- Decision actions remain visible.
- Detail drawer for metadata, provenance, duplicate candidates, and rights notes.

States:

- Pending review.
- Needs revision.
- On hold.
- Duplicate warning.
- Rights hold.

Avoid:

- Burying audio preview.
- Making approve and reject visually equivalent to destructive actions without hierarchy.

## 13. Shared Visual Governance

All pages must use:

- Shared design tokens.
- Shared layout primitives.
- Shared button hierarchy.
- Shared status badge language.
- Shared form and validation patterns.
- Shared empty, loading, error, and success states.

Visual consistency must be checked before a page is accepted.

