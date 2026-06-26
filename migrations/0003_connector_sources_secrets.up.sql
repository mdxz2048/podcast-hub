CREATE TABLE IF NOT EXISTS connector_sources (
  id UUID PRIMARY KEY,
  connector_version_id UUID NOT NULL REFERENCES connector_versions(id) ON DELETE RESTRICT,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK (status IN ('draft', 'active', 'disabled')),
  trigger_type TEXT NOT NULL CHECK (trigger_type IN ('manual')),
  auth_mode TEXT NOT NULL CHECK (auth_mode IN ('none', 'reusable_session')),
  execution_mode TEXT NOT NULL CHECK (execution_mode IN ('unattended')),
  config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  network_mode TEXT NOT NULL CHECK (network_mode IN ('disabled', 'trusted_admin')),
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_connector_sources_created ON connector_sources (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_connector_sources_version ON connector_sources (connector_version_id);

CREATE TABLE IF NOT EXISTS secret_records (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  secret_type TEXT NOT NULL CHECK (secret_type IN ('text', 'file')),
  encrypted_payload TEXT NOT NULL,
  encryption_version TEXT NOT NULL,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  rotated_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_secret_records_created ON secret_records (created_at DESC);

CREATE TABLE IF NOT EXISTS source_secret_bindings (
  id UUID PRIMARY KEY,
  connector_source_id UUID NOT NULL REFERENCES connector_sources(id) ON DELETE CASCADE,
  secret_name TEXT NOT NULL,
  secret_record_id UUID NOT NULL REFERENCES secret_records(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (connector_source_id, secret_name)
);
CREATE INDEX IF NOT EXISTS idx_source_secret_bindings_source ON source_secret_bindings (connector_source_id);

CREATE TABLE IF NOT EXISTS source_events (
  id UUID PRIMARY KEY,
  connector_source_id UUID REFERENCES connector_sources(id) ON DELETE CASCADE,
  secret_record_id UUID REFERENCES secret_records(id) ON DELETE SET NULL,
  actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  detail_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_source_events_source_created ON source_events (connector_source_id, created_at DESC);
