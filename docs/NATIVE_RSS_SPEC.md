# Native RSS Import Specification

## 1. Scope

Native RSS import is a platform built-in Importer.

It is not:

- An administrator-uploaded Connector ZIP.
- A Python Connector.
- A general web scraper.
- A bypass mechanism for restricted content.

Native RSS still follows:

- Source.
- Import Job.
- Imported Episode.
- Review.
- Publication.

## 2. Source Configuration

Frozen dimensions:

- `ingestion_type`: `native_rss`
- `trigger_type`: `manual` or `scheduled`
- `auth_mode`: `none`
- `execution_mode`: `unattended`

Rules:

- Native RSS can be manually triggered.
- Native RSS can be scheduled.
- Native RSS does not require QR or reusable source sessions in phase 1.
- Native RSS outputs enter review before publication.

## 3. Rights Requirements

Administrators must record:

- RSS source URL.
- Rights note.
- Redistribution scope.
- Whether media may be copied into controlled storage.
- Whether external source URLs are provenance-only.

External URLs do not automatically become formal RSS enclosures in Podcast Hub output.

## 4. Job Behavior

Native RSS Import Jobs should record:

- Source ID.
- Program ID.
- Trigger type.
- Import started and finished time.
- Feed URL snapshot.
- Episodes discovered.
- Episodes staged.
- Errors and warnings.

## 5. Output Normalization

Native RSS imported items should normalize into the same Imported Episode shape used by Connector outputs.

Required provenance:

- Source ID.
- Source feed URL.
- Source item GUID, if available.
- Source item URL, if available.
- Import Job ID.

## 6. Non-Goals

Native RSS does not:

- Bypass authentication.
- Circumvent paywalls.
- Circumvent DRM.
- Generate final Podcast Hub RSS directly.
- Skip review.
- Write directly to publication outputs.

