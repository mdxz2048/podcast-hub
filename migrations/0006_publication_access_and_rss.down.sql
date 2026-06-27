DROP TABLE IF EXISTS rss_feeds;
DROP TABLE IF EXISTS program_access_grants;

ALTER TABLE media_assets
  DROP CONSTRAINT IF EXISTS media_assets_delivery_status_check;

ALTER TABLE media_assets
  DROP CONSTRAINT IF EXISTS media_assets_status_check;

ALTER TABLE media_assets
  DROP COLUMN IF EXISTS published_at,
  DROP COLUMN IF EXISTS published_storage_key,
  DROP COLUMN IF EXISTS delivery_status;

ALTER TABLE media_assets
  ADD CONSTRAINT media_assets_status_check
  CHECK (status IN ('staged', 'approved', 'published', 'quarantined', 'deleted'));