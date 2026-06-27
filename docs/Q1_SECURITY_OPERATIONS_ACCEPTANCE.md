# Q1 Security And Operations Acceptance

Q1 adds minimum User Beta security and operations hardening.

Implemented:

- Public, redacted `GET /metrics` endpoint for process and dependency readiness gauges.
- JSON request logging with server-generated `request_id`, method, redacted path, status, and duration.
- RSS private token path redaction before request paths are written to application logs.
- `scripts/cleanup-staging.sh`
- `scripts/cleanup-job-workspaces.sh`
- `scripts/cleanup-connector-packages.sh`
- User Beta preflight checks for API non-root compose shape, Docker socket boundary, RSS token proxy redaction, private storage exposure, backup directory safety, migration table reachability when local `psql` and non-placeholder `DATABASE_URL` are available, and `.env.user-beta` git-ignore behavior.

Security notes:

- Metrics do not include email addresses, tokens, cookies, URLs, filesystem paths, storage keys, or database connection strings.
- Request logging does not include query strings, request headers, cookies, authorization headers, RSS token plaintext, complete private RSS URLs, or storage keys.
- Cleanup scripts default to dry-run. Destructive behavior requires `--apply`.
- Cleanup output reports counts or safe basenames only, not absolute paths or storage keys.
- Reverse proxy or infrastructure must restrict `/metrics` and operational endpoints in private deployments.

Validation commands:

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
bash -n scripts/preflight-user-beta.sh
bash -n scripts/cleanup-staging.sh
bash -n scripts/cleanup-job-workspaces.sh
bash -n scripts/cleanup-connector-packages.sh
```
