ALTER TABLE overlay_settings
    DROP COLUMN IF EXISTS lb_title,
    DROP COLUMN IF EXISTS lb_bg_color,
    DROP COLUMN IF EXISTS lb_text_color,
    DROP COLUMN IF EXISTS lb_font_weight,
    DROP COLUMN IF EXISTS lb_no_border,
    DROP COLUMN IF EXISTS lb_hide_amount,
    DROP COLUMN IF EXISTS lb_font_title,
    DROP COLUMN IF EXISTS lb_font_content,
    DROP COLUMN IF EXISTS lb_time_range,
    DROP COLUMN IF EXISTS lb_limit;
