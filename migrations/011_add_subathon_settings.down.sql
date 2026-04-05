ALTER TABLE overlay_settings
    DROP COLUMN IF EXISTS sub_initial_hours,
    DROP COLUMN IF EXISTS sub_initial_minutes,
    DROP COLUMN IF EXISTS sub_initial_seconds,
    DROP COLUMN IF EXISTS sub_no_border,
    DROP COLUMN IF EXISTS sub_bg_color,
    DROP COLUMN IF EXISTS sub_auto_play,
    DROP COLUMN IF EXISTS sub_text_color,
    DROP COLUMN IF EXISTS sub_font_weight,
    DROP COLUMN IF EXISTS sub_font_content,
    DROP COLUMN IF EXISTS sub_time_rules;
