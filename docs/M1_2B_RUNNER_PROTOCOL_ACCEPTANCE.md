# M1.2B Runner Protocol Acceptance

## Implemented

- Independent `cmd/runner` process.
- Explicit `RUNNER_MODE=fixture_subprocess` mode for local protocol validation.
- Default `RUNNER_MODE=disabled` behavior that does not execute Connector code.
- Runner-side queued Job claim through metadata store.
- Workspace creation under `RUNNER_WORKSPACE_ROOT` or `.local/runner-workspaces`.
- `/work/input/job.json` generation without Secret values.
- Fixture subprocess execution only.
- JSON Lines parser for:
  - `log`
  - `progress`
  - `artifact_ready`
  - `completed`
  - `failed`
- Protocol validation:
  - max line length
  - max total stdout
  - max event count
  - invalid JSON rejection
  - unknown event rejection
  - missing terminal rejection
  - duplicate terminal rejection
  - event after terminal rejection
  - exit code and terminal event consistency
- Event redaction before persistence.
- Artifact validation:
  - relative path only
  - no path traversal
  - no symlink, directory, socket, FIFO, or device Artifact
  - max count
  - max single-file size
  - max total size
  - every output file must be declared
  - duplicate Artifact rejection
  - SHA-256 and size recording
- Workspace cleanup after each run.

## Tests

- Fixture completed path.
- Invalid JSON Lines.
- Unknown event type.
- Missing terminal event.
- Duplicate terminal event.
- Event after completed.
- Exit code mismatch.
- Stdout line length limit.
- Event count limit.
- Redaction of token, cookie, authorization, password, session, and secret-like text.
- Normal Artifact metadata.
- Artifact path escape and absolute path rejection.
- Symlink and directory rejection.
- Artifact count and size limits.
- Undeclared Artifact rejection.
- SHA-256 and size correctness.
- Workspace cleanup.

## Explicit Boundaries

- M1.2B does not use Docker.
- M1.2B does not execute real duoting or external Connectors.
- M1.2B does not create Program, Episode, Review, RSS, subscription, or media download capability.
- M1.2B does not support scheduled Jobs, interactive Jobs, or QR Jobs.
- The API service does not read runner workspaces, execute Connector code, read Secret plaintext, or call Docker.
- Artifact APIs return metadata only and never return file contents or absolute paths.

## Deferred

- Trusted admin Docker Runner: M1.2C.
- Admin task console and Alpha deployment preparation: M1.2D.
