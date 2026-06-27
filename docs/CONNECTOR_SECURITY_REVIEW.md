# Connector Security Review

Every external Connector must pass security review before approval.

## Package Requirements

- `manifest.yaml`
- Python entrypoint source
- dependency lock file
- `README.md`
- tests
- sample output fixture

The package must execute once, write only to its assigned temporary output directory, emit JSON Lines events to stdout, and produce standardized episode metadata JSON files.

## Required Checks

- No `.env`, cookies, sessions, tokens, API hashes, passwords, QR session data, media, logs, databases, caches, or private keys.
- No Secret values in `manifest.yaml`.
- No database, Redis, Docker socket, host root filesystem, or production secret access.
- No daemon behavior, internal scheduler, final RSS XML generation, direct publishing, or production table writes.
- No attempts to bypass paywalls, DRM, CAPTCHA, authentication controls, platform restrictions, or access controls.
- Network requirements are explicit and minimal.
- Output metadata does not contain secret-like keys or values.
- Tests and fixture data use only non-sensitive synthetic or authorized sample data.

## Review Decision

Approve only when:

- the operator has authorization for the pilot content
- package structure is complete
- static validation passes
- no secrets or media are embedded
- runtime behavior matches the Connector protocol
- first-run scope is limited to staging

Reject or request revision when any check is incomplete.

## Main Repository Boundary

Podcast Hub main repository must not contain duoting-specific Connector code, real source credentials, source session material, downloaded media, external service logs, or private pilot analysis. Future duoting adaptation belongs in the external workbench and enters Podcast Hub only as an uploaded Connector ZIP after review.
