# Environment

## 1. Scope

This document defines environment and local configuration policy.

M0.0 and M0.1 do not create real services, Docker, databases, API servers, authentication, email, Turnstile, RSS, upload, or Connector execution environments.

## 2. Environment Files

Allowed in Git:

- `.env.example`

Forbidden in Git:

- `.env`
- `.env.*` except `.env.example`
- real passwords
- real cookies
- real tokens
- SSH private keys
- server credentials
- database credentials
- object storage root keys

## 3. Placeholder Policy

`.env.example` may contain placeholders only.

Examples:

- `change_me`
- `example.invalid`
- local development URLs

It must not contain:

- real secrets
- real production hostnames with credentials
- real access keys
- real session keys

## 4. M0.1 Local Environment

M0.1 may require only frontend local development configuration once implementation begins.

It must not require:

- PostgreSQL.
- Redis.
- S3-compatible object storage.
- Cloudflare Turnstile.
- Email provider.
- Docker.
- Connector runner.

## 5. Long-Term Backend Environment

Frozen long-term backend architecture:

- Go platform backend.
- PostgreSQL main database.
- Redis for cache, rate limiting, and task queue.
- S3-compatible object storage for authorized media and task artifacts.
- Python only for Connector SDK and Connector execution environment.

Python FastAPI is not used as the platform main API.

