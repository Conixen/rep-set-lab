CREATE TABLE IF NOT EXISTS exercise_gif_cache (
    exercisedb_id TEXT        PRIMARY KEY,
    gif_bytes     BYTEA       NOT NULL,
    content_type  TEXT        NOT NULL DEFAULT 'image/gif',
    cached_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
