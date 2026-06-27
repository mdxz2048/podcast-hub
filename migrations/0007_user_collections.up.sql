CREATE TABLE IF NOT EXISTS user_collections (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_collections_user ON user_collections(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS user_collection_programs (
  collection_id UUID NOT NULL REFERENCES user_collections(id) ON DELETE CASCADE,
  program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY(collection_id, program_id)
);
CREATE INDEX IF NOT EXISTS idx_user_collection_programs_program ON user_collection_programs(program_id);
