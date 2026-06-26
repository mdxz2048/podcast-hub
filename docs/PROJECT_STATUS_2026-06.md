# Project Status (2026-06)

## Milestones

- M1.0: **Completed**
- M1.0C: **Completed**
- M1.1A: **Completed** (generic Connector Registry upload, static validation, version review, enable/disable)
- M1.1B: **Completed** (Connector Source instances, encrypted Secret records, Secret Reference bindings)
- M1.2A: **Completed** (ImportJob lifecycle, admin APIs, precondition checks, cancellation state rules)
- M1.2B: **Completed** (fixture-only Runner protocol, JSON Lines validation, Artifact metadata validation)
- M1.2C: **Not Started**
- M1.2D: **Not Started**

## Clarifications

- Duoting-specific connector code and planning materials are **not** part of the Podcast Hub main repository capability baseline.
- Podcast Hub platform capability is connector-agnostic: connectors are uploaded as versioned external packages and validated by platform policy.
- “Repository-internal duoting first-party connector” is **not** counted as a completed project capability.
- Docker Connector execution, Program creation, Episode creation, staging review, private RSS, user subscription, scheduled jobs, interactive/QR jobs, real duoting, untrusted third-party Connector isolation, and production deployment are **not** implemented yet.
