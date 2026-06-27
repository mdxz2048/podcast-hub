ALTER TABLE media_assets
  DROP CONSTRAINT IF EXISTS media_assets_status_check;

ALTER TABLE media_assets
  ADD COLUMN IF NOT EXISTS delivery_status TEXT NOT NULL DEFAULT 'staged',
  ADD COLUMN IF NOT EXISTS published_storage_key TEXT,
  ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;

ALTER TABLE media_assets
  ADD CONSTRAINT media_assets_status_check
  CHECK (status IN ('staged', 'approved', 'published', 'archived', 'quarantined', 'deleted'));

ALTER TABLE media_assets
  ADD CONSTRAINT media_assets_delivery_status_check
  CHECK (delivery_status IN ('staged', 'approved', 'published', 'archived', 'quarantined', 'deleted'));

UPDATE media_assets
SET delivery_status = status
WHERE delivery_status IS DISTINCT FROM status;

CREATE TABLE IF NOT EXISTS program_access_grants (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('active', 'revoked')),
  granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
  revoked_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_program_access_active
  ON program_access_grants(user_id, program_id)
  WHERE status='active';
CREATE INDEX IF NOT EXISTS idx_program_access_program ON program_access_grants(program_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_program_access_user ON program_access_grants(user_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS rss_feeds (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  token_prefix TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'revoked', 'expired')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_used_at TIMESTAMPTZ,
  rotated_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_rss_feeds_user ON rss_feeds(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rss_feeds_status ON rss_feeds(status, created_at DESC);