ALTER TABLE overlay_settings
    DROP COLUMN IF EXISTS ms_title,
    DROP COLUMN IF EXISTS ms_target,
    DROP COLUMN IF EXISTS ms_start_date,
    DROP COLUMN IF EXISTS ms_bg_color,
    DROP COLUMN IF EXISTS ms_text_color_ms,
    DROP COLUMN IF EXISTS ms_no_border_ms,
    DROP COLUMN IF EXISTS ms_font_weight_ms,
    DROP COLUMN IF EXISTS ms_font_title,
    DROP COLUMN IF EXISTS ms_font_content;
