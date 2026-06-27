# D2 User Beta Deployment Acceptance

D2 adds private User Beta deployment candidate files and static validation. No public deployment was performed.

Implemented:

- `deploy/compose.user-beta.yml`
- `deploy/runner-compose.user-beta.yml`
- `deploy/Caddyfile.user-beta.template`
- `.env.user-beta.example`
- `scripts/preflight-user-beta.sh`
- `scripts/backup-postgres.sh`
- `scripts/restore-postgres.sh`
- `scripts/rotate-secrets-check.sh`
- `docs/OPERATIONS_USER_BETA.md`

Acceptance checks:

- API and Runner are separate services.
- API runs as non-root and does not mount Docker socket.
- Runner is separate; only Runner compose mounts Docker socket.
- PostgreSQL and Redis are private compose services with no public host port mapping.
- Connector packages, Import artifacts, staging files, Secret material, and private media use isolated paths or volumes.
- Caddy template supports HTTPS deployment shape without a real domain or certificate in Git.
- RSS token path logging is configured for redaction in the reverse proxy template.
- `RUNNER_MODE` defaults to `disabled`.
- `RUNNER_TRUSTED_ADMIN_ENABLED` is explicit.
- Preflight is dry-run by default and does not print sensitive values.
- Backup and restore scripts do not print passwords.
- Restore requires `--confirm-restore`.

Validation commands:

```bash
docker compose -f deploy/compose.user-beta.yml config
docker compose -f deploy/runner-compose.user-beta.yml config
bash -n scripts/preflight-user-beta.sh
bash -n scripts/backup-postgres.sh
bash -n scripts/restore-postgres.sh
bash -n scripts/rotate-secrets-check.sh
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```

Docker image build is optional and must be reported honestly if local Docker is unavailable.
