DROP TABLE IF EXISTS intake_runs;
DROP TABLE IF EXISTS publication_events;
DROP TABLE IF EXISTS review_items;
DROP TABLE IF EXISTS media_assets;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS program_sources;
DROP TABLE IF EXISTS programs;
ALTER TABLE import_job_artifacts DROP COLUMN IF EXISTS storage_key;
