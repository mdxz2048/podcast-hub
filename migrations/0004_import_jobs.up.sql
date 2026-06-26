CREATE TABLE IF NOT EXISTS import_jobs (
  id UUID PRIMARY KEY,
  connector_source_id UUID NOT NULL REFERENCES connector_sources(id) ON DELETE RESTRICT,
  connector_version_id UUID NOT NULL REFERENCES connector_versions(id) ON DELETE RESTRICT,
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed', 'cancelled')),
  requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
  trigger_type TEXT NOT NULL CHECK (trigger_type IN ('manual')),
  auth_mode TEXT NOT NULL CHECK (auth_mode IN ('none', 'reusable_session')),
  execution_mode TEXT NOT NULL CHECK (execution_mode IN ('unattended')),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  timeout_at TIMESTAMPTZ,
  cancellation_requested_at TIMESTAMPTZ,
  failure_code TEXT,
  failure_message_redacted TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_import_jobs_created ON import_jobs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_import_jobs_source ON import_jobs (connector_source_id);
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs (status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_import_jobs_one_active_per_source
  ON import_jobs (connector_source_id)
  WHERE status IN ('queued', 'running');

CREATE TABLE IF NOT EXISTS import_job_events (
  id UUID PRIMARY KEY,
  import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  level TEXT NOT NULL CHECK (level IN ('debug', 'info', 'warning', 'error')),
  message_redacted TEXT NOT NULL,
  metadata_redacted JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_import_job_events_job_created ON import_job_events (import_job_id, created_at ASC);

CREATE TABLE IF NOT EXISTS import_job_artifacts (
  id UUID PRIMARY KEY,
  import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
  artifact_type TEXT NOT NULL,
  relative_path TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  sha256 TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_import_job_artifacts_job_created ON import_job_artifacts (import_job_id, created_at ASC);
