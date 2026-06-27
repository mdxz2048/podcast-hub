# M1.3C User Catalog And Collections Acceptance

M1.3C replaces user-facing Program and Collection mock data with real API contracts.

Implemented backend APIs:

- `GET /programs`
- `GET /programs/{programId}`
- `GET /programs/{programId}/episodes`
- `GET /episodes/{episodeId}`
- `GET /me/collections`
- `POST /me/collections`
- `PATCH /me/collections/{collectionId}`
- `DELETE /me/collections/{collectionId}`
- `POST /me/collections/{collectionId}/programs`
- `DELETE /me/collections/{collectionId}/programs/{programId}`
- `GET /admin/programs/{programId}/access-grants`
- `POST /admin/programs/{programId}/access-grants`
- `POST /admin/program-access/{grantId}/revoke`

Security acceptance:

- User catalog reads require an active logged-in user.
- Program reads require an active Program access grant and `published` Program state.
- Episode reads require active grant, `published` Program, `published` Episode, and `published` audio media.
- Unauthorized Program and Episode reads return a generic not-found response without titles, Source, Connector, Job, Artifact, storage key, Secret, or token data.
- Personal collection membership may retain revoked Program IDs internally, but collection reads only return currently authorized Programs.
- Collection writes are owner-scoped and CSRF-protected.
- Admin access-grant writes require admin role, active target user, and write audit events.

Frontend acceptance:

- `/programs`, `/programs/:programId`, `/collections`, and `/collections/:collectionId` use real API clients.
- User pages do not fall back to `src/mock/data` for catalog or collection content.
- User pages do not display Connector, Source, Job, Artifact, storage keys, RSS tokens, or download links.
- Playwright tests mock real API contracts and cover authorized catalog, Program detail, collection create/edit/remove, and admin grant/revoke.

Validation:

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```
