ALTER TABLE overlay_settings
    DROP COLUMN IF EXISTS qr_background_color,
    DROP COLUMN IF EXISTS qr_barcode_color,
    DROP COLUMN IF EXISTS qr_label_top,
    DROP COLUMN IF EXISTS qr_label_bottom,
    DROP COLUMN IF EXISTS qr_font_family;
