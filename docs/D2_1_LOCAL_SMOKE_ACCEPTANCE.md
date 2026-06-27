# D2.1 Local User Beta Smoke Acceptance

D2.1 adds a repeatable local-only smoke script for the User Beta deployment candidate.

Implemented:

- `scripts/smoke-user-beta-local.sh`
- User Beta compose volume ownership initialization for the non-root API container.
- A Postgres media authorization query fix needed by real database-backed media reads.

Smoke scope:

- Starts the User Beta API stack under an isolated Compose project.
- Binds API only to `127.0.0.1`.
- Creates temporary admin, authorized user, and no-access user records.
- Inserts a minimal published Program, Episode, active access grant, and small text media fixture.
- Verifies authorized catalog visibility and no-access catalog hiding.
- Verifies private user media `GET`, `HEAD`, and `Range`.
- Creates a private RSS Feed and verifies Feed XML plus token-backed enclosure.
- Rotates the Feed token and verifies the old Feed and enclosure URLs fail.
- Revokes the Program grant and verifies user catalog, user media, Feed contents, and enclosure access update immediately.
- Runs PostgreSQL backup and restores it into a temporary database for verification.
- Cleans the temporary Compose project and volumes unless `--keep` is provided.

Safety rules:

- Default mode is dry-run.
- `--apply` is required to start containers or create temporary data.
- No real accounts, real email, real RSS tokens, real Connector packages, real media, Telegram, duoting, SSH, or public deployment are used.
- The script does not print plaintext RSS tokens, session cookies, generated secrets, database passwords, or complete private RSS URLs.
- Fixture media is a small text payload written only into the temporary local Compose media volume.
- Backup verification checks that the generated database password is not present in the backup file.

Validation commands:

```bash
bash -n scripts/smoke-user-beta-local.sh
scripts/smoke-user-beta-local.sh --dry-run
bash scripts/smoke-user-beta-local.sh --apply
```

If local Docker cannot build or pull the required base images, the `--apply` smoke must be reported as environment-blocked. Static script validation and the repository test suite still remain required.

Current workspace verification note:

- `bash scripts/smoke-user-beta-local.sh --apply` was attempted.
- The run was environment-blocked while Docker was pulling `alpine:3.20`; it did not reach application-level smoke assertions.
- The temporary Compose project was cleaned up and no smoke containers remained.
