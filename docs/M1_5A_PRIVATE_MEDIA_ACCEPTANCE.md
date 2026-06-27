# M1.5A Private Media Delivery Acceptance

Status: implemented in current workspace changes.

## Scope

M1.5A adds authorization-aware private media delivery for published Episodes.

Implemented:

- `media_assets.delivery_status`, `published_storage_key`, `published_at`, and expanded media lifecycle states.
- Isolated staging and published media filesystem roots through `STAGING_STORE_DIR` and `MEDIA_STORE_DIR`.
- Episode publish now requires successful media promotion before the Episode can enter `published`.
- Authorization-aware user media endpoints backed by published media storage only.
- Feed-token-aware enclosure delivery path for future RSS clients.
- HTTP `HEAD`, `GET`, `Range`, `ETag`, and `Last-Modified` behavior for private media responses.
- Private cache headers and non-indexing response headers for sensitive delivery paths.
- Access-grant-aware backend checks so revocation takes effect immediately.
- HTTP and service tests covering unauthorized access, conditional requests, range handling, and publish/promotion preconditions.

Not implemented:

- Public media URLs.
- Browser player UI.
- Download buttons.
- External object storage backend.
- CDN caching for private media.
- Real user catalog APIs beyond the delivery and authorization checks needed by private media.

## Delivery Rules

- Only `approved` media can be promoted during Episode publish.
- Promotion failure prevents Episode publication.
- Staging and published media roots must remain isolated.
- `archived`, `deleted`, and `quarantined` media are not deliverable.
- User-facing APIs never return staging paths, published storage keys, Runner workspace paths, or artifact IDs.
- Unauthorized callers receive a uniform safe failure response.
- Program access revocation immediately blocks further media reads.

## API

- `HEAD /media/episodes/{episodeId}`
- `GET /media/episodes/{episodeId}`
- `HEAD /rss/private/{opaqueToken}/episodes/{episodeId}/media`
- `GET /rss/private/{opaqueToken}/episodes/{episodeId}/media`

Rules:

- Session-backed media access requires an authenticated active user with active Program access.
- Token-backed media access requires an active feed token whose owner still has active Program access.
- Episode and Program must both be `published`.
- Media must be `published` with a non-empty private published storage key.

## Security Boundary

Private media responses do not expose:

- Staging paths.
- Published storage keys.
- Runner workspaces.
- Artifact IDs.
- Secret values.
- Session cookies.
- Feed tokens in response bodies.

Response headers enforce:

- `Cache-Control: private, no-store, max-age=0`
- `Pragma: no-cache`
- `Referrer-Policy: no-referrer` on token-based enclosure responses
- `X-Robots-Tag: noindex, nofollow, noarchive` on token-based enclosure responses

## Verification

Executed for this change set:

```bash
find cmd config internal -name '*.go' -print0 | xargs -0 gofmt -w
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
corepack pnpm run --if-present test
corepack pnpm build
git diff --check
```

Screenshot workflows were intentionally not run.