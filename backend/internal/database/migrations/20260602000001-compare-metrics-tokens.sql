ALTER TABLE compare_metrics
    ADD COLUMN IF NOT EXISTS input_tokens  INT              NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS output_tokens INT              NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_usd      DECIMAL(10,6)    NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS latency_ms    INT              NOT NULL DEFAULT 0;
