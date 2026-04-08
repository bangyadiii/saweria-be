ALTER TABLE overlay_settings
    ADD COLUMN mabar_enabled          BOOLEAN      NOT NULL DEFAULT FALSE,
    ADD COLUMN mabar_keyword          VARCHAR(50)  NOT NULL DEFAULT '!mabar',
    ADD COLUMN mabar_minimum_amount   BIGINT       NOT NULL DEFAULT 10000,
    ADD COLUMN mabar_gold_threshold   BIGINT       NOT NULL DEFAULT 50000,
    ADD COLUMN mabar_silver_threshold BIGINT       NOT NULL DEFAULT 20000;
