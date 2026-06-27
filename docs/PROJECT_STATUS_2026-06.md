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

## Clarifications

- Duoting-specific connector code and planning materials are **not** part of the Podcast Hub main repository capability baseline.
- Podcast Hub platform capability is connector-agnostic: connectors are uploaded as versioned external packages and validated by platform policy.
- “Repository-internal duoting first-party connector” is **not** counted as a completed project capability.
- Real external Connector execution, user catalog authorization, private RSS, user subscription, scheduled jobs, interactive/QR jobs, real duoting, untrusted third-party Connector isolation, and production deployment are **not** implemented yet.
- M1.3A Program/Episode records are admin-only staging candidates. They are not published and are not visible to normal users.
- M1.3B published Program/Episode states are still admin-only until M1.3C adds user access grants and catalog APIs.
