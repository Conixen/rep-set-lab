DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'workouts') THEN
        CREATE TABLE workouts (
            id               BIGSERIAL PRIMARY KEY,
            user_id          BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            prompt           TEXT        NOT NULL DEFAULT '',
            muscle_group     VARCHAR(100) NOT NULL,
            duration_minutes INT         NOT NULL,
            injuries         TEXT,
            goals            TEXT,
            ai_provider      VARCHAR(50) NOT NULL,
            ai_response      JSONB       NOT NULL DEFAULT '{}',
            xp_earned        INT         NOT NULL DEFAULT 0,
            completed_at     TIMESTAMPTZ,
            created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        CREATE INDEX idx_workouts_user_id ON workouts(user_id);
    END IF;
END $$;
