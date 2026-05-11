DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'exercises' AND column_name = 'thumbnail_url'
    ) THEN
        ALTER TABLE exercises ADD COLUMN thumbnail_url VARCHAR(500);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'exercises' AND column_name = 'gif_url'
    ) THEN
        ALTER TABLE exercises ADD COLUMN gif_url VARCHAR(500);
    END IF;
END $$;
