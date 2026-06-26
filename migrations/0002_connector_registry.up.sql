CREATE TABLE IF NOT EXISTS connectors (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS connector_versions (
  id UUID PRIMARY KEY,
  connector_id UUID NOT NULL REFERENCES connectors(id) ON DELETE CASCADE,
  version TEXT NOT NULL,
  review_status TEXT NOT NULL CHECK (review_status IN ('pending_review', 'approved', 'rejected', 'disabled')),
  runtime_profile TEXT NOT NULL,
  entrypoint TEXT NOT NULL,
  manifest_json JSONB NOT NULL,
  package_sha256 TEXT NOT NULL,
  package_size_bytes BIGINT NOT NULL,
  package_storage_key TEXT NOT NULL,
  validation_summary_json JSONB NOT NULL,
  uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (connector_id, version)
);
CREATE INDEX IF NOT EXISTS idx_connector_versions_connector_created ON connector_versions (connector_id, created_at DESC);

CREATE TABLE IF NOT EXISTS connector_events (
  id UUID PRIMARY KEY,
  connector_id UUID REFERENCES connectors(id) ON DELETE CASCADE,
  connector_version_id UUID REFERENCES connector_versions(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  detail_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_connector_events_connector_created ON connector_events (connector_id, created_at DESC);
