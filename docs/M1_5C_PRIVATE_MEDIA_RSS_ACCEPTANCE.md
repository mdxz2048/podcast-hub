# M1.5C Private Media And RSS User Flow Acceptance

M1.5C completes the user-facing private RSS management loop on top of the M1.5A/M1.5B backend.

Implemented:

- `/rss` now uses real `/me/rss-feeds` APIs.
- `src/api/rssFeeds.ts` provides user and admin RSS feed clients.
- `POST /me/rss-feeds` and rotate responses display plaintext token URLs only once in React state.
- List views display only Feed name, token prefix, status, created time, last-used time, expiration time, and revocation time.
- `/admin/rss-feeds` lists safe RSS metadata and allows admin revocation without plaintext token access.
- RSS XML and enclosure access audit metadata uses redacted token values and never stores complete private RSS URLs.

Security acceptance:

- Plaintext tokens are returned only on create or rotate success.
- Plaintext tokens are not written to localStorage, sessionStorage, route state, frontend logs, or list responses.
- Revoked feeds no longer display an effective URL.
- Admin views cannot read token plaintext or token hashes.
- User and admin pages do not display Connector, Source, Job, Artifact, storage keys, or Secret material.
- RSS XML and enclosure endpoints continue to enforce current feed token state, user status, Program grants, published Program state, published Episode state, and published media state.

Validation:

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```
