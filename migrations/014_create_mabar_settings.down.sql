ALTER TABLE overlay_settings
    DROP COLUMN IF EXISTS mabar_enabled,
    DROP COLUMN IF EXISTS mabar_keyword,
    DROP COLUMN IF EXISTS mabar_minimum_amount,
    DROP COLUMN IF EXISTS mabar_gold_threshold,
    DROP COLUMN IF EXISTS mabar_silver_threshold;
