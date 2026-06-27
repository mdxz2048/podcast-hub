# Project Status (2026-06)

## Milestones

- M1.0: **Completed**
- M1.0C: **Completed**
- M1.1A: **Completed** (generic Connector Registry upload, static validation, version review, enable/disable)
- M1.1B: **Completed** (Connector Source instances, encrypted Secret records, Secret Reference bindings)
- M1.2A: **Completed** (ImportJob lifecycle, admin APIs, precondition checks, cancellation state rules)
- M1.2B: **Completed** (fixture-only Runner protocol, JSON Lines validation, Artifact metadata validation)
- M1.2C: **Completed** (trusted-admin Docker Runner mode for fixture execution only)
- M1.2D: **Completed** (admin Import Job console, disabled Runner visibility, and local/internal Alpha deployment preparation)
- M1.2E: **Completed** (Docker fixture integration smoke, Runner Secret injection boundary, separate Runner compose)
- M1.3A: **Completed** (completed ImportJob artifact intake into review-pending staging Program/Episode candidates)
- M1.3B: **Completed** (admin review, metadata edit, approve/reject, publish/archive state machine)
- M1.5A: **Completed** (private published-media delivery with authorization-aware storage promotion and HTTP range support)
- M1.5B: **Completed** (private RSS feeds, feed token rotation/revocation, and token-backed private enclosures)
- P0 User Beta baseline: **Completed** (request correlation IDs are now server-generated random values with no hostname, `.local`, path, token, cookie, or caller-supplied content)
- M1.3C: **Completed** (real authorized user catalog APIs, personal collections, and admin Program access-grant UI/API integration)
- M1.5C: **Completed** (real user RSS management page, admin RSS metadata/revoke page, and RSS token redaction audit call points)
- D2: **Completed** (private User Beta deployment candidate files, preflight, backup/restore scripts, and operations docs)
- D2.1: **Implemented in current workspace changes; local apply smoke environment-blocked** (local-only User Beta smoke script with fixture catalog/media/RSS, revocation checks, backup, and temporary restore verification; current Docker pull did not reach app assertions)

## Clarifications

- Duoting-specific connector code and planning materials are **not** part of the Podcast Hub main repository capability baseline.
- Podcast Hub platform capability is connector-agnostic: connectors are uploaded as versioned external packages and validated by platform policy.
- “Repository-internal duoting first-party connector” is **not** counted as a completed project capability.
- Real external Connector execution, scheduled jobs, interactive/QR jobs, real duoting, untrusted third-party Connector isolation, and real public production deployment are **not** implemented yet.
- M1.3A Program/Episode records are admin-only staging candidates. They are not published and are not visible to normal users.
- Private media delivery and private RSS now depend on explicit selected-user Program access grants rather than public catalog exposure.
- The `/rss` frontend page now uses real RSS API contracts and stores plaintext token URLs only in transient React state after create or rotate.
- User Program and Collection pages now read real API contracts. Collection reads intentionally hide Programs after access revocation without deleting historical membership rows.
- Error responses may include `request_id` for support correlation, but the value is opaque and non-semantic. Clients must not treat it as stable identity or derive environment information from it.
