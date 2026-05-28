DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'exercises' AND column_name = 'aliases'
    ) THEN
        ALTER TABLE exercises ADD COLUMN aliases TEXT[] NOT NULL DEFAULT '{}';
    END IF;
END $$;
