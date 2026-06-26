# Mock Data

## 1. Scope

M0.1 and M0.2 use Mock data only.

Mock data must not:

- Call real APIs.
- Store real user credentials.
- Send email.
- Use real Cloudflare Turnstile.
- Upload audio.
- Generate RSS XML.
- Upload Connector ZIP packages.
- Execute Connectors.

## 2. Core Mock Models

### 2.1 User

Fields:

- `id`
- `email`
- `displayName`
- `role`: `user` or `admin`
- `status`: `pending_verification`, `active`, `suspended`, or `deleted`
- `responsibilityLabels`: for admin profiles such as `system_owner`, `operator`, `reviewer`

### 2.2 Program

Fields:

- `id`
- `title`
- `description`
- `coverUrl`
- `status`
- `rightsState`
- `publicationState`
- `episodeCount`
- `sourceCount`
- `accessState`

### 2.3 Source

Fields:

- `id`
- `programId`
- `name`
- `ingestionType`: `native_rss`, `connector`, or `manual_upload`
- `triggerType`: `manual` or `scheduled`
- `authMode`: `none`, `reusable_session`, or `qr_each_run`
- `executionMode`: `unattended` or `interactive`
- `status`
- `lastJobStatus`
- `nextRunAt`

### 2.4 ConnectorVersion

Fields:

- `id`
- `connectorId`
- `name`
- `version`
- `runtime`
- `status`
- `usageCount`
- `lastJobHealth`

### 2.5 ImportJob

Fields:

- `id`
- `programId`
- `sourceId`
- `ingestionType`
- `triggerType`
- `authMode`
- `executionMode`
- `status`
- `startedAt`
- `finishedAt`
- `errorCategory`
- `nextAction`

### 2.6 ReviewItem

Fields:

- `id`
- `programId`
- `jobId`
- `title`
- `durationSeconds`
- `reviewStatus`
- `rightsWarning`
- `duplicateWarning`

### 2.7 Collection

Fields:

- `id`
- `title`
- `programIds`
- `rssTokenState`
- `lastUpdatedAt`

## 3. Boundary Data

Mock data must include:

- Very long Program title.
- Missing cover image.
- Program with no episodes.
- Program with many Sources.
- Connector version revoked.
- Job waiting for QR scan.
- Job failed with long redacted error.
- Review queue with duplicate warning.
- User with no access.
- Suspended user.
- Revoked RSS token.

## 4. Static State Sets

Every M0.1 representative page should have Mock variants for:

- Loading.
- Empty.
- Error.
- Permission denied.
- Success feedback.
- Long text.

## 5. Naming Rules

Mock data should use:

- Clearly fake IDs.
- Example domains such as `example.invalid`.
- No real tokens, cookies, passwords, email verification codes, or session secrets.

