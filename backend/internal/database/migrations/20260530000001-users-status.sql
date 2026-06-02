DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'users' AND column_name = 'status'
  ) THEN
    ALTER TABLE users ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending';
    -- Existing accounts are already known users — activate them immediately.
    UPDATE users SET status = 'active';
  END IF;
END $$;
