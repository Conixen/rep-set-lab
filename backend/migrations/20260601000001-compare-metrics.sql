CREATE TABLE IF NOT EXISTS compare_metrics (
    id                   BIGSERIAL PRIMARY KEY,
    session_id           TEXT NOT NULL,
    user_id              BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider             TEXT NOT NULL,
    muscle_group         TEXT NOT NULL,
    duration_minutes     INT NOT NULL,
    environment          TEXT NOT NULL DEFAULT 'gym',
    has_injuries         BOOLEAN NOT NULL DEFAULT FALSE,
    library_match_rate   DOUBLE PRECISION NOT NULL DEFAULT 0,
    library_match_count  INT NOT NULL DEFAULT 0,
    library_total_count  INT NOT NULL DEFAULT 0,
    char_count           INT NOT NULL DEFAULT 0,
    emoji_count          INT NOT NULL DEFAULT 0,
    equipment_violations INT NOT NULL DEFAULT 0,
    completeness_score   INT NOT NULL DEFAULT 0,
    warm_up_count        INT NOT NULL DEFAULT 0,
    main_count           INT NOT NULL DEFAULT 0,
    cool_down_count      INT NOT NULL DEFAULT 0,
    tips_count           INT NOT NULL DEFAULT 0,
    avg_note_length      DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes_present_rate   DOUBLE PRECISION NOT NULL DEFAULT 0,
    estimated_minutes    DOUBLE PRECISION NOT NULL DEFAULT 0,
    groq_injury_grade    TEXT,
    groq_equipment_grade TEXT,
    groq_goal_grade      TEXT,
    groq_feedback        TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_compare_metrics_session_id ON compare_metrics(session_id);
CREATE INDEX IF NOT EXISTS idx_compare_metrics_provider   ON compare_metrics(provider);
CREATE INDEX IF NOT EXISTS idx_compare_metrics_user_id    ON compare_metrics(user_id);
