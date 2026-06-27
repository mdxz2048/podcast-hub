ALTER TABLE import_job_artifacts ADD COLUMN IF NOT EXISTS storage_key TEXT;

CREATE TABLE IF NOT EXISTS programs (
  id UUID PRIMARY KEY,
  canonical_key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  author TEXT NOT NULL,
  language TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('staging', 'review_pending', 'approved', 'published', 'archived', 'rejected')),
  created_from_source_id UUID NOT NULL REFERENCES connector_sources(id) ON DELETE RESTRICT,
  created_from_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_at TIMESTAMPTZ,
  archived_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_programs_status ON programs(status);

CREATE TABLE IF NOT EXISTS program_sources (
  id UUID PRIMARY KEY,
  program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  connector_source_id UUID NOT NULL REFERENCES connector_sources(id) ON DELETE RESTRICT,
  external_program_id TEXT NOT NULL,
  first_import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE RESTRICT,
  last_import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(connector_source_id, external_program_id)
);

CREATE TABLE IF NOT EXISTS episodes (
  id UUID PRIMARY KEY,
  program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  external_episode_id TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  published_at TIMESTAMPTZ NOT NULL,
  duration_seconds INTEGER NOT NULL CHECK (duration_seconds >= 0),
  status TEXT NOT NULL CHECK (status IN ('staging', 'review_pending', 'approved', 'rejected', 'published', 'archived')),
  source_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_to_users_at TIMESTAMPTZ,
  archived_at TIMESTAMPTZ,
  UNIQUE(program_id, external_episode_id)
);
CREATE INDEX IF NOT EXISTS idx_episodes_program ON episodes(program_id);
CREATE INDEX IF NOT EXISTS idx_episodes_status ON episodes(status);

CREATE TABLE IF NOT EXISTS media_assets (
  id UUID PRIMARY KEY,
  owner_type TEXT NOT NULL CHECK (owner_type IN ('program', 'episode')),
  owner_id UUID NOT NULL,
  import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE RESTRICT,
  artifact_id UUID NOT NULL REFERENCES import_job_artifacts(id) ON DELETE RESTRICT,
  media_kind TEXT NOT NULL CHECK (media_kind IN ('audio', 'cover', 'metadata')),
  staged_storage_key TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  sha256 TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('staged', 'approved', 'published', 'quarantined', 'deleted')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_media_assets_owner ON media_assets(owner_type, owner_id);

CREATE TABLE IF NOT EXISTS review_items (
  id UUID PRIMARY KEY,
  target_type TEXT NOT NULL CHECK (target_type IN ('program', 'episode')),
  target_id UUID NOT NULL,
  review_kind TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
  requested_by_job_id UUID REFERENCES import_jobs(id) ON DELETE SET NULL,
  assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  review_note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reviewed_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_review_items_one_pending
  ON review_items(target_type, target_id, review_kind)
  WHERE status='pending';

CREATE TABLE IF NOT EXISTS publication_events (
  id UUID PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
  metadata_redacted JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_publication_events_target ON publication_events(target_type, target_id, created_at);

CREATE TABLE IF NOT EXISTS intake_runs (
  id UUID PRIMARY KEY,
  import_job_id UUID NOT NULL UNIQUE REFERENCES import_jobs(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('succeeded', 'failed')),
  validation_issues_redacted JSONB NOT NULL DEFAULT '[]'::jsonb,
  program_id UUID REFERENCES programs(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
