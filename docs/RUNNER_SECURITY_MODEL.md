# Runner Security Model

## Current Mode

M1.2C supports:

- `RUNNER_MODE=disabled`
- `RUNNER_MODE=docker_trusted_admin`

The default is disabled. In disabled mode, no Connector code is executed and queued Jobs remain queued until an enabled Runner claims them.

`docker_trusted_admin` is an Alpha mode for administrator-trusted fixture execution. It is not a production-grade sandbox for arbitrary third-party Connectors.

## Separation of Responsibilities

API service:

- creates Import Jobs
- validates admin authorization and Source preconditions
- exposes Job metadata, redacted Events, and Artifact metadata
- records cancellation requests
- does not execute Connector code
- does not read Runner workspaces
- does not read Secret plaintext for Runner execution
- does not require Docker socket access

Runner:

- claims queued Jobs
- creates temporary workspaces
- writes `job.json` without Secret values
- executes the fixture Connector
- parses JSON Lines events
- validates Artifacts
- writes redacted Events and metadata only
- cleans workspaces
- may have Docker access only when explicitly started in `docker_trusted_admin` mode

Connector container:

- receives `/work/input/job.json`
- writes only under `/work/output`
- emits JSON Lines to stdout
- does not receive Docker socket
- does not receive host root mount
- does not receive database, Redis, or production Secret access

## Docker Controls

The Runner builds Docker specs with:

- `privileged=false`
- non-host network mode
- non-root user
- read-only root filesystem
- writable `/work` mount only
- read-only `/connector` package mount
- CPU quota
- memory limit
- PID limit
- timeout
- container removal after completion or failure

## Cancellation and Timeout

- A queued Job cancelled before claim is not executed.
- A running Job with `cancellation_requested_at` causes Runner context cancellation.
- The Docker client attempts graceful stop, then force kill.
- Timeout also stops execution and writes `failure_code=timeout`.

## Not Yet Implemented

- Strong sandbox for untrusted third-party Connectors.
- Domain-level network allowlist enforcement.
- Secret material injection into Connector runtime.
- Scheduled Jobs.
- Interactive or QR Jobs.
- Real external Connector execution.
- Program, Episode, RSS, subscription, or media publication paths.
