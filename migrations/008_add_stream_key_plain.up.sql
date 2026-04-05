CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE overlay_settings ADD COLUMN IF NOT EXISTS stream_key TEXT UNIQUE;

-- Backfill: generate a random stream key for existing rows that have none
UPDATE overlay_settings
SET stream_key = encode(gen_random_bytes(16), 'hex')
WHERE stream_key IS NULL;
