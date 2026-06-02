DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users') THEN
        CREATE TABLE users (
            id            BIGSERIAL PRIMARY KEY,
            email         VARCHAR(255) NOT NULL,
            username      VARCHAR(100) NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            xp            BIGINT NOT NULL DEFAULT 0,
            level         INT    NOT NULL DEFAULT 1,
            created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        CREATE UNIQUE INDEX idx_users_email    ON users(email);
        CREATE UNIQUE INDEX idx_users_username ON users(username);
    END IF;
END $$;
