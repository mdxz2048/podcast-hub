# M1.2E Runner Integration And Secret Boundary

## Implemented

- Real Docker fixture integration smoke test gated by `RUNNER_INTEGRATION_TEST=1`.
- Default Go test suite does not start Docker.
- Docker fixture uses:
  - no network
  - non-root user
  - read-only root filesystem
  - CPU, memory, PID, and timeout limits
  - `/work` as the only writable workspace
  - read-only Connector fixture mount
- Runner Secret boundary:
  - Runner resolves Secrets only after claiming a Job.
  - Runner revalidates Source, required Secret bindings, and revoked state before execution.
  - Runner decrypts Secret values only inside the Runner process.
  - Secret files are written under `/work/secrets`.
  - `job.json` contains only Secret logical names, types, and `/work/secrets/...` paths.
  - Secret values are not written to `job.json`, Docker command line, Docker environment, Job Events, Artifact metadata, or API responses.
  - Workspace cleanup removes `/work/secrets` on success, failure, cancellation, and timeout.
- Separate Runner compose:
  - `deploy/docker-compose.runner-alpha.yml`
  - only Runner mounts Docker socket
  - no public ports
  - `RUNNER_MODE=disabled` by default

## Tests

- Unit tests cover Secret injection without leakage.
- Unit tests cover missing required Secret and revoked Secret preflight failures.
- Unit tests cover API compose not mounting Docker socket and Runner compose owning the Docker socket.
- Integration test `TestIntegrationDockerTrustedAdminFixture` runs only when `RUNNER_INTEGRATION_TEST=1`.

## Boundaries

- This is not a duoting test.
- This does not use Telegram.
- This does not run real external Connectors.
- This does not download media.
- This does not create Program, Episode, Review, RSS, user subscription, or published content.
- `docker_trusted_admin` remains a trusted-admin Alpha mode, not a sandbox for untrusted third-party code.
- Network allowlist remains policy/audit metadata and is not domain-level enforcement.
