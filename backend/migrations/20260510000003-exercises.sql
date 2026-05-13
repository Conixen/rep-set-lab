DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exercises') THEN
        CREATE TABLE exercises (
            id           BIGSERIAL PRIMARY KEY,
            name         VARCHAR(200) NOT NULL,
            description  TEXT         NOT NULL DEFAULT '',
            muscle_group VARCHAR(100) NOT NULL,
            difficulty   VARCHAR(50)  NOT NULL,
            equipment    VARCHAR(100) NOT NULL DEFAULT '',
            created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
        );
        CREATE UNIQUE INDEX idx_exercises_name ON exercises(name);
        CREATE INDEX idx_exercises_muscle_group ON exercises(muscle_group);
    END IF;
END $$;
