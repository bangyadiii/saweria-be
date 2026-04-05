ALTER TABLE overlay_settings
    ADD COLUMN IF NOT EXISTS ms_title            TEXT    NOT NULL DEFAULT 'Pengumpulan Dana',
    ADD COLUMN IF NOT EXISTS ms_target           BIGINT  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ms_start_date       DATE,
    ADD COLUMN IF NOT EXISTS ms_bg_color         TEXT    NOT NULL DEFAULT '#faae2b',
    ADD COLUMN IF NOT EXISTS ms_text_color_ms    TEXT    NOT NULL DEFAULT '#333333',
    ADD COLUMN IF NOT EXISTS ms_no_border_ms     BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS ms_font_weight_ms   INTEGER NOT NULL DEFAULT 400,
    ADD COLUMN IF NOT EXISTS ms_font_title       TEXT    NOT NULL DEFAULT 'default',
    ADD COLUMN IF NOT EXISTS ms_font_content     TEXT    NOT NULL DEFAULT 'default';
