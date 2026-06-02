DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'workout_exercises') THEN
        CREATE TABLE workout_exercises (
            id               BIGSERIAL PRIMARY KEY,
            workout_id       BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
            exercise_name    VARCHAR(200) NOT NULL,
            sets             INT,
            reps             INT,
            duration_seconds INT,
            rest_seconds     INT,
            sort_order       INT NOT NULL DEFAULT 0
        );
        CREATE INDEX idx_workout_exercises_workout_id ON workout_exercises(workout_id);
    END IF;
END $$;
