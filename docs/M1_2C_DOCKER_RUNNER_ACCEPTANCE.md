# M1.2C Docker Runner Acceptance

## Implemented

- `RUNNER_MODE=disabled` default.
- `RUNNER_MODE=docker_trusted_admin` for Alpha fixture execution.
- `RUNNER_PYTHON_BASIC_IMAGE` image selection.
- `RUNNER_WORKSPACE_ROOT` workspace root selection.
- Independent `cmd/runner`; API service does not require Docker.
- Docker trusted-admin executor with structured `DockerRunSpec`.
- Docker CLI client that creates, starts, waits, logs, stops, kills when necessary, and removes containers.
- Running cancellation handling:
  - queued cancelled Jobs are not claimed
  - running cancellation requests cancel executor context
  - Runner records cancellation event
  - Job transitions to `cancelled`
- Timeout handling:
  - executor context deadline
  - Docker stop/kill path
  - Runner records timeout event
  - Job transitions to `failed`
  - `failure_code=timeout`
- Workspace cleanup after execution.

## Docker Security Assertions

Unit tests assert:

- `privileged=false`
- `network_mode != host`
- Docker socket is not mounted
- host root is not mounted
- `user != root`
- root filesystem is read-only
- `/work` is the only writable workspace mount
- connector package is mounted read-only at `/connector`
- CPU quota exists
- memory limit exists
- PID limit exists
- execution timeout exists

## Tests

- Disabled mode is explicit and non-executing.
- Docker parameter security assertions.
- Fixture Docker execution through a fake Docker client.
- queued cancel is not executed.
- running cancel stops execution and transitions to `cancelled`.
- timeout transitions to `failed` with `failure_code=timeout`.
- workspace cleanup remains enforced.
- API service has no Docker dependency.
- API responses continue to return metadata only, not Docker internals or file paths.

## Explicit Boundaries

- M1.2C still does not run real duoting or external Connectors.
- M1.2C still does not create Program, Episode, Review, RSS, subscription, or media download capability.
- M1.2C still does not support scheduled Jobs, interactive Jobs, or QR Jobs.
- `docker_trusted_admin` is not a strong sandbox for untrusted third-party Connectors.
- Domain-level network allowlist enforcement is not implemented; network policy remains auditable metadata for later hardening.
- Secrets are not passed through command line, image, Job Events, or `job.json`.
