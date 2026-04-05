ALTER TABLE overlay_settings
    ADD COLUMN IF NOT EXISTS qr_background_color TEXT    DEFAULT '#faae2b',
    ADD COLUMN IF NOT EXISTS qr_barcode_color    TEXT    DEFAULT '#000000',
    ADD COLUMN IF NOT EXISTS qr_label_top        TEXT    DEFAULT '',
    ADD COLUMN IF NOT EXISTS qr_label_bottom     TEXT    DEFAULT '',
    ADD COLUMN IF NOT EXISTS qr_font_family      TEXT    DEFAULT 'default';
