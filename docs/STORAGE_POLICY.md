# Storage Policy

## 1. Scope

This document freezes the storage policy for RSS authorization, media handling, Connector staging outputs, task artifacts, and failed job retention.

No storage implementation is created in M0.0 or M0.1.

## 2. RSS Authorization

Every RSS request must perform real-time authorization checks.

Required checks:

- Validate RSS Token.
- Validate User status.
- Validate that the User still has Program or Collection access.
- Validate that the Program is published and not on hold.
- Validate that each included Episode is approved and publishable.

Rules:

- Token revocation must take effect immediately.
- User permission revocation must take effect immediately.
- User suspension or deletion must take effect immediately.
- Cache layers must not bypass authorization checks.
- Cached feed fragments must be scoped so that they cannot leak unauthorized content.

## 3. Media Storage

Future storage target:

- S3-compatible object storage for authorized media and task artifacts.

Rules:

- Manual uploads, approved media, and published media may move into controlled object storage in later implementation phases.
- Connector outputs first enter an isolated staging area.
- Unreviewed content must not enter formal RSS publication.
- Published media must be served only through authorization-aware paths.

## 4. External Source URLs

External URLs from source systems are stored as provenance only:

- `source_url`
- canonical source reference
- external source metadata

Rules:

- Do not default to using external URLs as formal RSS `enclosure` values.
- A future explicit policy is required before an external media URL can be used as an RSS enclosure.
- External source URLs do not prove redistribution rights.

## 5. Connector Staging

Connector outputs are staged before review.

Staging area rules:

- Isolated from published media.
- Scoped by Import Job.
- Not writable by other jobs.
- Not directly served to RSS clients.
- Cleared according to retention rules after output processing.

## 6. Failed Job Retention

Default policy:

- Failed job raw media artifacts are immediately deleted.
- Sanitized logs and metadata are retained for 30 days by default.
- Raw secrets, session material, cookies, tokens, QR session data, and passwords are never retained in logs.

Retained failed-job metadata may include:

- Job ID.
- Program ID.
- Source ID.
- Connector version.
- Trigger type.
- Auth mode.
- Execution mode.
- Redacted error classification.
- Redacted JSON Lines event summary.

## 7. Manual Upload Retention

Manual upload media remains staged until review.

Rules:

- Rejected manual media should be deleted according to later retention policy.
- Approved manual media may move to controlled object storage.
- Published manual media must remain tied to authorization-aware RSS access.

## 8. Non-Goals for M0 and M0.1

Do not implement:

- Object storage clients.
- Upload APIs.
- Media processing.
- RSS XML generation.
- Signed media URLs.
- Connector artifact upload.
- Cleanup workers.

